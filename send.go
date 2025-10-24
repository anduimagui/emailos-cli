package mailos

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/russross/blackfriday/v2"
)

type EmailMessage struct {
	To              []string
	CC              []string
	BCC             []string
	Subject         string
	Body            string
	BodyHTML        string
	Attachments     []string
	IncludeSignature bool
	SignatureText   string
	UseTemplate     bool     // Whether to apply HTML template
	InReplyTo       string   // Message-ID being replied to
	References      []string // Chain of Message-IDs in conversation
}

// SavedEmail represents an email saved to local storage
type SavedEmail struct {
	ID          string    `json:"id"`
	From        string    `json:"from"`
	To          []string  `json:"to"`
	CC          []string  `json:"cc,omitempty"`
	BCC         []string  `json:"bcc,omitempty"`
	Subject     string    `json:"subject"`
	Body        string    `json:"body"`
	BodyHTML    string    `json:"body_html,omitempty"`
	Date        time.Time `json:"date"`
	RawMessage  string    `json:"raw_message"`
	Attachments []string  `json:"attachments,omitempty"`
}

// SendDraftsOptions contains options for sending draft emails
type SendDraftsOptions struct {
	DraftDir    string
	DryRun      bool
	Filter      string
	Confirm     bool
	DeleteAfter bool
	LogFile     string
}

func Send(msg *EmailMessage) error {
	return SendWithAccount(msg, "")
}

func SendFromMarkdownFile(filePath string, accountEmail string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", filePath, err)
	}
	
	frontmatter, bodyContent, err := ParseFrontmatter(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse frontmatter: %v", err)
	}
	
	var msg *EmailMessage
	if frontmatter != nil {
		msg = frontmatter.ToEmailMessage(bodyContent)
	} else {
		msg = &EmailMessage{
			Body: bodyContent,
		}
		if !strings.Contains(bodyContent, "<") {
			msg.BodyHTML = MarkdownToHTMLContent(bodyContent)
		}
	}
	
	return SendWithAccount(msg, accountEmail)
}

func ProcessEmailWithFrontmatter(msg *EmailMessage) (*EmailMessage, error) {
	if msg.Body == "" {
		return msg, nil
	}
	
	frontmatter, bodyContent, err := ParseFrontmatter(msg.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %v", err)
	}
	
	if frontmatter == nil {
		return msg, nil
	}
	
	processedMsg := frontmatter.ToEmailMessage(bodyContent)
	
	if len(msg.To) > 0 && len(processedMsg.To) == 0 {
		processedMsg.To = msg.To
	}
	if len(msg.CC) > 0 && len(processedMsg.CC) == 0 {
		processedMsg.CC = msg.CC
	}
	if len(msg.BCC) > 0 && len(processedMsg.BCC) == 0 {
		processedMsg.BCC = msg.BCC
	}
	if msg.Subject != "" && processedMsg.Subject == "" {
		processedMsg.Subject = msg.Subject
	}
	if len(msg.Attachments) > 0 && len(processedMsg.Attachments) == 0 {
		processedMsg.Attachments = msg.Attachments
	}
	
	return processedMsg, nil
}

// PreviewEmail displays the complete email content without sending it
func PreviewEmail(msg *EmailMessage, accountEmail string) error {
	setup, err := InitializeMailSetup(accountEmail)
	if err != nil {
		return fmt.Errorf("failed to initialize mail setup: %v", err)
	}
	
	config := setup.Config

	// Prepare from email
	fromEmail := config.Email
	if config.FromEmail != "" {
		fromEmail = config.FromEmail
	}
	
	from := fromEmail
	if config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", config.FromName, fromEmail)
	}

	// Build recipients list
	allRecipients := append([]string{}, msg.To...)
	allRecipients = append(allRecipients, msg.CC...)
	allRecipients = append(allRecipients, msg.BCC...)

	// Add signature if requested
	body := msg.Body
	bodyHTML := msg.BodyHTML
	if msg.IncludeSignature && msg.SignatureText != "" {
		body += msg.SignatureText
		if bodyHTML != "" {
			bodyHTML += strings.ReplaceAll(msg.SignatureText, "\n", "<br>")
		}
	}
	
	// Add EmailOS footer only for non-subscribed users
	if !IsSubscribed() {
		emailOSFooter := "\n\nSent with EmailOS https://email-os.com/"
		emailOSFooterHTML := "<br><br>Sent with <a href=\"https://email-os.com/\">EmailOS</a>"
		
		body += emailOSFooter
		if bodyHTML != "" {
			bodyHTML += emailOSFooterHTML
		} else {
			bodyHTML = strings.ReplaceAll(body, "\n", "<br>")
		}
	} else if bodyHTML == "" {
		bodyHTML = strings.ReplaceAll(body, "\n", "<br>")
	}
	
	// Apply template with profile image if it exists and UseTemplate is true
	if msg.UseTemplate && TemplateExists() {
		if config.ProfileImage != "" {
			bodyHTML = ApplyTemplateWithProfile(body, bodyHTML, config.ProfileImage)
		} else if bodyHTML != "" {
			bodyHTML = ApplyTemplate(body, bodyHTML)
		}
	}

	// Build email message for preview
	var message strings.Builder
	message.WriteString(fmt.Sprintf("From: %s\r\n", from))
	message.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))
	if len(msg.CC) > 0 {
		message.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(msg.CC, ", ")))
	}
	if len(msg.BCC) > 0 {
		message.WriteString(fmt.Sprintf("Bcc: %s\r\n", strings.Join(msg.BCC, ", ")))
	}
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	message.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	
	// Add threading headers if this is a reply
	if msg.InReplyTo != "" {
		inReplyTo := msg.InReplyTo
		if !strings.HasPrefix(inReplyTo, "<") {
			inReplyTo = "<" + inReplyTo + ">"
		}
		message.WriteString(fmt.Sprintf("In-Reply-To: %s\r\n", inReplyTo))
	}
	
	if len(msg.References) > 0 {
		var refs []string
		for _, ref := range msg.References {
			if !strings.HasPrefix(ref, "<") {
				refs = append(refs, "<"+ref+">")
			} else {
				refs = append(refs, ref)
			}
		}
		message.WriteString(fmt.Sprintf("References: %s\r\n", strings.Join(refs, " ")))
	}
	
	message.WriteString("MIME-Version: 1.0\r\n")

	// Add body and show attachment info
	if len(msg.Attachments) > 0 {
		// Show attachments in preview
		message.WriteString("Content-Type: multipart/mixed; boundary=\"preview_boundary\"\r\n")
		message.WriteString("\r\n")
		message.WriteString("--preview_boundary\r\n")
		
		if bodyHTML != "" {
			message.WriteString("Content-Type: multipart/alternative; boundary=\"alt_boundary\"\r\n")
			message.WriteString("\r\n")
			message.WriteString("--alt_boundary\r\n")
			message.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
			message.WriteString("\r\n")
			message.WriteString(body)
			message.WriteString("\r\n")
			message.WriteString("--alt_boundary\r\n")
			message.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
			message.WriteString("\r\n")
			message.WriteString(bodyHTML)
			message.WriteString("\r\n")
			message.WriteString("--alt_boundary--\r\n")
		} else {
			message.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
			message.WriteString("\r\n")
			message.WriteString(body)
			message.WriteString("\r\n")
		}

		// Add attachment info (not actual data in preview)
		for _, attachmentPath := range msg.Attachments {
			filename := filepath.Base(attachmentPath)
			message.WriteString("--preview_boundary\r\n")
			message.WriteString(fmt.Sprintf("Content-Type: application/octet-stream\r\n"))
			message.WriteString("Content-Transfer-Encoding: base64\r\n")
			message.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", filename))
			message.WriteString("\r\n")
			message.WriteString(fmt.Sprintf("[ATTACHMENT: %s - file will be included when sent]\r\n", filename))
		}
		message.WriteString("--preview_boundary--\r\n")
	} else if bodyHTML != "" {
		// No attachments, but has HTML
		boundary := fmt.Sprintf("==boundary_%d==", time.Now().Unix())
		message.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
		message.WriteString("\r\n")

		// Plain text part
		message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		message.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		message.WriteString("\r\n")
		message.WriteString(body)
		message.WriteString("\r\n")

		// HTML part
		message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		message.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
		message.WriteString("\r\n")
		message.WriteString(bodyHTML)
		message.WriteString("\r\n")

		message.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		// Plain text only, no attachments
		message.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		message.WriteString("\r\n")
		message.WriteString(body)
	}

	// Display the preview
	fmt.Println("=== EMAIL PREVIEW ===")
	fmt.Println(message.String())
	fmt.Println("=== END PREVIEW ===")
	
	return nil
}

// SendWithAccountVerbose sends an email using a specific account with optional verbose logging
func SendWithAccountVerbose(msg *EmailMessage, accountEmail string, verbose bool) error {
	return sendWithAccount(msg, accountEmail, verbose)
}

// SendWithAccount sends an email using a specific account
func SendWithAccount(msg *EmailMessage, accountEmail string) error {
	return sendWithAccount(msg, accountEmail, false)
}

// sendWithAccount is the internal implementation
func sendWithAccount(msg *EmailMessage, accountEmail string, verbose bool) error {
	processedMsg, err := ProcessEmailWithFrontmatter(msg)
	if err != nil {
		return fmt.Errorf("failed to process frontmatter: %v", err)
	}
	msg = processedMsg

	// Process attachments if provided
	var attachmentData map[string][]byte
	if len(msg.Attachments) > 0 {
		attachmentData = make(map[string][]byte)
		for _, attachmentPath := range msg.Attachments {
			// Check if file exists
			if _, err := os.Stat(attachmentPath); os.IsNotExist(err) {
				return fmt.Errorf("attachment file not found: %s", attachmentPath)
			}
			
			// Read file data
			data, err := os.ReadFile(attachmentPath)
			if err != nil {
				return fmt.Errorf("failed to read attachment %s: %v", attachmentPath, err)
			}
			
			// Store using filename as key
			filename := filepath.Base(attachmentPath)
			attachmentData[filename] = data
			
			if verbose {
				fmt.Printf("Debug: Added attachment: %s (%d bytes)\n", filename, len(data))
			}
		}
	}

	// Initialize mail setup with optional account
	if verbose {
		fmt.Printf("Debug: Initializing mail setup for account: %s\n", accountEmail)
	}
	setup, err := InitializeMailSetup(accountEmail)
	if err != nil {
		return fmt.Errorf("failed to initialize mail setup: %v", err)
	}
	
	config := setup.Config
	if verbose {
		fmt.Printf("Debug: Using SMTP account: %s\n", config.Email)
		fmt.Printf("Debug: Sending from: %s\n", config.FromEmail)
	}

	// Prepare email headers and body
	// Use FromEmail if specified, otherwise use the account email
	fromEmail := config.Email
	if config.FromEmail != "" {
		fromEmail = config.FromEmail
	}
	
	from := fromEmail
	if config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", config.FromName, fromEmail)
	}

	// Build recipients list
	allRecipients := append([]string{}, msg.To...)
	allRecipients = append(allRecipients, msg.CC...)
	allRecipients = append(allRecipients, msg.BCC...)

	// Add signature if requested
	body := msg.Body
	bodyHTML := msg.BodyHTML
	if msg.IncludeSignature && msg.SignatureText != "" {
		body += msg.SignatureText
		if bodyHTML != "" {
			bodyHTML += strings.ReplaceAll(msg.SignatureText, "\n", "<br>")
		}
	}
	
	// Add EmailOS footer only for non-subscribed users
	if !IsSubscribed() {
		emailOSFooter := "\n\nSent with EmailOS https://email-os.com/"
		emailOSFooterHTML := "<br><br>Sent with <a href=\"https://email-os.com/\">EmailOS</a>"
		
		body += emailOSFooter
		if bodyHTML != "" {
			bodyHTML += emailOSFooterHTML
		} else {
			bodyHTML = strings.ReplaceAll(body, "\n", "<br>")
		}
	} else if bodyHTML == "" {
		bodyHTML = strings.ReplaceAll(body, "\n", "<br>")
	}
	
	// Apply template with profile image if it exists and UseTemplate is true
	if msg.UseTemplate && (config.ProfileImage != "" || TemplateExists()) {
		bodyHTML = ApplyTemplateWithProfile(body, bodyHTML, config.ProfileImage)
	} else if msg.UseTemplate && TemplateExists() && bodyHTML != "" {
		bodyHTML = ApplyTemplate(body, bodyHTML)
	}

	// Build email message
	var message strings.Builder
	message.WriteString(fmt.Sprintf("From: %s\r\n", from))
	message.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))
	if len(msg.CC) > 0 {
		message.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(msg.CC, ", ")))
	}
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	message.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	
	// Add threading headers if this is a reply
	if msg.InReplyTo != "" {
		// Ensure Message-ID is wrapped in angle brackets
		inReplyTo := msg.InReplyTo
		if !strings.HasPrefix(inReplyTo, "<") {
			inReplyTo = "<" + inReplyTo + ">"
		}
		message.WriteString(fmt.Sprintf("In-Reply-To: %s\r\n", inReplyTo))
	}
	
	if len(msg.References) > 0 {
		// Ensure all Message-IDs in References are wrapped in angle brackets
		var refs []string
		for _, ref := range msg.References {
			if !strings.HasPrefix(ref, "<") {
				refs = append(refs, "<"+ref+">")
			} else {
				refs = append(refs, ref)
			}
		}
		message.WriteString(fmt.Sprintf("References: %s\r\n", strings.Join(refs, " ")))
	}
	
	message.WriteString("MIME-Version: 1.0\r\n")

	// Add body and attachments
	if len(attachmentData) > 0 {
		// Mixed multipart for attachments
		mixedBoundary := fmt.Sprintf("==mixed_%d==", time.Now().Unix())
		message.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", mixedBoundary))
		message.WriteString("\r\n")

		// Add the email body part
		message.WriteString(fmt.Sprintf("--%s\r\n", mixedBoundary))
		
		if bodyHTML != "" {
			// Alternative part for text/html
			altBoundary := fmt.Sprintf("==alt_%d==", time.Now().Unix())
			message.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", altBoundary))
			message.WriteString("\r\n")

			// Plain text part
			message.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
			message.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
			message.WriteString("\r\n")
			message.WriteString(body)
			message.WriteString("\r\n")

			// HTML part
			message.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
			message.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
			message.WriteString("\r\n")
			message.WriteString(bodyHTML)
			message.WriteString("\r\n")

			message.WriteString(fmt.Sprintf("--%s--\r\n", altBoundary))
		} else {
			// Plain text only
			message.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
			message.WriteString("\r\n")
			message.WriteString(body)
			message.WriteString("\r\n")
		}

		// Add attachments
		for filename, data := range attachmentData {
			message.WriteString(fmt.Sprintf("--%s\r\n", mixedBoundary))
			
			// Detect MIME type
			mimeType := mime.TypeByExtension(filepath.Ext(filename))
			if mimeType == "" {
				mimeType = "application/octet-stream"
			}
			
			message.WriteString(fmt.Sprintf("Content-Type: %s\r\n", mimeType))
			message.WriteString("Content-Transfer-Encoding: base64\r\n")
			message.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", filename))
			message.WriteString("\r\n")
			
			// Encode in base64
			encoded := base64.StdEncoding.EncodeToString(data)
			// Split into 76-character lines as per RFC 2045
			for i := 0; i < len(encoded); i += 76 {
				end := i + 76
				if end > len(encoded) {
					end = len(encoded)
				}
				message.WriteString(encoded[i:end] + "\r\n")
			}
		}

		message.WriteString(fmt.Sprintf("--%s--\r\n", mixedBoundary))
	} else if bodyHTML != "" {
		// No attachments, but has HTML - multipart/alternative
		boundary := fmt.Sprintf("==boundary_%d==", time.Now().Unix())
		message.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
		message.WriteString("\r\n")

		// Plain text part
		message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		message.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		message.WriteString("\r\n")
		message.WriteString(body)
		message.WriteString("\r\n")

		// HTML part
		message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		message.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
		message.WriteString("\r\n")
		message.WriteString(bodyHTML)
		message.WriteString("\r\n")

		message.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		// Plain text only, no attachments
		message.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		message.WriteString("\r\n")
		message.WriteString(body)
	}

	// Get SMTP settings from provider
	smtpHost, smtpPort, useTLS, useSSL, err := config.GetSMTPSettings()
	if err != nil {
		return fmt.Errorf("failed to get SMTP settings: %v", err)
	}

	if verbose {
		fmt.Printf("Debug: SMTP Host: %s:%d\n", smtpHost, smtpPort)
		fmt.Printf("Debug: TLS: %v, SSL: %v\n", useTLS, useSSL)
		fmt.Printf("Debug: SMTP Auth User: %s\n", config.Email)
		fmt.Printf("Debug: From Email in message: %s\n", fromEmail)
		fmt.Printf("Debug: Recipients: %v\n", allRecipients)
	}

	// Send email
	auth := smtp.PlainAuth("", config.Email, config.Password, smtpHost)

	if useTLS {
		// Use STARTTLS
		err = sendWithSTARTTLS(
			smtpHost,
			smtpPort,
			auth,
			fromEmail,
			allRecipients,
			message.String(),
		)
		if err != nil {
			return handleSendError(err, fromEmail, config.Email)
		}
		// After successfully sending, save to Sent folder
		return saveToSentFolder(message.String(), config, msg, from)
	} else if useSSL {
		// Use SMTPS (SMTP over SSL)
		err = sendWithSMTPS(
			smtpHost,
			smtpPort,
			auth,
			fromEmail,
			allRecipients,
			message.String(),
		)
		if err != nil {
			return handleSendError(err, fromEmail, config.Email)
		}
		// After successfully sending, save to Sent folder
		return saveToSentFolder(message.String(), config, msg, from)
	}

	// Plain SMTP (not recommended)
	addr := fmt.Sprintf("%s:%d", smtpHost, smtpPort)
	err = smtp.SendMail(addr, auth, fromEmail, allRecipients, []byte(message.String()))
	if err != nil {
		return handleSendError(err, fromEmail, config.Email)
	}
	
	// After successfully sending, save to Sent folder
	return saveToSentFolder(message.String(), config, msg, from)
}

// saveToSentFolder saves the sent email to both local storage and IMAP Sent folder
func saveToSentFolder(messageContent string, config *Config, msg *EmailMessage, from string) error {
	// First, save to local .email/sent folder
	if err := saveToLocalSentFolder(messageContent, config, msg, from); err != nil {
		// Log error but don't fail the send
		fmt.Printf("Note: Could not save to local sent folder: %v\n", err)
	}
	
	// Get IMAP settings from provider
	imapHost, imapPort, err := config.GetIMAPSettings()
	if err != nil {
		// If we can't get IMAP settings, just log and continue (email was sent)
		fmt.Println("Note: Could not save to IMAP Sent folder (IMAP not configured)")
		return nil
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
		fmt.Printf("Note: Could not save to IMAP Sent folder (connection failed: %v)\n", err)
		return nil
	}
	defer c.Logout()

	// Login
	if err := c.Login(config.Email, config.Password); err != nil {
		fmt.Printf("Note: Could not save to IMAP Sent folder (login failed: %v)\n", err)
		return nil
	}

	// Try common sent folder names
	sentFolderNames := []string{"Sent", "Sent Items", "Sent Messages", "[Gmail]/Sent Mail", "INBOX.Sent"}
	var selectedFolder string
	
	for _, folderName := range sentFolderNames {
		// Try to select the folder
		_, err := c.Select(folderName, false)
		if err == nil {
			selectedFolder = folderName
			c.Close() // Close the selected folder
			break
		}
	}
	
	if selectedFolder == "" {
		fmt.Println("Note: Could not find IMAP Sent folder to save message")
		return nil
	}

	// Append the message to the Sent folder
	// IMAP requires CRLF line endings
	messageWithCRLF := strings.ReplaceAll(messageContent, "\r\n", "\n")
	messageWithCRLF = strings.ReplaceAll(messageWithCRLF, "\n", "\r\n")
	
	flags := []string{imap.SeenFlag}
	date := time.Now()
	err = c.Append(selectedFolder, flags, date, strings.NewReader(messageWithCRLF))
	if err != nil {
		fmt.Printf("Note: Could not save to IMAP Sent folder (append failed: %v)\n", err)
		return nil
	}

	// Verify the email was saved to sent folder
	if err := verifySentEmail(msg, config); err != nil {
		fmt.Printf("Warning: Could not verify email was saved to sent folder: %v\n", err)
	} else {
		fmt.Println("✓ Email confirmed in sent folder")
	}

	return nil
}

// verifySentEmail checks that the sent email appears in the sent folder and validates delivery
func verifySentEmail(msg *EmailMessage, config *Config) error {
	// Wait a moment for the email to appear in the sent folder
	time.Sleep(2 * time.Second)
	
	// Search for the email in sent folder using subject and recent timestamp
	opts := SentOptions{
		Limit:   5,
		Subject: msg.Subject,
		Since:   time.Now().Add(-5 * time.Minute), // Look for emails sent in last 5 minutes
	}
	
	sentEmails, err := ReadSentEmails(opts)
	if err != nil {
		return fmt.Errorf("failed to read sent emails for verification: %v", err)
	}
	
	// Check if any of the found emails match our sent email
	emailFound := false
	for _, email := range sentEmails {
		if email.Subject == msg.Subject {
			// Additional checks to ensure it's the right email
			if len(msg.To) > 0 && len(email.To) > 0 {
				// Check if at least one recipient matches
				for _, sentTo := range msg.To {
					for _, emailTo := range email.To {
						if strings.Contains(emailTo, sentTo) || strings.Contains(sentTo, emailTo) {
							emailFound = true
							break
						}
					}
					if emailFound {
						break
					}
				}
			} else {
				// If no To addresses to compare, just match on subject and recent time
				emailFound = true
			}
		}
		if emailFound {
			break
		}
	}
	
	if !emailFound {
		return fmt.Errorf("sent email not found in sent folder")
	}
	
	// Check for bounce notifications
	if err := checkForBounces(msg, config); err != nil {
		return fmt.Errorf("delivery verification failed: %v", err)
	}
	
	return nil
}

// checkForBounces checks the inbox for bounce notifications related to the sent email
func checkForBounces(msg *EmailMessage, config *Config) error {
	// Wait additional time for potential bounces to arrive
	time.Sleep(3 * time.Second)
	
	// Read recent emails from inbox to check for bounces
	readOpts := ReadOptions{
		Limit:       10,
		FromAddress: "Mail Delivery System", // Common bounce sender
		Since:       time.Now().Add(-10 * time.Minute),
	}
	
	inboxEmails, err := Read(readOpts)
	if err != nil {
		// If we can't read inbox, don't fail verification - just warn
		fmt.Printf("Note: Could not check for bounces: %v\n", err)
		return nil
	}
	
	// Check for bounce notifications that match our recipients
	for _, email := range inboxEmails {
		if isBounceNotification(email, msg) {
			// Extract failed recipients from bounce message
			failedRecipients := extractFailedRecipients(email, msg.To)
			if len(failedRecipients) > 0 {
				return fmt.Errorf("email delivery failed for: %s", strings.Join(failedRecipients, ", "))
			}
		}
	}
	
	// Also check for bounces from common bounce senders
	bounceFromAddresses := []string{
		"mailer-daemon",
		"postmaster",
		"noreply",
		"Mail Delivery System",
		"Mail Delivery Subsystem",
	}
	
	for _, bounceFrom := range bounceFromAddresses {
		readOpts.FromAddress = bounceFrom
		bounceEmails, err := Read(readOpts)
		if err != nil {
			continue // Skip if can't read with this sender
		}
		
		for _, email := range bounceEmails {
			if isBounceNotification(email, msg) {
				failedRecipients := extractFailedRecipients(email, msg.To)
				if len(failedRecipients) > 0 {
					return fmt.Errorf("email delivery failed for: %s", strings.Join(failedRecipients, ", "))
				}
			}
		}
	}
	
	return nil
}

// isBounceNotification checks if an email is a bounce notification for our sent message
func isBounceNotification(email *Email, sentMsg *EmailMessage) bool {
	// Check common bounce indicators in subject
	bounceIndicators := []string{
		"Undelivered Mail Returned to Sender",
		"Delivery Status Notification",
		"Mail delivery failed",
		"Returned mail",
		"Message not delivered",
		"Delivery failure",
		"Undeliverable:",
		"Failed:",
	}
	
	subjectLower := strings.ToLower(email.Subject)
	for _, indicator := range bounceIndicators {
		if strings.Contains(subjectLower, strings.ToLower(indicator)) {
			// Check if bounce refers to our original message
			bodyLower := strings.ToLower(email.Body)
			
			// Look for our original subject in the bounce message
			if strings.Contains(bodyLower, strings.ToLower(sentMsg.Subject)) {
				return true
			}
			
			// Look for our recipients in the bounce message
			for _, recipient := range sentMsg.To {
				if strings.Contains(bodyLower, strings.ToLower(recipient)) {
					return true
				}
			}
		}
	}
	
	return false
}

// extractFailedRecipients extracts the failed recipient addresses from a bounce message
func extractFailedRecipients(bounceEmail *Email, originalRecipients []string) []string {
	var failedRecipients []string
	bodyLower := strings.ToLower(bounceEmail.Body)
	
	// Check each original recipient to see if it's mentioned in the bounce
	for _, recipient := range originalRecipients {
		if strings.Contains(bodyLower, strings.ToLower(recipient)) {
			failedRecipients = append(failedRecipients, recipient)
		}
	}
	
	return failedRecipients
}

// saveToLocalSentFolder saves the email to the local .email/sent directory
func saveToLocalSentFolder(messageContent string, config *Config, msg *EmailMessage, from string) error {
	// Ensure directories exist
	if err := EnsureEmailDirectories(); err != nil {
		return fmt.Errorf("failed to create email directories: %v", err)
	}
	
	// Get sent directory
	sentDir, err := GetSentDir()
	if err != nil {
		return fmt.Errorf("failed to get sent directory: %v", err)
	}
	
	// Convert to EmailData for saving
	emailData := EmailData{
		From:        from,
		To:          msg.To,
		CC:          msg.CC,
		BCC:         msg.BCC,
		Subject:     msg.Subject,
		Body:        msg.Body,
		Attachments: msg.Attachments,
		Date:        time.Now(),
	}
	
	// Generate filename
	filename := GenerateEmailFilename(msg.Subject, time.Now(), "sent")
	filepath := filepath.Join(sentDir, filename)
	
	// Save using the common function
	if err := SaveEmailToMarkdown(emailData, filepath); err != nil {
		return fmt.Errorf("failed to write email file: %v", err)
	}
	
	return nil
}

func sendWithSTARTTLS(host string, port int, auth smtp.Auth, from string, to []string, msg string) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()

	// Start TLS
	tlsConfig := &tls.Config{ServerName: host}
	if err = c.StartTLS(tlsConfig); err != nil {
		return err
	}

	// Authenticate
	if err = c.Auth(auth); err != nil {
		return err
	}

	// Set sender and recipients
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}

	// Send the email body
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}

func sendWithSMTPS(host string, port int, auth smtp.Auth, from string, to []string, msg string) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	
	// Connect with TLS
	tlsConfig := &tls.Config{ServerName: host}
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return err
	}
	defer c.Close()

	// Authenticate
	if err = c.Auth(auth); err != nil {
		return err
	}

	// Set sender and recipients
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}

	// Send the email body
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}

// handleSendError provides specific error handling for email sending failures
func handleSendError(err error, fromEmail, configEmail string) error {
	errStr := err.Error()
	
	// Check for common alias/authentication errors
	if strings.Contains(errStr, "authentication failed") || 
	   strings.Contains(errStr, "invalid credentials") ||
	   strings.Contains(errStr, "535") {
		
		if fromEmail != configEmail {
			return fmt.Errorf("authentication failed when sending from %s. This may be because:\n"+
				"1. The alias '%s' is not configured in your email provider\n"+
				"2. Your email provider doesn't allow sending from this alias\n"+
				"3. You need to add '%s' as an authorized sending address\n\n"+
				"For Fastmail: Go to Settings > Identities and add '%s' as a new identity\n"+
				"Original error: %v", fromEmail, fromEmail, fromEmail, fromEmail, err)
		}
	}
	
	// Check for "from address not allowed" type errors
	if strings.Contains(errStr, "from address") || 
	   strings.Contains(errStr, "sender") ||
	   strings.Contains(errStr, "not allowed") {
		return fmt.Errorf("sending from '%s' is not allowed. Please configure this email as an alias in your email provider settings. Original error: %v", fromEmail, err)
	}
	
	// Return original error if no specific handling applies
	return err
}

// MarkdownToHTMLContent converts markdown text to HTML content (without full document wrapper)
func MarkdownToHTMLContent(markdown string) string {
	html := blackfriday.Run([]byte(markdown))
	return string(html)
}

// SendDrafts processes and sends all draft emails from the drafts folder
func SendDrafts(opts SendDraftsOptions) error {
	// Set default draft directory
	if opts.DraftDir == "" {
		draftsDir, err := GetDraftsDir()
		if err != nil {
			return fmt.Errorf("failed to get drafts directory: %v", err)
		}
		opts.DraftDir = draftsDir
	}

	// Check if draft directory exists
	if _, err := os.Stat(opts.DraftDir); os.IsNotExist(err) {
		return fmt.Errorf("draft directory does not exist: %s", opts.DraftDir)
	}

	// Get all markdown files in the draft directory
	pattern := filepath.Join(opts.DraftDir, "*.md")
	draftFiles, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to list draft files: %v", err)
	}

	if len(draftFiles) == 0 {
		fmt.Printf("No draft files found in %s/\n", opts.DraftDir)
		return nil
	}

	fmt.Printf("Found %d draft(s) to process\n", len(draftFiles))

	// Confirm before sending if requested
	if opts.Confirm && !opts.DryRun {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Send all drafts? (y/n): ")
		response, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Create failed directory if needed
	failedDir := filepath.Join(opts.DraftDir, "failed")
	if !opts.DryRun {
		os.MkdirAll(failedDir, 0755)
	}

	// Process each draft
	successCount := 0
	failCount := 0
	
	for i, draftFile := range draftFiles {
		fmt.Printf("\n[%d/%d] Processing: %s\n", i+1, len(draftFiles), filepath.Base(draftFile))
		
		// Parse the draft file using our enhanced frontmatter parser
		draft, err := parseDraftFileWithFrontmatter(draftFile)
		if err != nil {
			fmt.Printf("  ❌ Failed to parse: %v\n", err)
			failCount++
			continue
		}

		// Apply filter if specified
		if opts.Filter != "" && !matchesFilter(draft, opts.Filter) {
			fmt.Println("  ⏭️  Skipped (doesn't match filter)")
			continue
		}

		// Check if scheduled for later
		if draft.SendAfter != nil && draft.SendAfter.After(time.Now()) {
			fmt.Printf("  ⏰ Scheduled for later: %s\n", draft.SendAfter.Format("Jan 2, 3:04 PM"))
			continue
		}

		// Dry run mode - just show what would be sent
		if opts.DryRun {
			fmt.Printf("  📧 Would send to: %s\n", strings.Join(draft.To, ", "))
			fmt.Printf("     Subject: %s\n", draft.Subject)
			if len(draft.CC) > 0 {
				fmt.Printf("     CC: %s\n", strings.Join(draft.CC, ", "))
			}
			successCount++
			continue
		}

		// Load config to get signature settings
		config, err := LoadConfig()
		if err != nil {
			fmt.Printf("  ⚠️  Warning: Could not load config for signature: %v\n", err)
		}

		// Create email message
		msg := &EmailMessage{
			To:          draft.To,
			CC:          draft.CC,
			BCC:         draft.BCC,
			Subject:     draft.Subject,
			Body:        draft.Body,
			Attachments: draft.Attachments,
			InReplyTo:   draft.InReplyTo,
			References:  draft.References,
		}

		// Add signature if config loaded successfully
		if config != nil {
			var sig string
			// Check for signature override first
			if config.SignatureOverride != "" {
				sig = config.SignatureOverride
			} else {
				// Use FromEmail if specified, otherwise use Email
				emailToShow := config.Email
				if config.FromEmail != "" {
					emailToShow = config.FromEmail
				}
				name := config.FromName
				if name == "" {
					name = strings.Split(emailToShow, "@")[0]
				}
				sig = fmt.Sprintf("\n--\n%s\n%s", name, emailToShow)
			}
			msg.IncludeSignature = true
			msg.SignatureText = sig
		}

		// Convert markdown to HTML
		if !strings.Contains(draft.Body, "<html>") {
			msg.BodyHTML = MarkdownToHTMLContent(draft.Body)
		}

		// Send the email
		err = Send(msg)
		if err != nil {
			fmt.Printf("  ❌ Failed to send: %v\n", err)
			// Move to failed directory
			failedPath := filepath.Join(failedDir, filepath.Base(draftFile))
			os.Rename(draftFile, failedPath)
			failCount++
			continue
		}

		fmt.Printf("  ✅ Sent successfully!\n")
		successCount++

		// Log the sent email
		if opts.LogFile != "" {
			logSentEmail(opts.LogFile, draft, draftFile)
		}

		// Delete the draft file if requested (default behavior)
		if opts.DeleteAfter {
			os.Remove(draftFile)
		} else {
			// Move to sent directory
			sentDir := filepath.Join(opts.DraftDir, "sent")
			os.MkdirAll(sentDir, 0755)
			sentPath := filepath.Join(sentDir, filepath.Base(draftFile))
			os.Rename(draftFile, sentPath)
		}
	}

	// Summary
	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("  ✅ Sent: %d\n", successCount)
	if failCount > 0 {
		fmt.Printf("  ❌ Failed: %d (moved to %s/)\n", failCount, failedDir)
	}
	
	return nil
}

// parseDraftFileWithFrontmatter reads and parses a markdown draft file using enhanced frontmatter
func parseDraftFileWithFrontmatter(filePath string) (*DraftEmail, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Use our enhanced frontmatter parser
	frontmatter, bodyContent, err := ParseFrontmatter(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %v", err)
	}

	var draft *DraftEmail
	if frontmatter != nil {
		// Convert frontmatter to draft
		draftFromFM := frontmatter.ToDraftEmail(bodyContent)
		draft = &draftFromFM
	} else {
		// No frontmatter, treat entire content as body
		draft = &DraftEmail{
			Body: bodyContent,
		}
	}
	
	// Validate required fields
	if len(draft.To) == 0 {
		return nil, fmt.Errorf("draft missing 'to' field")
	}
	if draft.Subject == "" {
		return nil, fmt.Errorf("draft missing 'subject' field")
	}
	
	return draft, nil
}

// matchesFilter checks if a draft matches the filter criteria
func matchesFilter(draft *DraftEmail, filter string) bool {
	// Parse filter (e.g., "priority:high" or "to:*@example.com")
	parts := strings.SplitN(filter, ":", 2)
	if len(parts) != 2 {
		return true
	}
	
	key := strings.ToLower(parts[0])
	value := strings.ToLower(parts[1])
	
	switch key {
	case "priority":
		return strings.ToLower(draft.Priority) == value
	case "to":
		for _, to := range draft.To {
			if strings.Contains(strings.ToLower(to), value) {
				return true
			}
		}
		return false
	case "subject":
		return strings.Contains(strings.ToLower(draft.Subject), value)
	default:
		return true
	}
}

// logSentEmail logs a sent email to a file
func logSentEmail(logFile string, draft *DraftEmail, draftFile string) error {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] Sent: %s | To: %s | From draft: %s\n",
		timestamp,
		draft.Subject,
		strings.Join(draft.To, ", "),
		filepath.Base(draftFile),
	)
	
	_, err = file.WriteString(logEntry)
	return err
}
