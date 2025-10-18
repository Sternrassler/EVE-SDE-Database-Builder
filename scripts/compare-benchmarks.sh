#!/bin/bash
# Script to compare current benchmarks against baseline
set -e

BENCH_DIR="benchmarks"
TMP_DIR="/tmp/bench-compare-$$"

# Check if benchstat is installed
if ! command -v benchstat &> /dev/null; then
    echo "❌ benchstat not found. Installing..."
    go install golang.org/x/perf/cmd/benchstat@latest
fi

mkdir -p "$TMP_DIR"

echo "🔍 Comparing benchmarks against baseline..."
echo ""

has_regression=false

# Compare worker benchmarks
if [ -f "$BENCH_DIR/baseline-worker.txt" ]; then
    echo "📊 Worker Pool Benchmarks:"
    echo "=========================="
    go test -bench='^BenchmarkPool_.*Workers[^_]' -benchmem -benchtime=10x ./internal/worker/ 2>&1 > "$TMP_DIR/current-worker.txt"
    
    if benchstat "$BENCH_DIR/baseline-worker.txt" "$TMP_DIR/current-worker.txt" 2>&1 | tee "$TMP_DIR/worker-comparison.txt"; then
        # Check for regressions (>10% slower)
        if grep -q "~" "$TMP_DIR/worker-comparison.txt" && grep -E '\+[1-9][0-9](\.[0-9]+)?%' "$TMP_DIR/worker-comparison.txt" | grep -qv '^#'; then
            echo "⚠️  Potential regression detected in worker benchmarks"
            has_regression=true
        fi
    fi
    echo ""
else
    echo "⚠️  No worker baseline found. Run 'make bench-baseline' first."
    echo ""
fi

# Compare parser benchmarks
if [ -f "$BENCH_DIR/baseline-parser.txt" ]; then
    echo "📊 Parser Benchmarks:"
    echo "====================="
    go test -bench='^BenchmarkParseJSONL' -benchmem -benchtime=10x ./internal/parser/ 2>&1 > "$TMP_DIR/current-parser.txt"
    
    if benchstat "$BENCH_DIR/baseline-parser.txt" "$TMP_DIR/current-parser.txt" 2>&1 | tee "$TMP_DIR/parser-comparison.txt"; then
        if grep -q "~" "$TMP_DIR/parser-comparison.txt" && grep -E '\+[1-9][0-9](\.[0-9]+)?%' "$TMP_DIR/parser-comparison.txt" | grep -qv '^#'; then
            echo "⚠️  Potential regression detected in parser benchmarks"
            has_regression=true
        fi
    fi
    echo ""
else
    echo "⚠️  No parser baseline found. Run 'make bench-baseline' first."
    echo ""
fi

# Compare database benchmarks
if [ -f "$BENCH_DIR/baseline-database.txt" ]; then
    echo "📊 Database Benchmarks:"
    echo "======================="
    go test -bench=. -benchmem -benchtime=10x ./internal/database/ 2>&1 > "$TMP_DIR/current-database.txt"
    
    if benchstat "$BENCH_DIR/baseline-database.txt" "$TMP_DIR/current-database.txt" 2>&1 | tee "$TMP_DIR/database-comparison.txt"; then
        if grep -q "~" "$TMP_DIR/database-comparison.txt" && grep -E '\+[1-9][0-9](\.[0-9]+)?%' "$TMP_DIR/database-comparison.txt" | grep -qv '^#'; then
            echo "⚠️  Potential regression detected in database benchmarks"
            has_regression=true
        fi
    fi
    echo ""
else
    echo "⚠️  No database baseline found. Run 'make bench-baseline' first."
    echo ""
fi

# Cleanup
rm -rf "$TMP_DIR"

if [ "$has_regression" = true ]; then
    echo "❌ Performance regressions detected!"
    echo ""
    echo "💡 To update baseline if changes are expected, run: make bench-baseline"
    exit 1
else
    echo "✅ No significant performance regressions detected"
    exit 0
fi
