package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int //当前客户端模式
}

func (c *Client) menu() bool {
	var flag int
	fmt.Println("1.public chat mode")
	fmt.Println("2.private chat mode")
	fmt.Println("3.update username")
	fmt.Println("0.exit")
	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		c.flag = flag
		return true
	} else {
		fmt.Println("invalid mod")
		return false
	}
}

// 处理server回应的消息，直接显示到标准输出上
func (c *Client) DealResponse() {
	//一旦client.conn有数据，就直接copy到stdout标准输出上，永久阻塞监听
	io.Copy(os.Stdout, c.conn)
}

// 1.公聊模式
func (c *Client) PublicChat() {
	var chatMsg string
	fmt.Println("please input chat content, exit by inputting 'exit'")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := c.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write error:", err)
				return
			}
		}
		chatMsg = ""
		fmt.Println("please input chat content, exit by inputting 'exit'")
		fmt.Scanln(&chatMsg)
	}
}

// 2.私聊模式
// 查询在线用户
func (c *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write error:", err)
		return
	}
}

// 发送私聊消息
func (c *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	c.SelectUsers()
	fmt.Println("please input the username you want to chat with, exit by inputting 'exit'")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println("please input chat content, exit by inputting 'exit'")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := c.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn.Write error:", err)
					return
				}
			}
			chatMsg = ""
			fmt.Println("please input chat content, exit by inputting 'exit'")
			fmt.Scanln(&chatMsg)
		}
		c.SelectUsers()
		fmt.Println("please input the username you want to chat with, exit by inputting 'exit'")
		fmt.Scanln(&remoteName)
	}
}

// 3.更新用户名
func (c *Client) UpdateName() bool {
	fmt.Println("please input your username:")
	fmt.Scanln(&c.Name)
	sendMsg := "rename|" + c.Name + "\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write error:", err)
		return false
	}
	return true
}

func (c *Client) Run() {
	for c.flag != 0 {
		for c.menu() != true {
		}
		switch c.flag {
		case 1:
			c.PublicChat()
			break
		case 2:
			c.PrivateChat()
			break
		case 3:
			c.UpdateName()
			break
		}
	}
}

// 创建一个客户端的API
func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn
	return client
}

var serverIp string
var serverPort int

// 在init函数中接收命令行参数
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "server ip address")
	flag.IntVar(&serverPort, "port", 8080, "server port")
}

func main() {
	//命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("connection to the server failed...")
		return
	}

	go client.DealResponse() //单独开启一个goroutine处理server的回执消息

	fmt.Println("connection to the server success...")

	client.Run()
}
