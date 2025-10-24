package helpers

import (
	"os"
	"path/filepath"
	"testing"
	"time"
	
	mailos "github.com/anduimagui/emailos"
)

// TestConfig holds configuration for test setup
type TestConfig struct {
	TempDir      string
	MockEmail    string
	MockPassword string
	Provider     string
}

// TestClient wraps the mailos client with test-specific functionality
type TestClient struct {
	Client   *mailos.Client
	Config   *TestConfig
	TempDir  string
	cleanup  []func()
}

// SetupTest initializes a test environment with proper cleanup
func SetupTest(t *testing.T) (*TestClient, func()) {
	tempDir, err := os.MkdirTemp("", "mailos-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	config := &TestConfig{
		TempDir:      tempDir,
		MockEmail:    "test@example.com",
		MockPassword: "mock-password",
		Provider:     "mock",
	}

	testClient := &TestClient{
		Config:  config,
		TempDir: tempDir,
		cleanup: []func(){},
	}

	// Add cleanup function for temp directory
	testClient.AddCleanup(func() {
		os.RemoveAll(tempDir)
	})

	cleanup := func() {
		testClient.Cleanup()
	}

	return testClient, cleanup
}

// AddCleanup adds a cleanup function to be called when test finishes
func (tc *TestClient) AddCleanup(fn func()) {
	tc.cleanup = append(tc.cleanup, fn)
}

// Cleanup runs all registered cleanup functions
func (tc *TestClient) Cleanup() {
	for i := len(tc.cleanup) - 1; i >= 0; i-- {
		tc.cleanup[i]()
	}
}

// CreateTestFile creates a temporary file with specified content
func (tc *TestClient) CreateTestFile(filename, content string) (string, error) {
	filePath := filepath.Join(tc.TempDir, filename)
	
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", err
	}

	tc.AddCleanup(func() {
		os.Remove(filePath)
	})

	return filePath, nil
}

// CreateTestAttachment creates a test attachment file
func (tc *TestClient) CreateTestAttachment(filename string, size int) (string, error) {
	content := make([]byte, size)
	for i := range content {
		content[i] = byte('A' + (i % 26))
	}
	
	return tc.CreateTestFile(filename, string(content))
}

// MockConfig creates a mock configuration for testing
func (tc *TestClient) MockConfig() *mailos.Config {
	return &mailos.Config{
		Email:        tc.Config.MockEmail,
		Password:     tc.Config.MockPassword,
		Provider:     tc.Config.Provider,
		FromName:     "Test User",
		FromEmail:    tc.Config.MockEmail,
		DefaultAICLI: "mock",
	}
}

// SetEnvVars sets up test environment variables
func SetupTestEnv(t *testing.T) func() {
	originalVars := map[string]string{
		"FROM_EMAIL": os.Getenv("FROM_EMAIL"),
		"TO_EMAIL":   os.Getenv("TO_EMAIL"),
	}

	os.Setenv("FROM_EMAIL", "test-from@example.com")
	os.Setenv("TO_EMAIL", "test-to@example.com")

	return func() {
		for key, value := range originalVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}
}

// CreateTestEmailData creates test email data for various scenarios
func CreateTestEmailData() []*mailos.Email {
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
			Subject: "Important: Meeting Tomorrow",
			Body:    "Don't forget about our meeting tomorrow at 2 PM",
			Date:    now.Add(-2 * time.Hour),
			Attachments: []string{"agenda.pdf"},
		},
		{
			ID:      3,
			From:    "newsletter@company.com",
			To:      []string{"test@example.com"},
			Subject: "Weekly Newsletter",
			Body:    "Here's what happened this week...",
			Date:    now.Add(-24 * time.Hour),
		},
	}
}

// AssertEmailEqual compares two emails for equality in tests
func AssertEmailEqual(t *testing.T, expected, actual *mailos.Email) {
	t.Helper()
	
	if expected.ID != actual.ID {
		t.Errorf("Email ID mismatch: expected %d, got %d", expected.ID, actual.ID)
	}
	
	if expected.From != actual.From {
		t.Errorf("Email From mismatch: expected %s, got %s", expected.From, actual.From)
	}
	
	if expected.Subject != actual.Subject {
		t.Errorf("Email Subject mismatch: expected %s, got %s", expected.Subject, actual.Subject)
	}
	
	if len(expected.To) != len(actual.To) {
		t.Errorf("Email To length mismatch: expected %d, got %d", len(expected.To), len(actual.To))
	}
	
	for i, to := range expected.To {
		if i < len(actual.To) && to != actual.To[i] {
			t.Errorf("Email To[%d] mismatch: expected %s, got %s", i, to, actual.To[i])
		}
	}
}

// SkipIfNoEmailConfig skips test if email configuration is not available
func SkipIfNoEmailConfig(t *testing.T) {
	if os.Getenv("FROM_EMAIL") == "" || os.Getenv("TO_EMAIL") == "" {
		t.Skip("Skipping test: email configuration not available")
	}
}

// CreateTestDraftsDir creates a temporary drafts directory for testing
func (tc *TestClient) CreateTestDraftsDir() string {
	draftsDir := filepath.Join(tc.TempDir, "drafts")
	err := os.MkdirAll(draftsDir, 0755)
	if err != nil {
		panic("Failed to create drafts directory: " + err.Error())
	}
	
	tc.AddCleanup(func() {
		os.RemoveAll(draftsDir)
	})
	
	return draftsDir
}

// CreateTestDraft creates a test draft file
func (tc *TestClient) CreateTestDraft(filename, to, subject, body string) (string, error) {
	draftsDir := tc.CreateTestDraftsDir()
	
	draftContent := "---\n"
	draftContent += "to: " + to + "\n"
	draftContent += "subject: " + subject + "\n"
	draftContent += "---\n\n"
	draftContent += body
	
	draftPath := filepath.Join(draftsDir, filename)
	err := os.WriteFile(draftPath, []byte(draftContent), 0644)
	if err != nil {
		return "", err
	}
	
	return draftPath, nil
}