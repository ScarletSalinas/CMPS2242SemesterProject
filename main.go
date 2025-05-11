package main

import (
	"flag"
	"log"
	"strconv"
	"strings"
	"github.com/ScarletSalinas/SemesterProject/tcp"
)

func main() {
	// Define port flag (default: 4000)
	port := flag.String("port", "4000", "TCP port to listen on")
	flag.Parse()

	// Clean and validate the port
	portStr := strings.TrimSpace(*port)
	if _, err := strconv.Atoi(portStr); err != nil {
		log.Fatalf("Invalid port number: %v", err)
	}
	address := ":" + portStr

	// Start server
	server := tcp.NewServer()
	if err := server.Start(address); err != nil {
		log.Fatal(err)
	}
}