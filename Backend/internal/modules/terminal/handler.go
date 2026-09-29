package terminal

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"ai-agent/internal/middleware"
	"ai-agent/internal/modules/terminal/domain"
	apierrors "ai-agent/internal/shared/errors"
	"ai-agent/internal/shared/realtime"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	module *Module
}

func NewHandler(module *Module) *Handler {
	return &Handler{module: module}
}

func (h *Handler) GetSandboxByProject(c *gin.Context) {
	projectID := c.Param("project_id")
	if projectID == "" {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "project_id is required", nil)
		return
	}

	sandbox, err := h.module.Service.GetByProject(c.Request.Context(), projectID)
	if err != nil {
		if errors.Is(err, domain.ErrSandboxNotFound) && c.Query("ensure") == "true" {
			if accessErr := h.module.Service.ValidateProjectAccess(c.Request.Context(), userID(c), projectID); accessErr != nil {
				h.writeError(c, accessErr)
				return
			}
			h.module.OnProjectCreated(c.Request.Context(), projectID, userID(c))
		}
		h.writeError(c, err)
		return
	}
	if accessErr := h.module.Service.ValidateSandboxAccess(c.Request.Context(), userID(c), sandbox.ID); accessErr != nil {
		h.writeError(c, accessErr)
		return
	}
	if c.Query("ensure") == "true" && (sandbox.Status == domain.StatusFailed || sandbox.Status == domain.StatusRunning) {
		if accessErr := h.module.Service.ValidateProjectAccess(c.Request.Context(), userID(c), projectID); accessErr != nil {
			h.writeError(c, accessErr)
			return
		}
		h.module.OnProjectCreated(c.Request.Context(), projectID, userID(c))
	}

	apierrors.Success(c, http.StatusOK, "sandbox fetched", sandbox)
}

func (h *Handler) ListSandboxFiles(c *gin.Context) {
	sandboxID, ok := h.requireSandboxAccess(c)
	if !ok {
		return
	}
	path := strings.TrimSpace(c.Query("path"))
	if path == "" {
		path = "/workspace"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	entries, err := h.module.Service.ListFiles(c.Request.Context(), sandboxID, path)
	if err != nil {
		h.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "files fetched", gin.H{"files": entries})
}

func (h *Handler) ReadSandboxFile(c *gin.Context) {
	sandboxID, ok := h.requireSandboxAccess(c)
	if !ok {
		return
	}
	filePath := c.Param("path")
	if filePath == "" {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "path is required", nil)
		return
	}
	filePath = "/" + strings.TrimPrefix(filePath, "/")

	content, err := h.module.Service.ReadFile(c.Request.Context(), sandboxID, filePath)
	if err != nil {
		h.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "file fetched", gin.H{"path": filePath, "content": content})
}

func (h *Handler) SaveSandboxFile(c *gin.Context) {
	sandboxID, ok := h.requireSandboxAccess(c)
	if !ok {
		return
	}

	var payload struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "invalid save payload", nil)
		return
	}
	if strings.TrimSpace(payload.Path) == "" {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "path is required", nil)
		return
	}
	if err := h.module.Service.WriteFile(c.Request.Context(), sandboxID, payload.Path, payload.Content); err != nil {
		h.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "file saved", gin.H{"path": payload.Path})
}

func (h *Handler) CreateSandboxPath(c *gin.Context) {
	sandboxID, ok := h.requireSandboxAccess(c)
	if !ok {
		return
	}

	var payload struct {
		Path string `json:"path"`
		Type string `json:"type"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "invalid create payload", nil)
		return
	}
	if strings.TrimSpace(payload.Path) == "" {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "path is required", nil)
		return
	}

	if payload.Type == "directory" {
		if err := h.module.Service.CreateDirectory(c.Request.Context(), sandboxID, payload.Path); err != nil {
			h.writeError(c, err)
			return
		}
		apierrors.Success(c, http.StatusOK, "directory created", gin.H{"path": payload.Path})
		return
	}

	if err := h.module.Service.WriteFile(c.Request.Context(), sandboxID, payload.Path, ""); err != nil {
		h.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "file created", gin.H{"path": payload.Path})
}

func (h *Handler) DeleteSandboxPath(c *gin.Context) {
	sandboxID, ok := h.requireSandboxAccess(c)
	if !ok {
		return
	}

	var payload struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "invalid delete payload", nil)
		return
	}
	if strings.TrimSpace(payload.Path) == "" {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "path is required", nil)
		return
	}
	if err := h.module.Service.DeletePath(c.Request.Context(), sandboxID, payload.Path); err != nil {
		h.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "path deleted", gin.H{"path": payload.Path})
}

func (h *Handler) RenameSandboxPath(c *gin.Context) {
	sandboxID, ok := h.requireSandboxAccess(c)
	if !ok {
		return
	}

	var payload struct {
		OldPath string `json:"old_path"`
		NewPath string `json:"new_path"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "invalid rename payload", nil)
		return
	}
	if strings.TrimSpace(payload.OldPath) == "" || strings.TrimSpace(payload.NewPath) == "" {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "old_path and new_path are required", nil)
		return
	}
	if err := h.module.Service.RenamePath(c.Request.Context(), sandboxID, payload.OldPath, payload.NewPath); err != nil {
		h.writeError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "path renamed", gin.H{"old_path": payload.OldPath, "new_path": payload.NewPath})
}

func (h *Handler) writeError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code, message := apierrors.CodeOf(err)

	switch {
	case errors.Is(err, domain.ErrSandboxNotFound):
		status, code, message = http.StatusNotFound, apierrors.ErrCodeNotFound, "sandbox not found"
	case errors.Is(err, domain.ErrSandboxAccessDenied):
		status, code, message = http.StatusForbidden, apierrors.ErrCodeForbidden, "workspace access denied"
	case errors.Is(err, domain.ErrInvalidSandbox):
		status, code, message = http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error()
	}

	apierrors.Error(c, status, code, message, nil)
}

func (h *Handler) requireSandboxAccess(c *gin.Context) (string, bool) {
	sandboxID := c.Param("sandbox_id")
	if sandboxID == "" {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "sandbox_id is required", nil)
		return "", false
	}
	if err := h.module.Service.ValidateSandboxAccess(c.Request.Context(), userID(c), sandboxID); err != nil {
		h.writeError(c, err)
		return "", false
	}
	return sandboxID, true
}

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.GET("/projects/:project_id/ws", handler.ProjectEvents)
	router.GET("/projects/:project_id/terminals/shells", handler.TerminalShells)
	router.POST("/projects/:project_id/terminals", handler.CreateTerminal)
	router.DELETE("/projects/:project_id/terminals/:session_id", handler.CloseTerminal)
	router.GET("/projects/:project_id/terminals/:session_id/ws", handler.TerminalSessionWebSocket)
	router.GET("/projects/:project_id/sandbox", handler.GetSandboxByProject)
	router.GET("/sandboxes/:sandbox_id/files", handler.ListSandboxFiles)
	router.GET("/sandboxes/:sandbox_id/files/*path", handler.ReadSandboxFile)
	router.POST("/sandboxes/:sandbox_id/files/save", handler.SaveSandboxFile)
	router.POST("/sandboxes/:sandbox_id/files/create", handler.CreateSandboxPath)
	router.DELETE("/sandboxes/:sandbox_id/files", handler.DeleteSandboxPath)
	router.PATCH("/sandboxes/:sandbox_id/files/rename", handler.RenameSandboxPath)
}

func (h *Handler) ProjectEvents(c *gin.Context) {
	projectID := strings.TrimSpace(c.Param("project_id"))
	userID := userID(c)
	if projectID == "" || userID == "" {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "project_id is required", nil)
		return
	}
	if err := h.module.Service.ValidateProjectAccess(c.Request.Context(), userID, projectID); err != nil {
		apierrors.Error(c, http.StatusForbidden, apierrors.ErrCodeForbidden, "workspace access denied", nil)
		return
	}
	if h.module.Events == nil {
		apierrors.Error(c, http.StatusServiceUnavailable, apierrors.ErrCodeInternalServer, "realtime events are not configured", nil)
		return
	}

	connection, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		Subprotocols:   []string{"forgeai"},
		OriginPatterns: h.module.WebSocketOrigins,
	})
	if err != nil {
		return
	}
	defer connection.Close(websocket.StatusNormalClosure, "")

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()
	go func() {
		for {
			if _, _, err := connection.Read(ctx); err != nil {
				cancel()
				return
			}
		}
	}()

	events, unsubscribe := h.module.Events.Subscribe(projectID)
	defer unsubscribe()
	authorizationCheck := time.NewTicker(30 * time.Second)
	defer authorizationCheck.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-authorizationCheck.C:
			if err := h.module.Service.ValidateProjectAccess(ctx, userID, projectID); err != nil {
				_ = connection.Close(websocket.StatusPolicyViolation, "project access revoked")
				return
			}
		case event, open := <-events:
			if !open {
				return
			}
			if err := h.module.Service.ValidateProjectAccess(ctx, userID, projectID); err != nil {
				_ = connection.Close(websocket.StatusPolicyViolation, "project access revoked")
				return
			}
			if err := writeRealtimeEvent(ctx, connection, event); err != nil {
				return
			}
		}
	}
}

func writeRealtimeEvent(ctx context.Context, connection *websocket.Conn, event realtime.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return connection.Write(ctx, websocket.MessageText, data)
}

func userID(c *gin.Context) string {
	value, _ := middleware.UserID(c)
	return value
}
