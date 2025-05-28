package models

import (
	"time"
)

// Cancao represents a song in the system
type Cancao struct {
	ID          int       `json:"id" db:"id"`
	Nome        string    `json:"nome" db:"nome"`
	LinkYoutube string    `json:"link_youtube" db:"link_youtube"`
	Letra       string    `json:"letra" db:"letra"`
	UserID      int       `json:"user_id" db:"user_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	
	// Related entities (not stored in the database directly)
	Tags  []TagCancao `json:"tags,omitempty" db:"-"`
	Ramos []Ramo      `json:"ramos,omitempty" db:"-"`
}

// NewCancao creates a new song with default values
func NewCancao(nome, linkYoutube, letra string, userID int) *Cancao {
	now := time.Now()
	return &Cancao{
		Nome:        nome,
		LinkYoutube: linkYoutube,
		Letra:       letra,
		UserID:      userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}