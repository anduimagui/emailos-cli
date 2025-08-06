package mailos

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var APP_SITE = "email-os.com"

type Config struct {
	Provider     string `json:"provider"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	FromName     string `json:"from_name,omitempty"`
	FromEmail    string `json:"from_email,omitempty"`
	LicenseKey   string `json:"license_key,omitempty"`
	DefaultAICLI string `json:"default_ai_cli,omitempty"`
}

// LegacyConfig represents the old config format
type LegacyConfig struct {
	EmailProvider string `json:"emailProvider"`
	AppPassword   string `json:"appPassword"`
	FromEmail     string `json:"fromEmail"`
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
				if localOverrides.DefaultAICLI != "" {
					localConfig.DefaultAICLI = localOverrides.DefaultAICLI
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

For more information, visit: https://github.com/emailos/mailos
`
	
	return os.WriteFile(readmePath, []byte(content), 0644)
}