# Draft Management

The `mailos draft` command (alias for `drafts`) provides comprehensive draft email management with automatic synchronization to your email account's Drafts folder.

## Overview

Create, manage, and send draft emails with full IMAP synchronization. Drafts are saved both locally and to your email provider's Drafts folder, ensuring they're accessible from any email client.

## Basic Usage

```bash
# Create a single draft interactively
mailos draft --interactive

# Create multiple drafts (alias)
mailos drafts --interactive

# List drafts from IMAP Drafts folder
mailos draft --list

# Read full draft content from IMAP
mailos draft --read

# Generate drafts with AI
mailos draft --ai "create follow-up email for meeting" --count 3

# Send all drafts
mailos send --drafts
```

## Key Features

### 📮 Automatic IMAP Synchronization
- Drafts are automatically saved to your email account's Drafts folder
- Compatible with Gmail, Outlook, Yahoo, and other IMAP providers
- Drafts appear instantly in your email client
- Proper RFC 822 email format with `\Draft` IMAP flag

### 📝 Local and Remote Storage
- Local files saved to `draft-emails/` directory
- Remote copies in your email provider's Drafts folder
- Dual storage provides backup and cross-device access

### 🎯 Smart Draft Detection
- Automatically detects the correct Drafts folder name
- Handles provider-specific folders ([Gmail]/Drafts, INBOX.Drafts, etc.)
- Creates Drafts folder if it doesn't exist

## Command Options

```bash
mailos draft [options]
```

| Option | Description |
|--------|-------------|
| `--list`, `-l` | List drafts from IMAP Drafts folder |
| `--read`, `-r` | Read full content of drafts from IMAP |
| `--interactive`, `-i` | Create drafts interactively with prompts |
| `--ai` | Use AI to generate drafts from a query |
| `--count N`, `-n N` | Number of drafts to generate (with AI) |
| `--template STRING` | Use a template for draft generation |
| `--data FILE` | Data file (CSV/JSON) for bulk generation |
| `--output DIR` | Output directory (default: "draft-emails") |

## Draft File Format

Drafts are saved as Markdown files with YAML frontmatter:

```markdown
---
to: recipient@example.com
cc: cc@example.com
bcc: bcc@example.com
subject: Meeting Follow-up
priority: high
send_after: 2024-01-15 09:00:00
attachments:
  - report.pdf
  - presentation.pptx
---

Dear Team,

Following up on our meeting yesterday...

Best regards,
Your Name
```

## Sending Drafts

### Send All Drafts
```bash
mailos send --drafts
```

### Preview Before Sending
```bash
mailos send --drafts --dry-run
```

### Filter by Priority
```bash
# Only send high priority drafts
mailos send --drafts --filter="priority:high"
```

### Filter by Recipient
```bash
# Only send to specific domain
mailos send --drafts --filter="to:*@company.com"
```

### Additional Send Options
```bash
# Confirm each draft before sending
mailos send --drafts --confirm

# Keep drafts after sending
mailos send --drafts --delete-after=false

# Log sent emails
mailos send --drafts --log-file="sent.log"
```

## Reading Drafts from IMAP

### List All Drafts
```bash
# Show draft headers (subject, to, date)
mailos draft --list
```

### Read Full Draft Content
```bash
# Show complete draft with body content
mailos draft --read
```

The draft reader will:
- Connect to your email provider
- Automatically detect the Drafts folder
- Display all drafts with proper formatting
- Show flags (Draft, Seen, etc.)
- Format dates in readable format

## Interactive Draft Creation

When using `--interactive`, you'll be prompted for:

1. **To**: Primary recipient email address
2. **CC**: Carbon copy recipients (optional)
3. **Subject**: Email subject line
4. **Body**: Email content (Markdown supported)
5. **Priority**: high/normal/low (default: normal)

Press Enter twice after the body to finish.

## AI-Powered Draft Generation

```bash
# Generate drafts from natural language
mailos draft --ai "create 3 follow-up emails for recent client meetings"

# Specify number of drafts
mailos draft --ai "thank you email" --count 5
```

## Template-Based Drafts

```bash
# Use a template (coming soon)
mailos draft --template=follow-up --data=contacts.csv
```

## Draft Organization

```
draft-emails/
├── 001-meeting-follow-up-2024-01-15-143022.md
├── 002-project-update-2024-01-15-143045.md
├── 003-thank-you-2024-01-15-143102.md
├── sent/                 # Successfully sent (if --delete-after=false)
│   └── 001-meeting-follow-up-2024-01-15-143022.md
└── failed/               # Failed to send
    └── 003-thank-you-2024-01-15-143102.md
```

## IMAP Folder Detection

The system automatically detects your provider's Drafts folder:

| Provider | Common Folder Names |
|----------|-------------------|
| Gmail | [Gmail]/Drafts |
| Outlook | Drafts |
| Yahoo | Draft |
| Fastmail | INBOX.Drafts |
| Generic | Drafts, Draft |

## Error Handling

- If IMAP save fails, local draft is still created
- Warning displayed but operation continues
- Drafts can still be sent using local files

## Examples

### Weekly Newsletter Draft
```bash
# Create draft interactively
mailos draft --interactive

# Enter details:
To: team@company.com
Subject: Weekly Update - Week 3
Body: 
## This Week's Highlights
- Completed project milestone
- New team member onboarded
- Client meeting scheduled

# Draft saved locally and to email account
```

### Batch Follow-up Emails
```bash
# Create multiple drafts
mailos draft --interactive
# Answer 'y' when asked "Create another draft?"

# Review all drafts
ls draft-emails/

# Send with confirmation
mailos send --drafts --confirm
```

### Priority-Based Sending
```bash
# Create drafts with different priorities
mailos draft --interactive  # Set priority: high
mailos draft --interactive  # Set priority: normal

# Send only high priority
mailos send --drafts --filter="priority:high"

# Later, send normal priority
mailos send --drafts --filter="priority:normal"
```

## Troubleshooting

### Draft Not Appearing in Email Client
- Check your email client's Drafts folder
- Refresh/sync your email client
- Verify IMAP is enabled in your email settings

### IMAP Connection Failed
- Ensure app-specific password is configured
- Check IMAP is enabled for your account
- Verify network connectivity

### Drafts Folder Not Found
- System will attempt to create Drafts folder
- Falls back to INBOX if creation fails
- Local draft is always created successfully

## Best Practices

1. **Use Priority Levels**: Organize drafts by importance
2. **Preview Before Sending**: Always use `--dry-run` first
3. **Keep Backups**: Use `--delete-after=false` for important drafts
4. **Batch Similar Emails**: Create related drafts together
5. **Review IMAP Status**: Check for sync confirmation messages

## Related Commands

- `mailos send` - Send individual emails
- `mailos template` - Manage email templates
- `mailos interactive` - Interactive email mode
- `mailos read` - Read emails from inbox

## See Also

- [Send Command](send.md) - Detailed sending options
- [Template Management](template.md) - Create reusable templates
- [Interactive Mode](interactive.md) - Full interactive interface