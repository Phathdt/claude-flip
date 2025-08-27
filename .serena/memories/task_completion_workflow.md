# Task Completion Workflow

## Standard Completion Steps

When completing any coding task in Claude Flip, follow these mandatory steps:

### 1. Code Quality Checks
```bash
# Format code
make fmt

# Run linter (requires golangci-lint)
make lint
```

### 2. Testing
```bash
# Run all tests
make test

# Generate coverage report (recommended)
make test-coverage
```

### 3. Build Verification
```bash
# Development build
make dev

# Full optimized build  
make build
```

### 4. Security Validation
- Verify no hardcoded credentials introduced
- Check file permissions for sensitive files (600/700)
- Ensure no sensitive data in logs
- Validate input sanitization

### 5. Integration Testing
```bash
# Test CLI commands manually
./bin/cflip help
./bin/cflip --version
```

### 6. Cross-platform Testing (if applicable)
```bash
# Test on both target platforms
make cross-compile
```

## Pre-commit Requirements
- All tests must pass
- No linting errors
- Code properly formatted
- No security vulnerabilities
- Documentation updated if needed

## Rollback Plan
- Always have rollback mechanism for major changes
- Atomic file operations with backup/restore
- Document undo plan for destructive operations

## Error Handling Verification
- Test failure scenarios
- Verify error messages are actionable
- Check timeout and cancellation mechanisms work
- Validate precondition checks function properly

## Performance Considerations
- Check memory usage for large operations  
- Verify binary size impact
- Test cleanup of temporary files
- Validate resource usage limits respected