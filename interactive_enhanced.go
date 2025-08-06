package mailos

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
)

// InteractiveModeWithMenu runs the enhanced interactive mode with logo and input-first design
func InteractiveModeWithMenu() error {
	// Display EmailOS logo and branding
	displayEmailOSLogo()

	// Check configuration
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Check for slash config
	slashConfig := loadSlashConfig()
	
	// If no AI provider configured, show setup prompt inline
	needsProvider := (config.DefaultAICLI == "" || config.DefaultAICLI == "none") && !hasConfiguredProvider(slashConfig)
	
	// Main interactive loop
	for {
		if err := showEnhancedInteractiveMenu(needsProvider); err != nil {
			if err.Error() == "exit" {
				fmt.Println("\n👋 Goodbye!")
				return nil
			}
			// After provider setup, update needsProvider flag
			if needsProvider {
				config, _ = LoadConfig()
				needsProvider = (config.DefaultAICLI == "" || config.DefaultAICLI == "none")
			}
		}
	}
}

// displayEmailOSLogo shows the EmailOS ASCII art logo
func displayEmailOSLogo() {
	// Check if license is already validated to avoid showing it again
	lm := GetLicenseManager()
	if lm != nil && lm.GetCachedLicense() != nil && lm.GetCachedLicense().Status == "valid" {
		logo := `
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║   ███████╗███╗   ███╗ █████╗ ██╗██╗      ██████╗ ███████╗   ║
║   ██╔════╝████╗ ████║██╔══██╗██║██║     ██╔═══██╗██╔════╝   ║
║   █████╗  ██╔████╔██║███████║██║██║     ██║   ██║███████╗   ║
║   ██╔══╝  ██║╚██╔╝██║██╔══██║██║██║     ██║   ██║╚════██║   ║
║   ███████╗██║ ╚═╝ ██║██║  ██║██║███████╗╚██████╔╝███████║   ║
║   ╚══════╝╚═╝     ╚═╝╚═╝  ╚═╝╚═╝╚══════╝ ╚═════╝ ╚══════╝   ║
║                                                               ║
║            Your AI-Powered Email Management System            ║
╚═══════════════════════════════════════════════════════════════╝
`
		fmt.Print("\033[36m" + logo + "\033[0m\n")
	}
	// If not licensed, the license check will show its own branding
}

// showEnhancedInteractiveMenu displays the input-first interactive menu
func showEnhancedInteractiveMenu(needsProvider bool) error {
	// Show current status
	config, _ := LoadConfig()
	fmt.Printf("\n📬 Account: %s", config.Email)
	if config.DefaultAICLI != "" && config.DefaultAICLI != "none" {
		fmt.Printf(" | 🤖 AI: %s", GetAICLIName(config.DefaultAICLI))
	} else if needsProvider {
		fmt.Printf(" | ⚠️  No AI provider (use /provider to setup)")
	}
	fmt.Println()

	// Show input prompt with command hints
	fmt.Println("💡 Enter a query for AI or type '/' to see commands")
	fmt.Println("────────────────────────────────────────────────────────────────────────────")
	
	// Get user input
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("▸ ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("exit")
	}
	
	input = strings.TrimSpace(input)
	
	// Handle empty input
	if input == "" {
		return nil
	}
	
	// Check if it's a command or show command menu
	if input == "/" {
		// Show command selection menu
		fmt.Println("\n📋 Available Commands:")
		fmt.Println("Debug: About to call showCommandMenu()")
		err := showCommandMenu()
		if err != nil {
			fmt.Printf("Error from showCommandMenu: %v\n", err)
		}
		return err
	} else if strings.HasPrefix(input, "/") {
		// Direct command execution
		return executeCommand(input)
	} else {
		// Process as AI query
		if config.DefaultAICLI == "" || config.DefaultAICLI == "none" {
			fmt.Println("\n⚠️  No AI provider configured.")
			setupPrompt := promptui.Select{
				Label: "Would you like to set up an AI provider now?",
				Items: []string{"Yes, set up AI provider", "No, continue without AI"},
			}
			idx, _, err := setupPrompt.Run()
			if err == nil && idx == 0 {
				return SelectAndConfigureAIProvider()
			}
			return nil
		}
		
		fmt.Printf("\n🤔 Processing: %s\n\n", input)
		return InvokeAIProvider(input)
	}
}

// showCommandMenu displays the command selection menu with arrow navigation
func showCommandMenu() error {
	// Define all available commands
	commands := []struct {
		Label       string
		Command     string
		Description string
		Icon        string
		Action      func() error
	}{
		{"Read Emails", "/read", "Browse and read your emails", "📧", handleInteractiveRead},
		{"Send Email", "/send", "Compose and send a new email", "✉️", handleInteractiveSend},
		{"Email Report", "/report", "Generate email analytics", "📊", handleInteractiveReport},
		{"Unsubscribe Links", "/unsubscribe", "Find unsubscribe links", "🔗", handleInteractiveUnsubscribe},
		{"Delete Emails", "/delete", "Delete emails by criteria", "🗑️", handleInteractiveDelete},
		{"Mark as Read", "/mark-read", "Mark emails as read", "✓", handleInteractiveMarkRead},
		{"Templates", "/template", "Manage email templates", "📝", func() error { return ManageTemplate() }},
		{"Configuration", "/configure", "Settings & configuration", "⚙️", handleInteractiveConfigure},
		{"AI Provider", "/provider", "Set AI provider", "🤖", func() error { return SelectAndConfigureAIProvider() }},
		{"Show Info", "/info", "Display configuration", "ℹ️", func() error { return showInfo() }},
		{"Help", "/help", "Show help information", "❓", showInteractiveHelp},
		{"Exit", "/exit", "Exit EmailOS", "👋", func() error { return fmt.Errorf("exit") }},
	}

	// Display commands as a list first for visibility
	fmt.Println()
	fmt.Println("────────────────────────────────────────")
	for i, cmd := range commands {
		fmt.Printf("%2d. %s %s - %s\n", i+1, cmd.Icon, cmd.Label, cmd.Description)
	}
	fmt.Println("────────────────────────────────────────")
	fmt.Println()

	// Create menu items for promptui selection
	items := make([]string, len(commands))
	for i, cmd := range commands {
		items[i] = fmt.Sprintf("%s %s", cmd.Icon, cmd.Label)
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "▸ {{ . | cyan }}",
		Inactive: "  {{ . }}",
		Selected: "✓ {{ . | green }}",
	}

	prompt := promptui.Select{
		Label:     "Use ↑↓ arrows to navigate, Enter to select",
		Items:     items,
		Templates: templates,
		Size:      12,
	}

	index, _, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			fmt.Println("\nCancelled.")
			return nil
		}
		fmt.Printf("\nError with menu selection: %v\n", err)
		return err
	}

	// Execute selected command
	cmd := commands[index]
	fmt.Printf("\n%s Executing: %s\n", cmd.Icon, cmd.Command)
	return cmd.Action()
}

// executeCommand executes a slash command directly
func executeCommand(input string) error {
	// Remove leading slash and get command parts
	input = strings.TrimPrefix(input, "/")
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}
	
	cmd := parts[0]
	args := parts[1:]
	
	// Map commands to functions
	switch cmd {
	case "read":
		return handleReadCommand(args)
	case "send":
		return handleInteractiveSend()
	case "report":
		return handleInteractiveReport()
	case "unsubscribe":
		return handleInteractiveUnsubscribe()
	case "delete":
		return handleInteractiveDelete()
	case "mark-read":
		return handleInteractiveMarkRead()
	case "template":
		return ManageTemplate()
	case "configure":
		return handleInteractiveConfigure()
	case "provider":
		return SelectAndConfigureAIProvider()
	case "info":
		return showInfo()
	case "help":
		return showInteractiveHelp()
	case "exit":
		return fmt.Errorf("exit")
	default:
		fmt.Printf("Unknown command: /%s\n", cmd)
		fmt.Println("Type /help for available commands")
		return nil
	}
}

// showInteractiveHelp displays comprehensive help
func showInteractiveHelp() error {
	help := `
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
EMAILOS HELP
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

QUICK START:
  • Type any question to ask your AI assistant
  • Type "/" to see available commands
  • Type "/command" to execute directly
  • Press Ctrl+C to go back

COMMANDS:
  /read         - Browse and read your emails
  /send         - Compose and send a new email  
  /report       - Generate email analytics report
  /unsubscribe  - Find and manage unsubscribe links
  /delete       - Delete emails by various criteria
  /mark-read    - Mark selected emails as read
  /template     - Customize email templates
  /configure    - Manage email and AI settings
  /provider     - Select or change AI provider
  /info         - Display current configuration
  /help         - Show this help message
  /exit         - Exit EmailOS

AI QUERIES:
  Just type naturally! Examples:
  • "Summarize my emails from today"
  • "Draft a reply to John's last email"
  • "Find all unread emails from this week"
  • "Help me write a professional follow-up"

KEYBOARD SHORTCUTS:
  • Enter      - Submit query or select option
  • /          - Show command menu
  • ↑↓         - Navigate menu options
  • Ctrl+C     - Cancel/Go back
  • Tab        - Auto-complete (where available)

TIPS:
  • Configure AI provider for best experience
  • Email templates support Markdown formatting
  • Use /provider if AI is not configured
  • Commands can be typed directly (e.g., /read)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`
	fmt.Println(help)
	
	fmt.Print("Press Enter to continue...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
	
	return nil
}