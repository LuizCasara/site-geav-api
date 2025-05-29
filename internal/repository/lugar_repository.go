package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/site-geav-api/internal/models"
)

// PostgresLugarRepository is an implementation of LugarRepository using PostgreSQL
type PostgresLugarRepository struct {
	db *sql.DB
}

// NewPostgresLugarRepository creates a new PostgresLugarRepository
func NewPostgresLugarRepository(db *sql.DB) *PostgresLugarRepository {
	return &PostgresLugarRepository{db: db}
}

// GetByID retrieves a place by ID
func (r *PostgresLugarRepository) GetByID(ctx context.Context, id int) (*models.Lugar, error) {
	query := `
		SELECT l.id, l.nome_local, l.nome_dono_local, l.telefone_para_contato, 
		       l.link_google_maps, l.link_site, l.endereco_completo, 
		       l.local_publico, l.valor_fixo, l.valor_individual, 
		       l.user_id, l.created_at, l.updated_at,
		       COALESCE(lwr.average_rating, 0) as average_rating,
		       COALESCE(lwr.rating_count, 0) as rating_count
		FROM lugares l
		LEFT JOIN lugares_with_ratings lwr ON l.id = lwr.id
		WHERE l.id = $1
	`

	var lugar models.Lugar
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&lugar.ID,
		&lugar.NomeLocal,
		&lugar.NomeDonoLocal,
		&lugar.TelefoneParaContato,
		&lugar.LinkGoogleMaps,
		&lugar.LinkSite,
		&lugar.EnderecoCompleto,
		&lugar.LocalPublico,
		&lugar.ValorFixo,
		&lugar.ValorIndividual,
		&lugar.UserID,
		&lugar.CreatedAt,
		&lugar.UpdatedAt,
		&lugar.AverageRating,
		&lugar.RatingCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil without error to indicate not found
		}
		return nil, fmt.Errorf("error getting lugar by ID: %w", err)
	}

	// Get images
	images, err := r.GetImages(ctx, lugar.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting images for lugar: %w", err)
	}
	lugar.Images = images

	// Get tags
	tags, err := r.GetTags(ctx, lugar.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting tags for lugar: %w", err)
	}
	lugar.Tags = tags

	// Get ramos
	ramos, err := r.GetRamos(ctx, lugar.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting ramos for lugar: %w", err)
	}
	lugar.Ramos = ramos

	return &lugar, nil
}

// List retrieves all places
func (r *PostgresLugarRepository) List(ctx context.Context) ([]*models.Lugar, error) {
	query := `
		SELECT l.id, l.nome_local, l.nome_dono_local, l.telefone_para_contato, 
		       l.link_google_maps, l.link_site, l.endereco_completo, 
		       l.local_publico, l.valor_fixo, l.valor_individual, 
		       l.user_id, l.created_at, l.updated_at,
		       COALESCE(lwr.average_rating, 0) as average_rating,
		       COALESCE(lwr.rating_count, 0) as rating_count
		FROM lugares l
		LEFT JOIN lugares_with_ratings lwr ON l.id = lwr.id
		ORDER BY l.id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error listing lugares: %w", err)
	}
	defer rows.Close()

	var lugares []*models.Lugar
	for rows.Next() {
		var lugar models.Lugar
		if err := rows.Scan(
			&lugar.ID,
			&lugar.NomeLocal,
			&lugar.NomeDonoLocal,
			&lugar.TelefoneParaContato,
			&lugar.LinkGoogleMaps,
			&lugar.LinkSite,
			&lugar.EnderecoCompleto,
			&lugar.LocalPublico,
			&lugar.ValorFixo,
			&lugar.ValorIndividual,
			&lugar.UserID,
			&lugar.CreatedAt,
			&lugar.UpdatedAt,
			&lugar.AverageRating,
			&lugar.RatingCount,
		); err != nil {
			return nil, fmt.Errorf("error scanning lugar row: %w", err)
		}
		lugares = append(lugares, &lugar)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lugar rows: %w", err)
	}

	// Get related entities for each lugar
	for _, lugar := range lugares {
		// Get images
		images, err := r.GetImages(ctx, lugar.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting images for lugar: %w", err)
		}
		lugar.Images = images

		// Get tags
		tags, err := r.GetTags(ctx, lugar.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting tags for lugar: %w", err)
		}
		lugar.Tags = tags

		// Get ramos
		ramos, err := r.GetRamos(ctx, lugar.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting ramos for lugar: %w", err)
		}
		lugar.Ramos = ramos
	}

	return lugares, nil
}

// Create creates a new place
func (r *PostgresLugarRepository) Create(ctx context.Context, lugar *models.Lugar) (int, error) {
	query := `
		INSERT INTO lugares (
			nome_local, nome_dono_local, telefone_para_contato, 
			link_google_maps, link_site, endereco_completo, 
			local_publico, valor_fixo, valor_individual, 
			user_id, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		lugar.NomeLocal,
		lugar.NomeDonoLocal,
		lugar.TelefoneParaContato,
		lugar.LinkGoogleMaps,
		lugar.LinkSite,
		lugar.EnderecoCompleto,
		lugar.LocalPublico,
		lugar.ValorFixo,
		lugar.ValorIndividual,
		lugar.UserID,
		lugar.CreatedAt,
		lugar.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error creating lugar: %w", err)
	}

	return id, nil
}

// Update updates an existing place
func (r *PostgresLugarRepository) Update(ctx context.Context, lugar *models.Lugar) error {
	query := `
		UPDATE lugares
		SET nome_local = $1, nome_dono_local = $2, telefone_para_contato = $3, 
		    link_google_maps = $4, link_site = $5, endereco_completo = $6, 
		    local_publico = $7, valor_fixo = $8, valor_individual = $9, 
		    user_id = $10, updated_at = $11
		WHERE id = $12
	`

	lugar.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		lugar.NomeLocal,
		lugar.NomeDonoLocal,
		lugar.TelefoneParaContato,
		lugar.LinkGoogleMaps,
		lugar.LinkSite,
		lugar.EnderecoCompleto,
		lugar.LocalPublico,
		lugar.ValorFixo,
		lugar.ValorIndividual,
		lugar.UserID,
		lugar.UpdatedAt,
		lugar.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating lugar: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("lugar with ID %d not found", lugar.ID)
	}

	return nil
}

// Delete deletes a place by ID
func (r *PostgresLugarRepository) Delete(ctx context.Context, id int) error {
	query := `
		DELETE FROM lugares
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting lugar: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("lugar with ID %d not found", id)
	}

	return nil
}

// AddImage adds an image to a place
func (r *PostgresLugarRepository) AddImage(ctx context.Context, image *models.LugarImage) (int, error) {
	query := `
		INSERT INTO lugares_images (lugar_id, image_url, display_order, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		image.LugarID,
		image.ImageURL,
		image.DisplayOrder,
		image.CreatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error adding image to lugar: %w", err)
	}

	return id, nil
}

// DeleteImage deletes an image from a place
func (r *PostgresLugarRepository) DeleteImage(ctx context.Context, imageID int) error {
	query := `
		DELETE FROM lugares_images
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, imageID)
	if err != nil {
		return fmt.Errorf("error deleting image: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("image with ID %d not found", imageID)
	}

	return nil
}

// GetImages gets all images for a place
func (r *PostgresLugarRepository) GetImages(ctx context.Context, lugarID int) ([]*models.LugarImage, error) {
	query := `
		SELECT id, lugar_id, image_url, display_order, created_at
		FROM lugares_images
		WHERE lugar_id = $1
		ORDER BY display_order
	`

	rows, err := r.db.QueryContext(ctx, query, lugarID)
	if err != nil {
		return nil, fmt.Errorf("error getting images for lugar: %w", err)
	}
	defer rows.Close()

	var images []*models.LugarImage
	for rows.Next() {
		image := &models.LugarImage{}
		if err := rows.Scan(
			&image.ID,
			&image.LugarID,
			&image.ImageURL,
			&image.DisplayOrder,
			&image.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning image row: %w", err)
		}
		images = append(images, image)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating image rows: %w", err)
	}

	return images, nil
}

// AddTag adds a tag to a place
func (r *PostgresLugarRepository) AddTag(ctx context.Context, lugarID, tagID int) error {
	query := `
		INSERT INTO lugares_tags (lugar_id, tag_id)
		VALUES ($1, $2)
		ON CONFLICT (lugar_id, tag_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, lugarID, tagID)
	if err != nil {
		return fmt.Errorf("error adding tag to lugar: %w", err)
	}

	return nil
}

// RemoveTag removes a tag from a place
func (r *PostgresLugarRepository) RemoveTag(ctx context.Context, lugarID, tagID int) error {
	query := `
		DELETE FROM lugares_tags
		WHERE lugar_id = $1 AND tag_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, lugarID, tagID)
	if err != nil {
		return fmt.Errorf("error removing tag from lugar: %w", err)
	}

	return nil
}

// GetTags gets all tags for a place
func (r *PostgresLugarRepository) GetTags(ctx context.Context, lugarID int) ([]*models.TagLugar, error) {
	query := `
		SELECT t.id, t.name, t.created_at
		FROM tags_lugares t
		JOIN lugares_tags lt ON t.id = lt.tag_id
		WHERE lt.lugar_id = $1
		ORDER BY t.name
	`

	rows, err := r.db.QueryContext(ctx, query, lugarID)
	if err != nil {
		return nil, fmt.Errorf("error getting tags for lugar: %w", err)
	}
	defer rows.Close()

	var tags []*models.TagLugar
	for rows.Next() {
		tag := &models.TagLugar{}
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

// AddRamo adds a ramo to a place
func (r *PostgresLugarRepository) AddRamo(ctx context.Context, lugarID, ramoID int) error {
	query := `
		INSERT INTO lugares_ramos (lugar_id, ramo_id)
		VALUES ($1, $2)
		ON CONFLICT (lugar_id, ramo_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, lugarID, ramoID)
	if err != nil {
		return fmt.Errorf("error adding ramo to lugar: %w", err)
	}

	return nil
}

// RemoveRamo removes a ramo from a place
func (r *PostgresLugarRepository) RemoveRamo(ctx context.Context, lugarID, ramoID int) error {
	query := `
		DELETE FROM lugares_ramos
		WHERE lugar_id = $1 AND ramo_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, lugarID, ramoID)
	if err != nil {
		return fmt.Errorf("error removing ramo from lugar: %w", err)
	}

	return nil
}

// GetRamos gets all ramos for a place
func (r *PostgresLugarRepository) GetRamos(ctx context.Context, lugarID int) ([]*models.Ramo, error) {
	query := `
		SELECT r.id, r.name, r.created_at
		FROM ramos r
		JOIN lugares_ramos lr ON r.id = lr.ramo_id
		WHERE lr.lugar_id = $1
		ORDER BY r.name
	`

	rows, err := r.db.QueryContext(ctx, query, lugarID)
	if err != nil {
		return nil, fmt.Errorf("error getting ramos for lugar: %w", err)
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

// AddRating adds a rating to a place
func (r *PostgresLugarRepository) AddRating(ctx context.Context, rating *models.LugarRating) (int, error) {
	query := `
		INSERT INTO lugares_ratings (lugar_id, user_id, rating, date)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (lugar_id, user_id) DO UPDATE
		SET rating = EXCLUDED.rating, date = EXCLUDED.date
		RETURNING id
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		rating.LugarID,
		rating.UserID,
		rating.Rating,
		rating.Date,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error adding rating to lugar: %w", err)
	}

	return id, nil
}

// UpdateRating updates a rating for a place
func (r *PostgresLugarRepository) UpdateRating(ctx context.Context, rating *models.LugarRating) error {
	query := `
		UPDATE lugares_ratings
		SET rating = $1, date = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query,
		rating.Rating,
		rating.Date,
		rating.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating rating: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("rating with ID %d not found", rating.ID)
	}

	return nil
}

// DeleteRating deletes a rating for a place
func (r *PostgresLugarRepository) DeleteRating(ctx context.Context, ratingID int) error {
	query := `
		DELETE FROM lugares_ratings
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, ratingID)
	if err != nil {
		return fmt.Errorf("error deleting rating: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("rating with ID %d not found", ratingID)
	}

	return nil
}

// GetRatings gets all ratings for a place
func (r *PostgresLugarRepository) GetRatings(ctx context.Context, lugarID int) ([]*models.LugarRating, error) {
	query := `
		SELECT id, lugar_id, user_id, rating, date
		FROM lugares_ratings
		WHERE lugar_id = $1
		ORDER BY date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, lugarID)
	if err != nil {
		return nil, fmt.Errorf("error getting ratings for lugar: %w", err)
	}
	defer rows.Close()

	var ratings []*models.LugarRating
	for rows.Next() {
		rating := &models.LugarRating{}
		if err := rows.Scan(
			&rating.ID,
			&rating.LugarID,
			&rating.UserID,
			&rating.Rating,
			&rating.Date,
		); err != nil {
			return nil, fmt.Errorf("error scanning rating row: %w", err)
		}
		ratings = append(ratings, rating)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rating rows: %w", err)
	}

	return ratings, nil
}
