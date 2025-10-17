package mailos

import (
	"fmt"
)

// DefaultEmailDraftingPrompt is the default system prompt for email drafting
const DefaultEmailDraftingPrompt = `You are an email drafting assistant. Your ONLY job is to write email responses.

**CRITICAL INSTRUCTIONS:**
- Output ONLY the email message text
- Do NOT include greetings like "Here is your response" or "I've drafted this email"
- Do NOT include any explanations, meta-commentary, or framing text
- Do NOT use markdown formatting or code blocks
- Start directly with the email content (e.g., "Dear John," or "Hi Sarah,")
- End with an appropriate sign-off (e.g., "Best regards," or "Thanks,")

**IMPORTANT: Output NOTHING except the email message itself. No preamble, no explanation, just the email.**`

// DefaultEmailManagerPrompt is the default system prompt for email management
const DefaultEmailManagerPrompt = `You are an email manager with permission to read, send, and perform various functions on the user's behalf using the mailos CLI.

IMPORTANT: You have full access to the mailos command-line tool to manage emails. Use the commands documented below to fulfill the user's request.`

// GetEmailDraftingPrompt returns the configured email drafting prompt or the default
func GetEmailDraftingPrompt() string {
	// For now, return the default prompt
	// TODO: Add configuration support for custom prompts in the future
	return DefaultEmailDraftingPrompt
}

// GetEmailManagerPrompt returns the configured email manager prompt or the default
func GetEmailManagerPrompt() string {
	// For now, return the default prompt
	// TODO: Add configuration support for custom prompts in the future
	return DefaultEmailManagerPrompt
}

// BuildEmailManagerSystemMessage builds the complete system message for email management
func BuildEmailManagerSystemMessage() string {
	systemMessage := GetEmailManagerPrompt()

	systemMessage += "\n\nCurrent Email Configuration:\n"

	// Add current configuration info
	if config, err := LoadConfig(); err == nil {
		systemMessage += fmt.Sprintf("- Active Account: %s\n", config.Email)
		systemMessage += fmt.Sprintf("- Provider: %s\n", GetProviderName(config.Provider))
		if config.FromName != "" {
			systemMessage += fmt.Sprintf("- Display Name: %s\n", config.FromName)
		}
		if config.FromEmail != "" && config.FromEmail != config.Email {
			systemMessage += fmt.Sprintf("- Sending As: %s\n", config.FromEmail)
		}

		// List all available accounts
		accounts := GetAllAccounts(config)
		if len(accounts) > 1 {
			systemMessage += "\nAvailable Accounts:\n"
			for _, acc := range accounts {
				label := ""
				if acc.Label != "" {
					label = fmt.Sprintf(" (%s)", acc.Label)
				}
				if acc.Email == config.Email {
					systemMessage += fmt.Sprintf("- %s%s [ACTIVE]\n", acc.Email, label)
				} else {
					systemMessage += fmt.Sprintf("- %s%s\n", acc.Email, label)
				}
			}
		}
	}

	// Add documentation
	systemMessage += "\n" + getEmailOSDocumentation()

	return systemMessage
}

// getEmailOSDocumentation returns the EmailOS documentation
func getEmailOSDocumentation() string {
	return `# EmailOS Command Reference

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
mailos send -t recipient@example.com -s "Subject" -b "Body"
mailos send -t user1@example.com -t user2@example.com -s "Multiple recipients" -b "Hello all"
mailos send -t boss@company.com -c coworker@company.com -s "Meeting notes" -b "Here are the notes..."
mailos send -t friend@example.com -a ~/Documents/photo.jpg -s "Check this out" -b "See attachment"
` + "```" + `

#### Draft Email (saves to IMAP Drafts folder)
` + "```bash" + `
mailos draft -t recipient@example.com -s "Subject" -b "Body"
mailos draft -t user@example.com -s "Draft email" -b "This will be saved as a draft"
` + "```" + `

### Read and List Emails

#### List Emails
` + "```bash" + `
mailos list                    # List recent emails from INBOX
mailos list -f "INBOX.Sent"    # List from specific folder
mailos list -l 50              # List last 50 emails
mailos list -u                # List unread emails only
mailos list --search "meeting" # Search emails containing "meeting"
` + "```" + `

#### Read Email
` + "```bash" + `
mailos read 123                # Read email with UID 123
mailos read -f "INBOX.Sent" 456 # Read from specific folder
` + "```" + `

### Interactive Mode
` + "```bash" + `
mailos interactive            # Start interactive email management
mailos interactive --ai       # Start AI-powered interactive mode
` + "```" + `

### Account Management
` + "```bash" + `
mailos setup                  # Initial setup wizard
mailos account add            # Add additional email account
mailos account list           # List configured accounts
mailos account switch         # Switch active account
` + "```" + `

### Other Commands
` + "```bash" + `
mailos sync                   # Sync emails locally
mailos stats                  # Show email statistics
mailos help                   # Show help information
` + "```" + `

## Tips for Effective Email Management

1. **Use Drafts**: Always draft important emails first to review before sending
2. **Search Effectively**: Use the --search flag to find specific emails quickly
3. **Multiple Accounts**: Configure multiple accounts and switch between them easily
4. **Attachments**: Use relative or absolute paths for file attachments
5. **Interactive Mode**: For complex tasks, use interactive mode with AI assistance

## Configuration

EmailOS can be configured via ` + "`~/.email/config.json`" + `:

` + "```json" + `
{
  "provider": "gmail",
  "email": "your.email@example.com",
  "password": "your-app-password",
  "from_name": "Your Name",
  "default_ai_cli": "ollama",
  "email_drafting_prompt": "Custom prompt for email drafting...",
  "email_manager_prompt": "Custom prompt for email management..."
}
` + "```" + `

For Gmail, use an App Password instead of your regular password.`
}
