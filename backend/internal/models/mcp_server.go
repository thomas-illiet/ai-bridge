package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MCPHeaders map[string]string

func (h MCPHeaders) Value() (driver.Value, error) {
	b, err := json.Marshal(h)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (h *MCPHeaders) Scan(v interface{}) error {
	switch val := v.(type) {
	case []byte:
		return json.Unmarshal(val, h)
	case string:
		return json.Unmarshal([]byte(val), h)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

type MCPServer struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Name         string     `gorm:"not null;uniqueIndex" json:"name"`
	DisplayName  string     `gorm:"not null;default:''" json:"displayName"`
	URL          string     `gorm:"not null" json:"url"`
	Headers      MCPHeaders `gorm:"type:jsonb;default:'{}'" json:"headers"`
	AllowPattern string     `gorm:"not null;default:''" json:"allowPattern"`
	DenyPattern  string     `gorm:"not null;default:''" json:"denyPattern"`
	Enabled      bool       `gorm:"not null;default:true" json:"enabled"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

func (m *MCPServer) BeforeCreate(_ *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
