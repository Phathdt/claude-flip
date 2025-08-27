# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Claude Flip is a Go CLI tool for managing and switching between multiple Claude Code accounts. It provides secure account storage and quick switching while preserving user settings and preferences.

## Development Commands

Use the provided Makefile for all development tasks:

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

# Format code
make fmt

# Run linter (requires golangci-lint)
make lint

# Cross-compile for all platforms (Linux/macOS, AMD64/ARM64)
make cross-compile

# Create full release with binaries and checksums
make release

# Install to GOPATH/bin
make install

# Setup development environment (install golangci-lint)
make setup

# Clean all build artifacts
make clean

# Show all available targets
make help
```

### Direct Go Commands (if needed)
```bash
# Run specific command
go run cmd/cflip/main.go [command]

# Run single test file
go test ./internal/[package] -v

# Build with version info
go build -ldflags "-X main.version=0.1.0" -o bin/cflip ./cmd/cflip
```

## Architecture

### Current Project Structure
```
/
├── go.mod                  # Module: claude-flip, uses github.com/urfave/cli/v2
├── Makefile               # Complete build automation with cross-compilation
├── cmd/cflip/main.go      # CLI entry point with urfave/cli framework
├── internal/              # Private packages (to be implemented)
├── pkg/                   # Public packages (to be implemented)
└── bin/                   # Build output directory
```

### CLI Implementation
Built with **github.com/urfave/cli/v2** framework providing:
- `cflip add [--alias]` - Add current account with optional custom alias
- `cflip list [--verbose]` - List accounts with active indicator
- `cflip switch [account] [--confirm] [--force]` - Switch accounts
- `cflip remove <account>` - Remove account from management
- `cflip current` - Show current active account
- `cflip rename <account> <alias>` - Rename account alias
- `cflip validate` - Validate all stored accounts

All commands have aliases and proper help text. Core logic is stubbed for implementation.

### Core Components

**Authentication Layer**: Handles Claude Code config file parsing and token extraction. Must support different config formats and validate authentication data integrity.

**Storage Layer**: Platform-specific secure credential storage:
- macOS: Native Keychain Services integration
- Linux: AES-encrypted files with restricted permissions (600)

**Account Manager**: Core business logic for adding, listing, switching, and removing accounts. Implements safety checks and rollback mechanisms.

**CLI Layer**: Built with urfave/cli/v2 framework. Main entry point handles all command routing and flag parsing. Each command function is currently stubbed with TODO comments.

### Key Requirements

**Security First**: All credential operations must be secure by default. Never log sensitive data. Implement atomic file operations with backup/restore capability.

**Cross-Platform**: Must work seamlessly on both macOS and Linux with platform-appropriate storage mechanisms.

**Safety Checks**: Always verify Claude Code is not running before switching accounts. Implement data integrity validation and rollback mechanisms.

**User Experience**: Provide clear error messages, progress indicators, and confirmation prompts for destructive operations.

## Development Guidelines

### Error Handling
Implement structured error handling with specific failure modes. All errors should include actionable information for users.

### Testing Strategy
- Unit tests for all core packages
- Integration tests for CLI commands
- Platform-specific tests for storage mechanisms
- Mock tests using fake Claude Code configurations

### Performance Considerations
- Minimize external dependencies
- Use efficient JSON parsing for config files
- Implement proper cleanup of temporary files
- Optimize binary size for distribution

### Current Dependencies
- **github.com/urfave/cli/v2** - CLI framework (already integrated)
- Future dependencies to add:
  - Cross-platform keychain library for macOS
  - Crypto libraries for Linux encryption
  - Standard library for most operations

Keep dependencies minimal beyond these core requirements.

## Claude Code Integration

**Config Location Discovery**: Must reliably locate Claude Code configuration directory across different installation methods and platforms.

**Config Format Handling**: Parse and manipulate Claude Code's configuration files without breaking other settings. Only modify authentication-related fields.

**Process Detection**: Implement reliable detection of running Claude Code processes to prevent corruption during account switches.

## Security Considerations

**Credential Isolation**: Each account's credentials must be completely isolated. Switching should be atomic - either fully complete or fully rolled back.

**Permission Management**: All files created should have restrictive permissions (600 for credential files, 700 for directories).

**Input Validation**: Validate all user inputs, file paths, and configuration data before processing.

**Audit Trail**: Maintain logs of account operations (excluding sensitive data) for troubleshooting.
