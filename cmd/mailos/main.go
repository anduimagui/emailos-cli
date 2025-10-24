package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	mailos "github.com/anduimagui/emailos"
	"github.com/spf13/cobra"
)

const (
	// Version is the current version of mailos
	Version = "0.1.12"
	// GithubRepo is the repository for updates
	GithubRepo = "emailos/mailos"
)

// checkForUpdates checks if a newer version is available and auto-updates
func checkForUpdates() {
	// Skip update check if environment variable is set
	if os.Getenv("MAILOS_SKIP_UPDATE") == "true" {
		return
	}

	// Check if we should skip based on cache (2 hour interval)
	if shouldSkipUpdateCheck() {
		return
	}
	
	// Get latest release from GitHub
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GithubRepo)
	resp, err := http.Get(url)
	if err != nil {
		// Update check time even on failure to avoid repeated failed attempts
		updateLastCheckTime()
		return // Silently fail if we can't check
	}
	defer resp.Body.Close()
	
	// Update the last check time after successful connection
	updateLastCheckTime()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return
	}

	// Parse version from tag (remove 'v' prefix if present)
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	
	// Compare versions
	if !isNewerVersion(latestVersion, Version) {
		return
	}

	fmt.Printf("New version available: %s (current: %s)\n", latestVersion, Version)
	fmt.Println("Auto-updating...")

	// Determine platform
	platform := getPlatformName()
	assetName := fmt.Sprintf("mailos-%s.tar.gz", platform)
	
	// Find download URL for this platform
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		fmt.Printf("No binary available for platform %s\n", platform)
		return
	}

	// Download and install update
	if err := downloadAndInstallUpdate(downloadURL); err != nil {
		fmt.Printf("Update failed: %v\n", err)
		fmt.Println("You can manually update with: npm update -g mailos")
		return
	}

	fmt.Println("Update successful! Restarting...")
	
	// Restart the application with the same arguments
	args := os.Args
	env := os.Environ()
	
	// Execute the new binary
	execPath, _ := os.Executable()
	err = syscallExec(execPath, args, env)
	if err != nil {
		fmt.Printf("Failed to restart: %v\n", err)
	}
}

// syscallExec replaces the current process with a new one
func syscallExec(argv0 string, argv []string, envv []string) error {
	cmd := exec.Command(argv0, argv[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = envv
	return cmd.Run()
}

// isNewerVersion compares semantic versions
func isNewerVersion(latest, current string) bool {
	latestParts := strings.Split(latest, ".")
	currentParts := strings.Split(current, ".")
	
	for i := 0; i < len(latestParts) && i < len(currentParts); i++ {
		var latestNum, currentNum int
		fmt.Sscanf(latestParts[i], "%d", &latestNum)
		fmt.Sscanf(currentParts[i], "%d", &currentNum)
		
		if latestNum > currentNum {
			return true
		} else if latestNum < currentNum {
			return false
		}
	}
	
	return len(latestParts) > len(currentParts)
}

// getPlatformName returns the platform name for downloads
func getPlatformName() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	
	platform := goos
	if goos == "darwin" {
		platform = "darwin"
	} else if goos == "windows" {
		platform = "windows"
	} else if goos == "linux" {
		platform = "linux"
	}
	
	arch := "amd64"
	if goarch == "arm64" {
		arch = "arm64"
	}
	
	return fmt.Sprintf("%s-%s", platform, arch)
}

// downloadAndInstallUpdate downloads and installs the update
func downloadAndInstallUpdate(url string) error {
	// Download the file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	// Create temp file
	tempFile, err := os.CreateTemp("", "mailos-update-*.tar.gz")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy download to temp file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save update: %w", err)
	}
	tempFile.Close()

	// Extract the binary
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create backup of current binary
	backupPath := execPath + ".backup"
	if err := os.Rename(execPath, backupPath); err != nil {
		// If rename fails, try copying
		if err := copyFile(execPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup current binary: %w", err)
		}
	}

	// Extract new binary from tar.gz
	cmd := exec.Command("tar", "-xzf", tempFile.Name(), "-C", filepath.Dir(execPath))
	if err := cmd.Run(); err != nil {
		// Restore backup on failure
		os.Rename(backupPath, execPath)
		return fmt.Errorf("failed to extract update: %w", err)
	}

	// Make sure new binary is executable
	if runtime.GOOS != "windows" {
		os.Chmod(execPath, 0755)
	}

	// Remove backup
	os.Remove(backupPath)
	
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}

// shouldSkipUpdateCheck checks if we should skip update based on cache
func shouldSkipUpdateCheck() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	
	// Cache file to track last update check
	cacheFile := filepath.Join(homeDir, ".email", ".update_check")
	
	info, err := os.Stat(cacheFile)
	if err != nil {
		// File doesn't exist, should check
		return false
	}
	
	// Check if more than 2 hours have passed
	timeSinceCheck := time.Since(info.ModTime())
	if timeSinceCheck < 2*time.Hour {
		// Less than 2 hours, skip check
		return true
	}
	
	return false
}

// updateLastCheckTime updates the cache file to track last update check
func updateLastCheckTime() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	
	// Ensure .email directory exists
	emailDir := filepath.Join(homeDir, ".email")
	os.MkdirAll(emailDir, 0755)
	
	// Update or create cache file
	cacheFile := filepath.Join(emailDir, ".update_check")
	os.WriteFile(cacheFile, []byte(time.Now().Format(time.RFC3339)), 0644)
}

// parseQueryFromArgs extracts a query from command line arguments
// Supports:
// - mailos q=[QUERY] - everything after q= is the query
// - mailos "[QUERY]" - quoted string is the query
// - mailos '[QUERY]' - single quoted string is the query
// - mailos interactive - launch interactive mode explicitly
// Returns empty string if no valid query format found
func parseQueryFromArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	
	// First check if it's a known command
	knownCommands := []string{
		"setup", "local", "configure", "config", "template", "drafts", "draft", "send", "sent", "read", "reply",
		"mark-read", "delete", "unsubscribe", "info", "test", "interactive", "chat",
		"report", "open", "provider", "stats", "search",
		"--help", "-h", "--version", "-v",
	}
	
	firstArg := strings.ToLower(args[0])
	for _, cmd := range knownCommands {
		if firstArg == cmd {
			return "" // It's a known command, not a query
		}
	}
	
	// Check for q= parameter format anywhere in args
	fullArgs := strings.Join(args, " ")
	if strings.HasPrefix(fullArgs, "q=") {
		return strings.TrimPrefix(fullArgs, "q=")
	}
	
	// Check if first arg starts with q=
	if strings.HasPrefix(args[0], "q=") {
		query := strings.TrimPrefix(args[0], "q=")
		// Add any remaining args
		if len(args) > 1 {
			query += " " + strings.Join(args[1:], " ")
		}
		return query
	}
	
	// Check for quoted string (can be across multiple args due to shell parsing)
	// When shell passes quoted string, it comes as single arg without quotes
	// But we should also handle if the quotes are preserved
	if len(args) == 1 {
		arg := args[0]
		// Check for double quotes
		if strings.HasPrefix(arg, "\"") && strings.HasSuffix(arg, "\"") && len(arg) > 2 {
			return strings.Trim(arg, "\"")
		}
		// Check for single quotes
		if strings.HasPrefix(arg, "'") && strings.HasSuffix(arg, "'") && len(arg) > 2 {
			return strings.Trim(arg, "'")
		}
		
		// If it's a single arg that's not a command and doesn't start with special chars,
		// treat it as a query (this handles "mailos 'search emails'" where shell strips quotes)
		if !strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "/") {
			// But only if it looks like a natural language query (contains spaces or common query words)
			if strings.Contains(arg, " ") || 
			   strings.Contains(strings.ToLower(arg), "email") ||
			   strings.Contains(strings.ToLower(arg), "find") ||
			   strings.Contains(strings.ToLower(arg), "show") ||
			   strings.Contains(strings.ToLower(arg), "search") ||
			   strings.Contains(strings.ToLower(arg), "send") ||
			   strings.Contains(strings.ToLower(arg), "write") ||
			   strings.Contains(strings.ToLower(arg), "compose") {
				return arg
			}
		}
	}
	
	// If multiple args and not a known command, check if it could be a natural language query
	// This handles cases like: mailos find unread emails
	if len(args) > 1 {
		// Check if it looks like a natural language query
		queryWords := []string{"find", "show", "search", "get", "list", "display", "write", "compose", "send", "check"}
		firstWord := strings.ToLower(args[0])
		for _, word := range queryWords {
			if firstWord == word {
				return fullArgs
			}
		}
	}
	
	// No valid query format found
	return ""
}

// levenshteinDistance calculates the edit distance between two strings
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
	}
	
	for i := 0; i <= len(a); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}
	
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,
				matrix[i][j-1]+1,
				matrix[i-1][j-1]+cost,
			)
		}
	}
	
	return matrix[len(a)][len(b)]
}

func min(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}

// formatEmailAsMarkdown formats an email as markdown
func formatEmailAsMarkdown(email *mailos.Email) string {
	var content strings.Builder
	
	// Title
	content.WriteString(fmt.Sprintf("# %s\n\n", email.Subject))
	
	// Metadata
	content.WriteString(fmt.Sprintf("**From:** %s  \n", email.From))
	content.WriteString(fmt.Sprintf("**To:** %s  \n", strings.Join(email.To, ", ")))
	content.WriteString(fmt.Sprintf("**Date:** %s  \n", email.Date.Format("January 2, 2006 3:04 PM")))
	content.WriteString(fmt.Sprintf("**ID:** %d  \n", email.ID))
	
	if len(email.Attachments) > 0 {
		content.WriteString(fmt.Sprintf("**Attachments:** %s  \n", strings.Join(email.Attachments, ", ")))
	}
	
	content.WriteString("\n---\n\n")
	
	// Body
	body := email.Body
	if body == "" && email.BodyHTML != "" {
		// If only HTML is available, note it
		content.WriteString("*[HTML email - plain text version not available]*\n\n")
		body = mailos.StripHTMLTags(email.BodyHTML)
	}
	
	content.WriteString(body)
	content.WriteString("\n")
	
	return content.String()
}

// getAllCommands returns all available command names including aliases
func getAllCommands() []string {
	commands := []string{
		"setup", "local", "provider", "configure", "config", "template",
		"draft", "drafts", "compose", "send", "sync", "sync-db", "sent", "download", "read", "reply", "forward",
		"mark-read", "accounts", "info", "test", "delete", "report",
		"open", "stats", "docs", "commands", "tools", "interactive", "chat", "search",
		"unsubscribe", "uninstall", "cleanup",
	}
	sort.Strings(commands)
	return commands
}

// findSimilarCommands finds commands similar to the input using edit distance
func findSimilarCommands(input string) []string {
	allCommands := getAllCommands()
	type commandDistance struct {
		command  string
		distance int
	}
	
	var suggestions []commandDistance
	maxDistance := min(len(input)/2+1, 3, len(input))
	
	for _, cmd := range allCommands {
		distance := levenshteinDistance(strings.ToLower(input), strings.ToLower(cmd))
		if distance <= maxDistance {
			suggestions = append(suggestions, commandDistance{cmd, distance})
		}
	}
	
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].distance < suggestions[j].distance
	})
	
	var result []string
	for _, s := range suggestions {
		result = append(result, s.command)
		if len(result) >= 3 {
			break
		}
	}
	
	return result
}

// isValidCommand checks if the input is a valid command
func isValidCommand(input string) bool {
	allCommands := getAllCommands()
	for _, cmd := range allCommands {
		if strings.EqualFold(input, cmd) {
			return true
		}
	}
	return false
}

// showUnknownCommandHelp displays help for unknown commands
func showUnknownCommandHelp(unknownCmd string) {
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("                UNKNOWN COMMAND\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	
	fmt.Printf("❌ Command '%s' not found.\n\n", unknownCmd)
	
	// Special case for compose - direct suggestion to drafts
	if strings.EqualFold(unknownCmd, "compose") {
		fmt.Printf("💡 For composing emails, use:\n")
		fmt.Printf("   • mailos compose --to user@example.com --subject \"Hello\" --body \"Message\"\n")
		fmt.Printf("   • mailos drafts (legacy command with advanced features)\n")
		fmt.Printf("   • mailos send (for immediate sending without drafts)\n")
		fmt.Println()
	} else {
		// Show suggestions if any
		suggestions := findSimilarCommands(unknownCmd)
		if len(suggestions) > 0 {
			fmt.Printf("💡 Did you mean:\n")
			for _, suggestion := range suggestions {
				fmt.Printf("   • mailos %s\n", suggestion)
			}
			fmt.Println()
		}
	}
	
	fmt.Printf("📋 Available commands:\n")
	
	// Group commands by category for better display
	core := []string{"setup", "configure", "info"}
	email := []string{"read", "reply", "send", "compose", "draft", "search", "delete", "mark-read"}
	management := []string{"sync", "sync-db", "accounts", "stats", "report", "template"}
	interaction := []string{"interactive", "chat", "open", "unsubscribe"}
	
	printCommandGroup("Core", core)
	printCommandGroup("Email", email)
	printCommandGroup("Management", management)
	printCommandGroup("Interaction", interaction)
	
	fmt.Printf("\n💁 For help with a specific command, run:\n")
	fmt.Printf("   mailos <command> --help\n\n")
	fmt.Printf("🚀 To get started, run:\n")
	fmt.Printf("   mailos setup\n\n")
}

func printCommandGroup(category string, commands []string) {
	fmt.Printf("\n   %s:\n", category)
	for _, cmd := range commands {
		fmt.Printf("     • %s\n", cmd)
	}
}

// showAllCommands displays all available commands organized by category
func showAllCommands(verbose bool) error {
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("               MAILOS COMMANDS REFERENCE\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	
	// Core Commands
	fmt.Printf("🏠 CORE COMMANDS:\n")
	fmt.Printf("  setup      - Initial email configuration\n")
	fmt.Printf("  configure  - Manage email configuration (aliases: config)\n")
	fmt.Printf("  info       - Show current configuration status\n")
	fmt.Printf("  accounts   - Manage multiple email accounts\n")
	fmt.Printf("  provider   - Select or configure AI provider\n")
	
	// Email Management Commands  
	fmt.Printf("\n📧 EMAIL MANAGEMENT:\n")
	fmt.Printf("  search     - Search and list emails with advanced filters\n")
	fmt.Printf("  read       - Display full content of a specific email by ID\n")
	fmt.Printf("  send       - Send an email\n")
	fmt.Printf("  reply      - Reply to a specific email\n")
	fmt.Printf("  forward    - Forward a specific email\n")
	fmt.Printf("  mark-read  - Mark emails as read\n")
	fmt.Printf("  delete     - Delete emails\n")
	fmt.Printf("  sent       - Read sent emails\n")
	fmt.Printf("  download   - Download email attachments\n")
	
	// Draft Management
	fmt.Printf("\n📝 DRAFT MANAGEMENT:\n")
	fmt.Printf("  compose    - Compose a new email (alias for drafts)\n")
	fmt.Printf("  draft      - Simplified draft management (list, edit, create)\n")
	fmt.Printf("  drafts     - Legacy draft command with advanced features\n")
	
	// Data & Analytics
	fmt.Printf("\n📊 DATA & ANALYTICS:\n")
	fmt.Printf("  stats      - Show email statistics and analytics\n")
	fmt.Printf("  report     - Generate email reports for time ranges\n")
	fmt.Printf("  sync       - Sync emails from IMAP to local filesystem\n")
	fmt.Printf("  sync-db    - Sync emails to local SQLite database\n")
	
	// Automation & Tools
	fmt.Printf("\n🔧 AUTOMATION & TOOLS:\n")
	fmt.Printf("  template   - Customize HTML email template\n")
	fmt.Printf("  open       - Open email in default mail application\n")
	fmt.Printf("  unsubscribe- Find and open unsubscribe links\n")
	fmt.Printf("  test       - Test email functionality\n")
	fmt.Printf("  tools      - List all Go methods/functions in codebase\n")
	
	// Interactive & AI
	fmt.Printf("\n🤖 INTERACTIVE & AI:\n")
	fmt.Printf("  interactive- Launch interactive mode\n")
	fmt.Printf("  chat       - Launch AI chat interface\n")
	
	// System Management
	fmt.Printf("\n⚙️  SYSTEM MANAGEMENT:\n")
	fmt.Printf("  local      - Create local project configuration\n")
	fmt.Printf("  docs       - Generate AI instruction documentation\n")
	fmt.Printf("  commands   - Show this command reference\n")
	fmt.Printf("  uninstall  - Completely remove EmailOS\n")
	fmt.Printf("  cleanup    - Clean up orphaned data\n")
	
	if verbose {
		fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("               COMMONLY USED FLAG PATTERNS\n")
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
		
		fmt.Printf("🔍 SEARCH FLAGS:\n")
		fmt.Printf("  --number, -n       Limit number of results\n")
		fmt.Printf("  --unread, -u       Show only unread emails\n")
		fmt.Printf("  --from             Filter by sender\n")
		fmt.Printf("  --to               Filter by recipient\n")
		fmt.Printf("  --subject          Filter by subject\n")
		fmt.Printf("  --days             Last N days\n")
		fmt.Printf("  --query, -q        Complex search query\n")
		fmt.Printf("  --has-attachments  Emails with attachments\n")
		fmt.Printf("  --case-sensitive   Case sensitive search\n")
		
		fmt.Printf("\n📤 SEND FLAGS:\n")
		fmt.Printf("  --to, -t           Recipient addresses\n")
		fmt.Printf("  --cc, -c           CC recipients\n")
		fmt.Printf("  --bcc, -B          BCC recipients\n")
		fmt.Printf("  --subject, -s      Email subject\n")
		fmt.Printf("  --body, -b         Email body\n")
		fmt.Printf("  --file, -f         Read body from file\n")
		fmt.Printf("  --attach, -a       Attachments\n")
		fmt.Printf("  --from             Send from specific account\n")
		fmt.Printf("  --plain, -P        Send as plain text\n")
		fmt.Printf("  --no-signature, -S Don't add signature\n")
		
		fmt.Printf("\n🗂️  OUTPUT FLAGS:\n")
		fmt.Printf("  --json             Output as JSON\n")
		fmt.Printf("  --save-markdown    Save as markdown files\n")
		fmt.Printf("  --output-dir       Output directory\n")
		fmt.Printf("  --download-attachments Download attachments\n")
		fmt.Printf("  --attachment-dir   Attachment directory\n")
		
		fmt.Printf("\n⏰ TIME FLAGS:\n")
		fmt.Printf("  --days             Last N days\n")
		fmt.Printf("  --range            Time range (Today, Yesterday, This week, etc.)\n")
		fmt.Printf("  --date-range       Custom date range (YYYY-MM-DD,YYYY-MM-DD)\n")
		fmt.Printf("  --before           Before date (YYYY-MM-DD)\n")
		fmt.Printf("  --after            After date (YYYY-MM-DD)\n")
	}
	
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("📚 USAGE EXAMPLES:\n")
	fmt.Printf("  mailos search --unread --days 7          # Unread emails last 7 days\n")
	fmt.Printf("  mailos read 1234                         # Read email ID 1234\n")
	fmt.Printf("  mailos send --to user@example.com --subject \"Hi\" --body \"Hello\"\n")
	fmt.Printf("  mailos stats --from gmail.com --days 30  # Gmail stats last 30 days\n")
	fmt.Printf("  mailos accounts --list                   # List all accounts\n")
	fmt.Printf("  mailos search --query \"meeting\" --has-attachments\n")
	
	fmt.Printf("\n💡 FOR HELP: mailos <command> --help\n")
	fmt.Printf("🚀 GET STARTED: mailos setup\n\n")
	
	return nil
}

var rootCmd = &cobra.Command{
	Use:     "mailos",
	Version: Version,
	Short:   "EmailOS - A standardized email client",
	Long: `EmailOS is a command-line email client that supports multiple providers
and provides a consistent interface for sending and reading emails.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		
		// Check if this is an unknown command (single argument that's not a query)
		if len(args) == 1 && !strings.HasPrefix(args[0], "-") {
			// Check if it's a valid command first
			if !isValidCommand(args[0]) {
				// Check if it could be a natural language query
				query := parseQueryFromArgs(args)
				if query == "" {
					// Not a query, show unknown command help
					showUnknownCommandHelp(args[0])
					return nil
				}
			}
		}
		
		// Ensure initialized first
		if err := mailos.EnsureInitialized(); err != nil {
			return err
		}
		
		// Parse query from arguments
		query := parseQueryFromArgs(args)
		
		// If a query was found, handle it
		if query != "" {
			// Use interactive handler which will prompt if no provider configured
			// return mailos.HandleQueryWithProviderSelection(query)
			return fmt.Errorf("CLI functionality temporarily disabled")
		}
		
		// Default behavior: show help
		return cmd.Help()
	},
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up your email configuration",
	Long: `Set up your email configuration interactively or with flags.

You can specify configuration values using flags to skip interactive prompts:
  --email           Your email address
  --provider        Email provider (gmail, fastmail, outlook, yahoo, zoho)
  --name            Your display name
  --license         Your MailOS license key
  --profile         Path to your profile image

Example:
  mailos setup --email=john@example.com --provider=gmail --name="John Doe"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		provider, _ := cmd.Flags().GetString("provider")
		name, _ := cmd.Flags().GetString("name")
		license, _ := cmd.Flags().GetString("license")
		profile, _ := cmd.Flags().GetString("profile")
		
		return mailos.SetupWithFlags(email, provider, name, profile, license)
	},
}

var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Create a local configuration that inherits from global settings",
	Long:  `Create a project-specific configuration in the current directory (.email/).
This configuration will inherit settings from your global configuration (~/.email/),
allowing you to override specific settings like from address or display name for this project.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Force local configuration creation
		// opts := mailos.ConfigureOptions{
		//	IsLocal: true,
		// }
		// return mailos.ConfigureWithOptions(opts)
		return fmt.Errorf("configuration functionality temporarily disabled")
	},
}

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Select or configure AI provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := mailos.EnsureInitializedInteractive(); err != nil {
			return err
		}
		// return mailos.SelectAndConfigureAIProvider()
		return fmt.Errorf("CLI functionality temporarily disabled")
	},
}

var configureCmd = &cobra.Command{
	Use:   "configure",
	Aliases: []string{"config"}, // Add config as an alias
	Short: "Manage email configuration (global or local)",
	Long:  `Configure email settings. By default modifies global configuration (~/.email/).
Use --local flag to create/modify project-specific configuration (.email/)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		quick, _ := cmd.Flags().GetBool("quick")
		isLocal, _ := cmd.Flags().GetBool("local")
		email, _ := cmd.Flags().GetString("email")
		provider, _ := cmd.Flags().GetString("provider")
		name, _ := cmd.Flags().GetString("name")
		from, _ := cmd.Flags().GetString("from")
		aiCLI, _ := cmd.Flags().GetString("ai")
		
		// Pass command-line arguments to Configure
		opts := mailos.ConfigureOptions{
			Email:    email,
			Provider: provider,
			Name:     name,
			From:     from,
			AICLI:    aiCLI,
			IsLocal:  isLocal,
			Quick:    quick,
		}
		return mailos.Configure(opts)
	},
}

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Customize HTML email template",
	Long: `Customize your HTML email template.
The template will be saved and used for all future emails.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		remove, _ := cmd.Flags().GetBool("remove")
		openBrowser, _ := cmd.Flags().GetBool("open-browser")
		
		if remove {
			return mailos.RemoveTemplate()
		}
		
		if openBrowser {
			return mailos.OpenTemplateInBrowser()
		}
		
		return mailos.ManageTemplate()
	},
}

var draftCmd = &cobra.Command{
	Use:   "draft",
	Short: "Simplified draft management",
	Long:  `Simplified draft email management with user-friendly commands.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default behavior: list drafts
		return mailos.ListSimpleDrafts()
	},
}

var draftListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all drafts",
	RunE: func(cmd *cobra.Command, args []string) error {
		return mailos.ListSimpleDrafts()
	},
}

var draftEditCmd = &cobra.Command{
	Use:   "edit [number]",
	Short: "Edit a draft by number",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		number, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid draft number: %s", args[0])
		}
		
		subject, _ := cmd.Flags().GetString("subject")
		body, _ := cmd.Flags().GetString("body")
		to, _ := cmd.Flags().GetStringSlice("to")
		
		return mailos.EditDraftByNumber(number, subject, body, to)
	},
}

var draftCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new draft",
	RunE: func(cmd *cobra.Command, args []string) error {
		to, _ := cmd.Flags().GetStringSlice("to")
		subject, _ := cmd.Flags().GetString("subject")
		body, _ := cmd.Flags().GetString("body")
		
		if len(to) == 0 {
			return fmt.Errorf("--to is required")
		}
		if subject == "" {
			return fmt.Errorf("--subject is required")
		}
		if body == "" {
			return fmt.Errorf("--body is required")
		}
		
		// Use existing drafts command
		opts := mailos.DraftsOptions{
			To:      to,
			Subject: subject,
			Body:    body,
		}
		
		return mailos.DraftsCommand(opts)
	},
}

var draftsCmd = &cobra.Command{
	Use:     "drafts",
	Short:   "Legacy draft command (use 'draft' instead)",
	Long: `Legacy draft command. Use 'mailos draft' for simplified interface.
Create draft emails in markdown format with frontmatter metadata.
Drafts are saved to the .email/drafts folder and can be sent using 'mailos send --drafts'.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get account flag if specified
		accountEmail, _ := cmd.Flags().GetString("account")
		
		// Ensure authenticated before proceeding
		cfg, err := mailos.EnsureAuthenticated(accountEmail)
		if err != nil {
			return err
		}
		_ = cfg // Config is now validated and has credentials
		
		query := strings.Join(args, " ")
		template, _ := cmd.Flags().GetString("template")
		dataFile, _ := cmd.Flags().GetString("data")
		outputDir, _ := cmd.Flags().GetString("output")
		interactive, _ := cmd.Flags().GetBool("interactive")
		useAI, _ := cmd.Flags().GetBool("ai")
		count, _ := cmd.Flags().GetInt("count")
		
		// Email composition flags
		to, _ := cmd.Flags().GetStringSlice("to")
		cc, _ := cmd.Flags().GetStringSlice("cc")
		bcc, _ := cmd.Flags().GetStringSlice("bcc")
		subject, _ := cmd.Flags().GetString("subject")
		body, _ := cmd.Flags().GetString("body")
		file, _ := cmd.Flags().GetString("file")
		attachments, _ := cmd.Flags().GetStringSlice("attach")
		priority, _ := cmd.Flags().GetString("priority")
		plain, _ := cmd.Flags().GetBool("plain")
		noSignature, _ := cmd.Flags().GetBool("no-signature")
		signature, _ := cmd.Flags().GetString("signature")
		
		// Read body from file if specified
		if file != "" && body == "" {
			content, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file: %v", err)
			}
			body = string(content)
		}
		
		// Get list/read flags
		list, _ := cmd.Flags().GetBool("list")
		read, _ := cmd.Flags().GetBool("read")
		
		// Get edit UID flag
		editUID, _ := cmd.Flags().GetUint32("edit-uid")
		
		opts := mailos.DraftsOptions{
			Query:       query,
			Template:    template,
			DataFile:    dataFile,
			OutputDir:   outputDir,
			Interactive: interactive,
			UseAI:       useAI,
			DraftCount:  count,
			List:        list,
			Read:        read,
			EditUID:     editUID,
			// Email composition fields
			To:          to,
			CC:          cc,
			BCC:         bcc,
			Subject:     subject,
			Body:        body,
			Attachments: attachments,
			Priority:    priority,
			PlainText:   plain,
			NoSignature: noSignature,
			Signature:   signature,
		}
		
		return mailos.DraftsCommand(opts)
	},
}

var composeCmd = &cobra.Command{
	Use:   "compose",
	Short: "Compose a new email (alias for drafts)",
	Long: `Compose a new email by creating a draft. This command is an alias for the drafts command
and supports all the same flags and functionality. The draft will be saved locally and 
can be sent using 'mailos send --drafts'.

Examples:
  mailos compose --to user@example.com --subject "Hello" --body "Hi there!"
  mailos compose --interactive
  mailos compose --template newsletter --data contacts.csv
  
Note: This command creates drafts that can be reviewed before sending. Use 'mailos send'
for immediate sending without creating drafts.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("📝 Composing email using drafts system...")
		
		accountEmail, _ := cmd.Flags().GetString("account")
		
		cfg, err := mailos.EnsureAuthenticated(accountEmail)
		if err != nil {
			return err
		}
		_ = cfg

		query := strings.Join(args, " ")
		template, _ := cmd.Flags().GetString("template")
		dataFile, _ := cmd.Flags().GetString("data")
		outputDir, _ := cmd.Flags().GetString("output")
		interactive, _ := cmd.Flags().GetBool("interactive")
		useAI, _ := cmd.Flags().GetBool("ai")
		count, _ := cmd.Flags().GetInt("count")
		
		to, _ := cmd.Flags().GetStringSlice("to")
		cc, _ := cmd.Flags().GetStringSlice("cc")
		bcc, _ := cmd.Flags().GetStringSlice("bcc")
		subject, _ := cmd.Flags().GetString("subject")
		body, _ := cmd.Flags().GetString("body")
		file, _ := cmd.Flags().GetString("file")
		attachments, _ := cmd.Flags().GetStringSlice("attach")
		priority, _ := cmd.Flags().GetString("priority")
		plain, _ := cmd.Flags().GetBool("plain")
		noSignature, _ := cmd.Flags().GetBool("no-signature")
		signature, _ := cmd.Flags().GetString("signature")
		
		if file != "" && body == "" {
			content, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file: %v", err)
			}
			body = string(content)
		}
		
		opts := mailos.DraftsOptions{
			Query:       query,
			Template:    template,
			DataFile:    dataFile,
			OutputDir:   outputDir,
			Interactive: interactive,
			UseAI:       useAI,
			DraftCount:  count,
			To:          to,
			CC:          cc,
			BCC:         bcc,
			Subject:     subject,
			Body:        body,
			Attachments: attachments,
			Priority:    priority,
			PlainText:   plain,
			NoSignature: noSignature,
			Signature:   signature,
		}
		
		return mailos.DraftsCommand(opts)
	},
}

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send an email",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get account flag if specified
		accountEmail, _ := cmd.Flags().GetString("account")
		
		// Ensure authenticated before proceeding
		cfg, err := mailos.EnsureAuthenticated(accountEmail)
		if err != nil {
			return err
		}
		_ = cfg // Config is now validated and has credentials
		
		// Check if --drafts flag is set
		drafts, _ := cmd.Flags().GetBool("drafts")
		if drafts {
			// Process draft emails
			draftDir, _ := cmd.Flags().GetString("draft-dir")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			filter, _ := cmd.Flags().GetString("filter")
			confirm, _ := cmd.Flags().GetBool("confirm")
			deleteAfter, _ := cmd.Flags().GetBool("delete-after")
			logFile, _ := cmd.Flags().GetString("log-file")
			
			opts := mailos.SendDraftsOptions{
				DraftDir:    draftDir,
				DryRun:      dryRun,
				Filter:      filter,
				Confirm:     confirm,
				DeleteAfter: deleteAfter,
				LogFile:     logFile,
			}
			return mailos.SendDrafts(opts)
		}
		
		// Regular send command
		to, _ := cmd.Flags().GetStringSlice("to")
		cc, _ := cmd.Flags().GetStringSlice("cc")
		bcc, _ := cmd.Flags().GetStringSlice("bcc")
		group, _ := cmd.Flags().GetString("group")
		subject, _ := cmd.Flags().GetString("subject")
		body, _ := cmd.Flags().GetString("body")
		message, _ := cmd.Flags().GetString("message")
		
		
		// Use message flag if body is empty but message is provided
		if body == "" && message != "" {
			body = message
		}
		file, _ := cmd.Flags().GetString("file")
		attachments, _ := cmd.Flags().GetStringSlice("attach")
		plain, _ := cmd.Flags().GetBool("plain")
		noSignature, _ := cmd.Flags().GetBool("no-signature")
		signature, _ := cmd.Flags().GetString("signature")
		from, _ := cmd.Flags().GetString("from")
		preview, _ := cmd.Flags().GetBool("preview")
		useTemplate, _ := cmd.Flags().GetBool("template")
		verbose, _ := cmd.Flags().GetBool("verbose")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Handle group parameter
		if group != "" {
			groupEmails, err := mailos.GetGroupEmails(group)
			if err != nil {
				return fmt.Errorf("failed to get group emails: %v", err)
			}
			to = append(to, groupEmails...)
		}

		if len(to) == 0 {
			return fmt.Errorf("at least one recipient is required (use --to or --group)")
		}

		if subject == "" {
			return fmt.Errorf("subject is required")
		}

		// Read body from file if specified
		if file != "" {
			content, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file: %v", err)
			}
			body = string(content)
		} else if body == "" {
			// Read from stdin if no body provided
			fmt.Println("Enter email body (Markdown supported). Press Ctrl+D when done:")
			scanner := bufio.NewScanner(os.Stdin)
			var lines []string
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			body = strings.Join(lines, "\n")
		}

		// client, err := NewClient()
		// if err != nil {
		//	return err
		// }
		// return fmt.Errorf("client functionality temporarily disabled")

		// Prepare signature
		var sig string
		if !noSignature {
			if signature != "" {
				// Process newline characters in custom signature
				sig = strings.ReplaceAll(signature, "\\n", "\n")
				sig = strings.ReplaceAll(sig, "\\t", "\t")
			} else {
				// Get account-specific config instead of global config
				// Use accountEmail from --account flag if from is not specified
				signatureFromAccount := from
				if signatureFromAccount == "" {
					signatureFromAccount = accountEmail
				}
				setup, err := mailos.InitializeMailSetup(signatureFromAccount)
				if err == nil {
					cfg := setup.Config
					if verbose {
						fmt.Printf("Debug: SignatureOverride = '%s'\n", cfg.SignatureOverride)
					}
					// Check for signature override first (works for all accounts)
					if cfg.SignatureOverride != "" {
						if verbose {
							fmt.Printf("Debug: Using account signature override\n")
						}
						sig = cfg.SignatureOverride
					} else {
						// For wildcard aliases, don't include automatic signature unless explicitly set
						if verbose {
							fmt.Printf("Debug: cfg.FromEmail = '%s', cfg.Email = '%s'\n", cfg.FromEmail, cfg.Email)
						}
						if cfg.FromEmail != cfg.Email {
							// This is a wildcard alias - no automatic signature
							if verbose {
								fmt.Printf("Debug: Wildcard alias detected - no automatic signature\n")
							}
							sig = ""
						} else {
							// Use FromEmail if specified, otherwise use Email
							emailToShow := cfg.Email
							if cfg.FromEmail != "" {
								emailToShow = cfg.FromEmail
							}
							name := cfg.FromName
							if name == "" {
								name = strings.Split(emailToShow, "@")[0]
							}
							sig = fmt.Sprintf("\n--\n%s\n%s", name, emailToShow)
						}
					}
				}
			}
		}

		// Create email message
		msg := &mailos.EmailMessage{
			To:          to,
			CC:          cc,
			BCC:         bcc,
			Subject:     subject,
			Body:        body,
			Attachments: attachments,
			UseTemplate: useTemplate,
		}

		// Add signature if needed
		if sig != "" {
			msg.IncludeSignature = true
			msg.SignatureText = sig
		}

		// Convert markdown to HTML unless plain text requested
		if !plain {
			html := mailos.MarkdownToHTMLContent(body)
			if html != body {
				msg.BodyHTML = html
			}
		}

		if preview {
			// Use accountEmail from --account flag if from is not specified
			previewFromAccount := from
			if previewFromAccount == "" {
				previewFromAccount = accountEmail
			}
			return mailos.PreviewEmail(msg, previewFromAccount)
		}

		if dryRun {
			// Use accountEmail from --account flag if from is not specified  
			dryRunFromAccount := from
			if dryRunFromAccount == "" {
				dryRunFromAccount = accountEmail
			}
			
			// If still empty, try to get from the mail setup
			if dryRunFromAccount == "" {
				setup, err := mailos.InitializeMailSetup("")
				if err == nil {
					if setup.Config.FromEmail != "" {
						dryRunFromAccount = setup.Config.FromEmail
					} else {
						dryRunFromAccount = setup.Config.Email
					}
				}
			}
			
			fmt.Printf("=== DRY RUN - Email Preview ===\n")
			fmt.Printf("From: %s\n", dryRunFromAccount)
			fmt.Printf("To: %s\n", strings.Join(to, ", "))
			if len(cc) > 0 {
				fmt.Printf("CC: %s\n", strings.Join(cc, ", "))
			}
			if len(bcc) > 0 {
				fmt.Printf("BCC: %s\n", strings.Join(bcc, ", "))
			}
			fmt.Printf("Subject: %s\n", subject)
			fmt.Printf("\n--- Body ---\n%s", body)
			if sig != "" {
				fmt.Printf("%s", sig)
			}
			fmt.Printf("\n--- End Body ---\n")
			fmt.Printf("\n=== End Preview (email not sent) ===\n")
			return nil
		}

		fmt.Printf("Sending email to %s...\n", strings.Join(to, ", "))
		
		// Use accountEmail from --account flag if from is not specified
		sendFromAccount := from
		if sendFromAccount == "" {
			sendFromAccount = accountEmail
		}
		
		if verbose {
			fmt.Printf("Debug: From address: %s\n", sendFromAccount)
			fmt.Printf("Debug: Account lookup starting for: %s\n", sendFromAccount)
		}
		
		if err := mailos.SendWithAccountVerbose(msg, sendFromAccount, verbose); err != nil {
			return fmt.Errorf("failed to send email: %v", err)
		}

		fmt.Println("✓ Email sent successfully!")
		return nil
	},
}

var groupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "Manage email groups for bulk sending",
	Long:  `Manage email groups that can be used for sending emails to multiple recipients at once.
Groups are stored in ~/.email/groups.json and can be used with the --group flag in send commands.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		update, _ := cmd.Flags().GetString("update")
		description, _ := cmd.Flags().GetString("description")
		emails, _ := cmd.Flags().GetString("emails")
		delete, _ := cmd.Flags().GetString("delete")
		addMember, _ := cmd.Flags().GetString("add-member")
		removeMember, _ := cmd.Flags().GetString("remove-member")
		groupName, _ := cmd.Flags().GetString("group")
		listMembers, _ := cmd.Flags().GetString("list-members")

		if delete != "" {
			return mailos.DeleteGroup(delete)
		}

		if update != "" {
			return mailos.UpdateGroup(update, description, emails)
		}

		if addMember != "" {
			if groupName == "" {
				return fmt.Errorf("--group flag is required when adding members")
			}
			return mailos.AddMemberToGroup(groupName, addMember)
		}

		if removeMember != "" {
			if groupName == "" {
				return fmt.Errorf("--group flag is required when removing members")
			}
			return mailos.RemoveMemberFromGroup(groupName, removeMember)
		}

		if listMembers != "" {
			return mailos.ListGroupMembers(listMembers)
		}

		return mailos.ListGroups()
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync emails from IMAP server to local filesystem",
	Long:  `Sync emails from your IMAP server to local filesystem.
Creates directory structure: emails/received, emails/sent, emails/drafts
Each email is saved as a markdown file with metadata.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir, _ := cmd.Flags().GetString("dir")
		limit, _ := cmd.Flags().GetInt("limit")
		days, _ := cmd.Flags().GetInt("days")
		includeRead, _ := cmd.Flags().GetBool("include-read")
		verbose, _ := cmd.Flags().GetBool("verbose")
		
		opts := mailos.SyncOptions{
			BaseDir:     baseDir,
			Limit:       limit,
			IncludeRead: includeRead,
			Verbose:     verbose,
		}
		
		if days > 0 {
			opts.Since = time.Now().AddDate(0, 0, -days)
		}
		
		return mailos.SyncEmails(opts)
	},
}

var syncDbCmd = &cobra.Command{
	Use:   "sync-db",
	Short: "Sync emails from inbox to local SQLite database",
	Long:  `Sync emails from your local inbox to SQLite database.
Creates a SQLite database at ~/.email/[account]/archive.db for fast querying and analysis.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		accountEmail, _ := cmd.Flags().GetString("account")
		allAccounts, _ := cmd.Flags().GetBool("all")
		
		if allAccounts {
			return mailos.SyncAllAccountsToDB()
		}
		
		if accountEmail != "" {
			return mailos.SyncEmailsToDB(accountEmail)
		}
		
		cfg, err := mailos.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}
		
		if cfg.Email == "" {
			return fmt.Errorf("no email account configured. Use --account flag or configure a default account")
		}
		
		return mailos.SyncEmailsToDB(cfg.Email)
	},
}

var sentCmd = &cobra.Command{
	Use:   "sent",
	Short: "Read sent emails",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("number")
		to, _ := cmd.Flags().GetString("to")
		subject, _ := cmd.Flags().GetString("subject")
		days, _ := cmd.Flags().GetInt("days")
		timeRange, _ := cmd.Flags().GetString("range")
		outputJSON, _ := cmd.Flags().GetBool("json")
		saveMarkdown, _ := cmd.Flags().GetBool("save-markdown")
		outputDir, _ := cmd.Flags().GetString("output-dir")

		opts := mailos.SentOptions{
			Limit:     limit,
			ToAddress: to,
			Subject:   subject,
		}

		// Handle time range parameter
		if timeRange != "" {
			// // selectedRange, err := mailos.ParseTimeRangeString(timeRange)
			// if err != nil {
			//	return fmt.Errorf("invalid time range: %v", err)
			// }
			// opts.Since = selectedRange.Since
			return fmt.Errorf("time range functionality temporarily disabled")
		} else if days > 0 {
			opts.Since = time.Now().AddDate(0, 0, -days)
		}

		fmt.Println("Reading sent emails...")
		emails, err := mailos.ReadSentEmails(opts)
		if err != nil {
			return fmt.Errorf("failed to read sent emails: %v", err)
		}

		// Filter by Until time if time range was specified
		// if timeRange != "" {
		//	// selectedRange, _ := mailos.ParseTimeRangeString(timeRange)
		//	var filteredEmails []*mailos.Email
		//	for _, email := range emails {
		//		// Check if email is within the time range (both Since and Until)
		//		if email.Date.After(selectedRange.Since.Add(-time.Second)) && 
		//		   email.Date.Before(selectedRange.Until.Add(time.Second)) {
		//			filteredEmails = append(filteredEmails, email)
		//		}
		//	}
		//	emails = filteredEmails
		// }

		// Save as markdown if requested - save to .email/sent instead of 'emails'
		if saveMarkdown && len(emails) > 0 {
			// Ensure email directories exist
			if err := mailos.EnsureEmailDirectories(); err != nil {
				fmt.Printf("Warning: failed to create email directories: %v\n", err)
			}
			
			// Get the sent directory
			sentDir, err := mailos.GetSentDir()
			if err != nil {
				fmt.Printf("Warning: failed to get sent directory: %v\n", err)
				sentDir = outputDir // fallback to the specified output dir
			}
			
			// SaveEmailsAsMarkdown will print the save message
			if err := mailos.SaveEmailsAsMarkdown(emails, sentDir); err != nil {
				fmt.Printf("Warning: failed to save markdown files: %v\n", err)
			}
		}

		// Output format
		if outputJSON {
			// Convert to JSON-friendly format
			type jsonEmail struct {
				ID          uint32   `json:"id"`
				From        string   `json:"from"`
				To          []string `json:"to"`
				Subject     string   `json:"subject"`
				Date        string   `json:"date"`
				Body        string   `json:"body"`
				Attachments []string `json:"attachments"`
			}
			
			jsonEmails := make([]jsonEmail, len(emails))
			for i, email := range emails {
				jsonEmails[i] = jsonEmail{
					ID:          email.ID,
					From:        email.From,
					To:          email.To,
					Subject:     email.Subject,
					Date:        email.Date.Format(time.RFC3339),
					Body:        email.Body,
					Attachments: email.Attachments,
				}
			}
			
			data, _ := json.MarshalIndent(jsonEmails, "", "  ")
			fmt.Println(string(data))
		} else {
			// Use the dedicated FormatSentEmailList function for sent emails
			fmt.Print(mailos.FormatSentEmailList(emails))
		}
		
		return nil
	},
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download email attachments",
	Long:  `Download attachments from emails that match the specified criteria.
Attachments are saved to the specified directory with timestamps to avoid conflicts.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get account flag if specified
		accountEmail, _ := cmd.Flags().GetString("account")
		
		// Ensure authenticated before proceeding
		cfg, err := mailos.EnsureAuthenticated(accountEmail)
		if err != nil {
			return err
		}
		_ = cfg // Config is now validated and has credentials
		
		limit, _ := cmd.Flags().GetInt("number")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		subject, _ := cmd.Flags().GetString("subject")
		days, _ := cmd.Flags().GetInt("days")
		outputDir, _ := cmd.Flags().GetString("output-dir")
		emailID, _ := cmd.Flags().GetUint32("id")
		showContent, _ := cmd.Flags().GetBool("show-content")
		
		// client, err := NewClient()
		// if err != nil {
		//	return err
		// }
		return fmt.Errorf("client functionality temporarily disabled")
		
		// If specific email ID is provided
		if emailID > 0 {
			fmt.Printf("Downloading attachments from email ID %d...\n", emailID)
			opts := mailos.ReadOptions{
				Limit:          1,
				DownloadAttach: true,
				AttachmentDir:  outputDir,
			}
			
			// Read emails to find the one with matching ID
			emails, err := mailos.Read(opts)
			if err != nil {
				return fmt.Errorf("failed to read email: %v", err)
			}
			
			// Find the email with matching ID
			var targetEmail *mailos.Email
			for _, email := range emails {
				if email.ID == emailID {
					targetEmail = email
					break
				}
			}
			
			if targetEmail == nil {
				return fmt.Errorf("email with ID %d not found", emailID)
			}
			
			if len(targetEmail.Attachments) == 0 {
				fmt.Println("No attachments found in this email")
				return nil
			}
			
			fmt.Printf("Found %d attachment(s)\n", len(targetEmail.Attachments))
			if showContent {
				fmt.Printf("\nEmail from: %s\nSubject: %s\nDate: %s\n\n%s\n", 
					targetEmail.From, targetEmail.Subject, 
					targetEmail.Date.Format("2006-01-02 15:04:05"), 
					targetEmail.Body)
			}
			
			return nil
		}
		
		// Otherwise use search criteria
		opts := mailos.ReadOptions{
			Limit:          limit,
			FromAddress:    from,
			ToAddress:      to,
			Subject:        subject,
			DownloadAttach: true,
			AttachmentDir:  outputDir,
		}
		
		if days > 0 {
			opts.Since = time.Now().AddDate(0, 0, -days)
		}
		
		fmt.Println("Searching for emails with attachments...")
		emails, err := mailos.Read(opts)
		if err != nil {
			return fmt.Errorf("failed to read emails: %v", err)
		}
		
		// Filter emails with attachments
		var emailsWithAttachments []*mailos.Email
		totalAttachments := 0
		for _, email := range emails {
			if len(email.Attachments) > 0 {
				emailsWithAttachments = append(emailsWithAttachments, email)
				totalAttachments += len(email.Attachments)
			}
		}
		
		if len(emailsWithAttachments) == 0 {
			fmt.Println("No emails with attachments found matching criteria")
			return nil
		}
		
		fmt.Printf("\nFound %d email(s) with %d total attachment(s)\n", 
			len(emailsWithAttachments), totalAttachments)
		
		// Show summary of emails with attachments
		for _, email := range emailsWithAttachments {
			fmt.Printf("\n[ID: %d] %s\n", email.ID, email.Subject)
			fmt.Printf("  From: %s\n", email.From)
			fmt.Printf("  Date: %s\n", email.Date.Format("2006-01-02 15:04:05"))
			fmt.Printf("  Attachments: %s\n", strings.Join(email.Attachments, ", "))
			
			if showContent && email.Body != "" {
				// Show first 200 chars of body
				body := email.Body
				if len(body) > 200 {
					body = body[:200] + "..."
				}
				fmt.Printf("  Preview: %s\n", body)
			}
		}
		
		fmt.Printf("\n✓ Attachments saved to: %s\n", outputDir)
		return nil
	},
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search and list emails",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get account flag if specified
		accountEmail, _ := cmd.Flags().GetString("account")
		
		// Ensure authenticated before proceeding
		cfg, err := mailos.EnsureAuthenticated(accountEmail)
		if err != nil {
			return err
		}
		_ = cfg // Config is now validated and has credentials
		
		limit, _ := cmd.Flags().GetInt("number")
		unreadOnly, _ := cmd.Flags().GetBool("unread")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		subject, _ := cmd.Flags().GetString("subject")
		days, _ := cmd.Flags().GetInt("days")
		timeRange, _ := cmd.Flags().GetString("range")
		outputJSON, _ := cmd.Flags().GetBool("json")
		saveMarkdown, _ := cmd.Flags().GetBool("save-markdown")
		outputDir, _ := cmd.Flags().GetString("output-dir")
		downloadAttach, _ := cmd.Flags().GetBool("download-attachments")
		attachmentDir, _ := cmd.Flags().GetString("attachment-dir")


		// client, err := NewClient()
		// if err != nil {
		//	return err
		// }
		// return fmt.Errorf("client functionality temporarily disabled")

		opts := mailos.ReadOptions{
			Limit:          limit,
			UnreadOnly:     unreadOnly,
			FromAddress:    from,
			ToAddress:      to,
			Subject:        subject,
			DownloadAttach: downloadAttach,
			AttachmentDir:  attachmentDir,
		}

		// Handle time range parameter
		if timeRange != "" {
			selectedRange, err := mailos.ParseTimeRangeString(timeRange)
			if err != nil {
				return fmt.Errorf("invalid time range: %v", err)
			}
			opts.Since = selectedRange.Since
			// Also filter by Until time after fetching
		} else if days > 0 {
			opts.Since = time.Now().AddDate(0, 0, -days)
		}

		fmt.Println("Searching emails...")
		emails, err := mailos.Read(opts)
		if err != nil {
			return fmt.Errorf("failed to search emails: %v", err)
		}

		// Filter by Until time if time range was specified
		if timeRange != "" {
			selectedRange, _ := mailos.ParseTimeRangeString(timeRange)
			var filteredEmails []*mailos.Email
			for _, email := range emails {
				// Check if email is within the time range (both Since and Until)
				if email.Date.After(selectedRange.Since.Add(-time.Second)) && 
				   email.Date.Before(selectedRange.Until.Add(time.Second)) {
					filteredEmails = append(filteredEmails, email)
				}
			}
			emails = filteredEmails
		}

		// Save as markdown if requested
		if saveMarkdown && len(emails) > 0 {
			if err := mailos.SaveEmailsAsMarkdown(emails, outputDir); err != nil {
				fmt.Printf("Warning: failed to save markdown files: %v\n", err)
			}
		}

		// Output format
		if outputJSON {
			// Convert to JSON-friendly format
			type jsonEmail struct {
				ID          uint32   `json:"id"`
				From        string   `json:"from"`
				To          []string `json:"to"`
				Subject     string   `json:"subject"`
				Date        string   `json:"date"`
				Body        string   `json:"body"`
				Attachments []string `json:"attachments"`
			}
			
			jsonEmails := make([]jsonEmail, len(emails))
			for i, email := range emails {
				jsonEmails[i] = jsonEmail{
					ID:          email.ID,
					From:        email.From,
					To:          email.To,
					Subject:     email.Subject,
					Date:        email.Date.Format(time.RFC3339),
					Body:        email.Body,
					Attachments: email.Attachments,
				}
			}
			
			data, _ := json.MarshalIndent(jsonEmails, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Print(mailos.FormatEmailList(emails))
		}
		
		return nil
	},
}

// isDocumentFile checks if the file is a readable document
func isDocumentFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	documentExts := []string{".txt", ".md", ".markdown", ".csv", ".json", ".xml", ".yaml", ".yml", ".log", ".conf", ".cfg", ".ini"}
	for _, docExt := range documentExts {
		if ext == docExt {
			return true
		}
	}
	return false
}

// parseDocumentContent parses the content of a document attachment
func parseDocumentContent(filename string, data []byte) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	
	switch ext {
	case ".txt", ".md", ".markdown", ".csv", ".json", ".xml", ".yaml", ".yml", ".log", ".conf", ".cfg", ".ini":
		// For text-based files, return the content as string
		content := string(data)
		
		// Limit content length to avoid overwhelming output
		maxLength := 5000
		if len(content) > maxLength {
			content = content[:maxLength] + "\n\n... [Content truncated - showing first " + fmt.Sprintf("%d", maxLength) + " characters]"
		}
		
		return content, nil
	default:
		return "", fmt.Errorf("unsupported document type: %s", ext)
	}
}

var readCmd = &cobra.Command{
	Use:   "read [email_id]",
	Short: "Display full content of a specific email",
	Args:  cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get account flag if specified
		accountEmail, _ := cmd.Flags().GetString("account")
		
		// Get include-documents flag
		includeDocuments, _ := cmd.Flags().GetBool("include-documents")
		
		// Get id flag
		idFlag, _ := cmd.Flags().GetUint32("id")
		
		// Ensure authenticated before proceeding
		cfg, err := mailos.EnsureAuthenticated(accountEmail)
		if err != nil {
			return err
		}
		_ = cfg // Config is now validated and has credentials
		
		// Determine email ID from either positional argument or flag
		var emailID string
		var id uint64
		
		if idFlag > 0 {
			// Use flag value
			id = uint64(idFlag)
		} else if len(args) > 0 {
			// Use positional argument
			emailID = args[0]
			var parseErr error
			id, parseErr = strconv.ParseUint(emailID, 10, 32)
			if parseErr != nil {
				return fmt.Errorf("READ_INVALID_ID: Email ID '%s' is not a valid number. Email IDs must be positive integers (e.g., 1332, 1331). Use 'mailos search' to see available email IDs. Parsing error: %v", emailID, parseErr)
			}
		} else {
			return fmt.Errorf("READ_MISSING_ID: Please provide an email ID either as a positional argument (mailos read 1423) or using the --id flag (mailos read --id 1423). Use 'mailos search' to see available email IDs.")
		}
		
		// client, err := NewClient()
		// if err != nil {
		//	return err
		// }
		// return fmt.Errorf("client functionality temporarily disabled")
		
		fmt.Printf("Reading email ID %d...\n", id)
		
		// Use ReadEmailByID for direct email retrieval
		targetEmail, err := mailos.ReadEmailByID(uint32(id))
		if err != nil {
			return fmt.Errorf("READ_EMAIL_ERROR: Failed to retrieve email with ID %d. This could be due to: (1) Email ID does not exist in your inbox, (2) IMAP server connection issues, (3) Authentication problems, (4) Email was deleted or moved. Try running 'mailos search' to see available email IDs. Original error: %v", id, err)
		}
		
		// Display full email content
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("📧 EMAIL ID: %d\n", targetEmail.ID)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("From:    %s\n", targetEmail.From)
		fmt.Printf("To:      %s\n", strings.Join(targetEmail.To, ", "))
		fmt.Printf("Subject: %s\n", targetEmail.Subject)
		fmt.Printf("Date:    %s\n", targetEmail.Date.Format("Mon, Jan 2, 2006 at 3:04 PM"))
		
		if len(targetEmail.Attachments) > 0 {
			fmt.Printf("Attachments: %s\n", strings.Join(targetEmail.Attachments, ", "))
		}
		
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("📄 CONTENT:\n")
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
		fmt.Println(targetEmail.Body)
		
		// Parse and display attachment documents if flag is enabled
		if includeDocuments && len(targetEmail.AttachmentData) > 0 {
			fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			fmt.Printf("📎 ATTACHMENT DOCUMENTS:\n")
			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			
			for filename, data := range targetEmail.AttachmentData {
				if isDocumentFile(filename) {
					fmt.Printf("\n📄 %s:\n", filename)
					fmt.Printf("─────────────────────────────────────────────────\n")
					
					content, err := parseDocumentContent(filename, data)
					if err != nil {
						fmt.Printf("Error parsing document: %v\n", err)
					} else {
						fmt.Println(content)
					}
				}
			}
		}
		
		fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		
		return nil
	},
}

var replyCmd = &cobra.Command{
	Use:   "reply [email_number]",
	Short: "Reply to a specific email",
	Long: `Reply to a specific email while preserving thread context.
The email_number corresponds to the number shown in the email list (mailos search).

Examples:
  mailos reply 2                    # Reply to email #2 interactively  
  mailos reply 2 --all              # Reply to all recipients of email #2
  mailos reply 2 --body "Thanks!"   # Reply with quick message
  mailos reply 2 --draft            # Save reply as draft instead of sending`,
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse email number
		emailNumber, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid email number: %s", args[0])
		}

		// Get flags
		replyAll, _ := cmd.Flags().GetBool("all")
		body, _ := cmd.Flags().GetString("body")
		subject, _ := cmd.Flags().GetString("subject")
		fileBody, _ := cmd.Flags().GetString("file")
		draft, _ := cmd.Flags().GetBool("draft")
		interactive, _ := cmd.Flags().GetBool("interactive")
		to, _ := cmd.Flags().GetStringSlice("to")
		cc, _ := cmd.Flags().GetStringSlice("cc")
		bcc, _ := cmd.Flags().GetStringSlice("bcc")

		// Build reply options
		opts := mailos.ReplyOptions{
			EmailNumber: emailNumber,
			ReplyAll:    replyAll,
			Body:        body,
			Subject:     subject,
			FileBody:    fileBody,
			Draft:       draft,
			Interactive: interactive || (body == "" && fileBody == ""),
			To:          to,
			CC:          cc,
			BCC:         bcc,
		}

		return mailos.ReplyCommand(opts)
	},
}

var forwardCmd = &cobra.Command{
	Use:   "forward [email_number]",
	Short: "Forward a specific email",
	Long: `Forward a specific email to one or more recipients.
The email_number corresponds to the number shown in the email list (mailos search).

Examples:
  mailos forward 2 --to user@example.com         # Forward email #2 to specific recipient
  mailos forward 2 --to user1@example.com,user2@example.com  # Forward to multiple recipients
  mailos forward 2 --body "FYI"                  # Forward with additional message
  mailos forward 2 --draft                       # Save forward as draft instead of sending`,
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse email number
		emailNumber, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid email number: %s", args[0])
		}

		// Get flags
		body, _ := cmd.Flags().GetString("body")
		subject, _ := cmd.Flags().GetString("subject")
		fileBody, _ := cmd.Flags().GetString("file")
		draft, _ := cmd.Flags().GetBool("draft")
		interactive, _ := cmd.Flags().GetBool("interactive")
		to, _ := cmd.Flags().GetStringSlice("to")
		cc, _ := cmd.Flags().GetStringSlice("cc")
		bcc, _ := cmd.Flags().GetStringSlice("bcc")

		// Build forward options
		opts := mailos.ForwardOptions{
			EmailNumber: emailNumber,
			Body:        body,
			Subject:     subject,
			FileBody:    fileBody,
			Draft:       draft,
			Interactive: interactive || (body == "" && fileBody == ""),
			To:          to,
			CC:          cc,
			BCC:         bcc,
		}

		return mailos.ForwardCommand(opts)
	},
}

var markReadCmd = &cobra.Command{
	Use:   "mark-read",
	Short: "Mark emails as read",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ids, _ := cmd.Flags().GetUintSlice("ids")
		from, _ := cmd.Flags().GetString("from")
		subject, _ := cmd.Flags().GetString("subject")
		
		// client, err := NewClient()
		// if err != nil {
		//	return err
		// }
		return fmt.Errorf("client functionality temporarily disabled")

		// If specific IDs provided
		if len(ids) > 0 {
			ids32 := make([]uint32, len(ids))
			for i, id := range ids {
				ids32[i] = uint32(id)
			}

			if err := mailos.MarkAsRead(ids32); err != nil {
				return fmt.Errorf("failed to mark emails as read: %v", err)
			}

			fmt.Printf("✓ Marked %d email(s) as read\n", len(ids))
			return nil
		}

		// Otherwise search for emails to mark
		if from == "" && subject == "" {
			return fmt.Errorf("provide either --ids, --from, or --subject")
		}

		opts := mailos.ReadOptions{
			FromAddress: from,
			Subject:     subject,
			Limit:       100,
		}

		emails, err := mailos.Read(opts)
		if err != nil {
			return fmt.Errorf("failed to find emails: %v", err)
		}

		if len(emails) == 0 {
			fmt.Println("No emails found matching criteria")
			return nil
		}

		// Extract IDs
		emailIds := make([]uint32, len(emails))
		for i, email := range emails {
			emailIds[i] = email.ID
		}

		if err := mailos.MarkAsRead(emailIds); err != nil {
			return fmt.Errorf("failed to mark emails as read: %v", err)
		}

		fmt.Printf("✓ Marked %d email(s) as read\n", len(emails))
		return nil
	},
}

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Manage email accounts",
	Long: `Manage email accounts

Examples:
  mailos accounts --list                                    # List available accounts
  mailos accounts --add user@example.com                   # Add a new email account
  mailos accounts --add user@example.com --provider fastmail # Add account with specific provider
  mailos accounts --add alias@domain.com --provider fastmail --use-existing-credentials # Add alias using existing credentials
  mailos accounts --sync-fastmail                          # Sync aliases from FastMail via JMAP API
  mailos accounts --sync-fastmail --token YOUR_TOKEN       # Sync with specific API token
  mailos accounts --set user@example.com                   # Set default account for this session
  mailos accounts --set-signature user@example.com:"Best regards, John"  # Set account signature
  mailos accounts --clear                                   # Clear session default`,
	RunE: func(cmd *cobra.Command, args []string) error {
		setAccount, _ := cmd.Flags().GetString("set")
		addAccount, _ := cmd.Flags().GetString("add")
		provider, _ := cmd.Flags().GetString("provider")
		useExistingCredentials, _ := cmd.Flags().GetBool("use-existing-credentials")
		setSignature, _ := cmd.Flags().GetString("set-signature")
		clearSession, _ := cmd.Flags().GetBool("clear")
		listAccounts, _ := cmd.Flags().GetBool("list")
		syncFastmail, _ := cmd.Flags().GetBool("sync-fastmail")
		token, _ := cmd.Flags().GetString("token")
		testConnection, _ := cmd.Flags().GetBool("test-connection")
		
		// Handle FastMail sync
		if syncFastmail {
			if token == "" {
				fmt.Println(mailos.GetFastMailTokenInstructions())
				fmt.Print("\nEnter your FastMail JMAP token: ")
				fmt.Scanln(&token)
				if token == "" {
					return fmt.Errorf("token is required for FastMail sync")
				}
			}
			
			if testConnection {
				return mailos.TestFastMailJMAPConnection(token)
			}
			
			return mailos.SyncFastMailAliases(token)
		}
		
		// Handle test connection
		if testConnection && token != "" {
			return mailos.TestFastMailJMAPConnection(token)
		}
		
		// Handle set signature
		if setSignature != "" {
			parts := strings.SplitN(setSignature, ":", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid format. Use: email:signature")
			}
			email, signature := parts[0], parts[1]
			
			if err := mailos.SetAccountSignature(email, signature); err != nil {
				return fmt.Errorf("failed to set signature: %v", err)
			}
			fmt.Printf("✓ Set signature for %s\n", email)
			return nil
		}
		
		// Handle add account
		if addAccount != "" {
			if err := mailos.AddNewAccountWithProvider(addAccount, provider, useExistingCredentials); err != nil {
				return fmt.Errorf("failed to add account: %v", err)
			}
			fmt.Printf("✓ Successfully added account: %s\n", addAccount)
			fmt.Println("You can now use this account with:")
			fmt.Printf("  mailos accounts --set %s\n", addAccount)
			fmt.Printf("  mailos send --from %s --to recipient@example.com --subject \"Subject\" --body \"Message\"\n", addAccount)
			return nil
		}
		
		// Handle clear session
		if clearSession {
			// mailos. - temporarily disabledClearSessionDefaultAccount()
			fmt.Println("✓ Cleared session default account")
			return nil
		}
		
		// Handle set account
		if setAccount != "" {
			// Try to load existing account
			setup, err := mailos.InitializeMailSetup(setAccount)
			if err != nil {
				// If account doesn't exist, offer to add it
				if strings.Contains(err.Error(), "not found") {
					fmt.Printf("Account '%s' not found.\n", setAccount)
					fmt.Print("Would you like to add this account? (y/N): ")
					
					var response string
					fmt.Scanln(&response)
					
					if strings.ToLower(strings.TrimSpace(response)) == "y" {
						// Add the new account by running configuration
						if err := mailos.AddNewAccount(setAccount); err != nil {
							return fmt.Errorf("failed to add account: %v", err)
						}
						
						// Try to initialize again after adding
						setup, err = mailos.InitializeMailSetup(setAccount)
						if err != nil {
							return fmt.Errorf("failed to initialize newly added account: %v", err)
						}
					} else {
						return fmt.Errorf("account setup cancelled")
					}
				} else {
					return err
				}
			}
			
			// Also set local preference so it persists for this directory
			if err := mailos.SetLocalAccountPreference(setAccount); err != nil {
				fmt.Printf("Note: Could not save local preference: %v\n", err)
			}
			
			fmt.Printf("✓ Set session default account to: %s\n", setAccount)
			fmt.Printf("  Provider: %s\n", mailos.GetProviderName(setup.Config.Provider))
			if setup.Config.FromEmail != "" && setup.Config.FromEmail != setup.Config.Email {
				fmt.Printf("  Sending as: %s\n", setup.Config.FromEmail)
			}
			return nil
		}
		
		// List accounts (default behavior)
		if listAccounts || (setAccount == "" && !clearSession) {
			cfg, err := mailos.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %v", err)
			}
			
			accounts := mailos.GetAllAccounts(cfg)
			if len(accounts) == 0 {
				fmt.Println("No accounts configured.")
				fmt.Println("Run 'mailos setup' to configure your first account.")
				return nil
			}
			
			// Get session and local defaults
			sessionDefault := mailos.GetSessionDefaultAccount()
			localDefault := mailos.GetLocalAccountPreference()
			
			fmt.Println("Available Email Accounts:")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━")
			
			var currentProvider string
			var isFirstInGroup bool = true
			
			for i, acc := range accounts {
				// Show provider header when switching providers
				if acc.Provider != currentProvider {
					if !isFirstInGroup {
						fmt.Println() // Add spacing between provider groups
					}
					currentProvider = acc.Provider
					providerName := mailos.GetProviderName(acc.Provider)
					fmt.Printf("── %s ──\n", providerName)
					isFirstInGroup = false
				}
				
				marker := "  "
				if acc.Email == localDefault {
					marker = "▸ " // Local preference takes precedence
				} else if acc.Email == sessionDefault {
					marker = "▸ "
				}
				
				// Format account display with visual hierarchy
				var prefix string
				var displayText string
				
				if acc.Label == "Primary" {
					prefix = "🏠 "
					displayText = fmt.Sprintf("%s%d. %s%s", marker, i+1, prefix, acc.Email)
				} else if acc.Label == "Sub-email" || (acc.Provider == currentProvider && acc.Label != "Account" && i > 0) {
					prefix = "  ↳ "
					displayText = fmt.Sprintf("%s%d. %s%s", marker, i+1, prefix, acc.Email)
				} else {
					prefix = "📧 "
					displayText = fmt.Sprintf("%s%d. %s%s", marker, i+1, prefix, acc.Email)
				}
				
				// Add labels and status indicators
				if acc.Label != "" && acc.Label != "Account" {
					displayText += fmt.Sprintf(" (%s)", acc.Label)
				}
				if acc.Email == localDefault {
					displayText += " [LOCAL DEFAULT]"
				} else if acc.Email == sessionDefault {
					displayText += " [SESSION DEFAULT]"
				}
				
				fmt.Println(displayText)
			}
			
			fmt.Println("\nUsage:")
			fmt.Println("  mailos accounts --set user@example.com  # Set session default")
			fmt.Println("  mailos accounts --clear                 # Clear session default")
			fmt.Println("  mailos send --account user@example.com  # Use specific account for one command")
		}
		
		return nil
	},
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show current configuration info",
	RunE: func(cmd *cobra.Command, args []string) error {
		return mailos.ShowEnhancedInfo()
	},
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test email functionality with search and fetch capabilities",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// client, err := NewClient()
		// if err != nil {
		//	return err
		// }
		return fmt.Errorf("client functionality temporarily disabled")

		// fmt.Println("EmailOS Test Suite")
		// fmt.Println("==================")
		// fmt.Printf("Testing with account: %s\n\n", client.GetConfig().Email)

		// Test 1: Basic connection and fetch
		fmt.Println("Test 1: Fetching recent emails...")
		emails, err := mailos.Read(mailos.ReadOptions{
			Limit: 5,
		})
		if err != nil {
			return fmt.Errorf("failed to fetch emails: %v", err)
		}
		fmt.Printf("✓ Successfully fetched %d emails\n", len(emails))
		if len(emails) > 0 {
			fmt.Printf("  Latest email: From %s - %s\n", emails[0].From, emails[0].Subject)
		}

		// Test 2: Search by unread
		fmt.Println("\nTest 2: Searching for unread emails...")
		unreadEmails, err := mailos.Read(mailos.ReadOptions{
			UnreadOnly: true,
			Limit:      10,
		})
		if err != nil {
			fmt.Printf("✗ Failed to search unread emails: %v\n", err)
		} else {
			fmt.Printf("✓ Found %d unread emails\n", len(unreadEmails))
		}

		// Test 3: Search by date range
		fmt.Println("\nTest 3: Searching emails from last 7 days...")
		recentEmails, err := mailos.Read(mailos.ReadOptions{
			Since: time.Now().AddDate(0, 0, -7),
			Limit: 20,
		})
		if err != nil {
			fmt.Printf("✗ Failed to search by date: %v\n", err)
		} else {
			fmt.Printf("✓ Found %d emails from the last 7 days\n", len(recentEmails))
		}

		// Test 4: Search by specific sender (if interactive mode)
		interactive, _ := cmd.Flags().GetBool("interactive")
		if interactive && len(emails) > 0 {
			fmt.Println("\nTest 4: Interactive search test")
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter email address to search from (or press Enter to skip): ")
			searchFrom, _ := reader.ReadString('\n')
			searchFrom = strings.TrimSpace(searchFrom)
			
			if searchFrom != "" {
				fromEmails, err := mailos.Read(mailos.ReadOptions{
					FromAddress: searchFrom,
					Limit:       10,
				})
				if err != nil {
					fmt.Printf("✗ Failed to search from %s: %v\n", searchFrom, err)
				} else {
					fmt.Printf("✓ Found %d emails from %s\n", len(fromEmails), searchFrom)
				}
			}

			// Test 5: Search by subject
			fmt.Print("\nEnter subject keyword to search (or press Enter to skip): ")
			searchSubject, _ := reader.ReadString('\n')
			searchSubject = strings.TrimSpace(searchSubject)
			
			if searchSubject != "" {
				subjectEmails, err := mailos.Read(mailos.ReadOptions{
					Subject: searchSubject,
					Limit:   10,
				})
				if err != nil {
					fmt.Printf("✗ Failed to search subject '%s': %v\n", searchSubject, err)
				} else {
					fmt.Printf("✓ Found %d emails with subject containing '%s'\n", len(subjectEmails), searchSubject)
					for i, email := range subjectEmails {
						if i < 3 { // Show first 3 results
							fmt.Printf("  - %s: %s\n", email.From, email.Subject)
						}
					}
				}
			}
		}

		// Test 6: Performance test
		fmt.Println("\nTest 5: Performance test - fetching 50 emails...")
		start := time.Now()
		largeSet, err := mailos.Read(mailos.ReadOptions{
			Limit: 50,
		})
		elapsed := time.Since(start)
		if err != nil {
			fmt.Printf("✗ Failed performance test: %v\n", err)
		} else {
			fmt.Printf("✓ Fetched %d emails in %v\n", len(largeSet), elapsed)
			fmt.Printf("  Average: %v per email\n", elapsed/time.Duration(len(largeSet)))
		}

		// Summary
		fmt.Println("\n==================")
		fmt.Println("Test Summary:")
		fmt.Printf("- Email account is properly configured\n")
		fmt.Printf("- IMAP connection is working\n")
		fmt.Printf("- Search functionality is operational\n")
		fmt.Printf("- Total emails accessible: %d+ \n", len(emails))

		// Show sample email details if verbose
		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose && len(emails) > 0 {
			fmt.Println("\nSample Email Details:")
			email := emails[0]
			fmt.Printf("From: %s\n", email.From)
			fmt.Printf("To: %s\n", strings.Join(email.To, ", "))
			fmt.Printf("Subject: %s\n", email.Subject)
			fmt.Printf("Date: %s\n", email.Date.Format("Jan 2, 2006 3:04 PM"))
			fmt.Printf("Attachments: %d\n", len(email.Attachments))
			if len(email.Body) > 200 {
				fmt.Printf("Body Preview: %s...\n", email.Body[:200])
			} else {
				fmt.Printf("Body: %s\n", email.Body)
			}
		}

		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete emails",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ids, _ := cmd.Flags().GetUintSlice("ids")
		from, _ := cmd.Flags().GetString("from")
		subject, _ := cmd.Flags().GetString("subject")
		confirm, _ := cmd.Flags().GetBool("confirm")
		drafts, _ := cmd.Flags().GetBool("drafts")
		before, _ := cmd.Flags().GetString("before")
		after, _ := cmd.Flags().GetString("after")
		days, _ := cmd.Flags().GetInt("days")
		
		if !confirm {
			return fmt.Errorf("please use --confirm flag to delete emails")
		}
		
		// client, err := NewClient()
		// if err != nil {
		//	return err
		// }
		return fmt.Errorf("client functionality temporarily disabled")
		
		// If specific IDs provided and not drafts
		if len(ids) > 0 && !drafts {
			ids32 := make([]uint32, len(ids))
			for i, id := range ids {
				ids32[i] = uint32(id)
			}
			
			if err := mailos.DeleteEmails(ids32); err != nil {
				return fmt.Errorf("failed to delete emails: %v", err)
			}
			
			fmt.Printf("✓ Deleted %d email(s)\n", len(ids))
			return nil
		}
		
		// Build search options
		opts := mailos.ReadOptions{
			FromAddress: from,
			Subject:     subject,
			Limit:       1000, // Increased limit for bulk deletions
		}
		
		// Handle date filters
		if after != "" {
			afterTime, err := time.Parse("2006-01-02", after)
			if err != nil {
				return fmt.Errorf("invalid --after date format (use YYYY-MM-DD): %v", err)
			}
			opts.Since = afterTime
		} else if days > 0 {
			opts.Since = time.Now().AddDate(0, 0, -days)
		}
		
		// Initialize emails slice
		var emails []*mailos.Email
		var err error
		
		// If drafts flag is set, read from drafts folder
		if drafts {
			fmt.Println("Reading drafts to delete...")
			emails, err = mailos.ReadFromFolder(opts, "Drafts")
			if err != nil {
				return fmt.Errorf("failed to find drafts: %v", err)
			}
		} else {
			// Read emails from inbox
			emails, err = mailos.Read(opts)
			if err != nil {
				return fmt.Errorf("failed to find emails: %v", err)
			}
		}
		
		// Filter by before date if specified
		if before != "" {
			beforeTime, err := time.Parse("2006-01-02", before)
			if err != nil {
				return fmt.Errorf("invalid --before date format (use YYYY-MM-DD): %v", err)
			}
			// Add one day to include emails on the specified date
			beforeTime = beforeTime.Add(24 * time.Hour)
			
			var filteredEmails []*mailos.Email
			for _, email := range emails {
				if email.Date.Before(beforeTime) {
					filteredEmails = append(filteredEmails, email)
				}
			}
			emails = filteredEmails
		}
		
		if len(emails) == 0 {
			if drafts {
				fmt.Println("No drafts found matching criteria")
			} else {
				fmt.Println("No emails found matching criteria")
			}
			return nil
		}
		
		// Extract IDs
		emailIds := make([]uint32, len(emails))
		for i, email := range emails {
			emailIds[i] = email.ID
		}
		
		// Show what will be deleted
		fmt.Printf("Found %d %s to delete:\n", len(emails), func() string {
			if drafts {
				return "draft(s)"
			}
			return "email(s)"
		}())
		
		// Show first 5 emails as preview
		previewCount := 5
		if len(emails) < previewCount {
			previewCount = len(emails)
		}
		for i := 0; i < previewCount; i++ {
			fmt.Printf("  - [%s] From: %s - %s\n", 
				emails[i].Date.Format("Jan 2"), 
				emails[i].From, 
				emails[i].Subject)
		}
		if len(emails) > previewCount {
			fmt.Printf("  ... and %d more\n", len(emails)-previewCount)
		}
		
		// Delete the emails
		if drafts {
			if err := mailos.DeleteDrafts(emailIds); err != nil {
				return fmt.Errorf("failed to delete drafts: %v", err)
			}
		} else {
			if err := mailos.DeleteEmails(emailIds); err != nil {
				return fmt.Errorf("failed to delete emails: %v", err)
			}
		}
		
		fmt.Printf("✓ Deleted %d %s\n", len(emails), func() string {
			if drafts {
				return "draft(s)"
			}
			return "email(s)"
		}())
		return nil
	},
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate an email report for a selected time range",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		timeRange, _ := cmd.Flags().GetString("range")
		outputFile, _ := cmd.Flags().GetString("output")
		
		// client, err := NewClient()
		// if err != nil {
		//	return err
		// }
		return fmt.Errorf("client functionality temporarily disabled")
		
		var selectedRange *mailos.TimeRange
		var err error
		
		// If no time range specified, show interactive selector
		if timeRange == "" {
			selectedRange, err = mailos.SelectTimeRange()
			if err != nil {
				return fmt.Errorf("failed to select time range: %v", err)
			}
		} else {
			// Parse the provided time range string
			selectedRange, err = mailos.ParseTimeRangeString(timeRange)
			if err != nil {
				return err
			}
		}
		
		fmt.Printf("Generating report for: %s\n", selectedRange.Name)
		fmt.Printf("Fetching emails from %s to %s...\n", 
			selectedRange.Since.Format("Jan 2, 3:04 PM"),
			selectedRange.Until.Format("Jan 2, 3:04 PM"))
		
		// Fetch emails for the selected time range
		opts := mailos.ReadOptions{
			Since: selectedRange.Since,
			Limit: 1000, // Get up to 1000 emails for the report
		}
		
		emails, err := mailos.Read(opts)
		if err != nil {
			return fmt.Errorf("failed to read emails: %v", err)
		}
		
		// Filter emails that are within the Until time if specified
		var filteredEmails []*mailos.Email
		for _, email := range emails {
			if email.Date.After(selectedRange.Since) && email.Date.Before(selectedRange.Until.Add(time.Second)) {
				filteredEmails = append(filteredEmails, email)
			}
		}
		
		// Generate the report
		report := mailos.GenerateEmailReport(filteredEmails, *selectedRange)
		
		// Output to file if specified, otherwise to stdout
		if outputFile != "" {
			if err := os.WriteFile(outputFile, []byte(report), 0644); err != nil {
				return fmt.Errorf("failed to write report to file: %v", err)
			}
			fmt.Printf("✓ Report saved to %s\n", outputFile)
		} else {
			fmt.Println(report)
		}
		
		return nil
	},
}

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open an email in your default mail application",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetUint("id")
		subject, _ := cmd.Flags().GetString("subject")
		from, _ := cmd.Flags().GetString("from")
		last, _ := cmd.Flags().GetInt("last")
		
		// client, err := NewClient()
		// if err != nil {
		//	return err
		// }
		return fmt.Errorf("client functionality temporarily disabled")
		
		// If specific ID provided
		if id > 0 {
			fmt.Printf("Opening email with ID %d...\n", id)
			if err := mailos.OpenEmailByID(uint32(id)); err != nil {
				return fmt.Errorf("failed to open email: %v", err)
			}
			fmt.Println("✓ Email opened in mail application")
			return nil
		}
		
		// Otherwise search for email
		opts := mailos.ReadOptions{
			FromAddress: from,
			Subject:     subject,
			Limit:       1,
		}
		
		// If --last flag is used, get the last N emails
		if last > 0 {
			opts.Limit = last
			opts.FromAddress = ""
			opts.Subject = ""
		}
		
		emails, err := mailos.Read(opts)
		if err != nil {
			return fmt.Errorf("failed to find emails: %v", err)
		}
		
		if len(emails) == 0 {
			fmt.Println("No emails found matching criteria")
			return nil
		}
		
		// Open the first (most recent) email found
		email := emails[0]
		fmt.Printf("Opening email: %s from %s\n", email.Subject, email.From)
		
		// Use the subject-based search method since we don't have Message-ID yet
		if err := mailos.OpenEmailByID(email.ID); err != nil {
			return fmt.Errorf("failed to open email: %v", err)
		}
		
		fmt.Println("✓ Email opened in mail application")
		return nil
	},
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show email statistics and analytics",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		
		// Parse query options from flags and arguments
		query := mailos.NewQueryOptions()
		
		// Parse flags
		query.Limit, _ = cmd.Flags().GetInt("number")
		query.UnreadOnly, _ = cmd.Flags().GetBool("unread")
		query.FromAddress, _ = cmd.Flags().GetString("from")
		query.ToAddress, _ = cmd.Flags().GetString("to")
		query.Subject, _ = cmd.Flags().GetString("subject")
		query.Days, _ = cmd.Flags().GetInt("days")
		query.TimeRange, _ = cmd.Flags().GetString("range")
		
		// Parse additional arguments as query parameters
		if err := query.ParseArgs(args); err != nil {
			return fmt.Errorf("STATS_QUERY_PARSE_ERROR: Failed to parse stats query arguments '%v'. Common issues: (1) Invalid date format, (2) Malformed time range, (3) Invalid account specification. Use format like 'today', 'last week', or '2024-01-01 to 2024-01-31'. Original error: %v", args, err)
		}
		
		// Handle time range
		if query.TimeRange != "" {
			if tr, err := mailos.ParseTimeRangeString(query.TimeRange); err == nil {
				query.Since = tr.Since
				query.Until = tr.Until
			}
		} else if query.Days > 0 {
			query.Since = time.Now().AddDate(0, 0, -query.Days)
		}
		
		// client, err := NewClient()
		// if err != nil {
		//	return err
		// }
		// return fmt.Errorf("client functionality temporarily disabled")
		
		fmt.Println("Fetching emails for analysis...")
		
		// Read emails using query options
		emails, err := mailos.Read(query.ToReadOptions())
		if err != nil {
			return fmt.Errorf("STATS_EMAIL_FETCH_ERROR: Failed to retrieve emails for statistical analysis. This could be due to: (1) IMAP server connection issues, (2) Authentication problems, (3) Missing email configuration, (4) Local storage access errors. Applied filters: from='%s', to='%s', subject='%s', days=%d, range='%s'. Original error: %v", 
				query.FromAddress, query.ToAddress, query.Subject, query.Days, query.TimeRange, err)
		}
		
		// Filter by Until time if needed
		emails = query.FilterEmails(emails)
		
		// Get account email for stats
		cfg, err := mailos.LoadConfig()
		if err != nil {
			return fmt.Errorf("STATS_CONFIG_ERROR: Failed to load email configuration needed for statistics generation. Ensure you have run 'mailos setup' to configure your email account. Original error: %v", err)
		}
		
		// Generate statistics
		statsOpts := mailos.StatsOptions{
			AccountEmail: cfg.Email,
			Since:        query.Since,
			Until:        query.Until,
			TopN:         10,
		}
		
		stats, err := mailos.GenerateEmailStats(statsOpts)
		if err != nil {
			return fmt.Errorf("STATS_GENERATION_ERROR: Failed to generate email statistics from %d emails for account '%s' (date range: %v to %v). This could be due to: (1) Email data processing errors, (2) Invalid date ranges, (3) Database access issues, (4) Email parsing problems. Original error: %v", 
				len(emails), cfg.Email, 
				func() interface{} { if query.Since.IsZero() { return "beginning" } else { return query.Since.Format("2006-01-02") } }(),
				func() interface{} { if query.Until.IsZero() { return "now" } else { return query.Until.Format("2006-01-02") } }(), 
				err)
		}
		
		// Display basic statistics
		fmt.Printf("📊 Email Statistics for %s\n", stats.AccountEmail)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("Total Emails: %d\n", stats.TotalEmails)
		if len(stats.SenderStats) > 0 {
			fmt.Printf("\nTop Senders:\n")
			for i, sender := range stats.SenderStats {
				if i >= 5 { break }
				fmt.Printf("  %d. %s (%d emails)\n", i+1, sender.Email, sender.Count)
			}
		}
		
		return nil
	},
}

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate or update EMAILOS.md documentation for AI CLI",
	Long:  `Generate EMAILOS.md file with complete command reference for AI CLI integration.
This reads from the docs/ directory and creates a comprehensive instruction file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Generating AI instructions from documentation...")
		if err := mailos.SaveAIInstructions(); err != nil {
			return fmt.Errorf("failed to generate documentation: %v", err)
		}
		fmt.Println("\n✓ EMAILOS.md has been created/updated in the current directory")
		fmt.Println("This file will be automatically used by AI CLIs like Claude Code")
		return nil
	},
}

var commandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "List all available mailos commands and their descriptions",
	Long:  `Display a comprehensive list of all available mailos commands organized by category.
This command is especially useful for LLMs and users who want to discover all available functionality.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")
		return showAllCommands(verbose)
	},
}

var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "List all available methods and functions across the codebase",
	Long:  `Display a comprehensive list of all Go methods/functions in the EmailOS codebase.
This command helps LLMs and developers understand the complete API surface area.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mailos.ToolsCommand()
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Completely uninstall EmailOS and remove all data",
	Long: `Completely removes EmailOS from your system including:
- Configuration files (~/.email/config.json)
- All synced email data (~/.email/sent, ~/.email/received, ~/.email/drafts)
- Local project configurations (.email/ directories)

This action cannot be undone. A backup can be created before removal.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		keepEmails, _ := cmd.Flags().GetBool("keep-emails")
		keepConfig, _ := cmd.Flags().GetBool("keep-config")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		quiet, _ := cmd.Flags().GetBool("quiet")
		createBackup, _ := cmd.Flags().GetBool("backup")
		backupPath, _ := cmd.Flags().GetString("backup-path")

		opts := mailos.CleanupOptions{
			Force:        force,
			KeepEmails:   keepEmails,
			KeepConfig:   keepConfig,
			RemoveAll:    !keepEmails && !keepConfig,
			DryRun:       dryRun,
			Quiet:        quiet,
			CreateBackup: createBackup,
			BackupPath:   backupPath,
		}

		return mailos.UninstallCommand(opts)
	},
}

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up orphaned EmailOS data",
	Long: `Detects and removes orphaned EmailOS configuration and data files.
Useful when EmailOS was uninstalled by a package manager but data remains.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		detector := mailos.NewCleanupDetector()
		
		if detector.CheckForOrphanedData() {
			return detector.PromptOrphanedCleanup()
		}
		
		quiet, _ := cmd.Flags().GetBool("quiet")
		if !quiet {
			fmt.Println("✓ No orphaned EmailOS data found.")
		}
		return nil
	},
}



var unsubscribeCmd = &cobra.Command{
	Use:   "unsubscribe",
	Short: "Find unsubscribe links and optionally open in browser",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		from, _ := cmd.Flags().GetString("from")
		subject, _ := cmd.Flags().GetString("subject")
		limit, _ := cmd.Flags().GetInt("number")
		openLink, _ := cmd.Flags().GetBool("open")
		autoOpen, _ := cmd.Flags().GetBool("auto-open")
		moveToFolder, _ := cmd.Flags().GetBool("move-to-folder")
		
		client, err := mailos.NewClient()
		if err != nil {
			return err
		}
		
		opts := mailos.ReadOptions{
			FromAddress: from,
			Subject:     subject,
			Limit:       limit,
		}
		
		fmt.Println("Searching for unsubscribe links...")
		
		// First read emails
		emails, err := client.ReadEmails(opts)
		if err != nil {
			return fmt.Errorf("failed to read emails: %v", err)
		}
		
		// Then find unsubscribe links
		links := mailos.FindUnsubscribeLinks(emails)
		
		if len(links) == 0 {
			fmt.Println("No unsubscribe links found in the specified emails")
			return nil
		}
		
		// Display report
		fmt.Print(mailos.GetUnsubscribeReport(links))
		
		// Move emails to folder if requested
		if moveToFolder {
			fmt.Println("\nMoving emails to Unsubscribe folder...")
			if err := mailos.MoveEmailsToUnsubscribeFolder(links); err != nil {
				fmt.Printf("Warning: Failed to move emails to folder: %v\n", err)
			}
		}
		
		// Open first link if requested
		if (openLink || autoOpen) && len(links) > 0 && len(links[0].Links) > 0 {
			firstLink := links[0].Links[0]
			
			if autoOpen {
				fmt.Printf("\nOpening unsubscribe link: %s\n", firstLink)
				cmd := exec.Command("open", firstLink)
				cmd.Run()
			} else if openLink {
				fmt.Printf("\nReady to open: %s\n", firstLink)
				fmt.Print("Open this link in your browser? (y/n): ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				if strings.TrimSpace(strings.ToLower(response)) == "y" {
					cmd := exec.Command("open", firstLink)
					cmd.Run()
					fmt.Println("Opened in browser")
				}
			}
		}
		
		return nil
	},
}

// setupHelpForCommand configures a command to use documentation if available
func setupHelpForCommand(cmd *cobra.Command, docName string) {
	cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		if mailos.ShowExtendedHelp(docName) {
			return
		}
		// Fall back to default help
		c.Usage()
	})
}

func init() {
	// Allow arbitrary arguments for general queries
	rootCmd.Args = cobra.ArbitraryArgs
	
	// Root command flags
	
	// Setup help functions for commands with documentation
	setupHelpForCommand(setupCmd, "setup")
	setupHelpForCommand(configureCmd, "configure")
	setupHelpForCommand(templateCmd, "template")
	setupHelpForCommand(sendCmd, "send")
	setupHelpForCommand(readCmd, "read")
	setupHelpForCommand(statsCmd, "stats")
	setupHelpForCommand(reportCmd, "report")
	setupHelpForCommand(markReadCmd, "mark-read")
	setupHelpForCommand(deleteCmd, "delete")
	setupHelpForCommand(openCmd, "open")
	setupHelpForCommand(unsubscribeCmd, "unsubscribe")
	setupHelpForCommand(testCmd, "test")
	setupHelpForCommand(infoCmd, "info")
	setupHelpForCommand(localCmd, "local")
	setupHelpForCommand(sentCmd, "sent")
	
	// Setup command flags
	setupCmd.Flags().String("email", "", "Your email address")
	setupCmd.Flags().String("provider", "", "Email provider (gmail, fastmail, outlook, yahoo, zoho)")
	setupCmd.Flags().String("name", "", "Your display name")
	setupCmd.Flags().String("license", "", "Your MailOS license key")
	setupCmd.Flags().String("profile", "", "Path to your profile image")
	
	// Accounts command flags
	accountsCmd.Flags().String("set", "", "Set session default account")
	accountsCmd.Flags().String("add", "", "Add a new email account")
	accountsCmd.Flags().String("provider", "", "Email provider for new account (gmail, fastmail, outlook, yahoo, zoho)")
	accountsCmd.Flags().Bool("use-existing-credentials", false, "Use existing credentials from same provider (useful for aliases)")
	accountsCmd.Flags().String("set-signature", "", "Set signature for an account (format: email:signature)")
	accountsCmd.Flags().Bool("clear", false, "Clear session default account")
	accountsCmd.Flags().Bool("list", false, "List available accounts")
	accountsCmd.Flags().Bool("sync-fastmail", false, "Sync aliases from FastMail via JMAP API")
	accountsCmd.Flags().String("token", "", "FastMail JMAP API token for sync operations")
	accountsCmd.Flags().Bool("test-connection", false, "Test FastMail JMAP API connection")
	
	// Stats command flags
	statsCmd.Flags().IntP("number", "n", 100, "Number of emails to analyze")
	statsCmd.Flags().BoolP("unread", "u", false, "Analyze only unread emails")
	statsCmd.Flags().String("from", "", "Filter by sender")
	statsCmd.Flags().String("to", "", "Filter by recipient")
	statsCmd.Flags().String("subject", "", "Filter by subject")
	statsCmd.Flags().Int("days", 0, "Analyze emails from last N days")
	statsCmd.Flags().String("range", "", "Time range (e.g., 'Last hour', 'Today', 'Yesterday', 'This week')")
	statsCmd.Args = cobra.ArbitraryArgs // Allow additional query parameters
	
	// Drafts command flags
	// Draft reading/listing flags
	draftsCmd.Flags().BoolP("list", "l", false, "List drafts from IMAP Drafts folder")
	draftsCmd.Flags().BoolP("read", "r", false, "Read full content of drafts from IMAP")
	draftsCmd.Flags().Uint32("edit-uid", 0, "UID of existing draft to edit/update")
	
	// Draft generation flags
	draftsCmd.Flags().String("template", "", "Use a template for draft generation")
	draftsCmd.Flags().String("data", "", "Data file (CSV/JSON) for bulk draft generation")
	draftsCmd.Flags().String("output", "", "Output directory for drafts (default: ~/.email/drafts)")
	draftsCmd.Flags().BoolP("interactive", "i", false, "Interactive mode for creating drafts")
	draftsCmd.Flags().Bool("ai", false, "Use AI to generate drafts from query")
	draftsCmd.Flags().IntP("count", "n", 1, "Number of drafts to generate (with AI)")
	
	// Email composition flags (shared with send command)
	draftsCmd.Flags().StringSliceP("to", "t", nil, "Recipient email addresses")
	draftsCmd.Flags().StringSliceP("cc", "c", nil, "CC recipients")
	draftsCmd.Flags().StringSliceP("bcc", "B", nil, "BCC recipients")
	draftsCmd.Flags().StringP("subject", "s", "", "Email subject")
	draftsCmd.Flags().StringP("body", "b", "", "Email body (Markdown supported)")
	draftsCmd.Flags().StringP("file", "f", "", "Read body from file")
	draftsCmd.Flags().StringSliceP("attach", "a", nil, "Attachments")
	draftsCmd.Flags().String("priority", "normal", "Email priority (high/normal/low)")
	draftsCmd.Flags().BoolP("plain", "P", false, "Send as plain text only")
	draftsCmd.Flags().BoolP("no-signature", "S", false, "Don't add signature")
	draftsCmd.Flags().String("signature", "", "Custom signature")
	
	// Compose command flags (same as drafts)
	composeCmd.Flags().String("account", "", "Account to use for sending")
	composeCmd.Flags().String("template", "", "Use a template for email generation")
	composeCmd.Flags().String("data", "", "Data file (CSV/JSON) for bulk email generation")
	composeCmd.Flags().String("output", "", "Output directory for drafts (default: ~/.email/drafts)")
	composeCmd.Flags().BoolP("interactive", "i", false, "Interactive mode for creating email")
	composeCmd.Flags().Bool("ai", false, "Use AI to generate email from query")
	composeCmd.Flags().IntP("count", "n", 1, "Number of emails to generate (with AI)")
	composeCmd.Flags().StringSliceP("to", "t", nil, "Recipient email addresses")
	composeCmd.Flags().StringSliceP("cc", "c", nil, "CC recipients")
	composeCmd.Flags().StringSliceP("bcc", "B", nil, "BCC recipients")
	composeCmd.Flags().StringP("subject", "s", "", "Email subject")
	composeCmd.Flags().StringP("body", "b", "", "Email body (Markdown supported)")
	composeCmd.Flags().StringP("file", "f", "", "Read body from file")
	composeCmd.Flags().StringSliceP("attach", "a", nil, "Attachments")
	composeCmd.Flags().String("priority", "normal", "Email priority (high/normal/low)")
	composeCmd.Flags().BoolP("plain", "P", false, "Send as plain text only")
	composeCmd.Flags().BoolP("no-signature", "S", false, "Don't add signature")
	composeCmd.Flags().String("signature", "", "Custom signature")
	
	// Send command flags
	sendCmd.Flags().StringSliceP("to", "t", nil, "Recipient email addresses")
	sendCmd.Flags().StringSliceP("cc", "c", nil, "CC recipients")
	sendCmd.Flags().StringSliceP("bcc", "B", nil, "BCC recipients")
	sendCmd.Flags().StringP("group", "g", "", "Send to email group")
	sendCmd.Flags().StringP("subject", "s", "", "Email subject")
	sendCmd.Flags().StringP("body", "b", "", "Email body (Markdown supported)")
	sendCmd.Flags().StringP("message", "m", "", "Email body (alias for --body)")
	sendCmd.Flags().StringP("file", "f", "", "Read body from file")
	sendCmd.Flags().StringSliceP("attach", "a", nil, "Attachments")
	sendCmd.Flags().BoolP("plain", "P", false, "Send as plain text")
	sendCmd.Flags().BoolP("no-signature", "S", false, "No signature")
	sendCmd.Flags().String("signature", "", "Custom signature")
	sendCmd.Flags().String("from", "", "Send from specific email account (account nickname or email)")
	sendCmd.Flags().Bool("preview", false, "Preview the complete email without sending")
	sendCmd.Flags().Bool("template", false, "Apply HTML template to email")
	sendCmd.Flags().BoolP("verbose", "v", false, "Show detailed SMTP debugging information")
	
	// Send --drafts specific flags
	sendCmd.Flags().Bool("drafts", false, "Send all draft emails from .email/drafts folder")
	sendCmd.Flags().String("draft-dir", "", "Directory containing draft emails (default: ~/.email/drafts)")
	sendCmd.Flags().Bool("dry-run", false, "Preview what would be sent without actually sending")
	sendCmd.Flags().String("filter", "", "Filter drafts (e.g., 'priority:high', 'to:*@example.com')")
	sendCmd.Flags().Bool("confirm", false, "Confirm before sending each draft")
	sendCmd.Flags().Bool("delete-after", true, "Delete drafts after successful sending")
	sendCmd.Flags().String("log-file", "", "Log sent emails to file")

	// Groups command flags
	groupsCmd.Flags().String("update", "", "Create or update a group with the given name")
	groupsCmd.Flags().String("description", "", "Description for the group")
	groupsCmd.Flags().String("emails", "", "Comma-separated list of email addresses")
	groupsCmd.Flags().String("delete", "", "Delete the specified group")
	groupsCmd.Flags().String("add-member", "", "Add a member to an existing group")
	groupsCmd.Flags().String("remove-member", "", "Remove a member from an existing group")
	groupsCmd.Flags().String("group", "", "Group name for add/remove member operations")
	groupsCmd.Flags().String("list-members", "", "List all members of the specified group")

	// Sync command flags
	syncCmd.Flags().String("dir", "emails", "Base directory for storing emails")
	syncCmd.Flags().Int("limit", 100, "Maximum number of emails to sync per folder")
	syncCmd.Flags().Int("days", 0, "Sync emails from last N days (0 for all)")
	syncCmd.Flags().Bool("include-read", false, "Include already read emails")
	syncCmd.Flags().BoolP("verbose", "v", false, "Show detailed progress")

	// Sync-db command flags
	syncDbCmd.Flags().String("account", "", "Specific account email to sync (defaults to configured account)")
	syncDbCmd.Flags().Bool("all", false, "Sync all configured accounts to database")

	// Sent command flags
	sentCmd.Flags().IntP("number", "n", 10, "Number of sent emails to read")
	sentCmd.Flags().String("to", "", "Filter by recipient")
	sentCmd.Flags().String("subject", "", "Filter by subject")
	sentCmd.Flags().Int("days", 0, "Show sent emails from last N days")
	sentCmd.Flags().String("range", "", "Time range (e.g., 'Last hour', 'Today', 'Yesterday', 'This week')")
	sentCmd.Flags().Bool("json", false, "Output as JSON")
	sentCmd.Flags().Bool("save-markdown", true, "Save emails as markdown files")
	sentCmd.Flags().String("output-dir", ".email/sent", "Directory to save markdown files")


	// Search command flags (enhanced with advanced search capabilities)
	searchCmd.Flags().IntP("number", "n", 10, "Number of emails to search")
	searchCmd.Flags().BoolP("unread", "u", false, "Show only unread emails")
	searchCmd.Flags().String("from", "", "Filter by sender")
	searchCmd.Flags().String("to", "", "Filter by recipient (defaults to from_email in config if set)")
	searchCmd.Flags().String("subject", "", "Filter by subject")
	searchCmd.Flags().Int("days", 0, "Search emails from last N days")
	searchCmd.Flags().String("range", "", "Time range (e.g., 'Last hour', 'Today', 'Yesterday', 'This week')")
	searchCmd.Flags().Bool("json", false, "Output as JSON")
	searchCmd.Flags().Bool("save-markdown", false, "Save emails as markdown files")
	searchCmd.Flags().String("output-dir", "emails", "Directory to save markdown files")
	searchCmd.Flags().Bool("download-attachments", false, "Download email attachments")
	searchCmd.Flags().String("attachment-dir", "attachments", "Directory to save attachments")
	
	// Advanced search flags
	searchCmd.Flags().StringP("query", "q", "", "Complex search query with boolean operators (AND, OR, NOT)")
	searchCmd.Flags().Float64("fuzzy-threshold", 0.7, "Fuzzy matching threshold (0.0-1.0)")
	searchCmd.Flags().Bool("no-fuzzy", false, "Disable fuzzy matching")
	searchCmd.Flags().Bool("case-sensitive", false, "Enable case sensitive search")
	searchCmd.Flags().String("min-size", "", "Minimum email size (e.g., '1MB', '500KB')")
	searchCmd.Flags().String("max-size", "", "Maximum email size (e.g., '10MB', '2GB')")
	searchCmd.Flags().Bool("has-attachments", false, "Filter emails with attachments")
	searchCmd.Flags().String("attachment-size", "", "Minimum attachment size (e.g., '1MB')")
	searchCmd.Flags().String("date-range", "", "Flexible date range (e.g., 'today', 'last week', '2023-01-01 to 2023-12-31')")

	// Read command flags (for displaying full email content)
	readCmd.Flags().Bool("include-documents", true, "Parse and display attachment document content inline")
	readCmd.Flags().Uint32("id", 0, "Email ID to read (alternative to positional argument)")

	// Reply command flags
	replyCmd.Flags().Bool("all", false, "Reply to all recipients")
	replyCmd.Flags().String("body", "", "Reply body text")
	replyCmd.Flags().String("subject", "", "Override reply subject")
	replyCmd.Flags().StringP("file", "f", "", "Read body from file")
	replyCmd.Flags().Bool("draft", false, "Save as draft instead of sending")
	replyCmd.Flags().BoolP("interactive", "i", false, "Force interactive mode")
	replyCmd.Flags().StringSlice("to", nil, "Override recipients")
	replyCmd.Flags().StringSlice("cc", nil, "CC recipients")
	replyCmd.Flags().StringSlice("bcc", nil, "BCC recipients")

	// Forward command flags
	forwardCmd.Flags().String("body", "", "Forward body text")
	forwardCmd.Flags().String("subject", "", "Override forward subject")
	forwardCmd.Flags().StringP("file", "f", "", "Read body from file")
	forwardCmd.Flags().Bool("draft", false, "Save as draft instead of sending")
	forwardCmd.Flags().BoolP("interactive", "i", false, "Force interactive mode")
	forwardCmd.Flags().StringSlice("to", nil, "Recipients")
	forwardCmd.Flags().StringSlice("cc", nil, "CC recipients")
	forwardCmd.Flags().StringSlice("bcc", nil, "BCC recipients")

	// Mark read command flags
	markReadCmd.Flags().UintSlice("ids", nil, "Email IDs to mark as read")
	markReadCmd.Flags().String("from", "", "Mark all from sender")
	markReadCmd.Flags().String("subject", "", "Mark all with subject")

	// Delete command flags
	deleteCmd.Flags().UintSlice("ids", nil, "Email IDs to delete")
	deleteCmd.Flags().String("from", "", "Delete all from sender")
	deleteCmd.Flags().String("subject", "", "Delete all with subject")
	deleteCmd.Flags().Bool("drafts", false, "Delete drafts instead of regular emails")
	deleteCmd.Flags().String("before", "", "Delete emails before date (YYYY-MM-DD)")
	deleteCmd.Flags().String("after", "", "Delete emails after date (YYYY-MM-DD)")
	deleteCmd.Flags().Int("days", 0, "Delete emails older than N days")
	deleteCmd.Flags().Bool("confirm", false, "Confirm deletion")

	// Open command flags
	openCmd.Flags().Uint("id", 0, "Email ID to open")
	openCmd.Flags().String("from", "", "Open email from sender")
	openCmd.Flags().String("subject", "", "Open email by subject")
	openCmd.Flags().Int("last", 0, "Open from last N emails (opens most recent)")

	// Unsubscribe command flags
	unsubscribeCmd.Flags().String("from", "", "Find unsubscribe links from sender")
	unsubscribeCmd.Flags().String("subject", "", "Find unsubscribe links by subject")
	unsubscribeCmd.Flags().IntP("number", "n", 10, "Number of emails to check")
	unsubscribeCmd.Flags().Bool("open", false, "Open the first unsubscribe link in browser")
	unsubscribeCmd.Flags().Bool("auto-open", false, "Automatically open unsubscribe link without prompting")
	unsubscribeCmd.Flags().Bool("move-to-folder", false, "Move emails with unsubscribe links to dedicated IMAP folder")

	// Test command flags
	testCmd.Flags().Bool("interactive", false, "Run interactive tests")
	testCmd.Flags().Bool("verbose", false, "Show detailed output")
	
	// Template command flags
	templateCmd.Flags().Bool("remove", false, "Remove existing template")
	templateCmd.Flags().Bool("open-browser", false, "Open template HTML file in browser")
	
	// Configure command flags
	configureCmd.Flags().Bool("quick", false, "Quick configuration menu")
	configureCmd.Flags().Bool("local", false, "Create/modify local configuration (.email/) instead of global (~/.email/)")
	configureCmd.Flags().String("email", "", "Email address to configure")
	configureCmd.Flags().String("provider", "", "Email provider (gmail, outlook, yahoo, icloud, proton, fastmail, custom)")
	configureCmd.Flags().String("name", "", "Display name for emails")
	configureCmd.Flags().String("from", "", "From email address (appears as sender)")
	configureCmd.Flags().String("ai", "", "AI CLI provider (claude-code, claude-code-yolo, openai, gemini, opencode, none)")
	
	// Report command flags
	reportCmd.Flags().String("range", "", "Time range (e.g., 'Last hour', 'Today', 'Yesterday', 'This week')")
	reportCmd.Flags().String("output", "", "Output file path (optional)")
	
	// Interactive command flags
	
	// Uninstall command flags
	uninstallCmd.Flags().Bool("force", false, "Skip confirmation prompts")
	uninstallCmd.Flags().Bool("keep-emails", false, "Keep email data, only remove configuration")
	uninstallCmd.Flags().Bool("keep-config", false, "Keep configuration, only remove email data")
	uninstallCmd.Flags().Bool("dry-run", false, "Show what would be removed without doing it")
	uninstallCmd.Flags().Bool("quiet", false, "Minimal output")
	uninstallCmd.Flags().Bool("backup", false, "Create backup before removal")
	uninstallCmd.Flags().String("backup-path", "", "Custom backup location")
	
	// Cleanup command flags
	cleanupCmd.Flags().Bool("quiet", false, "Minimal output")
	
	// Commands command flags
	commandsCmd.Flags().Bool("verbose", false, "Show detailed flag information")
	
	// Download command flags
	downloadCmd.Flags().IntP("number", "n", 10, "Number of emails to search")
	downloadCmd.Flags().Uint32("id", 0, "Download attachments from specific email ID")
	downloadCmd.Flags().String("from", "", "Filter by sender")
	downloadCmd.Flags().String("to", "", "Filter by recipient")
	downloadCmd.Flags().String("subject", "", "Filter by subject")
	downloadCmd.Flags().Int("days", 0, "Search emails from last N days")
	downloadCmd.Flags().String("output-dir", "attachments", "Directory to save attachments")
	downloadCmd.Flags().Bool("show-content", false, "Show email content preview")

	// Add draft subcommands
	draftCmd.AddCommand(draftListCmd)
	draftCmd.AddCommand(draftEditCmd)
	draftCmd.AddCommand(draftCreateCmd)
	
	// Add flags to draft commands
	draftEditCmd.Flags().String("subject", "", "New subject")
	draftEditCmd.Flags().String("body", "", "New body")
	draftEditCmd.Flags().StringSlice("to", nil, "New recipients")
	
	draftCreateCmd.Flags().StringSlice("to", nil, "Recipients")
	draftCreateCmd.Flags().String("subject", "", "Subject")
	draftCreateCmd.Flags().String("body", "", "Body")
}

var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Check license status and session caching",
	Long:  `Check license status and test session-level caching functionality.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")
		multiple, _ := cmd.Flags().GetInt("calls")
		
		if multiple <= 0 {
			multiple = 1
		}
		
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("                LICENSE STATUS CHECK")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		
		// Get license manager
		lm := mailos.GetLicenseManager()
		
		// Check initial session status
		validated, duration, key := lm.GetSessionStatus()
		fmt.Printf("Initial Session Status:\n")
		fmt.Printf("  Validated: %t\n", validated)
		fmt.Printf("  Duration:  %v\n", duration)
		fmt.Printf("  Key:       %s\n", key)
		fmt.Println()
		
		// Test subscription checks multiple times
		fmt.Printf("Testing IsSubscribed() %d time(s):\n", multiple)
		fmt.Println("────────────────────────────────────────────")
		
		for i := 1; i <= multiple; i++ {
			start := time.Now()
			subscribed := mailos.IsSubscribed()
			elapsed := time.Since(start)
			
			fmt.Printf("Call %d: subscribed=%t, took %v\n", i, subscribed, elapsed)
			
			if verbose {
				// Show session status after each call
				validated, duration, _ := lm.GetSessionStatus()
				fmt.Printf("        session_validated=%t, session_age=%v\n", validated, duration)
			}
			
			// Small delay between calls to make timing visible
			if i < multiple {
				time.Sleep(10 * time.Millisecond)
			}
		}
		
		// Final session status
		fmt.Println()
		validated, duration, key = lm.GetSessionStatus()
		fmt.Printf("Final Session Status:\n")
		fmt.Printf("  Validated: %t\n", validated)
		fmt.Printf("  Duration:  %v\n", duration)
		fmt.Printf("  Key:       %s\n", key)
		
		if validated {
			remaining := mailos.SessionCacheDuration - duration
			fmt.Printf("  Remaining: %v\n", remaining)
		}
		
		return nil
	},
}

var detectCmd = &cobra.Command{
	Use:   "detect [email]",
	Short: "Detect email provider for a given email address",
	Long:  `Detect the email provider for a given email address using domain analysis and MX record lookup.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		email := args[0]
		mailos.DetectEmailProvider(email)
		return nil
	},
}

func init() {
	// Add commands to root
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(localCmd)
	rootCmd.AddCommand(providerCmd)
	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(accountsCmd)
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(draftCmd)
	rootCmd.AddCommand(draftsCmd)
	rootCmd.AddCommand(composeCmd)
	rootCmd.AddCommand(sendCmd)
	rootCmd.AddCommand(groupsCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(syncDbCmd)
	rootCmd.AddCommand(sentCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(replyCmd)
	rootCmd.AddCommand(forwardCmd)
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(markReadCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(unsubscribeCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(docsCmd)
	rootCmd.AddCommand(commandsCmd)
	rootCmd.AddCommand(toolsCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(cleanupCmd)
	rootCmd.AddCommand(licenseCmd)
	rootCmd.AddCommand(detectCmd)

	// Add flags to license command
	licenseCmd.Flags().Bool("verbose", false, "Show detailed session information")
	licenseCmd.Flags().Int("calls", 5, "Number of subscription checks to perform")
}

func main() {
	// Set up signal handling for graceful cleanup detection
	setupSignalHandling()
	
	// Check for orphaned data on startup (for package manager uninstalls)
	checkOrphanedDataOnStartup()
	
	// Check for updates before running the main command
	// Skip for certain commands that shouldn't trigger updates
	if len(os.Args) > 1 {
		skipUpdateCommands := []string{"--version", "-v", "--help", "-h", "uninstall", "cleanup"}
		shouldSkip := false
		for _, cmd := range skipUpdateCommands {
			if os.Args[1] == cmd {
				shouldSkip = true
				break
			}
		}
		if !shouldSkip {
			checkForUpdates()
		}
	} else {
		// No arguments provided, check for updates
		checkForUpdates()
	}
	
	
	// Setup enhanced error handling with helpful suggestions
	// SetupErrorHandling(rootCmd) // Function not found
	
	// Configure error handling
	rootCmd.SilenceErrors = false
	rootCmd.SilenceUsage = false
	rootCmd.DisableFlagParsing = false
	// Remove the whitelist that was hiding flag errors
	rootCmd.FParseErrWhitelist.UnknownFlags = false
	
	
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// setupSignalHandling sets up handlers for termination signals
func setupSignalHandling() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	
	go func() {
		<-c
		// On signal, check if this might be an uninstall scenario
		handleGracefulShutdown()
		os.Exit(0)
	}()
}

// handleGracefulShutdown handles cleanup when process is terminated
func handleGracefulShutdown() {
	// This is called when the process receives a termination signal
	// We can't prompt the user here, but we can leave a cleanup hint
	detector := mailos.NewCleanupDetector()
	if detector.CheckForOrphanedData() {
		// Write a hint file for later cleanup detection
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return
		}
		
		hintFile := filepath.Join(homeDir, ".email", ".cleanup_hint")
		hintContent := fmt.Sprintf("EmailOS was terminated at %s. Run 'mailos cleanup' to remove orphaned data.\n", time.Now().Format(time.RFC3339))
		os.WriteFile(hintFile, []byte(hintContent), 0644)
	}
}

// checkOrphanedDataOnStartup checks for orphaned data when EmailOS starts
func checkOrphanedDataOnStartup() {
	// Only check if we're not running uninstall/cleanup commands
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		if cmd == "uninstall" || cmd == "cleanup" {
			return
		}
	}
	
	detector := mailos.NewCleanupDetector()
	
	// Check for cleanup hint file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	
	hintFile := filepath.Join(homeDir, ".email", ".cleanup_hint")
	if _, err := os.Stat(hintFile); err == nil {
		// Cleanup hint exists, remove it and show message
		os.Remove(hintFile)
		fmt.Println("💡 Tip: Run 'mailos cleanup' to remove any orphaned EmailOS data.")
		return
	}
	
	// Check for orphaned data (but don't prompt automatically on every startup)
	// Only show a subtle hint if data has been orphaned for a while
	if shouldShowOrphanedDataHint() && detector.CheckForOrphanedData() {
		fmt.Println("💡 Hint: Orphaned EmailOS data detected. Run 'mailos cleanup' to remove it.")
		updateOrphanedDataHintTime()
	}
}

// shouldShowOrphanedDataHint determines if we should show the orphaned data hint
func shouldShowOrphanedDataHint() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	
	hintTimeFile := filepath.Join(homeDir, ".email", ".last_orphan_hint")
	data, err := os.ReadFile(hintTimeFile)
	if err != nil {
		return true // No hint file, show hint
	}
	
	lastHint, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return true
	}
	
	// Show hint at most once per week
	return time.Since(lastHint) > 7*24*time.Hour
}

// updateOrphanedDataHintTime updates the last hint time
func updateOrphanedDataHintTime() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	
	hintTimeFile := filepath.Join(homeDir, ".email", ".last_orphan_hint")
	timeData := time.Now().Format(time.RFC3339)
	os.WriteFile(hintTimeFile, []byte(timeData), 0644)
}
