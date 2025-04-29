package main

import (
	"bufio"
	"log"
	"net"
	"strings"
	"sync"
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
	client := &Client{Conn: conn}
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

	// Message handling loop
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		// Show input prompt
		conn.Write([]byte("> "))
		
		msg := strings.TrimSpace(scanner.Text())

		if len(msg) == 0 {
			continue
		}

		switch {
		case msg == "/quit":
			conn.Write([]byte("Goodbye!\n"))
			return

		case msg == "/who":
			clientsMu.Lock()
			var users []string
			for c := range clients {
				users = append(users, c.Username)
			}
			clientsMu.Unlock()
			conn.Write([]byte("Online users: " + strings.Join(users, ", ") + "\n"))

		case strings.HasPrefix(msg, "/"):
			conn.Write([]byte("Unknown command. Try /who or /quit\n"))

		default:
			broadcast(client, "["+username+"] "+msg)
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