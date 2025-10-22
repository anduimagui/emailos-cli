# Mailos Test Framework

A comprehensive, reusable test framework for all mailos CLI commands.

## Overview

This framework provides structured testing for all mailos commands with support for:
- Multiple command types (send, read, search, etc.)
- Test categories (help, errors, basic, advanced, etc.)
- Environment requirements filtering
- Shell script integration

## Structure

### Test Case Definition
```go
type CommandTest struct {
    Name         string // Test name for display
    Command      string // Actual command to run
    Description  string // Description of what the test does
    RequiresEnv  bool   // Whether test needs environment variables
    Category     string // Test category (help, errors, basic, etc.)
    Command_Type string // Command type (send, read, search, etc.)
}
```

### Available Command Types
- `send` - Email sending functionality
- `read` - Email reading by ID

### Available Categories
- `help` - Help command tests
- `errors` - Error handling tests  
- `basic` - Basic functionality tests
- `shortflags` - Short flag tests (-t, -s, -b, etc.)
- `recipients` - Recipient handling (TO, CC, BCC)
- `format` - Format options (plain, HTML, templates)
- `signature` - Signature handling
- `attachments` - Attachment functionality
- `preview` - Preview and dry-run tests
- `debug` - Verbose/debugging tests
- `drafts` - Draft operations
- `documents` - Document parsing
- `combined` - Complex flag combinations
- `account` - Account-specific tests

## Usage

### Command Line Interface

```bash
# Show summary of all tests
go run test/test_framework/main.go summary

# List all tests for a command type
go run test/test_framework/main.go send

# List tests by category
go run test/test_framework/main.go help

# List tests by command type and category
go run test/test_framework/main.go send help

# List tests that don't require environment variables
go run test/test_framework/main.go no-env

# List error tests for send command
go run test/test_framework/main.go errors send
```

### Shell Script Integration

The framework is designed to integrate with `scripts/test-all-commands.sh`:

```bash
# Run help tests for send command
run_command_tests "send" "help"

# Run error tests for read command  
run_command_tests "read" "errors"

# Run basic functionality tests (requires environment)
run_command_tests "send" "basic"
```

## Adding New Tests

### 1. Add Test Cases

Add new test cases to the appropriate test suite in `main.go`:

```go
var YourCommandTestSuite = []CommandTest{
    {
        Name:         "Your test name",
        Command:      "./mailos yourcommand --flag value",
        Description:  "Description of what this tests",
        RequiresEnv:  true, // or false
        Category:     "basic", // or appropriate category
        Command_Type: "yourcommand",
    },
    // ... more tests
}
```

### 2. Register Test Suite

Add your test suite to the `AllTests` variable:

```go
func init() {
    AllTests = append(AllTests, SendTestSuite...)
    AllTests = append(AllTests, ReadTestSuite...)
    AllTests = append(AllTests, YourCommandTestSuite...) // Add this line
}
```

### 3. Update Test Script

Add calls to `run_command_tests` in `scripts/test-all-commands.sh`:

```bash
# In the appropriate section
run_command_tests "yourcommand" "help"
run_command_tests "yourcommand" "errors"
```

## Examples

### Send Command Tests
- **Help**: `./mailos send --help`
- **Basic**: `./mailos send --to $TO_EMAIL --subject 'Test' --body 'Test'`
- **Short flags**: `./mailos send -t $TO_EMAIL -s 'Test' -b 'Test'`
- **Errors**: `./mailos send --subject 'Test' --body 'Test'` (missing recipient)

### Read Command Tests  
- **Help**: `./mailos read --help`
- **Basic**: `./mailos read 1` or `./mailos read --id 1`
- **Documents**: `./mailos read 1 --include-documents`
- **Errors**: `./mailos read` (missing ID)

## Environment Variables

Tests marked with `RequiresEnv: true` expect these environment variables:
- `FROM_EMAIL` - Configured account email
- `TO_EMAIL` - Test recipient email

These are loaded from `.env` file by the test script.

## Categories Explanation

- **help**: Tests command help output (`--help`)
- **errors**: Tests error conditions and validation
- **basic**: Tests core functionality with minimal flags
- **shortflags**: Tests short flag variants (`-t`, `-s`, `-v`, etc.)
- **recipients**: Tests TO, CC, BCC handling
- **format**: Tests text formats (plain, HTML, markdown)
- **signature**: Tests signature handling
- **attachments**: Tests file attachments
- **preview**: Tests preview/dry-run modes
- **debug**: Tests verbose/debugging output
- **drafts**: Tests draft operations
- **documents**: Tests document parsing
- **combined**: Tests complex flag combinations
- **account**: Tests account-specific operations

## Best Practices

1. **Use descriptive test names** that clearly indicate what's being tested
2. **Group related tests** into appropriate categories
3. **Mark environment requirements** accurately with `RequiresEnv`
4. **Test both success and failure cases** 
5. **Include edge cases** (empty values, invalid inputs, etc.)
6. **Test flag combinations** in the "combined" category
7. **Document new categories** when adding them