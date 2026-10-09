package project_test

import (
	"testing"

	"ai-agent/internal/modules/integration"
	"ai-agent/internal/modules/project"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutesDoesNotConflictWithIntegrationRoutes(t *testing.T) {
	router := gin.New()

	integration.RegisterRoutes(router, router, nil)
	project.RegisterRoutes(router, nil)
}
