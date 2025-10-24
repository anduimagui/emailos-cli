# EmailOS Testing Guide

A comprehensive guide to using the enhanced EmailOS testing framework with Jest/pytest-like features.

## Quick Start

```bash
# Copy environment template and configure
cp .env.example .env
# Edit .env with your test email addresses

# Run all tests
task test-enhanced

# Run with coverage
task test-coverage

# Watch mode (re-run on file changes)
task test-watch

# Run specific test categories
task test-unit
task test-integration
task test-send
task test-read
```

## Test Commands Overview

### Core Test Commands

| Command | Description | Example |
|---------|-------------|---------|
| `task test-enhanced` | Run complete test suite | `task test-enhanced` |
| `task test-watch` | Watch mode with auto-reload | `task test-watch` |
| `task test-coverage` | Generate coverage reports | `task test-coverage` |
| `task test-mock` | Use mocked dependencies | `task test-mock` |
| `task test-verbose` | Detailed test output | `task test-verbose` |

### Category-Specific Commands

| Command | Description | Example |
|---------|-------------|---------|
| `task test-unit` | Unit tests only | `task test-unit` |
| `task test-integration` | Integration tests | `task test-integration` |
| `task test-send` | Send command tests | `task test-send` |
| `task test-read` | Read command tests | `task test-read` |
| `task test-search` | Search command tests | `task test-search` |
| `task test-groups` | Groups functionality tests | `task test-groups` |

### Legacy Commands (Still Available)

| Command | Description | Example |
|---------|-------------|---------|
| `task test-all` | Original comprehensive test | `task test-all` |
| `task test-quick` | Quick validation tests | `task test-quick` |
| `task test-framework` | Test framework validation | `task test-framework` |

## Enhanced Features

### 🎯 Advanced Filtering

```bash
# Filter tests by name pattern
task test-enhanced -- -f "Email"

# Filter with specific command
task test-enhanced send -f "help"

# Run only error tests
task test-enhanced errors
```

### 📊 Coverage Reporting

```bash
# Generate HTML coverage report
task test-coverage

# View coverage in browser
open test/coverage/coverage.html

# Coverage data available at:
# - test/coverage/coverage.out (Go coverage format)
# - test/coverage/coverage.html (HTML report)
# - test/coverage/coverage.json (Badge data)
```

### 👀 Watch Mode

```bash
# Start watch mode
task test-watch

# Watch specific test category
task test-watch unit

# Watch with verbose output
task test-watch -- -v
```

### 🎭 Mock Mode

```bash
# Run without external dependencies
task test-mock

# Mock mode with coverage
task test-enhanced -- --mock --coverage

# Mock mode skips email server connections
# Perfect for CI/CD environments
```

## Test Categories

### Unit Tests (`unit`)
- **Purpose**: Test individual functions and components
- **Speed**: Fast (< 1s per test)
- **Dependencies**: None (fully isolated)
- **Location**: `test/unit/`

```bash
task test-unit
task test-enhanced unit -v
```

### Integration Tests (`integration`)
- **Purpose**: Test component interactions
- **Speed**: Moderate (< 5s per test)
- **Dependencies**: Mocked external services
- **Location**: `test/integration/`

```bash
task test-integration
task test-enhanced integration --coverage
```

### Command Tests (`send`, `read`, `search`, etc.)
- **Purpose**: Test CLI command functionality
- **Speed**: Fast (help/errors) to moderate (functionality)
- **Dependencies**: Built binary, optional email config
- **Location**: `test/test_framework/`

```bash
task test-send
task test-read
task test-search
```

### Framework Tests (`framework`)
- **Purpose**: Validate test framework itself
- **Speed**: Fast
- **Dependencies**: Test framework binary
- **Location**: `test/test_framework/`

```bash
task test-framework
task test-enhanced framework
```

## Environment Configuration

### Required Setup

1. **Copy Environment Template**
   ```bash
   cp .env.example .env
   ```

2. **Configure Test Emails**
   ```bash
   # Edit .env file
   FROM_EMAIL=your-configured-account@example.com
   TO_EMAIL=test-recipient@example.com
   ```

3. **Optional: Advanced Configuration**
   ```bash
   # Enable debug mode
   DEBUG=true
   
   # Set test timeout
   TEST_TIMEOUT=60s
   
   # Enable mock mode by default
   MAILOS_MOCK_MODE=true
   ```

### Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `FROM_EMAIL` | Configured account email | Required | `andrew@happysoft.dev` |
| `TO_EMAIL` | Test recipient email | Required | `test@example.com` |
| `MAILOS_TEST_MODE` | Enable test mode | `true` | `true` |
| `MAILOS_MOCK_MODE` | Use mocked dependencies | `false` | `true` |
| `TEST_TIMEOUT` | Test timeout duration | `30s` | `60s` |
| `DEBUG` | Enable debug output | `false` | `true` |

## Advanced Usage

### Custom Test Patterns

```bash
# Run specific test patterns
scripts/test-enhanced.sh help        # Help command tests
scripts/test-enhanced.sh errors      # Error handling tests
scripts/test-enhanced.sh quick       # Quick validation tests

# With advanced options
scripts/test-enhanced.sh --verbose --coverage unit
scripts/test-enhanced.sh --mock --filter "Email" integration
scripts/test-enhanced.sh --watch --timeout 60s send
```

### Parallel Execution

```bash
# Enable parallel test execution
task test-enhanced -- --parallel

# Parallel with coverage
task test-enhanced -- --parallel --coverage
```

### Custom Output Formats

```bash
# JSON output for CI/CD
task test-enhanced -- --output json

# TAP format
task test-enhanced -- --output tap

# Pretty format (default)
task test-enhanced -- --output pretty
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: 1.24
      
      - name: Install Task
        run: sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin
      
      - name: Run Tests
        run: |
          cp .env.example .env
          task test-enhanced -- --mock --coverage --output json
      
      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          file: test/coverage/coverage.out
```

### Docker Testing

```dockerfile
FROM golang:1.24-alpine
WORKDIR /app
COPY . .
RUN go install github.com/go-task/task/v3/cmd/task@latest
RUN cp .env.example .env
RUN task test-enhanced -- --mock --coverage
```

## Test Development

### Writing New Tests

1. **Unit Tests** (Go testing framework)
   ```go
   // test/unit/example_test.go
   func TestEmailValidation(t *testing.T) {
       t.Run("should validate email addresses", func(t *testing.T) {
           testClient, cleanup := helpers.SetupTest(t)
           defer cleanup()
           
           email := &mailos.Email{
               From: "test@example.com",
               To:   []string{"recipient@example.com"},
           }
           
           helpers.AssertEmailStructure(t, email)
       })
   }
   ```

2. **Framework Tests** (Command-line testing)
   ```go
   // test/test_framework/main.go
   var NewCommandTestSuite = []CommandTest{
       {
           Name:         "New command help",
           Command:      "./mailos new --help",
           Description:  "Test new command help output",
           RequiresEnv:  false,
           Category:     "help",
           Command_Type: "new",
       },
   }
   ```

3. **Integration Tests** (Component interaction)
   ```go
   // test/integration/workflow_test.go
   func TestEmailWorkflow(t *testing.T) {
       mockServer := mocks.NewMockIMAPServer()
       testMessages := mocks.CreateTestMessages()
       mockServer.WithMessages(testMessages)
       
       // Test complete workflow
       messages, err := mockServer.FetchMessages("INBOX", 10)
       helpers.AssertNoError(t, err)
       helpers.AssertLen(t, messages, 4)
   }
   ```

### Test Helpers

Use the provided test helpers for consistent testing:

```go
// Setup and cleanup
testClient, cleanup := helpers.SetupTest(t)
defer cleanup()

// Assertions
helpers.AssertEqual(t, expected, actual)
helpers.AssertNoError(t, err)
helpers.AssertEmailStructure(t, email)
helpers.AssertValidEmail(t, "test@example.com")

// Test data creation
email := helpers.CreateTestEmailData()[0]
attachment := testClient.CreateTestAttachment("test.pdf", 1024)
```

## Troubleshooting

### Common Issues

**Tests Hanging**
```bash
# Increase timeout
task test-enhanced -- --timeout 60s

# Use mock mode
task test-mock
```

**Coverage Not Generated**
```bash
# Ensure coverage flag is used
task test-coverage

# Check permissions
chmod -R 755 test/coverage/
```

**Watch Mode Not Working**
```bash
# Install file watching tools
# macOS
brew install fswatch

# Linux
sudo apt-get install inotify-tools
```

**Environment Issues**
```bash
# Check environment setup
cat .env

# Skip environment checks
task test-enhanced -- --skip-env

# Use mock mode
task test-enhanced -- --mock
```

### Debug Mode

```bash
# Enable verbose output
task test-verbose

# Enable debug in environment
echo "DEBUG=true" >> .env

# Check test logs
cat test/logs/test_results.csv
```

## Best Practices

### Test Organization
- Keep tests focused and isolated
- Use descriptive test names
- Group related tests with `t.Run()`
- Follow AAA pattern (Arrange, Act, Assert)

### Performance
- Use mock mode for CI/CD
- Run unit tests first (fastest)
- Use parallel execution for independent tests
- Set appropriate timeouts

### Maintenance
- Run `task test-coverage` regularly
- Keep coverage above 80%
- Update tests when adding new features
- Use watch mode during development

### CI/CD Integration
- Always use mock mode in CI
- Generate coverage reports
- Run all test categories
- Use appropriate timeouts for environment

## Migration from Legacy Tests

The enhanced testing framework maintains compatibility:

1. **Existing commands still work**
   - `task test-all` → `task test-enhanced`
   - `task test-quick` → `task test-enhanced quick`
   - `task test-groups` → `task test-groups` (enhanced)

2. **New features available**
   - Watch mode: `task test-watch`
   - Coverage: `task test-coverage`
   - Mocking: `task test-mock`
   - Filtering: `task test-enhanced -- -f "pattern"`

3. **Gradual adoption**
   - Use enhanced commands for new workflows
   - Keep legacy commands for existing scripts
   - Migrate CI/CD to enhanced framework when ready

## Performance Benchmarks

| Test Category | Count | Duration | Coverage |
|---------------|-------|----------|----------|
| Unit Tests | 25+ | <5s | 85%+ |
| Integration Tests | 15+ | <15s | 70%+ |
| Command Tests | 100+ | <30s | N/A |
| Framework Tests | 200+ | <45s | N/A |
| **Total** | **340+** | **<95s** | **80%+** |

## Support

For issues with the testing framework:

1. Check this documentation
2. Review test logs in `test/logs/`
3. Run with verbose mode: `task test-verbose`
4. Use mock mode to isolate issues: `task test-mock`
5. Check environment configuration in `.env`

For new feature requests or bugs, create an issue in the repository with:
- Test command used
- Expected vs actual behavior
- Environment details (OS, Go version)
- Log output with verbose mode enabled