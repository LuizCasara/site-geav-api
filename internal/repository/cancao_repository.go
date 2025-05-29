package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/site-geav-api/internal/models"
)

// PostgresCancaoRepository is an implementation of CancaoRepository using PostgreSQL
type PostgresCancaoRepository struct {
	db *sql.DB
}

// NewPostgresCancaoRepository creates a new PostgresCancaoRepository
func NewPostgresCancaoRepository(db *sql.DB) *PostgresCancaoRepository {
	return &PostgresCancaoRepository{db: db}
}

// GetByID retrieves a song by ID
func (r *PostgresCancaoRepository) GetByID(ctx context.Context, id int) (*models.Cancao, error) {
	query := `
		SELECT id, nome, link_youtube, letra, user_id, created_at, updated_at
		FROM cancoes
		WHERE id = $1
	`

	var cancao models.Cancao
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cancao.ID,
		&cancao.Nome,
		&cancao.LinkYoutube,
		&cancao.Letra,
		&cancao.UserID,
		&cancao.CreatedAt,
		&cancao.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil without error to indicate not found
		}
		return nil, fmt.Errorf("error getting cancao by ID: %w", err)
	}

	// Get tags
	tags, err := r.GetTags(ctx, cancao.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting tags for cancao: %w", err)
	}
	cancao.Tags = tags

	// Get ramos
	ramos, err := r.GetRamos(ctx, cancao.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting ramos for cancao: %w", err)
	}
	cancao.Ramos = ramos

	return &cancao, nil
}

// List retrieves all songs
func (r *PostgresCancaoRepository) List(ctx context.Context) ([]*models.Cancao, error) {
	query := `
		SELECT id, nome, link_youtube, letra, user_id, created_at, updated_at
		FROM cancoes
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error listing cancoes: %w", err)
	}
	defer rows.Close()

	var cancoes []*models.Cancao
	for rows.Next() {
		var cancao models.Cancao
		if err := rows.Scan(
			&cancao.ID,
			&cancao.Nome,
			&cancao.LinkYoutube,
			&cancao.Letra,
			&cancao.UserID,
			&cancao.CreatedAt,
			&cancao.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning cancao row: %w", err)
		}
		cancoes = append(cancoes, &cancao)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cancao rows: %w", err)
	}

	// Get related entities for each cancao
	for _, cancao := range cancoes {
		// Get tags
		tags, err := r.GetTags(ctx, cancao.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting tags for cancao: %w", err)
		}
		cancao.Tags = tags

		// Get ramos
		ramos, err := r.GetRamos(ctx, cancao.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting ramos for cancao: %w", err)
		}
		cancao.Ramos = ramos
	}

	return cancoes, nil
}

// Create creates a new song
func (r *PostgresCancaoRepository) Create(ctx context.Context, cancao *models.Cancao) (int, error) {
	query := `
		INSERT INTO cancoes (nome, link_youtube, letra, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		cancao.Nome,
		cancao.LinkYoutube,
		cancao.Letra,
		cancao.UserID,
		cancao.CreatedAt,
		cancao.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error creating cancao: %w", err)
	}

	return id, nil
}

// Update updates an existing song
func (r *PostgresCancaoRepository) Update(ctx context.Context, cancao *models.Cancao) error {
	query := `
		UPDATE cancoes
		SET nome = $1, link_youtube = $2, letra = $3, user_id = $4, updated_at = $5
		WHERE id = $6
	`

	cancao.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		cancao.Nome,
		cancao.LinkYoutube,
		cancao.Letra,
		cancao.UserID,
		cancao.UpdatedAt,
		cancao.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating cancao: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cancao with ID %d not found", cancao.ID)
	}

	return nil
}

// Delete deletes a song by ID
func (r *PostgresCancaoRepository) Delete(ctx context.Context, id int) error {
	query := `
		DELETE FROM cancoes
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting cancao: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cancao with ID %d not found", id)
	}

	return nil
}

// AddTag adds a tag to a song
func (r *PostgresCancaoRepository) AddTag(ctx context.Context, cancaoID, tagID int) error {
	query := `
		INSERT INTO cancoes_tags (cancao_id, tag_id)
		VALUES ($1, $2)
		ON CONFLICT (cancao_id, tag_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, cancaoID, tagID)
	if err != nil {
		return fmt.Errorf("error adding tag to cancao: %w", err)
	}

	return nil
}

// RemoveTag removes a tag from a song
func (r *PostgresCancaoRepository) RemoveTag(ctx context.Context, cancaoID, tagID int) error {
	query := `
		DELETE FROM cancoes_tags
		WHERE cancao_id = $1 AND tag_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, cancaoID, tagID)
	if err != nil {
		return fmt.Errorf("error removing tag from cancao: %w", err)
	}

	return nil
}

// GetTags gets all tags for a song
func (r *PostgresCancaoRepository) GetTags(ctx context.Context, cancaoID int) ([]*models.TagCancao, error) {
	query := `
		SELECT t.id, t.name, t.created_at
		FROM tags_cancoes t
		JOIN cancoes_tags ct ON t.id = ct.tag_id
		WHERE ct.cancao_id = $1
		ORDER BY t.name
	`

	rows, err := r.db.QueryContext(ctx, query, cancaoID)
	if err != nil {
		return nil, fmt.Errorf("error getting tags for cancao: %w", err)
	}
	defer rows.Close()

	var tags []*models.TagCancao
	for rows.Next() {
		tag := &models.TagCancao{}
		if err := rows.Scan(
			&tag.ID,
			&tag.Name,
			&tag.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning tag row: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tag rows: %w", err)
	}

	return tags, nil
}

// AddRamo adds a ramo to a song
func (r *PostgresCancaoRepository) AddRamo(ctx context.Context, cancaoID, ramoID int) error {
	query := `
		INSERT INTO cancoes_ramos (cancao_id, ramo_id)
		VALUES ($1, $2)
		ON CONFLICT (cancao_id, ramo_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, cancaoID, ramoID)
	if err != nil {
		return fmt.Errorf("error adding ramo to cancao: %w", err)
	}

	return nil
}

// RemoveRamo removes a ramo from a song
func (r *PostgresCancaoRepository) RemoveRamo(ctx context.Context, cancaoID, ramoID int) error {
	query := `
		DELETE FROM cancoes_ramos
		WHERE cancao_id = $1 AND ramo_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, cancaoID, ramoID)
	if err != nil {
		return fmt.Errorf("error removing ramo from cancao: %w", err)
	}

	return nil
}

// GetRamos gets all ramos for a song
func (r *PostgresCancaoRepository) GetRamos(ctx context.Context, cancaoID int) ([]*models.Ramo, error) {
	query := `
		SELECT r.id, r.name, r.created_at
		FROM ramos r
		JOIN cancoes_ramos cr ON r.id = cr.ramo_id
		WHERE cr.cancao_id = $1
		ORDER BY r.name
	`

	rows, err := r.db.QueryContext(ctx, query, cancaoID)
	if err != nil {
		return nil, fmt.Errorf("error getting ramos for cancao: %w", err)
	}
	defer rows.Close()

	var ramos []*models.Ramo
	for rows.Next() {
		ramo := &models.Ramo{}
		if err := rows.Scan(
			&ramo.ID,
			&ramo.Name,
			&ramo.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning ramo row: %w", err)
		}
		ramos = append(ramos, ramo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ramo rows: %w", err)
	}

	return ramos, nil
}
