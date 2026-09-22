package invite

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"ai-agent/internal/shared/email"
	"ai-agent/internal/shared/templates"
	"ai-agent/internal/shared/worker"

	"github.com/jmoiron/sqlx"
)

var ErrInvalidInviteInput = errors.New("invalid invite input")
var ErrInviteNotFound = errors.New("invite not found")

type Service struct {
	db    *sqlx.DB
	email email.Sender
	queue worker.Worker
}

func NewService(db *sqlx.DB, sender email.Sender, queue worker.Worker) *Service {
	return &Service{db: db, email: sender, queue: queue}
}

func (service *Service) CreateInvite(ctx context.Context, workspaceID, inviterID, emailAddress, role string) (Invite, error) {
	if service == nil || strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(inviterID) == "" {
		return Invite{}, ErrInvalidInviteInput
	}
	if strings.TrimSpace(emailAddress) == "" || !strings.Contains(emailAddress, "@") {
		return Invite{}, ErrInvalidInviteInput
	}
	role = strings.TrimSpace(role)
	if role == "" {
		role = "MEMBER"
	}

	invite := Invite{
		WorkspaceID: workspaceID,
		Email:       strings.TrimSpace(emailAddress),
		Role:        strings.ToUpper(role),
		Status:      InvitePending,
		InvitedBy:   inviterID,
		TokenHash:   fmt.Sprintf("invite-%s-%d", workspaceID, time.Now().UnixNano()),
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if service.email != nil && service.queue != nil {
		subject, text, html := templates.WorkspaceInviteTemplate(templates.WorkspaceInviteTemplateData{
			WorkspaceName: "Workspace",
			InviteLink:    fmt.Sprintf("/workspaces/%s/invites/%s/accept", workspaceID, invite.TokenHash),
			InviterName:   inviterID,
			Role:          invite.Role,
		})
		job := func(jobCtx context.Context) error {
			return service.email.Send(jobCtx, email.Message{
				To:      invite.Email,
				Subject: subject,
				Text:    text,
				HTML:    html,
			})
		}
		if err := service.queue.Enqueue(ctx, job); err != nil {
			return Invite{}, err
		}
	}

	return invite, nil
}

func (service *Service) ListInvites(ctx context.Context, workspaceID string) ([]Invite, error) {
	if strings.TrimSpace(workspaceID) == "" {
		return nil, ErrInvalidInviteInput
	}
	return nil, nil
}

func (service *Service) AcceptInvite(ctx context.Context, workspaceID, inviteID, userID string) error {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(inviteID) == "" || strings.TrimSpace(userID) == "" {
		return ErrInvalidInviteInput
	}
	return nil
}

func (service *Service) RejectInvite(ctx context.Context, workspaceID, inviteID, userID string) error {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(inviteID) == "" || strings.TrimSpace(userID) == "" {
		return ErrInvalidInviteInput
	}
	return nil
}

func (service *Service) UpdateInviteStatus(ctx context.Context, workspaceID, inviteID, userID, status string) error {
	switch InviteStatus(strings.ToUpper(strings.TrimSpace(status))) {
	case InviteAccepted:
		return service.AcceptInvite(ctx, workspaceID, inviteID, userID)
	case InviteRejected:
		return service.RejectInvite(ctx, workspaceID, inviteID, userID)
	default:
		return ErrInvalidInviteInput
	}
}

func (service *Service) RevokeInvite(ctx context.Context, workspaceID, inviteID, userID string) error {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(inviteID) == "" || strings.TrimSpace(userID) == "" {
		return ErrInvalidInviteInput
	}
	return nil
}
