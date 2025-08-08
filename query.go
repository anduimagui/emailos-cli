package mailos

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type QueryOptions struct {
	Limit          int
	UnreadOnly     bool
	FromAddress    string
	ToAddress      string
	Subject        string
	Since          time.Time
	Until          time.Time
	Days           int
	TimeRange      string
	SentOnly       bool
	ReceivedOnly   bool
	HasAttachments bool
	MinSize        int
	MaxSize        int
	Domains        []string
	ExcludeDomains []string
	Keywords       []string
	ExcludeWords   []string
	GroupBy        string
	SortBy         string
	Format         string
	TopN           int
}

func NewQueryOptions() QueryOptions {
	return QueryOptions{
		Limit: 100,
		TopN:  10,
	}
}

func (q *QueryOptions) ParseArgs(args []string) error {
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		
		switch key {
		case "from":
			q.FromAddress = value
		case "to":
			q.ToAddress = value
		case "subject":
			q.Subject = value
		case "limit", "n":
			if n, err := strconv.Atoi(value); err == nil {
				q.Limit = n
			}
		case "days":
			if days, err := strconv.Atoi(value); err == nil {
				q.Days = days
				q.Since = time.Now().AddDate(0, 0, -days)
			}
		case "range":
			q.TimeRange = value
			if tr, err := ParseTimeRangeString(value); err == nil {
				q.Since = tr.Since
				q.Until = tr.Until
			}
		case "unread":
			q.UnreadOnly = value == "true" || value == "yes" || value == "1"
		case "sent":
			q.SentOnly = value == "true" || value == "yes" || value == "1"
		case "received":
			q.ReceivedOnly = value == "true" || value == "yes" || value == "1"
		case "attachments", "has-attachments":
			q.HasAttachments = value == "true" || value == "yes" || value == "1"
		case "min-size":
			if size, err := parseSize(value); err == nil {
				q.MinSize = size
			}
		case "max-size":
			if size, err := parseSize(value); err == nil {
				q.MaxSize = size
			}
		case "domain", "domains":
			q.Domains = strings.Split(value, ",")
		case "exclude-domain", "exclude-domains":
			q.ExcludeDomains = strings.Split(value, ",")
		case "keyword", "keywords":
			q.Keywords = strings.Split(value, ",")
		case "exclude", "exclude-words":
			q.ExcludeWords = strings.Split(value, ",")
		case "group-by", "groupby":
			q.GroupBy = value
		case "sort-by", "sortby", "sort":
			q.SortBy = value
		case "format":
			q.Format = value
		case "top", "top-n":
			if n, err := strconv.Atoi(value); err == nil {
				q.TopN = n
			}
		}
	}
	
	return nil
}

func parseSize(value string) (int, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	
	multiplier := 1
	if strings.HasSuffix(value, "kb") || strings.HasSuffix(value, "k") {
		multiplier = 1024
		value = strings.TrimSuffix(strings.TrimSuffix(value, "kb"), "k")
	} else if strings.HasSuffix(value, "mb") || strings.HasSuffix(value, "m") {
		multiplier = 1024 * 1024
		value = strings.TrimSuffix(strings.TrimSuffix(value, "mb"), "m")
	} else if strings.HasSuffix(value, "gb") || strings.HasSuffix(value, "g") {
		multiplier = 1024 * 1024 * 1024
		value = strings.TrimSuffix(strings.TrimSuffix(value, "gb"), "g")
	}
	
	if n, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
		return n * multiplier, nil
	}
	
	return 0, fmt.Errorf("invalid size format")
}

func (q QueryOptions) ToReadOptions() ReadOptions {
	return ReadOptions{
		Limit:       q.Limit,
		UnreadOnly:  q.UnreadOnly,
		FromAddress: q.FromAddress,
		ToAddress:   q.ToAddress,
		Subject:     q.Subject,
		Since:       q.Since,
	}
}

func (q QueryOptions) FilterEmails(emails []*Email) []*Email {
	var filtered []*Email
	
	for _, email := range emails {
		// Filter by date range
		if !q.Until.IsZero() {
			if !email.Date.After(q.Since.Add(-time.Second)) || 
			   !email.Date.Before(q.Until.Add(time.Second)) {
				continue
			}
		}
		
		// Filter by attachments
		if q.HasAttachments && len(email.Attachments) == 0 {
			continue
		}
		
		// Filter by size
		bodySize := len(email.Body)
		if q.MinSize > 0 && bodySize < q.MinSize {
			continue
		}
		if q.MaxSize > 0 && bodySize > q.MaxSize {
			continue
		}
		
		// Filter by domain
		if len(q.Domains) > 0 {
			domain := extractDomain(email.From)
			found := false
			for _, d := range q.Domains {
				if strings.Contains(strings.ToLower(domain), strings.ToLower(strings.TrimSpace(d))) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		
		// Filter by excluded domains
		if len(q.ExcludeDomains) > 0 {
			domain := extractDomain(email.From)
			excluded := false
			for _, d := range q.ExcludeDomains {
				if strings.Contains(strings.ToLower(domain), strings.ToLower(strings.TrimSpace(d))) {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}
		
		// Filter by keywords in subject/body
		if len(q.Keywords) > 0 {
			found := false
			content := strings.ToLower(email.Subject + " " + email.Body)
			for _, keyword := range q.Keywords {
				if strings.Contains(content, strings.ToLower(strings.TrimSpace(keyword))) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		
		// Filter by excluded words
		if len(q.ExcludeWords) > 0 {
			excluded := false
			content := strings.ToLower(email.Subject + " " + email.Body)
			for _, word := range q.ExcludeWords {
				if strings.Contains(content, strings.ToLower(strings.TrimSpace(word))) {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}
		
		filtered = append(filtered, email)
	}
	
	return filtered
}

func extractDomain(from string) string {
	if idx := strings.Index(from, "@"); idx >= 0 {
		endIdx := strings.Index(from[idx:], ">")
		if endIdx > 0 {
			return from[idx+1 : idx+endIdx]
		}
		return from[idx+1:]
	}
	return ""
}

func (q QueryOptions) GetDescription() string {
	var parts []string
	
	if q.FromAddress != "" {
		parts = append(parts, fmt.Sprintf("from:%s", q.FromAddress))
	}
	if q.ToAddress != "" {
		parts = append(parts, fmt.Sprintf("to:%s", q.ToAddress))
	}
	if q.Subject != "" {
		parts = append(parts, fmt.Sprintf("subject:%s", q.Subject))
	}
	if q.UnreadOnly {
		parts = append(parts, "unread")
	}
	if q.TimeRange != "" {
		parts = append(parts, q.TimeRange)
	} else if q.Days > 0 {
		parts = append(parts, fmt.Sprintf("last %d days", q.Days))
	}
	if q.SentOnly {
		parts = append(parts, "sent")
	}
	if q.ReceivedOnly {
		parts = append(parts, "received")
	}
	if q.HasAttachments {
		parts = append(parts, "with attachments")
	}
	if len(q.Domains) > 0 {
		parts = append(parts, fmt.Sprintf("domains:%s", strings.Join(q.Domains, ",")))
	}
	if len(q.Keywords) > 0 {
		parts = append(parts, fmt.Sprintf("keywords:%s", strings.Join(q.Keywords, ",")))
	}
	if q.MinSize > 0 || q.MaxSize > 0 {
		if q.MinSize > 0 && q.MaxSize > 0 {
			parts = append(parts, fmt.Sprintf("size:%s-%s", formatSize(q.MinSize), formatSize(q.MaxSize)))
		} else if q.MinSize > 0 {
			parts = append(parts, fmt.Sprintf("size>%s", formatSize(q.MinSize)))
		} else {
			parts = append(parts, fmt.Sprintf("size<%s", formatSize(q.MaxSize)))
		}
	}
	
	if len(parts) == 0 {
		return "all emails"
	}
	
	return strings.Join(parts, ", ")
}

func formatSize(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%dB", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%dKB", bytes/1024)
	} else if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%dMB", bytes/(1024*1024))
	}
	return fmt.Sprintf("%dGB", bytes/(1024*1024*1024))
}