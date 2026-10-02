package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"ai-agent/internal/modules/terminal/application"
	"ai-agent/internal/modules/terminal/domain"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (repository *Repository) CanAccessWorkspace(ctx context.Context, organizationID, userID string) (bool, error) {
	var allowed bool
	err := repository.db.GetContext(ctx, &allowed, `
		SELECT EXISTS (
			SELECT 1 FROM tbl_organization o 
			WHERE o.id = $1 AND o.created_by = $2
			UNION ALL
			SELECT 1 FROM tbl_organization_member om
			WHERE om.organization_id = $1 AND om.user_id = $2 AND om.deleted_at IS NULL
		)`, organizationID, userID)
	return allowed, err
}

func (repository *Repository) FindProjectWorkspace(ctx context.Context, projectID string) (string, error) {
	var organizationID string
	err := repository.db.GetContext(ctx, &organizationID, `SELECT organization_id FROM tbl_project WHERE id = $1`, projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.ErrSandboxNotFound
	}
	if err != nil {
		return "", fmt.Errorf("find project workspace: %w", err)
	}
	return organizationID, nil
}

func (repository *Repository) FindActiveByProject(ctx context.Context, projectID string) (domain.Sandbox, error) {
	return repository.find(ctx, `
		WHERE s.project_id = $1 AND s.status <> 'DESTROYED'
		ORDER BY s.created_at DESC LIMIT 1`, projectID)
}

func (repository *Repository) FindByID(ctx context.Context, sandboxID string) (domain.Sandbox, error) {
	return repository.find(ctx, `WHERE s.id = $1`, sandboxID)
}

func (repository *Repository) Create(ctx context.Context, sandbox domain.Sandbox) (string, error) {
	err := repository.db.GetContext(ctx, &sandbox.ID, `
		INSERT INTO tbl_sandbox (project_id, status, container_name, volume_name, image, workspace_path)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		sandbox.ProjectID, string(domain.StatusCreating),
		sandbox.ContainerName, sandbox.VolumeName, sandbox.Image, sandbox.WorkspacePath)
	if err != nil {
		return "", fmt.Errorf("create sandbox: %w", err)
	}

	// Create resource limits if provided
	if sandbox.Limits.CPUShares > 0 || sandbox.Limits.MemoryBytes > 0 {
		_, err = repository.db.ExecContext(ctx, `
			INSERT INTO tbl_sandbox_resource_limit (sandbox_id, cpu_quota, memory_limit_mb, pids_limit)
			VALUES ($1, $2, $3, $4)`,
			sandbox.ID,
			sandbox.Limits.CPUShares,
			sandbox.Limits.MemoryBytes/(1024*1024), // Convert bytes to MB
			sandbox.Limits.PidsLimit)
		if err != nil {
			return "", fmt.Errorf("create sandbox resource limits: %w", err)
		}
	}

	return sandbox.ID, nil
}

func (repository *Repository) AttachContainer(ctx context.Context, sandboxID, containerID string) error {
	result, err := repository.db.ExecContext(ctx, `
		UPDATE tbl_sandbox SET container_id = $2, updated_at = NOW()
		WHERE id = $1`, sandboxID, containerID)
	if err != nil {
		return fmt.Errorf("attach sandbox container: %w", err)
	}
	return requireOneRow(result, "sandbox")
}

func (repository *Repository) TransitionStatus(ctx context.Context, sandboxID string, from, to domain.SandboxStatus) error {
	result, err := repository.db.ExecContext(ctx, `
		UPDATE tbl_sandbox 
		SET status = $3, updated_at = NOW()
		WHERE id = $1 AND status = $2`,
		sandboxID, string(from), string(to))
	if err != nil {
		return fmt.Errorf("transition sandbox status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check sandbox status transition: %w", err)
	}
	if rows != 1 {
		return domain.ErrSandboxStateChanged
	}
	return nil
}

func (repository *Repository) SetLastError(ctx context.Context, sandboxID, message string) error {
	result, err := repository.db.ExecContext(ctx, `
		UPDATE tbl_sandbox SET last_error = $2, updated_at = NOW() WHERE id = $1`, sandboxID, message)
	if err != nil {
		return fmt.Errorf("set sandbox last error: %w", err)
	}
	return requireOneRow(result, "sandbox")
}

func (repository *Repository) find(ctx context.Context, suffix string, args ...any) (domain.Sandbox, error) {
	var row struct {
		ID             string         `db:"id"`
		ProjectID      string         `db:"project_id"`
		OrganizationID string         `db:"organization_id"`
		Status         string         `db:"status"`
		ContainerID    sql.NullString `db:"container_id"`
		ContainerName  sql.NullString `db:"container_name"`
		VolumeName     sql.NullString `db:"volume_name"`
		Image          sql.NullString `db:"image"`
		WorkspacePath  sql.NullString `db:"workspace_path"`
		LastError      sql.NullString `db:"last_error"`
	}
	err := repository.db.GetContext(ctx, &row, `
		SELECT s.id, s.project_id, p.organization_id, s.status,
		       s.container_id, s.container_name, s.volume_name, s.image, s.workspace_path,
		       s.last_error
		FROM tbl_sandbox s
		JOIN tbl_project p ON p.id = s.project_id `+suffix, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Sandbox{}, domain.ErrSandboxNotFound
	}
	if err != nil {
		return domain.Sandbox{}, fmt.Errorf("find sandbox: %w", err)
	}

	// Load resource limits if they exist
	var limits domain.ResourceLimits
	var limitRow struct {
		CPUQuota      sql.NullInt64 `db:"cpu_quota"`
		MemoryLimitMB sql.NullInt64 `db:"memory_limit_mb"`
		PidsLimit     sql.NullInt64 `db:"pids_limit"`
	}
	err = repository.db.GetContext(ctx, &limitRow, `
		SELECT cpu_quota, memory_limit_mb, pids_limit
		FROM tbl_sandbox_resource_limit
		WHERE sandbox_id = $1`, row.ID)
	if err == nil {
		if limitRow.CPUQuota.Valid {
			limits.CPUShares = limitRow.CPUQuota.Int64
		}
		if limitRow.MemoryLimitMB.Valid {
			limits.MemoryBytes = limitRow.MemoryLimitMB.Int64 * 1024 * 1024 // MB to bytes
		}
		if limitRow.PidsLimit.Valid {
			limits.PidsLimit = limitRow.PidsLimit.Int64
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return domain.Sandbox{}, fmt.Errorf("find sandbox limits: %w", err)
	}

	return domain.Sandbox{
		ID:             row.ID,
		ProjectID:      row.ProjectID,
		OrganizationID: row.OrganizationID,
		ContainerID:    row.ContainerID.String,
		ContainerName:  row.ContainerName.String,
		VolumeName:     row.VolumeName.String,
		Image:          row.Image.String,
		ImageActual:    row.Image.String, // Use same image for now
		WorkspacePath:  row.WorkspacePath.String,
		Status:         domain.SandboxStatus(row.Status),
		LastError:      row.LastError.String,
		Limits:         limits,
	}, nil
}

func requireOneRow(result sql.Result, resource string) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check %s update: %w", resource, err)
	}
	if rows != 1 {
		return fmt.Errorf("%s was not found: %w", resource, domain.ErrSandboxNotFound)
	}
	return nil
}

var _ application.Store = (*Repository)(nil)
