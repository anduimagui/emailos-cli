package mailos

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	promptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	inputStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("255"))

	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	selectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	suggestionStyle = lipgloss.NewStyle().
		PaddingLeft(2)

	boxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)
)

type inputMode int

const (
	normalMode inputMode = iota
	commandMode
	fileMode
	accountMode
)

type model struct {
	textInput    textinput.Model
	suggestions  []AISuggestion
	commands     []command
	files        []string
	accounts     []AccountConfig
	mode         inputMode
	selectedIdx  int
	showList     bool
	width        int
	height       int
	result       string
	err          error
	quitting     bool
	showingAccountSelector bool
}

type command struct {
	name        string
	description string
	icon        string
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Ask me anything or type / for commands, @ for files (searches subdirs)..."
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 80
	ti.Prompt = promptStyle.Render("▸ ")

	return model{
		textInput:   ti,
		suggestions: GetDefaultAISuggestions(),
		commands: []command{
			{"read", "Browse and read your emails", ""},
			{"send", "Compose and send a new email", ""},
			{"report", "Generate email analytics", ""},
			{"unsubscribe", "Find unsubscribe links", ""},
			{"delete", "Delete emails by criteria", ""},
			{"mark-read", "Mark emails as read", ""},
			{"template", "Manage email templates", ""},
			{"configure", "Settings & configuration", ""},
			{"provider", "Set AI provider", ""},
			{"info", "Display configuration", ""},
			{"help", "Show help information", ""},
			{"exit", "Exit Mailos", ""},
		},
		mode:        normalMode,
		selectedIdx: 0,
		showList:    false,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textInput.Width = minInt(msg.Width-4, 80)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			m.err = fmt.Errorf("cancelled")
			return m, tea.Quit

		case tea.KeyCtrlD:
			if m.textInput.Value() == "" {
				m.quitting = true
				m.err = fmt.Errorf("exit")
				return m, tea.Quit
			}

		case tea.KeyEscape:
			if m.showingAccountSelector {
				m.showingAccountSelector = false
				m.mode = normalMode
				m.selectedIdx = 0
			} else if m.showList {
				m.showList = false
				m.mode = normalMode
				m.selectedIdx = 0
			}

		case tea.KeyTab:
			if m.showList && m.mode == fileMode && m.selectedIdx < len(m.files) {
				// Autocomplete with selected file
				value := m.textInput.Value()
				atPos := strings.LastIndex(value, "@")
				if atPos >= 0 {
					m.textInput.SetValue(value[:atPos] + "@" + m.files[m.selectedIdx] + " ")
				}
				m.showList = false
				m.mode = normalMode
				m.selectedIdx = 0
			} else if !m.showList && m.textInput.Value() == "" {
				// Show account selector when Tab is pressed with empty input
				config, _ := LoadConfig()
				m.accounts = GetAllAccounts(config)
				if len(m.accounts) > 0 {
					m.showingAccountSelector = true
					m.mode = accountMode
					m.selectedIdx = 0
				}
			}

		case tea.KeyUp:
			if m.showingAccountSelector {
				m.selectedIdx--
				if m.selectedIdx < 0 {
					m.selectedIdx = len(m.accounts) // Include "Add New Account" option
				}
			} else if m.showList {
				m.selectedIdx--
				if m.selectedIdx < 0 {
					switch m.mode {
					case normalMode:
						m.selectedIdx = len(m.suggestions) - 1
					case commandMode:
						m.selectedIdx = len(m.commands) - 1
					case fileMode:
						m.selectedIdx = len(m.files) - 1
					}
				}
			}

		case tea.KeyDown:
			if m.showingAccountSelector {
				m.selectedIdx++
				if m.selectedIdx > len(m.accounts) { // Include "Add New Account" option
					m.selectedIdx = 0
				}
			} else if m.showList {
				m.selectedIdx++
				switch m.mode {
				case normalMode:
					if m.selectedIdx >= len(m.suggestions) {
						m.selectedIdx = 0
					}
				case commandMode:
					if m.selectedIdx >= len(m.commands) {
						m.selectedIdx = 0
					}
				case fileMode:
					if m.selectedIdx >= len(m.files) {
						m.selectedIdx = 0
					}
				}
			}

		case tea.KeyCtrlS:
			// Ctrl+S for permanent local account setting
			if m.showingAccountSelector && m.selectedIdx < len(m.accounts) {
				selectedEmail := m.accounts[m.selectedIdx].Email
				
				// Permanently set account in local .email/config.json
				err := SetLocalAccountPreference(selectedEmail)
				if err == nil {
					m.showingAccountSelector = false
					m.mode = normalMode
					m.selectedIdx = 0
					
					// Force a refresh to update the display
					return m, tea.ClearScreen
				} else {
					// Handle error
					m.showingAccountSelector = false
					m.mode = normalMode
					m.selectedIdx = 0
					fmt.Printf("\n✗ Error setting local account: %v\n", err)
				}
				return m, nil
			}
			
		case tea.KeyEnter:
			// Handle account selector
			if m.showingAccountSelector {
				if m.selectedIdx < len(m.accounts) {
					selectedEmail := m.accounts[m.selectedIdx].Email
					
					// Session-only: Switch to selected account for sending
					_, err := InitializeMailSetup(selectedEmail)
					if err == nil {
						// The InitializeMailSetup already sets session default
						m.showingAccountSelector = false
						m.mode = normalMode
						m.selectedIdx = 0
						
						// Force a refresh to update the display
						return m, tea.ClearScreen
					} else {
						// Handle error
						m.showingAccountSelector = false
						m.mode = normalMode
						m.selectedIdx = 0
						fmt.Printf("\n✗ Error switching account: %v\n", err)
					}
				} else {
					// Add new account - show separate UI
					m.result = "/add-account"
					m.quitting = true
					return m, tea.Quit
				}
				return m, nil
			}

			value := m.textInput.Value()

			// Handle empty input - show suggestions
			if strings.TrimSpace(value) == "" && !m.showList {
				m.showList = true
				m.mode = normalMode
				m.selectedIdx = 0
				return m, nil
			}

			// Handle list selection
			if m.showList {
				switch m.mode {
				case normalMode:
					if m.selectedIdx < len(m.suggestions) {
						m.result = m.suggestions[m.selectedIdx].Command
						m.quitting = true
						return m, tea.Quit
					}
				case commandMode:
					if m.selectedIdx < len(m.commands) {
						m.result = "/" + m.commands[m.selectedIdx].name
						m.quitting = true
						return m, tea.Quit
					}
				case fileMode:
					if m.selectedIdx < len(m.files) {
						atPos := strings.LastIndex(value, "@")
						if atPos >= 0 {
							m.textInput.SetValue(value[:atPos] + "@" + m.files[m.selectedIdx] + " ")
						}
						m.showList = false
						m.mode = normalMode
						m.selectedIdx = 0
					}
				}
				return m, nil
			}

			// Normal submit
			m.result = value
			m.quitting = true
			return m, tea.Quit

		default:
			// Update text input
			m.textInput, cmd = m.textInput.Update(msg)
			
			// Check for mode triggers
			value := m.textInput.Value()
			if value == "/" {
				m.mode = commandMode
				m.showList = true
				m.selectedIdx = 0
			} else if strings.HasSuffix(value, "@") {
				m.mode = fileMode
				m.files = getLocalFiles()
				m.showList = true
				m.selectedIdx = 0
			} else if m.mode == commandMode && !strings.HasPrefix(value, "/") {
				m.mode = normalMode
				m.showList = false
			} else if m.mode == fileMode && !strings.Contains(value, "@") {
				m.mode = normalMode
				m.showList = false
			}

			// Filter commands if in command mode
			if m.mode == commandMode && len(value) > 1 {
				filter := strings.ToLower(value[1:])
				if filter != "" {
					var filtered []command
					for _, cmd := range m.commands {
						if strings.Contains(strings.ToLower(cmd.name), filter) ||
							strings.Contains(strings.ToLower(cmd.description), filter) {
							filtered = append(filtered, cmd)
						}
					}
					if len(filtered) > 0 {
						m.commands = filtered
						if m.selectedIdx >= len(m.commands) {
							m.selectedIdx = 0
						}
					}
				}
			}

			// Filter files if in file mode
			if m.mode == fileMode {
				atPos := strings.LastIndex(value, "@")
				if atPos >= 0 && atPos < len(value)-1 {
					filter := strings.ToLower(value[atPos+1:])
					if filter != "" {
						var filtered []string
						for _, file := range getLocalFiles() {
							if strings.Contains(strings.ToLower(file), filter) {
								filtered = append(filtered, file)
							}
						}
						m.files = filtered
						if m.selectedIdx >= len(m.files) {
							m.selectedIdx = 0
						}
					}
				}
			}
		}
	}

	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	// If showing account selector, show it instead
	if m.showingAccountSelector {
		return m.renderAccountSelector()
	}

	var s strings.Builder

	// Header with auth info
	config, _ := LoadConfig()
	
	// Check for account preference (local folder or session)
	localAccount := GetLocalAccountPreference()
	sessionAccount := GetSessionDefaultAccount()
	
	sendingAs := config.FromEmail
	accountSource := ""
	
	if localAccount != "" {
		sendingAs = localAccount
		accountSource = " (local)"
	} else if sessionAccount != "" {
		sendingAs = sessionAccount
		accountSource = " (session)"
	} else if sendingAs == "" {
		sendingAs = config.Email
	}
	
	var headerLines []string
	
	// First line: Account and AI provider with hint about clicking
	if config.Email != "" {
		aiDisplay := getFriendlyAIName(config.DefaultAICLI)
		headerLine := fmt.Sprintf("%s Account: %s │ %s AI: %s", IconAccount, config.Email, IconAI, aiDisplay)
		headerLines = append(headerLines, headerLine)
	}
	
	// Second line: Always show "Sending as" to indicate which account will be used
	if sendingAs != "" {
		fromLine := fmt.Sprintf("%s Sending as: %s%s", IconFromEmail, sendingAs, accountSource)
		headerLines = append(headerLines, fromLine)
	}
	
	// If no config, show default header
	if len(headerLines) == 0 {
		headerLines = append(headerLines, fmt.Sprintf("%s Mailos - Email Client", IconAccount))
	}
	
	for _, line := range headerLines {
		s.WriteString(titleStyle.Render(line) + "\n")
	}
	
	// Add hint about account selection
	if config.Email != "" {
		s.WriteString(helpStyle.Render("  Press Tab to switch accounts") + "\n")
	}
	s.WriteString("\n")

	// Input field
	s.WriteString(m.textInput.View() + "\n")

	// Show list based on mode
	if m.showList {
		s.WriteString("\n")
		switch m.mode {
		case normalMode:
			s.WriteString(helpStyle.Render("Suggestions (↑↓ to navigate, Enter to select):") + "\n\n")
			for i, sug := range m.suggestions {
				prefix := "  "
				style := suggestionStyle
				if i == m.selectedIdx {
					prefix = selectedStyle.Render("▸ ")
					style = selectedStyle
				}
				s.WriteString(fmt.Sprintf("%s%s\n", prefix, style.Render(sug.Title)))
			}

		case commandMode:
			s.WriteString(helpStyle.Render("Commands (↑↓ to navigate, Enter to select):") + "\n\n")
			for i, cmd := range m.commands {
				prefix := "  "
				style := suggestionStyle
				if i == m.selectedIdx {
					prefix = selectedStyle.Render("▸ ")
					style = selectedStyle
				}
				if cmd.icon != "" {
					s.WriteString(fmt.Sprintf("%s%s /%s - %s\n", prefix, cmd.icon, 
						style.Render(cmd.name), helpStyle.Render(cmd.description)))
				} else {
					s.WriteString(fmt.Sprintf("%s/%s - %s\n", prefix,
						style.Render(cmd.name), helpStyle.Render(cmd.description)))
				}
			}

		case fileMode:
			s.WriteString(helpStyle.Render("Files & Folders (↑↓ to navigate, Tab to autocomplete):") + "\n\n")
			maxShow := 15 // Increased to show more files
			start := 0
			if m.selectedIdx >= maxShow {
				start = m.selectedIdx - maxShow + 1
			}
			end := minInt(start+maxShow, len(m.files))
			
			for i := start; i < end; i++ {
				prefix := "  "
				style := suggestionStyle
				if i == m.selectedIdx {
					prefix = selectedStyle.Render("▸ ")
					style = selectedStyle
				}
				
				// Format file path for display
				filePath := m.files[i]
				icon := "📄"
				if strings.HasSuffix(filePath, "/") {
					icon = "📁"
				} else if strings.HasSuffix(filePath, ".md") {
					icon = "📝"
				} else if strings.HasSuffix(filePath, ".go") || strings.HasSuffix(filePath, ".py") || 
				          strings.HasSuffix(filePath, ".js") || strings.HasSuffix(filePath, ".ts") {
					icon = "💻"
				} else if strings.HasSuffix(filePath, ".json") || strings.HasSuffix(filePath, ".yaml") || 
				          strings.HasSuffix(filePath, ".yml") || strings.HasSuffix(filePath, ".toml") {
					icon = "⚙️"
				}
				
				s.WriteString(fmt.Sprintf("%s%s %s\n", prefix, icon, style.Render(filePath)))
			}
			
			if len(m.files) > maxShow {
				remaining := len(m.files) - end
				if remaining > 0 {
					s.WriteString(helpStyle.Render(fmt.Sprintf("\n  ... and %d more files/folders", remaining)) + "\n")
				}
			}
		}
	} else {
		// Help text when not showing list
		s.WriteString("\n" + helpStyle.Render("Enter: submit | /: commands | @: files | Tab: accounts | Ctrl+C: cancel | Ctrl+D: exit") + "\n")
	}

	return boxStyle.Render(s.String())
}

// renderAccountSelector renders the account selection UI with provider grouping
func (m model) renderAccountSelector() string {
	var s strings.Builder
	
	// Header
	s.WriteString(titleStyle.Render("📬 Select Email Account") + "\n\n")
	
	// Group accounts by provider for display
	var currentProvider string
	var isFirstInGroup bool = true
	
	// Show account list with provider grouping
	for i, acc := range m.accounts {
		cursor := "  "
		if i == m.selectedIdx {
			cursor = selectedStyle.Render("▸ ")
		}
		
		// Show provider header when switching providers
		if acc.Provider != currentProvider {
			if !isFirstInGroup {
				s.WriteString("\n") // Add spacing between provider groups
			}
			currentProvider = acc.Provider
			providerName := GetProviderName(acc.Provider)
			providerHeader := fmt.Sprintf("── %s ──", providerName)
			s.WriteString(helpStyle.Render(providerHeader) + "\n")
			isFirstInGroup = false
		}
		
		// Format account display with indentation and sub-email indicators
		var displayText string
		var prefix string
		
		if acc.Label == "Primary" {
			prefix = "🏠 "
			displayText = fmt.Sprintf("%s%s", prefix, acc.Email)
		} else if acc.Label == "Sub-email" || (acc.Provider == currentProvider && acc.Label != "Account") {
			prefix = "  ↳ "
			displayText = fmt.Sprintf("%s%s", prefix, acc.Email)
		} else {
			prefix = "📧 "
			if acc.Label != "" && acc.Label != "Account" {
				displayText = fmt.Sprintf("%s%s (%s)", prefix, acc.Email, acc.Label)
			} else {
				displayText = fmt.Sprintf("%s%s", prefix, acc.Email)
			}
		}
		
		// Apply styling
		if i == m.selectedIdx {
			displayText = selectedStyle.Render(displayText)
		} else {
			displayText = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Render(displayText)
		}
		
		s.WriteString(fmt.Sprintf("%s%s\n", cursor, displayText))
	}
	
	// Add spacing before "Add New Account" option
	if len(m.accounts) > 0 {
		s.WriteString("\n")
	}
	
	// Add "Add New Account" option
	cursor := "  "
	if m.selectedIdx == len(m.accounts) {
		cursor = selectedStyle.Render("▸ ")
	}
	
	addNewText := "➕ Add New Account"
	if m.selectedIdx == len(m.accounts) {
		addNewText = selectedStyle.Render(addNewText)
	} else {
		addNewText = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Render(addNewText)
	}
	s.WriteString(fmt.Sprintf("%s%s\n", cursor, addNewText))
	
	// Help text
	s.WriteString("\n" + helpStyle.Render("↑↓ Navigate • Enter: Session • Ctrl+S: Local Permanent • ESC: Cancel") + "\n")
	s.WriteString(helpStyle.Render("🏠 Main Account • ↳ Sub-email • 📧 Other Provider") + "\n")
	
	return boxStyle.Render(s.String())
}

// BubbleTeaInteractiveInput provides an interactive input experience using Bubble Tea
func BubbleTeaInteractiveInput() (string, error) {
	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		return "", err
	}

	finalModel := m.(model)
	if finalModel.err != nil {
		return "", finalModel.err
	}

	return finalModel.result, nil
}

// InteractiveModeWithBubbleTea runs the interactive mode using Bubble Tea for input
func InteractiveModeWithBubbleTea() error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Check for slash config
	slashConfig := loadSlashConfig()
	
	// If no AI provider configured, show setup prompt inline
	needsProvider := (config.DefaultAICLI == "" || config.DefaultAICLI == "none") && !hasConfiguredProvider(slashConfig)
	
	// Only show logo during initial setup
	if needsProvider {
		if ShouldShowLogo() {
			DisplayEmailOSLogo()
		}
	}

	// Main interactive loop
	for {
		// Get input using Bubble Tea
		input, err := BubbleTeaInteractiveInput()
		if err != nil {
			if err.Error() == "exit" || err.Error() == "cancelled" {
				fmt.Println("\nGoodbye!")
				return nil
			}
			continue
		}

		// Process the input
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Handle commands
		if strings.HasPrefix(input, "/") {
			// Special handling for add-account command
			if input == "/add-account" {
				if err := handleAddAccount(); err != nil {
					fmt.Printf("Error adding account: %v\n", err)
				}
				continue
			}
			
			if err := executeCommand(input); err != nil {
				if err.Error() == "exit" {
					fmt.Println("\nGoodbye!")
					return nil
				}
				fmt.Printf("Error: %v\n", err)
			}
			// After provider setup, update needsProvider flag
			if needsProvider {
				config, _ = LoadConfig()
				needsProvider = (config.DefaultAICLI == "" || config.DefaultAICLI == "none")
			}
			continue
		}

		// Handle file references
		processedInput := input
		if strings.Contains(input, "@") {
			processedInput = processFileReferences(input)
		}

		// Execute AI query
		if needsProvider {
			fmt.Println("\nNo AI provider configured. Use /provider to set one up.")
			continue
		}

		// Execute the AI query using existing handler
		if err := handleAIQuery(processedInput); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

// getLocalFiles returns a list of files recursively from the current directory
func getLocalFiles() []string {
	var files []string
	maxDepth := 3 // Limit recursion depth for performance
	
	// Walk through directory tree
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}
		
		// Skip hidden files/directories and common ignore patterns
		if strings.HasPrefix(info.Name(), ".") || 
		   strings.HasPrefix(path, ".git/") ||
		   strings.HasPrefix(path, "node_modules/") ||
		   strings.HasPrefix(path, "vendor/") ||
		   strings.HasPrefix(path, "__pycache__/") ||
		   strings.HasPrefix(path, ".venv/") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		
		// Check depth limit
		depth := strings.Count(path, string(os.PathSeparator))
		if depth > maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		
		// Skip the root directory itself
		if path == "." {
			return nil
		}
		
		// Add file or directory to list
		if info.IsDir() {
			files = append(files, path+"/")
		} else {
			files = append(files, path)
		}
		
		return nil
	})
	
	if err != nil {
		// Fallback to simple directory listing if walk fails
		entries, err := os.ReadDir(".")
		if err != nil {
			return files
		}
		
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() {
				name += "/"
			}
			files = append(files, name)
		}
	}
	
	return files
}

// handleAddAccount handles adding a new email account
func handleAddAccount() error {
	email, newAccount, err := ShowAccountSelector()
	if err != nil {
		return err
	}
	
	config, err := LoadConfig()
	if err != nil {
		return err
	}
	
	if newAccount != nil {
		// Add the new account
		return AddAccount(config, *newAccount)
	}
	
	// Switch to existing account
	if email != "" {
		return SwitchAccount(config, email)
	}
	
	return nil
}

// processFileReferences processes @ references in the input
func processFileReferences(input string) string {
	parts := strings.Split(input, "@")
	if len(parts) <= 1 {
		return input
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		// Find the end of the file reference (space or end of string)
		spaceIdx := strings.IndexAny(parts[i], " \t\n")
		if spaceIdx == -1 {
			spaceIdx = len(parts[i])
		}
		
		fileName := parts[i][:spaceIdx]
		remainder := ""
		if spaceIdx < len(parts[i]) {
			remainder = parts[i][spaceIdx:]
		}

		// Check if file exists
		if _, err := os.Stat(fileName); err == nil {
			// Read file content
			content, err := os.ReadFile(fileName)
			if err == nil {
				result += fmt.Sprintf("\n[File: %s]\n%s\n[End of %s]\n", 
					fileName, string(content), fileName)
			} else {
				result += fmt.Sprintf("@%s", fileName)
			}
		} else {
			result += fmt.Sprintf("@%s", fileName)
		}
		result += remainder
	}

	return result
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getFriendlyAIName returns a user-friendly display name for the AI provider
func getFriendlyAIName(provider string) string {
	switch provider {
	case AIProviderClaudeCode:
		return AIDisplayClaudeCode
	case AIProviderClaudeCodeYolo:
		return AIDisplayClaudeCodeYolo
	case AIProviderOpenAI:
		return AIDisplayOpenAI
	case AIProviderGemini:
		return AIDisplayGemini
	case AIProviderOpenCode:
		return AIDisplayOpenCode
	case AIProviderNone, "":
		return AIDisplayNone
	default:
		return provider
	}
}