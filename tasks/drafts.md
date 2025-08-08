# Email Drafts Feature Implementation Checklist

## Overview
Implement a `drafts` command that generates draft emails in markdown format with frontmatter, and a `send --drafts` command that processes and sends them.

## Implementation Tasks

### 1. Drafts Command Implementation
- [ ] Create `drafts.go` file with DraftsCommand function
- [ ] Parse user query to generate draft emails
- [ ] Create `draft-emails/` folder structure
- [ ] Generate markdown files with frontmatter containing:
  - `to:` recipient email(s)
  - `cc:` CC recipients (optional)
  - `bcc:` BCC recipients (optional)
  - `subject:` email subject line
  - `attachments:` list of file paths (optional)
  - `send_after:` scheduled send time (optional)
  - `priority:` high/normal/low (optional)
- [ ] Support AI-powered draft generation using configured AI provider
- [ ] Include template variables for personalization

### 2. Send --drafts Command Enhancement
- [ ] Add `--drafts` flag to existing send command
- [ ] Implement draft folder scanning functionality
- [ ] Parse markdown frontmatter from draft files
- [ ] Process each draft email sequentially
- [ ] Delete draft files after successful sending
- [ ] Move failed drafts to `draft-emails/failed/` folder
- [ ] Log sent emails to `draft-emails/sent.log`

### 3. Draft File Format
```markdown
---
to: recipient@example.com
cc: cc@example.com
subject: Meeting Follow-up
attachments:
  - /path/to/document.pdf
send_after: 2025-08-09 09:00:00
priority: normal
---

# Email Body

Dear [Name],

This is the email content in markdown format.

Best regards,
[Your Name]
```

### 4. Features to Implement
- [ ] Bulk email generation from CSV/JSON data
- [ ] Template system for common email types
- [ ] Draft preview before sending
- [ ] Dry-run mode for testing
- [ ] Progress indicator for batch sending
- [ ] Rate limiting for bulk sends
- [ ] Retry mechanism for failed sends
- [ ] Draft validation before sending

### 5. Error Handling
- [ ] Validate email addresses
- [ ] Check attachment file existence
- [ ] Handle network failures gracefully
- [ ] Provide clear error messages
- [ ] Implement rollback for partial batch sends

### 6. Command Examples
```bash
# Generate drafts from a query
mailos drafts "create 3 follow-up emails for recent meetings"

# Generate drafts from template
mailos drafts --template=follow-up --data=contacts.csv

# Send all drafts
mailos send --drafts

# Send specific drafts
mailos send --drafts --filter="priority:high"

# Dry run to preview what would be sent
mailos send --drafts --dry-run
```

### 7. Testing Requirements
- [ ] Unit tests for drafts command
- [ ] Unit tests for draft parsing
- [ ] Integration tests for full workflow
- [ ] Test with various email providers
- [ ] Test bulk sending limits
- [ ] Test error recovery

### 8. Documentation Updates
- [ ] Update README with new commands
- [ ] Add examples to docs/
- [ ] Create tutorial for bulk email workflow
- [ ] Document template format
- [ ] Add troubleshooting guide

## Dependencies
- Existing email sending infrastructure
- AI provider integration for draft generation
- Markdown parsing library (already in use)
- YAML frontmatter parser

## Timeline
- Phase 1: Basic prep and send --drafts (Week 1)
- Phase 2: AI integration and templates (Week 2)
- Phase 3: Advanced features and testing (Week 3)
- Phase 4: Documentation and polish (Week 4)