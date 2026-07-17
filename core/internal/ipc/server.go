package ipc

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/GkIgor/jay-ia/sdk/ipc"
)

// IPCEvent represents a strongly-typed push event
type IPCEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

// Server handles IPC connections
type Server struct {
	socketPath string
	listener   net.Listener
	quit       chan struct{}
	handler    func(ipc.Message) ipc.Message

	mu      sync.RWMutex
	clients map[net.Conn]chan interface{}
}

// NewServer creates a new IPC server.
func NewServer(handler func(ipc.Message) ipc.Message) (*Server, error) {
	socketPath := getSocketPath()

	if err := os.MkdirAll(filepath.Dir(socketPath), 0700); err != nil {
		return nil, fmt.Errorf("failed to create socket directory: %w", err)
	}

	_ = os.Remove(socketPath)

	return &Server{
		socketPath: socketPath,
		quit:       make(chan struct{}),
		handler:    handler,
		clients:    make(map[net.Conn]chan interface{}),
	}, nil
}

func getSocketPath() string {
	xdgRuntimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if xdgRuntimeDir != "" {
		return filepath.Join(xdgRuntimeDir, "jay", "jay.sock")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/jay/jay.sock"
	}
	return filepath.Join(home, ".jay", "jay.sock")
}

// Start begins listening for connections
func (s *Server) Start() error {
	l, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.socketPath, err)
	}
	s.listener = l
	log.Printf("IPC Server listening on %s", s.socketPath)

	go s.acceptLoop()
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				log.Printf("IPC Accept error: %v", err)
				continue
			}
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	writeChan := make(chan interface{}, 100)

	s.mu.Lock()
	s.clients[conn] = writeChan
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
		close(writeChan)
		conn.Close()
		log.Printf("IPC Client disconnected: %s", conn.RemoteAddr())
	}()

	log.Printf("IPC Client connected: %s", conn.RemoteAddr())

	// Start connection writer loop to serialize writes thread-safely
	go func() {
		encoder := json.NewEncoder(conn)
		for msg := range writeChan {
			if err := encoder.Encode(msg); err != nil {
				log.Printf("Failed to encode message: %v", err)
				conn.Close()
				return
			}
		}
	}()

	decoder := json.NewDecoder(conn)
	for {
		var msg ipc.Message
		if err := decoder.Decode(&msg); err != nil {
			log.Printf("IPC connection closed or decode error: %v", err)
			return
		}

		log.Printf("Received message: %+v", msg)

		// Process commands concurrently so the read loop remains free
		// to process subsequent messages (like permission responses).
		go func(m ipc.Message) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("IPC response dropped because client disconnected: %v", r)
				}
			}()

			var resp ipc.Message
			if s.handler != nil {
				resp = s.handler(m)
			} else {
				resp = ipc.Message{
					Type: "response",
					Payload: ipc.Response{
						Status: "error",
						Data:   "No message handler configured",
					},
				}
			}

			// Queue response to the writer loop safely
			select {
			case writeChan <- resp:
			default:
				log.Printf("Write queue full for %s, dropping response", conn.RemoteAddr())
			}
		}(msg)
	}
}

// Broadcast sends an event to all connected clients.
func (s *Server) Broadcast(event IPCEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, writeChan := range s.clients {
		select {
		case writeChan <- event:
		default:
			log.Printf("Broadcast queue full, dropping event")
		}
	}
}

// Stop gracefully stops the server
func (s *Server) Stop() {
	close(s.quit)
	if s.listener != nil {
		s.listener.Close()
	}
}
