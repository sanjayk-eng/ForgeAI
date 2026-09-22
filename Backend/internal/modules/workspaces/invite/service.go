package invite

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"ai-agent/internal/shared/email"
	"ai-agent/internal/shared/logger"
	appdatabase "ai-agent/pkg/database"

	"github.com/jmoiron/sqlx"
)

var ErrInvalidInviteInput = errors.New("invalid invite input")
var ErrInviteNotFound = errors.New("invite not found")

type Service struct {
	db          *sqlx.DB
	repo        Repository
	email       *email.Service
	frontendURL string
	logger      logger.Logger
}

func NewService(db *sqlx.DB, emailService *email.Service, frontendURL string, appLogger logger.Logger) *Service {
	return &Service{db: db, repo: NewRepository(db), email: emailService, frontendURL: strings.TrimRight(frontendURL, "/"), logger: appLogger}
}

func (service *Service) CreateInvite(ctx context.Context, workspaceID, inviterID, emailAddress, role string) (Invite, error) {
	if service == nil || service.db == nil || strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(inviterID) == "" {
		return Invite{}, ErrInvalidInviteInput
	}
	if strings.TrimSpace(emailAddress) == "" || !strings.Contains(emailAddress, "@") {
		return Invite{}, ErrInvalidInviteInput
	}
	role = strings.TrimSpace(role)
	if role == "" {
		role = "MEMBER"
	}

	token, err := generateToken()
	if err != nil {
		return Invite{}, fmt.Errorf("generate invite token: %w", err)
	}
	invite := Invite{
		WorkspaceID: workspaceID,
		Email:       strings.TrimSpace(emailAddress),
		Role:        strings.ToUpper(role),
		Status:      InvitePending,
		InvitedBy:   inviterID,
		TokenHash:   token,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	var created Invite
	if err := appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		var err error
		created, err = service.repo.Create(ctx, tx, invite)
		return err
	}); err != nil {
		return Invite{}, fmt.Errorf("persist workspace invite: %w", err)
	}
	invite = created

	// Send workspace invite email asynchronously
	if service.email != nil {
		inviteLink := fmt.Sprintf("%s/workspace/members?workspace=%s&invite=%s", service.frontendURL, workspaceID, invite.ID)
		if err := service.email.SendWorkspaceInvite(ctx, invite.Email, "Workspace", inviterID, inviteLink, invite.Role); err != nil && service.logger != nil {
			service.logger.Warn(ctx, "failed to queue workspace invite email", "invite_id", invite.ID, "error", err)
		}
	}

	return invite, nil
}

func (service *Service) ListInvites(ctx context.Context, workspaceID string) ([]Invite, error) {
	if service == nil || service.db == nil || strings.TrimSpace(workspaceID) == "" {
		return nil, ErrInvalidInviteInput
	}
	return service.repo.ListByWorkspace(ctx, strings.TrimSpace(workspaceID))
}

func (service *Service) AcceptInvite(ctx context.Context, workspaceID, inviteID, userID string) (Invite, error) {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(inviteID) == "" || strings.TrimSpace(userID) == "" {
		return Invite{}, ErrInvalidInviteInput
	}
	return service.updateStatus(ctx, workspaceID, inviteID, InviteAccepted)
}

func (service *Service) RejectInvite(ctx context.Context, workspaceID, inviteID, userID string) (Invite, error) {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(inviteID) == "" || strings.TrimSpace(userID) == "" {
		return Invite{}, ErrInvalidInviteInput
	}
	return service.updateStatus(ctx, workspaceID, inviteID, InviteRejected)
}

func (service *Service) UpdateInviteStatus(ctx context.Context, workspaceID, inviteID, userID, status string) (Invite, error) {
	switch InviteStatus(strings.ToUpper(strings.TrimSpace(status))) {
	case InviteAccepted:
		return service.AcceptInvite(ctx, workspaceID, inviteID, userID)
	case InviteRejected:
		return service.RejectInvite(ctx, workspaceID, inviteID, userID)
	default:
		return Invite{}, ErrInvalidInviteInput
	}
}

func (service *Service) updateStatus(ctx context.Context, workspaceID, inviteID string, status InviteStatus) (Invite, error) {
	if service == nil || service.db == nil {
		return Invite{}, ErrInvalidInviteInput
	}
	var updated Invite
	if err := appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		var err error
		updated, err = service.repo.UpdateStatus(ctx, tx, workspaceID, inviteID, string(status))
		return err
	}); err != nil {
		return Invite{}, fmt.Errorf("update invite status: %w", err)
	}
	return updated, nil
}

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (service *Service) RevokeInvite(ctx context.Context, workspaceID, inviteID, userID string) error {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(inviteID) == "" || strings.TrimSpace(userID) == "" {
		return ErrInvalidInviteInput
	}
	return nil
}
