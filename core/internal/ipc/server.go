package ipc

import (
	"context"
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
	rawHandler func(ctx context.Context, rawRequest []byte) []byte

	mu      sync.RWMutex
	clients map[net.Conn]chan interface{}
}

// NewServer creates a new IPC server using legacy Message handler.
func NewServer(handler func(ipc.Message) ipc.Message) (*Server, error) {
	return newServerInternal(handler, nil)
}

// NewRawServer creates a new IPC server using raw bytes RPC handler (Router).
func NewRawServer(rawHandler func(ctx context.Context, rawRequest []byte) []byte) (*Server, error) {
	return newServerInternal(nil, rawHandler)
}

func newServerInternal(handler func(ipc.Message) ipc.Message, rawHandler func(ctx context.Context, rawRequest []byte) []byte) (*Server, error) {
	socketPath := getSocketPath()

	if err := os.MkdirAll(filepath.Dir(socketPath), 0700); err != nil {
		return nil, fmt.Errorf("failed to create socket directory: %w", err)
	}

	_ = os.Remove(socketPath)

	return &Server{
		socketPath: socketPath,
		quit:       make(chan struct{}),
		handler:    handler,
		rawHandler: rawHandler,
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
			if rawBytes, ok := msg.([]byte); ok {
				var rawMsg json.RawMessage = rawBytes
				if err := encoder.Encode(rawMsg); err != nil {
					log.Printf("Failed to encode raw message: %v", err)
					conn.Close()
					return
				}
			} else {
				if err := encoder.Encode(msg); err != nil {
					log.Printf("Failed to encode message: %v", err)
					conn.Close()
					return
				}
			}
		}
	}()

	decoder := json.NewDecoder(conn)
	for {
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			log.Printf("IPC connection closed or decode error: %v", err)
			return
		}

		go func(rawBytes []byte) {
			log.Printf("[IPC] Requisição recebida (%d bytes)", len(rawBytes))
			defer func() {
				if r := recover(); r != nil {
					log.Printf("IPC response dropped because client disconnected: %v", r)
				}
			}()

			if s.rawHandler != nil {
				respBytes := s.rawHandler(context.Background(), rawBytes)
				select {
				case writeChan <- respBytes:
				default:
					log.Printf("Write queue full for %s, dropping response", conn.RemoteAddr())
				}
				return
			}

			// Legacy handler
			var msg ipc.Message
			if err := json.Unmarshal(rawBytes, &msg); err != nil {
				return
			}

			var resp ipc.Message
			if s.handler != nil {
				resp = s.handler(msg)
			} else {
				resp = ipc.Message{
					Type: "response",
					Payload: ipc.Response{
						Status: "error",
						Data:   "No message handler configured",
					},
				}
			}

			select {
			case writeChan <- resp:
			default:
				log.Printf("Write queue full for %s, dropping response", conn.RemoteAddr())
			}
		}(raw)
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
