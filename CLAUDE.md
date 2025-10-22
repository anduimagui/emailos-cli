# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MailOS is an AI-powered command-line email client built in Go that integrates with multiple email providers and AI systems. The CLI provides natural language email management, draft creation, template systems, and comprehensive email analytics.

## Core Architecture

### Main Components
- **CLI Entry Point**: `cmd/mailos/main.go` - Main application entry with Cobra command framework
- **Core Client**: `client.go` - Primary email client interface and operations
- **Configuration**: `config.go` - Config management supporting both global (`~/.email/config.json`) and local (`.email/config.json`) settings
- **Authentication**: `auth.go` - Email provider authentication and credential management
- **Email Operations**: 
  - `send.go` - Email sending with markdown-to-HTML conversion
  - `read.go` - Email retrieval with filtering and search
  - `drafts.go` - Draft management with IMAP synchronization
  - `search.go` - Advanced email search functionality
- **AI Integration**: `ai_*.go` files - Integration with Claude, GPT-4, and other AI providers
- **Interactive UI**: `interactive.go`, `input_handler*.go` - Terminal UI using BubbleTea and PromptUI

### Email Provider Support
- Gmail, Fastmail, Outlook, Yahoo, Zoho
- IMAP/SMTP with app-specific password authentication
- Dual storage: local drafts + IMAP draft folder synchronization

## Development Commands

### Building and Testing
```bash
# Build binary
go build -o mailos ./cmd/mailos
task build-simple              # Quick build
task build                     # Build with auth testing

# Install locally
task install                   # Install to GOPATH
task update                    # Update global installation

# Testing
go test ./...                  # Run all tests
task test                      # Run tests via Task
task test-alias                # Test alias functionality
task test-quick                # Quick command validation

# Development
task dev -- [args]            # Build and run with arguments
make run                       # Build and run via Makefile
```

### Code Quality
```bash
go fmt ./...                   # Format code
golangci-lint run ./...        # Run linter (if installed)
```

### Release Management
```bash
task release                   # Build cross-platform binaries
task publish-patch             # Auto-increment patch version and publish
task simulate-github-release   # Test release process locally
```

## Email System Testing Protocol

When testing email system functionality, ALWAYS use the actual mailos commands directly within the Go system. Do NOT use Python scripts or external tools to query email data. 

**Required approach:**
1. Test mailos commands first (e.g., `./mailos read`, `./mailos search`, `./mailos draft`)
2. Verify functionality works through the intended CLI interface
3. Only use alternative methods if the mailos commands fail
4. Always use `andrew@happysoft.dev` as test sending email

This ensures the actual email system functionality is being tested, not just the data files.

## Configuration System

### Global vs Local Configuration
- **Global**: `~/.email/config.json` - Default settings for all projects
- **Local**: `.email/config.json` - Project-specific overrides
- Local configs inherit credentials from global but can override display settings

### Key Configuration Fields
```go
type Config struct {
    Provider     string // Email provider (gmail, fastmail, etc.)
    Email        string // User email address
    Password     string // App-specific password
    FromName     string // Display name
    FromEmail    string // From address override
    LicenseKey   string // EmailOS license key
    DefaultAICLI string // AI provider (claude-code, gpt-4, etc.)
    Accounts     []AccountConfig // Multi-account support
}
```

## Email Draft System

### Draft Architecture
- **Local Storage**: `draft-emails/` directory for offline drafts
- **IMAP Sync**: Automatic upload to email provider's Drafts folder
- **Dual Commands**: `mailos draft` (single) and `mailos drafts` (batch operations)

### Draft Workflow
1. Create drafts: `mailos draft` or `mailos draft --to email --subject "Subject"`
2. Drafts saved locally and synced to IMAP
3. Send all drafts: `mailos send --drafts`

## AI Integration

### Supported Providers
- Claude (via Claude Code CLI)
- GPT-4 (via OpenAI CLI)
- Gemini (via Google CLI)
- Custom AI integrations

### AI Features
- Natural language email composition
- Smart email search queries
- Email content suggestions
- Automated draft generation

## Command Structure

### Core Commands
- `send` - Send emails with markdown support
- `read` - Read emails with filtering
- `draft/drafts` - Create and manage drafts
- `search` - Advanced email search
- `stats` - Email analytics with charts
- `interactive` - Terminal UI mode
- `setup` - Configuration wizard

### Important Flags
- `--local` - Use local configuration
- `--json` - JSON output format
- `--dry-run` - Preview without executing
- `--debug` - Enable debug output

## File Structure Patterns

### Primary Go Files
- `*_commands.go` - Command implementations
- `*_handler*.go` - Input/UI handling
- `*_sync.go` - Synchronization logic
- `test_*.go` - Test files

### Key Directories
- `cmd/mailos/` - CLI entry point
- `docs/` - User documentation
- `scripts/` - Build and deployment scripts
- `npm/` - NPM package distribution
- `Formula/` - Homebrew package

## Dependencies

### Core Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/bubbletea` - Terminal UI
- `github.com/emersion/go-imap` - IMAP client
- `github.com/manifoldco/promptui` - Interactive prompts
- `github.com/russross/blackfriday/v2` - Markdown processing

### License System
- Uses Polar (polarsource/polar-go) for license validation
- Requires active license for full functionality
- 7-day offline grace period

## Testing Best Practices

1. Always test with actual mailos commands, not external scripts
2. Use `andrew@happysoft.dev` for test emails
3. Test both global and local configuration scenarios
4. Verify IMAP draft synchronization works
5. Test cross-platform build compatibility