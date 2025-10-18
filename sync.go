package mailos

import (
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

type SyncOptions struct {
	BaseDir      string
	Limit        int
	Since        time.Time
	IncludeRead  bool
	Verbose      bool
}

func SyncEmails(opts SyncOptions) error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Use new global inbox system for primary account
	if config.Email != "" {
		fmt.Printf("🚀 Using new global inbox system for %s...\n", config.Email)
		if err := FetchEmailsIncremental(config, opts.Limit); err != nil {
			fmt.Printf("❌ Warning: Failed to sync to global inbox: %v\n", err)
			fmt.Printf("📦 Falling back to legacy sync method...\n")
			// Continue with legacy sync as fallback
		} else {
			fmt.Printf("✅ Global inbox sync completed successfully!\n")
			// Also update legacy sync time
			if err := UpdateLastSyncTime(); err != nil && opts.Verbose {
				fmt.Printf("Warning: failed to update last sync time: %v\n", err)
			}
			return nil
		}
	}

	// Set default base directory from config or use default
	if opts.BaseDir == "" {
		if config.SyncDir != "" {
			opts.BaseDir = config.SyncDir
		} else {
			// Use .email folder (same as GetEmailStorageDir)
			baseDir, err := GetEmailStorageDir()
			if err != nil {
				opts.BaseDir = ".email"
			} else {
				opts.BaseDir = baseDir
			}
		}
	}

	// Create directory structure
	receivedDir := filepath.Join(opts.BaseDir, "received")
	sentDir := filepath.Join(opts.BaseDir, "sent")
	draftsDir := filepath.Join(opts.BaseDir, "drafts")

	// Ensure directories exist and add to .gitignore if needed
	if err := EnsureEmailDirectories(); err != nil {
		return fmt.Errorf("failed to create email directories: %v", err)
	}
	
	for _, dir := range []string{receivedDir, sentDir, draftsDir} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	// Get IMAP settings from provider
	imapHost, imapPort, err := config.GetIMAPSettings()
	if err != nil {
		return fmt.Errorf("failed to get IMAP settings: %v", err)
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
		return fmt.Errorf("failed to connect to IMAP server: %v", err)
	}
	defer c.Logout()

	// Login
	if err := c.Login(config.Email, config.Password); err != nil {
		return fmt.Errorf("failed to login: %v", err)
	}

	// Sync inbox/received emails
	if opts.Verbose {
		fmt.Println("Syncing inbox emails...")
	}
	receivedCount, err := syncFolder(c, "INBOX", receivedDir, opts)
	if err != nil {
		return fmt.Errorf("failed to sync inbox: %v", err)
	}

	// Sync sent emails
	if opts.Verbose {
		fmt.Println("Syncing sent emails...")
	}
	sentCount, err := syncSentFolder(c, sentDir, opts)
	if err != nil {
		return fmt.Errorf("failed to sync sent folder: %v", err)
	}

	// Sync drafts
	if opts.Verbose {
		fmt.Println("Syncing draft emails...")
	}
	draftsCount, err := syncDraftsFolder(c, draftsDir, opts)
	if err != nil {
		return fmt.Errorf("failed to sync drafts folder: %v", err)
	}

	fmt.Printf("Sync complete!\n")
	fmt.Printf("  Received: %d emails\n", receivedCount)
	fmt.Printf("  Sent: %d emails\n", sentCount)
	fmt.Printf("  Drafts: %d emails\n", draftsCount)
	fmt.Printf("Files saved to: %s\n", opts.BaseDir)

	// Update last sync time in config
	if err := UpdateLastSyncTime(); err != nil && opts.Verbose {
		fmt.Printf("Warning: failed to update last sync time: %v\n", err)
	}

	return nil
}


func syncFolder(c *client.Client, folderName, outputDir string, opts SyncOptions) (int, error) {
	_, err := c.Select(folderName, false)
	if err != nil {
		return 0, fmt.Errorf("failed to select folder %s: %v", folderName, err)
	}

	// Build search criteria
	criteria := imap.NewSearchCriteria()
	if !opts.IncludeRead {
		criteria.WithoutFlags = []string{imap.SeenFlag}
	}
	if !opts.Since.IsZero() {
		criteria.Since = opts.Since
	}

	// Search for messages
	ids, err := c.Search(criteria)
	if err != nil {
		return 0, fmt.Errorf("failed to search messages: %v", err)
	}

	// Limit results
	if opts.Limit > 0 && len(ids) > opts.Limit {
		ids = ids[len(ids)-opts.Limit:]
	}

	if len(ids) == 0 {
		return 0, nil
	}

	// Create sequence set
	seqSet := new(imap.SeqSet)
	for _, id := range ids {
		seqSet.AddNum(id)
	}

	// Fetch messages
	messages := make(chan *imap.Message, len(ids))
	section := &imap.BodySectionName{}
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqSet, []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, section.FetchItem()}, messages)
	}()

	count := 0
	for msg := range messages {
		email, err := parseMessageForSync(msg, section)
		if err != nil {
			if opts.Verbose {
				fmt.Printf("Warning: failed to parse message: %v\n", err)
			}
			continue
		}

		// Save email to file
		if err := saveEmailToFile(email, outputDir); err != nil {
			if opts.Verbose {
				fmt.Printf("Warning: failed to save email: %v\n", err)
			}
			continue
		}
		count++
	}

	if err := <-done; err != nil {
		return count, fmt.Errorf("failed to fetch messages: %v", err)
	}

	return count, nil
}

func syncSentFolder(c *client.Client, outputDir string, opts SyncOptions) (int, error) {
	// Try common sent folder names
	sentFolderNames := []string{"Sent", "Sent Items", "Sent Messages", "[Gmail]/Sent Mail", "INBOX.Sent"}
	
	for _, folderName := range sentFolderNames {
		count, err := syncFolder(c, folderName, outputDir, opts)
		if err == nil {
			return count, nil
		}
	}

	// If no sent folder found, list available folders for debugging
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()
	
	availableFolders := []string{}
	for mbox := range mailboxes {
		availableFolders = append(availableFolders, mbox.Name)
	}
	
	if err := <-done; err != nil {
		return 0, fmt.Errorf("failed to list mailboxes: %v", err)
	}

	if opts.Verbose {
		fmt.Printf("Warning: Could not find sent folder. Available folders: %v\n", availableFolders)
	}
	
	return 0, nil
}

func syncDraftsFolder(c *client.Client, outputDir string, opts SyncOptions) (int, error) {
	// Try common drafts folder names
	draftsFolderNames := []string{"Drafts", "Draft", "[Gmail]/Drafts", "INBOX.Drafts"}
	
	for _, folderName := range draftsFolderNames {
		count, err := syncFolder(c, folderName, outputDir, opts)
		if err == nil {
			return count, nil
		}
	}

	if opts.Verbose {
		fmt.Println("Warning: Could not find drafts folder")
	}
	
	return 0, nil
}

func parseMessageForSync(msg *imap.Message, section *imap.BodySectionName) (*Email, error) {
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
	if r != nil {
		m, err := mail.CreateReader(r)
		if err != nil {
			// Try to read as plain text
			body, _ := io.ReadAll(r)
			email.Body = string(body)
		} else {
			// Process parts
			for {
				p, err := m.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					continue
				}

				switch h := p.Header.(type) {
				case *mail.InlineHeader:
					// This is the body
					b, _ := io.ReadAll(p.Body)
					contentType, _, _ := h.ContentType()
					if strings.HasPrefix(contentType, "text/plain") {
						email.Body = string(b)
					} else if strings.HasPrefix(contentType, "text/html") && email.Body == "" {
						email.BodyHTML = string(b)
					}
				case *mail.AttachmentHeader:
					// This is an attachment
					filename, _ := h.Filename()
					if filename != "" {
						email.Attachments = append(email.Attachments, filename)
					}
				}
			}
		}
	}

	// If we only have HTML body, use that as the main body
	if email.Body == "" && email.BodyHTML != "" {
		email.Body = email.BodyHTML
	}

	return email, nil
}

func saveEmailToFile(email *Email, outputDir string) error {
	// Create filename from subject and date
	subject := sanitizeFilename(email.Subject)
	if subject == "" {
		subject = "No-Subject"
	}
	
	dateStr := email.Date.Format("2025-08-11")
	filename := fmt.Sprintf("%s-%s.md", subject, dateStr)
	filepath := filepath.Join(outputDir, filename)

	// Create markdown content
	var content strings.Builder
	content.WriteString("# " + email.Subject + "\n\n")
	content.WriteString("**From:** " + email.From + "\n")
	content.WriteString("**To:** " + strings.Join(email.To, ", ") + "\n")
	content.WriteString("**Date:** " + email.Date.Format("January 2, 2006 3:04 PM") + "\n")
	
	if len(email.Attachments) > 0 {
		content.WriteString("**Attachments:** " + strings.Join(email.Attachments, ", ") + "\n")
	}
	
	content.WriteString("\n---\n\n")
	content.WriteString(email.Body)

	// Write to file
	return os.WriteFile(filepath, []byte(content.String()), 0644)
}

