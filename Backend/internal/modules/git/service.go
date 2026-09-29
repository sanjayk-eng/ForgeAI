package git

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	terminalapp "ai-agent/internal/modules/terminal/application"
)

const defaultWorkspaceRoot = "/workspace"

type SandboxExecutor interface {
	Execute(ctx context.Context, sandboxID, command string) (terminalapp.ExecutionResult, error)
}

// Service executes git commands inside the sandbox workspace.
type Service struct {
	workspaceRoot string
	executor      SandboxExecutor
}

func NewService(workspaceRoot string, executor SandboxExecutor) *Service {
	root := strings.TrimSpace(workspaceRoot)
	if root == "" {
		root = defaultWorkspaceRoot
	}
	return &Service{workspaceRoot: filepath.Clean(root), executor: executor}
}

func (s *Service) Status(ctx context.Context, sandboxID string) (GitStatus, error) {
	output, err := s.runGit(ctx, sandboxID, "status", "--short", "--branch")
	if err != nil {
		return GitStatus{}, err
	}
	status, err := parseStatusOutput(output)
	if err != nil {
		return GitStatus{}, err
	}
	return status, nil
}

func (s *Service) Diff(ctx context.Context, sandboxID string) (GitDiffResult, error) {
	output, err := s.runGit(ctx, sandboxID, "diff", "--no-ext-diff", "--binary", "HEAD")
	if err != nil {
		return GitDiffResult{}, err
	}
	return GitDiffResult{Files: parseDiffOutput(output)}, nil
}

func (s *Service) Commit(ctx context.Context, sandboxID string, req CommitRequest) (CommitResult, error) {
	message := strings.TrimSpace(req.Message)
	if message == "" {
		return CommitResult{}, fmt.Errorf("commit message is required")
	}
	if req.All {
		if _, err := s.runGit(ctx, sandboxID, "add", "-A"); err != nil {
			return CommitResult{}, err
		}
	}
	if _, err := s.runGit(ctx, sandboxID, "commit", "-m", message); err != nil {
		return CommitResult{}, err
	}
	output, err := s.runGit(ctx, sandboxID, "rev-parse", "HEAD")
	if err != nil {
		return CommitResult{}, err
	}
	return CommitResult{
		Message: message,
		Hash:    strings.TrimSpace(output),
		Time:    time.Now().UTC(),
	}, nil
}

func (s *Service) Push(ctx context.Context, sandboxID string, req PushRequest) (PushResult, error) {
	args := []string{"push"}
	if strings.TrimSpace(req.Remote) != "" {
		args = append(args, req.Remote)
	}
	if strings.TrimSpace(req.Branch) != "" {
		args = append(args, req.Branch)
	}
	if _, err := s.runGit(ctx, sandboxID, args...); err != nil {
		return PushResult{}, err
	}
	return PushResult{Message: "push complete", Status: "ok"}, nil
}

func (s *Service) runGit(ctx context.Context, sandboxID string, args ...string) (string, error) {
	if sandboxID == "" {
		return "", fmt.Errorf("sandbox_id is required")
	}
	if s.executor == nil {
		return "", fmt.Errorf("sandbox executor is not configured")
	}
	command := "cd " + quoteShell(s.workspaceRoot) + " && git " + quoteShellArgs(args)
	result, err := s.executor.Execute(ctx, sandboxID, command)
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(result.Output))
	}
	if result.ExitCode != 0 {
		return "", fmt.Errorf("git %s: exit code %d: %s", strings.Join(args, " "), result.ExitCode, strings.TrimSpace(result.Output))
	}
	return result.Output, nil
}

func quoteShell(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func quoteShellArgs(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, quoteShell(arg))
	}
	return strings.Join(quoted, " ")
}

func parseStatusOutput(output string) (GitStatus, error) {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return GitStatus{}, fmt.Errorf("git status output is empty")
	}
	if strings.Contains(trimmed, "not a git repository") || strings.Contains(trimmed, "fatal:") {
		return GitStatus{}, fmt.Errorf("workspace is not a git repository")
	}

	status := GitStatus{}
	for _, rawLine := range strings.Split(trimmed, "\n") {
		line := strings.TrimRight(rawLine, "\r")
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "## ") {
			status.Branch = strings.TrimPrefix(line, "## ")
			if idx := strings.Index(status.Branch, "..."); idx >= 0 {
				tracking := status.Branch[idx+3:]
				status.Branch = strings.TrimSpace(status.Branch[:idx])
				if left, right, ok := strings.Cut(tracking, " ["); ok {
					_ = left
					tracking = strings.TrimSuffix(right, "]")
					for _, marker := range strings.Split(tracking, ", ") {
						parts := strings.SplitN(marker, " ", 2)
						if len(parts) != 2 {
							continue
						}
						count := 0
						if _, err := fmt.Sscanf(parts[1], "%d", &count); err != nil {
							continue
						}
						switch parts[0] {
						case "ahead":
							status.Ahead = count
						case "behind":
							status.Behind = count
						}
					}
				}
			}
			continue
		}
		if len(line) < 3 {
			continue
		}
		prefix := line[:2]
		fileName := strings.TrimSpace(line[3:])
		if strings.Contains(fileName, " -> ") {
			fileName = strings.TrimSpace(strings.SplitN(fileName, " -> ", 2)[1])
		}
		switch {
		case prefix == "??":
			status.Untracked = append(status.Untracked, fileName)
		default:
			if prefix[0] != ' ' {
				status.Staged = append(status.Staged, fileName)
			}
			if prefix[1] != ' ' {
				status.Modified = append(status.Modified, fileName)
			}
		}
	}
	status.IsDirty = len(status.Modified)+len(status.Staged)+len(status.Untracked) > 0
	return status, nil
}

func parseDiffOutput(output string) []GitDiffEntry {
	if strings.TrimSpace(output) == "" {
		return nil
	}
	files := make([]GitDiffEntry, 0)
	for _, block := range strings.Split(output, "diff --git ") {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		lines := strings.Split(block, "\n")
		if len(lines) < 2 {
			continue
		}
		header := strings.TrimSpace(lines[0])
		path := strings.TrimSpace(header)
		if strings.Contains(path, " b/") {
			path = strings.TrimSpace(strings.SplitN(path, " b/", 2)[1])
		}
		files = append(files, GitDiffEntry{Path: path, Content: block, Status: "modified"})
	}
	if len(files) == 0 {
		return nil
	}
	return files
}

func (s *Service) repoRoot() string {
	return s.workspaceRoot
}
