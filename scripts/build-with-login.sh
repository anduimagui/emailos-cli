#!/bin/bash
# build-with-login.sh - Build and test authentication pipeline

set -e

echo "Building mailos binary..."
go build -o mailos cmd/mailos/main.go cmd/mailos/error_handler.go

echo "Testing authentication pipeline..."
./mailos --version
echo "Binary built successfully"
echo ""

echo "Testing login functionality..."
if [ -f ~/.email/config.json ]; then
    echo "Existing email configuration found"
    ./mailos config --show || echo "Could not display config (may be normal)"
else
    echo "No email configuration found"
    echo "Run 'mailos setup' to configure authentication"
fi
echo ""

echo "Testing deployment integration..."
if command -v hcloud >/dev/null 2>&1; then
    echo "Hetzner Cloud CLI found"
    if hcloud context active >/dev/null 2>&1; then
        echo "Hetzner Cloud context is active"
        echo "Current context: $(hcloud context active)"
    else
        echo "No active Hetzner Cloud context"
        echo "Run 'hcloud context create' to set up deployment"
    fi
else
    echo "Hetzner Cloud CLI not found"
    echo "Install with: brew install hcloud"
fi
echo ""

echo "Checking deployment configuration..."
if [ -f deployment/Taskfile.yml ]; then
    echo "Deployment taskfile found"
    if [ -f deployment/config/cloud-init.yaml ]; then
        echo "Cloud-init configuration found"
    else
        echo "Cloud-init configuration missing"
    fi
else
    echo "Deployment taskfile not found"
fi
echo ""

echo "Pipeline validation complete!"
echo "========================================"
echo "Binary: mailos built successfully"
echo "Authentication: $([ -f ~/.email/config.json ] && echo 'configured' || echo 'needs setup')"
echo "Deployment: $(command -v hcloud >/dev/null 2>&1 && echo 'ready' || echo 'needs hcloud')"
echo "========================================"
echo ""
echo "Next steps:"
echo "- To configure email: mailos setup"
echo "- To test deployment: cd deployment && task deploy"
echo "- To publish release: task publish-patch"