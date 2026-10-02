package worker

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"ai-agent/internal/shared/realtime"

	"github.com/fsnotify/fsnotify"
)

var ignoredWorkspaceDirectories = map[string]struct{}{
	".git": {}, "node_modules": {}, "dist": {}, "build": {}, "cache": {},
	".next": {}, "vendor": {}, "target": {}, ".venv": {}, "coverage": {},
}

func (watcher *WorkspaceWatcher) flush(observed map[string]fsnotify.Op) {
	if watcher.publisher == nil || len(observed) == 0 {
		return
	}
	paths := make([]string, 0, len(observed))
	for filePath := range observed {
		paths = append(paths, filePath)
	}
	sort.Strings(paths)

	var renames, creates []string
	for _, filePath := range paths {
		operation := observed[filePath]
		if operation&fsnotify.Rename != 0 {
			renames = append(renames, filePath)
		}
		if operation&fsnotify.Create != 0 {
			creates = append(creates, filePath)
		}
	}
	pairedCreates := make(map[string]struct{})
	pairedRenames := make(map[string]struct{})
	for _, oldPath := range renames {
		if _, paired := pairedRenames[oldPath]; paired {
			continue
		}
		candidate := ""
		for _, newPath := range creates {
			if _, paired := pairedCreates[newPath]; paired {
				continue
			}
			if filepath.Dir(oldPath) == filepath.Dir(newPath) {
				candidate = newPath
				break
			}
		}
		if candidate == "" && len(renames) == 1 && len(creates) == 1 {
			if _, paired := pairedCreates[creates[0]]; !paired {
				candidate = creates[0]
			}
		}
		if candidate == "" && len(renames) == 2 {
			for _, newPath := range renames {
				if newPath != oldPath {
					candidate = newPath
					break
				}
			}
		}
		if candidate != "" {
			watcher.publishPathEvent("file.renamed", candidate, oldPath, isDirectory(candidate))
			pairedCreates[candidate] = struct{}{}
			pairedRenames[oldPath] = struct{}{}
			pairedRenames[candidate] = struct{}{}
		}
	}

	published := len(pairedCreates) > 0
	for _, filePath := range paths {
		if _, paired := pairedCreates[filePath]; paired {
			continue
		}
		if _, paired := pairedRenames[filePath]; paired {
			continue
		}
		operation := observed[filePath]
		if operation&(fsnotify.Remove|fsnotify.Rename) != 0 && operation&fsnotify.Create == 0 {
			watcher.publishPathEvent("file.deleted", filePath, "", false)
			published = true
			continue
		}
		if operation&fsnotify.Create != 0 {
			eventType := "file.created"
			if operation&fsnotify.Remove != 0 {
				eventType = "file.changed"
			}
			watcher.publishPathEvent(eventType, filePath, "", isDirectory(filePath))
			published = true
			continue
		}
		if operation&fsnotify.Write != 0 {
			watcher.publishPathEvent("file.changed", filePath, "", false)
			published = true
		}
	}
	if published {
		watcher.publisher.Publish(realtime.Event{
			Version: 1, Event: "git.status.changed", organizationID: watcher.organizationID,
			ProjectID: watcher.projectID, SandboxID: watcher.sandboxID,
		})
	}
}

func (watcher *WorkspaceWatcher) publishPathEvent(eventType, filePath, oldPath string, directory bool) {
	relativePath, ok := workspaceRelativePath(watcher.root, filePath)
	if !ok {
		return
	}
	relativeOldPath := ""
	if oldPath != "" {
		relativeOldPath, ok = workspaceRelativePath(watcher.root, oldPath)
		if !ok {
			return
		}
	}
	changeType := strings.TrimPrefix(eventType, "file.")
	watcher.publisher.Publish(realtime.Event{
		Version: 1, Event: eventType, organizationID: watcher.organizationID, ProjectID: watcher.projectID,
		SandboxID: watcher.sandboxID, Path: relativePath, OldPath: relativeOldPath,
		ChangeType: changeType, IsDirectory: directory,
	})
}

func workspaceRelativePath(root, absolutePath string) (string, bool) {
	relativePath, err := filepath.Rel(root, absolutePath)
	if err != nil || relativePath == "." || relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		return "", false
	}
	relativePath = filepath.ToSlash(relativePath)
	if isIgnoredWorkspacePath(relativePath, false) {
		return "", false
	}
	return relativePath, true
}

func isIgnoredWorkspacePath(relativePath string, includeSelf bool) bool {
	segments := strings.Split(filepath.ToSlash(relativePath), "/")
	limit := len(segments)
	if !includeSelf && limit > 0 {
		limit--
	}
	for _, segment := range segments[:limit] {
		if _, ignored := ignoredWorkspaceDirectories[strings.ToLower(segment)]; ignored {
			return true
		}
	}
	return false
}

func isDirectory(filePath string) bool {
	info, err := os.Stat(filePath)
	return err == nil && info.IsDir()
}
