package infrastructure

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"ai-agent/internal/modules/terminal/domain"
)

type ContainerSpec struct {
	Name            string
	Image           string
	VolumeName      string
	WorkspacePath   string
	NetworkMode     string
	ReadOnlyRootFS  bool
	NoNewPrivileges bool
	CapDrop         []string
	Limits          domain.ResourceLimits
}

type ExecutionResult struct {
	Output   string
	ExitCode int
	Command  string
}

type Runtime interface {
	CreateVolume(ctx context.Context, name string) error
	RemoveVolume(ctx context.Context, name string) error
	CreateContainer(ctx context.Context, spec ContainerSpec) (string, error)
	StartContainer(ctx context.Context, containerID string) error
	StopContainer(ctx context.Context, containerID string) error
	RemoveContainer(ctx context.Context, containerID string) error
	Execute(ctx context.Context, containerID, command string, timeout time.Duration) (ExecutionResult, error)
}

type DockerRuntime struct {
	binary string
}

func NewDockerRuntime(binary string) *DockerRuntime {
	if strings.TrimSpace(binary) == "" {
		binary = "docker"
	}
	return &DockerRuntime{binary: binary}
}

func (runtime *DockerRuntime) CreateVolume(ctx context.Context, name string) error {
	_, err := runtime.run(ctx, "volume", "create", "--label", "com.forgeai.managed=true", name)
	return err
}

func (runtime *DockerRuntime) RemoveVolume(ctx context.Context, name string) error {
	_, err := runtime.run(ctx, "volume", "rm", "--force", name)
	return err
}

func (runtime *DockerRuntime) CreateContainer(ctx context.Context, spec ContainerSpec) (string, error) {
	args := []string{"create", "--name", spec.Name, "--label", "com.forgeai.managed=true"}
	args = append(args, "--mount", "type=volume,src="+spec.VolumeName+",dst="+spec.WorkspacePath)
	args = append(args, "--workdir", spec.WorkspacePath, "--network", spec.NetworkMode)
	if spec.ReadOnlyRootFS {
		args = append(args, "--read-only")
	}
	if spec.NoNewPrivileges {
		args = append(args, "--security-opt", "no-new-privileges:true")
	}
	for _, capability := range spec.CapDrop {
		if strings.TrimSpace(capability) != "" {
			args = append(args, "--cap-drop", capability)
		}
	}
	if spec.Limits.CPUShares > 0 {
		args = append(args, "--cpu-shares", strconv.FormatInt(spec.Limits.CPUShares, 10))
	}
	if spec.Limits.MemoryBytes > 0 {
		args = append(args, "--memory", strconv.FormatInt(spec.Limits.MemoryBytes, 10))
	}
	if spec.Limits.PidsLimit > 0 {
		args = append(args, "--pids-limit", strconv.FormatInt(spec.Limits.PidsLimit, 10))
	}
	args = append(args, spec.Image, "sh", "-c", "while true; do sleep 3600; done")
	output, err := runtime.run(ctx, args...)
	if err != nil {
		return "", err
	}
	containerID := strings.TrimSpace(output)
	if containerID == "" {
		return "", fmt.Errorf("docker returned an empty container ID")
	}
	return containerID, nil
}

func (runtime *DockerRuntime) StartContainer(ctx context.Context, containerID string) error {
	_, err := runtime.run(ctx, "start", containerID)
	return err
}

func (runtime *DockerRuntime) StopContainer(ctx context.Context, containerID string) error {
	_, err := runtime.run(ctx, "stop", "--time", "10", containerID)
	return err
}

func (runtime *DockerRuntime) RemoveContainer(ctx context.Context, containerID string) error {
	_, err := runtime.run(ctx, "rm", "--force", containerID)
	return err
}

func (runtime *DockerRuntime) Execute(ctx context.Context, containerID, command string, timeout time.Duration) (ExecutionResult, error) {
	if strings.TrimSpace(command) == "" || timeout <= 0 {
		return ExecutionResult{}, fmt.Errorf("command and timeout are required")
	}
	commandContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	output, err := runtime.run(commandContext, "exec", containerID, "sh", "-lc", command)
	result := ExecutionResult{Output: output, ExitCode: 0, Command: command}
	if err != nil {
		result.ExitCode = 1
	}
	return result, err
}

func (runtime *DockerRuntime) run(ctx context.Context, args ...string) (string, error) {
	command := exec.CommandContext(ctx, runtime.binary, args...)
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Run(); err != nil {
		return output.String(), fmt.Errorf("docker %s: %w: %s", args[0], err, strings.TrimSpace(output.String()))
	}
	return output.String(), nil
}
