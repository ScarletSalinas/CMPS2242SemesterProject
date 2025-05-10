package main

import (
	"log"
)

func main() {
	server := NewServer()
	if err := server.Start(":4000"); err != nil {
		log.Fatal(err)
	}
}