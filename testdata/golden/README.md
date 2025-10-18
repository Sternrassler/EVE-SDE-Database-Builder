# Golden File Test Data

This directory contains golden files for parser output verification tests.

## Purpose

Golden file tests verify that parser output remains stable and consistent across code changes. Each golden file contains the expected parsed output for a specific EVE SDE table.

## Structure

```
testdata/golden/
├── README.md                           # This file
├── invTypes.golden.json                # Golden file for invTypes parser
├── invGroups.golden.json               # Golden file for invGroups parser
├── dogmaAttributes.golden.json         # Golden file for dogmaAttributes parser
└── [table_name].golden.json            # One file per parser
```

## File Format

All golden files are JSON-formatted arrays of parsed records:

```json
[
  {
    "typeID": 34,
    "typeName": "Tritanium",
    "groupID": 18,
    "mass": 0.01,
    "volume": 0.01,
    "published": 1
  },
  {
    "typeID": 35,
    "typeName": "Pyerite",
    "groupID": 18,
    "mass": 0.01,
    "volume": 0.01,
    "published": 1
  }
]
```

## Usage

### Running Golden File Tests

```bash
# Run all golden file tests
go test -v ./internal/parser/golden/...

# Run with race detector
go test -race -v ./internal/parser/golden/...

# Run only core parsers (faster)
go test -v ./internal/parser/golden/... -run TestGoldenFiles_CoreParsers

# Run single parser test
go test -v ./internal/parser/golden/... -run TestGoldenFiles_SingleParser
```

### Updating Golden Files

When you intentionally change parser behavior or update test data, regenerate golden files:

```bash
# Update all golden files
go test -v ./internal/parser/golden/... -update

# Update and show verbose output
go test -v ./internal/parser/golden/... -update -count=1

# Update only after verifying changes
git diff testdata/golden/
```

### Make Targets

```bash
# Run golden file tests
make test-golden

# Update golden files
make update-golden

# Run golden tests as part of full test suite
make test
```

## Workflow

### 1. Initial Setup

When adding golden file tests for the first time:

```bash
# Generate golden files from current parser output
go test -v ./internal/parser/golden/... -update

# Review generated files
git diff testdata/golden/

# Commit if output looks correct
git add testdata/golden/
git commit -m "Add golden files for parser tests"
```

### 2. Development Workflow

When making changes to parsers:

```bash
# 1. Make your code changes
vim internal/parser/parsers.go

# 2. Run tests to see what changed
go test -v ./internal/parser/golden/...

# 3. If changes are intentional, update golden files
go test -v ./internal/parser/golden/... -update

# 4. Review changes
git diff testdata/golden/

# 5. Commit both code and golden file updates
git add internal/parser/ testdata/golden/
git commit -m "Update parser and golden files"
```

### 3. CI/CD Integration

In CI/CD pipelines, golden file tests run automatically:

```bash
# CI runs tests without -update flag
go test -v ./internal/parser/golden/...

# If tests fail, developer must update golden files locally
# and commit the changes
```

## Best Practices

### DO ✅

- **Review golden file changes**: Always inspect diffs before committing
- **Commit golden files with code changes**: Keep them in sync
- **Use descriptive commit messages**: Explain why output changed
- **Run tests before updating**: Understand what changed and why
- **Keep test data minimal**: Golden files should be small and focused

### DON'T ❌

- **Blindly update golden files**: Always review changes first
- **Commit generated files without review**: Verify output is correct
- **Mix unrelated changes**: Update golden files in separate commits
- **Update golden files in CI**: Only update locally, then commit
- **Use excessive test data**: Keep golden files concise

## Troubleshooting

### Golden file test fails

```
--- FAIL: TestGoldenFiles_AllParsers/invTypes (0.00s)
    golden_test.go:45: Parser output mismatch for invTypes
        Golden file: testdata/golden/invTypes.golden.json
        Run with -update to update the golden file
```

**Solution:**

1. Review the diff shown in test output
2. If change is intentional: `go test -v ./internal/parser/golden/... -update`
3. If change is not intentional: Fix your code

### Golden file missing

```
--- FAIL: TestGoldenFiles_AllParsers/newTable (0.00s)
    golden.go:85: Golden file does not exist: testdata/golden/newTable.golden.json
        Run with -update to create it
```

**Solution:**

```bash
go test -v ./internal/parser/golden/... -update
```

### Test data file missing

```
--- SKIP: TestGoldenFiles_AllParsers/someTable (0.00s)
    golden_test.go:32: Test data file not found: testdata/sde/someTable.jsonl
```

**Solution:**

1. Add test data file: `testdata/sde/someTable.jsonl`
2. Or skip the test if table is not yet implemented

### Large diff output

If golden file changes are too large to review in terminal:

```bash
# Export diff to file
go test -v ./internal/parser/golden/... 2>&1 | tee test_output.txt

# Or compare files directly
diff testdata/golden/invTypes.golden.json <(go test -v ./internal/parser/golden/... -run invTypes -update 2>&1)
```

## Maintenance

### Adding New Parsers

When adding a new parser:

1. Add test data: `testdata/sde/newTable.jsonl`
2. Register parser in `internal/parser/parsers.go` or `parsers_extended.go`
3. Generate golden file: `go test -v ./internal/parser/golden/... -update`
4. Verify output: `cat testdata/golden/newTable.golden.json`
5. Commit both test data and golden file

### Updating Test Data

When updating test data in `testdata/sde/`:

1. Update `.jsonl` file: `testdata/sde/tableName.jsonl`
2. Regenerate golden file: `go test -v ./internal/parser/golden/... -update`
3. Review changes: `git diff testdata/golden/tableName.golden.json`
4. Commit both files

### Archiving Old Golden Files

If a parser is removed or renamed:

```bash
# Remove corresponding golden file
rm testdata/golden/oldTableName.golden.json

# Or rename if parser was renamed
mv testdata/golden/oldName.golden.json testdata/golden/newName.golden.json
```

## Technical Details

### Implementation

Golden file tests are implemented in:

- `internal/parser/golden/golden.go` - Core golden file utilities
- `internal/parser/golden/golden_test.go` - Golden file tests
- `internal/parser/golden/golden_unit_test.go` - Unit tests for golden package

### Format

- **Encoding**: UTF-8
- **Format**: JSON (pretty-printed with 2-space indentation)
- **Extension**: `.golden.json`
- **Structure**: Array of parsed records

### Dependencies

- `encoding/json` - JSON marshaling/unmarshaling
- `flag` - Command-line flag for `-update`
- `testing` - Go testing framework

## References

- [Golden File Testing Pattern](https://github.com/sebdah/goldie)
- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
- [EVE SDE Documentation](https://developers.eveonline.com/)
