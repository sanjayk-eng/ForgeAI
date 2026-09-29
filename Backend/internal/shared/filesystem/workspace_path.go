package filesystem

import (
	"errors"
	"path"
	"strings"
)

var ErrInvalidWorkspacePath = errors.New("invalid workspace path")

func ResolveWorkspacePath(workspaceRoot, rawPath string) (string, error) {
	root := path.Clean(strings.TrimSpace(workspaceRoot))
	if !path.IsAbs(root) || root == "/" {
		return "", ErrInvalidWorkspacePath
	}

	candidate := strings.TrimSpace(strings.ReplaceAll(rawPath, "\\", "/"))
	if candidate == "" {
		return "", ErrInvalidWorkspacePath
	}
	if strings.HasPrefix(candidate, root+"/") {
		candidate = strings.TrimPrefix(candidate, root+"/")
	} else if candidate == root || strings.HasPrefix(candidate, "/") || strings.HasPrefix(candidate, "~") ||
		(len(candidate) >= 2 && candidate[1] == ':') {
		return "", ErrInvalidWorkspacePath
	}

	candidate = strings.TrimPrefix(candidate, "./")
	candidate = strings.Trim(candidate, "/")
	if candidate == "" {
		return "", ErrInvalidWorkspacePath
	}
	for _, segment := range strings.Split(candidate, "/") {
		if segment == "" || segment == "." || segment == ".." || strings.EqualFold(segment, ".git") {
			return "", ErrInvalidWorkspacePath
		}
	}

	resolved := path.Join(root, candidate)
	if resolved == root || !strings.HasPrefix(resolved, root+"/") {
		return "", ErrInvalidWorkspacePath
	}
	return resolved, nil
}
