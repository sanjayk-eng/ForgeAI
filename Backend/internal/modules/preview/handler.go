package preview

import (
	"errors"
	"net/http"

	"ai-agent/internal/middleware"
	apierrors "ai-agent/internal/shared/errors"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
	gateway *Gateway
}

func NewHandler(access projectAccessService, previews previewTargetService, publicOrigin, signingKey string) (*Handler, error) {
	service, err := NewService(access, previews, publicOrigin, signingKey)
	if err != nil {
		return nil, err
	}
	return &Handler{service: service, gateway: NewGateway(service)}, nil
}

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.GET("/projects/:project_id/preview", handler.Info)
}

func (handler *Handler) Info(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		apierrors.Error(c, http.StatusUnauthorized, apierrors.ErrCodeUnauthorized, "authentication is required", nil)
		return
	}
	info, err := handler.service.GetInfo(c.Request.Context(), userID, c.Param("project_id"))
	if errors.Is(err, ErrProjectAccessDenied) {
		apierrors.Error(c, http.StatusForbidden, apierrors.ErrCodeForbidden, "workspace access denied", nil)
		return
	}
	if err != nil {
		apierrors.Error(c, http.StatusBadGateway, apierrors.ErrCodeInternalServer, "preview gateway is unavailable", nil)
		return
	}
	apierrors.Success(c, http.StatusOK, "preview status fetched", info)
}

func (handler *Handler) Gateway(c *gin.Context) {
	handler.gateway.ServeHTTP(c.Writer, c.Request)
}
