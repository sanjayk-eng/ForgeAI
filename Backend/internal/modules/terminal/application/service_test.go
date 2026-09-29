package application

import (
	"context"
	"testing"

	"ai-agent/internal/modules/terminal/domain"
	"ai-agent/internal/modules/terminal/policy"
)

func TestWriteFileConfinesPathsToWorkspace(t *testing.T) {
	runtime := &fileWriteRuntime{}
	service := &Service{
		store:  fileWriteStore{},
		files:  runtime,
		policy: policy.Sandbox{WorkspacePath: "/workspace"},
	}

	for _, path := range []string{"../outside.txt", "src/../../outside.txt", "/etc/passwd", "src/.git/config"} {
		if err := service.WriteFile(context.Background(), "sandbox-1", path, "content"); err != domain.ErrInvalidSandbox {
			t.Errorf("WriteFile(%q) error = %v, want ErrInvalidSandbox", path, err)
		}
	}
	if runtime.path != "" {
		t.Fatalf("invalid paths reached runtime write: %q", runtime.path)
	}

	if err := service.WriteFile(context.Background(), "sandbox-1", "src/main.go", "package main"); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if runtime.path != "/workspace/src/main.go" || runtime.content != "package main" {
		t.Fatalf("unexpected runtime write path/content: %q %q", runtime.path, runtime.content)
	}
}

func TestWorkspaceOperationsRejectTraversalAndAllowValidPaths(t *testing.T) {
	runtime := &fileWriteRuntime{}
	service := &Service{
		store:  fileWriteStore{},
		files:  runtime,
		policy: policy.Sandbox{WorkspacePath: "/workspace"},
	}

	for _, path := range []string{"../outside", "../../etc", "/tmp/x", "src/.git/secret", ".."} {
		if err := service.CreateDirectory(context.Background(), "sandbox-1", path); err != domain.ErrInvalidSandbox {
			t.Fatalf("CreateDirectory(%q) error = %v, want ErrInvalidSandbox", path, err)
		}
		if err := service.RenamePath(context.Background(), "sandbox-1", path, "safe.txt"); err != domain.ErrInvalidSandbox {
			t.Fatalf("RenamePath(%q) error = %v, want ErrInvalidSandbox", path, err)
		}
	}

	if err := service.CreateDirectory(context.Background(), "sandbox-1", "services/api"); err != nil {
		t.Fatalf("CreateDirectory() error = %v", err)
	}
	if runtime.path != "/workspace/services/api" {
		t.Fatalf("unexpected directory path: %q", runtime.path)
	}

	if err := service.CreateDirectory(context.Background(), "sandbox-1", "/workspace/services/api2"); err != nil {
		t.Fatalf("CreateDirectory(absolute workspace path) error = %v", err)
	}
	if runtime.path != "/workspace/services/api2" {
		t.Fatalf("unexpected absolute workspace path: %q", runtime.path)
	}

	if err := service.RenamePath(context.Background(), "sandbox-1", "src/main.go", "src/renamed.go"); err != nil {
		t.Fatalf("RenamePath() error = %v", err)
	}
	if runtime.path != "/workspace/src/renamed.go" {
		t.Fatalf("unexpected renamed path: %q", runtime.path)
	}
}

type fileWriteStore struct{ Store }

func (fileWriteStore) FindByID(context.Context, string) (domain.Sandbox, error) {
	return domain.Sandbox{
		ID: "sandbox-1", ContainerID: "container-1", WorkspacePath: "/workspace", Status: domain.StatusRunning,
	}, nil
}

type fileWriteRuntime struct {
	FileStore
	path    string
	content string
}

func (runtime *fileWriteRuntime) WriteFile(_ context.Context, _ string, path, content string) error {
	runtime.path = path
	runtime.content = content
	return nil
}

func (runtime *fileWriteRuntime) CreateDirectory(_ context.Context, _ string, path string) error {
	runtime.path = path
	return nil
}

func (runtime *fileWriteRuntime) DeletePath(_ context.Context, _ string, path string) error {
	runtime.path = path
	return nil
}

func (runtime *fileWriteRuntime) RenamePath(_ context.Context, _ string, oldPath, newPath string) error {
	runtime.path = newPath
	runtime.content = oldPath
	return nil
}
