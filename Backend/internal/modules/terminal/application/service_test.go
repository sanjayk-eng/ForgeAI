package application

import (
	"context"
	"testing"

	"ai-agent/internal/modules/terminal/domain"
	"ai-agent/internal/modules/terminal/infrastructure"
	"ai-agent/internal/modules/terminal/policy"
)

func TestWriteFileConfinesPathsToWorkspace(t *testing.T) {
	runtime := &fileWriteRuntime{}
	service := &Service{
		store:   fileWriteStore{},
		runtime: runtime,
		policy:  policy.Sandbox{},
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

type fileWriteStore struct{ Store }

func (fileWriteStore) FindByID(context.Context, string) (domain.Sandbox, error) {
	return domain.Sandbox{
		ID: "sandbox-1", ContainerID: "container-1", WorkspacePath: "/workspace", Status: domain.StatusRunning,
	}, nil
}

type fileWriteRuntime struct {
	infrastructure.Runtime
	path    string
	content string
}

func (runtime *fileWriteRuntime) WriteFile(_ context.Context, _ string, path, content string) error {
	runtime.path = path
	runtime.content = content
	return nil
}
