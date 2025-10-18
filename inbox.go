package mailos

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

type InboxData struct {
	AccountEmail     string    `json:"account_email"`
	LastFetchTime    time.Time `json:"last_fetch_time"`
	LastEmailDate    time.Time `json:"last_email_date"`
	TotalEmails      int       `json:"total_emails"`
	Emails           []*Email  `json:"emails"`
	LastSyncVersion  int       `json:"last_sync_version"`
}

type GlobalInbox struct {
	Version    int                    `json:"version"`
	Accounts   map[string]*InboxData  `json:"accounts"`
	LastUpdate time.Time              `json:"last_update"`
}

// GetGlobalInboxPath returns the path to the global inbox.json for an account
func GetGlobalInboxPath(accountEmail string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %v", err)
	}
	
	accountDir := filepath.Join(homeDir, ".email", accountEmail)
	if err := os.MkdirAll(accountDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create account directory: %v", err)
	}
	
	return filepath.Join(accountDir, "inbox.json"), nil
}

// LoadGlobalInbox loads the global inbox for an account
func LoadGlobalInbox(accountEmail string) (*InboxData, error) {
	inboxPath, err := GetGlobalInboxPath(accountEmail)
	if err != nil {
		return nil, err
	}
	
	// Check if file exists
	if _, err := os.Stat(inboxPath); os.IsNotExist(err) {
		// Create new inbox data
		return &InboxData{
			AccountEmail:    accountEmail,
			LastFetchTime:   time.Time{},
			LastEmailDate:   time.Time{},
			TotalEmails:     0,
			Emails:          []*Email{},
			LastSyncVersion: 1,
		}, nil
	}
	
	data, err := os.ReadFile(inboxPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read inbox file: %v", err)
	}
	
	var inboxData InboxData
	if err := json.Unmarshal(data, &inboxData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal inbox data: %v", err)
	}
	
	return &inboxData, nil
}

// SaveGlobalInbox saves the global inbox for an account
func SaveGlobalInbox(accountEmail string, inboxData *InboxData) error {
	inboxPath, err := GetGlobalInboxPath(accountEmail)
	if err != nil {
		return err
	}
	
	inboxData.LastFetchTime = time.Now()
	inboxData.TotalEmails = len(inboxData.Emails)
	
	// Update last email date if we have emails
	if len(inboxData.Emails) > 0 {
		latestDate := inboxData.Emails[0].Date
		for _, email := range inboxData.Emails {
			if email.Date.After(latestDate) {
				latestDate = email.Date
			}
		}
		inboxData.LastEmailDate = latestDate
	}
	
	data, err := json.MarshalIndent(inboxData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal inbox data: %v", err)
	}
	
	return os.WriteFile(inboxPath, data, 0600)
}

// FetchEmailsIncremental fetches emails incrementally based on last fetch time
func FetchEmailsIncremental(config *Config, limit int) error {
	if config.Email == "" {
		return fmt.Errorf("no email account configured")
	}
	
	// Load existing inbox data
	inboxData, err := LoadGlobalInbox(config.Email)
	if err != nil {
		return fmt.Errorf("failed to load inbox data: %v", err)
	}
	
	// Connect to IMAP server
	c, err := connectToIMAPServer(config)
	if err != nil {
		return err
	}
	defer c.Logout()
	
	// Select INBOX
	_, err = c.Select("INBOX", false)
	if err != nil {
		return fmt.Errorf("failed to select INBOX: %v", err)
	}
	
	// Build search criteria for incremental fetch
	criteria := imap.NewSearchCriteria()
	
	// For a complete history sync, don't set any date criteria
	// This will fetch ALL emails from the beginning of the account
	fmt.Printf("Fetching ALL emails from account history...\n")
	
	// Search for messages
	ids, err := c.Search(criteria)
	if err != nil {
		return fmt.Errorf("failed to search messages: %v", err)
	}
	
	if len(ids) == 0 {
		fmt.Printf("No new emails found for %s\n", config.Email)
		// Still update last fetch time
		return SaveGlobalInbox(config.Email, inboxData)
	}
	
	fmt.Printf("Found %d new emails for %s\n", len(ids), config.Email)
	
	// Limit results if specified
	if limit > 0 && len(ids) > limit {
		// Get the most recent messages
		ids = ids[len(ids)-limit:]
	}
	
	// Create sequence set
	seqSet := new(imap.SeqSet)
	for _, id := range ids {
		seqSet.AddNum(id)
	}
	
	// Fetch messages
	messages := make(chan *imap.Message, len(ids))
	section := &imap.BodySectionName{}
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqSet, []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, section.FetchItem()}, messages)
	}()
	
	var newEmails []*Email
	for msg := range messages {
		email, err := parseMessageWithOptions(msg, section, false)
		if err != nil {
			continue
		}
		newEmails = append(newEmails, email)
	}
	
	if err := <-done; err != nil {
		return fmt.Errorf("failed to fetch messages: %v", err)
	}
	
	// Add new emails to existing inbox data
	inboxData.Emails = append(inboxData.Emails, newEmails...)
	
	// Sort emails by date (newest first)
	sort.Slice(inboxData.Emails, func(i, j int) bool {
		return inboxData.Emails[i].Date.After(inboxData.Emails[j].Date)
	})
	
	// Remove duplicates based on MessageID
	inboxData.Emails = removeDuplicateEmails(inboxData.Emails)
	
	// Save updated inbox data
	if err := SaveGlobalInbox(config.Email, inboxData); err != nil {
		return fmt.Errorf("failed to save inbox data: %v", err)
	}
	
	fmt.Printf("✓ Fetched and saved %d new emails for %s\n", len(newEmails), config.Email)
	fmt.Printf("✓ Total emails in inbox: %d\n", len(inboxData.Emails))
	
	return nil
}

// GetEmailsFromInbox returns emails from the global inbox with filtering options
func GetEmailsFromInbox(accountEmail string, opts ReadOptions) ([]*Email, error) {
	inboxData, err := LoadGlobalInbox(accountEmail)
	if err != nil {
		return nil, err
	}
	
	var filteredEmails []*Email
	
	// Apply filters
	for _, email := range inboxData.Emails {
		// Apply filters similar to readFromLocalStorage
		if opts.FromAddress != "" && !containsIgnoreCase(email.From, opts.FromAddress) {
			continue
		}
		if opts.Subject != "" && !containsIgnoreCase(email.Subject, opts.Subject) {
			continue
		}
		if !opts.Since.IsZero() && email.Date.Before(opts.Since) {
			continue
		}
		
		filteredEmails = append(filteredEmails, email)
	}
	
	// Apply limit
	if opts.Limit > 0 && len(filteredEmails) > opts.Limit {
		filteredEmails = filteredEmails[:opts.Limit]
	}
	
	return filteredEmails, nil
}

// SyncAllAccounts fetches emails for all configured accounts
func SyncAllAccounts(limit int) error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}
	
	accounts := GetAllAccounts(config)
	if len(accounts) == 0 {
		return fmt.Errorf("no accounts configured")
	}
	
	fmt.Printf("Syncing emails for %d accounts...\n", len(accounts))
	
	for _, account := range accounts {
		fmt.Printf("\n--- Syncing %s (%s) ---\n", account.Email, account.Provider)
		
		// Create config for this account
		accountConfig := &Config{
			Provider: account.Provider,
			Email:    account.Email,
			Password: account.Password,
		}
		
		if err := FetchEmailsIncremental(accountConfig, limit); err != nil {
			fmt.Printf("Error syncing %s: %v\n", account.Email, err)
			continue
		}
	}
	
	fmt.Printf("\n✓ Finished syncing all accounts\n")
	return nil
}

// connectToIMAPServer establishes connection to IMAP server
func connectToIMAPServer(config *Config) (*client.Client, error) {
	imapHost, imapPort, err := config.GetIMAPSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to get IMAP settings: %v", err)
	}
	
	var c *client.Client
	if imapPort == 993 {
		c, err = client.DialTLS(fmt.Sprintf("%s:%d", imapHost, imapPort), nil)
	} else {
		c, err = client.Dial(fmt.Sprintf("%s:%d", imapHost, imapPort))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to IMAP server: %v", err)
	}
	
	if err := c.Login(config.Email, config.Password); err != nil {
		c.Logout()
		return nil, fmt.Errorf("failed to login: %v", err)
	}
	
	return c, nil
}

// removeDuplicateEmails removes duplicate emails based on MessageID
func removeDuplicateEmails(emails []*Email) []*Email {
	seen := make(map[string]bool)
	var unique []*Email
	
	for _, email := range emails {
		key := email.MessageID
		if key == "" {
			// Fallback: use subject + date + from as key
			key = fmt.Sprintf("%s_%d_%s", email.Subject, email.Date.Unix(), email.From)
		}
		
		if !seen[key] {
			seen[key] = true
			unique = append(unique, email)
		}
	}
	
	return unique
}

// containsIgnoreCase checks if haystack contains needle (case insensitive)
func containsIgnoreCase(haystack, needle string) bool {
	return len(needle) == 0 || 
		   len(haystack) >= len(needle) && 
		   findIgnoreCase(haystack, needle) >= 0
}

// findIgnoreCase finds needle in haystack (case insensitive), returns index or -1
func findIgnoreCase(haystack, needle string) int {
	if len(needle) == 0 {
		return 0
	}
	if len(needle) > len(haystack) {
		return -1
	}
	
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if equalFoldSubstring(haystack[i:i+len(needle)], needle) {
			return i
		}
	}
	return -1
}

// equalFoldSubstring compares two strings case-insensitively
func equalFoldSubstring(s1, s2 string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		c1, c2 := s1[i], s2[i]
		if c1 != c2 {
			// Convert to lowercase for comparison
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 'a' - 'A'
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 'a' - 'A'
			}
			if c1 != c2 {
				return false
			}
		}
	}
	return true
}

// GetInboxStats returns statistics about the inbox
func GetInboxStats(accountEmail string) (*InboxData, error) {
	return LoadGlobalInbox(accountEmail)
}

// ListAccountInboxes lists all accounts that have inbox data
func ListAccountInboxes() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}
	
	emailDir := filepath.Join(homeDir, ".email")
	entries, err := os.ReadDir(emailDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read .email directory: %v", err)
	}
	
	var accounts []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if this directory has an inbox.json
			inboxPath := filepath.Join(emailDir, entry.Name(), "inbox.json")
			if _, err := os.Stat(inboxPath); err == nil {
				accounts = append(accounts, entry.Name())
			}
		}
	}
	
	return accounts, nil
}

// CleanupOldEmails removes emails older than the specified duration
func CleanupOldEmails(accountEmail string, olderThan time.Duration) error {
	inboxData, err := LoadGlobalInbox(accountEmail)
	if err != nil {
		return err
	}
	
	cutoffDate := time.Now().Add(-olderThan)
	var keptEmails []*Email
	removedCount := 0
	
	for _, email := range inboxData.Emails {
		if email.Date.After(cutoffDate) {
			keptEmails = append(keptEmails, email)
		} else {
			removedCount++
		}
	}
	
	if removedCount > 0 {
		inboxData.Emails = keptEmails
		if err := SaveGlobalInbox(accountEmail, inboxData); err != nil {
			return err
		}
		fmt.Printf("Removed %d emails older than %v for %s\n", removedCount, olderThan, accountEmail)
	}
	
	return nil
}

// SyncEmailsForAccount is a convenience function to sync emails for a specific account
func SyncEmailsForAccount(accountEmail string, limit int) error {
	config, err := LoadAccountConfig(accountEmail)
	if err != nil {
		return fmt.Errorf("failed to load config for %s: %v", accountEmail, err)
	}
	
	return FetchEmailsIncremental(config, limit)
}

// UpdateLastSyncTime updates the last sync timestamp in the config
func UpdateLastSyncTime() error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	config.LastSyncTime = time.Now().Format(time.RFC3339)

	return SaveConfig(config)
}

// ShouldAutoSync checks if auto-sync should run based on last sync time
func ShouldAutoSync() bool {
	config, err := LoadConfig()
	if err != nil {
		return false
	}

	// Check if auto-sync is enabled (default to true if not set)
	if config.AutoSync == false && config.LastSyncTime != "" {
		// If AutoSync is explicitly set to false, don't auto-sync
		return false
	}

	// If never synced, should sync
	if config.LastSyncTime == "" {
		return true
	}

	// Parse last sync time
	lastSync, err := time.Parse(time.RFC3339, config.LastSyncTime)
	if err != nil {
		return true // If can't parse, sync to be safe
	}

	// Check if more than 24 hours have passed
	return time.Since(lastSync) > 24*time.Hour
}

// RunAutoSyncIfNeeded runs sync automatically if needed
func RunAutoSyncIfNeeded() error {
	if !ShouldAutoSync() {
		return nil
	}

	fmt.Println("Auto-syncing emails (last sync was more than 24 hours ago)...")
	
	// Use new global inbox system
	config, err := LoadConfig()
	if err == nil && config.Email != "" {
		if err := FetchEmailsIncremental(config, 50); err != nil {
			fmt.Printf("Warning: Failed to auto-sync to global inbox: %v\n", err)
			return err
		} else {
			// Update last sync time
			if err := UpdateLastSyncTime(); err != nil {
				fmt.Printf("Warning: failed to update last sync time: %v\n", err)
			}
			return nil
		}
	}
	
	return fmt.Errorf("no email account configured for auto-sync")
}

// sanitizeFilename removes or replaces invalid filename characters
func sanitizeFilename(s string) string {
	// Remove or replace invalid filename characters
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"<", "-",
		">", "-",
		"|", "-",
		"\n", " ",
		"\r", " ",
	)
	s = replacer.Replace(s)
	
	// Trim spaces and limit length
	s = strings.TrimSpace(s)
	if len(s) > 100 {
		s = s[:100]
	}
	
	return s
}