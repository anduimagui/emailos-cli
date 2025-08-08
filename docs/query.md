# EmailOS Advanced Query Documentation

EmailOS provides powerful query capabilities for searching, filtering, and analyzing emails using natural language and advanced parameters.

## Query Syntax

### Natural Language Queries
```bash
mailos q="unread emails from john about the project"
mailos "find all emails with attachments from last week"
```

### Direct Query
```bash
# Quoted query (recommended for complex queries)
mailos "emails from CEO sent yesterday"

# Using q= parameter
mailos q="urgent emails with attachments"
```

## Query Parameters

### Basic Filters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `from` | Filter by sender email | `from=john@example.com` |
| `to` | Filter by recipient | `to=me@example.com` |
| `subject` | Filter by subject text | `subject=invoice` |
| `limit`, `n` | Number of results | `limit=50` |
| `days` | Emails from last N days | `days=7` |
| `range` | Time range preset | `range="Last week"` |

### Boolean Filters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `unread` | Only unread emails | `unread=true` |
| `sent` | Only sent emails | `sent=true` |
| `received` | Only received emails | `received=true` |
| `attachments` | Has attachments | `attachments=true` |

### Size Filters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `min-size` | Minimum email size | `min-size=100KB` |
| `max-size` | Maximum email size | `max-size=10MB` |

Size formats supported:
- Bytes: `1000`, `1000b`
- Kilobytes: `10KB`, `10k`
- Megabytes: `5MB`, `5m`
- Gigabytes: `1GB`, `1g`

### Domain Filters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `domains` | Include specific domains | `domains=gmail.com,yahoo.com` |
| `exclude-domains` | Exclude domains | `exclude-domains=spam.com` |

### Content Filters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `keywords` | Must contain keywords | `keywords=urgent,important` |
| `exclude-words` | Must not contain words | `exclude-words=unsubscribe,spam` |

### Aggregation Options

| Parameter | Description | Example |
|-----------|-------------|---------|
| `group-by` | Group results by field | `group-by=sender` |
| `sort-by` | Sort results | `sort-by=date` |
| `top-n` | Limit to top N results | `top-n=10` |
| `format` | Output format | `format=json` |

## Time Range Presets

Use with `range=` parameter:
- `"Last hour"` - Past 60 minutes
- `"Today"` - Since midnight
- `"Yesterday"` - Previous day
- `"This week"` - Current week
- `"Last week"` - Previous week
- `"This month"` - Current month
- `"Last month"` - Previous month
- `"This year"` - Current year

## Complex Query Examples

### 1. Find large attachments from specific sender
```bash
mailos q="from=john@company.com attachments=true min-size=5MB days=30"
```

### 2. Urgent unread emails excluding spam
```bash
mailos "keywords=urgent,asap unread=true exclude-domains=marketing.com,spam.com"
```

### 3. Weekly report from manager
```bash
mailos q="from=manager@company.com subject=report range='This week'"
```

### 4. Cleanup old large emails
```bash
mailos "days=90 min-size=10MB sort-by=size"
```

### 5. Find meeting invites
```bash
mailos q="keywords=meeting,calendar,invite days=7 group-by=sender"
```

### 6. Customer emails with attachments
```bash
mailos "domains=customer.com,client.com attachments=true days=14"
```

### 7. Exclude newsletters and marketing
```bash
mailos q="exclude-words=unsubscribe,newsletter,promotional days=1"
```

### 8. Project-specific emails
```bash
mailos "subject=ProjectX keywords=deadline,milestone from=team.com"
```

## Natural Language Processing

EmailOS understands various natural language patterns:

### Temporal Queries
- "emails from yesterday"
- "messages received last week"
- "mail from the past 3 days"
- "today's unread emails"

### Sender Queries
- "emails from John"
- "messages from @gmail.com"
- "mail from the marketing team"

### Content Queries
- "emails about the project"
- "messages with attachments"
- "urgent emails"
- "unread newsletters"

### Combined Queries
- "unread emails from John about the meeting"
- "large attachments from last month"
- "urgent messages from the CEO today"

## Query Operators

### Logical Operators (in natural language)
- **AND**: "emails from John AND with attachments"
- **OR**: "emails from John OR Jane"
- **NOT**: "emails NOT from marketing"

### Comparison Operators
- Size: `min-size`, `max-size`
- Date: `days`, `range`
- Count: `limit`, `top-n`

## Output Formats

### Default Format
Human-readable text with formatting:
```
📧 Email #12345 (2024-01-15 10:30 AM)
From: john@example.com
Subject: Project Update
Preview: Here's the latest status...
```

### JSON Format
Machine-readable JSON:
```bash
mailos q="from=john@example.com format=json"
```

Output:
```json
[
  {
    "id": 12345,
    "from": "john@example.com",
    "subject": "Project Update",
    "date": "2024-01-15T10:30:00Z",
    "body": "...",
    "attachments": []
  }
]
```

### CSV Format (future)
```bash
mailos q="format=csv" > emails.csv
```

## Grouping and Aggregation

### Group by Sender
```bash
mailos q="days=30 group-by=sender top-n=10"
```
Shows top 10 senders in the last 30 days.

### Group by Domain
```bash
mailos q="group-by=domain days=7"
```
Shows email distribution by domain.

### Group by Date
```bash
mailos q="group-by=date range='This month'"
```
Shows daily email counts for the current month.

## Sorting Options

### Sort by Date (default)
```bash
mailos q="sort-by=date"  # or just sort=date
```

### Sort by Size
```bash
mailos q="attachments=true sort-by=size"
```

### Sort by Sender
```bash
mailos q="sort-by=sender days=7"
```

## Performance Optimization

### Use Specific Filters
More specific queries run faster:
```bash
# Slow
mailos q="emails"

# Fast
mailos q="from=john@example.com days=7 unread=true"
```

### Limit Results
Always use limit for large datasets:
```bash
mailos q="attachments=true limit=100"
```

### Use Time Ranges
Recent emails are cached and faster:
```bash
mailos q="range='Today' unread=true"
```

## Query Chaining

Combine EmailOS commands for powerful workflows:

```bash
# Find and delete old newsletters
mailos q="keywords=newsletter days=30" | \
  grep "Email #" | \
  cut -d'#' -f2 | \
  xargs mailos delete --ids

# Export important emails
mailos q="from=ceo@company.com format=json" > important.json

# Statistics on query results
mailos q="domains=client.com days=30" | mailos stats
```

## Saved Queries

Save frequently used queries in your shell configuration:

```bash
# ~/.bashrc or ~/.zshrc
alias mailos-urgent="mailos q='keywords=urgent,asap unread=true'"
alias mailos-today="mailos q='range=Today sort-by=sender'"
alias mailos-large="mailos q='min-size=5MB days=30'"
```

## Query Tips

1. **Start Broad, Then Narrow**: Begin with simple queries and add filters
2. **Use Quotes**: Always quote complex queries to prevent shell interpretation
3. **Combine Filters**: Multiple filters are AND-ed together
4. **Check Spelling**: Misspelled domains or email addresses return no results
5. **Use Wildcards**: Partial matches work for subject and keywords
6. **Time Ranges**: Use presets for common time periods
7. **Format for Scripts**: Use JSON format when piping to other tools

## Troubleshooting

### No Results
- Check email configuration: `mailos info`
- Verify spelling of email addresses
- Try broader time range
- Remove filters one by one

### Slow Queries
- Add more specific filters
- Use smaller time ranges
- Limit number of results
- Check internet connection for IMAP

### Unexpected Results
- Keywords are case-insensitive
- Domains match subdomains
- Check boolean values (true/false)
- Verify time range

## Advanced Use Cases

### Email Analytics Dashboard
```bash
#!/bin/bash
echo "=== Email Analytics ==="
echo "Today's emails:"
mailos q="range=Today" | wc -l
echo "Unread count:"
mailos q="unread=true" | wc -l
echo "With attachments:"
mailos q="attachments=true days=7" | wc -l
```

### Auto-responder Script
```bash
# Find emails needing response
mailos q="unread=true keywords=request,please days=1 format=json" | \
  jq -r '.[] | .from' | \
  while read email; do
    mailos send --to "$email" --subject "Auto-Reply" \
      --body "Thank you for your email. We'll respond within 24 hours."
  done
```

### Email Archiver
```bash
# Archive old emails to files
mailos q="days=365 format=json" > archive-$(date +%Y).json
mailos delete days=365  # Optional: delete after archiving
```

## Future Enhancements

Planned query features:
- Regex support for advanced pattern matching
- Full-text search indexing
- Saved query management
- Query history and suggestions
- Machine learning-based smart queries
- Integration with external search engines