package mailos

import (
	"crypto/tls"
	"fmt"
	"regexp"
	"strings"
	
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

// UnsubscribeLinks holds unsubscribe information for an email
type UnsubscribeLinks struct {
	Email   *Email
	Links   []string
	Sender  string
	Subject string
}

// FindUnsubscribeLinks searches for unsubscribe links in emails
func FindUnsubscribeLinks(emails []*Email) []UnsubscribeLinks {
	var results []UnsubscribeLinks
	
	for _, email := range emails {
		links := extractUnsubscribeLinks(email)
		if len(links) > 0 {
			results = append(results, UnsubscribeLinks{
				Email:   email,
				Links:   links,
				Sender:  email.From,
				Subject: email.Subject,
			})
		}
	}
	
	return results
}

// extractUnsubscribeLinks extracts unsubscribe links from an email
func extractUnsubscribeLinks(email *Email) []string {
	var unsubscribeLinks []string
	seen := make(map[string]bool)
	
	// First check List-Unsubscribe headers (RFC 2369/8058)
	if email.Headers != nil {
		if listUnsub, exists := email.Headers["List-Unsubscribe"]; exists {
			for _, header := range listUnsub {
				// Extract URLs from angle brackets
				urlPattern := regexp.MustCompile(`<(https?://[^>]+)>`)
				matches := urlPattern.FindAllStringSubmatch(header, -1)
				for _, match := range matches {
					if len(match) >= 2 {
						cleanedURL := cleanURL(match[1])
						if cleanedURL != "" && !seen[cleanedURL] {
							seen[cleanedURL] = true
							unsubscribeLinks = append(unsubscribeLinks, cleanedURL)
						}
					}
				}
			}
		}
	}
	
	// Enhanced unsubscribe URL patterns
	patterns := []string{
		// Basic patterns
		`https?://[^\s<>"']+unsubscribe[^\s<>"']*`,
		`https?://[^\s<>"']+/unsub[^\s<>"']*`,
		`https?://[^\s<>"']+/opt-out[^\s<>"']*`,
		`https?://[^\s<>"']+/opt_out[^\s<>"']*`,
		`https?://[^\s<>"']+/preferences[^\s<>"']*`,
		`https?://[^\s<>"']+/email-preferences[^\s<>"']*`,
		`https?://[^\s<>"']+/email_preferences[^\s<>"']*`,
		`https?://[^\s<>"']+/manage[^\s<>"']*subscription[^\s<>"']*`,
		`https?://[^\s<>"']+/email/preferences[^\s<>"']*`,
		`https?://[^\s<>"']+/settings/notifications[^\s<>"']*`,
		`https?://[^\s<>"']+/remove[^\s<>"']*`,
		`https?://[^\s<>"']+/stop[^\s<>"']*`,
		`https?://[^\s<>"']+/cancel[^\s<>"']*`,
		`https?://[^\s<>"']+/leave[^\s<>"']*`,
		`https?://[^\s<>"']+/signout[^\s<>"']*`,
		`https?://[^\s<>"']+/sign-out[^\s<>"']*`,
		`https?://[^\s<>"']+/update[^\s<>"']*profile[^\s<>"']*`,
		
		// Service-specific patterns
		`https?://[^\s<>"']*beehiiv\.com/[^\s<>"']*unsubscribe[^\s<>"']*`,
		`https?://[^\s<>"']*beehiiv\.com/[^\s<>"']*preferences[^\s<>"']*`,
		`https?://[^\s<>"']*mailchimp\.com/[^\s<>"']*unsubscribe[^\s<>"']*`,
		`https?://[^\s<>"']*sendgrid\.net/[^\s<>"']*unsubscribe[^\s<>"']*`,
		`https?://[^\s<>"']*convertkit[^\s<>"']*\.com/[^\s<>"']*unsubscribe[^\s<>"']*`,
		`https?://[^\s<>"']*convertkit[^\s<>"']*\.com/[^\s<>"']*preferences[^\s<>"']*`,
		`https?://[^\s<>"']*constantcontact\.com/[^\s<>"']*unsubscribe[^\s<>"']*`,
		`https?://[^\s<>"']*campaignmonitor\.com/[^\s<>"']*unsubscribe[^\s<>"']*`,
		`https?://[^\s<>"']*aweber\.com/[^\s<>"']*unsubscribe[^\s<>"']*`,
		`https?://[^\s<>"']*getresponse\.com/[^\s<>"']*unsubscribe[^\s<>"']*`,
		`https?://[^\s<>"']*activecampaign\.com/[^\s<>"']*unsubscribe[^\s<>"']*`,
		`https?://[^\s<>"']*hubspot\.com/[^\s<>"']*unsubscribe[^\s<>"']*`,
		`https?://[^\s<>"']*pardot\.com/[^\s<>"']*unsubscribe[^\s<>"']*`,
		`https?://[^\s<>"']*marketo\.com/[^\s<>"']*unsubscribe[^\s<>"']*`,
	}
	
	// Search in both plain text and HTML body
	bodies := []string{email.Body}
	if email.BodyHTML != "" {
		bodies = append(bodies, email.BodyHTML)
	}
	
	for _, body := range bodies {
		// Look for pattern matches
		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindAllString(body, -1)
			for _, match := range matches {
				// Clean up the URL
				cleanURL := cleanURL(match)
				if cleanURL != "" && !seen[cleanURL] {
					seen[cleanURL] = true
					unsubscribeLinks = append(unsubscribeLinks, cleanURL)
				}
			}
		}
		
		// Look for links with unsubscribe-related text in HTML
		if strings.Contains(body, "<a") {
			// Find links with unsubscribe text
			linkPattern := regexp.MustCompile(`<a[^>]+href=["']([^"']+)["'][^>]*>([^<]*)</a>`)
			linkMatches := linkPattern.FindAllStringSubmatch(body, -1)
			
			for _, match := range linkMatches {
				if len(match) >= 3 {
					url := match[1]
					linkText := match[2]
					
					// Check if link text contains unsubscribe-related words
					unsubPattern := regexp.MustCompile(`(?i)unsubscribe|opt.?out|preferences|manage|settings|stop.?receiv|remove|cancel|leave|sign.?out|update.?profile|email.?preferences`)
					if unsubPattern.MatchString(linkText) && strings.HasPrefix(url, "http") {
						cleanedURL := cleanURL(url)
						if cleanedURL != "" && !seen[cleanedURL] {
							seen[cleanedURL] = true
							unsubscribeLinks = append(unsubscribeLinks, cleanedURL)
						}
					}
				}
			}
			
			// Look for unsubscribe text near links
			// Pattern: unsubscribe...<a href="...">
			beforePattern := regexp.MustCompile(`(?i)unsubscribe[^<]*<a[^>]+href=["']([^"']+)["']`)
			beforeMatches := beforePattern.FindAllStringSubmatch(body, -1)
			for _, match := range beforeMatches {
				if len(match) >= 2 && strings.HasPrefix(match[1], "http") {
					cleanedURL := cleanURL(match[1])
					if cleanedURL != "" && !seen[cleanedURL] {
						seen[cleanedURL] = true
						unsubscribeLinks = append(unsubscribeLinks, cleanedURL)
					}
				}
			}
			
			// Look for standalone "Unsubscribe" text followed by pipe and URLs
			// Pattern: Unsubscribe | other text with URLs nearby
			pipePattern := regexp.MustCompile(`(?i)unsubscribe\s*\|[^|]*?(https?://[^\s<>"|']+)`)
			pipeMatches := pipePattern.FindAllStringSubmatch(body, -1)
			for _, match := range pipeMatches {
				if len(match) >= 2 {
					cleanedURL := cleanURL(match[1])
					if cleanedURL != "" && !seen[cleanedURL] {
						seen[cleanedURL] = true
						unsubscribeLinks = append(unsubscribeLinks, cleanedURL)
					}
				}
			}
			
			// Look for "Unsubscribe" text with URLs in the same line or nearby context
			// This catches simple cases where "Unsubscribe" appears as plain text
			lines := strings.Split(body, "\n")
			for i, line := range lines {
				if strings.Contains(strings.ToLower(line), "unsubscribe") {
					// Check current line and next few lines for URLs
					for j := i; j < len(lines) && j < i+3; j++ {
						urlPattern := regexp.MustCompile(`https?://[^\s<>"|']+`)
						urls := urlPattern.FindAllString(lines[j], -1)
						for _, url := range urls {
							cleanedURL := cleanURL(url)
							if cleanedURL != "" && !seen[cleanedURL] {
								seen[cleanedURL] = true
								unsubscribeLinks = append(unsubscribeLinks, cleanedURL)
							}
						}
					}
				}
			}
			
			// Pattern: <a href="...">...unsubscribe
			inLinkPattern := regexp.MustCompile(`<a[^>]+href=["']([^"']+)["'][^>]*>[^<]*unsubscribe`)
			inLinkMatches := inLinkPattern.FindAllStringSubmatch(body, -1)
			for _, match := range inLinkMatches {
				if len(match) >= 2 && strings.HasPrefix(match[1], "http") {
					cleanedURL := cleanURL(match[1])
					if cleanedURL != "" && !seen[cleanedURL] {
						seen[cleanedURL] = true
						unsubscribeLinks = append(unsubscribeLinks, cleanedURL)
					}
				}
			}
		}
	}
	
	return unsubscribeLinks
}

// cleanURL cleans and validates a URL
func cleanURL(url string) string {
	// Trim common trailing characters
	url = strings.TrimRight(url, ".,;:)]}'\"")
	
	// Decode HTML entities
	url = strings.ReplaceAll(url, "&amp;", "&")
	url = strings.ReplaceAll(url, "&lt;", "<")
	url = strings.ReplaceAll(url, "&gt;", ">")
	url = strings.ReplaceAll(url, "&quot;", "\"")
	
	// Ensure it's a valid HTTP(S) URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return ""
	}
	
	return url
}

// GetUnsubscribeReport generates a report of unsubscribe links found
func GetUnsubscribeReport(links []UnsubscribeLinks) string {
	if len(links) == 0 {
		return "No unsubscribe links found."
	}
	
	var report strings.Builder
	report.WriteString(fmt.Sprintf("Found unsubscribe links in %d emails:\n\n", len(links)))
	
	for i, item := range links {
		report.WriteString(fmt.Sprintf("%d. From: %s\n", i+1, item.Sender))
		report.WriteString(fmt.Sprintf("   Subject: %s\n", item.Subject))
		report.WriteString("   Unsubscribe links:\n")
		for _, link := range item.Links {
			report.WriteString(fmt.Sprintf("   - %s\n", link))
		}
		report.WriteString("\n")
	}
	
	return report.String()
}

// MoveEmailsToUnsubscribeFolder moves emails containing unsubscribe links to a dedicated IMAP folder
func MoveEmailsToUnsubscribeFolder(links []UnsubscribeLinks) error {
	if len(links) == 0 {
		return nil
	}
	
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}
	
	// Connect to IMAP server
	c, err := client.DialTLS(getIMAPServer(config.Provider), &tls.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to IMAP server: %v", err)
	}
	defer c.Close()
	
	// Login
	if err := c.Login(config.Email, config.Password); err != nil {
		return fmt.Errorf("failed to login to IMAP: %v", err)
	}
	defer c.Logout()
	
	// Create or find "Unsubscribe" folder
	unsubscribeFolder := "Unsubscribe"
	if err := createFolderIfNotExists(c, unsubscribeFolder); err != nil {
		return fmt.Errorf("failed to create unsubscribe folder: %v", err)
	}
	
	// Select INBOX first to move emails
	_, err = c.Select("INBOX", false)
	if err != nil {
		return fmt.Errorf("failed to select INBOX: %v", err)
	}
	
	// Collect unique email IDs
	emailIDs := make(map[uint32]bool)
	for _, link := range links {
		emailIDs[link.Email.ID] = true
	}
	
	// Convert map to slice
	var ids []uint32
	for id := range emailIDs {
		ids = append(ids, id)
	}
	
	if len(ids) == 0 {
		return nil
	}
	
	// Create sequence set from IDs
	seqSet := new(imap.SeqSet)
	for _, id := range ids {
		seqSet.AddNum(id)
	}
	
	// Move emails to Unsubscribe folder
	if err := c.Move(seqSet, unsubscribeFolder); err != nil {
		return fmt.Errorf("failed to move emails to unsubscribe folder: %v", err)
	}
	
	fmt.Printf("Moved %d emails with unsubscribe links to %s folder\n", len(ids), unsubscribeFolder)
	return nil
}

// createFolderIfNotExists creates an IMAP folder if it doesn't exist
func createFolderIfNotExists(c *client.Client, folderName string) error {
	// List folders to check if it exists
	folders := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", folders)
	}()
	
	folderExists := false
	for folder := range folders {
		if folder.Name == folderName {
			folderExists = true
			break
		}
	}
	
	if err := <-done; err != nil {
		return err
	}
	
	// Create folder if it doesn't exist
	if !folderExists {
		if err := c.Create(folderName); err != nil {
			return fmt.Errorf("failed to create folder %s: %v", folderName, err)
		}
		fmt.Printf("Created %s folder\n", folderName)
	}
	
	return nil
}

// getIMAPServer returns the IMAP server for a given provider
func getIMAPServer(provider string) string {
	switch provider {
	case "gmail":
		return "imap.gmail.com:993"
	case "fastmail":
		return "imap.fastmail.com:993"
	case "outlook":
		return "outlook.office365.com:993"
	case "yahoo":
		return "imap.mail.yahoo.com:993"
	case "zoho":
		return "imap.zoho.com:993"
	default:
		return "imap.gmail.com:993"
	}
}