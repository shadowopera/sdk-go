package archmage

import (
	"time"
)

type VersionInfo struct {
	Workspace string         `json:"workspace"`
	Tags      []string       `json:"tags"`
	Branch    string         `json:"branch"`
	ID        string         `json:"id"`
	ShortID   string         `json:"shortId"`
	Timestamp time.Time      `json:"timestamp"`
	Message   string         `json:"message"`
	Author    string         `json:"author"`
	Status    []string       `json:"status"`
	Extra     map[string]any `json:"extra"`
}
