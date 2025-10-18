package mailos

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// AdvancedSearchOptions extends ReadOptions with advanced search capabilities
type AdvancedSearchOptions struct {
	ReadOptions
	
	// Advanced filters
	Query           string    // Complex query with boolean operators
	FuzzyThreshold  float64   // Fuzzy matching threshold (0.0-1.0)
	MinSize         int64     // Minimum email size in bytes
	MaxSize         int64     // Maximum email size in bytes
	HasAttachments  bool      // Filter emails with attachments
	AttachmentSize  int64     // Minimum attachment size
	DateRange       string    // Flexible date range (e.g., "last week", "2023-01-01 to 2023-12-31")
	
	// Search behavior
	EnableFuzzy     bool      // Enable fuzzy matching
	CaseSensitive   bool      // Case sensitive search
	WholeWords      bool      // Match whole words only
}

// SearchQuery represents a parsed search query with boolean operators
type SearchQuery struct {
	Terms     []SearchTerm
	Operator  string // "AND", "OR"
}

// SearchTerm represents individual search terms
type SearchTerm struct {
	Text     string
	Field    string // "from", "to", "subject", "body", "any"
	Negate   bool   // NOT operator
	Fuzzy    bool   // Enable fuzzy for this term
}

// FuzzyMatch calculates similarity between two strings using Levenshtein distance
func FuzzyMatch(s1, s2 string, threshold float64) bool {
	if threshold <= 0 {
		return strings.Contains(strings.ToLower(s1), strings.ToLower(s2))
	}
	
	distance := levenshteinDistance(strings.ToLower(s1), strings.ToLower(s2))
	maxLen := max(len(s1), len(s2))
	if maxLen == 0 {
		return true
	}
	
	similarity := 1.0 - float64(distance)/float64(maxLen)
	return similarity >= threshold
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}
	
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}
	
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			
			matrix[i][j] = minInt3(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}
	
	return matrix[len(s1)][len(s2)]
}

// ParseSearchQuery parses a complex search query with boolean operators
func ParseSearchQuery(query string) (*SearchQuery, error) {
	if query == "" {
		return &SearchQuery{}, nil
	}
	
	// Simple parser for boolean queries
	// Supports: term1 AND term2, term1 OR term2, NOT term, field:value
	
	query = strings.TrimSpace(query)
	
	// Split by AND/OR operators (case insensitive)
	var terms []SearchTerm
	var operator string = "AND" // default
	
	// Check for OR operator
	if strings.Contains(strings.ToUpper(query), " OR ") {
		operator = "OR"
		parts := regexp.MustCompile(`(?i)\s+or\s+`).Split(query, -1)
		for _, part := range parts {
			term, err := parseSearchTerm(strings.TrimSpace(part))
			if err != nil {
				return nil, err
			}
			terms = append(terms, term)
		}
	} else {
		// Split by AND (default) or spaces
		parts := regexp.MustCompile(`(?i)\s+and\s+|\s+`).Split(query, -1)
		for _, part := range parts {
			if strings.TrimSpace(part) == "" {
				continue
			}
			term, err := parseSearchTerm(strings.TrimSpace(part))
			if err != nil {
				return nil, err
			}
			terms = append(terms, term)
		}
	}
	
	return &SearchQuery{
		Terms:    terms,
		Operator: operator,
	}, nil
}

// parseSearchTerm parses individual search terms with field specifiers and NOT operator
func parseSearchTerm(term string) (SearchTerm, error) {
	searchTerm := SearchTerm{
		Field: "any",
		Fuzzy: true, // Enable fuzzy by default
	}
	
	// Handle NOT operator
	if strings.HasPrefix(strings.ToUpper(term), "NOT ") {
		searchTerm.Negate = true
		term = strings.TrimSpace(term[4:])
	}
	
	// Handle field specifiers (field:value)
	if strings.Contains(term, ":") {
		parts := strings.SplitN(term, ":", 2)
		if len(parts) == 2 {
			field := strings.ToLower(strings.TrimSpace(parts[0]))
			value := strings.TrimSpace(parts[1])
			
			// Remove quotes if present
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			   (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
			
			searchTerm.Field = field
			searchTerm.Text = value
		} else {
			searchTerm.Text = term
		}
	} else {
		// Remove quotes if present
		if (strings.HasPrefix(term, "\"") && strings.HasSuffix(term, "\"")) ||
		   (strings.HasPrefix(term, "'") && strings.HasSuffix(term, "'")) {
			term = term[1 : len(term)-1]
			searchTerm.Fuzzy = false // Exact match for quoted terms
		}
		searchTerm.Text = term
	}
	
	return searchTerm, nil
}

// ParseDateRange parses flexible date range expressions
func ParseDateRange(dateRange string) (since, until time.Time, err error) {
	dateRange = strings.ToLower(strings.TrimSpace(dateRange))
	now := time.Now()
	
	switch dateRange {
	case "today":
		since = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		until = since.Add(24 * time.Hour)
		
	case "yesterday":
		yesterday := now.AddDate(0, 0, -1)
		since = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, now.Location())
		until = since.Add(24 * time.Hour)
		
	case "this week":
		// Start of current week (Monday)
		weekday := int(now.Weekday())
		if weekday == 0 { // Sunday
			weekday = 7
		}
		since = now.AddDate(0, 0, -(weekday-1))
		since = time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, now.Location())
		until = since.AddDate(0, 0, 7)
		
	case "last week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		since = now.AddDate(0, 0, -(weekday-1+7))
		since = time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, now.Location())
		until = since.AddDate(0, 0, 7)
		
	case "this month":
		since = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		until = since.AddDate(0, 1, 0)
		
	case "last month":
		since = time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
		until = since.AddDate(0, 1, 0)
		
	case "this year":
		since = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		until = since.AddDate(1, 0, 0)
		
	default:
		// Try to parse date ranges like "2023-01-01 to 2023-12-31" or "last 30 days"
		if strings.Contains(dateRange, " to ") {
			parts := strings.Split(dateRange, " to ")
			if len(parts) == 2 {
				since, err = parseFlexibleDate(strings.TrimSpace(parts[0]))
				if err != nil {
					return time.Time{}, time.Time{}, fmt.Errorf("invalid start date: %v", err)
				}
				until, err = parseFlexibleDate(strings.TrimSpace(parts[1]))
				if err != nil {
					return time.Time{}, time.Time{}, fmt.Errorf("invalid end date: %v", err)
				}
			}
		} else if strings.HasPrefix(dateRange, "last ") && strings.HasSuffix(dateRange, " days") {
			// Parse "last N days"
			dayStr := strings.TrimSpace(dateRange[5 : len(dateRange)-5])
			days, err := strconv.Atoi(dayStr)
			if err != nil {
				return time.Time{}, time.Time{}, fmt.Errorf("invalid number of days: %v", err)
			}
			since = now.AddDate(0, 0, -days)
			until = now
		} else {
			// Try to parse as single date
			since, err = parseFlexibleDate(dateRange)
			if err != nil {
				return time.Time{}, time.Time{}, fmt.Errorf("invalid date format: %v", err)
			}
			until = since.Add(24 * time.Hour) // End of day
		}
	}
	
	return since, until, nil
}

// parseFlexibleDate parses various date formats
func parseFlexibleDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	
	// Try common date formats
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"01/02/2006",
		"01-02-2006",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"Jan 2, 2006",
		"January 2, 2006",
		"2 Jan 2006",
		"2 January 2006",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// ParseSize parses size expressions like "1MB", "500KB", "2.5GB"
func ParseSize(sizeStr string) (int64, error) {
	if sizeStr == "" {
		return 0, nil
	}
	
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))
	
	// Extract number and unit
	var numStr string
	var unit string
	
	for i, r := range sizeStr {
		if unicode.IsDigit(r) || r == '.' {
			numStr += string(r)
		} else {
			unit = sizeStr[i:]
			break
		}
	}
	
	if numStr == "" {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}
	
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number in size: %v", err)
	}
	
	multiplier := int64(1)
	switch unit {
	case "B", "":
		multiplier = 1
	case "KB", "K":
		multiplier = 1024
	case "MB", "M":
		multiplier = 1024 * 1024
	case "GB", "G":
		multiplier = 1024 * 1024 * 1024
	case "TB", "T":
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unknown size unit: %s", unit)
	}
	
	return int64(num * float64(multiplier)), nil
}

// AdvancedSearchEmails performs advanced email search with all capabilities
func AdvancedSearchEmails(emails []*Email, opts AdvancedSearchOptions) ([]*Email, error) {
	var results []*Email
	
	// Parse search query if provided
	var searchQuery *SearchQuery
	var err error
	if opts.Query != "" {
		searchQuery, err = ParseSearchQuery(opts.Query)
		if err != nil {
			return nil, fmt.Errorf("invalid search query: %v", err)
		}
	}
	
	// Parse date range if provided
	var sinceDateRange, untilDateRange time.Time
	if opts.DateRange != "" {
		sinceDateRange, untilDateRange, err = ParseDateRange(opts.DateRange)
		if err != nil {
			return nil, fmt.Errorf("invalid date range: %v", err)
		}
	}
	
	for _, email := range emails {
		// Apply size filters
		if opts.MinSize > 0 || opts.MaxSize > 0 {
			emailSize := calculateEmailSize(email)
			if opts.MinSize > 0 && emailSize < opts.MinSize {
				continue
			}
			if opts.MaxSize > 0 && emailSize > opts.MaxSize {
				continue
			}
		}
		
		// Apply attachment filters
		if opts.HasAttachments && len(email.Attachments) == 0 {
			continue
		}
		
		if opts.AttachmentSize > 0 {
			hasLargeAttachment := false
			for _, data := range email.AttachmentData {
				if int64(len(data)) >= opts.AttachmentSize {
					hasLargeAttachment = true
					break
				}
			}
			// If no attachment data, estimate from filename (rough estimate)
			if !hasLargeAttachment && len(email.AttachmentData) == 0 && len(email.Attachments) > 0 {
				hasLargeAttachment = true // Assume it meets criteria if we can't check
			}
			if !hasLargeAttachment {
				continue
			}
		}
		
		// Apply date range filter
		if opts.DateRange != "" {
			if email.Date.Before(sinceDateRange) || email.Date.After(untilDateRange) {
				continue
			}
		}
		
		// Apply basic filters from ReadOptions
		if opts.UnreadOnly {
			// Note: This would need to be implemented based on email flags
			// For now, skip this filter
		}
		
		if opts.FromAddress != "" {
			if !matchesField(email.From, opts.FromAddress, opts.EnableFuzzy, opts.FuzzyThreshold, opts.CaseSensitive) {
				continue
			}
		}
		
		if opts.ToAddress != "" {
			found := false
			for _, to := range email.To {
				if matchesField(to, opts.ToAddress, opts.EnableFuzzy, opts.FuzzyThreshold, opts.CaseSensitive) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		
		if opts.Subject != "" {
			if !matchesField(email.Subject, opts.Subject, opts.EnableFuzzy, opts.FuzzyThreshold, opts.CaseSensitive) {
				continue
			}
		}
		
		// Apply complex query search
		if searchQuery != nil && len(searchQuery.Terms) > 0 {
			if !matchesQuery(email, searchQuery, opts) {
				continue
			}
		}
		
		results = append(results, email)
	}
	
	// Apply limit
	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}
	
	return results, nil
}

// calculateEmailSize estimates the size of an email in bytes
func calculateEmailSize(email *Email) int64 {
	size := int64(0)
	
	// Add text content size
	size += int64(len(email.From))
	for _, to := range email.To {
		size += int64(len(to))
	}
	size += int64(len(email.Subject))
	size += int64(len(email.Body))
	size += int64(len(email.BodyHTML))
	
	// Add attachment size
	for _, data := range email.AttachmentData {
		size += int64(len(data))
	}
	
	// If no attachment data but has attachments, estimate
	if len(email.AttachmentData) == 0 && len(email.Attachments) > 0 {
		size += int64(len(email.Attachments)) * 50000 // Rough estimate: 50KB per attachment
	}
	
	return size
}

// matchesField checks if a field matches the search term with fuzzy/exact matching
func matchesField(field, term string, enableFuzzy bool, fuzzyThreshold float64, caseSensitive bool) bool {
	if !caseSensitive {
		field = strings.ToLower(field)
		term = strings.ToLower(term)
	}
	
	if enableFuzzy && fuzzyThreshold > 0 {
		return FuzzyMatch(field, term, fuzzyThreshold)
	}
	
	return strings.Contains(field, term)
}

// matchesQuery checks if an email matches a complex search query
func matchesQuery(email *Email, query *SearchQuery, opts AdvancedSearchOptions) bool {
	matches := make([]bool, len(query.Terms))
	
	for i, term := range query.Terms {
		matches[i] = matchesTerm(email, term, opts)
		
		// Apply NOT operator
		if term.Negate {
			matches[i] = !matches[i]
		}
	}
	
	// Apply boolean operator
	if query.Operator == "OR" {
		for _, match := range matches {
			if match {
				return true
			}
		}
		return false
	} else { // AND (default)
		for _, match := range matches {
			if !match {
				return false
			}
		}
		return true
	}
}

// matchesTerm checks if an email matches a single search term
func matchesTerm(email *Email, term SearchTerm, opts AdvancedSearchOptions) bool {
	fuzzyThreshold := opts.FuzzyThreshold
	if !term.Fuzzy {
		fuzzyThreshold = 0 // Exact match for non-fuzzy terms
	}
	
	switch term.Field {
	case "from":
		return matchesField(email.From, term.Text, opts.EnableFuzzy && term.Fuzzy, fuzzyThreshold, opts.CaseSensitive)
	case "to":
		for _, to := range email.To {
			if matchesField(to, term.Text, opts.EnableFuzzy && term.Fuzzy, fuzzyThreshold, opts.CaseSensitive) {
				return true
			}
		}
		return false
	case "subject":
		return matchesField(email.Subject, term.Text, opts.EnableFuzzy && term.Fuzzy, fuzzyThreshold, opts.CaseSensitive)
	case "body":
		return matchesField(email.Body, term.Text, opts.EnableFuzzy && term.Fuzzy, fuzzyThreshold, opts.CaseSensitive)
	case "any", "":
		// Search in all fields
		return matchesField(email.From, term.Text, opts.EnableFuzzy && term.Fuzzy, fuzzyThreshold, opts.CaseSensitive) ||
			   matchesField(email.Subject, term.Text, opts.EnableFuzzy && term.Fuzzy, fuzzyThreshold, opts.CaseSensitive) ||
			   matchesField(email.Body, term.Text, opts.EnableFuzzy && term.Fuzzy, fuzzyThreshold, opts.CaseSensitive) ||
			   matchesFieldSlice(email.To, term.Text, opts.EnableFuzzy && term.Fuzzy, fuzzyThreshold, opts.CaseSensitive)
	}
	
	return false
}

// matchesFieldSlice checks if any string in a slice matches the search term
func matchesFieldSlice(fields []string, term string, enableFuzzy bool, fuzzyThreshold float64, caseSensitive bool) bool {
	for _, field := range fields {
		if matchesField(field, term, enableFuzzy, fuzzyThreshold, caseSensitive) {
			return true
		}
	}
	return false
}

// Helper functions
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt3(a, b, c int) int {
	return minInt(minInt(a, b), c)
}