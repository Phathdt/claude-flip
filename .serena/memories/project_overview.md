# Claude Flip - Project Overview

## Purpose
Claude Flip is a Go CLI tool designed to manage and switch between multiple Claude Code accounts. It provides secure account storage and quick switching while preserving user settings and preferences.

## Tech Stack
- **Language**: Go 1.24.3
- **CLI Framework**: github.com/urfave/cli/v2 v2.27.7
- **Module Name**: claude-flip
- **Target Platforms**: macOS and Linux (cross-platform support)

## Key Features
- Quick account switching between Claude Code accounts
- Multi-account management (add, remove, list accounts)
- Secure credential storage:
  - macOS: Native Keychain Services
  - Linux: AES-encrypted files with restricted permissions (600)
- Settings preservation (only switches authentication)
- Cross-platform compatibility
- Lightweight with minimal dependencies

## Security Model
- All credentials stored securely using OS-native methods
- Authentication files use restrictive permissions (600)
- No sensitive data logged or transmitted
- Requires Claude Code to be closed during switches for safety
- Atomic file operations with backup/restore capability

## Current Implementation Status
- CLI framework fully implemented using urfave/cli/v2
- Main commands defined with proper flags and aliases
- Core logic is stubbed - requires implementation of:
  - Authentication layer
  - Storage layer  
  - Account manager
  - Platform-specific credential storage