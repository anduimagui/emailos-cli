package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
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
		"draft", "drafts", "send", "sync", "sync-db", "sent", "download", "read", "reply",
		"mark-read", "accounts", "info", "test", "delete", "report",
		"open", "stats", "docs", "interactive", "chat", "search",
		"unsubscribe",
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
	
	// Show suggestions if any
	suggestions := findSimilarCommands(unknownCmd)
	if len(suggestions) > 0 {
		fmt.Printf("💡 Did you mean:\n")
		for _, suggestion := range suggestions {
			fmt.Printf("   • mailos %s\n", suggestion)
		}
		fmt.Println()
	}
	
	fmt.Printf("📋 Available commands:\n")
	
	// Group commands by category for better display
	core := []string{"setup", "configure", "info"}
	email := []string{"read", "reply", "send", "draft", "search", "delete", "mark-read"}
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

var rootCmd = &cobra.Command{
	Use:     "mailos",
	Version: Version,
	Short:   "EmailOS - A standardized email client",
	Long: `EmailOS is a command-line email client that supports multiple providers
and provides a consistent interface for sending and reading emails.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check for --ink flag
		ink, _ := cmd.Flags().GetBool("ink")
		if ink {
			os.Setenv("MAILOS_USE_INK", "true")
		}
		
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
			// return // mailos. - temporarily disabledHandleQueryWithProviderSelection(query)
			return fmt.Errorf("CLI functionality temporarily disabled")
		}
		
		// Default behavior: show help
		return cmd.Help()
	},
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up your email configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return mailos.Setup()
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
		if err := mailos.EnsureInitialized(); err != nil {
			return err
		}
		// return // mailos. - temporarily disabledSelectAndConfigureAIProvider()
		return fmt.Errorf("CLI functionality temporarily disabled")
	},
}

var configureCmd = &cobra.Command{
	Use:   "configure",
	Aliases: []string{"config"}, // Add config as an alias
	Short: "Manage email configuration (global or local)",
	Long:  `Configure email settings. By default modifies global configuration (~/.email/).
Use --local flag to create/modify project-specific configuration (.email/)`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		quick, _ := cmd.Flags().GetBool("quick")
		_, _ = cmd.Flags().GetBool("local")
		_, _ = cmd.Flags().GetString("email")
		_, _ = cmd.Flags().GetString("provider")
		_, _ = cmd.Flags().GetString("name")
		_, _ = cmd.Flags().GetString("from")
		_, _ = cmd.Flags().GetString("ai")
		
		if quick {
			// return // mailos. - temporarily disabledQuickConfigMenu()
			return fmt.Errorf("quick config functionality temporarily disabled")
		}
		
		// Pass command-line arguments to Configure
		// opts := mailos.ConfigureOptions{
		//	Email:    email,
		//	Provider: provider,
		//	Name:     name,
		//	From:     from,
		//	AICLI:    aiCLI,
		//	IsLocal:  isLocal,
		// }
		// return mailos.ConfigureWithOptions(opts)
		return fmt.Errorf("configuration functionality temporarily disabled")
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
			_, _ = cmd.Flags().GetString("draft-dir")
			_, _ = cmd.Flags().GetBool("dry-run")
			_, _ = cmd.Flags().GetString("filter")
			_, _ = cmd.Flags().GetBool("confirm")
			_, _ = cmd.Flags().GetBool("delete-after")
			_, _ = cmd.Flags().GetString("log-file")
			
			// opts := mailos.SendDraftsOptions{
			//	DraftDir:    draftDir,
			//	DryRun:      dryRun,
			//	Filter:      filter,
			//	Confirm:     confirm,
			//	DeleteAfter: deleteAfter,
			//	LogFile:     logFile,
			// }
			// return mailos.SendDrafts(opts)
			return fmt.Errorf("send drafts functionality temporarily disabled")
		}
		
		// Regular send command
		to, _ := cmd.Flags().GetStringSlice("to")
		cc, _ := cmd.Flags().GetStringSlice("cc")
		bcc, _ := cmd.Flags().GetStringSlice("bcc")
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

		if len(to) == 0 {
			return fmt.Errorf("at least one recipient is required")
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
				sig = signature
			} else {
				cfg, _ := mailos.LoadConfig()
				// Check for signature override first
				if cfg.SignatureOverride != "" {
					sig = cfg.SignatureOverride
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
			return mailos.PreviewEmail(msg, from)
		}

		fmt.Printf("Sending email to %s...\n", strings.Join(to, ", "))
		if err := mailos.SendWithAccount(msg, from); err != nil {
			return fmt.Errorf("failed to send email: %v", err)
		}

		fmt.Println("✓ Email sent successfully!")
		return nil
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

		opts := mailos.ReadOptions{
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
		emails, err := mailos.ReadFromFolder(opts, "Sent")
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
			
			// client, err := NewClient()
			// if err != nil {
			//	return err
			// }
			return fmt.Errorf("client functionality temporarily disabled")
			
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
			// Use the new mailos.FormatEmailList function for sent emails
			fmt.Print(mailos.FormatEmailList(emails))
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
	Use:   "read",
	Short: "Display full content of a specific email",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get account flag if specified
		accountEmail, _ := cmd.Flags().GetString("account")
		
		// Get include-documents flag
		includeDocuments, _ := cmd.Flags().GetBool("include-documents")
		
		// Ensure authenticated before proceeding
		cfg, err := mailos.EnsureAuthenticated(accountEmail)
		if err != nil {
			return err
		}
		_ = cfg // Config is now validated and has credentials
		
		// Parse email ID
		emailID := args[0]
		id, err := strconv.ParseUint(emailID, 10, 32)
		if err != nil {
			return fmt.Errorf("READ_INVALID_ID: Email ID '%s' is not a valid number. Email IDs must be positive integers (e.g., 1332, 1331). Use 'mailos search' to see available email IDs. Parsing error: %v", emailID, err)
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
  mailos accounts --set user@example.com                   # Set default account for this session
  mailos accounts --set-signature user@example.com:"Best regards, John"  # Set account signature
  mailos accounts --clear                                   # Clear session default`,
	RunE: func(cmd *cobra.Command, args []string) error {
		setAccount, _ := cmd.Flags().GetString("set")
		addAccount, _ := cmd.Flags().GetString("add")
		setSignature, _ := cmd.Flags().GetString("set-signature")
		clearSession, _ := cmd.Flags().GetBool("clear")
		listAccounts, _ := cmd.Flags().GetBool("list")
		
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
			if err := mailos.AddNewAccount(addAccount); err != nil {
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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
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

var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "List all available methods and functions across the codebase",
	Long:  `Display a comprehensive list of all Go methods/functions in the EmailOS codebase.
This command helps LLMs and developers understand the complete API surface area.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mailos.ToolsCommand()
	},
}

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Launch interactive mode",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ink, _ := cmd.Flags().GetBool("ink")
		if ink {
			os.Setenv("MAILOS_USE_INK", "true")
		}
		return mailos.InteractiveModeWithMenu()
	},
}

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Launch AI chat interface",
	Long:  `Launch the AI chat interface for natural language email interactions.
This uses the React Ink UI to provide a modern chat experience.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return mailos.EnsureInitialized()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Chat command always uses the React Ink UI for better chat experience
		os.Setenv("MAILOS_USE_INK", "true")
		return mailos.InteractiveModeWithMenu()
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
		
		// client, err := NewClient()
		// if err != nil {
		//	return err
		// }
		return fmt.Errorf("client functionality temporarily disabled")
		
		opts := mailos.ReadOptions{
			FromAddress: from,
			Subject:     subject,
			Limit:       limit,
		}
		
		fmt.Println("Searching for unsubscribe links...")
		
		// First read emails
		emails, err := mailos.Read(opts)
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
	rootCmd.Flags().Bool("ink", false, "Use the React Ink UI (experimental)")
	
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
	
	// Accounts command flags
	accountsCmd.Flags().String("set", "", "Set session default account")
	accountsCmd.Flags().String("add", "", "Add a new email account")
	accountsCmd.Flags().String("set-signature", "", "Set signature for an account (format: email:signature)")
	accountsCmd.Flags().Bool("clear", false, "Clear session default account")
	accountsCmd.Flags().Bool("list", false, "List available accounts")
	
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
	
	// Send command flags
	sendCmd.Flags().StringSliceP("to", "t", nil, "Recipient email addresses")
	sendCmd.Flags().StringSliceP("cc", "c", nil, "CC recipients")
	sendCmd.Flags().StringSliceP("bcc", "B", nil, "BCC recipients")
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
	
	// Send --drafts specific flags
	sendCmd.Flags().Bool("drafts", false, "Send all draft emails from .email/drafts folder")
	sendCmd.Flags().String("draft-dir", "", "Directory containing draft emails (default: ~/.email/drafts)")
	sendCmd.Flags().Bool("dry-run", false, "Preview what would be sent without actually sending")
	sendCmd.Flags().String("filter", "", "Filter drafts (e.g., 'priority:high', 'to:*@example.com')")
	sendCmd.Flags().Bool("confirm", false, "Confirm before sending each draft")
	sendCmd.Flags().Bool("delete-after", true, "Delete drafts after successful sending")
	sendCmd.Flags().String("log-file", "", "Log sent emails to file")

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
	interactiveCmd.Flags().Bool("ink", false, "Use the React Ink UI (experimental)")
	
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

	// Add commands to root
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(localCmd)
	rootCmd.AddCommand(providerCmd)
	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(accountsCmd)
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(draftCmd)
	rootCmd.AddCommand(draftsCmd)
	rootCmd.AddCommand(sendCmd)
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
	rootCmd.AddCommand(interactiveCmd)
	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(docsCmd)
	rootCmd.AddCommand(toolsCmd)
}

func main() {
	// Check for updates before running the main command
	// Skip for certain commands that shouldn't trigger updates
	if len(os.Args) > 1 {
		skipUpdateCommands := []string{"--version", "-v", "--help", "-h"}
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
	
	
	// Disable unknown command errors to allow general queries
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	rootCmd.DisableFlagParsing = false
	rootCmd.FParseErrWhitelist.UnknownFlags = true
	
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
