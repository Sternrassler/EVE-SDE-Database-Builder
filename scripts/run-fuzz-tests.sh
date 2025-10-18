#!/usr/bin/env bash
#
# run-fuzz-tests.sh - Execute fuzz tests for parser robustness
#
# This script runs Go native fuzz tests with configurable iteration counts
# to ensure parser robustness against malformed, edge case, and random inputs.
#
# Usage:
#   ./scripts/run-fuzz-tests.sh [iterations]
#
# Arguments:
#   iterations - Number of fuzzing iterations to run (default: 100000)
#
# Environment Variables:
#   FUZZ_TIME - Fuzzing duration (e.g., "30s", "5m") overrides iteration count

set -euo pipefail

# Configuration
ITERATIONS="${1:-100000}"
FUZZ_TIME="${FUZZ_TIME:-}"
PACKAGE="./internal/parser"

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "EVE SDE Parser Fuzz Testing"
echo "========================================="
echo ""

# List of fuzz test functions
FUZZ_TESTS=(
    "FuzzJSONLParser"
    "FuzzJSONLParserNestedData"
    "FuzzJSONLParserLargeInput"
)

# Calculate time or use iterations
if [ -n "$FUZZ_TIME" ]; then
    FUZZ_ARG="-fuzztime=$FUZZ_TIME"
    echo "Running fuzz tests for duration: $FUZZ_TIME"
else
    # Convert iterations to approximate time (assuming ~5000-10000 iterations/second)
    # Use a slightly longer time to ensure we hit the iteration count
    SECONDS=$((ITERATIONS / 5000))
    if [ $SECONDS -lt 1 ]; then
        SECONDS=1
    fi
    FUZZ_ARG="-fuzztime=${SECONDS}s"
    echo "Running fuzz tests for ~${ITERATIONS} iterations (${SECONDS}s)"
fi
echo ""

# Track results
PASSED=0
FAILED=0
TOTAL=${#FUZZ_TESTS[@]}

# Run each fuzz test
for fuzz_test in "${FUZZ_TESTS[@]}"; do
    echo "----------------------------------------"
    echo "Running: $fuzz_test"
    echo "----------------------------------------"
    
    # Run the fuzz test
    if go test -v "$PACKAGE" -fuzz="^${fuzz_test}$" "$FUZZ_ARG"; then
        echo -e "${GREEN}✓ $fuzz_test PASSED${NC}"
        ((PASSED++))
    else
        echo -e "${RED}✗ $fuzz_test FAILED${NC}"
        ((FAILED++))
    fi
    echo ""
done

# Summary
echo "========================================="
echo "Fuzz Testing Summary"
echo "========================================="
echo "Total tests:  $TOTAL"
echo -e "Passed:       ${GREEN}$PASSED${NC}"
if [ $FAILED -gt 0 ]; then
    echo -e "Failed:       ${RED}$FAILED${NC}"
else
    echo -e "Failed:       $FAILED"
fi
echo ""

# Exit with appropriate status
if [ $FAILED -gt 0 ]; then
    echo -e "${RED}Fuzz testing FAILED - found crashes or issues${NC}"
    exit 1
else
    echo -e "${GREEN}Fuzz testing PASSED - all tests completed without crashes${NC}"
    exit 0
fi
