package infrastructure

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
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

type FileEntry struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	IsDirectory bool   `json:"is_directory"`
	Size        int64  `json:"size,omitempty"`
	ModifiedAt  string `json:"modified_at,omitempty"`
}

type Runtime interface {
	CreateVolume(ctx context.Context, name string) error
	RemoveVolume(ctx context.Context, name string) error
	CreateContainer(ctx context.Context, spec ContainerSpec) (string, error)
	StartContainer(ctx context.Context, containerID string) error
	StopContainer(ctx context.Context, containerID string) error
	RemoveContainer(ctx context.Context, containerID string) error
	Execute(ctx context.Context, containerID, command string, timeoutSeconds int) (ExecutionResult, error)
	ListFiles(ctx context.Context, containerID, path string) ([]FileEntry, error)
	ReadFile(ctx context.Context, containerID, path string) (string, error)
	WriteFile(ctx context.Context, containerID, path, content string) error
	CloneRepository(ctx context.Context, volumeName, workspacePath, helperImage, repositoryURL, branch, accessToken string, timeout time.Duration) error
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

func hostWorkspacePath(name string) string {
	root := filepath.Join(os.TempDir(), "forgeai-workspaces")
	_ = os.MkdirAll(root, 0o755)
	return filepath.Join(root, name)
}

func (runtime *DockerRuntime) CreateVolume(ctx context.Context, name string) error {
	hostPath := hostWorkspacePath(name)
	if err := os.MkdirAll(hostPath, 0o755); err != nil {
		return fmt.Errorf("prepare host workspace path %s: %w", hostPath, err)
	}
	return nil
}

func (runtime *DockerRuntime) RemoveVolume(ctx context.Context, name string) error {
	hostPath := hostWorkspacePath(name)
	if err := os.RemoveAll(hostPath); err != nil {
		return fmt.Errorf("remove host workspace path %s: %w", hostPath, err)
	}
	return nil
}

func (runtime *DockerRuntime) CreateContainer(ctx context.Context, spec ContainerSpec) (string, error) {
	hostPath := hostWorkspacePath(spec.VolumeName)
	localMount := fmt.Sprintf("type=bind,src=%s,dst=%s", hostPath, spec.WorkspacePath)
	args := []string{"create", "--quiet", "--name", spec.Name, "--label", "com.forgeai.managed=true"}
	args = append(args, "--mount", localMount)
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
	return parseContainerID(output)
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
	if isMissingDockerResource(err, "container") {
		return nil
	}
	return err
}

func isMissingDockerResource(err error, resource string) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "no such "+strings.ToLower(resource))
}

func (runtime *DockerRuntime) Execute(ctx context.Context, containerID, command string, timeoutSeconds int) (ExecutionResult, error) {
	if strings.TrimSpace(command) == "" || timeoutSeconds <= 0 {
		return ExecutionResult{}, fmt.Errorf("command and timeout are required")
	}
	commandContext, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()
	output, err := runtime.run(commandContext, "exec", containerID, "sh", "-lc", command)
	result := ExecutionResult{Output: output, ExitCode: 0, Command: command}
	if err != nil {
		result.ExitCode = 1
	}
	return result, err
}

func parseListFiles(output, path string) []FileEntry {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	entries := make([]FileEntry, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "total ") {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) == 0 {
			continue
		}
		name := fields[len(fields)-1]
		isDir := strings.HasSuffix(name, "/")
		name = strings.TrimSuffix(name, "/")
		if name == "." || name == ".." {
			continue
		}
		entryPath := strings.TrimRight(path, "/") + "/" + name
		if path == "/" {
			entryPath = "/" + name
		}
		entries = append(entries, FileEntry{
			Name:        name,
			Path:        entryPath,
			IsDirectory: isDir,
		})
	}
	return entries
}

func (runtime *DockerRuntime) ListFiles(ctx context.Context, containerID, path string) ([]FileEntry, error) {
	if strings.TrimSpace(containerID) == "" {
		return nil, fmt.Errorf("container id is required")
	}
	if strings.TrimSpace(path) == "" {
		path = "/workspace"
	}
	commandContext, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	output, err := runtime.run(commandContext, "exec", containerID, "sh", "-lc", `ls -A1p --full-time "$1"`, "sh", path)
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}
	return parseListFiles(output, path), nil
}

func (runtime *DockerRuntime) ReadFile(ctx context.Context, containerID, path string) (string, error) {
	if strings.TrimSpace(containerID) == "" {
		return "", fmt.Errorf("container id is required")
	}
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("file path is required")
	}
	commandContext, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	output, err := runtime.run(commandContext, "exec", containerID, "sh", "-lc", `sed -n '1,250p' "$1"`, "sh", path)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	return output, nil
}

func (runtime *DockerRuntime) WriteFile(ctx context.Context, containerID, path, content string) error {
	if strings.TrimSpace(containerID) == "" || strings.TrimSpace(path) == "" {
		return fmt.Errorf("container id and file path are required")
	}
	commandContext, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	command := exec.CommandContext(commandContext, runtime.binary,
		"exec", "-i", containerID, "sh", "-lc",
		`mkdir -p "$(dirname "$1")" && cat > "$1"`, "sh", path,
	)
	command.Stdin = strings.NewReader(content)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		diagnostic := strings.TrimSpace(strings.Join([]string{stderr.String(), stdout.String()}, "\n"))
		return fmt.Errorf("write file: %w: %s", err, diagnostic)
	}
	return nil
}

func (runtime *DockerRuntime) CloneRepository(ctx context.Context, volumeName, workspacePath, helperImage, repositoryURL, branch, accessToken string, timeout time.Duration) error {
	parsedURL, err := url.ParseRequestURI(repositoryURL)
	if err != nil || parsedURL.Scheme != "https" || parsedURL.Host != "github.com" || parsedURL.User != nil {
		return fmt.Errorf("clone repository: only HTTPS GitHub URLs are allowed")
	}
	if strings.TrimSpace(volumeName) == "" || strings.TrimSpace(workspacePath) == "" || strings.TrimSpace(helperImage) == "" || timeout <= 0 {
		return fmt.Errorf("clone repository: volume, workspace, helper image, and positive timeout are required")
	}

	const script = `set -eu
IFS= read -r token || token=""
[ -e "$3/.forgeai-repository-cloned" ] && exit 0
clone_dir=$(mktemp -d)
trap 'rm -rf "$clone_dir"' EXIT
if [ -n "$token" ]; then
	auth=$(printf 'x-access-token:%s' "$token" | base64 | tr -d '\n')
	if [ -n "$2" ]; then
		git -c "http.extraheader=AUTHORIZATION: basic $auth" clone --depth 1 --branch "$2" -- "$1" "$clone_dir/repo"
	else
		git -c "http.extraheader=AUTHORIZATION: basic $auth" clone --depth 1 -- "$1" "$clone_dir/repo"
	fi
else
	if [ -n "$2" ]; then
		git clone --depth 1 --branch "$2" -- "$1" "$clone_dir/repo"
	else
		git clone --depth 1 -- "$1" "$clone_dir/repo"
	fi
fi
cp -a "$clone_dir/repo/." "$3/"
: > "$3/.forgeai-repository-cloned"`

	commandContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	hostPath := hostWorkspacePath(volumeName)
	command := exec.CommandContext(commandContext, runtime.binary,
		"run", "--rm", "-i", "--network", "bridge",
		"--label", "com.forgeai.managed=true",
		"--mount", "type=bind,src="+hostPath+",dst="+workspacePath,
		"--entrypoint", "sh", helperImage,
		"-c", script, "sh", repositoryURL, branch, workspacePath,
	)
	command.Stdin = strings.NewReader(accessToken + "\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		diagnostic := strings.TrimSpace(strings.Join([]string{stderr.String(), stdout.String()}, "\n"))
		return fmt.Errorf("clone GitHub repository: %w: %s", err, diagnostic)
	}
	return nil
}

func (runtime *DockerRuntime) run(ctx context.Context, args ...string) (string, error) {
	command := exec.CommandContext(ctx, runtime.binary, args...)
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
