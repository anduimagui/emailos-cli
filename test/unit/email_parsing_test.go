package unit

import (
	"regexp"
	"testing"
	"time"
	
	"github.com/anduimagui/emailos/test/helpers"
	"github.com/anduimagui/emailos/test/mocks"
	mailos "github.com/anduimagui/emailos"
)

func TestEmailStructure(t *testing.T) {
	t.Run("should create valid email structure", func(t *testing.T) {
		
		email := &mailos.Email{
			ID:      1,
			From:    "test@example.com",
			To:      []string{"recipient@example.com"},
			Subject: "Test Subject",
			Body:    "Test body content",
			Date:    time.Now(),
		}
		
		helpers.AssertEmailStructure(t, email)
	})
	
	t.Run("should handle multiple recipients", func(t *testing.T) {
		
		email := &mailos.Email{
			ID:      1,
			From:    "test@example.com",
			To:      []string{"recipient1@example.com", "recipient2@example.com"},
			Subject: "Test Subject",
			Body:    "Test body content",
			Date:    time.Now(),
		}
		
		helpers.AssertLen(t, email.To, 2, "Should have 2 recipients")
		helpers.AssertEqual(t, "recipient1@example.com", email.To[0])
		helpers.AssertEqual(t, "recipient2@example.com", email.To[1])
	})
	
	t.Run("should validate email addresses", func(t *testing.T) {
		validEmails := []string{
			"test@example.com",
			"user.name@domain.co.uk",
			"support+tag@service.org",
		}
		
		for _, email := range validEmails {
			helpers.AssertValidEmail(t, email, "Should be valid email: %s", email)
		}
		
		invalidEmails := []string{
			"invalid-email",
			"@domain.com",
			"user@",
			"",
		}
		
		for _, email := range invalidEmails {
			t.Run("invalid_email_"+email, func(t *testing.T) {
				// Test that email validation correctly identifies invalid emails
				// We'll use our own validation logic instead of the helper
				emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
				if emailRegex.MatchString(email) {
					t.Errorf("Expected email '%s' to be invalid but it passed validation", email)
				}
				// If we get here, the email was correctly identified as invalid
			})
		}
	})
}

func TestEmailAttachments(t *testing.T) {
	t.Run("should handle single attachment", func(t *testing.T) {
		
		email := &mailos.Email{
			ID:      1,
			From:    "test@example.com",
			To:      []string{"recipient@example.com"},
			Subject: "Email with attachment",
			Body:    "Please see attached file",
			Date:    time.Now(),
			Attachments: []string{"document.pdf"},
			AttachmentData: map[string][]byte{
				"document.pdf": []byte("PDF content here"),
			},
		}
		
		helpers.AssertAttachmentsValid(t, email)
		helpers.AssertLen(t, email.Attachments, 1, "Should have 1 attachment")
		helpers.AssertNotEmpty(t, email.AttachmentData["document.pdf"], "Attachment data should not be empty")
	})
	
	t.Run("should handle multiple attachments", func(t *testing.T) {
		
		email := &mailos.Email{
			ID:      1,
			From:    "test@example.com",
			To:      []string{"recipient@example.com"},
			Subject: "Email with multiple attachments",
			Body:    "Please see attached files",
			Date:    time.Now(),
			Attachments: []string{"document.pdf", "image.jpg", "data.csv"},
			AttachmentData: map[string][]byte{
				"document.pdf": []byte("PDF content"),
				"image.jpg":    []byte("JPEG content"),
				"data.csv":     []byte("CSV content"),
			},
		}
		
		helpers.AssertAttachmentsValid(t, email)
		helpers.AssertLen(t, email.Attachments, 3, "Should have 3 attachments")
		
		for _, filename := range email.Attachments {
			helpers.AssertNotEmpty(t, email.AttachmentData[filename], "Attachment data should not be empty for %s", filename)
		}
	})
	
	t.Run("should handle large attachments", func(t *testing.T) {
		
		// Create 1MB attachment
		largeData := make([]byte, 1024*1024)
		for i := range largeData {
			largeData[i] = byte('A' + (i % 26))
		}
		
		email := &mailos.Email{
			ID:      1,
			From:    "test@example.com",
			To:      []string{"recipient@example.com"},
			Subject: "Email with large attachment",
			Body:    "Please see attached large file",
			Date:    time.Now(),
			Attachments: []string{"large_file.txt"},
			AttachmentData: map[string][]byte{
				"large_file.txt": largeData,
			},
		}
		
		helpers.AssertAttachmentsValid(t, email)
		helpers.AssertTrue(t, len(email.AttachmentData["large_file.txt"]) > 900000, "Large attachment should be close to 1MB")
	})
	
	t.Run("should handle email without attachments", func(t *testing.T) {
		
		email := &mailos.Email{
			ID:      1,
			From:    "test@example.com",
			To:      []string{"recipient@example.com"},
			Subject: "Simple email",
			Body:    "This email has no attachments",
			Date:    time.Now(),
		}
		
		helpers.AssertLen(t, email.Attachments, 0, "Should have no attachments")
		helpers.AssertLen(t, email.AttachmentData, 0, "Should have no attachment data")
	})
}

func TestEmailDateHandling(t *testing.T) {
	t.Run("should handle current date", func(t *testing.T) {
		now := time.Now()
		email := &mailos.Email{
			ID:      1,
			From:    "test@example.com",
			To:      []string{"recipient@example.com"},
			Subject: "Test Subject",
			Body:    "Test body content",
			Date:    now,
		}
		
		helpers.AssertTimeAlmostEqual(t, now, email.Date, time.Second, "Email date should match current time")
	})
	
	t.Run("should handle past dates", func(t *testing.T) {
		pastDate := time.Now().Add(-24 * time.Hour)
		email := &mailos.Email{
			ID:      1,
			From:    "test@example.com",
			To:      []string{"recipient@example.com"},
			Subject: "Yesterday's email",
			Body:    "This email is from yesterday",
			Date:    pastDate,
		}
		
		helpers.AssertTrue(t, email.Date.Before(time.Now()), "Email date should be in the past")
		helpers.AssertTimeAlmostEqual(t, pastDate, email.Date, time.Second, "Email date should match expected past date")
	})
	
	t.Run("should handle timezone differences", func(t *testing.T) {
		utc := time.Now().UTC()
		local := utc.Local()
		
		email := &mailos.Email{
			ID:      1,
			From:    "test@example.com",
			To:      []string{"recipient@example.com"},
			Subject: "Timezone test",
			Body:    "Testing timezone handling",
			Date:    utc,
		}
		
		helpers.AssertTimeAlmostEqual(t, local, email.Date.Local(), time.Second, "Should handle timezone conversion correctly")
	})
}

func TestEmailComparison(t *testing.T) {
	t.Run("should identify identical emails", func(t *testing.T) {
		email1 := &mailos.Email{
			ID:      1,
			From:    "test@example.com",
			To:      []string{"recipient@example.com"},
			Subject: "Test Subject",
			Body:    "Test body content",
			Date:    time.Now(),
		}
		
		email2 := &mailos.Email{
			ID:      1,
			From:    "test@example.com",
			To:      []string{"recipient@example.com"},
			Subject: "Test Subject",
			Body:    "Test body content",
			Date:    email1.Date,
		}
		
		helpers.AssertEmailEqual(t, email1, email2)
	})
	
	t.Run("should identify different emails", func(t *testing.T) {
		email1 := &mailos.Email{
			ID:      1,
			From:    "test@example.com",
			To:      []string{"recipient@example.com"},
			Subject: "Test Subject",
			Body:    "Test body content",
			Date:    time.Now(),
		}
		
		email2 := &mailos.Email{
			ID:      2,
			From:    "test@example.com",
			To:      []string{"recipient@example.com"},
			Subject: "Different Subject",
			Body:    "Different body content",
			Date:    time.Now(),
		}
		
		// Test that emails are different by checking individual fields
		if email1.ID == email2.ID {
			t.Error("Expected different IDs")
		}
		if email1.Subject == email2.Subject {
			t.Error("Expected different subjects")
		}
		if email1.Body == email2.Body {
			t.Error("Expected different bodies")
		}
		// If we get here, the emails are correctly identified as different
	})
}

func TestEmailWithMockData(t *testing.T) {
	t.Run("should work with mock email data", func(t *testing.T) {
		
		mockServer := mocks.NewMockIMAPServer()
		testMessages := mocks.CreateTestMessages()
		mockServer.WithMessages(testMessages)
		
		helpers.AssertLen(t, mockServer.Messages, 4, "Should have 4 test messages")
		
		for _, msg := range mockServer.Messages {
			helpers.AssertEmailStructure(t, msg)
		}
	})
	
	t.Run("should handle emails with attachments from mock", func(t *testing.T) {
		
		mockServer := mocks.NewMockIMAPServer()
		testMessages := mocks.CreateTestMessages()
		mockServer.WithMessages(testMessages)
		
		// Find the message with attachments
		var attachmentMessage *mailos.Email
		for _, msg := range mockServer.Messages {
			if len(msg.Attachments) > 0 {
				attachmentMessage = msg
				break
			}
		}
		
		helpers.AssertNotNil(t, attachmentMessage, "Should find a message with attachments")
		helpers.AssertAttachmentsValid(t, attachmentMessage)
	})
}