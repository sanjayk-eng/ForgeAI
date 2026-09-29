package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ai-agent/internal/modules/terminal/application"
)

type DockerFileStore struct {
	docker *DockerCLI
}

func NewDockerFileStore(docker *DockerCLI) *DockerFileStore {
	return &DockerFileStore{docker: docker}
}

func (store *DockerFileStore) ListFiles(ctx context.Context, containerID, path string) ([]application.FileEntry, error) {
	if strings.TrimSpace(containerID) == "" {
		return nil, fmt.Errorf("container id is required")
	}
	if strings.TrimSpace(path) == "" {
		path = "/workspace"
	}
	commandContext, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	output, err := store.docker.Run(commandContext, "exec", containerID, "sh", "-lc", `ls -A1p --full-time "$1"`, "sh", path)
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}
	return parseListFiles(output, path), nil
}

func parseListFiles(output, path string) []application.FileEntry {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	entries := make([]application.FileEntry, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "total ") {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) == 0 {
			continue
		}
		name := fields[len(fields)-1]
		isDir := strings.HasSuffix(name, "/")
		name = strings.TrimSuffix(name, "/")
		if name == "." || name == ".." {
			continue
		}
		entryPath := strings.TrimRight(path, "/") + "/" + name
		if path == "/" {
			entryPath = "/" + name
		}
		entries = append(entries, application.FileEntry{
			Name:        name,
			Path:        entryPath,
			IsDirectory: isDir,
		})
	}
	return entries
}

func (store *DockerFileStore) ReadFile(ctx context.Context, containerID, path string) (string, error) {
	if strings.TrimSpace(containerID) == "" {
		return "", fmt.Errorf("container id is required")
	}
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("file path is required")
	}
	commandContext, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	output, err := store.docker.Run(commandContext, "exec", containerID, "sh", "-lc", `sed -n '1,250p' "$1"`, "sh", path)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	return output, nil
}

func (store *DockerFileStore) WriteFile(ctx context.Context, containerID, path, content string) error {
	if strings.TrimSpace(containerID) == "" || strings.TrimSpace(path) == "" {
		return fmt.Errorf("container id and file path are required")
	}
	commandContext, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	_, err := store.docker.RunWithInput(commandContext, strings.NewReader(content),
		"exec", "-i", containerID, "sh", "-lc",
		`mkdir -p "$(dirname "$1")" && cat > "$1"`, "sh", path,
	)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func (store *DockerFileStore) CreateDirectory(ctx context.Context, containerID, path string) error {
	if strings.TrimSpace(containerID) == "" || strings.TrimSpace(path) == "" {
		return fmt.Errorf("container id and directory path are required")
	}
	commandContext, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	_, err := store.docker.Run(commandContext, "exec", containerID, "sh", "-lc", `mkdir -p "$1"`, "sh", path)
	if err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	return nil
}

func (store *DockerFileStore) DeletePath(ctx context.Context, containerID, path string) error {
	if strings.TrimSpace(containerID) == "" || strings.TrimSpace(path) == "" {
		return fmt.Errorf("container id and path are required")
	}
	commandContext, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	_, err := store.docker.Run(commandContext, "exec", containerID, "sh", "-lc", `rm -rf "$1"`, "sh", path)
	if err != nil {
		return fmt.Errorf("delete path: %w", err)
	}
	return nil
}

func (store *DockerFileStore) RenamePath(ctx context.Context, containerID, oldPath, newPath string) error {
	if strings.TrimSpace(containerID) == "" || strings.TrimSpace(oldPath) == "" || strings.TrimSpace(newPath) == "" {
		return fmt.Errorf("container id, old path and new path are required")
	}
	commandContext, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	_, err := store.docker.Run(commandContext, "exec", containerID, "sh", "-lc", `mkdir -p "$(dirname "$2")" && mv "$1" "$2"`, "sh", oldPath, newPath)
	if err != nil {
		return fmt.Errorf("rename path: %w", err)
	}
	return nil
}

var _ application.FileStore = (*DockerFileStore)(nil)
