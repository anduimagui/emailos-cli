package mocks

import (
	"fmt"
	"strings"
	"time"
	
	mailos "github.com/anduimagui/emailos"
)

// MockIMAPServer simulates an IMAP server for testing
type MockIMAPServer struct {
	Messages      []*mailos.Email
	Folders       []string
	Connected     bool
	LoginAttempts int
	Behavior      MockBehavior
}

// MockBehavior defines how the mock should behave
type MockBehavior struct {
	ShouldFailConnection bool
	ShouldFailLogin      bool
	ShouldFailFetch      bool
	ConnectionDelay      time.Duration
	LoginDelay          time.Duration
	FetchDelay          time.Duration
}

// NewMockIMAPServer creates a new mock IMAP server
func NewMockIMAPServer() *MockIMAPServer {
	return &MockIMAPServer{
		Messages: []*mailos.Email{},
		Folders:  []string{"INBOX", "Drafts", "Sent", "Trash"},
		Connected: false,
		LoginAttempts: 0,
		Behavior: MockBehavior{},
	}
}

// WithMessages adds test messages to the mock server
func (m *MockIMAPServer) WithMessages(messages []*mailos.Email) *MockIMAPServer {
	m.Messages = append(m.Messages, messages...)
	return m
}

// WithFolders sets the available folders
func (m *MockIMAPServer) WithFolders(folders []string) *MockIMAPServer {
	m.Folders = folders
	return m
}

// WithBehavior sets the mock behavior
func (m *MockIMAPServer) WithBehavior(behavior MockBehavior) *MockIMAPServer {
	m.Behavior = behavior
	return m
}

// Connect simulates connecting to the IMAP server
func (m *MockIMAPServer) Connect() error {
	if m.Behavior.ShouldFailConnection {
		return fmt.Errorf("mock connection failed")
	}
	
	if m.Behavior.ConnectionDelay > 0 {
		time.Sleep(m.Behavior.ConnectionDelay)
	}
	
	m.Connected = true
	return nil
}

// Login simulates logging into the IMAP server
func (m *MockIMAPServer) Login(email, password string) error {
	m.LoginAttempts++
	
	if !m.Connected {
		return fmt.Errorf("not connected to server")
	}
	
	if m.Behavior.ShouldFailLogin {
		return fmt.Errorf("mock login failed for %s", email)
	}
	
	if m.Behavior.LoginDelay > 0 {
		time.Sleep(m.Behavior.LoginDelay)
	}
	
	return nil
}

// FetchMessages simulates fetching messages from a folder
func (m *MockIMAPServer) FetchMessages(folder string, limit int) ([]*mailos.Email, error) {
	if !m.Connected {
		return nil, fmt.Errorf("not connected to server")
	}
	
	if m.Behavior.ShouldFailFetch {
		return nil, fmt.Errorf("mock fetch failed")
	}
	
	if m.Behavior.FetchDelay > 0 {
		time.Sleep(m.Behavior.FetchDelay)
	}
	
	// Filter messages by folder if needed
	var messages []*mailos.Email
	for _, msg := range m.Messages {
		messages = append(messages, msg)
		if len(messages) >= limit {
			break
		}
	}
	
	return messages, nil
}

// SearchMessages simulates searching for messages
func (m *MockIMAPServer) SearchMessages(criteria SearchCriteria) ([]*mailos.Email, error) {
	if !m.Connected {
		return nil, fmt.Errorf("not connected to server")
	}
	
	if m.Behavior.ShouldFailFetch {
		return nil, fmt.Errorf("mock search failed")
	}
	
	var results []*mailos.Email
	for _, msg := range m.Messages {
		if m.matchesCriteria(msg, criteria) {
			results = append(results, msg)
		}
	}
	
	return results, nil
}

// SearchCriteria defines search parameters
type SearchCriteria struct {
	From        string
	To          string
	Subject     string
	Body        string
	Since       *time.Time
	Before      *time.Time
	UnreadOnly  bool
	HasAttachments bool
}

// matchesCriteria checks if a message matches search criteria
func (m *MockIMAPServer) matchesCriteria(msg *mailos.Email, criteria SearchCriteria) bool {
	if criteria.From != "" && !strings.Contains(strings.ToLower(msg.From), strings.ToLower(criteria.From)) {
		return false
	}
	
	if criteria.To != "" {
		found := false
		for _, to := range msg.To {
			if strings.Contains(strings.ToLower(to), strings.ToLower(criteria.To)) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	if criteria.Subject != "" && !strings.Contains(strings.ToLower(msg.Subject), strings.ToLower(criteria.Subject)) {
		return false
	}
	
	if criteria.Body != "" && !strings.Contains(strings.ToLower(msg.Body), strings.ToLower(criteria.Body)) {
		return false
	}
	
	if criteria.Since != nil && msg.Date.Before(*criteria.Since) {
		return false
	}
	
	if criteria.Before != nil && msg.Date.After(*criteria.Before) {
		return false
	}
	
	if criteria.HasAttachments && len(msg.Attachments) == 0 {
		return false
	}
	
	return true
}

// GetMessage simulates fetching a specific message by ID
func (m *MockIMAPServer) GetMessage(id int) (*mailos.Email, error) {
	if !m.Connected {
		return nil, fmt.Errorf("not connected to server")
	}
	
	for _, msg := range m.Messages {
		if msg.ID == uint32(id) {
			return msg, nil
		}
	}
	
	return nil, fmt.Errorf("message with ID %d not found", id)
}

// AddMessage adds a message to the mock server
func (m *MockIMAPServer) AddMessage(msg *mailos.Email) {
	if msg.ID == 0 {
		msg.ID = uint32(len(m.Messages) + 1)
	}
	m.Messages = append(m.Messages, msg)
}

// DeleteMessage removes a message from the mock server
func (m *MockIMAPServer) DeleteMessage(id int) error {
	for i, msg := range m.Messages {
		if msg.ID == uint32(id) {
			m.Messages = append(m.Messages[:i], m.Messages[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("message with ID %d not found", id)
}

// GetFolders returns the list of available folders
func (m *MockIMAPServer) GetFolders() []string {
	return m.Folders
}

// CreateFolder creates a new folder
func (m *MockIMAPServer) CreateFolder(name string) error {
	for _, folder := range m.Folders {
		if folder == name {
			return fmt.Errorf("folder %s already exists", name)
		}
	}
	m.Folders = append(m.Folders, name)
	return nil
}

// GetStats returns mock statistics
func (m *MockIMAPServer) GetStats() (int, int, error) {
	if !m.Connected {
		return 0, 0, fmt.Errorf("not connected to server")
	}
	
	total := len(m.Messages)
	unread := total / 2 // Mock behavior: assume half are unread
	
	return total, unread, nil
}

// Disconnect simulates disconnecting from the server
func (m *MockIMAPServer) Disconnect() error {
	m.Connected = false
	return nil
}

// Reset clears all data and resets the mock server
func (m *MockIMAPServer) Reset() {
	m.Messages = []*mailos.Email{}
	m.Connected = false
	m.LoginAttempts = 0
	m.Folders = []string{"INBOX", "Drafts", "Sent", "Trash"}
	m.Behavior = MockBehavior{}
}

// CreateTestMessages creates a set of test messages for common scenarios
func CreateTestMessages() []*mailos.Email {
	now := time.Now()
	
	return []*mailos.Email{
		{
			ID:      1,
			From:    "sender1@example.com",
			To:      []string{"test@example.com"},
			Subject: "Test Email 1",
			Body:    "This is a test email body",
			Date:    now.Add(-1 * time.Hour),
		},
		{
			ID:      2,
			From:    "sender2@example.com",
			To:      []string{"test@example.com"},
			Subject: "Important Meeting",
			Body:    "Don't forget about our meeting tomorrow",
			Date:    now.Add(-2 * time.Hour),
			Attachments: []string{"agenda.pdf"},
			AttachmentData: map[string][]byte{
				"agenda.pdf": []byte("PDF content here"),
			},
		},
		{
			ID:      3,
			From:    "newsletter@company.com",
			To:      []string{"test@example.com"},
			Subject: "Weekly Newsletter",
			Body:    "Here's what happened this week...",
			Date:    now.Add(-24 * time.Hour),
		},
		{
			ID:      4,
			From:    "support@service.com",
			To:      []string{"test@example.com"},
			Subject: "Your Support Request",
			Body:    "Thank you for contacting support",
			Date:    now.Add(-72 * time.Hour),
		},
	}
}