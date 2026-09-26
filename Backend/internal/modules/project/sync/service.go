package sync

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"ai-agent/internal/modules/project/github"
	"ai-agent/internal/shared/logger"
)

type Service struct {
	repo     Repository
	github   github.Client
	log      logger.Logger
	interval time.Duration
}

func NewService(repo Repository, githubClient github.Client, log logger.Logger, interval time.Duration) *Service {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &Service{
		repo:     repo,
		github:   githubClient,
		log:      log,
		interval: interval,
	}
}

func (s *Service) SyncProject(ctx context.Context, projectID string) error {
	return s.syncProject(ctx, projectID, "")
}

func (s *Service) SyncProjectWithToken(ctx context.Context, projectID, accessToken string) error {
	return s.syncProject(ctx, projectID, accessToken)
}

func (s *Service) syncProject(ctx context.Context, projectID, accessToken string) error {
	target, err := s.repo.ClaimRepositoryForSync(ctx, projectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSyncUnavailable
		}
		return fmt.Errorf("claim project repository: %w", err)
	}
	return s.syncTarget(ctx, target, accessToken)
}

func (s *Service) SyncNext(ctx context.Context) error {
	target, err := s.repo.ClaimNextRepositoryForSync(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("claim next project repository: %w", err)
	}
	return s.syncTarget(ctx, target, "")
}

func (s *Service) Run(ctx context.Context) {
	for {
		if err := s.SyncNext(ctx); err != nil && s.log != nil {
			s.log.Warn(ctx, "project repository sync failed", "error", err)
		}

		timer := time.NewTimer(s.interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (s *Service) syncTarget(ctx context.Context, target SyncTarget, accessToken string) error {
	if s.github == nil {
		return fmt.Errorf("GitHub repository client is not configured")
	}

	var repo github.Repository
	var err error

	if accessToken != "" {
		repo, err = s.github.InspectRepositoryWithToken(ctx, target.Owner, target.Name, accessToken)
	} else {
		repo, err = s.github.InspectRepository(ctx, target.Owner, target.Name)
	}

	if err != nil {
		if markErr := s.repo.MarkRepositorySyncFailed(ctx, target.ID); markErr != nil && s.log != nil {
			s.log.Error(ctx, "mark project repository sync failed", "repository_id", target.ID, "error", markErr)
		}
		return fmt.Errorf("sync GitHub repository %s/%s: %w", target.Owner, target.Name, err)
	}

	if err := s.repo.MarkRepositorySynced(ctx, target.ID, repo); err != nil {
		return fmt.Errorf("complete project repository sync: %w", err)
	}

	return nil
}
