package github

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"ai-agent/internal/modules/terminal/application"
	"ai-agent/internal/modules/terminal/infrastructure/docker"
)

type GitHubRepositoryCloner struct {
	docker *docker.DockerCLI
}

func NewGitHubRepositoryCloner(dockerCLI *docker.DockerCLI) *GitHubRepositoryCloner {
	return &GitHubRepositoryCloner{docker: dockerCLI}
}

func (cloner *GitHubRepositoryCloner) CloneRepository(ctx context.Context, volumeName, workspacePath, helperImage, repositoryURL, branch, accessToken string, timeout time.Duration) error {
	parsedURL, err := url.ParseRequestURI(repositoryURL)
	if err != nil || parsedURL.Scheme != "https" || parsedURL.Host != "github.com" || parsedURL.User != nil {
		return fmt.Errorf("clone repository: only HTTPS GitHub URLs are allowed")
	}
	if strings.TrimSpace(volumeName) == "" || strings.TrimSpace(workspacePath) == "" || strings.TrimSpace(helperImage) == "" || timeout <= 0 {
		return fmt.Errorf("clone repository: volume, workspace, helper image, and positive timeout are required")
	}

	const script = `set -eu
IFS= read -r token || token=""
repo_mark="$3/.forgeai-repository-cloned"
current_branch=""
if [ -d "$3/.git" ]; then
	current_branch=$(git -C "$3" rev-parse --abbrev-ref HEAD 2>/dev/null || true)
fi
if [ -e "$repo_mark" ] && [ -n "$current_branch" ] && [ "$2" = "$current_branch" ]; then
	exit 0
fi
find "$3" -mindepth 1 -maxdepth 1 ! -name ".forgeai-repository-cloned" -exec rm -rf {} +
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
: > "$repo_mark"`

	commandContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	hostPath := docker.HostWorkspacePath(volumeName)
	_, err = cloner.docker.RunWithInput(commandContext, strings.NewReader(accessToken+"\n"),
		"run", "--rm", "-i", "--network", "bridge",
		"--label", "com.forgeai.managed=true",
		"--mount", "type=bind,src="+hostPath+",dst="+workspacePath,
		"--entrypoint", "sh", helperImage,
		"-c", script, "sh", repositoryURL, branch, workspacePath,
	)
	if err != nil {
		return fmt.Errorf("clone GitHub repository: %w", err)
	}
	return nil
}

var _ application.RepositoryCloner = (*GitHubRepositoryCloner)(nil)
