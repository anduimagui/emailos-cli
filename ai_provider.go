package mailos

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// AI provider command mappings
var aiProviderCommands = map[string]string{
	"claude-code":       "claude",
	"claude-code-yolo":  "claude",
	"openai-codex":      "openai",
	"gemini-cli":        "gemini",
	"opencode":          "opencode",
}

// GetAIProviderCommand returns the command for the configured AI provider
func GetAIProviderCommand(provider string) (string, bool) {
	cmd, exists := aiProviderCommands[provider]
	return cmd, exists
}

// InvokeAIProvider invokes the configured AI provider with the given query
func InvokeAIProvider(query string) error {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Check if AI provider is configured
	if config.DefaultAICLI == "" || config.DefaultAICLI == "none" {
		return fmt.Errorf("no AI CLI provider configured. Run 'mailos setup' or 'mailos configure' to select an AI provider")
	}

	// Get the command for the AI provider
	aiCommand, exists := GetAIProviderCommand(config.DefaultAICLI)
	if !exists {
		return fmt.Errorf("unknown AI provider: %s", config.DefaultAICLI)
	}

	// Check if the AI command exists in PATH
	if _, err := exec.LookPath(aiCommand); err != nil {
		return fmt.Errorf("AI CLI '%s' not found. Please install %s CLI first", aiCommand, GetAICLIName(config.DefaultAICLI))
	}

	// Build the system message with documentation
	systemMessage := buildAISystemMessage()
	
	// Combine system message with user query
	fullQuery := fmt.Sprintf("%s\n\nUser Query: %s", systemMessage, query)

	// Build command based on provider type
	var cmd *exec.Cmd
	if config.DefaultAICLI == "claude-code-yolo" {
		// YOLO mode with dangerous permissions flag
		cmd = exec.Command(aiCommand, "--dangerously-skip-permissions", fullQuery)
	} else {
		// Regular interactive mode
		cmd = exec.Command(aiCommand, fullQuery)
	}
	
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Invoking %s for email management...\n\n", GetAICLIName(config.DefaultAICLI))
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute AI provider: %v", err)
	}

	return nil
}

// buildAISystemMessage builds the system message with email documentation
func buildAISystemMessage() string {
	readme := getEmailOSDocumentation()
	
	systemMessage := `You are an email manager with permission to read, send, and perform various functions on the user's behalf using the mailos CLI.

IMPORTANT: You have full access to the mailos command-line tool to manage emails. Use the commands documented below to fulfill the user's request.

Current Email Configuration:
`
	
	// Add current configuration info
	if config, err := LoadConfig(); err == nil {
		systemMessage += fmt.Sprintf("- Email: %s\n", config.Email)
		systemMessage += fmt.Sprintf("- Provider: %s\n", GetProviderName(config.Provider))
		if config.FromName != "" {
			systemMessage += fmt.Sprintf("- Display Name: %s\n", config.FromName)
		}
	}
	
	systemMessage += fmt.Sprintf("\n%s", readme)
	
	return systemMessage
}

// getEmailOSDocumentation returns the EmailOS documentation
func getEmailOSDocumentation() string {
	return `# EmailOS Command Reference

## Available Commands

### Send Email
` + "```bash" + `
mailos send -t <recipient> -s <subject> -m <message> [-c <cc>] [-b <bcc>] [-f <file>]

# Examples:
mailos send -t user@example.com -s "Hello" -m "This is a test email"
mailos send -t alice@example.com -t bob@example.com -s "Team Update" -m "Meeting at 3pm"
mailos send -t recipient@example.com -s "Report" -f report.md
` + "```" + `

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
mailos configure  # Manage existing configuration
` + "```" + `

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
- ` + "`-m, --body`" + `: Email body (supports Markdown)
- ` + "`-c, --cc`" + `: CC recipients (can be used multiple times)
- ` + "`-b, --bcc`" + `: BCC recipients (can be used multiple times)
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
mailos send -t boss@company.com -s "Project Update" -m "The project is on track for Friday delivery."
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
mailos send -t client@example.com -s "Report Attached" -m "Please find the report attached." -a report.pdf
` + "```" + `
`
}

// CheckAIProviderAvailable checks if an AI provider is configured and available
func CheckAIProviderAvailable() (bool, string) {
	config, err := LoadConfig()
	if err != nil {
		return false, ""
	}
	
	if config.DefaultAICLI == "" || config.DefaultAICLI == "none" {
		return false, ""
	}
	
	aiCommand, exists := GetAIProviderCommand(config.DefaultAICLI)
	if !exists {
		return false, ""
	}
	
	if _, err := exec.LookPath(aiCommand); err != nil {
		return false, ""
	}
	
	return true, config.DefaultAICLI
}

// IsGeneralQuery checks if the arguments represent a general query
func IsGeneralQuery(args []string) bool {
	if len(args) == 0 {
		return false
	}
	
	// Check if first argument is a known command
	knownCommands := []string{
		"setup", "configure", "config", "template", "send", "read",
		"mark-read", "delete", "unsubscribe", "info", "test",
		"report", "open", "provider",
		"--help", "-h", "--version", "-v",
	}
	
	firstArg := strings.ToLower(args[0])
	for _, cmd := range knownCommands {
		if firstArg == cmd {
			return false
		}
	}
	
	// If not a known command, it's a general query
	return true
}