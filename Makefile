.PHONY: help test test-tools lint build clean coverage fmt vet tidy check-hooks secrets-check commit-lint generate-parsers

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

test: ## Run all tests (core packages)
	go test -v ./cmd/... ./internal/...

test-tools: ## Run tests for tools (separate main packages)
	@echo "Testing add-tomap-methods..."
	@go test -v ./tools/add-tomap-methods*.go
	@echo ""
	@echo "Testing scrape-rift-schemas..."
	@go test -v ./tools/scrape-rift-schemas*.go

coverage: ## Run tests with coverage report
	go test -coverprofile=coverage.out ./...
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

.PHONY: migrate-status migrate-up migrate-down migrate-clean migrate-reset
