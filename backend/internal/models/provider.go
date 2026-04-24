package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProviderType string

const (
	ProviderTypeOpenAI    ProviderType = "openai"
	ProviderTypeOllama    ProviderType = "ollama"
	ProviderTypeAnthropic ProviderType = "anthropic"
)

type ProviderConfig map[string]interface{}

func (c ProviderConfig) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (c *ProviderConfig) Scan(v interface{}) error {
	switch val := v.(type) {
	case []byte:
		return json.Unmarshal(val, c)
	case string:
		return json.Unmarshal([]byte(val), c)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

type Provider struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string         `gorm:"not null;uniqueIndex" json:"name"`
	DisplayName string         `gorm:"not null;default:''" json:"displayName"`
	Type        ProviderType   `gorm:"not null" json:"type"`
	BaseURL     string         `gorm:"not null;default:''" json:"baseUrl"`
	APIKey      string         `gorm:"not null;default:''" json:"-"`
	Config      ProviderConfig `gorm:"type:jsonb;default:'{}'" json:"config"`
	Enabled     bool           `gorm:"not null;default:true" json:"enabled"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (p *Provider) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
