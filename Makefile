.PHONY: lint build test ci install-tools clean

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod

# Tools
GOPATH_BIN=$(shell $(GOCMD) env GOPATH)/bin
GOLANGCI_LINT=$(GOPATH_BIN)/golangci-lint

# Build output directory
BUILD_DIR=bin

# Example packages
EXAMPLES=./example ./example/retriable ./example/tls/client ./example/tls/server

## install-tools: Install development tools (golangci-lint)
install-tools:
	@echo "Installing golangci-lint..."
	@test -f $(GOLANGCI_LINT) || \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH_BIN) v1.62.2
	@echo "Tools installed successfully"

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
