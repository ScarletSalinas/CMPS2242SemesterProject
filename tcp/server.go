package tcp

import(
		"fmt"
		"log"
		"net"
		"strings"
		"sync"
		"time"
)

type Server struct {
	clients   		map[*Client]bool
	clientsMu 		sync.Mutex
	listener  		net.Listener
	running   		bool
	BenchmarkMode  	bool
}

// NewServer creates a new chat server instance
func NewServer() *Server {
	return &Server{
		clients: make(map[*Client]bool),
		running: false,
	}
}

// Start begins listening for connections
func (s *Server) Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	
	s.listener = listener
	s.running = true

	go func() {
		for s.running {
			conn, err := listener.Accept()
			if err != nil {
				if s.running {
					log.Printf("Accept error: %v", err)
				}
				continue
			}
			go s.handleConnection(conn)
		}
	}()
	return nil
}

// handleConnection manages a new client connection
func (s *Server) handleConnection(conn net.Conn) {

	// Panic recovery
	defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered from panic in handler: %v", r)
        }
    }()

	client := NewClient(conn)
	defer s.cleanupClient(client)
	buf := make([]byte, 1024)
	conn.SetDeadline(time.Time{}) // Remove any read/write timeouts
	
	// Benchmark mode handling (simple echo)
	if s.BenchmarkMode {
		conn.SetDeadline(time.Time{}) // Remove any read/write timeouts
		client.writer.benchmarkMode = true // Enable optimized writing
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			if _, err := conn.Write(buf[:n]); err != nil {
				return
			}
		}
    }

	// Normal chat flow
	if err := s.registerClient(client); err != nil {
		return
	}

	s.broadcastSystemMessage(fmt.Sprintf("\033[1;33m%s joined\033[0m", client.Username))
	client.SendMessage(fmt.Sprintf("\033[1;32mWelcome %s!\033[0m", client.Username))

	for {
		msg, err := client.ReadInput()
		if err != nil {
			break
		}
		s.handleMessage(client, msg)
	}
}

// registerClient: gets and sets the client's username
func (s *Server) registerClient(client *Client) error {
	client.prompt("\033[1;36mEnter your username: \033[0m")
	username, err := client.ReadInput()

	if err != nil {
		return err
	}
	client.Username = username

	s.clientsMu.Lock()
	s.clients[client] = true
	s.clientsMu.Unlock()

	return nil
}

// handleMessage processes a single message/command
func (s *Server) handleMessage(client *Client, msg string) {
	client.mu.Lock()
	if client.closed {
		client.mu.Unlock()
		return
	}
	client.mu.Unlock()

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
		client.SendMessage("\033[1;35mAvailable commands:\n/help    - Show help\n/who     - List online users\n/quit    - Disconnect from chat\033[0m")

	case msg == "/who":
		s.listUsers(client)

	case strings.HasPrefix(msg, "/"):
		client.SendMessage("\033[1;31mUnknown command. Try /help\033[0m")

	default:
		chatMsg := fmt.Sprintf("\033[1;34m[%s]\033[0m \033[1;36m%s\033[0m: %s", time.Now().Format("15:04"), client.Username, msg)
		s.broadcast(client, chatMsg)
	}

	if time.Since(client.LastMessage) < 1*time.Second {
		client.SendMessage("\033[1;31mMessage rate limit exceeded\033[0m")
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
    
	client.SendMessage(fmt.Sprintf("\033[1;35mOnline users: \033[1;33m%s\033[0m", strings.Join(users, ", ")))
}

// broadcast sends a message to all clients
func (s *Server) broadcast(sender *Client, msg string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	for client := range s.clients {
		if client != sender {
			client.SendMessage(msg)
		}
	}
}

func (s *Server) broadcastSystemMessage(msg string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	for client := range s.clients {
		client.SendMessage(msg)
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
                remainingClient.SendMessage(leaveMsg)
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

    // Optional: Notify clients (without waiting for sends to complete)
    for client := range s.clients {
        go client.SendMessage("Server shutting down...")
    }

    // Close all connections immediately
    for client := range s.clients {
        client.Conn.Close()  // This will interrupt any blocked operations
    }
    
    // Mark server as stopped
    s.running = false
    
    // Close the listener to stop accepting new connections
    if s.listener != nil {
        s.listener.Close()
    }
}

