package models

import "time"

const (
	AccessRequestPending  = "pending"
	AccessRequestApproved = "approved"
	AccessRequestRejected = "rejected"
)

// AccessRequest tracks a user's request to be granted a role.
type AccessRequest struct {
	ID         string          `gorm:"primaryKey" json:"id"` // UUID
	UserID     string          `gorm:"not null;index" json:"userId"`
	Status     string          `gorm:"not null;default:'pending'" json:"status"` // pending|approved|rejected
	Reason     string          `gorm:"not null" json:"reason"`
	ReviewNote string          `json:"reviewNote"` // admin note on rejection
	ReviewedBy string          `json:"reviewedBy"` // admin user ID
	ReviewedAt *time.Time      `json:"reviewedAt"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
	User       *User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}
