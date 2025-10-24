package mailos

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

// InteractiveAIProviderSelect presents an interactive menu to select AI provider
func InteractiveAIProviderSelect() (string, error) {
	providers := []string{
		"Claude Code",
		"Claude Code YOLO Mode (skip permissions)",
		"OpenAI Codex",
		"Gemini CLI",
		"OpenCode",
		"None (Manual only)",
		"Configure Email Settings",
	}

	prompt := promptui.Select{
		Label: "Select AI Provider or Action",
		Items: providers,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   "▸ {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "✓ {{ . | green }}",
		},
	}

	index, _, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("selection cancelled: %v", err)
	}

	// Map selection to config value
	switch index {
	case 0:
		return "claude-code", nil
	case 1:
		return "claude-code-yolo", nil
	case 2:
		return "openai-codex", nil
	case 3:
		return "gemini-cli", nil
	case 4:
		return "opencode", nil
	case 5:
		return "none", nil
	case 6:
		return "configure", nil
	default:
		return "", fmt.Errorf("invalid selection")
	}
}

// HandleQueryWithProviderSelection handles a query with optional provider selection
func HandleQueryWithProviderSelection(query string) error {
	// Load current configuration (assumes EnsureInitialized was called)
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Check if AI provider is configured
	if config.DefaultAICLI == "" || config.DefaultAICLI == "none" {
		fmt.Println("No AI CLI provider configured.")
		fmt.Println()
		
		provider, err := InteractiveAIProviderSelect()
		if err != nil {
			return err
		}

		if provider == "configure" {
			return Configure(ConfigureOptions{})
		}

		if provider != "none" {
			// Update configuration with selected provider
			config.DefaultAICLI = provider
			if err := SaveConfig(config); err != nil {
				return fmt.Errorf("failed to save configuration: %v", err)
			}
			fmt.Printf("\n✓ AI Provider set to: %s\n\n", GetAICLIName(provider))
			
			// Now invoke with the query
			return InvokeAIProvider(query)
		}
		
		fmt.Println("No AI provider selected. Cannot process query.")
		return nil
	}

	// AI provider is configured, invoke it
	return InvokeAIProvider(query)
}

// QuickConfigMenu shows a quick configuration menu
func QuickConfigMenu() error {
	config, err := LoadConfig()
	if err != nil {
		return Setup()
	}

	options := []string{
		fmt.Sprintf("Email: %s", config.Email),
		fmt.Sprintf("Display Name: %s", getDisplayName(config.FromName)),
		fmt.Sprintf("Provider: %s", GetProviderName(config.Provider)),
		fmt.Sprintf("AI CLI: %s", GetAICLIName(config.DefaultAICLI)),
		"Full Configuration Menu",
		"Exit",
	}

	prompt := promptui.Select{
		Label: "Quick Configuration",
		Items: options,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "▸ {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "✓ {{ . | green }}",
		},
	}

	index, _, err := prompt.Run()
	if err != nil {
		return err
	}

	switch index {
	case 0: // Change email
		return Setup() // Run full setup for email change
	case 1: // Change display name
		return editDisplayName(config)
	case 2: // Change provider
		return Setup() // Run full setup for provider change
	case 3: // Change AI CLI
		return editAICLI(config)
	case 4: // Full configuration
		return Configure(ConfigureOptions{})
	case 5: // Exit
		return nil
	default:
		return nil
	}
}

func editDisplayName(config *Config) error {
	prompt := promptui.Prompt{
		Label:   "Display Name",
		Default: config.FromName,
	}

	name, err := prompt.Run()
	if err != nil {
		return err
	}

	config.FromName = strings.TrimSpace(name)
	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %v", err)
	}

	fmt.Printf("✓ Display name updated to: %s\n", config.FromName)
	return nil
}

func editAICLI(config *Config) error {
	provider, err := InteractiveAIProviderSelect()
	if err != nil {
		return err
	}

	if provider == "configure" {
		return Configure(ConfigureOptions{})
	}

	config.DefaultAICLI = provider
	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %v", err)
	}

	fmt.Printf("✓ AI Provider updated to: %s\n", GetAICLIName(provider))
	return nil
}

func getDisplayName(name string) string {
	if name == "" {
		return "(not set)"
	}
	return name
}