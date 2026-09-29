package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	appjwt "ai-agent/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func TestAuthenticateAcceptsJWTInWebSocketSubprotocol(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager, err := appjwt.NewManager(strings.Repeat("x", 32), time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	pair, err := manager.GeneratePair("user-1")
	if err != nil {
		t.Fatalf("GeneratePair() error = %v", err)
	}
	router := gin.New()
	router.GET("/projects/:project_id/ws", Authenticate(manager, nil), func(c *gin.Context) {
		userID, ok := UserID(c)
		if !ok || userID != "user-1" {
			c.Status(http.StatusUnauthorized)
			return
		}
		c.Status(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodGet, "/projects/project-1/ws", nil)
	request.Header.Set("Connection", "Upgrade")
	request.Header.Set("Upgrade", "websocket")
	request.Header.Set("Sec-WebSocket-Protocol", "forgeai, forgeai-auth."+pair.AccessToken)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("websocket auth status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestWebSocketSubprotocolTokenRequiresUpgrade(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/projects/project-1/ws", nil)
	request.Header.Set("Sec-WebSocket-Protocol", "forgeai-auth.token")
	if isWebSocketUpgrade(request) {
		t.Fatal("ordinary HTTP request was treated as a WebSocket upgrade")
	}
}
