package terminal

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"ai-agent/internal/modules/terminal/domain"
	apierrors "ai-agent/internal/shared/errors"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
)

type terminalClientMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

type terminalServerMessage struct {
	Type    string               `json:"type"`
	Data    string               `json:"data,omitempty"`
	Code    int                  `json:"code,omitempty"`
	Message string               `json:"message,omitempty"`
	Session *TerminalSessionInfo `json:"session,omitempty"`
}

func (h *Handler) TerminalShells(c *gin.Context) {
	shells, err := h.module.Service.TerminalShells(c.Request.Context(), userID(c), c.Param("project_id"))
	if err != nil {
		h.writeTerminalError(c, err)
		return
	}
	apierrors.Success(c, http.StatusOK, "terminal shells fetched", gin.H{"shells": shells})
}

func (h *Handler) CreateTerminal(c *gin.Context) {
	var payload struct {
		Shell string `json:"shell"`
		Cols  int    `json:"cols"`
		Rows  int    `json:"rows"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil || strings.TrimSpace(payload.Shell) == "" {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "a supported shell is required", nil)
		return
	}
	if h.module.Sessions == nil {
		apierrors.Error(c, http.StatusServiceUnavailable, apierrors.ErrCodeInternalServer, "terminal sessions are unavailable", nil)
		return
	}
	session, err := h.module.Sessions.Create(c.Request.Context(), userID(c), c.Param("project_id"), payload.Shell, payload.Cols, payload.Rows)
	if err != nil {
		h.writeTerminalError(c, err)
		return
	}
	apierrors.Success(c, http.StatusCreated, "terminal created", session)
}

func (h *Handler) CloseTerminal(c *gin.Context) {
	projectID := strings.TrimSpace(c.Param("project_id"))
	sessionID := strings.TrimSpace(c.Param("session_id"))
	userID := userID(c)
	if err := h.module.Service.ValidateProjectAccess(c.Request.Context(), userID, projectID); err != nil {
		h.writeTerminalError(c, err)
		return
	}
	if h.module.Sessions == nil || !h.module.Sessions.CloseOwned(sessionID, userID, projectID) {
		apierrors.Error(c, http.StatusNotFound, apierrors.ErrCodeNotFound, "terminal session not found", nil)
		return
	}
	apierrors.Success(c, http.StatusOK, "terminal closed", gin.H{"session_id": sessionID})
}

func (h *Handler) TerminalSessionWebSocket(c *gin.Context) {
	projectID := strings.TrimSpace(c.Param("project_id"))
	sessionID := strings.TrimSpace(c.Param("session_id"))
	userID := userID(c)
	if projectID == "" || sessionID == "" || userID == "" {
		apierrors.Error(c, http.StatusBadRequest, apierrors.ErrCodeValidation, "project_id and session_id are required", nil)
		return
	}
	if err := h.module.Service.ValidateProjectAccess(c.Request.Context(), userID, projectID); err != nil {
		h.writeTerminalError(c, err)
		return
	}
	if h.module.Sessions == nil {
		apierrors.Error(c, http.StatusServiceUnavailable, apierrors.ErrCodeInternalServer, "terminal sessions are unavailable", nil)
		return
	}
	session, err := h.module.Sessions.Attach(sessionID, userID, projectID)
	if err != nil {
		status := http.StatusNotFound
		if errors.Is(err, errTerminalSessionAttached) {
			status = http.StatusConflict
		}
		apierrors.Error(c, status, apierrors.ErrCodeNotFound, err.Error(), nil)
		return
	}
	connection, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		Subprotocols:   []string{"forgeai"},
		OriginPatterns: h.module.WebSocketOrigins,
	})
	if err != nil {
		h.module.Sessions.Detach(session)
		return
	}
	defer connection.Close(websocket.StatusNormalClosure, "")
	defer h.module.Sessions.Detach(session)
	connection.SetReadLimit(64 * 1024)

	info := session.snapshot()
	if err := writeTerminalMessage(c.Request.Context(), connection, terminalServerMessage{Type: "session", Session: &info}); err != nil {
		return
	}

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()
	incoming := make(chan terminalClientMessage, 8)
	readErrors := make(chan error, 1)
	go readTerminalMessages(ctx, connection, incoming, readErrors)
	authorizationCheck := time.NewTicker(30 * time.Second)
	defer authorizationCheck.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case err := <-readErrors:
			if err != nil {
				return
			}
		case message := <-incoming:
			switch message.Type {
			case "input":
				if len(message.Data) > 32*1024 {
					if writeTerminalMessage(ctx, connection, terminalServerMessage{Type: "error", Message: "terminal input is too large"}) != nil {
						return
					}
					continue
				}
				if err := session.write([]byte(message.Data)); err != nil {
					_ = writeTerminalMessage(ctx, connection, terminalServerMessage{Type: "error", Message: "terminal input failed"})
					return
				}
			case "resize":
				if err := session.resize(message.Cols, message.Rows); err != nil {
					if writeTerminalMessage(ctx, connection, terminalServerMessage{Type: "error", Message: err.Error()}) != nil {
						return
					}
				}
			case "close":
				h.module.Sessions.Close(sessionID)
				_ = writeTerminalMessage(ctx, connection, terminalServerMessage{Type: "exit", Code: 0})
				return
			default:
				if writeTerminalMessage(ctx, connection, terminalServerMessage{Type: "error", Message: "unknown terminal message type"}) != nil {
					return
				}
			}
		case output := <-session.output:
			if err := writeTerminalMessage(ctx, connection, terminalServerMessage{Type: "output", Data: string(output)}); err != nil {
				return
			}
		case <-session.done:
			for {
				select {
				case output := <-session.output:
					if writeTerminalMessage(ctx, connection, terminalServerMessage{Type: "output", Data: string(output)}) != nil {
						return
					}
				default:
					session.mu.Lock()
					exitCode, exitErr := session.exitCode, session.exitErr
					session.mu.Unlock()
					if exitErr != nil {
						_ = writeTerminalMessage(ctx, connection, terminalServerMessage{Type: "error", Message: "terminal process ended unexpectedly"})
					}
					_ = writeTerminalMessage(ctx, connection, terminalServerMessage{Type: "exit", Code: exitCode})
					return
				}
			}
		case <-authorizationCheck.C:
			if err := h.module.Service.ValidateProjectAccess(ctx, userID, projectID); err != nil {
				_ = writeTerminalMessage(ctx, connection, terminalServerMessage{Type: "error", Message: "project access was revoked"})
				h.module.Sessions.Close(sessionID)
				_ = connection.Close(websocket.StatusPolicyViolation, "project access revoked")
				return
			}
		}
	}
}

func readTerminalMessages(ctx context.Context, connection *websocket.Conn, incoming chan<- terminalClientMessage, readErrors chan<- error) {
	for {
		messageType, payload, err := connection.Read(ctx)
		if err != nil {
			select {
			case readErrors <- err:
			case <-ctx.Done():
			}
			return
		}
		if messageType != websocket.MessageText {
			continue
		}
		var message terminalClientMessage
		if err := json.Unmarshal(payload, &message); err != nil {
			message.Type = "invalid"
		}
		select {
		case incoming <- message:
		case <-ctx.Done():
			return
		}
	}
}

func writeTerminalMessage(ctx context.Context, connection *websocket.Conn, message terminalServerMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return connection.Write(ctx, websocket.MessageText, data)
}

func (h *Handler) writeTerminalError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code := apierrors.ErrCodeInternalServer
	message := "terminal operation failed"
	switch {
	case errors.Is(err, domain.ErrSandboxNotFound):
		status, code, message = http.StatusNotFound, apierrors.ErrCodeNotFound, "project sandbox not found"
	case errors.Is(err, domain.ErrSandboxAccessDenied):
		status, code, message = http.StatusForbidden, apierrors.ErrCodeForbidden, "workspace access denied"
	case errors.Is(err, domain.ErrTerminalShellUnavailable):
		status, code, message = http.StatusBadRequest, apierrors.ErrCodeValidation, "selected shell is unavailable"
	case errors.Is(err, domain.ErrSandboxNotRunning):
		status, code, message = http.StatusServiceUnavailable, apierrors.ErrCodeInternalServer, "project sandbox is not running"
	case errors.Is(err, domain.ErrInvalidSandbox):
		status, code, message = http.StatusBadRequest, apierrors.ErrCodeValidation, err.Error()
	default:
		message = err.Error()
	}
	apierrors.Error(c, status, code, message, nil)
}
