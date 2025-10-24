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
	countOnly := false

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
		case "--count-only", "-c":
			countOnly = true
		}
	}

	fmt.Println("Searching emails...")
	
	// First get emails using basic read options
	emails, err := Read(advOpts.ReadOptions)
	if err != nil {
		return fmt.Errorf("SEARCH_READ_ERROR: Failed to retrieve emails from IMAP server or local storage. This could be due to: (1) IMAP connection issues, (2) Authentication problems, (3) Missing email configuration, (4) Local storage access errors. Original error: %v", err)
	}

	// Apply advanced filtering
	if advOpts.Query != "" || advOpts.MinSize > 0 || advOpts.MaxSize > 0 || 
	   advOpts.HasAttachments || advOpts.AttachmentSize > 0 || advOpts.DateRange != "" {
		emails, err = AdvancedSearchEmails(emails, advOpts)
		if err != nil {
			return fmt.Errorf("SEARCH_FILTER_ERROR: Failed to apply advanced search filters (query: '%s', min_size: %d, max_size: %d, has_attachments: %v, date_range: '%s'). This indicates an issue with search criteria processing or email content analysis. Original error: %v", 
				advOpts.Query, advOpts.MinSize, advOpts.MaxSize, advOpts.HasAttachments, advOpts.DateRange, err)
		}
	}

	if len(emails) == 0 {
		fmt.Printf("SEARCH_NO_RESULTS: No emails found matching search criteria. Applied filters: from='%s', to='%s', subject='%s', unread_only=%v, days_back=%v, limit=%d\n", 
			advOpts.FromAddress, advOpts.ToAddress, advOpts.Subject, advOpts.UnreadOnly, 
			func() interface{} { if advOpts.Since.IsZero() { return "none" } else { return advOpts.Since.Format("2006-01-02") } }(), 
			advOpts.Limit)
		return nil
	}

	if countOnly {
		fmt.Printf("Found %d emails matching search criteria\n", len(emails))
	} else {
		fmt.Printf("Found %d emails matching search criteria:\n\n", len(emails))
		// Display emails as snippets with IDs
		fmt.Print(FormatEmailList(emails))
	}

	// Optionally save to files if requested
	if saveToFile {
		err = SaveEmailsAsMarkdown(emails, outputDir)
		if err != nil {
			fmt.Printf("\nSEARCH_SAVE_WARNING: Failed to save %d emails to markdown files in directory '%s'. This could be due to: (1) Permission issues, (2) Disk space, (3) Invalid directory path, (4) File system errors. Error: %v\n", len(emails), outputDir, err)
		} else {
			fmt.Printf("\nSEARCH_SAVE_SUCCESS: %d emails saved to %s directory as markdown files\n", len(emails), outputDir)
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