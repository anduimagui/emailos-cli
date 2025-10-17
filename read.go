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
	ID              uint32
	From            string
	To              []string
	Subject         string
	Date            time.Time
	Body            string
	BodyHTML        string
	Attachments     []string
	AttachmentData  map[string][]byte // Map of filename to attachment data
	MessageID       string             // Message-ID header for threading
	InReplyTo       string             // In-Reply-To header for threading
}

type ReadOptions struct {
	Limit            int
	UnreadOnly       bool
	FromAddress      string
	ToAddress        string
	Subject          string
	Since            time.Time
	LocalOnly        bool  // Only read from local storage
	SyncLocal        bool  // Sync received emails to local storage
	DownloadAttach   bool  // Download attachment content
	AttachmentDir    string // Directory to save attachments (if empty, returns in memory)
}

func Read(opts ReadOptions) ([]*Email, error) {
	// If LocalOnly is set, try to read from global inbox first
	if opts.LocalOnly {
		config, err := LoadConfig()
		if err == nil && config.Email != "" {
			emails, err := GetEmailsFromInbox(config.Email, opts)
			if err == nil {
				fmt.Printf("Read %d emails from global inbox\n", len(emails))
				return emails, nil
			}
		}
		// Fallback to old local storage method
		return readFromLocalStorage(opts)
	}
	
	// For live IMAP reading, also sync to global inbox if SyncLocal is set
	emails, err := ReadFromFolder(opts, "INBOX")
	if err == nil && opts.SyncLocal {
		// Auto-sync emails to global inbox
		config, configErr := LoadConfig()
		if configErr == nil && config.Email != "" {
			// Load existing inbox
			inboxData, inboxErr := LoadGlobalInbox(config.Email)
			if inboxErr == nil {
				// Add new emails
				inboxData.Emails = append(inboxData.Emails, emails...)
				// Remove duplicates and sort
				inboxData.Emails = removeDuplicateEmails(inboxData.Emails)
				// Save
				SaveGlobalInbox(config.Email, inboxData)
				fmt.Printf("Synced %d emails to global inbox\n", len(emails))
			}
		}
	}
	
	return emails, err
}

// ReadFromFolder reads emails from a specific IMAP folder
func ReadFromFolder(opts ReadOptions, folder string) ([]*Email, error) {
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

	// Select the specified folder
	_, err = c.Select(folder, false)
	if err != nil {
		// Try alternative folder names for Drafts
		if folder == "Drafts" {
			// Common draft folder names by provider
			draftFolders := []string{
				"[Gmail]/Drafts",     // Gmail
				"INBOX.Drafts",       // Some IMAP servers
				"Draft",              // Alternative singular
				"INBOX.Draft",        // Alternative singular with INBOX prefix
				"[Imap]/Drafts",      // Some providers
				"[Mail]/Drafts",      // Some providers
			}
			
			folderFound := false
			for _, draftFolder := range draftFolders {
				_, err = c.Select(draftFolder, false)
				if err == nil {
					folderFound = true
					break
				}
			}
			
			if !folderFound {
				// List all folders to help debug
				mailboxes := make(chan *imap.MailboxInfo, 10)
				done := make(chan error, 1)
				go func() {
					done <- c.List("", "*", mailboxes)
				}()
				
				var availableFolders []string
				for m := range mailboxes {
					availableFolders = append(availableFolders, m.Name)
					// Check if this might be a drafts folder
					if strings.Contains(strings.ToLower(m.Name), "draft") {
						// Try to select it
						_, err = c.Select(m.Name, false)
						if err == nil {
							folderFound = true
							break
						}
					}
				}
				<-done
				
				if !folderFound {
					// Return a more helpful error message
					return nil, fmt.Errorf("failed to find Drafts folder. Available folders: %v", availableFolders)
				}
			}
		} else if err != nil {
			return nil, fmt.Errorf("failed to select %s folder: %v", folder, err)
		}
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
		email, err := parseMessageWithOptions(msg, section, opts.DownloadAttach)
		if err != nil {
			continue
		}
		// Save attachments to disk if directory specified
		if opts.DownloadAttach && opts.AttachmentDir != "" && len(email.AttachmentData) > 0 {
			if err := saveAttachmentsToDisk(email, opts.AttachmentDir); err != nil {
				// Log error but don't fail the read
				fmt.Printf("Note: Could not save attachments: %v\n", err)
			}
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
	return parseMessageWithOptions(msg, section, false)
}

func parseMessageWithOptions(msg *imap.Message, section *imap.BodySectionName, downloadAttachments bool) (*Email, error) {
	if msg == nil {
		return nil, fmt.Errorf("message is nil")
	}

	email := &Email{
		ID:             msg.SeqNum,
		AttachmentData: make(map[string][]byte),
	}

	// Parse envelope
	if msg.Envelope != nil {
		email.Subject = msg.Envelope.Subject
		email.Date = msg.Envelope.Date
		email.MessageID = msg.Envelope.MessageId
		email.InReplyTo = msg.Envelope.InReplyTo

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
				// Read attachment data if requested
				if downloadAttachments {
					data, err := io.ReadAll(p.Body)
					if err == nil {
						email.AttachmentData[filename] = data
					}
				}
			}
		}
	}

	// If no plain text body, strip HTML tags from HTML body
	if email.Body == "" && email.BodyHTML != "" {
		email.Body = StripHTMLTags(email.BodyHTML)
	}

	return email, nil
}

func StripHTMLTags(html string) string {
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
	return DeleteEmailsFromFolder(ids, "INBOX")
}

// DeleteDrafts deletes the given draft IDs from the Drafts folder
func DeleteDrafts(ids []uint32) error {
	return DeleteEmailsFromFolder(ids, "Drafts")
}

// DeleteEmailsFromFolder deletes emails from a specific folder
func DeleteEmailsFromFolder(ids []uint32, folder string) error {
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

	// Select the specified folder
	_, err = c.Select(folder, false)
	if err != nil {
		// Try with [Gmail]/Drafts for Gmail
		if folder == "Drafts" && config.Provider == ProviderGmail {
			_, err = c.Select("[Gmail]/Drafts", false)
		}
		if err != nil {
			return fmt.Errorf("failed to select %s folder: %v", folder, err)
		}
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

// saveAttachmentsToDisk saves email attachments to the specified directory
func saveAttachmentsToDisk(email *Email, dir string) error {
	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create attachment directory: %v", err)
	}
	
	for filename, data := range email.AttachmentData {
		// Sanitize filename
		safeFilename := strings.ReplaceAll(filename, "/", "_")
		safeFilename = strings.ReplaceAll(safeFilename, "..", "_")
		
		// Create a unique filename with timestamp
		timestamp := email.Date.Format("20060102_150405")
		finalFilename := fmt.Sprintf("%s_%s", timestamp, safeFilename)
		
		filepath := filepath.Join(dir, finalFilename)
		
		// Write attachment to file
		if err := os.WriteFile(filepath, data, 0644); err != nil {
			return fmt.Errorf("failed to save attachment %s: %v", filename, err)
		}
		
		fmt.Printf("Saved attachment: %s\n", filepath)
	}
	
	return nil
}