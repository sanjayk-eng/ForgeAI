package worker

import (
	"context"
	"sync"

	"ai-agent/internal/modules/terminal/application"
	"ai-agent/internal/shared/logger"
	"ai-agent/internal/shared/realtime"
)

type ProjectEvent struct {
	Type      string
	ProjectID string
	UserID    string
}

type SandboxWorker struct {
	service         *application.Service
	projectRepo     ProjectRepository
	events          chan ProjectEvent
	log             logger.Logger
	wg              sync.WaitGroup
	stopOnce        sync.Once
	stopChan        chan struct{}
	projectLocksMu  sync.Mutex
	projectLocks    map[string]*projectOperationLock
	watchersMu      sync.Mutex
	watchers        map[string]*WorkspaceWatcher
	eventsPublisher realtime.Publisher
}

type ProjectRepository interface {
	FindByID(ctx context.Context, projectID string) (Project, error)
}

type Project struct {
	ID         string
	Type       string
	UserID     string
	Repository *Repository
}

type Repository struct {
	URL         string
	Branch      string
	AccessToken string
}

func NewSandboxWorker(service *application.Service, projectRepo ProjectRepository, log logger.Logger, publishers ...realtime.Publisher) *SandboxWorker {
	worker := &SandboxWorker{
		service:      service,
		projectRepo:  projectRepo,
		events:       make(chan ProjectEvent, 100),
		log:          log,
		stopChan:     make(chan struct{}),
		projectLocks: make(map[string]*projectOperationLock),
		watchers:     make(map[string]*WorkspaceWatcher),
	}
	if len(publishers) > 0 {
		worker.eventsPublisher = publishers[0]
	}
	return worker
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
	})
	w.wg.Wait()
	w.stopAllWatchers()
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
