package docker

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"ai-agent/internal/modules/terminal/application"
	"ai-agent/internal/modules/terminal/policy"
)

type DockerRuntime struct {
	docker *DockerCLI
}

func NewDockerRuntime(binary string) *DockerRuntime {
	return NewDockerRuntimeWithCLI(NewDockerCLI(binary))
}

func NewDockerRuntimeWithCLI(docker *DockerCLI) *DockerRuntime {
	return &DockerRuntime{docker: docker}
}

func (runtime *DockerRuntime) CreateVolume(ctx context.Context, name string) error {
	hostPath := HostWorkspacePath(name)
	if err := os.MkdirAll(hostPath, 0o755); err != nil {
		return fmt.Errorf("prepare host workspace path %s: %w", hostPath, err)
	}
	cachePath := HostNPMCachePath(name)
	if err := os.MkdirAll(cachePath, 0o755); err != nil {
		return fmt.Errorf("prepare host npm cache path %s: %w", cachePath, err)
	}
	return nil
}

func (runtime *DockerRuntime) RemoveVolume(ctx context.Context, name string) error {
	hostPath := HostWorkspacePath(name)
	if err := os.RemoveAll(hostPath); err != nil {
		return fmt.Errorf("remove host workspace path %s: %w", hostPath, err)
	}
	cachePath := HostNPMCachePath(name)
	if err := os.RemoveAll(cachePath); err != nil {
		return fmt.Errorf("remove host npm cache path %s: %w", cachePath, err)
	}
	return nil
}

func (runtime *DockerRuntime) CreateContainer(ctx context.Context, spec application.ContainerSpec) (string, error) {
	hostPath := HostWorkspacePath(spec.VolumeName)
	localMount := fmt.Sprintf("type=bind,src=%s,dst=%s", hostPath, spec.WorkspacePath)
	args := []string{"create", "--quiet", "--name", spec.Name, "--label", "com.forgeai.managed=true"}
	args = append(args, "--mount", localMount)
	cacheMount := fmt.Sprintf("type=bind,src=%s,dst=/npm-cache", HostNPMCachePath(spec.VolumeName))
	args = append(args, "--mount", cacheMount, "--env", "npm_config_cache=/npm-cache")
	args = append(args, "--workdir", spec.WorkspacePath, "--network", spec.NetworkMode)
	if spec.NetworkMode == "bridge" {
		previewPort := spec.PreviewPort
		if previewPort == 0 {
			previewPort = application.PreviewContainerPort
		}
		if previewPort < policy.MinPreviewPort || previewPort > policy.MaxPreviewPort {
			return "", fmt.Errorf("preview port must be between %d and %d", policy.MinPreviewPort, policy.MaxPreviewPort)
		}
		args = append(args, "--publish", fmt.Sprintf("127.0.0.1::%d/tcp", previewPort))
	}
	if spec.ReadOnlyRootFS {
		args = append(args, "--read-only")
		args = append(args, "--tmpfs", "/tmp:rw,exec,nosuid,size=256m")
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
	args = append(args, "--entrypoint", "sh", spec.Image, "-c", "while true; do sleep 3600; done")
	output, err := runtime.docker.Run(ctx, args...)
	if err != nil {
		return "", err
	}
	return parseContainerID(output)
}

func (runtime *DockerRuntime) StartContainer(ctx context.Context, containerID string) error {
	_, err := runtime.docker.Run(ctx, "start", containerID)
	return err
}

func (runtime *DockerRuntime) StopContainer(ctx context.Context, containerID string) error {
	_, err := runtime.docker.Run(ctx, "stop", "--time", "10", containerID)
	return err
}

func (runtime *DockerRuntime) RemoveContainer(ctx context.Context, containerID string) error {
	_, err := runtime.docker.Run(ctx, "rm", "--force", containerID)
	if isMissingDockerResource(err, "container") {
		return nil
	}
	return err
}

func isMissingDockerResource(err error, resource string) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "no such "+strings.ToLower(resource))
}

func (runtime *DockerRuntime) Execute(ctx context.Context, containerID, command string, timeoutSeconds int) (application.ExecutionResult, error) {
	if strings.TrimSpace(command) == "" || timeoutSeconds <= 0 {
		return application.ExecutionResult{}, fmt.Errorf("command and timeout are required")
	}
	commandContext, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()
	output, err := runtime.docker.Run(commandContext, "exec", containerID, "sh", "-lc", command)
	result := application.ExecutionResult{Output: output, ExitCode: 0, Command: command}
	if err != nil {
		result.ExitCode = 1
	}
	return result, err
}

func (runtime *DockerRuntime) ResolvePreviewAddress(ctx context.Context, containerID string, containerPort int) (string, error) {
	if strings.TrimSpace(containerID) == "" || containerPort < 1 || containerPort > 65535 {
		return "", fmt.Errorf("container and valid preview port are required")
	}
	output, err := runtime.docker.Run(ctx, "port", containerID, strconv.Itoa(containerPort)+"/tcp")
	if err != nil {
		return "", err
	}
	address, err := parseLoopbackPort(output)
	if err != nil {
		return "", err
	}
	return "http://" + address, nil
}

func parseLoopbackPort(output string) (string, error) {
	for _, line := range strings.Fields(output) {
		host, port, err := net.SplitHostPort(strings.TrimSpace(line))
		if err != nil {
			continue
		}
		ip := net.ParseIP(host)
		parsedPort, portErr := strconv.Atoi(port)
		if ip == nil || !ip.IsLoopback() || portErr != nil || parsedPort < 1 || parsedPort > 65535 {
			continue
		}
		return net.JoinHostPort(host, port), nil
	}
	return "", fmt.Errorf("Docker did not report a loopback preview port")
}

func parseContainerID(output string) (string, error) {
	containerID := strings.TrimSpace(output)
	if containerID == "" {
		return "", fmt.Errorf("docker returned an empty container ID")
	}
	if len(containerID) > 255 || strings.ContainsAny(containerID, "\r\n") {
		return "", fmt.Errorf("docker returned an invalid container ID")
	}
	return containerID, nil
}

var _ application.Runtime = (*DockerRuntime)(nil)
