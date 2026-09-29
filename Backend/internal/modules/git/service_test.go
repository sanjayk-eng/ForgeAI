package git

import (
	"context"
	"strings"
	"testing"

	terminalapp "ai-agent/internal/modules/terminal/application"
)

func TestParseStatusOutput(t *testing.T) {
	output := "## main\n M README.md\n?? notes.txt\nA  app.go\n"

	status, err := parseStatusOutput(output)
	if err != nil {
		t.Fatalf("parseStatusOutput() returned error: %v", err)
	}
	if status.Branch != "main" {
		t.Fatalf("parseStatusOutput() branch = %q, want %q", status.Branch, "main")
	}
	if !status.IsDirty {
		t.Fatal("parseStatusOutput() IsDirty = false, want true")
	}
	if len(status.Modified) != 1 || status.Modified[0] != "README.md" {
		t.Fatalf("parseStatusOutput() modified = %#v, want [README.md]", status.Modified)
	}
	if len(status.Untracked) != 1 || status.Untracked[0] != "notes.txt" {
		t.Fatalf("parseStatusOutput() untracked = %#v, want [notes.txt]", status.Untracked)
	}
	if len(status.Staged) != 1 || status.Staged[0] != "app.go" {
		t.Fatalf("parseStatusOutput() staged = %#v, want [app.go]", status.Staged)
	}
}

func TestNewServiceKeepsContainerPathPOSIX(t *testing.T) {
	service := NewService("/workspace", &pushExecutor{})
	if service.workspaceRoot != "/workspace" {
		t.Fatalf("workspace root = %q, want POSIX container path /workspace", service.workspaceRoot)
	}
}

func TestParseStatusOutputRequiresRepoState(t *testing.T) {
	if _, err := parseStatusOutput("fatal: not a git repository"); err == nil {
		t.Fatal("parseStatusOutput() should reject non-repository output")
	}
}

func TestParseStatusOutputReturnsEmptyArraysForCleanRepository(t *testing.T) {
	status, err := parseStatusOutput("## main...origin/main\n")
	if err != nil {
		t.Fatalf("parseStatusOutput() returned error: %v", err)
	}
	if status.Modified == nil || status.Staged == nil || status.Untracked == nil {
		t.Fatalf("clean status contains nil arrays: %+v", status)
	}
}

func TestParseStatusOutputTracksDivergenceAndRenames(t *testing.T) {
	status, err := parseStatusOutput("## feature...origin/feature [ahead 2, behind 1]\nR  old.txt -> new.txt\n")
	if err != nil {
		t.Fatalf("parseStatusOutput() returned error: %v", err)
	}
	if status.Branch != "feature" || status.Ahead != 2 || status.Behind != 1 {
		t.Fatalf("parseStatusOutput() tracking = %#v, want feature/ahead=2/behind=1", status)
	}
	if len(status.Staged) != 1 || status.Staged[0] != "new.txt" {
		t.Fatalf("parseStatusOutput() staged = %#v, want [new.txt]", status.Staged)
	}
}

func TestPushPassesGitHubTokenThroughStdin(t *testing.T) {
	executor := &pushExecutor{remoteURL: "https://github.com/acme/project.git"}
	service := NewService("/workspace", executor)

	if _, err := service.Push(context.Background(), "sandbox-1", PushRequest{}, "github-token"); err != nil {
		t.Fatalf("Push() error = %v", err)
	}
	if executor.pushRemote != "origin" || executor.pushToken != "github-token" {
		t.Fatalf("push target/token = %q/%q, want origin/github-token", executor.pushRemote, executor.pushToken)
	}
}

func TestCommitUsesSuppliedIdentityForSingleCommand(t *testing.T) {
	executor := &pushExecutor{}
	service := NewService("/workspace", executor)

	if _, err := service.Commit(context.Background(), "sandbox-1", CommitRequest{Message: "update", All: true}, "Forge User", "forge@example.com"); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	joined := strings.Join(executor.commands, "\n")
	if !strings.Contains(joined, "'-c' 'user.name=Forge User'") || !strings.Contains(joined, "'-c' 'user.email=forge@example.com'") {
		t.Fatalf("commit command did not use supplied identity: %s", joined)
	}
	if strings.Contains(joined, "'config'") {
		t.Fatalf("commit identity should not be persisted: %s", joined)
	}
}

func TestPushRejectsNonGitHubRemote(t *testing.T) {
	executor := &pushExecutor{remoteURL: "https://example.com/acme/project.git"}
	service := NewService("/workspace", executor)

	if _, err := service.Push(context.Background(), "sandbox-1", PushRequest{}, "github-token"); err == nil {
		t.Fatal("Push() should reject a non-GitHub remote")
	}
	if executor.pushToken != "" {
		t.Fatal("push credentials were sent for a non-GitHub remote")
	}
}

type pushExecutor struct {
	remoteURL  string
	pushRemote string
	pushToken  string
	commands   []string
}

func (executor *pushExecutor) Execute(_ context.Context, _ string, command string) (terminalapp.ExecutionResult, error) {
	executor.commands = append(executor.commands, command)
	if strings.Contains(command, "'remote' 'get-url'") {
		return terminalapp.ExecutionResult{Output: executor.remoteURL}, nil
	}
	if strings.Contains(command, "'rev-parse' 'HEAD'") {
		return terminalapp.ExecutionResult{Output: "commit-hash"}, nil
	}
	return terminalapp.ExecutionResult{}, nil
}

func (executor *pushExecutor) PushRepository(_ context.Context, _ string, remote, _ string, accessToken string) error {
	executor.pushRemote = remote
	executor.pushToken = accessToken
	return nil
}
