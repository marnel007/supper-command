# Security Tests

This directory contains security-focused tests to verify protection against various attack vectors.

## Test Categories

- `injection/` - Command injection attack tests
- `validation/` - Input validation tests
- `privilege/` - Privilege escalation tests
- `fuzzing/` - Fuzzing tests for input validation

## Running Tests

```bash
# Run all security tests
go test ./tests/security/...

# Run with security test environment
SECURITY_TESTS=enabled go test ./tests/security/...
```

## Test Guidelines

- Test all known attack vectors
- Include malformed and malicious inputs
- Verify proper error handling
- Test privilege boundary enforcement
- Document security assumptions