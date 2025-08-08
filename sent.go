package mailos

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

type SentOptions struct {
	Limit       int
	ToAddress   string
	Subject     string
	Since       time.Time
}

func ReadSentEmails(opts SentOptions) ([]*Email, error) {
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

	// Try common sent folder names
	sentFolderNames := []string{"Sent", "Sent Items", "Sent Messages", "[Gmail]/Sent Mail", "INBOX.Sent"}
	var selectedFolder string
	
	// List all available mailboxes to find the sent folder
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()
	
	availableFolders := []string{}
	for mbox := range mailboxes {
		availableFolders = append(availableFolders, mbox.Name)
		// Check if this is a sent folder
		for _, sentName := range sentFolderNames {
			if mbox.Name == sentName {
				selectedFolder = mbox.Name
				break
			}
		}
	}
	
	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to list mailboxes: %v", err)
	}
	
	// If no sent folder found, try to select from common names
	if selectedFolder == "" {
		for _, folderName := range sentFolderNames {
			_, err := c.Select(folderName, false)
			if err == nil {
				selectedFolder = folderName
				break
			}
		}
	}
	
	if selectedFolder == "" {
		return nil, fmt.Errorf("could not find sent folder. Available folders: %v", availableFolders)
	}

	// Select the sent folder
	_, err = c.Select(selectedFolder, false)
	if err != nil {
		return nil, fmt.Errorf("failed to select sent folder '%s': %v", selectedFolder, err)
	}

	// Build search criteria
	criteria := imap.NewSearchCriteria()
	
	// For sent emails, don't filter by FROM as all emails in Sent folder are from us
	// Just use the provided filters
	
	if opts.ToAddress != "" {
		criteria.Header.Add("To", opts.ToAddress)
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
	fetchDone := make(chan error, 1)
	go func() {
		fetchDone <- c.Fetch(seqSet, []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, section.FetchItem()}, messages)
	}()

	emails := make([]*Email, 0, len(ids))
	for msg := range messages {
		email, err := parseMessage(msg, section)
		if err != nil {
			continue
		}
		emails = append(emails, email)
	}

	if err := <-fetchDone; err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %v", err)
	}

	// Reverse to get newest first
	for i, j := 0, len(emails)-1; i < j; i, j = i+1, j-1 {
		emails[i], emails[j] = emails[j], emails[i]
	}

	return emails, nil
}