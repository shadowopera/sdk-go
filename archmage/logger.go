package archmage

import (
	"fmt"
)

// Logger provides formatted logging methods with info and warning levels.
type Logger interface {
	// Infof logs an info-level message using a printf-style format string.
	Infof(format string, args ...any)
	// Warnf logs a warning-level message using a printf-style format string.
	Warnf(format string, args ...any)
}

// defaultLogger is the built-in Logger implementation that writes to stdout.
type defaultLogger struct{}

// Infof logs an info-level message with "INF" prefix to stdout.
func (l *defaultLogger) Infof(format string, args ...any) {
	format = "INF " + format + "\n"
	fmt.Printf(format, args...)
}

// Warnf logs a warning-level message with "WRN" prefix to stdout.
func (l *defaultLogger) Warnf(format string, args ...any) {
	format = "WRN " + format + "\n"
	fmt.Printf(format, args...)
}
