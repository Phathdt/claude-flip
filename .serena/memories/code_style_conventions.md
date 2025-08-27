# Code Style and Conventions

## Go Language Conventions
- **Go Version**: 1.24.3
- **Module**: claude-flip
- **Package Structure**: Standard Go layout with cmd/, internal/, pkg/
- **Dependency Management**: Go modules (go.mod/go.sum)

## Code Style Guidelines
Based on CLAUDE.md instructions and Go best practices:

### SOLID Principles
- **Single Responsibility**: Each module/function does one thing
- **Open-Closed**: Open for extension, closed for modification  
- **Liskov Substitution**: Derived classes substitutable for base types
- **Interface Segregation**: Many specific interfaces over general ones
- **Dependency Inversion**: Depend on abstractions, not concrete implementations

### Design Philosophy
- **KISS**: Keep solutions straightforward and easy to understand
- **YAGNI**: No speculative features unless explicitly required
- **Minimal Dependencies**: Use standard library when possible

## Error Handling Standards
- Implement structured error handling with specific failure modes
- Include actionable information for users in error messages
- No mock/fallback/synthetic data in production
- All preconditions must be validated before destructive operations

## Documentation Requirements
- Every function must include concise, purpose-driven docstring
- Documentation synchronized with code changes
- Clear outline: purpose, usage, parameters, examples
- Technical terms explained inline or linked

## Security Guidelines
- No hardcoded credentials - use secure storage mechanisms
- All inputs validated, sanitized, type-checked
- Avoid eval, unsanitized shell calls, command injection
- File operations follow principle of least privilege
- Log operations excluding sensitive data values
- File permissions: 600 for credentials, 700 for directories

## Testing Requirements
- Unit tests for all core packages
- Integration tests for CLI commands
- Platform-specific tests for storage mechanisms
- Test coverage should exceed 85%
- All tests must pass before deployment

## CLI Framework Patterns
- Using urfave/cli/v2 framework
- Commands have proper aliases and help text
- Flags follow consistent naming (--verbose, --confirm, --force)
- Proper argument validation and error handling