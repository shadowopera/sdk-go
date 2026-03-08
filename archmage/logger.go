package archmage

import (
	"fmt"
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
	fmt.Println("INF " + msg)
}
