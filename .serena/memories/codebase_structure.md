# Codebase Structure

## Directory Layout
```
/
├── go.mod                  # Module: claude-flip, uses github.com/urfave/cli/v2
├── go.sum                  # Dependency checksums
├── Makefile               # Complete build automation with cross-compilation
├── README.md              # Comprehensive documentation
├── CLAUDE.md              # Project instructions for Claude Code
├── TODO.md                # Task tracking
├── .gitignore             # Git ignore patterns
├── cmd/cflip/main.go      # CLI entry point with urfave/cli framework
├── internal/              # Private packages (empty - to be implemented)
├── pkg/                   # Public packages (empty - to be implemented)
└── bin/                   # Build output directory
```

## Main Entry Point (cmd/cflip/main.go)
- Uses urfave/cli/v2 framework
- Defines all CLI commands with proper aliases and flags
- Contains stubbed functions for core functionality:
  - `addAccount` - Add current account with optional alias
  - `listAccounts` - List accounts with active indicator  
  - `switchAccount` - Switch accounts with confirmation/force options
  - `removeAccount` - Remove account from management
  - `currentAccount` - Show current active account
  - `renameAccount` - Rename account alias
  - `validateAccounts` - Validate all stored accounts

## Architecture Components (To Be Implemented)
- **Authentication Layer**: Parse Claude Code config, extract tokens
- **Storage Layer**: Platform-specific secure credential storage
- **Account Manager**: Core business logic with safety checks
- **CLI Layer**: Already implemented with urfave/cli/v2

## Dependencies
- **github.com/urfave/cli/v2 v2.27.7**: CLI framework (implemented)
- **Standard library**: For most operations (preferred)
- Future: Cross-platform keychain library, crypto libraries for Linux