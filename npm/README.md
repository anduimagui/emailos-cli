# mailos

EmailOS - A standardized email client CLI with support for multiple providers

## Installation

```bash
npm install -g mailos
```

## Requirements

- Node.js 14.0.0 or higher
- Go 1.21 or higher (for building from source)

## Quick Start

1. **Setup your email account:**
   ```bash
   mailos setup
   ```

2. **Send an email:**
   ```bash
   mailos send -t recipient@example.com -s "Hello" -b "This is a test email"
   ```

3. **Read emails:**
   ```bash
   mailos read -n 10
   ```

## Features

- 📧 Multiple email provider support (Gmail, Fastmail, Zoho, Outlook, Yahoo)
- 📝 Markdown email composition (automatically converted to HTML)
- 🔍 Advanced email search and filtering
- 📎 Attachment support
- 🔗 Unsubscribe link detection
- 💾 Export emails to markdown files
- 🔒 Secure credential storage

## Commands

### Setup
```bash
mailos setup
```
Interactive setup wizard to configure your email account.

### Send
```bash
mailos send -t to@email.com -s "Subject" -b "Body"

# With attachments
mailos send -t to@email.com -s "Files" -b "See attached" -a file1.pdf -a file2.docx

# With CC and BCC
mailos send -t to@email.com -c cc@email.com -B bcc@email.com -s "Subject" -b "Body"
```

### Read
```bash
# Read last 10 emails
mailos read

# Read unread emails
mailos read --unread

# Search by sender
mailos read --from sender@example.com

# Save as markdown files
mailos read --save-markdown --output-dir emails/
```

### Mark as Read
```bash
# Mark specific emails
mailos mark-read --ids 1,2,3

# Mark all from sender
mailos mark-read --from notifications@example.com
```

### Delete
```bash
# Delete specific emails
mailos delete --ids 1,2,3 --confirm

# Delete all from sender
mailos delete --from spam@example.com --confirm
```

### Unsubscribe
```bash
# Find unsubscribe links
mailos unsubscribe --from newsletter@example.com

# Open unsubscribe link in browser
mailos unsubscribe --from newsletter@example.com --open
```

## Configuration

Configuration is stored in `~/.email/config.json` or in a local `.email/config.json` file.

## Building from Source

If you want to build from source instead of using pre-built binaries:

1. Clone the repository
2. Install Go 1.21+
3. Run `npm install -g .` from the npm directory

## License

MIT

## Support

For issues and feature requests, visit: https://github.com/anduimagui/emailos/issues