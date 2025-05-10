package main

import (
	"log"
	"net"
)

func main() {
	server := NewServer(":4000")
	log.Println("Chat server running on :4000")
	server.Start()
}