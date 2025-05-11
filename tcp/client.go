package tcp

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type Client struct {
	Conn       net.Conn
	Username   string
	writer     *syncWriter
	LastMessage time.Time
	mu         sync.Mutex
	closed     bool
}

type syncWriter struct {
	sync.Mutex
	conn net.Conn
}

func NewClient(conn net.Conn) *Client {
	conn.SetDeadline(time.Time{})
	return &Client{
		Conn:   conn,
		writer: &syncWriter{conn: conn},
	}
}

func (c *Client) prompt(text string) error {
	return c.writer.write(text)
}

func (c *Client) SendMessage(msg string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return fmt.Errorf("connection closed")
	}
	return c.writer.writeMessage(msg)
}

func (c *Client) ReadInput() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return "", fmt.Errorf("connection closed")
	}
	
	input, err := bufio.NewReader(c.Conn).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		c.Conn.Close()
		c.closed = true
	}
}

func (w *syncWriter) write(text string) error {
	w.Lock()
	defer w.Unlock()
	
	if _, err := w.conn.Write([]byte("\033[2K\r")); err != nil {
		return err
	}
	
	if _, err := w.conn.Write([]byte(text)); err != nil {
		return err
	}
	
	return nil
}

func (w *syncWriter) writeMessage(msg string) error {
	w.Lock()
	defer w.Unlock()
	
	if _, err := w.conn.Write([]byte("\033[2K\r")); err != nil {
		return err
	}
	if _, err := w.conn.Write([]byte(msg + "\n")); err != nil {
		return err
	}
	_, err := w.conn.Write([]byte("> "))
	return err
}