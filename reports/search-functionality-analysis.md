# EmailOS Search Functionality Analysis

## Overview

EmailOS implements email search functionality primarily through IMAP search criteria and filtering mechanisms. The search functionality is distributed across multiple components including CLI commands, interactive modes, IMAP operations, and a React-based UI component.

## Core Search Implementation

### 1. IMAP Search Criteria (`read.go:179-200`)

The primary search implementation uses the `go-imap` library's search criteria builder:

```go
// Build search criteria
criteria := imap.NewSearchCriteria()
if opts.UnreadOnly {
    criteria.WithoutFlags = []string{imap.SeenFlag}
}
if opts.FromAddress != "" {
    criteria.Header.Add("From", opts.FromAddress)
}
// Use config.FromEmail if no ToAddress is explicitly specified
toAddress := opts.ToAddress
if toAddress == "" && config.FromEmail != "" {
    toAddress = config.FromEmail
}
if toAddress != "" {
    criteria.Header.Add("To", toAddress)
}
if opts.Subject != "" {
    criteria.Header.Add("Subject", opts.Subject)
}
if !opts.Since.IsZero() {
    criteria.Since = opts.Since
}

// Search for messages
ids, err := c.Search(criteria)
```

**Location**: [`read.go:179-203`](read.go:179-203)

### 2. Search Options Structure (`read.go:33-44`)

The search functionality is configured through the `ReadOptions` struct:

```go
type ReadOptions struct {
    Limit            int
    UnreadOnly       bool
    FromAddress      string
    ToAddress        string
    Subject          string
    Since            time.Time
    LocalOnly        bool  // Only read from local storage
    SyncLocal        bool  // Sync received emails to local storage
    DownloadAttach   bool  // Download attachment content
    AttachmentDir    string // Directory to save attachments
}
```

**Location**: [`read.go:33-44`](read.go:33-44)

## Search Implementation Locations

### 1. Live IMAP Search (`read.go:84-267`)

- **Function**: `ReadFromFolder(opts ReadOptions, folder string)`
- **Capabilities**: 
  - IMAP server-side search using search criteria
  - Supports multiple filters: unread status, from/to addresses, subject, date ranges
  - Handles different folder types (INBOX, Drafts, etc.)
  - Connection management with TLS support

### 2. Local Storage Search (`read.go:526-601`)

- **Function**: `readFromLocalStorage(opts ReadOptions)`
- **Capabilities**:
  - Searches locally saved emails in `.email/received` directory
  - Client-side filtering by from address, subject, and date
  - Case-insensitive string matching
  - Sorting by date (newest first)

### 3. Global Inbox Search (`inbox.go:217-247`)

- **Function**: `GetEmailsFromInbox(accountEmail string, opts ReadOptions)`
- **Capabilities**:
  - Searches the global inbox cache for an account
  - Uses the same filtering logic as local storage
  - Optimized for quick access to recently synchronized emails

### 4. Incremental Search/Sync (`inbox.go:132-149`)

- **Function**: `FetchEmailsIncremental(config *Config, limit int)`
- **Implementation**:
```go
// Build search criteria for incremental fetch
criteria := imap.NewSearchCriteria()

// If we have a last email date, fetch emails since then
if !inboxData.LastEmailDate.IsZero() {
    // Add 1 second to avoid duplicate fetch of the last email
    since := inboxData.LastEmailDate.Add(1 * time.Second)
    criteria.Since = since
    fmt.Printf("Fetching emails since: %v\n", since.Format(time.RFC3339))
} else {
    // First time fetch - get emails from last 30 days
    since := time.Now().AddDate(0, 0, -30)
    criteria.Since = since
    fmt.Printf("First fetch - getting emails from last 30 days since: %v\n", since.Format(time.RFC3339))
}

// Search for messages
ids, err := c.Search(criteria)
```

**Location**: [`inbox.go:132-149`](inbox.go:132-149)

## Interactive Search Interface

### 1. CLI Interactive Mode (`interactive.go:193-269`)

The `/read` command in interactive mode supports search parameters:

```go
func handleReadCommand(args []string) error {
    opts := ReadOptions{
        Limit: 20, // Default
    }
    
    // Parse arguments
    for i := 0; i < len(args); i++ {
        switch args[i] {
        case "--unread":
            opts.UnreadOnly = true
        case "--from":
            if i+1 < len(args) {
                opts.FromAddress = args[i+1]
                i++
            }
        case "--days":
            if i+1 < len(args) {
                opts.Since = getTimeFromDays(args[i+1])
                i++
            }
        case "-n", "--number":
            if i+1 < len(args) {
                if n := parseNumber(args[i+1]); n > 0 {
                    opts.Limit = n
                }
                i++
            }
        }
    }
    // ... rest of implementation
}
```

**Location**: [`interactive.go:193-269`](interactive.go:193-269)

### 2. React-based Search UI (`ui/src/components/SearchBar.tsx`)

The frontend includes a search component:

```tsx
export const SearchBar: React.FC<SearchBarProps> = ({ onClose }) => {
  const [value, setValue] = useState('');
  const { setSearchQuery } = useStore();

  const handleSubmit = () => {
    setSearchQuery(value);
    onClose();
  };

  return (
    <Box borderStyle="single" paddingX={1}>
      <Text color="cyan">Search: </Text>
      <TextInput
        value={value}
        onChange={setValue}
        onSubmit={handleSubmit}
        placeholder="Type to search emails..."
      />
    </Box>
  );
};
```

**Location**: [`ui/src/components/SearchBar.tsx:10-29`](ui/src/components/SearchBar.tsx:10-29)

## Additional Search-Related Functionality

### 1. Sent Messages Search (`sent.go:100-117`)

Similar search implementation for sent emails folder:

```go
// Build search criteria
criteria := imap.NewSearchCriteria()
if opts.FromAddress != "" {
    criteria.Header.Add("From", opts.FromAddress)
}
if opts.ToAddress != "" {
    criteria.Header.Add("To", opts.ToAddress)
}
if opts.Subject != "" {
    criteria.Header.Add("Subject", opts.Subject)
}
if !opts.Since.IsZero() {
    criteria.Since = opts.Since
}

ids, err := c.Search(criteria)
```

**Location**: [`sent.go:100-117`](sent.go:100-117)

### 2. Subject-based Email Opening (`open.go:265`)

```go
func openBySubjectSearch(subject string) error {
    // Implementation for finding and opening emails by subject
}
```

**Location**: [`open.go:265`](open.go:265)

### 3. Case-Insensitive Search Utilities (`inbox.go:330-375`)

Helper functions for local search operations:

```go
// containsIgnoreCase checks if haystack contains needle (case insensitive)
func containsIgnoreCase(haystack, needle string) bool {
    return len(needle) == 0 || 
           len(haystack) >= len(needle) && 
           findIgnoreCase(haystack, needle) >= 0
}

// findIgnoreCase finds needle in haystack (case insensitive), returns index or -1
func findIgnoreCase(haystack, needle string) int {
    // ... implementation
}

// equalFoldSubstring compares two strings case-insensitively
func equalFoldSubstring(s1, s2 string) bool {
    // ... implementation
}
```

**Location**: [`inbox.go:330-375`](inbox.go:330-375)

## Search Flow Architecture

### 1. Command-Line Search Flow
```
User Input → CLI Parsing → ReadOptions → ReadFromFolder → IMAP Search → Results
```

### 2. Interactive Mode Search Flow
```
/read command → handleReadCommand → Parse Args → ReadOptions → Email Results
```

### 3. Local Storage Search Flow
```
ReadOptions → readFromLocalStorage → File System Scan → JSON Parse → Filter → Results
```

### 4. Global Inbox Search Flow
```
Account Email → LoadGlobalInbox → Filter by ReadOptions → Return Cached Results
```

## Key Search Features

1. **Server-side IMAP Search**: Leverages IMAP server search capabilities for efficiency
2. **Local Caching**: Maintains local copies for offline search
3. **Multiple Filter Types**: Supports filtering by sender, recipient, subject, date, and read status
4. **Case-Insensitive Matching**: Implements custom case-insensitive search for local operations
5. **Incremental Sync**: Only fetches new emails since last sync
6. **Folder Support**: Can search different email folders (INBOX, Drafts, Sent)
7. **Interactive Interface**: Both CLI and UI search interfaces
8. **Attachment Handling**: Can include attachment metadata in search results

## Error Handling

The search implementation includes comprehensive error handling for:
- IMAP connection failures
- Folder access issues (with fallback folder names)
- Local storage access problems
- JSON parsing errors
- Network connectivity issues

## Performance Considerations

1. **Limit Controls**: All search functions respect result limits to prevent memory issues
2. **Server-side Filtering**: Uses IMAP search to reduce data transfer
3. **Caching Strategy**: Maintains local caches to reduce server queries
4. **Incremental Updates**: Only fetches new emails to minimize bandwidth usage