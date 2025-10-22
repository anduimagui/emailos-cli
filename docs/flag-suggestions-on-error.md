# Flag Suggestions on Error System

## Overview

The EmailOS CLI now includes an intelligent error handling system that provides contextual suggestions when users use incorrect flags or arguments. Instead of generic "unknown flag" errors, the system guides users toward the correct commands and workflows.

## Implementation

### Core Components

The system consists of three main files:

1. **`cmd/mailos/error_handler.go`** - Core error handling logic
2. **`cmd/mailos/main.go`** - Integration with cobra commands  
3. **This documentation** - Usage patterns and extension guidelines

### Error Handler Architecture

```go
// CommandFlagMap defines available flags for each command
var CommandFlagMap = map[string][]string{
    "read": {"include-documents", "id"},
    "search": {"number", "n", "unread", "u", "from", "to", "subject", ...},
    // ... other commands
}

// CommandSuggestions provides workflow guidance
var CommandSuggestions = map[string]map[string]string{
    "read": {
        "number": "Use 'mailos search --number N' to show N emails, then 'mailos read <id>' for specific email",
        // ... other suggestions
    },
}
```

## How It Works

### 1. Flag Detection

When a user enters an invalid flag, the system:
- Extracts the flag name from the error message
- Searches `CommandFlagMap` to find which commands support that flag
- Looks up workflow suggestions in `CommandSuggestions`

### 2. Suggestion Generation

The error handler generates multiple types of suggestions:

**Command Availability**: Shows which commands support the attempted flag
```
'search' command supports --subject
'send' command supports --subject
```

**Workflow Guidance**: Explains the correct multi-step process
```
Use 'mailos search --subject <text>' to find emails, then 'mailos read <id>' for specific email
```

**Available Options**: Lists valid flags for the current command
```
Available flags for 'read': include-documents, id
```

### 3. Enhanced Output

**Before:**
```
Error: unknown flag: --subject
```

**After:**
```
Error: unknown flag: --subject

💡 Suggestions:
   1. 'search' command supports --subject
   2. 'send' command supports --subject  
   3. Use 'mailos search --subject <text>' to find emails, then 'mailos read <id>' for specific email
   4. Available flags for 'read': include-documents, id

Use 'mailos read --help' for complete usage information.
```

## Current Command Coverage

### Fully Mapped Commands

| Command | Key Flags | Common Mistakes |
|---------|-----------|-----------------|
| `read` | `--id`, `--include-documents` | Users try `--number`, `--from`, `--subject` |
| `search` | `--number`, `--from`, `--to`, `--subject`, `--unread` | Comprehensive filtering |
| `send` | `--to`, `--subject`, `--body`, `--attach` | Complex composition |
| `sent` | `--number`, `--to`, `--subject`, `--days` | Historical querying |
| `stats` | `--number`, `--unread`, `--from`, `--days` | Analytics filtering |
| `download` | `--id`, `--from`, `--subject`, `--output-dir` | Attachment handling |

### Commands Needing Enhancement

The following commands could benefit from expanded error handling:

| Command | Current State | Enhancement Opportunities |
|---------|---------------|---------------------------|
| `delete` | Basic mapping | Add workflow suggestions for bulk operations |
| `reply` | Basic mapping | Guide users through reply vs reply-all |
| `forward` | Basic mapping | Explain attachment handling |
| `accounts` | Not mapped | Add account management guidance |
| `configure` | Not mapped | Provider-specific setup help |
| `draft` | Not mapped | Draft workflow explanations |

## Application to Other Commands

### 1. Adding New Command Support

To add error handling for a new command:

```go
// 1. Add to CommandFlagMap
"newcmd": {
    "flag1", "f", "flag2", "other-flag",
},

// 2. Add workflow suggestions  
"newcmd": {
    "common-mistake": "Use 'correct command' instead of 'newcmd --common-mistake'",
},
```

### 2. Common Error Patterns

**Search vs Action Commands**
```go
// Users often confuse filtering (search) with action (read/delete/etc)
"read": {
    "number": "Use 'mailos search --number N' to show N emails, then 'mailos read <id>' for specific email",
}
```

**Bulk vs Individual Operations**
```go
"delete": {
    "all": "Use 'mailos search' to find emails, then 'mailos delete --ids 1,2,3' for specific emails",
}
```

**Provider-Specific Features**
```go
"configure": {
    "gmail": "Use 'mailos configure --provider gmail' to set up Gmail integration",
}
```

### 3. Workflow-Based Suggestions

The system excels at explaining multi-step workflows:

```go
"download": {
    "subject": "Use 'mailos search --subject <text>' to find emails, then 'mailos download --id <id>' for attachments",
}
```

## Benefits for AI Integration

### LLM-Friendly Error Messages

The structured error output helps LLMs understand and correct user commands:

1. **Clear Context**: Errors explain what went wrong and why
2. **Actionable Steps**: Numbered suggestions provide clear next actions  
3. **Command Discovery**: Shows related commands and their capabilities
4. **Workflow Guidance**: Explains the intended multi-step processes

### Example AI Interaction

```
User: "mailos read --from john@example.com"
CLI: Error with workflow suggestion
AI: "I see you want to read emails from john@example.com. Let me search for those emails first, then show you how to read a specific one..."
```

## Extension Patterns

### 1. Provider-Specific Errors

```go
// Handle provider-specific configuration errors
"configure": {
    "gmail-token": "Gmail requires OAuth2. Use 'mailos configure --provider gmail' for guided setup",
    "fastmail-api": "FastMail requires JMAP token. Get one from Settings > App Passwords",
}
```

### 2. Context-Aware Suggestions

```go
// Different suggestions based on user's configuration state
func getContextualSuggestion(cmd, flag, userState string) string {
    if userState == "unconfigured" {
        return "Run 'mailos setup' first to configure your email account"
    }
    // ... other context checks
}
```

### 3. Interactive Fallbacks

```go
// Offer to launch interactive mode for complex operations
"send": {
    "complex-composition": "This looks complex. Try 'mailos send --interactive' for guided composition",
}
```

## Integration with Other CLI Tools

### Universal Patterns

These patterns can be applied to any cobra-based CLI:

1. **Flag Mapping**: Maintain comprehensive flag inventories
2. **Cross-Command Awareness**: Show where flags work across commands
3. **Workflow Guidance**: Explain multi-step processes
4. **Context Sensitivity**: Tailor suggestions to user state

### Implementation Template

```go
// For any new CLI tool
type ErrorHandler struct {
    CommandFlags    map[string][]string
    WorkflowSuggestions map[string]map[string]string
    ContextProviders []ContextProvider
}

func (eh *ErrorHandler) HandleFlagError(cmd, flag string) error {
    // 1. Find commands supporting this flag
    // 2. Look up workflow suggestions  
    // 3. Generate contextual help
    // 4. Format user-friendly output
}
```

## Future Enhancements

### 1. Machine Learning Integration

- Learn from user error patterns
- Suggest personalized workflows
- Adapt suggestions based on usage history

### 2. Interactive Error Recovery

- Offer to run corrected commands automatically
- Provide command completion suggestions
- Launch relevant interactive modes

### 3. Documentation Integration

- Link to specific documentation sections
- Show relevant examples from help system
- Integrate with online tutorials

## Real-World Application Examples

### 1. EmailOS CLI Implementation 

The complete implementation now covers 25+ commands with comprehensive error handling:

```bash
# Before enhancement
$ mailos accounts --create test@example.com
Error: unknown flag: --create

# After enhancement  
$ mailos accounts --create test@example.com
Error: unknown flag: --create

💡 Suggestions:
   1. Use 'mailos accounts --add <email>' to add a new account
   2. Available flags for 'accounts': set, add, provider, use-existing-credentials, set-signature, clear, list, sync-fastmail, token, test-connection

Use 'mailos accounts --help' for complete usage information.
```

### 2. Pattern Applied to Other CLI Tools

This approach can be applied to any cobra-based CLI tool:

**Git-like tools:**
```go
"checkout": {
    "create": "Use 'git checkout -b <branch>' to create and switch to new branch",
    "new":    "Use 'git checkout -b <branch>' or 'git switch -c <branch>' for new branches",
}
```

**Docker-like tools:**
```go
"run": {
    "background": "Use 'docker run -d' to run container in background",
    "interactive": "Use 'docker run -it' for interactive container",
}
```

**Kubernetes-like tools:**
```go
"get": {
    "all": "Use 'kubectl get all' to see all resources, or 'kubectl get <resource>' for specific types",
    "pods": "Use 'kubectl get pods' to list pods",
}
```

### 3. Multi-Repository Application

For organizations with multiple CLI tools, create a shared error handling library:

```go
// shared-cli-errors/errorhandler.go
package cliErrors

type ErrorHandler struct {
    CommandMap map[string][]string
    Suggestions map[string]map[string]string
    ContextProviders []func(string, string) string
}

// Apply to different tools
func (eh *ErrorHandler) ApplyToTool(toolName string, rootCmd *cobra.Command) {
    // Universal error handling setup
}
```

## Repository Analysis: Other Applications

### Current EmailOS Repository Structure

The EmailOS repository contains several areas where this pattern could be extended:

**Web Interface (`/web`):**
- Could implement similar contextual help for form validation
- Frontend error messages could use same suggestion patterns
- API error responses could provide workflow guidance

**Marketing Tools (`/marketing`):**
- CLI tools for outreach management could benefit from similar error handling
- Template generation scripts could provide better guidance

**Deployment Scripts (`/deployment`):**
- Infrastructure setup commands could use enhanced error messaging
- Configuration validation could provide specific fix suggestions

### Integration Opportunities

**1. MCP Tools Integration**
```json
// mcp-tools.json could define error handling patterns
{
  "error_handling": {
    "patterns": ["flag-suggestions", "workflow-guidance"],
    "suggestion_sources": ["CommandFlagMap", "online-docs"]
  }
}
```

**2. AI Provider Integration**
```go
// ai_provider.go could leverage error handling
func (ai *AIProvider) InterpretError(cmdErr error) string {
    // Use error handler suggestions to improve AI responses
    return enhancedSuggestion
}
```

**3. Interactive Mode Enhancement** 
```go
// interactive_enhanced.go could use error patterns
func (i *Interactive) HandleInvalidInput(input string) {
    // Apply same suggestion logic to interactive mode
}
```

## Scaling Patterns

### 1. Documentation Integration

Connect error handling with existing documentation:

```go
"configure": {
    "gmail": "See docs/configure.md#gmail-setup for Gmail configuration guide",
    "oauth": "Visit web/docs/LICENSE_INTEGRATION.md for OAuth setup details",
}
```

### 2. Community Contributions

Enable community-driven error message improvements:

```yaml
# .github/ISSUE_TEMPLATE/error-improvement.yml
name: Error Message Improvement
description: Suggest better error handling for a command
body:
  - type: input
    attributes:
      label: Command
      description: Which command needs better error handling?
  - type: textarea
    attributes:
      label: Current Error
      description: What error message did you see?
  - type: textarea
    attributes:
      label: Suggested Improvement
      description: How could the error message be more helpful?
```

### 3. Analytics Integration

Track which errors occur most frequently:

```go
func (eh *ErrorHandler) LogError(command, flag string) {
    // Send to analytics to identify common pain points
    analytics.Track("cli_error", map[string]string{
        "command": command,
        "flag": flag,
        "suggestion_shown": "yes",
    })
}
```

## Conclusion

The flag suggestions system transforms CLI error handling from a frustration point into a learning and guidance opportunity. By understanding user intent and providing clear pathways to success, it makes the EmailOS CLI more accessible to both human users and AI systems.

**Key Benefits:**
- **User Experience**: Reduces frustration and learning curve
- **AI Integration**: Provides structured error information for LLMs
- **Maintainability**: Centralizes error handling logic
- **Scalability**: Easily extensible to new commands and workflows

**Implementation Success Metrics:**
- Reduced support requests about command usage
- Improved command discovery and adoption
- Better AI assistant integration
- Faster user onboarding

The pattern is highly reusable and can significantly improve user experience across any command-line application that uses similar flag-based interfaces. Organizations can adopt this approach incrementally, starting with their most problematic commands and expanding coverage over time.