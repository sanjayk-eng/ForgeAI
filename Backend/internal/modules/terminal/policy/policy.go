package policy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"ai-agent/internal/modules/terminal/domain"
)

type File struct {
	Sandbox Sandbox `yaml:"sandbox"`
}

type Sandbox struct {
	Image             string        `yaml:"image"`
	WorkspacePath     string        `yaml:"workspace_path"`
	NetworkMode       string        `yaml:"network_mode"`
	ReadOnlyRootFS    bool          `yaml:"read_only_rootfs"`
	NoNewPrivileges   bool          `yaml:"no_new_privileges"`
	CapDrop           []string      `yaml:"cap_drop"`
	CommandTimeout    time.Duration `yaml:"-"`
	CommandTimeoutRaw string        `yaml:"command_timeout"`
	Limits            Limits        `yaml:"limits"`
}

type Limits struct {
	CPUShares   int64 `yaml:"cpu_shares"`
	MemoryBytes int64 `yaml:"memory_bytes"`
	PidsLimit   int64 `yaml:"pids_limit"`
}

func Load(path string) (Sandbox, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Sandbox{}, fmt.Errorf("read sandbox policy %s: %w", path, err)
	}
	var file File
	if err := yaml.Unmarshal(data, &file); err != nil {
		return Sandbox{}, fmt.Errorf("decode sandbox policy: %w", err)
	}
	if err := file.Sandbox.validate(); err != nil {
		return Sandbox{}, err
	}
	return file.Sandbox, nil
}

func LoadFromWorkspace(workspace, path string) (Sandbox, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(workspace, path)
	}
	return Load(path)
}

func (sandbox Sandbox) ResourceLimits() domain.ResourceLimits {
	return domain.ResourceLimits{
		CPUShares:   sandbox.Limits.CPUShares,
		MemoryBytes: sandbox.Limits.MemoryBytes,
		PidsLimit:   sandbox.Limits.PidsLimit,
	}
}

func (sandbox *Sandbox) validate() error {
	if strings.TrimSpace(sandbox.Image) == "" || strings.TrimSpace(sandbox.WorkspacePath) == "" {
		return fmt.Errorf("sandbox policy image and workspace_path are required")
	}
	if sandbox.NetworkMode == "" {
		return fmt.Errorf("sandbox policy network_mode is required")
	}
	if sandbox.CommandTimeoutRaw == "" {
		return fmt.Errorf("sandbox policy command_timeout is required")
	}
	parsed, err := time.ParseDuration(sandbox.CommandTimeoutRaw)
	if err != nil || parsed <= 0 {
		return fmt.Errorf("sandbox policy command_timeout must be a positive duration")
	}
	if sandbox.Limits.CPUShares <= 0 || sandbox.Limits.MemoryBytes <= 0 || sandbox.Limits.PidsLimit <= 0 {
		return fmt.Errorf("sandbox policy resource limits must be positive")
	}
	sandbox.CommandTimeout = parsed
	return nil
}
