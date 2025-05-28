package models

import (
	"time"
)

// TagLugar represents a tag that can be applied to a place
type TagLugar struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TagCancao represents a tag that can be applied to a song
type TagCancao struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// NewTagLugar creates a new tag for places
func NewTagLugar(name string) *TagLugar {
	return &TagLugar{
		Name:      name,
		CreatedAt: time.Now(),
	}
}

// NewTagCancao creates a new tag for songs
func NewTagCancao(name string) *TagCancao {
	return &TagCancao{
		Name:      name,
		CreatedAt: time.Now(),
	}
}

// LugarTag represents the many-to-many relationship between lugares and tags
type LugarTag struct {
	LugarID int `json:"lugar_id" db:"lugar_id"`
	TagID   int `json:"tag_id" db:"tag_id"`
}

// CancaoTag represents the many-to-many relationship between cancoes and tags
type CancaoTag struct {
	CancaoID int `json:"cancao_id" db:"cancao_id"`
	TagID    int `json:"tag_id" db:"tag_id"`
}