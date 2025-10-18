package mailos

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type EmailStats struct {
	AccountEmail     string                 `json:"account_email"`
	TotalEmails      int                    `json:"total_emails"`
	DateRange        DateRange              `json:"date_range"`
	SenderStats      []ContactFrequency     `json:"sender_stats"`
	RecipientStats   []ContactFrequency     `json:"recipient_stats"`
	HourlyStats      map[int]int           `json:"hourly_stats"`
	DailyStats       map[string]int        `json:"daily_stats"`
	MonthlyStats     map[string]int        `json:"monthly_stats"`
	TopDomains       []DomainFrequency     `json:"top_domains"`
}

type ContactFrequency struct {
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Count     int       `json:"count"`
	LastEmail time.Time `json:"last_email"`
}

type DomainFrequency struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
}

type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type StatsOptions struct {
	AccountEmail string
	Since        time.Time
	Until        time.Time
	TopN         int
	IncludeBody  bool
}

func GenerateEmailStats(opts StatsOptions) (*EmailStats, error) {
	if opts.TopN == 0 {
		opts.TopN = 10
	}

	emails, err := GetEmailsFromInbox(opts.AccountEmail, ReadOptions{
		Since: opts.Since,
		Limit: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get emails: %v", err)
	}

	if !opts.Until.IsZero() {
		var filteredEmails []*Email
		for _, email := range emails {
			if email.Date.Before(opts.Until) || email.Date.Equal(opts.Until) {
				filteredEmails = append(filteredEmails, email)
			}
		}
		emails = filteredEmails
	}

	stats := &EmailStats{
		AccountEmail: opts.AccountEmail,
		TotalEmails:  len(emails),
		HourlyStats:  make(map[int]int),
		DailyStats:   make(map[string]int),
		MonthlyStats: make(map[string]int),
	}

	if len(emails) == 0 {
		return stats, nil
	}

	stats.DateRange = DateRange{
		Start: emails[len(emails)-1].Date,
		End:   emails[0].Date,
	}

	senderCounts := make(map[string]*ContactFrequency)
	recipientCounts := make(map[string]*ContactFrequency)
	domainCounts := make(map[string]int)

	for _, email := range emails {
		processSender(email, senderCounts, domainCounts)
		processRecipients(email, recipientCounts, opts.AccountEmail)
		processTimeStats(email, stats)
	}

	stats.SenderStats = sortContactFrequencies(senderCounts, opts.TopN)
	stats.RecipientStats = sortContactFrequencies(recipientCounts, opts.TopN)
	stats.TopDomains = sortDomainFrequencies(domainCounts, opts.TopN)

	return stats, nil
}

func processSender(email *Email, senderCounts map[string]*ContactFrequency, domainCounts map[string]int) {
	fromEmail := extractEmailAddress(email.From)
	if fromEmail == "" {
		return
	}

	if _, exists := senderCounts[fromEmail]; !exists {
		senderCounts[fromEmail] = &ContactFrequency{
			Email: fromEmail,
			Name:  extractName(email.From),
		}
	}
	
	senderCounts[fromEmail].Count++
	if email.Date.After(senderCounts[fromEmail].LastEmail) {
		senderCounts[fromEmail].LastEmail = email.Date
	}

	domain := extractDomain(fromEmail)
	if domain != "" {
		domainCounts[domain]++
	}
}

func processRecipients(email *Email, recipientCounts map[string]*ContactFrequency, accountEmail string) {
	for _, recipient := range email.To {
		recipientEmail := extractEmailAddress(strings.TrimSpace(recipient))
		if recipientEmail == "" || recipientEmail == accountEmail {
			continue
		}

		if _, exists := recipientCounts[recipientEmail]; !exists {
			recipientCounts[recipientEmail] = &ContactFrequency{
				Email: recipientEmail,
				Name:  extractName(recipient),
			}
		}
		
		recipientCounts[recipientEmail].Count++
		if email.Date.After(recipientCounts[recipientEmail].LastEmail) {
			recipientCounts[recipientEmail].LastEmail = email.Date
		}
	}
}

func processTimeStats(email *Email, stats *EmailStats) {
	hour := email.Date.Hour()
	stats.HourlyStats[hour]++

	day := email.Date.Format("2006-01-02")
	stats.DailyStats[day]++

	month := email.Date.Format("2006-01")
	stats.MonthlyStats[month]++
}

func sortContactFrequencies(counts map[string]*ContactFrequency, topN int) []ContactFrequency {
	var frequencies []ContactFrequency
	for _, freq := range counts {
		frequencies = append(frequencies, *freq)
	}

	sort.Slice(frequencies, func(i, j int) bool {
		return frequencies[i].Count > frequencies[j].Count
	})

	if len(frequencies) > topN {
		frequencies = frequencies[:topN]
	}

	return frequencies
}

func sortDomainFrequencies(counts map[string]int, topN int) []DomainFrequency {
	var frequencies []DomainFrequency
	for domain, count := range counts {
		frequencies = append(frequencies, DomainFrequency{
			Domain: domain,
			Count:  count,
		})
	}

	sort.Slice(frequencies, func(i, j int) bool {
		return frequencies[i].Count > frequencies[j].Count
	})

	if len(frequencies) > topN {
		frequencies = frequencies[:topN]
	}

	return frequencies
}

// extractEmailAddress is already defined in save.go
// func extractEmailAddress(fromField string) string {
// 	if strings.Contains(fromField, "<") && strings.Contains(fromField, ">") {
// 		start := strings.Index(fromField, "<")
// 		end := strings.Index(fromField, ">")
// 		if start < end {
// 			return strings.TrimSpace(fromField[start+1 : end])
// 		}
// 	}
// 	return strings.TrimSpace(fromField)
// }

func extractName(fromField string) string {
	if strings.Contains(fromField, "<") {
		name := strings.TrimSpace(fromField[:strings.Index(fromField, "<")])
		name = strings.Trim(name, "\"")
		return name
	}
	return ""
}

// extractDomain is already defined in query.go
// func extractDomain(email string) string {
// 	parts := strings.Split(email, "@")
// 	if len(parts) == 2 {
// 		return strings.ToLower(parts[1])
// 	}
// 	return ""
// }

func (stats *EmailStats) ToMarkdown() string {
	var md strings.Builder

	md.WriteString(fmt.Sprintf("# Email Statistics for %s\n\n", stats.AccountEmail))
	
	md.WriteString("## Overview\n\n")
	md.WriteString(fmt.Sprintf("- **Total Emails**: %d\n", stats.TotalEmails))
	if !stats.DateRange.Start.IsZero() && !stats.DateRange.End.IsZero() {
		md.WriteString(fmt.Sprintf("- **Date Range**: %s to %s\n", 
			stats.DateRange.Start.Format("2006-01-02"), 
			stats.DateRange.End.Format("2006-01-02")))
	}
	md.WriteString(fmt.Sprintf("- **Analysis Generated**: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	if len(stats.SenderStats) > 0 {
		md.WriteString("## Top Email Senders\n\n")
		md.WriteString("*People who send me the most emails*\n\n")
		md.WriteString("| Rank | Sender | Name | Email Count | Last Email |\n")
		md.WriteString("|------|--------|------|-------------|------------|\n")
		for i, sender := range stats.SenderStats {
			name := sender.Name
			if name == "" {
				name = "_No name_"
			}
			md.WriteString(fmt.Sprintf("| %d | %s | %s | %d | %s |\n", 
				i+1, sender.Email, name, sender.Count, 
				sender.LastEmail.Format("2006-01-02")))
		}
		md.WriteString("\n")
	}

	if len(stats.RecipientStats) > 0 {
		md.WriteString("## Top Email Recipients\n\n")
		md.WriteString("*People I email the most*\n\n")
		md.WriteString("| Rank | Recipient | Name | Email Count | Last Email |\n")
		md.WriteString("|------|-----------|------|-------------|------------|\n")
		for i, recipient := range stats.RecipientStats {
			name := recipient.Name
			if name == "" {
				name = "_No name_"
			}
			md.WriteString(fmt.Sprintf("| %d | %s | %s | %d | %s |\n", 
				i+1, recipient.Email, name, recipient.Count, 
				recipient.LastEmail.Format("2006-01-02")))
		}
		md.WriteString("\n")
	}

	if len(stats.TopDomains) > 0 {
		md.WriteString("## Top Email Domains\n\n")
		md.WriteString("*Most common email domains in my inbox*\n\n")
		md.WriteString("| Rank | Domain | Email Count |\n")
		md.WriteString("|------|--------|-------------|\n")
		for i, domain := range stats.TopDomains {
			md.WriteString(fmt.Sprintf("| %d | %s | %d |\n", i+1, domain.Domain, domain.Count))
		}
		md.WriteString("\n")
	}

	md.WriteString(stats.generateHourlyStatsMarkdown())
	md.WriteString(stats.generateMonthlyStatsMarkdown())

	return md.String()
}

func (stats *EmailStats) generateHourlyStatsMarkdown() string {
	if len(stats.HourlyStats) == 0 {
		return ""
	}

	var md strings.Builder
	md.WriteString("## Email Distribution by Hour of Day\n\n")
	md.WriteString("*When I receive the most emails*\n\n")
	md.WriteString("| Hour | Email Count | Visual |\n")
	md.WriteString("|------|-------------|--------|\n")

	maxCount := 0
	for _, count := range stats.HourlyStats {
		if count > maxCount {
			maxCount = count
		}
	}

	for hour := 0; hour < 24; hour++ {
		count := stats.HourlyStats[hour]
		visual := ""
		if maxCount > 0 {
			barLength := (count * 20) / maxCount
			visual = strings.Repeat("█", barLength)
		}
		
		timeLabel := fmt.Sprintf("%02d:00", hour)
		md.WriteString(fmt.Sprintf("| %s | %d | %s |\n", timeLabel, count, visual))
	}
	md.WriteString("\n")

	return md.String()
}

func (stats *EmailStats) generateMonthlyStatsMarkdown() string {
	if len(stats.MonthlyStats) == 0 {
		return ""
	}

	var md strings.Builder
	md.WriteString("## Email Distribution by Month\n\n")
	md.WriteString("*Email volume trends over time*\n\n")
	md.WriteString("| Month | Email Count |\n")
	md.WriteString("|-------|-------------|\n")

	var months []string
	for month := range stats.MonthlyStats {
		months = append(months, month)
	}
	sort.Strings(months)

	for _, month := range months {
		count := stats.MonthlyStats[month]
		md.WriteString(fmt.Sprintf("| %s | %d |\n", month, count))
	}
	md.WriteString("\n")

	return md.String()
}

func GetAccountStats(accountEmail string, topN int) (string, error) {
	if topN == 0 {
		topN = 10
	}

	opts := StatsOptions{
		AccountEmail: accountEmail,
		TopN:         topN,
	}

	stats, err := GenerateEmailStats(opts)
	if err != nil {
		return "", err
	}

	return stats.ToMarkdown(), nil
}

func GetAccountStatsWithDateRange(accountEmail string, since, until time.Time, topN int) (string, error) {
	if topN == 0 {
		topN = 10
	}

	opts := StatsOptions{
		AccountEmail: accountEmail,
		Since:        since,
		Until:        until,
		TopN:         topN,
	}

	stats, err := GenerateEmailStats(opts)
	if err != nil {
		return "", err
	}

	return stats.ToMarkdown(), nil
}

func GetAllAccountsStats(topN int) (string, error) {
	accounts, err := ListAccountInboxes()
	if err != nil {
		return "", fmt.Errorf("failed to list accounts: %v", err)
	}

	if len(accounts) == 0 {
		return "No email accounts found with inbox data.\n", nil
	}

	var md strings.Builder
	md.WriteString("# Email Statistics for All Accounts\n\n")

	for _, account := range accounts {
		accountStats, err := GetAccountStats(account, topN)
		if err != nil {
			md.WriteString(fmt.Sprintf("## Error for %s\n\nFailed to generate statistics: %v\n\n", account, err))
			continue
		}
		md.WriteString(accountStats)
		md.WriteString("---\n\n")
	}

	return md.String(), nil
}