# EmailOS Command Reference

This file provides command references for EmailOS (mailos) to enable LLM integration.

## Available Commands

### Create Email Drafts
```bash
mailos draft [-t <recipient>] [-s <subject>] [-b <body>] [-c <cc>] [-B <bcc>] [-f <file>]

# Examples:
mailos draft                                              # Create draft interactively
mailos draft -t user@example.com -s "Meeting" -b "Let's meet at 3pm"
mailos draft -f email_content.md -s "Report"             # Body from file
mailos draft --interactive                               # Create multiple drafts

# Notes:
# - Drafts are saved both locally (draft-emails/) and to your email provider's Drafts folder
# - Use 'mailos send --drafts' to send all saved drafts
```

### Send Email
```bash
mailos send -t <recipient> -s <subject> -b <body> [-c <cc>] [-B <bcc>] [-f <file>]
mailos send --drafts                                     # Send all draft emails

# Examples:
mailos send -t user@example.com -s "Hello" -b "This is a test email"
mailos send -t alice@example.com -t bob@example.com -s "Team Update" -b "Meeting at 3pm"
mailos send -t recipient@example.com -s "Report" -f report.md
mailos send --drafts                                      # Send all drafts from draft-emails/
mailos send --drafts --dry-run                          # Preview what would be sent
```

### Read Emails
```bash
mailos read [--unread] [--from <sender>] [--days <n>] [-n <limit>]

# Examples:
mailos read                          # Read last 10 emails
mailos read --unread                 # Read only unread emails
mailos read --from sender@example.com # Read from specific sender
mailos read --days 7                 # Read emails from last 7 days
mailos read -n 20                    # Read last 20 emails
```

### Configure Email Settings
```bash
# Global configuration (default)
mailos configure [--email <email>] [--provider <provider>] [--name <name>] [--ai <ai>]

# Local configuration (project-specific)
mailos configure --local [--email <email>] [--provider <provider>] [--name <name>] [--ai <ai>]

# Examples:
mailos configure                     # Interactive global configuration
mailos configure --local             # Interactive local configuration
mailos configure --email user@gmail.com --provider gmail
mailos configure --local --email user@outlook.com --provider outlook --name "John Doe"
mailos configure --ai claude-code    # Set AI CLI provider globally

# Providers: gmail, outlook, yahoo, icloud, proton, fastmail, custom
# AI Options: claude-code, claude-code-yolo, openai, gemini, opencode, none
```

### Mark Emails as Read
```bash
mailos mark-read --ids <comma-separated-ids>

# Example:
mailos mark-read --ids 1,2,3
```

### Show Configuration
```bash
mailos info  # Display current email configuration (shows local or global)
```

### Setup/Reconfigure
```bash
mailos setup  # Run initial setup wizard (global configuration)
```

## Configuration Management

EmailOS supports both global and local configuration:

- **Global**: Stored in `~/.email/config.json`, used by default
- **Local**: Stored in `.email/config.json` in current directory, overrides global

Use `mailos configure --local` to create a local configuration for project-specific settings.

## Email Body Formatting

All email bodies support Markdown formatting:
- **Bold**: `**text**`
- *Italic*: `*text*`
- Headers: `# H1`, `## H2`, `### H3`
- Links: `[text](https://example.com)`
- Code blocks: ` ```code``` `
- Lists: `- item` or `* item`

## Notes for LLM Usage

1. The `mailos` command is available globally after installation
2. Local configuration (`.email/`) overrides global (`~/.email/`)
3. All commands return appropriate exit codes for error handling
4. Use `-f` flag to send email content from a file
5. Multiple recipients can be specified with multiple `-t` flags
6. The read command returns emails in chronological order
7. Email IDs are provided in the read output for use with mark-read

## Security Notes

- Credentials are stored locally in `.email/` or `~/.email/`
- Uses app-specific passwords, not main account passwords
- Configuration files have restricted permissions (600)
