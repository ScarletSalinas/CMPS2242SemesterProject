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
func (s *Server) Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("Chat server running on %s", listener.Addr())

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

	// Panic recovery
	defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered from panic in handler: %v", r)
        }
    }()

	inputChan := make(chan string)
	conn.SetDeadline(time.Now().Add(5 * time.Minute)) // Timeout
	client := newClient(conn)
	defer func() {
		s.cleanupClient(client)
		close(inputChan)
		conn.Close()
	}()

	// Get username
	if err := s.registerClient(client); err != nil {
		return
	}

	// Start chat session
	s.broadcastSystemMessage(fmt.Sprintf("\033[1;33m%s has joined the chat\033[0m", client.Username))
	client.sendMessage(fmt.Sprintf("\033[1;32mWelcome, %s!\033[0m Type /help for commands\n", client.Username))
	
	s.startChatLoop(client, inputChan)
}

// registerClient: gets and sets the client's username
func (s *Server) registerClient(client *Client) error {
	client.prompt("\033[1;36mEnter your username: \033[0m")
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

// startChatLoop: handles the main chat session for a client
func (s *Server) startChatLoop(client *Client,  inputChan chan string) {
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
	client.Conn.SetDeadline(time.Now().Add(5 * time.Minute))  // timeout renewal
	for msg := range inputChan {
		if len(msg) == 0 {
			continue
		}
		s.handleMessage(client, msg)
	}
}

// handleMessage processes a single message/command
func (s *Server) handleMessage(client *Client, msg string) {

	// Check
	if _, err := client.Conn.Write([]byte("> ")); err != nil {
        s.cleanupClient(client)
        return
    }

	switch {
	case msg == "/quit":
		s.broadcastSystemMessage(fmt.Sprintf("\033[1;31m%s has left the chat\033[0m", client.Username))
		return  // Exit the handler after quitting

	case msg == "/help":
		client.sendMessage("\033[1;35mAvailable commands:\n/help    - Show help\n/who     - List online users\n/quit    - Disconnect from chat\033[0m")

	case msg == "/who":
		s.listUsers(client)

	case strings.HasPrefix(msg, "/"):
		client.sendMessage("\033[1;31mUnknown command. Try /help\033[0m")

	default:
		chatMsg := fmt.Sprintf("\033[1;34m[%s]\033[0m \033[1;36m%s\033[0m: %s", time.Now().Format("15:04"), client.Username, msg)
		s.broadcast(client, chatMsg)
	}

	if time.Since(client.LastMessage) < 1*time.Second {
		client.sendMessage("\033[1;31mMessage rate limit exceeded\033[0m")
		return
	}
	client.LastMessage = time.Now()
}

// listUsers sends the online user list to a client
func (s *Server) listUsers(client *Client) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	var users []string
	for c := range s.clients {
		c.mu.Lock()
        if !c.closed {
            users = append(users, c.Username)
        }
        c.mu.Unlock()
    }
    
	client.sendMessage(fmt.Sprintf("\033[1;35mOnline users: \033[1;33m%s\033[0m", strings.Join(users, ", ")))
}

// broadcast sends a message to all clients
func (s *Server) broadcast(sender *Client, msg string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	for client := range s.clients {
		if client != sender {
			client.sendMessage(msg)
		}
	}
}

func (s *Server) broadcastSystemMessage(msg string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	for client := range s.clients {
		client.sendMessage(msg)
	}
}

// cleanupClient removes a disconnected client
func (s *Server) cleanupClient(client *Client) {
	// Broadcast while client is still in map
    s.clientsMu.Lock()
	defer s.clientsMu.Unlock()    

	if _, exists := s.clients[client]; exists {
        leaveMsg := fmt.Sprintf("\033[1;31m%s has left the chat\033[0m", client.Username)
        for remainingClient := range s.clients {
            if remainingClient != client {
                remainingClient.sendMessage(leaveMsg)
            }
        }

		 // Cleanup
		 delete(s.clients, client)
		 client.Close()
		 log.Printf("[%s] %s@%s disconnected (%d active connections)", 
			 time.Now().Format("2006-01-02 15:04:05"),
			 client.Username, 
			 client.Conn.RemoteAddr().String(), 
			 len(s.clients))
	}
    
}

// Graceful shutdown
func (s *Server) Stop() {
    s.clientsMu.Lock()
    defer s.clientsMu.Unlock()

    for client := range s.clients {
        client.sendMessage("Server shutting down...")
    }

	time.Sleep(5*time.Second)
	for client := range s.clients {
		client.Conn.Close()
	}
}

