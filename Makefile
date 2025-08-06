.PHONY: build install clean test deps

# Binary name
BINARY_NAME=mailos
BINARY_PATH=cmd/mailos/main.go

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) $(BINARY_PATH)
	@echo "Build complete: ./$(BINARY_NAME)"

# Install the binary to GOPATH
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install $(BINARY_PATH)
	@echo "Installation complete. Run 'mailos' to get started."

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@go clean
	@echo "Clean complete."

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies downloaded."

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run the setup wizard
setup: build
	@./$(BINARY_NAME) setup

# Quick run for development
run: build
	@./$(BINARY_NAME)

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted."

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./... || echo "Install golangci-lint: https://golangci-lint.run/usage/install/"

# Show help
help:
	@echo "Available commands:"
	@echo "  make build    - Build the binary"
	@echo "  make install  - Install to GOPATH"
	@echo "  make clean    - Clean build artifacts"
	@echo "  make deps     - Download dependencies"
	@echo "  make test     - Run tests"
	@echo "  make setup    - Run the setup wizard"
	@echo "  make run      - Build and run"
	@echo "  make fmt      - Format code"
	@echo "  make lint     - Run linter"
	@echo "  make help     - Show this help"

# Default target
all: deps build