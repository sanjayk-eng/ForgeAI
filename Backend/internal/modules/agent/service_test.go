package agent

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	terminalapp "ai-agent/internal/modules/terminal/application"
	"ai-agent/internal/modules/terminal/domain"
	"ai-agent/internal/shared/realtime"
)

func TestValidateChangePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{name: "relative path", path: "src/main.go", want: "src/main.go"},
		{name: "workspace path", path: "/workspace/config/app.go", want: "config/app.go"},
		{name: "traversal", path: "../outside.txt", wantErr: true},
		{name: "absolute path", path: "/etc/passwd", wantErr: true},
		{name: "git metadata", path: "src/.git/config", wantErr: true},
		{name: "environment file", path: ".env.local", wantErr: true},
		{name: "credentials", path: "config/credentials.json", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := validateChangePath(test.path)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateChangePath() error = %v, wantErr %v", err, test.wantErr)
			}
			if got != test.want {
				t.Fatalf("validateChangePath() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestRunTaskReadsProjectAndAppliesChanges(t *testing.T) {
	sandbox := &fakeSandbox{
		files: map[string][]terminalapp.FileEntry{
			"/workspace": {
				{Name: ".env", Path: "/workspace/.env"},
				{Name: ".git", Path: "/workspace/.git", IsDirectory: true},
				{Name: "src", Path: "/workspace/src", IsDirectory: true},
			},
			"/workspace/src": {{Name: "main.go", Path: "/workspace/src/main.go"}},
		},
		contents: map[string]string{
			"/workspace/.env":        "SECRET=must-not-be-sent",
			"/workspace/src/main.go": "package main\n",
		},
		writes: map[string]string{},
	}
	model := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected model authorization header")
		}
		var body struct {
			Messages []chatMessage `json:"messages"`
		}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Errorf("decode model request: %v", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		if !strings.Contains(body.Messages[1].Content, "src/main.go") {
			t.Errorf("model context did not include project source")
		}
		if strings.Contains(body.Messages[1].Content, "must-not-be-sent") {
			t.Errorf("model context included a secret file")
		}
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"choices": []any{map[string]any{"message": map[string]string{
				"role":    "assistant",
				"content": `{"message":"Updated the entry point.","changes":[{"path":"src/main.go","content":"package main\\n\\nfunc main() {}\\n"}]}`,
			}}},
		})
	}))
	defer model.Close()

	events := &recordingEventPublisher{}
	service := NewService(sandbox, Config{BaseURL: model.URL, APIKey: "test-key", Model: "test-model"}, events)
	result, err := service.RunTask(context.Background(), "user-1", "sandbox-1", "Add an empty main function")
	if err != nil {
		t.Fatalf("RunTask() error = %v", err)
	}
	if result.Message != "Updated the entry point." || len(result.ChangedFiles) != 1 || result.ChangedFiles[0] != "src/main.go" {
		t.Fatalf("unexpected task result: %+v", result)
	}
	if !strings.Contains(sandbox.writes["src/main.go"], "func main()") {
		t.Fatalf("expected model change to be written, got %q", sandbox.writes["src/main.go"])
	}
	if len(events.events) != 4 || events.events[0].Event != "agent.started" || events.events[1].Event != "agent.progress" || events.events[2].Event != "agent.progress" || events.events[3].Event != "agent.completed" {
		t.Fatalf("unexpected Agent event sequence: %+v", events.events)
	}
	if events.events[0].ProjectID != "project-1" || events.events[0].SandboxID != "sandbox-1" {
		t.Fatalf("Agent event is missing its project scope: %+v", events.events[0])
	}
}

func TestRunTaskRejectsInaccessibleSandbox(t *testing.T) {
	sandbox := &fakeSandbox{accessError: errors.New("denied")}
	service := NewService(sandbox, Config{BaseURL: "https://example.com/v1", APIKey: "key", Model: "model"})
	_, err := service.RunTask(context.Background(), "user-1", "sandbox-1", "Do a task")
	if err == nil || err.Error() != "denied" {
		t.Fatalf("RunTask() error = %v, want access denial", err)
	}
}

type fakeSandbox struct {
	files       map[string][]terminalapp.FileEntry
	contents    map[string]string
	writes      map[string]string
	accessError error
}

func (sandbox *fakeSandbox) ValidateSandboxAccess(context.Context, string, string) error {
	return sandbox.accessError
}

func (sandbox *fakeSandbox) Get(context.Context, string) (domain.Sandbox, error) {
	return domain.Sandbox{ID: "sandbox-1", ProjectID: "project-1", WorkspaceID: "workspace-1", Status: domain.StatusRunning}, nil
}

func (sandbox *fakeSandbox) ListFiles(_ context.Context, _ string, dir string) ([]terminalapp.FileEntry, error) {
	return sandbox.files[dir], nil
}

func (sandbox *fakeSandbox) ReadFile(_ context.Context, _ string, filePath string) (string, error) {
	content, exists := sandbox.contents[filePath]
	if !exists {
		return "", errors.New("not found")
	}
	return content, nil
}

func (sandbox *fakeSandbox) WriteFile(_ context.Context, _ string, relativePath, content string) error {
	sandbox.writes[relativePath] = content
	return nil
}

type recordingEventPublisher struct {
	events []realtime.Event
}

func (publisher *recordingEventPublisher) Publish(event realtime.Event) {
	publisher.events = append(publisher.events, event)
}
