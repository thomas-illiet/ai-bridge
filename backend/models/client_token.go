package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClientToken struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     string         `gorm:"not null;index" json:"userId"`
	Name       string         `gorm:"not null" json:"name"`
	TokenHash  string         `gorm:"not null;uniqueIndex" json:"-"`
	LastUsedAt *time.Time     `json:"lastUsedAt"`
	RevokedAt  *time.Time     `json:"revokedAt"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (t *ClientToken) BeforeCreate(_ *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (t *ClientToken) IsRevoked() bool {
	return t.RevokedAt != nil
}
