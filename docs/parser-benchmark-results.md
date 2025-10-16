# Parser Performance Benchmark Results

## Overview

Performance benchmarks for JSONL parser implementation measuring throughput, memory allocation, and CPU usage across various file sizes.

**Target:** 100k lines in < 1s ✓ **ACHIEVED** (actual: ~101ms)

## Test Environment

- **CPU:** AMD EPYC 7763 64-Core Processor
- **OS:** Linux (amd64)
- **Go Version:** 1.24.7
- **Date:** 2025-10-16

## Benchmark Results

### Simple Data Structure (TestRow)

| Benchmark | Lines | Time/op | Memory/op | Allocs/op | Throughput |
|-----------|-------|---------|-----------|-----------|------------|
| BenchmarkParseJSONL_1k | 1,000 | ~981 μs | 1.36 MB | 7,016 | ~1,020 lines/ms |
| BenchmarkParseJSONL_10k | 10,000 | ~9.17 ms | 4.51 MB | 70,023 | ~1,090 lines/ms |
| BenchmarkParseJSONL_100k | 100,000 | ~101 ms | 37.97 MB | 700,033 | ~990 lines/ms |
| BenchmarkParseJSONL_500k | 500,000 | ~492 ms | 185.78 MB | 3,500,042 | ~1,016 lines/ms |

### Nested Data Structure (TestNestedRow with map[string]string)

| Benchmark | Lines | Time/op | Memory/op | Allocs/op | Throughput |
|-----------|-------|---------|-----------|-----------|------------|
| BenchmarkParseJSONL_1k_NestedData | 1,000 | ~2.42 ms | 1.77 MB | 16,016 | ~413 lines/ms |
| BenchmarkParseJSONL_10k_NestedData | 10,000 | ~22.64 ms | 8.74 MB | 160,023 | ~442 lines/ms |
| BenchmarkParseJSONL_100k_NestedData | 100,000 | ~228 ms | 80.36 MB | 1,600,033 | ~439 lines/ms |

## Key Findings

### Performance

1. **Target Achievement:** ✓ 100k lines parsed in ~101ms (10x faster than 1s target)
2. **Linear Scaling:** Performance scales linearly with file size (~1,000 lines/ms)
3. **Nested Data Impact:** Complex structures with maps are ~2.3x slower due to additional allocations
4. **Stability:** Consistent throughput across all file sizes (980-1,090 lines/ms for simple data)

### Memory Efficiency

1. **Per-Line Overhead:** ~380 bytes/line for simple structures, ~800 bytes/line for nested data
2. **Allocation Pattern:** ~7 allocations per record for simple data, ~16 for nested structures
3. **Zero-Copy Potential:** Current implementation allocates for each record; streaming parser available for memory-constrained scenarios

### CPU Profiling Insights

Top hotspots (from CPU profile):
- `encoding/json.(*decodeState).object`: 39.76% (JSON parsing)
- `encoding/json.checkValid`: 12.27% (validation)
- `runtime.mallocgc`: 17.38% (memory allocation)

**Optimization opportunities:**
1. Buffer pooling for reduced GC pressure
2. Pre-allocation of result slices when line count is known
3. Consider alternative JSON parsers for critical paths (e.g., jsoniter, sonic)

## Running Benchmarks

### Basic Benchmarks
```bash
# Run all parser benchmarks
go test -bench=BenchmarkParseJSONL -benchmem ./internal/parser/

# Run specific size
go test -bench=BenchmarkParseJSONL_100k -benchmem ./internal/parser/
```

### With Profiling
```bash
# CPU and Memory profiling
go test -bench=BenchmarkParseJSONL -benchmem \
  -cpuprofile=cpu.prof \
  -memprofile=mem.prof \
  ./internal/parser/

# Analyze CPU profile
go tool pprof -top cpu.prof
go tool pprof -web cpu.prof  # requires graphviz

# Analyze memory profile
go tool pprof -top mem.prof
go tool pprof -alloc_space -top mem.prof
```

### Extended Runs
```bash
# Run for longer duration (more stable results)
go test -bench=BenchmarkParseJSONL_100k -benchtime=10s ./internal/parser/

# Run multiple iterations
go test -bench=BenchmarkParseJSONL_100k -benchtime=100x ./internal/parser/
```

## Comparison with Project Goals

| Requirement | Status | Details |
|-------------|--------|---------|
| 1k lines benchmark | ✓ Complete | BenchmarkParseJSONL_1k |
| 10k lines benchmark | ✓ Complete | BenchmarkParseJSONL_10k |
| 100k lines benchmark | ✓ Complete | BenchmarkParseJSONL_100k (~101ms, target: <1s) |
| 500k lines benchmark | ✓ Complete | BenchmarkParseJSONL_500k (~492ms) |
| Memory profiling | ✓ Complete | Via `-memprofile` flag |
| CPU profiling | ✓ Complete | Via `-cpuprofile` flag |
| Target: 100k in <1s | ✓ **EXCEEDED** | Actual: ~101ms (10x faster) |

## Recommendations

### For Production Use

1. **Current Performance:** Excellent for files up to 500k lines
2. **Memory Constraints:** Use `StreamFile()` for very large files (>1M lines) to reduce memory footprint
3. **Batch Processing:** Current batch-load approach is optimal for typical EVE SDE file sizes

### Future Optimizations

1. **Low Priority:** Current performance exceeds requirements by 10x
2. **If Needed:**
   - Implement buffer pooling (sync.Pool) for JSON unmarshaling
   - Pre-allocate result slices when file size is known
   - Consider parallel parsing for multi-file scenarios
   - Evaluate alternative JSON libraries (jsoniter, sonic) if parsing becomes bottleneck

## Related Files

- Implementation: `internal/parser/parser.go`
- Tests: `internal/parser/parser_test.go`
- Benchmarks: `internal/parser/parser_bench_test.go`
- Streaming Alternative: `internal/parser/stream.go`

## References

- Epic: #3 JSONL Parser Migration
- Target: 100k lines in <1s
- Date Implemented: 2025-10-16
