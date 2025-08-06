package mailos

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	
	"github.com/charmbracelet/lipgloss"
)

// EnsureInitialized validates that the system is properly configured with a valid license
// This should be called before any command that requires email configuration
func EnsureInitialized() error {
	// Check if config exists
	if !ConfigExists() {
		// Run setup automatically when no config exists
		return Setup()
	}
	
	// Load config to validate license
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}
	
	// Check if license key exists
	if config.LicenseKey == "" {
		// Run setup automatically when no license key exists
		return Setup()
	}
	
	// Validate license using quick validation (with cache)
	lm := GetLicenseManager()
	if err := lm.QuickValidate(config.LicenseKey); err != nil {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("⚠️  LICENSE VALIDATION FAILED")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		fmt.Println(errorStyle.Render("License validation failed"))
		fmt.Println()
		
		// Check if we're in grace period
		if lm.IsInGracePeriod() {
			fmt.Println("📌 You are currently in a grace period.")
			fmt.Println("   Please ensure you have an active internet connection")
			fmt.Println("   to revalidate your license.")
			fmt.Println()
			return nil // Allow operation in grace period
		}
		
		fmt.Println("Your license may have expired or been revoked.")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("1. Check your internet connection and try again")
		fmt.Printf("2. Visit https://%s/checkout to renew your license\n", APP_SITE)
		fmt.Println("3. Contact support if you believe this is an error")
		fmt.Println()
		fmt.Println("Run 'mailos setup' to enter a new license key.")
		return fmt.Errorf("valid license required to continue")
	}
	
	return nil
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