package internal

import (
	"sync"
)

type scavenger struct {
	mu    sync.Mutex
	Lines []string
}

func newScavenger() *scavenger {
	return &scavenger{}
}

func (scv *scavenger) Info(msg string, args ...any) {
	if len(args) > 0 {
		panic("unreachable")
	}

	msg = "INF " + msg
	scv.mu.Lock()
	defer scv.mu.Unlock()
	scv.Lines = append(scv.Lines, msg)
}
