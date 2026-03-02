package main

import "net"

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

//用户上线
func (u *User) Online() {
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()
	//广播当前用户上线消息
	u.server.BroadCast(u, "is online")
}

//用户下线
func (u *User) Offline() {
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()
	//广播当前用户下线消息
	u.server.BroadCast(u, "is offline")
}

//用户消息发送
func (u *User) DoMsg(msg string) {
	u.server.BroadCast(u, "say:"+msg)
}

//监听当前User Channel的方法，一旦有消息就直接发送给对端客户端
func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}
