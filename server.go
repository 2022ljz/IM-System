package main

import (
	"fmt"
	"net"
)

type Server struct {
	IP   string
	Port int
}

func NewServer(ip string, port int) *Server {
	return &Server{
		IP:   ip,
		Port: port,
	}
}

func (s *Server) Handler(conn net.Conn) {
	// 处理连接
	fmt.Println("连接建立成功")
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

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept error:", err)
			continue
		}

		//do handler
		go s.Handler(conn)
	}
}
