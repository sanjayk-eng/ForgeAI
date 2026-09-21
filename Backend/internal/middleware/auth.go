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

		authorization := strings.TrimSpace(c.GetHeader("Authorization"))
		parts := strings.SplitN(authorization, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			unauthorized(c, "Bearer access token is required")
			return
		}

		claims, err := jwtManager.Parse(strings.TrimSpace(parts[1]), appjwt.AccessToken)
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
