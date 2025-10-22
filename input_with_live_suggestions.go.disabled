package mailos

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

// LiveInputWithSuggestions provides an input field with live suggestions
func LiveInputWithSuggestions() (string, error) {
	suggestions := GetDefaultAISuggestions()
	
	// Custom validation that shows suggestions on empty input
	validate := func(input string) error {
		// No actual validation errors - we just use this for display
		return nil
	}
	
	// Create templates that show suggestions below
	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . }}",
		Valid:   "{{ . }}",
		Invalid: "{{ . }}",
		Success: "{{ . }}",
	}
	
	prompt := promptui.Prompt{
		Label:     "▸",
		Templates: templates,
		Validate:  validate,
		Default:   "",
	}
	
	// Show hint about suggestions
	fmt.Println("\n💡 Press Enter on empty input for AI suggestions, or type your query:")
	fmt.Println("   Use arrow keys to navigate suggestions when shown")
	fmt.Println()
	
	// Get initial input
	input, err := prompt.Run()
	if err != nil {
		return "", err
	}
	
	// If empty, show suggestions menu
	if strings.TrimSpace(input) == "" {
		return showSuggestionsMenu(suggestions)
	}
	
	return input, nil
}

// showSuggestionsMenu displays the suggestions as a selectable menu
func showSuggestionsMenu(suggestions []AISuggestion) (string, error) {
	items := make([]string, len(suggestions)+1)
	for i, s := range suggestions {
		items[i] = fmt.Sprintf("%s %s", s.Icon, s.Title)
	}
	items[len(suggestions)] = "💬 Type custom query..."
	
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "▸ {{ . | cyan }}",
		Inactive: "  {{ . }}",
		Selected: "✓ {{ . | green }}",
		Help:     "Use arrow keys to navigate, Enter to select",
	}
	
	prompt := promptui.Select{
		Label:     "Select an AI suggestion",
		Items:     items,
		Templates: templates,
		Size:      9,
		HideHelp:  false,
	}
	
	index, _, err := prompt.Run()
	if err != nil {
		return "", err
	}
	
	// If custom query selected
	if index == len(suggestions) {
		customPrompt := promptui.Prompt{
			Label: "Enter your query",
		}
		return customPrompt.Run()
	}
	
	// Return the selected suggestion's command
	return suggestions[index].Command, nil
}

// EnhancedInteractiveModeV2 uses the live input with suggestions
func EnhancedInteractiveModeV2(config *Config) error {
	for {
		// Get input with suggestions
		query, err := LiveInputWithSuggestions()
		if err != nil {
			if err == promptui.ErrInterrupt {
				fmt.Println("\n👋 Goodbye!")
				return nil
			}
			return err
		}
		
		// Handle empty input
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}
		
		// Check for exit commands
		if query == "exit" || query == "quit" || query == "/exit" {
			fmt.Println("\n👋 Goodbye!")
			return nil
		}
		
		// Check for slash commands
		if strings.HasPrefix(query, "/") {
			if query == "/" {
				if err := showCommandMenu(); err != nil {
					if err.Error() == "exit" {
						fmt.Println("\n👋 Goodbye!")
						return nil
					}
				}
				continue
			}
			if err := executeCommand(query); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			continue
		}
		
		// Process as AI query
		if config.DefaultAICLI == "" || config.DefaultAICLI == "none" {
			fmt.Println("\n⚠️  No AI provider configured.")
			fmt.Println("Run 'mailos provider' to set up an AI provider.")
			continue
		}
		
		fmt.Printf("\n🤔 Processing: %s\n\n", query)
		if err := InvokeAIProvider(query); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		
		// Add a separator for clarity
		fmt.Println("\n" + strings.Repeat("─", 60))
	}
}