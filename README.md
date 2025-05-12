# TCP Chat Server

A concurrent TCP chat server implementation in Go.

## Features

- 🚀 Multi-client support
- 💬 Command-based chat interface
- 🌈 ANSI-colored output
- ⏱️ Connection timeouts
- 📊 Built-in benchmarking

## Installation

```bash
git clone https://github.com/ScarletSalinas/SemesterProject.git
cd SemesterProject
go build
```

## Usage

### Start Server
```bash
# Default port (4000)
./SemesterProject

# Custom port
./SemesterProject -port 5000
```
### Connect Clients
```bash
# Linux/Mac
nc localhost 4000

# Windows
telnet localhost 4000
```
### Chat Commands
```text
Command	  Description
/help   Show available commands
/who    List online users
/quit   Disconnect from chat
```

## 📂 Project Structure

```text
.
├── main.go         # Server entry point
├── tcp/
│   ├── server.go   # Core server logic
│   ├── client.go   # Client management
│   └── bench_test.go # Performance tests
├── go.mod         # Dependency management
└── README.md      # Project documentation
```
## Benchmarks

### Run performance tests:

```bash
go test -bench=. ./tcp/ -benchmem
```

### Sample Output
```bash
Latency: 17.86 μs/op
Throughput: 39,229 msg/s
```

## References
- [Practical Go Lessons by Maximilien Andile](https://www.practical-go-lessons.com/)
- LLM: DeepSeek, for tutoring, learning necessary concepts, and for guidance when needed.
- [W3Schools](https://www.w3schools.com/go/go_switch.php)

# Link to video and presentation
[Watch demo here]() 
[View Slides here](https://docs.google.com/presentation/d/1Z5EwuB8ZRvuQEp95mF7FgSsF6ac3ZR0v/edit#slide=id.p1) 


