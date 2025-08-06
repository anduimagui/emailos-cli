# EmailOS - Standardized Email Client

EmailOS (mailos) is a command-line email client written in Go that provides a standardized interface for managing emails across multiple providers.

## Features

- 🔧 Interactive setup wizard with provider selection
- 📧 Support for multiple email providers (Gmail, Fastmail, Zoho, Outlook, Yahoo)
- 🔒 Secure credential storage with app-specific passwords
- ✉️ Send emails with Markdown formatting (automatically converted to HTML)
- 📥 Read emails with filtering options
- 🎯 Mark emails as read
- 🚀 Fast and efficient IMAP/SMTP implementation

## Installation

### Via Homebrew (macOS/Linux)

```bash
brew tap emailos/mailos
brew install mailos
```

### Via npm (All Platforms)

```bash
npm install -g mailos
```

### Via Go

```bash
go install github.com/emailos/mailos@latest
```

### Build from Source

```bash
git clone https://github.com/emailos/mailos
cd mailos
go build -o mailos .
```

### Direct Download

Download pre-built binaries from [GitHub Releases](https://github.com/emailos/mailos/releases)

## Quick Start

### 1. Initial Setup

Run the interactive setup wizard:

```bash
mailos setup
```

The setup will:
- Display "Welcome to EmailOS!"
- Let you select your email provider using arrow keys
- Prompt for your email address
- Open your browser to create an app-specific password
- Securely store your configuration in `~/.email/config.json`

### 2. Send an Email

```bash
# Simple email
mailos send -t recipient@example.com -s "Hello" -m "This is a test email"

# With CC and BCC
mailos send -t to@example.com -c cc@example.com -b bcc@example.com -s "Meeting" -m "Let's meet tomorrow"

# From a Markdown file
mailos send -t recipient@example.com -s "Report" -f report.md

# Multiple recipients
mailos send -t alice@example.com -t bob@example.com -s "Team Update" -m "Here's the update..."
```

### 3. Read Emails

```bash
# Read last 10 emails
mailos read

# Read unread emails only
mailos read --unread

# Read emails from specific sender
mailos read --from sender@example.com

# Read emails from last 7 days
mailos read --days 7

# Read emails from specific time ranges
mailos read --range "Last hour"
mailos read --range "Today"
mailos read --range "Yesterday"
mailos read --range "This morning"

# Read with custom limit
mailos read -n 20
```

### 4. Mark Emails as Read

```bash
# Mark specific emails as read
mailos mark-read --ids 1,2,3
```

### 5. Generate Email Report

```bash
# Interactive time range selection
mailos report

# Specify time range directly
mailos report --range "Last hour"
mailos report --range "Today"
mailos report --range "Yesterday"
mailos report --range "This week"
mailos report --range "Last week"
mailos report --range "This morning"
mailos report --range "Yesterday morning"
mailos report --range "Last 3 days"
mailos report --range "Last 30 days"

# Save report to file
mailos report --range "Today" --output today_report.txt
```

### 6. Show Configuration

```bash
mailos info
```

## Configuration

EmailOS supports both global and local configuration:

### Global Configuration
Stored in `~/.email/config.json` - used by default across all directories.

```bash
# Interactive global configuration (default)
mailos configure

# With command-line arguments
mailos configure --email user@gmail.com --provider gmail
mailos configure --email user@outlook.com --provider outlook --name "John Doe"
mailos configure --ai claude-code  # Set AI CLI provider
```

### Local Configuration
Stored in `.email/config.json` in your project directory - overrides global configuration when present.

```bash
# Interactive local configuration
mailos configure --local

# With command-line arguments
mailos configure --local --email user@gmail.com --provider gmail
mailos configure --local --email user@outlook.com --provider outlook --name "John Doe"
mailos configure --local --ai claude-code  # Set AI CLI provider for this project
```

**Available Providers:** gmail, outlook, yahoo, icloud, proton, fastmail, custom
**AI CLI Options:** claude-code, claude-code-yolo, openai, gemini, opencode, none

### Configuration Structure

```json
{
  "provider": "gmail",
  "email": "your@email.com",
  "password": "your-app-password",
  "smtp_host": "smtp.gmail.com",
  "smtp_port": 587,
  "smtp_use_tls": true,
  "imap_host": "imap.gmail.com",
  "imap_port": 993,
  "from_name": "Your Name",
  "default_ai_cli": "claude-code"
}
```

## Markdown Support

All email bodies are treated as Markdown and automatically converted to HTML:

- **Bold text**: `**bold**`
- *Italic text*: `*italic*`
- Headers: `# H1`, `## H2`, `### H3`
- Links: `[text](https://example.com)`
- Lists: `- item` or `* item`
- Code blocks: ` ```code``` `

## Security

- Credentials are stored with restricted file permissions (600)
- App-specific passwords are recommended for all providers
- Never share or commit your `.email` directory
- Local configurations (`.email/`) override global (`~/.email/`)
- Add `.email/` to your `.gitignore` when using local configuration

## Supported Providers

| Provider | SMTP Server | IMAP Server | Notes |
|----------|------------|-------------|--------|
| Gmail | smtp.gmail.com:587 | imap.gmail.com:993 | Requires app password |
| Fastmail | smtp.fastmail.com:465 | imap.fastmail.com:993 | Use device password |
| Zoho | smtp.zoho.com:465 | imap.zoho.com:993 | Generate app password |
| Outlook | smtp-mail.outlook.com:587 | outlook.office365.com:993 | Enable 2FA first |
| Yahoo | smtp.mail.yahoo.com:587 | imap.mail.yahoo.com:993 | Create app password |

## Command Reference

This section documents all available emailos commands for automated usage:

### Send Email
```bash
mailos send -t <recipient> -s <subject> -m <message> [-c <cc>] [-b <bcc>] [-f <file>]
```

### Read Emails
```bash
mailos read [--unread] [--from <sender>] [--days <n>] [--range <time-range>] [-n <limit>]
```

### Generate Report
```bash
mailos report [--range <time-range>] [--output <file>]
```

### Mark as Read
```bash
mailos mark-read --ids <comma-separated-ids>
```

### Setup
```bash
mailos setup           # Initial setup wizard (creates global config)
mailos configure       # Modify global configuration
mailos configure --local  # Create/modify local project configuration
```

### Show Info
```bash
mailos info  # Display current configuration (shows local or global)
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.