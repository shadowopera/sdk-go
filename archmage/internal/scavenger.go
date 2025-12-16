package internal

import (
	"fmt"
)

type Logger interface {
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
}

type scavenger struct {
	Lines []string
}

func newScavenger() *scavenger {
	return &scavenger{}
}

func (scv *scavenger) Infof(format string, args ...any) {
	format = "INF " + format
	scv.Lines = append(scv.Lines, fmt.Sprintf(format, args...))
}

func (scv *scavenger) Warnf(format string, args ...any) {
	format = "WRN " + format
	scv.Lines = append(scv.Lines, fmt.Sprintf(format, args...))
}
