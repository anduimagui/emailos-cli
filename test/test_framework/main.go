package main

import (
	"fmt"
	"os"
	"strings"
)

// CommandTest represents a test case for any mailos command
type CommandTest struct {
	Name        string
	Command     string
	Description string
	RequiresEnv bool
	Category    string
	Command_Type string // "send", "read", "search", etc.
}

// TestSuite contains all command test cases
var AllTests = []CommandTest{}

// Initialize all test suites
func init() {
	AllTests = append(AllTests, SendTestSuite...)
	AllTests = append(AllTests, ReadTestSuite...)
	AllTests = append(AllTests, GroupsTestSuite...)
}

// SendTestSuite contains all send command test cases
var SendTestSuite = []CommandTest{
	// Basic send tests
	{
		Name:         "Send help",
		Command:      "./mailos send --help",
		Description:  "Test send command help output",
		RequiresEnv:  false,
		Category:     "help",
		Command_Type: "send",
	},

	// Basic send functionality tests
	{
		Name:         "Send basic syntax test",
		Command:      "./mailos send --to $TO_EMAIL --subject 'Test' --body 'Test message'",
		Description:  "Basic send command with required flags",
		RequiresEnv:  true,
		Category:     "basic",
		Command_Type: "send",
	},
	{
		Name:         "Send with --from flag",
		Command:      "./mailos send --to $TO_EMAIL --from $FROM_EMAIL --subject 'Test' --body 'Test'",
		Description:  "Send with explicit from address",
		RequiresEnv:  true,
		Category:     "basic",
		Command_Type: "send",
	},
	{
		Name:         "Send with --message flag (alias for --body)",
		Command:      "./mailos send --to $TO_EMAIL --subject 'Test' --message 'Test message'",
		Description:  "Send using --message flag instead of --body",
		RequiresEnv:  true,
		Category:     "basic",
		Command_Type: "send",
	},

	// Recipient tests
	{
		Name:         "Send with CC",
		Command:      "./mailos send --to $TO_EMAIL --cc cc@example.com --subject 'Test' --body 'Test'",
		Description:  "Send with CC recipients",
		RequiresEnv:  true,
		Category:     "recipients",
		Command_Type: "send",
	},
	{
		Name:         "Send with BCC",
		Command:      "./mailos send --to $TO_EMAIL --bcc bcc@example.com --subject 'Test' --body 'Test'",
		Description:  "Send with BCC recipients",
		RequiresEnv:  true,
		Category:     "recipients",
		Command_Type: "send",
	},
	{
		Name:         "Send with multiple CC",
		Command:      "./mailos send --to $TO_EMAIL --cc cc1@example.com,cc2@example.com --subject 'Test' --body 'Test'",
		Description:  "Send with multiple CC recipients",
		RequiresEnv:  true,
		Category:     "recipients",
		Command_Type: "send",
	},
	{
		Name:         "Send with multiple BCC",
		Command:      "./mailos send --to $TO_EMAIL --bcc bcc1@example.com,bcc2@example.com --subject 'Test' --body 'Test'",
		Description:  "Send with multiple BCC recipients",
		RequiresEnv:  true,
		Category:     "recipients",
		Command_Type: "send",
	},
	{
		Name:         "Send with all recipient types",
		Command:      "./mailos send --to $TO_EMAIL --cc cc@example.com --bcc bcc@example.com --subject 'Test' --body 'Test'",
		Description:  "Send with TO, CC, and BCC recipients",
		RequiresEnv:  true,
		Category:     "recipients",
		Command_Type: "send",
	},

	// Format and content tests
	{
		Name:         "Send plain text",
		Command:      "./mailos send --to $TO_EMAIL --subject 'Test' --body 'Test' --plain",
		Description:  "Send as plain text (no markdown conversion)",
		RequiresEnv:  true,
		Category:     "format",
		Command_Type: "send",
	},
	{
		Name:         "Send with template",
		Command:      "./mailos send --to $TO_EMAIL --subject 'Test' --body 'Test' --template",
		Description:  "Send with HTML template applied",
		RequiresEnv:  true,
		Category:     "format",
		Command_Type: "send",
	},
	{
		Name:         "Send with file body",
		Command:      "./mailos send --to $TO_EMAIL --subject 'Test' --file nonexistent.txt",
		Description:  "Send with body read from file",
		RequiresEnv:  true,
		Category:     "format",
		Command_Type: "send",
	},

	// Signature tests
	{
		Name:         "Send no signature",
		Command:      "./mailos send --to $TO_EMAIL --subject 'Test' --body 'Test' --no-signature",
		Description:  "Send without signature",
		RequiresEnv:  true,
		Category:     "signature",
		Command_Type: "send",
	},
	{
		Name:         "Send custom signature",
		Command:      "./mailos send --to $TO_EMAIL --subject 'Test' --body 'Test' --signature 'Custom sig'",
		Description:  "Send with custom signature",
		RequiresEnv:  true,
		Category:     "signature",
		Command_Type: "send",
	},

	// Attachment tests
	{
		Name:         "Send with single attachment",
		Command:      "./mailos send --to $TO_EMAIL --subject 'Test' --body 'Test' --attach file.txt",
		Description:  "Send with single attachment",
		RequiresEnv:  true,
		Category:     "attachments",
		Command_Type: "send",
	},
	{
		Name:         "Send with multiple attachments",
		Command:      "./mailos send --to $TO_EMAIL --subject 'Test' --body 'Test' --attach file1.txt,file2.txt",
		Description:  "Send with multiple attachments",
		RequiresEnv:  true,
		Category:     "attachments",
		Command_Type: "send",
	},

	// Preview and dry-run tests
	{
		Name:         "Send preview",
		Command:      "./mailos send --to $TO_EMAIL --subject 'Test' --body 'Test' --preview",
		Description:  "Preview email without sending",
		RequiresEnv:  true,
		Category:     "preview",
		Command_Type: "send",
	},
	{
		Name:         "Send dry-run",
		Command:      "./mailos send --to $TO_EMAIL --subject 'Test' --body 'Test' --dry-run",
		Description:  "Dry-run email without sending",
		RequiresEnv:  true,
		Category:     "preview",
		Command_Type: "send",
	},

	// Verbose and debugging tests
	{
		Name:         "Send verbose",
		Command:      "./mailos send --to $TO_EMAIL --subject 'Test' --body 'Test' --verbose",
		Description:  "Send with verbose debugging output",
		RequiresEnv:  true,
		Category:     "debug",
		Command_Type: "send",
	},

	// Draft-related send tests
	{
		Name:         "Send drafts help",
		Command:      "./mailos send --drafts --help",
		Description:  "Test send drafts help output",
		RequiresEnv:  false,
		Category:     "drafts",
		Command_Type: "send",
	},
	{
		Name:         "Send all drafts",
		Command:      "./mailos send --drafts",
		Description:  "Send all draft emails",
		RequiresEnv:  true,
		Category:     "drafts",
		Command_Type: "send",
	},
	{
		Name:         "Send drafts dry-run",
		Command:      "./mailos send --drafts --dry-run",
		Description:  "Preview all drafts without sending",
		RequiresEnv:  true,
		Category:     "drafts",
		Command_Type: "send",
	},
	{
		Name:         "Send drafts with filter",
		Command:      "./mailos send --drafts --filter 'priority:high'",
		Description:  "Send filtered drafts",
		RequiresEnv:  true,
		Category:     "drafts",
		Command_Type: "send",
	},
	{
		Name:         "Send drafts with confirmation",
		Command:      "./mailos send --drafts --confirm",
		Description:  "Send drafts with confirmation prompts",
		RequiresEnv:  true,
		Category:     "drafts",
		Command_Type: "send",
	},
	{
		Name:         "Send drafts with custom directory",
		Command:      "./mailos send --drafts --draft-dir ./custom-drafts",
		Description:  "Send drafts from custom directory",
		RequiresEnv:  true,
		Category:     "drafts",
		Command_Type: "send",
	},
	{
		Name:         "Send drafts with log file",
		Command:      "./mailos send --drafts --log-file ./send.log",
		Description:  "Send drafts and log to file",
		RequiresEnv:  true,
		Category:     "drafts",
		Command_Type: "send",
	},
	{
		Name:         "Send drafts without delete-after",
		Command:      "./mailos send --drafts --delete-after=false",
		Description:  "Send drafts without deleting after send",
		RequiresEnv:  true,
		Category:     "drafts",
		Command_Type: "send",
	},

	// Short flag tests
	{
		Name:         "Send with short flags",
		Command:      "./mailos send -t $TO_EMAIL -s 'Test' -b 'Test message'",
		Description:  "Send using short flags (-t, -s, -b)",
		RequiresEnv:  true,
		Category:     "shortflags",
		Command_Type: "send",
	},
	{
		Name:         "Send with short CC flag",
		Command:      "./mailos send -t $TO_EMAIL -c cc@example.com -s 'Test' -b 'Test'",
		Description:  "Send using short CC flag (-c)",
		RequiresEnv:  true,
		Category:     "shortflags",
		Command_Type: "send",
	},
	{
		Name:         "Send with short BCC flag",
		Command:      "./mailos send -t $TO_EMAIL -B bcc@example.com -s 'Test' -b 'Test'",
		Description:  "Send using short BCC flag (-B)",
		RequiresEnv:  true,
		Category:     "shortflags",
		Command_Type: "send",
	},
	{
		Name:         "Send with short file flag",
		Command:      "./mailos send -t $TO_EMAIL -s 'Test' -f nonexistent.txt",
		Description:  "Send using short file flag (-f)",
		RequiresEnv:  true,
		Category:     "shortflags",
		Command_Type: "send",
	},
	{
		Name:         "Send with short attach flag",
		Command:      "./mailos send -t $TO_EMAIL -s 'Test' -b 'Test' -a file.txt",
		Description:  "Send using short attach flag (-a)",
		RequiresEnv:  true,
		Category:     "shortflags",
		Command_Type: "send",
	},
	{
		Name:         "Send with short plain flag",
		Command:      "./mailos send -t $TO_EMAIL -s 'Test' -b 'Test' -P",
		Description:  "Send using short plain flag (-P)",
		RequiresEnv:  true,
		Category:     "shortflags",
		Command_Type: "send",
	},
	{
		Name:         "Send with short no-signature flag",
		Command:      "./mailos send -t $TO_EMAIL -s 'Test' -b 'Test' -S",
		Description:  "Send using short no-signature flag (-S)",
		RequiresEnv:  true,
		Category:     "shortflags",
		Command_Type: "send",
	},
	{
		Name:         "Send with short verbose flag",
		Command:      "./mailos send -t $TO_EMAIL -s 'Test' -b 'Test' -v",
		Description:  "Send using short verbose flag (-v)",
		RequiresEnv:  true,
		Category:     "shortflags",
		Command_Type: "send",
	},
	{
		Name:         "Send with message alias flag",
		Command:      "./mailos send -t $TO_EMAIL -s 'Test' -m 'Test message'",
		Description:  "Send using short message flag (-m)",
		RequiresEnv:  true,
		Category:     "shortflags",
		Command_Type: "send",
	},

	// Error handling tests
	{
		Name:         "Send without recipients",
		Command:      "./mailos send --subject 'Test' --body 'Test'",
		Description:  "Error: Send without TO recipients",
		RequiresEnv:  false,
		Category:     "errors",
		Command_Type: "send",
	},
	{
		Name:         "Send without subject",
		Command:      "./mailos send --to $TO_EMAIL --body 'Test'",
		Description:  "Error: Send without subject",
		RequiresEnv:  true,
		Category:     "errors",
		Command_Type: "send",
	},
	{
		Name:         "Send nonexistent from account",
		Command:      "./mailos send --to $TO_EMAIL --from nonexistent@example.com --subject 'Test' --body 'Test'",
		Description:  "Error: Send from nonexistent account",
		RequiresEnv:  true,
		Category:     "errors",
		Command_Type: "send",
	},
	{
		Name:         "Send with nonexistent file",
		Command:      "./mailos send --to $TO_EMAIL --subject 'Test' --file /nonexistent/file.txt",
		Description:  "Error: Send with nonexistent file",
		RequiresEnv:  true,
		Category:     "errors",
		Command_Type: "send",
	},
}

// ReadTestSuite contains all read command test cases
var ReadTestSuite = []CommandTest{
	// Basic read tests
	{
		Name:         "Read help",
		Command:      "./mailos read --help",
		Description:  "Test read command help output",
		RequiresEnv:  false,
		Category:     "help",
		Command_Type: "read",
	},

	// Basic read functionality tests
	{
		Name:         "Read with positional ID",
		Command:      "./mailos read 1",
		Description:  "Read email using positional ID argument",
		RequiresEnv:  true,
		Category:     "basic",
		Command_Type: "read",
	},
	{
		Name:         "Read with --id flag",
		Command:      "./mailos read --id 1",
		Description:  "Read email using --id flag",
		RequiresEnv:  true,
		Category:     "basic",
		Command_Type: "read",
	},
	{
		Name:         "Read with higher ID",
		Command:      "./mailos read 1234",
		Description:  "Read email with higher ID number",
		RequiresEnv:  true,
		Category:     "basic",
		Command_Type: "read",
	},

	// Include documents tests
	{
		Name:         "Read with include-documents",
		Command:      "./mailos read 1 --include-documents",
		Description:  "Read email with document parsing enabled",
		RequiresEnv:  true,
		Category:     "documents",
		Command_Type: "read",
	},
	{
		Name:         "Read without include-documents",
		Command:      "./mailos read 1 --include-documents=false",
		Description:  "Read email with document parsing disabled",
		RequiresEnv:  true,
		Category:     "documents",
		Command_Type: "read",
	},

	// Account-specific tests
	{
		Name:         "Read with account flag",
		Command:      "./mailos read 1 --account $FROM_EMAIL",
		Description:  "Read email from specific account",
		RequiresEnv:  true,
		Category:     "account",
		Command_Type: "read",
	},

	// Combined flag tests
	{
		Name:         "Read with all flags",
		Command:      "./mailos read --id 1 --include-documents --account $FROM_EMAIL",
		Description:  "Read email with all flags combined",
		RequiresEnv:  true,
		Category:     "combined",
		Command_Type: "read",
	},
	{
		Name:         "Read with ID flag and documents disabled",
		Command:      "./mailos read --id 1 --include-documents=false",
		Description:  "Read with ID flag and documents disabled",
		RequiresEnv:  true,
		Category:     "combined",
		Command_Type: "read",
	},

	// Error handling tests
	{
		Name:         "Read without ID",
		Command:      "./mailos read",
		Description:  "Error: Read without email ID",
		RequiresEnv:  false,
		Category:     "errors",
		Command_Type: "read",
	},
	{
		Name:         "Read with invalid ID",
		Command:      "./mailos read abc",
		Description:  "Error: Read with invalid ID format",
		RequiresEnv:  false,
		Category:     "errors",
		Command_Type: "read",
	},
	{
		Name:         "Read with negative ID",
		Command:      "./mailos read -1",
		Description:  "Error: Read with negative ID",
		RequiresEnv:  false,
		Category:     "errors",
		Command_Type: "read",
	},
	{
		Name:         "Read with zero ID",
		Command:      "./mailos read 0",
		Description:  "Error: Read with zero ID",
		RequiresEnv:  false,
		Category:     "errors",
		Command_Type: "read",
	},
	{
		Name:         "Read with nonexistent ID",
		Command:      "./mailos read 999999",
		Description:  "Error: Read with nonexistent email ID",
		RequiresEnv:  true,
		Category:     "errors",
		Command_Type: "read",
	},
	{
		Name:         "Read with invalid account",
		Command:      "./mailos read 1 --account nonexistent@example.com",
		Description:  "Error: Read with invalid account",
		RequiresEnv:  true,
		Category:     "errors",
		Command_Type: "read",
	},
}

// GroupsTestSuite contains all groups command test cases
var GroupsTestSuite = []CommandTest{
	// Basic groups tests
	{
		Name:         "Groups help",
		Command:      "./mailos groups --help",
		Description:  "Test groups command help output",
		RequiresEnv:  false,
		Category:     "help",
		Command_Type: "groups",
	},

	// Basic groups functionality tests
	{
		Name:         "List groups (empty)",
		Command:      "./mailos groups",
		Description:  "List groups when no groups exist",
		RequiresEnv:  false,
		Category:     "basic",
		Command_Type: "groups",
	},
	{
		Name:         "Create basic group",
		Command:      "./mailos groups --update 'test-group' --emails 'andrew@happysoft.dev,andremg212@gmail.com' --description 'Test group'",
		Description:  "Create a basic email group",
		RequiresEnv:  false,
		Category:     "basic",
		Command_Type: "groups",
	},
	{
		Name:         "List groups (with content)",
		Command:      "./mailos groups",
		Description:  "List groups when groups exist",
		RequiresEnv:  false,
		Category:     "basic",
		Command_Type: "groups",
	},
	{
		Name:         "Delete group",
		Command:      "./mailos groups --delete 'test-group'",
		Description:  "Delete an existing group",
		RequiresEnv:  false,
		Category:     "basic",
		Command_Type: "groups",
	},

	// Member management tests
	{
		Name:         "Add member to group",
		Command:      "./mailos groups --group 'test-contacts' --add-member 'newuser@example.com'",
		Description:  "Add a member to an existing group",
		RequiresEnv:  false,
		Category:     "members",
		Command_Type: "groups",
	},
	{
		Name:         "Remove member from group",
		Command:      "./mailos groups --group 'test-contacts' --remove-member 'newuser@example.com'",
		Description:  "Remove a member from an existing group",
		RequiresEnv:  false,
		Category:     "members",
		Command_Type: "groups",
	},
	{
		Name:         "List group members",
		Command:      "./mailos groups --list-members 'test-contacts'",
		Description:  "List all members of a specific group",
		RequiresEnv:  false,
		Category:     "members",
		Command_Type: "groups",
	},

	// Advanced group operations
	{
		Name:         "Update group with new description",
		Command:      "./mailos groups --update 'test-contacts' --emails 'andrew@happysoft.dev,andremg212@gmail.com' --description 'Updated test contacts'",
		Description:  "Update existing group with new description",
		RequiresEnv:  false,
		Category:     "advanced",
		Command_Type: "groups",
	},
	{
		Name:         "Create group with many members",
		Command:      "./mailos groups --update 'large-group' --emails 'user1@example.com,user2@example.com,user3@example.com,user4@example.com,user5@example.com' --description 'Large group test'",
		Description:  "Create group with multiple members",
		RequiresEnv:  false,
		Category:     "advanced",
		Command_Type: "groups",
	},

	// Send to groups tests
	{
		Name:         "Send to group (dry-run)",
		Command:      "./mailos send --group 'test-contacts' --subject 'Group Test' --body 'Testing group sending' --dry-run",
		Description:  "Send email to group in dry-run mode",
		RequiresEnv:  true,
		Category:     "sending",
		Command_Type: "groups",
	},
	{
		Name:         "Send to group with individual recipients",
		Command:      "./mailos send --group 'test-contacts' --to 'individual@example.com' --subject 'Combined Test' --body 'Testing combined recipients' --dry-run",
		Description:  "Send to both group and individual recipients",
		RequiresEnv:  true,
		Category:     "sending",
		Command_Type: "groups",
	},
	{
		Name:         "Preview send to group",
		Command:      "./mailos send --group 'test-contacts' --subject 'Preview Test' --body 'Testing group preview' --preview",
		Description:  "Preview email sending to group",
		RequiresEnv:  true,
		Category:     "sending",
		Command_Type: "groups",
	},

	// Error handling tests
	{
		Name:         "Create group without name",
		Command:      "./mailos groups --update '' --emails 'test@example.com'",
		Description:  "Error: Create group without name",
		RequiresEnv:  false,
		Category:     "errors",
		Command_Type: "groups",
	},
	{
		Name:         "Create group without emails",
		Command:      "./mailos groups --update 'test-group' --emails ''",
		Description:  "Error: Create group without emails",
		RequiresEnv:  false,
		Category:     "errors",
		Command_Type: "groups",
	},
	{
		Name:         "Delete non-existent group",
		Command:      "./mailos groups --delete 'non-existent-group'",
		Description:  "Error: Delete non-existent group",
		RequiresEnv:  false,
		Category:     "errors",
		Command_Type: "groups",
	},
	{
		Name:         "Add member without group name",
		Command:      "./mailos groups --add-member 'test@example.com'",
		Description:  "Error: Add member without specifying group",
		RequiresEnv:  false,
		Category:     "errors",
		Command_Type: "groups",
	},
	{
		Name:         "Remove member without group name",
		Command:      "./mailos groups --remove-member 'test@example.com'",
		Description:  "Error: Remove member without specifying group",
		RequiresEnv:  false,
		Category:     "errors",
		Command_Type: "groups",
	},
	{
		Name:         "Add member to non-existent group",
		Command:      "./mailos groups --group 'non-existent' --add-member 'test@example.com'",
		Description:  "Error: Add member to non-existent group",
		RequiresEnv:  false,
		Category:     "errors",
		Command_Type: "groups",
	},
	{
		Name:         "Remove member from non-existent group",
		Command:      "./mailos groups --group 'non-existent' --remove-member 'test@example.com'",
		Description:  "Error: Remove member from non-existent group",
		RequiresEnv:  false,
		Category:     "errors",
		Command_Type: "groups",
	},
	{
		Name:         "List members of non-existent group",
		Command:      "./mailos groups --list-members 'non-existent-group'",
		Description:  "Error: List members of non-existent group",
		RequiresEnv:  false,
		Category:     "errors",
		Command_Type: "groups",
	},
	{
		Name:         "Send to non-existent group",
		Command:      "./mailos send --group 'non-existent-group' --subject 'Test' --body 'Test' --dry-run",
		Description:  "Error: Send to non-existent group",
		RequiresEnv:  true,
		Category:     "errors",
		Command_Type: "groups",
	},

	// Validation tests
	{
		Name:         "Create group with invalid emails",
		Command:      "./mailos groups --update 'validation-test' --emails 'valid@example.com,invalid-email,another@valid.com'",
		Description:  "Create group with mix of valid and invalid emails",
		RequiresEnv:  false,
		Category:     "validation",
		Command_Type: "groups",
	},
	{
		Name:         "Add invalid email to group",
		Command:      "./mailos groups --group 'test-contacts' --add-member 'invalid-email-format'",
		Description:  "Error: Add invalid email format to group",
		RequiresEnv:  false,
		Category:     "validation",
		Command_Type: "groups",
	},
	{
		Name:         "Add duplicate member to group",
		Command:      "./mailos groups --group 'test-contacts' --add-member 'andrew@happysoft.dev'",
		Description:  "Error: Add duplicate member to group",
		RequiresEnv:  false,
		Category:     "validation",
		Command_Type: "groups",
	},

	// Integration tests
	{
		Name:         "Create multiple groups",
		Command:      "./mailos groups --update 'marketing' --emails 'marketing@example.com' --description 'Marketing team'",
		Description:  "Create multiple groups for integration testing",
		RequiresEnv:  false,
		Category:     "integration",
		Command_Type: "groups",
	},
	{
		Name:         "Full workflow test",
		Command:      "./mailos groups --update 'workflow-test' --emails 'user1@example.com,user2@example.com' --description 'Workflow test group'",
		Description:  "Complete workflow: create, modify, use, delete group",
		RequiresEnv:  false,
		Category:     "integration",
		Command_Type: "groups",
	},
}

// Filter functions
func GetTestsByCommandType(commandType string) []CommandTest {
	var filtered []CommandTest
	for _, test := range AllTests {
		if test.Command_Type == commandType {
			filtered = append(filtered, test)
		}
	}
	return filtered
}

func GetTestsByCategory(category string) []CommandTest {
	var filtered []CommandTest
	for _, test := range AllTests {
		if test.Category == category {
			filtered = append(filtered, test)
		}
	}
	return filtered
}

func GetTestsByCommandTypeAndCategory(commandType, category string) []CommandTest {
	var filtered []CommandTest
	for _, test := range AllTests {
		if test.Command_Type == commandType && test.Category == category {
			filtered = append(filtered, test)
		}
	}
	return filtered
}

func GetTestsRequiringEnv() []CommandTest {
	var filtered []CommandTest
	for _, test := range AllTests {
		if test.RequiresEnv {
			filtered = append(filtered, test)
		}
	}
	return filtered
}

func GetTestsNotRequiringEnv() []CommandTest {
	var filtered []CommandTest
	for _, test := range AllTests {
		if !test.RequiresEnv {
			filtered = append(filtered, test)
		}
	}
	return filtered
}

func GetAllCommandTypes() []string {
	types := make(map[string]bool)
	for _, test := range AllTests {
		types[test.Command_Type] = true
	}
	
	var result []string
	for cmdType := range types {
		result = append(result, cmdType)
	}
	return result
}

func GetAllCategories() []string {
	categories := make(map[string]bool)
	for _, test := range AllTests {
		categories[test.Category] = true
	}
	
	var result []string
	for category := range categories {
		result = append(result, category)
	}
	return result
}

// Output functions for shell script consumption
func OutputTestsForShell(commandType, category string, requiresEnv *bool) {
	var tests []CommandTest
	
	if commandType != "" && category != "" {
		tests = GetTestsByCommandTypeAndCategory(commandType, category)
	} else if commandType != "" {
		tests = GetTestsByCommandType(commandType)
	} else if category != "" {
		tests = GetTestsByCategory(category)
	} else if requiresEnv != nil {
		if *requiresEnv {
			tests = GetTestsRequiringEnv()
		} else {
			tests = GetTestsNotRequiringEnv()
		}
	} else {
		tests = AllTests
	}
	
	for _, test := range tests {
		fmt.Printf("%s|%s\n", test.Name, test.Command)
	}
}

func PrintTestSummary() {
	fmt.Printf("Mailos Test Suite Summary:\n")
	fmt.Printf("==========================\n")
	fmt.Printf("Total tests: %d\n", len(AllTests))
	
	commandTypes := GetAllCommandTypes()
	fmt.Printf("Command types: %v\n", commandTypes)
	
	for _, cmdType := range commandTypes {
		tests := GetTestsByCommandType(cmdType)
		fmt.Printf("  %s: %d tests\n", cmdType, len(tests))
	}
	
	categories := GetAllCategories()
	fmt.Printf("Categories: %v\n", categories)
	
	for _, category := range categories {
		tests := GetTestsByCategory(category)
		fmt.Printf("  %s: %d tests\n", category, len(tests))
	}
	
	envTests := GetTestsRequiringEnv()
	noEnvTests := GetTestsNotRequiringEnv()
	fmt.Printf("Tests requiring environment: %d\n", len(envTests))
	fmt.Printf("Tests not requiring environment: %d\n", len(noEnvTests))
}

func main() {
	if len(os.Args) < 2 {
		PrintTestSummary()
		return
	}

	command := os.Args[1]
	
	switch command {
	case "summary":
		PrintTestSummary()
		
	case "list":
		if len(os.Args) > 2 {
			commandType := os.Args[2]
			category := ""
			if len(os.Args) > 3 {
				category = os.Args[3]
			}
			OutputTestsForShell(commandType, category, nil)
		} else {
			OutputTestsForShell("", "", nil)
		}
		
	case "help":
		if len(os.Args) > 2 {
			commandType := os.Args[2]
			tests := GetTestsByCommandTypeAndCategory(commandType, "help")
			for _, test := range tests {
				fmt.Printf("%s|%s\n", test.Name, test.Command)
			}
		} else {
			tests := GetTestsByCategory("help")
			for _, test := range tests {
				fmt.Printf("%s|%s\n", test.Name, test.Command)
			}
		}
		
	case "errors":
		if len(os.Args) > 2 {
			commandType := os.Args[2]
			tests := GetTestsByCommandTypeAndCategory(commandType, "errors")
			for _, test := range tests {
				fmt.Printf("%s|%s\n", test.Name, test.Command)
			}
		} else {
			tests := GetTestsByCategory("errors")
			for _, test := range tests {
				fmt.Printf("%s|%s\n", test.Name, test.Command)
			}
		}
		
	case "no-env":
		if len(os.Args) > 2 {
			commandType := os.Args[2]
			tests := GetTestsByCommandType(commandType)
			var filtered []CommandTest
			for _, test := range tests {
				if !test.RequiresEnv {
					filtered = append(filtered, test)
				}
			}
			for _, test := range filtered {
				fmt.Printf("%s|%s\n", test.Name, test.Command)
			}
		} else {
			requiresEnv := false
			OutputTestsForShell("", "", &requiresEnv)
		}
		
	case "env":
		if len(os.Args) > 2 {
			commandType := os.Args[2]
			tests := GetTestsByCommandType(commandType)
			var filtered []CommandTest
			for _, test := range tests {
				if test.RequiresEnv {
					filtered = append(filtered, test)
				}
			}
			for _, test := range filtered {
				fmt.Printf("%s|%s\n", test.Name, test.Command)
			}
		} else {
			requiresEnv := true
			OutputTestsForShell("", "", &requiresEnv)
		}
		
	default:
		// Check if it's a command type
		commandTypes := GetAllCommandTypes()
		for _, cmdType := range commandTypes {
			if command == cmdType {
				if len(os.Args) > 2 {
					category := os.Args[2]
					OutputTestsForShell(cmdType, category, nil)
				} else {
					OutputTestsForShell(cmdType, "", nil)
				}
				return
			}
		}
		
		// Check if it's a category
		categories := GetAllCategories()
		for _, category := range categories {
			if command == category {
				OutputTestsForShell("", category, nil)
				return
			}
		}
		
		fmt.Printf("Usage: %s [summary|list|help|errors|no-env|env|<command_type>|<category>] [<command_type>] [<category>]\n", os.Args[0])
		fmt.Printf("Available command types: %s\n", strings.Join(commandTypes, ", "))
		fmt.Printf("Available categories: %s\n", strings.Join(categories, ", "))
	}
}