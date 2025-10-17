#!/bin/bash
# tools/generate-parsers.sh
# Batch Code Generation Script for EVE SDE JSONL Parsers
# 
# ADR Reference: ADR-003 (JSONL Parser Architecture)
# Epic: #3 JSONL Parser Migration
#
# Usage: ./tools/generate-parsers.sh
#        make generate-parsers

set -euo pipefail

# Configuration
SCHEMAS_DIR="${SCHEMAS_DIR:-schemas}"
OUTPUT_DIR="${OUTPUT_DIR:-internal/parser/generated}"
PACKAGE_NAME="${PACKAGE_NAME:-generated}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if ! command -v quicktype &> /dev/null; then
        log_error "quicktype not found. Install with: npm install -g quicktype"
        exit 1
    fi
    
    local quicktype_version
    quicktype_version=$(quicktype --version | head -n 1)
    log_info "Found $quicktype_version"
    
    if [ ! -d "$SCHEMAS_DIR" ]; then
        log_error "Schemas directory not found: $SCHEMAS_DIR"
        log_info "Run 'go run tools/scrape-rift-schemas.go' first to generate schemas"
        exit 1
    fi
    
    local schema_count
    schema_count=$(find "$SCHEMAS_DIR" -name "*.json" -type f | wc -l)
    if [ "$schema_count" -eq 0 ]; then
        log_error "No JSON schemas found in $SCHEMAS_DIR"
        log_info "Run 'go run tools/scrape-rift-schemas.go' first to generate schemas"
        exit 1
    fi
    
    log_info "Found $schema_count schema file(s)"
}

# Create output directory
setup_output_dir() {
    log_info "Setting up output directory: $OUTPUT_DIR"
    mkdir -p "$OUTPUT_DIR"
}

# Generate Go structs from JSON schemas
generate_parsers() {
    log_info "Generating Go structs from JSON schemas..."
    
    local total=0
    local success=0
    local failed=0
    
    for schema in "$SCHEMAS_DIR"/*.json; do
        if [ ! -f "$schema" ]; then
            continue
        fi
        
        total=$((total + 1))
        
        # Extract table name (basename without extension)
        local table
        table=$(basename "$schema" .json)
        
        # Output Go file path
        local output_file="$OUTPUT_DIR/${table}.go"
        
        log_info "Processing: $table"
        
        # Run quicktype
        if quicktype --src "$schema" \
                     --lang go \
                     --package "$PACKAGE_NAME" \
                     --out "$output_file" 2>&1 | grep -q "error"; then
            log_error "Failed to generate: $table"
            failed=$((failed + 1))
        else
            success=$((success + 1))
            log_info "Generated: $output_file"
        fi
    done
    
    log_info "Generation complete: $success/$total successful, $failed failed"
    
    if [ "$failed" -gt 0 ]; then
        return 1
    fi
    
    return 0
}

# Add ToMap methods to generated structs
add_tomap_methods() {
    log_info "Adding ToMap methods to generated structs..."
    
    if [ ! -d "$OUTPUT_DIR" ]; then
        log_warn "Output directory not found: $OUTPUT_DIR"
        return 0
    fi
    
    local go_files
    go_files=$(find "$OUTPUT_DIR" -name "*.go" -type f)
    
    if [ -z "$go_files" ]; then
        log_warn "No Go files found in $OUTPUT_DIR"
        return 0
    fi
    
    # Run add-tomap-methods tool
    # Note: Tool must be run from repo root, processes files in-place
    for go_file in $go_files; do
        if ! go run ./tools/add-tomap-methods "$go_file"; then
            log_error "Failed to add ToMap methods to: $go_file"
            return 1
        fi
    done
    log_info "ToMap methods added successfully"
    
    return 0
}

# Format generated code
format_code() {
    log_info "Formatting generated code..."
    
    if command -v gofmt &> /dev/null; then
        gofmt -w "$OUTPUT_DIR"/*.go 2>/dev/null || true
        log_info "Code formatted with gofmt"
    else
        log_warn "gofmt not found, skipping formatting"
    fi
}

# Verify generated code
verify_code() {
    log_info "Verifying generated code compiles..."
    
    if command -v go &> /dev/null; then
        if go build ./internal/parser/generated/... 2>&1; then
            log_info "Generated code compiles successfully"
        else
            log_error "Generated code has compilation errors"
            return 1
        fi
    else
        log_warn "Go compiler not found, skipping verification"
    fi
}

# Main execution
main() {
    log_info "Starting EVE SDE Parser Generation..."
    log_info "Schemas: $SCHEMAS_DIR"
    log_info "Output: $OUTPUT_DIR"
    log_info "Package: $PACKAGE_NAME"
    echo
    
    check_prerequisites
    setup_output_dir
    generate_parsers
    add_tomap_methods
    format_code
    verify_code
    
    echo
    log_info "✓ Parser generation complete!"
    log_info "Generated Go structs are in: $OUTPUT_DIR"
    log_info "All structs include ToMap() methods for database operations"
}

# Run main function
main "$@"
