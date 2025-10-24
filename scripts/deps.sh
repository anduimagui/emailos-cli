#!/bin/bash
set -e

echo "Downloading dependencies..."
go mod download
go mod tidy
echo "Dependencies downloaded."