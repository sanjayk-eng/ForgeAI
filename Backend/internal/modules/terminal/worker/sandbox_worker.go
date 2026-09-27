package worker

import (
	"context"
	"sync"
	"time"

	"ai-agent/internal/modules/terminal/application"
	"ai-agent/internal/modules/terminal/domain"
	"ai-agent/internal/shared/logger"
)

type ProjectEvent struct {
	Type      string
	ProjectID string
	UserID    string
}

type SandboxWorker struct {
	service     *application.Service
	projectRepo ProjectRepository
	events      chan ProjectEvent
	log         logger.Logger
	wg          sync.WaitGroup
	stopOnce    sync.Once
	stopChan    chan struct{}
	creatingMu  sync.Mutex
	creating    map[string]struct{}
}

type ProjectRepository interface {
	FindByID(ctx context.Context, projectID string) (Project, error)
}

type Project struct {
	ID         string
	Type       string
	Repository *Repository
}

type Repository struct {
	URL    string
	Branch string
}

func NewSandboxWorker(service *application.Service, projectRepo ProjectRepository, log logger.Logger) *SandboxWorker {
	return &SandboxWorker{
		service:     service,
		projectRepo: projectRepo,
		events:      make(chan ProjectEvent, 100),
		log:         log,
		stopChan:    make(chan struct{}),
		creating:    make(map[string]struct{}),
	}
}

func (w *SandboxWorker) Start(ctx context.Context, workerCount int) {
	if workerCount <= 0 {
		workerCount = 3
	}

	w.wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go w.process(ctx, i+1)
	}

	if w.log != nil {
		w.log.Info(ctx, "sandbox workers started", "count", workerCount)
	}
}

func (w *SandboxWorker) Stop() {
	w.stopOnce.Do(func() {
		close(w.stopChan)
		close(w.events)
	})
	w.wg.Wait()
}

func (w *SandboxWorker) Publish(event ProjectEvent) {
	select {
	case w.events <- event:
	case <-w.stopChan:
	default:
		if w.log != nil {
			w.log.Warn(context.Background(), "sandbox worker queue full", "event", event.Type)
		}
	}
}

func (w *SandboxWorker) process(ctx context.Context, workerID int) {
	defer w.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			if w.log != nil {
				w.log.Error(ctx, "sandbox worker panic recovered",
					"worker", workerID,
					"panic", r,
				)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopChan:
			return
		case event, ok := <-w.events:
			if !ok {
				return
			}
			w.handleEvent(ctx, workerID, event)
		}
	}
}

func (w *SandboxWorker) handleEvent(ctx context.Context, workerID int, event ProjectEvent) {
	if w.log != nil {
		w.log.Info(ctx, "processing sandbox event",
			"worker", workerID,
			"event", event.Type,
			"project_id", event.ProjectID,
		)
	}

	switch event.Type {
	case "project.created":
		w.handleProjectCreated(ctx, event.ProjectID, event.UserID)
	case "project.deleted":
		w.handleProjectDeleted(ctx, event.ProjectID)
	}
}

func (w *SandboxWorker) handleProjectCreated(ctx context.Context, projectID, userID string) {
	w.creatingMu.Lock()
	if _, exists := w.creating[projectID]; exists {
		w.creatingMu.Unlock()
		return
	}
	w.creating[projectID] = struct{}{}
	w.creatingMu.Unlock()
	defer func() {
		w.creatingMu.Lock()
		delete(w.creating, projectID)
		w.creatingMu.Unlock()
	}()

	sandbox, err := w.service.GetByProject(ctx, projectID)
	if err == nil && sandbox.ID != "" {
		if sandbox.Status != domain.StatusFailed {
			if w.log != nil {
				w.log.Info(ctx, "sandbox already exists", "project_id", projectID)
			}
			return
		}
		if _, err := w.service.Destroy(ctx, sandbox.ID); err != nil {
			if w.log != nil {
				w.log.Error(ctx, "failed sandbox cleanup failed", "sandbox_id", sandbox.ID, "error", err)
			}
			return
		}
	}

	sandbox, err = w.service.Create(ctx, userID, projectID)
	if err != nil {
		if w.log != nil {
			w.log.Error(ctx, "sandbox creation failed", "project_id", projectID, "error", err)
		}
		return
	}

	sandbox, err = w.service.Start(ctx, sandbox.ID)
	if err != nil {
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

	if w.log != nil {
		w.log.Info(ctx, "sandbox provisioned", "sandbox_id", sandbox.ID, "project_id", projectID)
	}
}

func (w *SandboxWorker) handleProjectDeleted(ctx context.Context, projectID string) {
	sandbox, err := w.service.GetByProject(ctx, projectID)
	if err != nil {
		return
	}

	if sandbox.Status == domain.StatusDestroyed {
		return
	}

	_, err = w.service.Destroy(ctx, sandbox.ID)
	if err != nil && w.log != nil {
		w.log.Error(ctx, "sandbox destruction failed", "sandbox_id", sandbox.ID, "error", err)
	}
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

	installGit := "command -v git >/dev/null 2>&1 || apk add --no-cache git"
	if _, err := w.service.Execute(timeoutCtx, sandbox.ID, installGit); err != nil {
		return err
	}

	cloneCmd := "git clone --depth 1 --branch " + project.Repository.Branch + " " + project.Repository.URL + " ."
	if _, err := w.service.Execute(timeoutCtx, sandbox.ID, cloneCmd); err != nil {
		return err
	}

	return nil
}
