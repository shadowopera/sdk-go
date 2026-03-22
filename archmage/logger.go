package archmage

import (
	"fmt"
	"os"
)

var (
	_ Logger = (*defaultLogger)(nil)
	_ Logger = (*NullLogger)(nil)
)

// Logger represents a simple logging capability.
type Logger interface {
	Info(msg string)
}

type defaultLogger struct{}

func (l *defaultLogger) Info(msg string) {
	_, _ = fmt.Fprintf(os.Stderr, "INF %s\n", msg)
}

// NullLogger is a Logger that silently discards all log messages.
type NullLogger struct{}

func (l *NullLogger) Info(string) {}
