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

# Interactive mode
mailos interactive

# AI-powered email (requires AI setup)
mailos "send an email to john@example.com thanking him for the meeting"
```

## 📖 Documentation

- [Installation Guide](docs/installation.md) - Detailed installation instructions
- [Setup Guide](docs/setup.md) - Configuration and provider setup
- [Usage Guide](docs/usage.md) - Complete command reference
- [AI Integration](docs/ai-integration.md) - Setting up AI features
- [License Integration](docs/LICENSE_INTEGRATION.md) - License system details

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

MailOS stores configuration in `~/.email/config.json`:

```json
{
  "provider": "gmail",
  "email": "your-email@gmail.com",
  "from_name": "Your Name",
  "license_key": "your-license-key",
  "default_ai_cli": "claude-code"
}
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

- **Latest Version**: 0.1.6
- **Downloads**: Available on [npm](https://www.npmjs.com/package/mailos)
- **Stars**: [![GitHub stars](https://img.shields.io/github/stars/corp-os/emailos.svg)](https://github.com/corp-os/emailos/stargazers)
- **License**: [Proprietary](https://email-os.com)

## 🎉 What's New

### Version 0.1.6
- ✅ Multi-platform support (macOS, Linux, Windows)
- ✅ npm package distribution
- ✅ AI integration with multiple providers
- ✅ Interactive TUI mode
- ✅ Markdown email support
- ✅ License validation system

### Coming Soon
- 📱 Mobile companion app
- 🔄 Email synchronization
- 📊 Analytics dashboard
- 🎨 Custom themes
- 🔌 Plugin system

---

Made with ❤️ by the [EmailOS Team](https://email-os.com)

© 2024 EmailOS. All rights reserved.