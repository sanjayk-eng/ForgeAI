package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"ai-agent/internal/modules/terminal/domain"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (repository *Repository) CanAccessWorkspace(ctx context.Context, workspaceID, userID string) (bool, error) {
	var allowed bool
	err := repository.db.GetContext(ctx, &allowed, `
		SELECT EXISTS (
			SELECT 1 FROM tbl_workspace w WHERE w.id = $1 AND w.owner_id = $2
			UNION ALL
			SELECT 1 FROM tbl_workspace_member wm
			WHERE wm.workspace_id = $1 AND wm.user_id = $2 AND wm.deleted_at IS NULL
		)`, workspaceID, userID)
	return allowed, err
}

func (repository *Repository) FindProjectWorkspace(ctx context.Context, projectID string) (string, error) {
	var workspaceID string
	err := repository.db.GetContext(ctx, &workspaceID, `SELECT workspace_id FROM tbl_project WHERE id = $1`, projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.ErrSandboxNotFound
	}
	if err != nil {
		return "", fmt.Errorf("find project workspace: %w", err)
	}
	return workspaceID, nil
}

func (repository *Repository) FindActiveByProject(ctx context.Context, projectID string) (domain.Sandbox, error) {
	return repository.find(ctx, `
		WHERE s.project_id = $1 AND status_enum.code <> 'DESTROYED'
		ORDER BY s.created_at DESC LIMIT 1`, projectID)
}

func (repository *Repository) FindByID(ctx context.Context, sandboxID string) (domain.Sandbox, error) {
	return repository.find(ctx, `WHERE s.id = $1`, sandboxID)
}

func (repository *Repository) Create(ctx context.Context, sandbox domain.Sandbox) (string, error) {
	tx, err := repository.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin sandbox transaction: %w", err)
	}
	defer tx.Rollback()

	if err := tx.GetContext(ctx, &sandbox.ID, `
		INSERT INTO tbl_sandbox (workspace_id, project_id, status_id, volume_config, resource_limit, last_error)
		SELECT $1, $2, e.id, '{}'::jsonb, '{}'::jsonb, NULL FROM tbl_enum e
		WHERE e.category = 'SANDBOX_STATUS' AND e.code = $3
		RETURNING id`, sandbox.WorkspaceID, sandbox.ProjectID, string(domain.StatusCreating)); err != nil {
		return "", fmt.Errorf("create sandbox state: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO tbl_sandbox_container (sandbox_id, source_type, config)
		VALUES ($1, 'docker', jsonb_build_object('name', $2, 'image', $3, 'image_used', $4))`, sandbox.ID, sandbox.ContainerName, sandbox.Image, sandbox.ImageActual); err != nil {
		return "", fmt.Errorf("create sandbox container state: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE tbl_sandbox
		SET volume_config = jsonb_build_object('name', $2, 'mount_path', $3),
		    resource_limit = jsonb_build_object('cpu_shares', $4, 'memory_bytes', $5, 'pids_limit', $6)
		WHERE id = $1`, sandbox.ID, sandbox.VolumeName, sandbox.WorkspacePath, sandbox.Limits.CPUShares, sandbox.Limits.MemoryBytes, sandbox.Limits.PidsLimit); err != nil {
		return "", fmt.Errorf("create sandbox configuration: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit sandbox state: %w", err)
	}
	return sandbox.ID, nil
}

func (repository *Repository) AttachContainer(ctx context.Context, sandboxID, containerID string) error {
	result, err := repository.db.ExecContext(ctx, `
		UPDATE tbl_sandbox_container SET source_id = $2, updated_at = NOW()
		WHERE sandbox_id = $1`, sandboxID, containerID)
	if err != nil {
		return fmt.Errorf("attach sandbox container: %w", err)
	}
	return requireOneRow(result, "sandbox container")
}

func (repository *Repository) TransitionStatus(ctx context.Context, sandboxID string, from, to domain.SandboxStatus) error {
	result, err := repository.db.ExecContext(ctx, `
		UPDATE tbl_sandbox SET status_id = (
			SELECT next_status.id FROM tbl_enum next_status
			WHERE next_status.category = 'SANDBOX_STATUS' AND next_status.code = $3
		), updated_at = NOW()
		WHERE id = $1 AND status_id = (
			SELECT current_status.id FROM tbl_enum current_status
			WHERE current_status.category = 'SANDBOX_STATUS' AND current_status.code = $2
		)`, sandboxID, string(from), string(to))
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
		ID              string `db:"id"`
		WorkspaceID     string `db:"workspace_id"`
		ProjectID       string `db:"project_id"`
		Status          string `db:"status"`
		ContainerID     string `db:"container_id"`
		ContainerConfig []byte `db:"container_config"`
		VolumeConfig    []byte `db:"volume_config"`
		ResourceLimit   []byte `db:"resource_limit"`
		LastError       string `db:"last_error"`
	}
	err := repository.db.GetContext(ctx, &row, `
		SELECT s.id, s.workspace_id, s.project_id, status_enum.code AS status,
		       c.source_id AS container_id, c.config AS container_config,
		       s.volume_config, s.resource_limit, s.last_error
		FROM tbl_sandbox s
		JOIN tbl_enum status_enum ON status_enum.id = s.status_id
		JOIN tbl_sandbox_container c ON c.sandbox_id = s.id `+suffix, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Sandbox{}, domain.ErrSandboxNotFound
	}
	if err != nil {
		return domain.Sandbox{}, fmt.Errorf("find sandbox: %w", err)
	}
	var containerConfig struct {
		Name        string `json:"name"`
		Image       string `json:"image"`
		ImageActual string `json:"image_used"`
	}
	var volumeConfig struct {
		Name      string `json:"name"`
		MountPath string `json:"mount_path"`
	}
	var limits domain.ResourceLimits
	if err := json.Unmarshal(row.ContainerConfig, &containerConfig); err != nil {
		return domain.Sandbox{}, fmt.Errorf("decode sandbox container config: %w", err)
	}
	if err := json.Unmarshal(row.VolumeConfig, &volumeConfig); err != nil {
		return domain.Sandbox{}, fmt.Errorf("decode sandbox volume config: %w", err)
	}
	if err := json.Unmarshal(row.ResourceLimit, &limits); err != nil {
		return domain.Sandbox{}, fmt.Errorf("decode sandbox resource limit: %w", err)
	}
	return domain.Sandbox{
		ID: row.ID, WorkspaceID: row.WorkspaceID, ProjectID: row.ProjectID,
		ContainerID: row.ContainerID, ContainerName: containerConfig.Name,
		VolumeName: volumeConfig.Name, Image: containerConfig.Image, ImageActual: containerConfig.ImageActual,
		WorkspacePath: volumeConfig.MountPath, Status: domain.SandboxStatus(row.Status),
		LastError: row.LastError, Limits: limits,
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
