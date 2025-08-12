# EmailOS MCP Server Setup Guide

## Overview
Your EmailOS CLI (`mailos`) has been configured to work as an MCP (Model Context Protocol) server. This allows Claude Desktop and other AI tools to interact with your email system directly.

## What's Been Added
1. **MCP Server Support**: Added via the Ophis library for Cobra CLIs
2. **MCP Command**: `mailos mcp` command with subcommands:
   - `mailos mcp start` - Start the MCP server
   - `mailos mcp tools` - Export available tools as JSON
   - `mailos mcp claude` - Claude-specific integration

## Available MCP Tools
The MCP server exposes all your CLI commands as tools that Claude can use:
- Email sending and reading
- Draft management
- Email statistics and reports
- Template management
- Configuration management
- And all other mailos commands

## Installation Steps

### 1. Build the CLI (Already Done)
```bash
go build -o mailos cmd/mailos/main.go
```

### 2. Install Globally (Optional)
To make the CLI available system-wide:
```bash
# Option A: Copy to /usr/local/bin
sudo cp mailos /usr/local/bin/

# Option B: Add current directory to PATH
export PATH="$PATH:/Users/andrewmaguire/LOCAL/Github/_code-main/emailos"
```

### 3. Configure Claude Desktop

#### Method 1: Direct Configuration
1. Open Claude Desktop settings
2. Go to Developer → Edit Config
3. Add the following to your configuration:

```json
{
  "mcpServers": {
    "mailos": {
      "command": "/Users/andrewmaguire/LOCAL/Github/_code-main/emailos/mailos",
      "args": ["mcp", "start"],
      "env": {}
    }
  }
}
```

#### Method 2: Using Configuration File
We've created a `claude_mcp_config.json` file in this directory. 

To add it to Claude:
1. Copy the configuration from `claude_mcp_config.json`
2. Open Claude Desktop → Developer → Edit Config
3. Merge the "mailos" entry into your existing "mcpServers" section

### 4. Restart Claude Desktop
After adding the configuration, restart Claude Desktop for the changes to take effect.

## Testing the Integration

### Local Testing (Command Line)
```bash
# Test MCP server startup
./mailos mcp start

# Export available tools
./mailos mcp tools

# Test with MCP protocol
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}' | ./mailos mcp start
```

### Testing in Claude
Once configured, you can test in Claude by asking:
- "Can you list my emails?"
- "Send an email to someone@example.com"
- "Show me email statistics for today"
- "Create a draft email"

## Troubleshooting

### Issue: Claude doesn't see the mailos server
**Solution**: 
1. Check that the path in the config is absolute and correct
2. Ensure the mailos binary is executable: `chmod +x mailos`
3. Restart Claude Desktop

### Issue: Commands fail with permission errors
**Solution**: 
1. Ensure mailos has been configured: `./mailos setup`
2. Check that ~/.email/config.json exists and is readable

### Issue: MCP server doesn't start
**Solution**:
1. Test the binary directly: `./mailos mcp start`
2. Check for error messages in Claude's Developer Console
3. Verify Go dependencies: `go mod tidy`

## Advanced Configuration

### Custom Environment Variables
You can pass environment variables to the MCP server:
```json
{
  "mcpServers": {
    "mailos": {
      "command": "/path/to/mailos",
      "args": ["mcp", "start"],
      "env": {
        "MAILOS_CONFIG_PATH": "/custom/path/to/config"
      }
    }
  }
}
```

### Multiple Configurations
You can run multiple instances with different configurations:
```json
{
  "mcpServers": {
    "mailos-personal": {
      "command": "/path/to/mailos",
      "args": ["mcp", "start"],
      "env": {
        "MAILOS_CONFIG": "personal"
      }
    },
    "mailos-work": {
      "command": "/path/to/mailos",
      "args": ["mcp", "start"],
      "env": {
        "MAILOS_CONFIG": "work"
      }
    }
  }
}
```

## Development Notes

### How It Works
- The Ophis library automatically converts Cobra commands to MCP tools
- Each command becomes a tool with its flags as parameters
- The MCP server communicates via stdio using JSON-RPC
- Claude can discover and call these tools dynamically

### Extending the MCP Server
To add custom MCP-specific functionality:
1. Modify the `ophis.Command(nil)` call in main.go
2. Pass custom options or handlers as needed
3. Rebuild the binary

## Security Considerations
- The MCP server has access to your email configuration
- It can send emails on your behalf
- Only add to trusted AI applications
- Consider using read-only configurations for testing

## Next Steps
1. Configure your email account if not done: `./mailos setup`
2. Test basic commands: `./mailos read`, `./mailos stats`
3. Add to Claude Desktop using the configuration above
4. Start using natural language email commands in Claude!