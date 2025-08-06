package mailos

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/charmbracelet/lipgloss"
	"github.com/manifoldco/promptui"
	"golang.org/x/term"
)

func Setup() error {
	// Define styles
	logoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")) // Bright blue
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("87")) // Light cyan
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82")) // Green
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow
	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")) // Orange
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")) // White bold
	
	// Display ASCII art
	fmt.Println()
	fmt.Println(logoStyle.Render("    ███╗   ███╗ █████╗ ██╗██╗      ██████╗ ███████╗"))
	fmt.Println(logoStyle.Render("    ████╗ ████║██╔══██╗██║██║     ██╔═══██╗██╔════╝"))
	fmt.Println(logoStyle.Render("    ██╔████╔██║███████║██║██║     ██║   ██║███████╗"))
	fmt.Println(logoStyle.Render("    ██║╚██╔╝██║██╔══██║██║██║     ██║   ██║╚════██║"))
	fmt.Println(logoStyle.Render("    ██║ ╚═╝ ██║██║  ██║██║███████╗╚██████╔╝███████║"))
	fmt.Println(logoStyle.Render("    ╚═╝     ╚═╝╚═╝  ╚═╝╚═╝╚══════╝ ╚═════╝ ╚══════╝"))
	fmt.Println()
	fmt.Println(titleStyle.Render("    Email Client for the Command Line"))
	fmt.Println(titleStyle.Render("    ═══════════════════════════════════"))
	fmt.Println()
	
	// License validation
	fmt.Println(headerStyle.Render("LICENSE VALIDATION"))
	fmt.Println(headerStyle.Render("━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Println("EmailOS requires a valid license key to operate.")
	fmt.Println("If you don't have a license key, please visit:")
	fmt.Printf("https://%s/checkout\n", APP_SITE)
	fmt.Println()
	
	reader := bufio.NewReader(os.Stdin)
	var licenseKey string
	validLicense := false
	
	// Try to load existing license from config
	if existingConfig, err := LoadConfig(); err == nil && existingConfig.LicenseKey != "" {
		fmt.Printf("Found existing license key: %s...\n", existingConfig.LicenseKey[:min(8, len(existingConfig.LicenseKey))])
		fmt.Print("Use existing license? (Y/n): ")
		useExisting, _ := reader.ReadString('\n')
		useExisting = strings.TrimSpace(strings.ToLower(useExisting))
		
		if useExisting != "n" && useExisting != "no" {
			licenseKey = existingConfig.LicenseKey
			fmt.Println("Validating existing license...")
			lm := GetLicenseManager()
			if err := lm.ValidateLicense(licenseKey); err == nil {
				validLicense = true
				fmt.Println(successStyle.Render("✓ License validated successfully!"))
			} else {
				fmt.Println(errorStyle.Render("✗ Invalid license key"))
				fmt.Println("Please enter a new license key.")
			}
		}
	}
	
	// Get new license key if needed
	for !validLicense {
		fmt.Print(promptStyle.Render("\nEnter your license key (or press Ctrl+C to exit): "))
		licenseKey, _ = reader.ReadString('\n')
		licenseKey = strings.TrimSpace(licenseKey)
		
		if licenseKey == "" {
			fmt.Println(errorStyle.Render("\n✗ License key cannot be empty."))
			fmt.Println("\nTo purchase a license, please visit:")
			fmt.Printf(warningStyle.Render("→ https://%s/checkout\n"), APP_SITE)
			fmt.Println("\nYou can:")
			fmt.Println("1. Open the link above in your browser to purchase a license")
			fmt.Println("2. Enter your license key when you have one")
			fmt.Println("3. Press Ctrl+C to exit")
			fmt.Print("\nPress ENTER to open the checkout page, or type your license key: ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			
			if input == "" {
				// User pressed enter, open browser
				fmt.Println("Opening checkout page...")
				openBrowser(fmt.Sprintf("https://%s/checkout", APP_SITE))
				fmt.Println("Once you have your license key, enter it below.")
			} else {
				// User typed something, treat it as a license key
				licenseKey = input
				// Continue to validation below
			}
			
			if licenseKey == "" {
				continue
			}
		}
		
		fmt.Println("\nValidating license key...")
		lm := GetLicenseManager()
		if err := lm.ValidateLicense(licenseKey); err != nil {
			fmt.Println(errorStyle.Render("\n✗ Invalid license key"))
			fmt.Println()
			fmt.Printf(warningStyle.Render("Please visit → https://%s/checkout to purchase a valid license.\n"), APP_SITE)
			fmt.Println("\nYou can:")
			fmt.Println("1. Try entering a different license key")
			fmt.Println("2. Visit the checkout page to purchase a license")
			fmt.Println("3. Press Ctrl+C to exit")
		} else {
			validLicense = true
			fmt.Println(successStyle.Render("✓ License validated successfully!"))
			
			// Show customer info if available
			if cache := lm.GetCachedLicense(); cache != nil && cache.CustomerEmail != "" {
				fmt.Printf("%s %s\n", successStyle.Render("✓ Licensed to:"), cache.CustomerEmail)
			}
		}
	}
	
	fmt.Println()
	
	// Security disclaimer
	fmt.Println(warningStyle.Render("IMPORTANT SECURITY NOTICE:"))
	fmt.Println(warningStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Println("MailOS is a command-line email client that:")
	fmt.Println("• Stores your email configuration ONLY on your local machine (~/.email/)")
	fmt.Println("• Does NOT transmit or store your credentials anywhere else")
	fmt.Println("• Requires you to generate and manage app-specific passwords")
	fmt.Println("• Is NOT responsible for security issues from app password usage")
	fmt.Println()
	fmt.Println("By continuing, you acknowledge that:")
	fmt.Println("• You are responsible for securing your app passwords")
	fmt.Println("• Your credentials are stored locally in plain text")
	fmt.Println("• You should use app-specific passwords, never your main password")
	fmt.Println()
	
	fmt.Print(promptStyle.Render("Press ENTER to confirm you understand and want to continue..."))
	reader.ReadString('\n')
	
	fmt.Println()
	fmt.Println("Let's set up your email client.")
	fmt.Println()

	// Select provider
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

	selectedKey := providerKeys[index]
	provider := Providers[selectedKey]

	fmt.Printf("\n%s %s\n", successStyle.Render("You selected:"), provider.Name)

	// Get email address
	var email string
	for {
		fmt.Print(promptStyle.Render("\nEnter your email address: "))
		emailInput, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read email: %v", err)
		}
		email = strings.TrimSpace(emailInput)
		
		// Validate email format
		if !isValidEmail(email) {
			fmt.Println("This doesn't look right. Please enter a valid email address.")
			continue
		}
		break
	}

	// Get from name (optional)
	fmt.Print(promptStyle.Render("\nEnter your display name (optional, press Enter to skip): "))
	fromName, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read display name: %v", err)
	}
	fromName = strings.TrimSpace(fromName)

	// Select AI CLI provider
	fmt.Println("\n" + headerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(headerStyle.Render("AI CLI CONFIGURATION"))
	fmt.Println(headerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Println("Select your preferred AI CLI for email automation:")
	fmt.Println()

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
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   "▸ {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "✓ {{ . | green }}",
		},
	}

	aiIndex, selectedAI, err := aiPrompt.Run()
	if err != nil {
		// Default to None if selection cancelled
		selectedAI = "None (Manual only)"
	}

	// Map selection to config value
	var defaultAICLI string
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
	case 5:
		defaultAICLI = "none"
	default:
		defaultAICLI = "none"
	}

	fmt.Printf("\n%s %s\n", successStyle.Render("Selected AI CLI:"), selectedAI)

	// Explain app passwords
	fmt.Println("\n" + headerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(headerStyle.Render("ABOUT APP PASSWORDS"))
	fmt.Println(headerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Println("What is an App Password?")
	fmt.Println("An app password is like an API key for your email account.")
	fmt.Println("It's a special password that:")
	fmt.Println("• Works only for this specific application")
	fmt.Println("• Can be revoked without changing your main password")
	fmt.Println("• Provides limited access (email only, not full account)")
	fmt.Println("• Is more secure than using your regular password")
	fmt.Println()
	fmt.Println("Think of it as giving a valet key to your car:")
	fmt.Println("• It can drive the car (send/read emails)")
	fmt.Println("• But can't open the trunk (access account settings)")
	fmt.Println("• You can take it back anytime (revoke access)")
	fmt.Println()
	fmt.Printf("%s requires an app-specific password for security.\n", provider.Name)
	fmt.Printf("%s\n", provider.AppPasswordHelp)
	fmt.Println()
	fmt.Printf("Direct link: %s\n", provider.AppPasswordURL)
	fmt.Println()
	fmt.Print(promptStyle.Render("Press ENTER to open the app password page in your browser, or visit the link above manually..."))
	reader.ReadString('\n')

	// Open browser to app password URL
	fmt.Printf("\nOpening %s app password page...\n", provider.Name)
	if err := openBrowser(provider.AppPasswordURL); err != nil {
		fmt.Printf("Could not open browser automatically.\n")
		fmt.Printf("Please manually visit: %s\n", provider.AppPasswordURL)
	}

	fmt.Print(promptStyle.Render("\nOnce you've generated your app password, press ENTER to continue..."))
	reader.ReadString('\n')

	// Get app password
	fmt.Print(promptStyle.Render("\nEnter your app password: "))
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %v", err)
	}
	password := string(passwordBytes)
	fmt.Println() // New line after password input

	// Create config
	config := &Config{
		Provider:     selectedKey,
		Email:        email,
		Password:     password,
		FromName:     fromName,
		LicenseKey:   licenseKey,
		DefaultAICLI: defaultAICLI,
	}

	// Save config
	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %v", err)
	}

	// Create README in .email directory
	if err := CreateReadme(); err != nil {
		fmt.Printf("Warning: failed to create README in .email: %v\n", err)
	}

	// Copy EmailOS README to current directory
	if err := copyReadmeToCurrentDir(); err != nil {
		fmt.Printf("Warning: failed to copy EmailOS README to current directory: %v\n", err)
	} else {
		fmt.Println(successStyle.Render("\n✓ EmailOS README.md has been added to your current directory"))
	}

	fmt.Println(successStyle.Render("\n✓ Configuration saved successfully!"))
	fmt.Println(successStyle.Render("\n🎉 Congratulations! EmailOS has been successfully installed!"))
	fmt.Println()
	fmt.Println(headerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	
	// Customize message based on AI CLI provider
	switch defaultAICLI {
	case "claude-code", "claude-code-yolo":
		fmt.Println(headerStyle.Render("CLAUDE CODE INTEGRATION"))
		fmt.Println(headerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()
		fmt.Println("The EmailOS guide and instructions have been added to your")
		fmt.Println("current folder as 'EMAILOS.md'.")
		fmt.Println()
		fmt.Println("You can now use EmailOS with Claude Code by running:")
		fmt.Println()
		fmt.Printf("  mailos \"your query here\"\n")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  mailos \"send an email to john@example.com about the meeting\"")
		fmt.Println("  mailos \"read my unread emails from today\"")
		fmt.Println("  mailos \"find emails from alice about the project\"")
		fmt.Println()
		fmt.Println("Claude Code will automatically have access to the EMAILOS.md")
		fmt.Println("file with all necessary commands and instructions.")
		
	case "openai-codex", "gemini-cli", "opencode":
		aiName := map[string]string{
			"openai-codex": "OpenAI Codex",
			"gemini-cli": "Gemini CLI",
			"opencode": "OpenCode",
		}[defaultAICLI]
		
		fmt.Println(headerStyle.Render(fmt.Sprintf("%s INTEGRATION", aiName)))
		fmt.Println(headerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()
		fmt.Println("The EmailOS guide and instructions have been added to your")
		fmt.Println("current folder as 'EMAILOS.md'.")
		fmt.Println()
		fmt.Printf("You can now use EmailOS with %s by running:\n", aiName)
		fmt.Println()
		fmt.Printf("  mailos \"your query here\"\n")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  mailos \"send an email to john@example.com about the meeting\"")
		fmt.Println("  mailos \"read my unread emails from today\"")
		fmt.Println("  mailos \"find emails from alice about the project\"")
		
	default:
		fmt.Println(headerStyle.Render("EMAILOS COMMANDS & AI INTEGRATION"))
		fmt.Println(headerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println()
		fmt.Println("The EmailOS guide and instructions have been added to your")
		fmt.Println("current folder as 'EMAILOS.md'.")
		fmt.Println()
		fmt.Println("When using Claude Code or Cursor in this directory, simply")
		fmt.Println("include @EMAILOS.md in your prompt to provide the LLM with")
		fmt.Println("all necessary email commands and instructions.")
		fmt.Println()
		fmt.Println("Example: \"@EMAILOS.md send an email to user@example.com\"")
	}
	
	fmt.Println()
	fmt.Println(headerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Println("Configuration has been saved to ~/.email/config.json")
	fmt.Println("You can now use the mailos commands to send and read emails.")

	return nil
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

