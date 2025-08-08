package mailos

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SendDraftsOptions contains options for sending draft emails
type SendDraftsOptions struct {
	DraftDir    string
	DryRun      bool
	Filter      string
	Confirm     bool
	DeleteAfter bool
	LogFile     string
}

// SendDrafts processes and sends all draft emails from the drafts folder
func SendDrafts(opts SendDraftsOptions) error {
	// Set default draft directory
	if opts.DraftDir == "" {
		opts.DraftDir = "draft-emails"
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
		
		// Parse the draft file
		draft, err := parseDraftFile(draftFile)
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

		// Create email message
		msg := &EmailMessage{
			To:          draft.To,
			CC:          draft.CC,
			BCC:         draft.BCC,
			Subject:     draft.Subject,
			Body:        draft.Body,
			Attachments: draft.Attachments,
		}

		// Convert markdown to HTML
		if !strings.Contains(draft.Body, "<html>") {
			msg.BodyHTML = MarkdownToHTML(draft.Body)
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

// parseDraftFile reads and parses a markdown draft file
func parseDraftFile(filepath string) (*DraftEmail, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	draft := &DraftEmail{}
	
	// Parse frontmatter
	if len(lines) > 0 && lines[0] == "---" {
		inFrontmatter := true
		bodyStart := 0
		
		for i := 1; i < len(lines); i++ {
			line := lines[i]
			
			if line == "---" {
				inFrontmatter = false
				bodyStart = i + 1
				break
			}
			
			if inFrontmatter {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					
					switch key {
					case "to":
						draft.To = parseEmailList(value)
					case "cc":
						draft.CC = parseEmailList(value)
					case "bcc":
						draft.BCC = parseEmailList(value)
					case "subject":
						draft.Subject = value
					case "priority":
						draft.Priority = value
					case "send_after":
						if t, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
							draft.SendAfter = &t
						}
					case "attachments":
						// Handle multi-line attachments
						if value == "" && i+1 < len(lines) {
							for j := i + 1; j < len(lines); j++ {
								if strings.HasPrefix(lines[j], "  - ") {
									attachment := strings.TrimPrefix(lines[j], "  - ")
									draft.Attachments = append(draft.Attachments, strings.TrimSpace(attachment))
									i = j
								} else if !strings.HasPrefix(lines[j], "  ") {
									break
								}
							}
						}
					}
				}
			}
		}
		
		// Extract body
		if bodyStart > 0 && bodyStart < len(lines) {
			bodyLines := lines[bodyStart:]
			draft.Body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
		}
	} else {
		// No frontmatter, treat entire content as body
		draft.Body = string(content)
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

// parseEmailList parses a comma-separated list of email addresses
func parseEmailList(value string) []string {
	if value == "" {
		return []string{}
	}
	
	emails := strings.Split(value, ",")
	result := make([]string, 0, len(emails))
	for _, email := range emails {
		trimmed := strings.TrimSpace(email)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
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