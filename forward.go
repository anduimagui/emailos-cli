package mailos

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type ForwardOptions struct {
	EmailNumber int      // User-friendly email number from list
	EmailUID    uint32   // IMAP UID if known
	MessageID   string   // Message-ID to forward
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	Body        string
	FileBody    string   // Read body from file
	Interactive bool     // Interactive mode
	Draft       bool     // Save as draft instead of sending
}

func ForwardCommand(opts ForwardOptions) error {
	var originalEmail *Email
	var err error

	// Find the original email to forward
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
		return fmt.Errorf("must specify either email number, UID, or Message-ID to forward")
	}

	if originalEmail == nil {
		return fmt.Errorf("original email not found")
	}

	fmt.Printf("📧 Forwarding: %s\n", originalEmail.Subject)
	fmt.Printf("   From: %s\n", originalEmail.From)
	fmt.Printf("   Date: %s\n", originalEmail.Date.Format("Jan 2, 2006 at 3:04 PM"))
	fmt.Printf("   Message-ID: %s\n", originalEmail.MessageID)

	// Prepare forward
	forward := DraftEmail{}

	// Set recipients
	if len(opts.To) > 0 {
		forward.To = opts.To
	}
	if len(opts.CC) > 0 {
		forward.CC = opts.CC
	}
	if len(opts.BCC) > 0 {
		forward.BCC = opts.BCC
	}

	// Set subject
	subject := originalEmail.Subject
	if !strings.HasPrefix(strings.ToLower(subject), "fwd:") && !strings.HasPrefix(strings.ToLower(subject), "fw:") {
		subject = "Fwd: " + subject
	}
	if opts.Subject != "" {
		subject = opts.Subject
	}
	forward.Subject = subject

	// Set body
	if opts.FileBody != "" {
		fileContent, err := os.ReadFile(opts.FileBody)
		if err != nil {
			return fmt.Errorf("failed to read body from file %s: %v", opts.FileBody, err)
		}
		forward.Body = string(fileContent)
	} else if opts.Body != "" {
		forward.Body = opts.Body
	} else if opts.Interactive {
		// Interactive composition
		body, err := composeForwardInteractively(originalEmail)
		if err != nil {
			return fmt.Errorf("failed to compose forward: %v", err)
		}
		forward.Body = body
	} else {
		// Default: create a basic forward template
		forward.Body = createForwardTemplate(originalEmail)
	}

	// Save or send the forward
	if opts.Draft {
		// Save as draft
		uid, err := saveDraftToIMAP(forward)
		if err != nil {
			return fmt.Errorf("failed to save forward as draft: %v", err)
		}
		fmt.Printf("✓ Forward saved as draft (UID: %d)\n", uid)
		
		// Also save to local drafts
		if err := saveLocalDraft(forward); err != nil {
			fmt.Printf("⚠️  Could not save to local drafts: %v\n", err)
		}
		
		return nil
	} else {
		// Send the forward
		msg := &EmailMessage{
			To:      forward.To,
			CC:      forward.CC,
			BCC:     forward.BCC,
			Subject: forward.Subject,
			Body:    forward.Body,
		}
		
		fmt.Printf("📤 Sending forward...\n")
		if err := Send(msg); err != nil {
			return fmt.Errorf("failed to send forward: %v", err)
		}
		fmt.Printf("✓ Forward sent successfully!\n")
		return nil
	}
}

func composeForwardInteractively(originalEmail *Email) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("\n" + strings.Repeat("─", 60))
	fmt.Println("📝 Compose your forward message:")
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
	
	// Add original message
	forwardContent := createForwardedMessageContent(originalEmail)
	if body != "" {
		body = body + "\n\n" + forwardContent
	} else {
		body = forwardContent
	}
	
	return body, nil
}

func createForwardTemplate(originalEmail *Email) string {
	// Create a basic forward with original message
	template := fmt.Sprintf("\n\n%s", createForwardedMessageContent(originalEmail))
	return template
}

func createForwardedMessageContent(originalEmail *Email) string {
	// Create forwarded message content
	var content strings.Builder
	
	content.WriteString("---------- Forwarded message ----------\n")
	content.WriteString(fmt.Sprintf("From: %s\n", originalEmail.From))
	content.WriteString(fmt.Sprintf("Date: %s\n", originalEmail.Date.Format("Jan 2, 2006 at 3:04 PM")))
	content.WriteString(fmt.Sprintf("Subject: %s\n", originalEmail.Subject))
	if len(originalEmail.To) > 0 {
		content.WriteString(fmt.Sprintf("To: %s\n", strings.Join(originalEmail.To, ", ")))
	}
	content.WriteString("\n")
	content.WriteString(originalEmail.Body)
	
	return content.String()
}

// ForwardEmail is a helper function that can be called with just an email number
func ForwardEmail(emailNumber int, interactive bool) error {
	opts := ForwardOptions{
		EmailNumber: emailNumber,
		Interactive: interactive,
		Draft:       false,
	}
	return ForwardCommand(opts)
}

// DraftForwardEmail creates a forward draft
func DraftForwardEmail(emailNumber int, interactive bool) error {
	opts := ForwardOptions{
		EmailNumber: emailNumber,
		Interactive: interactive,
		Draft:       true,
	}
	return ForwardCommand(opts)
}