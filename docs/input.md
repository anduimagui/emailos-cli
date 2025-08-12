# EmailOS Input System Documentation

## Overview

EmailOS features an advanced interactive input system with AI-powered suggestions, file autocomplete, and command shortcuts. The system is built using the **promptui** library for robust terminal UI interactions with arrow key navigation.

## Table of Contents

1. [Input Modes](#input-modes)
2. [AI Suggestions](#ai-suggestions)
3. [Keyboard Shortcuts](#keyboard-shortcuts)
4. [File Autocomplete](#file-autocomplete)
5. [Slash Commands](#slash-commands)
6. [Configuration](#configuration)
7. [Implementation Details](#implementation-details)

## Input Modes

EmailOS supports multiple input modes, selectable via the `MAILOS_SUGGESTION_MODE` environment variable:

### Dynamic Mode (Default)
- **File**: `dynamic_suggestions.go`
- **Features**:
  - Live filtering of suggestions as you type
  - Searchable suggestion list
  - Press Enter on empty input for full suggestions
  - Start typing to filter suggestions in real-time
  - Arrow keys navigate filtered results

### Clean Mode
- **File**: `refined_input_suggestions.go`
- **Features**:
  - Minimalist interface
  - Focused input prompt
  - Press Enter on empty input for suggestion menu
  - Clean separation between input and suggestions

### Simple Mode
- **File**: `simple_input_suggestions.go`
- **Features**:
  - Basic implementation
  - Enter on empty input shows suggestions
  - Straightforward selection menu

### Live Mode
- **File**: `input_with_live_suggestions.go`
- **Features**:
  - Hybrid approach
  - Shows hints about available suggestions
  - Interactive suggestion display

## AI Suggestions

### Default AI Commands

The system provides 8 pre-configured AI command suggestions:

1. **📊 Summarize yesterday's emails**
   - Command: `"Summarize all emails from yesterday"`
   - Quick overview of previous day's messages

2. **📬 Show unread emails**
   - Command: `"Show me all unread emails with a brief summary"`
   - Lists and summarizes unread messages

3. **✍️ Draft a professional reply**
   - Command: `"Help me draft a professional reply to the last email"`
   - Assists in composing well-formatted responses

4. **⭐ Find important emails**
   - Command: `"Find important emails from this week"`
   - Identifies high-priority messages

5. **📈 Email statistics**
   - Command: `"Show me email statistics for this week"`
   - Provides insights about email activity

6. **🔔 Schedule follow-ups**
   - Command: `"Find emails that need follow-ups"`
   - Identifies messages requiring responses

7. **🧹 Clean up inbox**
   - Command: `"Help me clean up my inbox - find emails to delete"`
   - Helps identify removable emails

8. **📅 Today's agenda from emails**
   - Command: `"Extract today's agenda and tasks from my emails"`
   - Extracts action items and meetings

### Accessing Suggestions

#### In Dynamic Mode (Default):
1. Press Enter on empty input to see all suggestions
2. Start typing to filter suggestions
3. Use ↑↓ arrow keys to navigate
4. Press Enter to select
5. Select "💬 Type your own query..." for custom input

#### In Other Modes:
1. Press Enter on empty input
2. Navigate with arrow keys
3. Select suggestion or custom query option

## Keyboard Shortcuts

### Global Shortcuts

| Key | Action | Description |
|-----|--------|-------------|
| `Enter` | Submit/Select | Submit query or select highlighted option |
| `↑↓` | Navigate | Move through suggestions or menu items |
| `Esc` | Cancel | Cancel current operation or autocomplete |
| `Ctrl+C` | Interrupt | Exit current operation |
| `Ctrl+D` | Exit | Exit EmailOS (when input is empty) |

### Input-Specific Shortcuts

| Key | Action | Description |
|-----|--------|-------------|
| `@` | File Autocomplete | Activate file/folder autocomplete |
| `/` | Command Menu | Show command menu or execute command |
| `Tab` | Complete | Auto-complete selected file (in @ mode) |
| `Esc Esc` | Clear Line | Double ESC to clear current input |
| `Backspace` | Delete | Delete character before cursor |
| `←→` | Move Cursor | Navigate within input line |

## File Autocomplete

### Activation
Type `@` to activate file autocomplete mode. The system will:
1. Display files and folders from current directory
2. Show relative paths
3. Support fuzzy matching as you type

### Navigation
- **↑↓**: Navigate through file suggestions
- **Tab**: Complete selected file/folder
- **Enter**: Insert selected path
- **Esc**: Cancel autocomplete mode

### Features
- Directories shown with `/` suffix
- Fuzzy search across filenames
- Maximum 10 suggestions displayed
- Relative path preservation

## Slash Commands

### Available Commands

| Command | Description | Icon |
|---------|-------------|------|
| `/read` | Browse and read emails | 📧 |
| `/send` | Compose and send email | ✉️ |
| `/report` | Generate email analytics | 📊 |
| `/unsubscribe` | Find unsubscribe links | 🔗 |
| `/delete` | Delete emails by criteria | 🗑️ |
| `/mark-read` | Mark emails as read | ✓ |
| `/template` | Manage email templates | 📝 |
| `/configure` | Settings & configuration | ⚙️ |
| `/provider` | Set AI provider | 🤖 |
| `/info` | Display configuration | ℹ️ |
| `/help` | Show help information | ❓ |
| `/exit` | Exit EmailOS | 👋 |

### Usage
- Type `/` alone to see command menu with arrow navigation
- Type `/command` directly to execute
- Commands support additional arguments

## Configuration

### Environment Variables

#### MAILOS_SUGGESTION_MODE
Controls the input suggestion mode:
```bash
export MAILOS_SUGGESTION_MODE=dynamic  # Default, live filtering
export MAILOS_SUGGESTION_MODE=clean    # Minimalist interface
export MAILOS_SUGGESTION_MODE=simple   # Basic implementation
export MAILOS_SUGGESTION_MODE=live     # Hybrid approach
```

#### MAILOS_USE_INK
Enable React Ink UI (experimental):
```bash
export MAILOS_USE_INK=true
```

### Settings File
The input system respects configurations in:
- `~/.emailos/config.json` - Main configuration
- `~/.emailos/slash_config.json` - Slash command settings

## Implementation Details

### Architecture

```
interactive_enhanced.go
    ├── Dynamic Mode (default)
    │   └── dynamic_suggestions.go
    │       ├── DynamicInputWithSuggestions()
    │       └── InteractiveModeWithDynamicSuggestions()
    │
    ├── Clean Mode
    │   └── refined_input_suggestions.go
    │       ├── RefinedInputWithSuggestions()
    │       └── CleanInteractiveMode()
    │
    ├── Simple Mode
    │   └── simple_input_suggestions.go
    │       ├── SimpleInputWithSuggestions()
    │       └── EnhancedInteractiveMode()
    │
    └── Live Mode
        └── input_with_live_suggestions.go
            ├── LiveInputWithSuggestions()
            └── EnhancedInteractiveModeV2()
```

### Key Components

#### 1. AI Suggestions System
- **File**: `ai_suggestions.go`
- **Structure**: `AISuggestion` struct with Title, Command, Description, Icon
- **Function**: `GetDefaultAISuggestions()` returns 8 pre-configured suggestions

#### 2. File Autocomplete
- **File**: `file_autocomplete.go`
- **Uses**: promptui for selection interface
- **Features**: Fuzzy matching, directory detection, relative paths

#### 3. Promptui Integration
- **Library**: `github.com/manifoldco/promptui`
- **Templates**: Custom templates for consistent UI
- **Features**: Arrow navigation, search, custom styling

### Input Flow

```mermaid
graph TD
    A[User Input] --> B{Empty?}
    B -->|Yes| C[Show AI Suggestions]
    B -->|No| D{Special Character?}
    D -->|@| E[File Autocomplete]
    D -->|/| F[Command Menu/Execute]
    D -->|Text| G[Process Query]
    C --> H[Arrow Navigation]
    H --> I[Select Suggestion]
    I --> G
    E --> J[Select File]
    J --> G
    F --> K[Execute Command]
    G --> L[Send to AI Provider]
```

### Suggestion Selection Process

1. **Initial Display**: Show input prompt with hint
2. **Empty Enter**: Display suggestion menu
3. **Navigation**: Arrow keys move selection
4. **Filtering** (Dynamic Mode): Type to filter in real-time
5. **Selection**: Enter confirms choice
6. **Custom Query**: Option to type own query

### Terminal UI Rendering

The system uses ANSI escape sequences for:
- Cursor positioning: `\033[s` (save), `\033[u` (restore)
- Line clearing: `\033[K` (clear to end of line)
- Movement: `\033[A` (up), `\033[B` (down)
- Colors: Via promptui templates (cyan, green, dim)

## Usage Examples

### Basic AI Query
```
▸ What emails did I receive today?
[Enter]
🤔 Processing: What emails did I receive today?
```

### Using Suggestions
```
▸ [Press Enter on empty]

AI Suggestions
▸ 📊 Summarize yesterday's emails
  📬 Show unread emails
  ✍️ Draft a professional reply
  [Navigate with ↑↓, Enter to select]
```

### File Reference
```
▸ Summarize @reports/quarterly.pdf
[Tab to complete, Enter to submit]
```

### Direct Command
```
▸ /read
📧 Executing: /read
[Email reading interface opens]
```

## Troubleshooting

### Suggestions Not Appearing
1. Check terminal supports ANSI escape sequences
2. Verify promptui is installed: `go get github.com/manifoldco/promptui`
3. Try different mode: `MAILOS_SUGGESTION_MODE=simple mailos`

### Arrow Keys Not Working
1. Ensure terminal is in raw mode
2. Check terminal emulator compatibility
3. Try different terminal (iTerm2, Terminal.app, etc.)

### File Autocomplete Issues
1. Verify current directory permissions
2. Check for hidden files starting with `.`
3. Ensure @ symbol is at word boundary

## Advanced Features

### Custom Suggestion Sets
Modify `GetDefaultAISuggestions()` in `ai_suggestions.go` to add custom suggestions:

```go
func GetDefaultAISuggestions() []AISuggestion {
    return []AISuggestion{
        {
            Title:       "Your Custom Command",
            Command:     "Actual command to execute",
            Description: "What this does",
            Icon:        "🎯",
        },
        // ... more suggestions
    }
}
```

### Mode-Specific Customization
Each mode can be customized independently:
- Adjust `Size` parameter for visible items
- Modify templates for different styling
- Add mode-specific features

### Integration with AI Providers
The selected suggestion's `Command` field is passed directly to `InvokeAIProvider()`, which routes to the configured AI backend (OpenAI, Claude, etc.).

## Best Practices

1. **Use Dynamic Mode** for power users who want quick filtering
2. **Use Clean Mode** for minimal distraction
3. **Configure AI provider** before using suggestions
4. **Learn keyboard shortcuts** for faster navigation
5. **Use @ for files** instead of typing full paths
6. **Use / commands** for common operations

## Future Enhancements

Potential improvements under consideration:
- Context-aware suggestions based on email history
- Learning from user selections
- Custom suggestion categories
- Multi-line input support
- Rich text formatting in suggestions
- Suggestion history and favorites