package main

import (
	"net"
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