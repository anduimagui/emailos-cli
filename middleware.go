package mailos

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
	
	"github.com/charmbracelet/lipgloss"
)

// isAIEnvironment detects if running in an AI CLI environment
func isAIEnvironment() bool {
	// Check for common AI CLI environment variables and parent processes
	aiIndicators := []string{
		"CLAUDE_CODE", "CLAUDE_CLI", "CODEX", "OPENAI_CLI", "GEMINI_CLI", 
		"AI_CLI", "ANTHROPIC_CLI", "GPT_CLI", "LLM_CLI",
	}
	
	for _, indicator := range aiIndicators {
		if os.Getenv(indicator) != "" {
			return true
		}
	}
	
	// Check parent process name for AI CLI tools
	if parent := os.Getenv("_"); parent != "" {
		aiProcesses := []string{"claude", "codex", "gpt", "gemini", "ai-cli"}
		for _, process := range aiProcesses {
			if strings.Contains(strings.ToLower(parent), process) {
				return true
			}
		}
	}
	
	// Check if stdin/stdout is not a terminal (pipe/redirect)
	if !isInteractiveTerminal() {
		return true
	}
	
	return false
}

// isInteractiveTerminal checks if we're running in an interactive terminal
func isInteractiveTerminal() bool {
	// Check if stdin is a terminal
	if fileInfo, _ := os.Stdin.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return false
	}
	
	// Check if stdout is a terminal  
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return false
	}
	
	return true
}

// EnsureInitialized validates that the system is properly configured
// This should be called before any command that requires email configuration
// Now defaults to showing error messages instead of launching interactive setup
func EnsureInitialized() error {
	// Check if config exists
	if !ConfigExists() {
		return fmt.Errorf("MailOS is not configured. Run: mailos setup\n\nYou'll need:\n• Your email address\n• App-specific password (not your regular password)\n• Email provider info (Gmail, Fastmail, Outlook, etc.)")
	}
	
	// Load config 
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}
	
	// Email configuration is required, but license is now optional
	if config.Email == "" || config.Password == "" {
		return fmt.Errorf("MailOS configuration is incomplete. Run: mailos setup\n\nYou'll need:\n• Your email address\n• App-specific password (not your regular password)")
	}
	
	// Check if auto-sync is needed (run in background, don't block)
	go func() {
		if err := RunAutoSyncIfNeeded(); err != nil {
			// Silently fail - auto-sync is not critical
		}
	}()
	
	return nil
}

// EnsureInitializedInteractive is the old behavior that launches interactive setup
// This should only be used for commands that explicitly want interactive setup
func EnsureInitializedInteractive() error {
	// Check if config exists
	if !ConfigExists() {
		// If running in AI environment, don't run interactive setup
		if isAIEnvironment() {
			return fmt.Errorf("email configuration not found. Please run 'mailos setup' in an interactive terminal to configure EmailOS")
		}
		// Run setup automatically when no config exists in interactive mode
		return Setup()
	}
	
	// Load config 
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}
	
	// Email configuration is required, but license is now optional
	if config.Email == "" || config.Password == "" {
		// If running in AI environment, don't run interactive setup
		if isAIEnvironment() {
			return fmt.Errorf("email configuration incomplete. Please run 'mailos setup' in an interactive terminal to complete configuration")
		}
		// Run setup to configure email if missing in interactive mode
		return Setup()
	}
	
	// Check if auto-sync is needed (run in background, don't block)
	go func() {
		if err := RunAutoSyncIfNeeded(); err != nil {
			// Silently fail - auto-sync is not critical
		}
	}()
	
	return nil
}

// IsSubscribed checks if the user has a valid subscription
// Returns true if user has valid license, false otherwise
// This function prioritizes cache and minimizes API calls for security
func IsSubscribed() bool {
	config, err := LoadConfig()
	if err != nil {
		return false
	}
	
	// No license key means not subscribed
	if config.LicenseKey == "" {
		return false
	}
	
	// Get license manager but only use cached data for security
	lm := GetLicenseManager()
	
	// Check session cache first (fastest path)
	if validated, duration, sessionKey := lm.GetSessionStatus(); validated && 
		sessionKey == config.LicenseKey && duration < SessionCacheDuration {
		return true
	}
	
	// First check if we have valid cached license (no API call)
	if cache := lm.GetCachedLicense(); cache != nil && cache.Key == config.LicenseKey {
		// Check if cache is still valid (not expired)
		if cache.ExpiresAt.After(time.Now()) {
			return true
		}
	}
	
	// If cache expired, check grace period (no API call)
	if lm.IsInGracePeriod() {
		return true
	}
	
	// Only make API call in background, don't block or expose errors
	go func() {
		// Silent background validation - don't expose results to user context
		lm.QuickValidate(config.LicenseKey)
	}()
	
	// Default to not subscribed if no valid cache
	return false
}

// enterLicenseKey prompts the user to enter a license key and saves it
func enterLicenseKey(config *Config) error {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println()
	for {
		fmt.Print("Enter your license key (or 'cancel' to exit): ")
		licenseKey, _ := reader.ReadString('\n')
		licenseKey = strings.TrimSpace(licenseKey)
		
		if strings.ToLower(licenseKey) == "cancel" {
			return fmt.Errorf("license entry cancelled")
		}
		
		if licenseKey == "" {
			fmt.Println("License key cannot be empty.")
			continue
		}
		
		fmt.Println("Validating license key...")
		lm := GetLicenseManager()
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
		
		if err := lm.ValidateLicense(licenseKey); err != nil {
			fmt.Println(errorStyle.Render("✗ Invalid license key"))
			fmt.Println()
			fmt.Printf("Visit https://%s/checkout to purchase a valid license.\n", APP_SITE)
			fmt.Println()
			continue
		}
		
		// License is valid, save it
		config.LicenseKey = licenseKey
		if err := SaveConfig(config); err != nil {
			return fmt.Errorf("failed to save configuration: %v", err)
		}
		
		fmt.Println(successStyle.Render("✓ License validated and saved successfully!"))
		
		// Show customer info if available
		if cache := lm.GetCachedLicense(); cache != nil && cache.CustomerEmail != "" {
			fmt.Printf("%s %s\n", successStyle.Render("✓ Licensed to:"), cache.CustomerEmail)
		}
		
		fmt.Println()
		return nil
	}
}

// ValidateLicenseOnly performs only license validation without full initialization
// Used by commands that already have their own config loading logic
func ValidateLicenseOnly() error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}
	
	if config.LicenseKey == "" {
		return fmt.Errorf("no license key configured")
	}
	
	lm := GetLicenseManager()
	if err := lm.QuickValidate(config.LicenseKey); err != nil {
		// Check grace period
		if lm.IsInGracePeriod() {
			return nil // Allow operation in grace period
		}
		return fmt.Errorf("license validation failed: %v", err)
	}
	
	return nil
}