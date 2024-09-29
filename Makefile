# Makefile to build and run go-ios and the cdc-ncm network driver
# cdc-ncm needs to be executed with sudo on Linux for USB Access and setting
# up virtual TAP network devices.
# use Make build to build both binaries. 
# Make run is a simple target that just runs the cdc-ncm driver with sudo
# For development, use "make up" to rebuild and run cdc-ncm quickly

# Name of your Go binaries
GO_IOS_BINARY_NAME=ios
NCM_BINARY_NAME=go-ncm

# Define only if compiling for system different than our own
OS=
ARCH=

# Prepend each non-empty OS/ARCH definition to "go" command
GOEXEC=$(strip $(foreach v,OS ARCH,$(and $($v),GO$v=$($v) )) go)

# Build the Go program
build:
	@$(GOEXEC) work use .
	@$(GOEXEC) build -o $(GO_IOS_BINARY_NAME) ./main.go
	@$(GOEXEC) work use ./ncm
	@CGO_ENABLED=1 $(GOEXEC) build -o $(NCM_BINARY_NAME) ./cmd/cdc-ncm/main.go

# Run the Go program with sudo
run: build
	@sudo ./$(NCM_BINARY_NAME) --prometheusport=8080

# Build and run
up: build run

# Phony targets
.PHONY: build run up
