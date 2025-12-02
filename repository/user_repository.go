// Package repository provides data access layer for database operations.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"go-realtime-workspace/models"
)

// UserRepository handles user database operations.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user.
func (r *UserRepository) Create(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
	query := `
		INSERT INTO users (username, email, full_name, org_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, username, email, full_name, org_id, created_at, updated_at
	`

	user := &models.User{}
	err := r.db.QueryRowContext(
		ctx, query,
		req.Username, req.Email, req.FullName, req.OrgID,
	).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName,
		&user.OrgID, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return user, nil
}

// GetByID retrieves a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, username, email, full_name, org_id, created_at, updated_at
		FROM users WHERE id = $1
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName,
		&user.OrgID, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return user, nil
}

// GetByUsername retrieves a user by username.
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, email, full_name, org_id, created_at, updated_at
		FROM users WHERE username = $1
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName,
		&user.OrgID, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return user, nil
}

// GetByOrgID retrieves all users in an organization.
func (r *UserRepository) GetByOrgID(ctx context.Context, orgID string) ([]models.User, error) {
	query := `
		SELECT id, username, email, full_name, org_id, created_at, updated_at
		FROM users WHERE org_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("error getting users: %w", err)
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.FullName,
			&user.OrgID, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// Update updates a user.
func (r *UserRepository) Update(ctx context.Context, id string, req models.UpdateUserRequest) (*models.User, error) {
	query := `
		UPDATE users
		SET username = COALESCE(NULLIF($1, ''), username),
		    email = COALESCE(NULLIF($2, ''), email),
		    full_name = COALESCE(NULLIF($3, ''), full_name),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING id, username, email, full_name, org_id, created_at, updated_at
	`

	user := &models.User{}
	err := r.db.QueryRowContext(
		ctx, query,
		req.Username, req.Email, req.FullName, id,
	).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName,
		&user.OrgID, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return user, nil
}

// Delete deletes a user.
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
