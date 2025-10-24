package mailos_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	
	mailos "github.com/anduimagui/emailos-cli"
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

func TestEmailMessageWithAttachments(t *testing.T) {
	// Test EmailMessage struct with attachments
	msg := &mailos.EmailMessage{
		To:          []string{"test@example.com"},
		Subject:     "Test with attachments",
		Body:        "Test email body",
		Attachments: []string{"file1.txt", "file2.pdf"},
	}
	
	if len(msg.Attachments) != 2 {
		t.Errorf("Expected 2 attachments, got %d", len(msg.Attachments))
	}
	
	expectedFiles := []string{"file1.txt", "file2.pdf"}
	for i, attachment := range msg.Attachments {
		if attachment != expectedFiles[i] {
			t.Errorf("Expected attachment %s, got %s", expectedFiles[i], attachment)
		}
	}
	
	fmt.Println("✓ EmailMessage attachment structure test passed")
}

func TestAttachmentFileHandling(t *testing.T) {
	// Create temporary test files
	tempDir, err := ioutil.TempDir("", "mailos_test_attachments")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create test files
	testFiles := []struct {
		name    string
		content string
	}{
		{"test.txt", "This is a test text file"},
		{"test.pdf", "This simulates PDF content"},
		{"image.jpg", "This simulates JPEG image data"},
		{"document.docx", "This simulates Word document content"},
		{"spreadsheet.xlsx", "This simulates Excel spreadsheet content"},
	}
	
	var attachmentPaths []string
	for _, file := range testFiles {
		path := filepath.Join(tempDir, file.name)
		err := ioutil.WriteFile(path, []byte(file.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file.name, err)
		}
		attachmentPaths = append(attachmentPaths, path)
	}
	
	// Test EmailMessage with real file paths
	msg := &mailos.EmailMessage{
		To:          []string{"andrew@happysoft.dev"},
		Subject:     "Test with multiple attachments",
		Body:        "This email has multiple attachments for testing",
		Attachments: attachmentPaths,
	}
	
	if len(msg.Attachments) != len(testFiles) {
		t.Errorf("Expected %d attachments, got %d", len(testFiles), len(msg.Attachments))
	}
	
	// Verify files exist and can be read
	for _, path := range attachmentPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Attachment file does not exist: %s", path)
		}
		
		content, err := ioutil.ReadFile(path)
		if err != nil {
			t.Errorf("Failed to read attachment file %s: %v", path, err)
		}
		
		if len(content) == 0 {
			t.Errorf("Attachment file %s is empty", path)
		}
	}
	
	fmt.Println("✓ Attachment file handling test passed")
}

func TestLargeAttachmentHandling(t *testing.T) {
	// Create temporary directory
	tempDir, err := ioutil.TempDir("", "mailos_large_attachment_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a larger test file (1MB)
	largeContent := strings.Repeat("This is test data for a large attachment file. ", 20000)
	largePath := filepath.Join(tempDir, "large_file.txt")
	err = ioutil.WriteFile(largePath, []byte(largeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}
	
	// Test with large attachment
	msg := &mailos.EmailMessage{
		To:          []string{"andrew@happysoft.dev"},
		Subject:     "Test with large attachment",
		Body:        "This email has a large attachment for testing",
		Attachments: []string{largePath},
	}
	
	// Verify message structure
	if len(msg.Attachments) != 1 {
		t.Errorf("Expected 1 attachment, got %d", len(msg.Attachments))
	}
	
	// Verify the large file can be handled
	content, err := ioutil.ReadFile(largePath)
	if err != nil {
		t.Fatalf("Failed to read large attachment: %v", err)
	}
	
	if len(content) < 900000 { // Should be close to 1MB
		t.Errorf("Large attachment seems too small: %d bytes", len(content))
	}
	
	fmt.Println("✓ Large attachment handling test passed")
}

func TestMultipleAttachmentsLimit(t *testing.T) {
	// Create temporary directory
	tempDir, err := ioutil.TempDir("", "mailos_multiple_attachments_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create 10 test files
	var attachmentPaths []string
	for i := 0; i < 10; i++ {
		filename := fmt.Sprintf("test_file_%d.txt", i+1)
		content := fmt.Sprintf("This is test file number %d", i+1)
		path := filepath.Join(tempDir, filename)
		
		err := ioutil.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
		attachmentPaths = append(attachmentPaths, path)
	}
	
	// Test EmailMessage with 10 attachments
	msg := &mailos.EmailMessage{
		To:          []string{"andrew@happysoft.dev"},
		Subject:     "Test with 10 attachments",
		Body:        "This email has 10 attachments for testing",
		Attachments: attachmentPaths,
	}
	
	if len(msg.Attachments) != 10 {
		t.Errorf("Expected 10 attachments, got %d", len(msg.Attachments))
	}
	
	// Verify all files exist
	for i, path := range attachmentPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Attachment file %d does not exist: %s", i+1, path)
		}
	}
	
	fmt.Printf("✓ Multiple attachments test passed (10 files)\n")
}

func TestRealImageAttachments(t *testing.T) {
	// Test with real image files from test_attachments/images/
	testImagePaths := []string{
		"../../test_attachments/images/life-quote-image.jpg",
		"../../test_attachments/images/raggle.png", 
		"../../test_attachments/images/wolf.webp",
	}
	
	// Verify image files exist
	for _, imagePath := range testImagePaths {
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			t.Skipf("Skipping test - image file not found: %s", imagePath)
		}
	}
	
	// Test EmailMessage with real image attachments
	msg := &mailos.EmailMessage{
		To:          []string{"andrew@happysoft.dev"},
		Subject:     "Test with real image attachments",
		Body:        "This email has real image attachments: JPG, PNG, and WebP",
		Attachments: testImagePaths,
	}
	
	if len(msg.Attachments) != 3 {
		t.Errorf("Expected 3 image attachments, got %d", len(msg.Attachments))
	}
	
	// Verify images can be read and have content
	for _, imagePath := range testImagePaths {
		content, err := ioutil.ReadFile(imagePath)
		if err != nil {
			t.Errorf("Failed to read image file %s: %v", imagePath, err)
			continue
		}
		
		if len(content) == 0 {
			t.Errorf("Image file %s is empty", imagePath)
		}
		
		// Check file extension and content
		ext := strings.ToLower(filepath.Ext(imagePath))
		switch ext {
		case ".jpg", ".jpeg":
			// JPEG files should start with FFD8
			if len(content) >= 2 && !(content[0] == 0xFF && content[1] == 0xD8) {
				t.Errorf("File %s doesn't appear to be a valid JPEG", imagePath)
			}
		case ".png":
			// PNG files should start with PNG signature
			if len(content) >= 8 && string(content[1:4]) != "PNG" {
				t.Errorf("File %s doesn't appear to be a valid PNG", imagePath)
			}
		case ".webp":
			// WebP files should contain "WEBP" signature
			if len(content) >= 12 && string(content[8:12]) != "WEBP" {
				t.Errorf("File %s doesn't appear to be a valid WebP", imagePath)
			}
		}
		
		fmt.Printf("✓ Image file validated: %s (%d bytes)\n", filepath.Base(imagePath), len(content))
	}
	
	fmt.Println("✓ Real image attachments test passed")
}