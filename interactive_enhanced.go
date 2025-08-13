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
	return InteractiveModeWithMenuOptions(false, true)
}

// InteractiveModeWithMenuOptions runs the enhanced interactive mode with control over display
func InteractiveModeWithMenuOptions(showLogo bool, showInitialStatus bool) error {
	// Check for MAILOS_USE_BUBBLETEA to use new Bubble Tea UI (default)
	useBubbleTea := os.Getenv("MAILOS_USE_BUBBLETEA") != "false"
	
	// Check for legacy UI modes
	useInkUI := os.Getenv("MAILOS_USE_INK") == "true"
	
	if useBubbleTea && !useInkUI {
		// Use the new Bubble Tea implementation by default
		return InteractiveModeWithBubbleTea()
	}
	
	if useInkUI {
		return InteractiveModeWithReactInk()
	}

	// Fall back to classic UI
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
	firstIteration := true
	for {
		// Show appropriate header on first iteration
		if firstIteration {
			if showLogo && ShouldShowLogo() {
				DisplayEmailOSLogo()
			} else if showInitialStatus {
				// Show single-line header instead of full logo
				DisplaySingleLineHeader(config)
			}
		}
		
		// Show status on subsequent iterations but not on first if we already showed header
		shouldShowStatus := !firstIteration
		if err := showEnhancedInteractiveMenuWithOptions(needsProvider, shouldShowStatus); err != nil {
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
		firstIteration = false
	}
}


// showEnhancedInteractiveMenu displays the input-first interactive menu
func showEnhancedInteractiveMenu(needsProvider bool) error {
	return showEnhancedInteractiveMenuWithOptions(needsProvider, true)
}

// showEnhancedInteractiveMenuWithOptions displays the input-first interactive menu with control over status display
func showEnhancedInteractiveMenuWithOptions(needsProvider bool, showStatus bool) error {
	// Show current status
	config, _ := LoadConfig()
	
	if showStatus {
		// Use the new DisplayStatusBox function from logo.go
		DisplayStatusBox(config)
	}
	
	// Check for suggestion mode preference
	suggestionMode := os.Getenv("MAILOS_SUGGESTION_MODE")
	
	switch suggestionMode {
	case "dynamic":
		// Dynamic filtering as you type
		return InteractiveModeWithDynamicSuggestions(config)
	case "simple":
		// Simple Enter for suggestions
		return EnhancedInteractiveMode(config)
	case "live":
		// Live input with suggestions
		return EnhancedInteractiveModeV2(config)
	case "clean":
		// Cleanest implementation
		return CleanInteractiveMode(config)
	default:
		// Default to dynamic mode with live suggestions
		return InteractiveModeWithDynamicSuggestions(config)
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
		{"Read Emails", "/read", "Browse and read your emails", IconEmail, handleInteractiveRead},
		{"Send Email", "/send", "Compose and send a new email", IconSend, handleInteractiveSend},
		{"Open Inbox", "/inbox", "Open inbox in web browser", IconInbox, func() error { return OpenInbox() }},
		{"Open Sent Mail", "/sent", "Open sent mail in web browser", IconSent, func() error { return OpenSentMail() }},
		{"Open Drafts", "/drafts", "Open drafts in web browser", IconDrafts, func() error { return OpenDrafts() }},
		{"Email Report", "/report", "Generate email analytics", IconReport, handleInteractiveReport},
		{"Unsubscribe Links", "/unsubscribe", "Find unsubscribe links", IconUnsubscribe, handleInteractiveUnsubscribe},
		{"Delete Emails", "/delete", "Delete emails by criteria", IconDelete, handleInteractiveDelete},
		{"Mark as Read", "/mark-read", "Mark emails as read", IconCheck, handleInteractiveMarkRead},
		{"Templates", "/template", "Manage email templates", IconTemplate, func() error { return ManageTemplate() }},
		{"Configuration", "/configure", "Settings & configuration", IconSettings, handleInteractiveConfigure},
		{"AI Provider", "/provider", "Set AI provider", IconAI, func() error { return SelectAndConfigureAIProvider() }},
		{"Show Info", "/info", "Display configuration", IconInfo, func() error { return showInfo() }},
		{"Help", "/help", "Show help information", IconHelp, showInteractiveHelp},
		{"Exit", "/exit", "Exit EmailOS", IconExit, func() error { return fmt.Errorf("exit") }},
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
	case "inbox":
		return OpenInbox()
	case "sent":
		return OpenSentMail()
	case "drafts":
		return OpenDrafts()
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
  • Type "@" to tag files/folders from current directory
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
  • ESC ESC     - Clear current input (press ESC twice quickly)
  • @          - Show file/folder autocomplete
  • /          - Show command menu
  • ↑↓         - Navigate menu options (also in @ mode)
  • Tab        - Auto-complete selected file (in @ mode)
  • ESC        - Cancel autocomplete (in @ mode)
  • Ctrl+C     - Cancel/Go back
  • Ctrl+D     - Exit (when input is empty)
  • Backspace  - Delete character

TIPS:
  • Configure AI provider for best experience
  • Email templates support Markdown formatting
  • Use @ to quickly reference files in your queries
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

// GetInputWithEscapeClear is deprecated - use ReadLineWithFileAutocomplete instead
// Kept for backward compatibility
func GetInputWithEscapeClear() string {
	return ReadLineWithFileAutocomplete("▸ ")
}

