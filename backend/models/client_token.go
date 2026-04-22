package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type APIToken struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      string         `gorm:"not null;index" json:"userId"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description"`
	TokenHash   string         `gorm:"not null;uniqueIndex" json:"-"`
	ExpiresAt   *time.Time     `json:"expiresAt"`
	LastUsedAt  *time.Time     `json:"lastUsedAt"`
	RevokedAt   *time.Time     `json:"revokedAt"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (t *APIToken) BeforeCreate(_ *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (t *APIToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

func (t *APIToken) IsExpired() bool {
	return t.ExpiresAt != nil && time.Now().After(*t.ExpiresAt)
}
