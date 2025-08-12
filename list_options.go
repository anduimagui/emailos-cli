package mailos

import (
	"fmt"
	"strings"
)

// ListOption represents a single item in a list
type ListOption struct {
	Label       string // Main text to display
	Icon        string // Optional icon/emoji
	Description string // Optional description
	Value       interface{} // Associated value
}

// ListRenderer handles rendering lists in the terminal
type ListRenderer struct {
	Title           string       // Title for the list
	Items           []ListOption // Items to display
	SelectedIndex   int          // Currently selected item
	MaxDisplay      int          // Maximum items to show at once
	ShowNumbers     bool         // Show line numbers
	ShowIcons       bool         // Show icons
	CompactMode     bool         // Compact display mode
	HeaderText      string       // Text to show in header
	FooterText      string       // Text to show in footer
}

// NewListRenderer creates a new list renderer with defaults
func NewListRenderer(title string, items []ListOption) *ListRenderer {
	return &ListRenderer{
		Title:         title,
		Items:         items,
		SelectedIndex: -1,
		MaxDisplay:    15,
		ShowNumbers:   false,
		ShowIcons:     true,
		CompactMode:   false,
	}
}

// MoveSelection moves the selection up or down
func (lr *ListRenderer) MoveSelection(direction int) {
	if len(lr.Items) == 0 {
		return
	}
	
	lr.SelectedIndex += direction
	
	// Wrap around
	if lr.SelectedIndex < 0 {
		lr.SelectedIndex = len(lr.Items) - 1
	} else if lr.SelectedIndex >= len(lr.Items) {
		lr.SelectedIndex = 0
	}
}

// GetSelected returns the currently selected item
func (lr *ListRenderer) GetSelected() *ListOption {
	if lr.SelectedIndex < 0 || lr.SelectedIndex >= len(lr.Items) {
		return nil
	}
	return &lr.Items[lr.SelectedIndex]
}

// RenderList renders the list to a string
func (lr *ListRenderer) RenderList() string {
	if len(lr.Items) == 0 {
		return ""
	}
	
	var output strings.Builder
	
	// Add spacing before list
	if !lr.CompactMode {
		output.WriteString("\n")
	}
	
	// Render header
	if lr.Title != "" {
		if lr.CompactMode {
			output.WriteString(fmt.Sprintf("\033[90m╭─── %s ───╮\033[0m\n", lr.Title))
		} else {
			output.WriteString(fmt.Sprintf("  %s\n", lr.Title))
			output.WriteString("  ──────────────────────────────────────────────────────────\n")
		}
	}
	
	// Determine items to display
	displayItems := lr.Items
	if lr.MaxDisplay > 0 && len(lr.Items) > lr.MaxDisplay {
		displayItems = lr.Items[:lr.MaxDisplay]
	}
	
	// Render items
	for i, item := range displayItems {
		lr.renderItem(&output, i, &item)
	}
	
	// Render footer
	if lr.CompactMode && lr.Title != "" {
		output.WriteString("\033[90m╰────────────────────────────────────────────────────╯\033[0m\n")
	} else if !lr.CompactMode {
		if len(lr.Items) > lr.MaxDisplay {
			output.WriteString(fmt.Sprintf("  ────────────────── Showing %d of %d ──────────────────\n", lr.MaxDisplay, len(lr.Items)))
		} else {
			output.WriteString("  ──────────────────────────────────────────────────────────\n")
		}
	}
	
	return output.String()
}

// renderItem renders a single item
func (lr *ListRenderer) renderItem(output *strings.Builder, index int, item *ListOption) {
	// Indentation and selection marker
	if lr.CompactMode {
		output.WriteString("\033[90m│\033[0m ")
	} else {
		if index == lr.SelectedIndex {
			output.WriteString("> ")
		} else {
			output.WriteString("  ")
		}
	}
	
	// Line number if enabled
	if lr.ShowNumbers {
		output.WriteString(fmt.Sprintf("%2d. ", index+1))
	}
	
	// Icon if enabled
	if lr.ShowIcons && item.Icon != "" {
		output.WriteString(item.Icon + " ")
	}
	
	// Label with selection highlighting
	if index == lr.SelectedIndex {
		// Highlighted with inverse colors (works in both modes)
		output.WriteString(fmt.Sprintf("\033[7m%s\033[0m", item.Label))
	} else if lr.CompactMode {
		// Colored text in compact mode
		output.WriteString(fmt.Sprintf("\033[36m%s\033[0m", item.Label))
	} else {
		// Normal text
		output.WriteString(item.Label)
	}
	
	// Description if present
	if item.Description != "" && !lr.CompactMode {
		output.WriteString(fmt.Sprintf(" \033[90m- %s\033[0m", item.Description))
	}
	
	output.WriteString("\n")
}

// ClearList returns ANSI codes to clear the rendered list
func (lr *ListRenderer) ClearList() string {
	if len(lr.Items) == 0 {
		return ""
	}
	
	// Calculate number of lines to clear
	numLines := 0
	
	// Space before
	if !lr.CompactMode {
		numLines++
	}
	
	// Header lines
	if lr.Title != "" {
		if lr.CompactMode {
			numLines++ // Single header line
		} else {
			numLines += 2 // Title + separator
		}
	}
	
	// Item lines
	displayCount := len(lr.Items)
	if lr.MaxDisplay > 0 && displayCount > lr.MaxDisplay {
		displayCount = lr.MaxDisplay
	}
	numLines += displayCount
	
	// Footer line
	numLines++
	
	// Build clear sequence
	var output strings.Builder
	for i := 0; i < numLines; i++ {
		output.WriteString("\033[A\033[2K") // Move up and clear line
	}
	
	return output.String()
}

// RenderInline renders the list inline below current cursor position
func (lr *ListRenderer) RenderInline() string {
	rendered := lr.RenderList()
	if rendered == "" {
		return ""
	}
	
	var output strings.Builder
	
	// Save cursor position
	output.WriteString("\033[s")
	
	// Render the list
	output.WriteString(rendered)
	
	// Restore cursor position
	output.WriteString("\033[u")
	
	return output.String()
}

// GetLineCount returns the number of lines this list will occupy
func (lr *ListRenderer) GetLineCount() int {
	if len(lr.Items) == 0 {
		return 0
	}
	
	count := 0
	
	// Space before
	if !lr.CompactMode {
		count++
	}
	
	// Header
	if lr.Title != "" {
		if lr.CompactMode {
			count++
		} else {
			count += 2
		}
	}
	
	// Items
	displayCount := len(lr.Items)
	if lr.MaxDisplay > 0 && displayCount > lr.MaxDisplay {
		displayCount = lr.MaxDisplay
	}
	count += displayCount
	
	// Footer
	count++
	
	return count
}

// CreateFileListOptions converts file suggestions to list options
func CreateFileListOptions(suggestions []FileSuggestion) []ListOption {
	options := make([]ListOption, len(suggestions))
	for i, s := range suggestions {
		icon := "📄"
		if s.IsDirectory {
			icon = "📁"
		}
		
		label := s.RelPath
		if s.IsDirectory && !strings.HasSuffix(label, "/") {
			label += "/"
		}
		
		// Keep full path for better visibility - don't truncate
		// This allows users to see the complete folder structure
		
		options[i] = ListOption{
			Label:       label,
			Icon:        icon,
			Description: s.ModTime,
			Value:       s,
		}
	}
	return options
}

// CreateAICommandOptions converts AI suggestions to list options
func CreateAICommandOptions(suggestions []AISuggestion) []ListOption {
	options := make([]ListOption, len(suggestions))
	for i, s := range suggestions {
		options[i] = ListOption{
			Label:       s.Title,
			Icon:        s.Icon,
			Description: s.Description,
			Value:       s.Command,
		}
	}
	return options
}