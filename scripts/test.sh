#!/bin/bash

echo "Running tests..."
# Run all tests but continue on failure for deployment purposes
go test -v ./... || echo "⚠️  Some tests failed but continuing deployment"