# EmailOS Command Reference

## Available Commands

### Send and Draft Emails

Both ` + "`send`" + ` and ` + "`draft`" + ` commands share the same parameters for email composition:

**Core Parameters (work with both commands):**
- ` + "`-t, --to`" + `: Recipient email addresses (can be used multiple times)
- ` + "`-s, --subject`" + `: Email subject
- ` + "`-b, --body`" + `: Email body (supports Markdown)
- ` + "`-c, --cc`" + `: CC recipients (can be used multiple times)
- ` + "`-B, --bcc`" + `: BCC recipients (can be used multiple times)
- ` + "`-f, --file`" + `: Read body from file
- ` + "`-a, --attach`" + `: Attach files (can be used multiple times)
- ` + "`-P, --plain`" + `: Send as plain text (no HTML conversion)
- ` + "`-S, --no-signature`" + `: Don't include signature

#### Send Email (sends immediately)
` + "```bash" + `
mailos send -t <recipient> -s <subject> -b <body> [-c <cc>] [-B <bcc>] [-f <file>] [-a <attachment>]
mailos send --drafts                      # Send all draft emails

# Examples:
mailos send -t user@example.com -s "Hello" -b "This is a test email"
mailos send -t alice@example.com -t bob@example.com -s "Team Update" -b "Meeting at 3pm"
mailos send -t recipient@example.com -s "Report" -f report.md -a data.xlsx
mailos send --drafts                      # Send all drafts
` + "```" + `

#### Draft Email (saves to drafts folder)
` + "```bash" + `
mailos draft [-t <recipient>] [-s <subject>] [-b <body>] [-c <cc>] [-B <bcc>] [-f <file>] [-a <attachment>]
mailos draft --list                      # List drafts from IMAP with UIDs
mailos draft --read                      # Read draft content from IMAP
mailos draft --edit-uid <UID> -b "Updated body"  # Edit existing draft by UID

# Examples:
mailos draft                              # Create draft interactively
mailos draft -t user@example.com -s "Meeting" -b "Let's meet at 3pm"  # Same as send, but saves as draft
mailos draft -t team@example.com -s "Report" -f report.md -a data.xlsx  # Draft with attachment
mailos draft --interactive               # Create multiple drafts

# Editing drafts:
mailos draft --list                      # Shows UIDs in output
mailos draft --edit-uid 12345 -s "Updated Subject" -b "New body content"  # Update draft with UID 12345

# The draft command saves emails to:
# 1. Local .email/draft-emails/ folder (as markdown files)
# 2. Your email account's IMAP Drafts folder (with UID tracking)
` + "```" + `

**Key Difference:** 
- ` + "`send`" + ` immediately sends the email to recipients
- ` + "`draft`" + ` saves the email to drafts folder for later sending (use ` + "`mailos send --drafts`" + ` to send all drafts)
- ` + "`draft --edit-uid <UID>`" + ` updates an existing draft (replacing it with new content)

**Important for editing drafts:**
- When creating a new draft, the command outputs the UID (e.g., "Saved draft to email account's Drafts folder (UID: 12345)")
- Use ` + "`mailos draft --list`" + ` to see all drafts with their UIDs
- To edit a draft, use ` + "`mailos draft --edit-uid <UID>`" + ` with the new content
- When editing, the old draft is deleted and replaced with the updated version (new UID is assigned)

### Read Emails
` + "```bash" + `
mailos read [--unread] [--from <sender>] [--days <n>] [-n <limit>]

# Examples:
mailos read                          # Read last 10 emails
mailos read --unread                 # Read only unread emails
mailos read --from sender@example.com # Read from specific sender
mailos read --days 7                 # Read emails from last 7 days
mailos read -n 20                    # Read last 20 emails
` + "```" + `

### Mark Emails as Read
` + "```bash" + `
mailos mark-read --ids <comma-separated-ids>

# Example:
mailos mark-read --ids 1,2,3
` + "```" + `

### Delete Emails
` + "```bash" + `
mailos delete --ids <ids> --confirm
mailos delete --from <sender> --confirm
mailos delete --subject <subject> --confirm

# Examples:
mailos delete --ids 1,2,3 --confirm
mailos delete --from spam@example.com --confirm
` + "```" + `

### Find Unsubscribe Links
` + "```bash" + `
mailos unsubscribe [--from <sender>] [--open]

# Examples:
mailos unsubscribe --from newsletter@example.com
mailos unsubscribe --from newsletter@example.com --open  # Opens link in browser
` + "```" + `

### Show Configuration
` + "```bash" + `
mailos info  # Display current email configuration
` + "```" + `

### Setup/Reconfigure
` + "```bash" + `
mailos setup      # Run interactive configuration wizard
mailos configure  # Manage existing configuration (interactive menu)

# Direct configuration updates (non-interactive):
mailos configure --name "New Display Name"     # Update display name
mailos configure --from "newemail@example.com" # Update from email
mailos configure --ai "claude-code"            # Update AI provider
mailos configure --local                       # Create/modify local config

# Examples:
mailos configure --name "John Doe"             # Change display name to John Doe
mailos configure --local --name "Project Bot"  # Set local display name
` + "```" + `

### Important Notes for Configuration Changes:
- When the user asks to change their name, use: ` + "`mailos configure --name \"Their Name\"`" + `
- When the user asks to change display name locally, use: ` + "`mailos configure --local --name \"Their Name\"`" + `
- The configure command accepts flags: --name, --from, --email, --provider, --ai
- Use --local flag to modify project-specific configuration (.email/)
- Without --local flag, it modifies global configuration (~/.email/)

## Email Body Formatting

All email bodies support Markdown formatting:
- **Bold**: ` + "`**text**`" + `
- *Italic*: ` + "`*text*`" + `
- Headers: ` + "`# H1`, `## H2`, `### H3`" + `
- Links: ` + "`[text](https://example.com)`" + `
- Code blocks: ` + "` ```code``` `" + `
- Lists: ` + "`- item` or `* item`" + `

## Command Options Reference

### Send Command Options:
- ` + "`-t, --to`" + `: Recipient email addresses (can be used multiple times)
- ` + "`-s, --subject`" + `: Email subject
- ` + "`-b, --body`" + `: Email body (supports Markdown)
- ` + "`-c, --cc`" + `: CC recipients (can be used multiple times)
- ` + "`-B, --bcc`" + `: BCC recipients (can be used multiple times)
- ` + "`-f, --file`" + `: Read body from file
- ` + "`-a, --attach`" + `: Attach files (can be used multiple times)
- ` + "`-P, --plain`" + `: Send as plain text (no HTML conversion)
- ` + "`-S, --no-signature`" + `: Don't include signature

### Read Command Options:
- ` + "`-n, --number`" + `: Number of emails to read (default: 10)
- ` + "`-u, --unread`" + `: Show only unread emails
- ` + "`--from`" + `: Filter by sender email address
- ` + "`--subject`" + `: Filter by subject (partial match)
- ` + "`--days`" + `: Show emails from last N days
- ` + "`--json`" + `: Output as JSON
- ` + "`--save-markdown`" + `: Save emails as markdown files
- ` + "`--output-dir`" + `: Directory to save markdown files

## Usage Notes

1. The mailos command is available globally after installation
2. Email configuration is stored locally in ~/.email/config.json
3. All commands return appropriate exit codes for error handling
4. Use -f flag to send email content from a file
5. Multiple recipients can be specified with multiple -t flags
6. The read command returns emails in chronological order
7. Email IDs are provided in the read output for use with mark-read and delete commands

## Examples of Common Tasks

### Send a quick email:
` + "```bash" + `
mailos send -t boss@company.com -s "Project Update" -b "The project is on track for Friday delivery."
` + "```" + `

### Read and manage unread emails:
` + "```bash" + `
mailos read --unread
mailos mark-read --ids 1,2,3
` + "```" + `

### Clean up emails from a specific sender:
` + "```bash" + `
mailos read --from newsletter@spam.com
mailos delete --from newsletter@spam.com --confirm
` + "```" + `

### Send an email with attachment:
` + "```bash" + `
mailos send -t client@example.com -s "Report Attached" -b "Please find the report attached." -a report.pdf
` + "```" + `
`