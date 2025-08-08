# Changelog

All notable changes to MailOS will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Draft email management**: New `mailos drafts` command for creating and managing email drafts
- **Batch email sending**: `mailos send --drafts` to send all prepared draft emails
- **Draft filtering**: Send drafts selectively with `--filter` option (e.g., `priority:high`)
- **Dry-run mode**: Preview drafts before sending with `--dry-run` flag
- **Draft templates**: Create drafts in markdown with YAML frontmatter
- **Escape key clearing**: Press ESC twice quickly to clear input in interactive mode
- Advanced query filters: `--sent`, `--attachments`, `--min-size`, `--max-size`, `--domains`, `--keywords`
- Profile image support in email templates with `{{PROFILE_IMAGE}}` placeholder
- JSON output format for read command with `--json` flag
- `mailos open` command to open emails in default mail client
- `mailos provider` command for AI provider configuration
- `mailos chat` command for dedicated AI chat interface
- Slash commands in interactive mode (e.g., `/read`, `/send`, `/stats`)
- Enhanced keyboard shortcuts in interactive mode (Ctrl+D to exit, ESC ESC to clear)
- Visual statistics charts with Unicode bar graphs
- Group-by and sort-by options for query results
- File input support for email body with `--file` flag
- Plain text email sending with `--plain` flag

### Changed
- Interactive mode now supports raw terminal input for better keyboard handling
- Statistics command displays visual activity charts by hour and weekday
- Templates now support profile images and enhanced placeholders

### Fixed
- Terminal input handling for special key sequences
- Escape key detection in various terminal environments

## [0.1.8] - 2024-01-08

### Added
- React Ink UI as default interactive interface
- Automatic UI installation on first launch
- Fallback to classic UI if React UI unavailable
- API server for UI-backend communication
- Email tagging and filtering in UI
- Production-ready UI deployment

### Changed
- Default interactive mode now uses React Ink
- Classic UI available with MAILOS_CLASSIC_UI=true
- Improved UI path resolution
- Better error handling for UI installation

### Fixed
- UI installation path issues
- Build process for different environments

## [0.1.7] - 2024-01-XX

### Added
- Natural language query support with `q=` parameter
- Email statistics command (`mailos stats`) with time range filtering
- Advanced email filtering with multiple criteria
- Batch operations: mark-read, delete, export
- Enhanced template management system
- Query system documentation

### Changed
- Improved command parsing for better user experience
- Enhanced error messages for invalid commands

### Fixed
- Query parameter handling for complex searches
- Template loading issues in certain environments

## [0.1.6] - 2024-01-XX

### Added
- Multi-platform support (macOS, Linux, Windows)
- npm package distribution for easy installation
- AI integration with multiple providers (Claude, GPT-4, Gemini)
- Interactive TUI mode for better user experience
- Markdown email support with automatic HTML conversion
- License validation system for enterprise features

### Changed
- Improved email provider configuration
- Enhanced setup wizard for first-time users

### Fixed
- IMAP connection stability issues
- Email threading in conversation view

## [0.1.5] - 2024-01-XX

### Added
- Initial public release
- Basic email send/read functionality
- Support for major email providers (Gmail, Fastmail, Outlook, Yahoo, Zoho)
- Interactive mode for email management
- Template system for email composition

### Security
- App-specific password support
- Encrypted credential storage
- Secure IMAP/SMTP connections

## Command Reference

### Basic Commands
```bash
mailos send <email> <subject> <body>  # Send an email
mailos read [options]                 # Read emails
mailos interactive                    # Interactive mode
mailos setup                         # Configuration wizard
```

### Query Commands
```bash
mailos q="<query>"                   # Natural language query
mailos "<query>"                     # Alternative query syntax
mailos stats [options]               # Email statistics
```

### Advanced Options
```bash
# Read filters
--from <email>      # Filter by sender
--to <email>        # Filter by recipient
--subject <text>    # Filter by subject
--unread           # Only unread emails
--days <n>         # Last n days
--limit <n>        # Limit results
--range <range>    # Time range (e.g., "last week")

# Stats options
--days <n>         # Statistics for last n days
--range <range>    # Statistics for time range
--from <email>     # Statistics for specific sender

# Batch operations
mark-read [filters]   # Mark emails as read
delete [filters]      # Delete emails
export [options]      # Export emails
```

## Migration Guide

### From 0.1.5 to 0.1.6
- Run `mailos setup` to configure AI providers
- Update npm package: `npm update -g mailos`

### From 0.1.6 to 0.1.7
- New query syntax available, update scripts using old format
- Statistics feature requires no additional configuration