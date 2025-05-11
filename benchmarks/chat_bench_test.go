package benchmarks

import (
	"bufio"
	"net"
	"testing"
	"time"
	"github.com/ScarletSalinas/SemesterProject/tcp"
)

const (
	testPort    = ":5000"
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
    // 1. Start server
    b.Log("Starting server...")
    server := startTestServer(b)
    defer func() {
        b.Log("Stopping server...")
        server.Stop()
    }()

    // 2. Create connection
    conn, err := net.Dial("tcp", testPort)
    if err != nil {
        b.Fatalf("Dial failed: %v", err)
    }
    defer conn.Close()
    
    // 3. Configure connection
    if tcpConn, ok := conn.(*net.TCPConn); ok {
        tcpConn.SetNoDelay(true)
        tcpConn.SetLinger(0)
    }

    // 4. Prepare buffers
    msg := []byte(testMessage)
    recvBuf := make([]byte, len(msg))

    // 5. Warm-up
    if _, err := conn.Write(msg); err != nil {
        b.Fatal("Warm-up failed:", err)
    }
    conn.Read(recvBuf)

    // 6. Benchmark
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        start := time.Now()
        
        if _, err := conn.Write(msg); err != nil {
            b.Error("Write error:", err)
            continue
        }
        
        if _, err := conn.Read(recvBuf); err != nil {
            b.Error("Read error:", err)
            continue
        }
        
        elapsed := time.Since(start)
        b.ReportMetric(float64(elapsed.Nanoseconds()), "ns/op")
        
        // Safe debug logging (fixed division by zero)
        if b.N > 10 && i%(b.N/10) == 0 && i > 0 {
            b.Logf("Progress: %d/%d (%.1f%%)", i, b.N, float64(i)/float64(b.N)*100)
        }
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
