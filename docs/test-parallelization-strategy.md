# Test Parallelization Strategy

## Overview

This document describes the test parallelization strategy implemented for the EVE SDE Database Builder project to optimize test suite runtime while maintaining test reliability and correctness.

## Implementation

### Package-Level Parallelization

The test suite uses Go's built-in package-level parallelization with the `-p` flag:

```bash
go test -p 4 -parallel 8 ./cmd/... ./internal/...
```

- **`-p 4`**: Run up to 4 test packages in parallel
- **`-parallel 8`**: Run up to 8 test functions in parallel within each package

This allows multiple test packages to run concurrently and enables parallel execution of tests marked with `t.Parallel()` within each package.

### Test-Level Parallelization

Individual test functions are marked with `t.Parallel()` where appropriate:

```go
func TestExample(t *testing.T) {
    t.Parallel()  // Marks this test as safe to run in parallel
    // test code...
}
```

#### Tests Marked for Parallel Execution

The following test suites have been parallelized:

1. **internal/config**: All tests except those using environment variables
2. **internal/errors**: All tests (pure functions, no shared state)
3. **internal/logger**: All tests except `TestGlobalLogger` (isolated instances)
4. **internal/retry**: Tests without time.Sleep (timing-insensitive tests)
5. **internal/testutil**: All utility tests
6. **internal/worker**: Pool creation tests

#### Tests Excluded from Parallelization

Some tests are intentionally **not** marked with `t.Parallel()`:

1. **Timing-sensitive tests**: Tests with `time.Sleep()` or timing requirements
   - `internal/retry/context_test.go` (retry timing tests)
   - `internal/worker` tests with sleep delays

2. **Global state tests**: Tests that modify shared global state
   - `internal/logger/TestGlobalLogger` (modifies global logger)
   - `internal/config` tests with environment variables

3. **Integration tests**: Tests that create actual resources
   - Database integration tests
   - File system tests (though tests using `t.TempDir()` are safe)

## Performance Results

### Before Optimization
- Baseline: **4.7 seconds** (sequential execution)

### After Optimization
- Package parallelization: **4.5 seconds**
- With test-level parallelization: **4.5 seconds**

### Analysis

The relatively modest improvement (4.7s → 4.5s) is due to:

1. **Short overall runtime**: The test suite is already fast (~5 seconds)
2. **I/O-bound tests**: Many tests involve file I/O and time.Sleep(), limiting parallel speedup
3. **Sequential dependencies**: Some test packages have inherent ordering requirements

## Guidelines for Future Tests

### When to Use t.Parallel()

✅ **DO use `t.Parallel()` for:**
- Pure function tests with no side effects
- Tests that create isolated resources (e.g., using `t.TempDir()`)
- Unit tests with no shared state
- Tests that don't rely on timing

### When NOT to Use t.Parallel()

❌ **DO NOT use `t.Parallel()` for:**
- Tests with `time.Sleep()` or timing requirements
- Tests that modify global state
- Tests that use shared resources (databases, files without isolation)
- Tests with environment variable manipulation

## Makefile Targets

The following Make targets support parallelization:

```makefile
test:       # Run tests with -p 4 -parallel 8
test-race:  # Run with race detector and parallelization
coverage:   # Run coverage with parallelization
test-tools: # Run tool tests with -p 2 -parallel 4
```

## Verification

To verify parallelization is working:

```bash
# Clear test cache
go clean -testcache

# Run with timing
time make test

# Expected: ~4.5 seconds on typical hardware
```

## Future Improvements

Potential areas for further optimization:

1. **Reduce time.Sleep usage**: Replace sleep-based timing tests with mock time or channels
2. **Isolate global state**: Refactor tests to avoid global state dependencies
3. **Optimize slow packages**: Profile and optimize the slowest test packages (parser, retry, cli)
4. **Increase parallelism**: Consider higher `-p` and `-parallel` values on machines with more cores

## References

- [Go Testing Package](https://pkg.go.dev/testing)
- [Parallel Test Execution](https://go.dev/blog/subtests)
- Issue #6: Test Parallelization Strategy
