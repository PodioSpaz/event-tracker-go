package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/PodioSpaz/event-tracker-go/internal/repository"
	"github.com/PodioSpaz/event-tracker-go/internal/util"
)

// RoleRepository implements repository.RoleRepository for SQLite
type RoleRepository struct {
	db *DB
}

// NewRoleRepository creates a new RoleRepository
func NewRoleRepository(db *DB) repository.RoleRepository {
	return &RoleRepository{db: db}
}

// Create creates a new role
func (r *RoleRepository) Create(ctx context.Context, role *domain.Role) error {
	if err := role.Validate(); err != nil {
		return err
	}

	// Check for duplicate name
	exists, err := r.ExistsByName(ctx, role.Name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("role with name %s already exists", role.Name)
	}

	query := `
		INSERT INTO roles (name, display_name, description, active)
		VALUES (?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		role.Name,
		role.DisplayName,
		role.Description,
		role.Active,
	)

	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	role.ID = id

	return nil
}

// GetByID retrieves a role by ID
func (r *RoleRepository) GetByID(ctx context.Context, id int64) (*domain.Role, error) {
	query := `
		SELECT id, name, display_name, description, active
		FROM roles
		WHERE id = ?
	`

	role := &domain.Role{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&role.ID,
		&role.Name,
		&role.DisplayName,
		&role.Description,
		&role.Active,
	)

	if err == sql.ErrNoRows {
		return nil, util.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return role, nil
}

// GetByName retrieves a role by name
func (r *RoleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	query := `
		SELECT id, name, display_name, description, active
		FROM roles
		WHERE name = ?
	`

	role := &domain.Role{}

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&role.ID,
		&role.Name,
		&role.DisplayName,
		&role.Description,
		&role.Active,
	)

	if err == sql.ErrNoRows {
		return nil, util.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}

	return role, nil
}

// GetAll retrieves all roles
func (r *RoleRepository) GetAll(ctx context.Context) ([]*domain.Role, error) {
	query := `
		SELECT id, name, display_name, description, active
		FROM roles
		ORDER BY display_name
	`

	return r.queryRoles(ctx, query)
}

// GetActive retrieves all active roles
func (r *RoleRepository) GetActive(ctx context.Context) ([]*domain.Role, error) {
	query := `
		SELECT id, name, display_name, description, active
		FROM roles
		WHERE active = 1
		ORDER BY display_name
	`

	return r.queryRoles(ctx, query)
}

// Update updates an existing role
func (r *RoleRepository) Update(ctx context.Context, role *domain.Role) error {
	if err := role.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE roles
		SET display_name = ?, description = ?, active = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		role.DisplayName,
		role.Description,
		role.Active,
		role.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return util.ErrRecordNotFound
	}

	return nil
}

// Delete deletes a role by ID
func (r *RoleRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM roles WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return util.ErrRecordNotFound
	}

	return nil
}

// ExistsByName checks if a role with the given name exists
func (r *RoleRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	query := `SELECT COUNT(*) FROM roles WHERE name = ?`

	var count int
	err := r.db.QueryRowContext(ctx, query, name).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check role existence: %w", err)
	}

	return count > 0, nil
}

// queryRoles is a helper method to query roles with optional parameters
func (r *RoleRepository) queryRoles(ctx context.Context, query string, args ...interface{}) ([]*domain.Role, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	var roles []*domain.Role

	for rows.Next() {
		role := &domain.Role{}

		err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.DisplayName,
			&role.Description,
			&role.Active,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating roles: %w", err)
	}

	return roles, nil
}
