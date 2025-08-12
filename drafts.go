package mailos

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
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
	List          bool   // List drafts from IMAP
	Read          bool   // Read drafts from IMAP
	// Email composition fields (same as send command)
	To            []string
	CC            []string
	BCC           []string
	Subject       string
	Body          string
	FileBody      string   // Read body from file (-f flag)
	Attachments   []string
	Priority      string
	PlainText     bool    // Send as plain text (-P flag)
	NoSignature   bool    // Don't include signature (-S flag)
	Signature     string  // Custom signature text
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
	// Handle listing drafts from IMAP or local storage
	if opts.List || opts.Read {
		// First try to list from local storage
		if err := listLocalDrafts(opts.Read); err != nil {
			fmt.Printf("Note: Could not read local drafts: %v\n", err)
		}
		// Then list from IMAP
		return listDraftsFromIMAP(opts.Read)
	}

	// Ensure local .email directories exist
	if err := EnsureEmailDirectories(); err != nil {
		return fmt.Errorf("failed to create email directories: %v", err)
	}

	// Set default output directory to .email/drafts
	if opts.OutputDir == "" {
		draftsDir, err := GetDraftsDir()
		if err != nil {
			return fmt.Errorf("failed to get drafts directory: %v", err)
		}
		opts.OutputDir = draftsDir
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
		// Check if we have command-line specified fields
		if len(opts.To) > 0 || opts.Subject != "" || opts.Body != "" || opts.FileBody != "" {
			// Handle body from file if specified
			body := opts.Body
			if opts.FileBody != "" {
				fileContent, err := os.ReadFile(opts.FileBody)
				if err != nil {
					return fmt.Errorf("failed to read body from file %s: %v", opts.FileBody, err)
				}
				body = string(fileContent)
			}
			
			// Create draft from command-line arguments
			draft := DraftEmail{
				To:          opts.To,
				CC:          opts.CC,
				BCC:         opts.BCC,
				Subject:     opts.Subject,
				Body:        body,
				Attachments: opts.Attachments,
				Priority:    opts.Priority,
			}
			drafts = []DraftEmail{draft}
		} else {
			// Default: create a single draft interactively
			draft, err := createSingleDraftInteractively()
			if err != nil {
				return fmt.Errorf("failed to create draft: %v", err)
			}
			drafts = []DraftEmail{draft}
		}
	}

	// Save drafts to both local files and IMAP Drafts folder
	for i, draft := range drafts {
		// Save to local .email/drafts as JSON
		if err := saveLocalDraft(draft); err != nil {
			fmt.Printf("⚠️  Could not save draft to local storage: %v\n", err)
		} else {
			fmt.Printf("✓ Saved draft to local .email/drafts folder\n")
		}
		
		// Also save as markdown file if OutputDir is not the default drafts dir
		draftsDir, _ := GetDraftsDir()
		if opts.OutputDir != draftsDir {
			filename := generateDraftFilename(draft.Subject, i+1)
			filepath := filepath.Join(opts.OutputDir, filename)
			
			if err := saveDraftToFile(draft, filepath); err != nil {
				return fmt.Errorf("failed to save draft %d to file: %v", i+1, err)
			}
			
			fmt.Printf("✓ Created draft file: %s\n", filepath)
		}
		
		// Save to IMAP Drafts folder
		if err := saveDraftToIMAP(draft); err != nil {
			// Don't fail the whole operation if IMAP save fails
			fmt.Printf("⚠️  Could not save draft to email account: %v\n", err)
		} else {
			fmt.Printf("✓ Saved draft to email account's Drafts folder\n")
		}
	}

	fmt.Printf("\n📧 Created %d draft(s) in %s/\n", len(drafts), opts.OutputDir)
	fmt.Printf("📤 To send all drafts, run: mailos send --drafts\n")
	fmt.Printf("📮 Drafts are also saved in your email account's Drafts folder\n")
	
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

// saveDraftToIMAP saves a draft email to the IMAP Drafts folder
func saveDraftToIMAP(draft DraftEmail) error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Build the email message in RFC 822 format
	var message bytes.Buffer
	
	// Use FromEmail if specified, otherwise use the account email
	fromEmail := config.Email
	if config.FromEmail != "" {
		fromEmail = config.FromEmail
	}
	
	from := fromEmail
	if config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", config.FromName, fromEmail)
	}

	// Write headers
	message.WriteString(fmt.Sprintf("From: %s\r\n", from))
	message.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(draft.To, ", ")))
	if len(draft.CC) > 0 {
		message.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(draft.CC, ", ")))
	}
	if len(draft.BCC) > 0 {
		message.WriteString(fmt.Sprintf("Bcc: %s\r\n", strings.Join(draft.BCC, ", ")))
	}
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", draft.Subject))
	message.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	
	// Add a draft header to indicate this is a draft
	message.WriteString("X-Draft: true\r\n")
	if draft.Priority != "" && draft.Priority != "normal" {
		message.WriteString(fmt.Sprintf("X-Priority: %s\r\n", draft.Priority))
	}
	
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	message.WriteString("\r\n")
	message.WriteString(draft.Body)

	// Connect to IMAP server
	imapHost, imapPort, err := config.GetIMAPSettings()
	if err != nil {
		return fmt.Errorf("failed to get IMAP settings: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", imapHost, imapPort)
	
	// Connect with TLS
	tlsConfig := &tls.Config{ServerName: imapHost}
	c, err := client.DialTLS(addr, tlsConfig)
	if err != nil {
		// Try without TLS
		c, err = client.Dial(addr)
		if err != nil {
			return fmt.Errorf("failed to connect to IMAP server: %v", err)
		}
		
		// Start TLS if not already encrypted
		if ok, _ := c.SupportStartTLS(); ok {
			if err := c.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("failed to start TLS: %v", err)
			}
		}
	}
	defer c.Logout()

	// Login
	if err := c.Login(config.Email, config.Password); err != nil {
		return fmt.Errorf("failed to login: %v", err)
	}

	// Find the Drafts folder
	// Common draft folder names
	draftFolderNames := []string{"Drafts", "INBOX.Drafts", "[Gmail]/Drafts", "Draft", "INBOX.Draft"}
	
	// List all folders to find the drafts folder
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()
	
	var selectedFolder string
	availableFolders := []string{}
	for m := range mailboxes {
		availableFolders = append(availableFolders, m.Name)
		// Check if this is a drafts folder
		for _, draftName := range draftFolderNames {
			if strings.EqualFold(m.Name, draftName) || strings.Contains(strings.ToLower(m.Name), "draft") {
				selectedFolder = m.Name
				break
			}
		}
	}
	
	if err := <-done; err != nil {
		return fmt.Errorf("failed to list folders: %v", err)
	}
	
	// If no drafts folder found, try common names
	if selectedFolder == "" {
		for _, folderName := range draftFolderNames {
			_, err := c.Select(folderName, false)
			if err == nil {
				selectedFolder = folderName
				break
			}
		}
	}
	
	// If still no folder, create one or use INBOX
	if selectedFolder == "" {
		// Try to create a Drafts folder
		err := c.Create("Drafts")
		if err == nil {
			selectedFolder = "Drafts"
		} else {
			// Fall back to INBOX if we can't create Drafts
			selectedFolder = "INBOX"
		}
	}

	// Append the draft to the folder
	flags := []string{imap.DraftFlag} // Mark as draft
	date := time.Now()
	
	messageStr := message.String()
	// Ensure CRLF line endings
	messageWithCRLF := strings.ReplaceAll(messageStr, "\n", "\r\n")
	messageWithCRLF = strings.ReplaceAll(messageWithCRLF, "\r\r\n", "\r\n")
	
	err = c.Append(selectedFolder, flags, date, strings.NewReader(messageWithCRLF))
	if err != nil {
		return fmt.Errorf("failed to save draft to %s folder: %v", selectedFolder, err)
	}

	return nil
}

// listDraftsFromIMAP lists or reads drafts from the IMAP Drafts folder
func listDraftsFromIMAP(showFullContent bool) error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Connect to IMAP server
	imapHost, imapPort, err := config.GetIMAPSettings()
	if err != nil {
		return fmt.Errorf("failed to get IMAP settings: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", imapHost, imapPort)
	
	// Connect with TLS
	tlsConfig := &tls.Config{ServerName: imapHost}
	c, err := client.DialTLS(addr, tlsConfig)
	if err != nil {
		// Try without TLS
		c, err = client.Dial(addr)
		if err != nil {
			return fmt.Errorf("failed to connect to IMAP server: %v", err)
		}
		
		// Start TLS if not already encrypted
		if ok, _ := c.SupportStartTLS(); ok {
			if err := c.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("failed to start TLS: %v", err)
			}
		}
	}
	defer c.Logout()

	// Login
	if err := c.Login(config.Email, config.Password); err != nil {
		return fmt.Errorf("failed to login: %v", err)
	}

	// Find and select the Drafts folder
	selectedFolder, err := findDraftsFolder(c)
	if err != nil {
		return fmt.Errorf("failed to find Drafts folder: %v", err)
	}

	// Select the drafts folder
	mbox, err := c.Select(selectedFolder, false)
	if err != nil {
		return fmt.Errorf("failed to select drafts folder: %v", err)
	}

	fmt.Printf("📮 Reading drafts from %s folder\n", selectedFolder)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// If mailbox is empty
	if mbox.Messages == 0 {
		fmt.Println("No drafts found in your Drafts folder")
		return nil
	}

	// Fetch all drafts
	seqSet := new(imap.SeqSet)
	seqSet.AddRange(1, mbox.Messages)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate}
	if showFullContent {
		items = append(items, section.FetchItem())
	}

	go func() {
		done <- c.Fetch(seqSet, items, messages)
	}()

	count := 0
	for msg := range messages {
		count++
		
		// Display draft information
		envelope := msg.Envelope
		if envelope != nil {
			fmt.Printf("\n📧 Draft #%d\n", count)
			
			// From
			if len(envelope.From) > 0 {
				from := envelope.From[0]
				fmt.Printf("  From: %s <%s>\n", from.PersonalName, from.Address())
			}
			
			// To
			if len(envelope.To) > 0 {
				toAddrs := []string{}
				for _, addr := range envelope.To {
					if addr.PersonalName != "" {
						toAddrs = append(toAddrs, fmt.Sprintf("%s <%s>", addr.PersonalName, addr.Address()))
					} else {
						toAddrs = append(toAddrs, addr.Address())
					}
				}
				fmt.Printf("  To: %s\n", strings.Join(toAddrs, ", "))
			}
			
			// Subject
			fmt.Printf("  Subject: %s\n", envelope.Subject)
			
			// Date
			if !msg.InternalDate.IsZero() {
				fmt.Printf("  Date: %s\n", msg.InternalDate.Format("Jan 2, 2006 at 3:04 PM"))
			}
			
			// Flags
			flagStrs := []string{}
			for _, flag := range msg.Flags {
				flagStrs = append(flagStrs, flag)
			}
			if len(flagStrs) > 0 {
				fmt.Printf("  Flags: %s\n", strings.Join(flagStrs, ", "))
			}
			
			// Show body if requested
			if showFullContent {
				body := msg.GetBody(section)
				if body != nil {
					bodyBytes, err := ioutil.ReadAll(body)
					if err == nil && len(bodyBytes) > 0 {
						bodyStr := string(bodyBytes)
						// Extract just the text content
						lines := strings.Split(bodyStr, "\n")
						inBody := false
						bodyContent := []string{}
						for _, line := range lines {
							if inBody {
								bodyContent = append(bodyContent, line)
							} else if line == "" {
								inBody = true
							}
						}
						if len(bodyContent) > 0 {
							fmt.Println("\n  Body:")
							fmt.Println("  ──────────────────────────────────────────")
							bodyText := strings.Join(bodyContent, "\n")
							// Indent body content
							for _, line := range strings.Split(bodyText, "\n") {
								if strings.TrimSpace(line) != "" {
									fmt.Printf("  %s\n", line)
								}
							}
							fmt.Println("  ──────────────────────────────────────────")
						}
					}
				}
			}
		}
	}

	if err := <-done; err != nil {
		return fmt.Errorf("failed to fetch drafts: %v", err)
	}

	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Total: %d draft(s) in %s\n", count, selectedFolder)
	
	if !showFullContent {
		fmt.Println("\nTip: Use 'mailos drafts --read' to see full draft content")
	}

	return nil
}

// findDraftsFolder locates the Drafts folder on the IMAP server
func findDraftsFolder(c *client.Client) (string, error) {
	// Common draft folder names
	draftFolderNames := []string{"Drafts", "INBOX.Drafts", "[Gmail]/Drafts", "Draft", "INBOX.Draft"}
	
	// List all folders to find the drafts folder
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()
	
	var selectedFolder string
	for m := range mailboxes {
		// Check if this is a drafts folder
		for _, draftName := range draftFolderNames {
			if strings.EqualFold(m.Name, draftName) || strings.Contains(strings.ToLower(m.Name), "draft") {
				selectedFolder = m.Name
				break
			}
		}
		if selectedFolder != "" {
			break
		}
	}
	
	if err := <-done; err != nil {
		return "", fmt.Errorf("failed to list folders: %v", err)
	}
	
	// If no drafts folder found, try common names
	if selectedFolder == "" {
		for _, folderName := range draftFolderNames {
			_, err := c.Select(folderName, false)
			if err == nil {
				selectedFolder = folderName
				break
			}
		}
	}
	
	if selectedFolder == "" {
		return "", fmt.Errorf("no Drafts folder found")
	}
	
	return selectedFolder, nil
}