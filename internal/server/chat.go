package server

import (
	"fmt"
	"log"
	"net"
	"sync"
)

type ChatServer struct {
	address  string
	clients  map[*Client]bool
	incoming chan string
	join     chan *Client
	leave    chan *Client
	mu       sync.RWMutex
}

func NewChatServer(address string) *ChatServer {
	return &ChatServer{
		address:  address,
		clients:  make(map[*Client]bool),
		incoming: make(chan string),
		join:     make(chan *Client),
		leave:    make(chan *Client),
		mu:       sync.RWMutex{},
	}
}

func (s *ChatServer) Run() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}
	defer func(listener net.Listener) {
		_ = listener.Close()
	}(listener)

	go s.broadcast()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("error accepting connection:", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *ChatServer) handleConnection(conn net.Conn) {
	client := &Client{
		conn:     conn,
		name:     conn.RemoteAddr().String(),
		outgoing: make(chan string),
	}

	s.join <- client

	go client.write()
	client.read(s)

	s.leave <- client
	_ = conn.Close()
}

func (s *ChatServer) broadcast() {
	for {
		select {
		case msg := <-s.incoming:
			s.mu.RLock()
			for client := range s.clients {
				client.outgoing <- msg
			}
			s.mu.RUnlock()
		case client := <-s.join:
			s.mu.Lock()
			s.clients[client] = true
			s.mu.Unlock()
			s.incoming <- fmt.Sprintf("%s joined the chat", client.name)
		case client := <-s.leave:
			s.mu.Lock()
			delete(s.clients, client)
			close(client.outgoing)
			s.mu.Unlock()
			s.incoming <- fmt.Sprintf("%s left the chat", client.name)
		}
	}
}
