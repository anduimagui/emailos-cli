# Query Parsing Test Checklist

## Current Behavior
- `mailos [ANYTHING]` = interpreted as query

## New Expected Behavior
- `mailos` (no args) = show landing page/help
- `mailos q=[QUERY]` = interpret everything after `q=` as query
- `mailos "[QUERY]"` = interpret quoted content as query
- `mailos '[QUERY]'` = interpret quoted content as query
- `mailos [KNOWN_COMMAND]` = execute known command (read, compose, etc.)
- `mailos [UNKNOWN]` = show error/help suggesting proper query format

## Test Cases

### 1. Landing Page Tests
- [ ] `mailos` → Should show landing page with available commands
- [ ] `mailos help` → Should show help/landing page

### 2. Query Parameter Tests
- [ ] `mailos q=unread emails` → Query: "unread emails"
- [ ] `mailos q=emails from john` → Query: "emails from john"
- [ ] `mailos q=meetings today` → Query: "meetings today"
- [ ] `mailos q=` → Should handle empty query gracefully

### 3. Quoted Query Tests
- [ ] `mailos "unread emails"` → Query: "unread emails"
- [ ] `mailos 'emails from john'` → Query: "emails from john"
- [ ] `mailos "emails with 'quotes' inside"` → Query: "emails with 'quotes' inside"
- [ ] `mailos 'emails with "quotes" inside'` → Query: "emails with "quotes" inside"

### 4. Known Command Tests
- [ ] `mailos read` → Execute read command
- [ ] `mailos compose` → Execute compose command
- [ ] `mailos search` → Execute search command
- [ ] `mailos config` → Execute config command

### 5. Edge Cases
- [ ] `mailos "q=test"` → Query: "q=test" (quoted takes precedence)
- [ ] `mailos q="quoted query"` → Query: "quoted query"
- [ ] `mailos random text` → Show error suggesting proper format
- [ ] `mailos read q=test` → Execute read command (known command takes precedence)

### 6. Multi-word Query Tests
- [ ] `mailos q=find all attachments from last week` → Query: "find all attachments from last week"
- [ ] `mailos "find all attachments from last week"` → Query: "find all attachments from last week"
- [ ] `mailos 'find all attachments from last week'` → Query: "find all attachments from last week"

## Implementation Notes
- Parse arguments before any other processing
- Check for known commands first
- Then check for quoted strings
- Then check for q= parameter
- Default to landing page if no valid input
- Provide clear error messages for invalid formats