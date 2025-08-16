// auth.go - Authentication validation and management
// This file handles email authentication checks and ensures valid credentials
// are available before allowing email operations.

package mailos

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AuthError represents an authentication error with detailed information
type AuthError struct {
	Type        string // "missing_config", "missing_password", "missing_email", "invalid_provider"
	Message     string
	Suggestion  string
	Provider    string
	Email       string
}

func (e *AuthError) Error() string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("❌ Authentication Error\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	sb.WriteString(fmt.Sprintf("Problem: %s\n", e.Message))
	
	if e.Email != "" {
		sb.WriteString(fmt.Sprintf("Account: %s\n", e.Email))
	}
	if e.Provider != "" {
		sb.WriteString(fmt.Sprintf("Provider: %s\n", GetProviderName(e.Provider)))
	}
	
	sb.WriteString(fmt.Sprintf("\n✨ Solution: %s\n", e.Suggestion))
	sb.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	
	return sb.String()
}

// ValidateAuthentication checks if valid email authentication is available
// It checks both local and global configurations to ensure credentials exist
func ValidateAuthentication(config *Config) error {
	// No config at all
	if config == nil {
		return &AuthError{
			Type:    "missing_config",
			Message: "No email configuration found",
			Suggestion: `Run 'mailos setup' to configure your email account.
   This will guide you through setting up your email provider and app password.`,
		}
	}
	
	// Check for email address
	email := config.Email
	if config.FromEmail != "" {
		email = config.FromEmail
	}
	
	if email == "" {
		return &AuthError{
			Type:    "missing_email",
			Message: "No email address configured",
			Suggestion: `Run 'mailos setup' to configure your email address.
   Or use 'mailos accounts' to select an existing account.`,
		}
	}
	
	// Check for provider
	if config.Provider == "" {
		return &AuthError{
			Type:    "invalid_provider",
			Message: "No email provider configured",
			Email:   email,
			Suggestion: `Run 'mailos setup' to configure your email provider.
   Supported providers: Gmail, Fastmail, Outlook, Yahoo, Zoho.`,
		}
	}
	
	// Check for app password - this is the critical authentication piece
	if config.Password == "" {
		// Try to get password from global config if not in local
		password := getPasswordFromGlobalConfig(email, config.Provider)
		if password != "" {
			// Found password in global config, update the current config
			config.Password = password
			return nil
		}
		
		provider, exists := Providers[config.Provider]
		providerName := config.Provider
		if exists {
			providerName = provider.Name
		}
		
		suggestion := fmt.Sprintf(`You need to set up an app-specific password for %s.
   
   Steps:
   1. Go to your %s account settings
   2. Enable 2-factor authentication (if not already enabled)
   3. Generate an app-specific password`, email, providerName)
		
		if exists && provider.AppPasswordURL != "" {
			suggestion += fmt.Sprintf("\n   4. Visit: %s", provider.AppPasswordURL)
		}
		
		suggestion += fmt.Sprintf("\n   5. Run 'mailos setup' and enter the app password when prompted")
		
		return &AuthError{
			Type:       "missing_password",
			Message:    fmt.Sprintf("No app password configured for %s", email),
			Email:      email,
			Provider:   config.Provider,
			Suggestion: suggestion,
		}
	}
	
	return nil
}

// getPasswordFromGlobalConfig attempts to retrieve password from global config
func getPasswordFromGlobalConfig(email, provider string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	
	globalConfigPath := filepath.Join(homeDir, ".email", "config.json")
	globalConfig, err := loadConfigFromPath(globalConfigPath)
	if err != nil || globalConfig == nil {
		return ""
	}
	
	// Check if the main account matches
	if globalConfig.Email == email && globalConfig.Password != "" {
		return globalConfig.Password
	}
	
	// Check if FromEmail matches
	if globalConfig.FromEmail == email && globalConfig.Password != "" {
		return globalConfig.Password
	}
	
	// Check accounts array
	for _, acc := range globalConfig.Accounts {
		if acc.Email == email {
			if acc.Password != "" {
				return acc.Password
			}
			// If account has same provider as global, use global password
			if acc.Provider == globalConfig.Provider && globalConfig.Password != "" {
				return globalConfig.Password
			}
		}
	}
	
	// If same provider, might be a sub-email - use provider's password
	if provider == globalConfig.Provider && globalConfig.Password != "" {
		return globalConfig.Password
	}
	
	return ""
}

// EnsureAuthenticated is a wrapper that validates authentication before proceeding
// This should be called at the start of any command that requires email access
func EnsureAuthenticated(accountEmail string) (*Config, error) {
	// Initialize mail setup with the specified account
	setup, err := InitializeMailSetup(accountEmail)
	if err != nil {
		return nil, err
	}
	
	// Validate authentication
	if err := ValidateAuthentication(setup.Config); err != nil {
		return nil, err
	}
	
	return setup.Config, nil
}

// IsAuthenticated checks if authentication is valid without returning detailed errors
func IsAuthenticated(config *Config) bool {
	if config == nil {
		return false
	}
	
	email := config.Email
	if config.FromEmail != "" {
		email = config.FromEmail
	}
	
	if email == "" || config.Provider == "" {
		return false
	}
	
	// Check for password locally or globally
	if config.Password != "" {
		return true
	}
	
	password := getPasswordFromGlobalConfig(email, config.Provider)
	if password != "" {
		config.Password = password
		return true
	}
	
	return false
}

// GetAuthenticationStatus returns a human-readable authentication status
func GetAuthenticationStatus(config *Config) string {
	if config == nil {
		return "❌ Not configured"
	}
	
	email := config.Email
	if config.FromEmail != "" {
		email = config.FromEmail
	}
	
	if email == "" {
		return "❌ No email configured"
	}
	
	if config.Provider == "" {
		return "❌ No provider configured"
	}
	
	if config.Password == "" {
		password := getPasswordFromGlobalConfig(email, config.Provider)
		if password == "" {
			return fmt.Sprintf("⚠️  No app password for %s", email)
		}
	}
	
	return fmt.Sprintf("✅ Authenticated as %s (%s)", email, GetProviderName(config.Provider))
}