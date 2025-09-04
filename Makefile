# Makefile for the mangle-service project

.PHONY: build run test clean

# Build the Go application
build:
	@echo "Building the application..."
	go build -o mangle-service ./cmd/server/main.go

# Run the application
run:
	@echo "Running the application..."
	go run ./cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean the built binary
clean:
	@echo "Cleaning up..."
	rm -f mangle-service
