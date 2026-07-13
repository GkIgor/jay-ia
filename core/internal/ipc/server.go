package ipc

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/GkIgor/jay-ia/sdk/ipc"
)

// Server handles IPC connections
type Server struct {
	socketPath string
	listener   net.Listener
	quit       chan struct{}
	handler    func(ipc.Message) ipc.Message
}

// NewServer creates a new IPC server. It prefers XDG_RUNTIME_DIR on Linux.
func NewServer(handler func(ipc.Message) ipc.Message) (*Server, error) {
	socketPath := getSocketPath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(socketPath), 0700); err != nil {
		return nil, fmt.Errorf("failed to create socket directory: %w", err)
	}

	// Remove existing socket if it exists
	_ = os.Remove(socketPath)

	return &Server{
		socketPath: socketPath,
		quit:       make(chan struct{}),
		handler:    handler,
	}, nil
}

func getSocketPath() string {
	// XDG_RUNTIME_DIR is standard for Linux runtime files
	xdgRuntimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if xdgRuntimeDir != "" {
		return filepath.Join(xdgRuntimeDir, "jay", "jay.sock")
	}

	// Fallback/compatibility layer for Mac/Windows or missing XDG
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
	defer conn.Close()
	log.Printf("IPC Client connected: %s", conn.RemoteAddr())

	decoder := json.NewDecoder(conn)
	for {
		var msg ipc.Message
		if err := decoder.Decode(&msg); err != nil {
			log.Printf("IPC connection closed or decode error: %v", err)
			return
		}

		log.Printf("Received message: %+v", msg)
		
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
		
		encoder := json.NewEncoder(conn)
		if err := encoder.Encode(resp); err != nil {
			log.Printf("Failed to encode response: %v", err)
			return
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
