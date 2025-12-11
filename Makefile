.PHONY: help build test clean install lint fmt vet coverage run release

# Default target
help:
	@echo "Foundagent Makefile Commands:"
	@echo "  make build         - Build the binary"
	@echo "  make test          - Run tests"
	@echo "  make test-v        - Run tests with verbose output"
	@echo "  make coverage      - Generate test coverage report"
	@echo "  make lint          - Run linters"
	@echo "  make fmt           - Format code"
	@echo "  make vet           - Run go vet"
	@echo "  make install       - Install binary to GOPATH/bin"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make run           - Build and run with arguments (make run ARGS='init myworkspace')"
	@echo "  make release       - Build release binaries for all platforms"

# Variables
BINARY_NAME=foundagent
BINARY_ALIAS=fa
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
# Strip 'v' prefix if present (version.go adds it back in output)
VERSION_NO_V=$(shell echo $(VERSION) | sed 's/^v//')
LDFLAGS=-ldflags "-s -w -X github.com/foundagent/foundagent/internal/version.Version=$(VERSION_NO_V)"
BUILD_DIR=dist
CMD_DIR=./cmd/foundagent
COVERAGE_FILE=coverage.out

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) $(CMD_DIR)
	@echo "Build complete: ./$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	go test -race ./...

# Run tests with verbose output
test-v:
	@echo "Running tests (verbose)..."
	go test -v -race ./...

# Generate test coverage
coverage:
	@echo "Generating coverage report..."
	go test -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linters
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Install binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) $(CMD_DIR)
	@if [ -n "$(GOPATH)" ]; then \
		ln -sf $(GOPATH)/bin/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_ALIAS) 2>/dev/null || true; \
	fi
	@echo "Installed to GOPATH/bin"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME) $(BINARY_ALIAS)
	rm -rf $(BUILD_DIR)
	rm -f $(COVERAGE_FILE) coverage.html
	go clean
	@echo "Clean complete"

# Run the binary with arguments
run: build
	./$(BINARY_NAME) $(ARGS)

# Build release binaries for all platforms
release:
	@echo "Building release binaries..."
	@mkdir -p $(BUILD_DIR)
	
	@echo "Building Linux AMD64..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	
	@echo "Building Linux ARM64..."
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)
	
	@echo "Building macOS AMD64..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	
	@echo "Building macOS ARM64..."
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)
	
	@echo "Building Windows AMD64..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)
	
	@echo "Building Windows ARM64..."
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe $(CMD_DIR)
	
	@echo "Creating checksums..."
	cd $(BUILD_DIR) && sha256sum * > checksums.txt
	
	@echo "Release binaries built in $(BUILD_DIR)/"
	@ls -lh $(BUILD_DIR)/

# Development workflow - format, vet, lint, test
check: fmt vet lint test
	@echo "All checks passed!"

# Quick development build and test
dev: fmt build test
	@echo "Development build complete!"
