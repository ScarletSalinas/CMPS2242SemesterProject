package tcp_test

import (
	"net"
	"testing"
	"time"
)

const testMessage = "test\n"

func BenchmarkTCP(b *testing.B) {
	// Start echo server
	ln, err := net.Listen("tcp", ":0") // Random port
	if err != nil {
		b.Fatal(err)
	}
	defer ln.Close()

	// Simple echo handler
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleEcho(conn)
		}
	}()

	// Client connection
	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	msg := []byte(testMessage)
	reply := make([]byte, len(msg))
	
	// Start timing here
	start := time.Now()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Single operation timing
		opStart := time.Now()
		if _, err := conn.Write(msg); err != nil {
			b.Fatal("Write failed:", err)
		}
		if _, err := conn.Read(reply); err != nil {
			b.Fatal("Read failed:", err)
		}
		latency := time.Since(opStart)
		
		// Report per-operation latency
		b.ReportMetric(float64(latency.Nanoseconds())/1000, "latency-µs")
	}
	
	// Calculate throughput
	elapsed := time.Since(start)
	b.ReportMetric(float64(b.N)/elapsed.Seconds(), "throughput-msg/s")
}

func handleEcho(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		conn.Write(buf[:n])
	}
}