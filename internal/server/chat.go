package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
)

// ChatServer is a simple TCP chat server that broadcasts messages to all connected clients.
type ChatServer struct {
	addr      string
	listener  net.Listener
	clients   map[net.Conn]struct{}
	mu        sync.RWMutex
	shutdownC chan struct{}
}

func New(addr string) *ChatServer {
	return &ChatServer{
		addr:      addr,
		clients:   make(map[net.Conn]struct{}),
		shutdownC: make(chan struct{}),
	}
}

// Start begins listening and serving new connections. It blocks until shutdown is triggered.
func (s *ChatServer) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.addr, err)
	}
	s.listener = ln
	log.Printf("chat server listening on %s", s.addr)

	// Handle Ctrl+C for a graceful shutdown inside Start for simplicity.
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt)
	go func() {
		<-sigC
		log.Println("interrupt signal received, shutting down...")
		s.Shutdown()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-s.shutdownC:
				return nil // listener closed due to shut down
			default:
				// transient accept error
				return fmt.Errorf("accept error: %w", err)
			}
		}
		go s.handleConn(conn)
	}
}

func (s *ChatServer) handleConn(c net.Conn) {
	remote := c.RemoteAddr().String()
	log.Printf("client connected: %s", remote)

	s.mu.Lock()
	s.clients[c] = struct{}{}
	s.mu.Unlock()

	s.broadcast(fmt.Sprintf("[server] %s joined\n", remote), c)

	scanner := bufio.NewScanner(c)
	for scanner.Scan() {
		msg := scanner.Text()
		msg = strings.TrimSpace(msg)
		if msg == "" {
			continue
		}
		if strings.EqualFold(msg, "/quit") {
			break
		}
		// broadcast with prefix
		s.broadcast(fmt.Sprintf("[%s] %s\n", remote, msg), c)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("read error from %s: %v", remote, err)
	}

	s.mu.Lock()
	delete(s.clients, c)
	s.mu.Unlock()
	_ = c.Close()
	s.broadcast(fmt.Sprintf("[server] %s left\n", remote), c)
	log.Printf("client disconnected: %s", remote)
}

func (s *ChatServer) broadcast(msg string, except net.Conn) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for c := range s.clients {
		if except != nil && c == except {
			continue
		}
		_, _ = c.Write([]byte(msg))
	}
}

// Shutdown gracefully closes the listener and disconnects clients.
func (s *ChatServer) Shutdown() {
	select {
	case <-s.shutdownC:
		return
	default:
		close(s.shutdownC)
	}
	if s.listener != nil {
		_ = s.listener.Close()
	}

	s.mu.Lock()
	for c := range s.clients {
		_ = c.Close()
	}
	s.clients = make(map[net.Conn]struct{})
	s.mu.Unlock()
	log.Println("chat server stopped")
}
