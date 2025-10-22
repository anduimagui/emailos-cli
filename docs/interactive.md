# EmailOS Interactive Mode Documentation

The interactive mode provides a rich terminal user interface for managing emails and interacting with AI assistants.

## Launching Interactive Mode

```bash
# Classic terminal UI (default)
mailos interactive

# AI Chat mode
mailos chat
```

## Interactive Mode Interfaces

### Classic UI
The default interface with menu-driven navigation and keyboard input.


## Keyboard Shortcuts

### Input Controls
- **Enter** - Submit query or select option
- **ESC ESC** - Clear current input (press ESC twice quickly within 500ms)
- **Backspace** - Delete character
- **Tab** - Auto-complete (where available)

### Navigation
- **↑↓** - Navigate menu options
- **←→** - Move cursor in input field
- **Page Up/Down** - Scroll through results

### System Controls
- **Ctrl+C** - Cancel current operation or go back
- **Ctrl+D** - Exit program (when input is empty)
- **Ctrl+L** - Clear screen

## Slash Commands

Type `/` to see available commands or execute directly:

### Email Management
- `/read` - Browse and read emails with filters
- `/send` - Compose and send new email
- `/delete` - Delete emails by criteria
- `/mark-read` - Mark selected emails as read
- `/unsubscribe` - Find unsubscribe links

### Analysis & Reports
- `/stats` - View email statistics with visual charts
- `/report` - Generate detailed email reports
- `/search` - Advanced email search with fuzzy matching and boolean operators

### Configuration
- `/template` - Manage email templates
- `/configure` - Modify settings
- `/provider` - Set up or change AI provider
- `/local` - Create project-specific configuration
- `/info` - Display current configuration

### System
- `/help` - Show help information
- `/exit` - Exit EmailOS

## AI Integration

### Natural Language Queries
Simply type your request without any prefix:
```
"Summarize emails from John from last week"
"Draft a reply to the latest email from Sarah"
"Find all unread emails with attachments"
```

### AI Provider Commands
- Configure provider: `/provider`
- Switch providers on the fly
- Supports Claude, GPT-4, Gemini, and more

## Interactive Features

### Smart Input Handling
- **Double ESC to Clear**: Quickly clear your input by pressing ESC twice
- **Auto-suggestion**: Common commands and emails addresses
- **History**: Navigate previous commands with arrow keys
- **Context Awareness**: Commands adapt based on current state

### Visual Feedback
- Color-coded messages (errors in red, success in green)
- Progress indicators for long operations
- Real-time email count updates
- Visual charts in statistics

### Email Preview
When browsing emails:
- See sender, subject, and preview
- Attachment indicators
- Read/unread status
- Quick actions (reply, delete, mark)

## Template Shortcuts

Use `@` prefix for quick template access:
```
@meeting John tomorrow at 3pm
@followup regarding our discussion
@thank for your help with the project
```

Available templates:
- `@meeting` - Schedule a meeting
- `@followup` - Follow-up email
- `@thank` - Thank you email
- `@intro` - Introduction email
- `@request` - Request information
- `@reminder` - Send reminder
- `@apologize` - Apology email
- `@decline` - Polite decline

## Interactive Workflows

### Email Triage Workflow
1. Launch interactive mode: `mailos interactive`
2. Type `/read --unread` to see unread emails
3. Use arrow keys to navigate
4. Press Enter to read full email
5. Use quick actions: `r` for reply, `d` for delete, `m` for mark read

### Batch Processing
1. `/read --from newsletter@example.com`
2. Select multiple emails with Space
3. Choose bulk action: delete, mark read, or archive

### AI-Assisted Composition
1. Type: "compose an email to John about the quarterly report"
2. AI drafts the email
3. Review and edit if needed
4. Confirm and send

### Advanced Search Workflow
1. Type: `/search -q "urgent AND project OR deadline"`
2. Use boolean operators for complex queries
3. Apply filters: `/search --date-range "last week" --has-attachments`
4. Fuzzy search: `/search --from "supprt"` (finds "support")
5. Size filters: `/search --min-size 1MB --attachment-size 500KB`

## Configuration

### Setting Default UI
```bash
# Configure in settings
mailos configure
```

### Customizing Shortcuts
Edit `~/.email/.slash_config.json`:
```json
{
  "shortcuts": {
    "r": "/read --unread",
    "s": "/send",
    "q": "/exit"
  }
}
```

## Tips and Tricks

### Quick Email Check
```
mailos interactive
/read --unread --limit 5
```

### Daily Email Review
```
mailos interactive
/stats --range "Today"
/read --range "Today" --unread
```

### Advanced Search Examples
```
# Boolean search with fuzzy matching
/search -q "urgent OR important AND NOT spam"

# Field-specific search
/search -q "from:manager AND subject:project"

# Date range with size filters
/search --date-range "last week" --min-size 1MB

# Complex query with attachments
/search -q "contract OR agreement" --has-attachments --attachment-size 500KB
```

### Project-Specific Email
```
cd my-project
mailos local  # Configure project email
mailos interactive  # Uses project settings
```

### Keyboard-Only Navigation
- Never need to use mouse
- All functions accessible via keyboard
- Faster than clicking through menus

## Troubleshooting

### ESC Key Not Working
- Ensure terminal supports raw input mode
- Try different terminal emulator
- Fallback: use Ctrl+U to clear line


### Slow Performance
- Limit email fetch: `/read --limit 20`
- Use time filters: `--days 7`
- Clear cache: `rm -rf ~/.email/cache`

### Command Not Recognized
- Check spelling
- Type `/` alone to see all commands
- Update EmailOS to latest version

## Advanced Features

### Custom Key Bindings
Future versions will support custom key bindings via configuration file.

### Scripting Support
Interactive mode can be scripted:
```bash
echo -e "/read --unread\n/exit" | mailos interactive
```

### Integration with Tools
- Export emails: `/export --format json`
- Pipe to other commands
- Use with automation tools

## Best Practices

1. **Learn Keyboard Shortcuts**: Much faster than menus
2. **Use Slash Commands**: Direct access to functions
3. **Leverage AI**: Natural language is often quickest
4. **Create Templates**: Save time on common emails
5. **Configure Locally**: Project-specific settings
6. **Regular Updates**: Keep EmailOS updated for new features

## Getting Help

- In-app help: `/help` or press `?`
- Documentation: This file and others in docs/
- GitHub Issues: Report bugs and request features
- Community: Share tips and workflows