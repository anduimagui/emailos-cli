package mailos

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// CleanupOptions defines options for cleanup operations
type CleanupOptions struct {
	Force           bool // Skip confirmation prompts
	KeepEmails      bool // Keep email data but remove config
	KeepConfig      bool // Keep config but remove email data
	RemoveAll       bool // Remove everything including global config
	DryRun          bool // Show what would be removed without doing it
	Quiet           bool // Minimal output
	CreateBackup    bool // Create backup before removal
	BackupPath      string // Custom backup location
}

// CleanupReport contains information about what was cleaned up
type CleanupReport struct {
	RemovedDirs    []string
	RemovedFiles   []string
	BackupCreated  string
	EmailsRemoved  int
	ConfigRemoved  bool
	BytesFreed     int64
	Errors         []error
}

// UninstallCommand handles complete uninstallation of EmailOS
func UninstallCommand(opts CleanupOptions) error {
	// Define styles for output
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82")) // Green
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")) // Blue

	if !opts.Quiet {
		fmt.Println(headerStyle.Render("📦 EmailOS Complete Uninstallation"))
		fmt.Println(headerStyle.Render("═════════════════════════════════"))
		fmt.Println()
	}

	// Detect current installation
	installation := DetectInstallation()
	if !opts.Quiet {
		fmt.Printf("Detected installation: %s\n", installation.Method)
		if installation.BinaryPath != "" {
			fmt.Printf("Binary location: %s\n", installation.BinaryPath)
		}
		fmt.Println()
	}

	// Check what data exists
	dataExists, dataSize := CheckEmailDataExists()
	configExists := CheckConfigExists()

	if !dataExists && !configExists {
		if !opts.Quiet {
			fmt.Println(successStyle.Render("✓ No EmailOS data found to clean up."))
		}
		return nil
	}

	// Show what will be removed
	if !opts.Quiet {
		fmt.Println(warningStyle.Render("⚠️  The following will be removed:"))
		fmt.Println()
		
		if configExists {
			fmt.Println("🗂️  Configuration files:")
			homeDir, _ := os.UserHomeDir()
			configPath := filepath.Join(homeDir, ".email", "config.json")
			fmt.Printf("   • %s\n", configPath)
		}
		
		if dataExists {
			fmt.Println("📧 Email data:")
			fmt.Printf("   • All synced emails (~%s)\n", formatBytes(dataSize))
			homeDir, _ := os.UserHomeDir()
			emailDir := filepath.Join(homeDir, ".email")
			fmt.Printf("   • Directory: %s\n", emailDir)
		}

		// Check for local configs
		localConfigs := FindLocalConfigs()
		if len(localConfigs) > 0 {
			fmt.Println("📁 Local project configurations:")
			for _, config := range localConfigs {
				fmt.Printf("   • %s\n", config)
			}
		}

		fmt.Println()
		fmt.Println(errorStyle.Render("⚠️  This action cannot be undone!"))
		fmt.Println()
	}

	// Get confirmation unless force flag is set
	if !opts.Force && !opts.DryRun {
		if !ConfirmUninstall() {
			if !opts.Quiet {
				fmt.Println("Uninstallation cancelled.")
			}
			return nil
		}
	}

	// Perform cleanup
	report, err := PerformCleanup(opts)
	if err != nil {
		return fmt.Errorf("cleanup failed: %v", err)
	}

	// Display results
	if !opts.Quiet {
		DisplayCleanupReport(report, opts.DryRun)
	}

	// Attempt to remove binary (best effort)
	if !opts.DryRun {
		RemoveBinary(installation, opts.Quiet)
	}

	return nil
}

// DetectInstallation identifies how EmailOS was installed
func DetectInstallation() InstallationInfo {
	execPath, err := os.Executable()
	if err != nil {
		return InstallationInfo{Method: "unknown"}
	}

	absPath, _ := filepath.Abs(execPath)
	
	// Check for common installation methods
	if strings.Contains(absPath, "/usr/local/bin") {
		return InstallationInfo{
			Method:     "manual",
			BinaryPath: absPath,
		}
	}
	if strings.Contains(absPath, "node_modules") || strings.Contains(absPath, ".npm") {
		return InstallationInfo{
			Method:     "npm",
			BinaryPath: absPath,
		}
	}
	if strings.Contains(absPath, "homebrew") || strings.Contains(absPath, "Cellar") {
		return InstallationInfo{
			Method:     "homebrew",
			BinaryPath: absPath,
		}
	}
	
	return InstallationInfo{
		Method:     "unknown",
		BinaryPath: absPath,
	}
}

// InstallationInfo contains information about how EmailOS was installed
type InstallationInfo struct {
	Method     string // npm, homebrew, manual, unknown
	BinaryPath string
}

// CheckEmailDataExists checks if email data directory exists and returns size
func CheckEmailDataExists() (bool, int64) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, 0
	}

	emailDir := filepath.Join(homeDir, ".email")
	info, err := os.Stat(emailDir)
	if err != nil {
		return false, 0
	}

	if !info.IsDir() {
		return false, 0
	}

	// Calculate directory size
	size := calculateDirSize(emailDir)
	return true, size
}

// CheckConfigExists checks if configuration files exist
func CheckConfigExists() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	configPath := filepath.Join(homeDir, ".email", "config.json")
	_, err = os.Stat(configPath)
	return err == nil
}

// FindLocalConfigs finds all local .email directories
func FindLocalConfigs() []string {
	var configs []string
	
	// Search common project directories
	searchDirs := []string{
		".",
		filepath.Join(os.Getenv("HOME"), "projects"),
		filepath.Join(os.Getenv("HOME"), "Documents"),
		filepath.Join(os.Getenv("HOME"), "Desktop"),
	}

	for _, dir := range searchDirs {
		if dir == "" {
			continue
		}
		
		found := findEmailDirsRecursive(dir, 3) // Limit depth to 3
		configs = append(configs, found...)
	}

	return configs
}

// findEmailDirsRecursive searches for .email directories recursively with depth limit
func findEmailDirsRecursive(root string, maxDepth int) []string {
	var results []string
	
	if maxDepth <= 0 {
		return results
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return results
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		fullPath := filepath.Join(root, entry.Name())
		
		if entry.Name() == ".email" {
			// Check if it contains config.json
			configPath := filepath.Join(fullPath, "config.json")
			if _, err := os.Stat(configPath); err == nil {
				results = append(results, fullPath)
			}
		} else if !strings.HasPrefix(entry.Name(), ".") {
			// Recurse into non-hidden directories
			subResults := findEmailDirsRecursive(fullPath, maxDepth-1)
			results = append(results, subResults...)
		}
	}

	return results
}

// ConfirmUninstall prompts user for confirmation
func ConfirmUninstall() bool {
	fmt.Print("Do you want to proceed with complete uninstallation? (yes/no): ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "yes" || response == "y"
}

// PerformCleanup performs the actual cleanup operations
func PerformCleanup(opts CleanupOptions) (*CleanupReport, error) {
	report := &CleanupReport{
		RemovedDirs:  make([]string, 0),
		RemovedFiles: make([]string, 0),
		Errors:       make([]error, 0),
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return report, fmt.Errorf("failed to get home directory: %v", err)
	}

	emailDir := filepath.Join(homeDir, ".email")

	// Create backup if requested
	if opts.CreateBackup && !opts.DryRun {
		backupPath, err := CreateBackup(emailDir, opts.BackupPath)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Errorf("backup failed: %v", err))
		} else {
			report.BackupCreated = backupPath
		}
	}

	// Remove email data
	if !opts.KeepEmails {
		if opts.DryRun {
			report.RemovedDirs = append(report.RemovedDirs, emailDir)
		} else {
			// Count emails before removal
			emailCount := countEmailFiles(emailDir)
			report.EmailsRemoved = emailCount

			// Calculate size
			report.BytesFreed = calculateDirSize(emailDir)

			// Remove directory
			if err := os.RemoveAll(emailDir); err != nil {
				report.Errors = append(report.Errors, fmt.Errorf("failed to remove %s: %v", emailDir, err))
			} else {
				report.RemovedDirs = append(report.RemovedDirs, emailDir)
			}
		}
	}

	// Remove local configs if found
	localConfigs := FindLocalConfigs()
	for _, config := range localConfigs {
		if opts.DryRun {
			report.RemovedDirs = append(report.RemovedDirs, config)
		} else {
			if err := os.RemoveAll(config); err != nil {
				report.Errors = append(report.Errors, fmt.Errorf("failed to remove local config %s: %v", config, err))
			} else {
				report.RemovedDirs = append(report.RemovedDirs, config)
			}
		}
	}

	return report, nil
}

// CreateBackup creates a backup of the email directory
func CreateBackup(emailDir, customPath string) (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	
	var backupPath string
	if customPath != "" {
		backupPath = filepath.Join(customPath, fmt.Sprintf("emailos-backup-%s", timestamp))
	} else {
		homeDir, _ := os.UserHomeDir()
		backupPath = filepath.Join(homeDir, "Downloads", fmt.Sprintf("emailos-backup-%s", timestamp))
	}

	// Create backup directory
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return "", err
	}

	// Copy email directory to backup
	if err := copyDir(emailDir, filepath.Join(backupPath, ".email")); err != nil {
		return "", err
	}

	return backupPath, nil
}

// RemoveBinary attempts to remove the EmailOS binary
func RemoveBinary(installation InstallationInfo, quiet bool) {
	if installation.BinaryPath == "" {
		return
	}

	switch installation.Method {
	case "npm":
		if !quiet {
			fmt.Println("📦 For npm installation, run: npm uninstall -g mailos")
		}
	case "homebrew":
		if !quiet {
			fmt.Println("🍺 For Homebrew installation, run: brew uninstall mailos")
		}
	case "manual":
		// Try to remove manually installed binary
		if err := os.Remove(installation.BinaryPath); err != nil {
			if !quiet {
				fmt.Printf("⚠️  Could not remove binary at %s (may require sudo)\n", installation.BinaryPath)
				fmt.Printf("   Run: sudo rm %s\n", installation.BinaryPath)
			}
		} else if !quiet {
			fmt.Printf("✓ Removed binary: %s\n", installation.BinaryPath)
		}
	default:
		if !quiet {
			fmt.Printf("ℹ️  Unknown installation method. Binary location: %s\n", installation.BinaryPath)
		}
	}
}

// DisplayCleanupReport shows the results of cleanup
func DisplayCleanupReport(report *CleanupReport, dryRun bool) {
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82")) // Green
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")) // Blue

	if dryRun {
		fmt.Println(infoStyle.Render("🔍 Dry Run - What would be removed:"))
	} else {
		fmt.Println(successStyle.Render("✅ Cleanup Complete"))
	}
	fmt.Println()

	if report.BackupCreated != "" {
		fmt.Printf("💾 Backup created: %s\n", report.BackupCreated)
		fmt.Println()
	}

	if len(report.RemovedDirs) > 0 {
		fmt.Println("📁 Directories removed:")
		for _, dir := range report.RemovedDirs {
			fmt.Printf("   • %s\n", dir)
		}
		fmt.Println()
	}

	if len(report.RemovedFiles) > 0 {
		fmt.Println("📄 Files removed:")
		for _, file := range report.RemovedFiles {
			fmt.Printf("   • %s\n", file)
		}
		fmt.Println()
	}

	if report.EmailsRemoved > 0 {
		fmt.Printf("📧 Emails removed: %d\n", report.EmailsRemoved)
	}

	if report.BytesFreed > 0 {
		fmt.Printf("💾 Space freed: %s\n", formatBytes(report.BytesFreed))
	}

	if len(report.Errors) > 0 {
		fmt.Println()
		fmt.Println(errorStyle.Render("❌ Errors encountered:"))
		for _, err := range report.Errors {
			fmt.Printf("   • %v\n", err)
		}
	}

	if !dryRun && len(report.Errors) == 0 {
		fmt.Println()
		fmt.Println(successStyle.Render("🎉 EmailOS has been completely uninstalled from your system."))
		fmt.Println()
		fmt.Println("Thank you for using EmailOS! 👋")
	}
}

// Helper functions

func calculateDirSize(path string) int64 {
	var size int64
	
	filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	
	return size
}

func countEmailFiles(path string) int {
	count := 0
	
	filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && (strings.HasSuffix(filePath, ".md") || strings.HasSuffix(filePath, ".json")) {
			// Check if it's in sent, received, or drafts directory
			dir := filepath.Dir(filePath)
			if strings.Contains(dir, "sent") || strings.Contains(dir, "received") || strings.Contains(dir, "drafts") {
				count++
			}
		}
		return nil
	})
	
	return count
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

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

	if _, err := destFile.ReadFrom(sourceFile); err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

// CleanupDetector checks for orphaned EmailOS data
type CleanupDetector struct {
	CheckInterval time.Duration
	LastCheck     time.Time
}

// NewCleanupDetector creates a new cleanup detector
func NewCleanupDetector() *CleanupDetector {
	return &CleanupDetector{
		CheckInterval: 24 * time.Hour, // Check daily
	}
}

// CheckForOrphanedData detects if EmailOS data exists without binary
func (cd *CleanupDetector) CheckForOrphanedData() bool {
	// Check if we should run detection
	if time.Since(cd.LastCheck) < cd.CheckInterval {
		return false
	}

	cd.LastCheck = time.Now()

	// Check if email data exists
	dataExists, _ := CheckEmailDataExists()
	if !dataExists {
		return false
	}

	// Check if binary exists in common locations
	commonPaths := []string{
		"/usr/local/bin/mailos",
		"/opt/homebrew/bin/mailos",
		"/usr/bin/mailos",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return false // Binary found, not orphaned
		}
	}

	// Check npm global packages
	if runtime.GOOS != "windows" {
		homeDir, _ := os.UserHomeDir()
		npmPaths := []string{
			filepath.Join(homeDir, ".npm-global", "bin", "mailos"),
			filepath.Join(homeDir, ".nvm", "versions", "node", "*", "bin", "mailos"),
		}
		
		for _, pattern := range npmPaths {
			matches, _ := filepath.Glob(pattern)
			if len(matches) > 0 {
				return false // Binary found
			}
		}
	}

	return true // Data exists but no binary found
}

// PromptOrphanedCleanup prompts user to clean up orphaned data
func (cd *CleanupDetector) PromptOrphanedCleanup() error {
	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow
	
	fmt.Println()
	fmt.Println(warningStyle.Render("🧹 Orphaned EmailOS Data Detected"))
	fmt.Println(warningStyle.Render("═══════════════════════════════"))
	fmt.Println()
	fmt.Println("EmailOS configuration and email data was found on your system,")
	fmt.Println("but the EmailOS binary appears to have been removed.")
	fmt.Println()
	fmt.Print("Would you like to clean up this orphaned data? (y/n): ")
	
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	
	response = strings.TrimSpace(strings.ToLower(response))
	if response == "y" || response == "yes" {
		opts := CleanupOptions{
			Force:        false,
			RemoveAll:    true,
			CreateBackup: true,
		}
		return UninstallCommand(opts)
	}
	
	return nil
}