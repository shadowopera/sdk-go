package archmage

import (
	"time"
)

// VersionInfo represents VCS version metadata.
type VersionInfo struct {
	// Workspace is the workspace name.
	Workspace string `json:"workspace"`
	// Tags are tags associated with this version.
	Tags []string `json:"tags"`
	// Branch is the source control branch name.
	Branch string `json:"branch"`
	// ID is the full commit ID.
	ID string `json:"id"`
	// ShortID is the abbreviated commit ID.
	ShortID string `json:"shortId"`
	// Timestamp is the commit timestamp.
	Timestamp time.Time `json:"timestamp"`
	// Message is the commit message.
	Message string `json:"message"`
	// Author is the commit author.
	Author string `json:"author"`
	// Status contains working tree status entries.
	Status []string `json:"status"`
	// Extra holds additional metadata.
	Extra map[string]any `json:"extra"`
}
