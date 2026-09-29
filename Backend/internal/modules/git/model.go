package git

import "time"

// GitStatus represents a parsed git status result for a sandbox workspace.
type GitStatus struct {
	Branch    string   `json:"branch"`
	IsDirty   bool     `json:"is_dirty"`
	Modified  []string `json:"modified"`
	Staged    []string `json:"staged"`
	Untracked []string `json:"untracked"`
	Ahead     int      `json:"ahead,omitempty"`
	Behind    int      `json:"behind,omitempty"`
}

// GitDiffEntry is a single file diff for a git change.
type GitDiffEntry struct {
	Path            string  `json:"path"`
	OldPath         string  `json:"old_path,omitempty"`
	Content         string  `json:"content"`
	Status          string  `json:"status"`
	OriginalContent *string `json:"original_content,omitempty"`
}

// GitDiffResult contains the diff summary for the workspace.
type GitDiffResult struct {
	Files []GitDiffEntry `json:"files"`
}

// CommitRequest carries the commit message for a staged workspace snapshot.
type CommitRequest struct {
	Message string   `json:"message"`
	All     bool     `json:"all,omitempty"`
	Files   []string `json:"files,omitempty"`
}

// RevertRequest carries optional file-scoped restore instructions.
type RevertRequest struct {
	All   bool     `json:"all,omitempty"`
	Files []string `json:"files,omitempty"`
}

// CommitResult captures the response from a commit operation.
type CommitResult struct {
	Message string    `json:"message"`
	Hash    string    `json:"hash,omitempty"`
	Time    time.Time `json:"time"`
}

// RevertResult captures the response from a revert operation.
type RevertResult struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// PushRequest carries optional remote branch information.
type PushRequest struct {
	Remote string `json:"remote,omitempty"`
	Branch string `json:"branch,omitempty"`
}

// PushResult is returned when a push is attempted.
type PushResult struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// StatusResponse is the API shape for git status queries.
type StatusResponse struct {
	Status GitStatus `json:"status"`
}
