package invite

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"ai-agent/internal/modules/organization/core"

	"github.com/jmoiron/sqlx"
)

type EmailSender interface {
	SendWorkspaceInvite(ctx context.Context, to, organizationName, inviterName, inviteLink, role string) error
}

type Service interface {
	Create(context.Context, string, string, CreateRequest) (Invite, error)
	List(context.Context, string, string, string) ([]Invite, error)
	GetByToken(context.Context, string) (Invite, error)
	Accept(context.Context, string, string) (string, error)
	Reject(context.Context, string, string) error
	UpdateStatus(context.Context, string, string, string) (StatusUpdateResult, error)
	ListPending(context.Context, string) ([]Invite, error)
}

type service struct {
	db            *sqlx.DB
	repo          Repository
	organizations core.Service
	email         EmailSender
	frontendURL   string
}

func NewService(db *sqlx.DB, repo Repository, organizations core.Service, email EmailSender, frontendURL string) Service {
	return &service{db: db, repo: repo, organizations: organizations, email: email, frontendURL: strings.TrimRight(frontendURL, "/")}
}

func (s *service) Create(ctx context.Context, userID, organizationID string, input CreateRequest) (Invite, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Role = strings.ToUpper(strings.TrimSpace(input.Role))
	if !strings.Contains(input.Email, "@") || (input.Role != RoleAdmin && input.Role != RoleMember) {
		return Invite{}, ErrInvalidInvite
	}
	if err := s.requireManager(ctx, userID, organizationID); err != nil {
		return Invite{}, err
	}
	organization, err := s.organizations.FindByID(ctx, organizationID)
	if err != nil {
		return Invite{}, err
	}
	token, err := newToken()
	if err != nil {
		return Invite{}, fmt.Errorf("generate invitation token: %w", err)
	}
	inviterName, err := s.repo.UserName(ctx, userID)
	if err != nil {
		return Invite{}, err
	}
	invite, err := s.repo.Create(ctx, Invite{
		OrganizationID:   organizationID,
		OrganizationName: organization.Name,
		Email:            input.Email,
		Role:             input.Role,
		InvitedBy:        userID,
		InviterName:      inviterName,
		ExpiresAt:        time.Now().UTC().Add(7 * 24 * time.Hour),
	}, hashToken(token))
	if err != nil {
		return Invite{}, err
	}
	if s.email != nil {
		link := s.frontendURL + "/accept-invite/" + url.PathEscape(token)
		if err := s.email.SendWorkspaceInvite(ctx, input.Email, organization.Name, inviterName, link, input.Role); err != nil {
			return Invite{}, fmt.Errorf("queue organization invitation email: %w", err)
		}
	}
	return invite, nil
}

func (s *service) List(ctx context.Context, userID, organizationID, status string) ([]Invite, error) {
	if err := s.requireManager(ctx, userID, organizationID); err != nil {
		return nil, err
	}
	status = strings.ToUpper(strings.TrimSpace(status))
	if status != "" && !validStatus(status) {
		return nil, ErrInvalidStatus
	}
	return s.repo.ListByOrganization(ctx, organizationID, status)
}

func (s *service) GetByToken(ctx context.Context, token string) (Invite, error) {
	if strings.TrimSpace(token) == "" {
		return Invite{}, ErrNotFound
	}
	return s.repo.FindByToken(ctx, hashToken(token))
}

func (s *service) Accept(ctx context.Context, userID, token string) (string, error) {
	if strings.TrimSpace(token) == "" || strings.TrimSpace(userID) == "" {
		return "", ErrNotFound
	}
	email, err := s.repo.UserEmail(ctx, userID)
	if err != nil {
		return "", err
	}
	return s.repo.Accept(ctx, hashToken(token), userID, email)
}

func (s *service) Reject(ctx context.Context, token, email string) error {
	if strings.TrimSpace(token) == "" || strings.TrimSpace(email) == "" {
		return ErrNotFound
	}
	return s.repo.Reject(ctx, hashToken(token), email)
}

func (s *service) UpdateStatus(ctx context.Context, userID, inviteID, status string) (StatusUpdateResult, error) {
	status = strings.ToUpper(strings.TrimSpace(status))
	switch status {
	case StatusAccepted:
		email, err := s.repo.UserEmail(ctx, userID)
		if err != nil {
			return StatusUpdateResult{}, err
		}
		organizationID, err := s.repo.AcceptByID(ctx, inviteID, userID, email)
		if err != nil {
			return StatusUpdateResult{}, err
		}
		return StatusUpdateResult{Status: StatusAccepted, OrganizationID: organizationID}, nil
	case StatusRejected:
		email, err := s.repo.UserEmail(ctx, userID)
		if err != nil {
			return StatusUpdateResult{}, err
		}
		if err := s.repo.RejectByID(ctx, inviteID, email); err != nil {
			return StatusUpdateResult{}, err
		}
		return StatusUpdateResult{Status: StatusRejected}, nil
	case StatusRevoked:
		organizationID, err := s.repo.OrganizationIDByID(ctx, inviteID)
		if err != nil {
			return StatusUpdateResult{}, err
		}
		if err := s.requireManager(ctx, userID, organizationID); err != nil {
			return StatusUpdateResult{}, err
		}
		if err := s.repo.Revoke(ctx, organizationID, inviteID); err != nil {
			return StatusUpdateResult{}, err
		}
		return StatusUpdateResult{Status: StatusRevoked}, nil
	default:
		return StatusUpdateResult{}, ErrInvalidStatus
	}
}

func (s *service) ListPending(ctx context.Context, userID string) ([]Invite, error) {
	email, err := s.repo.UserEmail(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListPendingByEmail(ctx, email)
}

func (s *service) requireManager(ctx context.Context, userID, organizationID string) error {
	var allowed bool
	err := s.db.GetContext(ctx, &allowed, `
		SELECT EXISTS (
			SELECT 1 FROM tbl_organization_member
			WHERE organization_id = $1 AND user_id = $2
			  AND deleted_at IS NULL AND role IN ('OWNER', 'ADMIN')
		)`, organizationID, userID)
	if err != nil {
		return fmt.Errorf("check organization invite permission: %w", err)
	}
	if !allowed {
		return ErrInviteForbidden
	}
	return nil
}

func newToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

var (
	ErrInvalidInvite   = errors.New("invitation email or role is invalid")
	ErrInviteForbidden = errors.New("organization owner or admin permission is required")
	ErrInvalidStatus   = errors.New("invalid invitation status transition")
)

type StatusUpdateResult struct {
	Status         string `json:"status"`
	OrganizationID string `json:"organization_id,omitempty"`
}

func validStatus(status string) bool {
	switch status {
	case StatusPending, StatusAccepted, StatusRejected, StatusExpired, StatusRevoked:
		return true
	default:
		return false
	}
}
