package mailos

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetDocumentationPath returns the path to documentation files
func GetDocumentationPath(command string) string {
	// First, try to find docs relative to the executable
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		
		// Check common locations relative to executable
		possiblePaths := []string{
			filepath.Join(execDir, "docs", command+".md"),
			filepath.Join(execDir, "..", "docs", command+".md"),
			filepath.Join(execDir, "..", "..", "docs", command+".md"),
		}
		
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}
	
	// Try to find docs in the current working directory
	cwd, err := os.Getwd()
	if err == nil {
		possiblePaths := []string{
			filepath.Join(cwd, "docs", command+".md"),
			filepath.Join(cwd, "..", "docs", command+".md"),
		}
		
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}
	
	// Try to find docs in the home directory (for installed version)
	homeDir, err := os.UserHomeDir()
	if err == nil {
		possiblePaths := []string{
			filepath.Join(homeDir, ".emailos", "docs", command+".md"),
			filepath.Join(homeDir, ".email", "docs", command+".md"),
		}
		
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}
	
	// Check if we're in development environment
	if devPath := os.Getenv("EMAILOS_DEV_PATH"); devPath != "" {
		docPath := filepath.Join(devPath, "docs", command+".md")
		if _, err := os.Stat(docPath); err == nil {
			return docPath
		}
	}
	
	return ""
}

// ReadDocumentation reads and returns the documentation content
func ReadDocumentation(command string) (string, error) {
	docPath := GetDocumentationPath(command)
	if docPath == "" {
		return "", fmt.Errorf("documentation not found for command: %s", command)
	}
	
	content, err := os.ReadFile(docPath)
	if err != nil {
		return "", fmt.Errorf("failed to read documentation: %v", err)
	}
	
	return string(content), nil
}

// FormatDocumentationForTerminal formats markdown documentation for terminal display
func FormatDocumentationForTerminal(content string) string {
	lines := strings.Split(content, "\n")
	var formatted []string
	inCodeBlock := false
	
	for _, line := range lines {
		// Handle code blocks
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			if inCodeBlock {
				formatted = append(formatted, "")
			} else {
				formatted = append(formatted, "")
			}
			continue
		}
		
		if inCodeBlock {
			formatted = append(formatted, "  "+line)
			continue
		}
		
		// Handle headers
		if strings.HasPrefix(line, "# ") {
			formatted = append(formatted, "")
			formatted = append(formatted, "═══════════════════════════════════════════════")
			formatted = append(formatted, strings.TrimPrefix(line, "# "))
			formatted = append(formatted, "═══════════════════════════════════════════════")
			formatted = append(formatted, "")
		} else if strings.HasPrefix(line, "## ") {
			formatted = append(formatted, "")
			formatted = append(formatted, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			formatted = append(formatted, strings.TrimPrefix(line, "## "))
			formatted = append(formatted, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			formatted = append(formatted, "")
		} else if strings.HasPrefix(line, "### ") {
			formatted = append(formatted, "")
			formatted = append(formatted, "▶ "+strings.TrimPrefix(line, "### "))
			formatted = append(formatted, "")
		} else if strings.HasPrefix(line, "| ") {
			// Tables - keep as is but ensure spacing
			formatted = append(formatted, line)
		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			// Lists
			formatted = append(formatted, "  • "+strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* "))
		} else if strings.HasPrefix(line, "  - ") || strings.HasPrefix(line, "  * ") {
			// Nested lists
			formatted = append(formatted, "    ◦ "+strings.TrimPrefix(strings.TrimPrefix(line, "  - "), "  * "))
		} else if line == "" {
			// Empty lines
			formatted = append(formatted, "")
		} else {
			// Regular text
			formatted = append(formatted, line)
		}
	}
	
	return strings.Join(formatted, "\n")
}

// GetCommandHelp returns help text for a command, preferring documentation if available
func GetCommandHelp(command string, defaultHelp string) string {
	// Try to read documentation
	doc, err := ReadDocumentation(command)
	if err == nil {
		// Format the documentation for terminal display
		formatted := FormatDocumentationForTerminal(doc)
		return formatted
	}
	
	// Fall back to default help
	return defaultHelp
}

// ShowExtendedHelp displays extended help from documentation
func ShowExtendedHelp(command string) bool {
	doc, err := ReadDocumentation(command)
	if err != nil {
		return false
	}
	
	// Use a pager if available and the content is long
	lines := strings.Split(doc, "\n")
	if len(lines) > 40 {
		// Try to use a pager
		if pager := os.Getenv("PAGER"); pager != "" {
			ShowWithPager(FormatDocumentationForTerminal(doc), pager)
			return true
		} else if HasCommand("less") {
			ShowWithPager(FormatDocumentationForTerminal(doc), "less")
			return true
		} else if HasCommand("more") {
			ShowWithPager(FormatDocumentationForTerminal(doc), "more")
			return true
		}
	}
	
	// Otherwise just print it
	fmt.Println(FormatDocumentationForTerminal(doc))
	return true
}

// HasCommand checks if a command exists in PATH
func HasCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// ShowWithPager displays content using a pager program
func ShowWithPager(content string, pager string) {
	cmd := exec.Command(pager)
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}