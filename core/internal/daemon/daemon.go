package daemon

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/GkIgor/jay-ia/core/internal/ipc"
	"github.com/GkIgor/jay-ia/core/internal/memory"
	"github.com/GkIgor/jay-ia/core/internal/state"
)

// Daemon coordinates all core components
type Daemon struct {
	currentState state.State
	memoryStore  memory.MemoryStore
	ipcServer    *ipc.Server
}

// New creates a new Jay Daemon
func New() (*Daemon, error) {
	ipcServer, err := ipc.NewServer()
	if err != nil {
		return nil, err
	}

	return &Daemon{
		currentState: state.Idle,
		memoryStore:  memory.NewInMemoryStore(),
		ipcServer:    ipcServer,
	}, nil
}

// Start starts the daemon and blocks until stopped
func (d *Daemon) Start() error {
	log.Printf("Starting Jay Daemon... Current State: %s", d.currentState)

	if err := d.ipcServer.Start(); err != nil {
		return err
	}

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	<-sigChan
	log.Println("Termination signal received. Shutting down...")

	d.Stop()
	return nil
}

// Stop safely shuts down the daemon
func (d *Daemon) Stop() {
	if d.ipcServer != nil {
		d.ipcServer.Stop()
	}
	log.Println("Jay Daemon stopped.")
}
