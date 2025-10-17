package mailos

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"
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
	EditUID       uint32 // UID of draft to edit (0 means create new)
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
	InReplyTo   string   // For threading: the Message-ID of the email being replied to
	References  []string // For threading: chain of Message-IDs in the conversation
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

	// Handle editing an existing draft
	if opts.EditUID > 0 {
		return editDraftInIMAP(opts)
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
	for _, draft := range drafts {
		// Save to local .email/drafts as markdown
		if err := saveLocalDraft(draft); err != nil {
			fmt.Printf("⚠️  Could not save draft to local storage: %v\n", err)
		} else {
			fmt.Printf("✓ Saved draft to local .email/drafts folder\n")
		}
		
		// Save to IMAP Drafts folder (without sending)
		uid, err := saveDraftToIMAP(draft)
		if err != nil {
			// Don't fail the whole operation if IMAP save fails
			fmt.Printf("⚠️  Could not save draft to email account: %v\n", err)
		} else {
			fmt.Printf("✓ Saved draft to email account's Drafts folder (UID: %d)\n", uid)
		}
	}

	fmt.Printf("\n📧 Created %d draft(s) in .email/drafts folder\n", len(drafts))
	fmt.Printf("📤 To send all drafts, run: mailos send --drafts\n")
	fmt.Printf("📮 Drafts are also saved in your email account's Drafts folder (not sent)\n")
	
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


// saveDraftToIMAP saves a draft email to the IMAP Drafts folder and returns its UID
func saveDraftToIMAP(draft DraftEmail) (uint32, error) {
	config, err := LoadConfig()
	if err != nil {
		return 0, fmt.Errorf("failed to load config: %v", err)
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
	
	// Add threading headers if this is a reply
	if draft.InReplyTo != "" {
		// Ensure Message-ID is wrapped in angle brackets
		inReplyTo := draft.InReplyTo
		if !strings.HasPrefix(inReplyTo, "<") {
			inReplyTo = "<" + inReplyTo + ">"
		}
		message.WriteString(fmt.Sprintf("In-Reply-To: %s\r\n", inReplyTo))
	}
	
	if len(draft.References) > 0 {
		// Ensure all Message-IDs in References are wrapped in angle brackets
		var refs []string
		for _, ref := range draft.References {
			if !strings.HasPrefix(ref, "<") {
				refs = append(refs, "<"+ref+">")
			} else {
				refs = append(refs, ref)
			}
		}
		message.WriteString(fmt.Sprintf("References: %s\r\n", strings.Join(refs, " ")))
	}
	
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
		return 0, fmt.Errorf("failed to get IMAP settings: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", imapHost, imapPort)
	
	// Connect with TLS
	tlsConfig := &tls.Config{ServerName: imapHost}
	c, err := client.DialTLS(addr, tlsConfig)
	if err != nil {
		// Try without TLS
		c, err = client.Dial(addr)
		if err != nil {
			return 0, fmt.Errorf("failed to connect to IMAP server: %v", err)
		}
		
		// Start TLS if not already encrypted
		if ok, _ := c.SupportStartTLS(); ok {
			if err := c.StartTLS(tlsConfig); err != nil {
				return 0, fmt.Errorf("failed to start TLS: %v", err)
			}
		}
	}
	defer c.Logout()

	// Login
	if err := c.Login(config.Email, config.Password); err != nil {
		return 0, fmt.Errorf("failed to login: %v", err)
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
		return 0, fmt.Errorf("failed to list folders: %v", err)
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
		return 0, fmt.Errorf("failed to save draft to %s folder: %v", selectedFolder, err)
	}

	// Get the UID of the newly appended message
	// Note: This is a simplified approach. In production, you might need to use UIDPLUS extension
	// or search for the message you just added
	mbox, err := c.Select(selectedFolder, false)
	if err != nil {
		return 0, fmt.Errorf("failed to select folder to get UID: %v", err)
	}
	
	// The last message should be the one we just added (in most cases)
	// This is not guaranteed but works for most IMAP servers
	if mbox.Messages > 0 {
		seqSet := new(imap.SeqSet)
		seqSet.AddNum(mbox.Messages) // Get the last message
		
		messages := make(chan *imap.Message, 1)
		done := make(chan error, 1)
		go func() {
			done <- c.Fetch(seqSet, []imap.FetchItem{imap.FetchUid}, messages)
		}()
		
		if msg := <-messages; msg != nil {
			<-done
			return msg.Uid, nil
		}
		<-done
	}

	return 0, nil
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
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate, imap.FetchUid}
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
			fmt.Printf("\n📧 Draft #%d (UID: %d)\n", count, msg.Uid)
			
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

// editDraftInIMAP edits an existing draft in the IMAP Drafts folder
func editDraftInIMAP(opts DraftsOptions) error {
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
	if _, err := c.Select(selectedFolder, false); err != nil {
		return fmt.Errorf("failed to select drafts folder: %v", err)
	}

	// Fetch the existing draft by UID
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(opts.EditUID)

	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)
	
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchUid, section.FetchItem()}

	go func() {
		done <- c.UidFetch(seqSet, items, messages)
	}()

	var oldMsg *imap.Message
	for msg := range messages {
		oldMsg = msg
		break
	}

	if err := <-done; err != nil {
		return fmt.Errorf("failed to fetch draft: %v", err)
	}

	if oldMsg == nil {
		return fmt.Errorf("draft with UID %d not found", opts.EditUID)
	}

	// Delete the old draft
	deleteSet := new(imap.SeqSet)
	deleteSet.AddNum(opts.EditUID)
	
	// Mark as deleted
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.DeletedFlag}
	if err := c.UidStore(deleteSet, item, flags, nil); err != nil {
		return fmt.Errorf("failed to mark old draft for deletion: %v", err)
	}

	// Expunge to actually delete
	if err := c.Expunge(nil); err != nil {
		return fmt.Errorf("failed to expunge old draft: %v", err)
	}

	// Create the updated draft
	draft := DraftEmail{
		To:          opts.To,
		CC:          opts.CC,
		BCC:         opts.BCC,
		Subject:     opts.Subject,
		Body:        opts.Body,
		Attachments: opts.Attachments,
		Priority:    opts.Priority,
	}

	// If body was specified from file, read it
	if opts.FileBody != "" {
		fileContent, err := os.ReadFile(opts.FileBody)
		if err != nil {
			return fmt.Errorf("failed to read body from file %s: %v", opts.FileBody, err)
		}
		draft.Body = string(fileContent)
	}

	// Save the updated draft
	newUID, err := saveDraftToIMAP(draft)
	if err != nil {
		return fmt.Errorf("failed to save updated draft: %v", err)
	}

	fmt.Printf("✓ Updated draft (old UID: %d, new UID: %d)\n", opts.EditUID, newUID)
	return nil
}