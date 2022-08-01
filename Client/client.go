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
	mode       int // 当前client的模式
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端的对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		mode:       999,
	}

	// 连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn

	// 返回对象
	return client
}

// DealResponse 处理server回应的消息，直接显示到标准输出即可
func (client *Client) DealResponse() {
	// 一旦client.conn有数据，就直接copy到stdout标准输出上，永久阻塞监听
	io.Copy(os.Stdout, client.conn)
	// 等价写法啊
	//for {
	//	buf := make([]byte, 4096)
	//	client.conn.Read(buf)
	//	fmt.Println(buf)
	//}
}

// 菜单显示
func (client *Client) menu() bool {

	var mode int
	fmt.Println("1.PublicChatMode")
	fmt.Println("2.PrivateChatMode")
	fmt.Println("3.UpdateUsername")
	fmt.Println("0.Quit")

	fmt.Scanln(&mode)

	if mode >= 0 && mode <= 3 {
		client.mode = mode
		return true
	} else {
		fmt.Println(">>>>>> Please check if your input is valid <<<<<<")
		return false
	}
}

// UpdateName 更新用户名业务
func (client *Client) UpdateName() bool {
	fmt.Println(">>>>>>Please enter user name:")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

// PublicChat 公聊业务
func (client *Client) PublicChat() {
	var chatMsg string
	fmt.Println(">>>>>>Please enter chat content or 'exit' quit")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		// 发送给服务器

		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>>>Please enter chat content or 'exit' to quit")
		fmt.Scanln(&chatMsg)
	}
}

// SelectUsers 查询在线用户
func (client *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write err:", err)
		return
	}
}

// PrivateChat 私聊业务
func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string
	client.SelectUsers()
	fmt.Println(">>>>>>Please enter the chat partner's[username] or 'exit' to quit")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>>>Please enter the message content or 'exit' to quit")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err:", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println(">>>>>>Please enter the message content or 'exit' to quit")
			fmt.Scanln(&chatMsg)
		}
		client.SelectUsers()
		fmt.Println(">>>>>>Please enter the chat partner's[username] or 'exit' to quit")
		fmt.Scanln(&remoteName)
	}
}

// Run 业务入口
func (client *Client) Run() {
	for client.mode != 0 {
		for client.menu() != true {
		}

		// 根据不同模式处理不同业务
		switch client.mode {
		case 1:
			// 公聊
			//fmt.Println("Public chat mode selection...")
			client.PublicChat()
			break
		case 2:
			// 私聊
			//fmt.Println("Private chat mode selection...")
			client.PrivateChat()
			break
		case 3:
			// 更新用户名
			//fmt.Println("Update username selection...")
			client.UpdateName()
			break
		}

	}
}

var serverIp string
var serverPort int

// ./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "Set the server IP address(Default:127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "Set server port(Default:8888)")
}

func main() {

	// 解析命令行
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>> Connection failure...")
		return
	}

	fmt.Println(">>>>>> Connection success...")

	// 单独开启一个goroutine处理server的回执消息
	go client.DealResponse()

	// 启动客户端业务
	client.Run()
}
