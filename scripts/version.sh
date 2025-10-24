#!/bin/bash
set -e

echo "Current version: v$(cd npm && node -p "require('./package.json').version")"