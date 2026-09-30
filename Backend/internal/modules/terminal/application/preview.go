package application

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"ai-agent/internal/modules/terminal/domain"
)

type PreviewService struct {
	store   Store
	runtime PreviewRuntime
}

func NewPreviewService(store Store, runtime PreviewRuntime) *PreviewService {
	return &PreviewService{store: store, runtime: runtime}
}

func (service *PreviewService) ResolvePreviewTarget(ctx context.Context, projectID string) (domain.Sandbox, string, int, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" || service.store == nil || service.runtime == nil {
		return domain.Sandbox{}, "", 0, domain.ErrInvalidSandbox
	}
	sandbox, err := service.store.FindActiveByProject(ctx, projectID)
	if err != nil {
		return domain.Sandbox{}, "", 0, err
	}
	if sandbox.Status != domain.StatusRunning || sandbox.ContainerID == "" {
		return sandbox, "", 0, nil
	}
	port, err := service.runtime.DetectPreviewPort(ctx, sandbox.ContainerID)
	if errors.Is(err, ErrPreviewPortNotFound) {
		return sandbox, sandbox.ContainerID, 0, nil
	}
	if err != nil {
		return sandbox, "", 0, fmt.Errorf("detect sandbox preview port: %w", err)
	}
	return sandbox, sandbox.ContainerID, port, nil
}

func (service *PreviewService) OpenPreviewTunnel(ctx context.Context, containerID string, containerPort int) (net.Conn, error) {
	if service.runtime == nil {
		return nil, domain.ErrInvalidSandbox
	}
	return service.runtime.OpenPreviewTunnel(ctx, containerID, containerPort)
}
