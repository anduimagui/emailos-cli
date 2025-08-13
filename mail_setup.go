package mailos

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// MailSetup holds the configuration and account information for email operations
type MailSetup struct {
	Config        *Config
	AccountEmail  string
	IsInitialized bool
}

// Global mail setup instance
var globalMailSetup *MailSetup

// CommandFlags holds common flags for email commands
type CommandFlags struct {
	Account string // Email account to use
}

// RegisterAccountFlag registers the --account flag for a command
func RegisterAccountFlag(fs *flag.FlagSet) *string {
	return fs.String("account", "", "Email account to use (e.g., user@example.com)")
}

// InitializeMailSetup initializes the mail configuration with optional account switching
func InitializeMailSetup(accountEmail string) (*MailSetup, error) {
	var config *Config
	var err error
	
	// Priority order for account selection:
	// 1. Explicitly specified account (command line --account flag)
	// 2. Local folder preference (.email/config.json active_account)
	// 3. Session default (MAILOS_SESSION_ACCOUNT environment variable)
	// 4. Default config
	
	if accountEmail == "" {
		// Check for local folder preference first
		localAccount := GetLocalAccountPreference()
		if localAccount != "" {
			accountEmail = localAccount
		} else {
			// Fall back to session default
			accountEmail = GetSessionDefaultAccount()
		}
	}
	
	// Load configuration for specific account or default
	if accountEmail != "" {
		config, err = LoadAccountConfig(accountEmail)
		if err != nil {
			// If account not found, list available accounts
			availableAccounts := listAvailableAccounts()
			if len(availableAccounts) > 0 {
				return nil, fmt.Errorf("account '%s' not found. Available accounts:\n%s", 
					accountEmail, strings.Join(availableAccounts, "\n"))
			}
			return nil, fmt.Errorf("account '%s' not found and no accounts configured", accountEmail)
		}
		
		// Set as session default for subsequent commands (but don't override local preference)
		if GetLocalAccountPreference() == "" {
			SetSessionDefaultAccount(accountEmail)
		}
	} else {
		config, err = LoadConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load configuration: %v", err)
		}
	}
	
	// Create and cache the setup
	setup := &MailSetup{
		Config:        config,
		AccountEmail:  accountEmail,
		IsInitialized: true,
	}
	
	globalMailSetup = setup
	return setup, nil
}

// GetMailSetup returns the current mail setup, initializing if needed
func GetMailSetup() (*MailSetup, error) {
	if globalMailSetup != nil && globalMailSetup.IsInitialized {
		return globalMailSetup, nil
	}
	
	// Initialize with default config if not already initialized
	return InitializeMailSetup("")
}

// GetMailSetupForAccount returns mail setup for a specific account
func GetMailSetupForAccount(accountEmail string) (*MailSetup, error) {
	// If requesting same account that's already loaded, return it
	if globalMailSetup != nil && 
	   globalMailSetup.IsInitialized && 
	   globalMailSetup.AccountEmail == accountEmail {
		return globalMailSetup, nil
	}
	
	// Initialize with the requested account
	return InitializeMailSetup(accountEmail)
}

// listAvailableAccounts returns a list of configured account emails with labels
func listAvailableAccounts() []string {
	// Try to load config to get accounts
	config, err := LoadConfig()
	if err != nil {
		return []string{}
	}
	
	accounts := GetAllAccounts(config)
	result := make([]string, 0, len(accounts))
	
	for _, acc := range accounts {
		if acc.Label != "" {
			result = append(result, fmt.Sprintf("  - %s (%s)", acc.Email, acc.Label))
		} else {
			result = append(result, fmt.Sprintf("  - %s", acc.Email))
		}
	}
	
	return result
}

// ParseAccountFromArgs checks for --account flag in command arguments
func ParseAccountFromArgs(args []string) (string, []string) {
	var accountEmail string
	var cleanArgs []string
	
	skipNext := false
	for i, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		
		// Check for --account flag
		if arg == "--account" || arg == "-a" {
			if i+1 < len(args) {
				accountEmail = args[i+1]
				skipNext = true
			}
			continue
		}
		
		// Check for --account=email format
		if strings.HasPrefix(arg, "--account=") {
			accountEmail = strings.TrimPrefix(arg, "--account=")
			continue
		}
		
		if strings.HasPrefix(arg, "-a=") {
			accountEmail = strings.TrimPrefix(arg, "-a=")
			continue
		}
		
		// Add to clean args if not account-related
		cleanArgs = append(cleanArgs, arg)
	}
	
	return accountEmail, cleanArgs
}

// WithAccount is a helper function to wrap command functions with account setup
func WithAccount(accountEmail string, fn func(*Config) error) error {
	setup, err := InitializeMailSetup(accountEmail)
	if err != nil {
		return err
	}
	
	return fn(setup.Config)
}

// GetConfigForCommand gets the config for a command, checking for --account flag
func GetConfigForCommand(args []string) (*Config, []string, error) {
	accountEmail, cleanArgs := ParseAccountFromArgs(args)
	
	setup, err := InitializeMailSetup(accountEmail)
	if err != nil {
		return nil, cleanArgs, err
	}
	
	return setup.Config, cleanArgs, nil
}

// ResetMailSetup resets the global mail setup (useful for testing)
func ResetMailSetup() {
	globalMailSetup = nil
}

// GetFromAddress returns the appropriate from address for the current setup
func (ms *MailSetup) GetFromAddress() string {
	if ms.Config.FromEmail != "" {
		return ms.Config.FromEmail
	}
	return ms.Config.Email
}

// GetFromDisplay returns the formatted from address with display name
func (ms *MailSetup) GetFromDisplay() string {
	fromEmail := ms.GetFromAddress()
	if ms.Config.FromName != "" {
		return fmt.Sprintf("%s <%s>", ms.Config.FromName, fromEmail)
	}
	return fromEmail
}

// GetSMTPConfig returns SMTP configuration for the current setup
func (ms *MailSetup) GetSMTPConfig() (host string, port int, useTLS bool, useSSL bool, err error) {
	return ms.Config.GetSMTPSettings()
}

// GetIMAPConfig returns IMAP configuration for the current setup
func (ms *MailSetup) GetIMAPConfig() (host string, port int, err error) {
	return ms.Config.GetIMAPSettings()
}

// PrintAccountInfo prints information about the current account
func (ms *MailSetup) PrintAccountInfo() {
	if ms.AccountEmail != "" {
		fmt.Printf("Using account: %s\n", ms.Config.Email)
	}
	fmt.Printf("Provider: %s\n", GetProviderName(ms.Config.Provider))
	if ms.Config.FromEmail != "" && ms.Config.FromEmail != ms.Config.Email {
		fmt.Printf("Sending as: %s\n", ms.Config.FromEmail)
	}
	if ms.Config.FromName != "" {
		fmt.Printf("Display name: %s\n", ms.Config.FromName)
	}
}

// Session account management using environment variable
const SessionAccountEnvVar = "MAILOS_SESSION_ACCOUNT"

// SetSessionDefaultAccount sets the default account for the current terminal session
func SetSessionDefaultAccount(accountEmail string) {
	if accountEmail != "" {
		os.Setenv(SessionAccountEnvVar, accountEmail)
	}
}

// GetSessionDefaultAccount gets the default account for the current terminal session
func GetSessionDefaultAccount() string {
	return os.Getenv(SessionAccountEnvVar)
}

// ClearSessionDefaultAccount clears the session default account
func ClearSessionDefaultAccount() {
	os.Unsetenv(SessionAccountEnvVar)
}

// PrintSessionAccount prints the current session account if set
func PrintSessionAccount() {
	sessionAccount := GetSessionDefaultAccount()
	if sessionAccount != "" {
		fmt.Printf("Session default account: %s\n", sessionAccount)
	}
}