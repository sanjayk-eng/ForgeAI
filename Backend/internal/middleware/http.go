package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Setup(engine *gin.Engine, allowedOrigins string) {
	engine.Use(gin.Logger(), gin.Recovery(), corsMiddleware(allowedOrigins))
}

func corsMiddleware(allowedOrigins string) gin.HandlerFunc {
	origins := strings.Split(allowedOrigins, ",")
	for index := range origins {
		origins[index] = strings.TrimSpace(origins[index])
	}
	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: allowedOrigins != "*",
		MaxAge:           12 * time.Hour,
	})
}
