# EmailOS Main Command Execution Report

## Overview
This report documents how the main command is executed in the EmailOS CLI application, tracing the execution flow from `task dev` through the interactive UI.

## 1. Build and Execution Process

### Entry Point: Taskfile
**Location:** [`Taskfile.yml:36-40`](/Users/andrewmaguire/LOCAL/Github/_code-main/emialos/cli-go/Taskfile.yml#L36)

The `task dev` command triggers two steps:
1. **Build:** Compiles the Go binary from `cmd/mailos/main.go`
2. **Execute:** Runs the compiled `mailos` binary

```yaml
dev:
  desc: Build and run locally
  cmds:
    - task: build
    - ./mailos {{.CLI_ARGS}}
```

### Build Task
**Location:** [`Taskfile.yml:9-12`](/Users/andrewmaguire/LOCAL/Github/_code-main/emialos/cli-go/Taskfile.yml#L9)

```yaml
build:
  desc: Build the mailos binary locally
  cmds:
    - go build -o mailos cmd/mailos/main.go
```

## 2. Main Function Entry

### Main Function
**Location:** [`cmd/mailos/main.go:2041`](/Users/andrewmaguire/LOCAL/Github/_code-main/emialos/cli-go/cmd/mailos/main.go#L2041)

The main function performs these key operations:

```go
func main() {
    // 1. Auto-update check
    checkForUpdates()
    
    // 2. Initialize mailos package
    mailos.Initialize()
    
    // 3. Register all commands
    registerCommands()
    
    // 4. Execute the root command
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
```

## 3. Command Routing

### Root Command Handler
**Location:** [`cmd/mailos/main.go:389-427`](/Users/andrewmaguire/LOCAL/Github/_code-main/emialos/cli-go/cmd/mailos/main.go#L389)

The root command (`mailos` without subcommands) handles:

1. **Query Parsing:** Checks if arguments contain a natural language query
   - [`parseQueryFromArgs()` at line 408](/Users/andrewmaguire/LOCAL/Github/_code-main/emialos/cli-go/cmd/mailos/main.go#L305)
   - Supports formats: `mailos "query"`, `mailos q=query`, `mailos find emails`

2. **Query Handling:** If query found, processes it with AI
   - [`HandleQueryWithProviderSelection()` at line 413](/Users/andrewmaguire/LOCAL/Github/_code-main/emialos/cli-go/cmd/mailos/main.go#L413)

3. **Interactive Mode:** If no query, launches interactive UI
   - [`InteractiveModeWithMenu()` at line 425](/Users/andrewmaguire/LOCAL/Github/_code-main/emialos/cli-go/cmd/mailos/main.go#L425)

## 4. Interactive UI Implementation

### UI Entry Point
**Location:** [`interactive_enhanced.go:13-95`](/Users/andrewmaguire/LOCAL/Github/_code-main/emialos/cli-go/interactive_enhanced.go#L13)

```go
func InteractiveModeWithMenu() error {
    return InteractiveModeWithMenuOptions(false, true)
}
```

### UI Implementation Selection
**Location:** [`interactive_enhanced.go:35-53`](/Users/andrewmaguire/LOCAL/Github/_code-main/emialos/cli-go/interactive_enhanced.go#L35)

The system supports multiple UI implementations, selected via environment variables:

1. **OpenTUI** (`MAILOS_USE_OPENTUI=true`) - Experimental
2. **BubbleTea** (Default) - Current default UI
3. **React Ink** (`MAILOS_USE_INK=true`) - Legacy
4. **Classic** - Fallback option

### Main Interactive Loop
**Location:** [`interactive_enhanced.go:67-94`](/Users/andrewmaguire/LOCAL/Github/_code-main/emialos/cli-go/interactive_enhanced.go#L67)

The interactive loop:
1. Shows logo/header on first iteration
2. Displays status box with current configuration
3. Handles user input with dynamic suggestions
4. Processes commands and queries

## 5. Command Processing Flow

### Available Commands
**Location:** [`interactive_enhanced.go:144-160`](/Users/andrewmaguire/LOCAL/Github/_code-main/emialos/cli-go/interactive_enhanced.go#L144)

The system supports these commands:
- `/read` - Read emails
- `/send` - Send email
- `/inbox` - Open inbox in browser
- `/sent` - Open sent mail
- `/drafts` - Open drafts
- `/report` - Generate email report
- `/unsubscribe` - Find unsubscribe links
- `/delete` - Delete emails
- `/mark-read` - Mark emails as read
- `/template` - Manage templates
- `/configure` - Settings
- `/provider` - Set AI provider
- `/info` - Show configuration
- `/help` - Show help
- `/exit` - Exit application

### Command Handlers
Each command maps to a specific handler function:
- [`handleInteractiveRead()`](/Users/andrewmaguire/LOCAL/Github/_code-main/emialos/cli-go/interactive_menu.go#L134)
- [`handleInteractiveSend()`](/Users/andrewmaguire/LOCAL/Github/_code-main/emialos/cli-go/interactive_menu.go#L226)
- [`handleInteractiveReport()`](/Users/andrewmaguire/LOCAL/Github/_code-main/emialos/cli-go/interactive_menu.go#L329)
- [`handleInteractiveDelete()`](/Users/andrewmaguire/LOCAL/Github/_code-main/emialos/cli-go/interactive_menu.go#L435)

## 6. Execution Flow Diagram

```mermaid
graph TD
    A[task dev] --> B[Build: go build -o mailos]
    B --> C[Execute: ./mailos]
    C --> D[main.go:main()]
    D --> E[checkForUpdates()]
    D --> F[mailos.Initialize()]
    D --> G[rootCmd.Execute()]
    G --> H{Arguments?}
    H -->|Query| I[HandleQueryWithProviderSelection()]
    H -->|No Query| J[InteractiveModeWithMenu()]
    J --> K{UI Selection}
    K -->|BubbleTea| L[Functionality moved to OLD directory]
    K -->|Classic| M[showEnhancedInteractiveMenu()]
    M --> N[User Input Loop]
    N --> O{Command Type}
    O -->|/command| P[Command Handler]
    O -->|Text| Q[AI Query Handler]
    O -->|/exit| R[Exit]
```

## 7. Key Features

### Auto-Update System
**Location:** [`cmd/mailos/main.go:29-117`](/Users/andrewmaguire/LOCAL/Github/_code-main/emialos/cli-go/cmd/mailos/main.go#L29)

- Checks GitHub for new releases
- Downloads and installs updates automatically
- Can be disabled with `MAILOS_SKIP_UPDATE=true`

### Authentication Flow
**Location:** [`interactive_enhanced.go:19-31`](/Users/andrewmaguire/LOCAL/Github/_code-main/emialos/cli-go/interactive_enhanced.go#L19)

- Validates authentication before interactive mode
- Guides users through setup if not configured
- Handles multiple email accounts

### Dynamic Input Suggestions
**Location:** [`interactive_enhanced.go:114-133`](/Users/andrewmaguire/LOCAL/Github/_code-main/emialos/cli-go/interactive_enhanced.go#L114)

Multiple suggestion modes available:
- `dynamic` - Filter suggestions as you type
- `simple` - Press Enter for suggestions
- `live` - Live input with suggestions
- `clean` - Minimal interface

## 8. Configuration

### Environment Variables
- `MAILOS_USE_OPENTUI` - Use OpenTUI interface
- `MAILOS_USE_BUBBLETEA` - Use BubbleTea interface (default)
- `MAILOS_USE_INK` - Use React Ink interface
- `MAILOS_SUGGESTION_MODE` - Set suggestion mode (dynamic/simple/live/clean)
- `MAILOS_SKIP_UPDATE` - Skip automatic updates

### Configuration Files
- Global: `~/.email/config.json`
- Local: `.email/config.json`
- Slash config: `.email/.slash_config.json`

## Summary

The EmailOS main command execution follows a well-structured flow:

1. **Build & Launch:** Task runner builds and executes the binary
2. **Initialization:** Main function sets up the environment
3. **Command Routing:** Cobra framework routes to appropriate handlers
4. **Interactive UI:** Multiple UI implementations provide user interaction
5. **Command Processing:** Specific handlers execute user commands
6. **AI Integration:** Natural language queries processed by configured AI provider

The system is designed for flexibility with multiple UI options, automatic updates, and comprehensive command support for email management operations.