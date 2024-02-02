# Makefile to build and run a Go program with sudo

# Name of your Go binary
BINARY_NAME=myprogram

# Build the Go program
build:
	@go build -o $(BINARY_NAME) ./cmd/cdc-ncm/main.go

# Run the Go program with sudo
run: build
	@sudo ./$(BINARY_NAME)

# Build and run
up: build run

# Phony targets
.PHONY: build run up
