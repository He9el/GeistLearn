package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type Client struct {
	ID   int
	Conn net.Conn
}

type ConnectionManager struct {
	clients    map[int]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	nextID     int
	mu         sync.Mutex
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		clients:    make(map[int]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}
}

func (m *ConnectionManager) Run() {
	for {
		select {
		case client := <-m.register:
			m.clients[client.ID] = client
			log.Printf("register client, id: %d\n", client.ID)
		case client := <-m.unregister:
			delete(m.clients, client.ID)
			log.Printf("unregister client, id: %d\n", client.ID)
		case data := <-m.broadcast:
			for _, client := range m.clients {
				_, err := client.Conn.Write(data)
				if err != nil {
					log.Printf("Error writing to client %d: %v\n", client.ID, err)
				}
			}
		}
	}
}

func (m *ConnectionManager) BroadcastMessage(data []byte) {
	m.broadcast <- data
}

func (m *ConnectionManager) GetNextID() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	return m.nextID
}

// HandleNewConnection 负责处理新的网络连接，创建客户端并启动其处理goroutine
func (m *ConnectionManager) HandleNewConnection(conn net.Conn) {
	client := &Client{
		ID:   m.GetNextID(),
		Conn: conn,
	}
	m.register <- client

	go handleConnection(m, client)
}

func main() {
	// 创建连接管理器
	manager := NewConnectionManager()
	go manager.Run()

	// 启动TCP服务器
	startListener(manager)
}

func startListener(manager *ConnectionManager) {
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatalf("listen failed, err: %v\n", err)
	}
	fmt.Println("listen success on 127.0.0.1:8080")
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept conn failed, err: %v\n", err)
			continue
		}
		fmt.Printf("accept conn success, remote addr: %v\n", conn.RemoteAddr())

		// 将连接处理的职责交给 ConnectionManager
		manager.HandleNewConnection(conn)
	}
}

func handleConnection(manager *ConnectionManager, client *Client) {
	defer client.Conn.Close()
	defer func() {
		manager.unregister <- client
	}()

	buf := make([]byte, 1024)
	for {
		n, err := client.Conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("Client %d disconnected\n", client.ID)
			} else {
				log.Printf("read data from client %d failed, err: %v\n", client.ID, err)
			}
			break
		}

		data := buf[:n]
		log.Printf("read data from client %d, data: %s\n", client.ID, string(data))

		response := processRequest(data)
		manager.BroadcastMessage(response)
		log.Printf("broadcasted response: %s\n", string(response))
	}
}

func processRequest(data []byte) []byte {
	// 实际的业务逻辑处理
	return []byte(fmt.Sprintf("learn1! (received: %s)\n", string(data)))
}
