package preview

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"ai-agent/internal/modules/terminal/domain"

	"github.com/coder/websocket"
)

func TestGetInfoReturnsSignedPreviewOnlyWhenApplicationResponds(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	sandbox := testSandboxService{
		sandbox: domain.Sandbox{Status: domain.StatusRunning},
		target:  upstream.URL,
	}
	service, err := NewService(sandbox, sandbox, "http://preview.localhost:8080", "test-preview-signing-key")
	if err != nil {
		t.Fatal(err)
	}
	info, err := service.GetInfo(context.Background(), "user-1", "project-123")
	if err != nil {
		t.Fatalf("GetInfo() error = %v", err)
	}
	if info.Status != "running" || info.Port == nil || *info.Port != 5173 || info.URL == nil {
		t.Fatalf("unexpected preview info: %+v", info)
	}
	parsedURL, err := url.Parse(*info.URL)
	if err != nil {
		t.Fatal(err)
	}
	projectID, ok := service.projectFromHost(parsedURL.Host)
	if !ok || projectID != "project-123" {
		t.Fatalf("preview hostname did not identify project: %q, %v", projectID, ok)
	}
	if _, ok := service.projectFromHost("other-project.preview.localhost:8080"); ok {
		t.Fatal("unsigned hostname resolved to a project")
	}
}

func TestGetInfoReportsApplicationUnavailable(t *testing.T) {
	sandbox := testSandboxService{
		sandbox: domain.Sandbox{Status: domain.StatusRunning},
		target:  "http://127.0.0.1:1",
	}
	service, err := NewService(sandbox, sandbox, "http://preview.localhost:8080", "test-preview-signing-key")
	if err != nil {
		t.Fatal(err)
	}
	info, err := service.GetInfo(context.Background(), "user-1", "project-123")
	if err != nil || info.Status != "application_unavailable" || info.URL != nil || info.Port != nil {
		t.Fatalf("GetInfo() = %+v, %v; want application_unavailable", info, err)
	}
}

func TestGatewayRoutesOnlySignedProjectHost(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(request.URL.Path))
	}))
	defer upstream.Close()

	sandbox := testSandboxService{
		sandbox: domain.Sandbox{Status: domain.StatusRunning},
		target:  upstream.URL,
	}
	service, err := NewService(sandbox, sandbox, "http://preview.localhost:8080", "test-preview-signing-key")
	if err != nil {
		t.Fatal(err)
	}
	previewURL, err := service.previewURL("project-123")
	if err != nil {
		t.Fatal(err)
	}
	previewHost, err := url.Parse(previewURL)
	if err != nil {
		t.Fatal(err)
	}
	gateway := NewGateway(service)
	request := httptest.NewRequest(http.MethodGet, "http://preview.localhost:8080/assets/app.js", nil)
	request.Host = previewHost.Host
	response := httptest.NewRecorder()
	gateway.ServeHTTP(response, request)
	if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "/assets/app.js" {
		t.Fatalf("gateway response = %d %q", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodGet, "http://preview.localhost:8080/", nil)
	request.Host = "other-project.preview.localhost:8080"
	response = httptest.NewRecorder()
	gateway.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("unsigned host returned %d, want 404", response.Code)
	}
}

func TestGatewayProxiesWebSocketUpgrade(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		connection, err := websocket.Accept(writer, request, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			return
		}
		defer connection.Close(websocket.StatusNormalClosure, "")
		messageType, message, err := connection.Read(request.Context())
		if err == nil {
			_ = connection.Write(request.Context(), messageType, message)
		}
	}))
	defer upstream.Close()

	sandbox := testSandboxService{
		sandbox: domain.Sandbox{Status: domain.StatusRunning},
		target:  upstream.URL,
	}
	service, err := NewService(sandbox, sandbox, "http://preview.localhost", "test-preview-signing-key")
	if err != nil {
		t.Fatal(err)
	}
	previewURL, err := service.previewURL("project-123")
	if err != nil {
		t.Fatal(err)
	}
	previewHost, err := url.Parse(previewURL)
	if err != nil {
		t.Fatal(err)
	}
	gatewayServer := httptest.NewServer(NewGateway(service))
	defer gatewayServer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, "ws://"+gatewayServer.Listener.Addr().String()+"/hmr", &websocket.DialOptions{
		Host: previewHost.Host,
	})
	if err != nil {
		t.Fatalf("WebSocket dial through preview gateway: %v", err)
	}
	defer connection.Close(websocket.StatusNormalClosure, "")
	if err := connection.Write(ctx, websocket.MessageText, []byte("vite-hmr")); err != nil {
		t.Fatalf("write through preview WebSocket: %v", err)
	}
	_, message, err := connection.Read(ctx)
	if err != nil || string(message) != "vite-hmr" {
		t.Fatalf("WebSocket echo = %q, %v", message, err)
	}
}

func TestNewServiceRequiresValidOriginAndSigningKey(t *testing.T) {
	validSandboxService := testSandboxService{}
	for _, test := range []struct {
		origin string
		key    string
	}{
		{origin: "ftp://preview.example.com", key: "test-preview-signing-key"},
		{origin: "https://preview.example.com/path", key: "test-preview-signing-key"},
		{origin: "https://preview.example.com", key: "short"},
	} {
		if _, err := NewService(validSandboxService, validSandboxService, test.origin, test.key); err == nil {
			t.Errorf("NewService(%q, %q) unexpectedly succeeded", test.origin, test.key)
		}
	}
}

type testSandboxService struct {
	sandbox    domain.Sandbox
	target     string
	accessErr  error
	resolveErr error
}

func (service testSandboxService) ValidateProjectAccess(context.Context, string, string) error {
	return service.accessErr
}

func (service testSandboxService) ResolvePreviewTarget(context.Context, string) (domain.Sandbox, string, error) {
	return service.sandbox, service.target, service.resolveErr
}
