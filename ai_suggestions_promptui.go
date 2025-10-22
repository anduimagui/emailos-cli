package mailos

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

// AISuggestionItem represents an AI suggestion for promptui
type AISuggestionItem struct {
	Title       string
	Command     string
	Description string
	Icon        string
}

// GetDefaultAISuggestionItems returns default AI suggestions for promptui
func GetDefaultAISuggestionItems() []AISuggestionItem {
	return []AISuggestionItem{
		{
			Title:       "Summarize yesterday's emails",
			Command:     "Summarize all emails from yesterday",
			Description: "Get a quick overview of yesterday's messages",
			Icon:        "📊",
		},
		{
			Title:       "Show unread emails",
			Command:     "Show me all unread emails with a brief summary",
			Description: "List and summarize unread messages",
			Icon:        "📬",
		},
		{
			Title:       "Draft a professional reply",
			Command:     "Help me draft a professional reply to the last email",
			Description: "Compose a well-formatted response",
			Icon:        "✍️",
		},
		{
			Title:       "Find important emails",
			Command:     "Find important emails from this week",
			Description: "Identify high-priority messages",
			Icon:        "⭐",
		},
		{
			Title:       "Email statistics",
			Command:     "Show me email statistics for this week",
			Description: "Get insights about your email activity",
			Icon:        "📈",
		},
		{
			Title:       "Schedule follow-ups",
			Command:     "Find emails that need follow-ups",
			Description: "Identify messages requiring responses",
			Icon:        "🔔",
		},
		{
			Title:       "Clean up inbox",
			Command:     "Help me clean up my inbox - find emails to delete",
			Description: "Identify emails that can be removed",
			Icon:        "🧹",
		},
		{
			Title:       "Today's agenda from emails",
			Command:     "Extract today's agenda and tasks from my emails",
			Description: "Find action items and meetings",
			Icon:        "📅",
		},
		{
			Title:       "Type custom query...",
			Command:     "",
			Description: "Enter your own query",
			Icon:        "💬",
		},
	}
}

// ShowAISuggestionsPromptUI displays AI suggestions using promptui
func ShowAISuggestionsPromptUI() (string, error) {
	suggestions := GetDefaultAISuggestionItems()
	
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "▸ {{ .Icon }} {{ .Title | cyan }}",
		Inactive: "  {{ .Icon }} {{ .Title }}",
		Selected: "✓ {{ .Icon }} {{ .Title | green }}",
		Details: `
--------- Suggestion Details ----------
{{ "Description:" | faint }}	{{ .Description }}
{{ "Command:" | faint }}	{{ .Command }}`,
	}

	searcher := func(input string, index int) bool {
		suggestion := suggestions[index]
		name := strings.Replace(strings.ToLower(suggestion.Title), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     "AI Suggestions (type to filter, ↑↓ to navigate, Enter to select)",
		Items:     suggestions,
		Templates: templates,
		Size:      9,
		Searcher:  searcher,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	selected := suggestions[index]
	
	// If "Type custom query" was selected, prompt for input
	if selected.Command == "" {
		inputPrompt := promptui.Prompt{
			Label: "Enter your query",
			Validate: func(input string) error {
				if strings.TrimSpace(input) == "" {
					return fmt.Errorf("query cannot be empty")
				}
				return nil
			},
		}
		return inputPrompt.Run()
	}
	
	return selected.Command, nil
}

// ShowFileAutocompletePromptUI displays file selection using promptui
func ShowFileAutocompletePromptUI(baseDir string, query string) (string, error) {
	// Get file suggestions
	autocomplete := NewFileAutocomplete(baseDir)
	autocomplete.UpdateSuggestions(query)
	
	if len(autocomplete.Suggestions) == 0 {
		return "", fmt.Errorf("no files found matching '%s'", query)
	}
	
	// Convert to display items
	type FileItem struct {
		Path        string
		Display     string
		IsDirectory bool
		Icon        string
	}
	
	items := make([]FileItem, len(autocomplete.Suggestions))
	for i, s := range autocomplete.Suggestions {
		icon := "📄"
		if s.IsDirectory {
			icon = "📁"
		}
		
		display := s.RelPath
		if s.IsDirectory && !strings.HasSuffix(display, "/") {
			display += "/"
		}
		
		items[i] = FileItem{
			Path:        s.RelPath,
			Display:     display,
			IsDirectory: s.IsDirectory,
			Icon:        icon,
		}
	}
	
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "▸ {{ .Icon }} {{ .Display | cyan }}",
		Inactive: "  {{ .Icon }} {{ .Display }}",
		Selected: "✓ {{ .Icon }} {{ .Display | green }}",
	}
	
	searcher := func(input string, index int) bool {
		item := items[index]
		return fuzzyMatch(strings.ToLower(item.Path), strings.ToLower(input))
	}
	
	prompt := promptui.Select{
		Label:     fmt.Sprintf("Files matching '@%s' (type to filter, ↑↓ to navigate)", query),
		Items:     items,
		Templates: templates,
		Size:      15,
		Searcher:  searcher,
	}
	
	index, _, err := prompt.Run()
	if err != nil {
		return "", err
	}
	
	return items[index].Path, nil
}

// InteractiveQueryWithSuggestions shows AI suggestions and handles user input
func InteractiveQueryWithSuggestions(config *Config) error {
	// First check if AI is configured
	if config.DefaultAICLI == "" || config.DefaultAICLI == "none" {
		fmt.Println("\n⚠️  No AI provider configured.")
		setupPrompt := promptui.Select{
			Label: "Would you like to set up an AI provider now?",
			Items: []string{"Yes, set up AI provider", "No, continue without AI"},
		}
		idx, _, err := setupPrompt.Run()
		if err == nil && idx == 0 {
			// return SelectAndConfigureAIProvider()
			fmt.Println("AI provider setup would be called here")
		}
		return nil
	}
	
	// Show AI suggestions
	query, err := ShowAISuggestionsPromptUI()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return nil
		}
		return err
	}
	
	// Process the selected query
	fmt.Printf("\n🤔 Processing: %s\n\n", query)
	return InvokeAIProvider(query)
}