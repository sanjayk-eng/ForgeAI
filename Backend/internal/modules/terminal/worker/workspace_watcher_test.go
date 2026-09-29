package worker

import (
	"ai-agent/internal/shared/realtime"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWorkspaceRelativePathNormalizesAndRejectsOutsidePaths(t *testing.T) {
	root := filepath.Join(t.TempDir(), "workspace")
	inside := filepath.Join(root, "src", "App.tsx")
	if got, ok := workspaceRelativePath(root, inside); !ok || got != "src/App.tsx" {
		t.Fatalf("workspaceRelativePath() = %q, %v; want src/App.tsx, true", got, ok)
	}
	if _, ok := workspaceRelativePath(root, filepath.Join(t.TempDir(), "outside.tsx")); ok {
		t.Fatal("workspaceRelativePath() accepted a path outside workspace")
	}
}

func TestIgnoredWorkspacePaths(t *testing.T) {
	for _, path := range []string{
		".git/config", "node_modules/pkg/index.js", "dist/app.js", "build/output.js", "cache/data",
	} {
		if _, ok := workspaceRelativePath("/workspace", filepath.FromSlash("/workspace/"+path)); ok {
			t.Errorf("workspaceRelativePath() accepted ignored path %q", path)
		}
	}
	if _, ok := workspaceRelativePath("/workspace", "/workspace/src/app.go"); !ok {
		t.Fatal("workspaceRelativePath() rejected source file")
	}
}

func TestWorkspaceWatcherDetectsCreateAndDebouncesWrites(t *testing.T) {
	root := t.TempDir()
	created := make(chan string, 4)
	publisher := testEventPublisher{events: created}
	watcher, err := NewWorkspaceWatcher(root, "workspace-a", "project-a", "sandbox-a", publisher, nil)
	if err != nil {
		t.Fatalf("NewWorkspaceWatcher() error = %v", err)
	}
	defer watcher.Close()

	filePath := filepath.Join(root, "app.tsx")
	if err := os.WriteFile(filePath, []byte("first"), 0o600); err != nil {
		t.Fatalf("create watched file: %v", err)
	}
	if err := os.WriteFile(filePath, []byte("second"), 0o600); err != nil {
		t.Fatalf("modify watched file: %v", err)
	}

	select {
	case got := <-created:
		if got != "file.created:app.tsx" {
			t.Fatalf("first event = %q, want file.created:app.tsx", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("watcher did not publish file create")
	}
	select {
	case got := <-created:
		t.Fatalf("duplicate write event was not debounced: %q", got)
	case <-time.After(250 * time.Millisecond):
	}
}

func TestWorkspaceWatcherReportsRename(t *testing.T) {
	root := t.TempDir()
	oldPath := filepath.Join(root, "before.tsx")
	newPath := filepath.Join(root, "after.tsx")
	if err := os.WriteFile(oldPath, []byte("content"), 0o600); err != nil {
		t.Fatalf("create source file: %v", err)
	}
	events := make(chan string, 4)
	watcher, err := NewWorkspaceWatcher(root, "workspace-a", "project-a", "sandbox-a", testEventPublisher{events: events}, nil)
	if err != nil {
		t.Fatalf("NewWorkspaceWatcher() error = %v", err)
	}
	defer watcher.Close()
	if err := os.Rename(oldPath, newPath); err != nil {
		t.Fatalf("rename watched file: %v", err)
	}
	select {
	case got := <-events:
		if got != "file.renamed:after.tsx" {
			t.Fatalf("rename event = %q, want file.renamed:after.tsx", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("watcher did not publish file rename")
	}
}

type testEventPublisher struct {
	events chan string
}

func (publisher testEventPublisher) Publish(event realtime.Event) {
	if event.Event == "git.status.changed" {
		return
	}
	publisher.events <- event.Event + ":" + event.Path
}
