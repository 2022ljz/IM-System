package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	//直接监听当前user channel的消息的goroutine
	go user.ListenMessage()

	return user
}

// 用户上线
func (u *User) Online() {
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()
	//广播当前用户上线消息
	u.server.BroadCast(u, "is online")
}

// 用户下线
func (u *User) Offline() {
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()
	//广播当前用户下线消息
	u.server.BroadCast(u, "is offline")
}

// 用户消息发送
func (u *User) DoMsg(msg string) {
	//查询在线用户
	if msg == "who" {
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "is online...\n"
			u.conn.Write([]byte(onlineMsg)) //给当前user对应的客户端发消息
		}
		u.server.mapLock.Unlock()

		//修改用户名，格式：rename|张三
	} else if len(msg) > 7 && msg[0:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]
		_, ok := u.server.OnlineMap[newName]
		if ok {
			u.conn.Write([]byte("the username is already exist\n"))
		} else {
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name) //删除原来的用户名
			u.server.OnlineMap[newName] = u
			u.server.mapLock.Unlock()

			u.Name = newName
			u.conn.Write([]byte("username changed to " + newName + "\n"))
		}

		//私聊功能 to|张三|消息内容
	} else if len(msg) > 4 && msg[:3] == "to|" {
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			u.conn.Write([]byte("the username cannot be empty\n"))
			return
		}

		remoteUser, ok := u.server.OnlineMap[remoteName]
		if !ok {
			u.conn.Write([]byte("the username does not exist\n"))
			return
		}

		content := strings.Split(msg, "|")[2]
		if content == "" {
			u.conn.Write([]byte("the message cannot be empty\n"))
			return
		}

		remoteUser.conn.Write([]byte(u.Name + " says to you: " + content + "\n"))

	} else {
		u.server.BroadCast(u, "say:"+msg)
	}

}

// 监听当前User Channel的方法，一旦有消息就直接发送给对端客户端
func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}
