# Unit Tests

This directory contains unit tests for individual components and functions.

## Structure

- `app/` - Tests for application orchestration
- `shell/` - Tests for shell components
- `commands/` - Tests for command implementations
- `security/` - Tests for security and validation
- `config/` - Tests for configuration system
- `monitoring/` - Tests for monitoring and logging

## Running Tests

```bash
# Run all unit tests
go test ./tests/unit/...

# Run tests with coverage
go test -cover ./tests/unit/...

# Run tests with verbose output
go test -v ./tests/unit/...
```

## Test Guidelines

- Each test file should end with `_test.go`
- Use table-driven tests where appropriate
- Mock external dependencies
- Aim for 80%+ code coverage
- Test both success and error cases