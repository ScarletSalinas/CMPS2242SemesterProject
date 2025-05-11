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
    benchmarkMode bool
}

// newClient creates and initializes a new Client
func NewClient(conn net.Conn) *Client {
    conn.SetDeadline(time.Time{})
    return &Client{
        Conn:   conn,
        writer: &syncWriter{conn: conn, benchmarkMode: false},
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
// Simplified writeMessage for benchmarks
func (w *syncWriter) writeMessage(msg string) error {
    w.Lock()
    defer w.Unlock()
    
    if w.benchmarkMode {
        // Bypass all ANSI formatting for benchmarks
        _, err := w.conn.Write([]byte(msg + "\n"))
        return err
    }
    
    // Original ANSI-formatted writing
    if _, err := w.conn.Write([]byte("\033[2K\r")); err != nil {
        return err
    }
    if _, err := w.conn.Write([]byte(msg + "\n")); err != nil {
        return err
    }
    _, err := w.conn.Write([]byte("> "))
    return err
}

