# EmailOS Send Command Documentation

The `mailos send` command sends emails with support for markdown formatting, attachments, and signatures.

## Basic Usage

```bash
mailos send --to recipient@example.com --subject "Subject" --body "Message"
```

## Command-Line Flags

### Required Flags

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--to` | `-t` | Recipient email(s) | `--to john@example.com,jane@example.com` |
| `--subject` | `-s` | Email subject | `--subject "Meeting Tomorrow"` |

### Message Content

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--body` | `-b` | Email body (Markdown supported) | `--body "Hello **world**"` |
| `--message` | `-m` | Alias for --body | `--message "Hello"` |
| `--file` | `-f` | Read body from file | `--file ./email.md` |

### Recipients

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--to` | `-t` | To recipients (comma-separated) | `--to user1@example.com,user2@example.com` |
| `--cc` | `-c` | CC recipients | `--cc manager@example.com` |
| `--bcc` | `-B` | BCC recipients | `--bcc archive@example.com` |

### Format & Attachments

| Flag | Short | Description | Default | Example |
|------|-------|-------------|---------|---------|
| `--plain` | `-P` | Send as plain text only | false | `--plain` |
| `--attach` | `-a` | File attachments | | `--attach file1.pdf,file2.docx` |
| `--no-signature` | `-S` | Omit signature | false | `--no-signature` |
| `--signature` | | Custom signature text | | `--signature "Best regards,\nJohn"` |

## Markdown Support

By default, markdown in the message body is converted to HTML:

### Supported Markdown

- **Bold**: `**text**` or `__text__`
- *Italic*: `*text*` or `_text_`
- `Code`: `` `code` ``
- Links: `[text](url)`
- Lists: `- item` or `1. item`
- Headers: `# Header`
- Code blocks: ``` ```language ```

### Example with Markdown

```bash
mailos send \
  --to team@example.com \
  --subject "Project Update" \
  --body "## Status Update

**Completed:**
- Feature A
- Feature B

*In Progress:*
- Feature C

See [documentation](https://example.com)"
```

## Input Methods

### 1. Command Line
```bash
mailos send --to user@example.com --subject "Test" --body "Message"
```

### 2. From File
```bash
mailos send --to user@example.com --subject "Report" --file ./report.md
```

### 3. Interactive (stdin)
```bash
mailos send --to user@example.com --subject "Notes"
# Then type your message and press Ctrl+D
```

### 4. Piped Input
```bash
echo "Automated message" | mailos send --to admin@example.com --subject "Alert"
```

## Signature Management

### Default Signature
Automatically appends:
```
--
Your Name
your.email@example.com
```

### Custom Signature
```bash
mailos send \
  --to client@example.com \
  --subject "Proposal" \
  --body "Please find attached..." \
  --signature "Best regards,\nJohn Smith\nSales Manager\nCompany Inc."
```

### No Signature
```bash
mailos send --to user@example.com --subject "Test" --body "Message" --no-signature
```

## Multiple Recipients

### Using Comma Separation
```bash
mailos send \
  --to "user1@example.com,user2@example.com,user3@example.com" \
  --cc "manager@example.com" \
  --bcc "archive@example.com" \
  --subject "Team Update" \
  --body "Meeting at 3pm"
```

### Using Multiple Flags
```bash
mailos send \
  --to user1@example.com \
  --to user2@example.com \
  --cc manager@example.com \
  --subject "Notice" \
  --body "Important information"
```

## Attachments

### Single Attachment
```bash
mailos send \
  --to recipient@example.com \
  --subject "Report" \
  --body "Please see attached" \
  --attach report.pdf
```

### Multiple Attachments
```bash
mailos send \
  --to team@example.com \
  --subject "Documents" \
  --body "All requested files attached" \
  --attach "file1.pdf,file2.docx,data.xlsx"
```

## HTML Templates

If you have configured an HTML template using `mailos template`, it will automatically be applied to your emails unless you use `--plain`.

## Examples

### Simple Text Email
```bash
mailos send \
  --to friend@example.com \
  --subject "Hello" \
  --body "How are you?" \
  --plain
```

### Professional Email with Markdown
```bash
mailos send \
  --to client@company.com \
  --subject "Project Proposal" \
  --body "Dear Client,

**Project Overview**

We are pleased to present our proposal for your consideration.

## Deliverables
1. Design phase
2. Implementation
3. Testing
4. Deployment

Please review and let us know if you have questions.

Best regards"
```

### Email from Template File
```bash
# Create template file
cat > email_template.md << EOF
Dear {{Name}},

Thank you for your interest in our services.

**Next Steps:**
- Schedule a call
- Review documentation
- Sign agreement

Best regards,
Sales Team
EOF

# Send using template
mailos send \
  --to prospect@example.com \
  --subject "Follow-up" \
  --file email_template.md
```

### Batch Email Script
```bash
#!/bin/bash
for email in $(cat recipients.txt); do
  mailos send \
    --to "$email" \
    --subject "Newsletter" \
    --file newsletter.md \
    --attach newsletter.pdf
done
```

## Error Handling

Common errors and solutions:

### Authentication Failed
- Verify email credentials: `mailos info`
- Check app-specific password for Gmail/Outlook
- Ensure 2FA is properly configured

### Attachment Not Found
- Use absolute paths for attachments
- Check file permissions
- Verify file exists before sending

### SMTP Connection Failed
- Check internet connection
- Verify SMTP settings for your provider
- Ensure firewall allows SMTP ports

## Notes

- Email body supports UTF-8 encoding
- Maximum attachment size depends on provider (usually 25MB)
- HTML and plain text versions are sent as multipart/alternative
- Sent emails are saved to Sent folder when IMAP is configured
- Use `--plain` to force plain text only
- Markdown tables are supported in HTML output