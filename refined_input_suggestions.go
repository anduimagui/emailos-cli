package mailos

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

// RefinedInputWithSuggestions provides a clean input experience with optional suggestions
func RefinedInputWithSuggestions() (string, error) {
	// Show a clean, focused input prompt
	prompt := promptui.Prompt{
		Label:   "▸",
		Default: "",
	}
	
	// Show minimal hint
	fmt.Print("\n💡 ")
	fmt.Println("Enter your query (or press Enter for suggestions):")
	
	// Get input
	input, err := prompt.Run()
	if err != nil {
		return "", err
	}
	
	// If empty, show suggestions in a clean menu
	if strings.TrimSpace(input) == "" {
		return showCleanSuggestionsMenu()
	}
	
	return input, nil
}

// showCleanSuggestionsMenu displays a clean, focused suggestions menu
func showCleanSuggestionsMenu() (string, error) {
	suggestions := GetDefaultAISuggestions()
	
	// Create clean menu items
	items := []struct {
		Name        string
		Value       string
		Description string
	}{
		{"📊 Summarize yesterday's emails", suggestions[0].Command, "Get a quick overview"},
		{"📬 Show unread emails", suggestions[1].Command, "List unread messages"},
		{"✍️  Draft a professional reply", suggestions[2].Command, "Compose a response"},
		{"⭐ Find important emails", suggestions[3].Command, "Identify priority messages"},
		{"📈 Email statistics", suggestions[4].Command, "Get email insights"},
		{"🔔 Schedule follow-ups", suggestions[5].Command, "Find pending responses"},
		{"🧹 Clean up inbox", suggestions[6].Command, "Identify deletable emails"},
		{"📅 Today's agenda from emails", suggestions[7].Command, "Extract tasks and meetings"},
		{"", "", ""},
		{"💬 Type custom query...", "__CUSTOM__", "Enter your own question"},
	}
	
	// Create display items
	displayItems := make([]string, len(items))
	for i, item := range items {
		if item.Name == "" {
			displayItems[i] = "────────────────────────"
		} else {
			displayItems[i] = item.Name
		}
	}
	
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "▸ {{ . | cyan }}",
		Inactive: "  {{ . | dim }}",
		Selected: "{{ . | green }}",
	}
	
	prompt := promptui.Select{
		Label:        "AI Suggestions",
		Items:        displayItems,
		Templates:    templates,
		Size:         10,
		HideSelected: true,
		CursorPos:    0,
	}
	
	index, _, err := prompt.Run()
	if err != nil {
		return "", err
	}
	
	// Handle selection
	selectedItem := items[index]
	if selectedItem.Value == "__CUSTOM__" {
		customPrompt := promptui.Prompt{
			Label: "▸",
		}
		fmt.Println("\n💬 Enter your custom query:")
		return customPrompt.Run()
	}
	
	if selectedItem.Value == "" {
		// Separator was selected, retry
		return showCleanSuggestionsMenu()
	}
	
	return selectedItem.Value, nil
}

// CleanInteractiveMode provides the cleanest interactive experience
func CleanInteractiveMode(config *Config) error {
	for {
		// Get input with refined suggestions
		query, err := RefinedInputWithSuggestions()
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
		
		// Check for @ file references
		if strings.Contains(query, "@") {
			fmt.Println("\n📁 File reference detected. Processing with context...")
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
		
		// Add a clean separator
		fmt.Println("\n" + strings.Repeat("─", 60))
	}
}