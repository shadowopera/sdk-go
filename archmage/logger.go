package archmage

import (
	"log/slog"
)

var (
	_ Logger = (*defaultLogger)(nil)
	_ Logger = (*NullLogger)(nil)
)

// Logger represents a simple logging capability.
type Logger interface {
	// Info logs an informational message.
	// The args parameter is unused and exists solely to maintain signature
	// compatibility with slog.Logger.
	Info(msg string, args ...any)
}

type defaultLogger struct{}

func (l *defaultLogger) Info(msg string, args ...any) {
	slog.Default().Info(msg, args...)
}

// NullLogger is a Logger that silently discards all log messages.
type NullLogger struct{}

func (l *NullLogger) Info(msg string, args ...any) {}
