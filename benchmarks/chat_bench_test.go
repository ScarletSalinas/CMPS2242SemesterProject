package benchmarks

import (
	"bufio"
	"net"
	"testing"
	"time"
	"github.com/ScarletSalinas/SemesterProject/tcp"
)

const (
	testPort    = ":4000"
	testMessage = "benchmark\n"
)

// startTestServer starts the TCP server and returns when ready
func startTestServer(b *testing.B) *tcp.Server {
	server := tcp.NewServer()
	server.BenchmarkMode = true 
	
	errChan := make(chan error, 1)
	
	go func() {
		errChan <- server.Start(testPort)
	}()

	// Wait for server to start or fail
	select {
	case err := <-errChan:
		if err != nil {
			b.Fatalf("Failed to start server: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		// Verify server is actually running
		conn, err := net.Dial("tcp", testPort)
		if err != nil {
			b.Fatal("Server did not start in time")
		}
		conn.Close()
	}

	return server
}

func BenchmarkLatency(b *testing.B) {
	server := startTestServer(b)
	defer func() {
		server.Stop()
		time.Sleep(100 * time.Millisecond) // Allow cleanup
	}()

	conn, err := net.Dial("tcp", testPort)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	buf := make([]byte, len(testMessage))
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		
		if _, err := conn.Write([]byte(testMessage)); err != nil {
			b.Error(err)
			continue
		}
		
		if _, err := reader.Read(buf); err != nil {
			b.Error(err)
			continue
		}
		
		b.ReportMetric(float64(time.Since(start).Nanoseconds()), "ns/op")
	}
}

func BenchmarkPacketLoss(b *testing.B) {
	server := startTestServer(b)
	defer func() {
		server.Stop()
		time.Sleep(100 * time.Millisecond)
	}()

	conn, err := net.Dial("tcp", testPort)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	var packetsLost int
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i > 0 && i%10 == 0 {
			packetsLost++
			continue
		}

		start := time.Now()
		if _, err := conn.Write([]byte(testMessage)); err != nil {
			b.Error(err)
			continue
		}
		
		if _, err := reader.Read(make([]byte, len(testMessage))); err != nil {
			b.Error(err)
			continue
		}
		
		b.ReportMetric(float64(time.Since(start).Nanoseconds()), "ns/op")
	}
	
	b.ReportMetric(float64(packetsLost)*100/float64(b.N), "%_lost")
}