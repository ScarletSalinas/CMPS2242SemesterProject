package tcp

import (
		"bufio"
		"fmt"
		"net"
		"strings"
		"sync"
		"time"
)

// Client represents a connected chat user
type Client struct {
	Conn     net.Conn
	Username string
	writer   *syncWriter // private thread-safe writer
	LastMessage time.Time
	mu sync.Mutex
	closed bool
}

// syncWriter provides thread-safe writing with line clearing
type syncWriter struct {
	sync.Mutex
	conn net.Conn
}

// newClient creates and initializes a new Client
func NewClient(conn net.Conn) *Client {
	return &Client{
		Conn:   conn,
		writer: newSyncWriter(conn),
	}
}

// prompt sends a prompt to the client 
func (c *Client) prompt(text string) error {
	return c.writer.write(text)
}

// sendMessage safely writes a message to the client's connection
func (c *Client) SendMessage(msg string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return fmt.Errorf("connection closed")
	}
	return c.writer.writeMessage(msg)
}

// readInput reads a line of input from the client
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

// Creates a new thread-safe writer for a connection
func newSyncWriter(conn net.Conn) *syncWriter {
	return &syncWriter{conn: conn}
}


// Prompts-no newline
func (w *syncWriter) write(text string) error {
	w.Lock()
	defer w.Unlock()
	
	// Clear current line
	if _, err := w.conn.Write([]byte("\033[2K\r")); err != nil {
		return err
	}
	
	if _, err := w.conn.Write([]byte(text)); err != nil {
		return err
	}
	
	return nil
}

// Method for chat messages-adds newline
func (w *syncWriter) writeMessage(msg string) error {
	w.Lock()
	defer w.Unlock()

	if _, err := w.conn.Write([]byte("\033[2K\r")); err != nil {
		return fmt.Errorf("clear failed: %w", err)
	}
	
	// Clear line, write message, then newline
	if _, err := w.conn.Write([]byte("\033[2K\r" + msg + "\n")); err != nil {
		return err
	}
	
	// Re-draw prompt if needed
	_, err := w.conn.Write([]byte("> "))
	return err
}

