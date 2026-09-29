package git

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	terminalapp "ai-agent/internal/modules/terminal/application"
)

const defaultWorkspaceRoot = "/workspace"

type SandboxExecutor interface {
	Execute(ctx context.Context, sandboxID, command string) (terminalapp.ExecutionResult, error)
}

type SandboxRepositoryPusher interface {
	PushRepository(ctx context.Context, sandboxID, remote, branch, accessToken string) error
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
	return &Service{workspaceRoot: path.Clean(root), executor: executor}
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

func (s *Service) Commit(ctx context.Context, sandboxID string, req CommitRequest, authorName, authorEmail string) (CommitResult, error) {
	message := strings.TrimSpace(req.Message)
	if message == "" {
		return CommitResult{}, fmt.Errorf("commit message is required")
	}
	authorName = strings.TrimSpace(authorName)
	authorEmail = strings.TrimSpace(authorEmail)
	if authorName == "" || authorEmail == "" {
		return CommitResult{}, fmt.Errorf("commit author name and email are required")
	}
	files, err := normalizeCommitPaths(req.Files)
	if err != nil {
		return CommitResult{}, err
	}
	if len(files) > 0 {
		addArgs := append([]string{"--literal-pathspecs", "add", "-A", "--"}, files...)
		if _, err := s.runGit(ctx, sandboxID, addArgs...); err != nil {
			return CommitResult{}, err
		}
	} else if req.All {
		if _, err := s.runGit(ctx, sandboxID, "add", "-A"); err != nil {
			return CommitResult{}, err
		}
	}
	commitArgs := []string{"--literal-pathspecs", "-c", "user.name=" + authorName, "-c", "user.email=" + authorEmail, "commit", "-m", message}
	if len(files) > 0 {
		commitArgs = append(commitArgs, "--")
		commitArgs = append(commitArgs, files...)
	}
	if _, err := s.runGit(ctx, sandboxID, commitArgs...); err != nil {
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

func normalizeCommitPaths(files []string) ([]string, error) {
	if len(files) == 0 {
		return nil, nil
	}

	normalized := make([]string, 0, len(files))
	seen := make(map[string]struct{}, len(files))
	for _, file := range files {
		cleaned := path.Clean(file)
		if strings.TrimSpace(file) == "" || strings.ContainsRune(file, '\x00') || path.IsAbs(file) || cleaned != file || cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
			return nil, fmt.Errorf("invalid commit path %q", file)
		}
		if _, exists := seen[file]; exists {
			continue
		}
		seen[file] = struct{}{}
		normalized = append(normalized, file)
	}
	return normalized, nil
}

func (s *Service) Push(ctx context.Context, sandboxID string, req PushRequest, accessToken string) (PushResult, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return PushResult{}, fmt.Errorf("GitHub account is not connected")
	}

	remote := strings.TrimSpace(req.Remote)
	if remote == "" {
		remote = "origin"
	}
	remoteURL, err := s.runGit(ctx, sandboxID, "remote", "get-url", remote)
	if err != nil {
		return PushResult{}, err
	}
	parsedRemote, err := url.Parse(strings.TrimSpace(remoteURL))
	if err != nil || !strings.EqualFold(parsedRemote.Scheme, "https") ||
		!strings.EqualFold(parsedRemote.Hostname(), "github.com") || parsedRemote.User != nil {
		return PushResult{}, fmt.Errorf("push requires an HTTPS GitHub remote")
	}

	pusher, ok := s.executor.(SandboxRepositoryPusher)
	if !ok {
		return PushResult{}, fmt.Errorf("sandbox executor does not support repository pushes")
	}
	if err := pusher.PushRepository(ctx, sandboxID, remote, req.Branch, accessToken); err != nil {
		return PushResult{}, fmt.Errorf("git push: %w", err)
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

	status := GitStatus{
		Modified:  make([]string, 0),
		Staged:    make([]string, 0),
		Untracked: make([]string, 0),
	}
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

func (s *Service) repoRoot() string {
	return s.workspaceRoot
}
