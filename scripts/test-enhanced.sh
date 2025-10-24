#!/bin/bash

# Enhanced EmailOS Test Runner - Modern Jest/Pytest-like experience
# Integrates with existing test-all-commands.sh structure while adding modern features

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
WATCH_MODE=false
COVERAGE=false
VERBOSE=false
PARALLEL=false
FILTER=""
TIMEOUT="30s"
OUTPUT_FORMAT="pretty"
TEST_PATTERN="all"
MOCK_MODE=false
ENV_CHECK=true

# Test statistics
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0
START_TIME=$(date +%s)

# Print usage information
usage() {
    cat << EOF
Enhanced EmailOS Test Runner

Usage: $0 [OPTIONS] [TEST_PATTERN]

OPTIONS:
    -w, --watch         Watch mode - re-run tests on file changes
    -c, --coverage      Generate test coverage report
    -v, --verbose       Verbose output
    -p, --parallel      Run tests in parallel
    -f, --filter FILTER Filter tests by name/pattern
    -t, --timeout TIME  Test timeout (default: 30s)
    -o, --output FORMAT Output format (pretty, json, tap)
    -m, --mock          Run with mocked dependencies only
    -s, --skip-env      Skip environment variable checks
    -h, --help          Show this help message

TEST PATTERNS:
    all                 Run all tests (default)
    unit                Run only unit tests
    integration         Run only integration tests
    framework           Run test framework tests
    help                Run help command tests
    errors              Run error handling tests
    send                Run send command tests
    read                Run read command tests
    search              Run search command tests
    groups              Run groups command tests
    drafts              Run drafts command tests
    quick               Run quick validation tests

EXAMPLES:
    $0                  Run all tests
    $0 -c unit          Run unit tests with coverage
    $0 -w -v            Watch mode with verbose output
    $0 -f "Email"       Run tests matching "Email"
    $0 --mock           Run with mocked dependencies only
    $0 send --verbose   Run send tests with verbose output

EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -w|--watch)
                WATCH_MODE=true
                shift
                ;;
            -c|--coverage)
                COVERAGE=true
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -p|--parallel)
                PARALLEL=true
                shift
                ;;
            -f|--filter)
                FILTER="$2"
                shift 2
                ;;
            -t|--timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_FORMAT="$2"
                shift 2
                ;;
            -m|--mock)
                MOCK_MODE=true
                shift
                ;;
            -s|--skip-env)
                ENV_CHECK=false
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                TEST_PATTERN="$1"
                shift
                ;;
        esac
    done
}

# Enhanced environment setup
setup_test_environment() {
    echo -e "${BLUE}🔧 Setting up enhanced test environment...${NC}"
    
    # Create necessary directories
    mkdir -p test/reports
    mkdir -p test/coverage
    mkdir -p test/temp
    mkdir -p test/logs
    
    # Set test-specific environment variables
    export MAILOS_TEST_MODE=true
    export MAILOS_CONFIG_DIR="test/temp"
    export GO_TEST_TIMEOUT="$TIMEOUT"
    
    if [ "$MOCK_MODE" = true ]; then
        export MAILOS_MOCK_MODE=true
        echo -e "${YELLOW}🎭 Running in mock mode - external dependencies will be mocked${NC}"
    fi
    
    # Environment variable check (from original test-all-commands.sh)
    if [ "$ENV_CHECK" = true ] && [ "$MOCK_MODE" = false ]; then
        if [ ! -f ".env" ]; then
            echo -e "${YELLOW}⚠️  .env file not found - some tests may be skipped${NC}"
            echo -e "${CYAN}💡 Create .env with FROM_EMAIL and TO_EMAIL for full testing${NC}"
        else
            source .env
            if [ -n "$FROM_EMAIL" ] && [ -n "$TO_EMAIL" ]; then
                echo -e "${GREEN}✓ Using FROM_EMAIL: $FROM_EMAIL${NC}"
                echo -e "${GREEN}✓ Using TO_EMAIL: $TO_EMAIL${NC}"
            else
                echo -e "${YELLOW}⚠️  FROM_EMAIL or TO_EMAIL not set - some tests will be skipped${NC}"
            fi
        fi
    fi
    
    echo -e "${GREEN}✓ Enhanced test environment ready${NC}"
}

# Build test binaries
build_test_binaries() {
    echo -e "${BLUE}🔨 Building test binaries...${NC}"
    
    # Use the same build process as the original script
    if ! go build -o mailos cmd/mailos/main.go cmd/mailos/error_handler.go; then
        echo -e "${RED}❌ Failed to build mailos${NC}"
        exit 1
    fi
    
    # Also build test framework
    if [ -f "test/test_framework/main.go" ]; then
        if ! go build -o test/temp/test-framework test/test_framework/main.go; then
            echo -e "${YELLOW}⚠️  Failed to build test framework, using go run instead${NC}"
        fi
    fi
    
    echo -e "${GREEN}✓ Test binaries built${NC}"
}

# Enhanced test function with timing and better output
run_enhanced_test() {
    local test_name="$1"
    local command="$2"
    local category="${3:-general}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    local start_time=$(date +%s.%N)
    
    if [ "$VERBOSE" = true ]; then
        echo -e "${CYAN}🧪 Running: $test_name${NC}"
        echo -e "${PURPLE}   Command: $command${NC}"
    else
        echo -n "  Testing: $test_name... "
    fi
    
    # Apply filter if specified
    if [ -n "$FILTER" ] && [[ ! "$test_name" =~ $FILTER ]]; then
        if [ "$VERBOSE" = true ]; then
            echo -e "${YELLOW}   ⏭️  SKIPPED (filter)${NC}"
        else
            echo -e "${YELLOW}⏭️  SKIP${NC}"
        fi
        SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
        return
    fi
    
    # Run command with timeout
    local output
    local exit_code=0
    
    if [ "$VERBOSE" = true ]; then
        if timeout 30 bash -c "$command" 2>&1; then
            local end_time=$(date +%s.%N)
            local duration=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0")
            echo -e "${GREEN}   ✓ PASS${NC} ${PURPLE}(${duration}s)${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            exit_code=$?
            local end_time=$(date +%s.%N)
            local duration=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0")
            if [ $exit_code -eq 124 ]; then
                echo -e "${YELLOW}   ⏱️  TIMEOUT${NC} ${PURPLE}(${duration}s)${NC}"
            else
                echo -e "${RED}   ✗ FAIL${NC} ${PURPLE}(${duration}s)${NC}"
            fi
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    else
        if timeout 30 bash -c "$command" >/dev/null 2>&1; then
            echo -e "${GREEN}✓ PASS${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            exit_code=$?
            if [ $exit_code -eq 124 ]; then
                echo -e "${YELLOW}⏱️  TIMEOUT${NC}"
            else
                echo -e "${RED}✗ FAIL${NC}"
            fi
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    fi
    
    # Log test results
    local end_time=$(date +%s.%N)
    local duration=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0")
    echo "$(date '+%Y-%m-%d %H:%M:%S'),$test_name,$category,$exit_code,$duration" >> test/logs/test_results.csv
}

# Run command tests from framework (enhanced version of original function)
run_enhanced_command_tests() {
    local command_type="$1"
    local test_category="$2"
    
    echo -e "${CYAN}📋 Running $command_type $test_category tests...${NC}"
    
    # Check if test framework exists
    if [ ! -f "test/test_framework/main.go" ]; then
        echo -e "${YELLOW}⚠️  test/test_framework/main.go not found, skipping ${command_type} tests${NC}"
        return
    fi
    
    # Get test cases from test framework
    local test_output
    if ! test_output=$(go run test/test_framework/main.go "$test_category" "$command_type" 2>/dev/null); then
        echo -e "${YELLOW}⚠️  Could not run test framework for ${command_type} ${test_category}${NC}"
        return
    fi
    
    # Parse and run each test
    while IFS='|' read -r test_name test_command; do
        if [ -n "$test_name" ] && [ -n "$test_command" ]; then
            run_enhanced_test "$test_name" "$test_command" "$command_type-$test_category"
        fi
    done <<< "$test_output"
}

# Run Go unit tests with enhanced features
run_go_unit_tests() {
    echo -e "${BLUE}🧪 Running Go unit tests...${NC}"
    
    local go_test_args=()
    
    if [ "$VERBOSE" = true ]; then
        go_test_args+=("-v")
    fi
    
    if [ "$COVERAGE" = true ]; then
        go_test_args+=("-coverprofile=test/coverage/coverage.out")
        go_test_args+=("-covermode=atomic")
    fi
    
    if [ -n "$FILTER" ]; then
        go_test_args+=("-run" "$FILTER")
    fi
    
    go_test_args+=("-timeout" "$TIMEOUT")
    
    # Run tests in test directory
    if [ -d "test/unit_tests" ]; then
        echo -e "${CYAN}  Running unit tests...${NC}"
        if go test "${go_test_args[@]}" ./test/unit_tests/...; then
            echo -e "${GREEN}  ✓ Unit tests passed${NC}"
        else
            echo -e "${RED}  ✗ Unit tests failed${NC}"
        fi
    fi
    
    # Run root level tests
    local root_tests=($(find . -maxdepth 1 -name "*_test.go" 2>/dev/null || true))
    if [ ${#root_tests[@]} -gt 0 ]; then
        echo -e "${CYAN}  Running root tests...${NC}"
        if go test "${go_test_args[@]}" .; then
            echo -e "${GREEN}  ✓ Root tests passed${NC}"
        else
            echo -e "${RED}  ✗ Root tests failed${NC}"
        fi
    fi
}

# Run specific test patterns
run_test_pattern() {
    local pattern="$1"
    
    case "$pattern" in
        "all")
            echo -e "${BLUE}🚀 Running comprehensive test suite...${NC}"
            run_help_tests
            run_error_tests
            run_command_tests
            run_go_unit_tests
            ;;
        "unit")
            run_go_unit_tests
            ;;
        "integration")
            echo -e "${BLUE}🔗 Running integration tests...${NC}"
            run_enhanced_command_tests "send" "basic"
            run_enhanced_command_tests "read" "basic"
            run_enhanced_command_tests "groups" "basic"
            ;;
        "framework")
            run_framework_summary
            ;;
        "help")
            run_help_tests
            ;;
        "errors")
            run_error_tests
            ;;
        "send")
            run_enhanced_command_tests "send" "help"
            run_enhanced_command_tests "send" "errors"
            if [ "$MOCK_MODE" = false ]; then
                run_enhanced_command_tests "send" "basic"
            fi
            ;;
        "read")
            run_enhanced_command_tests "read" "help"
            run_enhanced_command_tests "read" "errors"
            if [ "$MOCK_MODE" = false ]; then
                run_enhanced_command_tests "read" "basic"
            fi
            ;;
        "search")
            echo -e "${BLUE}🔍 Running search command tests...${NC}"
            run_enhanced_test "Search help" "./mailos search --help" "search"
            ;;
        "groups")
            run_enhanced_command_tests "groups" "help"
            run_enhanced_command_tests "groups" "basic"
            run_enhanced_command_tests "groups" "errors"
            ;;
        "drafts")
            echo -e "${BLUE}📝 Running drafts command tests...${NC}"
            run_enhanced_test "Drafts help" "./mailos drafts --help" "drafts"
            ;;
        "quick")
            run_quick_tests
            ;;
        *)
            echo -e "${RED}❌ Unknown test pattern: $pattern${NC}"
            echo "Available patterns: all, unit, integration, framework, help, errors, send, read, search, groups, drafts, quick"
            exit 1
            ;;
    esac
}

# Run help tests
run_help_tests() {
    echo -e "${BLUE}❓ Running help command tests...${NC}"
    
    run_enhanced_test "Main help" "./mailos --help" "help"
    run_enhanced_test "Version" "./mailos --version" "help"
    run_enhanced_test "Commands list" "./mailos commands" "help"
    
    # Framework help tests
    run_enhanced_command_tests "send" "help"
    run_enhanced_command_tests "read" "help"
    run_enhanced_command_tests "groups" "help"
}

# Run error tests
run_error_tests() {
    echo -e "${RED}🚨 Running error handling tests...${NC}"
    
    run_enhanced_test "Invalid command" "./mailos invalid-command" "errors"
    run_enhanced_test "Invalid flag combination" "./mailos --invalid-flag" "errors"
    
    # Framework error tests
    run_enhanced_command_tests "send" "errors"
    run_enhanced_command_tests "read" "errors"
    run_enhanced_command_tests "groups" "errors"
}

# Run command tests
run_command_tests() {
    echo -e "${BLUE}⚙️  Running command functionality tests...${NC}"
    
    # Configuration commands
    run_enhanced_test "Setup help" "./mailos setup --help" "config"
    run_enhanced_test "Config help" "./mailos config --help" "config"
    
    # Account management
    run_enhanced_test "Accounts help" "./mailos accounts --help" "accounts"
    run_enhanced_test "List accounts" "./mailos accounts --list" "accounts"
    
    # Stats and reporting
    run_enhanced_test "Stats help" "./mailos stats --help" "stats"
    run_enhanced_test "Report help" "./mailos report --help" "reports"
    
    # Templates
    run_enhanced_test "Template help" "./mailos template --help" "templates"
}

# Run quick validation tests
run_quick_tests() {
    echo -e "${BLUE}⚡ Running quick validation tests...${NC}"
    
    run_enhanced_test "Help command" "./mailos --help" "quick"
    run_enhanced_test "Version command" "./mailos --version" "quick"
    run_enhanced_test "Setup help" "./mailos setup --help" "quick"
    run_enhanced_test "Send help" "./mailos send --help" "quick"
    run_enhanced_test "Read help" "./mailos read --help" "quick"
    run_enhanced_test "Search help" "./mailos search --help" "quick"
}

# Show framework summary
run_framework_summary() {
    echo -e "${BLUE}📊 Test Framework Summary${NC}"
    if [ -f "test/test_framework/main.go" ]; then
        go run test/test_framework/main.go summary
    else
        echo -e "${YELLOW}⚠️  Test framework not found${NC}"
    fi
}

# Generate enhanced coverage report
generate_coverage_report() {
    if [ "$COVERAGE" = true ] && [ -f "test/coverage/coverage.out" ]; then
        echo -e "${BLUE}📊 Generating coverage report...${NC}"
        
        # Generate HTML coverage report
        go tool cover -html=test/coverage/coverage.out -o test/coverage/coverage.html
        
        # Generate text coverage summary
        local coverage_percent
        coverage_percent=$(go tool cover -func=test/coverage/coverage.out | tail -1 | awk '{print $3}')
        
        echo -e "${GREEN}✓ Coverage report generated${NC}"
        echo -e "${CYAN}📈 Coverage: $coverage_percent${NC}"
        echo -e "${CYAN}📄 HTML Report: test/coverage/coverage.html${NC}"
        
        # Generate coverage badge data
        echo "{\"coverage\": \"$coverage_percent\"}" > test/coverage/coverage.json
    fi
}

# Enhanced test summary with more details
print_enhanced_summary() {
    local end_time=$(date +%s)
    local duration=$((end_time - START_TIME))
    
    echo ""
    echo -e "${BLUE}=================================================${NC}"
    echo -e "${BLUE}📊 ENHANCED TEST RESULTS SUMMARY${NC}"
    echo -e "${BLUE}=================================================${NC}"
    echo -e "Test Pattern: ${CYAN}$TEST_PATTERN${NC}"
    echo -e "Total tests: ${CYAN}$TOTAL_TESTS${NC}"
    echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
    echo -e "Skipped: ${YELLOW}$SKIPPED_TESTS${NC}"
    echo -e "Duration: ${PURPLE}${duration}s${NC}"
    
    if [ $TOTAL_TESTS -gt 0 ]; then
        local pass_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
        echo -e "Pass Rate: ${CYAN}${pass_rate}%${NC}"
    fi
    
    if [ "$COVERAGE" = true ] && [ -f "test/coverage/coverage.out" ]; then
        local coverage_percent
        coverage_percent=$(go tool cover -func=test/coverage/coverage.out | tail -1 | awk '{print $3}' 2>/dev/null || echo "N/A")
        echo -e "Coverage: ${PURPLE}$coverage_percent${NC}"
    fi
    
    echo ""
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}🎉 All tests passed!${NC}"
        if [ -f "test/logs/test_results.csv" ]; then
            echo -e "${CYAN}📊 Detailed results: test/logs/test_results.csv${NC}"
        fi
    else
        echo -e "${RED}❌ Some tests failed${NC}"
        echo -e "${YELLOW}💡 Check test/logs/test_results.csv for details${NC}"
    fi
    
    # Show quick commands for common actions
    echo ""
    echo -e "${BLUE}💡 Quick Actions:${NC}"
    echo -e "  ${CYAN}task test-enhanced -v${NC}     # Verbose mode"
    echo -e "  ${CYAN}task test-enhanced -c${NC}     # With coverage"
    echo -e "  ${CYAN}task test-enhanced -w${NC}     # Watch mode"
    echo -e "  ${CYAN}task test-enhanced quick${NC}  # Quick validation"
}

# Watch mode implementation
run_watch_mode() {
    echo -e "${YELLOW}👀 Entering watch mode...${NC}"
    echo -e "${YELLOW}Watching for changes in Go files and test files...${NC}"
    echo -e "${YELLOW}Press Ctrl+C to exit${NC}"
    echo ""
    
    # Initial test run
    run_tests_once
    
    # Watch for file changes
    if command -v inotifywait >/dev/null 2>&1; then
        # Linux
        while inotifywait -e modify -r . --include='\.go$|\.sh$' >/dev/null 2>&1; do
            echo -e "${YELLOW}🔄 Files changed, re-running tests...${NC}"
            echo ""
            run_tests_once
        done
    elif command -v fswatch >/dev/null 2>&1; then
        # macOS
        fswatch -o . --include='\.go$|\.sh$' | while read num; do
            echo -e "${YELLOW}🔄 Files changed, re-running tests...${NC}"
            echo ""
            run_tests_once
        done
    else
        echo -e "${RED}❌ No file watching utility found (inotifywait or fswatch)${NC}"
        echo -e "${YELLOW}Please install inotify-tools (Linux) or fswatch (macOS)${NC}"
        exit 1
    fi
}

# Run tests once
run_tests_once() {
    # Reset statistics
    TOTAL_TESTS=0
    PASSED_TESTS=0
    FAILED_TESTS=0
    SKIPPED_TESTS=0
    START_TIME=$(date +%s)
    
    # Initialize CSV log
    echo "timestamp,test_name,category,exit_code,duration" > test/logs/test_results.csv
    
    setup_test_environment
    build_test_binaries
    
    # Run the specified test pattern
    run_test_pattern "$TEST_PATTERN"
    
    generate_coverage_report
    print_enhanced_summary
    
    return $FAILED_TESTS
}

# Cleanup function
cleanup() {
    echo -e "${BLUE}🧹 Cleaning up...${NC}"
    rm -rf test/temp
}

# Main execution
main() {
    parse_args "$@"
    
    # Set up cleanup trap
    trap cleanup EXIT
    
    echo -e "${PURPLE}🚀 Enhanced EmailOS Test Runner${NC}"
    echo -e "${PURPLE}=================================${NC}"
    
    if [ "$WATCH_MODE" = true ]; then
        run_watch_mode
    else
        run_tests_once
        exit $?
    fi
}

# Run main function with all arguments
main "$@"