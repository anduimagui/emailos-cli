package mailos

import (
	"strings"

	"github.com/manifoldco/promptui"
)

// ReadLineWithPromptUI reads a line with AI suggestions and file autocomplete using promptui
func ReadLineWithPromptUI(prompt string) string {
	// Main menu options
	options := []struct {
		Label string
		Icon  string
		Type  string
	}{
		{"Ask AI (with suggestions)", "🤖", "ai"},
		{"Type custom query", "💬", "custom"},
		{"Browse files (@)", "📁", "files"},
		{"Enter command (/)", "⚙️", "command"},
		{"Cancel", "❌", "cancel"},
	}
	
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "▸ {{ .Icon }} {{ .Label | cyan }}",
		Inactive: "  {{ .Icon }} {{ .Label }}",
		Selected: "✓ {{ .Icon }} {{ .Label | green }}",
	}
	
	selectPrompt := promptui.Select{
		Label:     prompt,
		Items:     options,
		Templates: templates,
		Size:      5,
	}
	
	index, _, err := selectPrompt.Run()
	if err != nil {
		return "__EXIT__"
	}
	
	selected := options[index]
	
	switch selected.Type {
	case "ai":
		// Show AI suggestions
		query, err := ShowAISuggestionsPromptUI()
		if err != nil {
			return ""
		}
		return query
		
	case "custom":
		// Custom input
		inputPrompt := promptui.Prompt{
			Label: "Enter your query",
		}
		result, err := inputPrompt.Run()
		if err != nil {
			return ""
		}
		return result
		
	case "files":
		// File browser
		inputPrompt := promptui.Prompt{
			Label:   "Enter file pattern (leave empty to browse all)",
			Default: "",
		}
		pattern, _ := inputPrompt.Run()
		
		filePath, err := ShowFileAutocompletePromptUI("", pattern)
		if err != nil {
			return "@" + pattern // Return the pattern they typed
		}
		return "@" + filePath
		
	case "command":
		// Command input
		inputPrompt := promptui.Prompt{
			Label:   "Enter command",
			Default: "/",
		}
		result, err := inputPrompt.Run()
		if err != nil {
			return ""
		}
		if !strings.HasPrefix(result, "/") {
			result = "/" + result
		}
		return result
		
	case "cancel":
		return "__EXIT__"
		
	default:
		return ""
	}
}

// HandleFileAutocompleteInteractive handles @ symbol file selection interactively
func HandleFileAutocompleteInteractive(currentInput string) (string, error) {
	// Extract the query after @ symbol
	atIndex := strings.LastIndex(currentInput, "@")
	if atIndex == -1 {
		return currentInput, nil
	}
	
	beforeAt := currentInput[:atIndex]
	query := currentInput[atIndex+1:]
	
	// Show file selector
	selectedPath, err := ShowFileAutocompletePromptUI("", query)
	if err != nil {
		return currentInput, err
	}
	
	// Replace the @query part with the selected path
	return beforeAt + "@" + selectedPath, nil
}

// ProcessInputWithPromptUI processes user input and handles special characters
func ProcessInputWithPromptUI(input string) (string, error) {
	// Check for @ symbol (file autocomplete)
	if strings.Contains(input, "@") {
		return HandleFileAutocompleteInteractive(input)
	}
	
	// Check for / command
	if strings.HasPrefix(input, "/") {
		return input, nil
	}
	
	// Regular input
	return input, nil
}