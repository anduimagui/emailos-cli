# EmailOS Main Entry Point

## Overview

The main entry point for the EmailOS CLI is located in `cmd/mailos/main.go`. This file handles:
- Command parsing and routing
- Auto-update checks
- All subcommands via Cobra CLI framework
- Natural language query parsing

## Key Files

### Main Command Entry
- **File**: `cmd/mailos/main.go`
- **Main function**: `main()` at line 2238
- **Purpose**: Entry point that handles auto-updates and executes the root command

### Interactive Terminal
- **Launch command**: `mailos interactive` or just `mailos` without arguments
- **Function calls**: 
  - `cmd/mailos/main.go:425` - Called when no query is provided
  - `cmd/mailos/main.go:1935` - Called by the `interactive` subcommand
- **Implementation**: `InteractiveModeWithMenu()` located in `interactive_enhanced.go`

## Interactive Mode Components

### Core Interactive File: `interactive_enhanced.go`
This file provides the main interactive mode orchestration with multiple UI options:

#### Main Functions:
- **`InteractiveModeWithMenu()`** (line 13) - Main entry point for interactive mode
- **`InteractiveModeWithMenuOptions()`** (line 18) - Configurable entry with display options
- **`showEnhancedInteractiveMenuWithOptions()`** (line 98) - Displays input-first menu with status
- **`showCommandMenu()`** (line 130) - Arrow-navigated command selection menu using promptui
- **`executeCommand()`** (line 202) - Executes slash commands directly
- **`showInteractiveHelp()`** (line 253) - Displays comprehensive help

#### UI Mode Selection (line 33-52):
The interactive mode can use different UI implementations based on environment variables:
1. **OpenTUI** (experimental) - `InteractiveModeWithOpenTUI()` from `opentui_input.go` when `MAILOS_USE_OPENTUI=true`
2. **BubbleTea UI** (deprecated) - functionality moved to OLD directory
3. **Classic UI** - Falls back to suggestion-based modes

To switch between UI implementations, set the appropriate environment variable:
```bash
# Use OpenTUI (experimental, high-performance)
export MAILOS_USE_OPENTUI=true

# Use BubbleTea (default, feature-rich)
export MAILOS_USE_BUBBLETEA=true  # or just don't set anything

# Disable BubbleTea to use classic UI
export MAILOS_USE_BUBBLETEA=false
```

#### Suggestion Modes (line 108-126):
Based on `MAILOS_SUGGESTION_MODE` environment variable:
- **dynamic** - `InteractiveModeWithDynamicSuggestions()` from `dynamic_suggestions.go`
- **simple** - `EnhancedInteractiveMode()` 
- **live** - `EnhancedInteractiveModeV2()`
- **clean** - `CleanInteractiveMode()`

#### Key Features:
1. **Arrow Navigation**: Uses promptui for up/down arrow navigation in menus
2. **Slash Commands**: Direct command execution via `/command` syntax
3. **File Autocomplete**: `ReadLineWithFileAutocomplete()` from `file_autocomplete.go` for `@` file tagging
4. **Status Display**: Shows current email configuration and AI provider status
5. **Help System**: Comprehensive keyboard shortcuts and command documentation

#### Keyboard Shortcuts (from help text, line 287-297):
- **Enter** - Submit query or select option
- **ESC ESC** - Clear current input (press ESC twice quickly)
- **@** - Show file/folder autocomplete
- **/** - Show command menu
- **↑↓** - Navigate menu options
- **Tab** - Auto-complete selected file (in @ mode)
- **ESC** - Cancel autocomplete (in @ mode)
- **Ctrl+C** - Cancel/Go back
- **Ctrl+D** - Exit (when input is empty)
- **Backspace** - Delete character

## Command Structure

The CLI uses Cobra for command handling with the following main commands:
- `setup` - Initial email configuration
- `send` - Send emails
- `read` - Read emails  
- `search` - Advanced email search with fuzzy matching, boolean operators, and filters
- `drafts` - Manage draft emails
- `interactive` - Launch interactive terminal
- `chat` - Launch AI chat interface
- `configure` - Manage email configuration
- `template` - Customize HTML email template
- And many more...

## Query Parsing

The `parseQueryFromArgs()` function (line 298) handles natural language queries:
- Checks for `q=` parameter format
- Handles quoted strings
- Distinguishes between commands and queries
- Supports natural language patterns like "find unread emails"

## Auto-Update System

The `checkForUpdates()` function (line 29) automatically:
- Checks GitHub for new releases
- Downloads and installs updates
- Restarts the application with new version
- Can be disabled with `MAILOS_SKIP_UPDATE=true`

## Interactive Mode Entry Points

1. **Default launch** (no arguments): Calls `InteractiveModeWithMenu()` at line 425
2. **Explicit command**: `mailos interactive` calls it at line 1935  
3. **Chat command**: `mailos chat` calls it at line 1950
4. **Query with provider selection**: `HandleQueryWithProviderSelection()` for natural language queries