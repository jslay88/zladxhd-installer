# ZLADXHD Installer Makefile
# Automated installer for Zelda: Link's Awakening DX HD on Linux

.PHONY: help build build-debug install test test-verbose cover cover-html cover-report lint fmt vet clean run tidy deps check ci

.DEFAULT_GOAL := help

# Go parameters
GOCMD=go
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOFMT=$(GOCMD) fmt
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GOMOD=$(GOCMD) mod
GORUN=$(GOCMD) run

# Build parameters
BINARY_NAME=zladxhd-installer
BINARY_PATH=./cmd/$(BINARY_NAME)
BUILD_DIR=./build
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Build flags
LDFLAGS=-ldflags "-s -w"

help: ## Show this help message
	@echo "ZLADXHD Installer - Zelda: Link's Awakening DX HD Linux Installer"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the installer binary
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(BINARY_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

build-debug: ## Build with debug symbols
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(BINARY_PATH)
	@echo "Built (debug): $(BUILD_DIR)/$(BINARY_NAME)"

install: ## Install the binary to $GOPATH/bin
	$(GOINSTALL) $(LDFLAGS) $(BINARY_PATH)
	@echo "Installed: $(BINARY_NAME)"

run: ## Run the installer directly
	$(GORUN) $(BINARY_PATH)

test: ## Run all tests
	$(GOTEST) ./... -count=1

test-verbose: ## Run all tests with verbose output
	$(GOTEST) ./... -v -count=1

cover: ## Run tests with coverage report
	$(GOTEST) ./... -coverprofile=$(COVERAGE_FILE)
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)

cover-html: cover ## Generate HTML coverage report
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

cover-report: cover ## Show coverage by package
	@echo ""
	@echo "Coverage by package:"
	@$(GOTEST) ./... -cover 2>&1 | grep -E "coverage:|ok"

lint: fmt vet ## Run linter (requires golangci-lint)
	@which golangci-lint > /dev/null || (echo "golangci-lint not found, install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...

fmt: ## Format Go source files
	$(GOFMT) ./...

vet: ## Run go vet
	$(GOVET) ./...

clean: ## Remove build artifacts and generated files
	rm -rf $(BUILD_DIR)
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)

tidy: ## Tidy and verify go.mod
	$(GOMOD) tidy
	$(GOMOD) verify

deps: ## Download dependencies
	$(GOMOD) download

check: fmt vet test ## Run fmt, vet, and tests

ci: lint check cover ## Run all CI checks (lint, fmt, vet, test, coverage)
