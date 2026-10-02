package worker

import (
	"context"
	"errors"
	"time"

	"ai-agent/internal/modules/terminal/domain"
	"ai-agent/internal/shared/realtime"
)

func (w *SandboxWorker) handleProjectCreated(ctx context.Context, projectID, userID string) {
	unlock := w.lockProject(projectID)
	defer unlock()

	sandbox, err := w.service.GetByProject(ctx, projectID)
	if err == nil && sandbox.ID != "" {
		if sandbox.Status != domain.StatusFailed {
			if sandbox.Status == domain.StatusRunning {
				if err := w.provisionFiles(ctx, sandbox, projectID); err != nil && w.log != nil {
					w.log.Warn(ctx, "file provisioning failed", "sandbox_id", sandbox.ID, "error", err)
				}
				w.startWorkspaceWatcher(ctx, sandbox, projectID)
			}
			if w.log != nil {
				w.log.Info(ctx, "sandbox already exists", "project_id", projectID)
			}
			return
		}
		w.stopWorkspaceWatcher(sandbox.ID)
		if _, err := w.service.Destroy(ctx, sandbox.ID); err != nil {
			if w.log != nil {
				w.log.Error(ctx, "failed sandbox cleanup failed", "sandbox_id", sandbox.ID, "error", err)
			}
			return
		}
	}

	sandbox, err = w.service.Create(ctx, userID, projectID)
	if err != nil {
		if failedSandbox, getErr := w.service.GetByProject(ctx, projectID); getErr == nil && failedSandbox.Status == domain.StatusFailed {
			w.publishSandboxStatus(failedSandbox, failedSandbox.Status)
		}
		if w.log != nil {
			w.log.Error(ctx, "sandbox creation failed", "project_id", projectID, "error", err)
		}
		return
	}
	w.publishSandboxStatus(sandbox, domain.StatusCreated)

	sandboxID := sandbox.ID
	sandbox, err = w.service.Start(ctx, sandbox.ID)
	if err != nil {
		if failedSandbox, getErr := w.service.Get(ctx, sandboxID); getErr == nil {
			w.publishSandboxStatus(failedSandbox, failedSandbox.Status)
		}
		if w.log != nil {
			w.log.Error(ctx, "sandbox start failed", "sandbox_id", sandbox.ID, "error", err)
		}
		return
	}

	if err := w.provisionFiles(ctx, sandbox, projectID); err != nil {
		if w.log != nil {
			w.log.Warn(ctx, "file provisioning failed", "sandbox_id", sandbox.ID, "error", err)
		}
	}
	w.startWorkspaceWatcher(ctx, sandbox, projectID)

	if w.log != nil {
		w.log.Info(ctx, "sandbox provisioned", "sandbox_id", sandbox.ID, "project_id", projectID)
	}
}

func (w *SandboxWorker) handleProjectDeleted(ctx context.Context, projectID string) {
	if err := w.DeleteProject(ctx, projectID); err != nil && w.log != nil {
		w.log.Error(ctx, "sandbox destruction failed", "project_id", projectID, "error", err)
	}
}

func (w *SandboxWorker) RefreshProjectBranch(ctx context.Context, projectID string) error {
	unlock := w.lockProject(projectID)
	defer unlock()

	sandbox, err := w.service.GetByProject(ctx, projectID)
	if err != nil {
		if errors.Is(err, domain.ErrSandboxNotFound) {
			return nil
		}
		return err
	}
	if sandbox.Status != domain.StatusRunning {
		return nil
	}
	if err := w.provisionFiles(ctx, sandbox, projectID); err != nil {
		return err
	}
	if w.log != nil {
		w.log.Info(ctx, "sandbox repository branch refreshed", "project_id", projectID, "sandbox_id", sandbox.ID)
	}
	return nil
}

func (w *SandboxWorker) DeleteProject(ctx context.Context, projectID string) error {
	unlock := w.lockProject(projectID)
	defer unlock()

	sandbox, err := w.service.GetByProject(ctx, projectID)
	if err != nil {
		if errors.Is(err, domain.ErrSandboxNotFound) {
			return nil
		}
		return err
	}
	w.stopWorkspaceWatcher(sandbox.ID)
	if sandbox.Status == domain.StatusDestroyed {
		return nil
	}
	_, err = w.service.Destroy(ctx, sandbox.ID)
	if err == nil && w.eventsPublisher != nil {
		w.eventsPublisher.Publish(realtime.Event{
			Version: 1, Event: "sandbox.status.changed", organizationID: sandbox.organizationID,
			ProjectID: sandbox.ProjectID, SandboxID: sandbox.ID, Status: string(domain.StatusDestroyed),
		})
	}
	return err
}

func (w *SandboxWorker) provisionFiles(ctx context.Context, sandbox domain.Sandbox, projectID string) error {
	if w.projectRepo == nil {
		return nil
	}
	project, err := w.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return err
	}
	if project.Type != "REPOSITORY" || project.Repository == nil {
		return nil
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	return w.service.CloneRepository(timeoutCtx, sandbox.ID, project.Repository.URL, project.Repository.Branch, project.Repository.AccessToken)
}
