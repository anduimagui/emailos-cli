package mailos

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Account selector states
type accountSelectorState int

const (
	selectingAccount accountSelectorState = iota
	addingNewAccount
	enteringEmail
	enteringPassword
	enteringProvider
)

// accountSelectorModel is the Bubble Tea model for account selection
type accountSelectorModel struct {
	accounts     []AccountConfig
	selectedIdx  int
	state        accountSelectorState
	newAccount   AccountConfig
	inputValue   string
	err          error
	selected     string
	cancelled    bool
	width        int
	height       int
}

// InitAccountSelector creates a new account selector model
func InitAccountSelector(config *Config) accountSelectorModel {
	accounts := GetAllAccounts(config)
	
	// Add "Add New Account" option
	return accountSelectorModel{
		accounts:    accounts,
		selectedIdx: 0,
		state:       selectingAccount,
	}
}

func (m accountSelectorModel) Init() tea.Cmd {
	return nil
}

func (m accountSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		switch m.state {
		case selectingAccount:
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEscape:
				m.cancelled = true
				return m, tea.Quit
				
			case tea.KeyUp:
				if m.selectedIdx > 0 {
					m.selectedIdx--
				} else {
					m.selectedIdx = len(m.accounts)
				}
				
			case tea.KeyDown:
				if m.selectedIdx < len(m.accounts) {
					m.selectedIdx++
				} else {
					m.selectedIdx = 0
				}
				
			case tea.KeyEnter:
				if m.selectedIdx < len(m.accounts) {
					// Selected an existing account
					m.selected = m.accounts[m.selectedIdx].Email
					return m, tea.Quit
				} else {
					// Selected "Add New Account"
					m.state = enteringEmail
					m.inputValue = ""
				}
			}
			
		case enteringEmail:
			switch msg.Type {
			case tea.KeyEscape:
				m.state = selectingAccount
				m.inputValue = ""
				
			case tea.KeyEnter:
				if m.inputValue != "" {
					m.newAccount.Email = m.inputValue
					m.newAccount.FromEmail = m.inputValue
					m.state = enteringProvider
					m.inputValue = ""
				}
				
			case tea.KeyBackspace:
				if len(m.inputValue) > 0 {
					m.inputValue = m.inputValue[:len(m.inputValue)-1]
				}
				
			default:
				if msg.Type == tea.KeyRunes {
					m.inputValue += string(msg.Runes)
				}
			}
			
		case enteringProvider:
			switch msg.Type {
			case tea.KeyEscape:
				m.state = enteringEmail
				m.inputValue = m.newAccount.Email
				
			case tea.KeyEnter:
				if m.inputValue != "" {
					m.newAccount.Provider = m.inputValue
					m.state = enteringPassword
					m.inputValue = ""
				}
				
			case tea.KeyBackspace:
				if len(m.inputValue) > 0 {
					m.inputValue = m.inputValue[:len(m.inputValue)-1]
				}
				
			default:
				if msg.Type == tea.KeyRunes {
					m.inputValue += string(msg.Runes)
				}
			}
			
		case enteringPassword:
			switch msg.Type {
			case tea.KeyEscape:
				m.state = enteringProvider
				m.inputValue = m.newAccount.Provider
				
			case tea.KeyEnter:
				if m.inputValue != "" {
					m.newAccount.Password = m.inputValue
					// Account is complete, add it
					m.selected = m.newAccount.Email
					return m, tea.Quit
				}
				
			case tea.KeyBackspace:
				if len(m.inputValue) > 0 {
					m.inputValue = m.inputValue[:len(m.inputValue)-1]
				}
				
			default:
				if msg.Type == tea.KeyRunes {
					m.inputValue += string(msg.Runes)
				}
			}
		}
	}
	
	return m, nil
}

func (m accountSelectorModel) View() string {
	var s strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginBottom(1)
	
	s.WriteString(headerStyle.Render("📬 Select Email Account") + "\n\n")
	
	switch m.state {
	case selectingAccount:
		// Show account list
		for i, acc := range m.accounts {
			cursor := "  "
			if i == m.selectedIdx {
				cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render("▸ ")
			}
			
			label := acc.Email
			if acc.Label != "" {
				label = fmt.Sprintf("%s (%s)", acc.Email, acc.Label)
			}
			
			if i == m.selectedIdx {
				label = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170")).Render(label)
			}
			
			s.WriteString(fmt.Sprintf("%s%s\n", cursor, label))
		}
		
		// Add "Add New Account" option
		cursor := "  "
		if m.selectedIdx == len(m.accounts) {
			cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render("▸ ")
		}
		
		addNewText := "➕ Add New Account"
		if m.selectedIdx == len(m.accounts) {
			addNewText = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170")).Render(addNewText)
		}
		s.WriteString(fmt.Sprintf("%s%s\n", cursor, addNewText))
		
		// Help text
		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(1)
		s.WriteString(helpStyle.Render("\n↑↓ Navigate • Enter: Select • ESC: Cancel"))
		
	case enteringEmail:
		s.WriteString("Enter email address:\n")
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render("▸ "))
		s.WriteString(m.inputValue)
		s.WriteString("\n\n")
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Enter: Continue • ESC: Back"))
		
	case enteringProvider:
		s.WriteString("Enter email provider (gmail, outlook, etc.):\n")
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render("▸ "))
		s.WriteString(m.inputValue)
		s.WriteString("\n\n")
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Enter: Continue • ESC: Back"))
		
	case enteringPassword:
		s.WriteString("Enter app password:\n")
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render("▸ "))
		// Show asterisks for password
		s.WriteString(strings.Repeat("*", len(m.inputValue)))
		s.WriteString("\n\n")
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Enter: Save Account • ESC: Back"))
	}
	
	// Wrap in a box
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(60)
	
	return boxStyle.Render(s.String())
}

// ShowAccountSelector displays the account selector and returns the selected email
func ShowAccountSelector() (string, *AccountConfig, error) {
	config, err := LoadConfig()
	if err != nil {
		return "", nil, err
	}
	
	model := InitAccountSelector(config)
	p := tea.NewProgram(model, tea.WithAltScreen())
	
	finalModel, err := p.Run()
	if err != nil {
		return "", nil, err
	}
	
	m := finalModel.(accountSelectorModel)
	if m.cancelled {
		return "", nil, fmt.Errorf("cancelled")
	}
	
	// If a new account was created, return it
	if m.newAccount.Email != "" {
		return m.selected, &m.newAccount, nil
	}
	
	return m.selected, nil, nil
}