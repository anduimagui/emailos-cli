package mailos

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ManageTemplate handles the template customization flow
func ManageTemplate() error {
	return manageTemplateWithName("")
}

// ManageTemplateWithName handles the template customization flow for a named template
func ManageTemplateWithName(name string) error {
	return manageTemplateWithName(name)
}

// manageTemplateWithName handles the template customization flow
func manageTemplateWithName(templateName string) error {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if templateName != "" {
		fmt.Printf("EMAIL TEMPLATE CUSTOMIZATION - %s\n", templateName)
	} else {
		fmt.Println("EMAIL TEMPLATE CUSTOMIZATION")
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	
	// Load config to show current from_email
	config, err := LoadConfig()
	if err != nil {
		fmt.Println("Warning: Could not load email configuration")
	} else {
		fromEmail := config.Email
		if config.FromEmail != "" {
			fromEmail = config.FromEmail
		}
		fmt.Printf("Current sender email: %s\n", fromEmail)
		if config.FromName != "" {
			fmt.Printf("Display name: %s\n", config.FromName)
		}
		fmt.Println()
	}
	
	fmt.Println("Customize your email template to create beautiful,")
	fmt.Println("branded emails with your own design.")
	fmt.Println()
	fmt.Println("Template customization allows you to:")
	fmt.Println("• Design a custom HTML email template")
	fmt.Println("• Preview your design in real-time")
	fmt.Println("• Use {{BODY}} placeholder for email content")
	fmt.Println("• Use {{PROFILE_IMAGE}} placeholder for profile image")
	fmt.Println("• Add your branding, colors, and styling")
	fmt.Println("• Preview templates locally before saving")
	fmt.Println()
	
	// Check if template already exists
	templatePath, err := GetTemplatePathWithName(templateName)
	if err == nil {
		if _, err := os.Stat(templatePath); err == nil {
			fmt.Printf("Current template location: %s\n", templatePath)
			fmt.Println()
		}
	}

	
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("INSTRUCTIONS")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("1. Design your template using any HTML editor or tool")
	fmt.Println("2. Use {{BODY}} where email content should appear")
	fmt.Println("3. Use {{PROFILE_IMAGE}} where profile image should appear (optional)")
	fmt.Println("4. Copy the HTML code when you're satisfied")
	fmt.Println("5. Come back here and paste it")
	fmt.Println()
	fmt.Println("The {{BODY}} placeholder will be replaced with your")
	fmt.Println("email content (converted from Markdown to HTML).")
	fmt.Println()
	fmt.Println("Example template structure:")
	fmt.Println("  <html>")
	fmt.Println("    <body style=\"font-family: Arial;\">")
	fmt.Println("      <div class=\"header\">My Company</div>")
	fmt.Println("      <div class=\"content\">{{BODY}}</div>")
	fmt.Println("      <div class=\"footer\">© 2024</div>")
	fmt.Println("    </body>")
	fmt.Println("  </html>")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	
	reader := bufio.NewReader(os.Stdin)
	
	// Ask what they want to do
	fmt.Println("What would you like to do?")
	fmt.Println("1. Create/Edit template")
	fmt.Println("2. Preview existing template")
	fmt.Println("3. List templates")
	fmt.Print("Enter choice (1-3): ")
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)
	
	switch response {
	case "1":
		return createEditTemplate(templateName, reader)
	case "2":
		return previewTemplate(templateName, reader)
	case "3":
		return listTemplates()
	default:
		fmt.Println("Invalid choice. Template customization cancelled.")
		return nil
	}
}

// createEditTemplate handles template creation/editing
func createEditTemplate(templateName string, reader *bufio.Reader) error {
	// Get template name if not provided
	if templateName == "" {
		fmt.Print("Enter template name: ")
		name, _ := reader.ReadString('\n')
		templateName = strings.TrimSpace(name)
		if templateName == "" {
			fmt.Println("Template name required.")
			return nil
		}
	}
	
	fmt.Println()
	fmt.Println("Paste your HTML template below.")
	fmt.Println("When done, type 'END' on a new line and press Enter:")
	fmt.Println()
	
	// Read multi-line template
	var templateLines []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read template: %v", err)
		}
		
		// Check for end marker
		if strings.TrimSpace(line) == "END" {
			break
		}
		
		templateLines = append(templateLines, line)
	}
	
	if len(templateLines) == 0 {
		fmt.Println("No template provided.")
		return nil
	}
	
	// Join lines to create template
	template := strings.Join(templateLines, "")
	
	// Validate template contains {{BODY}} placeholder
	if !strings.Contains(template, "{{BODY}}") {
		fmt.Println()
		fmt.Println("⚠️  Warning: Template doesn't contain {{BODY}} placeholder.")
		fmt.Println("Without {{BODY}}, your email content won't be inserted.")
		fmt.Print("Continue anyway? (y/n): ")
		
		confirm, _ := reader.ReadString('\n')
		confirm = strings.TrimSpace(strings.ToLower(confirm))
		if confirm != "y" && confirm != "yes" {
			fmt.Println("Template save cancelled.")
			return nil
		}
	}
	
	// Ask if they want to preview before saving
	fmt.Print("Preview template before saving? (y/n): ")
	previewResponse, _ := reader.ReadString('\n')
	previewResponse = strings.TrimSpace(strings.ToLower(previewResponse))
	
	if previewResponse == "y" || previewResponse == "yes" {
		if err := previewTemplateContent(template); err != nil {
			fmt.Printf("Preview failed: %v\n", err)
		}
		
		fmt.Print("Continue with saving? (y/n): ")
		saveResponse, _ := reader.ReadString('\n')
		saveResponse = strings.TrimSpace(strings.ToLower(saveResponse))
		if saveResponse != "y" && saveResponse != "yes" {
			fmt.Println("Template save cancelled.")
			return nil
		}
	}
	
	// Save template
	if err := SaveTemplateWithName(template, templateName); err != nil {
		return fmt.Errorf("failed to save template: %v", err)
	}
	
	fmt.Println()
	fmt.Printf("✓ Template '%s' saved successfully!\n", templateName)
	fmt.Println()
	fmt.Println("Your custom template will now be available for use when sending emails.")
	
	templatePath, _ := GetTemplatePathWithName(templateName)
	fmt.Printf("Saved to: %s\n", templatePath)
	fmt.Println()
	
	return nil
}

// previewTemplate handles template preview
func previewTemplate(templateName string, reader *bufio.Reader) error {
	// If no template name provided, ask for it or list available
	if templateName == "" {
		templates, err := ListTemplateNames()
		if err != nil {
			return fmt.Errorf("failed to list templates: %v", err)
		}
		
		if len(templates) == 0 {
			fmt.Println("No templates found. Create one first.")
			return nil
		}
		
		fmt.Println("Available templates:")
		for i, name := range templates {
			fmt.Printf("%d. %s\n", i+1, name)
		}
		
		fmt.Print("Enter template name or number: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		
		// Try to parse as number first
		if num := parseTemplateNumber(input, len(templates)); num > 0 {
			templateName = templates[num-1]
		} else {
			templateName = input
		}
	}
	
	// Load and preview template
	template, err := LoadTemplateWithName(templateName)
	if err != nil {
		return fmt.Errorf("failed to load template '%s': %v", templateName, err)
	}
	
	return previewTemplateContent(template)
}

// previewTemplateContent creates a temporary HTML file and opens it in browser
func previewTemplateContent(template string) error {
	// Replace placeholders with sample content for preview
	sampleHTML := strings.ReplaceAll(template, "{{BODY}}", "<h2>Sample Email Content</h2><p>This is a preview of your email template with sample content. Your actual email content will appear here when sending emails.</p><p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.</p>")
	sampleHTML = strings.ReplaceAll(sampleHTML, "{{PROFILE_IMAGE}}", `<img src="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMTUwIiBoZWlnaHQ9IjE1MCIgdmlld0JveD0iMCAwIDE1MCAxNTAiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+CjxyZWN0IHdpZHRoPSIxNTAiIGhlaWdodD0iMTUwIiByeD0iNzUiIGZpbGw9IiNFNUU3RUIiLz4KPHN2ZyB4PSI0NSIgeT0iNDAiIHdpZHRoPSI2MCIgaGVpZ2h0PSI3MCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSIjOTNBM0I4Ij4KPHA+PHBhdGggZD0iTTEyIDEyYzIuMjEgMCA0LTEuNzkgNC00cy0xLjc5LTQtNC00LTQgMS43OS00IDQgMS43OSA0IDQgNHptMCAyYy0yLjY3IDAtOCAxLjM0LTggNHYyaDE2di0yYzAtMi42Ni01LjMzLTQtOC00eiIvPjwvcGF0aD48L3N2Zz4KPC9zdmc+" alt="Sample Profile" style="max-width: 150px; height: auto; border-radius: 50%;">`)
	
	// Create temporary file
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("email-template-preview-%d.html", time.Now().Unix()))
	
	if err := ioutil.WriteFile(tempFile, []byte(sampleHTML), 0644); err != nil {
		return fmt.Errorf("failed to create preview file: %v", err)
	}
	
	fmt.Printf("Opening template preview: %s\n", tempFile)
	
	// Open in browser
	if err := openBrowserURL("file://" + tempFile); err != nil {
		fmt.Printf("Could not open browser automatically.\n")
		fmt.Printf("Please manually open: file://%s\n", tempFile)
	} else {
		fmt.Println("✓ Preview opened in browser")
	}
	
	return nil
}

// listTemplates shows all available templates
func listTemplates() error {
	templates, err := ListTemplateNames()
	if err != nil {
		return fmt.Errorf("failed to list templates: %v", err)
	}
	
	if len(templates) == 0 {
		fmt.Println("No templates found.")
		return nil
	}
	
	fmt.Println("Available templates:")
	for _, name := range templates {
		path, _ := GetTemplatePathWithName(name)
		fmt.Printf("• %s (%s)\n", name, path)
	}
	fmt.Println()
	
	return nil
}

// parseTemplateNumber parses user input as a template number
func parseTemplateNumber(input string, maxNum int) int {
	if len(input) == 0 {
		return 0
	}
	
	num := 0
	for _, r := range input {
		if r < '0' || r > '9' {
			return 0
		}
		num = num*10 + int(r-'0')
	}
	
	if num < 1 || num > maxNum {
		return 0
	}
	
	return num
}

// GetTemplatePath returns the path to the default template file
func GetTemplatePath() (string, error) {
	return GetTemplatePathWithName("default")
}

// GetTemplatePathWithName returns the path to a named template file
func GetTemplatePathWithName(templateName string) (string, error) {
	if templateName == "" {
		templateName = "default"
	}
	
	// First check for local .email/templates/[name].html
	localTemplate := filepath.Join(".email", "templates", templateName+".html")
	if _, err := os.Stat(localTemplate); err == nil {
		absPath, _ := filepath.Abs(localTemplate)
		return absPath, nil
	}
	
	// Fall back to global ~/.email/templates/[name].html
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".email", "templates", templateName+".html"), nil
}

// SaveTemplate saves the HTML template with default name
func SaveTemplate(template string) error {
	return SaveTemplateWithName(template, "default")
}

// SaveTemplateWithName saves the HTML template with a specific name
func SaveTemplateWithName(template string, templateName string) error {
	if templateName == "" {
		templateName = "default"
	}
	
	// Determine where to save based on existing config
	var templatePath string
	
	// Check if local .email exists
	if _, err := os.Stat(".email"); err == nil {
		// Save to local .email/templates
		templatesDir := filepath.Join(".email", "templates")
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			return err
		}
		templatePath = filepath.Join(templatesDir, templateName+".html")
		// Ensure .email is in .gitignore
		if err := EnsureGitIgnore(); err != nil {
			// Don't fail the operation, just warn
			fmt.Printf("Note: Could not update .gitignore: %v\n", err)
		}
	} else {
		// Save to global ~/.email/templates
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		
		templatesDir := filepath.Join(homeDir, ".email", "templates")
		if err := os.MkdirAll(templatesDir, 0700); err != nil {
			return err
		}
		
		templatePath = filepath.Join(templatesDir, templateName+".html")
	}
	
	// Write template file
	return os.WriteFile(templatePath, []byte(template), 0644)
}

// LoadTemplate loads the default HTML template if it exists
func LoadTemplate() (string, error) {
	return LoadTemplateWithName("default")
}

// LoadTemplateWithName loads a named HTML template if it exists
func LoadTemplateWithName(templateName string) (string, error) {
	templatePath, err := GetTemplatePathWithName(templateName)
	if err != nil {
		return "", err
	}
	
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", err
	}
	
	return string(content), nil
}

// ApplyTemplate applies the default template to the email body
func ApplyTemplate(body string, bodyHTML string) string {
	return ApplyTemplateWithName(body, bodyHTML, "default")
}

// ApplyTemplateWithName applies a named template to the email body
func ApplyTemplateWithName(body string, bodyHTML string, templateName string) string {
	// Try to load template
	template, err := LoadTemplateWithName(templateName)
	if err != nil || template == "" {
		// No template, return original HTML
		return bodyHTML
	}
	
	// Use HTML body if available, otherwise use plain body
	content := bodyHTML
	if content == "" {
		content = strings.ReplaceAll(body, "\n", "<br>")
	}
	
	// Replace {{BODY}} placeholder with content
	result := strings.ReplaceAll(template, "{{BODY}}", content)
	
	return result
}

// ApplyTemplateWithProfile applies the default template to the email body including profile image
func ApplyTemplateWithProfile(body string, bodyHTML string, profileImagePath string) string {
	return ApplyTemplateWithProfileAndName(body, bodyHTML, profileImagePath, "default")
}

// ApplyTemplateWithProfileAndName applies a named template to the email body including profile image
func ApplyTemplateWithProfileAndName(body string, bodyHTML string, profileImagePath string, templateName string) string {
	// Try to load template
	template, err := LoadTemplateWithName(templateName)
	if err != nil || template == "" {
		// If no template, create a simple default with profile image if provided
		if profileImagePath != "" && bodyHTML != "" {
			imageTag := getProfileImageTag(profileImagePath)
			if imageTag != "" {
				// Add profile image at the top of the email
				return imageTag + "<br><br>" + bodyHTML
			}
		}
		return bodyHTML
	}
	
	// Use HTML body if available, otherwise use plain body
	content := bodyHTML
	if content == "" {
		content = strings.ReplaceAll(body, "\n", "<br>")
	}
	
	// Replace {{BODY}} placeholder with content
	result := strings.ReplaceAll(template, "{{BODY}}", content)
	
	// Replace {{PROFILE_IMAGE}} placeholder if profile image is provided
	if profileImagePath != "" {
		imageTag := getProfileImageTag(profileImagePath)
		result = strings.ReplaceAll(result, "{{PROFILE_IMAGE}}", imageTag)
	} else {
		// Remove profile image placeholder if no image provided
		result = strings.ReplaceAll(result, "{{PROFILE_IMAGE}}", "")
	}
	
	return result
}

// getProfileImageTag creates an HTML img tag with base64 encoded image
func getProfileImageTag(imagePath string) string {
	// Read the image file
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return ""
	}
	
	// Detect image type from file extension
	ext := strings.ToLower(filepath.Ext(imagePath))
	var mimeType string
	switch ext {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".png":
		mimeType = "image/png"
	case ".gif":
		mimeType = "image/gif"
	case ".webp":
		mimeType = "image/webp"
	default:
		// Default to jpeg if unknown
		mimeType = "image/jpeg"
	}
	
	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(imageData)
	
	// Create img tag with embedded base64 data
	// Using a reasonable max width for email display
	return fmt.Sprintf(`<img src="data:%s;base64,%s" alt="Profile" style="max-width: 150px; height: auto; border-radius: 50%%;">`, mimeType, encoded)
}

// TemplateExists checks if the default template file exists
func TemplateExists() bool {
	return TemplateExistsWithName("default")
}

// TemplateExistsWithName checks if a named template file exists
func TemplateExistsWithName(templateName string) bool {
	templatePath, err := GetTemplatePathWithName(templateName)
	if err != nil {
		return false
	}
	_, err = os.Stat(templatePath)
	return err == nil
}

// RemoveTemplate deletes the default template file
func RemoveTemplate() error {
	return RemoveTemplateWithName("default")
}

// RemoveTemplateWithName deletes a named template file
func RemoveTemplateWithName(templateName string) error {
	templatePath, err := GetTemplatePathWithName(templateName)
	if err != nil {
		return err
	}
	
	if err := os.Remove(templatePath); err != nil {
		return fmt.Errorf("failed to remove template '%s': %v", templateName, err)
	}
	
	fmt.Printf("✓ Template '%s' removed successfully\n", templateName)
	return nil
}

// ListTemplateNames returns a list of all available template names
func ListTemplateNames() ([]string, error) {
	var templates []string
	
	// Check local templates
	localDir := filepath.Join(".email", "templates")
	if localTemplates, err := getTemplateNamesFromDir(localDir); err == nil {
		templates = append(templates, localTemplates...)
	}
	
	// Check global templates
	homeDir, err := os.UserHomeDir()
	if err == nil {
		globalDir := filepath.Join(homeDir, ".email", "templates")
		if globalTemplates, err := getTemplateNamesFromDir(globalDir); err == nil {
			for _, name := range globalTemplates {
				// Only add if not already in list (local takes precedence)
				found := false
				for _, existing := range templates {
					if existing == name {
						found = true
						break
					}
				}
				if !found {
					templates = append(templates, name)
				}
			}
		}
	}
	
	return templates, nil
}

// getTemplateNamesFromDir extracts template names from a directory
func getTemplateNamesFromDir(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	
	var names []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".html") {
			name := strings.TrimSuffix(file.Name(), ".html")
			names = append(names, name)
		}
	}
	
	return names, nil
}

// PreviewTemplate creates a preview of a template by name
func PreviewTemplate(templateName string) error {
	template, err := LoadTemplateWithName(templateName)
	if err != nil {
		return fmt.Errorf("failed to load template '%s': %v", templateName, err)
	}
	
	return previewTemplateContent(template)
}

func openBrowserURL(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default:
		return fmt.Errorf("unsupported platform")
	}

	return exec.Command(cmd, args...).Start()
}

// OpenTemplateInBrowser opens the template HTML file in the browser
func OpenTemplateInBrowser() error {
	templatePath, err := GetTemplatePath()
	if err != nil {
		return fmt.Errorf("failed to get template path: %v", err)
	}
	
	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template file does not exist at %s", templatePath)
	}
	
	// Convert to file:// URL
	fileURL := "file://" + templatePath
	
	fmt.Printf("Opening template in browser: %s\n", templatePath)
	
	return openBrowser(fileURL)
}