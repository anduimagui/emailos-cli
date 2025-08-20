// +build ignore

// This file contains the full-featured Bubble Tea UI implementation
// It's currently disabled but can be re-enabled by removing the build ignore tag above
// and uncommenting the code in bubbletea_input.go

package mailos

/*
// Full Update function with all features
func (m model) UpdateFull(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.showingEmailContent {
				m.showingEmailContent = false
				m.selectedEmail = nil
			} else if m.showingEmailList {
				m.showingEmailList = false
				m.mode = normalMode
				m.selectedIdx = 0
			} else if m.showingAccountSelector {
				m.showingAccountSelector = false
				m.mode = normalMode
				m.selectedIdx = 0
			} else if m.showList {
				m.showList = false
				m.mode = normalMode
				m.selectedIdx = 0
			}

		case tea.KeyTab:
			// Tab key switches to account selector
			if !m.showingAccountSelector && !m.showingEmailContent {
				// Fetch all configured accounts
				config, _ := LoadConfig()
				m.accounts = GetAllAccounts(config)
				if len(m.accounts) > 0 {
					m.showingAccountSelector = true
					m.selectedIdx = 0
				}
			} else if m.mode == fileMode && m.selectedIdx < len(m.files) {
				// In file mode, Tab autocompletes the selected file
				selectedFile := m.files[m.selectedIdx]
				currentValue := m.textInput.Value()
				
				// Find the @ symbol and replace everything after it
				atIndex := strings.LastIndex(currentValue, "@")
				if atIndex >= 0 {
					newValue := currentValue[:atIndex+1] + selectedFile
					m.textInput.SetValue(newValue)
					m.textInput.SetCursor(len(newValue))
				}
				
				m.showList = false
				m.mode = normalMode
				m.selectedIdx = 0
			}

		case tea.KeyUp:
			if msg.Alt && len(m.emails) > 0 {
				// Alt+Up navigates emails
				if !m.showingEmailList {
					m.showingEmailList = true
					m.mode = emailListMode
					m.selectedIdx = 0
				} else if m.selectedIdx > 0 {
					m.selectedIdx--
				}
			} else if m.showingAccountSelector && m.selectedIdx > 0 {
				m.selectedIdx--
			} else if m.showList && m.selectedIdx > 0 {
				m.selectedIdx--
			}

		case tea.KeyDown:
			if msg.Alt && len(m.emails) > 0 {
				// Alt+Down navigates emails
				if !m.showingEmailList {
					m.showingEmailList = true
					m.mode = emailListMode
					m.selectedIdx = 0
				} else if m.selectedIdx < len(m.emails)-1 {
					m.selectedIdx++
				}
			} else if m.showingAccountSelector && m.selectedIdx < len(m.accounts)-1 {
				m.selectedIdx++
			} else if m.showList {
				switch m.mode {
				case normalMode:
					if m.selectedIdx < len(m.suggestions)-1 {
						m.selectedIdx++
					}
				case commandMode:
					if m.selectedIdx < len(m.commands)-1 {
						m.selectedIdx++
					}
				case fileMode:
					if m.selectedIdx < len(m.files)-1 {
						m.selectedIdx++
					}
				}
			}

		case tea.KeyShiftTab:
			// Shift+Tab to make this the local account (project-specific)
			if m.showingAccountSelector && m.selectedIdx < len(m.accounts) {
				selectedEmail := m.accounts[m.selectedIdx].Email
				
				// Set as local account preference
				if err := SetLocalAccountPreference(selectedEmail); err == nil {
					// Success - close the selector and refresh
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
			// Check if Alt is pressed for opening email
			if msg.Alt && m.mode == emailListMode && m.selectedIdx < len(m.emails) {
				m.selectedEmail = m.emails[m.selectedIdx]
				m.showingEmailContent = true
				return m, nil
			}
			
			// Handle email list selection
			if m.showingEmailList && m.selectedIdx < len(m.emails) {
				m.selectedEmail = m.emails[m.selectedIdx]
				m.showingEmailContent = true
				return m, nil
			}
			
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
				}
				return m, nil
			}
			
			// Handle list selection
			if m.showList {
				switch m.mode {
				case normalMode:
					if m.selectedIdx < len(m.suggestions) {
						m.result = m.suggestions[m.selectedIdx].Query
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
						// Autocomplete the selected file
						selectedFile := m.files[m.selectedIdx]
						currentValue := m.textInput.Value()
						
						// Find the @ symbol and replace everything after it
						atIndex := strings.LastIndex(currentValue, "@")
						if atIndex >= 0 {
							newValue := currentValue[:atIndex+1] + selectedFile
							m.textInput.SetValue(newValue)
							m.textInput.SetCursor(len(newValue))
						}
						
						m.showList = false
						m.mode = normalMode
						m.selectedIdx = 0
					}
				}
			} else {
				// Normal enter - submit the input
				input := strings.TrimSpace(m.textInput.Value())
				if input != "" {
					m.result = input
					m.quitting = true
					return m, tea.Quit
				}
			}

		default:
			// Handle regular input
			prevValue := m.textInput.Value()
			m.textInput, cmd = m.textInput.Update(msg)
			newValue := m.textInput.Value()

			// Check for triggers
			if newValue != prevValue {
				// Check for @ trigger (file autocomplete)
				if strings.Contains(newValue, "@") && !strings.Contains(prevValue, "@") {
					// User just typed @, show file list
					m.mode = fileMode
					m.files = getFilesAndFolders()
					m.showList = true
					m.selectedIdx = 0
				} else if m.mode == fileMode && !strings.Contains(newValue, "@") {
					// User removed the @, hide file list
					m.mode = normalMode
					m.showList = false
					m.selectedIdx = 0
				} else if m.mode == fileMode && strings.Contains(newValue, "@") {
					// Update file list based on partial input after @
					atIndex := strings.LastIndex(newValue, "@")
					if atIndex >= 0 && atIndex < len(newValue)-1 {
						partial := newValue[atIndex+1:]
						m.files = filterFiles(getFilesAndFolders(), partial)
						if len(m.files) == 0 {
							m.showList = false
						}
					}
				} else if strings.HasPrefix(newValue, "/") {
					// Command mode
					if len(newValue) == 1 {
						// Just "/" - show all commands
						m.mode = commandMode
						m.showList = true
						m.selectedIdx = 0
					} else {
						// Filter commands based on input
						partial := newValue[1:]
						m.commands = filterCommands(m.commands, partial)
						if len(m.commands) > 0 {
							m.mode = commandMode
							m.showList = true
						} else {
							m.showList = false
						}
					}
				} else if prevValue != "" && strings.HasPrefix(prevValue, "/") && !strings.HasPrefix(newValue, "/") {
					// User removed the /, exit command mode
					m.mode = normalMode
					m.showList = false
					m.selectedIdx = 0
				} else if m.mode == normalMode && !strings.HasPrefix(newValue, "/") && !strings.Contains(newValue, "@") {
					// Normal mode - show suggestions if input is empty or matches
					if newValue == "" {
						m.showList = true
						m.selectedIdx = 0
					} else {
						// Filter suggestions based on input
						m.suggestions = filterSuggestions(GetDefaultAISuggestions(), newValue)
						m.showList = len(m.suggestions) > 0
						if m.selectedIdx >= len(m.suggestions) {
							m.selectedIdx = 0
						}
					}
				}
			}
		}
	}

	// Update text input
	if !m.quitting {
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}

// Full View function with all UI elements
func (m model) ViewFull() string {
	if m.quitting {
		return ""
	}

	// If showing email content, show it
	if m.showingEmailContent && m.selectedEmail != nil {
		return m.renderEmailContent()
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

	// Add email list at the bottom
	if len(m.emails) > 0 {
		s.WriteString("\n" + titleStyle.Render("📧 Recent Emails") + "\n")
		s.WriteString(helpStyle.Render("Alt+↑/↓: Navigate emails | Alt+Enter: Open email") + "\n\n")
		
		maxEmails := 5 // Show last 5 emails to keep it compact
		if len(m.emails) < maxEmails {
			maxEmails = len(m.emails)
		}
		
		for i := 0; i < maxEmails; i++ {
			email := m.emails[i]
			
			// Format the email line compactly
			from := email.From
			if len(from) > 25 {
				from = from[:22] + "..."
			}
			
			subject := email.Subject
			if len(subject) > 40 {
				subject = subject[:37] + "..."
			}
			
			date := email.Date.Format("Jan 2")
			
			line := fmt.Sprintf("%-25s │ %-40s │ %s", from, subject, date)
			
			// Highlight if this email is selected (when in email navigation mode)
			if m.mode == emailListMode && i == m.selectedIdx {
				s.WriteString(selectedStyle.Render("▸ " + line) + "\n")
			} else {
				s.WriteString("  " + line + "\n")
			}
		}
		
		if len(m.emails) > maxEmails {
			s.WriteString(helpStyle.Render(fmt.Sprintf("  ... and %d more emails", len(m.emails)-maxEmails)) + "\n")
		}
	}

	return boxStyle.Render(s.String())
}
*/