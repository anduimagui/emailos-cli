package mailos

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// saveLocalDraft saves a draft to the local .email/drafts directory as JSON
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
	
	// Create a SavedEmail struct for the draft
	savedEmail := SavedEmail{
		ID:          fmt.Sprintf("draft_%d", time.Now().Unix()),
		From:        "", // Will be filled when sending
		To:          draft.To,
		CC:          draft.CC,
		BCC:         draft.BCC,
		Subject:     draft.Subject,
		Body:        draft.Body,
		Date:        time.Now(),
		Attachments: draft.Attachments,
	}
	
	// Generate filename with timestamp
	filename := fmt.Sprintf("%s_draft_%s.json",
		time.Now().Format("20060102_150405"),
		strings.ReplaceAll(strings.ReplaceAll(draft.Subject, "/", "_"), " ", "_"))
	
	// Ensure filename is not too long
	if len(filename) > 100 {
		filename = filename[:100] + ".json"
	}
	
	// Clean filename of problematic characters
	filename = strings.ReplaceAll(filename, "<", "")
	filename = strings.ReplaceAll(filename, ">", "")
	filename = strings.ReplaceAll(filename, ":", "")
	filename = strings.ReplaceAll(filename, "\"", "")
	filename = strings.ReplaceAll(filename, "|", "")
	filename = strings.ReplaceAll(filename, "?", "")
	filename = strings.ReplaceAll(filename, "*", "")
	
	filepath := filepath.Join(draftsDir, filename)
	
	// Marshal to JSON
	data, err := json.MarshalIndent(savedEmail, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal draft: %v", err)
	}
	
	// Write to file
	if err := os.WriteFile(filepath, data, 0600); err != nil {
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
	
	// Read all JSON files from the directory
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
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		
		filepath := filepath.Join(draftsDir, file.Name())
		data, err := os.ReadFile(filepath)
		if err != nil {
			continue
		}
		
		var savedEmail SavedEmail
		if err := json.Unmarshal(data, &savedEmail); err != nil {
			continue
		}
		
		count++
		fmt.Printf("\n📧 Draft #%d\n", count)
		fmt.Printf("  File: %s\n", file.Name())
		fmt.Printf("  To: %s\n", strings.Join(savedEmail.To, ", "))
		if len(savedEmail.CC) > 0 {
			fmt.Printf("  CC: %s\n", strings.Join(savedEmail.CC, ", "))
		}
		fmt.Printf("  Subject: %s\n", savedEmail.Subject)
		fmt.Printf("  Date: %s\n", savedEmail.Date.Format("Jan 2, 2006 at 3:04 PM"))
		
		if showFullContent && savedEmail.Body != "" {
			fmt.Println("\n  Body:")
			fmt.Println("  ──────────────────────────────────────────")
			for _, line := range strings.Split(savedEmail.Body, "\n") {
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