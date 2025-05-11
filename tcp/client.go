package tcp

import (
	"bufio"
	"fmt"
	"log"
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
	mu       sync.Mutex // Add this
    closed   bool       // Add connection state flag
}

func (c *Client) Close() {
    c.mu.Lock()
    defer c.mu.Unlock()
    if !c.closed {
        c.Conn.Close()
        c.closed = true
    }
}

func (c *Client) IsConnected() bool {
    c.mu.Lock()
    defer c.mu.Unlock()
    return !c.closed && c.Conn != nil
}

// syncWriter provides thread-safe writing with line clearing
type syncWriter struct {
	sync.Mutex
	conn net.Conn
}

// Creates a new thread-safe writer for a connection
func newSyncWriter(conn net.Conn) *syncWriter {
	return &syncWriter{conn: conn}
}

// Method for prompts-no newline
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

// newClient creates and initializes a new Client
func newClient(conn net.Conn) *Client {
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
func (c *Client) sendMessage(msg string) error {
	return c.writer.writeMessage(msg)
}

// readInput reads a line of input from the client
func (c *Client) readInput() (string, error) {
	reader := bufio.NewReader(c.Conn)
	input, err := reader.ReadString('\n')
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			log.Printf("Timeout: %v", c.Username)	// Network error
		}
		return "", err
	}
	return strings.TrimSpace(input), nil
}
