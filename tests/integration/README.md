# Integration Tests

This directory contains integration tests that test complete workflows and component interactions.

## Structure

- `shell/` - End-to-end shell functionality tests
- `commands/` - Command execution integration tests
- `config/` - Configuration loading and validation tests
- `security/` - Security validation integration tests

## Running Tests

```bash
# Run all integration tests
go test ./tests/integration/...

# Run with test environment setup
TEST_ENV=integration go test ./tests/integration/...
```

## Test Guidelines

- Test complete user workflows
- Use real components where possible
- Mock external services and file systems
- Test cross-platform compatibility
- Include performance regression tests