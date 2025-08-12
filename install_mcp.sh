#!/bin/bash

# EmailOS MCP Server Installation Script
# This script builds and configures mailos as an MCP server for Claude

set -e

echo "EmailOS MCP Server Installation"
echo "================================"
echo ""

# Get the current directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BINARY_PATH="$SCRIPT_DIR/mailos"

# Step 1: Build the binary
echo "Step 1: Building mailos with MCP support..."
cd "$SCRIPT_DIR"
go build -o mailos cmd/mailos/main.go
chmod +x mailos
echo "✓ Binary built successfully"
echo ""

# Step 2: Test MCP functionality
echo "Step 2: Testing MCP server..."
if echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}}' | ./mailos mcp start 2>/dev/null | grep -q "serverInfo"; then
    echo "✓ MCP server is working"
else
    echo "✗ MCP server test failed"
    exit 1
fi
echo ""

# Step 3: Export tools
echo "Step 3: Exporting MCP tools..."
./mailos mcp tools
echo "✓ Tools exported to mcp-tools.json"
echo ""

# Step 4: Create Claude configuration
echo "Step 4: Creating Claude configuration..."
cat > claude_mcp_config.json << EOF
{
  "mcpServers": {
    "mailos": {
      "command": "$BINARY_PATH",
      "args": ["mcp", "start"],
      "env": {}
    }
  }
}
EOF
echo "✓ Configuration created in claude_mcp_config.json"
echo ""

# Step 5: Instructions
echo "Installation Complete!"
echo "====================="
echo ""
echo "To complete the setup:"
echo ""
echo "1. Open Claude Desktop"
echo "2. Go to: Claude → Settings → Developer → Edit Config"
echo "3. Add the following to your configuration:"
echo ""
cat claude_mcp_config.json
echo ""
echo "4. Save and restart Claude Desktop"
echo ""
echo "The MCP server binary is located at: $BINARY_PATH"
echo ""
echo "To test locally, run:"
echo "  $BINARY_PATH mcp start"
echo ""
echo "For more details, see SETUP_MCP.md"