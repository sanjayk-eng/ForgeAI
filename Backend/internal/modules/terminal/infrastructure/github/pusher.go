package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ai-agent/internal/modules/terminal/application"
	"ai-agent/internal/modules/terminal/infrastructure/docker"
)

type GitHubRepositoryPusher struct {
	docker *docker.DockerCLI
}

func NewGitHubRepositoryPusher(dockerCLI *docker.DockerCLI) *GitHubRepositoryPusher {
	return &GitHubRepositoryPusher{docker: dockerCLI}
}

func (pusher *GitHubRepositoryPusher) PushRepository(ctx context.Context, volumeName, workspacePath, helperImage, remote, branch, accessToken string, timeout time.Duration) error {
	if strings.TrimSpace(volumeName) == "" || strings.TrimSpace(workspacePath) == "" ||
		strings.TrimSpace(helperImage) == "" || strings.TrimSpace(remote) == "" ||
		strings.TrimSpace(accessToken) == "" || timeout <= 0 {
		return fmt.Errorf("push repository: volume, workspace, helper image, remote, token, and positive timeout are required")
	}

	const script = `set -eu
IFS= read -r token || exit 1
auth=$(printf 'x-access-token:%s' "$token" | base64 | tr -d '\n')
export GIT_CONFIG_COUNT=1
export GIT_CONFIG_KEY_0='http.https://github.com/.extraheader'
export GIT_CONFIG_VALUE_0="AUTHORIZATION: basic $auth"
cd "$1"
if [ -n "$3" ]; then
	exec git push -- "$2" "$3"
fi
exec git push -- "$2"`

	commandContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	hostPath := docker.HostWorkspacePath(volumeName)
	_, err := pusher.docker.RunWithInput(commandContext, strings.NewReader(accessToken+"\n"),
		"run", "--rm", "-i", "--network", "bridge",
		"--read-only", "--tmpfs", "/tmp:rw,noexec,nosuid,size=64m",
		"--security-opt", "no-new-privileges:true", "--cap-drop", "ALL",
		"--mount", "type=bind,src="+hostPath+",dst="+workspacePath,
		"--entrypoint", "sh", helperImage,
		"-c", script, "sh", workspacePath, remote, branch,
	)
	if err != nil {
		return fmt.Errorf("push GitHub repository: %w", err)
	}
	return nil
}

var _ application.RepositoryPusher = (*GitHubRepositoryPusher)(nil)
