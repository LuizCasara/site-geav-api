package repository

import (
	"context"

	"github.com/site-geav-api/internal/models"
)

// UserRepository defines the interface for user operations
type UserRepository interface {
	GetByID(ctx context.Context, id int) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	List(ctx context.Context) ([]*models.User, error)
	Create(ctx context.Context, user *models.User) (int, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id int) error
}

// LugarRepository defines the interface for lugar operations
type LugarRepository interface {
	GetByID(ctx context.Context, id int) (*models.Lugar, error)
	List(ctx context.Context) ([]*models.Lugar, error)
	Create(ctx context.Context, lugar *models.Lugar) (int, error)
	Update(ctx context.Context, lugar *models.Lugar) error
	Delete(ctx context.Context, id int) error
	
	// Related operations
	AddImage(ctx context.Context, image *models.LugarImage) (int, error)
	DeleteImage(ctx context.Context, imageID int) error
	GetImages(ctx context.Context, lugarID int) ([]*models.LugarImage, error)
	
	AddTag(ctx context.Context, lugarID, tagID int) error
	RemoveTag(ctx context.Context, lugarID, tagID int) error
	GetTags(ctx context.Context, lugarID int) ([]*models.TagLugar, error)
	
	AddRamo(ctx context.Context, lugarID, ramoID int) error
	RemoveRamo(ctx context.Context, lugarID, ramoID int) error
	GetRamos(ctx context.Context, lugarID int) ([]*models.Ramo, error)
	
	AddRating(ctx context.Context, rating *models.LugarRating) (int, error)
	UpdateRating(ctx context.Context, rating *models.LugarRating) error
	DeleteRating(ctx context.Context, ratingID int) error
	GetRatings(ctx context.Context, lugarID int) ([]*models.LugarRating, error)
}

// CancaoRepository defines the interface for cancao operations
type CancaoRepository interface {
	GetByID(ctx context.Context, id int) (*models.Cancao, error)
	List(ctx context.Context) ([]*models.Cancao, error)
	Create(ctx context.Context, cancao *models.Cancao) (int, error)
	Update(ctx context.Context, cancao *models.Cancao) error
	Delete(ctx context.Context, id int) error
	
	// Related operations
	AddTag(ctx context.Context, cancaoID, tagID int) error
	RemoveTag(ctx context.Context, cancaoID, tagID int) error
	GetTags(ctx context.Context, cancaoID int) ([]*models.TagCancao, error)
	
	AddRamo(ctx context.Context, cancaoID, ramoID int) error
	RemoveRamo(ctx context.Context, cancaoID, ramoID int) error
	GetRamos(ctx context.Context, cancaoID int) ([]*models.Ramo, error)
}

// TagLugarRepository defines the interface for tag_lugar operations
type TagLugarRepository interface {
	GetByID(ctx context.Context, id int) (*models.TagLugar, error)
	List(ctx context.Context) ([]*models.TagLugar, error)
	Create(ctx context.Context, tag *models.TagLugar) (int, error)
	Update(ctx context.Context, tag *models.TagLugar) error
	Delete(ctx context.Context, id int) error
}

// TagCancaoRepository defines the interface for tag_cancao operations
type TagCancaoRepository interface {
	GetByID(ctx context.Context, id int) (*models.TagCancao, error)
	List(ctx context.Context) ([]*models.TagCancao, error)
	Create(ctx context.Context, tag *models.TagCancao) (int, error)
	Update(ctx context.Context, tag *models.TagCancao) error
	Delete(ctx context.Context, id int) error
}

// RamoRepository defines the interface for ramo operations
type RamoRepository interface {
	GetByID(ctx context.Context, id int) (*models.Ramo, error)
	List(ctx context.Context) ([]*models.Ramo, error)
	Create(ctx context.Context, ramo *models.Ramo) (int, error)
	Update(ctx context.Context, ramo *models.Ramo) error
	Delete(ctx context.Context, id int) error
}