#!/bin/bash
set -e

echo "Building binaries for all platforms..."
mkdir -p dist
GOOS=darwin GOARCH=amd64 go build -o dist/mailos-darwin-amd64 ./cmd/mailos
GOOS=darwin GOARCH=arm64 go build -o dist/mailos-darwin-arm64 ./cmd/mailos
GOOS=linux GOARCH=amd64 go build -o dist/mailos-linux-amd64 ./cmd/mailos
GOOS=linux GOARCH=arm64 go build -o dist/mailos-linux-arm64 ./cmd/mailos
GOOS=windows GOARCH=amd64 go build -o dist/mailos-windows-amd64.exe ./cmd/mailos
echo "✓ Release binaries created in dist/"