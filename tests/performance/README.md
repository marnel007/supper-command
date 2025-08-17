# Performance Tests

This directory contains performance benchmarks and load tests.

## Test Categories

- `benchmarks/` - Go benchmark tests
- `memory/` - Memory usage tests
- `concurrency/` - Concurrent operation tests
- `startup/` - Application startup time tests

## Running Tests

```bash
# Run all benchmarks
go test -bench=. ./tests/performance/...

# Run with memory profiling
go test -bench=. -memprofile=mem.prof ./tests/performance/...

# Run with CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./tests/performance/...
```

## Performance Targets

- Startup time: < 2 seconds
- Command execution: < 100ms for basic commands
- Memory usage: < 50MB baseline
- Tab completion: < 50ms response time