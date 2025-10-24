// frontend.go - User interface and configuration setup functions
// This file contains all the interactive configuration UI logic,
// including menus, prompts, and setup wizards.

package mailos

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/manifoldco/promptui"
)




// editConfiguration allows editing an existing configuration
func editConfiguration(config *Config, configPath string, isGlobal bool) error {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if isGlobal {
		fmt.Println("                 EDIT GLOBAL CONFIGURATION              ")
	} else {
		fmt.Println("                 EDIT LOCAL CONFIGURATION               ")
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("\n📝 Press Enter to keep current value, or type new value:")
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

	// AI CLI configuration has been moved to setup.go
	fmt.Printf("AI CLI Provider [%s]: (use 'mailos setup' to change)\n", GetAICLIName(config.DefaultAICLI))

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

// changeFromAddress allows user to change just the from address
func changeFromAddress(config *Config, configPath string) error {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("                 CHANGE FROM ADDRESS                   ")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	if config.FromEmail != "" {
		fmt.Printf("\nCurrent from address: %s\n", config.FromEmail)
	} else {
		fmt.Printf("\nCurrent from address: %s (using account email)\n", config.Email)
	}
	
	fmt.Print("\nEnter new from address (press Enter to use account email): ")
	newFrom, _ := reader.ReadString('\n')
	newFrom = strings.TrimSpace(newFrom)
	
	if newFrom != "" && !isValidEmail(newFrom) {
		fmt.Println("Invalid email address. Keeping current setting.")
		return nil
	}
	
	config.FromEmail = newFrom
	
	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %v", err)
	}
	
	if newFrom == "" {
		fmt.Printf("\n✓ From address set to account email: %s\n", config.Email)
	} else {
		fmt.Printf("\n✓ From address updated to: %s\n", newFrom)
	}
	fmt.Println("📧 Configuration saved successfully!")
	
	return nil
}

// changeDisplayName allows user to change just the display name
func changeDisplayName(config *Config, configPath string) error {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("                 CHANGE DISPLAY NAME                  ")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	if config.FromName != "" {
		fmt.Printf("\nCurrent display name: %s\n", config.FromName)
	} else {
		fmt.Println("\nNo display name currently set.")
	}
	
	fmt.Print("\nEnter new display name (press Enter to remove): ")
	newName, _ := reader.ReadString('\n')
	newName = strings.TrimSpace(newName)
	
	config.FromName = newName
	
	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %v", err)
	}
	
	if newName == "" {
		fmt.Println("\n✓ Display name removed.")
	} else {
		fmt.Printf("\n✓ Display name updated to: %s\n", newName)
	}
	fmt.Println("📧 Configuration saved successfully!")
	
	return nil
}

// changeEmailProvider allows user to change the email provider
func changeEmailProvider(config *Config, configPath string) error {
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("                 CHANGE EMAIL PROVIDER                ")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	fmt.Printf("\nCurrent provider: %s\n", GetProviderName(config.Provider))
	fmt.Println("\n⚠️  Warning: Changing the provider will require re-entering your app password.")
	
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nContinue? (y/n): ")
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	
	if response != "y" && response != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}
	
	// Run setup to reconfigure completely
	return Setup()
}

// changeAICLI redirects to setup for AI CLI configuration
func changeAICLI(config *Config, configPath string) error {
	return fmt.Errorf("AI CLI configuration requires interactive setup. Run 'mailos setup' to configure AI CLI provider")
}

// changeFromAddressLocal allows user to change just the from address for local config
func changeFromAddressLocal(config *Config, configPath string) error {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("              CHANGE FROM ADDRESS (LOCAL)              ")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	if config.FromEmail != "" {
		fmt.Printf("\nCurrent from address: %s\n", config.FromEmail)
	} else {
		fmt.Printf("\nCurrent from address: %s (using account email)\n", config.Email)
	}
	
	fmt.Print("\nEnter new from address (press Enter to use account email): ")
	newFrom, _ := reader.ReadString('\n')
	newFrom = strings.TrimSpace(newFrom)
	
	if newFrom != "" && !isValidEmail(newFrom) {
		fmt.Println("Invalid email address. Keeping current setting.")
		return nil
	}
	
	config.FromEmail = newFrom
	
	if err := saveLocalConfig(config); err != nil {
		return fmt.Errorf("failed to save local configuration: %v", err)
	}
	
	if newFrom == "" {
		fmt.Printf("\n✓ From address set to account email: %s\n", config.Email)
	} else {
		fmt.Printf("\n✓ From address updated to: %s\n", newFrom)
	}
	fmt.Println("📁 Local configuration saved successfully!")
	
	return nil
}

// changeDisplayNameLocal allows user to change just the display name for local config
func changeDisplayNameLocal(config *Config, configPath string) error {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("              CHANGE DISPLAY NAME (LOCAL)             ")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	if config.FromName != "" {
		fmt.Printf("\nCurrent display name: %s\n", config.FromName)
	} else {
		fmt.Println("\nNo display name currently set.")
	}
	
	fmt.Print("\nEnter new display name (press Enter to remove): ")
	newName, _ := reader.ReadString('\n')
	newName = strings.TrimSpace(newName)
	
	config.FromName = newName
	
	if err := saveLocalConfig(config); err != nil {
		return fmt.Errorf("failed to save local configuration: %v", err)
	}
	
	if newName == "" {
		fmt.Println("\n✓ Display name removed.")
	} else {
		fmt.Printf("\n✓ Display name updated to: %s\n", newName)
	}
	fmt.Println("📁 Local configuration saved successfully!")
	
	return nil
}

// changeEmailProviderLocal allows user to change the email provider for local config
func changeEmailProviderLocal(config *Config, configPath string) error {
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("              CHANGE EMAIL PROVIDER (LOCAL)           ")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	fmt.Printf("\nCurrent provider: %s\n", GetProviderName(config.Provider))
	fmt.Println("\n⚠️  Warning: Changing the provider will require re-entering your app password.")
	
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nContinue? (y/n): ")
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	
	if response != "y" && response != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}
	
	// Create new local configuration with setup flow
	opts := ConfigureOptions{IsLocal: true}
	return setupConfigWithOptions(opts, true)
}

// changeAICLILocal allows user to change just the AI CLI setting for local config
func changeAICLILocal(config *Config, configPath string) error {
	return fmt.Errorf("AI CLI configuration requires interactive setup. Run 'mailos setup' to configure AI CLI provider")
}

// showAdvancedLocalOptions displays the advanced local configuration menu
func showAdvancedLocalOptions(localConfig *Config, localConfigPath string) error {
	menuItems := []string{
		"✓ Keep current configuration",
		"🔄 Replace local configuration",
		"🗑️  Remove local configuration",
	}
	
	prompt := promptui.Select{
		Label: "Advanced local configuration options",
		Items: menuItems,
	}
	
	index, _, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("selection cancelled: %v", err)
	}
	
	switch index {
	case 0: // Keep current
		fmt.Println("\n✓ Configuration unchanged.")
		return nil
	case 1: // Replace local
		opts := ConfigureOptions{IsLocal: true}
		return setupConfigWithOptions(opts, true)
	case 2: // Remove local
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("\n⚠️  Are you sure you want to remove the local configuration? (y/n): ")
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
	
	return nil
}

// showAdvancedOptions displays the advanced configuration menu
func showAdvancedOptions(globalConfig *Config, globalConfigPath string, localConfig *Config, localConfigPath string) error {
	var menuItems []string
	if localConfig != nil {
		menuItems = []string{
			"✓ Keep current configuration",
			"🔄 Replace global configuration",
			"📁 Edit local configuration for this folder",
			"🗑️  Remove local configuration",
		}
	} else {
		menuItems = []string{
			"✓ Keep current configuration", 
			"🔄 Replace global configuration",
			"➕ Add local configuration for this folder",
		}
	}
	
	prompt := promptui.Select{
		Label: "Advanced configuration options",
		Items: menuItems,
	}
	
	index, _, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("selection cancelled: %v", err)
	}
	
	switch index {
	case 0: // Keep current
		fmt.Println("\n✓ Configuration unchanged.")
		return nil
	case 1: // Replace global
		return Setup()
	case 2: // Local configuration
		if localConfig != nil {
			// Edit existing local config
			return editConfiguration(localConfig, localConfigPath, false)
		} else {
			// Create new local config
			opts := ConfigureOptions{IsLocal: true}
			return configureLocal(opts)
		}
	case 3: // Remove local (only available when local config exists)
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("\n⚠️  Are you sure you want to remove the local configuration? (y/n): ")
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
	
	return nil
}

// showAccountSwitcherInterface displays the main interface with account switching
func showAccountSwitcherInterface(globalConfig *Config, globalConfigPath string) error {
	accounts := GetAllAccounts(globalConfig)
	
	// Get current active account
	activeAccount := globalConfig.ActiveAccount
	if activeAccount == "" {
		activeAccount = globalConfig.Email
	}
	
	// Check for local config
	localConfigPath := filepath.Join(".email", "config.json")
	localConfig, _ := LoadConfigFromPath(localConfigPath)
	
	// Use account selector to choose account
	selectedEmail, newAccount, err := ShowAccountSelector()
	if err != nil {
		if err.Error() == "cancelled" {
			return nil
		}
		return err
	}
	
	// If a new account was created, add it to the config
	if newAccount != nil {
		if err := AddAccount(globalConfig, *newAccount); err != nil {
			return fmt.Errorf("failed to add new account: %v", err)
		}
		accounts = GetAllAccounts(globalConfig)
	}
	
	// Find the selected account
	var selectedAccount *AccountConfig
	for _, acc := range accounts {
		if acc.Email == selectedEmail {
			selectedAccount = &acc
			break
		}
	}
	
	if selectedAccount == nil {
		return fmt.Errorf("selected account not found")
	}
	
	// Switch to the selected account
	if err := SwitchAccount(globalConfig, selectedEmail); err != nil {
		return fmt.Errorf("failed to switch account: %v", err)
	}
	
	// Display the configuration interface for the selected account
	return showConfigurationForAccount(selectedAccount, globalConfig, globalConfigPath, localConfig, localConfigPath)
}

// showConfigurationForAccount displays the configuration interface for a specific account
func showConfigurationForAccount(account *AccountConfig, globalConfig *Config, globalConfigPath string, localConfig *Config, localConfigPath string) error {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("                  EMAIL CONFIGURATION                   ")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	
	// Show currently active account
	fmt.Println("📧 ACTIVE ACCOUNT")
	fmt.Println("─────────────────")
	fmt.Printf("Email:     %s\n", account.Email)
	if account.FromEmail != "" && account.FromEmail != account.Email {
		fmt.Printf("From:      %s\n", account.FromEmail)
	}
	if account.FromName != "" {
		fmt.Printf("Name:      %s\n", account.FromName)
	}
	fmt.Printf("Provider:  %s\n", GetProviderName(account.Provider))
	if account.Label != "" {
		fmt.Printf("Type:      %s\n", account.Label)
	}
	fmt.Printf("AI CLI:    %s\n", GetAICLIName(globalConfig.DefaultAICLI))
	fmt.Println()
	
	// Show local settings if they exist
	if localConfig != nil {
		fmt.Println("📁 LOCAL SETTINGS (in this folder)")
		fmt.Println("──────────────────────────────────")
		if localConfig.FromEmail != "" {
			fmt.Printf("From:      %s\n", localConfig.FromEmail)
		}
		if localConfig.FromName != "" {
			fmt.Printf("Name:      %s\n", localConfig.FromName)
		}
		if localConfig.DefaultAICLI != "" {
			fmt.Printf("AI CLI:    %s\n", GetAICLIName(localConfig.DefaultAICLI))
		}
		fmt.Println()
	}
	
	// Configuration options menu
	configOptions := []string{
		"📧 Switch account",
		"📧 Change from address (" + getFromAddress(account) + ")",
		"👤 Change display name (" + getAccountDisplayName(account) + ")",
		"📨 Change email provider (" + GetProviderName(account.Provider) + ")",
		"🤖 Change AI CLI (" + GetAICLIName(globalConfig.DefaultAICLI) + ")",
		"⚙️  Advanced options...",
	}
	
	prompt := promptui.Select{
		Label: "What would you like to configure?",
		Items: configOptions,
	}
	
	index, _, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("selection cancelled: %v", err)
	}
	
	switch index {
	case 0: // Switch account
		return showAccountSwitcherInterface(globalConfig, globalConfigPath)
	case 1: // Change from address
		return changeFromAddress(globalConfig, globalConfigPath)
	case 2: // Change display name
		return changeDisplayName(globalConfig, globalConfigPath)
	case 3: // Change email provider
		return changeEmailProvider(globalConfig, globalConfigPath)
	case 4: // Change AI CLI
		return changeAICLI(globalConfig, globalConfigPath)
	case 5: // Advanced options
		return showAdvancedOptions(globalConfig, globalConfigPath, localConfig, localConfigPath)
	}
	
	return nil
}

// getFromAddress returns the from address for display, handling empty values
func getFromAddress(account *AccountConfig) string {
	if account.FromEmail != "" {
		return account.FromEmail
	}
	return account.Email
}

// getAccountDisplayName returns the display name for display, handling empty values
func getAccountDisplayName(account *AccountConfig) string {
	if account.FromName != "" {
		return account.FromName
	}
	return "not set"
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default:
		return fmt.Errorf("unsupported platform")
	}

	return exec.Command(cmd, args...).Start()
}