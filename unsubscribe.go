package mailos

import (
	"fmt"
	"regexp"
	"strings"
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
	
	// Common unsubscribe URL patterns
	patterns := []string{
		`https?://[^\s<>"']+unsubscribe[^\s<>"']*`,
		`https?://[^\s<>"']+/unsub[^\s<>"']*`,
		`https?://[^\s<>"']+/opt-out[^\s<>"']*`,
		`https?://[^\s<>"']+/preferences[^\s<>"']*`,
		`https?://[^\s<>"']+/email-preferences[^\s<>"']*`,
		`https?://[^\s<>"']+/manage[^\s<>"']*subscription[^\s<>"']*`,
		`https?://[^\s<>"']+/email/preferences[^\s<>"']*`,
		`https?://[^\s<>"']+/settings/notifications[^\s<>"']*`,
		// Service-specific patterns
		`https?://[^\s<>"']*beehiiv\.com/[^\s<>"']*unsubscribe[^\s<>"']*`,
		`https?://[^\s<>"']*beehiiv\.com/[^\s<>"']*preferences[^\s<>"']*`,
		`https?://[^\s<>"']*mailchimp\.com/[^\s<>"']*unsubscribe[^\s<>"']*`,
		`https?://[^\s<>"']*sendgrid\.net/[^\s<>"']*unsubscribe[^\s<>"']*`,
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
					unsubPattern := regexp.MustCompile(`(?i)unsubscribe|opt.?out|preferences|manage|settings|stop.?receiv`)
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