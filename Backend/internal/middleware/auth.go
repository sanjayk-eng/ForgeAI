package middleware

import (
	"context"
	"net/http"
	"strings"

	"ai-agent/internal/shared/logger"
	appjwt "ai-agent/pkg/jwt"

	"github.com/gin-gonic/gin"
)

type contextKey string

const userIDKey contextKey = "user_id"

func Authenticate(jwtManager *appjwt.Manager, log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if jwtManager == nil {
			unauthorized(c, "authentication is not configured")
			return
		}

		accessToken := bearerToken(c.GetHeader("Authorization"))
		if accessToken == "" && isWebSocketUpgrade(c.Request) {
			accessToken = webSocketAccessToken(c.GetHeader("Sec-WebSocket-Protocol"))
		}
		if accessToken == "" {
			unauthorized(c, "Bearer access token is required")
			return
		}

		claims, err := jwtManager.Parse(accessToken, appjwt.AccessToken)
		if err != nil {
			unauthorized(c, "invalid access token")
			return
		}

		c.Set(string(userIDKey), claims.Subject)
		requestContext := context.WithValue(c.Request.Context(), userIDKey, claims.Subject)
		c.Request = c.Request.WithContext(requestContext)
		if log != nil {
			log.Info(requestContext, "authenticated request", "user_id", claims.Subject, "method", c.Request.Method, "path", c.Request.URL.Path)
		}
		c.Next()
	}
}

func bearerToken(authorization string) string {
	parts := strings.SplitN(strings.TrimSpace(authorization), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func isWebSocketUpgrade(request *http.Request) bool {
	return strings.EqualFold(strings.TrimSpace(request.Header.Get("Upgrade")), "websocket") &&
		strings.Contains(strings.ToLower(request.Header.Get("Connection")), "upgrade")
}

func webSocketAccessToken(protocolHeader string) string {
	for _, protocol := range strings.Split(protocolHeader, ",") {
		protocol = strings.TrimSpace(protocol)
		if strings.HasPrefix(protocol, "forgeai-auth.") {
			return strings.TrimPrefix(protocol, "forgeai-auth.")
		}
	}
	return ""
}

func ProtectedGroup(router gin.IRouter, jwtManager *appjwt.Manager, log logger.Logger) *gin.RouterGroup {
	return router.Group("", Authenticate(jwtManager, log))
}

func UserID(c *gin.Context) (string, bool) {
	value, exists := c.Get(string(userIDKey))
	userID, ok := value.(string)
	return userID, exists && ok && userID != ""
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok && userID != ""
}

func unauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "UNAUTHORIZED",
			"message": message,
		},
	})
}
