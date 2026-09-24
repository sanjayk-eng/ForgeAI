package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, tx *sqlx.Tx, email, name, passwordHash string) (string, error)
	CreateEmailVerification(ctx context.Context, tx *sqlx.Tx, userID, tokenHash string, expiresAt time.Time) error
	VerifyEmailToken(ctx context.Context, tx *sqlx.Tx, tokenHash string) error
	FindOAuthUserID(ctx context.Context, tx *sqlx.Tx, providerType ProviderType, providerUserID string) (string, error)
	FindUserIDByEmail(ctx context.Context, tx *sqlx.Tx, email string) (string, error)
	FindUserByID(ctx context.Context, userID string) (UserProfile, error)
	FindCredentials(ctx context.Context, email string) (UserCredentials, error)
	CreateOAuthUser(ctx context.Context, tx *sqlx.Tx, email, name string) (string, error)
	CreateOAuthAccount(ctx context.Context, tx *sqlx.Tx, providerType ProviderType, userID, providerUserID, accessToken string) error
	FindGitHubAccessToken(ctx context.Context, userID string) (string, error)
}

type repository struct {
	db *sqlx.DB
}

type UserCredentials struct {
	ID            string `db:"id"`
	PasswordHash  string `db:"password_hash"`
	EmailVerified bool   `db:"email_verified"`
}

func NewRepository(db *sqlx.DB) AuthRepository {
	return &repository{db: db}
}

func (repo *repository) FindOAuthUserID(ctx context.Context, tx *sqlx.Tx, providerType ProviderType, providerUserID string) (string, error) {
	var userID string
	err := tx.GetContext(ctx, &userID, `
		SELECT oa.user_id
		FROM tbl_oauth_account oa
		JOIN tbl_enum e ON e.id = oa.provider_id
		WHERE e.category = 'AUTH_PROVIDER' AND e.code = $1 AND oa.provider_user_id = $2`,
		strings.ToUpper(string(providerType)), providerUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", sql.ErrNoRows
	}
	if err != nil {
		return "", fmt.Errorf("find OAuth account: %w", err)
	}
	return userID, nil
}

func (repo *repository) FindGitHubAccessToken(ctx context.Context, userID string) (string, error) {
	var accessToken string
	err := repo.db.GetContext(ctx, &accessToken, `
		SELECT oa.access_token
		FROM tbl_oauth_account oa
		JOIN tbl_enum e ON e.id = oa.provider_id
		WHERE oa.user_id = $1 AND e.category = 'AUTH_PROVIDER' AND e.code = 'GITHUB'`, userID)
	if err != nil {
		return "", fmt.Errorf("find GitHub access token: %w", err)
	}
	if strings.TrimSpace(accessToken) == "" {
		return "", fmt.Errorf("GitHub account is not connected")
	}
	return accessToken, nil
}

func (repo *repository) FindUserIDByEmail(ctx context.Context, tx *sqlx.Tx, email string) (string, error) {
	var userID string
	err := tx.GetContext(ctx, &userID, `SELECT id FROM tbl_user WHERE LOWER(email) = LOWER($1)`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return "", sql.ErrNoRows
	}
	if err != nil {
		return "", fmt.Errorf("find user by email: %w", err)
	}
	return userID, nil
}

func (repo *repository) FindUserByID(ctx context.Context, userID string) (UserProfile, error) {
	var user UserProfile
	err := repo.db.GetContext(ctx, &user, `
		SELECT id, email, name
		FROM tbl_user
		WHERE id = $1`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return UserProfile{}, sql.ErrNoRows
	}
	if err != nil {
		return UserProfile{}, fmt.Errorf("find user by ID: %w", err)
	}
	return user, nil
}

func (repo *repository) FindCredentials(ctx context.Context, email string) (UserCredentials, error) {
	var credentials UserCredentials
	if err := repo.db.GetContext(ctx, &credentials, `
		SELECT id, password_hash, email_verified_at IS NOT NULL AS email_verified
		FROM tbl_user
		WHERE LOWER(email) = LOWER($1)`, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserCredentials{}, ErrInvalidCredentials
		}
		return UserCredentials{}, fmt.Errorf("find user credentials: %w", err)
	}
	if credentials.PasswordHash == "" {
		return UserCredentials{}, ErrInvalidCredentials
	}
	if !credentials.EmailVerified {
		return UserCredentials{}, ErrEmailNotVerified
	}
	return credentials, nil
}

func (repo *repository) CreateEmailVerification(ctx context.Context, tx *sqlx.Tx, userID, tokenHash string, expiresAt time.Time) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO tbl_email_verification (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`, userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("create email verification: %w", err)
	}
	return nil
}

func (repo *repository) VerifyEmailToken(ctx context.Context, tx *sqlx.Tx, tokenHash string) error {
	var userID string
	err := tx.GetContext(ctx, &userID, `
		SELECT user_id
		FROM tbl_email_verification
		WHERE token_hash = $1 AND expires_at > NOW()
		FOR UPDATE`, tokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalidVerificationToken
	}
	if err != nil {
		return fmt.Errorf("find email verification: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE tbl_user SET email_verified_at = NOW(), updated_at = NOW() WHERE id = $1`, userID); err != nil {
		return fmt.Errorf("verify email: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM tbl_email_verification WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("remove email verification: %w", err)
	}
	return nil
}

func (repo *repository) CreateUser(ctx context.Context, tx *sqlx.Tx, email, name, passwordHash string) (string, error) {
	var userID string
	if err := tx.GetContext(ctx, &userID, `
		INSERT INTO tbl_user (email, name, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id`, email, name, passwordHash); err != nil {
		var postgresErr *pq.Error
		if errors.As(err, &postgresErr) && postgresErr.Code == "23505" {
			return "", ErrEmailAlreadyExists
		}
		return "", fmt.Errorf("create user: %w", err)
	}
	return userID, nil
}

func (repo *repository) CreateOAuthUser(ctx context.Context, tx *sqlx.Tx, email, name string) (string, error) {
	var userID string
	if err := tx.GetContext(ctx, &userID, `
			INSERT INTO tbl_user (email, name, email_verified_at)
			VALUES ($1, $2, NOW())
			RETURNING id`, email, name); err != nil {
		return "", fmt.Errorf("create OAuth user: %w", err)
	}
	return userID, nil
}

func (repo *repository) CreateOAuthAccount(ctx context.Context, tx *sqlx.Tx, providerType ProviderType, userID, providerUserID, accessToken string) error {
	providerCode := strings.ToUpper(string(providerType))
	result, err := tx.ExecContext(ctx, `
			INSERT INTO tbl_oauth_account (user_id, provider_id, provider_user_id, access_token)
			SELECT $1, id, $3, $4
			FROM tbl_enum
			WHERE category = 'AUTH_PROVIDER' AND code = $2
			ON CONFLICT (provider_id, provider_user_id)
			DO UPDATE SET user_id = EXCLUDED.user_id, access_token = EXCLUDED.access_token, updated_at = NOW()
			`, userID, providerCode, providerUserID, accessToken)
	if err != nil {
		return fmt.Errorf("save oauth account: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check oauth account: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("auth provider %q is not configured", providerType)
	}
	return nil
}
