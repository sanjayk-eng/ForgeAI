package worker

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ai-agent/internal/shared/realtime"

	"github.com/fsnotify/fsnotify"
)

const watcherDebounce = 100 * time.Millisecond

type WorkspaceWatcher struct {
	root        string
	projectID   string
	workspaceID string
	sandboxID   string
	publisher   realtime.Publisher
	watcher     *fsnotify.Watcher
	done        chan struct{}
	closeOnce   sync.Once
	onError     func(error)
}

func NewWorkspaceWatcher(root, workspaceID, projectID, sandboxID string, publisher realtime.Publisher, onError func(error)) (*WorkspaceWatcher, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace watcher root: %w", err)
	}
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("stat workspace watcher root: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("workspace watcher root is not a directory")
	}
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create workspace watcher: %w", err)
	}
	workspaceWatcher := &WorkspaceWatcher{
		root: root, workspaceID: workspaceID, projectID: projectID, sandboxID: sandboxID,
		publisher: publisher, watcher: fsWatcher, done: make(chan struct{}), onError: onError,
	}
	if err := workspaceWatcher.addTree(root); err != nil {
		_ = fsWatcher.Close()
		return nil, err
	}
	go workspaceWatcher.run()
	return workspaceWatcher, nil
}

func (watcher *WorkspaceWatcher) Close() error {
	var err error
	watcher.closeOnce.Do(func() {
		close(watcher.done)
		err = watcher.watcher.Close()
	})
	return err
}

func (watcher *WorkspaceWatcher) addTree(root string) error {
	return filepath.WalkDir(root, func(currentPath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}
		if currentPath != watcher.root {
			relativePath, ok := workspaceRelativePath(watcher.root, currentPath)
			if !ok || isIgnoredWorkspacePath(relativePath, true) {
				return filepath.SkipDir
			}
		}
		if err := watcher.watcher.Add(currentPath); err != nil {
			return fmt.Errorf("watch workspace directory %s: %w", currentPath, err)
		}
		return nil
	})
}

func (watcher *WorkspaceWatcher) run() {
	observed := make(map[string]fsnotify.Op)
	timer := time.NewTimer(time.Hour)
	if !timer.Stop() {
		<-timer.C
	}
	var timerChannel <-chan time.Time
	defer timer.Stop()

	for {
		select {
		case <-watcher.done:
			return
		case event, open := <-watcher.watcher.Events:
			if !open {
				return
			}
			if event.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					if relativePath, ok := workspaceRelativePath(watcher.root, event.Name); ok && !isIgnoredWorkspacePath(relativePath, true) {
						if err := watcher.addTree(event.Name); err != nil {
							watcher.reportError(err)
						}
					}
				}
			}
			relativePath, ok := workspaceRelativePath(watcher.root, event.Name)
			if !ok || (event.Op&(fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 && isIgnoredWorkspacePath(relativePath, true)) {
				continue
			}
			observed[event.Name] |= event.Op
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(watcherDebounce)
			timerChannel = timer.C
		case err, open := <-watcher.watcher.Errors:
			if !open {
				return
			}
			watcher.reportError(err)
		case <-timerChannel:
			watcher.flush(observed)
			clear(observed)
			timerChannel = nil
		}
	}
}

func (watcher *WorkspaceWatcher) reportError(err error) {
	if watcher.onError != nil {
		watcher.onError(err)
	}
}
