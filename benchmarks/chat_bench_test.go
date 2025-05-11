package benchmarks

import (
	"testing"
	"net"
	"time"
	"github.com/ScarletSalinas/SemesterProject/tcp"
)

func BenchmarkTCPServer(b *testing.B) {
	// 1. Start the server
	server := tcp.NewServer()
	go server.Start(":4000")
	defer server.Stop()
	time.Sleep(100 * time.Millisecond) // Wait for server to start

	b.ResetTimer()

	// 2. Test different scenarios
	b.Run("SingleClient", func(b *testing.B) {
		conn, err := net.Dial("tcp", ":4000")
		if err != nil {
			b.Fatal(err)
		}
		defer conn.Close()

		testMessage := []byte("benchmark\n")
		response := make([]byte, len(testMessage))

		for i := 0; i < b.N; i++ {
			// Measure round-trip time
			start := time.Now()
			
			if _, err := conn.Write(testMessage); err != nil {
				b.Error(err)
				continue
			}
			
			if _, err := conn.Read(response); err != nil {
				b.Error(err)
				continue
			}
			
			// Record metrics
			b.ReportMetric(float64(time.Since(start).Nanoseconds()), "ns/op")
		}
	})

	b.Run("MultipleClients", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			conn, err := net.Dial("tcp", ":4000")
			if err != nil {
				b.Fatal(err)
			}
			defer conn.Close()

			testMessage := []byte("benchmark\n")
			response := make([]byte, len(testMessage))

			for pb.Next() {
				if _, err := conn.Write(testMessage); err != nil {
					b.Error(err)
					continue
				}
				if _, err := conn.Read(response); err != nil {
					b.Error(err)
					continue
				}
			}
		})
	})
}