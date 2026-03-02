package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	IP        string
	Port      int
	OnlineMap map[string]*User //在线用户列表
	mapLock   sync.RWMutex     //读写锁
	Message   chan string      //消息广播channel
}

// 创建一个服务器的API
func NewServer(ip string, port int) *Server {
	return &Server{
		IP:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}

func (s *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		fmt.Println("net.Listen error:", err)
		return
	}
	// 为了防止忘关，直接把close放到defer里
	defer listener.Close()

	//启动监听Message的goroutine
	go s.ListenMessager()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept error:", err)
			continue
		}

		// do handler
		go s.Handler(conn)
	}
}

// 监听Message广播消息channel的goroutine，一旦有消息就发送给全部在线User
func (s *Server) ListenMessager() {
	for {
		msg := <-s.Message
		s.mapLock.Lock()
		//广播给用户列表中所有的用户
		for _, u := range s.OnlineMap {
			u.C <- msg
		}
		s.mapLock.Unlock()
	}
}

// 向Message写入广播消息
func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Message <- sendMsg
}

func (s *Server) Handler(conn net.Conn) {
	user := NewUser(conn, s)
	//用户上线，将用户加入在线用户列表
	user.Online()

	//监听用户是否活跃的channel
	isLive := make(chan bool)

	//接受用户消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			//用户下线，将用户从在线用户列表删除
			if n == 0 {
				user.Offline()
				return
			}

			//捕捉异常
			if err != nil && err != io.EOF {
				fmt.Println("conn.Read error:", err)
				return
			}

			//处理用户消息
			msg := strings.TrimSpace(string(buf[:n])) //把客户端发来的原始字节 → 转成字符串 → 去掉首尾空白 → 得到干净的一条消息
			user.DoMsg(msg)
			//用户任意消息代表当前用户活跃
			isLive <- true
		}
	}()

	//超时强踢功能
	for {
		select {
		//如果当前用户活跃，则重置定时器
		case <-isLive:
		//什么都不做，只需要进入下一轮循环从而自动重置定时器

		//10秒后管道中会有数据，此时就会进入该case，强制当前用户下线
		case <-time.After(time.Second * 10):
			user.conn.Write([]byte("you have been removed\n"))
			close(user.C)
			conn.Close()
			return
		}

	}

}
