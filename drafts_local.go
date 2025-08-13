package mailos

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// saveLocalDraft saves a draft to the local .email/drafts directory as markdown
func saveLocalDraft(draft DraftEmail) error {
	// Ensure directories exist
	if err := EnsureEmailDirectories(); err != nil {
		return fmt.Errorf("failed to create email directories: %v", err)
	}
	
	// Get drafts directory
	draftsDir, err := GetDraftsDir()
	if err != nil {
		return fmt.Errorf("failed to get drafts directory: %v", err)
	}
	
	// Get config to get the from email
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}
	
	// Determine the from email
	fromEmail := config.Email
	if config.FromEmail != "" {
		fromEmail = config.FromEmail
	}
	if config.FromName != "" {
		fromEmail = fmt.Sprintf("%s <%s>", config.FromName, fromEmail)
	}
	
	// Convert draft to EmailData
	emailData := ConvertDraftToEmailData(draft, fromEmail)
	
	// Generate filename
	filename := GenerateEmailFilename(draft.Subject, time.Now(), "draft")
	filepath := filepath.Join(draftsDir, filename)
	
	// Save using the common function
	if err := SaveEmailToMarkdown(emailData, filepath); err != nil {
		return fmt.Errorf("failed to write draft file: %v", err)
	}
	
	return nil
}

// listLocalDrafts lists drafts from the local .email/drafts directory
func listLocalDrafts(showFullContent bool) error {
	// Get drafts directory
	draftsDir, err := GetDraftsDir()
	if err != nil {
		return fmt.Errorf("failed to get drafts directory: %v", err)
	}
	
	// Read all markdown files from the directory
	files, err := os.ReadDir(draftsDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No local drafts found")
			return nil
		}
		return fmt.Errorf("failed to read drafts directory: %v", err)
	}
	
	fmt.Println("📁 Local drafts from .email/drafts folder")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	count := 0
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") {
			continue
		}
		
		filepath := filepath.Join(draftsDir, file.Name())
		
		// Parse the markdown file
		emailData, err := ParseMarkdownEmail(filepath)
		if err != nil {
			continue
		}
		
		count++
		fmt.Printf("\n📧 Draft #%d\n", count)
		fmt.Printf("  File: %s\n", file.Name())
		if len(emailData.To) > 0 {
			fmt.Printf("  To: %s\n", strings.Join(emailData.To, ", "))
		}
		if len(emailData.CC) > 0 {
			fmt.Printf("  CC: %s\n", strings.Join(emailData.CC, ", "))
		}
		fmt.Printf("  Subject: %s\n", emailData.Subject)
		if !emailData.Date.IsZero() {
			fmt.Printf("  Date: %s\n", emailData.Date.Format("Jan 2, 2006 at 3:04 PM"))
		}
		
		if showFullContent && emailData.Body != "" {
			fmt.Println("\n  Body:")
			fmt.Println("  ──────────────────────────────────────────")
			for _, line := range strings.Split(emailData.Body, "\n") {
				if strings.TrimSpace(line) != "" {
					fmt.Printf("  %s\n", line)
				}
			}
			fmt.Println("  ──────────────────────────────────────────")
		}
	}
	
	if count == 0 {
		fmt.Println("No local drafts found")
	} else {
		fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("Total: %d local draft(s)\n\n", count)
	}
	
	return nil
}

// ParseMarkdownEmail parses a markdown file with front matter and returns EmailData
func ParseMarkdownEmail(filepath string) (*EmailData, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	emailData := &EmailData{}
	scanner := bufio.NewScanner(file)
	inFrontMatter := false
	frontMatterStarted := false
	var bodyLines []string
	
	for scanner.Scan() {
		line := scanner.Text()
		
		if !frontMatterStarted && line == "---" {
			inFrontMatter = true
			frontMatterStarted = true
			continue
		}
		
		if inFrontMatter && line == "---" {
			inFrontMatter = false
			continue
		}
		
		if inFrontMatter {
			// Parse front matter line
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				// Remove surrounding quotes if present
				value = strings.Trim(value, "\"")
				
				switch key {
				case "from":
					emailData.From = value
				case "to":
					emailData.To = strings.Split(value, ", ")
				case "cc":
					emailData.CC = strings.Split(value, ", ")
				case "bcc":
					emailData.BCC = strings.Split(value, ", ")
				case "subject":
					emailData.Subject = value
				case "date":
					if t, err := time.Parse(time.RFC3339, value); err == nil {
						emailData.Date = t
					}
				case "priority":
					emailData.Priority = value
				}
			}
		} else if frontMatterStarted {
			// We're in the body
			bodyLines = append(bodyLines, line)
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	emailData.Body = strings.Join(bodyLines, "\n")
	return emailData, nil
}