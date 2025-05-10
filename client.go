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