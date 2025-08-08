# EmailOS Stats Command Documentation

The `mailos stats` command provides comprehensive email analytics and statistics with powerful filtering and aggregation capabilities.

## Basic Usage

```bash
mailos stats
```

This will analyze the last 100 emails and display comprehensive statistics.

## Command-Line Flags

### Basic Filters

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--number` | `-n` | Number of emails to analyze (default: 100) | `mailos stats -n 500` |
| `--unread` | `-u` | Analyze only unread emails | `mailos stats -u` |
| `--from` | | Filter by sender email address | `mailos stats --from john@example.com` |
| `--to` | | Filter by recipient email address | `mailos stats --to me@example.com` |
| `--subject` | | Filter by subject (partial match) | `mailos stats --subject invoice` |
| `--days` | | Analyze emails from last N days | `mailos stats --days 30` |
| `--range` | | Time range preset | `mailos stats --range "This week"` |

### Time Range Presets

The `--range` flag accepts the following presets:
- `"Last hour"` - Emails from the last hour
- `"Today"` - Today's emails
- `"Yesterday"` - Yesterday's emails
- `"This week"` - Current week's emails
- `"Last week"` - Previous week's emails
- `"This month"` - Current month's emails
- `"Last month"` - Previous month's emails
- `"This year"` - Current year's emails

## Advanced Query Parameters

You can pass additional parameters as key=value pairs after the command:

```bash
mailos stats [flags] [param1=value1] [param2=value2] ...
```

### Filter Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `from` | Filter by sender | `from=john@example.com` |
| `to` | Filter by recipient | `to=me@example.com` |
| `subject` | Filter by subject | `subject=invoice` |
| `limit`, `n` | Number of emails to analyze | `limit=500` |
| `days` | Emails from last N days | `days=30` |
| `range` | Time range preset | `range="Last week"` |
| `unread` | Only unread emails | `unread=true` |
| `sent` | Only sent emails | `sent=true` |
| `received` | Only received emails | `received=true` |

### Advanced Filters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `attachments`, `has-attachments` | Only emails with attachments | `attachments=true` |
| `min-size` | Minimum email size | `min-size=10kb` |
| `max-size` | Maximum email size | `max-size=5mb` |
| `domain`, `domains` | Filter by sender domains (comma-separated) | `domains=gmail.com,outlook.com` |
| `exclude-domain`, `exclude-domains` | Exclude specific domains | `exclude-domains=spam.com,junk.com` |
| `keyword`, `keywords` | Must contain keywords (comma-separated) | `keywords=urgent,important` |
| `exclude`, `exclude-words` | Must not contain words | `exclude-words=spam,unsubscribe` |

### Size Format

Size parameters support the following formats:
- Bytes: `1000` or `1000b`
- Kilobytes: `10kb` or `10k`
- Megabytes: `5mb` or `5m`
- Gigabytes: `1gb` or `1g`

### Aggregation & Display Options

| Parameter | Description | Example |
|-----------|-------------|---------|
| `group-by`, `groupby` | Group results by field | `group-by=sender` |
| `sort-by`, `sortby`, `sort` | Sort results | `sort=date` |
| `format` | Output format | `format=json` |
| `top`, `top-n` | Number of top items to show | `top=20` |

## Complex Query Examples

### 1. Analyze large emails from specific domains in the last week

```bash
mailos stats range="Last week" domains=gmail.com,yahoo.com min-size=1mb
```

### 2. Find emails with attachments from specific senders

```bash
mailos stats from=john@example.com attachments=true days=30
```

### 3. Exclude spam and marketing emails

```bash
mailos stats exclude-words=unsubscribe,promotional exclude-domains=marketing.com days=7
```

### 4. Analyze urgent emails with keywords

```bash
mailos stats keywords=urgent,asap,important --days 3
```

### 5. Complex multi-filter query

```bash
mailos stats \
  from=boss@company.com \
  days=30 \
  keywords=project,deadline \
  min-size=5kb \
  attachments=true \
  top=20
```

### 6. Analyze emails by size range

```bash
mailos stats min-size=100kb max-size=10mb --days 14
```

### 7. Filter multiple domains with exclusions

```bash
mailos stats domains=company.com,partner.com exclude-domains=noreply.company.com
```

## Statistics Output

The stats command provides comprehensive analytics with **visual charts** using Unicode characters:

### Summary Statistics
- Total number of emails matching criteria
- Number of emails with attachments
- Average email size

### Top Senders
- Most frequent email senders
- Percentage of total emails
- Limited to top 10 by default (configurable with `top` parameter)

### Top Domains
- Most common sender domains
- Percentage distribution
- Shows top 10 domains

### Activity by Hour (Visual Chart)
- 24-hour distribution of email activity
- **Visual bar chart using Unicode blocks** (█▇▆▅▄▃▂▁)
- Shows peak email times at a glance
- Example output:
  ```
  00:00 ▂▂ 2 emails
  08:00 ████████ 24 emails
  12:00 ██████ 18 emails
  17:00 █████████ 28 emails
  ```

### Activity by Weekday (Visual Chart)
- Weekly email distribution with visual bars
- Shows which days are busiest
- Percentage breakdown by day
- Example output:
  ```
  Monday    ████████ 25%
  Tuesday   ██████ 18%
  Wednesday ███████ 22%
  ```

### Top Subject Keywords
- Most common words in email subjects
- Excludes common words (the, and, for, etc.)
- Shows top 15 keywords with frequency

### Daily Distribution (Visual Timeline)
- Last 30 days of email activity
- **Visual bar chart showing daily trends**
- Helps identify patterns over time
- Displays as a timeline with Unicode bars

## Combining Flags and Parameters

You can combine command-line flags with query parameters:

```bash
mailos stats --unread --days 7 domains=important.com keywords=urgent
```

This will analyze unread emails from the last 7 days, from important.com domain, containing the word "urgent".

## Performance Tips

1. **Use specific filters**: The more specific your query, the faster the analysis
2. **Limit the number of emails**: Use `limit` or `-n` to analyze fewer emails for faster results
3. **Use time ranges**: Analyzing recent emails is faster than analyzing all emails
4. **Combine filters**: Multiple filters reduce the dataset size

## Output Formats

Currently supports:
- **Default**: Human-readable terminal output with visual charts
- **JSON**: Machine-readable format (use `format=json`)

## Use Cases

### 1. Email Volume Analysis
```bash
mailos stats --days 30
```
Understand your email patterns over the last month.

### 2. Sender Analysis
```bash
mailos stats from=client@company.com --days 90
```
Analyze communication patterns with specific contacts.

### 3. Attachment Tracking
```bash
mailos stats attachments=true --days 7
```
Find all emails with attachments from the last week.

### 4. Spam Detection
```bash
mailos stats exclude-words=unsubscribe,click,offer domains=suspicious.com
```
Identify potential spam patterns.

### 5. Important Email Monitoring
```bash
mailos stats keywords=urgent,important,asap --unread
```
Find critical unread emails.

### 6. Storage Analysis
```bash
mailos stats min-size=5mb
```
Find large emails consuming storage space.

## Notes

- All text searches are case-insensitive
- Domain filters match partial domains (e.g., `gmail.com` matches `mail.gmail.com`)
- Keywords search both subject and body content
- The stats command reads emails but does not modify them
- Results are based on locally cached email data after initial fetch

## Troubleshooting

If stats are not showing expected results:

1. Ensure your email account is properly configured: `mailos info`
2. Check that emails are being fetched: `mailos read -n 10`
3. Verify your filters are correct: Try broader filters first
4. Use the `--days` flag to ensure you're looking at the right time period