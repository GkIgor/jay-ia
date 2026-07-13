package ipc

import (
	"encoding/json"
	"net"
	"path/filepath"
	"testing"

	sdkipc "github.com/GkIgor/jay-ia/sdk/ipc"
)

func TestIPCServer(t *testing.T) {
	// Set XDG_RUNTIME_DIR to a temporary directory to isolate the test socket
	tempDir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", tempDir)

	// Mock handler that returns static response
	mockHandler := func(msg sdkipc.Message) sdkipc.Message {
		return sdkipc.Message{
			Type: "response",
			Payload: sdkipc.Response{
				Status: "ok",
				Data:   "Message received",
			},
		}
	}

	server, err := NewServer(mockHandler)
	if err != nil {
		t.Fatalf("failed to create IPC server: %v", err)
	}
	defer server.Stop()

	// Verify socket path uses XDG_RUNTIME_DIR
	expectedSocketPath := filepath.Join(tempDir, "jay", "jay.sock")
	if server.socketPath != expectedSocketPath {
		t.Errorf("expected socket path %q, got %q", expectedSocketPath, server.socketPath)
	}

	// Start the server
	err = server.Start()
	if err != nil {
		t.Fatalf("failed to start IPC server: %v", err)
	}

	// Connect a client
	conn, err := net.Dial("unix", server.socketPath)
	if err != nil {
		t.Fatalf("failed to connect to socket: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Logf("failed to close connection: %v", err)
		}
	}()

	// Prepare and send a JSON message
	reqMsg := sdkipc.Message{
		Type: "command",
		Payload: sdkipc.Command{
			ID:     "test-cmd-id",
			Action: "test-action",
			Data:   "hello",
		},
	}

	encoder := json.NewEncoder(conn)
	err = encoder.Encode(reqMsg)
	if err != nil {
		t.Fatalf("failed to encode request message: %v", err)
	}

	// Read and validate the response
	var respMsg sdkipc.Message
	decoder := json.NewDecoder(conn)
	err = decoder.Decode(&respMsg)
	if err != nil {
		t.Fatalf("failed to decode response message: %v", err)
	}

	if respMsg.Type != "response" {
		t.Errorf("expected response type 'response', got %q", respMsg.Type)
	}

	// The payload is raw JSON/any, which decodes to map[string]any
	payloadMap, ok := respMsg.Payload.(map[string]any)
	if !ok {
		t.Fatalf("expected payload to be map[string]any, got %T", respMsg.Payload)
	}

	if payloadMap["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", payloadMap["status"])
	}

	if payloadMap["data"] != "Message received" {
		t.Errorf("expected data 'Message received', got %v", payloadMap["data"])
	}
}
