package tcp_test

import (
	"net"
	"testing"
	"time"
	"github.com/ScarletSalinas/SemesterProject/tcp"  
)

const (
	testPort    = ":5000"
	testMessage = "benchmark\n"
)

func startEchoServer() net.Listener {
	ln, err := net.Listen("tcp", testPort)
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 1024)
				for {
					n, err := c.Read(buf)
					if err != nil {
						return
					}
					c.Write(buf[:n])
				}
			}(conn)
		}
	}()
	return ln
}

func BenchmarkThroughput(b *testing.B) {
    server := tcp.NewServer()
    go server.Start(":5000")
    defer server.Stop()

    conn, err := net.Dial("tcp", ":5000")
    if err != nil {
        b.Fatal(err)
    }
    defer conn.Close()

    msg := []byte("test\n")
    reply := make([]byte, len(msg))

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        conn.Write(msg)
        conn.Read(reply)
    }
}

func BenchmarkLatency(b *testing.B) {
	ln := startEchoServer()
	defer ln.Close()

	conn, err := net.Dial("tcp", testPort)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	msg := []byte(testMessage)
	reply := make([]byte, len(msg))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		if _, err := conn.Write(msg); err != nil {
			b.Fatal(err)
		}
		if _, err := conn.Read(reply); err != nil {
			b.Fatal(err)
		}
		elapsed := time.Since(start)
		b.ReportMetric(float64(elapsed.Microseconds()), "µs/op")
	}
}

func BenchmarkPacketLoss(b *testing.B) {
	ln := startEchoServer()
	defer ln.Close()

	conn, err := net.Dial("tcp", testPort)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	msg := []byte(testMessage)
	reply := make([]byte, len(msg))
	lost := 0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%10 == 0 { // Simulate 10% packet loss
			lost++
			continue
		}
		if _, err := conn.Write(msg); err != nil {
			b.Fatal(err)
		}
		if _, err := conn.Read(reply); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportMetric(float64(lost)/float64(b.N)*100, "%_lost")
}