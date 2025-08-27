# Variables
BINARY_NAME=cflip
VERSION=0.1.0
BUILD_DIR=bin
MAIN_PACKAGE=./cmd/cflip

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

.PHONY: all build clean test deps lint install dev cross-compile help tag push-tag

# Default target
all: clean deps test build

# Build the binary
build:
	@mkdir -p $(BUILD_DIR)
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

# Build and install locally
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) $(MAIN_PACKAGE)

# Development build (faster, no optimizations)
dev:
	@mkdir -p $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) for development..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	$(GOCMD) run $(MAIN_PACKAGE)

# Cross-compile for multiple platforms
cross-compile: clean
	@echo "Cross-compiling for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)
	
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	
	# macOS ARM64 (M1/M2)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)

# Generate checksums for release binaries
checksums:
	@echo "Generating checksums..."
	@cd $(BUILD_DIR) && sha256sum * > checksums.sha256

# Create release (cross-compile + checksums)
release: cross-compile checksums
	@echo "Release artifacts created in $(BUILD_DIR)/"

# Create and push a new tag
tag:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make tag VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "Creating tag $(VERSION)..."
	@if git rev-parse "$(VERSION)" >/dev/null 2>&1; then \
		echo "Error: Tag $(VERSION) already exists"; \
		exit 1; \
	fi
	@git tag -a "$(VERSION)" -m "Release $(VERSION)"
	@echo "Tag $(VERSION) created successfully"
	@echo "Run 'make push-tag VERSION=$(VERSION)' to push the tag and trigger release"

# Push tag to trigger release
push-tag:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make push-tag VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "Pushing tag $(VERSION) to origin..."
	@git push origin "$(VERSION)"
	@echo "Tag $(VERSION) pushed! GitHub Actions will now build and create the release."
	@echo "Check: https://github.com/phathdt/claude-flip/actions"

# Create tag and push in one command
release-tag: 
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make release-tag VERSION=v1.0.0"; \
		exit 1; \
	fi
	@$(MAKE) tag VERSION=$(VERSION)
	@$(MAKE) push-tag VERSION=$(VERSION)


# Setup development environment
setup:
	@echo "Setting up development environment..."
	$(GOMOD) download
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin; \
	fi

# Show help
help:
	@echo "Available targets:"
	@echo "  all            - Clean, download deps, test, and build"
	@echo "  build          - Build the binary"
	@echo "  dev            - Fast development build"
	@echo "  install        - Build and install to GOPATH/bin"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Remove build artifacts"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  lint           - Run golangci-lint"
	@echo "  fmt            - Format code"
	@echo "  run            - Run the application"
	@echo "  cross-compile  - Build for multiple platforms"
	@echo "  checksums      - Generate SHA256 checksums"
	@echo "  release        - Create release with cross-platform binaries and checksums"
	@echo "  tag            - Create a new tag (usage: make tag VERSION=v1.0.0)"
	@echo "  push-tag       - Push tag to trigger GitHub release (usage: make push-tag VERSION=v1.0.0)"
	@echo "  release-tag    - Create and push tag in one command (usage: make release-tag VERSION=v1.0.0)"
	@echo "  setup          - Setup development environment"
	@echo "  help           - Show this help message"