# Suggested Commands for Claude Flip Development

## Development Commands (via Makefile)

### Primary Development Commands
```bash
# Build binary to bin/cflip
make build

# Fast development build (no optimizations)
make dev

# Run the application directly
make run

# Run tests
make test

# Run tests with HTML coverage report
make test-coverage
```

### Code Quality Commands
```bash
# Format code
make fmt

# Run linter (requires golangci-lint)
make lint

# Setup development environment (install golangci-lint)
make setup
```

### Release Commands
```bash
# Cross-compile for all platforms (Linux/macOS, AMD64/ARM64)
make cross-compile

# Create full release with binaries and checksums
make release

# Install to GOPATH/bin
make install
```

### Utility Commands
```bash
# Clean all build artifacts
make clean

# Download and tidy dependencies
make deps

# Show all available targets
make help
```

## Direct Go Commands (if needed)
```bash
# Run specific command
go run cmd/cflip/main.go [command]

# Run single test file
go test ./internal/[package] -v

# Build with version info
go build -ldflags "-X main.version=0.1.0" -o bin/cflip ./cmd/cflip
```

## System Commands (Linux)
```bash
# Standard Linux utilities available
ls, cd, grep, find, git

# File permissions (important for security)
chmod 600 ~/.claude-flip/credentials  # Restrict file permissions
ls -la ~/.claude-flip/                # Check permissions
```

## Application Commands
```bash
# Add current Claude Code account
./bin/cflip add [--alias name]

# List managed accounts
./bin/cflip list [--verbose]

# Switch accounts
./bin/cflip switch [account] [--confirm] [--force]

# Remove account
./bin/cflip remove <account>

# Show current account
./bin/cflip current

# Validate accounts
./bin/cflip validate
```
