package main

func main() {
	// 启动服务器
	server := NewServer("127.0.0.1", 8080)
	server.Start()
}
