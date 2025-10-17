package mailos

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// AI provider command mappings
var aiProviderCommands = map[string]string{
	"claude-code":        "claude",
	"claude-code-accept": "claude",
	"claude-code-yolo":   "claude",
	"openai-codex":       "codex",
	"gemini-cli":         "gemini",
	"opencode":           "opencode",
}

// GetAIProviderCommand returns the command for the configured AI provider
func GetAIProviderCommand(provider string) (string, bool) {
	cmd, exists := aiProviderCommands[provider]
	return cmd, exists
}

// InvokeAIProvider invokes the configured AI provider with the given query
func InvokeAIProvider(query string) error {
	return InvokeAIProviderWithMode(query, "")
}

// InvokeAIProviderNonInteractive invokes the AI provider in non-interactive mode (returns text response)
func InvokeAIProviderNonInteractive(query string) (string, error) {
	return InvokeAIProviderNonInteractiveWithSystemPrompt(query, "")
}

// InvokeAIProviderNonInteractiveWithSystemPrompt invokes the AI provider with a custom system prompt
func InvokeAIProviderNonInteractiveWithSystemPrompt(query string, customSystemPrompt string) (string, error) {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load configuration: %v", err)
	}

	// Check if AI provider is configured
	if config.DefaultAICLI == "" || config.DefaultAICLI == "none" {
		return "", fmt.Errorf("no AI CLI provider configured. Run 'mailos setup' or 'mailos configure' to select an AI provider")
	}

	// Get the command for the AI provider
	aiCommand, exists := GetAIProviderCommand(config.DefaultAICLI)
	if !exists {
		return "", fmt.Errorf("unknown AI provider: %s", config.DefaultAICLI)
	}

	// Check if the AI command exists in PATH
	_, err = exec.LookPath(aiCommand)
	if err != nil {
		return "", fmt.Errorf("AI CLI '%s' not found. Please install %s CLI first", aiCommand, GetAICLIName(config.DefaultAICLI))
	}

	// Use custom system prompt if provided, otherwise build default
	var fullQuery string
	if customSystemPrompt != "" {
		fullQuery = fmt.Sprintf("%s\n\n%s", customSystemPrompt, query)
	} else {
		systemMessage := BuildEmailManagerSystemMessage()
		fullQuery = fmt.Sprintf("%s\n\nUser Query: %s", systemMessage, query)
	}

	// Build command for non-interactive mode
	var cmd *exec.Cmd
	switch config.DefaultAICLI {
	case "claude-code", "claude-code-accept", "claude-code-yolo":
		// Claude uses --print flag for non-interactive output
		cmd = exec.Command(aiCommand, "--print", fullQuery)
	case "openai-codex":
		// OpenAI Codex command
		cmd = exec.Command(aiCommand, "exec", "--skip-git-repo-check", fullQuery)
	case "gemini-cli":
		// Gemini CLI command
		cmd = exec.Command(aiCommand, "-p", fullQuery)
	case "opencode":
		// OpenCode command (assuming it supports similar syntax)
		cmd = exec.Command(aiCommand, "--print", fullQuery)
	default:
		// Fallback
		cmd = exec.Command(aiCommand, "--print", fullQuery)
	}

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute AI provider: %v", err)
	}

	return string(output), nil
}

// InvokeAIProviderWithMode invokes the AI provider with a specific mode override
func InvokeAIProviderWithMode(query string, modeOverride string) error {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Check if AI provider is configured
	if config.DefaultAICLI == "" || config.DefaultAICLI == "none" {
		return fmt.Errorf("no AI CLI provider configured. Run 'mailos setup' or 'mailos configure' to select an AI provider")
	}

	// Get the command for the AI provider
	aiCommand, exists := GetAIProviderCommand(config.DefaultAICLI)
	if !exists {
		return fmt.Errorf("unknown AI provider: %s", config.DefaultAICLI)
	}

	// Check if the AI command exists in PATH
	if _, err := exec.LookPath(aiCommand); err != nil {
		return fmt.Errorf("AI CLI '%s' not found. Please install %s CLI first", aiCommand, GetAICLIName(config.DefaultAICLI))
	}

	// Build the system message with documentation
	systemMessage := BuildEmailManagerSystemMessage()

	// Combine system message with user query
	fullQuery := fmt.Sprintf("%s\n\nUser Query: %s", systemMessage, query)

	// Determine the effective mode (use override if provided)
	effectiveMode := config.DefaultAICLI
	if modeOverride != "" {
		effectiveMode = modeOverride
	}

	// Build command based on provider type
	var cmd *exec.Cmd
	switch effectiveMode {
	case "claude-code":
		// Regular Claude command with --print for non-interactive
		cmd = exec.Command(aiCommand, "--print", fullQuery)
	case "claude-code-yolo":
		// YOLO mode with dangerous permissions flag
		cmd = exec.Command(aiCommand, "--dangerously-skip-permissions", "--print", fullQuery)
	case "claude-code-accept":
		// Accept edits mode - automatically accepts file edits
		cmd = exec.Command(aiCommand, "--permission-mode", "acceptEdits", "--print", fullQuery)
	case "openai-codex":
		// OpenAI Codex command
		cmd = exec.Command(aiCommand, "exec", "--skip-git-repo-check", fullQuery)
	case "gemini-cli":
		// Gemini CLI command
		cmd = exec.Command(aiCommand, "-p", fullQuery)
	case "opencode":
		// OpenCode command
		cmd = exec.Command(aiCommand, fullQuery)
	default:
		// Fallback to regular interactive mode
		cmd = exec.Command(aiCommand, fullQuery)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Use the effective mode for display
	displayMode := config.DefaultAICLI
	if modeOverride != "" {
		displayMode = modeOverride
	}
	fmt.Printf("Invoking %s for email management...\n\n", GetAICLIName(displayMode))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute AI provider: %v", err)
	}

	return nil
}

// CheckAIProviderAvailable checks if an AI provider is configured and available
func CheckAIProviderAvailable() (bool, string) {
	config, err := LoadConfig()
	if err != nil {
		return false, ""
	}

	if config.DefaultAICLI == "" || config.DefaultAICLI == "none" {
		return false, ""
	}

	aiCommand, exists := GetAIProviderCommand(config.DefaultAICLI)
	if !exists {
		return false, ""
	}

	if _, err := exec.LookPath(aiCommand); err != nil {
		return false, ""
	}

	return true, config.DefaultAICLI
}

// IsGeneralQuery checks if the arguments represent a general query
func IsGeneralQuery(args []string) bool {
	if len(args) == 0 {
		return false
	}

	// Check if first argument is a known command
	knownCommands := []string{
		"setup", "configure", "config", "template", "send", "read",
		"mark-read", "delete", "unsubscribe", "info", "test",
		"report", "open", "provider",
		"--help", "-h", "--version", "-v",
	}

	firstArg := strings.ToLower(args[0])
	for _, cmd := range knownCommands {
		if firstArg == cmd {
			return false
		}
	}

	// If not a known command, it's a general query
	return true
}

// getEmailManagerPrompt returns the configured email manager prompt or a default
func getEmailManagerPrompt() string {
	return GetEmailManagerPrompt()
}
