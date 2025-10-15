.PHONY: help test lint build clean coverage

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

test: ## Run all tests
	go test -v ./...

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
