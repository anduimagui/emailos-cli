#!/bin/bash
set -e

BINARY_NAME=mailos

echo "Cleaning..."
rm -f ${BINARY_NAME}
go clean
echo "Clean complete."