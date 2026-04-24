package models

import "time"

type Interception struct {
	ID           string     `gorm:"primaryKey" json:"id"`
	InitiatorID  string     `gorm:"not null;index" json:"initiatorId"`
	Initiator    *User      `gorm:"foreignKey:InitiatorID;references:ID;constraint:OnDelete:RESTRICT" json:"-"`
	Provider     string     `gorm:"not null" json:"provider"`
	ProviderType string     `gorm:"not null;default:''" json:"providerType"`
	Model        string     `gorm:"not null" json:"model"`
	ClientIP     string     `gorm:"not null;default:''" json:"clientIp"`
	StartedAt    time.Time  `gorm:"not null" json:"startedAt"`
	EndedAt      *time.Time `json:"endedAt"`
	Metadata     string     `json:"metadata"`
}

type TokenUsage struct {
	ID                    string        `gorm:"primaryKey" json:"id"`
	InterceptionID        string        `gorm:"not null;index" json:"interceptionId"`
	Interception          *Interception `gorm:"foreignKey:InterceptionID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	ProviderResponseID    string    `gorm:"not null" json:"providerResponseId"`
	InputTokens           int64     `gorm:"not null;default:0" json:"inputTokens"`
	OutputTokens          int64     `gorm:"not null;default:0" json:"outputTokens"`
	CacheReadInputTokens  int64     `gorm:"not null;default:0" json:"cacheReadInputTokens"`
	CacheWriteInputTokens int64     `gorm:"not null;default:0" json:"cacheWriteInputTokens"`
	Metadata              string    `json:"metadata"`
	CreatedAt             time.Time `gorm:"not null" json:"createdAt"`
}

type UserPrompt struct {
	ID                 string        `gorm:"primaryKey" json:"id"`
	InterceptionID     string        `gorm:"not null;index" json:"interceptionId"`
	Interception       *Interception `gorm:"foreignKey:InterceptionID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	ProviderResponseID string    `gorm:"not null" json:"providerResponseId"`
	Prompt             string    `gorm:"not null" json:"prompt"`
	Metadata           string    `json:"metadata"`
	CreatedAt          time.Time `gorm:"not null" json:"createdAt"`
}

type ToolUsage struct {
	ID                 string        `gorm:"primaryKey" json:"id"`
	InterceptionID     string        `gorm:"not null;index" json:"interceptionId"`
	Interception       *Interception `gorm:"foreignKey:InterceptionID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	ProviderResponseID string     `gorm:"not null" json:"providerResponseId"`
	ServerURL          *string    `json:"serverUrl"`
	Tool               string     `gorm:"not null" json:"tool"`
	Input              string     `gorm:"not null" json:"input"`
	Injected           bool       `gorm:"not null;default:false" json:"injected"`
	InvocationError    *string    `json:"invocationError"`
	Metadata           string     `json:"metadata"`
	CreatedAt          time.Time  `gorm:"not null" json:"createdAt"`
}

type ModelThought struct {
	ID             string        `gorm:"primaryKey" json:"id"`
	InterceptionID string        `gorm:"not null;index" json:"interceptionId"`
	Interception   *Interception `gorm:"foreignKey:InterceptionID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	Content        string    `gorm:"not null" json:"content"`
	Metadata       string    `json:"metadata"`
	CreatedAt      time.Time `gorm:"not null" json:"createdAt"`
}
