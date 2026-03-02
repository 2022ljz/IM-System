package main

import (
	"fmt"
	"net"
	"sync"
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
	//用户上线，将用户加入在线用户列表
	user := NewUser(conn)
	s.mapLock.Lock()
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock()
	//广播当前用户上线消息
	s.BroadCast(user, "已上线")

	//阻塞handler，防退出
	select {}

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
