package main

import (
	"fmt"
	"strings"
	"github.com/spf13/cobra"
)

// CommandFlagMap defines the available flags for each command
var CommandFlagMap = map[string][]string{
	"read": {
		"include-documents", "id",
	},
	"search": {
		"number", "n", "unread", "u", "from", "to", "subject", "days", "range", 
		"json", "save-markdown", "output-dir", "download-attachments", "attachment-dir",
		"query", "q", "fuzzy-threshold", "no-fuzzy", "case-sensitive", "min-size", 
		"max-size", "has-attachments", "attachment-size", "date-range",
	},
	"send": {
		"to", "t", "cc", "c", "bcc", "B", "subject", "s", "body", "b", "message", "m",
		"file", "f", "attach", "a", "plain", "P", "no-signature", "S", "signature",
		"from", "preview", "template", "verbose", "v", "drafts", "draft-dir", "dry-run",
		"filter", "confirm", "delete-after", "log-file",
	},
	"sent": {
		"number", "n", "to", "subject", "days", "range", "json", "save-markdown", "output-dir",
	},
	"stats": {
		"number", "n", "unread", "u", "from", "to", "subject", "days", "range",
	},
	"download": {
		"number", "n", "id", "from", "to", "subject", "days", "output-dir", "show-content",
	},
	"delete": {
		"ids", "from", "subject", "drafts", "before", "after", "days", "confirm",
	},
	"reply": {
		"all", "body", "subject", "file", "f", "draft", "interactive", "i", "to", "cc", "bcc",
	},
	"forward": {
		"body", "subject", "file", "f", "draft", "interactive", "i", "to", "cc", "bcc",
	},
	"accounts": {
		"set", "add", "provider", "use-existing-credentials", "set-signature", "clear", 
		"list", "sync-fastmail", "token", "test-connection",
	},
	"configure": {
		"quick", "local", "email", "provider", "name", "from", "ai",
	},
	"mark-read": {
		"ids", "from", "subject",
	},
	"open": {
		"id", "from", "subject", "last",
	},
	"unsubscribe": {
		"from", "subject", "number", "n", "open", "auto-open",
	},
	"sync": {
		"dir", "limit", "days", "include-read", "verbose", "v",
	},
	"sync-db": {
		"account", "all",
	},
	"draft": {
		"list", "l", "read", "r", "edit-uid", "template", "data", "output", 
		"interactive", "i", "ai", "count", "n", "to", "t", "cc", "c", "bcc", "B",
		"subject", "s", "body", "b", "file", "f", "attach", "a", "priority",
		"plain", "P", "no-signature", "S", "signature",
	},
	"report": {
		"range", "output",
	},
	"test": {
		"interactive", "verbose",
	},
	"template": {
		"remove", "open-browser",
	},
	"uninstall": {
		"force", "keep-emails", "keep-config", "dry-run", "quiet", "backup", "backup-path",
	},
	"cleanup": {
		"quiet",
	},
	"commands": {
		"verbose",
	},
}

// CommandSuggestions provides command suggestions for common flag usage patterns
var CommandSuggestions = map[string]map[string]string{
	"read": {
		"number":  "Use 'mailos search --number N' to show N emails, then 'mailos read <id>' for specific email",
		"n":       "Use 'mailos search -n N' to show N emails, then 'mailos read <id>' for specific email", 
		"from":    "Use 'mailos search --from <email>' to find emails, then 'mailos read <id>' for specific email",
		"to":      "Use 'mailos search --to <email>' to find emails, then 'mailos read <id>' for specific email",
		"subject": "Use 'mailos search --subject <text>' to find emails, then 'mailos read <id>' for specific email",
		"unread":  "Use 'mailos search --unread' to find unread emails, then 'mailos read <id>' for specific email",
		"u":       "Use 'mailos search -u' to find unread emails, then 'mailos read <id>' for specific email",
		"days":    "Use 'mailos search --days N' to find recent emails, then 'mailos read <id>' for specific email",
		"range":   "Use 'mailos search --range <range>' to find emails in timeframe, then 'mailos read <id>' for specific email",
	},
	"accounts": {
		"create": "Use 'mailos accounts --add <email>' to add a new account",
		"new":    "Use 'mailos accounts --add <email>' to add a new account",
		"switch": "Use 'mailos accounts --set <email>' to switch default account",
		"config": "Use 'mailos configure' for email configuration, 'mailos accounts' for account management",
	},
	"configure": {
		"account": "Use 'mailos accounts' for account management, 'mailos configure' for settings",
		"setup":   "Use 'mailos setup' for initial configuration, 'mailos configure' for changes",
		"provider-setup": "Use 'mailos configure --provider <name>' to set up specific email provider",
	},
	"send": {
		"recipient": "Use 'mailos send --to <email>' to specify recipients",
		"message":   "Use 'mailos send --body <text>' or 'mailos send --file <path>' for message content",
		"draft":     "Use 'mailos draft create' to create drafts, 'mailos send' to send emails",
	},
	"delete": {
		"all":     "Use 'mailos search' to find emails, then 'mailos delete --ids <list>' for specific deletion",
		"bulk":    "Use 'mailos delete --from <email>' or 'mailos delete --days <N>' for bulk operations",
		"remove":  "Use 'mailos delete --ids <list>' to delete specific emails by ID",
	},
	"draft": {
		"create": "Use 'mailos draft create' to create new drafts",
		"edit":   "Use 'mailos draft edit <number>' to edit existing drafts", 
		"send":   "Use 'mailos send --drafts' to send all drafts, or 'mailos send' for individual emails",
	},
}

// FlagAliases maps common alternative flag names to their correct equivalents
var FlagAliases = map[string]map[string]string{
	"send": {
		"recipient":    "to",
		"recipients":   "to",
		"dest":         "to",
		"destination":  "to",
		"email":        "to",
		"msg":          "body",
		"message":      "body",
		"content":      "body",
		"text":         "body",
		"attachment":   "attach",
		"attachments":  "attach",
		"files":        "attach",
		"copy":         "cc",
		"blind-copy":   "bcc",
		"title":        "subject",
	},
	"search": {
		"sender":       "from",
		"author":       "from",
		"recipient":    "to",
		"dest":         "to",
		"title":        "subject",
		"limit":        "number",
		"count":        "number",
		"max":          "number",
		"recent":       "days",
		"since":        "days",
	},
	"read": {
		"email-id":     "id",
		"message-id":   "id",
		"mail-id":      "id",
		"docs":         "include-documents",
		"documents":    "include-documents",
		"attachments":  "include-documents",
	},
	"delete": {
		"sender":       "from",
		"author":       "from",
		"title":        "subject",
		"age":          "days",
		"older-than":   "days",
		"force":        "confirm",
	},
	"accounts": {
		"create":       "add",
		"new":          "add",
		"register":     "add",
		"switch":       "set",
		"change":       "set",
		"select":       "set",
		"remove":       "clear",
		"delete":       "clear",
		"show":         "list",
		"display":      "list",
	},
	"configure": {
		"setup":        "quick",
		"wizard":       "quick",
		"account":      "email",
		"user":         "email",
		"service":      "provider",
		"display":      "name",
		"sender":       "from",
	},
}

// tryFlagAlias attempts to find an alias for the unknown flag and execute the command with the correct flag
func tryFlagAlias(cmd *cobra.Command, flagName string, args []string) error {
	cmdName := cmd.Name()
	
	// Check if we have aliases for this command
	if aliases, exists := FlagAliases[cmdName]; exists {
		if correctFlag, hasAlias := aliases[flagName]; hasAlias {
			// Find the original flag value from the args
			flagValue := ""
			for i, arg := range args {
				if arg == "--"+flagName && i+1 < len(args) {
					flagValue = args[i+1]
					break
				}
			}
			
			// Provide helpful message about the alias
			msg := fmt.Sprintf("💡 Flag alias detected: --%s is equivalent to --%s", flagName, correctFlag)
			if flagValue != "" {
				msg += fmt.Sprintf("\n📝 Try: mailos %s --%s %s", cmdName, correctFlag, flagValue)
			} else {
				msg += fmt.Sprintf("\n📝 Try: mailos %s --%s <value>", cmdName, correctFlag)
			}
			msg += fmt.Sprintf("\n\nUse 'mailos %s --help' for all available flags.", cmdName)
			
			return fmt.Errorf(msg)
		}
	}
	
	return nil
}

// SuggestCorrectCommand analyzes the error and provides helpful suggestions
func SuggestCorrectCommand(cmd *cobra.Command, args []string, err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	cmdName := cmd.Name()
	
	// Handle unknown flag errors
	if strings.Contains(errStr, "unknown flag:") || strings.Contains(errStr, "unknown shorthand flag:") {
		flagName := extractFlagName(errStr)
		if flagName != "" {
			// First, try to find a flag alias
			if aliasErr := tryFlagAlias(cmd, flagName, args); aliasErr != nil {
				return aliasErr
			}
			// If no alias found, provide general suggestions
			return handleUnknownFlag(cmdName, flagName, errStr)
		}
	}
	
	// Handle other common errors
	if strings.Contains(errStr, "accepts") && strings.Contains(errStr, "arg(s)") {
		return handleArgCountError(cmdName, args, errStr)
	}
	
	return err
}

// extractFlagName extracts the flag name from error messages
func extractFlagName(errStr string) string {
	if strings.Contains(errStr, "unknown flag: --") {
		parts := strings.Split(errStr, "unknown flag: --")
		if len(parts) > 1 {
			return strings.TrimSpace(parts[1])
		}
	}
	if strings.Contains(errStr, "unknown shorthand flag: '") {
		parts := strings.Split(errStr, "unknown shorthand flag: '")
		if len(parts) > 1 {
			flagPart := strings.Split(parts[1], "'")
			if len(flagPart) > 0 {
				return strings.TrimSpace(flagPart[0])
			}
		}
	}
	return ""
}

// handleUnknownFlag provides suggestions for unknown flags
func handleUnknownFlag(cmdName, flagName, originalErr string) error {
	suggestions := []string{}
	
	// Check if this flag exists on other commands
	for cmd, flags := range CommandFlagMap {
		for _, validFlag := range flags {
			if validFlag == flagName {
				suggestions = append(suggestions, fmt.Sprintf("'%s' command supports --%s", cmd, flagName))
			}
		}
	}
	
	// Check for specific command suggestions
	if cmdSuggestions, exists := CommandSuggestions[cmdName]; exists {
		if suggestion, hasFlag := cmdSuggestions[flagName]; hasFlag {
			suggestions = append(suggestions, suggestion)
		}
	}
	
	// Show available flags for current command
	if availableFlags, exists := CommandFlagMap[cmdName]; exists {
		flagsList := strings.Join(availableFlags, ", ")
		suggestions = append(suggestions, fmt.Sprintf("Available flags for '%s': %s", cmdName, flagsList))
	}
	
	// Build helpful error message
	errorMsg := fmt.Sprintf("Error: %s\n\n", originalErr)
	
	if len(suggestions) > 0 {
		errorMsg += "💡 Suggestions:\n"
		for i, suggestion := range suggestions {
			errorMsg += fmt.Sprintf("   %d. %s\n", i+1, suggestion)
		}
	}
	
	errorMsg += fmt.Sprintf("\nUse 'mailos %s --help' for complete usage information.", cmdName)
	
	return fmt.Errorf("%s", errorMsg)
}

// handleArgCountError provides suggestions for argument count errors
func handleArgCountError(cmdName string, args []string, originalErr string) error {
	errorMsg := fmt.Sprintf("Error: %s\n\n", originalErr)
	
	switch cmdName {
	case "read":
		if len(args) == 0 {
			errorMsg += "💡 Suggestions:\n"
			errorMsg += "   1. Specify an email ID: 'mailos read <email_id>'\n"
			errorMsg += "   2. Use --id flag: 'mailos read --id <email_id>'\n"
			errorMsg += "   3. Find emails first: 'mailos search' then 'mailos read <id>'\n"
		}
	case "send":
		errorMsg += "💡 Suggestions:\n"
		errorMsg += "   1. Specify recipient: 'mailos send --to recipient@example.com --subject \"Subject\" --body \"Message\"'\n"
		errorMsg += "   2. Use interactive mode: 'mailos send' (will prompt for details)\n"
		errorMsg += "   3. Send from file: 'mailos send --file message.txt --to recipient@example.com'\n"
	}
	
	errorMsg += fmt.Sprintf("\nUse 'mailos %s --help' for complete usage information.", cmdName)
	
	return fmt.Errorf("%s", errorMsg)
}

// SetupErrorHandling adds the error handler to all commands
func SetupErrorHandling(rootCmd *cobra.Command) {
	// Set custom error handler for root command
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return SuggestCorrectCommand(cmd, nil, err)
	})
	
	// Apply to all subcommands
	for _, cmd := range rootCmd.Commands() {
		cmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
			return SuggestCorrectCommand(c, nil, err)
		})
		
		// Also handle RunE errors
		if cmd.RunE != nil {
			originalRunE := cmd.RunE
			cmd.RunE = func(c *cobra.Command, args []string) error {
				err := originalRunE(c, args)
				if err != nil {
					return SuggestCorrectCommand(c, args, err)
				}
				return nil
			}
		}
	}
}