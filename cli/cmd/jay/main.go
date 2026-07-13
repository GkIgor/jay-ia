package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/GkIgor/jay-ia/sdk/ipc"
)

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

func main() {
	socketPath := getSocketPath()
	log.Printf("Connecting to Jay Core daemon at %s...", socketPath)

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to connect to daemon socket: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}()

	cmd := ipc.Command{
		ID:     "cli-ping-1",
		Action: "ping",
		Data:   "Hello from Jay CLI client!",
	}

	msg := ipc.Message{
		Type:    "command",
		Payload: cmd,
	}

	log.Printf("Sending command: %+v", msg)

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(msg); err != nil {
		log.Fatalf("Failed to encode/send message: %v", err)
	}

	decoder := json.NewDecoder(conn)
	var resp ipc.Message
	if err := decoder.Decode(&resp); err != nil {
		log.Fatalf("Failed to read/decode response: %v", err)
	}

	log.Printf("Received response wrapper: %+v", resp)

	payloadBytes, err := json.Marshal(resp.Payload)
	if err != nil {
		log.Fatalf("Failed to marshal payload: %v", err)
	}

	var responsePayload ipc.Response
	if err := json.Unmarshal(payloadBytes, &responsePayload); err != nil {
		fmt.Printf("Raw Payload: %s\n", string(payloadBytes))
	} else {
		fmt.Printf("Response Status: %s\n", responsePayload.Status)
		if responsePayload.Data != nil {
			fmt.Printf("Response Data: %v\n", responsePayload.Data)
		}
	}
}
