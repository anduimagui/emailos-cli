// config.go - Core configuration data structures and file operations
// This file handles the backend logic for loading, saving, and managing
// email configurations (both global and local).

package mailos

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

var APP_SITE = AppSite // Using constant from constants.go

type Config struct {
	Provider          string          `json:"provider"`
	Email             string          `json:"email"`
	Password          string          `json:"password"`
	FromName          string          `json:"from_name,omitempty"`
	FromEmail         string          `json:"from_email,omitempty"`
	ProfileImage      string          `json:"profile_image,omitempty"`
	LicenseKey        string          `json:"license_key,omitempty"`
	DefaultAICLI      string          `json:"default_ai_cli,omitempty"`
	LastSyncTime      string          `json:"last_sync_time,omitempty"`
	AutoSync          bool            `json:"auto_sync,omitempty"`
	SyncDir           string          `json:"sync_dir,omitempty"`
	LocalStorageDir   string          `json:"local_storage_dir,omitempty"`
	SignatureOverride string          `json:"signature_override,omitempty"`
	Accounts          []AccountConfig `json:"accounts,omitempty"`
	ActiveAccount     string          `json:"active_account,omitempty"`
	Debug             bool            `json:"debug,omitempty"`
}

type AccountConfig struct {
	Email        string `json:"email"`
	Provider     string `json:"provider"`
	Password     string `json:"password"`
	FromName     string `json:"from_name,omitempty"`
	FromEmail    string `json:"from_email,omitempty"`
	ProfileImage string `json:"profile_image,omitempty"`
	Label        string `json:"label,omitempty"`
	Signature    string `json:"signature,omitempty"`
}

// LegacyConfig represents the old config format
type LegacyConfig struct {
	EmailProvider string `json:"emailProvider"`
	AppPassword   string `json:"appPassword"`
	FromEmail     string `json:"fromEmail"`
}

// IsDebugMode returns true if debug mode is enabled via environment variable or config
func IsDebugMode() bool {
	// Check environment variable first
	if os.Getenv("MAILOS_DEBUG") == "true" {
		return true
	}

	// Check config file
	config, err := LoadConfig()
	if err == nil && config.Debug {
		return true
	}

	return false
}

// DebugPrintf prints debug messages only if debug mode is enabled
func DebugPrintf(format string, args ...interface{}) {
	if IsDebugMode() {
		fmt.Printf(format, args...)
	}
}

func GetConfigPath() (string, error) {
	// First check for local .email/config.json in current directory
	localConfig := filepath.Join(".email", "config.json")
	if _, err := os.Stat(localConfig); err == nil {
		// Local config exists, use it
		absPath, _ := filepath.Abs(localConfig)
		return absPath, nil
	}

	// Fall back to global config in home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".email", "config.json"), nil
}

func LoadConfig() (*Config, error) {
	// Ensure email directories exist when loading config
	if err := EnsureEmailDirectories(); err != nil {
		// Don't fail loading config, just warn
		fmt.Printf("Note: Could not create email directories: %v\n", err)
	}

	// First check for local config
	localConfig := filepath.Join(".email", "config.json")
	if _, err := os.Stat(localConfig); err == nil {
		// Local config exists, load it with inheritance from global
		return LoadConfigWithInheritance()
	}

	// No local config, load global config only
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	globalConfigPath := filepath.Join(homeDir, ".email", "config.json")
	return loadConfigFromPath(globalConfigPath)
}

// LoadConfigWithInheritance loads local config and inherits missing fields from global
func LoadConfigWithInheritance() (*Config, error) {
	// Load local config
	localConfigPath := filepath.Join(".email", "config.json")
	localConfig, _ := loadConfigFromPath(localConfigPath) // Ignore error, might just be partial config

	// If local config is nil or invalid, try to load global
	if localConfig == nil || localConfig.Provider == "" {
		// Create a new config to populate
		if localConfig == nil {
			localConfig = &Config{}
		}

		// Load global config
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %v", err)
		}
		globalConfigPath := filepath.Join(homeDir, ".email", "config.json")
		globalConfig, _ := loadConfigFromPath(globalConfigPath) // Ignore error, might not exist

		if globalConfig == nil {
			return nil, fmt.Errorf("no valid configuration found")
		}

		// Copy all fields from global
		*localConfig = *globalConfig

		// Now reload the local config to get any local overrides
		localData, err := os.ReadFile(localConfigPath)
		if err == nil {
			var localOverrides Config
			if json.Unmarshal(localData, &localOverrides) == nil {
				// Apply local overrides
				if localOverrides.FromEmail != "" {
					localConfig.FromEmail = localOverrides.FromEmail
				}
				if localOverrides.FromName != "" {
					localConfig.FromName = localOverrides.FromName
				}
				if localOverrides.ProfileImage != "" {
					localConfig.ProfileImage = localOverrides.ProfileImage
				}
				if localOverrides.DefaultAICLI != "" {
					localConfig.DefaultAICLI = localOverrides.DefaultAICLI
				}
				if localOverrides.SignatureOverride != "" {
					localConfig.SignatureOverride = localOverrides.SignatureOverride
				}
			}
		}
	}

	return localConfig, nil
}

// loadConfigFromPath loads config from a specific path
func loadConfigFromPath(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Try to unmarshal as new format first
	var config Config
	if err := json.Unmarshal(data, &config); err == nil {
		// Config is valid if it has at least one field set
		if config.Provider != "" || config.Email != "" || config.FromEmail != "" || config.FromName != "" {
			return &config, nil
		}
	}

	// Try legacy format
	var legacy LegacyConfig
	if err := json.Unmarshal(data, &legacy); err == nil && legacy.EmailProvider != "" {
		// Convert legacy format to new format
		provider := "gmail" // Default to gmail
		if legacy.EmailProvider == "fastgmail" {
			provider = "gmail"
		}

		config = Config{
			Provider: provider,
			Email:    legacy.FromEmail,
			Password: legacy.AppPassword,
			FromName: "", // Can be derived from email
		}
		return &config, nil
	}

	// If we get here, the file doesn't match any known format
	return nil, fmt.Errorf("invalid config format")
}

func SaveConfig(config *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create .email directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	// Marshal config with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Write config file with restricted permissions
	return os.WriteFile(configPath, data, 0600)
}

func ConfigExists() bool {
	configPath, err := GetConfigPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(configPath)
	return err == nil
}

// GlobalConfigExists checks if global config exists specifically
func GlobalConfigExists() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	globalConfigPath := filepath.Join(homeDir, ".email", "config.json")
	_, err = os.Stat(globalConfigPath)
	return err == nil
}

// EnsureGitIgnore ensures that .email is added to .gitignore in the current directory
func EnsureGitIgnore() error {
	gitignorePath := ".gitignore"

	// Check if .gitignore exists
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// .gitignore doesn't exist, create it with .email entry
		return os.WriteFile(gitignorePath, []byte(".email/\n"), 0644)
	}

	// Check if .email is already in .gitignore
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == ".email" || trimmed == ".email/" || trimmed == "/.email" || trimmed == "/.email/" {
			// Already in .gitignore
			return nil
		}
	}

	// Add .email to .gitignore
	file, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Add newline if file doesn't end with one
	if len(content) > 0 && content[len(content)-1] != '\n' {
		if _, err := file.WriteString("\n"); err != nil {
			return err
		}
	}

	// Add .email entry
	if _, err := file.WriteString(".email/\n"); err != nil {
		return err
	}

	return nil
}

// LoadConfigFromPath loads config from a specific path
func LoadConfigFromPath(configPath string) (*Config, error) {
	return loadConfigFromPath(configPath)
}

// SaveConfigToPath saves config to a specific path
func SaveConfigToPath(config *Config, configPath string) error {
	// Create directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	// If creating a local .email folder, add it to .gitignore
	if strings.HasPrefix(configPath, ".email/") || strings.HasPrefix(configPath, "./.email/") {
		if err := EnsureGitIgnore(); err != nil {
			// Don't fail the operation, just warn
			fmt.Printf("Note: Could not update .gitignore: %v\n", err)
		}
	}

	// Marshal config with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Write config file with restricted permissions
	return os.WriteFile(configPath, data, 0600)
}

// GetConfigLocation returns a string describing where the config is loaded from
func GetConfigLocation() string {
	localConfig := filepath.Join(".email", "config.json")
	if _, err := os.Stat(localConfig); err == nil {
		absPath, _ := filepath.Abs(localConfig)
		return "local: " + absPath
	}

	homeDir, _ := os.UserHomeDir()
	return "global: " + filepath.Join(homeDir, ".email", "config.json")
}

// GetAllAccounts returns all available email accounts from home directory config only
// This includes provider main accounts and sub-email addresses
func GetAllAccounts(config *Config) []AccountConfig {
	accounts := []AccountConfig{}
	accountMap := make(map[string]AccountConfig) // Use map to avoid duplicates

	// Always load accounts from home directory config (not local)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If can't get home dir, use the provided config as fallback
		if config != nil && config.Email != "" {
			return []AccountConfig{{
				Email:        config.Email,
				Provider:     config.Provider,
				Password:     config.Password,
				FromName:     config.FromName,
				FromEmail:    config.FromEmail,
				ProfileImage: config.ProfileImage,
				Label:        "Current",
			}}
		}
		return accounts
	}

	globalConfigPath := filepath.Join(homeDir, ".email", "config.json")
	globalConfig, err := loadConfigFromPath(globalConfigPath)
	if err != nil || globalConfig == nil {
		// If no global config, use the provided config as fallback
		if config != nil && config.Email != "" {
			return []AccountConfig{{
				Email:        config.Email,
				Provider:     config.Provider,
				Password:     config.Password,
				FromName:     config.FromName,
				FromEmail:    config.FromEmail,
				ProfileImage: config.ProfileImage,
				Label:        "Current",
			}}
		}
		return accounts
	}

	// Group accounts by provider to identify main vs sub-emails
	providerGroups := make(map[string][]AccountConfig)

	// Add main account as provider main account
	if globalConfig.Email != "" {
		mainAcc := AccountConfig{
			Email:        globalConfig.Email,
			Provider:     globalConfig.Provider,
			Password:     globalConfig.Password,
			FromName:     globalConfig.FromName,
			FromEmail:    globalConfig.FromEmail,
			ProfileImage: globalConfig.ProfileImage,
			Label:        "Primary",
		}
		providerGroups[globalConfig.Provider] = append(providerGroups[globalConfig.Provider], mainAcc)
		accountMap[globalConfig.Email] = mainAcc
	}

	// Process additional accounts from accounts array
	for _, acc := range globalConfig.Accounts {
		// Skip if already exists (avoid duplicates)
		if _, exists := accountMap[acc.Email]; exists {
			continue
		}

		// Inherit password and provider from main account if not specified
		if acc.Password == "" && acc.Provider == globalConfig.Provider {
			acc.Password = globalConfig.Password
		}
		if acc.Provider == "" {
			acc.Provider = globalConfig.Provider
		}

		// Set label based on whether it's same provider as main or different
		if acc.Label == "" {
			if acc.Provider == globalConfig.Provider {
				acc.Label = "Sub-email"
			} else {
				acc.Label = "Account"
			}
		}

		providerGroups[acc.Provider] = append(providerGroups[acc.Provider], acc)
		accountMap[acc.Email] = acc
	}

	// Add from email as sub-email if different from main email
	if globalConfig.FromEmail != "" && globalConfig.FromEmail != globalConfig.Email {
		if _, exists := accountMap[globalConfig.FromEmail]; !exists {
			fromAcc := AccountConfig{
				Email:        globalConfig.FromEmail,
				Provider:     globalConfig.Provider,
				Password:     globalConfig.Password,
				FromName:     globalConfig.FromName,
				FromEmail:    globalConfig.FromEmail,
				ProfileImage: globalConfig.ProfileImage,
				Label:        "Sub-email",
			}
			providerGroups[globalConfig.Provider] = append(providerGroups[globalConfig.Provider], fromAcc)
			accountMap[globalConfig.FromEmail] = fromAcc
		}
	}

	// Sort accounts: Primary first, then by provider groups
	var primaryAccount *AccountConfig
	var providerAccounts []AccountConfig

	// Find primary account
	for _, acc := range accountMap {
		if acc.Email == globalConfig.Email || acc.Label == "Primary" {
			primaryAccount = &acc
			break
		}
	}

	// Add primary account first
	if primaryAccount != nil {
		accounts = append(accounts, *primaryAccount)
	}

	// Add sub-emails from the same provider as primary
	if primaryAccount != nil {
		primaryProviderAccounts := providerGroups[primaryAccount.Provider]
		for _, acc := range primaryProviderAccounts {
			if acc.Email != primaryAccount.Email { // Skip primary, already added
				providerAccounts = append(providerAccounts, acc)
			}
		}
	}

	// Add accounts from other providers
	for provider, providerAccs := range providerGroups {
		if primaryAccount != nil && provider == primaryAccount.Provider {
			continue // Already processed above
		}
		for _, acc := range providerAccs {
			providerAccounts = append(providerAccounts, acc)
		}
	}

	// Add all provider accounts to final list
	accounts = append(accounts, providerAccounts...)

	return accounts
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// LoadAccountConfig loads configuration for a specific email account from home directory
func LoadAccountConfig(accountEmail string) (*Config, error) {
	if accountEmail == "" {
		// Check if there's an active account set in the config
		config, err := LoadConfig()
		if err != nil {
			return nil, err
		}
		
		// If active account is set, use it
		if config.ActiveAccount != "" {
			return LoadAccountConfig(config.ActiveAccount)
		}
		
		// Otherwise return the default config
		return config, nil
	}

	// Load home directory config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	globalConfigPath := filepath.Join(homeDir, ".email", "config.json")
	globalConfig, err := loadConfigFromPath(globalConfigPath)
	if err != nil || globalConfig == nil {
		return nil, fmt.Errorf("failed to load configuration from home directory: %v", err)
	}

	// Get all accounts
	accounts := GetAllAccounts(globalConfig)

	// Find the requested account - first try exact match
	for _, acc := range accounts {
		if acc.Email == accountEmail {
			// Create a new config with the selected account
			config := &Config{
				Provider:          acc.Provider,
				Email:             acc.Email,
				Password:          acc.Password,
				FromName:          acc.FromName,
				FromEmail:         acc.FromEmail,
				ProfileImage:      acc.ProfileImage,
				SignatureOverride: acc.Signature,
				LicenseKey:        globalConfig.LicenseKey,
				DefaultAICLI:      globalConfig.DefaultAICLI,
				ActiveAccount:     acc.Email,
				Accounts:          globalConfig.Accounts,
			}

			// If account doesn't have all fields, inherit from global config
			if config.Provider == "" {
				config.Provider = globalConfig.Provider
			}
			if config.Password == "" {
				config.Password = globalConfig.Password
			}
			if config.FromEmail == "" {
				config.FromEmail = acc.Email
			}
			
			// For secondary accounts with same provider as primary, use primary email for SMTP auth
			// but keep the secondary email for the "from" field
			if acc.Provider == globalConfig.Provider && acc.Email != globalConfig.Email {
				// This is a secondary account/alias - use primary account for SMTP authentication
				config.Email = globalConfig.Email
				config.FromEmail = acc.Email
			}

			return config, nil
		}
	}

	// No exact match found - try domain matching to find authentication account
	requestedParts := strings.Split(accountEmail, "@")
	if len(requestedParts) == 2 {
		requestedDomain := requestedParts[1]
		
		// Find any configured account with the same domain for authentication
		for _, acc := range accounts {
			accParts := strings.Split(acc.Email, "@")
			if len(accParts) == 2 && accParts[1] == requestedDomain {
				// Found a configured account with the same domain - use it for authentication
				config := &Config{
					Provider:          acc.Provider,
					Email:             acc.Email, // Use configured account for SMTP auth
					Password:          acc.Password,
					FromName:          acc.FromName,
					FromEmail:         accountEmail, // Send from the requested email
					ProfileImage:      acc.ProfileImage,
					SignatureOverride: acc.Signature,
					LicenseKey:        globalConfig.LicenseKey,
					DefaultAICLI:      globalConfig.DefaultAICLI,
					ActiveAccount:     accountEmail,
					Accounts:          globalConfig.Accounts,
				}

				// If account doesn't have all fields, inherit from global config
				if config.Provider == "" {
					config.Provider = globalConfig.Provider
				}
				if config.Password == "" {
					config.Password = globalConfig.Password
				}

				return config, nil
			}
		}
		
		// If no domain match found, try using the primary account as fallback
		// This allows any email to be attempted with primary account credentials
		if globalConfig.Email != "" {
			config := &Config{
				Provider:          globalConfig.Provider,
				Email:             globalConfig.Email, // Use primary account for SMTP auth
				Password:          globalConfig.Password,
				FromName:          globalConfig.FromName,
				FromEmail:         accountEmail, // Send from the requested email
				ProfileImage:      globalConfig.ProfileImage,
				SignatureOverride: globalConfig.SignatureOverride,
				LicenseKey:        globalConfig.LicenseKey,
				DefaultAICLI:      globalConfig.DefaultAICLI,
				ActiveAccount:     accountEmail,
				Accounts:          globalConfig.Accounts,
			}
			
			return config, nil
		}
	}

	return nil, fmt.Errorf("no configured accounts available for authentication")
}

// SwitchAccount switches the active email account for sending
func SwitchAccount(config *Config, accountEmail string) error {
	accounts := GetAllAccounts(config)

	for _, acc := range accounts {
		if acc.Email == accountEmail {
			// Update FromEmail to send as the selected account
			config.FromEmail = acc.Email

			// Also update FromName if the account has one
			if acc.FromName != "" {
				config.FromName = acc.FromName
			}

			// Set as active account
			config.ActiveAccount = accountEmail

			// Don't save to file - this is session-only
			// The change persists in memory for the current session
			return nil
		}
	}

	return fmt.Errorf("account %s not found", accountEmail)
}

// AddAccount adds a new email account to the config
func AddAccount(config *Config, account AccountConfig) error {
	// Check if account already exists
	for i, acc := range config.Accounts {
		if acc.Email == account.Email {
			// Update existing account
			config.Accounts[i] = account
			return SaveConfig(config)
		}
	}

	// Add new account
	config.Accounts = append(config.Accounts, account)
	return SaveConfig(config)
}

// SetAccountSignature sets the signature for a specific account
func SetAccountSignature(email, signature string) error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Check if this is the primary account
	if config.Email == email {
		config.SignatureOverride = signature
		return SaveConfig(config)
	}

	// Check if this is a secondary account
	for i, acc := range config.Accounts {
		if acc.Email == email {
			config.Accounts[i].Signature = signature
			return SaveConfig(config)
		}
	}

	return fmt.Errorf("account %s not found", email)
}

// AddNewAccount prompts user for account details and adds it to the global config
func AddNewAccount(email string) error {
	// Load current global config
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load current config: %v", err)
	}

	// Check if account already exists
	accounts := GetAllAccounts(config)
	for _, acc := range accounts {
		if acc.Email == email {
			return fmt.Errorf("account %s already exists", email)
		}
	}

	fmt.Printf("\nSetting up account: %s\n", email)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Determine provider from email domain with detection details
	provider, detected, method := detectProviderFromEmail(email)
	if detected {
		fmt.Printf("✓ Detected provider: %s (via %s)\n", GetProviderName(provider), method)
	} else {
		fmt.Printf("? Using default provider: %s (unable to detect from domain/MX records)\n", GetProviderName(provider))
	}
	
	// For custom domains, confirm the provider
	domain := strings.ToLower(strings.Split(email, "@")[1])
	if !strings.Contains(domain, "gmail.com") && !strings.Contains(domain, "googlemail.com") && 
	   !strings.Contains(domain, "fastmail.") && !strings.Contains(domain, "fm.") &&
	   !strings.Contains(domain, "outlook.") && !strings.Contains(domain, "hotmail.") && 
	   !strings.Contains(domain, "live.") && !strings.Contains(domain, "yahoo.") && 
	   !strings.Contains(domain, "zoho.") {
		fmt.Printf("Is this correct? If not, available providers are:\n")
		fmt.Printf("  1. gmail (Google Gmail/G Suite)\n")
		fmt.Printf("  2. fastmail (Fastmail/custom domains)\n")
		fmt.Printf("  3. outlook (Microsoft Outlook/Hotmail)\n")
		fmt.Printf("  4. yahoo (Yahoo Mail)\n")
		fmt.Printf("  5. zoho (Zoho Mail)\n")
		fmt.Printf("Enter provider name or press Enter to use '%s': ", provider)
		
		var userProvider string
		fmt.Scanln(&userProvider)
		
		if userProvider != "" {
			switch strings.ToLower(userProvider) {
			case "gmail", "google":
				provider = "gmail"
			case "fastmail", "fastmail.com":
				provider = "fastmail"
			case "outlook", "microsoft", "hotmail":
				provider = "outlook"
			case "yahoo":
				provider = "yahoo"
			case "zoho":
				provider = "zoho"
			default:
				fmt.Printf("Unknown provider '%s', using detected provider: %s\n", userProvider, provider)
			}
		}
	}

	// Prompt for app password
	fmt.Printf("\nFor %s, you need an app-specific password.\n", provider)
	fmt.Print("Enter app password: ")
	var password string
	fmt.Scanln(&password)

	if password == "" {
		return fmt.Errorf("password is required")
	}

	// Create account config
	newAccount := AccountConfig{
		Email:    email,
		Provider: provider,
		Password: password,
		Label:    "Secondary",
	}

	// Add account to config
	if err := AddAccount(config, newAccount); err != nil {
		return fmt.Errorf("failed to save account: %v", err)
	}

	fmt.Printf("✓ Successfully added account: %s\n", email)
	return nil
}

// AddNewAccountWithProvider prompts user for account details and adds it to the global config with specified provider
func AddNewAccountWithProvider(email, provider string, useExistingCredentials bool) error {
	// Load global config explicitly (not local config)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}
	globalConfigPath := filepath.Join(homeDir, ".email", "config.json")
	config, err := loadConfigFromPath(globalConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load global config: %v", err)
	}

	// Check if account already exists
	accounts := GetAllAccounts(config)
	for _, acc := range accounts {
		if acc.Email == email {
			return fmt.Errorf("account %s already exists", email)
		}
	}

	fmt.Printf("\nSetting up account: %s\n", email)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Use specified provider or guess from email domain
	if provider == "" {
		var detected bool
		var method string
		provider, detected, method = detectProviderFromEmail(email)
		if detected {
			fmt.Printf("✓ Detected provider: %s (via %s)\n", GetProviderName(provider), method)
		} else {
			fmt.Printf("? Using default provider: %s (unable to detect from domain/MX records)\n", GetProviderName(provider))
		}
	} else {
		// Validate the specified provider
		switch strings.ToLower(provider) {
		case "gmail", "google":
			provider = ProviderGmail
		case "fastmail":
			provider = ProviderFastmail
		case "outlook", "microsoft":
			provider = ProviderOutlook
		case "yahoo":
			provider = ProviderYahoo
		case "zoho":
			provider = ProviderZoho
		default:
			return fmt.Errorf("unsupported provider: %s. Supported providers: gmail, fastmail, outlook, yahoo, zoho", provider)
		}
		fmt.Printf("Using specified provider: %s\n", provider)
	}

	// Handle credentials based on flag
	var password string
	var existingFromName string
	var existingFromEmail string
	var existingProfileImage string
	
	if useExistingCredentials {
		// Check if we already have credentials for this provider
		// Check primary account
		if config.Provider == provider && config.Password != "" {
			password = config.Password
			existingFromName = config.FromName
			existingFromEmail = config.FromEmail
			existingProfileImage = config.ProfileImage
			fmt.Printf("✓ Using existing %s credentials from primary account\n", provider)
		} else {
			// Check secondary accounts
			for _, acc := range config.Accounts {
				if acc.Provider == provider && acc.Password != "" {
					password = acc.Password
					existingFromName = acc.FromName
					existingFromEmail = acc.FromEmail
					existingProfileImage = acc.ProfileImage
					fmt.Printf("✓ Using existing %s credentials from account %s\n", provider, acc.Email)
					break
				}
			}
		}
		
		// If no existing credentials found, fall back to prompting
		if password == "" {
			fmt.Printf("No existing %s credentials found. Please provide new credentials.\n", provider)
		}
	}
	
	// Prompt for credentials if not using existing or no existing found
	if password == "" {
		fmt.Printf("For %s, you need an app-specific password.\n", provider)
		fmt.Print("Enter app password: ")
		fmt.Scanln(&password)
		if password == "" {
			return fmt.Errorf("password is required")
		}
	}

	// Create new account config
	newAccount := AccountConfig{
		Email:        email,
		Provider:     provider,
		Password:     password,
		FromName:     existingFromName,
		FromEmail:    existingFromEmail,
		ProfileImage: existingProfileImage,
	}

	// Add account to global config explicitly
	// Check if account already exists
	for i, acc := range config.Accounts {
		if acc.Email == newAccount.Email {
			// Update existing account
			config.Accounts[i] = newAccount
			if err := SaveConfigToPath(config, globalConfigPath); err != nil {
				return fmt.Errorf("failed to save account: %v", err)
			}
			return nil
		}
	}

	// Add new account
	config.Accounts = append(config.Accounts, newAccount)
	if err := SaveConfigToPath(config, globalConfigPath); err != nil {
		return fmt.Errorf("failed to save account: %v", err)
	}
	fmt.Printf("✓ Successfully added account: %s\n", email)
	return nil
}

// detectProviderFromMX attempts to determine the email provider from MX records
func detectProviderFromMX(domain string) (string, bool) {
	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		return "", false
	}

	for _, mx := range mxRecords {
		mxHost := strings.ToLower(mx.Host)
		
		switch {
		case strings.Contains(mxHost, "gmail.com") || strings.Contains(mxHost, "google.com") || strings.Contains(mxHost, "googlemail.com"):
			return ProviderGmail, true
		case strings.Contains(mxHost, "fastmail.com") || strings.Contains(mxHost, "messagingengine.com"):
			return ProviderFastmail, true
		case strings.Contains(mxHost, "outlook.com") || strings.Contains(mxHost, "office365.com") || strings.Contains(mxHost, "protection.outlook.com"):
			return ProviderOutlook, true
		case strings.Contains(mxHost, "yahoo.com") || strings.Contains(mxHost, "yahoodns.net"):
			return ProviderYahoo, true
		case strings.Contains(mxHost, "zoho.com") || strings.Contains(mxHost, "zohomail.com"):
			return ProviderZoho, true
		}
	}
	
	return "", false
}

// guessProviderFromEmail attempts to determine the email provider from the email domain
func guessProviderFromEmail(email string) string {
	domain := strings.ToLower(strings.Split(email, "@")[1])
	
	// First try direct domain matching for known providers
	switch {
	case strings.Contains(domain, "gmail.com") || strings.Contains(domain, "googlemail.com"):
		return ProviderGmail
	case strings.Contains(domain, "fastmail.") || strings.Contains(domain, "fm."):
		return ProviderFastmail
	case strings.Contains(domain, "outlook.") || strings.Contains(domain, "hotmail.") || strings.Contains(domain, "live."):
		return ProviderOutlook
	case strings.Contains(domain, "yahoo."):
		return ProviderYahoo
	case strings.Contains(domain, "zoho."):
		return ProviderZoho
	}
	
	// For custom domains, try MX record detection
	if provider, detected := detectProviderFromMX(domain); detected {
		return provider
	}
	
	// Default to fastmail since it works well with custom domains
	return ProviderFastmail
}

// detectProviderFromEmail detects the email provider and returns both the provider and detection method
func detectProviderFromEmail(email string) (provider string, detected bool, method string) {
	domain := strings.ToLower(strings.Split(email, "@")[1])
	
	// First try direct domain matching for known providers
	switch {
	case strings.Contains(domain, "gmail.com") || strings.Contains(domain, "googlemail.com"):
		return ProviderGmail, true, "domain"
	case strings.Contains(domain, "fastmail.") || strings.Contains(domain, "fm."):
		return ProviderFastmail, true, "domain"
	case strings.Contains(domain, "outlook.") || strings.Contains(domain, "hotmail.") || strings.Contains(domain, "live."):
		return ProviderOutlook, true, "domain"
	case strings.Contains(domain, "yahoo."):
		return ProviderYahoo, true, "domain"
	case strings.Contains(domain, "zoho."):
		return ProviderZoho, true, "domain"
	}
	
	// For custom domains, try MX record detection
	if provider, detected := detectProviderFromMX(domain); detected {
		return provider, true, "MX record"
	}
	
	// Default to fastmail since it works well with custom domains
	return ProviderFastmail, false, "default"
}

// SetLocalAccountPreference sets the preferred account for the current local directory
// This creates or updates a local .email/config.json with the active_account setting
func SetLocalAccountPreference(accountEmail string) error {
	// First validate that the account exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	globalConfigPath := filepath.Join(homeDir, ".email", "config.json")
	globalConfig, err := loadConfigFromPath(globalConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load global configuration: %v", err)
	}

	// Verify account exists
	accountFound := false
	var selectedAccount AccountConfig
	accounts := GetAllAccounts(globalConfig)
	for _, acc := range accounts {
		if acc.Email == accountEmail {
			accountFound = true
			selectedAccount = acc
			break
		}
	}

	if !accountFound {
		return fmt.Errorf("account %s not found", accountEmail)
	}

	// Create local .email directory if it doesn't exist
	localConfigDir := ".email"
	if err := os.MkdirAll(localConfigDir, 0700); err != nil {
		return fmt.Errorf("failed to create local .email directory: %v", err)
	}

	// Load existing local config or create new one
	localConfigPath := filepath.Join(localConfigDir, "config.json")
	var localConfig Config

	// Try to load existing local config
	existingData, err := os.ReadFile(localConfigPath)
	if err == nil {
		// Parse existing config, ignore errors to start fresh if corrupted
		json.Unmarshal(existingData, &localConfig)
	}

	// Update the local config with the selected account as active
	localConfig.ActiveAccount = accountEmail
	localConfig.FromEmail = selectedAccount.FromEmail
	if localConfig.FromEmail == "" {
		localConfig.FromEmail = accountEmail
	}
	localConfig.FromName = selectedAccount.FromName

	// Save the local config
	data, err := json.MarshalIndent(localConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal local config: %v", err)
	}

	if err := os.WriteFile(localConfigPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write local config: %v", err)
	}

	// Ensure .gitignore includes .email
	if err := EnsureGitIgnore(); err != nil {
		// Don't fail, just warn
		fmt.Printf("Note: Could not update .gitignore: %v\n", err)
	}

	return nil
}

// GetLocalAccountPreference returns the locally configured account preference if it exists
func GetLocalAccountPreference() string {
	localConfigPath := filepath.Join(".email", "config.json")
	data, err := os.ReadFile(localConfigPath)
	if err != nil {
		return ""
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return ""
	}

	// Return active account if set
	if config.ActiveAccount != "" {
		return config.ActiveAccount
	}

	// Fall back to FromEmail if set
	if config.FromEmail != "" {
		return config.FromEmail
	}

	return ""
}

// GetSMTPSettings returns SMTP configuration for the given provider
func (c *Config) GetSMTPSettings() (host string, port int, useTLS bool, useSSL bool, err error) {
	provider, exists := Providers[c.Provider]
	if !exists {
		return "", 0, false, false, fmt.Errorf("unknown provider: %s", c.Provider)
	}
	return provider.SMTPHost, provider.SMTPPort, provider.SMTPUseTLS, provider.SMTPUseSSL, nil
}

// GetIMAPSettings returns IMAP configuration for the given provider
func (c *Config) GetIMAPSettings() (host string, port int, err error) {
	provider, exists := Providers[c.Provider]
	if !exists {
		return "", 0, fmt.Errorf("unknown provider: %s", c.Provider)
	}
	return provider.IMAPHost, provider.IMAPPort, nil
}

// GetEmailStorageDir returns the base directory for email storage (.email folder)
func GetEmailStorageDir() (string, error) {
	// Check for local .email directory first
	localDir := ".email"
	if info, err := os.Stat(localDir); err == nil && info.IsDir() {
		return filepath.Abs(localDir)
	}

	// Fall back to home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".email"), nil
}

// GetSentDir returns the path to the sent emails directory
func GetSentDir() (string, error) {
	baseDir, err := GetEmailStorageDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, "sent"), nil
}

// GetReceivedDir returns the path to the received emails directory
func GetReceivedDir() (string, error) {
	baseDir, err := GetEmailStorageDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, "received"), nil
}

// GetDraftsDir returns the path to the drafts directory
func GetDraftsDir() (string, error) {
	baseDir, err := GetEmailStorageDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, "drafts"), nil
}

// EnsureEmailDirectories creates the necessary email storage directories
func EnsureEmailDirectories() error {
	baseDir, err := GetEmailStorageDir()
	if err != nil {
		return err
	}

	// Create main .email directory
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return err
	}

	// Create subdirectories
	subdirs := []string{"sent", "received", "drafts"}
	for _, subdir := range subdirs {
		path := filepath.Join(baseDir, subdir)
		if err := os.MkdirAll(path, 0700); err != nil {
			return err
		}
	}

	// Ensure .gitignore is updated
	if strings.HasPrefix(baseDir, ".email") || strings.Contains(baseDir, "/.email") {
		if err := EnsureGitIgnore(); err != nil {
			// Don't fail, just warn
			fmt.Printf("Note: Could not update .gitignore: %v\n", err)
		}
	}

	return nil
}

// ConfigureOptions holds command-line options for configuration
type ConfigureOptions struct {
	Email    string
	Provider string
	Name     string
	From     string
	AICLI    string
	IsLocal  bool
	Quick    bool
}

// Configure handles email configuration with command-line options
func Configure(opts ConfigureOptions) error {
	if opts.IsLocal && GlobalConfigExists() {
		// For local configurations when global config exists, skip license validation
		return configureWithOptions(opts)
	}
	
	// For global configurations or when no global config exists, ensure full initialization
	if err := EnsureInitializedInteractive(); err != nil {
		return err
	}
	
	return configureWithOptions(opts)
}

// configureWithOptions implements the core configuration logic
func configureWithOptions(opts ConfigureOptions) error {
	if opts.Quick {
		return fmt.Errorf("quick config functionality temporarily disabled")
	}
	
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
	
	// Check if we have command-line options to apply directly (for simple local config changes)
	if (opts.Name != "" || opts.From != "" || opts.AICLI != "") && opts.Provider == "" && opts.Email == "" {
		// Create new local config if it doesn't exist
		if localConfig == nil {
			localConfig = &Config{}
		}
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
			fmt.Println("Changing email provider requires full reconfiguration.")
			return setupConfigWithOptions(opts, true)
		}
		
		// Save the updated local configuration
		return saveLocalConfig(localConfig)
	}
	
	// Create new local configuration or handle interactive mode
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
			fmt.Println("Changing email provider requires full reconfiguration.")
			return setupConfigWithOptions(opts, false)
		}
		
		// Save the updated global configuration
		return SaveConfig(globalConfig)
	}
	
	// Create new global configuration or handle interactive mode
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("              CREATE GLOBAL CONFIGURATION               ")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("This will create your global email configuration")
	fmt.Println("in ~/.email/ that is used by default.")
	fmt.Println()
	
	return setupConfigWithOptions(opts, false)
}

// setupConfigWithOptions creates configuration with command-line options
func setupConfigWithOptions(opts ConfigureOptions, isLocal bool) error {
	// For now, redirect to interactive setup if not all parameters provided
	if opts.Provider == "" || opts.Email == "" {
		return fmt.Errorf("interactive configuration required. Run 'mailos setup' to configure provider, email, and other settings")
	}
	
	// Basic validation
	if !isValidEmail(opts.Email) {
		return fmt.Errorf("invalid email address: %s", opts.Email)
	}
	
	// Get license key from global config if doing local setup
	if isLocal {
		if globalConfig, _ := LoadConfig(); globalConfig != nil && globalConfig.LicenseKey != "" {
			// License key available for future use from global config
		}
	}
	
	// For now, require interactive setup for full configuration
	// This ensures proper provider setup and credential handling
	return fmt.Errorf("interactive configuration required for provider setup and credentials")
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
	// Basic email validation - just check for @ and domain
	parts := strings.Split(email, "@")
	return len(parts) == 2 && len(parts[0]) > 0 && len(parts[1]) > 0 && strings.Contains(parts[1], ".")
}

func CreateReadme() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	readmePath := filepath.Join(homeDir, ".email", "README.md")

	content := `# EmailOS Configuration

This directory contains your email client configuration.

## Files

- **config.json**: Your email account settings and credentials
  - Provider: Your email service provider
  - Email: Your email address
  - Password: Your app-specific password
  - SMTP/IMAP settings: Server configuration

## Security

- The config.json file contains sensitive information
- File permissions are set to 600 (read/write for owner only)
- Never share or commit this file to version control

## Setup

To reconfigure your email client, run:
` + "```bash\nmailos setup\n```" + `

## Supported Providers

- Gmail
- Fastmail
- Zoho Mail
- Outlook/Hotmail
- Yahoo Mail

## App Passwords

Most email providers require app-specific passwords for third-party clients:

1. Enable 2-factor authentication on your email account
2. Generate an app-specific password
3. Use this password instead of your regular account password

## Troubleshooting

If you're having issues:

1. Verify your app password is correct
2. Check that IMAP/SMTP access is enabled in your email settings
3. Some providers may require you to enable "less secure app access"

For more information, visit: ` + GitHubRepo + `
`

	return os.WriteFile(readmePath, []byte(content), 0644)
}
