#!/bin/bash
set -e

echo "Running linter..."
golangci-lint run ./... || echo "Install golangci-lint: https://golangci-lint.run/usage/install/"