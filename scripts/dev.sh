#!/bin/bash
set -e

echo "Building and running in development mode..."
bash scripts/build.sh
./mailos "$@"