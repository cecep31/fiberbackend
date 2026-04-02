package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime"
	"time"
)

// Logger wraps slog.Logger with additional helper methods
type Logger struct {
	*slog.Logger
}

// New creates a new structured logger with the specified configuration
func New(level slog.Level, format string, addSource bool) *Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: addSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize attribute names for better readability
			switch a.Key {
			case slog.TimeKey:
				a.Key = "timestamp"
			case slog.LevelKey:
				a.Key = "level"
			case slog.MessageKey:
				a.Key = "message"
			case slog.SourceKey:
				a.Key = "source"
			}
			return a
		},
	}

	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// NewWithWriter creates a logger that writes to the specified writer
func NewWithWriter(writer io.Writer, level slog.Level, format string) *Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(writer, opts)
	} else {
		handler = slog.NewTextHandler(writer, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// WithContext returns a new Logger with the given context
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// slog.Logger doesn't have WithContext, we just return self for now
	// Context is passed via Log methods
	return l
}

// With returns a new Logger with the given attributes
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(args...),
	}
}

// Debug logs a debug message with key-value pairs
func (l *Logger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

// Debugf logs a debug message with formatting
func (l *Logger) Debugf(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

// Info logs an info message with key-value pairs
func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

// Infof logs an info message with formatting
func (l *Logger) Infof(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

// Warn logs a warning message with key-value pairs
func (l *Logger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

// Warnf logs a warning message with formatting
func (l *Logger) Warnf(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

// Error logs an error message with key-value pairs
func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}

// Errorf logs an error message with formatting
func (l *Logger) Errorf(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}

// LogAttrs logs a message with attributes at the specified level
func (l *Logger) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	l.Logger.LogAttrs(ctx, level, msg, attrs...)
}

// Group creates a group of attributes for better organization
func Group(name string, args ...any) any {
	return slog.Group(name, args...)
}

// String creates a string attribute
func String(key, value string) slog.Attr {
	return slog.String(key, value)
}

// Int creates an int attribute
func Int(key string, value int) slog.Attr {
	return slog.Int(key, value)
}

// Int64 creates an int64 attribute
func Int64(key string, value int64) slog.Attr {
	return slog.Int64(key, value)
}

// Bool creates a bool attribute
func Bool(key string, value bool) slog.Attr {
	return slog.Bool(key, value)
}

// Duration creates a duration attribute
func Duration(key string, value time.Duration) slog.Attr {
	return slog.Duration(key, value)
}

// Any creates an attribute with any value
func Any(key string, value any) slog.Attr {
	return slog.Any(key, value)
}

// Err creates an error attribute
func Err(err error) slog.Attr {
	return slog.Any("error", err)
}

// Caller returns the caller information for logging
func Caller() (string, int) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown", 0
	}
	// Extract just the filename
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' || file[i] == '\\' {
			return file[i+1:], line
		}
	}
	return file, line
}
