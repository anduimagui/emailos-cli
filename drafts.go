package mailos

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DraftsOptions contains configuration for the drafts command
type DraftsOptions struct {
	Query         string
	Template      string
	DataFile      string
	OutputDir     string
	Interactive   bool
	UseAI         bool
	DraftCount    int
}

// DraftEmail represents an email draft with metadata
type DraftEmail struct {
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	Body        string
	Attachments []string
	SendAfter   *time.Time
	Priority    string
}

// DraftsCommand generates draft emails based on user input
func DraftsCommand(opts DraftsOptions) error {
	// Set default output directory
	if opts.OutputDir == "" {
		opts.OutputDir = "draft-emails"
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create draft directory: %v", err)
	}

	// Generate drafts based on the method
	var drafts []DraftEmail
	var err error

	if opts.Query != "" && opts.UseAI {
		// Use AI to generate drafts from query
		drafts, err = generateDraftsWithAI(opts.Query, opts.DraftCount)
		if err != nil {
			return fmt.Errorf("failed to generate drafts with AI: %v", err)
		}
	} else if opts.Template != "" {
		// Generate from template
		drafts, err = generateDraftsFromTemplate(opts.Template, opts.DataFile)
		if err != nil {
			return fmt.Errorf("failed to generate drafts from template: %v", err)
		}
	} else if opts.Interactive {
		// Interactive draft creation
		drafts, err = createDraftsInteractively()
		if err != nil {
			return fmt.Errorf("failed to create drafts interactively: %v", err)
		}
	} else {
		// Default: create a single draft interactively
		draft, err := createSingleDraftInteractively()
		if err != nil {
			return fmt.Errorf("failed to create draft: %v", err)
		}
		drafts = []DraftEmail{draft}
	}

	// Save drafts to files
	for i, draft := range drafts {
		filename := generateDraftFilename(draft.Subject, i+1)
		filepath := filepath.Join(opts.OutputDir, filename)
		
		if err := saveDraftToFile(draft, filepath); err != nil {
			return fmt.Errorf("failed to save draft %d: %v", i+1, err)
		}
		
		fmt.Printf("✓ Created draft: %s\n", filepath)
	}

	fmt.Printf("\n📧 Created %d draft(s) in %s/\n", len(drafts), opts.OutputDir)
	fmt.Printf("📤 To send all drafts, run: mailos send --drafts\n")
	
	return nil
}

// generateDraftsWithAI uses AI to generate email drafts from a query
func generateDraftsWithAI(query string, count int) ([]DraftEmail, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	// Check if AI provider is configured
	if config.DefaultAICLI == "" {
		return nil, fmt.Errorf("no AI provider configured. Run 'mailos provider' to set up AI")
	}

	// For now, create a simple example draft
	// In a full implementation, this would call the AI provider
	fmt.Printf("🤖 Using AI to generate %d draft(s) from query: %s\n", count, query)
	fmt.Println("Note: AI draft generation is a placeholder - implement with actual AI provider")
	
	drafts := []DraftEmail{}
	for i := 0; i < count; i++ {
		draft := DraftEmail{
			To:      []string{"recipient@example.com"},
			Subject: fmt.Sprintf("Draft %d: %s", i+1, query),
			Body:    fmt.Sprintf("This is draft %d generated from your query:\n\n%s\n\n[AI-generated content would go here]", i+1, query),
		}
		drafts = append(drafts, draft)
	}
	
	return drafts, nil
}

// generateDraftsFromTemplate generates drafts from a template file
func generateDraftsFromTemplate(templateName string, dataFile string) ([]DraftEmail, error) {
	// Placeholder for template-based generation
	fmt.Printf("📝 Generating drafts from template: %s\n", templateName)
	if dataFile != "" {
		fmt.Printf("📊 Using data from: %s\n", dataFile)
	}
	
	// For now, return a single example draft
	draft := DraftEmail{
		To:      []string{"template@example.com"},
		Subject: fmt.Sprintf("Email from template: %s", templateName),
		Body:    "This email was generated from a template.\n\n[Template content would be processed here]",
	}
	
	return []DraftEmail{draft}, nil
}

// createDraftsInteractively allows creating multiple drafts interactively
func createDraftsInteractively() ([]DraftEmail, error) {
	reader := bufio.NewReader(os.Stdin)
	drafts := []DraftEmail{}
	
	for {
		fmt.Println("\n📝 Create a new draft email")
		draft, err := createSingleDraftInteractively()
		if err != nil {
			return nil, err
		}
		drafts = append(drafts, draft)
		
		fmt.Print("\nCreate another draft? (y/n): ")
		response, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			break
		}
	}
	
	return drafts, nil
}

// createSingleDraftInteractively creates a single draft through user input
func createSingleDraftInteractively() (DraftEmail, error) {
	reader := bufio.NewReader(os.Stdin)
	draft := DraftEmail{}
	
	// Get recipient
	fmt.Print("To (email address): ")
	to, _ := reader.ReadString('\n')
	draft.To = []string{strings.TrimSpace(to)}
	
	// Get CC (optional)
	fmt.Print("CC (optional, press Enter to skip): ")
	cc, _ := reader.ReadString('\n')
	cc = strings.TrimSpace(cc)
	if cc != "" {
		draft.CC = []string{cc}
	}
	
	// Get subject
	fmt.Print("Subject: ")
	subject, _ := reader.ReadString('\n')
	draft.Subject = strings.TrimSpace(subject)
	
	// Get body
	fmt.Println("Body (press Enter twice to finish):")
	var bodyLines []string
	emptyCount := 0
	for {
		line, _ := reader.ReadString('\n')
		if line == "\n" {
			emptyCount++
			if emptyCount >= 2 {
				break
			}
		} else {
			emptyCount = 0
		}
		bodyLines = append(bodyLines, line)
	}
	draft.Body = strings.Join(bodyLines, "")
	
	// Get priority (optional)
	fmt.Print("Priority (high/normal/low, default: normal): ")
	priority, _ := reader.ReadString('\n')
	priority = strings.TrimSpace(priority)
	if priority == "" {
		priority = "normal"
	}
	draft.Priority = priority
	
	return draft, nil
}

// generateDraftFilename creates a filename for the draft
func generateDraftFilename(subject string, index int) string {
	// Sanitize subject for filename
	safe := strings.Map(func(r rune) rune {
		if r == ' ' {
			return '-'
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return -1
	}, subject)
	
	if safe == "" {
		safe = "draft"
	}
	
	// Limit length
	if len(safe) > 50 {
		safe = safe[:50]
	}
	
	timestamp := time.Now().Format("2006-01-02-150405")
	return fmt.Sprintf("%03d-%s-%s.md", index, safe, timestamp)
}

// saveDraftToFile saves a draft email to a markdown file with frontmatter
func saveDraftToFile(draft DraftEmail, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Write frontmatter
	file.WriteString("---\n")
	file.WriteString(fmt.Sprintf("to: %s\n", strings.Join(draft.To, ", ")))
	
	if len(draft.CC) > 0 {
		file.WriteString(fmt.Sprintf("cc: %s\n", strings.Join(draft.CC, ", ")))
	}
	
	if len(draft.BCC) > 0 {
		file.WriteString(fmt.Sprintf("bcc: %s\n", strings.Join(draft.BCC, ", ")))
	}
	
	file.WriteString(fmt.Sprintf("subject: %s\n", draft.Subject))
	
	if len(draft.Attachments) > 0 {
		file.WriteString("attachments:\n")
		for _, attachment := range draft.Attachments {
			file.WriteString(fmt.Sprintf("  - %s\n", attachment))
		}
	}
	
	if draft.SendAfter != nil {
		file.WriteString(fmt.Sprintf("send_after: %s\n", draft.SendAfter.Format("2006-01-02 15:04:05")))
	}
	
	if draft.Priority != "" {
		file.WriteString(fmt.Sprintf("priority: %s\n", draft.Priority))
	}
	
	file.WriteString("---\n\n")
	
	// Write body
	file.WriteString(draft.Body)
	
	return nil
}