#!/bin/bash
# Script to capture benchmark baselines for regression testing
set -e

BENCH_DIR="benchmarks"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo "🏁 Capturing benchmark baselines..."
echo ""

# Create backup of old baselines if they exist
if [ -f "$BENCH_DIR/baseline.txt" ]; then
    echo "📦 Backing up old baselines..."
    mkdir -p "$BENCH_DIR/archive"
    mv "$BENCH_DIR/baseline"*.txt "$BENCH_DIR/archive/" 2>/dev/null || true
fi

# Run benchmarks and save baselines
echo "🔨 Running worker pool benchmarks..."
go test -bench='^BenchmarkPool_.*Workers[^_]' -benchmem -benchtime=10x ./internal/worker/ 2>&1 | tee "$BENCH_DIR/baseline-worker.txt"

echo ""
echo "🔨 Running parser benchmarks..."
go test -bench='^BenchmarkParseJSONL' -benchmem -benchtime=10x ./internal/parser/ 2>&1 | tee "$BENCH_DIR/baseline-parser.txt"

echo ""
echo "🔨 Running database benchmarks..."
go test -bench=. -benchmem -benchtime=10x ./internal/database/ 2>&1 | tee "$BENCH_DIR/baseline-database.txt"

echo ""
echo "✅ Baselines captured successfully!"
echo ""
echo "📊 Baseline files:"
ls -lh "$BENCH_DIR"/baseline*.txt
echo ""
echo "💡 To compare against these baselines, run: make bench-compare"
