package tcp

import (
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
	listener  net.Listener
	running   bool
}

func NewServer() *Server {
	return &Server{
		clients: make(map[*Client]bool),
	}
}

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

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in handler: %v", r)
		}
	}()

	client := NewClient(conn)
	defer s.cleanupClient(client)
	conn.SetDeadline(time.Time{})

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

func (s *Server) handleMessage(client *Client, msg string) {
	client.mu.Lock()
	if client.closed {
		client.mu.Unlock()
		return
	}
	client.mu.Unlock()

	if _, err := client.Conn.Write([]byte("> ")); err != nil {
		s.cleanupClient(client)
		return
	}

	switch {
	case msg == "/quit":
		s.broadcastSystemMessage(fmt.Sprintf("\033[1;31m%s has left the chat\033[0m", client.Username))
		return

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

func (s *Server) cleanupClient(client *Client) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	if _, exists := s.clients[client]; exists {
		leaveMsg := fmt.Sprintf("\033[1;31m%s has left the chat\033[0m", client.Username)
		for remainingClient := range s.clients {
			if remainingClient != client {
				remainingClient.SendMessage(leaveMsg)
			}
		}

		delete(s.clients, client)
		client.Close()
		log.Printf("[%s] %s@%s disconnected (%d active connections)", 
			time.Now().Format("2006-01-02 15:04:05"),
			client.Username, 
			client.Conn.RemoteAddr().String(), 
			len(s.clients))
	}
}

func (s *Server) Stop() error {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	
	s.running = false
	
	if s.listener != nil {
		s.listener.Close()
	}
	
	for client := range s.clients {
		client.Close()
		delete(s.clients, client)
	}
	
	return nil
}