package mailos

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

type Email struct {
	ID          uint32
	From        string
	To          []string
	Subject     string
	Date        time.Time
	Body        string
	BodyHTML    string
	Attachments []string
}

type ReadOptions struct {
	Limit       int
	UnreadOnly  bool
	FromAddress string
	ToAddress   string
	Subject     string
	Since       time.Time
	LocalOnly   bool  // Only read from local storage
	SyncLocal   bool  // Sync received emails to local storage
}

func Read(opts ReadOptions) ([]*Email, error) {
	// If local only, read from local storage
	if opts.LocalOnly {
		return readFromLocalStorage(opts)
	}
	
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	// Get IMAP settings from provider
	imapHost, imapPort, err := config.GetIMAPSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to get IMAP settings: %v", err)
	}

	// Connect to IMAP server
	var c *client.Client
	if imapPort == 993 {
		// Use TLS
		tlsConfig := &tls.Config{ServerName: imapHost}
		c, err = client.DialTLS(fmt.Sprintf("%s:%d", imapHost, imapPort), tlsConfig)
	} else {
		// Use plain connection
		c, err = client.Dial(fmt.Sprintf("%s:%d", imapHost, imapPort))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to IMAP server: %v", err)
	}
	defer c.Logout()

	// Login
	if err := c.Login(config.Email, config.Password); err != nil {
		return nil, fmt.Errorf("failed to login: %v", err)
	}

	// Select inbox
	_, err = c.Select("INBOX", false)
	if err != nil {
		return nil, fmt.Errorf("failed to select inbox: %v", err)
	}

	// Build search criteria
	criteria := imap.NewSearchCriteria()
	if opts.UnreadOnly {
		criteria.WithoutFlags = []string{imap.SeenFlag}
	}
	if opts.FromAddress != "" {
		criteria.Header.Add("From", opts.FromAddress)
	}
	// Use config.FromEmail if no ToAddress is explicitly specified
	toAddress := opts.ToAddress
	if toAddress == "" && config.FromEmail != "" {
		toAddress = config.FromEmail
	}
	if toAddress != "" {
		criteria.Header.Add("To", toAddress)
	}
	if opts.Subject != "" {
		criteria.Header.Add("Subject", opts.Subject)
	}
	if !opts.Since.IsZero() {
		criteria.Since = opts.Since
	}

	// Search for messages
	ids, err := c.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %v", err)
	}

	// Limit results
	if opts.Limit > 0 && len(ids) > opts.Limit {
		// Get the most recent messages
		ids = ids[len(ids)-opts.Limit:]
	}

	if len(ids) == 0 {
		return []*Email{}, nil
	}

	// Create sequence set
	seqSet := new(imap.SeqSet)
	for i := len(ids) - 1; i >= 0; i-- {
		seqSet.AddNum(ids[i])
	}

	// Fetch messages
	messages := make(chan *imap.Message, len(ids))
	section := &imap.BodySectionName{}
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqSet, []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, section.FetchItem()}, messages)
	}()

	emails := make([]*Email, 0, len(ids))
	for msg := range messages {
		email, err := parseMessage(msg, section)
		if err != nil {
			continue
		}
		emails = append(emails, email)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %v", err)
	}

	// Reverse to get newest first
	for i, j := 0, len(emails)-1; i < j; i, j = i+1, j-1 {
		emails[i], emails[j] = emails[j], emails[i]
	}
	
	// Save to local storage if requested
	if opts.SyncLocal {
		for _, email := range emails {
			if err := saveReceivedEmail(email); err != nil {
				// Log error but don't fail the read
				fmt.Printf("Note: Could not save email to local storage: %v\n", err)
			}
		}
	}

	return emails, nil
}

func parseMessage(msg *imap.Message, section *imap.BodySectionName) (*Email, error) {
	if msg == nil {
		return nil, fmt.Errorf("message is nil")
	}

	email := &Email{
		ID: msg.SeqNum,
	}

	// Parse envelope
	if msg.Envelope != nil {
		email.Subject = msg.Envelope.Subject
		email.Date = msg.Envelope.Date

		// Parse From
		if len(msg.Envelope.From) > 0 {
			addr := msg.Envelope.From[0]
			if addr.PersonalName != "" {
				email.From = fmt.Sprintf("%s <%s@%s>", addr.PersonalName, addr.MailboxName, addr.HostName)
			} else {
				email.From = fmt.Sprintf("%s@%s", addr.MailboxName, addr.HostName)
			}
		}

		// Parse To
		email.To = make([]string, 0, len(msg.Envelope.To))
		for _, addr := range msg.Envelope.To {
			if addr.PersonalName != "" {
				email.To = append(email.To, fmt.Sprintf("%s <%s@%s>", addr.PersonalName, addr.MailboxName, addr.HostName))
			} else {
				email.To = append(email.To, fmt.Sprintf("%s@%s", addr.MailboxName, addr.HostName))
			}
		}
	}

	// Parse body
	r := msg.GetBody(section)
	if r == nil {
		return email, nil
	}

	mr, err := mail.CreateReader(r)
	if err != nil {
		return email, nil
	}

	// Process message parts
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			break
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			// Read body
			b, _ := io.ReadAll(p.Body)
			contentType, _, _ := h.ContentType()
			
			switch contentType {
			case "text/plain":
				email.Body = string(b)
			case "text/html":
				email.BodyHTML = string(b)
			}
		case *mail.AttachmentHeader:
			// Get attachment filename
			filename, _ := h.Filename()
			if filename != "" {
				email.Attachments = append(email.Attachments, filename)
			}
		}
	}

	// If no plain text body, strip HTML tags from HTML body
	if email.Body == "" && email.BodyHTML != "" {
		email.Body = stripHTMLTags(email.BodyHTML)
	}

	return email, nil
}

func stripHTMLTags(html string) string {
	// Very basic HTML tag stripping
	// In production, you'd want to use a proper HTML parser
	result := html
	result = strings.ReplaceAll(result, "<br>", "\n")
	result = strings.ReplaceAll(result, "<br/>", "\n")
	result = strings.ReplaceAll(result, "<br />", "\n")
	result = strings.ReplaceAll(result, "</p>", "\n\n")
	result = strings.ReplaceAll(result, "</div>", "\n")
	
	// Remove all other tags
	for strings.Contains(result, "<") && strings.Contains(result, ">") {
		start := strings.Index(result, "<")
		end := strings.Index(result[start:], ">")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}
	
	// Clean up whitespace
	lines := strings.Split(result, "\n")
	cleanLines := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}
	
	return strings.Join(cleanLines, "\n")
}

func MarkAsRead(ids []uint32) error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Get IMAP settings from provider
	imapHost, imapPort, err := config.GetIMAPSettings()
	if err != nil {
		return fmt.Errorf("failed to get IMAP settings: %v", err)
	}

	// Connect to IMAP server
	var c *client.Client
	if imapPort == 993 {
		tlsConfig := &tls.Config{ServerName: imapHost}
		c, err = client.DialTLS(fmt.Sprintf("%s:%d", imapHost, imapPort), tlsConfig)
	} else {
		c, err = client.Dial(fmt.Sprintf("%s:%d", imapHost, imapPort))
	}
	if err != nil {
		return fmt.Errorf("failed to connect to IMAP server: %v", err)
	}
	defer c.Logout()

	// Login
	if err := c.Login(config.Email, config.Password); err != nil {
		return fmt.Errorf("failed to login: %v", err)
	}

	// Select inbox
	_, err = c.Select("INBOX", false)
	if err != nil {
		return fmt.Errorf("failed to select inbox: %v", err)
	}

	// Create sequence set
	seqSet := new(imap.SeqSet)
	for _, id := range ids {
		seqSet.AddNum(id)
	}

	// Mark as read
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.SeenFlag}
	if err := c.Store(seqSet, item, flags, nil); err != nil {
		return fmt.Errorf("failed to mark messages as read: %v", err)
	}

	return nil
}

func DeleteEmails(ids []uint32) error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Get IMAP settings from provider
	imapHost, imapPort, err := config.GetIMAPSettings()
	if err != nil {
		return fmt.Errorf("failed to get IMAP settings: %v", err)
	}

	// Connect to IMAP server
	var c *client.Client
	if imapPort == 993 {
		tlsConfig := &tls.Config{ServerName: imapHost}
		c, err = client.DialTLS(fmt.Sprintf("%s:%d", imapHost, imapPort), tlsConfig)
	} else {
		c, err = client.Dial(fmt.Sprintf("%s:%d", imapHost, imapPort))
	}
	if err != nil {
		return fmt.Errorf("failed to connect to IMAP server: %v", err)
	}
	defer c.Logout()

	// Login
	if err := c.Login(config.Email, config.Password); err != nil {
		return fmt.Errorf("failed to login: %v", err)
	}

	// Select inbox
	_, err = c.Select("INBOX", false)
	if err != nil {
		return fmt.Errorf("failed to select inbox: %v", err)
	}

	// Create sequence set
	seqSet := new(imap.SeqSet)
	for _, id := range ids {
		seqSet.AddNum(id)
	}

	// Mark as deleted
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.DeletedFlag}
	if err := c.Store(seqSet, item, flags, nil); err != nil {
		return fmt.Errorf("failed to mark messages for deletion: %v", err)
	}

	// Expunge to permanently delete
	if err := c.Expunge(nil); err != nil {
		return fmt.Errorf("failed to expunge deleted messages: %v", err)
	}

	return nil
}

// readFromLocalStorage reads emails from the local .email/received directory
func readFromLocalStorage(opts ReadOptions) ([]*Email, error) {
	// Ensure directories exist
	if err := EnsureEmailDirectories(); err != nil {
		return nil, fmt.Errorf("failed to create email directories: %v", err)
	}
	
	// Get received directory
	receivedDir, err := GetReceivedDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get received directory: %v", err)
	}
	
	// Read all JSON files from the directory
	files, err := os.ReadDir(receivedDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Email{}, nil
		}
		return nil, fmt.Errorf("failed to read received directory: %v", err)
	}
	
	var emails []*Email
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		
		filepath := filepath.Join(receivedDir, file.Name())
		data, err := os.ReadFile(filepath)
		if err != nil {
			continue
		}
		
		var savedEmail SavedEmail
		if err := json.Unmarshal(data, &savedEmail); err != nil {
			continue
		}
		
		// Convert SavedEmail to Email
		email := &Email{
			ID:          0, // Local emails don't have IMAP IDs
			From:        savedEmail.From,
			To:          savedEmail.To,
			Subject:     savedEmail.Subject,
			Date:        savedEmail.Date,
			Body:        savedEmail.Body,
			BodyHTML:    savedEmail.BodyHTML,
			Attachments: savedEmail.Attachments,
		}
		
		// Apply filters
		if opts.FromAddress != "" && !strings.Contains(strings.ToLower(email.From), strings.ToLower(opts.FromAddress)) {
			continue
		}
		if opts.Subject != "" && !strings.Contains(strings.ToLower(email.Subject), strings.ToLower(opts.Subject)) {
			continue
		}
		if !opts.Since.IsZero() && email.Date.Before(opts.Since) {
			continue
		}
		
		emails = append(emails, email)
	}
	
	// Sort by date (newest first)
	sort.Slice(emails, func(i, j int) bool {
		return emails[i].Date.After(emails[j].Date)
	})
	
	// Apply limit
	if opts.Limit > 0 && len(emails) > opts.Limit {
		emails = emails[:opts.Limit]
	}
	
	return emails, nil
}

// saveReceivedEmail saves an email to the local .email/received directory
func saveReceivedEmail(email *Email) error {
	// Ensure directories exist
	if err := EnsureEmailDirectories(); err != nil {
		return fmt.Errorf("failed to create email directories: %v", err)
	}
	
	// Get received directory
	receivedDir, err := GetReceivedDir()
	if err != nil {
		return fmt.Errorf("failed to get received directory: %v", err)
	}
	
	// Create a SavedEmail struct
	savedEmail := SavedEmail{
		ID:          fmt.Sprintf("%d_%d", email.ID, email.Date.Unix()),
		From:        email.From,
		To:          email.To,
		Subject:     email.Subject,
		Body:        email.Body,
		BodyHTML:    email.BodyHTML,
		Date:        email.Date,
		Attachments: email.Attachments,
	}
	
	// Generate filename with timestamp
	filename := fmt.Sprintf("%s_%s_%s.json",
		email.Date.Format("20060102_150405"),
		strings.ReplaceAll(strings.ReplaceAll(email.From, "/", "_"), " ", "_"),
		strings.ReplaceAll(strings.ReplaceAll(email.Subject, "/", "_"), " ", "_"))
	
	// Ensure filename is not too long
	if len(filename) > 150 {
		filename = filename[:150] + ".json"
	}
	
	// Clean filename of problematic characters
	filename = strings.ReplaceAll(filename, "<", "")
	filename = strings.ReplaceAll(filename, ">", "")
	filename = strings.ReplaceAll(filename, ":", "")
	filename = strings.ReplaceAll(filename, "\"", "")
	filename = strings.ReplaceAll(filename, "|", "")
	filename = strings.ReplaceAll(filename, "?", "")
	filename = strings.ReplaceAll(filename, "*", "")
	
	filepath := filepath.Join(receivedDir, filename)
	
	// Check if file already exists
	if _, err := os.Stat(filepath); err == nil {
		// File already exists, skip
		return nil
	}
	
	// Marshal to JSON
	data, err := json.MarshalIndent(savedEmail, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal email: %v", err)
	}
	
	// Write to file
	if err := os.WriteFile(filepath, data, 0600); err != nil {
		return fmt.Errorf("failed to write email file: %v", err)
	}
	
	return nil
}