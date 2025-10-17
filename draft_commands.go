package mailos

import (
	"fmt"
	"strconv"
	"strings"
)

// DraftReference represents a way to reference a draft
type DraftReference struct {
	Number   int    // User-friendly number (1, 2, 3...)
	UID      uint32 // Internal IMAP UID
	Subject  string
	To       []string
	From     string
	IsLatest bool
}

// GetDraftList returns a list of drafts with user-friendly numbering
func GetDraftList() ([]*DraftReference, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	opts := ReadOptions{
		Limit: 1000, // Get all drafts
	}

	drafts, err := client.ReadDrafts(opts)
	if err != nil {
		return nil, err
	}

	var references []*DraftReference
	for i, draft := range drafts {
		ref := &DraftReference{
			Number:  i + 1, // Start from 1
			UID:     draft.ID,
			Subject: draft.Subject,
			To:      draft.To,
			From:    draft.From,
		}
		references = append(references, ref)
	}

	return references, nil
}

// FindDraftByReference finds a draft based on various reference methods
func FindDraftByReference(ref string, subject string, to string, latest bool) (*DraftReference, error) {
	drafts, err := GetDraftList()
	if err != nil {
		return nil, err
	}

	if len(drafts) == 0 {
		return nil, fmt.Errorf("no drafts found")
	}

	// Handle --latest flag
	if latest {
		return drafts[0], nil // Most recent draft
	}

	// Handle numeric reference (draft number)
	if ref != "" {
		num, err := strconv.Atoi(ref)
		if err == nil {
			if num < 1 || num > len(drafts) {
				return nil, fmt.Errorf("draft number %d not found (available: 1-%d)", num, len(drafts))
			}
			return drafts[num-1], nil // Convert to 0-based index
		}
	}

	// Handle subject search
	if subject != "" {
		for _, draft := range drafts {
			if strings.Contains(strings.ToLower(draft.Subject), strings.ToLower(subject)) {
				return draft, nil
			}
		}
		return nil, fmt.Errorf("no draft found with subject containing '%s'", subject)
	}

	// Handle recipient search
	if to != "" {
		for _, draft := range drafts {
			for _, recipient := range draft.To {
				if strings.Contains(strings.ToLower(recipient), strings.ToLower(to)) {
					return draft, nil
				}
			}
		}
		return nil, fmt.Errorf("no draft found with recipient containing '%s'", to)
	}

	return nil, fmt.Errorf("no valid reference provided")
}

// DraftCreateOptions represents options for creating a draft
type DraftCreateOptions struct {
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	Body        string
	Interactive bool
	UseAI       bool
}

// DraftEditOptions represents options for editing a draft
type DraftEditOptions struct {
	Reference   string
	Subject     string
	To          string
	Latest      bool
	NewSubject  string
	NewBody     string
	NewTo       []string
	NewCC       []string
	NewBCC      []string
	Interactive bool
}

// CreateDraft creates a new draft with simplified interface
func CreateDraft(opts DraftCreateOptions) error {
	if opts.Interactive {
		return createDraftInteractive(opts)
	}

	if opts.UseAI {
		return createDraftWithAI(opts)
	}

	// Create standard draft
	draftOpts := DraftsOptions{
		To:      opts.To,
		CC:      opts.CC,
		BCC:     opts.BCC,
		Subject: opts.Subject,
		Body:    opts.Body,
	}

	return DraftsCommand(draftOpts)
}

// EditDraft edits an existing draft with simplified interface
func EditDraft(opts DraftEditOptions) error {
	// Find the draft to edit
	draft, err := FindDraftByReference(opts.Reference, opts.Subject, opts.To, opts.Latest)
	if err != nil {
		return err
	}

	fmt.Printf("Editing draft #%d (UID: %d): %s\n", draft.Number, draft.UID, draft.Subject)

	if opts.Interactive {
		return editDraftInteractive(draft, opts)
	}

	// Update the draft
	updateOpts := DraftsOptions{
		EditUID: draft.UID,
	}

	// Apply updates
	if opts.NewSubject != "" {
		updateOpts.Subject = opts.NewSubject
	}
	if opts.NewBody != "" {
		updateOpts.Body = opts.NewBody
	}
	if len(opts.NewTo) > 0 {
		updateOpts.To = opts.NewTo
	}
	if len(opts.NewCC) > 0 {
		updateOpts.CC = opts.NewCC
	}
	if len(opts.NewBCC) > 0 {
		updateOpts.BCC = opts.NewBCC
	}

	return DraftsCommand(updateOpts)
}

// ListDrafts shows all drafts with user-friendly numbering
func ListDrafts() error {
	drafts, err := GetDraftList()
	if err != nil {
		return err
	}

	if len(drafts) == 0 {
		fmt.Println("No drafts found")
		return nil
	}

	fmt.Printf("📧 Found %d draft(s):\n\n", len(drafts))
	for _, draft := range drafts {
		toList := strings.Join(draft.To, ", ")
		if toList == "" {
			toList = "(no recipients)"
		}
		
		fmt.Printf("#%d - %s\n", draft.Number, draft.Subject)
		fmt.Printf("     To: %s\n", toList)
		fmt.Printf("     UID: %d\n\n", draft.UID)
	}

	fmt.Println("Commands:")
	fmt.Println("  mailos draft edit 1              # Edit draft #1")
	fmt.Println("  mailos draft edit --latest       # Edit latest draft")
	fmt.Println("  mailos draft send 1              # Send draft #1")
	fmt.Println("  mailos draft show 1              # Show draft #1 content")

	return nil
}

// ShowDraft displays the content of a specific draft
func ShowDraft(reference string) error {
	draft, err := FindDraftByReference(reference, "", "", false)
	if err != nil {
		return err
	}

	// Use existing read functionality
	opts := DraftsOptions{
		Read:    true,
		EditUID: draft.UID,
	}

	fmt.Printf("📧 Draft #%d (UID: %d)\n", draft.Number, draft.UID)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	return DraftsCommand(opts)
}

// SendDraft sends a specific draft
func SendDraft(reference string) error {
	draft, err := FindDraftByReference(reference, "", "", false)
	if err != nil {
		return err
	}

	fmt.Printf("Sending draft #%d: %s\n", draft.Number, draft.Subject)
	
	// TODO: Implement sending specific draft by UID
	// For now, return an error directing to current send command
	return fmt.Errorf("sending individual drafts not yet implemented. Use: mailos send --drafts")
}

// DeleteDraft deletes a specific draft
func DeleteDraft(reference string, confirm bool) error {
	if !confirm {
		return fmt.Errorf("use --confirm flag to delete drafts")
	}

	draft, err := FindDraftByReference(reference, "", "", false)
	if err != nil {
		return err
	}

	fmt.Printf("Deleting draft #%d: %s\n", draft.Number, draft.Subject)
	
	client, err := NewClient()
	if err != nil {
		return err
	}

	err = client.DeleteDrafts([]uint32{draft.UID})
	if err != nil {
		return fmt.Errorf("failed to delete draft: %v", err)
	}

	fmt.Printf("✓ Deleted draft #%d\n", draft.Number)
	return nil
}

// Helper functions for interactive and AI modes
func createDraftInteractive(opts DraftCreateOptions) error {
	// TODO: Implement interactive draft creation
	return fmt.Errorf("interactive draft creation not yet implemented")
}

func createDraftWithAI(opts DraftCreateOptions) error {
	// Use existing AI functionality
	draftOpts := DraftsOptions{
		UseAI:      true,
		DraftCount: 1,
		To:         opts.To,
		Subject:    opts.Subject,
		Body:       opts.Body,
	}
	
	return DraftsCommand(draftOpts)
}

func editDraftInteractive(draft *DraftReference, opts DraftEditOptions) error {
	// TODO: Implement interactive draft editing
	return fmt.Errorf("interactive draft editing not yet implemented")
}