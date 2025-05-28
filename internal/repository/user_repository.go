package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/site-geav-api/internal/models"
)

// PostgresUserRepository is an implementation of UserRepository using PostgreSQL
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository creates a new PostgresUserRepository
func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// GetByID retrieves a user by ID
func (r *PostgresUserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	query := `
		SELECT id, username, password, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	
	var user models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with ID %d not found", id)
		}
		return nil, fmt.Errorf("error getting user by ID: %w", err)
	}
	
	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *PostgresUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, password, role, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	
	var user models.User
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with username %s not found", username)
		}
		return nil, fmt.Errorf("error getting user by username: %w", err)
	}
	
	return &user, nil
}

// List retrieves all users
func (r *PostgresUserRepository) List(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT id, username, password, role, created_at, updated_at
		FROM users
		ORDER BY id
	`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error listing users: %w", err)
	}
	defer rows.Close()
	
	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Password,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning user row: %w", err)
		}
		users = append(users, &user)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}
	
	return users, nil
}

// Create creates a new user
func (r *PostgresUserRepository) Create(ctx context.Context, user *models.User) (int, error) {
	query := `
		INSERT INTO users (username, password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	
	var id int
	err := r.db.QueryRowContext(ctx, query,
		user.Username,
		user.Password,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&id)
	
	if err != nil {
		return 0, fmt.Errorf("error creating user: %w", err)
	}
	
	return id, nil
}

// Update updates an existing user
func (r *PostgresUserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET username = $1, password = $2, role = $3, updated_at = $4
		WHERE id = $5
	`
	
	result, err := r.db.ExecContext(ctx, query,
		user.Username,
		user.Password,
		user.Role,
		user.UpdatedAt,
		user.ID,
	)
	
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user with ID %d not found", user.ID)
	}
	
	return nil
}

// Delete deletes a user by ID
func (r *PostgresUserRepository) Delete(ctx context.Context, id int) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user with ID %d not found", id)
	}
	
	return nil
}