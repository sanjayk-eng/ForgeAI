package application

import (
	"context"
	"fmt"
	"strings"

	"ai-agent/internal/modules/terminal/domain"
)

type PreviewService struct {
	store       Store
	resolver    PreviewAddressResolver
	previewPort int
}

func NewPreviewService(store Store, resolver PreviewAddressResolver, previewPort int) *PreviewService {
	if previewPort == 0 {
		previewPort = PreviewContainerPort
	}
	return &PreviewService{store: store, resolver: resolver, previewPort: previewPort}
}

func (service *PreviewService) ResolvePreviewTarget(ctx context.Context, projectID string) (domain.Sandbox, string, int, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" || service.store == nil || service.resolver == nil {
		return domain.Sandbox{}, "", service.previewPort, domain.ErrInvalidSandbox
	}
	sandbox, err := service.store.FindActiveByProject(ctx, projectID)
	if err != nil {
		return domain.Sandbox{}, "", service.previewPort, err
	}
	if sandbox.Status != domain.StatusRunning || sandbox.ContainerID == "" {
		return sandbox, "", service.previewPort, nil
	}
	address, err := service.resolver.ResolvePreviewAddress(ctx, sandbox.ContainerID, service.previewPort)
	if err != nil {
		return sandbox, "", service.previewPort, fmt.Errorf("resolve sandbox preview port: %w", err)
	}
	return sandbox, address, service.previewPort, nil
}
