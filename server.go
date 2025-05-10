package main

import(
		"fmt"
		"log"
		"net"
		"strings"
		"sync"
		"time"
)


type Server struct {
	clients   map[*Client]bool
	clientsMu sync.Mutex
}

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

// broadcastJoin announces a new user to all clients
func (s *Server) broadcastJoin(client *Client) {
	joinMsg := fmt.Sprintf("[%s %s] has joined", time.Now().Format("15:04"), client.Username)
	s.broadcast(nil, joinMsg)
	client.sendMessage(fmt.Sprintf("Welcome, %s! Type /help for commands\n", client.Username))
	log.Printf("New connection: %s (%d active)", client.Username, len(s.clients))
}

// startChatLoop handles the main chat session for a client
func (s *Server) startChatLoop(client *Client) {
	inputChan := make(chan string)
	defer close(inputChan)

	// Input reader goroutine
	go func() {
		for {
			msg, err := client.readInput()
			if err != nil {
				log.Printf("Read error from %s: %v", client.Username, err)
				close(inputChan)
				return
			}
			inputChan <- msg
		}
	}()

	// Message processor
	for msg := range inputChan {
		if len(msg) == 0 {
			continue
		}
		s.handleMessage(client, msg)
	}
}

// handleMessage processes a single message/command
func (s *Server) handleMessage(client *Client, msg string) {
	switch {
	case msg == "/quit":
		client.sendMessage("Goodbye!")
		client.Conn.Close()

	case msg == "/help":
		client.sendMessage(`Available commands:
/help    - Show help 
/who     - List online users
/quit    - Disconnect from chat`)

	case msg == "/who":
		s.listUsers(client)

	case strings.HasPrefix(msg, "/"):
		client.sendMessage("Unknown command. Try /help")

	default:
		chatMsg := fmt.Sprintf("[%s %s]: %s", 
			time.Now().Format("15:04"), 
			client.Username, 
			msg)
		s.broadcast(client, chatMsg)
	}
}

// listUsers sends the online user list to a client
func (s *Server) listUsers(client *Client) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	var users []string
	for c := range s.clients {
		users = append(users, c.Username)
	}
	client.sendMessage("Online: " + strings.Join(users, ", "))
}