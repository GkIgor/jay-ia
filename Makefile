.PHONY: build build-debug test clean run-core run-daemon run-cli

# Default target runs build
all: build

# Build both CLI (jay) and daemon (jayd) into the bin/ directory
build:
	go build -o bin/jay ./cli/cmd/jay
	go build -o bin/jayd ./core/cmd/jayd

build-debug:
	go build -gcflags="all=-N -l" -o bin/jay ./cli/cmd/jay
	go build -gcflags="all=-N -l" -o bin/jayd ./core/cmd/jayd

# Run all project tests
test:
	go test ./...

# Clean build outputs and delete the bin/ directory, and clean up any stray root binaries
clean:
	rm -rf bin
	rm -f jay jayd

# Build and run the daemon
run-core: build
	./bin/jayd

run-daemon: run-core

# Build and run the CLI
run-cli: build
	./bin/jay
