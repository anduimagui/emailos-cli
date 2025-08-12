package mailos

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

// AISuggestion represents a suggested AI command
type AISuggestion struct {
	Title       string
	Command     string
	Description string
	Icon        string
}

// GetDefaultAISuggestions returns the default AI command suggestions
func GetDefaultAISuggestions() []AISuggestion {
	return []AISuggestion{
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
	}
}

// ShowAISuggestions displays the AI command suggestions
func ShowAISuggestions() {
	suggestions := GetDefaultAISuggestions()
	displaySuggestions := suggestions
	if len(suggestions) > 4 {
		displaySuggestions = suggestions[:4]
	}
	options := CreateAICommandOptions(displaySuggestions)
	
	renderer := NewListRenderer("AI Suggestions (↑↓ to select, Enter to use)", options)
	renderer.CompactMode = true
	renderer.ShowIcons = true
	renderer.MaxDisplay = 4
	
	fmt.Print("\n")
	fmt.Print(renderer.RenderList())
}

// ReadLineWithAISuggestions reads input with AI command suggestions support
func ReadLineWithAISuggestions(prompt string) string {
	// Check if stdin is a terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		// Fallback to regular input
		return ReadLineWithFileAutocomplete(prompt)
	}
	
	// Switch to raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return ReadLineWithFileAutocomplete(prompt)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	
	fmt.Print(prompt)
	
	var buffer []rune
	var cursorPos int
	suggestions := GetDefaultAISuggestions()
	selectedSuggestion := -1
	showingSuggestions := false
	autocomplete := NewFileAutocomplete("")
	var atSymbolPos int = -1
	var escapeSeq []byte
	inEscape := false
	lastEscapeTime := time.Time{}
	
	// Helper to clear suggestions display
	clearSuggestions := func() {
		if showingSuggestions {
			// Create renderer to calculate lines to clear
			displayedSuggestions := suggestions
			if len(suggestions) > 4 {
				displayedSuggestions = suggestions[:4]
			}
			options := CreateAICommandOptions(displayedSuggestions)
			
			renderer := NewListRenderer("AI Suggestions (↑↓ to select, Enter to use)", options)
			renderer.CompactMode = true
			renderer.MaxDisplay = 4
			
			numLines := renderer.GetLineCount() + 1 // +1 for initial newline
			for i := 0; i < numLines; i++ {
				fmt.Print("\033[B\033[2K") // Move down and clear line
			}
			for i := 0; i < numLines; i++ {
				fmt.Print("\033[A") // Move back up
			}
			showingSuggestions = false
		}
	}
	
	// Helper to show suggestions with selection
	displaySuggestions := func() {
		if len(buffer) > 0 {
			// Don't show suggestions if user is typing
			return
		}
		
		// Create list renderer for AI suggestions
		displayedSuggestions := suggestions
		if len(suggestions) > 4 {
			displayedSuggestions = suggestions[:4]
		}
		options := CreateAICommandOptions(displayedSuggestions)
		
		renderer := NewListRenderer("AI Suggestions (↑↓ to select, Enter to use)", options)
		renderer.CompactMode = true
		renderer.ShowIcons = true
		renderer.MaxDisplay = 4
		renderer.SelectedIndex = selectedSuggestion
		
		// Save cursor position
		fmt.Print("\033[s")
		
		// Render the list
		fmt.Print("\n" + renderer.RenderList())
		
		// Restore cursor position
		fmt.Print("\033[u")
		showingSuggestions = true
	}
	
	// Helper to redraw the line
	redraw := func() {
		// Clear current line and redraw input
		fmt.Print("\r\033[K" + prompt + string(buffer))
		
		// Check for @ symbol and update file autocomplete
		if atSymbolPos >= 0 && cursorPos > atSymbolPos {
			query := string(buffer[atSymbolPos+1:cursorPos])
			autocomplete.UpdateSuggestions(query)
			autocomplete.IsActive = true
			clearSuggestions() // Hide AI suggestions when file autocomplete is active
		} else {
			autocomplete.IsActive = false
			if len(buffer) == 0 && selectedSuggestion >= 0 {
				displaySuggestions()
			}
		}
		
		// Render file suggestions if active
		if autocomplete.IsActive {
			fmt.Print("\033[s")
			fmt.Print(autocomplete.RenderSuggestions())
			fmt.Print("\033[u")
		}
		
		// Position cursor correctly
		fmt.Printf("\r%s%s", prompt, string(buffer[:cursorPos]))
	}
	
	// Initial display of suggestions if buffer is empty
	if len(buffer) == 0 {
		displaySuggestions()
	}
	
	reader := bufio.NewReader(os.Stdin)
	
	for {
		b, err := reader.ReadByte()
		if err != nil {
			clearSuggestions()
			return ""
		}
		
		// Handle escape sequences
		if inEscape {
			escapeSeq = append(escapeSeq, b)
			
			if len(escapeSeq) >= 2 {
				if escapeSeq[0] == '[' {
					switch escapeSeq[1] {
					case 'A': // Up arrow
						if len(buffer) == 0 && !autocomplete.IsActive {
							// Navigate suggestions
							if selectedSuggestion < 0 {
								selectedSuggestion = 0
							} else if selectedSuggestion > 0 {
								selectedSuggestion--
							}
							displaySuggestions()
						} else if autocomplete.IsActive {
							autocomplete.MoveSelection(-1)
							redraw()
						}
						inEscape = false
						escapeSeq = nil
						
					case 'B': // Down arrow
						if len(buffer) == 0 && !autocomplete.IsActive {
							// Navigate suggestions
							if selectedSuggestion < len(suggestions)-1 && selectedSuggestion < 3 {
								selectedSuggestion++
							}
							displaySuggestions()
						} else if autocomplete.IsActive {
							autocomplete.MoveSelection(1)
							redraw()
						}
						inEscape = false
						escapeSeq = nil
						
					case 'D': // Left arrow
						if cursorPos > 0 {
							cursorPos--
							fmt.Print("\033[D")
						}
						inEscape = false
						escapeSeq = nil
						
					case 'C': // Right arrow
						if cursorPos < len(buffer) {
							cursorPos++
							fmt.Print("\033[C")
						}
						inEscape = false
						escapeSeq = nil
						
					default:
						inEscape = false
						escapeSeq = nil
					}
				} else {
					inEscape = false
					escapeSeq = nil
				}
			}
			continue
		}
		
		switch b {
		case 27: // ESC
			// Check for double ESC (clear line)
			now := time.Now()
			if !lastEscapeTime.IsZero() && now.Sub(lastEscapeTime) < 500*time.Millisecond {
				// Double ESC - clear line
				buffer = []rune{}
				cursorPos = 0
				atSymbolPos = -1
				autocomplete.IsActive = false
				selectedSuggestion = -1
				clearSuggestions()
				fmt.Print("\r" + prompt + "\033[K")
				displaySuggestions()
				lastEscapeTime = time.Time{}
			} else {
				// Start of escape sequence or cancel autocomplete
				if autocomplete.IsActive {
					autocomplete.IsActive = false
					atSymbolPos = -1
					redraw()
				} else {
					inEscape = true
					escapeSeq = []byte{}
					lastEscapeTime = now
				}
			}
			
		case 13, 10: // Enter
			clearSuggestions()
			
			// If a suggestion is selected and buffer is empty, use the suggestion
			if len(buffer) == 0 && selectedSuggestion >= 0 && selectedSuggestion < len(suggestions) {
				fmt.Println()
				return suggestions[selectedSuggestion].Command
			}
			
			// If file autocomplete is active and has selection
			if autocomplete.IsActive && autocomplete.SelectedIndex >= 0 && autocomplete.SelectedIndex < len(autocomplete.Suggestions) {
				selected := autocomplete.Suggestions[autocomplete.SelectedIndex]
				replacement := selected.RelPath
				if selected.IsDirectory && !strings.HasSuffix(replacement, "/") {
					replacement += "/"
				}
				
				// Replace from @ to cursor with the selected path
				newBuffer := make([]rune, 0)
				newBuffer = append(newBuffer, buffer[:atSymbolPos+1]...)
				newBuffer = append(newBuffer, []rune(replacement)...)
				buffer = newBuffer
				cursorPos = len(buffer)
				autocomplete.IsActive = false
				atSymbolPos = -1
				redraw()
				continue
			}
			
			fmt.Println()
			return string(buffer)
			
		case 9: // Tab
			if autocomplete.IsActive && autocomplete.SelectedIndex >= 0 && autocomplete.SelectedIndex < len(autocomplete.Suggestions) {
				selected := autocomplete.Suggestions[autocomplete.SelectedIndex]
				replacement := selected.RelPath
				if selected.IsDirectory && !strings.HasSuffix(replacement, "/") {
					replacement += "/"
				}
				
				// Replace from @ to cursor with the selected path
				newBuffer := make([]rune, 0)
				newBuffer = append(newBuffer, buffer[:atSymbolPos+1]...)
				newBuffer = append(newBuffer, []rune(replacement)...)
				buffer = newBuffer
				cursorPos = len(buffer)
				autocomplete.IsActive = false
				atSymbolPos = -1
				redraw()
			}
			
		case 3: // Ctrl+C
			clearSuggestions()
			fmt.Println("^C")
			return "__EXIT__"
			
		case 4: // Ctrl+D
			if len(buffer) == 0 {
				clearSuggestions()
				fmt.Println()
				return "__EXIT__"
			}
			
		case 127, 8: // Backspace
			if cursorPos > 0 {
				// Check if we're deleting the @ symbol
				if cursorPos-1 == atSymbolPos {
					atSymbolPos = -1
					autocomplete.IsActive = false
				}
				
				buffer = append(buffer[:cursorPos-1], buffer[cursorPos:]...)
				cursorPos--
				selectedSuggestion = -1 // Reset selection when typing
				
				// If buffer becomes empty, show suggestions again
				if len(buffer) == 0 {
					clearSuggestions()
					displaySuggestions()
				} else {
					clearSuggestions()
				}
				redraw()
			}
			
		case 64: // @ symbol
			if !autocomplete.IsActive {
				// Insert @ and activate autocomplete
				newBuffer := make([]rune, 0, len(buffer)+1)
				newBuffer = append(newBuffer, buffer[:cursorPos]...)
				newBuffer = append(newBuffer, '@')
				newBuffer = append(newBuffer, buffer[cursorPos:]...)
				buffer = newBuffer
				atSymbolPos = cursorPos
				cursorPos++
				clearSuggestions() // Hide AI suggestions
				redraw()
			} else {
				// Regular @ character
				newBuffer := make([]rune, 0, len(buffer)+1)
				newBuffer = append(newBuffer, buffer[:cursorPos]...)
				newBuffer = append(newBuffer, '@')
				newBuffer = append(newBuffer, buffer[cursorPos:]...)
				buffer = newBuffer
				cursorPos++
				redraw()
			}
			
		default:
			// Regular character
			if b >= 32 && b < 127 {
				// Clear suggestions when typing
				if showingSuggestions {
					clearSuggestions()
				}
				selectedSuggestion = -1
				
				// Insert at cursor position
				newBuffer := make([]rune, 0, len(buffer)+1)
				newBuffer = append(newBuffer, buffer[:cursorPos]...)
				newBuffer = append(newBuffer, rune(b))
				newBuffer = append(newBuffer, buffer[cursorPos:]...)
				buffer = newBuffer
				cursorPos++
				
				// Update display
				fmt.Print(string(rune(b)))
				if cursorPos < len(buffer) {
					fmt.Print(string(buffer[cursorPos:]))
					fmt.Printf("\033[%dD", len(buffer)-cursorPos)
				}
				
				// Check if we need to update autocomplete
				if atSymbolPos >= 0 && cursorPos > atSymbolPos {
					redraw()
				}
			}
		}
		
		// Reset escape time if not in escape mode
		if !inEscape {
			lastEscapeTime = time.Time{}
		}
	}
}