# Email Database Sync Documentation

## Overview

The `sync-db` command syncs emails from your local inbox storage to a SQLite database for fast querying and analysis.

## Prerequisites

### SQLite Installation

The sync-db feature requires SQLite to be available on your system. Here's how to install it:

#### macOS
```bash
# SQLite comes pre-installed on macOS
# To install the latest version via Homebrew:
brew install sqlite
```

#### Linux (Ubuntu/Debian)
```bash
sudo apt-get update
sudo apt-get install sqlite3 libsqlite3-dev
```

#### Linux (CentOS/RHEL/Fedora)
```bash
# CentOS/RHEL
sudo yum install sqlite sqlite-devel

# Fedora
sudo dnf install sqlite sqlite-devel
```

#### Windows
Download SQLite from: https://www.sqlite.org/download.html

Or install via Chocolatey:
```powershell
choco install sqlite
```

### Go CGO Requirements

This feature uses the `github.com/mattn/go-sqlite3` driver which requires CGO. Ensure you have:

1. **GCC/Clang compiler installed**
   - macOS: Install Xcode Command Line Tools (`xcode-select --install`)
   - Linux: Install build-essential (`sudo apt-get install build-essential`)
   - Windows: Install TDM-GCC or MinGW-w64

2. **CGO enabled** (default)
   ```bash
   export CGO_ENABLED=1
   ```

## Usage

### Sync default account
```bash
mailos sync-db
```

### Sync specific account
```bash
mailos sync-db --account user@example.com
```

### Sync all configured accounts
```bash
mailos sync-db --all
```

## Database Structure

The SQLite database is created at: `~/.email/[account]/archive.db`

### Tables

#### `emails` table
- `id` - Primary key
- `message_id` - Unique message identifier
- `from_address` - Sender email address
- `to_addresses` - JSON array of recipient addresses
- `subject` - Email subject
- `date_sent` - When the email was sent
- `body_text` - Plain text body
- `body_html` - HTML body
- `attachments` - JSON array of attachment filenames
- `attachment_data` - BLOB containing attachment data
- `in_reply_to` - Message ID this email replies to
- `created_at` - When record was created
- `updated_at` - When record was last updated

#### `sync_metadata` table
- `account_email` - Account email address
- `last_sync_time` - Last synchronization timestamp
- `total_emails` - Total number of emails synced
- `last_email_date` - Date of the most recent email
- `sync_version` - Sync version for compatibility

### Indexes
- Message ID (for fast lookups)
- From address (for sender filtering)
- Date sent (for chronological queries)
- Subject (for subject searching)

## Workflow

1. **First, sync emails from IMAP to local inbox:**
   ```bash
   mailos sync
   ```

2. **Then, sync from local inbox to database:**
   ```bash
   mailos sync-db
   ```

The sync-db command reads from the local inbox files created by the `sync` command and imports them into SQLite for fast querying.

## Benefits

- **Fast querying** - SQLite provides indexed searches
- **SQL analysis** - Use standard SQL for email analysis
- **Lightweight** - Single file database per account
- **Portable** - Database files can be backed up/shared
- **Structured data** - JSON fields for complex data types

## Troubleshooting

### CGO compilation errors
```bash
# Ensure CGO is enabled
export CGO_ENABLED=1

# Install build tools on macOS
xcode-select --install

# Install build tools on Linux
sudo apt-get install build-essential
```

### SQLite driver errors
```bash
# Reinstall the SQLite driver
go clean -modcache
go get github.com/mattn/go-sqlite3
```

### Database permissions
Ensure the `~/.email/[account]` directory is writable:
```bash
chmod 755 ~/.email
chmod 755 ~/.email/[account]
```

## Advanced Usage

### Direct SQL access
You can query the database directly:
```bash
sqlite3 ~/.email/user@example.com/archive.db

# Example queries:
SELECT COUNT(*) FROM emails;
SELECT from_address, COUNT(*) FROM emails GROUP BY from_address ORDER BY COUNT(*) DESC LIMIT 10;
SELECT subject, date_sent FROM emails WHERE date_sent > datetime('now', '-7 days');
```

### Backup databases
```bash
# Backup all account databases
cp -r ~/.email /path/to/backup/

# Backup specific account
cp ~/.email/user@example.com/archive.db /path/to/backup/
```