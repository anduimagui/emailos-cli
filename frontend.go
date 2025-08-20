// frontend.go - User interface and configuration setup functions
// This file contains all the interactive configuration UI logic,
// including menus, prompts, and setup wizards.

package mailos

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/charmbracelet/lipgloss"
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
	
	// Check if we have command-line options to apply directly
	if localConfig != nil && (opts.Name != "" || opts.From != "" || opts.AICLI != "" || opts.Provider != "") {
		// Apply command-line options directly to existing config
		if opts.Name != "" {
			localConfig.FromName = opts.Name
			fmt.Printf("✓ Updated display name to: %s\n", opts.Name)
		}
		if opts.From != "" {
			localConfig.FromEmail = opts.From
			fmt.Printf("✓ Updated from email to: %s\n", opts.From)
		}
		if opts.AICLI != "" {
			// Map command-line AI option to internal key
			aiMap := map[string]string{
				"claude-code":        "claude-code",
				"claude-code-accept": "claude-code-accept",
				"claude-code-yolo":   "claude-code-yolo",
				"claude":             "claude-code",
				"claude-accept":      "claude-code-accept",
				"claude-yolo":        "claude-code-yolo",
				"openai":             "openai-codex",
				"openai-codex":       "openai-codex",
				"gemini":             "gemini-cli",
				"gemini-cli":         "gemini-cli",
				"opencode":           "opencode",
				"none":               "none",
			}
			if key, ok := aiMap[strings.ToLower(opts.AICLI)]; ok {
				localConfig.DefaultAICLI = key
				fmt.Printf("✓ Updated AI CLI to: %s\n", GetAICLIName(key))
			} else {
				return fmt.Errorf("invalid AI CLI: %s", opts.AICLI)
			}
		}
		if opts.Provider != "" {
			// Note: changing provider requires re-entering credentials
			fmt.Println("Changing email provider requires full reconfiguration.")
			return setupConfigWithOptions(opts, true)
		}
		
		// Save the updated local configuration
		return saveLocalConfig(localConfig)
	}
	
	if localConfig != nil && opts.Email == "" {
		// Local config already exists and no command-line options provided
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("            LOCAL CONFIGURATION EXISTS                  ")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()
		fmt.Println("📁 LOCAL SETTINGS (in this folder)")
		fmt.Println("──────────────────────────────────")
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
		
		// New improved local menu - start with individual config options
		configOptions := []string{
			"📧 Change from address (" + localConfig.FromEmail + ")",
			"👤 Change display name (" + localConfig.FromName + ")",
			"📨 Change email provider (" + GetProviderName(localConfig.Provider) + ")",
			"🤖 Change AI CLI (" + GetAICLIName(localConfig.DefaultAICLI) + ")",
			"⚙️  Advanced options...",
		}
		
		// Handle empty values gracefully
		if localConfig.FromEmail == "" {
			configOptions[0] = "📧 Set from address (not set)"
		}
		if localConfig.FromName == "" {
			configOptions[1] = "👤 Set display name (not set)"
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
		case 0: // Change from address
			return changeFromAddressLocal(localConfig, localConfigPath)
		case 1: // Change display name
			return changeDisplayNameLocal(localConfig, localConfigPath)
		case 2: // Change email provider
			return changeEmailProviderLocal(localConfig, localConfigPath)
		case 3: // Change AI CLI
			return changeAICLILocal(localConfig, localConfigPath)
		case 4: // Advanced options
			return showAdvancedLocalOptions(localConfig, localConfigPath)
		}
	}
	
	// Create new local configuration
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("              CREATE LOCAL CONFIGURATION                ")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("📁 This will create a .email configuration in the current")
	fmt.Println("   directory for project-specific email settings.")
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
	
	// Check if we have command-line options to apply directly
	if globalConfig != nil && (opts.Name != "" || opts.From != "" || opts.AICLI != "" || opts.Provider != "") {
		// Apply command-line options directly to existing config
		if opts.Name != "" {
			globalConfig.FromName = opts.Name
			fmt.Printf("✓ Updated display name to: %s\n", opts.Name)
		}
		if opts.From != "" {
			globalConfig.FromEmail = opts.From
			fmt.Printf("✓ Updated from email to: %s\n", opts.From)
		}
		if opts.AICLI != "" {
			// Map command-line AI option to internal key
			aiMap := map[string]string{
				"claude-code":        "claude-code",
				"claude-code-accept": "claude-code-accept",
				"claude-code-yolo":   "claude-code-yolo",
				"claude":             "claude-code",
				"claude-accept":      "claude-code-accept",
				"claude-yolo":        "claude-code-yolo",
				"openai":             "openai-codex",
				"openai-codex":       "openai-codex",
				"gemini":             "gemini-cli",
				"gemini-cli":         "gemini-cli",
				"opencode":           "opencode",
				"none":               "none",
			}
		if key, ok := aiMap[strings.ToLower(opts.AICLI)]; ok {
				globalConfig.DefaultAICLI = key
				fmt.Printf("✓ Updated AI CLI to: %s\n", GetAICLIName(key))
			} else {
				return fmt.Errorf("invalid AI CLI: %s", opts.AICLI)
			}
		}
		if opts.Provider != "" {
			// Note: changing provider requires re-entering credentials
			fmt.Println("Changing email provider requires full reconfiguration.")
			return setupConfigWithOptions(opts, false)
		}
		
		// Save the updated global configuration
		return SaveConfig(globalConfig)
	}
	
	if globalConfig != nil && opts.Email == "" {
		// Global config already exists and no command-line options provided
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("                  EMAIL CONFIGURATION                   ")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()
		fmt.Println("📧 GLOBAL SETTINGS")
		fmt.Println("──────────────────")
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
		
		// Check if local config exists
		localConfigPath := filepath.Join(".email", "config.json")
		localConfig, _ := LoadConfigFromPath(localConfigPath)
		
		if localConfig != nil {
			fmt.Println("📁 LOCAL SETTINGS (in this folder)")
			fmt.Println("──────────────────────────────────")
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
		}
		
		// New improved menu - start with individual config options
		configOptions := []string{
			"📧 Change from address (" + globalConfig.FromEmail + ")",
			"👤 Change display name (" + globalConfig.FromName + ")",
			"📨 Change email provider (" + GetProviderName(globalConfig.Provider) + ")",
			"🤖 Change AI CLI (" + GetAICLIName(globalConfig.DefaultAICLI) + ")",
			"⚙️  Advanced options...",
		}
		
		// Handle empty values gracefully
		if globalConfig.FromEmail == "" {
			configOptions[0] = "📧 Set from address (not set)"
		}
		if globalConfig.FromName == "" {
			configOptions[1] = "👤 Set display name (not set)"
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
		case 0: // Change from address
			return changeFromAddress(globalConfig, globalConfigPath)
		case 1: // Change display name
			return changeDisplayName(globalConfig, globalConfigPath)
		case 2: // Change email provider
			return changeEmailProvider(globalConfig, globalConfigPath)
		case 3: // Change AI CLI
			return changeAICLI(globalConfig, globalConfigPath)
		case 4: // Advanced options
			return showAdvancedOptions(globalConfig, globalConfigPath, localConfig, localConfigPath)
		}
	}
	
	// Create new global configuration
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("              CREATE GLOBAL CONFIGURATION               ")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("This will create your global email configuration")
	fmt.Println("in ~/.email/ that is used by default.")
	fmt.Println()
	
	return setupConfigWithOptions(opts, false)
}

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

	// Edit AI CLI provider
	currentAI := GetAICLIName(config.DefaultAICLI)
	fmt.Printf("AI CLI Provider [%s]: ", currentAI)
	fmt.Println("\n  Options: claude-code, claude-code-accept, claude-code-yolo, openai-codex, gemini-cli, opencode, none")
	fmt.Print("  Enter choice: ")
	newAI, _ := reader.ReadString('\n')
	newAI = strings.TrimSpace(strings.ToLower(newAI))
	if newAI != "" {
		// Validate and set the AI CLI choice
		switch newAI {
		case "claude-code", "claude", "claude code":
			config.DefaultAICLI = "claude-code"
		case "claude-code-accept", "claude-accept", "claude accept":
			config.DefaultAICLI = "claude-code-accept"
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
	
	// Check if this is a local config setup and warn if different provider than global
	if isLocal {
		homeDir, _ := os.UserHomeDir()
		globalConfigPath := filepath.Join(homeDir, ".email", "config.json")
		if globalConfig, err := LoadConfigFromPath(globalConfigPath); err == nil && globalConfig != nil {
			if globalConfig.Provider != "" && globalConfig.Provider != selectedKey {
				// Define styles for the warning
				warningStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("196")). // Red
					Bold(true)
				headerStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("226")). // Yellow
					Bold(true).
					Background(lipgloss.Color("196")). // Red background
					Padding(0, 1)
				infoStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("214")) // Orange
				providerStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("39")). // Bright blue
					Bold(true)
				bulletStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("196")) // Red for bullets
				
				fmt.Println()
				fmt.Println(headerStyle.Render(" ⚠️  WARNING: DIFFERENT PROVIDER SELECTED "))
				fmt.Println(warningStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
				fmt.Println()
				fmt.Printf("%s %s\n", 
					infoStyle.Render("Your global configuration uses:"),
					providerStyle.Render(GetProviderName(globalConfig.Provider)))
				fmt.Printf("%s %s\n", 
					infoStyle.Render("You selected for local config: "),
					providerStyle.Render(provider.Name))
				fmt.Println()
				fmt.Println(warningStyle.Render("⚠️  This means you'll need to configure entirely separate"))
				fmt.Println(warningStyle.Render("   credentials for this local folder, including:"))
				fmt.Println()
				fmt.Printf("   %s Email address\n", bulletStyle.Render("•"))
				fmt.Printf("   %s App password\n", bulletStyle.Render("•"))
				fmt.Printf("   %s SMTP/IMAP settings\n", bulletStyle.Render("•"))
				fmt.Println()
				fmt.Println(infoStyle.Render("The local configuration will be saved in:"))
				fmt.Println(providerStyle.Render("  .email/config.json"))
				fmt.Println()
				
				confirmPrompt := promptui.Select{
					Label: warningStyle.Render("Do you want to continue with a different provider?"),
					Items: []string{"Yes, use " + provider.Name, "No, go back to selection"},
				}
				
				confirmIdx, _, err := confirmPrompt.Run()
				if err != nil || confirmIdx == 1 {
					return fmt.Errorf("configuration cancelled")
				}
				fmt.Println()
			}
		}
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
			"claude-code":        "claude-code",
			"claude-code-accept": "claude-code-accept",
			"claude-code-yolo":   "claude-code-yolo",
			"claude":             "claude-code",
			"claude-accept":      "claude-code-accept",
			"claude-yolo":        "claude-code-yolo",
			"openai":             "openai-codex",
			"openai-codex":       "openai-codex",
			"gemini":             "gemini-cli",
			"gemini-cli":         "gemini-cli",
			"opencode":           "opencode",
			"none":               "none",
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
			"Claude Code Accept Edits",
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
			defaultAICLI = "claude-code-accept"
		case 2:
			defaultAICLI = "claude-code-yolo"
		case 3:
			defaultAICLI = "openai-codex"
		case 4:
			defaultAICLI = "gemini-cli"
		case 5:
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
	
	// Load existing config to preserve accounts
	var existingConfig *Config
	if isLocal {
		localConfigPath := filepath.Join(".email", "config.json")
		existingConfig, _ = LoadConfigFromPath(localConfigPath)
	} else {
		existingConfig, _ = LoadConfig()
	}
	
	// Create config, preserving existing accounts
	config := &Config{
		Provider:     selectedKey,
		Email:        email,
		Password:     password,
		FromName:     fromName,
		FromEmail:    fromEmail,
		DefaultAICLI: defaultAICLI,
		LicenseKey:   licenseKey,
		ActiveAccount: email,
	}
	
	// Preserve existing accounts if they exist
	if existingConfig != nil && len(existingConfig.Accounts) > 0 {
		config.Accounts = existingConfig.Accounts
		
		// Check if we need to add/update the current account in the accounts list
		accountFound := false
		for i, acc := range config.Accounts {
			if acc.Email == email {
				// Update existing account
				config.Accounts[i].Provider = selectedKey
				config.Accounts[i].Password = password
				config.Accounts[i].FromName = fromName
				config.Accounts[i].FromEmail = fromEmail
				accountFound = true
				break
			}
		}
		
		// Add current account if it's not already in the list
		if !accountFound && email != "" {
			newAccount := AccountConfig{
				Email:        email,
				Provider:     selectedKey,
				Password:     password,
				FromName:     fromName,
				FromEmail:    fromEmail,
				Label:        "Configured",
			}
			config.Accounts = append(config.Accounts, newAccount)
		}
	} else if email != "" {
		// No existing accounts, create first one
		config.Accounts = []AccountConfig{{
			Email:        email,
			Provider:     selectedKey,
			Password:     password,
			FromName:     fromName,
			FromEmail:    fromEmail,
			Label:        "Primary",
		}}
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

// isValidEmail validates email format
func isValidEmail(email string) bool {
	// Basic email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

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

// changeAICLI allows user to change just the AI CLI setting
func changeAICLI(config *Config, configPath string) error {
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("                   CHANGE AI CLI                      ")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	fmt.Printf("\nCurrent AI CLI: %s\n", GetAICLIName(config.DefaultAICLI))
	
	aiProviders := []string{
		"Claude Code",
		"Claude Code Accept Edits",
		"Claude Code YOLO Mode (skip permissions)",
		"OpenAI Codex",
		"Gemini CLI", 
		"OpenCode",
		"None (Manual only)",
	}
	
	aiPrompt := promptui.Select{
		Label: "Select new AI CLI provider",
		Items: aiProviders,
	}
	
	aiIndex, _, err := aiPrompt.Run()
	if err != nil {
		fmt.Println("Selection cancelled.")
		return nil
	}
	
	var newAICLI string
	switch aiIndex {
	case 0:
		newAICLI = "claude-code"
	case 1:
		newAICLI = "claude-code-accept"
	case 2:
		newAICLI = "claude-code-yolo"
	case 3:
		newAICLI = "openai-codex"
	case 4:
		newAICLI = "gemini-cli"
	case 5:
		newAICLI = "opencode"
	default:
		newAICLI = "none"
	}
	
	config.DefaultAICLI = newAICLI
	
	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %v", err)
	}
	
	fmt.Printf("\n✓ AI CLI updated to: %s\n", GetAICLIName(newAICLI))
	fmt.Println("📧 Configuration saved successfully!")
	
	return nil
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
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("                 CHANGE AI CLI (LOCAL)                ")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	fmt.Printf("\nCurrent AI CLI: %s\n", GetAICLIName(config.DefaultAICLI))
	
	aiProviders := []string{
		"Claude Code",
		"Claude Code Accept Edits",
		"Claude Code YOLO Mode (skip permissions)",
		"OpenAI Codex",
		"Gemini CLI", 
		"OpenCode",
		"None (Manual only)",
	}
	
	aiPrompt := promptui.Select{
		Label: "Select new AI CLI provider",
		Items: aiProviders,
	}
	
	aiIndex, _, err := aiPrompt.Run()
	if err != nil {
		fmt.Println("Selection cancelled.")
		return nil
	}
	
	var newAICLI string
	switch aiIndex {
	case 0:
		newAICLI = "claude-code"
	case 1:
		newAICLI = "claude-code-accept"
	case 2:
		newAICLI = "claude-code-yolo"
	case 3:
		newAICLI = "openai-codex"
	case 4:
		newAICLI = "gemini-cli"
	case 5:
		newAICLI = "opencode"
	default:
		newAICLI = "none"
	}
	
	config.DefaultAICLI = newAICLI
	
	if err := saveLocalConfig(config); err != nil {
		return fmt.Errorf("failed to save local configuration: %v", err)
	}
	
	fmt.Printf("\n✓ AI CLI updated to: %s\n", GetAICLIName(newAICLI))
	fmt.Println("📁 Local configuration saved successfully!")
	
	return nil
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