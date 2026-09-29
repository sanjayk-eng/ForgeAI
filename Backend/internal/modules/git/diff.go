package git

import (
	"context"
	"fmt"
	"path"
	"strings"
)

func (s *Service) Diff(ctx context.Context, sandboxID string, filePaths ...string) (GitDiffResult, error) {
	args := []string{"diff", "--no-ext-diff", "--binary", "--find-renames", "HEAD"}
	filePath := ""
	if len(filePaths) > 0 && strings.TrimSpace(filePaths[0]) != "" {
		filePath = path.Clean(strings.TrimSpace(filePaths[0]))
		if path.IsAbs(filePath) || filePath == "." || filePath == ".." || strings.HasPrefix(filePath, "../") || strings.Contains(filePath, "\\") {
			return GitDiffResult{}, fmt.Errorf("invalid workspace file path")
		}
		args = append(args, "--", filePath)
	}
	output, err := s.runGit(ctx, sandboxID, args...)
	if err != nil {
		return GitDiffResult{}, err
	}
	files := parseDiffOutput(output)
	if filePath == "" {
		return GitDiffResult{Files: files}, nil
	}

	for index := range files {
		if files[index].Path != filePath && files[index].OldPath != filePath {
			continue
		}
		basePath := files[index].OldPath
		if basePath == "" {
			basePath = files[index].Path
		}
		original, showErr := s.runGit(ctx, sandboxID, "show", "HEAD:"+basePath)
		if showErr == nil {
			files[index].OriginalContent = &original
		}
		return GitDiffResult{Files: files}, nil
	}

	original, showErr := s.runGit(ctx, sandboxID, "show", "HEAD:"+filePath)
	if showErr == nil {
		files = append(files, GitDiffEntry{
			Path: filePath, Status: "unchanged", OriginalContent: &original,
		})
	}
	return GitDiffResult{Files: files}, nil
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
		if len(lines) == 0 {
			continue
		}
		file := GitDiffEntry{Path: diffHeaderPath(lines[0]), Content: block, Status: "modified"}
		for _, line := range lines[1:] {
			switch {
			case strings.HasPrefix(line, "rename from "):
				file.OldPath = strings.Trim(strings.TrimPrefix(line, "rename from "), "\"")
				file.Status = "renamed"
			case strings.HasPrefix(line, "rename to "):
				file.Path = strings.Trim(strings.TrimPrefix(line, "rename to "), "\"")
			case strings.HasPrefix(line, "new file mode "):
				file.Status = "added"
			case strings.HasPrefix(line, "deleted file mode "):
				file.Status = "deleted"
			}
		}
		files = append(files, file)
	}
	return files
}

func diffHeaderPath(header string) string {
	if _, newPath, found := strings.Cut(header, " b/"); found {
		return strings.TrimSpace(newPath)
	}
	if _, oldPath, found := strings.Cut(header, "a/"); found {
		return strings.TrimSpace(oldPath)
	}
	return strings.TrimSpace(header)
}
