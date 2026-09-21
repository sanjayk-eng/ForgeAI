package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, tx *sqlx.Tx, email, name, passwordHash string) (string, error)
	FindOAuthUserID(ctx context.Context, tx *sqlx.Tx, providerType ProviderType, providerUserID string) (string, error)
	FindUserIDByEmail(ctx context.Context, tx *sqlx.Tx, email string) (string, error)
	FindCredentials(ctx context.Context, email string) (UserCredentials, error)
	CreateOAuthUser(ctx context.Context, tx *sqlx.Tx, email, name string) (string, error)
	CreateOAuthAccount(ctx context.Context, tx *sqlx.Tx, providerType ProviderType, userID, providerUserID string) error
}

type repository struct {
	db *sqlx.DB
}

type UserCredentials struct {
	ID           string `db:"id"`
	PasswordHash string `db:"password_hash"`
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

func (repo *repository) FindUserIDByEmail(ctx context.Context, tx *sqlx.Tx, email string) (string, error) {
	var userID string
	err := tx.GetContext(ctx, &userID, `SELECT id FROM tbl_user WHERE email = $1`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return "", sql.ErrNoRows
	}
	if err != nil {
		return "", fmt.Errorf("find user by email: %w", err)
	}
	return userID, nil
}

func (repo *repository) FindCredentials(ctx context.Context, email string) (UserCredentials, error) {
	var credentials UserCredentials
	if err := repo.db.GetContext(ctx, &credentials, `
		SELECT id, password_hash
		FROM tbl_user
		WHERE email = $1`, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserCredentials{}, ErrInvalidCredentials
		}
		return UserCredentials{}, fmt.Errorf("find user credentials: %w", err)
	}
	if credentials.PasswordHash == "" {
		return UserCredentials{}, ErrInvalidCredentials
	}
	return credentials, nil
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
			INSERT INTO tbl_user (email, name)
			VALUES ($1, $2)
			RETURNING id`, email, name); err != nil {
		return "", fmt.Errorf("create OAuth user: %w", err)
	}
	return userID, nil
}

func (repo *repository) CreateOAuthAccount(ctx context.Context, tx *sqlx.Tx, providerType ProviderType, userID, providerUserID string) error {
	providerCode := strings.ToUpper(string(providerType))
	result, err := tx.ExecContext(ctx, `
			INSERT INTO tbl_oauth_account (user_id, provider_id, provider_user_id)
			SELECT $1, id, $3
			FROM tbl_enum
			WHERE category = 'AUTH_PROVIDER' AND code = $2
			`, userID, providerCode, providerUserID)
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
