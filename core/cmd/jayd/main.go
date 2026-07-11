package main

import (
	"log"

	"github.com/GkIgor/jay-ia/core/internal/daemon"
)

func main() {
	log.Println("Jay Core (Headless Daemon) initializing...")

	d, err := daemon.New()
	if err != nil {
		log.Fatalf("Failed to initialize daemon: %v", err)
	}

	if err := d.Start(); err != nil {
		log.Fatalf("Daemon error: %v", err)
	}
}
