# EmailOS Search Command Documentation

The `mailos search` command provides advanced email search capabilities with fuzzy matching, boolean operators, flexible date ranges, and size filters.

## Basic Usage

```bash
mailos search
```

Searches and displays the 10 most recent emails by default.

## Quick Examples

```bash
# Boolean search with OR operator
mailos search -q "urgent OR deadline OR important"

# Field-specific search
mailos search -q "from:support AND subject:invoice"

# Date range with size filter
mailos search --date-range "last week" --min-size 1MB

# Complex query with NOT operator
mailos search -q "project AND NOT spam" --has-attachments

# Fuzzy search for typos
mailos search --from "supprt" # matches "support"
```

## Command-Line Flags

### Basic Search Options

| Flag | Short | Description | Default | Example |
|------|-------|-------------|---------|---------|
| `--number` | `-n` | Number of emails to display | 10 | `mailos search -n 20` |
| `--unread` | `-u` | Show only unread emails | false | `mailos search -u` |
| `--from` | | Filter by sender address | | `mailos search --from john@example.com` |
| `--to` | | Filter by recipient | | `mailos search --to me@example.com` |
| `--subject` | | Filter by subject | | `mailos search --subject "meeting"` |
| `--days` | | Show emails from last N days | | `mailos search --days 7` |

### Advanced Search Options

| Flag | Short | Description | Default | Example |
|------|-------|-------------|---------|---------|
| `--query` | `-q` | Complex search query with boolean operators | | `mailos search -q "urgent AND project"` |
| `--fuzzy-threshold` | | Fuzzy matching threshold (0.0-1.0) | 0.7 | `mailos search --fuzzy-threshold 0.5` |
| `--no-fuzzy` | | Disable fuzzy matching | false | `mailos search --no-fuzzy` |
| `--case-sensitive` | | Enable case sensitive search | false | `mailos search --case-sensitive` |

### Date Range Options

| Flag | Description | Example |
|------|-------------|---------|
| `--date-range` | Flexible date range expressions | `mailos search --date-range "today"` |
| `--range` | Time range presets | `mailos search --range "This week"` |

**Date Range Expressions:**
- Natural language: `"today"`, `"yesterday"`, `"last week"`, `"this month"`
- Specific ranges: `"2023-01-01 to 2023-12-31"`
- Relative dates: `"last 30 days"`, `"last 7 days"`

### Size Filters

| Flag | Description | Example |
|------|-------------|---------|
| `--min-size` | Minimum email size | `mailos search --min-size 1MB` |
| `--max-size` | Maximum email size | `mailos search --max-size 10MB` |
| `--has-attachments` | Filter emails with attachments | `mailos search --has-attachments` |
| `--attachment-size` | Minimum attachment size | `mailos search --attachment-size 500KB` |

**Size Units Supported:** B, KB, MB, GB, TB

### Output Options

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `--save-markdown` | Save emails as markdown files | false | `mailos search --save-markdown` |
| `--output-dir` | Directory for markdown files | emails | `mailos search --output-dir ./results` |
| `--json` | Output as JSON format | false | `mailos search --json` |

## Advanced Query Syntax

### Boolean Operators

- **AND**: All terms must match (default)
  ```bash
  mailos search -q "urgent AND project AND deadline"
  ```

- **OR**: Any term can match
  ```bash
  mailos search -q "invoice OR billing OR payment"
  ```

- **NOT**: Exclude terms
  ```bash
  mailos search -q "meeting NOT cancelled"
  ```

### Field-Specific Search

Target specific email fields:

```bash
# Search in sender field
mailos search -q "from:support"

# Search in recipient field
mailos search -q "to:team@company.com"

# Search in subject line
mailos search -q "subject:urgent"

# Search in email body
mailos search -q "body:confidential"
```

### Exact Phrase Matching

Use quotes for exact phrases:

```bash
# Exact phrase (disables fuzzy matching)
mailos search -q '"urgent: please reply immediately"'

# Mixed exact and fuzzy
mailos search -q '"project alpha" AND deadline'
```

### Complex Query Examples

```bash
# Multiple conditions with field targeting
mailos search -q "from:manager AND (urgent OR important) AND NOT spam"

# Date range with boolean search
mailos search -q "invoice OR billing" --date-range "last month"

# Size and content filters combined
mailos search -q "contract OR agreement" --min-size 100KB --has-attachments
```

## Fuzzy Search

Fuzzy search helps find emails even with typos or slight variations:

### How It Works
- Uses Levenshtein distance algorithm
- Configurable similarity threshold (0.0 to 1.0)
- Default threshold: 0.7 (70% similarity)

### Examples

```bash
# Find "support" even if typed as "supprt"
mailos search --from "supprt"

# Adjust fuzzy sensitivity
mailos search --fuzzy-threshold 0.5 --from "suport"

# Disable fuzzy matching for exact searches
mailos search --no-fuzzy --from "support@company.com"
```

### Fuzzy Threshold Guide
- **0.9-1.0**: Very strict (minor typos only)
- **0.7-0.8**: Balanced (default, handles common typos)
- **0.5-0.6**: Loose (more variations accepted)
- **0.0-0.4**: Very loose (may match unrelated terms)

## Date Range Examples

### Natural Language Dates
```bash
mailos search --date-range "today"
mailos search --date-range "yesterday"
mailos search --date-range "last week"
mailos search --date-range "this month"
mailos search --date-range "last month"
mailos search --date-range "this year"
```

### Specific Date Ranges
```bash
mailos search --date-range "2023-01-01 to 2023-12-31"
mailos search --date-range "2024-06-01 to 2024-06-30"
```

### Relative Dates
```bash
mailos search --date-range "last 7 days"
mailos search --date-range "last 30 days"
mailos search --date-range "last 90 days"
```

## Size Filter Examples

### Email Size Filtering
```bash
# Large emails only
mailos search --min-size 1MB

# Small to medium emails
mailos search --max-size 500KB

# Specific size range
mailos search --min-size 100KB --max-size 5MB
```

### Attachment Filtering
```bash
# Emails with any attachments
mailos search --has-attachments

# Emails with large attachments
mailos search --attachment-size 1MB

# Combined attachment and size filters
mailos search --has-attachments --min-size 2MB
```

## Practical Use Cases

### Finding Important Communications
```bash
# Urgent emails from management
mailos search -q "from:manager AND (urgent OR important OR asap)"

# Contract-related emails with attachments
mailos search -q "contract OR agreement" --has-attachments --min-size 50KB
```

### Email Cleanup and Management
```bash
# Large emails consuming storage
mailos search --min-size 10MB

# Old promotional emails
mailos search -q "unsubscribe OR newsletter" --date-range "last 6 months"

# Emails with large attachments for archival
mailos search --attachment-size 5MB --date-range "last year"
```

### Project and Work Tracking
```bash
# Project-related communications
mailos search -q "project:alpha OR alpha-project" --date-range "this month"

# Invoice and billing emails
mailos search -q "invoice OR billing OR payment" --has-attachments
```

## Performance Tips

1. **Use specific filters** to reduce search scope
2. **Combine date ranges** with other filters for faster results
3. **Adjust fuzzy threshold** based on your needs
4. **Use field-specific searches** when possible
5. **Limit result count** with `-n` flag

## Troubleshooting

### Search Returns No Results
1. Check query syntax with simple terms first
2. Try lowering fuzzy threshold: `--fuzzy-threshold 0.5`
3. Verify date ranges are correct
4. Remove filters one by one to isolate issues

### Fuzzy Search Too Broad
1. Increase fuzzy threshold: `--fuzzy-threshold 0.8`
2. Use exact phrases with quotes
3. Enable case sensitivity: `--case-sensitive`
4. Disable fuzzy matching: `--no-fuzzy`

### Size Filters Not Working
1. Verify size units (B, KB, MB, GB, TB)
2. Check that emails have been synced locally
3. Use broader size ranges initially

### Boolean Queries Not Working
1. Ensure proper spacing around operators
2. Use quotes for phrases containing operators
3. Check parentheses for complex queries
4. Start with simple AND/OR queries

## Integration with Other Commands

### Combining with Read Command
```bash
# Search to find email IDs, then read specific emails
mailos search -q "important contract" --json | jq '.[].id'
mailos read 12345  # Read specific email by ID
```

### Saving Search Results
```bash
# Save search results to files
mailos search -q "quarterly report" --save-markdown --output-dir ./reports

# Export search results as JSON
mailos search --date-range "last month" --json > last_month_emails.json
```

## Notes

- Search uses IMAP server-side filtering when possible for performance
- Fuzzy matching works on all text fields (from, to, subject, body)
- Size calculations include email content and attachments
- Date ranges use your local timezone
- Boolean operators are case-insensitive (AND, and, And all work)
- Search results are sorted by date (newest first)