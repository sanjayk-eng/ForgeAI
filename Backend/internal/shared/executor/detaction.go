package executor

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"ai-agent/internal/shared/executor/spec"
)

func DetectShell() spec.Shell {
	switch runtime.GOOS {
	case "windows":
		return detectWindowsShell()

	case "linux", "darwin":
		return detectUnixShell()

	default:
		return ""
	}
}

func detectWindowsShell() spec.Shell {
	if strings.HasPrefix(strings.ToUpper(os.Getenv("MSYSTEM")), "MINGW") || filepath.Base(os.Getenv("SHELL")) == "bash" {
		return spec.ShellBash
	}

	if os.Getenv("PSModulePath") != "" {
		return spec.ShellPowerShell
	}

	if os.Getenv("COMSPEC") != "" {
		return spec.ShellCMD
	}

	return ""
}

func detectUnixShell() spec.Shell {
	currentShell := filepath.Base(os.Getenv("SHELL"))

	switch currentShell {
	case "zsh":
		return spec.ShellZsh

	case "bash":
		return spec.ShellBash

	default:
		if runtime.GOOS == "darwin" {
			return spec.ShellZsh
		}

		return spec.ShellBash
	}
}
