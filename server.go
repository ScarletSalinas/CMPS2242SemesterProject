package main

import(
		"log"
		"net"
		"strings"
		"time"
)

// NewServer creates a new chat server instance
func NewServer() *Server {
	return &Server{
		clients: make(map[*Client]bool),
	}
}

// Start begins listening for connections
func (s *Server) Start(port string) error {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("Chat server running on %s", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Connection error: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

// handleConnection manages a new client connection
func (s *Server) handleConnection(conn net.Conn) {
	client := newClient(conn)
	defer s.cleanupClient(client)

	// Get username
	if err := s.registerClient(client); err != nil {
		return
	}

	// Start chat session
	s.broadcastJoin(client)
	s.startChatLoop(client)
}

// handleConnection manages a new client connection
func (s *Server) handleConnection(conn net.Conn) {
	client := newClient(conn)
	defer s.cleanupClient(client)

	// Get username
	if err := s.registerClient(client); err != nil {
		return
	}

	// Start chat session
	s.broadcastJoin(client)
	s.startChatLoop(client)
}

// registerClient gets and sets the client's username
func (s *Server) registerClient(client *Client) error {
	if err := client.prompt("Enter your username: "); err != nil {
		return err
	}

	username, err := client.readInput()
	if err != nil {
		return err
	}
	client.Username = username

	s.clientsMu.Lock()
	s.clients[client] = true
	s.clientsMu.Unlock()

	return nil
}