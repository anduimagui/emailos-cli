package mailos

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// InteractiveMode runs the interactive input mode with slash commands
func InteractiveMode() error {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("EMAILOS INTERACTIVE MODE")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  /read      - Read emails")
	fmt.Println("  /send      - Send an email")
	fmt.Println("  /report    - Generate email report")
	fmt.Println("  /delete    - Delete emails")
	fmt.Println("  /unsubscribe - Find unsubscribe links")
	fmt.Println("  /mark-read - Mark emails as read")
	fmt.Println("  /template  - Manage email templates")
	fmt.Println("  /configure - Configure email settings")
	fmt.Println("  /info      - Show current configuration")
	fmt.Println("  /help      - Show this help message")
	fmt.Println("  /exit      - Exit interactive mode")
	fmt.Println()
	fmt.Println("Type any other text to send to your configured AI provider.")
	fmt.Println("Press Enter with no text to exit.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("mailos> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		input = strings.TrimSpace(input)

		// Exit on empty input
		if input == "" {
			fmt.Println("Exiting interactive mode.")
			return nil
		}

		// Check if it's a slash command
		if strings.HasPrefix(input, "/") {
			if err := handleSlashCommand(input); err != nil {
				if err.Error() == "exit" {
					fmt.Println("Exiting interactive mode.")
					return nil
				}
				fmt.Printf("Error: %v\n", err)
			}
		} else {
			// Handle as AI query
			if err := handleAIQuery(input); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}
		fmt.Println()
	}
}

// handleSlashCommand processes slash commands
func handleSlashCommand(command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {
	case "/exit":
		return fmt.Errorf("exit")

	case "/help":
		printInteractiveHelp()
		return nil

	case "/read":
		return handleReadCommand(args)

	case "/send":
		return handleSendCommand(args)

	case "/report":
		return handleReportCommand(args)

	case "/delete":
		return handleDeleteCommand(args)

	case "/unsubscribe":
		return handleUnsubscribeCommand(args)

	case "/mark-read":
		return handleMarkReadCommand(args)

	case "/template":
		return ManageTemplate()

	case "/configure":
		return Configure()

	case "/info":
		return showInfo()

	default:
		return fmt.Errorf("unknown command: %s (type /help for available commands)", cmd)
	}
}

// handleAIQuery sends the query to the configured AI provider
func handleAIQuery(query string) error {
	// Check if AI provider is configured
	config, err := LoadConfig()
	if err != nil {
		fmt.Println("No email configuration found. Run /configure to set up.")
		return nil
	}

	if config.DefaultAICLI == "" || config.DefaultAICLI == "none" {
		fmt.Println("No AI provider configured.")
		fmt.Println("Would you like to configure one now? (y/n)")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(response)) == "y" {
			provider, err := InteractiveAIProviderSelect()
			if err != nil {
				return err
			}

			if provider != "none" && provider != "configure" {
				config.DefaultAICLI = provider
				if err := SaveConfig(config); err != nil {
					return fmt.Errorf("failed to save configuration: %v", err)
				}
				fmt.Printf("AI Provider set to: %s\n\n", GetAICLIName(provider))
				return InvokeAIProvider(query)
			}
		}
		return nil
	}

	// Invoke the AI provider
	return InvokeAIProvider(query)
}

// printInteractiveHelp displays help information
func printInteractiveHelp() {
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("EMAILOS INTERACTIVE MODE HELP")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Slash Commands:")
	fmt.Println("  /read [options]     - Read emails")
	fmt.Println("    --unread          - Show only unread emails")
	fmt.Println("    --from <email>    - Filter by sender")
	fmt.Println("    --days <n>        - Emails from last n days")
	fmt.Println("    -n <number>       - Number of emails to read")
	fmt.Println()
	fmt.Println("  /send               - Interactive email composition")
	fmt.Println("  /report             - Generate email report")
	fmt.Println("  /delete             - Delete emails (interactive)")
	fmt.Println("  /unsubscribe        - Find unsubscribe links")
	fmt.Println("  /mark-read          - Mark emails as read")
	fmt.Println("  /template           - Manage email templates")
	fmt.Println("  /configure          - Configure email settings")
	fmt.Println("  /info               - Show current configuration")
	fmt.Println("  /help               - Show this help message")
	fmt.Println("  /exit               - Exit interactive mode")
	fmt.Println()
	fmt.Println("AI Queries:")
	fmt.Println("  Type any text without a slash to send it to your")
	fmt.Println("  configured AI provider for email-related assistance.")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  /read --unread")
	fmt.Println("  /send")
	fmt.Println("  Summarize my emails from today")
	fmt.Println("  Draft a reply to the last email from John")
}

// handleReadCommand handles the /read command
func handleReadCommand(args []string) error {
	opts := ReadOptions{
		Limit: 20, // Default
	}

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--unread":
			opts.UnreadOnly = true
		case "--from":
			if i+1 < len(args) {
				opts.FromAddress = args[i+1]
				i++
			}
		case "--days":
			if i+1 < len(args) {
				// Parse days - simplified for now
				opts.Since = getTimeFromDays(args[i+1])
				i++
			}
		case "-n":
			if i+1 < len(args) {
				// Parse number - simplified for now
				if n := parseNumber(args[i+1]); n > 0 {
					opts.Limit = n
				}
				i++
			}
		}
	}

	client, err := NewClient()
	if err != nil {
		return err
	}

	fmt.Println("Reading emails and drafts...")
	emails, err := client.ReadEmails(opts)
	if err != nil {
		return fmt.Errorf("failed to read emails: %v", err)
	}

	// Fetch drafts with the same options
	draftsOpts := opts
	draftsOpts.Limit = 10 // Fetch fewer drafts to balance with emails
	drafts, err := client.ReadDrafts(draftsOpts)
	if err != nil {
		fmt.Printf("Note: Could not read drafts: %v\n", err)
		// Continue without drafts
	}

	// Combine emails and drafts into a single list
	allItems := make([]*Email, 0, len(emails)+len(drafts))
	allItems = append(allItems, emails...)
	allItems = append(allItems, drafts...)

	fmt.Print(FormatEmailListWithDrafts(allItems, len(emails)))
	return nil
}

// handleSendCommand handles the /send command interactively
func handleSendCommand(args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n━━━ Compose Email ━━━")

	// Get recipient
	fmt.Print("To: ")
	to, _ := reader.ReadString('\n')
	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("recipient is required")
	}

	// Get subject
	fmt.Print("Subject: ")
	subject, _ := reader.ReadString('\n')
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return fmt.Errorf("subject is required")
	}

	// Get body
	fmt.Println("Body (press Enter twice to finish):")
	var bodyLines []string
	emptyCount := 0
	for {
		line, _ := reader.ReadString('\n')
		if line == "\n" {
			emptyCount++
			if emptyCount >= 2 {
				break
			}
		} else {
			emptyCount = 0
		}
		bodyLines = append(bodyLines, line)
	}
	body := strings.Join(bodyLines, "")
	body = strings.TrimSpace(body)

	// Confirm send
	fmt.Printf("\nReady to send email to: %s\n", to)
	fmt.Printf("Subject: %s\n", subject)
	fmt.Print("Send? (y/n): ")
	confirm, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
		fmt.Println("Email cancelled.")
		return nil
	}

	// Send email
	msg := &EmailMessage{
		To:      strings.Split(to, ","),
		Subject: subject,
		Body:    body,
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

// handleReportCommand handles the /report command
func handleReportCommand(args []string) error {
	selectedRange, err := SelectTimeRange()
	if err != nil {
		return fmt.Errorf("failed to select time range: %v", err)
	}

	client, err := NewClient()
	if err != nil {
		return err
	}

	fmt.Printf("Generating report for: %s\n", selectedRange.Name)

	opts := ReadOptions{
		Since: selectedRange.Since,
		Limit: 1000,
	}

	emails, err := client.ReadEmails(opts)
	if err != nil {
		return fmt.Errorf("failed to read emails: %v", err)
	}

	// Filter emails within the time range
	var filteredEmails []*Email
	for _, email := range emails {
		if email.Date.After(selectedRange.Since) && email.Date.Before(selectedRange.Until) {
			filteredEmails = append(filteredEmails, email)
		}
	}

	report := GenerateEmailReport(filteredEmails, *selectedRange)
	fmt.Println(report)
	return nil
}

// handleDeleteCommand handles the /delete command
func handleDeleteCommand(args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n━━━ Delete Emails ━━━")
	fmt.Println("Enter email IDs to delete (comma-separated) or")
	fmt.Print("enter sender email to delete all from sender: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return fmt.Errorf("no input provided")
	}

	client, err := NewClient()
	if err != nil {
		return err
	}

	// Check if it's email IDs or sender email
	if strings.Contains(input, "@") {
		// Delete by sender
		opts := ReadOptions{
			FromAddress: input,
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

		fmt.Printf("Found %d emails from %s\n", len(emails), input)
		fmt.Print("Delete all? (y/n): ")
		confirm, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
			fmt.Println("Deletion cancelled.")
			return nil
		}

		// Extract IDs
		var emailIds []uint32
		for _, email := range emails {
			emailIds = append(emailIds, email.ID)
		}

		if err := client.DeleteEmails(emailIds); err != nil {
			return fmt.Errorf("failed to delete emails: %v", err)
		}

		fmt.Printf("✓ Deleted %d email(s)\n", len(emails))
	} else {
		// Delete by IDs
		ids := parseEmailIDs(input)
		if len(ids) == 0 {
			return fmt.Errorf("invalid email IDs")
		}

		fmt.Printf("Delete %d email(s)? (y/n): ", len(ids))
		confirm, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
			fmt.Println("Deletion cancelled.")
			return nil
		}

		if err := client.DeleteEmails(ids); err != nil {
			return fmt.Errorf("failed to delete emails: %v", err)
		}

		fmt.Printf("✓ Deleted %d email(s)\n", len(ids))
	}

	return nil
}

// handleUnsubscribeCommand handles the /unsubscribe command
func handleUnsubscribeCommand(args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n━━━ Find Unsubscribe Links ━━━")
	fmt.Print("Enter sender email (or press Enter for all): ")

	from, _ := reader.ReadString('\n')
	from = strings.TrimSpace(from)

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

	fmt.Println("Searching for unsubscribe links...")
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
		fmt.Printf("\nOpen first link in browser? (y/n): ")
		confirm, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
			// Note: openBrowser function would need to be implemented
			fmt.Printf("Opening: %s\n", links[0].Links[0])
		}
	}

	return nil
}

// handleMarkReadCommand handles the /mark-read command
func handleMarkReadCommand(args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n━━━ Mark Emails as Read ━━━")
	fmt.Print("Enter email IDs (comma-separated): ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return fmt.Errorf("no email IDs provided")
	}

	ids := parseEmailIDs(input)
	if len(ids) == 0 {
		return fmt.Errorf("invalid email IDs")
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

// showInfo displays current configuration info
func showInfo() error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	config := client.GetConfig()
	
	// Email Configuration Section
	fmt.Println("Email Configuration (Global ~/.email/)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Location: ~/.email/config.json\n")
	fmt.Printf("  Provider: %s\n", client.GetProviderInfo())
	fmt.Printf("  Email: %s\n", config.Email)
	if config.DefaultAICLI != "" && config.DefaultAICLI != "none" {
		fmt.Printf("  AI CLI: %s\n", GetAICLIName(config.DefaultAICLI))
	} else {
		fmt.Printf("  AI CLI: Not configured (use /provider)\n")
	}
	if smtpHost, smtpPort, _, _, err := config.GetSMTPSettings(); err == nil {
		fmt.Printf("  SMTP: %s:%d\n", smtpHost, smtpPort)
	}
	if imapHost, imapPort, err := config.GetIMAPSettings(); err == nil {
		fmt.Printf("  IMAP: %s:%d\n", imapHost, imapPort)
	}
	if config.FromName != "" {
		fmt.Printf("  Display Name: %s\n", config.FromName)
	}

	fmt.Println("\nTip: Use 'mailos configure --local' to create a local config for this project")

	// Common Commands Section
	fmt.Println("\nCommon Commands")
	fmt.Println("━━━━━━━━━━━━━━━")
	fmt.Printf("  mailos                   Start interactive mode\n")
	fmt.Printf("  mailos read              Browse and read emails\n")
	fmt.Printf("  mailos send              Compose and send email\n")
	fmt.Printf("  mailos report            Generate email analytics\n")
	fmt.Printf("  mailos configure         Setup email configuration\n")
	fmt.Printf("  mailos configure --local Configure for current project\n")
	fmt.Printf("  mailos provider          Set AI provider (Claude, GPT, etc.)\n")
	fmt.Printf("  mailos help              Show detailed help\n")

	// Interactive Commands Section
	fmt.Println("\nInteractive Mode Commands")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  /read                    Browse emails interactively\n")
	fmt.Printf("  /send                    Compose new email\n")
	fmt.Printf("  /inbox                   Open inbox in browser\n")
	fmt.Printf("  /drafts                  Open drafts in browser\n")
	fmt.Printf("  /template                Manage email templates\n")
	fmt.Printf("  /unsubscribe             Find unsubscribe links\n")
	fmt.Printf("  /delete                  Delete emails by criteria\n")
	fmt.Printf("  /info                    Show this information\n")
	fmt.Printf("  /help                    Detailed help and shortcuts\n")

	// AI Navigation Section
	fmt.Println("\nAI Assistant Features")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Natural Language         Ask questions in plain English\n")
	fmt.Printf("  Email Summaries          'Summarize my emails from today'\n")
	fmt.Printf("  Draft Assistance         'Help me write a follow-up email'\n")
	fmt.Printf("  Email Analysis           'Find all emails about project X'\n")
	fmt.Printf("  File Integration         Use @ to reference files in queries\n")

	// Documentation and Resources
	fmt.Println("\nDocumentation & Resources")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Configuration Guide      mailos help configure\n")
	fmt.Printf("  AI Provider Setup        mailos help provider\n")
	fmt.Printf("  Command Reference        mailos help commands\n")
	fmt.Printf("  Troubleshooting          mailos help troubleshoot\n")
	fmt.Printf("  GitHub Repository        https://github.com/corp-os/emailos\n")

	// Environment Information
	fmt.Println("\nEnvironment")
	fmt.Println("━━━━━━━━━━━")
	fmt.Printf("  Current Directory        %s\n", getCurrentDirectory())
	if isGitRepo() {
		fmt.Printf("  Git Repository           Yes\n")
	}
	
	return nil
}

// Helper function to get current directory
func getCurrentDirectory() string {
	if dir, err := os.Getwd(); err == nil {
		return dir
	}
	return "Unknown"
}

// Helper function to check if current directory is a git repo
func isGitRepo() bool {
	if _, err := os.Stat(".git"); err == nil {
		return true
	}
	return false
}

// ShowEnhancedInfo displays comprehensive configuration and help information
func ShowEnhancedInfo() error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	config := client.GetConfig()
	
	// Email Configuration Section
	fmt.Println("Email Configuration (Global ~/.email/)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Location: ~/.email/config.json\n")
	fmt.Printf("  Provider: %s\n", client.GetProviderInfo())
	fmt.Printf("  Email: %s\n", config.Email)
	if config.DefaultAICLI != "" && config.DefaultAICLI != "none" {
		fmt.Printf("  AI CLI: %s\n", GetAICLIName(config.DefaultAICLI))
	} else {
		fmt.Printf("  AI CLI: Not configured (use /provider)\n")
	}
	if smtpHost, smtpPort, _, _, err := config.GetSMTPSettings(); err == nil {
		fmt.Printf("  SMTP: %s:%d\n", smtpHost, smtpPort)
	}
	if imapHost, imapPort, err := config.GetIMAPSettings(); err == nil {
		fmt.Printf("  IMAP: %s:%d\n", imapHost, imapPort)
	}
	if config.FromName != "" {
		fmt.Printf("  Display Name: %s\n", config.FromName)
	}

	fmt.Println("\nTip: Use 'mailos configure --local' to create a local config for this project")

	// Common Commands Section
	fmt.Println("\nCommon Commands")
	fmt.Println("━━━━━━━━━━━━━━━")
	fmt.Printf("  mailos                   Start interactive mode\n")
	fmt.Printf("  mailos read              Browse and read emails\n")
	fmt.Printf("  mailos send              Compose and send email\n")
	fmt.Printf("  mailos report            Generate email analytics\n")
	fmt.Printf("  mailos configure         Setup email configuration\n")
	fmt.Printf("  mailos configure --local Configure for current project\n")
	fmt.Printf("  mailos provider          Set AI provider (Claude, GPT, etc.)\n")
	fmt.Printf("  mailos help              Show detailed help\n")

	// Interactive Commands Section
	fmt.Println("\nInteractive Mode Commands")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  /read                    Browse emails interactively\n")
	fmt.Printf("  /send                    Compose new email\n")
	fmt.Printf("  /inbox                   Open inbox in browser\n")
	fmt.Printf("  /drafts                  Open drafts in browser\n")
	fmt.Printf("  /template                Manage email templates\n")
	fmt.Printf("  /unsubscribe             Find unsubscribe links\n")
	fmt.Printf("  /delete                  Delete emails by criteria\n")
	fmt.Printf("  /info                    Show this information\n")
	fmt.Printf("  /help                    Detailed help and shortcuts\n")

	// AI Navigation Section
	fmt.Println("\nAI Assistant Features")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Natural Language         Ask questions in plain English\n")
	fmt.Printf("  Email Summaries          'Summarize my emails from today'\n")
	fmt.Printf("  Draft Assistance         'Help me write a follow-up email'\n")
	fmt.Printf("  Email Analysis           'Find all emails about project X'\n")
	fmt.Printf("  File Integration         Use @ to reference files in queries\n")

	// Documentation and Resources
	fmt.Println("\nDocumentation & Resources")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Configuration Guide      mailos help configure\n")
	fmt.Printf("  AI Provider Setup        mailos help provider\n")
	fmt.Printf("  Command Reference        mailos help commands\n")
	fmt.Printf("  Troubleshooting          mailos help troubleshoot\n")
	fmt.Printf("  GitHub Repository        https://github.com/corp-os/emailos\n")

	// Environment Information
	fmt.Println("\nEnvironment")
	fmt.Println("━━━━━━━━━━━")
	fmt.Printf("  Current Directory        %s\n", getCurrentDirectory())
	if isGitRepo() {
		fmt.Printf("  Git Repository           Yes\n")
	}
	
	return nil
}

// Helper functions

func parseEmailIDs(input string) []uint32 {
	var ids []uint32
	parts := strings.Split(input, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if id := parseNumber(part); id > 0 {
			ids = append(ids, uint32(id))
		}
	}
	return ids
}

func parseNumber(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

func getTimeFromDays(days string) time.Time {
	n := parseNumber(days)
	if n <= 0 {
		n = 1
	}
	return time.Now().AddDate(0, 0, -n)
}
