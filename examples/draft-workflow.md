# Email Draft Workflow Examples

## Overview
The draft workflow in mailos allows you to prepare emails in advance, review them, and send them in batch. This is particularly useful for:
- Bulk email campaigns
- Scheduled communications
- Template-based messages
- Review before sending

## Basic Usage

### 1. Create a Single Draft Interactively
```bash
mailos drafts --interactive
```
This will prompt you for:
- To address
- CC (optional)
- Subject
- Body (markdown supported)
- Priority

### 2. Create Multiple Drafts
```bash
mailos drafts --interactive
# Answer 'y' when asked "Create another draft?"
```

### 3. Send All Drafts
```bash
mailos send --drafts
```

### 4. Preview Before Sending (Dry Run)
```bash
mailos send --drafts --dry-run
```

## Advanced Usage

### Filter Drafts by Priority
```bash
# Only send high priority drafts
mailos send --drafts --filter="priority:high"
```

### Filter by Recipient
```bash
# Only send drafts to specific domain
mailos send --drafts --filter="to:*@company.com"
```

### Confirm Before Sending
```bash
mailos send --drafts --confirm
```

### Keep Drafts After Sending
```bash
mailos send --drafts --delete-after=false
```
This moves sent drafts to `draft-emails/sent/` instead of deleting them.

### Log Sent Emails
```bash
mailos send --drafts --log-file="sent-emails.log"
```

## Draft File Format

Draft files are markdown files with YAML frontmatter:

```markdown
---
to: recipient@example.com, another@example.com
cc: cc@example.com
bcc: hidden@example.com
subject: Your Email Subject
priority: high
send_after: 2025-08-10 09:00:00
attachments:
  - /path/to/file1.pdf
  - /path/to/file2.docx
---

# Email Body in Markdown

Dear **Recipient**,

This is the email content with *markdown* formatting.

## Lists are supported
- Item 1
- Item 2
- Item 3

Best regards,
Your Name
```

## Workflow Examples

### Example 1: Weekly Team Updates
```bash
# Create drafts for team updates
mailos drafts --interactive

# Review what will be sent
mailos send --drafts --dry-run

# Send all drafts
mailos send --drafts
```

### Example 2: Event Invitations
```bash
# Create multiple invitation drafts
mailos drafts --interactive

# Send only to internal team first
mailos send --drafts --filter="to:*@ourcompany.com"

# Then send to external participants
mailos send --drafts --filter="to:*@external.com"
```

### Example 3: Follow-up Emails
```bash
# Create follow-up drafts with high priority
# (manually create files or use --interactive)

# Send high priority first
mailos send --drafts --filter="priority:high"

# Send normal priority later
mailos send --drafts --filter="priority:normal"
```

## Directory Structure

```
draft-emails/
├── 001-draft-name.md        # Pending drafts
├── 002-another-draft.md     
├── sent/                    # Successfully sent (if --delete-after=false)
│   └── 001-draft-name.md
└── failed/                  # Failed to send
    └── 003-failed-draft.md
```

## Tips and Best Practices

1. **Use Dry Run First**: Always test with `--dry-run` before sending bulk emails
2. **Organize by Priority**: Use the priority field to control sending order
3. **Template Variables**: You can use placeholders in drafts and replace them programmatically
4. **Scheduled Sending**: Use `send_after` field to delay sending until a specific time
5. **Backup Drafts**: Keep copies of important draft templates
6. **Review Failed Sends**: Check the `failed/` directory for emails that couldn't be sent

## Future Features (Planned)
- AI-powered draft generation: `mailos drafts --ai "create 3 follow-up emails"`
- Template system: `mailos drafts --template=follow-up --data=contacts.csv`
- Bulk generation from CSV/JSON data
- Mail merge with personalization
- Attachment support
- Rate limiting for bulk sends