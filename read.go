package mailos

import (
	"crypto/tls"
	"fmt"
	"io"
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
}

func Read(opts ReadOptions) ([]*Email, error) {
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