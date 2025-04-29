package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Client struct {
	Conn net.Conn
	Username string
}

var (
	clients = make(map[*Client]bool)   // Tracks all connected clients
	clientsMu sync.Mutex				// Protects the clients map
)

// broadcast sends a message to all connected clients
func broadcast(sender *Client, msg string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		// Skip sender if specified
		if sender != nil && client == sender {
			continue
		}
		_, err := client.Conn.Write([]byte(msg + "\n"))
		if err != nil {
			log.Printf("Failed to send to %s: %v", client.Username, err)
		}
	}
}

func handleConnection(conn net.Conn) {
	// Prompt for username
	conn.Write([]byte("Enter your username: "))
	username, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		conn.Close()
		return
	}

	username = strings.TrimSpace(username)

	// Create and register new client
	client := &Client{Conn: conn, Username: username}
	clientsMu.Lock()
	clients[client] = true
	clientsMu.Unlock()

	log.Printf("New connection (%d active)", len(clients))
	broadcast(nil, username + " has joined the chat")  // Notify all
	conn.Write([]byte("Welcome, " + username + "! Type /help for commands\n"))

	defer func() {
		// Cleanup on exit
		clientsMu.Lock()
		delete(clients, client)
		clientsMu.Unlock()
		broadcast(nil, username + " has left the chat")
		conn.Close()
		log.Printf("Connection closed (%d remaining)", len(clients))
	}()

	// Initialize chat
    conn.Write([]byte("\n=== Chat Started ===\n"))
	prompt := func() { conn.Write([]byte("> ")) }  // Helper function

	prompt()
    
    // Create a channel for user input
    inputChan := make(chan string)
    
    // Goroutine to show prompts and read input
    go func() {
        reader := bufio.NewReader(conn)
        for {
            msg, err := reader.ReadString('\n')
            if err != nil {
                close(inputChan)
                return
            }
            inputChan <- strings.TrimSpace(msg)
        }
    }()

	for msg := range inputChan {
        if len(msg) == 0 {
			prompt()
            continue
        }

		switch {
		case msg == "/quit":
			conn.Write([]byte("Goodbye!\n"))
			return
		
		case msg == "/help":
			helpMsg := `
		Available commands:
		/help    - Show this help message
		/who     - List online users
		/quit    - Disconnect from chat
		`
			conn.Write([]byte(helpMsg))
			conn.Write([]byte("> ")) // Restore prompt

		case msg == "/who":
			clientsMu.Lock()
			var users []string
			for c := range clients {
				users = append(users, c.Username)
			}
			clientsMu.Unlock()
			conn.Write([]byte("Online users: " + strings.Join(users, ", ") + "\n"))
			prompt()

		case strings.HasPrefix(msg, "/"):
			conn.Write([]byte("Unknown command. Try /who or /quit\n"))
			prompt()

		default:
			broadcast(client, fmt.Sprintf("[%s] %s: %s", time.Now().Format("15:04"), username, msg))
			prompt()
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":4000")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	log.Println("Chat server running on :4000")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Connection error: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}