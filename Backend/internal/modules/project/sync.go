package project

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"ai-agent/internal/shared/logger"
)

var ErrSyncUnavailable = errors.New("repository is already syncing or has no pending sync")

type SyncRepositoryStore interface {
	ClaimRepositoryForSync(ctx context.Context, projectID string) (SyncTarget, error)
	ClaimNextRepositoryForSync(ctx context.Context) (SyncTarget, error)
	MarkRepositorySynced(ctx context.Context, repositoryID string, repository GitHubRepository) error
	MarkRepositorySyncFailed(ctx context.Context, repositoryID string) error
}

type SyncTarget struct {
	ID    string `db:"id"`
	Owner string `db:"github_owner"`
	Name  string `db:"github_repository_name"`
}

type SyncService struct {
	store    SyncRepositoryStore
	github   GitHubRepositoryClient
	log      logger.Logger
	interval time.Duration
}

func NewSyncService(store SyncRepositoryStore, github GitHubRepositoryClient, log logger.Logger, interval time.Duration) *SyncService {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &SyncService{store: store, github: github, log: log, interval: interval}
}

func (service *SyncService) SyncProject(ctx context.Context, projectID string) error {
	return service.syncProject(ctx, projectID, "")
}

func (service *SyncService) SyncProjectWithToken(ctx context.Context, projectID, accessToken string) error {
	return service.syncProject(ctx, projectID, accessToken)
}

func (service *SyncService) syncProject(ctx context.Context, projectID, accessToken string) error {
	target, err := service.store.ClaimRepositoryForSync(ctx, projectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSyncUnavailable
		}
		return fmt.Errorf("claim project repository: %w", err)
	}
	return service.syncTarget(ctx, target, accessToken)
}

func (service *SyncService) SyncNext(ctx context.Context) error {
	target, err := service.store.ClaimNextRepositoryForSync(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("claim next project repository: %w", err)
	}
	return service.syncTarget(ctx, target, "")
}

func (service *SyncService) Run(ctx context.Context) {
	for {
		if err := service.SyncNext(ctx); err != nil && service.log != nil {
			service.log.Warn(ctx, "project repository sync failed", "error", err)
		}
		timer := time.NewTimer(service.interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (service *SyncService) syncTarget(ctx context.Context, target SyncTarget, accessToken string) error {
	if service.github == nil {
		return fmt.Errorf("GitHub repository client is not configured")
	}
	var repository GitHubRepository
	var err error
	if tokenClient, ok := service.github.(GitHubRepositoryTokenClient); ok {
		repository, err = tokenClient.InspectRepositoryWithToken(ctx, target.Owner, target.Name, accessToken)
	} else {
		repository, err = service.github.InspectRepository(ctx, target.Owner, target.Name)
	}
	if err != nil {
		if markErr := service.store.MarkRepositorySyncFailed(ctx, target.ID); markErr != nil && service.log != nil {
			service.log.Error(ctx, "mark project repository sync failed", "repository_id", target.ID, "error", markErr)
		}
		return fmt.Errorf("sync GitHub repository %s/%s: %w", target.Owner, target.Name, err)
	}
	if err := service.store.MarkRepositorySynced(ctx, target.ID, repository); err != nil {
		return fmt.Errorf("complete project repository sync: %w", err)
	}
	return nil
}
