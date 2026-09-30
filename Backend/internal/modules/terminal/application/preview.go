package application

import (
	"context"
	"fmt"
	"strings"

	"ai-agent/internal/modules/terminal/domain"
)

type PreviewService struct {
	store    Store
	resolver PreviewAddressResolver
}

func NewPreviewService(store Store, resolver PreviewAddressResolver) *PreviewService {
	return &PreviewService{store: store, resolver: resolver}
}

func (service *PreviewService) ResolvePreviewTarget(ctx context.Context, projectID string) (domain.Sandbox, string, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" || service.store == nil || service.resolver == nil {
		return domain.Sandbox{}, "", domain.ErrInvalidSandbox
	}
	sandbox, err := service.store.FindActiveByProject(ctx, projectID)
	if err != nil {
		return domain.Sandbox{}, "", err
	}
	if sandbox.Status != domain.StatusRunning || sandbox.ContainerID == "" {
		return sandbox, "", nil
	}
	address, err := service.resolver.ResolvePreviewAddress(ctx, sandbox.ContainerID, PreviewContainerPort)
	if err != nil {
		return sandbox, "", fmt.Errorf("resolve sandbox preview port: %w", err)
	}
	return sandbox, address, nil
}
