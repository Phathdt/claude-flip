# Claude Flip

A fast and intuitive CLI tool to manage and switch between multiple Claude Code accounts.

## Features

- 🔄 **Quick Account Switching**: Switch between Claude Code accounts with a single command
- 👥 **Multi-Account Management**: Add, remove, and list your Claude Code accounts
- 🔒 **Secure Storage**: Uses system keychain (macOS) or encrypted files (Linux)
- ⚙️ **Settings Preservation**: Only switches authentication - your themes, settings, and preferences stay intact
- 🌍 **Cross-Platform**: Works seamlessly on macOS and Linux
- ⚡ **Lightweight**: Minimal dependencies, maximum performance

## Installation

### Using Go (Recommended)
```bash
go install github.com/phathdt/claude-flip@latest
```

### Download Binary
Download the latest release from [GitHub Releases](https://github.com/phathdt/claude-flip/releases) for your platform.

### Build from Source
```bash
git clone https://github.com/phathdt/claude-flip.git
cd claude-flip
go build -o cflip
sudo mv cflip /usr/local/bin/
```

## Quick Start

1. **Log into Claude Code** with your first account
2. **Add the account** to claude-flip:
   ```bash
   cflip add
   ```
3. **Switch to your second account** in Claude Code and add it:
   ```bash
   cflip add
   ```
4. **Start flipping** between accounts:
   ```bash
   cflip switch
   ```

## Usage

### Basic Commands

```bash
# Add current Claude Code account to managed accounts
cflip add

# List all managed accounts (shows which one is active)
cflip list

# Switch to the next account in sequence
cflip switch

# Switch to a specific account by number or email
cflip switch 2
cflip switch user@example.com

# Remove an account from management
cflip remove user@example.com

# Show current active account
cflip current

# Display help
cflip help
```

### Advanced Usage

```bash
# Switch to account with confirmation prompt
cflip switch --confirm

# List accounts with detailed information
cflip list --verbose

# Force switch (skip safety checks)
cflip switch --force

# Add account with custom alias
cflip add --alias "work-account"
```

## How It Works

Claude Flip only changes your authentication credentials while preserving everything else:

✅ **What gets switched:**
- Authentication tokens
- Account credentials

❌ **What stays the same:**
- Themes and UI preferences
- Settings and configurations
- Chat history
- Extensions and customizations

The tool safely stores your authentication data:
- **macOS**: Credentials in Keychain, OAuth info in `~/.claude-flip/`
- **Linux**: Encrypted storage in `~/.claude-flip/` with restricted permissions

## Requirements

- **Claude Code**: Must be installed and have logged in at least once
- **Go**: 1.19 or higher (for building from source)

The binary has no external dependencies - `jq` is not required as JSON processing is handled natively by Go.

## Troubleshooting

### Account switching not working?
1. Make sure Claude Code is completely closed before switching
2. Verify you have accounts added: `cflip list`
3. Try restarting Claude Code after switching

### Permission errors?
- Ensure you have write permissions to your home directory
- On Linux, check file permissions: `ls -la ~/.claude-flip/`

### Can't see new account after switching?
- Restart Claude Code completely (quit and reopen)
- Check current account: `cflip current`

## Uninstall

To remove claude-flip:

```bash
# If installed with go install
rm $(go env GOPATH)/bin/cflip

# If installed manually
sudo rm /usr/local/bin/cflip

# Clean up data
rm -rf ~/.claude-flip
```

Your current Claude Code session will remain active.

## Security

- All credentials are stored securely using OS-native methods
- Authentication files use restricted permissions (600)
- No sensitive data is logged or transmitted
- Requires Claude Code to be closed during switches for safety

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Support

- 🐛 **Bug Reports**: [Open an issue](https://github.com/phathdt/claude-flip/issues)
- 💡 **Feature Requests**: [Start a discussion](https://github.com/phathdt/claude-flip/discussions)
- 📖 **Documentation**: [Wiki](https://github.com/phathdt/claude-flip/wiki)

---

Made with ❤️ for the Claude Code community
