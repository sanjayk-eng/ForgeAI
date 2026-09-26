package sync

import "errors"

type SyncTarget struct {
	ID    string `db:"id"`
	Owner string `db:"github_owner"`
	Name  string `db:"github_repository_name"`
}

type WorkspaceProjectsSyncResponse struct {
	Triggered int `json:"triggered"`
	Failed    int `json:"failed"`
}

var (
	ErrSyncUnavailable = errors.New("repository is already syncing or has no pending sync")
)
