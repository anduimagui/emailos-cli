package mailos

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

// InputWithSuggestions handles input with suggestions below
type InputWithSuggestions struct {
	Prompt            string
	Buffer            []rune
	CursorPos         int
	Suggestions       []AISuggestion
	SelectedIndex     int
	ShowSuggestions   bool
	SuggestionsHeight int
}

// NewInputWithSuggestions creates a new input handler
func NewInputWithSuggestions(prompt string) *InputWithSuggestions {
	return &InputWithSuggestions{
		Prompt:            prompt,
		Buffer:            []rune{},
		CursorPos:         0,
		Suggestions:       GetDefaultAISuggestions(),
		SelectedIndex:     -1,
		ShowSuggestions:   true,
		SuggestionsHeight: 0,
	}
}

// RenderSuggestions displays suggestions below input
func (iws *InputWithSuggestions) RenderSuggestions() {
	if !iws.ShowSuggestions || len(iws.Buffer) > 0 {
		return
	}
	
	// Save cursor position
	fmt.Print("\033[s")
	
	// Move to next line
	fmt.Print("\n")
	
	// Show up to 5 suggestions
	displayed := 0
	maxDisplay := 5
	if len(iws.Suggestions) < maxDisplay {
		maxDisplay = len(iws.Suggestions)
	}
	
	fmt.Println("\033[90m┌─── Suggestions (↑↓ to select, Enter to use) ───┐\033[0m")
	displayed++
	
	for i := 0; i < maxDisplay; i++ {
		suggestion := iws.Suggestions[i]
		if i == iws.SelectedIndex {
			// Highlighted
			fmt.Printf("\033[90m│\033[0m \033[7m%s %s\033[0m\n", suggestion.Icon, suggestion.Title)
		} else {
			fmt.Printf("\033[90m│\033[0m %s \033[36m%s\033[0m\n", suggestion.Icon, suggestion.Title)
		}
		displayed++
	}
	
	fmt.Println("\033[90m└──────────────────────────────────────────────────┘\033[0m")
	displayed++
	
	iws.SuggestionsHeight = displayed
	
	// Restore cursor position
	fmt.Print("\033[u")
}

// ClearSuggestions removes suggestions from display
func (iws *InputWithSuggestions) ClearSuggestions() {
	if iws.SuggestionsHeight == 0 {
		return
	}
	
	// Save cursor position
	fmt.Print("\033[s")
	
	// Clear suggestion lines
	for i := 0; i < iws.SuggestionsHeight; i++ {
		fmt.Print("\n\033[2K") // Move down and clear line
	}
	
	// Move back up
	for i := 0; i < iws.SuggestionsHeight; i++ {
		fmt.Print("\033[A")
	}
	
	// Restore cursor position
	fmt.Print("\033[u")
	
	iws.SuggestionsHeight = 0
}

// Redraw redraws the input line
func (iws *InputWithSuggestions) Redraw() {
	// Clear current line
	fmt.Print("\r\033[2K")
	
	// Draw prompt and buffer
	fmt.Print(iws.Prompt + string(iws.Buffer))
	
	// Position cursor
	if iws.CursorPos < len(iws.Buffer) {
		moveBack := len(iws.Buffer) - iws.CursorPos
		for i := 0; i < moveBack; i++ {
			fmt.Print("\033[D")
		}
	}
}

// ReadInput reads user input with suggestions
func (iws *InputWithSuggestions) ReadInput() (string, error) {
	// Check if stdin is a terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		// Fallback to regular input
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(iws.Prompt)
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(input), nil
	}
	
	// Switch to raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	
	// Initial display
	fmt.Print(iws.Prompt)
	iws.RenderSuggestions()
	
	reader := bufio.NewReader(os.Stdin)
	var escapeSeq []byte
	inEscape := false
	lastEscapeTime := time.Time{}
	
	for {
		b, err := reader.ReadByte()
		if err != nil {
			iws.ClearSuggestions()
			return "", err
		}
		
		// Handle escape sequences
		if inEscape {
			escapeSeq = append(escapeSeq, b)
			
			if len(escapeSeq) >= 2 && escapeSeq[0] == '[' {
				switch escapeSeq[1] {
				case 'A': // Up arrow
					if len(iws.Buffer) == 0 && iws.ShowSuggestions {
						iws.ClearSuggestions()
						if iws.SelectedIndex <= 0 {
							iws.SelectedIndex = len(iws.Suggestions) - 1
							if iws.SelectedIndex > 4 {
								iws.SelectedIndex = 4
							}
						} else {
							iws.SelectedIndex--
						}
						iws.RenderSuggestions()
					}
					inEscape = false
					escapeSeq = nil
					
				case 'B': // Down arrow
					if len(iws.Buffer) == 0 && iws.ShowSuggestions {
						iws.ClearSuggestions()
						if iws.SelectedIndex < 0 {
							iws.SelectedIndex = 0
						} else if iws.SelectedIndex < 4 && iws.SelectedIndex < len(iws.Suggestions)-1 {
							iws.SelectedIndex++
						} else {
							iws.SelectedIndex = -1
						}
						iws.RenderSuggestions()
					}
					inEscape = false
					escapeSeq = nil
					
				case 'C': // Right arrow
					if iws.CursorPos < len(iws.Buffer) {
						iws.CursorPos++
						fmt.Print("\033[C")
					}
					inEscape = false
					escapeSeq = nil
					
				case 'D': // Left arrow
					if iws.CursorPos > 0 {
						iws.CursorPos--
						fmt.Print("\033[D")
					}
					inEscape = false
					escapeSeq = nil
					
				default:
					inEscape = false
					escapeSeq = nil
				}
			}
			continue
		}
		
		switch b {
		case 27: // ESC
			now := time.Now()
			if !lastEscapeTime.IsZero() && now.Sub(lastEscapeTime) < 500*time.Millisecond {
				// Double ESC - clear line
				iws.ClearSuggestions()
				iws.Buffer = []rune{}
				iws.CursorPos = 0
				iws.SelectedIndex = -1
				iws.ShowSuggestions = true
				iws.Redraw()
				iws.RenderSuggestions()
				lastEscapeTime = time.Time{}
			} else {
				inEscape = true
				escapeSeq = []byte{}
				lastEscapeTime = now
			}
			
		case 13, 10: // Enter
			iws.ClearSuggestions()
			
			// If a suggestion is selected and buffer is empty, use it
			if len(iws.Buffer) == 0 && iws.SelectedIndex >= 0 && iws.SelectedIndex < len(iws.Suggestions) {
				fmt.Println()
				return iws.Suggestions[iws.SelectedIndex].Command, nil
			}
			
			fmt.Println()
			return string(iws.Buffer), nil
			
		case 3: // Ctrl+C
			iws.ClearSuggestions()
			fmt.Println("^C")
			return "", fmt.Errorf("interrupted")
			
		case 4: // Ctrl+D
			if len(iws.Buffer) == 0 {
				iws.ClearSuggestions()
				fmt.Println()
				return "", fmt.Errorf("EOF")
			}
			
		case 127, 8: // Backspace
			if iws.CursorPos > 0 {
				iws.Buffer = append(iws.Buffer[:iws.CursorPos-1], iws.Buffer[iws.CursorPos:]...)
				iws.CursorPos--
				
				// Show suggestions again if buffer becomes empty
				if len(iws.Buffer) == 0 {
					iws.ClearSuggestions()
					iws.ShowSuggestions = true
					iws.SelectedIndex = -1
					iws.Redraw()
					iws.RenderSuggestions()
				} else {
					iws.Redraw()
				}
			}
			
		case 21: // Ctrl+U - Clear line
			if len(iws.Buffer) > 0 {
				iws.ClearSuggestions()
				iws.Buffer = []rune{}
				iws.CursorPos = 0
				iws.ShowSuggestions = true
				iws.SelectedIndex = -1
				iws.Redraw()
				iws.RenderSuggestions()
			}
			
		default:
			// Regular character
			if b >= 32 && b < 127 {
				// Hide suggestions when typing
				if iws.ShowSuggestions && len(iws.Buffer) == 0 {
					iws.ClearSuggestions()
					iws.ShowSuggestions = false
				}
				
				// Insert character
				newBuffer := make([]rune, 0, len(iws.Buffer)+1)
				newBuffer = append(newBuffer, iws.Buffer[:iws.CursorPos]...)
				newBuffer = append(newBuffer, rune(b))
				newBuffer = append(newBuffer, iws.Buffer[iws.CursorPos:]...)
				iws.Buffer = newBuffer
				iws.CursorPos++
				
				// Redraw
				iws.Redraw()
			}
		}
		
		// Reset escape time if not in escape mode
		if !inEscape {
			lastEscapeTime = time.Time{}
		}
	}
}

// ReadLineWithSuggestions is the main entry point for input with suggestions
func ReadLineWithSuggestions(prompt string) string {
	handler := NewInputWithSuggestions(prompt)
	result, err := handler.ReadInput()
	if err != nil {
		if err.Error() == "interrupted" || err.Error() == "EOF" {
			return "__EXIT__"
		}
		return ""
	}
	return result
}

// InteractiveInputWithAI handles the interactive input with AI suggestions
func InteractiveInputWithAI(config *Config) error {
	// Check if AI is configured
	if config.DefaultAICLI == "" || config.DefaultAICLI == "none" {
		fmt.Println("\n⚠️  No AI provider configured.")
		fmt.Println("Run 'mailos provider' to set up an AI provider.")
		return nil
	}
	
	// Get input with suggestions
	input := ReadLineWithSuggestions("▸ ")
	
	if input == "__EXIT__" {
		return fmt.Errorf("exit")
	}
	
	input = strings.TrimSpace(input)
	
	// Handle empty input
	if input == "" {
		return nil
	}
	
	// Check for commands
	if strings.HasPrefix(input, "/") {
		if input == "/" {
			return showCommandMenu()
		}
		return executeCommand(input)
	}
	
	// Process as AI query
	fmt.Printf("\n🤔 Processing: %s\n\n", input)
	return InvokeAIProvider(input)
}