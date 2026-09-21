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
	FindCredentials(ctx context.Context, email string) (UserCredentials, error)
	SaveUser(ctx context.Context, tx *sqlx.Tx, user OAuthUser) (string, error)
	SaveOAuthAccount(ctx context.Context, tx *sqlx.Tx, providerType ProviderType, userID, providerUserID string) error
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

func (repo *repository) SaveUser(ctx context.Context, tx *sqlx.Tx, user OAuthUser) (string, error) {
	var userID string
	if err := tx.GetContext(ctx, &userID, `
			INSERT INTO tbl_user (email, name)
			VALUES ($1, $2)
			ON CONFLICT (email) DO UPDATE SET
				name = EXCLUDED.name,
				updated_at = NOW()
			RETURNING id`, user.Email, user.Name); err != nil {
		return "", fmt.Errorf("save user: %w", err)
	}
	return userID, nil
}

func (repo *repository) SaveOAuthAccount(ctx context.Context, tx *sqlx.Tx, providerType ProviderType, userID, providerUserID string) error {
	providerCode := strings.ToUpper(string(providerType))
	result, err := tx.ExecContext(ctx, `
			INSERT INTO tbl_oauth_account (user_id, provider_id, provider_user_id)
			SELECT $1, id, $3
			FROM tbl_enum
			WHERE category = 'AUTH_PROVIDER' AND code = $2
			ON CONFLICT (provider_id, provider_user_id) DO UPDATE SET
				user_id = EXCLUDED.user_id,
				updated_at = NOW()`, userID, providerCode, providerUserID)
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
