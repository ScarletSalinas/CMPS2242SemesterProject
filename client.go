package main

import (
	"bufio"
	"net"
	"strings"
	"sync"
)

// Client represents a connected chat user
type Client struct {
	Conn     net.Conn
	Username string
	writer   *syncWriter // private thread-safe writer
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

// write clears the line and writes text (thread-safe)
func (w *syncWriter) write(text string) error {
	w.Lock()
	defer w.Unlock()
	
	// Clear current line
	if _, err := w.conn.Write([]byte("\033[2K\r")); err != nil {
		return err
	}
	
	// Write the actual text
	if _, err := w.conn.Write([]byte(text)); err != nil {
		return err
	}
	
	return nil
}

// newClient creates and initializes a new Client
func newClient(conn net.Conn) *Client {
	return &Client{
		Conn:   conn,
		writer: newSyncWriter(conn),
	}
}

// sendMessage safely writes a message to the client's connection
func (c *Client) sendMessage(msg string) error {
	return c.writer.write(msg)
}

// prompt sends a prompt to the client (e.g., for username)
func (c *Client) prompt(text string) error {
	return c.sendMessage(text)
}

// readInput reads a line of input from the client
func (c *Client) readInput() (string, error) {
	reader := bufio.NewReader(c.Conn)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}