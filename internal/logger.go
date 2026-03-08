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

func (scv *scavenger) Info(msg string) {
	msg = "INF " + msg
	scv.mu.Lock()
	defer scv.mu.Unlock()
	scv.Lines = append(scv.Lines, msg)
}
