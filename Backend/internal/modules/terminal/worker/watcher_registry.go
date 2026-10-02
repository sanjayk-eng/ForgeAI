package worker

import (
	"context"

	"ai-agent/internal/modules/terminal/domain"
	"ai-agent/internal/modules/terminal/infrastructure/docker"
	"ai-agent/internal/shared/realtime"
)

func (w *SandboxWorker) startWorkspaceWatcher(ctx context.Context, sandbox domain.Sandbox, projectID string) {
	if w.eventsPublisher == nil {
		return
	}
	w.watchersMu.Lock()
	if w.watchers[sandbox.ID] != nil {
		w.watchersMu.Unlock()
		return
	}
	w.watchersMu.Unlock()

	watcher, err := NewWorkspaceWatcher(
		docker.HostWorkspacePath(sandbox.VolumeName), sandbox.OrganizationID, projectID, sandbox.ID,
		w.eventsPublisher, func(err error) {
			if w.log != nil {
				w.log.Warn(ctx, "sandbox filesystem watcher error", "sandbox_id", sandbox.ID, "error", err)
			}
		},
	)
	if err != nil {
		if w.log != nil {
			w.log.Warn(ctx, "sandbox filesystem watcher could not start", "sandbox_id", sandbox.ID, "error", err)
		}
		return
	}

	select {
	case <-w.stopChan:
		_ = watcher.Close()
		return
	default:
	}
	w.watchersMu.Lock()
	w.watchers[sandbox.ID] = watcher
	w.watchersMu.Unlock()
	w.publishSandboxStatus(sandbox, domain.StatusRunning)
}

func (w *SandboxWorker) publishSandboxStatus(sandbox domain.Sandbox, status domain.SandboxStatus) {
	if w.eventsPublisher == nil {
		return
	}
	w.eventsPublisher.Publish(realtime.Event{
		Version: 1, Event: "sandbox.status.changed", OrganizationID: sandbox.OrganizationID,
		ProjectID: sandbox.ProjectID, SandboxID: sandbox.ID, Status: string(status),
	})
}

func (w *SandboxWorker) stopWorkspaceWatcher(sandboxID string) {
	w.watchersMu.Lock()
	watcher := w.watchers[sandboxID]
	delete(w.watchers, sandboxID)
	w.watchersMu.Unlock()
	if watcher != nil {
		_ = watcher.Close()
	}
}

func (w *SandboxWorker) stopAllWatchers() {
	w.watchersMu.Lock()
	watchers := make([]*WorkspaceWatcher, 0, len(w.watchers))
	for sandboxID, watcher := range w.watchers {
		watchers = append(watchers, watcher)
		delete(w.watchers, sandboxID)
	}
	w.watchersMu.Unlock()
	for _, watcher := range watchers {
		_ = watcher.Close()
	}
}
