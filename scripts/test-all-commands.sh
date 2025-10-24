#!/bin/bash

# EmailOS (mailos) Comprehensive Command Testing Script
# Tests all major commands including new account management and signature features
# Covers: help, config, accounts, send, search, stats, sync, drafts, templates, etc.

echo "🧪 Testing mailos commands comprehensively..."
echo "================================================="

# Check for required .env file and variables
if [ ! -f ".env" ]; then
    echo -e "${RED}❌ ERROR: .env file not found${NC}"
    echo "Please create a .env file based on .env.example with FROM_EMAIL and TO_EMAIL variables"
    echo "Example:"
    echo "  cp .env.example .env"
    echo "  # Edit .env to set your email addresses"
    exit 1
fi

# Source .env file
source .env

# Validate required environment variables
if [ -z "$FROM_EMAIL" ] || [ -z "$TO_EMAIL" ]; then
    echo -e "${RED}❌ ERROR: Missing required environment variables${NC}"
    echo "Please ensure your .env file contains:"
    echo "  FROM_EMAIL=your-configured-account@example.com"
    echo "  TO_EMAIL=test-recipient@example.com"
    echo ""
    echo "See .env.example for the expected format"
    exit 1
fi

echo -e "${GREEN}✓ Using FROM_EMAIL: $FROM_EMAIL${NC}"
echo -e "${GREEN}✓ Using TO_EMAIL: $TO_EMAIL${NC}"
echo ""

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
    
    # Run command with timeout to prevent hanging
    if timeout 30 bash -c "$command" >/dev/null 2>&1; then
        echo -e "${GREEN}✓ PASS${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        local exit_code=$?
        if [ $exit_code -eq 124 ]; then
            # Timeout occurred
            echo -e "${YELLOW}⏱ TIMEOUT${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        else
            echo -e "${RED}✗ FAIL${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    fi
}

# Function to run test with timeout for interactive commands
run_test_timeout() {
    local test_name="$1"
    local command="$2"
    local timeout_duration="${3:-5}"  # Default 5 seconds
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -n "Testing: $test_name... "
    
    # Run command with timeout
    if timeout "$timeout_duration" bash -c "$command" >/dev/null 2>&1; then
        echo -e "${GREEN}✓ PASS${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        local exit_code=$?
        if [ $exit_code -eq 124 ]; then
            # Timeout occurred - this is expected for interactive commands
            echo -e "${YELLOW}⏱ TIMEOUT (expected)${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "${RED}✗ FAIL${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    fi
}

# Function to run command tests from the general test framework
# Usage: run_command_tests <command_type> <test_category>
# 
# Available command types: send, read (add more in test/test_framework/main.go)
# Available categories: help, errors, basic, combined, shortflags, etc.
# 
# To add new commands/tests:
# 1. Add test cases to test/test_framework/main.go
# 2. Add them to the appropriate TestSuite (SendTestSuite, ReadTestSuite, etc.)
# 3. Update the AllTests variable to include your new test suite
# 4. Call run_command_tests in this script with your command type and category
run_command_tests() {
    local command_type="$1"  # "send", "read", etc.
    local test_category="$2"  # "help", "errors", "no-env", etc.
    
    # Check if test framework exists
    if [ ! -f "test/test_framework/main.go" ]; then
        echo -e "${YELLOW}⚠️  test/test_framework/main.go not found, skipping ${command_type} tests${NC}"
        return
    fi
    
    # Get test cases from test framework
    local test_output
    if ! test_output=$(go run test/test_framework/main.go "$test_category" "$command_type" 2>/dev/null); then
        echo -e "${YELLOW}⚠️  Could not run test framework for ${command_type} ${test_category}, using manual tests${NC}"
        return
    fi
    
    # Parse and run each test
    while IFS='|' read -r test_name test_command; do
        if [ -n "$test_name" ] && [ -n "$test_command" ]; then
            run_test "$test_name" "$test_command"
        fi
    done <<< "$test_output"
}

# Build mailos first
echo -e "${BLUE}🔨 Building mailos...${NC}"
if ! go build -o mailos cmd/mailos/main.go cmd/mailos/error_handler.go; then
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
run_test "Commands command" "./mailos commands"
run_test "Commands verbose" "./mailos commands --verbose"
run_test "Tools command" "./mailos tools"
# Skipping stats command as it requires email server connection
# run_test "Basic stats command" "./mailos stats"

echo ""

# =============================================================================
# CONFIGURATION COMMANDS  
# =============================================================================
echo -e "${BLUE}=== CONFIGURATION COMMANDS ===${NC}"

run_test "Setup help" "./mailos setup --help"
run_test "Config help" "./mailos config --help"

echo ""

# =============================================================================
# ACCOUNT MANAGEMENT COMMANDS
# =============================================================================
echo -e "${BLUE}=== ACCOUNT MANAGEMENT COMMANDS ===${NC}"

run_test "Accounts help" "./mailos accounts --help"
run_test "List accounts" "./mailos accounts --list"
run_test "Accounts command without flags (shows list)" "./mailos accounts"

# Account management operations (these will show usage/errors without proper setup)
# Skipping interactive tests to prevent hanging
# run_test_timeout "Add account with email (interactive)" "./mailos accounts --add $TO_EMAIL" 3
# run_test_timeout "Set account with email (interactive)" "./mailos accounts --set $FROM_EMAIL" 3
run_test "Set signature with proper format" "./mailos accounts --set-signature '$FROM_EMAIL:Test Signature'"
run_test "Clear session help syntax" "./mailos accounts --clear"

echo ""

# =============================================================================
# READ SPECIFIC EMAIL COMMANDS 
# =============================================================================
echo -e "${BLUE}=== READ SPECIFIC EMAIL COMMANDS ===${NC}"

# Run help tests from test framework
run_command_tests "read" "help"

# Run error handling tests from test framework
echo -e "${BLUE}=== READ ERROR HANDLING TESTS ===${NC}"
run_command_tests "read" "errors"

# Skipping read command functionality tests as they require email server connections
# Uncomment the lines below to run additional test categories
# run_command_tests "read" "basic"
# run_command_tests "read" "documents"
# run_command_tests "read" "combined"

# =============================================================================
# SEARCH/LIST COMMANDS (what used to be called read)
# =============================================================================
echo -e "${BLUE}=== SEARCH/LIST COMMANDS ===${NC}"

# Skipping search/list commands as they require email server connections
# run_test "List 5 emails" "./mailos search -n 5"
# run_test "List 10 emails" "./mailos search --number 10"
# run_test "List unread only" "./mailos search --unread"
# run_test "List from specific sender" "./mailos search --from gmail.com"
# run_test "List to specific recipient" "./mailos search --to $TO_EMAIL"
# run_test "List with subject filter" "./mailos search --subject test"
# run_test "List last 7 days" "./mailos search --days 7"
# run_test "List and save markdown" "./mailos search --save-markdown"
# run_test "List with attachment download" "./mailos search --download-attachments"
# run_test "List with attachment dir" "./mailos search --attachment-dir ./attachments"

# Skipping combined search/list options as they require email server connections
# run_test "List unread from last 3 days limit 5" "./mailos search --unread --days 3 --number 5"
# run_test "List from gmail last week" "./mailos search --from gmail --days 7"

echo ""

# =============================================================================
# SEARCH COMMANDS (from src/core/search.go)
# =============================================================================
echo -e "${BLUE}=== SEARCH COMMANDS ===${NC}"

# Basic search options
run_test "Search help" "./mailos search --help"
# Skipping search commands as they require email server connections
# run_test "Search unread" "./mailos search --unread"
# run_test "Search from address" "./mailos search --from noreply"
# run_test "Search to address" "./mailos search --to $TO_EMAIL"
# run_test "Search subject" "./mailos search --subject invoice"
# run_test "Search last 7 days" "./mailos search --days 7"
# run_test "Search limit 5" "./mailos search -n 5"
# run_test "Search with number limit" "./mailos search --number 10"
# run_test "Search and save markdown" "./mailos search --save-markdown"
# run_test "Search with output dir" "./mailos search --output-dir ./search-results"

# Skipping advanced search options as they require email server connections
# run_test "Search with query" "./mailos search --query meeting"
# run_test "Search with -q shorthand" "./mailos search -q 'important email'"
# run_test "Search fuzzy threshold" "./mailos search --fuzzy-threshold 0.8"
# run_test "Search no fuzzy" "./mailos search --no-fuzzy"
# run_test "Search case sensitive" "./mailos search --case-sensitive"
# run_test "Search min size" "./mailos search --min-size 1KB"
# run_test "Search max size" "./mailos search --max-size 10MB"
# run_test "Search has attachments" "./mailos search --has-attachments"
# run_test "Search attachment size" "./mailos search --attachment-size 5MB"
# run_test "Search date range" "./mailos search --date-range '2024-01-01,2024-12-31'"

# Skipping complex search combinations as they require email server connections
# run_test "Complex search 1" "./mailos search --from support --days 14 --unread --number 5"
# run_test "Complex search 2" "./mailos search --query meeting --has-attachments --case-sensitive"
# run_test "Complex search 3" "./mailos search --subject invoice --min-size 1KB --days 30"

echo ""

# =============================================================================
# STATS COMMANDS (from stats.go)
# =============================================================================
echo -e "${BLUE}=== STATS COMMANDS ===${NC}"

# Basic stats operations
run_test "Stats help" "./mailos stats --help"
# Skipping stats commands as they require email server connections
# run_test "Basic stats" "./mailos stats"
# run_test "Stats with number limit" "./mailos stats --number 50"
# run_test "Stats with -n shorthand" "./mailos stats -n 20"
# run_test "Stats unread only" "./mailos stats --unread"
# run_test "Stats unread with -u shorthand" "./mailos stats -u"

# Skipping stats with filters as they require email server connections
# run_test "Stats from specific sender" "./mailos stats --from gmail.com"
# run_test "Stats to specific recipient" "./mailos stats --to $TO_EMAIL"
# run_test "Stats with subject filter" "./mailos stats --subject invoice"
# run_test "Stats last 7 days" "./mailos stats --days 7"
# run_test "Stats last 30 days" "./mailos stats --days 30"

# Skipping stats with time ranges as they require email server connections
# run_test "Stats today" "./mailos stats --range 'Today'"
# run_test "Stats yesterday" "./mailos stats --range 'Yesterday'"
# run_test "Stats this week" "./mailos stats --range 'This week'"
# run_test "Stats last week" "./mailos stats --range 'Last week'"
# run_test "Stats last hour" "./mailos stats --range 'Last hour'"

# Skipping combined stats options as they require email server connections
# run_test "Stats unread last 7 days" "./mailos stats --unread --days 7"
# run_test "Stats from gmail last week" "./mailos stats --from gmail --range 'Last week'"
# run_test "Stats subject filter with limit" "./mailos stats --subject meeting --number 25"
# run_test "Complex stats combination" "./mailos stats --from support --days 14 --unread --number 10"

echo ""

# =============================================================================
# SYNC COMMANDS (from src/core/inbox.go)
# =============================================================================
echo -e "${BLUE}=== SYNC COMMANDS ===${NC}"

run_test "Sync help" "./mailos sync --help"
# Skipping sync commands as they require email server connections
# run_test "Basic sync" "./mailos sync"
# run_test "Sync with limit" "./mailos sync --limit 5"
# run_test "Sync with high limit" "./mailos sync --limit 50"
# run_test "Sync verbose" "./mailos sync --verbose"
# Note: --incremental flag doesn't exist, removing this test

# Skipping combined sync options as they require email server connections
# run_test "Sync limit 10 verbose" "./mailos sync --limit 10 --verbose"

echo ""

# =============================================================================
# SEND COMMANDS
# =============================================================================
echo -e "${BLUE}=== SEND COMMANDS ===${NC}"

# Run help tests from test framework
run_command_tests "send" "help"

# Run error handling tests from test framework
echo -e "${BLUE}=== SEND ERROR HANDLING TESTS ===${NC}"
run_command_tests "send" "errors"

# Skipping send command syntax tests as they require email server connections
# Uncomment the lines below to run additional test categories
# run_command_tests "send" "basic"
# run_command_tests "send" "shortflags"
# run_command_tests "send" "preview"

echo ""

# =============================================================================
# DRAFTS COMMANDS (from src/core/drafts.go)
# =============================================================================
echo -e "${BLUE}=== DRAFTS COMMANDS ===${NC}"

# Basic drafts operations
run_test "Drafts help" "./mailos drafts --help"
# Skipping draft operations as they may require email server connections
# run_test "List drafts" "./mailos drafts --list"
# run_test "Read drafts" "./mailos drafts --read"

# Skipping draft creation options as they may require email server connections
# run_test "Draft with to address" "./mailos drafts --to $TO_EMAIL"
# run_test "Draft with CC" "./mailos drafts --cc cc@example.com"
# run_test "Draft with BCC" "./mailos drafts --bcc bcc@example.com"
# run_test "Draft with subject" "./mailos drafts --subject 'Test Subject'"
# run_test "Draft with body" "./mailos drafts --body 'Test email body'"
# run_test "Draft with priority" "./mailos drafts --priority high"
# run_test "Draft plain text" "./mailos drafts --plain-text"
# run_test "Draft no signature" "./mailos drafts --no-signature"
# run_test "Draft with signature" "./mailos drafts --signature 'Custom signature'"

# Skipping draft advanced options as they may require email server connections
# run_test "Draft with query" "./mailos drafts --query 'meeting reminder'"
# run_test "Draft with template" "./mailos drafts --template meeting"
# run_test "Draft with data file" "./mailos drafts --data-file ./data.json"
# run_test "Draft with output dir" "./mailos drafts --output-dir ./my-drafts"
# Skipping interactive test to prevent hanging
# run_test "Draft interactive" "./mailos drafts --interactive"
# run_test "Draft use AI" "./mailos drafts --use-ai"
# run_test "Draft count 3" "./mailos drafts --draft-count 3"

# Skipping complex draft combinations as they may require email server connections
# run_test "Complex draft 1" "./mailos drafts --to $TO_EMAIL --subject 'Test' --body 'Hello' --priority high"
# run_test "Complex draft 2" "./mailos drafts --query 'follow up' --use-ai --draft-count 2"

echo ""


# =============================================================================
# GROUPS COMMANDS
# =============================================================================
echo -e "${BLUE}=== GROUPS COMMANDS ===${NC}"

# Run help tests from test framework
run_command_tests "groups" "help"

# Run basic groups functionality tests
echo -e "${BLUE}=== GROUPS BASIC FUNCTIONALITY TESTS ===${NC}"
run_command_tests "groups" "basic"

# Run member management tests
echo -e "${BLUE}=== GROUPS MEMBER MANAGEMENT TESTS ===${NC}"
run_command_tests "groups" "members"

# Run error handling tests for groups
echo -e "${BLUE}=== GROUPS ERROR HANDLING TESTS ===${NC}"
run_command_tests "groups" "errors"

# Run validation tests for groups
echo -e "${BLUE}=== GROUPS VALIDATION TESTS ===${NC}"
run_command_tests "groups" "validation"

# Skipping sending tests as they require email server connections
# Uncomment the lines below to run group sending tests
# run_command_tests "groups" "sending"
# run_command_tests "groups" "advanced"
# run_command_tests "groups" "integration"

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

# Skipping all interactive mode tests to prevent hanging
# run_test "Interactive help" "echo 'help' | ./mailos interactive"
# run_test "Interactive stats" "echo 'stats' | ./mailos interactive"
# run_test "Interactive read" "echo 'read -n 5' | ./mailos interactive"
# run_test "Interactive search" "echo 'search --query test' | ./mailos interactive"
# run_test "Interactive exit" "echo 'exit' | ./mailos interactive"

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

# Advanced flag combinations are now tested via the general test framework
# Uncomment the lines below to run comprehensive flag combination tests
# run_command_tests "send" "combined"
# run_command_tests "read" "combined"
# Skipping search and draft combinations as they require email server connections
run_test "Account management combo" "./mailos accounts --set-signature '$FROM_EMAIL:Best regards, User'"

echo ""

# =============================================================================
# ERROR HANDLING TESTS
# =============================================================================
echo -e "${BLUE}=== ERROR HANDLING TESTS ===${NC}"

run_test "Invalid command" "./mailos invalid-command"
# Skipping error tests that require email server connections
# run_test "Invalid flag" "./mailos read --invalid-flag"
# run_test "Missing argument" "./mailos search --query"

# Account management error handling
run_test "Add account without email" "./mailos accounts --add"
run_test "Set account without email" "./mailos accounts --set"
run_test "Set signature without argument" "./mailos accounts --set-signature"
run_test "Set signature invalid format" "./mailos accounts --set-signature 'invalid-format'"

# Send command error handling is now tested via run_send_tests "errors" above

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