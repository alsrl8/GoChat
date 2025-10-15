package server

import (
	"bufio"
	"fmt"
	"net"
)

type Client struct {
	conn     net.Conn
	name     string
	outgoing chan string
}

func (c *Client) write() {
	for msg := range c.outgoing {
		_, _ = fmt.Fprintln(c.conn, msg)
	}
}

func (c *Client) read(s *ChatServer) {
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		msg := fmt.Sprintf("%s: %s", c.name, scanner.Text())
		s.incoming <- msg
	}
}
