package mailos

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGroupManagement(t *testing.T) {
	tmpDir := setupTestGroups(t)
	defer cleanupTestGroups(tmpDir)

	t.Run("CreateGroup", func(t *testing.T) {
		err := UpdateGroup("test-contacts", "Test contacts for EmailOS functionality", "andrew@happysoft.dev,andremg212@gmail.com")
		if err != nil {
			t.Fatalf("Failed to create group: %v", err)
		}
	})

	t.Run("ListGroups", func(t *testing.T) {
		err := ListGroups()
		if err != nil {
			t.Fatalf("Failed to list groups: %v", err)
		}
	})

	t.Run("GetGroup", func(t *testing.T) {
		group, err := GetGroup("test-contacts")
		if err != nil {
			t.Fatalf("Failed to get group: %v", err)
		}
		if group.Name != "test-contacts" {
			t.Errorf("Expected group name 'test-contacts', got '%s'", group.Name)
		}
		if len(group.Emails) != 2 {
			t.Errorf("Expected 2 emails, got %d", len(group.Emails))
		}
	})

	t.Run("AddMemberToGroup", func(t *testing.T) {
		err := AddMemberToGroup("test-contacts", "newmember@example.com")
		if err != nil {
			t.Fatalf("Failed to add member to group: %v", err)
		}

		group, err := GetGroup("test-contacts")
		if err != nil {
			t.Fatalf("Failed to get group after adding member: %v", err)
		}
		if len(group.Emails) != 3 {
			t.Errorf("Expected 3 emails after adding member, got %d", len(group.Emails))
		}
	})

	t.Run("RemoveMemberFromGroup", func(t *testing.T) {
		err := RemoveMemberFromGroup("test-contacts", "newmember@example.com")
		if err != nil {
			t.Fatalf("Failed to remove member from group: %v", err)
		}

		group, err := GetGroup("test-contacts")
		if err != nil {
			t.Fatalf("Failed to get group after removing member: %v", err)
		}
		if len(group.Emails) != 2 {
			t.Errorf("Expected 2 emails after removing member, got %d", len(group.Emails))
		}
	})

	t.Run("CreateSecondGroup", func(t *testing.T) {
		err := UpdateGroup("dev-team", "Development team for testing", "dev1@example.com,dev2@example.com,dev3@example.com")
		if err != nil {
			t.Fatalf("Failed to create second group: %v", err)
		}
	})

	t.Run("ProcessGroupsForSending", func(t *testing.T) {
		groupNames := []string{"test-contacts"}
		existingEmails := []string{"additional@example.com"}
		
		result, err := ProcessGroupsForSending(groupNames, existingEmails)
		if err != nil {
			t.Fatalf("Failed to process groups for sending: %v", err)
		}
		
		if len(result) != 3 {
			t.Errorf("Expected 3 total emails (2 from group + 1 individual), got %d", len(result))
		}
	})

	t.Run("DuplicateEmailRemoval", func(t *testing.T) {
		groupNames := []string{"test-contacts"}
		existingEmails := []string{"andrew@happysoft.dev"}
		
		result, err := ProcessGroupsForSending(groupNames, existingEmails)
		if err != nil {
			t.Fatalf("Failed to process groups with duplicate emails: %v", err)
		}
		
		if len(result) != 2 {
			t.Errorf("Expected 2 unique emails after duplicate removal, got %d", len(result))
		}
	})

	t.Run("UpdateGroup", func(t *testing.T) {
		err := UpdateGroup("test-contacts", "Updated test contacts with additional email", "andrew@happysoft.dev,andremg212@gmail.com,updated@example.com")
		if err != nil {
			t.Fatalf("Failed to update group: %v", err)
		}

		group, err := GetGroup("test-contacts")
		if err != nil {
			t.Fatalf("Failed to get updated group: %v", err)
		}
		if len(group.Emails) != 3 {
			t.Errorf("Expected 3 emails after update, got %d", len(group.Emails))
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		_, err := GetGroup("non-existent-group")
		if err == nil {
			t.Error("Expected error for non-existent group, got nil")
		}

		err = AddMemberToGroup("non-existent-group", "test@example.com")
		if err == nil {
			t.Error("Expected error when adding to non-existent group, got nil")
		}

		err = RemoveMemberFromGroup("test-contacts", "non-existent@example.com")
		if err == nil {
			t.Error("Expected error when removing non-existent member, got nil")
		}

		err = AddMemberToGroup("test-contacts", "andrew@happysoft.dev")
		if err == nil {
			t.Error("Expected error when adding duplicate member, got nil")
		}
	})

	t.Run("DeleteGroup", func(t *testing.T) {
		err := DeleteGroup("dev-team")
		if err != nil {
			t.Fatalf("Failed to delete group: %v", err)
		}

		_, err = GetGroup("dev-team")
		if err == nil {
			t.Error("Expected error for deleted group, got nil")
		}
	})

	t.Run("ListGroupMembers", func(t *testing.T) {
		err := ListGroupMembers("test-contacts")
		if err != nil {
			t.Fatalf("Failed to list group members: %v", err)
		}
	})

	t.Run("CleanupTestGroup", func(t *testing.T) {
		err := DeleteGroup("test-contacts")
		if err != nil {
			t.Fatalf("Failed to cleanup test group: %v", err)
		}
	})
}

func TestAdvancedGroupOperations(t *testing.T) {
	tmpDir := setupTestGroups(t)
	defer cleanupTestGroups(tmpDir)

	t.Run("MultipleGroupOperations", func(t *testing.T) {
		groups := []struct {
			name        string
			description string
			emails      string
		}{
			{"marketing", "Marketing team", "marketing1@example.com,marketing2@example.com"},
			{"sales", "Sales team", "sales1@example.com,sales2@example.com"},
			{"support", "Support team", "support1@example.com,support2@example.com"},
		}

		for _, group := range groups {
			err := UpdateGroup(group.name, group.description, group.emails)
			if err != nil {
				t.Fatalf("Failed to create group %s: %v", group.name, err)
			}
		}

		for _, group := range groups {
			g, err := GetGroup(group.name)
			if err != nil {
				t.Fatalf("Failed to retrieve group %s: %v", group.name, err)
			}
			if len(g.Emails) != 2 {
				t.Errorf("Group %s should have 2 emails, got %d", group.name, len(g.Emails))
			}
		}

		for _, group := range groups {
			err := DeleteGroup(group.name)
			if err != nil {
				t.Fatalf("Failed to delete group %s: %v", group.name, err)
			}
		}
	})

	t.Run("LargeGroupManagement", func(t *testing.T) {
		var emails []string
		for i := 1; i <= 100; i++ {
			emails = append(emails, fmt.Sprintf("user%d@example.com", i))
		}

		err := UpdateGroup("large-group", "Large test group", strings.Join(emails, ","))
		if err != nil {
			t.Fatalf("Failed to create large group: %v", err)
		}

		group, err := GetGroup("large-group")
		if err != nil {
			t.Fatalf("Failed to get large group: %v", err)
		}
		if len(group.Emails) != 100 {
			t.Errorf("Expected 100 emails in large group, got %d", len(group.Emails))
		}

		for i := 1; i <= 10; i++ {
			err := RemoveMemberFromGroup("large-group", fmt.Sprintf("user%d@example.com", i))
			if err != nil {
				t.Fatalf("Failed to remove member from large group: %v", err)
			}
		}

		group, err = GetGroup("large-group")
		if err != nil {
			t.Fatalf("Failed to get large group after removals: %v", err)
		}
		if len(group.Emails) != 90 {
			t.Errorf("Expected 90 emails after removals, got %d", len(group.Emails))
		}

		err = DeleteGroup("large-group")
		if err != nil {
			t.Fatalf("Failed to delete large group: %v", err)
		}
	})

	t.Run("EmailValidation", func(t *testing.T) {
		err := UpdateGroup("validation-test", "Validation test group", "valid@example.com,invalid-email,another@valid.com")
		if err != nil {
			t.Fatalf("Failed to create validation test group: %v", err)
		}

		group, err := GetGroup("validation-test")
		if err != nil {
			t.Fatalf("Failed to get validation test group: %v", err)
		}

		if len(group.Emails) != 2 {
			t.Errorf("Expected 2 valid emails after validation, got %d", len(group.Emails))
		}

		err = DeleteGroup("validation-test")
		if err != nil {
			t.Fatalf("Failed to delete validation test group: %v", err)
		}
	})
}

func TestLiveEmailGroupSending(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live email test in short mode")
	}

	if os.Getenv("EMAILOS_LIVE_TEST") != "true" {
		t.Skip("Skipping live email test (set EMAILOS_LIVE_TEST=true to enable)")
	}

	tmpDir := setupTestGroups(t)
	defer cleanupTestGroups(tmpDir)

	t.Run("LiveEmailSending", func(t *testing.T) {
		err := UpdateGroup("live-test", "Live test group for actual email sending", "andrew@happysoft.dev,andremg212@gmail.com")
		if err != nil {
			t.Fatalf("Failed to create live test group: %v", err)
		}

		timestamp := time.Now().Format("2006-01-02 15:04:05")
		msg := &EmailMessage{
			Subject: fmt.Sprintf("EmailOS Groups Live Test - %s", timestamp),
			Body: fmt.Sprintf(`This is a live test of the EmailOS groups functionality. This email was sent to the 'live-test' group which includes andrew@happysoft.dev and andremg212@gmail.com.

The groups feature allows sending emails to multiple recipients using a single --group parameter, perfect for cold email campaigns and bulk messaging.

Test timestamp: %s`, timestamp),
		}

		groupEmails, err := GetGroupEmails("live-test")
		if err != nil {
			t.Fatalf("Failed to get live test group emails: %v", err)
		}

		msg.To = groupEmails

		err = Send(msg)
		if err != nil {
			t.Fatalf("Failed to send live test email: %v", err)
		}

		err = DeleteGroup("live-test")
		if err != nil {
			t.Fatalf("Failed to cleanup live test group: %v", err)
		}

		t.Log("Live email test completed successfully!")
		t.Logf("Real emails were sent to: %v", groupEmails)
	})
}

func setupTestGroups(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "emailos_groups_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	originalHome := os.Getenv("HOME")
	testEmailDir := filepath.Join(tmpDir, ".email")
	err = os.MkdirAll(testEmailDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test email directory: %v", err)
	}

	os.Setenv("HOME", tmpDir)

	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})

	return tmpDir
}

func cleanupTestGroups(tmpDir string) {
	os.RemoveAll(tmpDir)
}