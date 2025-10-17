package mailos

import (
	"fmt"
)

// SearchCommand handles the search command functionality
func SearchCommand(args []string) error {
	opts := ReadOptions{
		Limit: 20, // Default
	}
	
	saveToFile := false
	outputDir := "emails"

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--unread":
			opts.UnreadOnly = true
		case "--from":
			if i+1 < len(args) {
				opts.FromAddress = args[i+1]
				i++
			}
		case "--to":
			if i+1 < len(args) {
				opts.ToAddress = args[i+1]
				i++
			}
		case "--subject":
			if i+1 < len(args) {
				opts.Subject = args[i+1]
				i++
			}
		case "--days":
			if i+1 < len(args) {
				opts.Since = getTimeFromDays(args[i+1])
				i++
			}
		case "-n", "--number":
			if i+1 < len(args) {
				if n := parseNumber(args[i+1]); n > 0 {
					opts.Limit = n
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
			opts.LocalOnly = true
		case "--sync":
			opts.SyncLocal = true
		}
	}

	client, err := NewClient()
	if err != nil {
		return err
	}

	fmt.Println("Searching emails...")
	emails, err := client.ReadEmails(opts)
	if err != nil {
		return fmt.Errorf("failed to search emails: %v", err)
	}

	if len(emails) == 0 {
		fmt.Println("No emails found matching search criteria.")
		return nil
	}

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