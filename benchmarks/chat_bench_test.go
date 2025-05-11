package benchmarks

import (
	"net"
	"testing"
	"time"
	"github.com/ScarletSalinas/SemesterProject/tcp"
)

func BenchmarkServerPerformance(b *testing.B) {
	// Create server instance
	server := tcp.NewServer()
	
	// Start server in goroutine
	go func() {
		if err := server.Start(":4000"); err != nil {
			b.Errorf("Server failed: %v", err)
		}
	}()
	defer server.Stop()
	
	time.Sleep(100 * time.Millisecond) // Wait for server start

	b.ResetTimer()
	
	// Metrics collectors
	var (
		totalLatency time.Duration
		successCount int
		messageSize  = 128 // bytes
	)

	for i := 0; i < b.N; i++ {
		start := time.Now()
		
		conn, err := net.Dial("tcp", ":4000")
		if err != nil {
			b.Logf("Connection failed: %v", err)
			continue
		}

		// Test message
		msg := make([]byte, messageSize)
		_, err = conn.Write(msg)
		if err != nil {
			b.Logf("Write failed: %v", err)
			conn.Close()
			continue
		}

		// Read response
		buf := make([]byte, messageSize)
		_, err = conn.Read(buf)
		latency := time.Since(start)
		conn.Close()

		if err == nil {
			successCount++
			totalLatency += latency
		}
	}

	// Calculate metrics
	throughput := float64(successCount*messageSize) / 1e6 // MB/s
	avgLatency := totalLatency / time.Duration(successCount)
	packetLoss := 1 - float64(successCount)/float64(b.N)

	b.ReportMetric(avgLatency.Seconds()*1000, "avg_latency_ms")
	b.ReportMetric(throughput, "throughput_mb_s")
	b.ReportMetric(packetLoss*100, "packet_loss_%")
}