package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	//在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播的channel
	Message chan string
}

//创建一个server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

//监听Message 广播消息 channel 的 goroutine， 一旦有消息就发送给在线的用户
func (this *Server) ListenMessager() {
	fmt.Println("// ListenMessager ...")
	for {
		msg := <-this.Message
		fmt.Println("// sendMsg to onlineMap")
		//将msg 发送给全部在线的User
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			fmt.Println("// range OnlineMap ...")
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

//广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	fmt.Println("// BroadCast ...")
	sendMsg := "[" + user.Name + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	fmt.Println("// do Handler ...")
	// ...当前链接的业务
	// fmt.Println("连接建立成功")

	user := NewUser(conn)

	//用户上线，将用户加入到 onlineMap 中
	this.mapLock.Lock()
	fmt.Println("// new user, and add to map")
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()

	//广播当前用户上线消息
	fmt.Println("// BroadCast user Online ...")
	this.BroadCast(user, "已上线")

	//接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				this.BroadCast(user, "下线")
				return
			}

			if err != nil {
				fmt.Println("Conn Read err:", err)
				return
			}

			//提取用户消息（去掉‘\n’）
			msg := string(buf[:n-1])
			this.BroadCast(user, msg)
		}
	}()
	//当前 handler阻塞
	select {}
}

//启动服务器的接口
func (this *Server) Start() {
	// TODO: socket listen
	fmt.Println("// Server starting ...")
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	// TODO: close socket listen
	defer listener.Close()
	defer fmt.Println("// close socket listen")

	//启动监听 Message 的 goroutine
	go this.ListenMessager()

	for {
		// TODO: accept socket
		fmt.Println("// accept socket ...")
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		// TODO: do handler
		go this.Handler(conn)
	}
}
