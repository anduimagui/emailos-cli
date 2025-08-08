package mailos

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"
	
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
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
}

func Send(msg *EmailMessage) error {
	// For now, we'll skip attachment support in the simple implementation
	if len(msg.Attachments) > 0 {
		return fmt.Errorf("attachment support not yet implemented")
	}

	// Otherwise use the simple implementation
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
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
	
	// Apply template with profile image if it exists
	if config.ProfileImage != "" || TemplateExists() {
		bodyHTML = ApplyTemplateWithProfile(body, bodyHTML, config.ProfileImage)
	} else if TemplateExists() && bodyHTML != "" {
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
	message.WriteString("MIME-Version: 1.0\r\n")

	// Add body
	if bodyHTML != "" {
		// Multipart message with HTML and plain text
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
		// Plain text only
		message.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		message.WriteString("\r\n")
		message.WriteString(body)
	}

	// Get SMTP settings from provider
	smtpHost, smtpPort, useTLS, useSSL, err := config.GetSMTPSettings()
	if err != nil {
		return fmt.Errorf("failed to get SMTP settings: %v", err)
	}

	// Send email
	auth := smtp.PlainAuth("", config.Email, config.Password, smtpHost)

	if useTLS {
		// Use STARTTLS
		err = sendWithSTARTTLS(
			smtpHost,
			smtpPort,
			auth,
			config.Email,
			allRecipients,
			message.String(),
		)
		if err != nil {
			return err
		}
		// After successfully sending, save to Sent folder
		return saveToSentFolder(message.String(), config)
	} else if useSSL {
		// Use SMTPS (SMTP over SSL)
		err = sendWithSMTPS(
			smtpHost,
			smtpPort,
			auth,
			config.Email,
			allRecipients,
			message.String(),
		)
		if err != nil {
			return err
		}
		// After successfully sending, save to Sent folder
		return saveToSentFolder(message.String(), config)
	}

	// Plain SMTP (not recommended)
	addr := fmt.Sprintf("%s:%d", smtpHost, smtpPort)
	err = smtp.SendMail(addr, auth, config.Email, allRecipients, []byte(message.String()))
	if err != nil {
		return err
	}
	
	// After successfully sending, save to Sent folder
	return saveToSentFolder(message.String(), config)
}

// saveToSentFolder saves the sent email to the IMAP Sent folder
func saveToSentFolder(messageContent string, config *Config) error {
	// Get IMAP settings from provider
	imapHost, imapPort, err := config.GetIMAPSettings()
	if err != nil {
		// If we can't get IMAP settings, just log and continue (email was sent)
		fmt.Println("Note: Could not save to Sent folder (IMAP not configured)")
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
		fmt.Printf("Note: Could not save to Sent folder (connection failed: %v)\n", err)
		return nil
	}
	defer c.Logout()

	// Login
	if err := c.Login(config.Email, config.Password); err != nil {
		fmt.Printf("Note: Could not save to Sent folder (login failed: %v)\n", err)
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
		fmt.Println("Note: Could not find Sent folder to save message")
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
		fmt.Printf("Note: Could not save to Sent folder (append failed: %v)\n", err)
		return nil
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