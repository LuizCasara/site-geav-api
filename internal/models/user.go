package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"-" db:"password"` // Password is not included in JSON responses
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserRole represents the possible roles for a user
type UserRole string

const (
	// RoleRead represents a user with read-only access
	RoleRead UserRole = "read"
	// RoleWrite represents a user with read and write access
	RoleWrite UserRole = "write"
)

// NewUser creates a new user with default values
func NewUser(username, password string, role UserRole) *User {
	now := time.Now()
	return &User{
		Username:  username,
		Password:  password, // Note: In a real application, this should be hashed
		Role:      string(role),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IsValidRole checks if the role is valid
func IsValidRole(role string) bool {
	return role == string(RoleRead) || role == string(RoleWrite)
}

// HasWriteAccess checks if the user has write access
func (u *User) HasWriteAccess() bool {
	return u.Role == string(RoleWrite)
}