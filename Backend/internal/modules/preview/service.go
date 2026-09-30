package preview

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
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
}

type Info struct {
	Status string  `json:"status"`
	Port   *int    `json:"port"`
	URL    *string `json:"url"`
}

type Service struct {
	access  projectAccessService
	preview previewTargetService
	origin  *url.URL
	key     []byte
	probe   *http.Client
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
		probe: &http.Client{
			Timeout: 1500 * time.Millisecond,
			Transport: &http.Transport{
				Proxy: nil,
				DialContext: (&net.Dialer{
					Timeout: 1000 * time.Millisecond,
				}).DialContext,
			},
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
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
		return Info{Status: "sandbox_unavailable"}, nil
	}
	if err != nil {
		return Info{}, err
	}
	switch sandbox.Status {
	case domain.StatusRunning:
		if target == "" || !service.applicationResponds(ctx, target) {
			return Info{Status: "application_unavailable"}, nil
		}
		previewURL, err := service.previewURL(projectID)
		if err != nil {
			return Info{}, err
		}
		return Info{Status: "running", Port: &port, URL: &previewURL}, nil
	case domain.StatusStopped:
		return Info{Status: "stopped"}, nil
	case domain.StatusCreated, domain.StatusStarting:
		return Info{Status: "starting"}, nil
	default:
		return Info{Status: "sandbox_unavailable"}, nil
	}
}

func (service *Service) ResolveTarget(ctx context.Context, projectID string) (string, error) {
	sandbox, target, _, err := service.preview.ResolvePreviewTarget(ctx, projectID)
	if err != nil || sandbox.Status != domain.StatusRunning || !validTarget(target) {
		return "", ErrPreviewUnavailable
	}
	return target, nil
}

func (service *Service) applicationResponds(ctx context.Context, target string) bool {
	if !validTarget(target) {
		return false
	}
	requestContext, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
	defer cancel()
	request, err := http.NewRequestWithContext(requestContext, http.MethodGet, target+"/", nil)
	if err != nil {
		return false
	}
	response, err := service.probe.Do(request)
	if err != nil {
		return false
	}
	_ = response.Body.Close()
	return true
}

func validTarget(target string) bool {
	parsed, err := url.Parse(target)
	if err != nil || parsed.Scheme != "http" || parsed.Hostname() != "127.0.0.1" || parsed.User != nil ||
		parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Path != "" {
		return false
	}
	port, err := strconv.Atoi(parsed.Port())
	return err == nil && port > 0 && port <= 65535
}
