package benchmarks

import (
	"bufio"
	"context"
	"net"
	"testing"
	"time"
	"github.com/ScarletSalinas/SemesterProject/tcp"
)

const (
	testPort    = ":4000"
	testMessage = "benchmark\n"
)

// startTestServer starts the TCP server in benchmark mode (echo only)
func startTestServer(b *testing.B) *tcp.Server {
	server := tcp.NewServer()
	server.BenchMarkMode = true 
	go func() {
		if err := server.Start(testPort); err != nil {
			b.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for server to start (max 500ms)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			b.Fatal("Server did not start in time")
		default:
			conn, err := net.Dial("tcp", testPort)
			if err == nil {
				conn.Close()
				return server
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// --- Benchmark Tests ---

// 1. Measures average round-trip latency per message
func BenchmarkLatency(b *testing.B) {
	server := startTestServer(b)
	defer server.Stop()

	conn, err := net.Dial("tcp", testPort)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		start := time.Now()
		if _, err := conn.Write([]byte(testMessage)); err != nil {
			b.Error(err)
			continue
		}
		if _, err := reader.ReadString('\n'); err != nil {
			b.Error(err)
			continue
		}
		elapsed := time.Since(start)
		b.ReportMetric(float64(elapsed.Nanoseconds()), "ns/op")
	}

}

// 2. Simulates 10% packet loss and measures reliability
func BenchmarkPacketLoss(b *testing.B) {
	server := startTestServer(b)
	defer server.Stop()

	var packetsLost int
	for i := 0; i < b.N; i++ {
		conn, err := net.Dial("tcp", testPort)
		if err != nil {
			b.Error(err)
			continue
		}

		// Simulate 10% packet loss by randomly dropping writes
		if i%10 == 0 { // Drop every 10th packet
			packetsLost++
			conn.Close()
			continue
		}

		if _, err := conn.Write([]byte(testMessage)); err != nil {
			b.Error(err)
		}
		conn.Close()
	}
	b.ReportMetric(float64(packetsLost), "packets_lost")
}