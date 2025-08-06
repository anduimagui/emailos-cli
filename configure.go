package mailos

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/manifoldco/promptui"
	"golang.org/x/term"
)

// ConfigureOptions holds command-line options for configuration
type ConfigureOptions struct {
	Email    string
	Provider string
	Name     string
	From     string
	AICLI    string
	IsLocal  bool  // If true, create/modify local config; otherwise global
}

// Configure manages email configuration (defaults to global)
func Configure() error {
	return ConfigureWithOptions(ConfigureOptions{})
}

// ConfigureWithOptions creates or modifies email configuration based on options
func ConfigureWithOptions(opts ConfigureOptions) error {
	if opts.IsLocal {
		return configureLocal(opts)
	} else {
		return configureGlobal(opts)
	}
}

// configureLocal handles local .email configuration
func configureLocal(opts ConfigureOptions) error {
	// Check if local .email already exists
	localConfigPath := filepath.Join(".email", "config.json")
	localConfig, _ := LoadConfigFromPath(localConfigPath)
	
	if localConfig != nil && opts.Email == "" {
		// Local config already exists and no command-line options provided
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("LOCAL CONFIGURATION EXISTS")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()
		fmt.Printf("Email:     %s\n", localConfig.Email)
		if localConfig.FromEmail != "" {
			fmt.Printf("From:      %s\n", localConfig.FromEmail)
		}
		if localConfig.FromName != "" {
			fmt.Printf("Name:      %s\n", localConfig.FromName)
		}
		fmt.Printf("Provider:  %s\n", GetProviderName(localConfig.Provider))
		fmt.Printf("AI CLI:    %s\n", GetAICLIName(localConfig.DefaultAICLI))
		fmt.Println()
		
		prompt := promptui.Select{
			Label: "What would you like to do?",
			Items: []string{
				"Keep current configuration",
				"Edit current configuration",
				"Replace with new configuration",
				"Remove local configuration",
			},
		}
		
		index, _, err := prompt.Run()
		if err != nil {
			return fmt.Errorf("selection cancelled: %v", err)
		}
		
		switch index {
		case 0: // Keep current
			fmt.Println("\n✓ Configuration unchanged.")
			return nil
		case 1: // Edit current
			return editConfiguration(localConfig, localConfigPath, false)
		case 2: // Replace with new
			// Continue with setup below
		case 3: // Remove
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("\nAre you sure you want to remove the local configuration? (y/n): ")
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			
			if response == "y" || response == "yes" {
				if err := os.RemoveAll(".email"); err != nil {
					return fmt.Errorf("failed to remove local configuration: %v", err)
				}
				fmt.Println("✓ Local configuration removed.")
			} else {
				fmt.Println("Cancelled.")
			}
			return nil
		}
	}
	
	// Create new local configuration
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("CREATE LOCAL CONFIGURATION")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("This will create a .email configuration in the current")
	fmt.Println("directory for project-specific email settings.")
	fmt.Println()
	
	return setupConfigWithOptions(opts, true)
}

// configureGlobal handles global ~/.email configuration
func configureGlobal(opts ConfigureOptions) error {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}
	
	globalConfigPath := filepath.Join(homeDir, ".email", "config.json")
	globalConfig, _ := LoadConfigFromPath(globalConfigPath)
	
	if globalConfig != nil && opts.Email == "" {
		// Global config already exists and no command-line options provided
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("GLOBAL CONFIGURATION EXISTS")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()
		fmt.Printf("Email:     %s\n", globalConfig.Email)
		if globalConfig.FromEmail != "" {
			fmt.Printf("From:      %s\n", globalConfig.FromEmail)
		}
		if globalConfig.FromName != "" {
			fmt.Printf("Name:      %s\n", globalConfig.FromName)
		}
		fmt.Printf("Provider:  %s\n", GetProviderName(globalConfig.Provider))
		fmt.Printf("AI CLI:    %s\n", GetAICLIName(globalConfig.DefaultAICLI))
		fmt.Println()
		
		prompt := promptui.Select{
			Label: "What would you like to do?",
			Items: []string{
				"Keep current configuration",
				"Edit current configuration",
				"Replace with new configuration",
			},
		}
		
		index, _, err := prompt.Run()
		if err != nil {
			return fmt.Errorf("selection cancelled: %v", err)
		}
		
		switch index {
		case 0: // Keep current
			fmt.Println("\n✓ Configuration unchanged.")
			return nil
		case 1: // Edit current
			return editConfiguration(globalConfig, globalConfigPath, true)
		case 2: // Replace with new
			// Continue with setup below
		}
	}
	
	// Create new global configuration
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("UPDATE GLOBAL CONFIGURATION")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("This will update your global email configuration")
	fmt.Println("in ~/.email/ that is used by default.")
	fmt.Println()
	
	return setupConfigWithOptions(opts, false)
}

// editConfiguration allows editing an existing configuration
func editConfiguration(config *Config, configPath string, isGlobal bool) error {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("EDIT CONFIGURATION")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("\nPress Enter to keep current value, or type new value:")
	fmt.Println()

	// Edit from email
	fmt.Printf("From Email [%s]: ", config.FromEmail)
	newFrom, _ := reader.ReadString('\n')
	newFrom = strings.TrimSpace(newFrom)
	if newFrom != "" {
		config.FromEmail = newFrom
	}

	// Edit display name
	fmt.Printf("Display Name [%s]: ", config.FromName)
	newName, _ := reader.ReadString('\n')
	newName = strings.TrimSpace(newName)
	if newName != "" {
		config.FromName = newName
	}

	// Edit AI CLI provider
	currentAI := GetAICLIName(config.DefaultAICLI)
	fmt.Printf("AI CLI Provider [%s]: ", currentAI)
	fmt.Println("\n  Options: claude-code, claude-code-yolo, openai-codex, gemini-cli, opencode, none")
	fmt.Print("  Enter choice: ")
	newAI, _ := reader.ReadString('\n')
	newAI = strings.TrimSpace(strings.ToLower(newAI))
	if newAI != "" {
		// Validate and set the AI CLI choice
		switch newAI {
		case "claude-code", "claude", "claude code":
			config.DefaultAICLI = "claude-code"
		case "claude-code-yolo", "claude yolo", "yolo":
			config.DefaultAICLI = "claude-code-yolo"
		case "openai-codex", "openai", "codex":
			config.DefaultAICLI = "openai-codex"
		case "gemini-cli", "gemini", "gemini cli":
			config.DefaultAICLI = "gemini-cli"
		case "opencode", "open code":
			config.DefaultAICLI = "opencode"
		case "none", "manual":
			config.DefaultAICLI = "none"
		default:
			fmt.Printf("  Keeping current: %s\n", currentAI)
		}
	}

	// Ask if they want to change provider/email
	fmt.Print("\nDo you want to change the email account? (y/n): ")
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	
	if response == "y" || response == "yes" {
		// Run setup to reconfigure
		return Setup()
	}

	// Save the updated configuration
	if isGlobal {
		if err := SaveConfig(config); err != nil {
			return fmt.Errorf("failed to save configuration: %v", err)
		}
	} else {
		if err := saveLocalConfig(config); err != nil {
			return fmt.Errorf("failed to save local configuration: %v", err)
		}
	}

	fmt.Println("\n✓ Configuration updated successfully!")
	return nil
}

// setupConfigWithOptions creates configuration with command-line options
func setupConfigWithOptions(opts ConfigureOptions, isLocal bool) error {
	reader := bufio.NewReader(os.Stdin)
	var selectedKey string
	var provider Provider
	
	// Handle provider selection
	if opts.Provider != "" {
		// Map command-line provider to internal key
		providerMap := map[string]string{
			"gmail":    "gmail",
			"outlook":  "outlook",
			"yahoo":    "yahoo",
			"icloud":   "icloud",
			"proton":   "proton",
			"fastmail": "fastmail",
			"custom":   "custom",
		}
		
		if key, ok := providerMap[strings.ToLower(opts.Provider)]; ok {
			selectedKey = key
			provider = Providers[selectedKey]
			fmt.Printf("Using provider: %s\n", provider.Name)
		} else {
			return fmt.Errorf("invalid provider: %s. Valid options: gmail, outlook, yahoo, icloud, proton, fastmail, custom", opts.Provider)
		}
	} else {
		// Interactive provider selection
		providerKeys := GetProviderKeys()
		providerNames := make([]string, len(providerKeys))
		for i, key := range providerKeys {
			providerNames[i] = Providers[key].Name
		}
		
		prompt := promptui.Select{
			Label: "Select your email provider",
			Items: providerNames,
		}
		
		index, _, err := prompt.Run()
		if err != nil {
			return fmt.Errorf("provider selection cancelled: %v", err)
		}
		
		selectedKey = providerKeys[index]
		provider = Providers[selectedKey]
		fmt.Printf("\nYou selected: %s\n", provider.Name)
	}
	
	// Handle email address
	var email string
	if opts.Email != "" {
		email = opts.Email
		if !isValidEmail(email) {
			return fmt.Errorf("invalid email address: %s", email)
		}
		fmt.Printf("Using email: %s\n", email)
	} else {
		// Interactive email input
		for {
			fmt.Print("\nEnter your email address: ")
			emailInput, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read email: %v", err)
			}
			email = strings.TrimSpace(emailInput)
			
			if !isValidEmail(email) {
				fmt.Println("This doesn't look right. Please enter a valid email address.")
				continue
			}
			break
		}
	}
	
	// Handle from email
	var fromEmail string
	if opts.From != "" {
		fromEmail = opts.From
		if !isValidEmail(fromEmail) {
			return fmt.Errorf("invalid from email address: %s", fromEmail)
		}
		fmt.Printf("Using from email: %s\n", fromEmail)
	} else {
		// Check if we have a global config for context
		var globalEmail string
		if isLocal {
			if globalConfig, _ := LoadConfig(); globalConfig != nil {
				globalEmail = globalConfig.Email
				if globalEmail != "" {
					fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
					fmt.Println("FROM EMAIL CONFIGURATION")
					fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
					fmt.Printf("\nYour global account email is: %s\n", globalEmail)
					fmt.Println("\nYou can specify a different 'from' email address that will")
					fmt.Println("appear as the sender, while still using your global account")
					fmt.Println("for authentication and sending.")
				}
			}
		}
		
		fmt.Print("\nEnter your 'from' email (optional, press Enter to use account email): ")
		fromInput, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read from email: %v", err)
		}
		fromEmail = strings.TrimSpace(fromInput)
		if fromEmail != "" && !isValidEmail(fromEmail) {
			fmt.Println("This doesn't look right. Using account email instead.")
			fromEmail = ""
		}
	}

	// Handle display name
	var fromName string
	if opts.Name != "" {
		fromName = opts.Name
		fmt.Printf("Using display name: %s\n", fromName)
	} else {
		fmt.Print("\nEnter your display name (optional, press Enter to skip): ")
		nameInput, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read display name: %v", err)
		}
		fromName = strings.TrimSpace(nameInput)
	}
	
	// Handle AI CLI selection
	var defaultAICLI string
	if opts.AICLI != "" {
		// Map command-line AI option to internal key
		aiMap := map[string]string{
			"claude-code":      "claude-code",
			"claude-code-yolo": "claude-code-yolo",
			"claude":           "claude-code",
			"claude-yolo":      "claude-code-yolo",
			"openai":           "openai-codex",
			"openai-codex":     "openai-codex",
			"gemini":           "gemini-cli",
			"gemini-cli":       "gemini-cli",
			"opencode":         "opencode",
			"none":             "none",
		}
		
		if key, ok := aiMap[strings.ToLower(opts.AICLI)]; ok {
			defaultAICLI = key
			fmt.Printf("Using AI CLI: %s\n", GetAICLIName(defaultAICLI))
		} else {
			return fmt.Errorf("invalid AI CLI: %s. Valid options: claude-code, claude-code-yolo, openai, gemini, opencode, none", opts.AICLI)
		}
	} else {
		// Interactive AI CLI selection
		fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("AI CLI CONFIGURATION")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		
		aiProviders := []string{
			"Claude Code",
			"Claude Code YOLO Mode (skip permissions)",
			"OpenAI Codex",
			"Gemini CLI",
			"OpenCode",
			"None (Manual only)",
		}
		
		aiPrompt := promptui.Select{
			Label: "Select AI CLI provider",
			Items: aiProviders,
		}
		
		aiIndex, _, err := aiPrompt.Run()
		if err != nil {
			// Default to None if selection cancelled
			aiIndex = 5
		}
		
		switch aiIndex {
		case 0:
			defaultAICLI = "claude-code"
		case 1:
			defaultAICLI = "claude-code-yolo"
		case 2:
			defaultAICLI = "openai-codex"
		case 3:
			defaultAICLI = "gemini-cli"
		case 4:
			defaultAICLI = "opencode"
		default:
			defaultAICLI = "none"
		}
	}
	
	// Get app password
	fmt.Printf("\n%s\n", provider.AppPasswordHelp)
	fmt.Printf("Visit: %s\n", provider.AppPasswordURL)
	fmt.Print("\nEnter your app password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %v", err)
	}
	password := string(passwordBytes)
	fmt.Println() // New line after password input
	
	// Check for license in global config if it exists
	var licenseKey string
	if globalConfig, _ := LoadConfig(); globalConfig != nil && globalConfig.LicenseKey != "" {
		licenseKey = globalConfig.LicenseKey
	}
	
	// Create config
	config := &Config{
		Provider:     selectedKey,
		Email:        email,
		Password:     password,
		FromName:     fromName,
		FromEmail:    fromEmail,
		DefaultAICLI: defaultAICLI,
		LicenseKey:   licenseKey,
	}
	
	// Save configuration
	if isLocal {
		if err := saveLocalConfig(config); err != nil {
			return fmt.Errorf("failed to save local configuration: %v", err)
		}
		
		// Copy EMAILOS.md to current directory
		if err := copyReadmeToCurrentDir(); err != nil {
			fmt.Printf("Warning: failed to copy EmailOS README: %v\n", err)
		}
		
		fmt.Println("\n✓ Local configuration saved to .email/config.json")
		fmt.Println()
		fmt.Println("This configuration will be used when running mailos from")
		fmt.Println("this directory or its subdirectories.")
		fmt.Println()
		fmt.Println("The EMAILOS.md file has been added for LLM integration.")
	} else {
		if err := SaveConfig(config); err != nil {
			return fmt.Errorf("failed to save global configuration: %v", err)
		}
		
		fmt.Println("\n✓ Global configuration saved to ~/.email/config.json")
		fmt.Println()
		fmt.Println("This configuration will be used by default in all directories.")
		fmt.Println()
		fmt.Println("Tip: Use 'mailos configure --local' to create project-specific configurations.")
	}
	
	return nil
}

// saveLocalConfig saves configuration to local .email directory
func saveLocalConfig(config *Config) error {
	configDir := ".email"
	configPath := filepath.Join(configDir, "config.json")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	// Save configuration
	return SaveConfigToPath(config, configPath)
}

// GetProviderName returns the display name for a provider key
func GetProviderName(key string) string {
	if provider, exists := Providers[key]; exists {
		return provider.Name
	}
	return key
}

// GetAICLIName returns the display name for an AI CLI key
func GetAICLIName(key string) string {
	switch key {
	case "claude-code":
		return "Claude Code"
	case "claude-code-yolo":
		return "Claude Code YOLO Mode"
	case "openai-codex":
		return "OpenAI Codex"
	case "gemini-cli":
		return "Gemini CLI"
	case "opencode":
		return "OpenCode"
	case "none", "":
		return "None (Manual only)"
	default:
		return key
	}
}

// isValidEmail validates email format
func isValidEmail(email string) bool {
	// Basic email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// copyReadmeToCurrentDir creates the EMAILOS.md file in current directory
func copyReadmeToCurrentDir() error {
	// Create the EmailOS documentation content
	content := `# EmailOS Command Reference

This file provides command references for EmailOS (mailos) to enable LLM integration.

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

### Configure Email Settings
` + "```bash" + `
# Global configuration (default)
mailos configure [--email <email>] [--provider <provider>] [--name <name>] [--ai <ai>]

# Local configuration (project-specific)
mailos configure --local [--email <email>] [--provider <provider>] [--name <name>] [--ai <ai>]

# Examples:
mailos configure                     # Interactive global configuration
mailos configure --local             # Interactive local configuration
mailos configure --email user@gmail.com --provider gmail
mailos configure --local --email user@outlook.com --provider outlook --name "John Doe"
mailos configure --ai claude-code    # Set AI CLI provider globally

# Providers: gmail, outlook, yahoo, icloud, proton, fastmail, custom
# AI Options: claude-code, claude-code-yolo, openai, gemini, opencode, none
` + "```" + `

### Mark Emails as Read
` + "```bash" + `
mailos mark-read --ids <comma-separated-ids>

# Example:
mailos mark-read --ids 1,2,3
` + "```" + `

### Show Configuration
` + "```bash" + `
mailos info  # Display current email configuration (shows local or global)
` + "```" + `

### Setup/Reconfigure
` + "```bash" + `
mailos setup  # Run initial setup wizard (global configuration)
` + "```" + `

## Configuration Management

EmailOS supports both global and local configuration:

- **Global**: Stored in ` + "`~/.email/config.json`" + `, used by default
- **Local**: Stored in ` + "`.email/config.json`" + ` in current directory, overrides global

Use ` + "`mailos configure --local`" + ` to create a local configuration for project-specific settings.

## Email Body Formatting

All email bodies support Markdown formatting:
- **Bold**: ` + "`**text**`" + `
- *Italic*: ` + "`*text*`" + `
- Headers: ` + "`# H1`, `## H2`, `### H3`" + `
- Links: ` + "`[text](https://example.com)`" + `
- Code blocks: ` + "` ```code``` `" + `
- Lists: ` + "`- item` or `* item`" + `

## Notes for LLM Usage

1. The ` + "`mailos`" + ` command is available globally after installation
2. Local configuration (` + "`.email/`" + `) overrides global (` + "`~/.email/`" + `)
3. All commands return appropriate exit codes for error handling
4. Use ` + "`-f`" + ` flag to send email content from a file
5. Multiple recipients can be specified with multiple ` + "`-t`" + ` flags
6. The read command returns emails in chronological order
7. Email IDs are provided in the read output for use with mark-read

## Security Notes

- Credentials are stored locally in ` + "`.email/`" + ` or ` + "`~/.email/`" + `
- Uses app-specific passwords, not main account passwords
- Configuration files have restricted permissions (600)
`

	// Write to EMAILOS.md in current directory
	return os.WriteFile("EMAILOS.md", []byte(content), 0644)
}