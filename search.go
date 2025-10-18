package mailos

import (
	"fmt"
	"strconv"
)

// SearchCommand handles the search command functionality with advanced capabilities
func SearchCommand(args []string) error {
	// Use advanced search options
	advOpts := AdvancedSearchOptions{
		ReadOptions: ReadOptions{
			Limit: 20, // Default
		},
		FuzzyThreshold: 0.7,  // Default fuzzy threshold
		EnableFuzzy:    true, // Enable fuzzy by default
		CaseSensitive:  false, // Case insensitive by default
	}
	
	saveToFile := false
	outputDir := "emails"

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--unread":
			advOpts.UnreadOnly = true
		case "--from":
			if i+1 < len(args) {
				advOpts.FromAddress = args[i+1]
				i++
			}
		case "--to":
			if i+1 < len(args) {
				advOpts.ToAddress = args[i+1]
				i++
			}
		case "--subject":
			if i+1 < len(args) {
				advOpts.Subject = args[i+1]
				i++
			}
		case "--days":
			if i+1 < len(args) {
				advOpts.Since = getTimeFromDays(args[i+1])
				i++
			}
		case "-n", "--number":
			if i+1 < len(args) {
				if n := parseNumber(args[i+1]); n > 0 {
					advOpts.Limit = n
				}
				i++
			}
		case "--save":
			saveToFile = true
		case "--output-dir":
			if i+1 < len(args) {
				outputDir = args[i+1]
				i++
			}
		case "--local":
			advOpts.LocalOnly = true
		case "--sync":
			advOpts.SyncLocal = true
		
		// Advanced search options
		case "--query", "-q":
			if i+1 < len(args) {
				advOpts.Query = args[i+1]
				i++
			}
		case "--fuzzy-threshold":
			if i+1 < len(args) {
				if threshold := parseFloat(args[i+1]); threshold >= 0 && threshold <= 1 {
					advOpts.FuzzyThreshold = threshold
				}
				i++
			}
		case "--no-fuzzy":
			advOpts.EnableFuzzy = false
		case "--case-sensitive":
			advOpts.CaseSensitive = true
		case "--min-size":
			if i+1 < len(args) {
				if size, err := ParseSize(args[i+1]); err == nil {
					advOpts.MinSize = size
				}
				i++
			}
		case "--max-size":
			if i+1 < len(args) {
				if size, err := ParseSize(args[i+1]); err == nil {
					advOpts.MaxSize = size
				}
				i++
			}
		case "--has-attachments":
			advOpts.HasAttachments = true
		case "--attachment-size":
			if i+1 < len(args) {
				if size, err := ParseSize(args[i+1]); err == nil {
					advOpts.AttachmentSize = size
				}
				i++
			}
		case "--date-range":
			if i+1 < len(args) {
				advOpts.DateRange = args[i+1]
				i++
			}
		}
	}

	client, err := NewClient()
	if err != nil {
		return err
	}

	fmt.Println("Searching emails...")
	
	// First get emails using basic read options
	emails, err := client.ReadEmails(advOpts.ReadOptions)
	if err != nil {
		return fmt.Errorf("failed to search emails: %v", err)
	}

	// Apply advanced filtering
	if advOpts.Query != "" || advOpts.MinSize > 0 || advOpts.MaxSize > 0 || 
	   advOpts.HasAttachments || advOpts.AttachmentSize > 0 || advOpts.DateRange != "" {
		emails, err = AdvancedSearchEmails(emails, advOpts)
		if err != nil {
			return fmt.Errorf("failed to apply advanced search: %v", err)
		}
	}

	if len(emails) == 0 {
		fmt.Println("No emails found matching search criteria.")
		return nil
	}

	fmt.Printf("Found %d emails matching search criteria:\n\n", len(emails))

	// Display emails as snippets with IDs
	fmt.Print(FormatEmailList(emails))

	// Optionally save to files if requested
	if saveToFile {
		err = SaveEmailsAsMarkdown(emails, outputDir)
		if err != nil {
			fmt.Printf("\nWarning: Failed to save emails to files: %v\n", err)
		} else {
			fmt.Printf("\nEmails saved to %s directory\n", outputDir)
		}
	}

	return nil
}

// handleSearchCommand handles the /search command in interactive mode
func handleSearchCommand(args []string) error {
	return SearchCommand(args)
}

// parseFloat parses a string to float64, returns 0 on error
func parseFloat(s string) float64 {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return 0
}