package internal

import (
	"fmt"
	"sync"
)

type scavenger struct {
	mu    sync.Mutex
	Lines []string
}

func newScavenger() *scavenger {
	return &scavenger{}
}

func (scv *scavenger) Infof(format string, args ...any) {
	format = "INF " + format
	scv.mu.Lock()
	defer scv.mu.Unlock()
	scv.Lines = append(scv.Lines, fmt.Sprintf(format, args...))
}

func (scv *scavenger) Warnf(format string, args ...any) {
	format = "WRN " + format
	scv.mu.Lock()
	defer scv.mu.Unlock()
	scv.Lines = append(scv.Lines, fmt.Sprintf(format, args...))
}
