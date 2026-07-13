package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

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

	// Parse arguments
	inputData := "Hello from Jay CLI client!"
	action := "ping"
	if len(os.Args) > 1 {
		inputData = strings.Join(os.Args[1:], " ")
		if strings.HasPrefix(inputData, "/") {
			parts := strings.SplitN(inputData, " ", 2)
			action = strings.TrimPrefix(parts[0], "/")
		} else {
			action = "chat"
		}
	}

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
		ID:     "cli-cmd-1",
		Action: action,
		Data:   inputData,
	}

	msg := ipc.Message{
		Type:    "command",
		Payload: cmd,
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(msg); err != nil {
		log.Fatalf("Failed to encode/send message: %v", err)
	}

	decoder := json.NewDecoder(conn)
	var resp ipc.Message
	if err := decoder.Decode(&resp); err != nil {
		log.Fatalf("Failed to read/decode response: %v", err)
	}

	payloadBytes, err := json.Marshal(resp.Payload)
	if err != nil {
		log.Fatalf("Failed to marshal payload: %v", err)
	}

	var responsePayload ipc.Response
	if err := json.Unmarshal(payloadBytes, &responsePayload); err != nil {
		fmt.Printf("Raw Payload: %s\n", string(payloadBytes))
	} else {
		if responsePayload.Status == "error" {
			fmt.Printf("Error: %v\n", responsePayload.Data)
		} else {
			fmt.Printf("%v\n", responsePayload.Data)
		}
	}
}
