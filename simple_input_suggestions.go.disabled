package mailos

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

// SimpleInputWithSuggestions provides input with suggestion dropdown
func SimpleInputWithSuggestions() (string, error) {
	// First show the input prompt with hint
	fmt.Println("\n💡 Type your query or press Enter for suggestions")
	
	inputPrompt := promptui.Prompt{
		Label:   "▸",
		Default: "",
	}
	
	// Try to get input
	input, err := inputPrompt.Run()
	if err != nil {
		return "", err
	}
	
	// If user pressed enter without typing (wants suggestions)
	if strings.TrimSpace(input) == "" {
		// Show AI suggestions menu
		suggestions := GetDefaultAISuggestions()
		
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
		}
		
		prompt := promptui.Select{
			Label:     "Select a suggestion or type custom",
			Items:     items,
			Templates: templates,
			Size:      9,
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
	
	// Return the typed input
	return input, nil
}

// EnhancedInteractiveMode provides a better interactive experience
func EnhancedInteractiveMode(config *Config) error {
	// Main loop
	for {
		// Get input or suggestion
		query, err := SimpleInputWithSuggestions()
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