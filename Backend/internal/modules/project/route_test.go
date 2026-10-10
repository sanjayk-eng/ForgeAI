package project_test

import (
	"testing"

	githubmodule "ai-agent/internal/modules/github"
	"ai-agent/internal/modules/project"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutesDoesNotConflictWithIntegrationRoutes(t *testing.T) {
	router := gin.New()

	githubmodule.RegisterRoutes(router, router, nil)
	project.RegisterRoutes(router, nil)
}
