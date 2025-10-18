# Email Database Sync

## New `sync-db` Command

The `sync-db` command has been added to sync emails from your local inbox to a SQLite database for fast querying and analysis.

### Quick Start

1. **Install SQLite** (if not already installed):
   ```bash
   # macOS (usually pre-installed)
   brew install sqlite
   
   # Linux
   sudo apt-get install sqlite3 libsqlite3-dev
   ```

2. **Sync emails from IMAP to local inbox**:
   ```bash
   mailos sync
   ```

3. **Sync from local inbox to database**:
   ```bash
   mailos sync-db
   ```

### Usage Examples

```bash
# Sync default account to database
mailos sync-db

# Sync specific account
mailos sync-db --account user@example.com

# Sync all configured accounts
mailos sync-db --all
```

### Database Location

Databases are created at: `~/.email/[account]/archive.db`

### Features

- **Fast querying** with SQLite indexes
- **SQL analysis** capabilities
- **Structured storage** with proper schema
- **Automatic deduplication** based on Message-ID
- **Incremental sync** support

### Architecture

```
IMAP Server → Local Inbox (JSON) → SQLite Database
     ↓              ↓                    ↓
  mailos sync   mailos sync-db      SQL queries
```

See [docs/sync-db.md](docs/sync-db.md) for detailed documentation.