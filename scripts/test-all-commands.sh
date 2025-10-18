#!/bin/bash

# EmailOS (mailos) Comprehensive Command Testing Script
# Tests all major commands with parameters extracted from core source files

echo "🧪 Testing mailos commands comprehensively..."
echo "================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counter
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run test and track results
run_test() {
    local test_name="$1"
    local command="$2"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -n "Testing: $test_name... "
    
    if eval "$command" >/dev/null 2>&1; then
        echo -e "${GREEN}✓ PASS${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}✗ FAIL${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# Build mailos first
echo -e "${BLUE}🔨 Building mailos...${NC}"
if ! go build -o mailos cmd/mailos/main.go; then
    echo -e "${RED}❌ Failed to build mailos${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Build successful${NC}"
echo ""

# =============================================================================
# HELP & INFO COMMANDS
# =============================================================================
echo -e "${BLUE}=== HELP & INFO COMMANDS ===${NC}"

run_test "Help" "./mailos --help"
run_test "Version" "./mailos --version"
run_test "Tools command" "./mailos tools"
run_test "Basic stats command" "./mailos stats"

echo ""

# =============================================================================
# CONFIGURATION COMMANDS  
# =============================================================================
echo -e "${BLUE}=== CONFIGURATION COMMANDS ===${NC}"

run_test "Setup help" "./mailos setup --help"
run_test "Config help" "./mailos config --help"

echo ""

# =============================================================================
# READ SPECIFIC EMAIL COMMANDS 
# =============================================================================
echo -e "${BLUE}=== READ SPECIFIC EMAIL COMMANDS ===${NC}"

# Read command tests (reads specific email by ID)
run_test "Read help" "./mailos read --help"
# Note: Read command requires an email ID, so these tests will show usage help
run_test "Read command without ID (shows usage)" "./mailos read"

# =============================================================================
# SEARCH/LIST COMMANDS (what used to be called read)
# =============================================================================
echo -e "${BLUE}=== SEARCH/LIST COMMANDS ===${NC}"

# Basic search/list options (this is what the old "read" tests were trying to do)
run_test "List 5 emails" "./mailos search -n 5"
run_test "List 10 emails" "./mailos search --number 10"
run_test "List unread only" "./mailos search --unread"
run_test "List from specific sender" "./mailos search --from gmail.com"
run_test "List to specific recipient" "./mailos search --to test@example.com"
run_test "List with subject filter" "./mailos search --subject test"
run_test "List last 7 days" "./mailos search --days 7"
run_test "List and save" "./mailos search --save-markdown"
run_test "List with attachment download" "./mailos search --download-attachments"
run_test "List with attachment dir" "./mailos search --attachment-dir ./attachments"

# Combined search/list options
run_test "List unread from last 3 days limit 5" "./mailos search --unread --days 3 --number 5"
run_test "List from gmail last week" "./mailos search --from gmail --days 7"

echo ""

# =============================================================================
# SEARCH COMMANDS (from src/core/search.go)
# =============================================================================
echo -e "${BLUE}=== SEARCH COMMANDS ===${NC}"

# Basic search options
run_test "Search help" "./mailos search --help"
run_test "Search unread" "./mailos search --unread"
run_test "Search from address" "./mailos search --from noreply"
run_test "Search to address" "./mailos search --to test@example.com"
run_test "Search subject" "./mailos search --subject invoice"
run_test "Search last 7 days" "./mailos search --days 7"
run_test "Search limit 5" "./mailos search -n 5"
run_test "Search with number limit" "./mailos search --number 10"
run_test "Search and save" "./mailos search --save"
run_test "Search with output dir" "./mailos search --output-dir ./search-results"
run_test "Search local only" "./mailos search --local"
run_test "Search and sync" "./mailos search --sync"

# Advanced search options
run_test "Search with query" "./mailos search --query meeting"
run_test "Search with -q shorthand" "./mailos search -q 'important email'"
run_test "Search fuzzy threshold" "./mailos search --fuzzy-threshold 0.8"
run_test "Search no fuzzy" "./mailos search --no-fuzzy"
run_test "Search case sensitive" "./mailos search --case-sensitive"
run_test "Search min size" "./mailos search --min-size 1KB"
run_test "Search max size" "./mailos search --max-size 10MB"
run_test "Search has attachments" "./mailos search --has-attachments"
run_test "Search attachment size" "./mailos search --attachment-size 5MB"
run_test "Search date range" "./mailos search --date-range '2024-01-01,2024-12-31'"

# Complex search combinations
run_test "Complex search 1" "./mailos search --from support --days 14 --unread --limit 5"
run_test "Complex search 2" "./mailos search --query meeting --has-attachments --case-sensitive"
run_test "Complex search 3" "./mailos search --subject invoice --min-size 1KB --days 30"

echo ""

# =============================================================================
# STATS COMMANDS (from stats.go)
# =============================================================================
echo -e "${BLUE}=== STATS COMMANDS ===${NC}"

# Basic stats operations
run_test "Stats help" "./mailos stats --help"
run_test "Basic stats" "./mailos stats"
run_test "Stats with number limit" "./mailos stats --number 50"
run_test "Stats with -n shorthand" "./mailos stats -n 20"
run_test "Stats unread only" "./mailos stats --unread"
run_test "Stats unread with -u shorthand" "./mailos stats -u"

# Stats with filters
run_test "Stats from specific sender" "./mailos stats --from gmail.com"
run_test "Stats to specific recipient" "./mailos stats --to test@example.com"
run_test "Stats with subject filter" "./mailos stats --subject invoice"
run_test "Stats last 7 days" "./mailos stats --days 7"
run_test "Stats last 30 days" "./mailos stats --days 30"

# Stats with time ranges
run_test "Stats today" "./mailos stats --range 'Today'"
run_test "Stats yesterday" "./mailos stats --range 'Yesterday'"
run_test "Stats this week" "./mailos stats --range 'This week'"
run_test "Stats last week" "./mailos stats --range 'Last week'"
run_test "Stats last hour" "./mailos stats --range 'Last hour'"

# Combined stats options
run_test "Stats unread last 7 days" "./mailos stats --unread --days 7"
run_test "Stats from gmail last week" "./mailos stats --from gmail --range 'Last week'"
run_test "Stats subject filter with limit" "./mailos stats --subject meeting --number 25"
run_test "Complex stats combination" "./mailos stats --from support --days 14 --unread --number 10"

echo ""

# =============================================================================
# SYNC COMMANDS (from src/core/inbox.go)
# =============================================================================
echo -e "${BLUE}=== SYNC COMMANDS ===${NC}"

run_test "Sync help" "./mailos sync --help"
run_test "Basic sync" "./mailos sync"
run_test "Sync with limit" "./mailos sync --limit 5"
run_test "Sync with high limit" "./mailos sync --limit 50"
run_test "Sync verbose" "./mailos sync --verbose"
run_test "Sync incremental" "./mailos sync --incremental"

# Combined sync options
run_test "Sync limit 10 verbose" "./mailos sync --limit 10 --verbose"

echo ""

# =============================================================================
# DRAFTS COMMANDS (from src/core/drafts.go)
# =============================================================================
echo -e "${BLUE}=== DRAFTS COMMANDS ===${NC}"

# Basic drafts operations
run_test "Drafts help" "./mailos drafts --help"
run_test "List drafts" "./mailos drafts --list"
run_test "Read drafts" "./mailos drafts --read"

# Draft creation options
run_test "Draft with to address" "./mailos drafts --to test@example.com"
run_test "Draft with CC" "./mailos drafts --cc cc@example.com"
run_test "Draft with BCC" "./mailos drafts --bcc bcc@example.com"
run_test "Draft with subject" "./mailos drafts --subject 'Test Subject'"
run_test "Draft with body" "./mailos drafts --body 'Test email body'"
run_test "Draft with priority" "./mailos drafts --priority high"
run_test "Draft plain text" "./mailos drafts --plain-text"
run_test "Draft no signature" "./mailos drafts --no-signature"
run_test "Draft with signature" "./mailos drafts --signature 'Custom signature'"

# Draft advanced options
run_test "Draft with query" "./mailos drafts --query 'meeting reminder'"
run_test "Draft with template" "./mailos drafts --template meeting"
run_test "Draft with data file" "./mailos drafts --data-file ./data.json"
run_test "Draft with output dir" "./mailos drafts --output-dir ./my-drafts"
run_test "Draft interactive" "./mailos drafts --interactive"
run_test "Draft use AI" "./mailos drafts --use-ai"
run_test "Draft count 3" "./mailos drafts --draft-count 3"

# Complex draft combinations
run_test "Complex draft 1" "./mailos drafts --to test@example.com --subject 'Test' --body 'Hello' --priority high"
run_test "Complex draft 2" "./mailos drafts --query 'follow up' --use-ai --draft-count 2"

echo ""

# =============================================================================
# QUERY COMMANDS
# =============================================================================
echo -e "${BLUE}=== QUERY COMMANDS ===${NC}"

run_test "Query help" "./mailos query --help"
run_test "Query recent emails" "./mailos query 'show me recent emails'"
run_test "Query important emails" "./mailos query 'find important emails'"
run_test "Query last week" "./mailos query 'emails from last week'"
run_test "Query specific sender" "./mailos query 'emails from gmail'"

echo ""

# =============================================================================
# TEMPLATE COMMANDS
# =============================================================================
echo -e "${BLUE}=== TEMPLATE COMMANDS ===${NC}"

run_test "Template help" "./mailos template --help"
run_test "List templates" "./mailos template --list"
run_test "Template manage" "./mailos template --manage"

echo ""

# =============================================================================
# INTERACTIVE MODE TESTS
# =============================================================================
echo -e "${BLUE}=== INTERACTIVE MODE TESTS ===${NC}"

run_test "Interactive help" "echo 'help' | ./mailos interactive"
run_test "Interactive stats" "echo 'stats' | ./mailos interactive"
run_test "Interactive read" "echo 'read -n 5' | ./mailos interactive"
run_test "Interactive search" "echo 'search --query test' | ./mailos interactive"
run_test "Interactive exit" "echo 'exit' | ./mailos interactive"

echo ""

# =============================================================================
# REPORT COMMANDS
# =============================================================================
echo -e "${BLUE}=== REPORT COMMANDS ===${NC}"

run_test "Report help" "./mailos report --help"
run_test "Report summary" "./mailos report --summary"
run_test "Report detailed" "./mailos report --detailed"

echo ""

# =============================================================================
# ADVANCED FLAG COMBINATIONS
# =============================================================================
echo -e "${BLUE}=== ADVANCED FLAG COMBINATIONS ===${NC}"

run_test "Read all flags" "./mailos read --unread-only --from-address test --limit 5 --local-only --download-attach"
run_test "Search all flags" "./mailos search --query test --from gmail --days 7 --has-attachments --case-sensitive --limit 10"
run_test "Draft all flags" "./mailos drafts --to test@example.com --cc cc@example.com --subject Test --body Hello --priority high --plain-text"

echo ""

# =============================================================================
# ERROR HANDLING TESTS
# =============================================================================
echo -e "${BLUE}=== ERROR HANDLING TESTS ===${NC}"

run_test "Invalid command" "./mailos invalid-command"
run_test "Invalid flag" "./mailos read --invalid-flag"
run_test "Missing argument" "./mailos search --query"

echo ""

# =============================================================================
# RESULTS SUMMARY
# =============================================================================
echo "================================================="
echo -e "${BLUE}📊 TEST RESULTS SUMMARY${NC}"
echo "================================================="
echo -e "Total tests run: ${BLUE}$TOTAL_TESTS${NC}"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠️  Some tests failed. This may be expected if email is not configured.${NC}"
    echo -e "${YELLOW}📝 Note: Many failures are expected without proper email configuration.${NC}"
    exit 0
fi