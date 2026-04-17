package models

import (
	"time"

	"github.com/google/uuid"
)

type IPWhitelistEntry struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CIDR        string    `gorm:"not null;uniqueIndex" json:"cidr"`
	Description string    `json:"description"`
	Enabled     bool      `gorm:"not null;default:true" json:"enabled"`
	CreatedBy   string    `gorm:"not null" json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
