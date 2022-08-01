package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// NewServer 创建一个Server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// ListenMessage 监听Message广播消息channel的goroutine
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message
		// 将msg发送给全部的在线User
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// BroadCast 广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	//...当前连接的业务
	fmt.Println("Connection established successfully")

	user := NewUser(conn, this)
	user.Online()

	// 监听用户是否活跃的channel
	isLive := make(chan bool)

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			// 提取客户端发送的消息（去除\n）
			msg := string(buf[:n-1])

			// 对msg进行处理
			user.DoMessage(msg)

			// 用户在处理消息，活跃
			isLive <- true
		}
	}()

	// 当前handler阻塞

	for {
		select {
		case <-isLive:
			// 当前用户是活跃的，应该重置定时器
			// 不做任何事情，为了激活select，更新下面的的定时器（条件case会执行但不会进去）
		case <-time.After(time.Second * 1000):
			// 已经超时
			// 将User强制关闭
			user.SendMsg("The current user timeout")

			// 销毁用户数据
			close(user.C)

			// 关闭连接
			conn.Close()

			// 退出
			return
		}
	}

}

// Start 启动服务器的接口
func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err", err)
		return
	}
	// close listen socket
	defer listener.Close()

	// 启动监听Message的goroutine
	go this.ListenMessage()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err", err)
			continue
		}

		// do handler
		go this.Handler(conn)
	}

}
