# MailOS - AI-Powered Command Line Email Client

[![npm version](https://img.shields.io/npm/v/mailos.svg)](https://www.npmjs.com/package/mailos)
[![GitHub release](https://img.shields.io/github/release/corp-os/emailos.svg)](https://github.com/corp-os/emailos/releases)
[![License](https://img.shields.io/badge/license-Proprietary-blue.svg)](https://email-os.com)

MailOS is a powerful command-line email client that brings AI automation to your terminal. Send, read, and manage emails across multiple providers with natural language commands.

## ✨ Features

### 🤖 AI Integration
- **Natural Language Commands**: "Send an email to John about the meeting tomorrow"
- **Smart Email Composition**: AI helps write professional emails
- **Intelligent Search**: Find emails using natural language queries
- **Multiple AI Providers**: Supports Claude, GPT-4, Gemini, and more

### 📧 Email Management
- **Multi-Provider Support**: Gmail, Fastmail, Outlook, Yahoo, Zoho
- **Markdown Formatting**: Write emails in Markdown, automatically converted to HTML
- **Interactive Mode**: Browse and manage emails with a TUI interface
- **Batch Operations**: Process multiple emails efficiently
- **Template System**: Save and reuse email templates

### 🔒 Security
- **App-Specific Passwords**: Never store your main password
- **Local Storage**: Credentials stored securely on your machine
- **License Protection**: Enterprise-grade license validation
- **Encrypted Communications**: Secure IMAP/SMTP connections

## 🚀 Quick Start

### Installation

```bash
# Install via npm (recommended)
npm install -g mailos

# Or download for your platform
# See full installation guide: docs/installation.md
```

### Initial Setup

```bash
# Run interactive setup wizard
mailos setup
```

The setup wizard will:
1. Validate your license key (get one at https://email-os.com)
2. Configure your email provider
3. Set up app-specific password
4. Configure AI integration (optional)

### Basic Usage

```bash
# Send an email
mailos send user@example.com "Meeting Tomorrow" "Let's discuss the project at 3pm"

# Read recent emails
mailos read --limit 10

# Interactive mode (classic UI)
mailos interactive

# Interactive mode with React Ink UI
mailos interactive --ink
mailos chat  # Always uses React Ink UI

# AI-powered email (requires AI setup)
mailos "send an email to john@example.com thanking him for the meeting"

# Query emails with natural language
mailos q="unread emails from last week"
mailos "emails with attachments"

# Generate email statistics with visual charts
mailos stats --days 30
mailos stats --range "last week"

# Advanced search options
mailos read --from john@example.com --unread --days 7
mailos read --subject "invoice" --limit 20
mailos read --json  # Output in JSON format
```

## 📘 Command Reference

```bash
# Core Commands
mailos send <email> <subject> <body>     # Send email
mailos send --file body.txt              # Send with body from file
mailos send --plain                      # Send as plain text only
mailos read [--limit N] [--unread]       # Read emails
mailos read --json                       # Output as JSON
mailos interactive                       # Interactive TUI mode
mailos interactive --ink                 # React Ink UI mode
mailos chat                               # AI chat interface
mailos setup                              # Configuration wizard
mailos local                              # Create local config for current directory
mailos configure [--local]                # Manage configuration
mailos provider                           # Configure AI provider
mailos open [--from email] [--last N]    # Open emails in mail client

# Draft Management
mailos drafts --interactive              # Create drafts interactively
mailos drafts --ai "query" --count N     # Generate drafts with AI
mailos send --drafts                      # Send all draft emails
mailos send --drafts --dry-run           # Preview drafts before sending
mailos send --drafts --filter="priority:high"  # Send filtered drafts

# Query & Search
mailos q="<natural language query>"      # AI-powered search
mailos "<query>"                          # Alternative query syntax
mailos stats [--days N] [--range "week"]  # Email statistics with charts

# Advanced Query Filters
--sent=true/false         # Filter sent vs received
--attachments=true/false  # Filter by attachments
--min-size=5KB           # Minimum email size
--max-size=10MB          # Maximum email size
--domains=gmail.com      # Filter by domains
--keywords=meeting       # Search keywords
--group-by=sender        # Group results
--format=json            # Output format

# Standard Filters (for read, delete, mark-read)
--from <email>    --to <email>    --subject <text>
--unread          --days <N>       --limit <N>
--range <range>   # e.g., "yesterday", "last week", "this month"

# Batch Operations  
mailos mark-read [filters]               # Mark as read
mailos delete [filters]                   # Delete emails
mailos export --format md --output dir    # Export emails
mailos unsubscribe [--auto-open]         # Find unsubscribe links

# Templates
mailos template [create|edit|list|delete] # Manage templates
# Templates support {{BODY}} and {{PROFILE_IMAGE}} placeholders
```

## 📖 Documentation

- [Installation Guide](docs/installation.md) - Detailed installation instructions
- [Setup Guide](docs/setup.md) - Configuration and provider setup
- [Usage Guide](docs/usage.md) - Complete command reference
- [AI Integration](docs/ai-integration.md) - Setting up AI features
- [License Integration](docs/LICENSE_INTEGRATION.md) - License system details

## ⌨️ Interactive Mode Keyboard Shortcuts

- **Enter** - Submit query or select option
- **ESC ESC** - Clear current input (press ESC twice quickly)
- **/** - Show command menu
- **↑↓** - Navigate menu options
- **Ctrl+C** - Cancel/Go back
- **Ctrl+D** - Exit (when input is empty)
- **Backspace** - Delete character
- **Tab** - Auto-complete (where available)

### Slash Commands in Interactive Mode

Type `/` to see available commands or use directly:
- `/read` - Browse and read emails
- `/send` - Compose new email
- `/stats` - View email statistics
- `/template` - Manage templates
- `/provider` - Configure AI provider
- `/exit` - Exit the application

## 🎯 Use Cases

### For Developers
- Send automated reports from CI/CD pipelines
- Monitor system alerts via email
- Quick email responses without leaving the terminal

### For Power Users
- Manage multiple email accounts from one interface
- Batch process emails with scripts
- Create email workflows with AI automation

### For Teams
- Standardized email templates
- Automated email responses
- Integration with existing CLI tools

## 🛠️ Supported Platforms

| Platform | Architecture | Status |
|----------|-------------|--------|
| macOS | Intel (x64) | ✅ Supported |
| macOS | Apple Silicon (ARM64) | ✅ Supported |
| Linux | x64 | ✅ Supported |
| Linux | ARM64 | ✅ Supported |
| Windows | x64 | ✅ Supported |

## 📦 Installation Options

### npm (All Platforms)
```bash
npm install -g mailos
```

### Direct Download
Download the latest binary for your platform from [GitHub Releases](https://github.com/corp-os/emailos/releases).

### Build from Source
```bash
git clone https://github.com/corp-os/emailos.git
cd emailos
go build -o mailos .
```

## 🔧 Configuration

MailOS supports both global and local configurations:

### Global Configuration
Stored in `~/.email/config.json` - applies to all projects:

```json
{
  "provider": "gmail",
  "email": "your-email@gmail.com",
  "from_name": "Your Name",
  "license_key": "your-license-key",
  "default_ai_cli": "claude-code"
}
```

### Local Configuration
Create project-specific settings with the `local` command:

```bash
# Create local config in current directory
mailos local

# Or use configure with --local flag
mailos configure --local
```

Local configuration (`.email/config.json`) inherits from global settings but allows you to override:
- From email address (different sender for this project)
- Display name (project-specific name)
- AI CLI provider (different AI for this project)

### Quick Configuration Updates
Update configuration settings directly via command-line flags:

```bash
# Update display name
mailos configure --name "John Doe"

# Update from email
mailos configure --from "john@example.com"

# Update AI provider
mailos configure --ai "claude-code"

# Update local configuration
mailos configure --local --name "Project Bot"
mailos configure --local --ai "claude-code-yolo"
```

**How it works:**
1. MailOS first checks for `.email/config.json` in the current directory
2. If found, it loads the global config and applies local overrides
3. Credentials (email/password) are always inherited from global for security
4. Local configs are automatically added to `.gitignore`

**Example use case:**
```bash
# Global config: personal@gmail.com
cd ~/work/project
mailos local
# Set from_email to work@company.com for this project
# Now emails sent from this directory appear from work@company.com
```

### Supported Email Providers

- **Gmail** - Full support with app passwords
- **Fastmail** - Native IMAP/SMTP support
- **Outlook/Hotmail** - Microsoft account integration
- **Yahoo Mail** - App password support
- **Zoho Mail** - Full IMAP/SMTP support
- **Custom IMAP/SMTP** - Any compatible provider

## 🤝 AI Providers

MailOS integrates with popular AI tools:

- Claude (via Claude Code CLI)
- GPT-4 (via OpenAI CLI)
- Gemini (via Google CLI)
- Custom AI integrations

## 📄 License

MailOS is proprietary software. A valid license is required for use.

- **Purchase License**: https://email-os.com/checkout
- **License Validation**: Automatic with internet connection
- **Offline Grace Period**: 7 days for offline usage

## 🆘 Getting Help

### Resources
- **Website**: https://email-os.com
- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/corp-os/emailos/issues)

### Common Issues

**License Validation Failed**
- Ensure you have an active internet connection
- Verify your license at https://email-os.com
- Run `mailos setup` to re-enter your license key

**Email Provider Connection Issues**
- Enable IMAP/SMTP in your email settings
- Generate an app-specific password
- Check firewall/antivirus settings

**Command Not Found**
- Ensure mailos is in your PATH
- Try reinstalling with `npm install -g mailos`
- See [installation troubleshooting](docs/installation.md#troubleshooting)

## 🚀 Development

### Prerequisites
- Go 1.23+
- Node.js 18+ (for npm package)
- Git

### Building
```bash
# Clone repository
git clone https://github.com/corp-os/emailos.git
cd emailos

# Build binary
go build -o mailos .

# Run tests
go test ./...

# Build for all platforms
task release
```

### Contributing
We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📊 Stats

- **Latest Version**: 0.1.8 (dev version with unreleased features)
- **Downloads**: Available on [npm](https://www.npmjs.com/package/mailos)
- **Stars**: [![GitHub stars](https://img.shields.io/github/stars/corp-os/emailos.svg)](https://github.com/corp-os/emailos/stargazers)
- **License**: [Proprietary](https://email-os.com)

## 📄 Changelog

See [CHANGELOG.md](CHANGELOG.md) for detailed version history and updates.

---

Made with ❤️ by the [EmailOS Team](https://email-os.com)

© 2024 EmailOS. All rights reserved.