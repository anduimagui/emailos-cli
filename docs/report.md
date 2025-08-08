# EmailOS Report Command Documentation

The `mailos report` command generates comprehensive email reports for specified time ranges with detailed analytics and summaries.

## Basic Usage

```bash
mailos report
```

Interactive time range selector followed by detailed report generation.

## Command-Line Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--range` | Time range for report | `--range "This week"` |
| `--output` | Save report to file | `--output report.md` |

## Time Range Options

### Interactive Selection
When run without `--range` flag, provides menu with:
- Last hour
- Today
- Yesterday  
- This week
- Last week
- This month
- Last month
- Last 7 days
- Last 30 days
- This year
- Custom range

### Preset Ranges
Use with `--range` flag:

| Range | Description |
|-------|-------------|
| `"Last hour"` | Previous 60 minutes |
| `"Today"` | Midnight to now |
| `"Yesterday"` | Previous calendar day |
| `"This week"` | Current week (Mon-Sun) |
| `"Last week"` | Previous full week |
| `"This month"` | Current calendar month |
| `"Last month"` | Previous calendar month |
| `"Last 7 days"` | Rolling 7-day window |
| `"Last 30 days"` | Rolling 30-day window |
| `"This year"` | Current calendar year |

## Report Contents

### Summary Section
- Time period covered
- Total emails analyzed
- Daily average
- Peak activity times

### Sender Analysis
- Top 10 most frequent senders
- Email count per sender
- Percentage of total volume
- New vs recurring senders

### Domain Statistics
- Most common email domains
- Corporate vs personal email ratio
- Domain distribution chart

### Time Analysis
- Hourly distribution
- Day of week patterns
- Peak communication times
- Response time estimates

### Subject Analysis
- Common keywords
- Topic clustering
- Thread identification
- Subject line patterns

### Attachment Summary
- Total attachments
- File type distribution
- Size statistics
- Sender attachment patterns

### Communication Patterns
- Conversation threads
- Reply frequencies
- Communication networks
- Key relationships

## Output Formats

### Terminal Display (Default)
Formatted markdown with:
- Headers and sections
- ASCII charts
- Tables
- Color highlighting (if supported)

### File Output
Save to markdown file:
```bash
mailos report --range "This month" --output monthly-report.md
```

Supports:
- Markdown (.md)
- Plain text (.txt)
- HTML (.html) - auto-converted
- PDF (.pdf) - requires pandoc

## Examples

### Weekly Report
```bash
mailos report --range "This week"
```

### Monthly Report to File
```bash
mailos report --range "Last month" --output reports/december-2024.md
```

### Today's Activity
```bash
mailos report --range "Today"
```

### Custom Date Range
```bash
mailos report
# Select "Custom range" from menu
# Enter start date: 2024-12-01
# Enter end date: 2024-12-15
```

### Quarterly Report
```bash
# Last 90 days
mailos report --range "Last 90 days" --output Q4-report.md
```

## Report Sections Detail

### 1. Executive Summary
```
Report Period: Dec 1-31, 2024
Total Emails: 847
Daily Average: 27.3
Busiest Day: Tuesday (156 emails)
Peak Hour: 9:00 AM - 10:00 AM
```

### 2. Sender Breakdown
```
Top Senders:
1. john@company.com (89 emails, 10.5%)
2. notifications@service.com (67 emails, 7.9%)
3. team@project.org (45 emails, 5.3%)
...
```

### 3. Communication Heatmap
```
Hour  Mon Tue Wed Thu Fri Sat Sun
00:00  2   1   0   3   1   0   0
01:00  0   0   1   0   0   0   0
...
09:00  15  22  18  20  16  2   1
```

### 4. Thread Analysis
```
Active Threads: 34
Longest Thread: "Project Alpha Discussion" (23 messages)
Most Participants: "Team Meeting Notes" (8 people)
```

### 5. Keyword Cloud
```
Common Terms:
- meeting (45 occurrences)
- project (38 occurrences)
- update (31 occurrences)
- review (28 occurrences)
```

## Advanced Features

### Filtering Reports
Combine with read filters:
```bash
# Report on emails from specific sender
mailos read --from manager@company.com --days 30 | mailos report

# Report on unread emails
mailos read --unread --days 7 | mailos report
```

### Scheduled Reports
Create automated reports with cron:
```bash
# Weekly report every Monday
0 9 * * 1 mailos report --range "Last week" --output ~/reports/weekly-$(date +%Y%m%d).md

# Monthly report on the 1st
0 8 1 * * mailos report --range "Last month" --output ~/reports/monthly-$(date +%Y%m).md
```

### Comparison Reports
Generate multiple reports for comparison:
```bash
mailos report --range "This week" --output this-week.md
mailos report --range "Last week" --output last-week.md
# Compare the two reports
```

## Report Customization

### Environment Variables
- `MAILOS_REPORT_TIMEZONE`: Set report timezone
- `MAILOS_REPORT_FORMAT`: Default output format
- `MAILOS_REPORT_DETAIL`: Detail level (summary/normal/detailed)

### Configuration Options
In `.email/report-config.json`:
```json
{
  "include_body_preview": false,
  "max_senders_shown": 20,
  "include_charts": true,
  "timezone": "America/New_York"
}
```

## Use Cases

### 1. Productivity Analysis
Track email patterns to identify:
- Peak productivity hours
- Communication bottlenecks
- Response time patterns

### 2. Relationship Management
Monitor:
- Key contact frequency
- Neglected relationships
- Communication balance

### 3. Workload Assessment
Measure:
- Email volume trends
- Seasonal patterns
- Department communication

### 4. Inbox Zero Progress
Track:
- Daily email processing
- Unread accumulation
- Response rates

## Performance Considerations

- Reports analyze up to 1000 emails by default
- Large date ranges may take longer
- Cache results for 15 minutes
- Use specific ranges for faster generation

## Integration

### With Other Tools
Export reports for use in:
- Spreadsheet analysis (CSV export planned)
- Dashboard tools (JSON export planned)
- Document management systems
- Time tracking applications

### API Usage (Future)
```bash
# Get report as JSON
mailos report --range "Today" --format json

# Pipe to analysis tools
mailos report --range "This week" | grep "Total:" | awk '{print $2}'
```

## Tips

1. **Regular Reviews**: Generate weekly reports for pattern recognition
2. **Archive Reports**: Save monthly reports for long-term analysis
3. **Compare Periods**: Look for trends across multiple reports
4. **Focus Areas**: Use filters to report on specific senders or subjects
5. **Time Management**: Identify and minimize peak email times

## Troubleshooting

### Empty Reports
- Verify emails exist in time range: `mailos read --range "..."`
- Check timezone settings
- Ensure proper authentication

### Slow Generation
- Reduce time range
- Clear email cache: `rm -rf ~/.email/cache`
- Check network connection

### Incorrect Counts
- Verify filter settings
- Check for duplicate emails
- Ensure IMAP sync is complete