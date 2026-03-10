package archmage

import (
	"fmt"
	"os"
)

var (
	_ Logger = (*defaultLogger)(nil)
)

// Logger represents a simple logging capability.
type Logger interface {
	Info(msg string)
}

type defaultLogger struct{}

func (l *defaultLogger) Info(msg string) {
	_, _ = fmt.Fprintf(os.Stderr, "INF %s\n", msg)
}
