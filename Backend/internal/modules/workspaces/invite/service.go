package invite

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"ai-agent/internal/shared/logger"
	appdatabase "ai-agent/pkg/database"

	"github.com/jmoiron/sqlx"
)

var (
	ErrInvalidInviteInput      = errors.New("invalid invite input")
	ErrInviteNotFound          = errors.New("invite not found")
	ErrInviteExpired           = errors.New("invite has expired")
	ErrInviteAlreadyUsed       = errors.New("invite has already been used")
	ErrInvalidToken            = errors.New("invalid invite token")
	ErrEmailMismatch           = errors.New("email does not match invite")
	ErrInviteAlreadySent       = errors.New("pending invite already exists for this email")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

// allowedTransitions defines valid state transitions (industry best practice)
var allowedTransitions = map[InviteStatus][]InviteStatus{
	InvitePending: {InviteAccepted, InviteRejected, InviteExpired, InviteRevoked},
	// Terminal states - no transitions allowed
	InviteAccepted: {},
	InviteRejected: {},
	InviteExpired:  {},
	InviteRevoked:  {},
}

type Service struct {
	db          *sqlx.DB
	repo        Repository
	email       EmailService
	frontendURL string
	logger      logger.Logger
}

func NewService(db *sqlx.DB, emailService EmailService, frontendURL string, appLogger logger.Logger) *Service {
	return &Service{
		db:          db,
		repo:        NewRepository(db),
		email:       emailService,
		frontendURL: strings.TrimRight(frontendURL, "/"),
		logger:      appLogger,
	}
}

// CreateInvite creates and sends a workspace invitation
func (service *Service) CreateInvite(ctx context.Context, workspaceID, inviterID, emailAddress, role string) (Invite, error) {
	if service == nil || service.db == nil || strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(inviterID) == "" {
		return Invite{}, ErrInvalidInviteInput
	}
	emailAddress = strings.TrimSpace(strings.ToLower(emailAddress))
	if emailAddress == "" || !strings.Contains(emailAddress, "@") {
		return Invite{}, ErrInvalidInviteInput
	}
	role = strings.TrimSpace(role)
	if role == "" {
		role = "MEMBER"
	}

	// Check for existing pending invites (prevent duplicates - best practice)
	existingInvites, err := service.repo.FindByWorkspaceAndEmail(ctx, workspaceID, emailAddress)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Invite{}, fmt.Errorf("check existing invites: %w", err)
	}
	
	for _, existing := range existingInvites {
		if existing.Status == InvitePending {
			return Invite{}, ErrInviteAlreadySent
		}
		// REJECTED, EXPIRED, REVOKED are OK - allow re-invite
	}

	token, err := generateToken()
	if err != nil {
		return Invite{}, fmt.Errorf("generate invite token: %w", err)
	}

	invite := Invite{
		WorkspaceID: workspaceID,
		Email:       emailAddress,
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
		workspaceName, inviterName := service.getInviteDetails(ctx, workspaceID, inviterID)
		inviteLink := fmt.Sprintf("%s/accept-invite/%s", service.frontendURL, invite.TokenHash)
		
		if err := service.email.SendWorkspaceInvite(ctx, invite.Email, workspaceName, inviterName, inviteLink, invite.Role); err != nil && service.logger != nil {
			service.logger.Warn(ctx, "failed to queue workspace invite email", "invite_id", invite.ID, "error", err)
		}
	}

	return invite, nil
}

// ListInvites returns all invites for a workspace
func (service *Service) ListInvites(ctx context.Context, workspaceID string) ([]Invite, error) {
	if service == nil || service.db == nil || strings.TrimSpace(workspaceID) == "" {
		return nil, ErrInvalidInviteInput
	}
	return service.repo.ListByWorkspace(ctx, strings.TrimSpace(workspaceID))
}

// GetInviteByToken retrieves an invite by its token (public endpoint)
func (service *Service) GetInviteByToken(ctx context.Context, token string) (Invite, error) {
	if strings.TrimSpace(token) == "" {
		return Invite{}, ErrInvalidToken
	}
	
	invite, err := service.repo.FindByToken(ctx, strings.TrimSpace(token))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Invite{}, ErrInviteNotFound
		}
		return Invite{}, fmt.Errorf("find invite by token: %w", err)
	}
	
	return invite, nil
}

// AcceptInviteByToken accepts an invite using token (authenticated)
func (service *Service) AcceptInviteByToken(ctx context.Context, token, userID, userEmail string) (Invite, error) {
	token = strings.TrimSpace(token)
	userID = strings.TrimSpace(userID)
	userEmail = strings.TrimSpace(strings.ToLower(userEmail))

	if token == "" || userID == "" || userEmail == "" {
		return Invite{}, ErrInvalidInviteInput
	}

	// Get invite by token
	invite, err := service.repo.FindByToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Invite{}, ErrInviteNotFound
		}
		return Invite{}, fmt.Errorf("find invite: %w", err)
	}

	// Validate invite (auto-expires if needed)
	if err := service.validateInvite(invite, userEmail); err != nil {
		return Invite{}, err
	}

	// Check state transition
	if !service.canTransition(invite.Status, InviteAccepted) {
		return Invite{}, ErrInvalidStatusTransition
	}

	// Accept invite and add member in transaction
	var updated Invite
	if err := appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		// Update invite status
		now := time.Now()
		updated, err = service.repo.UpdateStatusWithTimestamp(ctx, tx, invite.WorkspaceID, invite.ID, string(InviteAccepted), &now)
		if err != nil {
			return fmt.Errorf("update invite status: %w", err)
		}

		// Add user as workspace member
		if err := service.repo.AddWorkspaceMember(ctx, tx, invite.WorkspaceID, userID, invite.Role); err != nil {
			return fmt.Errorf("add workspace member: %w", err)
		}

		return nil
	}); err != nil {
		return Invite{}, err
	}

	if service.logger != nil {
		service.logger.Info(ctx, "invite accepted", "invite_id", invite.ID, "user_id", userID, "workspace_id", invite.WorkspaceID)
	}

	return updated, nil
}

// RejectInviteByToken rejects an invite using token
func (service *Service) RejectInviteByToken(ctx context.Context, token, userEmail string) (Invite, error) {
	token = strings.TrimSpace(token)
	userEmail = strings.TrimSpace(strings.ToLower(userEmail))

	if token == "" || userEmail == "" {
		return Invite{}, ErrInvalidInviteInput
	}

	// Get invite by token
	invite, err := service.repo.FindByToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Invite{}, ErrInviteNotFound
		}
		return Invite{}, fmt.Errorf("find invite: %w", err)
	}

	// Validate email matches
	if invite.Email != userEmail {
		return Invite{}, ErrEmailMismatch
	}

	// Check state transition
	if !service.canTransition(invite.Status, InviteRejected) {
		return Invite{}, ErrInvalidStatusTransition
	}

	// Update status
	var updated Invite
	if err := appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		now := time.Now()
		updated, err = service.repo.UpdateStatusWithTimestamp(ctx, tx, invite.WorkspaceID, invite.ID, string(InviteRejected), &now)
		return err
	}); err != nil {
		return Invite{}, fmt.Errorf("reject invite: %w", err)
	}

	if service.logger != nil {
		service.logger.Info(ctx, "invite rejected", "invite_id", invite.ID, "email", userEmail)
	}

	return updated, nil
}

// RevokeInvite revokes a pending invite (admin action)
func (service *Service) RevokeInvite(ctx context.Context, workspaceID, inviteID, userID string) error {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(inviteID) == "" || strings.TrimSpace(userID) == "" {
		return ErrInvalidInviteInput
	}

	// Get current invite
	invite, err := service.repo.FindByID(ctx, workspaceID, inviteID)
	if err != nil {
		return fmt.Errorf("find invite: %w", err)
	}

	// Check state transition
	if !service.canTransition(invite.Status, InviteRevoked) {
		return ErrInvalidStatusTransition
	}

	if err := appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		now := time.Now()
		_, err := service.repo.UpdateStatusWithTimestamp(ctx, tx, workspaceID, inviteID, string(InviteRevoked), &now)
		return err
	}); err != nil {
		return fmt.Errorf("revoke invite: %w", err)
	}

	return nil
}

// validateInvite checks if invite is valid for acceptance (Option A: check on access)
func (service *Service) validateInvite(invite Invite, userEmail string) error {
	// Check if email matches
	if invite.Email != userEmail {
		return ErrEmailMismatch
	}

	// Check if already used (not pending)
	if invite.Status != InvitePending {
		return ErrInviteAlreadyUsed
	}

	// Auto-expire if expired (Option A from best practices)
	if time.Now().After(invite.ExpiresAt) {
		go service.expireInvite(context.Background(), invite.ID, invite.WorkspaceID)
		return ErrInviteExpired
	}

	return nil
}

// expireInvite updates an invite to expired status
func (service *Service) expireInvite(ctx context.Context, inviteID, workspaceID string) {
	if err := appdatabase.WithTx(ctx, service.db, func(ctx context.Context, tx *sqlx.Tx) error {
		now := time.Now()
		_, err := service.repo.UpdateStatusWithTimestamp(ctx, tx, workspaceID, inviteID, string(InviteExpired), &now)
		return err
	}); err != nil && service.logger != nil {
		service.logger.Error(ctx, "failed to auto-expire invite", "invite_id", inviteID, "error", err)
	}
}

// canTransition checks if a status transition is allowed (best practice)
func (service *Service) canTransition(from, to InviteStatus) bool {
	allowed := allowedTransitions[from]
	for _, status := range allowed {
		if status == to {
			return true
		}
	}
	return false
}

// getInviteDetails fetches workspace and inviter names for email
func (service *Service) getInviteDetails(ctx context.Context, workspaceID, inviterID string) (workspaceName, inviterName string) {
	workspaceName = "Workspace" // default
	inviterName = "Team member" // default

	// Try to get workspace name
	if ws, err := service.repo.GetWorkspaceInfo(ctx, workspaceID); err == nil {
		workspaceName = ws.Name
	}

	// Try to get inviter name
	if user, err := service.repo.GetUserInfo(ctx, inviterID); err == nil {
		inviterName = user.Name
	}

	return workspaceName, inviterName
}

// generateToken generates a secure random token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
