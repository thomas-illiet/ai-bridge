package models

import "time"

const (
	RoleAdmin   = "admin"
	RoleManager = "manager"
	RoleUser    = "user"
	RoleNone    = "none"
	RoleService = "service"
)

// User is the persisted representation of an authenticated user.
// Role is managed by AI Bridge admins, not by Keycloak realm roles.
type User struct {
	ID            string     `gorm:"primaryKey" json:"id"` // Keycloak sub, or "svc-<uuid>" for service accounts
	Username      string     `gorm:"not null" json:"username"`
	Email         string     `json:"email"`
	Role          string     `gorm:"not null;default:'none'" json:"role"` // "admin", "user", "none", "service"
	RoleExpiresAt *time.Time `json:"roleExpiresAt"`                       // nil = never expires
	Description   string     `json:"description"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}
