package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type DockerCLI struct {
	binary string
}

func NewDockerCLI(binary string) *DockerCLI {
	if strings.TrimSpace(binary) == "" {
		binary = "docker"
	}
	return &DockerCLI{binary: binary}
}

func HostWorkspacePath(name string) string {
	root := filepath.Join(os.TempDir(), "forgeai-workspaces")
	_ = os.MkdirAll(root, 0o755)
	return filepath.Join(root, name)
}

func HostNPMCachePath(name string) string {
	root := filepath.Join(os.TempDir(), "forgeai-npm-cache")
	_ = os.MkdirAll(root, 0o755)
	return filepath.Join(root, name)
}

func (docker *DockerCLI) Run(ctx context.Context, args ...string) (string, error) {
	return docker.RunWithInput(ctx, nil, args...)
}

func (docker *DockerCLI) RunWithInput(ctx context.Context, input io.Reader, args ...string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("docker command is required")
	}
	command := exec.CommandContext(ctx, docker.binary, args...)
	command.Stdin = input
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		diagnostic := strings.TrimSpace(strings.Join([]string{stderr.String(), stdout.String()}, "\n"))
		return stdout.String(), fmt.Errorf("docker %s: %w: %s", args[0], err, diagnostic)
	}
	return stdout.String(), nil
}
