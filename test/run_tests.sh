#!/bin/bash

# Enhanced EmailOS Test Runner
# Provides Jest/pytest-like testing experience for Go projects

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Test configuration
WATCH_MODE=false
COVERAGE=false
VERBOSE=false
PARALLEL=false
FILTER=""
TIMEOUT="30s"
OUTPUT_FORMAT="pretty"
TEST_PATTERN="."

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
    -h, --help          Show this help message

TEST PATTERNS:
    unit                Run only unit tests
    integration         Run only integration tests
    e2e                 Run only end-to-end tests
    mocks               Run mock-related tests
    ./path/to/test      Run specific test file
    TestName            Run specific test function

EXAMPLES:
    $0                  Run all tests
    $0 -c unit          Run unit tests with coverage
    $0 -w -v            Watch mode with verbose output
    $0 -f "Email"       Run tests matching "Email"
    $0 --parallel integration  Run integration tests in parallel

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

# Setup test environment
setup_test_environment() {
    echo -e "${BLUE}🔧 Setting up test environment...${NC}"
    
    # Create necessary directories
    mkdir -p test/reports
    mkdir -p test/coverage
    mkdir -p test/temp
    
    # Source environment variables if available
    if [ -f ".env" ]; then
        source .env
        echo -e "${GREEN}✓ Loaded .env file${NC}"
    fi
    
    # Set test-specific environment variables
    export MAILOS_TEST_MODE=true
    export MAILOS_CONFIG_DIR="test/temp"
    export GO_TEST_TIMEOUT="$TIMEOUT"
    
    echo -e "${GREEN}✓ Test environment ready${NC}"
}

# Build test binaries
build_test_binaries() {
    echo -e "${BLUE}🔨 Building test binaries...${NC}"
    
    if ! go build -o test/temp/mailos-test ./cmd/mailos/; then
        echo -e "${RED}❌ Failed to build test binary${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✓ Test binaries built${NC}"
}

# Discover test files
discover_tests() {
    local pattern="$1"
    local test_files=()
    
    case "$pattern" in
        "unit")
            test_files=($(find test/unit -name "*_test.go" 2>/dev/null || true))
            ;;
        "integration")
            test_files=($(find test/integration -name "*_test.go" 2>/dev/null || true))
            ;;
        "e2e")
            test_files=($(find test/e2e -name "*_test.go" 2>/dev/null || true))
            ;;
        "mocks")
            test_files=($(find test/mocks -name "*_test.go" 2>/dev/null || true))
            ;;
        "framework")
            test_files=("test/test_framework/")
            ;;
        *)
            # Default: find all test files
            test_files=($(find test -name "*_test.go" 2>/dev/null || true))
            # Also include the current directory for any *_test.go files
            local root_tests=($(find . -maxdepth 1 -name "*_test.go" 2>/dev/null || true))
            test_files=("${test_files[@]}" "${root_tests[@]}")
            ;;
    esac
    
    echo "${test_files[@]}"
}

# Run Go tests with enhanced options
run_go_tests() {
    local test_files=("$@")
    local go_test_args=()
    
    # Build Go test arguments
    if [ "$VERBOSE" = true ]; then
        go_test_args+=("-v")
    fi
    
    if [ "$PARALLEL" = true ]; then
        go_test_args+=("-parallel" "4")
    fi
    
    if [ "$COVERAGE" = true ]; then
        go_test_args+=("-coverprofile=test/coverage/coverage.out")
        go_test_args+=("-covermode=atomic")
    fi
    
    if [ -n "$FILTER" ]; then
        go_test_args+=("-run" "$FILTER")
    fi
    
    go_test_args+=("-timeout" "$TIMEOUT")
    
    # Add test files or use default
    if [ ${#test_files[@]} -gt 0 ]; then
        go_test_args+=("${test_files[@]}")
    else
        go_test_args+=("./...")
    fi
    
    echo -e "${BLUE}🧪 Running Go tests...${NC}"
    echo -e "${PURPLE}Command: go test ${go_test_args[*]}${NC}"
    
    # Run tests and capture output
    local test_output
    local test_result=0
    
    if [ "$OUTPUT_FORMAT" = "json" ]; then
        go_test_args+=("-json")
    fi
    
    test_output=$(go test "${go_test_args[@]}" 2>&1) || test_result=$?
    
    # Parse test results
    parse_test_results "$test_output" "$test_result"
    
    return $test_result
}

# Parse test results and update statistics
parse_test_results() {
    local output="$1"
    local result="$2"
    
    echo "$output"
    
    # Extract test statistics from output
    local pass_count=$(echo "$output" | grep -c "PASS:" || echo "0")
    local fail_count=$(echo "$output" | grep -c "FAIL:" || echo "0")
    local skip_count=$(echo "$output" | grep -c "SKIP:" || echo "0")
    
    PASSED_TESTS=$((PASSED_TESTS + pass_count))
    FAILED_TESTS=$((FAILED_TESTS + fail_count))
    SKIPPED_TESTS=$((SKIPPED_TESTS + skip_count))
    TOTAL_TESTS=$((TOTAL_TESTS + pass_count + fail_count + skip_count))
}

# Run framework tests
run_framework_tests() {
    echo -e "${BLUE}🧪 Running test framework tests...${NC}"
    
    if [ ! -f "test/test_framework/main.go" ]; then
        echo -e "${YELLOW}⚠️  Test framework not found, skipping${NC}"
        return 0
    fi
    
    # Run framework summary
    echo -e "${CYAN}📋 Test Framework Summary:${NC}"
    go run test/test_framework/main.go summary
    
    # Run help tests
    echo -e "${CYAN}📋 Running help tests:${NC}"
    local help_output
    help_output=$(go run test/test_framework/main.go help 2>/dev/null || true)
    
    if [ -n "$help_output" ]; then
        echo "$help_output" | while IFS='|' read -r test_name test_command; do
            if [ -n "$test_name" ] && [ -n "$test_command" ]; then
                run_single_framework_test "$test_name" "$test_command"
            fi
        done
    fi
    
    # Run error tests
    echo -e "${CYAN}📋 Running error tests:${NC}"
    local error_output
    error_output=$(go run test/test_framework/main.go errors 2>/dev/null || true)
    
    if [ -n "$error_output" ]; then
        echo "$error_output" | while IFS='|' read -r test_name test_command; do
            if [ -n "$test_name" ] && [ -n "$test_command" ]; then
                run_single_framework_test "$test_name" "$test_command"
            fi
        done
    fi
}

# Run a single framework test
run_single_framework_test() {
    local test_name="$1"
    local test_command="$2"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -n "  Testing: $test_name... "
    
    if timeout 30 bash -c "$test_command" >/dev/null 2>&1; then
        echo -e "${GREEN}✓ PASS${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        local exit_code=$?
        if [ $exit_code -eq 124 ]; then
            echo -e "${YELLOW}⏱ TIMEOUT${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        else
            echo -e "${RED}✗ FAIL${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    fi
}

# Generate test coverage report
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
    fi
}

# Print test summary
print_test_summary() {
    local end_time=$(date +%s)
    local duration=$((end_time - START_TIME))
    
    echo ""
    echo -e "${BLUE}=================================================${NC}"
    echo -e "${BLUE}📊 TEST RESULTS SUMMARY${NC}"
    echo -e "${BLUE}=================================================${NC}"
    echo -e "Total tests: ${CYAN}$TOTAL_TESTS${NC}"
    echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
    echo -e "Skipped: ${YELLOW}$SKIPPED_TESTS${NC}"
    echo -e "Duration: ${PURPLE}${duration}s${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}🎉 All tests passed!${NC}"
    else
        echo -e "${RED}❌ Some tests failed${NC}"
    fi
}

# Watch mode implementation
run_watch_mode() {
    echo -e "${YELLOW}👀 Entering watch mode...${NC}"
    echo -e "${YELLOW}Watching for changes in Go files...${NC}"
    echo -e "${YELLOW}Press Ctrl+C to exit${NC}"
    echo ""
    
    # Initial test run
    run_tests_once
    
    # Watch for file changes
    if command -v inotifywait >/dev/null 2>&1; then
        # Linux
        while inotifywait -e modify -r . --include='\.go$' >/dev/null 2>&1; do
            echo -e "${YELLOW}🔄 Files changed, re-running tests...${NC}"
            run_tests_once
        done
    elif command -v fswatch >/dev/null 2>&1; then
        # macOS
        fswatch -o . --include='\.go$' | while read num; do
            echo -e "${YELLOW}🔄 Files changed, re-running tests...${NC}"
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
    
    setup_test_environment
    build_test_binaries
    
    local test_files
    test_files=($(discover_tests "$TEST_PATTERN"))
    
    # Run Go tests
    local go_result=0
    if [ ${#test_files[@]} -gt 0 ]; then
        run_go_tests "${test_files[@]}" || go_result=$?
    fi
    
    # Run framework tests
    if [ "$TEST_PATTERN" = "." ] || [ "$TEST_PATTERN" = "framework" ]; then
        run_framework_tests
    fi
    
    generate_coverage_report
    print_test_summary
    
    return $go_result
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
    
    echo -e "${PURPLE}🚀 EmailOS Enhanced Test Runner${NC}"
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