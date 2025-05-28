package models

import (
	"time"
)

// Lugar represents a place in the system
type Lugar struct {
	ID                 int       `json:"id" db:"id"`
	NomeLocal          string    `json:"nome_local" db:"nome_local"`
	NomeDonoLocal      string    `json:"nome_dono_local" db:"nome_dono_local"`
	TelefoneParaContato int64     `json:"telefone_para_contato" db:"telefone_para_contato"`
	LinkGoogleMaps     string    `json:"link_google_maps" db:"link_google_maps"`
	LinkSite           string    `json:"link_site" db:"link_site"`
	EnderecoCompleto   string    `json:"endereco_completo" db:"endereco_completo"`
	LocalPublico       bool      `json:"local_publico" db:"local_publico"`
	ValorFixo          float64   `json:"valor_fixo" db:"valor_fixo"`
	ValorIndividual    float64   `json:"valor_individual" db:"valor_individual"`
	UserID             int       `json:"user_id" db:"user_id"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
	
	// Related entities (not stored in the database directly)
	Images []LugarImage `json:"images,omitempty" db:"-"`
	Tags   []TagLugar   `json:"tags,omitempty" db:"-"`
	Ramos  []Ramo       `json:"ramos,omitempty" db:"-"`
	
	// Calculated fields from the materialized view
	AverageRating float64 `json:"average_rating,omitempty" db:"average_rating"`
	RatingCount   int     `json:"rating_count,omitempty" db:"rating_count"`
}

// LugarImage represents an image associated with a place
type LugarImage struct {
	ID           int       `json:"id" db:"id"`
	LugarID      int       `json:"lugar_id" db:"lugar_id"`
	ImageURL     string    `json:"image_url" db:"image_url"`
	DisplayOrder int       `json:"display_order" db:"display_order"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// LugarRating represents a rating given to a place
type LugarRating struct {
	ID      int       `json:"id" db:"id"`
	LugarID int       `json:"lugar_id" db:"lugar_id"`
	UserID  int       `json:"user_id" db:"user_id"`
	Rating  int       `json:"rating" db:"rating"`
	Date    time.Time `json:"date" db:"date"`
}

// NewLugar creates a new place with default values
func NewLugar(
	nomeLocal, nomeDonoLocal string,
	telefoneParaContato int64,
	linkGoogleMaps, linkSite, enderecoCompleto string,
	localPublico bool,
	valorFixo, valorIndividual float64,
	userID int,
) *Lugar {
	now := time.Now()
	return &Lugar{
		NomeLocal:           nomeLocal,
		NomeDonoLocal:       nomeDonoLocal,
		TelefoneParaContato: telefoneParaContato,
		LinkGoogleMaps:      linkGoogleMaps,
		LinkSite:            linkSite,
		EnderecoCompleto:    enderecoCompleto,
		LocalPublico:        localPublico,
		ValorFixo:           valorFixo,
		ValorIndividual:     valorIndividual,
		UserID:              userID,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

// NewLugarImage creates a new image for a place
func NewLugarImage(lugarID int, imageURL string, displayOrder int) *LugarImage {
	return &LugarImage{
		LugarID:      lugarID,
		ImageURL:     imageURL,
		DisplayOrder: displayOrder,
		CreatedAt:    time.Now(),
	}
}

// NewLugarRating creates a new rating for a place
func NewLugarRating(lugarID, userID, rating int) *LugarRating {
	return &LugarRating{
		LugarID: lugarID,
		UserID:  userID,
		Rating:  rating,
		Date:    time.Now(),
	}
}