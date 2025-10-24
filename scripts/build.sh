#!/bin/bash
set -e

BINARY_NAME=mailos
BINARY_PATH=cmd/mailos/main.go

echo "Building ${BINARY_NAME}..."
go build -o ${BINARY_NAME} ${BINARY_PATH}
echo "Build complete: ./${BINARY_NAME}"