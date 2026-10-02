package terminal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-agent/internal/middleware"
	"ai-agent/internal/modules/terminal/application"
	"ai-agent/internal/modules/terminal/policy"
	"ai-agent/internal/shared/realtime"
	appjwt "ai-agent/pkg/jwt"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
)

func TestProjectWebSocketRequiresProjectAccessAndStreamsEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtManager, err := appjwt.NewManager(strings.Repeat("j", 32), time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	userTokens, err := jwtManager.GeneratePair("member-1")
	if err != nil {
		t.Fatalf("GeneratePair() error = %v", err)
	}
	deniedTokens, err := jwtManager.GeneratePair("member-2")
	if err != nil {
		t.Fatalf("GeneratePair() error = %v", err)
	}
	service := application.NewService(webSocketTestStore{}, nil, nil, nil, policy.Sandbox{})
	hub := NewEventHub()
	module := &Module{Service: service, Events: hub, WebSocketOrigins: []string{"http://app.test"}}
	router := gin.New()
	RegisterRoutes(middleware.ProtectedGroup(router, jwtManager, nil), NewHandler(module))
	server := httptest.NewServer(router)
	defer server.Close()
	webSocketURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/projects/project-1/ws"

	deniedResponse, deniedResp, err := websocket.Dial(context.Background(), webSocketURL, &websocket.DialOptions{
		Subprotocols: []string{"forgeai", "forgeai-auth." + deniedTokens.AccessToken},
		HTTPHeader:   http.Header{"Origin": []string{"http://app.test"}},
	})
	if deniedResponse != nil {
		_ = deniedResponse.Close(websocket.StatusNormalClosure, "")
	}
	if err == nil || deniedResp == nil || deniedResp.StatusCode != http.StatusForbidden {
		t.Fatalf("unauthorized websocket dial error/response = %v/%v, want HTTP 403", err, deniedResp)
	}

	connection, _, err := websocket.Dial(context.Background(), webSocketURL, &websocket.DialOptions{
		Subprotocols: []string{"forgeai", "forgeai-auth." + userTokens.AccessToken},
		HTTPHeader:   http.Header{"Origin": []string{"http://app.test"}},
	})
	if err != nil {
		t.Fatalf("authorized websocket dial: %v", err)
	}
	defer connection.Close(websocket.StatusNormalClosure, "test complete")

	hub.Publish(realtime.Event{
		Version: 1, Event: "file.changed", organizationID: "workspace-1", ProjectID: "project-1",
		SandboxID: "sandbox-1", Path: "src/App.tsx", ChangeType: "changed",
	})
	readContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, payload, err := connection.Read(readContext)
	if err != nil {
		t.Fatalf("read websocket event: %v", err)
	}
	var event struct {
		ProjectID string `json:"project_id"`
		SandboxID string `json:"sandbox_id"`
		Path      string `json:"path"`
		Event     string `json:"event"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		t.Fatalf("decode websocket event: %v", err)
	}
	if event.ProjectID != "project-1" || event.SandboxID != "sandbox-1" || event.Path != "src/App.tsx" || event.Event != "file.changed" {
		t.Fatalf("unexpected websocket event: %+v", event)
	}
}

type webSocketTestStore struct{ application.Store }

func (webSocketTestStore) FindProjectWorkspace(context.Context, string) (string, error) {
	return "workspace-1", nil
}

func (webSocketTestStore) CanAccessWorkspace(_ context.Context, _ string, userID string) (bool, error) {
	return userID == "member-1", nil
}
