package models

import (
	"time"
)

// Ramo represents a scout branch/section
type Ramo struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// NewRamo creates a new scout branch/section
func NewRamo(name string) *Ramo {
	return &Ramo{
		Name:      name,
		CreatedAt: time.Now(),
	}
}

// LugarRamo represents the many-to-many relationship between lugares and ramos
type LugarRamo struct {
	LugarID int `json:"lugar_id" db:"lugar_id"`
	RamoID  int `json:"ramo_id" db:"ramo_id"`
}

// CancaoRamo represents the many-to-many relationship between cancoes and ramos
type CancaoRamo struct {
	CancaoID int `json:"cancao_id" db:"cancao_id"`
	RamoID   int `json:"ramo_id" db:"ramo_id"`
}