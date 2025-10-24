# EmailOS Testing Framework

A comprehensive, modern testing framework for the EmailOS CLI project, inspired by Jest and pytest patterns while maintaining Go best practices.

## Quick Start

```bash
# Run all tests
./test/run_tests.sh

# Run with coverage
./test/run_tests.sh -c

# Watch mode (re-run on file changes)
./test/run_tests.sh -w

# Run specific test categories
./test/run_tests.sh unit
./test/run_tests.sh integration
./test/run_tests.sh e2e

# Run with filter
./test/run_tests.sh -f "Email"

# Verbose output with parallel execution
./test/run_tests.sh -v -p
```

## Test Structure

```
test/
├── unit/                           # Unit tests (individual functions/components)
│   ├── email_parsing_test.go      # Email structure and parsing
│   ├── config_test.go             # Configuration management
│   └── cli_commands_test.go       # Command parsing and validation
├── integration/                   # Integration tests (component interactions)
│   ├── email_flow_test.go         # End-to-end email workflows
│   └── imap_integration_test.go   # IMAP server integration
├── e2e/                           # End-to-end tests (full user workflows)
│   ├── full_workflow_test.go      # Complete user scenarios
│   └── cross_platform_test.go    # Platform compatibility
├── mocks/                         # Mock implementations
│   ├── imap_mock.go              # IMAP server mock
│   └── smtp_mock.go              # SMTP server mock
├── helpers/                       # Test utilities and helpers
│   ├── test_setup.go             # Common setup functions
│   ├── assertions.go             # Custom assertion helpers
│   └── data_builders.go          # Test data builders
├── fixtures/                      # Test data and fixtures
│   ├── emails/                    # Sample email data
│   ├── configs/                   # Test configurations
│   └── attachments/               # Test attachment files
├── test_framework/                # Original command-line test framework
│   ├── main.go                   # Test case definitions
│   └── README.md                 # Framework documentation
└── scripts/                      # Test execution scripts
    ├── run_tests.sh              # Enhanced test runner
    └── setup_test_env.sh         # Environment setup
```

## Features

### Jest/Pytest-like Experience

- **Watch Mode**: Automatically re-run tests when files change
- **Test Filtering**: Run specific tests by name or pattern
- **Parallel Execution**: Run tests concurrently for faster feedback
- **Coverage Reports**: Generate HTML and text coverage reports
- **Descriptive Output**: Clear, colorful test results
- **Test Discovery**: Automatic test file discovery and execution

### Modern Testing Patterns

- **Test Organization**: Structured test suites with describe/it-like patterns
- **Setup/Teardown**: Comprehensive test lifecycle management
- **Mocking Infrastructure**: Mock external dependencies (IMAP, SMTP)
- **Parameterized Tests**: Data-driven test cases
- **Custom Assertions**: Domain-specific assertion helpers

### Enhanced Reliability

- **Test Isolation**: Each test runs in a clean environment
- **Deterministic Tests**: Reproducible test results
- **Mock Dependencies**: No external service requirements
- **Timeout Handling**: Prevents hanging tests
- **Error Recovery**: Graceful handling of test failures

## Test Runner Options

```bash
# Basic Usage
./test/run_tests.sh [OPTIONS] [TEST_PATTERN]

# Options
-w, --watch         Watch mode - re-run tests on file changes
-c, --coverage      Generate test coverage report
-v, --verbose       Verbose output
-p, --parallel      Run tests in parallel
-f, --filter FILTER Filter tests by name/pattern
-t, --timeout TIME  Test timeout (default: 30s)
-o, --output FORMAT Output format (pretty, json, tap)
-h, --help          Show help message

# Test Patterns
unit                Run only unit tests
integration         Run only integration tests
e2e                 Run only end-to-end tests
mocks               Run mock-related tests
framework           Run command-line framework tests
./path/to/test      Run specific test file
TestName            Run specific test function
```

## Writing Tests

### Unit Tests

```go
package unit

import (
    "testing"
    "../helpers"
    mailos "github.com/anduimagui/emailos"
)

func TestEmailValidation(t *testing.T) {
    t.Run("should validate email addresses", func(t *testing.T) {
        testClient, cleanup := helpers.SetupTest(t)
        defer cleanup()
        
        email := &mailos.Email{
            ID:      1,
            From:    "test@example.com",
            To:      []string{"recipient@example.com"},
            Subject: "Test Subject",
            Body:    "Test body content",
            Date:    time.Now(),
        }
        
        helpers.AssertEmailStructure(t, email)
        helpers.AssertValidEmail(t, email.From)
    })
}
```

### Integration Tests with Mocks

```go
func TestEmailFetching(t *testing.T) {
    t.Run("should fetch emails from IMAP server", func(t *testing.T) {
        testClient, cleanup := helpers.SetupTest(t)
        defer cleanup()
        
        // Setup mock IMAP server
        mockServer := mocks.NewMockIMAPServer()
        testMessages := mocks.CreateTestMessages()
        mockServer.WithMessages(testMessages)
        
        // Test email fetching
        messages, err := mockServer.FetchMessages("INBOX", 10)
        helpers.AssertNoError(t, err)
        helpers.AssertLen(t, messages, 4)
    })
}
```

### Using Test Helpers

```go
func TestAttachments(t *testing.T) {
    testClient, cleanup := helpers.SetupTest(t)
    defer cleanup()
    
    // Create test attachment
    attachmentPath, err := testClient.CreateTestAttachment("test.pdf", 1024)
    helpers.AssertNoError(t, err)
    
    // Test email with attachment
    email := &mailos.Email{
        Attachments: []string{attachmentPath},
        AttachmentData: map[string][]byte{
            "test.pdf": []byte("PDF content"),
        },
    }
    
    helpers.AssertAttachmentsValid(t, email)
}
```

## Test Categories

### Unit Tests (`test/unit/`)
- Individual function testing
- Pure logic validation
- No external dependencies
- Fast execution (< 1s per test)

### Integration Tests (`test/integration/`)
- Component interaction testing
- Mock external services
- Database operations
- Moderate execution time (< 5s per test)

### End-to-End Tests (`test/e2e/`)
- Full workflow testing
- Real or containerized services
- User scenario validation
- Longer execution time (< 30s per test)

### Framework Tests (`test/test_framework/`)
- Command-line interface testing
- Help text validation
- Error message verification
- Flag combination testing

## Mock Infrastructure

### IMAP Server Mock

```go
mockServer := mocks.NewMockIMAPServer()
mockServer.WithMessages(testMessages)
mockServer.WithBehavior(mocks.MockBehavior{
    ShouldFailConnection: false,
    ConnectionDelay:      100 * time.Millisecond,
})

err := mockServer.Connect()
messages, err := mockServer.FetchMessages("INBOX", 10)
```

### Test Data Builders

```go
email := helpers.NewEmailBuilder().
    WithFrom("test@example.com").
    WithTo("recipient@example.com").
    WithSubject("Test Email").
    WithAttachment("document.pdf", []byte("content")).
    Build()
```

## Coverage Reports

Generate coverage reports with:

```bash
./test/run_tests.sh -c
```

This creates:
- `test/coverage/coverage.out` - Coverage data
- `test/coverage/coverage.html` - HTML report
- Console output with coverage percentage

## Continuous Integration

The test framework integrates with CI/CD pipelines:

```yaml
# GitHub Actions example
- name: Run Tests
  run: |
    ./test/run_tests.sh -c -p --output json > test_results.json
    
- name: Upload Coverage
  uses: codecov/codecov-action@v1
  with:
    file: test/coverage/coverage.out
```

## Best Practices

### Test Organization
- Group related tests using `t.Run()`
- Use descriptive test names
- Follow AAA pattern (Arrange, Act, Assert)
- Keep tests focused and simple

### Test Data
- Use builders for complex objects
- Create minimal test data
- Clean up test artifacts
- Use fixtures for common scenarios

### Assertions
- Use domain-specific assertions
- Provide clear error messages
- Test both positive and negative cases
- Validate complete object state

### Performance
- Run fast tests first
- Use parallel execution for independent tests
- Mock external dependencies
- Set appropriate timeouts

## Troubleshooting

### Common Issues

**Tests hanging:**
- Check for infinite loops
- Verify timeout settings
- Review mock configurations

**Coverage not generated:**
- Ensure `-c` flag is used
- Check write permissions for coverage directory
- Verify Go coverage tools are installed

**Watch mode not working:**
- Install `inotify-tools` (Linux) or `fswatch` (macOS)
- Check file permissions
- Verify file patterns

### Debug Mode

Run tests with verbose output:

```bash
./test/run_tests.sh -v -f "TestName"
```

## Migration from Old Structure

The enhanced testing framework maintains compatibility with existing tests while providing modern features:

1. **Existing unit tests** in `test/unit_tests/` continue to work
2. **Framework tests** in `test/test_framework/` are integrated
3. **Shell scripts** in `scripts/test-*.sh` are enhanced
4. **New structure** provides better organization and tooling

## Contributing

When adding new tests:

1. Choose appropriate test category (unit/integration/e2e)
2. Use existing helpers and assertions
3. Follow naming conventions
4. Add documentation for complex test scenarios
5. Ensure tests are deterministic and isolated