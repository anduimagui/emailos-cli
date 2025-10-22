package mailos_test

import (
	"fmt"
	"testing"
	"time"
	
	mailos "github.com/anduimagui/emailos"
)

func TestAttachmentData(t *testing.T) {
	// Create a test email with attachment data
	email := &mailos.Email{
		ID:      1,
		From:    "test@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test Email with Attachment",
		Date:    time.Now(),
		Body:    "This is a test email with an attachment",
		Attachments: []string{"test.pdf"},
		AttachmentData: map[string][]byte{
			"test.pdf": []byte("PDF content here"),
		},
	}
	
	// Verify attachment data is stored
	if len(email.AttachmentData) != 1 {
		t.Errorf("Expected 1 attachment data, got %d", len(email.AttachmentData))
	}
	
	if data, ok := email.AttachmentData["test.pdf"]; !ok {
		t.Error("test.pdf not found in AttachmentData")
	} else if string(data) != "PDF content here" {
		t.Errorf("Unexpected attachment data: %s", string(data))
	}
	
	fmt.Println("✓ Attachment data structure test passed")
}