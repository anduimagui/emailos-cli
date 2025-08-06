#!/bin/bash

echo "📦 Preparing to publish mailos to NPM..."

# Clean up any existing binaries
rm -rf bin/

# Check if logged in to NPM
if ! npm whoami >/dev/null 2>&1; then
    echo "❌ Not logged in to NPM. Please run 'npm login' first."
    exit 1
fi

# Dry run first
echo "Running dry-run..."
npm publish --dry-run

echo ""
echo "📋 Review the above output. Continue with publish? (y/N)"
read -r response

if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo "Publishing to NPM..."
    npm publish
    echo "✅ Published successfully!"
else
    echo "❌ Publish cancelled."
fi