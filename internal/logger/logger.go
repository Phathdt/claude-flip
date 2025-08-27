package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Logger wraps slog.Logger with additional convenience methods
type Logger struct {
	*slog.Logger
	level LogLevel
}

// LogLevel represents logging levels
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// LogConfig holds configuration for the logger
type LogConfig struct {
	Level      LogLevel
	Format     string // "json" or "text"
	Output     string // "stdout", "stderr", or file path
	AddSource  bool   // Add source code position
	Structured bool   // Use structured logging for user messages
}

// DefaultConfig returns default logging configuration
func DefaultConfig() *LogConfig {
	return &LogConfig{
		Level:      LevelInfo,
		Format:     "text",
		Output:     "stderr",
		AddSource:  false,
		Structured: false,
	}
}

// New creates a new logger with the given configuration
func New(config *LogConfig) (*Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Convert our LogLevel to slog.Level
	var slogLevel slog.Level
	switch config.Level {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level:     slogLevel,
		AddSource: config.AddSource,
	}

	// Determine output destination
	var output *os.File
	switch config.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// Assume it's a file path
		if config.Output != "" {
			// Create directory if needed
			dir := filepath.Dir(config.Output)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("failed to create log directory: %w", err)
			}

			file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err != nil {
				return nil, fmt.Errorf("failed to open log file: %w", err)
			}
			output = file
		} else {
			output = os.Stderr
		}
	}

	// Create handler based on format
	var handler slog.Handler
	if config.Format == "json" {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	logger := slog.New(handler)

	return &Logger{
		Logger: logger,
		level:  config.Level,
	}, nil
}

// NewDefault creates a logger with default configuration
func NewDefault() *Logger {
	logger, _ := New(DefaultConfig())
	return logger
}

// User-facing output methods (for CLI interaction)

// Success prints a success message with green checkmark
func (l *Logger) Success(msg string, args ...any) {
	formatted := fmt.Sprintf("✅ "+msg, args...)
	fmt.Println(formatted)
	l.Info("Success: " + strings.TrimPrefix(formatted, "✅ "))
}

// Info prints an info message with blue info icon
func (l *Logger) InfoMsg(msg string, args ...any) {
	formatted := fmt.Sprintf("📋 "+msg, args...)
	fmt.Println(formatted)
	l.Info("Info: " + strings.TrimPrefix(formatted, "📋 "))
}

// Progress prints a progress message with spinner
func (l *Logger) Progress(msg string, args ...any) {
	formatted := fmt.Sprintf("🔄 "+msg, args...)
	fmt.Println(formatted)
	l.Info("Progress: " + strings.TrimPrefix(formatted, "🔄 "))
}

// Warning prints a warning message with yellow warning icon
func (l *Logger) Warning(msg string, args ...any) {
	formatted := fmt.Sprintf("⚠️  "+msg, args...)
	fmt.Println(formatted)
	l.Warn("Warning: " + strings.TrimPrefix(formatted, "⚠️  "))
}

// Error prints an error message with red X
func (l *Logger) ErrorMsg(msg string, args ...any) {
	formatted := fmt.Sprintf("❌ "+msg, args...)
	fmt.Fprintln(os.Stderr, formatted)
	l.Error("Error: " + strings.TrimPrefix(formatted, "❌ "))
}

// Question prints a question/prompt message
func (l *Logger) Question(msg string, args ...any) {
	formatted := fmt.Sprintf("❓ "+msg, args...)
	fmt.Print(formatted)
	l.Debug("Question: " + strings.TrimPrefix(formatted, "❓ "))
}

// Plain prints a message without icons (for normal output)
func (l *Logger) Plain(msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	fmt.Println(formatted)
	l.Debug("Plain: " + formatted)
}

// Bullet prints a bulleted list item
func (l *Logger) Bullet(msg string, args ...any) {
	formatted := fmt.Sprintf("  • "+msg, args...)
	fmt.Println(formatted)
	l.Debug("Bullet: " + strings.TrimPrefix(formatted, "  • "))
}

// Header prints a header message
func (l *Logger) Header(msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	fmt.Printf("\n%s\n", formatted)
	l.Info("Header: " + formatted)
}

// Structured logging methods (for debugging and auditing)

// DebugContext logs a debug message with context
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.Logger.DebugContext(ctx, msg, args...)
}

// InfoContext logs an info message with context
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.Logger.InfoContext(ctx, msg, args...)
}

// WarnContext logs a warning message with context
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.Logger.WarnContext(ctx, msg, args...)
}

// ErrorContext logs an error message with context
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.Logger.ErrorContext(ctx, msg, args...)
}

// WithAttrs returns a new logger with the given attributes
func (l *Logger) WithAttrs(attrs ...slog.Attr) *Logger {
	return &Logger{
		Logger: l.Logger.With(attrsToAny(attrs)...),
		level:  l.level,
	}
}

// WithGroup returns a new logger with the given group
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{
		Logger: l.Logger.WithGroup(name),
		level:  l.level,
	}
}

// Operation logs the start and end of an operation
func (l *Logger) Operation(name string, fn func() error) error {
	l.Debug("Starting operation", "operation", name)
	err := fn()
	if err != nil {
		l.Error("Operation failed", "operation", name, "error", err)
		return err
	}
	l.Debug("Operation completed", "operation", name)
	return nil
}

// Audit logs an audit event (always logged regardless of level)
func (l *Logger) Audit(action string, attrs ...slog.Attr) {
	// Force audit logs to always be written
	oldLevel := l.level
	if l.level > LevelInfo {
		// Temporarily lower level for audit logs
		tempLogger, _ := New(&LogConfig{
			Level:     LevelInfo,
			Format:    "json", // Audit logs should be structured
			Output:    "stderr",
			AddSource: true,
		})
		tempLogger.Info("AUDIT", append([]any{"action", action}, attrsToAny(attrs)...)...)
	} else {
		l.Info("AUDIT", append([]any{"action", action}, attrsToAny(attrs)...)...)
	}
	_ = oldLevel // Suppress unused variable warning
}

// Account-specific logging helpers

// AccountAdded logs when an account is added
func (l *Logger) AccountAdded(email, alias string) {
	attrs := []slog.Attr{
		slog.String("email", email),
	}
	if alias != "" {
		attrs = append(attrs, slog.String("alias", alias))
	}
	l.Audit("account_added", attrs...)
}

// AccountRemoved logs when an account is removed
func (l *Logger) AccountRemoved(email string) {
	l.Audit("account_removed", slog.String("email", email))
}

// AccountSwitched logs when accounts are switched
func (l *Logger) AccountSwitched(fromEmail, toEmail string) {
	l.Audit("account_switched",
		slog.String("from_email", fromEmail),
		slog.String("to_email", toEmail))
}

// AccountRenamed logs when an account is renamed
func (l *Logger) AccountRenamed(email, oldAlias, newAlias string) {
	l.Audit("account_renamed",
		slog.String("email", email),
		slog.String("old_alias", oldAlias),
		slog.String("new_alias", newAlias))
}

// Helper function to convert slog.Attr to []any
func attrsToAny(attrs []slog.Attr) []any {
	args := make([]any, 0, len(attrs)*2)
	for _, attr := range attrs {
		args = append(args, attr.Key, attr.Value.Any())
	}
	return args
}

// Global logger instance
var defaultLogger *Logger

func init() {
	defaultLogger = NewDefault()
}

// Global convenience functions

// Success logs a success message using the default logger
func Success(msg string, args ...any) {
	defaultLogger.Success(msg, args...)
}

// Info logs an info message using the default logger
func InfoMsg(msg string, args ...any) {
	defaultLogger.InfoMsg(msg, args...)
}

// Progress logs a progress message using the default logger
func Progress(msg string, args ...any) {
	defaultLogger.Progress(msg, args...)
}

// Warning logs a warning message using the default logger
func Warning(msg string, args ...any) {
	defaultLogger.Warning(msg, args...)
}

// ErrorMsg logs an error message using the default logger
func ErrorMsg(msg string, args ...any) {
	defaultLogger.ErrorMsg(msg, args...)
}

// Question logs a question message using the default logger
func Question(msg string, args ...any) {
	defaultLogger.Question(msg, args...)
}

// Plain logs a plain message using the default logger
func Plain(msg string, args ...any) {
	defaultLogger.Plain(msg, args...)
}

// SetDefault sets the default logger
func SetDefault(l *Logger) {
	defaultLogger = l
}
