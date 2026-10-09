package integration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrStateInvalid              = errors.New("installation state is invalid or expired")
	ErrIntegrationNotFound       = errors.New("GitHub integration not found")
	ErrProjectNotFound           = errors.New("project not found")
	ErrExportRepositoryNotFound  = errors.New("GitHub export repository not found")
)

type Repository interface {
	CreateState(ctx context.Context, stateHash, organizationID, userID string, expiresAt time.Time) error
	FindState(ctx context.Context, stateHash string) (State, error)
	CompleteInstallation(ctx context.Context, stateHash string, installation Installation) (Integration, error)
	ListGitHub(ctx context.Context, organizationID string) ([]Integration, error)
	FindExportContext(ctx context.Context, organizationID, projectID string) (ExportContext, error)
	SaveExportRepository(ctx context.Context, projectID, integrationID string, repository ExportRepository, userID string) (ExportRepository, error)
	FindExportRepository(ctx context.Context, organizationID, projectID string) (ExportRepository, error)
	FindExportCredentials(ctx context.Context, organizationID, projectID string) (string, string, error)
	DisconnectGitHub(ctx context.Context, organizationID, integrationID string) error
	SetInstallationStatus(ctx context.Context, installationID, status string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateState(ctx context.Context, stateHash, organizationID, userID string, expiresAt time.Time) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM tbl_github_installation_state WHERE expires_at <= NOW()`); err != nil {
		return fmt.Errorf("remove expired GitHub installation states: %w", err)
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tbl_github_installation_state (state_hash, organization_id, user_id, expires_at)
		VALUES ($1, $2, $3, $4)
	`, stateHash, organizationID, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("create GitHub installation state: %w", err)
	}
	return nil
}

func (r *repository) FindState(ctx context.Context, stateHash string) (State, error) {
	var state State
	err := r.db.GetContext(ctx, &state, `
		SELECT organization_id, user_id
		FROM tbl_github_installation_state
		WHERE state_hash = $1 AND expires_at > NOW()
	`, stateHash)
	if errors.Is(err, sql.ErrNoRows) {
		return State{}, ErrStateInvalid
	}
	if err != nil {
		return State{}, fmt.Errorf("find GitHub installation state: %w", err)
	}
	return state, nil
}

func (r *repository) CompleteInstallation(ctx context.Context, stateHash string, installation Installation) (Integration, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return Integration{}, fmt.Errorf("begin GitHub installation transaction: %w", err)
	}
	defer tx.Rollback()

	var organizationID, userID string
	var consumedAt sql.NullTime
	var previousIntegrationID sql.NullString
	err = tx.QueryRowxContext(ctx, `
		SELECT organization_id, user_id, consumed_at, integration_id
		FROM tbl_github_installation_state
		WHERE state_hash = $1 AND expires_at > NOW()
		FOR UPDATE
	`, stateHash).Scan(&organizationID, &userID, &consumedAt, &previousIntegrationID)
	if errors.Is(err, sql.ErrNoRows) {
		return Integration{}, ErrStateInvalid
	}
	if err != nil {
		return Integration{}, fmt.Errorf("lock GitHub installation state: %w", err)
	}
	if consumedAt.Valid {
		var existing Integration
		err := tx.GetContext(ctx, &existing, integrationSelect+`
			WHERE i.id = $1 AND p.code = 'GITHUB'
		`, previousIntegrationID.String)
		if errors.Is(err, sql.ErrNoRows) {
			return Integration{}, ErrStateInvalid
		}
		if err != nil {
			return Integration{}, fmt.Errorf("load previously completed GitHub integration: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return Integration{}, fmt.Errorf("commit idempotent GitHub callback: %w", err)
		}
		return existing, nil
	}

	var result Integration
	err = tx.GetContext(ctx, &result, `
		INSERT INTO tbl_organization_integration (
			organization_id, provider_id, external_account_id, external_account_name,
			external_account_type, installation_id, status, connected_by,
			connected_at, disconnected_at, provider_metadata
		)
		SELECT $1, p.id, $2, $3, $4, $5, 'ACTIVE', $6, NOW(), NULL, '{}'::jsonb
		FROM tbl_integration_provider p
		WHERE p.code = 'GITHUB' AND p.is_active = TRUE
		ON CONFLICT (organization_id, provider_id, external_account_id)
		DO UPDATE SET
			external_account_name = EXCLUDED.external_account_name,
			external_account_type = EXCLUDED.external_account_type,
			installation_id = EXCLUDED.installation_id,
			status = 'ACTIVE',
			connected_by = EXCLUDED.connected_by,
			connected_at = NOW(),
			disconnected_at = NULL,
			updated_at = NOW(),
			provider_metadata = '{}'::jsonb
		RETURNING id, organization_id, external_account_id, external_account_name,
			external_account_type, installation_id, status, connected_at
	`, organizationID, fmt.Sprint(installation.AccountID), installation.AccountName,
		installation.AccountType, fmt.Sprint(installation.ID), userID)
	if err != nil {
		return Integration{}, fmt.Errorf("save organization GitHub installation: %w", err)
	}
	result.ProviderCode = "GITHUB"
	_, err = tx.ExecContext(ctx, `
		UPDATE tbl_github_installation_state
		SET consumed_at = NOW(), integration_id = $2
		WHERE state_hash = $1 AND consumed_at IS NULL
	`, stateHash, result.ID)
	if err != nil {
		return Integration{}, fmt.Errorf("consume GitHub installation state: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Integration{}, fmt.Errorf("commit GitHub installation: %w", err)
	}
	return result, nil
}

const integrationSelect = `
	SELECT i.id, i.organization_id, p.code AS provider_code, i.external_account_id,
		i.external_account_name, COALESCE(i.external_account_type, '') AS external_account_type,
		COALESCE(i.installation_id, '') AS installation_id, i.status, i.connected_at
	FROM tbl_organization_integration i
	INNER JOIN tbl_integration_provider p ON p.id = i.provider_id
`

func (r *repository) ListGitHub(ctx context.Context, organizationID string) ([]Integration, error) {
	items := make([]Integration, 0)
	err := r.db.SelectContext(ctx, &items, integrationSelect+`
		WHERE i.organization_id = $1 AND p.code = 'GITHUB'
		ORDER BY i.connected_at DESC
	`, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list organization GitHub integrations: %w", err)
	}
	return items, nil
}

func (r *repository) FindExportContext(ctx context.Context, organizationID, projectID string) (ExportContext, error) {
	var result ExportContext
	err := r.db.GetContext(ctx, &result, `
		SELECT p.id AS project_id, p.organization_id, p.slug AS project_slug,
			i.id AS integration_id,
			i.installation_id, i.external_account_name AS account_name,
			COALESCE(i.external_account_type, '') AS account_type
		FROM tbl_project p
		INNER JOIN tbl_organization_integration i
			ON i.organization_id = p.organization_id AND i.status = 'ACTIVE'
		INNER JOIN tbl_integration_provider provider ON provider.id = i.provider_id
		WHERE p.id = $1 AND p.organization_id = $2 AND p.deleted_at IS NULL
			AND provider.code = 'GITHUB'
		ORDER BY i.connected_at DESC
		LIMIT 1
	`, projectID, organizationID)
	if errors.Is(err, sql.ErrNoRows) {
		var projectExists bool
		if lookupErr := r.db.GetContext(ctx, &projectExists,
			`SELECT EXISTS (SELECT 1 FROM tbl_project WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL)`,
			projectID, organizationID); lookupErr != nil {
			return ExportContext{}, fmt.Errorf("check project for GitHub export: %w", lookupErr)
		}
		if !projectExists {
			return ExportContext{}, ErrProjectNotFound
		}
		return ExportContext{}, ErrIntegrationNotFound
	}
	if err != nil {
		return ExportContext{}, fmt.Errorf("find project GitHub export context: %w", err)
	}
	var existing ExportRepository
	if err := r.db.GetContext(ctx, &existing, `
		SELECT project_id, integration_id, external_repository_id, repository_name,
			full_name, COALESCE(default_branch, '') AS default_branch,
			COALESCE(repository_url, '') AS repository_url, is_private
		FROM tbl_project_repository
		WHERE project_id = $1 AND organization_id = $2
	`, projectID, organizationID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return ExportContext{}, fmt.Errorf("find existing project GitHub export: %w", err)
	} else if err == nil {
		result.Existing = &existing
	}
	return result, nil
}

func (r *repository) SaveExportRepository(ctx context.Context, projectID, integrationID string, repository ExportRepository, userID string) (ExportRepository, error) {
	var result ExportRepository
	err := r.db.GetContext(ctx, &result, `
		INSERT INTO tbl_project_repository (
			organization_id, project_id, integration_id, external_repository_id,
			repository_name, full_name, default_branch, repository_url,
			is_private, linked_by
		)
		SELECT p.organization_id, p.id, $2, $3, $4, $5, $6, $7, $8, $9
		FROM tbl_project p
		WHERE p.id = $1 AND p.deleted_at IS NULL
		RETURNING project_id, integration_id, external_repository_id, repository_name,
			full_name, COALESCE(default_branch, '') AS default_branch,
			COALESCE(repository_url, '') AS repository_url, is_private
	`, projectID, integrationID, repository.RepositoryID, repository.RepositoryName,
		repository.FullName, repository.DefaultBranch, repository.RepositoryURL, repository.Private, userID)
	if err != nil {
		return ExportRepository{}, fmt.Errorf("save project GitHub export repository: %w", err)
	}
	return result, nil
}

func (r *repository) FindExportRepository(ctx context.Context, organizationID, projectID string) (ExportRepository, error) {
	var result ExportRepository
	err := r.db.GetContext(ctx, &result, `
		SELECT project_id, integration_id, external_repository_id, repository_name,
			full_name, COALESCE(default_branch, '') AS default_branch,
			COALESCE(repository_url, '') AS repository_url, is_private
		FROM tbl_project_repository
		WHERE organization_id = $1 AND project_id = $2
	`, organizationID, projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return ExportRepository{}, ErrExportRepositoryNotFound
	}
	if err != nil {
		return ExportRepository{}, fmt.Errorf("find project GitHub export repository: %w", err)
	}
	return result, nil
}

func (r *repository) FindExportCredentials(ctx context.Context, organizationID, projectID string) (string, string, error) {
	var result struct {
		InstallationID string `db:"installation_id"`
		RepositoryURL  string `db:"repository_url"`
	}
	err := r.db.GetContext(ctx, &result, `
		SELECT i.installation_id, pr.repository_url
		FROM tbl_project_repository pr
		INNER JOIN tbl_organization_integration i
			ON i.id = pr.integration_id AND i.organization_id = pr.organization_id
		INNER JOIN tbl_integration_provider provider ON provider.id = i.provider_id
		WHERE pr.organization_id = $1 AND pr.project_id = $2
			AND i.status = 'ACTIVE' AND provider.code = 'GITHUB'
	`, organizationID, projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", ErrExportRepositoryNotFound
	}
	if err != nil {
		return "", "", fmt.Errorf("find project GitHub export credentials: %w", err)
	}
	return result.InstallationID, result.RepositoryURL, nil
}

func (r *repository) DisconnectGitHub(ctx context.Context, organizationID, integrationID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE tbl_organization_integration i
		SET status = 'REVOKED', disconnected_at = COALESCE(disconnected_at, NOW()), updated_at = NOW()
		FROM tbl_integration_provider p
		WHERE i.provider_id = p.id AND p.code = 'GITHUB'
		  AND i.organization_id = $1 AND i.id = $2
	`, organizationID, integrationID)
	if err != nil {
		return fmt.Errorf("disconnect organization GitHub integration: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check GitHub disconnection result: %w", err)
	}
	if rows == 0 {
		return ErrIntegrationNotFound
	}
	return nil
}

func (r *repository) SetInstallationStatus(ctx context.Context, installationID, status string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE tbl_organization_integration
		SET status = $1,
			disconnected_at = CASE WHEN $1 = 'ACTIVE' THEN NULL ELSE NOW() END,
			updated_at = NOW()
		WHERE installation_id = $2 AND status <> 'REVOKED'
	`, status, installationID)
	if err != nil {
		return fmt.Errorf("synchronize GitHub installation status: %w", err)
	}
	return nil
}
