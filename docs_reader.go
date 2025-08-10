package mailos

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// GenerateAIInstructions reads documentation from docs/ directory and generates
// instructions for AI CLI providers
func GenerateAIInstructions() (string, error) {
	// Start with base instructions
	instructions := strings.Builder{}
	instructions.WriteString("You are an email manager with permission to read, send, and perform various functions on the user's behalf using the mailos CLI.\n\n")
	instructions.WriteString("IMPORTANT: You have full access to the mailos command-line tool to manage emails. Use the commands documented below to fulfill the user's request.\n\n")
	
	// Add current configuration
	config, err := LoadConfig()
	if err == nil {
		instructions.WriteString("Current Email Configuration:\n")
		instructions.WriteString(fmt.Sprintf("- Email: %s\n", config.Email))
		if config.FromEmail != "" && config.FromEmail != config.Email {
			instructions.WriteString(fmt.Sprintf("- From Email: %s\n", config.FromEmail))
		}
		instructions.WriteString(fmt.Sprintf("- Provider: %s\n", config.Provider))
		if config.FromName != "" {
			instructions.WriteString(fmt.Sprintf("- Display Name: %s\n", config.FromName))
		}
		instructions.WriteString("\n")
	}
	
	// Add command reference header
	instructions.WriteString("# EmailOS Command Reference\n\n")
	instructions.WriteString("## Available Commands\n\n")
	
	// Read documentation from docs directory
	docsDir := "docs"
	
	// Check if docs directory exists in current directory
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		// Try to find docs directory relative to executable
		exePath, _ := os.Executable()
		docsDir = filepath.Join(filepath.Dir(exePath), "docs")
		
		if _, err := os.Stat(docsDir); os.IsNotExist(err) {
			// Try home directory
			homeDir, _ := os.UserHomeDir()
			docsDir = filepath.Join(homeDir, ".email", "docs")
			
			if _, err := os.Stat(docsDir); os.IsNotExist(err) {
				// Fall back to embedded basic instructions
				instructions.WriteString(generateBasicInstructions())
				return instructions.String(), nil
			}
		}
	}
	
	// Read each documentation file
	docFiles := []string{
		"drafts.md",    // Draft management first (most recent feature)
		"send.md",      // Sending emails
		"read.md",      // Reading emails
		"query.md",     // Query syntax
		"stats.md",     // Statistics
		"configure.md", // Configuration
		"template.md",  // Templates
		"interactive.md", // Interactive mode
	}
	
	for _, filename := range docFiles {
		filepath := filepath.Join(docsDir, filename)
		if content, err := ioutil.ReadFile(filepath); err == nil {
			// Extract command examples and usage from the markdown
			commandSection := extractCommandSection(string(content))
			if commandSection != "" {
				// Add section header based on filename
				sectionName := strings.TrimSuffix(filename, ".md")
				sectionName = strings.Title(sectionName)
				instructions.WriteString(fmt.Sprintf("### %s\n", sectionName))
				instructions.WriteString(commandSection)
				instructions.WriteString("\n")
			}
		}
	}
	
	// Add general notes
	instructions.WriteString("\n## Important Notes\n\n")
	instructions.WriteString("1. The mailos command is available globally after installation\n")
	instructions.WriteString("2. Email configuration is stored locally in ~/.email/config.json\n")
	instructions.WriteString("3. All commands return appropriate exit codes for error handling\n")
	instructions.WriteString("4. Multiple recipients can be specified with multiple -t flags\n")
	instructions.WriteString("5. Email bodies support Markdown formatting\n")
	instructions.WriteString("6. Drafts are saved both locally (draft-emails/) and to IMAP Drafts folder\n")
	instructions.WriteString("7. Use 'mailos send --drafts' to send all saved drafts\n")
	instructions.WriteString("\n")
	
	// Add configuration change notes
	instructions.WriteString("## Configuration Management\n\n")
	instructions.WriteString("- When the user asks to change their name, use: `mailos configure --name \"Their Name\"`\n")
	instructions.WriteString("- When the user asks to change display name locally, use: `mailos configure --local --name \"Their Name\"`\n")
	instructions.WriteString("- The configure command accepts flags: --name, --from, --email, --provider, --ai\n")
	instructions.WriteString("- Use --local flag to modify project-specific configuration (.email/)\n")
	instructions.WriteString("- Without --local flag, it modifies global configuration (~/.email/)\n")
	
	return instructions.String(), nil
}

// extractCommandSection extracts command examples and usage from markdown documentation
func extractCommandSection(content string) string {
	var result strings.Builder
	lines := strings.Split(content, "\n")
	inCodeBlock := false
	inUsageSection := false
	
	for _, line := range lines {
		// Check for usage or examples section
		if strings.Contains(strings.ToLower(line), "## usage") || 
		   strings.Contains(strings.ToLower(line), "## examples") ||
		   strings.Contains(strings.ToLower(line), "## command") {
			inUsageSection = true
			continue
		}
		
		// Stop at next major section
		if inUsageSection && strings.HasPrefix(line, "## ") && 
		   !strings.Contains(strings.ToLower(line), "usage") && 
		   !strings.Contains(strings.ToLower(line), "example") {
			break
		}
		
		// Track code blocks
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			if inUsageSection {
				result.WriteString(line + "\n")
			}
			continue
		}
		
		// Include content from usage sections
		if inUsageSection {
			result.WriteString(line + "\n")
		}
		
		// Also include any line that starts with "mailos" (command examples)
		if !inUsageSection && strings.Contains(line, "mailos ") && !strings.HasPrefix(line, "#") {
			result.WriteString(line + "\n")
		}
	}
	
	return result.String()
}

// generateBasicInstructions provides fallback instructions when docs are not available
func generateBasicInstructions() string {
	var sb strings.Builder
	sb.WriteString(`### Create Email Drafts
` + "```bash" + `
mailos draft [-t <recipient>] [-s <subject>] [-b <body>] [-c <cc>] [-B <bcc>] [-f <file>]
mailos draft --list                      # List drafts from IMAP
mailos draft --read                      # Read draft content from IMAP

# Examples:
mailos draft                              # Create draft interactively
mailos draft -t user@example.com -s "Meeting" -b "Let's meet at 3pm"
mailos draft --interactive               # Create multiple drafts
` + "```" + `

### Send Email
` + "```bash" + `
mailos send -t <recipient> -s <subject> -b <body> [-c <cc>] [-B <bcc>] [-f <file>]
mailos send --drafts                      # Send all draft emails

# Examples:
mailos send -t user@example.com -s "Hello" -b "This is a test email"
mailos send --drafts                      # Send all drafts
mailos send --drafts --dry-run           # Preview drafts before sending
` + "```" + `

### Read Emails
` + "```bash" + `
mailos read [--unread] [--from <sender>] [--days <n>] [-n <limit>]

# Examples:
mailos read                              # Read last 10 emails
mailos read --unread                     # Read only unread emails
mailos read --from sender@example.com    # Read from specific sender
` + "```" + `

### Configure
` + "```bash" + `
mailos configure [--name <name>] [--email <email>] [--provider <provider>]
mailos configure --local                 # Local configuration

# Examples:
mailos configure --name "John Doe"       # Update display name
mailos configure --local --name "Bot"    # Local display name
` + "```" + `

### Other Commands
` + "```bash" + `
mailos info                              # Show configuration
mailos stats                             # Email statistics
mailos mark-read --ids 1,2,3            # Mark emails as read
mailos delete --ids 1,2,3 --confirm     # Delete emails
` + "```" + `
`)
	return sb.String()
}

// SaveAIInstructions saves the generated instructions to EMAILOS.md in the current directory
func SaveAIInstructions() error {
	instructions, err := GenerateAIInstructions()
	if err != nil {
		return fmt.Errorf("failed to generate AI instructions: %v", err)
	}
	
	// Save to EMAILOS.md in current directory
	filename := "EMAILOS.md"
	err = ioutil.WriteFile(filename, []byte(instructions), 0644)
	if err != nil {
		return fmt.Errorf("failed to write EMAILOS.md: %v", err)
	}
	
	fmt.Printf("✓ AI instructions saved to %s\n", filename)
	return nil
}

// UpdateAIInstructionsOnSetup updates or creates EMAILOS.md file after setup
func UpdateAIInstructionsOnSetup() error {
	return SaveAIInstructions()
}