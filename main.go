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
	// Create and register new client
    client := &Client{Conn: conn}

	// Cleanup on exit
    defer func() {
        clientsMu.Lock()
        delete(clients, client)
        clientsMu.Unlock()
        broadcast(nil, fmt.Sprintf("[%s %s] has left the chat", time.Now().Format("15:04"), client.Username))
        conn.Close()
        log.Printf("Connection closed (%d remaining)", len(clients))
    }()


    write := func(w *syncWriter, text string) {
        w.Lock()
        defer w.Unlock()
		
    // Clear current line and write
	_, err := w.conn.Write([]byte("\033[2K\r"))
        if err != nil {
            w.conn.Write([]byte("\n"))
        }
        w.conn.Write([]byte(text))

}
    

	// Prompt for username
	write(w, "Enter your username: ")  
	username, err := bufio.NewReader(conn).ReadString('\n')
    if err != nil {
        return
    }
    client.Username = strings.TrimRight(username, "\r\n") // Strict trim

	// Register client
    clientsMu.Lock()
    clients[client] = true
    clientsMu.Unlock()

	// Welcome client
	log.Printf("New connection (%d active)", len(clients))
	// Send join notification
	joinMsg := fmt.Sprintf("[%s %s] has joined", time.Now().Format("15:04"), client.Username)
	broadcast(nil, joinMsg)
	write(w, fmt.Sprintf("Welcome, %s! Type /help for commands\n\n=== Chat Started ===\n", client.Username))
	
	// Input handling
    inputChan := make(chan string)
    defer close(inputChan)

	// Input reader
    go func() {
        reader := bufio.NewReader(conn)

        for {
            msg, err := reader.ReadString('\n')
            if err != nil {
				log.Printf("Read error: %v", err)
                return
            }
            inputChan <- strings.TrimRight(msg, "\r\n")
        }
    }()

	// Message handling
	for msg := range inputChan {
        if len(msg) == 0 {
            continue
        }

        switch {
		case msg == "/quit":
			write(w, "You left.\n")
			return
		
		case msg == "/help":	// List available commands
			helpMsg := `Available commands:
/help    - Show help 
/who     - List online users
/quit    - Disconnect from chat
`
		write(w, helpMsg)

		case msg == "/who":  // List clients in server
			clientsMu.Lock()
			var users []string
			for c := range clients {
				users = append(users, c.Username)
			}
			clientsMu.Unlock()
			write(w, "Online: "+strings.Join(users, ", ")+"\n")

		case strings.HasPrefix(msg, "/"): // Unknown command
			write(w, "Unknown command. Try /help for options\n")

		default:  // Broadcast msg
			broadcast(client, fmt.Sprintf("[%s \033[1;36m%s\033[0m]: %s", time.Now().Format("15:04"), client.Username, msg))
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