.PHONY: lint build test ci install-tools clean

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod

# Tools
GOLANGCI_LINT_VERSION=v1.64.5
GOLANGCI_LINT=go run github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

# Build output directory
BUILD_DIR=bin

# Example packages
EXAMPLES=./example ./example/retriable ./example/tls/client ./example/tls/server

## install-tools: Pre-cache golangci-lint module
install-tools:
	@echo "Pre-caching golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@$(GOLANGCI_LINT) version
	@echo "Tools ready"

## lint: Run golangci-lint
lint:
	@echo "Running linter..."
	$(GOLANGCI_LINT) run ./...

## build: Build library and examples
build:
	@echo "Building library..."
	$(GOBUILD) ./...
	@echo "Building examples..."
	@mkdir -p $(BUILD_DIR)
	@for pkg in $(EXAMPLES); do \
		name=$$(basename $$pkg); \
		echo "  Building $$pkg -> $(BUILD_DIR)/$$name"; \
		$(GOBUILD) -o $(BUILD_DIR)/$$name $$pkg; \
	done
	@echo "Build completed successfully"

## test: Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race ./...

## ci: Run lint, build, and test (same as CI pipeline)
ci: lint build test
	@echo "CI pipeline completed successfully"

## tidy: Tidy go modules
tidy:
	$(GOMOD) tidy

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean completed"

## help: Show this help message
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
