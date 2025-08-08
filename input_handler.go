package mailos

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

// ReadLineWithArrows reads a line of input with full arrow key support for cursor movement
func ReadLineWithArrows(prompt string) string {
	// Check if stdin is a terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		// Fallback to regular input for non-terminal environments
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(prompt)
		input, err := reader.ReadString('\n')
		if err != nil {
			return ""
		}
		return strings.TrimSpace(input)
	}

	// Switch to raw mode for terminal input
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		// Fallback to regular input if we can't get raw mode
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(prompt)
		input, err := reader.ReadString('\n')
		if err != nil {
			return ""
		}
		return strings.TrimSpace(input)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	fmt.Print(prompt)
	
	var buffer []rune
	var cursorPos int
	var escapeSeq []byte
	inEscape := false
	lastEscapeTime := time.Time{}
	
	// Helper to redraw the line
	redraw := func() {
		// Move to beginning of line after prompt
		fmt.Print("\r" + prompt)
		fmt.Print(string(buffer))
		fmt.Print("\033[K") // Clear to end of line
		// Position cursor
		if cursorPos < len(buffer) {
			fmt.Printf("\r%s%s", prompt, string(buffer[:cursorPos]))
		}
	}
	
	reader := bufio.NewReader(os.Stdin)
	
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return ""
		}
		
		// Handle escape sequences
		if inEscape {
			escapeSeq = append(escapeSeq, b)
			
			// Check if sequence is complete
			if len(escapeSeq) >= 2 {
				if escapeSeq[0] == '[' {
					switch escapeSeq[1] {
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
						
					case 'H': // Home
						if cursorPos > 0 {
							fmt.Printf("\r%s", prompt)
							cursorPos = 0
						}
						inEscape = false
						escapeSeq = nil
						
					case 'F': // End
						if cursorPos < len(buffer) {
							fmt.Printf("\r%s%s", prompt, string(buffer))
							cursorPos = len(buffer)
						}
						inEscape = false
						escapeSeq = nil
						
					case '3': // Delete key starts with ESC[3
						if len(escapeSeq) >= 3 && escapeSeq[2] == '~' {
							if cursorPos < len(buffer) {
								buffer = append(buffer[:cursorPos], buffer[cursorPos+1:]...)
								redraw()
							}
							inEscape = false
							escapeSeq = nil
						}
						
					case 'A', 'B': // Up/Down arrows - ignore for now
						inEscape = false
						escapeSeq = nil
						
					default:
						if len(escapeSeq) >= 3 || (escapeSeq[1] >= '0' && escapeSeq[1] <= '9') {
							// Unknown or complex sequence, reset
							inEscape = false
							escapeSeq = nil
						}
					}
				} else {
					// Not a bracket sequence, reset
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
				fmt.Print("\r" + prompt + "\033[K")
				lastEscapeTime = time.Time{}
			} else {
				// Start of escape sequence
				inEscape = true
				escapeSeq = []byte{}
				lastEscapeTime = now
			}
			
		case 13, 10: // Enter
			fmt.Println()
			return string(buffer)
			
		case 3: // Ctrl+C
			fmt.Println("^C")
			return "__EXIT__"
			
		case 4: // Ctrl+D
			if len(buffer) == 0 {
				fmt.Println()
				return "__EXIT__"
			}
			
		case 127, 8: // Backspace
			if cursorPos > 0 {
				buffer = append(buffer[:cursorPos-1], buffer[cursorPos:]...)
				cursorPos--
				redraw()
			}
			
		case 1: // Ctrl+A (Home)
			if cursorPos > 0 {
				fmt.Printf("\r%s", prompt)
				cursorPos = 0
			}
			
		case 5: // Ctrl+E (End)
			if cursorPos < len(buffer) {
				fmt.Printf("\r%s%s", prompt, string(buffer))
				cursorPos = len(buffer)
			}
			
		case 11: // Ctrl+K (Kill to end)
			if cursorPos < len(buffer) {
				buffer = buffer[:cursorPos]
				fmt.Print("\033[K")
			}
			
		case 21: // Ctrl+U (Kill to beginning)
			if cursorPos > 0 {
				buffer = buffer[cursorPos:]
				cursorPos = 0
				redraw()
			}
			
		case 23: // Ctrl+W (Kill word backward)
			if cursorPos > 0 {
				// Find start of previous word
				i := cursorPos - 1
				// Skip trailing spaces
				for i > 0 && buffer[i] == ' ' {
					i--
				}
				// Skip word chars
				for i > 0 && buffer[i] != ' ' {
					i--
				}
				if buffer[i] == ' ' {
					i++
				}
				// Remove from i to cursorPos
				buffer = append(buffer[:i], buffer[cursorPos:]...)
				cursorPos = i
				redraw()
			}
			
		default:
			// Regular character
			if b >= 32 && b < 127 {
				// Insert at cursor position
				newBuffer := make([]rune, 0, len(buffer)+1)
				newBuffer = append(newBuffer, buffer[:cursorPos]...)
				newBuffer = append(newBuffer, rune(b))
				newBuffer = append(newBuffer, buffer[cursorPos:]...)
				buffer = newBuffer
				cursorPos++
				
				// Print the character
				fmt.Print(string(rune(b)))
				// If not at end, redraw rest and reposition
				if cursorPos < len(buffer) {
					fmt.Print(string(buffer[cursorPos:]))
					// Move cursor back
					fmt.Printf("\033[%dD", len(buffer)-cursorPos)
				}
			}
		}
		
		// Reset escape time if not in escape mode
		if !inEscape {
			lastEscapeTime = time.Time{}
		}
	}
}