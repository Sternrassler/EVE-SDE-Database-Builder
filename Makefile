.PHONY: help setup test test-tools lint build clean coverage fmt vet tidy check-hooks secrets-check commit-lint generate-parsers bench bench-baseline bench-compare fuzz fuzz-quick

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

setup: ## Complete setup: install dependencies + generate parsers
	@echo "🚀 Setting up EVE SDE Database Builder..."
	@echo ""
	@echo "1️⃣  Installing Go dependencies..."
	@go mod download
	@echo "✅ Go dependencies installed"
	@echo ""
	@echo "2️⃣  Checking quicktype (code generation tool)..."
	@if ! command -v quicktype &> /dev/null; then \
		echo "⚠️  quicktype not found - attempting to install..."; \
		if command -v npm &> /dev/null; then \
			npm install -g quicktype && echo "✅ quicktype installed"; \
		else \
			echo "❌ npm not found - please install Node.js first"; \
			echo "   Visit: https://nodejs.org/"; \
			exit 1; \
		fi; \
	else \
		echo "✅ quicktype already installed ($$(quicktype --version | head -n1))"; \
	fi
	@echo ""
	@echo "3️⃣  Generating parser code from schemas..."
	@$(MAKE) generate-parsers
	@echo ""
	@echo "🎉 Setup complete! You can now run 'make test' or 'make build'"

test: ## Run all tests (core packages)
	go test -v -p 4 -parallel 8 ./cmd/... ./internal/...

test-race: ## Run all tests with race detector
	go test -race -p 4 -parallel 8 ./...

test-tools: ## Run tests for tools (separate main packages)
	@echo "Testing add-tomap-methods..."
	@go test -v -p 2 -parallel 4 ./tools/add-tomap-methods/...
	@echo ""
	@echo "Testing scrape-rift-schemas..."
	@go test -v -p 2 -parallel 4 ./tools/scrape-rift-schemas/...

coverage: ## Run tests with coverage report
	go test -coverprofile=coverage.out -p 4 -parallel 8 ./...
	go tool cover -func=coverage.out
	@echo ""
	@echo "For HTML coverage report, run: go tool cover -html=coverage.out"

lint: ## Run linter (golangci-lint)
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install it from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

build: ## Build the project
	go build -v ./...

clean: ## Clean build artifacts
	go clean
	rm -f coverage.out

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

tidy: ## Tidy and verify dependencies
	go mod tidy
	go mod verify

check-hooks: ## Placeholder for pre-commit hook
	@echo "[check-hooks] Skipping - no hooks configured yet"

secrets-check: ## Placeholder for secrets check
	@echo "[secrets-check] Skipping - no secrets scanning configured yet"

commit-lint: ## Placeholder for commit message validation
	@echo "[commit-lint] Skipping - no commit linting configured yet"

generate-parsers: ## Generate Go parsers from JSON schemas (requires quicktype)
	@bash tools/generate-parsers.sh

bench: ## Run benchmarks for key packages
	@echo "Running benchmarks..."
	@echo ""
	@echo "Worker Pool Benchmarks:"
	@go test -bench='^BenchmarkPool_.*Workers[^_]' -benchmem ./internal/worker/
	@echo ""
	@echo "Parser Benchmarks:"
	@go test -bench='^BenchmarkParseJSONL' -benchmem ./internal/parser/
	@echo ""
	@echo "Database Benchmarks:"
	@go test -bench=. -benchmem ./internal/database/

bench-baseline: ## Capture benchmark baseline for regression testing
	@bash scripts/capture-baseline.sh

bench-compare: ## Compare current benchmarks against baseline
	@bash scripts/compare-benchmarks.sh

fuzz: ## Run fuzz tests (100k iterations for robustness testing)
	@echo "Running fuzz tests with ~100k iterations..."
	@bash scripts/run-fuzz-tests.sh 100000

fuzz-quick: ## Run quick fuzz tests (5 seconds for development)
	@echo "Running quick fuzz tests (5s)..."
	@FUZZ_TIME=5s bash scripts/run-fuzz-tests.sh

# Database Migration Targets
DB_FILE ?= eve_sde.db

migrate-status: ## Show current migration status
	@echo "Migration files in migrations/sqlite/:"
	@ls -1 migrations/sqlite/*.sql 2>/dev/null || echo "No migration files found"
	@echo ""
	@if [ -f "$(DB_FILE)" ]; then \
		echo "Database file: $(DB_FILE) (exists)"; \
		sqlite3 $(DB_FILE) "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name;" 2>/dev/null || echo "Failed to read database"; \
	else \
		echo "Database file: $(DB_FILE) (does not exist)"; \
	fi

migrate-up: ## Apply all migrations to database (creates if not exists)
	@echo "Applying migrations to $(DB_FILE)..."
	@for migration in migrations/sqlite/*.sql; do \
		echo "Applying $$migration..."; \
		sqlite3 $(DB_FILE) < "$$migration" || { echo "Failed to apply $$migration"; exit 1; }; \
	done
	@echo "All migrations applied successfully"
	@$(MAKE) migrate-status

migrate-down: ## Drop all tables (WARNING: destructive)
	@echo "WARNING: This will drop all tables in $(DB_FILE)"
	@read -p "Are you sure? (yes/NO): " confirm && [ "$$confirm" = "yes" ] || { echo "Aborted."; exit 1; }
	@if [ -f "$(DB_FILE)" ]; then \
		echo "Dropping all tables..."; \
		sqlite3 $(DB_FILE) "SELECT 'DROP TABLE IF EXISTS ' || name || ';' FROM sqlite_master WHERE type='table';" | sqlite3 $(DB_FILE); \
		echo "All tables dropped"; \
	else \
		echo "Database file $(DB_FILE) does not exist"; \
	fi
	@$(MAKE) migrate-status

migrate-clean: ## Delete database file (WARNING: destructive)
	@echo "WARNING: This will delete $(DB_FILE)"
	@read -p "Are you sure? (yes/NO): " confirm && [ "$$confirm" = "yes" ] || { echo "Aborted."; exit 1; }
	@if [ -f "$(DB_FILE)" ]; then \
		rm -f $(DB_FILE) $(DB_FILE)-shm $(DB_FILE)-wal; \
		echo "Database files deleted"; \
	else \
		echo "Database file $(DB_FILE) does not exist"; \
	fi

migrate-reset: migrate-clean migrate-up ## Reset database (clean + migrate-up)

# Release Targets
release-check: ## Check if repository is ready for release
	@echo "Checking release readiness..."
	@echo ""
	@echo "1. Checking VERSION file..."
	@if [ ! -f VERSION ]; then echo "❌ VERSION file not found"; exit 1; fi
	@cat VERSION
	@echo ""
	@echo "2. Checking CHANGELOG.md..."
	@if [ ! -f CHANGELOG.md ]; then echo "❌ CHANGELOG.md not found"; exit 1; fi
	@if ! grep -q "\[Unreleased\]" CHANGELOG.md; then echo "⚠️  Warning: No [Unreleased] section in CHANGELOG.md"; fi
	@echo "✅ CHANGELOG.md exists"
	@echo ""
	@echo "3. Running tests..."
	@$(MAKE) test
	@echo ""
	@echo "4. Running lint..."
	@$(MAKE) lint
	@echo ""
	@echo "✅ Release check passed!"
	@echo ""
	@echo "To create a release:"
	@echo "  1. Update VERSION file with new version"
	@echo "  2. Update CHANGELOG.md [Unreleased] → [X.Y.Z] - YYYY-MM-DD"
	@echo "  3. Commit changes: git commit -am 'chore: Release vX.Y.Z'"
	@echo "  4. Create tag: git tag vX.Y.Z"
	@echo "  5. Push: git push origin main && git push origin vX.Y.Z"

.PHONY: migrate-status migrate-up migrate-down migrate-clean migrate-reset release-check
