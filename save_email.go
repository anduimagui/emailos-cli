package mailos

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// EmailData represents the common email data structure for saving
type EmailData struct {
	From        string
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	Body        string
	Attachments []string
	Priority    string
	Date        time.Time
	SendAfter   *time.Time
	MessageID   string
	InReplyTo   string
	References  string
}

// SaveEmailToMarkdown saves an email to a markdown file with front matter
// This function is used by both draft and send commands
func SaveEmailToMarkdown(email EmailData, filePath string) error {
	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Write front matter
	if _, err := file.WriteString("---\n"); err != nil {
		return err
	}

	// Write from field if present
	if email.From != "" {
		if _, err := fmt.Fprintf(file, "from: \"%s\"\n", escapeYAMLString(email.From)); err != nil {
			return err
		}
	}

	// Write to field
	if len(email.To) > 0 {
		if _, err := fmt.Fprintf(file, "to: \"%s\"\n", escapeYAMLString(strings.Join(email.To, ", "))); err != nil {
			return err
		}
	}

	// Write cc field if present
	if len(email.CC) > 0 {
		if _, err := fmt.Fprintf(file, "cc: \"%s\"\n", escapeYAMLString(strings.Join(email.CC, ", "))); err != nil {
			return err
		}
	}

	// Write bcc field if present
	if len(email.BCC) > 0 {
		if _, err := fmt.Fprintf(file, "bcc: \"%s\"\n", escapeYAMLString(strings.Join(email.BCC, ", "))); err != nil {
			return err
		}
	}

	// Write subject
	if _, err := fmt.Fprintf(file, "subject: \"%s\"\n", escapeYAMLString(email.Subject)); err != nil {
		return err
	}

	// Write date
	if !email.Date.IsZero() {
		if _, err := fmt.Fprintf(file, "date: \"%s\"\n", email.Date.Format(time.RFC3339)); err != nil {
			return err
		}
	}

	// Write send_after if present
	if email.SendAfter != nil {
		if _, err := fmt.Fprintf(file, "send_after: \"%s\"\n", email.SendAfter.Format(time.RFC3339)); err != nil {
			return err
		}
	}

	// Write priority if not normal
	if email.Priority != "" && email.Priority != "normal" {
		if _, err := fmt.Fprintf(file, "priority: \"%s\"\n", email.Priority); err != nil {
			return err
		}
	}

	// Write message ID if present
	if email.MessageID != "" {
		if _, err := fmt.Fprintf(file, "message_id: \"%s\"\n", escapeYAMLString(email.MessageID)); err != nil {
			return err
		}
	}

	// Write in-reply-to if present
	if email.InReplyTo != "" {
		if _, err := fmt.Fprintf(file, "in_reply_to: \"%s\"\n", escapeYAMLString(email.InReplyTo)); err != nil {
			return err
		}
	}

	// Write references if present
	if email.References != "" {
		if _, err := fmt.Fprintf(file, "references: \"%s\"\n", escapeYAMLString(email.References)); err != nil {
			return err
		}
	}

	// Write attachments if present
	if len(email.Attachments) > 0 {
		if _, err := file.WriteString("attachments:\n"); err != nil {
			return err
		}
		for _, attachment := range email.Attachments {
			if _, err := fmt.Fprintf(file, "  - \"%s\"\n", escapeYAMLString(attachment)); err != nil {
				return err
			}
		}
	}

	// End front matter
	if _, err := file.WriteString("---\n\n"); err != nil {
		return err
	}

	// Write body - directly write the string without any formatting
	// This preserves dollar signs and all other special characters
	if _, err := file.WriteString(email.Body); err != nil {
		return err
	}

	return nil
}

// escapeYAMLString escapes special characters in YAML strings
func escapeYAMLString(s string) string {
	// Escape backslashes first to avoid double escaping
	s = strings.ReplaceAll(s, "\\", "\\\\")
	// Escape double quotes
	s = strings.ReplaceAll(s, "\"", "\\\"")
	// Escape newlines
	s = strings.ReplaceAll(s, "\n", "\\n")
	// Escape carriage returns
	s = strings.ReplaceAll(s, "\r", "\\r")
	// Escape tabs
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// GenerateEmailFilename creates a safe filename for an email
func GenerateEmailFilename(subject string, timestamp time.Time, prefix string) string {
	// Sanitize subject for filename
	safe := strings.Map(func(r rune) rune {
		if r == ' ' {
			return '-'
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return -1
	}, subject)

	if safe == "" {
		safe = "email"
	}

	// Limit length
	if len(safe) > 50 {
		safe = safe[:50]
	}

	// Format: prefix_YYYYMMDD_HHMMSS_subject.md
	return fmt.Sprintf("%s_%s_%s.md", prefix, timestamp.Format("20060102_150405"), safe)
}

// ConvertDraftToEmailData converts a DraftEmail to EmailData
func ConvertDraftToEmailData(draft DraftEmail, from string) EmailData {
	return EmailData{
		From:        from,
		To:          draft.To,
		CC:          draft.CC,
		BCC:         draft.BCC,
		Subject:     draft.Subject,
		Body:        draft.Body,
		Attachments: draft.Attachments,
		Priority:    draft.Priority,
		Date:        time.Now(),
		SendAfter:   draft.SendAfter,
	}
}

// ConvertSavedEmailToEmailData converts a SavedEmail to EmailData
func ConvertSavedEmailToEmailData(saved SavedEmail) EmailData {
	return EmailData{
		From:        saved.From,
		To:          saved.To,
		CC:          saved.CC,
		BCC:         saved.BCC,
		Subject:     saved.Subject,
		Body:        saved.Body,
		Attachments: saved.Attachments,
		Date:        saved.Date,
	}
}