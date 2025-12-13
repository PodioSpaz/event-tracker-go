.PHONY: build test run clean deps lint fmt migrate help

# Build variables
BINARY_NAME=event-tracker
MIGRATE_NAME=migrate
VERSION=0.1.0
BUILD_DIR=bin
DIST_DIR=dist
COVERAGE_FILE=coverage.out

# Version injection
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE) -X main.gitCommit=$(GIT_COMMIT)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the application binaries
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/app
	@echo "Building $(MIGRATE_NAME)..."
	$(GOBUILD) -o $(BUILD_DIR)/$(MIGRATE_NAME) ./cmd/migrate
	@echo "Build complete!"

test: ## Run tests with coverage
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) ./...
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "Test coverage report generated: coverage.html"

test-short: ## Run tests without coverage
	$(GOTEST) -v -short ./...

run: ## Run the application
	$(GOCMD) run ./cmd/app

migrate: ## Run database migrations
	@echo "Running database migrations..."
	@mkdir -p data
	@if [ ! -f $(BUILD_DIR)/$(MIGRATE_NAME) ]; then \
		echo "Building migrate tool..."; \
		$(GOBUILD) -o $(BUILD_DIR)/$(MIGRATE_NAME) ./cmd/migrate; \
	fi
	@./$(BUILD_DIR)/$(MIGRATE_NAME) up

clean: ## Clean build artifacts and test files
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)/
	@rm -f $(COVERAGE_FILE) coverage.html
	@rm -f data/*.db data/*.db-shm data/*.db-wal
	@echo "Clean complete!"

deps: ## Download and verify dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) verify
	@echo "Dependencies ready!"

tidy: ## Tidy go.mod and go.sum
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	@echo "Tidy complete!"

fmt: ## Format Go code
	@echo "Formatting code..."
	$(GOFMT) ./...
	@echo "Format complete!"

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...
	@echo "Vet complete!"

lint: fmt vet ## Run formatters and linters
	@echo "Linting complete!"

# Cross-compilation targets
build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/app
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/app
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/app
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/app
	@echo "Cross-compilation complete!"

build-macos: ## Build for macOS (ARM64)
	@echo "Building for macOS ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/app
	@echo "Build complete!"

bundle-macos: build-macos ## Create macOS app bundle
	@echo "Creating macOS app bundle..."
	@./scripts/bundle-macos.fish
	@echo "Bundle complete!"

# Release targets
build-release: ## Build for current platform with version info
	@echo "Building v$(VERSION) for current platform ($(shell go env GOOS)/$(shell go env GOARCH))..."
	@echo "Note: Fyne GUI requires native builds - cross-compilation not supported"
	@mkdir -p $(DIST_DIR)
	@echo "Building for $(shell go env GOOS)-$(shell go env GOARCH)..."
	$(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH) ./cmd/app
	@echo "✓ Build complete: $(DIST_DIR)/$(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH)"
	@echo ""
	@echo "For other platforms, build on native systems or use platform-specific Docker containers."

package-release: build-release bundle-macos ## Create distribution archives
	@echo "Packaging v$(VERSION) for $(shell go env GOOS)-$(shell go env GOARCH)..."
	@cd $(DIST_DIR) && tar -czf $(BINARY_NAME)-v$(VERSION)-$(shell go env GOOS)-$(shell go env GOARCH).tar.gz $(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH)
	@cd $(DIST_DIR) && zip -qr "Event-Tracker-v$(VERSION)-macOS.zip" "Event Tracker.app"
	@echo "✓ Release packages created:"
	@ls -lh $(DIST_DIR)/*.tar.gz $(DIST_DIR)/*.zip 2>/dev/null || true

# Development helpers
dev-setup: deps ## Set up development environment
	@echo "Setting up development environment..."
	@mkdir -p data logs
	@echo "Development environment ready!"

check: fmt vet test ## Run all checks (format, vet, test)
	@echo "All checks passed!"

# Database utilities
db-reset: ## Reset the database (WARNING: deletes all data)
	@echo "Resetting database..."
	@rm -f data/events.db data/events.db-shm data/events.db-wal
	@$(MAKE) migrate
	@echo "Database reset complete!"

db-backup: ## Backup the database
	@echo "Backing up database..."
	@mkdir -p backups
	@cp data/events.db backups/events-$(shell date +%Y%m%d-%H%M%S).db
	@echo "Backup complete!"

# Display project info
info: ## Display project information
	@echo "Project: $(BINARY_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Go version: $(shell go version)"
	@echo "Build directory: $(BUILD_DIR)"
	@echo "Dist directory: $(DIST_DIR)"
