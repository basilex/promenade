package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

// Logger wraps slog.Logger with additional convenience methods
type Logger struct {
	*slog.Logger
}

// Config holds logger configuration
type Config struct {
	Level      string // debug, info, warn, error
	Format     string // json, text
	Output     io.Writer
	AddSource  bool // Add source file/line to logs
	TimeFormat string
}

// ContextKey is the type for context keys used in logging
type ContextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// TraceIDKey is the context key for trace ID
	TraceIDKey ContextKey = "trace_id"
)

var defaultLogger *Logger

// New creates a new logger with the given configuration
func New(cfg Config) *Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}

	if cfg.TimeFormat == "" {
		cfg.TimeFormat = time.RFC3339
	}

	// Parse log level
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize time format
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(cfg.TimeFormat))
				}
			}
			// Shorten source file paths
			if a.Key == slog.SourceKey {
				if source, ok := a.Value.Any().(*slog.Source); ok {
					// Remove absolute path prefix, keep only from "promenade/"
					if idx := strings.Index(source.File, "promenade/"); idx != -1 {
						source.File = source.File[idx:]
					}
				}
			}
			return a
		},
	}

	// Create handler based on format
	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(cfg.Output, opts)
	} else {
		handler = slog.NewTextHandler(cfg.Output, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// Init initializes the default logger
func Init(cfg Config) {
	defaultLogger = New(cfg)
	slog.SetDefault(defaultLogger.Logger)
}

// Default returns the default logger
func Default() *Logger {
	if defaultLogger == nil {
		// Fallback to basic logger if not initialized
		defaultLogger = New(Config{
			Level:  "info",
			Format: "text",
			Output: os.Stdout,
		})
	}
	return defaultLogger
}

// FromContext extracts logger with context values
func FromContext(ctx context.Context) *Logger {
	logger := Default()

	// Add context fields if present
	attrs := make([]any, 0, 6)

	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		attrs = append(attrs, slog.String("request_id", requestID.(string)))
	}

	if userID := ctx.Value(UserIDKey); userID != nil {
		attrs = append(attrs, slog.String("user_id", userID.(string)))
	}

	if traceID := ctx.Value(TraceIDKey); traceID != nil {
		attrs = append(attrs, slog.String("trace_id", traceID.(string)))
	}

	if len(attrs) > 0 {
		return &Logger{
			Logger: logger.With(attrs...),
		}
	}

	return logger
}

// WithContext returns a logger with context values
func (l *Logger) WithContext(ctx context.Context) *Logger {
	attrs := make([]any, 0, 6)

	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		attrs = append(attrs, slog.String("request_id", requestID.(string)))
	}

	if userID := ctx.Value(UserIDKey); userID != nil {
		attrs = append(attrs, slog.String("user_id", userID.(string)))
	}

	if traceID := ctx.Value(TraceIDKey); traceID != nil {
		attrs = append(attrs, slog.String("trace_id", traceID.(string)))
	}

	if len(attrs) > 0 {
		return &Logger{
			Logger: l.With(attrs...),
		}
	}

	return l
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields map[string]any) *Logger {
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, slog.Any(k, v))
	}
	return &Logger{
		Logger: l.With(attrs...),
	}
}

// WithError returns a logger with error field
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.With(slog.String("error", err.Error())),
	}
}

// Package-level convenience functions

// Debug logs at debug level
func Debug(msg string, args ...any) {
	Default().Debug(msg, args...)
}

// Info logs at info level
func Info(msg string, args ...any) {
	Default().Info(msg, args...)
}

// Warn logs at warn level
func Warn(msg string, args ...any) {
	Default().Warn(msg, args...)
}

// Error logs at error level
func Error(msg string, args ...any) {
	Default().Error(msg, args...)
}

// DebugContext logs at debug level with context
func DebugContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Debug(msg, args...)
}

// InfoContext logs at info level with context
func InfoContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Info(msg, args...)
}

// WarnContext logs at warn level with context
func WarnContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Warn(msg, args...)
}

// ErrorContext logs at error level with context
func ErrorContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Error(msg, args...)
}

// Fatal logs at error level and exits
func Fatal(msg string, args ...any) {
	Default().Error(msg, args...)
	os.Exit(1)
}

// FatalContext logs at error level with context and exits
func FatalContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Error(msg, args...)
	os.Exit(1)
}
