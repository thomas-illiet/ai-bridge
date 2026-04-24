package models

import (
	"time"

	"github.com/google/uuid"
)

type FirewallRule struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CIDR        string    `gorm:"not null;uniqueIndex" json:"cidr"`
	Description string    `json:"description"`
	Action      string    `gorm:"not null;default:'allow'" json:"action"`
	Priority    int       `gorm:"not null;default:100" json:"priority"`
	Enabled   bool      `gorm:"not null;default:true" json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (FirewallRule) TableName() string { return "firewall_rules" }
