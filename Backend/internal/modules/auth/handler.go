package auth

import (
	"net/http"

	apierrors "ai-agent/internal/shared/errors"
	"ai-agent/pkg/validate"

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

	result, err := handler.service.AuthenticateOAuth(c.Request.Context(), ProviderType(providerType), code)
	if err != nil {
		code, message := apierrors.CodeOf(err)
		status := http.StatusInternalServerError
		if code == apierrors.ErrCodeValidation || code == apierrors.ErrCodeBadRequest {
			status = http.StatusBadRequest
		}
		apierrors.Error(c, status, code, message, nil)
		return
	}
	apierrors.Success(c, http.StatusOK, "authentication successful", result)
}

func (handler *Handler) Register(c *gin.Context) {
	var input RegisterRequest
	if err := validate.BindAndValidate(c, &input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}

	result, err := handler.service.Register(c.Request.Context(), input)
	if err != nil {
		code, message := apierrors.CodeOf(err)
		status := http.StatusInternalServerError
		if code == apierrors.ErrCodeConflict {
			status = http.StatusConflict
		}
		apierrors.Error(c, status, code, message, nil)
		return
	}
	apierrors.Success(c, http.StatusCreated, "registration successful", result)
}

func (handler *Handler) Login(c *gin.Context) {
	var input LoginRequest
	if err := validate.BindAndValidate(c, &input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}

	result, err := handler.service.Login(c.Request.Context(), input)
	if err != nil {
		code, message := apierrors.CodeOf(err)
		status := http.StatusInternalServerError
		if code == apierrors.ErrCodeUnauthorized {
			status = http.StatusUnauthorized
		}
		apierrors.Error(c, status, code, message, nil)
		return
	}
	apierrors.Success(c, http.StatusOK, "login successful", result)
}
