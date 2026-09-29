package agent

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	terminalapp "ai-agent/internal/modules/terminal/application"
	"ai-agent/internal/shared/filesystem"
)

const (
	maxContextLength     = 80_000
	maxContextFiles      = 50
	maxFileContextLength = 8_000
)

func (service *Service) projectContext(ctx context.Context, sandboxID string) (string, error) {
	directories := []string{"/workspace"}
	files := make([]terminalapp.FileEntry, 0, maxContextFiles)
	for index := 0; index < len(directories) && len(files) < maxContextFiles; index++ {
		entries, err := service.sandbox.ListFiles(ctx, sandboxID, directories[index])
		if err != nil {
			return "", err
		}
		sort.Slice(entries, func(left, right int) bool { return entries[left].Name < entries[right].Name })
		for _, entry := range entries {
			if skipContextEntry(entry.Name, entry.IsDirectory) {
				continue
			}
			if entry.IsDirectory {
				if len(directories) < maxContextFiles {
					directories = append(directories, entry.Path)
				}
				continue
			}
			if !validWorkspacePath(entry.Path) {
				continue
			}
			files = append(files, entry)
			if len(files) >= maxContextFiles {
				break
			}
		}
	}

	var contextBuilder strings.Builder
	for _, file := range files {
		remaining := maxContextLength - contextBuilder.Len()
		if remaining <= 0 {
			break
		}
		content, err := service.sandbox.ReadFile(ctx, sandboxID, file.Path)
		if err != nil || !utf8.ValidString(content) {
			continue
		}
		if len(content) > maxFileContextLength {
			const marker = "\n[truncated]"
			content = truncateUTF8(content, maxFileContextLength-len(marker)) + marker
		}
		if len(content) > remaining {
			content = content[:remaining]
			for !utf8.ValidString(content) && len(content) > 0 {
				content = content[:len(content)-1]
			}
		}
		relativePath := strings.TrimPrefix(file.Path, "/workspace/")
		fmt.Fprintf(&contextBuilder, "\n--- FILE: %s ---\n%s\n", relativePath, content)
	}
	if contextBuilder.Len() == 0 {
		return "No readable source files were found in /workspace.", nil
	}
	return contextBuilder.String(), nil
}

func validWorkspacePath(filePath string) bool {
	_, err := filesystem.ResolveWorkspacePath("/workspace", filePath)
	return err == nil
}

func skipContextEntry(name string, isDirectory bool) bool {
	base := strings.ToLower(strings.TrimSpace(name))
	if base == "" || base == ".git" || base == "node_modules" || base == "vendor" || base == ".next" || base == "dist" || base == "build" || base == "target" || base == ".venv" || base == "coverage" {
		return true
	}
	if strings.HasPrefix(base, ".env") || strings.Contains(base, "secret") || strings.Contains(base, "credential") {
		return true
	}
	if isDirectory {
		return false
	}
	return strings.HasSuffix(base, ".pem") || strings.HasSuffix(base, ".key") || strings.HasSuffix(base, ".p12") || strings.HasSuffix(base, ".pfx")
}

func truncateUTF8(value string, maxBytes int) string {
	if len(value) <= maxBytes {
		return value
	}
	cut := maxBytes
	for cut > 0 && !utf8.RuneStart(value[cut]) {
		cut--
	}
	return value[:cut]
}
