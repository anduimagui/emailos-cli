package mailos

import (
	"fmt"
	"strings"

	"github.com/russross/blackfriday/v2"
)

// Client is the main email client interface
type Client struct {
	config *Config
}

// NewClient creates a new email client
// Note: Assumes EnsureInitialized() has been called by middleware
func NewClient() (*Client, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	return &Client{config: config}, nil
}

// SendEmail sends an email with the given parameters
func (c *Client) SendEmail(to []string, subject, body string, cc []string, bcc []string) error {
	msg := &EmailMessage{
		To:      to,
		CC:      cc,
		BCC:     bcc,
		Subject: subject,
		Body:    body,
	}

	// Convert markdown to HTML
	html := MarkdownToHTML(body)
	if html != body {
		msg.BodyHTML = html
	}

	return Send(msg)
}

// ReadEmails reads emails with the given options
func (c *Client) ReadEmails(opts ReadOptions) ([]*Email, error) {
	// Try to read from global inbox first if LocalOnly is set
	if opts.LocalOnly {
		if c.config.Email != "" {
			emails, err := GetEmailsFromInbox(c.config.Email, opts)
			if err == nil && len(emails) > 0 {
				return emails, nil
			}
		}
		// Fallback to old local storage method
		return readFromLocalStorage(opts)
	}
	return Read(opts)
}

// SyncEmails fetches emails incrementally and stores in global inbox
func (c *Client) SyncEmails(limit int) error {
	return FetchEmailsIncremental(c.config, limit)
}

// ReadEmailsFromInbox reads emails from the global inbox
func (c *Client) ReadEmailsFromInbox(opts ReadOptions) ([]*Email, error) {
	if c.config.Email == "" {
		return nil, fmt.Errorf("no email account configured")
	}
	return GetEmailsFromInbox(c.config.Email, opts)
}

// MarkEmailsAsRead marks the given email IDs as read
func (c *Client) MarkEmailsAsRead(ids []uint32) error {
	return MarkAsRead(ids)
}

// DeleteEmails deletes the given email IDs
func (c *Client) DeleteEmails(ids []uint32) error {
	return DeleteEmails(ids)
}

// DeleteDrafts deletes the given draft IDs from the Drafts folder
func (c *Client) DeleteDrafts(ids []uint32) error {
	return DeleteDrafts(ids)
}

// ReadDrafts reads drafts from the IMAP Drafts folder
func (c *Client) ReadDrafts(opts ReadOptions) ([]*Email, error) {
	return ReadFromFolder(opts, "Drafts")
}

// FindUnsubscribeLinks finds unsubscribe links in emails
func (c *Client) FindUnsubscribeLinks(opts ReadOptions) ([]UnsubscribeLinks, error) {
	emails, err := Read(opts)
	if err != nil {
		return nil, err
	}
	return FindUnsubscribeLinks(emails), nil
}

// SaveEmailsAsMarkdown saves emails as markdown files
func (c *Client) SaveEmailsAsMarkdown(emails []*Email, outputDir string) error {
	return SaveEmailsAsMarkdown(emails, outputDir)
}

// SendEmailWithAttachments sends an email with attachments
func (c *Client) SendEmailWithAttachments(to []string, subject, body string, cc []string, bcc []string, attachments []string) error {
	msg := &EmailMessage{
		To:          to,
		CC:          cc,
		BCC:         bcc,
		Subject:     subject,
		Body:        body,
		Attachments: attachments,
	}

	// Convert markdown to HTML
	html := MarkdownToHTML(body)
	if html != body {
		msg.BodyHTML = html
	}

	return Send(msg)
}

// SendEmailWithSignature sends an email with a signature
func (c *Client) SendEmailWithSignature(to []string, subject, body string, signature string) error {
	msg := &EmailMessage{
		To:               to,
		Subject:          subject,
		Body:             body,
		IncludeSignature: true,
		SignatureText:    signature,
	}

	// Convert markdown to HTML
	html := MarkdownToHTML(body)
	if html != body {
		msg.BodyHTML = html
	}

	return Send(msg)
}

// GetConfig returns the current configuration
func (c *Client) GetConfig() *Config {
	return c.config
}

// MarkdownToHTML converts markdown text to HTML
func MarkdownToHTML(markdown string) string {
	// Use blackfriday for markdown parsing
	html := blackfriday.Run([]byte(markdown))

	// Wrap in basic HTML template
	template := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        h1, h2, h3 { color: #2c3e50; margin-top: 20px; margin-bottom: 10px; }
        p { margin: 10px 0; }
        a { color: #3498db; text-decoration: none; }
        a:hover { text-decoration: underline; }
        pre { background: #f4f4f4; padding: 10px; border-radius: 4px; overflow-x: auto; }
        code { background: #f4f4f4; padding: 2px 4px; border-radius: 2px; }
        blockquote { border-left: 4px solid #ddd; margin: 0; padding-left: 16px; color: #666; }
    </style>
</head>
<body>
%s
</body>
</html>`

	return fmt.Sprintf(template, string(html))
}

// GetProviderInfo returns information about the configured email provider
func (c *Client) GetProviderInfo() string {
	provider, ok := Providers[c.config.Provider]
	if !ok {
		return "Unknown provider"
	}
	return provider.Name
}

// FormatEmailList formats a list of emails for display
func FormatEmailList(emails []*Email) string {
	return FormatEmailListWithDrafts(emails, len(emails))
}

// FormatEmailListWithDrafts formats a list of emails and drafts for display
func FormatEmailListWithDrafts(emails []*Email, emailCount int) string {
	if len(emails) == 0 {
		return "No emails or drafts found."
	}

	var result strings.Builder
	for i, email := range emails {
		isDraft := i >= emailCount
		draftIndicator := ""
		if isDraft {
			draftIndicator = " [DRAFT]"
		}

		result.WriteString(fmt.Sprintf("\n%d. From: %s\n", i+1, email.From))
		result.WriteString(fmt.Sprintf("   Subject: %s\n", email.Subject))
		result.WriteString(fmt.Sprintf("   Date: %s%s\n", email.Date.Format("Jan 2, 2006 3:04 PM"), draftIndicator))

		// Show preview of body
		preview := email.Body
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		preview = strings.ReplaceAll(preview, "\n", " ")
		result.WriteString(fmt.Sprintf("   Preview: %s\n", preview))

		if len(email.Attachments) > 0 {
			result.WriteString(fmt.Sprintf("   Attachments: %s\n", strings.Join(email.Attachments, ", ")))
		}
	}

	return result.String()
}
