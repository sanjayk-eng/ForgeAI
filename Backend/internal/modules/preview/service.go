package preview

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ai-agent/internal/modules/terminal/domain"
)

var (
	ErrProjectAccessDenied = errors.New("preview project access denied")
	ErrPreviewUnavailable  = errors.New("project preview is unavailable")
)

type projectAccessService interface {
	ValidateProjectAccess(ctx context.Context, userID, projectID string) error
}

type previewTargetService interface {
	ResolvePreviewTarget(ctx context.Context, projectID string) (domain.Sandbox, string, int, error)
	OpenPreviewTunnel(ctx context.Context, containerID string, containerPort int) (net.Conn, error)
}

type Info struct {
	SandboxID     string  `json:"sandbox_id"`
	Status        string  `json:"status"`
	ContainerPort *int    `json:"container_port"`
	HostPort      *int    `json:"host_port,omitempty"`
	Protocol      string  `json:"protocol"`
	URL           *string `json:"url"`
}

type Service struct {
	access  projectAccessService
	preview previewTargetService
	origin  *url.URL
	key     []byte
}

func NewService(access projectAccessService, previews previewTargetService, publicOrigin, signingKey string) (*Service, error) {
	origin, err := url.Parse(publicOrigin)
	if err != nil || (origin.Scheme != "http" && origin.Scheme != "https") || origin.Hostname() == "" ||
		origin.User != nil || (origin.Path != "" && origin.Path != "/") || origin.RawQuery != "" || origin.Fragment != "" {
		return nil, fmt.Errorf("preview public origin must be an HTTP(S) origin without a path")
	}
	if len(signingKey) < 16 {
		return nil, fmt.Errorf("preview signing key must contain at least 16 bytes")
	}
	if access == nil || previews == nil {
		return nil, fmt.Errorf("preview access and target services are required")
	}
	return &Service{
		access:  access,
		preview: previews,
		origin:  origin,
		key:     []byte(signingKey),
	}, nil
}

func (service *Service) GetInfo(ctx context.Context, userID, projectID string) (Info, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return Info{}, fmt.Errorf("project ID is required")
	}
	if err := service.access.ValidateProjectAccess(ctx, userID, projectID); err != nil {
		return Info{}, ErrProjectAccessDenied
	}
	sandbox, target, port, err := service.preview.ResolvePreviewTarget(ctx, projectID)
	if errors.Is(err, domain.ErrSandboxNotFound) {
		return Info{Status: "sandbox_unavailable", Protocol: "http"}, nil
	}
	if err != nil {
		return Info{}, err
	}
	switch sandbox.Status {
	case domain.StatusRunning:
		if target == "" || port == 0 || !service.applicationResponds(ctx, target, port) {
			return Info{SandboxID: sandbox.ID, Status: "application_unavailable", Protocol: "http"}, nil
		}
		previewURL, err := service.previewURL(projectID)
		if err != nil {
			return Info{}, err
		}
		return Info{SandboxID: sandbox.ID, Status: "running", ContainerPort: &port, Protocol: "http", URL: &previewURL}, nil
	case domain.StatusStopped:
		return Info{SandboxID: sandbox.ID, Status: "stopped", Protocol: "http"}, nil
	case domain.StatusCreated, domain.StatusStarting:
		return Info{SandboxID: sandbox.ID, Status: "starting", Protocol: "http"}, nil
	default:
		return Info{SandboxID: sandbox.ID, Status: "sandbox_unavailable", Protocol: "http"}, nil
	}
}

func (service *Service) ResolveTarget(ctx context.Context, projectID string) (string, int, error) {
	sandbox, containerID, port, err := service.preview.ResolvePreviewTarget(ctx, projectID)
	if err != nil || sandbox.Status != domain.StatusRunning || containerID == "" || port < 1 || port > 65535 {
		return "", 0, ErrPreviewUnavailable
	}
	return containerID, port, nil
}

func (service *Service) applicationResponds(ctx context.Context, containerID string, port int) bool {
	if containerID == "" || port < 1 || port > 65535 {
		return false
	}
	requestContext, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
	defer cancel()
	request, err := http.NewRequestWithContext(requestContext, http.MethodGet, "http://preview.internal/", nil)
	if err != nil {
		return false
	}
	transport := service.previewTransport(containerID, port)
	defer transport.CloseIdleConnections()
	client := &http.Client{Timeout: 1500 * time.Millisecond, Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.Do(request)
	if err != nil {
		return false
	}
	_ = response.Body.Close()
	return true
}

func (service *Service) previewTransport(containerID string, containerPort int) *http.Transport {
	return &http.Transport{
		Proxy:                 nil,
		DisableKeepAlives:     true,
		ResponseHeaderTimeout: 30 * time.Second,
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return service.preview.OpenPreviewTunnel(ctx, containerID, containerPort)
		},
	}
}
