package mailos

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/term"
)

// FileAutocomplete handles file/folder suggestions with @ symbol
type FileAutocomplete struct {
	BaseDir         string
	CurrentInput    string
	Suggestions     []FileSuggestion
	SelectedIndex   int
	IsActive        bool
	SearchQuery     string
	MaxDisplayItems int
}

// FileSuggestion represents a file or folder suggestion
type FileSuggestion struct {
	Name        string
	Path        string
	IsDirectory bool
	RelPath     string
	Size        int64
	ModTime     string
}

// NewFileAutocomplete creates a new file autocomplete instance
func NewFileAutocomplete(baseDir string) *FileAutocomplete {
	if baseDir == "" {
		baseDir, _ = os.Getwd()
	}
	return &FileAutocomplete{
		BaseDir:         baseDir,
		Suggestions:     []FileSuggestion{},
		SelectedIndex:   0,
		IsActive:        false,
		MaxDisplayItems: 20,
	}
}

// fuzzyMatch performs a fuzzy string matching
// Returns true if all characters in the query appear in the target in the same order
func fuzzyMatch(target, query string) bool {
	if query == "" {
		return true
	}
	
	queryIndex := 0
	queryLen := len(query)
	
	for i := 0; i < len(target) && queryIndex < queryLen; i++ {
		if target[i] == query[queryIndex] {
			queryIndex++
		}
	}
	
	return queryIndex == queryLen
}

// UpdateSuggestions updates the suggestion list based on the search query
func (fa *FileAutocomplete) UpdateSuggestions(query string) {
	fa.SearchQuery = strings.ToLower(query)
	fa.Suggestions = []FileSuggestion{}
	fa.SelectedIndex = 0
	
	// Recursively collect all files
	var allFiles []FileSuggestion
	collectFiles(fa.BaseDir, fa.BaseDir, &allFiles)
	
	// Filter based on fuzzy match
	for _, file := range allFiles {
		// Skip hidden files unless explicitly searching for them or they're in .email folder
		if strings.Contains(file.RelPath, "/.") && !strings.Contains(fa.SearchQuery, ".") && 
			!strings.Contains(file.RelPath, ".email/") {
			continue
		}
		
		// Fuzzy match against the relative path
		if fa.SearchQuery == "" || fuzzyMatch(strings.ToLower(file.RelPath), fa.SearchQuery) {
			fa.Suggestions = append(fa.Suggestions, file)
		}
	}
	
	// Sort suggestions alphabetically by path
	sort.Slice(fa.Suggestions, func(i, j int) bool {
		return strings.ToLower(fa.Suggestions[i].RelPath) < strings.ToLower(fa.Suggestions[j].RelPath)
	})
	
	// Limit to max display items
	if len(fa.Suggestions) > fa.MaxDisplayItems {
		fa.Suggestions = fa.Suggestions[:fa.MaxDisplayItems]
	}
}

// collectFiles recursively collects all files from a directory
func collectFiles(baseDir, currentDir string, files *[]FileSuggestion) {
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return
	}
	
	for _, entry := range entries {
		// Skip hidden directories (starting with .) except for .email folder
		if entry.IsDir() && strings.HasPrefix(entry.Name(), ".") && entry.Name() != ".email" {
			continue
		}
		
		// Skip common build/cache directories
		if entry.IsDir() && (entry.Name() == "node_modules" || 
			entry.Name() == ".next" || entry.Name() == "dist" || entry.Name() == ".turbo" ||
			entry.Name() == "__pycache__" || entry.Name() == "build") {
			continue
		}
		
		fullPath := filepath.Join(currentDir, entry.Name())
		relPath, _ := filepath.Rel(baseDir, fullPath)
		
		// Get file info
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		// Format the relative path to show folder hierarchy clearly
		// Use forward slashes for consistency
		relPath = strings.ReplaceAll(relPath, string(filepath.Separator), "/")
		
		// Add to list
		suggestion := FileSuggestion{
			Name:        entry.Name(),
			Path:        fullPath,
			IsDirectory: entry.IsDir(),
			RelPath:     relPath,
			Size:        info.Size(),
			ModTime:     info.ModTime().Format("Jan 02 15:04"),
		}
		*files = append(*files, suggestion)
		
		// Recurse into directories
		if entry.IsDir() {
			collectFiles(baseDir, fullPath, files)
		}
	}
}

// MoveSelection moves the selection up or down
func (fa *FileAutocomplete) MoveSelection(direction int) {
	if len(fa.Suggestions) == 0 {
		return
	}
	
	fa.SelectedIndex += direction
	
	// Wrap around
	if fa.SelectedIndex < 0 {
		fa.SelectedIndex = len(fa.Suggestions) - 1
	} else if fa.SelectedIndex >= len(fa.Suggestions) {
		fa.SelectedIndex = 0
	}
}

// GetSelected returns the currently selected suggestion
func (fa *FileAutocomplete) GetSelected() *FileSuggestion {
	if len(fa.Suggestions) == 0 || fa.SelectedIndex >= len(fa.Suggestions) {
		return nil
	}
	return &fa.Suggestions[fa.SelectedIndex]
}

// RenderSuggestions renders the suggestion list to the terminal
func (fa *FileAutocomplete) RenderSuggestions() string {
	if !fa.IsActive || len(fa.Suggestions) == 0 {
		return ""
	}
	
	// Convert to list options
	options := CreateFileListOptions(fa.Suggestions)
	
	// Create renderer
	renderer := NewListRenderer("Files (↑↓ navigate, Enter select, Tab complete, ESC cancel)", options)
	renderer.SelectedIndex = fa.SelectedIndex
	renderer.MaxDisplay = fa.MaxDisplayItems
	renderer.ShowIcons = true // Show folder/file icons for better visibility
	renderer.CompactMode = false
	
	// Add spacing before list
	return "\n" + renderer.RenderList()
}

// formatFileSize formats file size in human-readable format
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%dB", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%c", float64(size)/float64(div), "KMGTPE"[exp])
}

// ClearSuggestions clears the rendered suggestions from the terminal
func (fa *FileAutocomplete) ClearSuggestions() string {
	if !fa.IsActive || len(fa.Suggestions) == 0 {
		return ""
	}
	
	// Convert to list options
	options := CreateFileListOptions(fa.Suggestions)
	
	// Create renderer to calculate lines
	renderer := NewListRenderer("Files (↑↓ navigate, Enter select, Tab complete, ESC cancel)", options)
	renderer.SelectedIndex = fa.SelectedIndex
	renderer.MaxDisplay = fa.MaxDisplayItems
	renderer.ShowIcons = true
	renderer.CompactMode = false
	
	// Get line count and clear
	numLines := renderer.GetLineCount() + 1 // +1 for the initial newline
	
	var output strings.Builder
	for i := 0; i < numLines; i++ {
		output.WriteString("\033[A\033[2K") // Move up and clear line
	}
	
	return output.String()
}

// ReadLineWithFileAutocomplete reads input with file autocomplete support
func ReadLineWithFileAutocomplete(prompt string) string {
	// Check if stdin is a terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		// Fallback to regular ReadLineWithArrows
		return ReadLineWithArrows(prompt)
	}
	
	// Switch to raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return ReadLineWithArrows(prompt)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	
	fmt.Print(prompt)
	
	var buffer []rune
	var cursorPos int
	autocomplete := NewFileAutocomplete("")
	var atSymbolPos int = -1
	
	// Helper to redraw the line
	redraw := func() {
		// Clear current line
		fmt.Print("\r\033[2K")
		// Redraw prompt and buffer
		fmt.Print(prompt + string(buffer))
		
		// Check for @ symbol and update autocomplete
		if atSymbolPos >= 0 && cursorPos > atSymbolPos {
			query := string(buffer[atSymbolPos+1:cursorPos])
			autocomplete.UpdateSuggestions(query)
			autocomplete.IsActive = true
		} else {
			autocomplete.IsActive = false
		}
		
		// Render suggestions if active
		if autocomplete.IsActive {
			fmt.Print(autocomplete.RenderSuggestions())
			// Move cursor back up to input line
			options := CreateFileListOptions(autocomplete.Suggestions)
			renderer := NewListRenderer("", options)
			renderer.MaxDisplay = autocomplete.MaxDisplayItems
			numLines := renderer.GetLineCount() + 1 // +1 for initial newline
			for i := 0; i < numLines; i++ {
				fmt.Print("\033[A")
			}
		}
		
		// Position cursor correctly
		if cursorPos < len(buffer) {
			fmt.Printf("\r%s%s", prompt, string(buffer[:cursorPos]))
		} else {
			fmt.Printf("\r%s%s", prompt, string(buffer))
		}
	}
	
	b := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(b)
		if err != nil || n == 0 {
			return ""
		}
		
		// Handle special keys
		switch b[0] {
		case 27: // ESC
			// Read next bytes for arrow keys
			n, _ = os.Stdin.Read(b)
			if n > 0 && b[0] == '[' {
				n, _ = os.Stdin.Read(b)
				if n > 0 {
					switch b[0] {
					case 'A': // Up arrow
						if autocomplete.IsActive {
							// Clear old suggestions
							fmt.Print(autocomplete.ClearSuggestions())
							autocomplete.MoveSelection(-1)
							redraw()
						}
					case 'B': // Down arrow
						if autocomplete.IsActive {
							// Clear old suggestions
							fmt.Print(autocomplete.ClearSuggestions())
							autocomplete.MoveSelection(1)
							redraw()
						}
					case 'C': // Right arrow
						if cursorPos < len(buffer) {
							cursorPos++
							fmt.Print("\033[C")
						}
					case 'D': // Left arrow
						if cursorPos > 0 {
							cursorPos--
							fmt.Print("\033[D")
						}
					}
				}
			} else {
				// Single ESC - cancel autocomplete
				if autocomplete.IsActive {
					fmt.Print(autocomplete.ClearSuggestions())
					autocomplete.IsActive = false
					atSymbolPos = -1
					redraw()
				}
			}
			
		case 13, 10: // Enter
			if autocomplete.IsActive {
				// Insert selected file
				selected := autocomplete.GetSelected()
				if selected != nil {
					// Replace from @ to cursor with selected path
					newBuffer := make([]rune, 0)
					newBuffer = append(newBuffer, buffer[:atSymbolPos]...)
					newBuffer = append(newBuffer, []rune(selected.RelPath)...)
					if cursorPos < len(buffer) {
						newBuffer = append(newBuffer, buffer[cursorPos:]...)
					}
					buffer = newBuffer
					cursorPos = atSymbolPos + len(selected.RelPath)
					
					// Clear suggestions and reset
					fmt.Print(autocomplete.ClearSuggestions())
					autocomplete.IsActive = false
					atSymbolPos = -1
					redraw()
				}
			} else {
				// Normal enter - submit input
				if autocomplete.IsActive {
					fmt.Print(autocomplete.ClearSuggestions())
				}
				fmt.Println()
				return string(buffer)
			}
			
		case 3: // Ctrl+C
			if autocomplete.IsActive {
				fmt.Print(autocomplete.ClearSuggestions())
			}
			fmt.Println("^C")
			return "__EXIT__"
			
		case 4: // Ctrl+D
			if len(buffer) == 0 {
				if autocomplete.IsActive {
					fmt.Print(autocomplete.ClearSuggestions())
				}
				fmt.Println()
				return "__EXIT__"
			}
			
		case 127, 8: // Backspace
			if cursorPos > 0 {
				// Check if we're deleting the @ symbol
				if cursorPos-1 == atSymbolPos {
					if autocomplete.IsActive {
						fmt.Print(autocomplete.ClearSuggestions())
					}
					autocomplete.IsActive = false
					atSymbolPos = -1
				}
				
				buffer = append(buffer[:cursorPos-1], buffer[cursorPos:]...)
				cursorPos--
				redraw()
			}
			
		case 9: // Tab
			if autocomplete.IsActive && len(autocomplete.Suggestions) > 0 {
				// Auto-complete with first suggestion
				selected := &autocomplete.Suggestions[0]
				if len(autocomplete.Suggestions) == 1 || autocomplete.SelectedIndex > 0 {
					selected = autocomplete.GetSelected()
				}
				
				// Replace from @ to cursor with selected path
				newBuffer := make([]rune, 0)
				newBuffer = append(newBuffer, buffer[:atSymbolPos]...)
				newBuffer = append(newBuffer, []rune(selected.RelPath)...)
				if cursorPos < len(buffer) {
					newBuffer = append(newBuffer, buffer[cursorPos:]...)
				}
				buffer = newBuffer
				cursorPos = atSymbolPos + len(selected.RelPath)
				
				// Clear suggestions and reset
				fmt.Print(autocomplete.ClearSuggestions())
				autocomplete.IsActive = false
				atSymbolPos = -1
				redraw()
			}
			
		case '@': // @ symbol - trigger autocomplete
			// Clear any existing autocomplete
			if autocomplete.IsActive {
				fmt.Print(autocomplete.ClearSuggestions())
			}
			
			// Insert @ and mark position
			newBuffer := make([]rune, 0, len(buffer)+1)
			newBuffer = append(newBuffer, buffer[:cursorPos]...)
			newBuffer = append(newBuffer, '@')
			newBuffer = append(newBuffer, buffer[cursorPos:]...)
			buffer = newBuffer
			atSymbolPos = cursorPos
			cursorPos++
			
			// Start autocomplete
			autocomplete.UpdateSuggestions("")
			autocomplete.IsActive = true
			redraw()
			
		default:
			// Regular character
			if b[0] >= 32 && b[0] < 127 {
				// Insert at cursor position
				newBuffer := make([]rune, 0, len(buffer)+1)
				newBuffer = append(newBuffer, buffer[:cursorPos]...)
				newBuffer = append(newBuffer, rune(b[0]))
				newBuffer = append(newBuffer, buffer[cursorPos:]...)
				buffer = newBuffer
				cursorPos++
				
				// Update autocomplete if active
				if autocomplete.IsActive && cursorPos > atSymbolPos {
					// Clear old suggestions first
					fmt.Print(autocomplete.ClearSuggestions())
				}
				
				redraw()
			}
		}
	}
}