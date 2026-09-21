package executor
package executor

import (
	"os"
	"testing"

	"ai-agent/internal/shared/executor/spec"
)

func TestDetectWindowsShellRecognizesGitBash(t *testing.T) {
	t.Setenv("MSYSTEM", "MINGW64")
	t.Setenv("SHELL", "/usr/bin/bash")
	t.Setenv("PSModulePath", "C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\Modules")

	if got := detectWindowsShell(); got != spec.ShellBash {
		t.Fatalf("detectWindowsShell() = %q, want %q", got, spec.ShellBash)
	}
}

func TestDetectWindowsShellFallsBackToPowerShell(t *testing.T) {
	t.Setenv("MSYSTEM", "")
	t.Setenv("SHELL", "")
	t.Setenv("PSModulePath", "present")
	t.Setenv("COMSPEC", os.Getenv("COMSPEC"))

	if got := detectWindowsShell(); got != spec.ShellPowerShell {
		t.Fatalf("detectWindowsShell() = %q, want %q", got, spec.ShellPowerShell)
	}
}