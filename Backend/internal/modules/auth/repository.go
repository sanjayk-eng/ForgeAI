package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type AuthRepository interface {
	SaveUser(ctx context.Context, tx *sqlx.Tx, user OAuthUser) (string, error)
	SaveOAuthAccount(ctx context.Context, tx *sqlx.Tx, providerType ProviderType, userID, providerUserID string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) AuthRepository {
	return &repository{db: db}
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
