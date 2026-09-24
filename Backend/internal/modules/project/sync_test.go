package project

import (
	"context"
	"errors"
	"testing"
)

type fakeSyncStore struct {
	target  SyncTarget
	claimed bool
	synced  *GitHubRepository
	failed  bool
}

func (store *fakeSyncStore) ClaimRepositoryForSync(context.Context, string) (SyncTarget, error) {
	if store.claimed {
		return SyncTarget{}, errors.New("already claimed")
	}
	store.claimed = true
	return store.target, nil
}

func (store *fakeSyncStore) ClaimNextRepositoryForSync(context.Context) (SyncTarget, error) {
	return store.ClaimRepositoryForSync(context.Background(), "")
}

func (store *fakeSyncStore) MarkRepositorySynced(_ context.Context, _ string, repository GitHubRepository) error {
	store.synced = &repository
	return nil
}

func (store *fakeSyncStore) MarkRepositorySyncFailed(context.Context, string) error {
	store.failed = true
	return nil
}

type fakeGitHubClient struct {
	repository GitHubRepository
	err        error
}

func (client fakeGitHubClient) InspectRepository(context.Context, string, string) (GitHubRepository, error) {
	return client.repository, client.err
}

func TestSyncProjectMarksRepositorySynced(t *testing.T) {
	store := &fakeSyncStore{target: SyncTarget{ID: "repository-1", Owner: "forgeai", Name: "agent"}}
	want := GitHubRepository{ID: 42, Owner: "forgeai", Name: "agent", URL: "https://github.com/forgeai/agent", DefaultBranch: "main"}
	service := NewSyncService(store, fakeGitHubClient{repository: want}, nil, 0)

	if err := service.SyncProject(context.Background(), "project-1"); err != nil {
		t.Fatalf("SyncProject() error = %v", err)
	}
	if store.synced == nil || store.synced.DefaultBranch != "main" {
		t.Fatalf("expected repository to be marked synced, got %#v", store.synced)
	}
	if store.failed {
		t.Fatal("repository should not be marked failed")
	}
}

func TestSyncProjectMarksRepositoryFailed(t *testing.T) {
	store := &fakeSyncStore{target: SyncTarget{ID: "repository-1", Owner: "forgeai", Name: "missing"}}
	service := NewSyncService(store, fakeGitHubClient{err: errors.New("not found")}, nil, 0)

	if err := service.SyncProject(context.Background(), "project-1"); err == nil {
		t.Fatal("expected SyncProject() to fail")
	}
	if !store.failed {
		t.Fatal("expected repository to be marked failed")
	}
	if store.synced != nil {
		t.Fatal("repository should not be marked synced")
	}
}
