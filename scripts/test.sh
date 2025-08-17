#!/bin/bash

# Test runner script for SuperShell

set -e

echo "🧪 Running SuperShell Tests"
echo "=========================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Run unit tests
echo ""
echo "📋 Running Unit Tests..."
if go test -v ./tests/unit/...; then
    print_status "Unit tests passed"
else
    print_error "Unit tests failed"
    exit 1
fi

# Run tests with coverage
echo ""
echo "📊 Running Tests with Coverage..."
if go test -cover ./tests/unit/...; then
    print_status "Coverage tests completed"
else
    print_warning "Coverage tests had issues"
fi

# Generate detailed coverage report
echo ""
echo "📈 Generating Coverage Report..."
go test -coverprofile=coverage.out ./tests/unit/...
go tool cover -html=coverage.out -o coverage.html
print_status "Coverage report generated: coverage.html"

# Run integration tests if they exist
if [ -d "tests/integration" ] && [ "$(ls -A tests/integration)" ]; then
    echo ""
    echo "🔗 Running Integration Tests..."
    if go test -v ./tests/integration/...; then
        print_status "Integration tests passed"
    else
        print_error "Integration tests failed"
        exit 1
    fi
fi

# Run security tests if they exist
if [ -d "tests/security" ] && [ "$(ls -A tests/security)" ]; then
    echo ""
    echo "🔒 Running Security Tests..."
    if SECURITY_TESTS=enabled go test -v ./tests/security/...; then
        print_status "Security tests passed"
    else
        print_error "Security tests failed"
        exit 1
    fi
fi

# Run performance benchmarks if they exist
if [ -d "tests/performance" ] && [ "$(ls -A tests/performance)" ]; then
    echo ""
    echo "⚡ Running Performance Benchmarks..."
    if go test -bench=. ./tests/performance/...; then
        print_status "Performance benchmarks completed"
    else
        print_warning "Performance benchmarks had issues"
    fi
fi

echo ""
print_status "All tests completed successfully!"
echo ""
echo "📋 Test Summary:"
echo "  - Unit tests: ✅ Passed"
echo "  - Coverage report: coverage.html"
if [ -d "tests/integration" ] && [ "$(ls -A tests/integration)" ]; then
    echo "  - Integration tests: ✅ Passed"
fi
if [ -d "tests/security" ] && [ "$(ls -A tests/security)" ]; then
    echo "  - Security tests: ✅ Passed"
fi
if [ -d "tests/performance" ] && [ "$(ls -A tests/performance)" ]; then
    echo "  - Performance benchmarks: ✅ Completed"
fi