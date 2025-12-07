package archmage

import (
	"fmt"
)

type Logger interface {
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
}

type defaultLogger struct{}

func (l *defaultLogger) Infof(format string, args ...any) {
	format = "INF " + format + "\n"
	fmt.Printf(format, args...)
}

func (l *defaultLogger) Warnf(format string, args ...any) {
	format = "WRN " + format + "\n"
	fmt.Printf(format, args...)
}
