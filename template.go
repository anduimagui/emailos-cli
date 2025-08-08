package mailos

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ManageTemplate handles the template customization flow
func ManageTemplate() error {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("EMAIL TEMPLATE CUSTOMIZATION")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Customize your email template to create beautiful,")
	fmt.Println("branded emails with your own design.")
	fmt.Println()
	fmt.Println("The template editor allows you to:")
	fmt.Println("• Design a custom HTML email template")
	fmt.Println("• Preview your design in real-time")
	fmt.Println("• Use {{BODY}} placeholder for email content")
	fmt.Println("• Use {{PROFILE_IMAGE}} placeholder for profile image")
	fmt.Println("• Add your branding, colors, and styling")
	fmt.Println()
	
	// Check if template already exists
	templatePath, err := GetTemplatePath()
	if err == nil {
		if _, err := os.Stat(templatePath); err == nil {
			fmt.Printf("Current template location: %s\n", templatePath)
			fmt.Println()
		}
	}

	fmt.Println("Opening the EmailOS Template Editor in your browser...")
	fmt.Println("URL: https://email-os.com/editor")
	fmt.Println()
	
	// Open browser
	editorURL := "https://email-os.com/editor"
	if err := openBrowserURL(editorURL); err != nil {
		fmt.Printf("Could not open browser automatically.\n")
		fmt.Printf("Please manually visit: %s\n", editorURL)
	} else {
		fmt.Println("✓ Browser opened")
	}
	
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("INSTRUCTIONS")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("1. Design your template in the browser editor")
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
	
	// Ask if they want to paste template
	fmt.Print("Ready to paste your template? (y/n): ")
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	
	if response != "y" && response != "yes" {
		fmt.Println("Template customization cancelled.")
		return nil
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
	
	// Save template
	if err := SaveTemplate(template); err != nil {
		return fmt.Errorf("failed to save template: %v", err)
	}
	
	fmt.Println()
	fmt.Println("✓ Template saved successfully!")
	fmt.Println()
	fmt.Println("Your custom template will now be used when sending emails.")
	fmt.Println("To remove the template and use default formatting, delete:")
	
	templatePath, _ = GetTemplatePath()
	fmt.Printf("  %s\n", templatePath)
	fmt.Println()
	
	return nil
}

// GetTemplatePath returns the path to the template file
func GetTemplatePath() (string, error) {
	// First check for local .email/template.html
	localTemplate := filepath.Join(".email", "template.html")
	if _, err := os.Stat(localTemplate); err == nil {
		absPath, _ := filepath.Abs(localTemplate)
		return absPath, nil
	}
	
	// Fall back to global ~/.email/template.html
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".email", "template.html"), nil
}

// SaveTemplate saves the HTML template
func SaveTemplate(template string) error {
	// Determine where to save based on existing config
	var templatePath string
	
	// Check if local .email exists
	if _, err := os.Stat(".email"); err == nil {
		// Save to local .email
		templatePath = filepath.Join(".email", "template.html")
		// Ensure .email is in .gitignore
		if err := EnsureGitIgnore(); err != nil {
			// Don't fail the operation, just warn
			fmt.Printf("Note: Could not update .gitignore: %v\n", err)
		}
	} else {
		// Save to global ~/.email
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		
		emailDir := filepath.Join(homeDir, ".email")
		if err := os.MkdirAll(emailDir, 0700); err != nil {
			return err
		}
		
		templatePath = filepath.Join(emailDir, "template.html")
	}
	
	// Write template file
	return os.WriteFile(templatePath, []byte(template), 0644)
}

// LoadTemplate loads the HTML template if it exists
func LoadTemplate() (string, error) {
	templatePath, err := GetTemplatePath()
	if err != nil {
		return "", err
	}
	
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", err
	}
	
	return string(content), nil
}

// ApplyTemplate applies the template to the email body
func ApplyTemplate(body string, bodyHTML string) string {
	// Try to load template
	template, err := LoadTemplate()
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

// ApplyTemplateWithProfile applies the template to the email body including profile image
func ApplyTemplateWithProfile(body string, bodyHTML string, profileImagePath string) string {
	// Try to load template
	template, err := LoadTemplate()
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

// TemplateExists checks if a template file exists
func TemplateExists() bool {
	templatePath, err := GetTemplatePath()
	if err != nil {
		return false
	}
	_, err = os.Stat(templatePath)
	return err == nil
}

// RemoveTemplate deletes the template file
func RemoveTemplate() error {
	templatePath, err := GetTemplatePath()
	if err != nil {
		return err
	}
	
	if err := os.Remove(templatePath); err != nil {
		return fmt.Errorf("failed to remove template: %v", err)
	}
	
	fmt.Println("✓ Template removed successfully")
	return nil
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