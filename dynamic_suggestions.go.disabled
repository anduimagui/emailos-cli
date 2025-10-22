package mailos

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

// DynamicInputWithSuggestions provides an enhanced input experience with dynamic suggestions
func DynamicInputWithSuggestions() (string, error) {
	suggestions := GetDefaultAISuggestions()
	
	// First, try to get input with a custom prompt
	fmt.Println("\n┌─────────────────────────────────────────────────────────┐")
	fmt.Println("│ 💡 AI Assistant Ready                                   │")
	fmt.Println("│                                                         │")
	fmt.Println("│ • Type your question or command                        │")
	fmt.Println("│ • Press Enter on empty input to see suggestions        │")
	fmt.Println("│ • Type '/' for commands, '@' for files                 │")
	fmt.Println("└─────────────────────────────────────────────────────────┘")
	fmt.Println()
	
	// Custom searcher that filters suggestions based on input
	searcher := func(input string, index int) bool {
		if index >= len(suggestions) {
			// The "Type custom query" option should always be visible
			return true
		}
		
		// If input is empty, show all suggestions
		if strings.TrimSpace(input) == "" {
			return true
		}
		
		// Filter suggestions based on input
		suggestion := suggestions[index]
		lowerInput := strings.ToLower(input)
		
		// Check if input matches title or command
		return strings.Contains(strings.ToLower(suggestion.Title), lowerInput) ||
			strings.Contains(strings.ToLower(suggestion.Command), lowerInput)
	}
	
	// Create a searchable prompt
	items := make([]string, len(suggestions)+1)
	for i, s := range suggestions {
		items[i] = fmt.Sprintf("%s %s", s.Icon, s.Title)
	}
	items[len(suggestions)] = "💬 Type your own query..."
	
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "▸ {{ . | cyan }}",
		Inactive: "  {{ . }}",
		Selected: "✓ {{ . | green }}",
		Details: `
--------- Suggestion Details ----------
{{ if lt .Index %d }}{{ with index $.Suggestions .Index }}{{ .Description }}
Command: {{ .Command }}{{ end }}{{ else }}Type a custom query for the AI assistant{{ end }}`,
	}
	
	// Format the details template with the suggestions length
	detailsTemplate := fmt.Sprintf(templates.Details, len(suggestions))
	templates.Details = detailsTemplate
	
	prompt := promptui.Select{
		Label:             "▸ Type to filter or select a suggestion",
		Items:             items,
		Templates:         templates,
		Size:              8,
		Searcher:          searcher,
		StartInSearchMode: true,
		HideSelected:      true,
	}
	
	// Add the suggestions as extra data for the template
	prompt.Stdin = nil // Use default stdin
	
	index, result, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return "", err
		}
		// If they typed something and pressed enter without selecting
		if strings.Contains(err.Error(), "^M") {
			// Extract what they typed from the result
			return strings.TrimSpace(result), nil
		}
		return "", err
	}
	
	// If custom query selected or they typed something
	if index == len(suggestions) {
		// Check if they already typed something in search mode
		if result != "" && result != items[index] {
			// They typed something, use it
			query := strings.TrimSpace(result)
			// Remove the icon and title if present
			for _, item := range items {
				if strings.HasPrefix(query, item) {
					query = ""
					break
				}
			}
			if query != "" {
				return query, nil
			}
		}
		
		// Otherwise prompt for custom input
		customPrompt := promptui.Prompt{
			Label: "Enter your query",
		}
		return customPrompt.Run()
	}
	
	// Return the selected suggestion's command
	return suggestions[index].Command, nil
}

// InteractiveModeWithDynamicSuggestions provides the main loop with dynamic suggestions
func InteractiveModeWithDynamicSuggestions(config *Config) error {
	for {
		// Get input with dynamic suggestions
		query, err := DynamicInputWithSuggestions()
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