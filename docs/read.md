# EmailOS Read Command Documentation

The `mailos read` command retrieves and displays emails from your inbox with powerful filtering options.

## Basic Usage

```bash
mailos read
```

Displays the 10 most recent emails by default.

## Command-Line Flags

### Display Options

| Flag | Short | Description | Default | Example |
|------|-------|-------------|---------|---------|
| `--number` | `-n` | Number of emails to display | 10 | `mailos read -n 20` |
| `--json` | | Output as JSON format | false | `mailos read --json` |
| `--save-markdown` | | Save emails as markdown files | true | `mailos read --save-markdown=false` |
| `--output-dir` | | Directory for markdown files | emails | `mailos read --output-dir ./inbox` |

### Filter Options

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--unread` | `-u` | Show only unread emails | `mailos read -u` |
| `--from` | | Filter by sender address | `mailos read --from john@example.com` |
| `--to` | | Filter by recipient | `mailos read --to me@example.com` |
| `--subject` | | Filter by subject | `mailos read --subject "meeting"` |
| `--days` | | Show emails from last N days | `mailos read --days 7` |
| `--range` | | Time range preset | `mailos read --range "Today"` |

## Time Range Options

The `--range` flag accepts:
- `"Last hour"` - Emails from the past hour
- `"Today"` - Today's emails only
- `"Yesterday"` - Yesterday's emails
- `"This week"` - Current week
- `"Last week"` - Previous week
- `"This month"` - Current month
- `"Last month"` - Previous month

## Output Formats

### Default Format
Shows emails in a readable format with:
- Email ID and timestamp
- Sender information
- Subject line
- Preview of email body
- Attachment indicators

### JSON Format
Use `--json` flag for machine-readable output:
```json
[
  {
    "id": 12345,
    "from": "sender@example.com",
    "to": ["recipient@example.com"],
    "subject": "Email Subject",
    "date": "2024-01-15T10:30:00Z",
    "body": "Email content...",
    "attachments": []
  }
]
```

### Markdown Files
By default, emails are saved as markdown files in the `emails` directory:
- Filename format: `subject-date.md`
- Includes full email metadata
- Preserves formatting
- Lists attachments

## Examples

### Read recent unread emails
```bash
mailos read --unread -n 20
```

### Read emails from specific sender
```bash
mailos read --from boss@company.com --days 7
```

### Read today's emails as JSON
```bash
mailos read --range "Today" --json
```

### Read without saving markdown files
```bash
mailos read --save-markdown=false
```

### Search by subject in last month
```bash
mailos read --subject "invoice" --range "Last month"
```

### Read emails with custom output directory
```bash
mailos read --output-dir ./important-emails --from client@company.com
```

## Advanced Filtering

Combine multiple filters for precise results:

```bash
mailos read \
  --from manager@company.com \
  --subject "project" \
  --days 14 \
  --unread \
  -n 50
```

## Notes

- Filters are case-insensitive
- Subject filter matches partial strings
- The `--to` flag defaults to your configured from_email if set
- Markdown files are automatically organized by date
- Use `--json` for integration with other tools
- Email body preview is truncated in default display

## Troubleshooting

### No emails appearing
1. Check your configuration: `mailos info`
2. Verify credentials are correct
3. Try without filters first: `mailos read`
4. Check IMAP settings for your provider

### Markdown files not saving
1. Ensure write permissions in output directory
2. Check disk space
3. Use `--output-dir` to specify writable location

### Slow performance
1. Reduce number of emails: use `-n` flag
2. Use specific time ranges
3. Add more filters to reduce dataset