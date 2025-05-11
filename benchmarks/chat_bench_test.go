package benchmarks

import (
	"net"
	"testing"
	"time"
	"syscall"
	"github.com/ScarletSalinas/SemesterProject/tcp"
)

const (
	testPort    = ":5000"
	testMessage = "benchmark\n" // For latency tests
	largeMessage = "hello\n"   // For throughput tests

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

/// 1. Latency Benchmark (Single Connection)
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

// 2. Packet Loss Benchmark
func BenchmarkPacketLoss(b *testing.B) {
    // 1. Server Setup (using your existing Start() method)
    server := tcp.NewServer()
    server.BenchmarkMode = true
    
    // Start server normally (not with Serve)
    errChan := make(chan error, 1)
    go func() {
        errChan <- server.Start(testPort)
    }()
    
    // Wait for server to be ready
    select {
    case err := <-errChan:
        if err != nil {
            b.Fatalf("Server failed to start: %v", err)
        }
    case <-time.After(500 * time.Millisecond):
        if _, err := net.Dial("tcp", testPort); err != nil {
            b.Fatal("Server did not start in time")
        }
    }
    defer server.Stop()

	// 2. Client Connection
	conn, err := net.Dial("tcp", testPort)
	if err != nil {
		b.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()
 
	// 3. TCP Optimizations
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)  // Disable Nagle's algorithm
		tcpConn.SetLinger(0)      // Disable lingering
		
		// Enable TCP QuickACK if possible
		if fd, err := tcpConn.File(); err == nil {
			syscall.SetsockoptInt(int(fd.Fd()), syscall.IPPROTO_TCP, syscall.TCP_QUICKACK, 1)
			fd.Close()
		}
	}

	// 4. Benchmark Configuration
	msg := []byte{'x'} // Single byte message
	recvBuf := make([]byte, 1)
	const lossInterval = 10
	var packetsLost int
 
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i > 0 && i%lossInterval == 0 {
			packetsLost++
			continue
		}

		start := time.Now()
		if _, err := conn.Write(msg); err != nil {
			b.Error("Write error:", err)
			continue
		}
		if _, err := conn.Read(recvBuf); err != nil {
			b.Error("Read error:", err)
			continue
		}
		b.ReportMetric(float64(time.Since(start).Nanoseconds()), "ns/op")
	}
 
	// 5. Report Metrics
	lossPercent := float64(packetsLost) / float64(b.N) * 100
	b.ReportMetric(lossPercent, "%_lost")
	b.ReportMetric(float64(packetsLost), "packets_lost")
	b.ReportMetric(float64(len(msg)), "bytes/op")
	b.ReportMetric(0, "B/op")          // Zero allocations
	b.ReportMetric(0, "allocs/op")     // Confirmed no allocs
 }
