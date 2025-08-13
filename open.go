package mailos

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// OpenEmailInMailApp opens an email in the user's default mail application
// using the message-id header. This works on macOS and Windows.
func OpenEmailInMailApp(messageID string) error {
	// Clean up the message ID (remove angle brackets if present)
	messageID = strings.Trim(messageID, "<>")
	
	// Construct the email URL using the message: URI scheme
	// This is supported by most mail clients
	emailURL := fmt.Sprintf("message:<%s>", messageID)
	
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "darwin":
		// macOS - use 'open' command
		cmd = exec.Command("open", emailURL)
	case "windows":
		// Windows - use 'start' command via cmd
		cmd = exec.Command("cmd", "/c", "start", "", emailURL)
	case "linux":
		// Linux - try xdg-open
		cmd = exec.Command("xdg-open", emailURL)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	
	// Execute the command
	if err := cmd.Run(); err != nil {
		// Try alternative approach with mailto: scheme as fallback
		return tryAlternativeOpen(messageID)
	}
	
	return nil
}

// tryAlternativeOpen tries alternative methods to open the email
func tryAlternativeOpen(messageID string) error {
	// Some mail clients support a special URL format
	// Try opening with a search query
	searchURL := fmt.Sprintf("mailto:?search=%s", messageID)
	
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", searchURL)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", searchURL)
	case "linux":
		cmd = exec.Command("xdg-open", searchURL)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	
	return cmd.Run()
}

// OpenEmailByID fetches an email by its IMAP sequence number and opens it
func OpenEmailByID(id uint32) error {
	// First, try to fetch from inbox
	opts := ReadOptions{
		Limit: 100, // Fetch enough to find the email
	}
	
	emails, err := Read(opts)
	if err == nil {
		// Find the email with the matching ID in inbox
		for _, email := range emails {
			if email.ID == id {
				fmt.Printf("Opening email: %s from %s\n", email.Subject, email.From)
				return openEmailInClient(email)
			}
		}
	}
	
	// If not found in inbox, try sent folder
	sentOpts := SentOptions{
		Limit: 100,
	}
	sentEmails, err := ReadSentEmails(sentOpts)
	if err == nil {
		for _, email := range sentEmails {
			if email.ID == id {
				fmt.Printf("Opening sent email: %s to %s\n", email.Subject, strings.Join(email.To, ", "))
				return openEmailInClient(email)
			}
		}
	}
	
	return fmt.Errorf("email with ID %d not found in inbox or sent folder", id)
}

// openEmailInClient opens a specific email in the mail client
func openEmailInClient(email *Email) error {
	switch runtime.GOOS {
	case "darwin":
		return openEmailMacOS(email)
	case "windows":
		return openEmailWindows(email)
	case "linux":
		return openEmailLinux(email)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// openEmailMacOS opens email on macOS using AppleScript
func openEmailMacOS(email *Email) error {
	// First try to open Mail.app
	openCmd := exec.Command("open", "-a", "Mail")
	if err := openCmd.Run(); err != nil {
		// If Mail.app doesn't exist, try opening default mail client
		openCmd = exec.Command("open", "mailto:")
		openCmd.Run()
	}
	
	// Wait a moment for Mail to open
	time.Sleep(500 * time.Millisecond)
	
	// Extract sender email from the From field
	senderEmail := extractEmailFromString(email.From)
	
	// Create AppleScript to find and display the email
	// We'll search by multiple criteria for better accuracy
	appleScript := fmt.Sprintf(`
on run
	tell application "Mail"
		activate
		
		-- Try to find the email by subject and sender
		set searchSubject to "%s"
		set searchFrom to "%s"
		
		-- Get all messages
		set allAccounts to every account
		set foundMessage to missing value
		
		repeat with eachAccount in allAccounts
			try
				set allMailboxes to every mailbox of eachAccount
				repeat with eachMailbox in allMailboxes
					try
						-- Search for messages with matching subject
						set matchingMessages to (every message of eachMailbox whose subject contains searchSubject)
						
						-- Filter by sender if we found matches
						repeat with eachMessage in matchingMessages
							set senderAddress to (extract address from sender of eachMessage)
							if senderAddress contains searchFrom then
								set foundMessage to eachMessage
								exit repeat
							end if
						end repeat
						
						if foundMessage is not missing value then
							exit repeat
						end if
					end try
				end repeat
				
				if foundMessage is not missing value then
					exit repeat
				end if
			end try
		end repeat
		
		-- If we found the message, display it
		if foundMessage is not missing value then
			try
				-- Open the message in a new window
				open foundMessage
			on error
				-- If that fails, try to select it in the message viewer
				try
					set selected messages of message viewer 1 to {foundMessage}
				end try
			end try
		else
			-- If not found, at least search for the subject
			-- This will show the search field with the query
			tell application "System Events"
				tell process "Mail"
					keystroke "f" using {command down, option down}
					delay 0.5
					keystroke searchSubject
				end tell
			end tell
		end if
	end tell
end run
`, escapeAppleScriptString(email.Subject), senderEmail)
	
	// Execute the AppleScript
	cmd := exec.Command("osascript", "-e", appleScript)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		// Fallback: just open Mail and show search instructions
		fmt.Printf("Mail.app opened. Please search for:\n")
		fmt.Printf("  Subject: %s\n", email.Subject)
		fmt.Printf("  From: %s\n", email.From)
		return nil
	}
	
	if len(output) > 0 {
		fmt.Printf("Mail.app response: %s\n", string(output))
	}
	
	return nil
}

// openEmailWindows opens email on Windows
func openEmailWindows(email *Email) error {
	// Try to create a search URL for Outlook
	searchQuery := url.QueryEscape(email.Subject)
	outlookURL := fmt.Sprintf("outlook://search?query=%s", searchQuery)
	
	cmd := exec.Command("cmd", "/c", "start", "", outlookURL)
	if err := cmd.Run(); err != nil {
		// Fallback to mailto with subject
		mailtoURL := fmt.Sprintf("mailto:?subject=%s", url.QueryEscape(email.Subject))
		cmd = exec.Command("cmd", "/c", "start", "", mailtoURL)
		cmd.Run()
		
		fmt.Printf("Mail client opened. Please search for:\n")
		fmt.Printf("  Subject: %s\n", email.Subject)
		fmt.Printf("  From: %s\n", email.From)
	}
	
	return nil
}

// openEmailLinux opens email on Linux
func openEmailLinux(email *Email) error {
	// Try Thunderbird-specific URL if available
	thunderbirdURL := fmt.Sprintf("thunderbird://search?query=%s", url.QueryEscape(email.Subject))
	
	cmd := exec.Command("xdg-open", thunderbirdURL)
	if err := cmd.Run(); err != nil {
		// Fallback to generic mailto
		mailtoURL := fmt.Sprintf("mailto:?subject=%s", url.QueryEscape(email.Subject))
		cmd = exec.Command("xdg-open", mailtoURL)
		cmd.Run()
		
		fmt.Printf("Mail client opened. Please search for:\n")
		fmt.Printf("  Subject: %s\n", email.Subject)
		fmt.Printf("  From: %s\n", email.From)
	}
	
	return nil
}

// openBySubjectSearch opens mail app with a search for the subject
func openBySubjectSearch(subject string) error {
	// Create a dummy email object with just the subject
	email := &Email{
		Subject: subject,
		From:    "",
	}
	
	return openEmailInClient(email)
}

// extractEmailFromString extracts the email address from a From field
// e.g., "John Doe <john@example.com>" -> "john@example.com"
func extractEmailFromString(from string) string {
	if strings.Contains(from, "<") && strings.Contains(from, ">") {
		start := strings.Index(from, "<")
		end := strings.Index(from, ">")
		if start < end {
			return from[start+1 : end]
		}
	}
	return from
}

// escapeAppleScriptString escapes a string for use in AppleScript
func escapeAppleScriptString(s string) string {
	// Escape backslashes first
	s = strings.ReplaceAll(s, `\`, `\\`)
	// Escape quotes
	s = strings.ReplaceAll(s, `"`, `\"`)
	// Escape newlines
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	return s
}

// OpenWebMailFolder opens a specific folder in the user's webmail interface
// based on their configured provider
func OpenWebMailFolder(folder string) error {
	// Load config to get the user's email provider and email address
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}
	
	if config.Provider == "" {
		return fmt.Errorf("no email provider configured")
	}
	
	if config.Email == "" {
		return fmt.Errorf("no email address configured")
	}
	
	// Get the URL for the specific folder based on provider
	url := getWebMailFolderURL(config.Provider, config.Email, folder)
	if url == "" {
		return fmt.Errorf("webmail not supported for provider: %s", config.Provider)
	}
	
	// Open the URL in the default browser
	return openURL(url)
}

// getWebMailFolderURL returns the appropriate webmail URL for the provider and folder
func getWebMailFolderURL(provider, email, folder string) string {
	// URL encode the email for use in query parameters
	encodedEmail := url.QueryEscape(email)
	
	switch provider {
	case ProviderGmail:
		switch folder {
		case "sent":
			return GmailSentURL
		case "inbox":
			return GmailInboxURL
		case "drafts":
			return GmailDraftsURL
		case "all":
			return GmailAllMailURL
		default:
			return GmailWebURL
		}
		
	case ProviderFastmail:
		switch folder {
		case "sent":
			return fmt.Sprintf("%s/search:from%%3A%s", FastmailSentURL, encodedEmail)
		case "inbox":
			return FastmailInboxURL
		case "drafts":
			return FastmailDraftsURL
		case "all":
			return FastmailAllMailURL
		default:
			return FastmailWebURL
		}
		
	case ProviderOutlook:
		switch folder {
		case "sent":
			return OutlookSentURL
		case "inbox":
			return OutlookInboxURL
		case "drafts":
			return OutlookDraftsURL
		case "all":
			return OutlookWebURL
		default:
			return OutlookWebURL
		}
		
	case ProviderYahoo:
		switch folder {
		case "sent":
			return YahooSentURL
		case "inbox":
			return YahooInboxURL
		case "drafts":
			return YahooDraftsURL
		case "all":
			return YahooWebURL
		default:
			return YahooWebURL
		}
		
	case ProviderZoho:
		switch folder {
		case "sent":
			return ZohoSentURL
		case "inbox":
			return ZohoInboxURL
		case "drafts":
			return ZohoDraftsURL
		case "all":
			return ZohoInboxURL
		default:
			return ZohoWebURL
		}
		
	default:
		// For unknown providers, try a generic mailto link
		return ""
	}
}

// openURL opens a URL in the system's default browser
func openURL(url string) error {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open URL: %v", err)
	}
	
	fmt.Printf("✓ Opened in browser: %s\n", url)
	return nil
}

// OpenSentMail opens the sent mail folder in the webmail interface
func OpenSentMail() error {
	return OpenWebMailFolder("sent")
}

// OpenInbox opens the inbox in the webmail interface
func OpenInbox() error {
	return OpenWebMailFolder("inbox")
}

// OpenDrafts opens the drafts folder in the webmail interface
func OpenDrafts() error {
	return OpenWebMailFolder("drafts")
}

// OpenAllMail opens all mail in the webmail interface
func OpenAllMail() error {
	return OpenWebMailFolder("all")
}