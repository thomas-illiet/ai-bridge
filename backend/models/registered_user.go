package models

import "time"

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
	RoleNone  = "none"
)

// RegisteredUser is the persisted representation of an authenticated user.
// Role is managed by AI Bridge admins, not by Keycloak realm roles.
type RegisteredUser struct {
	ID            string     `gorm:"primaryKey" json:"id"` // Keycloak sub
	Username      string     `gorm:"not null" json:"username"`
	Email         string     `json:"email"`
	Role          string     `gorm:"not null;default:'none'" json:"role"` // "admin", "user", "none"
	RoleExpiresAt *time.Time `json:"roleExpiresAt"`                       // nil = never expires
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}
