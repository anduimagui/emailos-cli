package mailos

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type EmailStats struct {
	TotalEmails      int
	UnreadCount      int
	ReadCount        int
	EmailsByDate     map[string]int
	EmailsBySender   map[string]int
	EmailsByDomain   map[string]int
	EmailsByHour     map[int]int
	EmailsByWeekday  map[string]int
	SubjectKeywords  map[string]int
	AttachmentCount  int
	AverageBodySize  int
	TimeRange        string
	QueryDescription string
}

func GenerateEmailStats(emails []*Email, query QueryOptions) *EmailStats {
	// Apply advanced filters first
	emails = query.FilterEmails(emails)
	
	stats := &EmailStats{
		TotalEmails:      len(emails),
		EmailsByDate:     make(map[string]int),
		EmailsBySender:   make(map[string]int),
		EmailsByDomain:   make(map[string]int),
		EmailsByHour:     make(map[int]int),
		EmailsByWeekday:  make(map[string]int),
		SubjectKeywords:  make(map[string]int),
		QueryDescription: query.GetDescription(),
	}
	
	if query.TimeRange != "" {
		stats.TimeRange = query.TimeRange
	} else if query.Days > 0 {
		stats.TimeRange = fmt.Sprintf("Last %d days", query.Days)
	} else if !query.Since.IsZero() {
		stats.TimeRange = fmt.Sprintf("Since %s", query.Since.Format("Jan 2, 2006"))
	} else {
		stats.TimeRange = "All time"
	}
	
	totalBodySize := 0
	
	for _, email := range emails {
		// Count by date
		dateKey := email.Date.Format("2006-01-02")
		stats.EmailsByDate[dateKey]++
		
		// Count by sender
		sender := extractEmailAddress(email.From)
		stats.EmailsBySender[sender]++
		
		// Count by domain
		if idx := strings.Index(sender, "@"); idx >= 0 {
			domain := sender[idx+1:]
			stats.EmailsByDomain[domain]++
		}
		
		// Count by hour
		hour := email.Date.Hour()
		stats.EmailsByHour[hour]++
		
		// Count by weekday
		weekday := email.Date.Weekday().String()
		stats.EmailsByWeekday[weekday]++
		
		// Extract subject keywords (simple implementation)
		words := strings.Fields(strings.ToLower(email.Subject))
		for _, word := range words {
			// Skip common words
			if len(word) > 3 && !isCommonWord(word) {
				cleaned := strings.Trim(word, ".,!?;:")
				if cleaned != "" {
					stats.SubjectKeywords[cleaned]++
				}
			}
		}
		
		// Count attachments
		if len(email.Attachments) > 0 {
			stats.AttachmentCount += len(email.Attachments)
		}
		
		// Sum body sizes
		totalBodySize += len(email.Body)
	}
	
	if stats.TotalEmails > 0 {
		stats.AverageBodySize = totalBodySize / stats.TotalEmails
	}
	
	return stats
}

func isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the": true, "and": true, "for": true, "are": true,
		"you": true, "your": true, "with": true, "from": true,
		"this": true, "that": true, "have": true, "will": true,
		"been": true, "were": true, "what": true, "when": true,
		"where": true, "which": true, "while": true, "about": true,
	}
	return commonWords[word]
}

func FormatEmailStats(stats *EmailStats) string {
	var output strings.Builder
	
	output.WriteString("\n")
	output.WriteString("📊 Email Statistics Report\n")
	output.WriteString("══════════════════════════════════════════════\n\n")
	
	output.WriteString(fmt.Sprintf("Query: %s\n", stats.QueryDescription))
	output.WriteString(fmt.Sprintf("Time Range: %s\n", stats.TimeRange))
	output.WriteString(fmt.Sprintf("Total Emails: %d\n", stats.TotalEmails))
	
	if stats.TotalEmails == 0 {
		output.WriteString("\nNo emails found matching the criteria.\n")
		return output.String()
	}
	
	output.WriteString(fmt.Sprintf("Emails with Attachments: %d\n", stats.AttachmentCount))
	output.WriteString(fmt.Sprintf("Average Email Size: %d bytes\n\n", stats.AverageBodySize))
	
	// Top senders
	output.WriteString("📧 Top Senders:\n")
	output.WriteString("──────────────────────────────────────────────\n")
	topSenders := getTopItems(stats.EmailsBySender, 10)
	for _, item := range topSenders {
		percentage := float64(item.Count) * 100 / float64(stats.TotalEmails)
		output.WriteString(fmt.Sprintf("  %-40s %4d (%5.1f%%)\n", 
			truncateString(item.Key, 40), item.Count, percentage))
	}
	output.WriteString("\n")
	
	// Top domains
	output.WriteString("🌐 Top Domains:\n")
	output.WriteString("──────────────────────────────────────────────\n")
	topDomains := getTopItems(stats.EmailsByDomain, 10)
	for _, item := range topDomains {
		percentage := float64(item.Count) * 100 / float64(stats.TotalEmails)
		output.WriteString(fmt.Sprintf("  %-30s %4d (%5.1f%%)\n", 
			truncateString(item.Key, 30), item.Count, percentage))
	}
	output.WriteString("\n")
	
	// Activity by hour
	output.WriteString("⏰ Activity by Hour:\n")
	output.WriteString("──────────────────────────────────────────────\n")
	for hour := 0; hour < 24; hour++ {
		count := stats.EmailsByHour[hour]
		if count > 0 {
			bar := strings.Repeat("█", min(count, 50))
			output.WriteString(fmt.Sprintf("  %02d:00  %s %d\n", hour, bar, count))
		}
	}
	output.WriteString("\n")
	
	// Activity by weekday
	output.WriteString("📅 Activity by Weekday:\n")
	output.WriteString("──────────────────────────────────────────────\n")
	weekdays := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	for _, day := range weekdays {
		count := stats.EmailsByWeekday[day]
		if count > 0 {
			percentage := float64(count) * 100 / float64(stats.TotalEmails)
			bar := strings.Repeat("█", int(percentage/2))
			output.WriteString(fmt.Sprintf("  %-10s %s %d (%.1f%%)\n", day, bar, count, percentage))
		}
	}
	output.WriteString("\n")
	
	// Top subject keywords
	output.WriteString("🔤 Top Subject Keywords:\n")
	output.WriteString("──────────────────────────────────────────────\n")
	topKeywords := getTopItems(stats.SubjectKeywords, 15)
	for i, item := range topKeywords {
		if i > 0 && i%5 == 0 {
			output.WriteString("\n")
		}
		output.WriteString(fmt.Sprintf("  %-12s(%d)", item.Key, item.Count))
	}
	output.WriteString("\n\n")
	
	// Daily distribution (last 30 days or available range)
	output.WriteString("📈 Daily Distribution:\n")
	output.WriteString("──────────────────────────────────────────────\n")
	
	// Sort dates
	var dates []string
	for date := range stats.EmailsByDate {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	
	// Show last 30 days or all if less
	startIdx := 0
	if len(dates) > 30 {
		startIdx = len(dates) - 30
	}
	
	maxCount := 0
	for _, date := range dates[startIdx:] {
		if stats.EmailsByDate[date] > maxCount {
			maxCount = stats.EmailsByDate[date]
		}
	}
	
	for _, date := range dates[startIdx:] {
		count := stats.EmailsByDate[date]
		// Parse date for better formatting
		if t, err := time.Parse("2006-01-02", date); err == nil {
			dateFormatted := t.Format("Jan 02")
			barLength := 0
			if maxCount > 0 {
				barLength = (count * 30) / maxCount
			}
			bar := strings.Repeat("█", barLength)
			output.WriteString(fmt.Sprintf("  %s  %s %d\n", dateFormatted, bar, count))
		}
	}
	
	output.WriteString("\n══════════════════════════════════════════════\n")
	
	return output.String()
}

type countItem struct {
	Key   string
	Count int
}

func getTopItems(items map[string]int, limit int) []countItem {
	var result []countItem
	for key, count := range items {
		result = append(result, countItem{Key: key, Count: count})
	}
	
	// Sort by count (descending)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})
	
	// Limit results
	if len(result) > limit {
		result = result[:limit]
	}
	
	return result
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

