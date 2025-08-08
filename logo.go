package mailos

import (
	"fmt"
	"os"
)

// DisplayEmailOSLogo shows the EmailOS ASCII art logo with colors
func DisplayEmailOSLogo() {
	// New stylized ASCII logo with mail theme
	logo := `
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║     ███╗   ███╗ █████╗ ██╗██╗      ██████╗ ███████╗                        ║
║     ████╗ ████║██╔══██╗██║██║     ██╔═══██╗██╔════╝    ✉  ✉  ✉           ║
║     ██╔████╔██║███████║██║██║     ██║   ██║███████╗       📮              ║
║     ██║╚██╔╝██║██╔══██║██║██║     ██║   ██║╚════██║    ✉  ✉  ✉           ║
║     ██║ ╚═╝ ██║██║  ██║██║███████╗╚██████╔╝███████║                        ║
║     ╚═╝     ╚═╝╚═╝  ╚═╝╚═╝╚══════╝ ╚═════╝ ╚══════╝                        ║
║                                                                              ║
║                   🚀 Your AI-Powered Email Command Center 🚀                ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
`
	// Cyan color for the logo
	fmt.Print("\033[36m" + logo + "\033[0m")
}

// DisplayCompactLogo shows a compact version of the logo for quick commands
func DisplayCompactLogo() {
	compactLogo := `
 ███╗   ███╗ █████╗ ██╗██╗      ██████╗ ███████╗
 ████╗ ████║██╔══██╗██║██║     ██╔═══██╗██╔════╝  📮
 ██╔████╔██║███████║██║██║     ██║   ██║███████╗
 ██║╚██╔╝██║██╔══██║██║██║     ██║   ██║╚════██║
 ██║ ╚═╝ ██║██║  ██║██║███████╗╚██████╔╝███████║
 ╚═╝     ╚═╝╚═╝  ╚═╝╚═╝╚══════╝ ╚═════╝ ╚══════╝`
	fmt.Print("\033[36m" + compactLogo + "\033[0m\n")
}

// DisplayStatusBox shows the enhanced status display with account and AI info
func DisplayStatusBox(config *Config) {
	// Check if we have a local config with a different from_email
	localConfigPath := ".email/config.json"
	isLocal := false
	if _, err := os.Stat(localConfigPath); err == nil {
		isLocal = true
	}
	
	// Display the appropriate email address
	displayEmail := config.Email
	if isLocal && config.FromEmail != "" && config.FromEmail != config.Email {
		displayEmail = config.FromEmail
	}
	
	// Enhanced status display with colors and better formatting
	fmt.Println()
	fmt.Print("\033[90m╭─────────────────────────────────────────────────────────────────────────────╮\033[0m\n")
	fmt.Print("\033[90m│\033[0m ")
	fmt.Printf("\033[32m📬 Account:\033[0m \033[36m%s\033[0m", displayEmail)
	
	if config.DefaultAICLI != "" && config.DefaultAICLI != "none" {
		fmt.Printf(" \033[90m│\033[0m \033[32m🤖 AI:\033[0m \033[35m%s\033[0m", GetAICLIName(config.DefaultAICLI))
	} else {
		fmt.Printf(" \033[90m│\033[0m \033[33m⚠️  No AI provider\033[0m \033[90m(use /provider)\033[0m")
	}
	fmt.Print(" \033[90m│\033[0m\n")
	fmt.Print("\033[90m├─────────────────────────────────────────────────────────────────────────────┤\033[0m\n")
	fmt.Print("\033[90m│\033[0m \033[36m💡 Enter a query for AI or type \033[33m'/'\033[36m to see commands\033[0m")
	// Pad to right edge
	fmt.Print("                          \033[90m│\033[0m\n")
	fmt.Print("\033[90m╰─────────────────────────────────────────────────────────────────────────────╯\033[0m\n")
}

// DisplayMinimalStatus shows just the account and AI provider info in one line
func DisplayMinimalStatus(config *Config) {
	// Check if we have a local config with a different from_email
	localConfigPath := ".email/config.json"
	isLocal := false
	if _, err := os.Stat(localConfigPath); err == nil {
		isLocal = true
	}
	
	// Display the appropriate email address
	displayEmail := config.Email
	if isLocal && config.FromEmail != "" && config.FromEmail != config.Email {
		displayEmail = config.FromEmail
	}
	
	fmt.Printf("\033[32m📬\033[0m \033[36m%s\033[0m", displayEmail)
	if config.DefaultAICLI != "" && config.DefaultAICLI != "none" {
		fmt.Printf(" \033[90m│\033[0m \033[32m🤖\033[0m \033[35m%s\033[0m", GetAICLIName(config.DefaultAICLI))
	}
	fmt.Println()
}

// DisplaySingleLineHeader displays a compact single-line header with configuration info
func DisplaySingleLineHeader(config *Config) {
	// Check if we have a local config with a different from_email
	localConfigPath := ".email/config.json"
	isLocal := false
	if _, err := os.Stat(localConfigPath); err == nil {
		isLocal = true
	}
	
	// Display the appropriate email address
	displayEmail := config.Email
	if isLocal && config.FromEmail != "" && config.FromEmail != config.Email {
		displayEmail = config.FromEmail
	}
	
	// Display single line header with configuration
	fmt.Print("\033[90m╭─────────────────────────────────────────────────────────────────────────────╮\033[0m\n")
	fmt.Print("\033[90m│\033[0m ")
	fmt.Printf("\033[32m📬 Account:\033[0m \033[36m%s\033[0m", displayEmail)
	
	if config.DefaultAICLI != "" && config.DefaultAICLI != "none" {
		fmt.Printf(" \033[90m│\033[0m \033[32m🤖 AI:\033[0m \033[35m%s\033[0m", GetAICLIName(config.DefaultAICLI))
	} else {
		fmt.Printf(" \033[90m│\033[0m \033[33m⚠️  No AI provider\033[0m \033[90m(use /provider)\033[0m")
	}
	fmt.Print(" \033[90m│\033[0m\n")
	fmt.Print("\033[90m╰─────────────────────────────────────────────────────────────────────────────╯\033[0m\n")
}

// ShouldShowLogo determines if the logo should be shown based on environment and context
func ShouldShowLogo() bool {
	// Check for environment variable to suppress logo
	if os.Getenv("MAILOS_NO_LOGO") == "true" {
		return false
	}
	
	// Check if running in a non-interactive mode
	if os.Getenv("CI") == "true" {
		return false
	}
	
	return true
}

// DisplayWelcome shows the appropriate welcome message based on context
func DisplayWelcome(showFull bool) {
	if !ShouldShowLogo() {
		return
	}
	
	if showFull {
		DisplayEmailOSLogo()
	} else {
		DisplayCompactLogo()
	}
}