package main

import (
	"flag"
	"log"
	"strconv"
)

func main() {
	// Define port flag (default: 4000)
	port := flag.String("port", "4000", "TCP port to listen on")
	flag.Parse()

	// Convert and validate port
	p, err := strconv.Atoi(*port)
	if err != nil || p < 1 || p > 1024 {
		log.Fatal("Port must be a number between 1 and 1024")
	}

	// Start server
	server := NewServer()

	if err := server.Start(": + port"); err != nil {
		log.Fatal(err)
	}
}