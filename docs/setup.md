# EmailOS Setup Command Documentation

The `mailos setup` command provides a guided wizard for initial EmailOS configuration, setting up your email account for the first time.

## Basic Usage

```bash
mailos setup
```

Launches the interactive setup wizard.

## Setup Process

### Step 1: Welcome
- Introduction to EmailOS
- Overview of setup process
- Requirements checklist

### Step 2: Email Provider Selection
Choose from supported providers:
- Gmail
- Outlook/Hotmail
- Yahoo Mail
- iCloud Mail
- ProtonMail
- Fastmail
- Custom (manual configuration)

### Step 3: Email Credentials
Enter your:
- Email address
- App-specific password (not regular password)

### Step 4: Test Connection
Automatic testing of:
- SMTP connection (sending)
- IMAP connection (receiving)
- Credential validation

### Step 5: Optional Configuration
- Display name
- Custom sender address
- AI provider selection (claude-code, openai, gemini, etc.)
- Profile image (used in email templates with {{PROFILE_IMAGE}})
- Interactive mode preference (Classic or React Ink UI)

### Step 6: Confirmation
- Review settings
- Save configuration
- Create README file

## Provider-Specific Setup

### Gmail Setup
1. Enable 2-factor authentication
2. Generate app password:
   - Go to Google Account settings
   - Security → 2-Step Verification
   - App passwords → Mail
3. Enable IMAP:
   - Gmail Settings → Forwarding and POP/IMAP
   - Enable IMAP

### Outlook/Hotmail Setup
1. Enable two-step verification
2. Create app password:
   - Security settings → Advanced security
   - Create a new app password
3. Settings automatically configured

### Yahoo Mail Setup
1. Enable two-step verification
2. Generate app password:
   - Account Security → Generate app password
   - Select "Other App"
3. Allow less secure apps (if needed)

### iCloud Mail Setup
1. Enable two-factor authentication
2. Generate app-specific password:
   - Apple ID → Security
   - App-Specific Passwords → Generate
3. Use @icloud.com email address

### ProtonMail Setup
1. Install ProtonMail Bridge
2. Get Bridge credentials:
   - Open ProtonMail Bridge
   - Get password for email client
3. Use Bridge-provided settings

### Fastmail Setup
1. Create app password:
   - Settings → Passwords & Security
   - App Passwords → New App Password
2. Use provided credentials

## Custom Provider Setup

For unlisted providers, you'll need:

### Required Information
- SMTP server hostname
- SMTP port (usually 587, 465, or 25)
- SMTP security (TLS/SSL/None)
- IMAP server hostname
- IMAP port (usually 993 or 143)
- IMAP security (TLS/SSL/None)

### Common Settings

| Provider | SMTP Server | SMTP Port | IMAP Server | IMAP Port |
|----------|------------|-----------|-------------|-----------|
| Gmail | smtp.gmail.com | 587 | imap.gmail.com | 993 |
| Outlook | smtp-mail.outlook.com | 587 | outlook.office365.com | 993 |
| Yahoo | smtp.mail.yahoo.com | 587 | imap.mail.yahoo.com | 993 |
| iCloud | smtp.mail.me.com | 587 | imap.mail.me.com | 993 |

## First-Time Setup

### Prerequisites
1. Email account with IMAP enabled
2. 2FA enabled (recommended)
3. App-specific password generated
4. Internet connection

### Quick Start
```bash
# Install EmailOS (if not already installed)
npm install -g @emailos/mailos

# Run setup
mailos setup

# Follow the prompts
```

### Verification
After setup, verify with:
```bash
# Check configuration
mailos info

# Test email reading
mailos read -n 1

# Test email sending
mailos send --to your-email@example.com --subject "Test" --body "Test message"
```

## Configuration Files

Setup creates:

### Global Configuration
`~/.email/config.json`
- Main configuration file
- Encrypted credentials
- Provider settings

### README File
`~/.email/README.md`
- Configuration documentation
- Troubleshooting guide
- Security notes

### Directory Structure
```
~/.email/
├── config.json         # Main configuration
├── README.md           # Documentation
├── template.html       # Email template (optional)
├── .slash_config.json  # Slash command preferences
├── ui/                 # React Ink UI components (auto-installed)
└── cache/              # Email cache (created on use)
```

## Security Considerations

### Password Security
- Never use your main account password
- Always use app-specific passwords
- Passwords are stored locally only
- File permissions set to 600 (owner only)

### Best Practices
1. Enable 2FA on email account
2. Use unique app passwords
3. Regularly rotate passwords
4. Don't share configuration files
5. Add `.email/` to `.gitignore`

## Troubleshooting Setup

### Common Issues

#### "Authentication Failed"
- Verify app password is correct
- Check 2FA is enabled
- Ensure IMAP/SMTP is enabled
- Try regenerating app password

#### "Connection Timeout"
- Check internet connection
- Verify firewall settings
- Try different ports
- Check VPN/proxy settings

#### "Invalid Provider"
- Use exact provider name
- Check spelling
- Use "custom" for unlisted providers

#### "Permission Denied"
- Check file permissions
- Ensure write access to home directory
- Run without sudo

### Reset Configuration
```bash
# Remove existing configuration
rm -rf ~/.email

# Run setup again
mailos setup
```

### Manual Configuration
If setup fails, create manually:
```bash
mkdir -p ~/.email
cat > ~/.email/config.json << 'EOF'
{
  "provider": "gmail",
  "email": "your-email@gmail.com",
  "password": "your-app-password"
}
EOF
chmod 600 ~/.email/config.json
```

## Multiple Accounts

### Primary Account
```bash
mailos setup
# Sets up global default account
```

### Secondary Accounts
```bash
# Create project-specific configuration
cd my-project
mailos configure --local
```

### Account Switching
```bash
# Use global account
mailos read

# Use local account (in project directory)
cd my-project
mailos read
```

## Advanced Setup Options

### Environment Variables
Override setup with environment:
```bash
export MAILOS_PROVIDER=gmail
export MAILOS_EMAIL=user@gmail.com
export MAILOS_PASSWORD=app-password
mailos setup
```

### Automated Setup
For scripted installation:
```bash
mailos setup --non-interactive \
  --provider gmail \
  --email user@gmail.com \
  --password "app-password"
```

### Import Configuration
```bash
# Copy from another machine
scp other-machine:~/.email/config.json ~/.email/
chmod 600 ~/.email/config.json
```

## Post-Setup

### Recommended Next Steps
1. Test email reading: `mailos read -n 5`
2. Configure template: `mailos template`
3. Set up AI provider: `mailos provider`
4. Create local config: `mailos local`

### Useful Commands
```bash
# View configuration
mailos info

# Configure AI provider
mailos provider

# Test email system
mailos test

# Read recent emails
mailos read
mailos read --json  # JSON output

# Send test email
mailos send --to me@example.com --subject "Test" --body "Hello"
mailos send --plain --file message.txt  # Send plain text from file

# Interactive mode
mailos interactive      # Classic UI
mailos interactive --ink  # React Ink UI
mailos chat            # AI chat interface

# View email statistics
mailos stats --days 30  # With visual charts
```

## Migration from Other Clients

### From Gmail Web
- Export is not needed
- EmailOS connects directly via IMAP
- All emails remain on server

### From Outlook
- No migration needed
- Uses existing email account
- Maintains folder structure

### From Thunderbird
- Copy SMTP/IMAP settings
- Use same app password
- Folders sync automatically

## Getting Help

### Resources
- Documentation: `mailos help`
- GitHub Issues: https://github.com/corp-os/emailos
- Community Forum: https://email-os.com/forum

### Support Commands
```bash
# Check system
mailos test

# View logs
mailos --debug

# Get version
mailos --version
```