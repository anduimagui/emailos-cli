# Reply Command

Reply to emails while preserving thread context and conversation history.

## Usage

```bash
mailos reply [email_number] [flags]
```

## Description

The `reply` command allows you to respond to specific emails while maintaining proper email threading. Unlike the `send` command which creates new emails, `reply` ensures your response becomes part of the existing conversation thread.

## Email Number Reference

The `email_number` parameter refers to the position of emails in your most recent search results:

- `mailos reply 1` = Reply to the **first email** in search results
- `mailos reply 2` = Reply to the **second email** in search results  
- `mailos reply 3` = Reply to the **third email** in search results

**Example:**
```bash
# First, search for emails
mailos search --number 5

# Results show:
# 1. From: john@example.com Subject: Project Update
# 2. From: sarah@company.com Subject: Meeting Tomorrow  
# 3. From: boss@work.com Subject: Quarterly Review

# Reply to the first email (Project Update)
mailos reply 1 --body "Thanks for the update!"

# Reply to the second email (Meeting Tomorrow)
mailos reply 2 --body "I'll be there at 10am"
```

## Threading Behavior

The reply command automatically:

- **Preserves conversation threads** by setting proper email headers (`In-Reply-To`, `References`)
- **Maintains subject line** with "Re:" prefix if not already present
- **Sets correct recipients** (reply to sender, or reply-all to all recipients)
- **Includes original message context** in interactive mode

## Flags

### Basic Options
- `--body string` - Reply body text
- `--subject string` - Override reply subject (keeps "Re:" prefix)
- `-f, --file string` - Read body content from file
- `-i, --interactive` - Force interactive composition mode

### Recipient Options  
- `--all` - Reply to all recipients (reply-all)
- `--to strings` - Override recipients
- `--cc strings` - Add CC recipients
- `--bcc strings` - Add BCC recipients

### Draft Options
- `--draft` - Save as draft instead of sending immediately

## Examples

### Basic Reply
```bash
# Interactive reply (opens composition)
mailos reply 2

# Quick reply with message
mailos reply 1 --body "Thanks for your email!"

# Reply with custom subject
mailos reply 3 --subject "Re: Updated timeline" --body "Looks good"
```

### Reply All
```bash
# Reply to all recipients
mailos reply 1 --all --body "Thanks everyone for the feedback"

# Reply all with additional CC
mailos reply 2 --all --cc team@company.com --body "Looping in the team"
```

### File-based Reply
```bash
# Read reply content from file
mailos reply 1 --file response.txt

# Combine file content with custom recipients
mailos reply 2 --file template.txt --cc manager@company.com
```

### Draft Mode
```bash
# Save reply as draft for later editing
mailos reply 1 --body "Draft response" --draft

# Create draft reply-all
mailos reply 2 --all --body "Team update" --draft
```

### Advanced Examples
```bash
# Override all recipients  
mailos reply 1 --to different@email.com --body "Forwarding to you"

# Reply with BCC to keep manager informed
mailos reply 3 --body "Will handle this" --bcc manager@company.com

# Interactive reply all with custom subject
mailos reply 1 --all --subject "Re: Revised proposal" --interactive
```

## Interactive Mode

When no `--body` or `--file` is specified, or when `--interactive` is used, the reply command enters interactive mode:

1. Shows original email context
2. Opens composition interface
3. Automatically includes quoted original message
4. Press Enter twice to finish composition

```bash
mailos reply 1 --interactive
```

Output:
```
📧 Replying to: Project Update
   From: john@example.com
   Date: Oct 18, 2025 at 2:30 PM
   Message-ID: <abc123@example.com>

────────────────────────────────────────────
📝 Compose your reply:
   (Press Enter twice to finish)
────────────────────────────────────────────
Thanks for the update!

[Your message here]


✓ Reply sent successfully!
```

## Threading Details

### Message Headers
The reply command automatically sets these headers for proper threading:

- **In-Reply-To**: `<original-message-id@domain.com>`
- **References**: Chain of Message-IDs in the conversation
- **Subject**: "Re: Original Subject" (if not already prefixed)

### Conversation Chains
For multi-reply conversations:

```
Original Email: <msg1@domain.com>
├── Reply 1: References: <msg1@domain.com>
├── Reply 2: References: <msg1@domain.com> <reply1@domain.com>  
└── Reply 3: References: <msg1@domain.com> <reply1@domain.com> <reply2@domain.com>
```

## Error Handling

Common errors and solutions:

```bash
# Email not found
Error: email #5 not found (available: 1-3)
# Solution: Check available emails with recent search

# No recent search results  
Error: failed to find email #1: no emails found
# Solution: Run mailos search first to populate email list

# Invalid email number
Error: invalid email number: abc
# Solution: Use numeric email position (1, 2, 3, etc.)
```

## Integration with Other Commands

### Typical Workflow
```bash
# 1. Search for emails
mailos search --from important@client.com --number 10

# 2. Read specific email for context  
mailos read 1

# 3. Reply to the email
mailos reply 1 --body "Thanks for reaching out!"

# 4. Check sent emails
mailos sent --number 3
```

### Draft Workflow
```bash
# Create reply draft
mailos reply 2 --body "Initial response" --draft

# Edit the draft later
mailos draft edit 1 --body "Updated response"

# Send the draft
mailos send --drafts
```

## Tips

1. **Search First**: Always run `mailos search` before using `reply` to ensure you're replying to the intended email
2. **Use Read**: Use `mailos read [number]` to view full email content before replying
3. **Draft for Review**: Use `--draft` for important replies that need review
4. **Reply-All Carefully**: Double-check recipients when using `--all` flag
5. **File Templates**: Create reusable response templates with `--file` option

## See Also

- [`mailos search`](search.md) - Find emails to reply to
- [`mailos read`](read.md) - View email content before replying  
- [`mailos send`](send.md) - Send new emails (non-threaded)
- [`mailos draft`](draft.md) - Manage email drafts
- [`mailos sent`](sent.md) - View sent emails