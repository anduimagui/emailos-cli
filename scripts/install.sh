#!/bin/bash
set -e

BINARY_PATH=cmd/mailos/main.go

echo "Installing mailos..."
go install ${BINARY_PATH}
echo "✓ Installed mailos to $(go env GOPATH)/bin/mailos"