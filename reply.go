package mailos

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type ReplyOptions struct {
	EmailNumber int      // User-friendly email number from list
	EmailUID    uint32   // IMAP UID if known
	MessageID   string   // Message-ID to reply to
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	Body        string
	FileBody    string   // Read body from file
	ReplyAll    bool     // Reply to all recipients
	Interactive bool     // Interactive mode
	Draft       bool     // Save as draft instead of sending
}

func ReplyCommand(opts ReplyOptions) error {
	var originalEmail *Email
	var err error

	// Find the original email to reply to
	if opts.EmailNumber > 0 {
		originalEmail, err = findEmailByNumber(opts.EmailNumber)
		if err != nil {
			return fmt.Errorf("failed to find email #%d: %v", opts.EmailNumber, err)
		}
	} else if opts.EmailUID > 0 {
		originalEmail, err = findEmailByUID(opts.EmailUID)
		if err != nil {
			return fmt.Errorf("failed to find email with UID %d: %v", opts.EmailUID, err)
		}
	} else if opts.MessageID != "" {
		originalEmail, err = findEmailByMessageID(opts.MessageID)
		if err != nil {
			return fmt.Errorf("failed to find email with Message-ID %s: %v", opts.MessageID, err)
		}
	} else {
		return fmt.Errorf("must specify either email number, UID, or Message-ID to reply to")
	}

	if originalEmail == nil {
		return fmt.Errorf("original email not found")
	}

	fmt.Printf("📧 Replying to: %s\n", originalEmail.Subject)
	fmt.Printf("   From: %s\n", originalEmail.From)
	fmt.Printf("   Date: %s\n", originalEmail.Date.Format("Jan 2, 2006 at 3:04 PM"))
	fmt.Printf("   Message-ID: %s\n", originalEmail.MessageID)

	// Prepare reply
	reply := DraftEmail{}

	// Set threading headers
	if originalEmail.MessageID != "" {
		reply.InReplyTo = originalEmail.MessageID
		// Build References chain
		if originalEmail.InReplyTo != "" {
			// This was already a reply, so add to the chain
			reply.References = []string{originalEmail.InReplyTo, originalEmail.MessageID}
		} else {
			// This is the first reply in the thread
			reply.References = []string{originalEmail.MessageID}
		}
	}

	// Set recipients
	if opts.ReplyAll {
		// Reply to all - include original sender and all original recipients
		reply.To = []string{extractEmailAddress(originalEmail.From)}
		
		// Add original To recipients (excluding our own address)
		config, _ := LoadConfig()
		ourEmail := config.Email
		if config.FromEmail != "" {
			ourEmail = config.FromEmail
		}
		
		for _, to := range originalEmail.To {
			emailAddr := extractEmailAddress(to)
			if !strings.EqualFold(emailAddr, ourEmail) && !contains(reply.To, emailAddr) {
				reply.To = append(reply.To, emailAddr)
			}
		}
	} else {
		// Reply only to sender
		reply.To = []string{extractEmailAddress(originalEmail.From)}
	}

	// Override recipients if specified
	if len(opts.To) > 0 {
		reply.To = opts.To
	}
	if len(opts.CC) > 0 {
		reply.CC = opts.CC
	}
	if len(opts.BCC) > 0 {
		reply.BCC = opts.BCC
	}

	// Set subject
	subject := originalEmail.Subject
	if !strings.HasPrefix(strings.ToLower(subject), "re:") {
		subject = "Re: " + subject
	}
	if opts.Subject != "" {
		subject = opts.Subject
	}
	reply.Subject = subject

	// Set body
	if opts.FileBody != "" {
		fileContent, err := os.ReadFile(opts.FileBody)
		if err != nil {
			return fmt.Errorf("failed to read body from file %s: %v", opts.FileBody, err)
		}
		reply.Body = string(fileContent)
	} else if opts.Body != "" {
		reply.Body = opts.Body
	} else if opts.Interactive {
		// Interactive composition
		body, err := composeReplyInteractively(originalEmail)
		if err != nil {
			return fmt.Errorf("failed to compose reply: %v", err)
		}
		reply.Body = body
	} else {
		// Default: create a basic reply template
		reply.Body = createReplyTemplate(originalEmail)
	}

	// Save or send the reply
	if opts.Draft {
		// Save as draft
		uid, err := saveDraftToIMAP(reply)
		if err != nil {
			return fmt.Errorf("failed to save reply as draft: %v", err)
		}
		fmt.Printf("✓ Reply saved as draft (UID: %d)\n", uid)
		
		// Also save to local drafts
		if err := saveLocalDraft(reply); err != nil {
			fmt.Printf("⚠️  Could not save to local drafts: %v\n", err)
		}
		
		return nil
	} else {
		// Send the reply
		msg := &EmailMessage{
			To:         reply.To,
			CC:         reply.CC,
			BCC:        reply.BCC,
			Subject:    reply.Subject,
			Body:       reply.Body,
			InReplyTo:  reply.InReplyTo,
			References: reply.References,
		}
		
		fmt.Printf("📤 Sending reply...\n")
		if err := Send(msg); err != nil {
			return fmt.Errorf("failed to send reply: %v", err)
		}
		fmt.Printf("✓ Reply sent successfully!\n")
		return nil
	}
}

func findEmailByNumber(number int) (*Email, error) {
	// Read recent emails and find by number (1-indexed)
	opts := ReadOptions{
		Limit:     50, // Get recent emails
		LocalOnly: false,
	}
	
	emails, err := Read(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to read emails: %v", err)
	}
	
	if number < 1 || number > len(emails) {
		return nil, fmt.Errorf("email #%d not found (available: 1-%d)", number, len(emails))
	}
	
	return emails[number-1], nil
}

func findEmailByUID(uid uint32) (*Email, error) {
	// For now, read recent emails and find by ID (which maps to UID for IMAP emails)
	opts := ReadOptions{
		Limit:     100,
		LocalOnly: false,
	}
	
	emails, err := Read(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to read emails: %v", err)
	}
	
	for _, email := range emails {
		if email.ID == uid {
			return email, nil
		}
	}
	
	return nil, fmt.Errorf("email with UID %d not found", uid)
}

func findEmailByMessageID(messageID string) (*Email, error) {
	// Read emails and find by Message-ID
	opts := ReadOptions{
		Limit:     100,
		LocalOnly: false,
	}
	
	emails, err := Read(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to read emails: %v", err)
	}
	
	// Normalize Message-ID (remove angle brackets if present)
	normalizeID := strings.Trim(messageID, "<>")
	
	for _, email := range emails {
		emailMsgID := strings.Trim(email.MessageID, "<>")
		if emailMsgID == normalizeID {
			return email, nil
		}
	}
	
	return nil, fmt.Errorf("email with Message-ID %s not found", messageID)
}


func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

func composeReplyInteractively(originalEmail *Email) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("\n" + strings.Repeat("─", 60))
	fmt.Println("📝 Compose your reply:")
	fmt.Println("   (Press Enter twice to finish)")
	fmt.Println(strings.Repeat("─", 60))
	
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
	
	// Add original message quote
	quote := createOriginalMessageQuote(originalEmail)
	if body != "" {
		body = body + "\n\n" + quote
	} else {
		body = quote
	}
	
	return body, nil
}

func createReplyTemplate(originalEmail *Email) string {
	// Create a basic reply with quoted original message
	template := fmt.Sprintf("\n\n%s", createOriginalMessageQuote(originalEmail))
	return template
}

func createOriginalMessageQuote(originalEmail *Email) string {
	// Create a quoted version of the original message
	var quote strings.Builder
	
	quote.WriteString(fmt.Sprintf("On %s, %s wrote:\n", 
		originalEmail.Date.Format("Jan 2, 2006 at 3:04 PM"),
		originalEmail.From))
	
	// Quote the original body
	lines := strings.Split(originalEmail.Body, "\n")
	for _, line := range lines {
		quote.WriteString("> " + line + "\n")
	}
	
	return quote.String()
}

// ReplyToEmail is a helper function that can be called with just an email number
func ReplyToEmail(emailNumber int, interactive bool) error {
	opts := ReplyOptions{
		EmailNumber: emailNumber,
		Interactive: interactive,
		Draft:       false,
	}
	return ReplyCommand(opts)
}

// ReplyToEmailAll replies to all recipients
func ReplyToEmailAll(emailNumber int, interactive bool) error {
	opts := ReplyOptions{
		EmailNumber: emailNumber,
		Interactive: interactive,
		ReplyAll:    true,
		Draft:       false,
	}
	return ReplyCommand(opts)
}

// DraftReplyToEmail creates a reply draft
func DraftReplyToEmail(emailNumber int, interactive bool) error {
	opts := ReplyOptions{
		EmailNumber: emailNumber,
		Interactive: interactive,
		Draft:       true,
	}
	return ReplyCommand(opts)
}