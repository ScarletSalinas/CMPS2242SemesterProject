package main

import (
	"bufio"
	"log"
	"net"
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
func broadcast(msg string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		if _, err := client.Conn.Write([]byte(msg + "\n")); err != nil {
			log.Printf("Broadcast failed: %v", err)
			// Don't delete here - handle in connection cleanup
		}
	}
}
func handleConnection(conn net.Conn) {
	// Create new client
	client := &Client{Conn: conn}

	// Register client
	clientsMu.Lock()
	clients[client] = true
	clientsMu.Unlock()

	log.Printf("New connection (%d active)", len(clients))
	broadcast("New user joined!")  // Notify all

	defer func() {
		// Cleanup on exit
		clientsMu.Lock()
		delete(clients, client)
		clientsMu.Unlock()
		conn.Close()
		log.Printf("Connection closed (%d remaining)", len(clients))
	}()

	// Read loop
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		broadcast(">> " + msg) // Send to all clients
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Read error: %v", err)
	}
}

func main() {
	listener, err := net.Listen("tcp", ":4000")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	log.Println("Server started on :4000")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Connection error: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}