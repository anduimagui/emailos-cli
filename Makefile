.PHONY: build install clean test deps task-% help-tasks

# Build the binary (uses shared script)
build:
	@bash scripts/build.sh

# Install the binary to GOPATH (uses shared script)
install:
	@bash scripts/install.sh

# Install locally for system-wide use on Mac
local-install:
	@echo "Installing mailos locally for system-wide use..."
	@./build-local.sh

# Clean build artifacts (uses shared script)
clean:
	@bash scripts/clean.sh

# Download dependencies (uses shared script)
deps:
	@bash scripts/deps.sh

# Run tests (uses shared script)
test:
	@bash scripts/test.sh

# Run the setup wizard
setup: build
	@./mailos setup

# Quick run for development
run: build
	@./mailos

# Format code (uses shared script)
fmt:
	@bash scripts/format.sh

# Run linter (uses shared script)
lint:
	@bash scripts/lint.sh

# Delegate to task runner for complex commands
task-%:
	@task $*

# Show available tasks from Taskfile
help-tasks:
	@echo "Available Taskfile commands (use 'make task-<command>'):"
	@task --list

# Build for all platforms (delegates to task)
release:
	@task release

# Version info (delegates to task)
version:
	@task version

# Update global installation (delegates to task)
update:
	@task update

# Development build and run (delegates to task)
dev:
	@task dev

# Comprehensive testing (delegates to task)
test-all:
	@task test-all

# Quick tests (delegates to task)
test-quick:
	@task test-quick

# Patch release (delegates to task)
publish-patch:
	@task publish-patch

# Show help
help:
	@echo "Available commands:"
	@echo ""
	@echo "Basic commands (using shared scripts):"
	@echo "  make build         - Build the binary"
	@echo "  make install       - Install to GOPATH"
	@echo "  make local-install - Install system-wide on Mac"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make deps          - Download dependencies"
	@echo "  make test          - Run tests"
	@echo "  make fmt           - Format code"
	@echo "  make lint          - Run linter"
	@echo ""
	@echo "Application commands:"
	@echo "  make setup         - Run the setup wizard"
	@echo "  make run           - Build and run"
	@echo ""
	@echo "Advanced commands (delegated to Taskfile):"
	@echo "  make release       - Build for all platforms"
	@echo "  make version       - Show current version"
	@echo "  make update        - Update global installation"
	@echo "  make dev           - Development build and run"
	@echo "  make test-all      - Comprehensive testing"
	@echo "  make test-quick    - Quick tests"
	@echo "  make publish-patch - Patch release"
	@echo ""
	@echo "Task delegation:"
	@echo "  make task-<name>   - Run any Taskfile command"
	@echo "  make help-tasks    - Show all available Taskfile commands"
	@echo "  make help          - Show this help"

# Default target
all: deps build