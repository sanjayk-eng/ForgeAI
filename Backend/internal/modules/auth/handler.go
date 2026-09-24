package auth

import (
	"net/http"

	"ai-agent/internal/middleware"
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

func (handler *Handler) OAuthConnect(c *gin.Context) {
	providerType := ProviderType(c.Param("provider"))
	if providerType == "" {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "OAuth provider type is required", nil)
		return
	}

	redirectURL, err := handler.service.OAuthAuthorizationURL(providerType)
	if err != nil {
		code, message := apierrors.CodeOf(err)
		status := http.StatusInternalServerError
		if code == apierrors.ErrCodeBadRequest || code == apierrors.ErrCodeValidation {
			status = http.StatusBadRequest
		}
		apierrors.Error(c, status, code, message, nil)
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
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

func (handler *Handler) VerifyEmail(c *gin.Context) {
	if err := handler.service.VerifyEmail(c.Request.Context(), c.Query("token")); err != nil {
		code, message := apierrors.CodeOf(err)
		status := http.StatusInternalServerError
		if code == apierrors.ErrCodeBadRequest {
			status = http.StatusBadRequest
		}
		apierrors.Error(c, status, code, message, nil)
		return
	}
	apierrors.Success(c, http.StatusOK, "email verified successfully", nil)
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

func (handler *Handler) Refresh(c *gin.Context) {
	var input RefreshTokenRequest
	if err := validate.BindAndValidate(c, &input); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error(), nil)
		return
	}

	result, err := handler.service.Refresh(c.Request.Context(), input)
	if err != nil {
		code, message := apierrors.CodeOf(err)
		status := http.StatusInternalServerError
		if code == apierrors.ErrCodeUnauthorized {
			status = http.StatusUnauthorized
		}
		apierrors.Error(c, status, code, message, nil)
		return
	}
	apierrors.Success(c, http.StatusOK, "token refreshed", result)
}

func (handler *Handler) Logout(c *gin.Context) {
	apierrors.Success(c, http.StatusOK, "logout successful", nil)
}

func (handler *Handler) Me(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		apierrors.Error(c, http.StatusUnauthorized, apierrors.ErrCodeUnauthorized, "authenticated user is required", nil)
		return
	}

	result, err := handler.service.GetUser(c.Request.Context(), userID)
	if err != nil {
		code, message := apierrors.CodeOf(err)
		status := http.StatusInternalServerError
		if code == apierrors.ErrCodeNotFound {
			status = http.StatusNotFound
		}
		apierrors.Error(c, status, code, message, nil)
		return
	}
	apierrors.Success(c, http.StatusOK, "user fetched", result)
}
