package mailos

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
)

// SlashConfig represents the configuration for slash commands
type SlashConfig struct {
	DefaultProvider string `json:"default_provider,omitempty"`
	LastCommand     string `json:"last_command,omitempty"`
}

// Removed - now in interactive_enhanced.go

// showInteractiveMenuLegacy displays the main interactive menu (legacy version)
// This is kept for backward compatibility but the main entry is in interactive_enhanced.go
func showInteractiveMenuLegacy() error {
	options := []struct {
		Label       string
		Description string
		Action      func() error
	}{
		{"Ask AI Assistant", "Send a query to your configured AI provider", handleAIAssistantQuery},
		{"Read Emails", "Browse and read your emails", handleInteractiveRead},
		{"Send Email", "Compose and send a new email", handleInteractiveSend},
		{"Generate Report", "Create an email report for a time range", handleInteractiveReport},
		{"Find Unsubscribe Links", "Locate unsubscribe links in emails", handleInteractiveUnsubscribe},
		{"Delete Emails", "Remove emails by various criteria", handleInteractiveDelete},
		{"Mark Emails as Read", "Mark selected emails as read", handleInteractiveMarkRead},
		{"Manage Templates", "Customize email templates", func() error { return ManageTemplate() }},
		{"Configure Settings", "Manage email and AI provider settings", handleInteractiveConfigure},
		{"Set AI Provider", "Select or change AI provider", func() error { return SelectAndConfigureAIProvider() }},
		{"Show Info", "Display current configuration", func() error { return showInfo() }},
		{"Exit", "Exit EmailOS", func() error { return fmt.Errorf("exit") }},
	}

	// Create menu items with descriptions
	items := make([]string, len(options))
	for i, opt := range options {
		items[i] = opt.Label
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "▸ {{ . | cyan }}",
		Inactive: "  {{ . }}",
		Selected: "✓ {{ . | green }}",
		Details: `
--------- Action Details ----------
{{ "Description:" | faint }}	{{ .Description }}`,
	}

	// Create a searcher function for fuzzy finding
	searcher := func(input string, index int) bool {
		option := options[index]
		name := strings.Replace(strings.ToLower(option.Label), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     "What would you like to do?",
		Items:     items,
		Templates: templates,
		Size:      12,
		Searcher:  searcher,
	}

	for {
		index, _, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				fmt.Println("\nExiting EmailOS...")
				return nil
			}
			return err
		}

		// Execute the selected action
		if err := options[index].Action(); err != nil {
			if err.Error() == "exit" {
				fmt.Println("Exiting EmailOS...")
				return nil
			}
			fmt.Printf("Error: %v\n", err)
		}

		// After action completes, ask if they want to continue
		if index != len(options)-1 { // Not exit option
			fmt.Println()
			continuePrompt := promptui.Select{
				Label: "Continue?",
				Items: []string{"Return to Main Menu", "Exit"},
			}
			idx, _, err := continuePrompt.Run()
			if err != nil || idx == 1 {
				fmt.Println("Exiting EmailOS...")
				return nil
			}
			fmt.Println()
		} else {
			return nil
		}
	}
}

// handleAIAssistantQuery handles AI assistant queries (legacy - kept for compatibility)
func handleAIAssistantQuery() error {
	prompt := promptui.Prompt{
		Label: "Enter your query for the AI assistant",
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("query cannot be empty")
			}
			return nil
		},
	}

	query, err := prompt.Run()
	if err != nil {
		return err
	}

	return HandleQueryWithProviderSelection(query)
}

// handleInteractiveRead provides an interactive email reading experience
func handleInteractiveRead() error {
	options := []string{
		"Read all emails",
		"Read unread emails only",
		"Read from specific sender",
		"Read by subject",
		"Read from last N days",
		"Read with custom filters",
		"Back to main menu",
	}

	prompt := promptui.Select{
		Label: "Select read option",
		Items: options,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return err
	}

	opts := ReadOptions{
		Limit: 10,
	}

	switch index {
	case 0: // Read all
		// Use default options
	case 1: // Unread only
		opts.UnreadOnly = true
	case 2: // From specific sender
		fromPrompt := promptui.Prompt{
			Label: "Enter sender email address",
		}
		from, err := fromPrompt.Run()
		if err != nil {
			return err
		}
		opts.FromAddress = from
	case 3: // By subject
		subjectPrompt := promptui.Prompt{
			Label: "Enter subject keyword",
		}
		subject, err := subjectPrompt.Run()
		if err != nil {
			return err
		}
		opts.Subject = subject
	case 4: // Last N days
		daysPrompt := promptui.Prompt{
			Label:   "Number of days",
			Default: "7",
		}
		days, err := daysPrompt.Run()
		if err != nil {
			return err
		}
		opts.Since = getTimeFromDays(days)
	case 5: // Custom filters
		return handleCustomReadFilters()
	case 6: // Back
		return nil
	}

	// Ask for number of emails
	limitPrompt := promptui.Prompt{
		Label:   "Number of emails to read",
		Default: "10",
	}
	limitStr, err := limitPrompt.Run()
	if err == nil {
		if limit := parseNumber(limitStr); limit > 0 {
			opts.Limit = limit
		}
	}

	client, err := NewClient()
	if err != nil {
		return err
	}

	fmt.Println("\nReading emails...")
	emails, err := client.ReadEmails(opts)
	if err != nil {
		return fmt.Errorf("failed to read emails: %v", err)
	}

	fmt.Print(FormatEmailList(emails))
	return nil
}

// handleInteractiveSend provides guided email composition
func handleInteractiveSend() error {
	fmt.Println("\n━━━ Compose New Email ━━━")
	
	// Get recipients
	toPrompt := promptui.Prompt{
		Label: "To (comma-separated for multiple)",
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("at least one recipient is required")
			}
			return nil
		},
	}
	to, err := toPrompt.Run()
	if err != nil {
		return err
	}

	// Get CC (optional)
	ccPrompt := promptui.Prompt{
		Label: "CC (optional, comma-separated)",
	}
	cc, _ := ccPrompt.Run()

	// Get subject
	subjectPrompt := promptui.Prompt{
		Label: "Subject",
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("subject is required")
			}
			return nil
		},
	}
	subject, err := subjectPrompt.Run()
	if err != nil {
		return err
	}

	// Get body
	fmt.Println("Enter email body (press Enter twice to finish):")
	var bodyLines []string
	emptyCount := 0
	for {
		var line string
		fmt.Scanln(&line)
		if line == "" {
			emptyCount++
			if emptyCount >= 2 {
				break
			}
		} else {
			emptyCount = 0
		}
		bodyLines = append(bodyLines, line)
	}
	body := strings.Join(bodyLines, "\n")

	// Confirm send
	fmt.Printf("\n━━━ Email Preview ━━━\n")
	fmt.Printf("To: %s\n", to)
	if cc != "" {
		fmt.Printf("CC: %s\n", cc)
	}
	fmt.Printf("Subject: %s\n", subject)
	fmt.Printf("Body:\n%s\n", body)
	
	confirmPrompt := promptui.Select{
		Label: "Send this email?",
		Items: []string{"Send", "Cancel"},
	}
	idx, _, err := confirmPrompt.Run()
	if err != nil || idx == 1 {
		fmt.Println("Email cancelled.")
		return nil
	}

	// Prepare message
	msg := &EmailMessage{
		To:      strings.Split(to, ","),
		Subject: subject,
		Body:    body,
	}
	
	if cc != "" {
		msg.CC = strings.Split(cc, ",")
	}

	// Convert markdown to HTML
	html := MarkdownToHTML(body)
	if html != body {
		msg.BodyHTML = html
	}

	if err := Send(msg); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	fmt.Println("✓ Email sent successfully!")
	return nil
}

// handleInteractiveReport generates email reports
func handleInteractiveReport() error {
	selectedRange, err := SelectTimeRange()
	if err != nil {
		return err
	}
	
	client, err := NewClient()
	if err != nil {
		return err
	}

	fmt.Printf("\nGenerating report for: %s\n", selectedRange.Name)
	
	opts := ReadOptions{
		Since: selectedRange.Since,
		Limit: 1000,
	}
	
	emails, err := client.ReadEmails(opts)
	if err != nil {
		return fmt.Errorf("failed to read emails: %v", err)
	}
	
	var filteredEmails []*Email
	for _, email := range emails {
		if email.Date.After(selectedRange.Since) && email.Date.Before(selectedRange.Until) {
			filteredEmails = append(filteredEmails, email)
		}
	}
	
	report := GenerateEmailReport(filteredEmails, *selectedRange)
	
	// Ask if they want to save to file
	savePrompt := promptui.Select{
		Label: "Save report to file?",
		Items: []string{"Display only", "Save to file", "Both"},
	}
	saveIdx, _, err := savePrompt.Run()
	if err == nil && saveIdx > 0 {
		filenamePrompt := promptui.Prompt{
			Label:   "Filename",
			Default: fmt.Sprintf("email-report-%s.md", selectedRange.Name),
		}
		filename, err := filenamePrompt.Run()
		if err == nil {
			if err := os.WriteFile(filename, []byte(report), 0644); err == nil {
				fmt.Printf("✓ Report saved to %s\n", filename)
			}
		}
	}
	
	if saveIdx != 1 { // Not "Save to file" only
		fmt.Println(report)
	}
	
	return nil
}

// handleInteractiveUnsubscribe finds unsubscribe links
func handleInteractiveUnsubscribe() error {
	fromPrompt := promptui.Prompt{
		Label: "Enter sender email (or press Enter for all)",
	}
	from, _ := fromPrompt.Run()
	
	client, err := NewClient()
	if err != nil {
		return err
	}
	
	opts := ReadOptions{
		Limit: 20,
	}
	if from != "" {
		opts.FromAddress = from
	}
	
	fmt.Println("\nSearching for unsubscribe links...")
	links, err := client.FindUnsubscribeLinks(opts)
	if err != nil {
		return fmt.Errorf("failed to find unsubscribe links: %v", err)
	}
	
	if len(links) == 0 {
		fmt.Println("No unsubscribe links found.")
		return nil
	}
	
	fmt.Print(GetUnsubscribeReport(links))
	
	if len(links) > 0 && len(links[0].Links) > 0 {
		openPrompt := promptui.Select{
			Label: "Open first link in browser?",
			Items: []string{"Yes", "No"},
		}
		idx, _, err := openPrompt.Run()
		if err == nil && idx == 0 {
			fmt.Printf("Opening: %s\n", links[0].Links[0])
			// Implementation would open the link
		}
	}
	
	return nil
}

// handleInteractiveDelete handles email deletion
func handleInteractiveDelete() error {
	options := []string{
		"Delete by email IDs",
		"Delete all from sender",
		"Delete by subject",
		"Back to main menu",
	}

	prompt := promptui.Select{
		Label: "Select delete option",
		Items: options,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return err
	}

	if index == 3 {
		return nil
	}

	client, err := NewClient()
	if err != nil {
		return err
	}

	switch index {
	case 0: // Delete by IDs
		idsPrompt := promptui.Prompt{
			Label: "Enter email IDs (comma-separated)",
		}
		idsStr, err := idsPrompt.Run()
		if err != nil {
			return err
		}
		
		ids := parseEmailIDs(idsStr)
		if len(ids) == 0 {
			return fmt.Errorf("no valid IDs provided")
		}
		
		confirmPrompt := promptui.Select{
			Label: fmt.Sprintf("Delete %d email(s)?", len(ids)),
			Items: []string{"Yes", "No"},
		}
		idx, _, err := confirmPrompt.Run()
		if err != nil || idx == 1 {
			fmt.Println("Deletion cancelled.")
			return nil
		}
		
		if err := client.DeleteEmails(ids); err != nil {
			return fmt.Errorf("failed to delete emails: %v", err)
		}
		fmt.Printf("✓ Deleted %d email(s)\n", len(ids))
		
	case 1: // Delete from sender
		fromPrompt := promptui.Prompt{
			Label: "Enter sender email address",
		}
		from, err := fromPrompt.Run()
		if err != nil {
			return err
		}
		
		opts := ReadOptions{
			FromAddress: from,
			Limit:       100,
		}
		
		emails, err := client.ReadEmails(opts)
		if err != nil {
			return fmt.Errorf("failed to find emails: %v", err)
		}
		
		if len(emails) == 0 {
			fmt.Println("No emails found from that sender.")
			return nil
		}
		
		confirmPrompt := promptui.Select{
			Label: fmt.Sprintf("Delete %d email(s) from %s?", len(emails), from),
			Items: []string{"Yes", "No"},
		}
		idx, _, err := confirmPrompt.Run()
		if err != nil || idx == 1 {
			fmt.Println("Deletion cancelled.")
			return nil
		}
		
		var emailIds []uint32
		for _, email := range emails {
			emailIds = append(emailIds, email.ID)
		}
		
		if err := client.DeleteEmails(emailIds); err != nil {
			return fmt.Errorf("failed to delete emails: %v", err)
		}
		fmt.Printf("✓ Deleted %d email(s)\n", len(emails))
		
	case 2: // Delete by subject
		subjectPrompt := promptui.Prompt{
			Label: "Enter subject keyword",
		}
		subject, err := subjectPrompt.Run()
		if err != nil {
			return err
		}
		
		opts := ReadOptions{
			Subject: subject,
			Limit:   100,
		}
		
		emails, err := client.ReadEmails(opts)
		if err != nil {
			return fmt.Errorf("failed to find emails: %v", err)
		}
		
		if len(emails) == 0 {
			fmt.Println("No emails found with that subject.")
			return nil
		}
		
		fmt.Printf("Found %d email(s) with subject containing '%s'\n", len(emails), subject)
		for i, email := range emails {
			if i < 5 {
				fmt.Printf("  - %s: %s\n", email.From, email.Subject)
			}
		}
		if len(emails) > 5 {
			fmt.Printf("  ... and %d more\n", len(emails)-5)
		}
		
		confirmPrompt := promptui.Select{
			Label: fmt.Sprintf("Delete all %d email(s)?", len(emails)),
			Items: []string{"Yes", "No"},
		}
		idx, _, err := confirmPrompt.Run()
		if err != nil || idx == 1 {
			fmt.Println("Deletion cancelled.")
			return nil
		}
		
		var emailIds []uint32
		for _, email := range emails {
			emailIds = append(emailIds, email.ID)
		}
		
		if err := client.DeleteEmails(emailIds); err != nil {
			return fmt.Errorf("failed to delete emails: %v", err)
		}
		fmt.Printf("✓ Deleted %d email(s)\n", len(emails))
	}
	
	return nil
}

// handleInteractiveMarkRead marks emails as read
func handleInteractiveMarkRead() error {
	idsPrompt := promptui.Prompt{
		Label: "Enter email IDs to mark as read (comma-separated)",
	}
	idsStr, err := idsPrompt.Run()
	if err != nil {
		return err
	}
	
	ids := parseEmailIDs(idsStr)
	if len(ids) == 0 {
		return fmt.Errorf("no valid IDs provided")
	}
	
	client, err := NewClient()
	if err != nil {
		return err
	}
	
	if err := client.MarkEmailsAsRead(ids); err != nil {
		return fmt.Errorf("failed to mark emails as read: %v", err)
	}
	
	fmt.Printf("✓ Marked %d email(s) as read\n", len(ids))
	return nil
}

// handleInteractiveConfigure manages configuration
func handleInteractiveConfigure() error {
	options := []string{
		"Quick Configuration",
		"Full Configuration",
		"Change Email Account",
		"Change Display Name",
		"Change AI Provider",
		"Back to main menu",
	}

	prompt := promptui.Select{
		Label: "Configuration Options",
		Items: options,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return err
	}

	switch index {
	case 0:
		return QuickConfigMenu()
	case 1:
		return Configure()
	case 2:
		return Setup()
	case 3:
		config, err := LoadConfig()
		if err != nil {
			return err
		}
		return editDisplayName(config)
	case 4:
		return SelectAndConfigureAIProvider()
	case 5:
		return nil
	}
	
	return nil
}

// handleCustomReadFilters handles custom email filters
func handleCustomReadFilters() error {
	opts := ReadOptions{}
	
	// Unread filter
	unreadPrompt := promptui.Select{
		Label: "Show unread only?",
		Items: []string{"All emails", "Unread only"},
	}
	unreadIdx, _, _ := unreadPrompt.Run()
	opts.UnreadOnly = (unreadIdx == 1)
	
	// From filter
	fromPrompt := promptui.Prompt{
		Label: "From address (or press Enter to skip)",
	}
	from, _ := fromPrompt.Run()
	if from != "" {
		opts.FromAddress = from
	}
	
	// Subject filter
	subjectPrompt := promptui.Prompt{
		Label: "Subject contains (or press Enter to skip)",
	}
	subject, _ := subjectPrompt.Run()
	if subject != "" {
		opts.Subject = subject
	}
	
	// Date filter
	datePrompt := promptui.Select{
		Label: "Date range",
		Items: []string{"All time", "Today", "Last 7 days", "Last 30 days", "Custom"},
	}
	dateIdx, _, _ := datePrompt.Run()
	switch dateIdx {
	case 1:
		opts.Since = getTimeFromDays("0")
	case 2:
		opts.Since = getTimeFromDays("7")
	case 3:
		opts.Since = getTimeFromDays("30")
	case 4:
		daysPrompt := promptui.Prompt{
			Label: "Days ago",
		}
		days, _ := daysPrompt.Run()
		opts.Since = getTimeFromDays(days)
	}
	
	// Limit
	limitPrompt := promptui.Prompt{
		Label:   "Maximum emails to fetch",
		Default: "20",
	}
	limitStr, _ := limitPrompt.Run()
	if limit := parseNumber(limitStr); limit > 0 {
		opts.Limit = limit
	} else {
		opts.Limit = 20
	}
	
	client, err := NewClient()
	if err != nil {
		return err
	}
	
	fmt.Println("\nReading emails with custom filters...")
	emails, err := client.ReadEmails(opts)
	if err != nil {
		return fmt.Errorf("failed to read emails: %v", err)
	}
	
	fmt.Print(FormatEmailList(emails))
	return nil
}

// SelectAndConfigureAIProvider allows interactive selection of AI provider
func SelectAndConfigureAIProvider() error {
	provider, err := InteractiveAIProviderSelect()
	if err != nil {
		return err
	}
	
	if provider == "configure" {
		return Configure()
	}
	
	if provider == "none" {
		fmt.Println("Continuing without AI provider.")
		return nil
	}
	
	// Load and update config
	config, err := LoadConfig()
	if err != nil {
		return err
	}
	
	config.DefaultAICLI = provider
	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %v", err)
	}
	
	// Also save to slash config
	slashConfig := &SlashConfig{
		DefaultProvider: provider,
	}
	saveSlashConfig(slashConfig)
	
	fmt.Printf("✓ AI Provider set to: %s\n", GetAICLIName(provider))
	return nil
}

// loadSlashConfig loads the slash command configuration
func loadSlashConfig() *SlashConfig {
	configPath := ".email/.slash_config.json"
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return &SlashConfig{}
	}
	
	var config SlashConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return &SlashConfig{}
	}
	
	return &config
}

// saveSlashConfig saves the slash command configuration
func saveSlashConfig(config *SlashConfig) error {
	configDir := ".email"
	configPath := filepath.Join(configDir, ".slash_config.json")
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(configPath, data, 0644)
}

// hasConfiguredProvider checks if there's a provider in slash config
func hasConfiguredProvider(config *SlashConfig) bool {
	return config != nil && config.DefaultProvider != "" && config.DefaultProvider != "none"
}