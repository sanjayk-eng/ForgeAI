package git

import (
	"context"
	"errors"
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

func TestHistoryReturnsAllCommitDetails(t *testing.T) {
	executor := &historyExecutor{output: "abc123\x00parent1 parent2\x00Forge User\x001791567200\x00HEAD -> main, origin/main\x00Merge feature\n" +
		"parent1\x00\x00Forge User\x001791567100\x00\x00Add feature\n"}
	service := NewService("/workspace", executor)

	commits, err := service.History(context.Background(), "sandbox-1")
	if err != nil {
		t.Fatalf("History() error = %v", err)
	}
	if len(commits) != 2 || commits[0].Hash != "abc123" || len(commits[0].Parents) != 2 ||
		commits[0].Author != "Forge User" || commits[0].Subject != "Merge feature" ||
		len(commits[0].Refs) != 2 || commits[0].Refs[0] != "HEAD -> main" || commits[0].Refs[1] != "origin/main" {
		t.Fatalf("History() = %#v, want commit details and branch references", commits)
	}
	if commits[0].Time.Unix() != 1791567200 {
		t.Fatalf("History() timestamp = %v, want unix timestamp 1791567200", commits[0].Time)
	}
}

func TestHistoryReturnsEmptyListForUnbornRepository(t *testing.T) {
	service := NewService("/workspace", &historyExecutor{})
	commits, err := service.History(context.Background(), "sandbox-1")
	if err != nil {
		t.Fatalf("History() error = %v", err)
	}
	if commits == nil || len(commits) != 0 {
		t.Fatalf("History() = %#v, want an empty list", commits)
	}
}

func TestStatusAndDiffAcceptWorkspaceWithoutInitialCommit(t *testing.T) {
	executor := emptyRepositoryExecutor{}
	service := NewService("/workspace", &executor)

	status, err := service.Status(context.Background(), "sandbox-1")
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.Branch != "main" || len(status.Untracked) != 2 ||
		status.Untracked[0] != "sanjaykh/README.md" || status.Untracked[1] != "sanjaykh/index.html" {
		t.Fatalf("Status() = %+v, want main branch with each nested project file untracked", status)
	}

	diff, err := service.Diff(context.Background(), "sandbox-1")
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}
	if len(diff.Files) != 0 {
		t.Fatalf("Diff() files = %#v, want no diff before the initial commit", diff.Files)
	}
	if len(executor.commands) < 2 || !strings.Contains(executor.commands[0], "'--untracked-files=all'") {
		t.Fatalf("Status() commands = %#v, want untracked files listed individually", executor.commands)
	}
}

func TestDiffForUntrackedFileBeforeInitialCommit(t *testing.T) {
	executor := &emptyRepositoryExecutor{}
	service := NewService("/workspace", executor)

	result, err := service.Diff(context.Background(), "sandbox-1", "sanjaykh/README.md")
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}
	if len(result.Files) != 1 || result.Files[0].Path != "sanjaykh/README.md" || result.Files[0].Status != "added" {
		t.Fatalf("Diff() files = %#v, want an added diff for the untracked file", result.Files)
	}
	if result.Files[0].OriginalContent == nil || *result.Files[0].OriginalContent != "" {
		t.Fatalf("Diff() original content = %#v, want empty baseline", result.Files[0].OriginalContent)
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

func TestDiffForFileReturnsGitBaseContentAndScopesPath(t *testing.T) {
	executor := &fileDiffExecutor{}
	service := NewService("/workspace", executor)

	result, err := service.Diff(context.Background(), "sandbox-1", "src/App.tsx")
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}
	if len(result.Files) != 1 || result.Files[0].Path != "src/App.tsx" {
		t.Fatalf("Diff() files = %#v, want only src/App.tsx", result.Files)
	}
	if result.Files[0].OriginalContent == nil || *result.Files[0].OriginalContent != "<h1>Hello</h1>\n" {
		t.Fatalf("Diff() original content = %#v, want Git base content", result.Files[0].OriginalContent)
	}
	if len(executor.commands) != 3 ||
		!strings.Contains(executor.commands[1], "'--' 'src/App.tsx'") ||
		!strings.Contains(executor.commands[2], "'show' 'HEAD:src/App.tsx'") {
		t.Fatalf("Diff() commands = %#v, want a scoped diff and Git base read", executor.commands)
	}
}

func TestDiffForTrackedFileReturnsBaseEvenWhenPatchIsEmpty(t *testing.T) {
	executor := &fileDiffExecutor{unchanged: true}
	service := NewService("/workspace", executor)

	result, err := service.Diff(context.Background(), "sandbox-1", "src/App.tsx")
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}
	if len(result.Files) != 1 || result.Files[0].Status != "unchanged" {
		t.Fatalf("Diff() files = %#v, want the tracked file baseline", result.Files)
	}
	if result.Files[0].OriginalContent == nil || *result.Files[0].OriginalContent != "<h1>Hello</h1>\n" {
		t.Fatalf("Diff() original content = %#v, want Git base content", result.Files[0].OriginalContent)
	}
}

func TestDiffForNewUntrackedFileReturnsSyntheticDiff(t *testing.T) {
	executor := &newFileDiffExecutor{}
	service := NewService("/workspace", executor)

	result, err := service.Diff(context.Background(), "sandbox-1", "notes.txt")
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}
	if len(result.Files) != 1 || result.Files[0].Status != "added" {
		t.Fatalf("Diff() files = %#v, want synthetic added diff for notes.txt", result.Files)
	}
	if result.Files[0].Path != "notes.txt" {
		t.Fatalf("Diff() path = %q, want notes.txt", result.Files[0].Path)
	}
	if result.Files[0].OriginalContent == nil || *result.Files[0].OriginalContent != "" {
		t.Fatalf("Diff() original content = %#v, want empty original for a newly created file", result.Files[0].OriginalContent)
	}
}

func TestDiffForRenamedFileReadsOriginalFromOldPath(t *testing.T) {
	executor := &renamedFileDiffExecutor{}
	service := NewService("/workspace", executor)

	result, err := service.Diff(context.Background(), "sandbox-1", "src/App.tsx")
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}
	if len(result.Files) != 1 || result.Files[0].Path != "src/App.tsx" || result.Files[0].OldPath != "src/App.ts" {
		t.Fatalf("Diff() files = %#v, want src/App.ts renamed to src/App.tsx", result.Files)
	}
	if result.Files[0].OriginalContent == nil || *result.Files[0].OriginalContent != "export const App = () => null;\n" {
		t.Fatalf("Diff() original content = %#v, want content from old Git path", result.Files[0].OriginalContent)
	}
	if len(executor.commands) != 3 || !strings.Contains(executor.commands[2], "'show' 'HEAD:src/App.ts'") {
		t.Fatalf("Diff() commands = %#v, want the renamed file's old Git path", executor.commands)
	}
}

func TestDiffRejectsPathsOutsideWorkspace(t *testing.T) {
	executor := &fileDiffExecutor{}
	service := NewService("/workspace", executor)
	if _, err := service.Diff(context.Background(), "sandbox-1", "../outside.tsx"); err == nil {
		t.Fatal("Diff() accepted a path outside the workspace")
	}
	if len(executor.commands) != 0 {
		t.Fatalf("Diff() executed commands for an invalid path: %#v", executor.commands)
	}
}

func TestPushPassesGitHubTokenThroughStdin(t *testing.T) {
	executor := &pushExecutor{remoteURL: "https://github.com/acme/project.git", currentBranch: "main"}
	service := NewService("/workspace", executor)

	if _, err := service.Push(context.Background(), "sandbox-1", PushRequest{}, "github-token", "https://github.com/acme/project.git"); err != nil {
		t.Fatalf("Push() error = %v", err)
	}
	if executor.pushRemote != "origin" || executor.pushBranch != "main" || executor.pushToken != "github-token" {
		t.Fatalf("push target/branch/token = %q/%q/%q, want origin/main/github-token",
			executor.pushRemote, executor.pushBranch, executor.pushToken)
	}
	if len(executor.commands) != 2 || !strings.Contains(executor.commands[1], "'branch' '--show-current'") {
		t.Fatalf("Push() commands = %#v, want current branch resolved before pushing", executor.commands)
	}
}

func TestPushUsesRequestedBranchWithoutResolvingCurrentBranch(t *testing.T) {
	executor := &pushExecutor{remoteURL: "https://github.com/acme/project.git"}
	service := NewService("/workspace", executor)

	if _, err := service.Push(context.Background(), "sandbox-1", PushRequest{Branch: "release"}, "github-token", "https://github.com/acme/project.git"); err != nil {
		t.Fatalf("Push() error = %v", err)
	}
	if executor.pushBranch != "release" {
		t.Fatalf("pushed branch = %q, want release", executor.pushBranch)
	}
	if len(executor.commands) != 1 {
		t.Fatalf("Push() commands = %#v, want no current-branch lookup", executor.commands)
	}
}

func TestPushRejectsDetachedHeadWithoutRequestedBranch(t *testing.T) {
	executor := &pushExecutor{remoteURL: "https://github.com/acme/project.git"}
	service := NewService("/workspace", executor)

	if _, err := service.Push(context.Background(), "sandbox-1", PushRequest{}, "github-token", "https://github.com/acme/project.git"); err == nil ||
		!strings.Contains(err.Error(), "HEAD is detached") {
		t.Fatalf("Push() error = %v, want detached HEAD error", err)
	}
	if executor.pushToken != "" {
		t.Fatal("push credentials were sent without a branch")
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

func TestCommitOnlyStagesAndCommitsSelectedFiles(t *testing.T) {
	executor := &pushExecutor{}
	service := NewService("/workspace", executor)
	selected := []string{"Backend/internal/config/env.go", "Backend/internal/config/validation.go"}

	if _, err := service.Commit(context.Background(), "sandbox-1", CommitRequest{Message: "selected config files", Files: selected}, "Forge User", "forge@example.com"); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	if len(executor.commands) != 3 {
		t.Fatalf("Commit() commands = %#v, want add, commit, and rev-parse", executor.commands)
	}
	for _, command := range executor.commands[:2] {
		for _, file := range selected {
			if !strings.Contains(command, "'"+file+"'") {
				t.Fatalf("selected file %q missing from command: %s", file, command)
			}
		}
		if strings.Contains(command, "config_test.go") {
			t.Fatalf("unselected file was included in command: %s", command)
		}
	}
}

func TestCommitRejectsPathsOutsideWorkspace(t *testing.T) {
	executor := &pushExecutor{}
	service := NewService("/workspace", executor)
	if _, err := service.Commit(context.Background(), "sandbox-1", CommitRequest{Message: "bad path", Files: []string{"../outside.go"}}, "Forge User", "forge@example.com"); err == nil {
		t.Fatal("Commit() accepted a path outside the workspace")
	}
	if len(executor.commands) != 0 {
		t.Fatalf("Commit() executed commands for an invalid path: %#v", executor.commands)
	}
}

func TestRevertRestoresFilesTrackedInHead(t *testing.T) {
	executor := &pushExecutor{}
	service := NewService("/workspace", executor)
	selected := []string{"Backend/internal/config/env.go", "Backend/internal/config/validation.go", "internal/scheduler/method.go"}

	if err := service.Revert(context.Background(), "sandbox-1", RevertRequest{Files: selected}); err != nil {
		t.Fatalf("Revert() error = %v", err)
	}
	if len(executor.commands) != 6 {
		t.Fatalf("Revert() commands = %#v, want HEAD lookup plus restore for each selected file", executor.commands)
	}
	joined := strings.Join(executor.commands, "\n")
	for _, file := range selected {
		if !strings.Contains(joined, "'"+file+"'") {
			t.Fatalf("selected file %q missing from revert command: %s", file, joined)
		}
		if !strings.Contains(joined, "'ls-tree' '-r' '--name-only' 'HEAD'") {
			t.Fatalf("revert did not check whether %q exists in HEAD: %s", file, joined)
		}
		if !strings.Contains(joined, "'restore' '--source=HEAD' '--staged' '--worktree' '--' '"+file+"'") {
			t.Fatalf("tracked file %q was not restored with HEAD source: %s", file, joined)
		}
	}
	if strings.Contains(joined, "'.'") {
		t.Fatalf("selected-file revert used workspace-wide restore: %s", joined)
	}
}

func TestRevertRemovesUntrackedNewFiles(t *testing.T) {
	executor := &pushExecutor{}
	service := NewService("/workspace", executor)

	if err := service.Revert(context.Background(), "sandbox-1", RevertRequest{Files: []string{"notes.txt"}}); err != nil {
		t.Fatalf("Revert() error = %v", err)
	}
	joined := strings.Join(executor.commands, "\n")
	if !strings.Contains(joined, "'ls-tree' '-r' '--name-only' 'HEAD'") {
		t.Fatalf("untracked revert should check whether the file exists in HEAD: %s", joined)
	}
	if !strings.Contains(joined, "'rm' '--cached' '--ignore-unmatch' '-f'") || !strings.Contains(joined, "'clean'") {
		t.Fatalf("reverting a new file should remove it from the workspace: %s", joined)
	}
}

func TestIsTrackedFileReturnsFalseWhenAbsentFromHead(t *testing.T) {
	service := NewService("/workspace", &untrackedFileExecutor{})
	tracked, err := service.isTrackedFile(context.Background(), "sandbox-1", "demo1.go")
	if err != nil {
		t.Fatalf("isTrackedFile() error = %v", err)
	}
	if tracked {
		t.Fatal("isTrackedFile() incorrectly marked a file absent from HEAD as tracked")
	}
}

func TestRevertRejectsPathsOutsideWorkspace(t *testing.T) {
	executor := &pushExecutor{}
	service := NewService("/workspace", executor)
	if err := service.Revert(context.Background(), "sandbox-1", RevertRequest{Files: []string{"../outside.go"}}); err == nil {
		t.Fatal("Revert() accepted a path outside the workspace")
	}
	if len(executor.commands) != 0 {
		t.Fatalf("Revert() executed commands for an invalid path: %#v", executor.commands)
	}
}

func TestPushRejectsNonGitHubRemote(t *testing.T) {
	executor := &pushExecutor{remoteURL: "https://example.com/acme/project.git"}
	service := NewService("/workspace", executor)

	if _, err := service.Push(context.Background(), "sandbox-1", PushRequest{}, "github-token", "https://example.com/acme/project.git"); err == nil {
		t.Fatal("Push() should reject a non-GitHub remote")
	}
	if executor.pushToken != "" {
		t.Fatal("push credentials were sent for a non-GitHub remote")
	}
}

type pushExecutor struct {
	remoteURL     string
	currentBranch string
	pushRemote    string
	pushBranch    string
	pushToken     string
	commands      []string
}

type emptyRepositoryExecutor struct {
	commands []string
}

type historyExecutor struct {
	output string
}

func (executor *historyExecutor) Execute(_ context.Context, _ string, command string) (terminalapp.ExecutionResult, error) {
	if !strings.Contains(command, "'log'") {
		return terminalapp.ExecutionResult{}, errors.New("unexpected Git command")
	}
	return terminalapp.ExecutionResult{Output: executor.output}, nil
}

func (executor *emptyRepositoryExecutor) Execute(_ context.Context, _ string, command string) (terminalapp.ExecutionResult, error) {
	executor.commands = append(executor.commands, command)
	if strings.Contains(command, "'status'") {
		return terminalapp.ExecutionResult{Output: "## No commits yet on main\n?? sanjaykh/README.md\n?? sanjaykh/index.html\n"}, nil
	}
	if strings.Contains(command, "'rev-parse' '--verify' '--quiet' 'HEAD'") {
		return terminalapp.ExecutionResult{
			ExitCode: 1,
		}, errors.New("docker exec: exit status 1")
	}
	if strings.Contains(command, "'diff' '--no-ext-diff' '--binary' '--no-index'") {
		if strings.Contains(command, "'sanjaykh/README.md'") {
			return terminalapp.ExecutionResult{
				Output:   "diff --git a/dev/null b/sanjaykh/README.md\nnew file mode 100644\n--- /dev/null\n+++ b/sanjaykh/README.md\n@@ -0,0 +1 @@\n+# project\n",
				ExitCode: 1,
			}, errors.New("docker exec: exit status 1")
		}
	}
	return terminalapp.ExecutionResult{}, nil
}

type untrackedFileExecutor struct{}

func (executor *untrackedFileExecutor) Execute(_ context.Context, _ string, command string) (terminalapp.ExecutionResult, error) {
	if strings.Contains(command, "'ls-tree'") {
		return terminalapp.ExecutionResult{Output: "", ExitCode: 0}, nil
	}
	return terminalapp.ExecutionResult{Output: "", ExitCode: 0}, nil
}

type fileDiffExecutor struct {
	commands  []string
	unchanged bool
}

type renamedFileDiffExecutor struct {
	commands []string
}

type newFileDiffExecutor struct {
	commands []string
}

func (executor *newFileDiffExecutor) Execute(_ context.Context, _ string, command string) (terminalapp.ExecutionResult, error) {
	executor.commands = append(executor.commands, command)
	if strings.Contains(command, "'rev-parse' '--verify' '--quiet' 'HEAD'") {
		return terminalapp.ExecutionResult{Output: "commit-hash\n"}, nil
	}
	if strings.Contains(command, "'diff'") {
		return terminalapp.ExecutionResult{Output: "diff --git a/notes.txt b/notes.txt\nnew file mode 100644\nindex 0000000..1111111\n--- /dev/null\n+++ b/notes.txt\n@@ -0,0 +1 @@\n+hello\n"}, nil
	}
	if strings.Contains(command, "'show' 'HEAD:notes.txt'") {
		return terminalapp.ExecutionResult{Output: "fatal: invalid object name HEAD:notes.txt", ExitCode: 128}, nil
	}
	if strings.Contains(command, "'diff' '--' '--' 'notes.txt'") {
		return terminalapp.ExecutionResult{Output: "diff --git a/notes.txt b/notes.txt\nnew file mode 100644\nindex 0000000..1111111\n--- /dev/null\n+++ b/notes.txt\n@@ -0,0 +1 @@\n+hello\n"}, nil
	}
	return terminalapp.ExecutionResult{}, nil
}

func (executor *renamedFileDiffExecutor) Execute(_ context.Context, _ string, command string) (terminalapp.ExecutionResult, error) {
	executor.commands = append(executor.commands, command)
	if strings.Contains(command, "'rev-parse' '--verify' '--quiet' 'HEAD'") {
		return terminalapp.ExecutionResult{Output: "commit-hash\n"}, nil
	}
	if strings.Contains(command, "'diff'") {
		return terminalapp.ExecutionResult{Output: "diff --git a/src/App.ts b/src/App.tsx\nsimilarity index 95%\nrename from src/App.ts\nrename to src/App.tsx\n"}, nil
	}
	if strings.Contains(command, "'show' 'HEAD:src/App.ts'") {
		return terminalapp.ExecutionResult{Output: "export const App = () => null;\n"}, nil
	}
	return terminalapp.ExecutionResult{}, nil
}

func (executor *fileDiffExecutor) Execute(_ context.Context, _ string, command string) (terminalapp.ExecutionResult, error) {
	executor.commands = append(executor.commands, command)
	if strings.Contains(command, "'rev-parse' '--verify' '--quiet' 'HEAD'") {
		return terminalapp.ExecutionResult{Output: "commit-hash\n"}, nil
	}
	if strings.Contains(command, "'diff'") {
		if executor.unchanged {
			return terminalapp.ExecutionResult{}, nil
		}
		return terminalapp.ExecutionResult{Output: "diff --git a/src/App.tsx b/src/App.tsx\nindex 1..2 100644\n--- a/src/App.tsx\n+++ b/src/App.tsx\n@@ -1 +1 @@\n-<h1>Hello</h1>\n+<h1>Welcome</h1>\n"}, nil
	}
	if strings.Contains(command, "'show' 'HEAD:src/App.tsx'") {
		return terminalapp.ExecutionResult{Output: "<h1>Hello</h1>\n"}, nil
	}
	return terminalapp.ExecutionResult{}, nil
}

func (executor *pushExecutor) Execute(_ context.Context, _ string, command string) (terminalapp.ExecutionResult, error) {
	executor.commands = append(executor.commands, command)
	if strings.Contains(command, "'remote' 'get-url'") {
		return terminalapp.ExecutionResult{Output: executor.remoteURL}, nil
	}
	if strings.Contains(command, "'branch' '--show-current'") {
		return terminalapp.ExecutionResult{Output: executor.currentBranch}, nil
	}
	if strings.Contains(command, "'rev-parse' 'HEAD'") {
		return terminalapp.ExecutionResult{Output: "commit-hash"}, nil
	}
	if strings.Contains(command, "'ls-tree' '-r' '--name-only' 'HEAD'") {
		start := strings.Index(command, "'--' '")
		if start >= 0 {
			rest := command[start+len("'--' '"):]
			if end := strings.Index(rest, "'"); end >= 0 {
				file := rest[:end]
				if file == "notes.txt" {
					return terminalapp.ExecutionResult{ExitCode: 0}, nil
				}
				return terminalapp.ExecutionResult{Output: file + "\n"}, nil
			}
		}
		return terminalapp.ExecutionResult{Output: "tracked.txt\n"}, nil
	}
	return terminalapp.ExecutionResult{}, nil
}

func (executor *pushExecutor) PushRepository(_ context.Context, _ string, remote, branch, accessToken string) error {
	executor.pushRemote = remote
	executor.pushBranch = branch
	executor.pushToken = accessToken
	return nil
}
