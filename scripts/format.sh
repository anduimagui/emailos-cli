#!/bin/bash
set -e

echo "Formatting code..."
go fmt ./...
echo "Code formatted."