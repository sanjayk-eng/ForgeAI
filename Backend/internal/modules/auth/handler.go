package auth

import (
	"net/http"

	apierrors "ai-agent/internal/shared/errors"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) OAuthCallback(c *gin.Context) {
	providerType := c.Query("type")
	if providerType == "" {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "OAuth provider type is required", nil)
		return
	}

	code := c.Query("code")
	if code == "" {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "OAuth code is required", nil)
		return
	}

	user, err := handler.service.ExchangeCode(c.Request.Context(), ProviderType(providerType), code)
	if err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeBadRequest, err.Error(), nil)
		return
	}
	apierrors.Success(c, http.StatusOK, "OAuth code exchanged", user)
}
