package agent

import (
	"fmt"
	"strings"

	"ai-agent/internal/shared/filesystem"
)

const (
	maxChangedFiles       = 20
	maxChangedFileLength  = 200_000
	maxTotalChangesLength = 512_000
)

func validateChanges(changes []fileChange) ([]fileChange, error) {
	if len(changes) > maxChangedFiles {
		return nil, fmt.Errorf("%w: too many changed files", ErrModelResponse)
	}
	validated := make([]fileChange, 0, len(changes))
	seen := make(map[string]struct{}, len(changes))
	totalLength := 0
	for _, change := range changes {
		relativePath, err := validateChangePath(change.Path)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid changed file path", ErrModelResponse)
		}
		if len(change.Content) > maxChangedFileLength {
			return nil, fmt.Errorf("%w: changed file exceeds the size limit", ErrModelResponse)
		}
		if _, exists := seen[relativePath]; exists {
			return nil, fmt.Errorf("%w: duplicate changed file path", ErrModelResponse)
		}
		seen[relativePath] = struct{}{}
		totalLength += len(change.Content)
		if totalLength > maxTotalChangesLength {
			return nil, fmt.Errorf("%w: total changes exceed the size limit", ErrModelResponse)
		}
		validated = append(validated, fileChange{Path: relativePath, Content: change.Content})
	}
	return validated, nil
}

func validateChangePath(filePath string) (string, error) {
	fullPath, err := filesystem.ResolveWorkspacePath("/workspace", filePath)
	if err != nil {
		return "", fmt.Errorf("invalid path")
	}
	relativePath := strings.TrimPrefix(fullPath, "/workspace/")
	if len(relativePath) > 300 {
		return "", fmt.Errorf("invalid path")
	}
	for _, segment := range strings.Split(relativePath, "/") {
		if skipContextEntry(segment, false) {
			return "", fmt.Errorf("protected path")
		}
	}
	return relativePath, nil
}
