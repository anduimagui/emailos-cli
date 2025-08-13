package mailos

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SaveEmailsAsMarkdown saves emails as markdown files
func SaveEmailsAsMarkdown(emails []*Email, outputDir string) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	savedCount := 0
	for _, email := range emails {
		if err := saveEmailAsMarkdown(email, outputDir); err != nil {
			// Log error but continue with other emails
			fmt.Printf("Warning: failed to save email %d: %v\n", email.ID, err)
			continue
		}
		savedCount++
	}

	fmt.Printf("Saved %d emails to %s\n", savedCount, outputDir)
	return nil
}

// saveEmailAsMarkdown saves a single email as a markdown file
func saveEmailAsMarkdown(email *Email, outputDir string) error {
	// Create filename from subject and date
	subjectClean := cleanFilename(email.Subject)
	if subjectClean == "" {
		subjectClean = "no-subject"
	}
	
	// Limit filename length
	if len(subjectClean) > 50 {
		subjectClean = subjectClean[:50]
	}
	
	dateFormatted := email.Date.Format("2006-01-02")
	filename := fmt.Sprintf("%s-%s.md", subjectClean, dateFormatted)
	filepath := filepath.Join(outputDir, filename)
	
	// Create markdown content
	content := formatEmailAsMarkdown(email)
	
	// Write file
	return os.WriteFile(filepath, []byte(content), 0644)
}

// cleanFilename removes characters that are invalid in filenames
func cleanFilename(s string) string {
	// Remove invalid characters
	re := regexp.MustCompile(`[^\w\s-]`)
	s = re.ReplaceAllString(s, "")
	
	// Replace spaces with hyphens
	s = strings.TrimSpace(s)
	re = regexp.MustCompile(`[-\s]+`)
	s = re.ReplaceAllString(s, "-")
	
	return s
}

// formatEmailAsMarkdown formats an email as markdown
func formatEmailAsMarkdown(email *Email) string {
	var content strings.Builder
	
	// Title
	content.WriteString(fmt.Sprintf("# %s\n\n", email.Subject))
	
	// Metadata
	content.WriteString(fmt.Sprintf("**From:** %s  \n", email.From))
	content.WriteString(fmt.Sprintf("**To:** %s  \n", strings.Join(email.To, ", ")))
	content.WriteString(fmt.Sprintf("**Date:** %s  \n", email.Date.Format("January 2, 2006 3:04 PM")))
	content.WriteString(fmt.Sprintf("**ID:** %d  \n", email.ID))
	
	if len(email.Attachments) > 0 {
		content.WriteString(fmt.Sprintf("**Attachments:** %s  \n", strings.Join(email.Attachments, ", ")))
	}
	
	content.WriteString("\n---\n\n")
	
	// Body
	body := email.Body
	if body == "" && email.BodyHTML != "" {
		// If only HTML is available, note it
		content.WriteString("*[HTML email - plain text version not available]*\n\n")
		body = stripHTMLTags(email.BodyHTML)
	}
	
	content.WriteString(body)
	content.WriteString("\n")
	
	// Add unsubscribe links if found
	unsubLinks := extractUnsubscribeLinks(email)
	if len(unsubLinks) > 0 {
		content.WriteString("\n---\n\n")
		content.WriteString("## Unsubscribe Links\n\n")
		for _, link := range unsubLinks {
			content.WriteString(fmt.Sprintf("- <%s>\n", link))
		}
	}
	
	return content.String()
}

// EmailSaveOptions contains options for saving emails
type EmailSaveOptions struct {
	OutputDir      string
	IncludeHTML    bool
	GroupBySender  bool
	DateFormat     string
}

// SaveEmailsWithOptions saves emails with custom options
func SaveEmailsWithOptions(emails []*Email, opts EmailSaveOptions) error {
	if opts.OutputDir == "" {
		// Use .email folder by default
		baseDir, err := GetEmailStorageDir()
		if err != nil {
			opts.OutputDir = ".email"
		} else {
			opts.OutputDir = baseDir
		}
	}
	
	if opts.GroupBySender {
		// Group emails by sender
		grouped := make(map[string][]*Email)
		for _, email := range emails {
			sender := cleanFilename(extractEmailAddress(email.From))
			grouped[sender] = append(grouped[sender], email)
		}
		
		// Save each group in its own directory
		for sender, senderEmails := range grouped {
			senderDir := filepath.Join(opts.OutputDir, sender)
			if err := SaveEmailsAsMarkdown(senderEmails, senderDir); err != nil {
				return err
			}
		}
		return nil
	}
	
	return SaveEmailsAsMarkdown(emails, opts.OutputDir)
}

// extractEmailAddress extracts the email address from a "Name <email>" string
func extractEmailAddress(from string) string {
	// Try to extract email from "Name <email>" format
	if idx := strings.Index(from, "<"); idx >= 0 {
		if endIdx := strings.Index(from[idx:], ">"); endIdx > 0 {
			return from[idx+1 : idx+endIdx]
		}
	}
	// Otherwise return as-is
	return from
}
