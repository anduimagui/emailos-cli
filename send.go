package mailos

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"
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
	
	// Apply template if it exists
	if TemplateExists() && bodyHTML != "" {
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
		return sendWithSTARTTLS(
			smtpHost,
			smtpPort,
			auth,
			config.Email,
			allRecipients,
			message.String(),
		)
	} else if useSSL {
		// Use SMTPS (SMTP over SSL)
		return sendWithSMTPS(
			smtpHost,
			smtpPort,
			auth,
			config.Email,
			allRecipients,
			message.String(),
		)
	}

	// Plain SMTP (not recommended)
	addr := fmt.Sprintf("%s:%d", smtpHost, smtpPort)
	return smtp.SendMail(addr, auth, config.Email, allRecipients, []byte(message.String()))
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