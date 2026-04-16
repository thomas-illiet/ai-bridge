package aibridge

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	aibrecorder "github.com/coder/aibridge/recorder"
	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/models"
)

// GORMRecorder implements aibridge.Recorder using the PostgreSQL database via GORM.
type GORMRecorder struct{}

var _ aibrecorder.Recorder = (*GORMRecorder)(nil)

func NewGORMRecorder() *GORMRecorder {
	return &GORMRecorder{}
}

func marshalMeta(m aibrecorder.Metadata) string {
	if m == nil {
		return "{}"
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func marshalAny(v any) string {
	if v == nil {
		return "{}"
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func (r *GORMRecorder) RecordInterception(_ context.Context, req *aibrecorder.InterceptionRecord) error {
	row := models.AibridgeInterception{
		ID:          req.ID,
		InitiatorID: req.InitiatorID,
		Provider:    req.Provider,
		Model:       req.Model,
		StartedAt:   req.StartedAt,
		Metadata:    marshalMeta(req.Metadata),
	}
	return database.DB.Create(&row).Error
}

func (r *GORMRecorder) RecordInterceptionEnded(_ context.Context, req *aibrecorder.InterceptionRecordEnded) error {
	return database.DB.Model(&models.AibridgeInterception{}).
		Where("id = ?", req.ID).
		Update("ended_at", req.EndedAt).Error
}

func (r *GORMRecorder) RecordTokenUsage(_ context.Context, req *aibrecorder.TokenUsageRecord) error {
	meta := aibrecorder.Metadata{}
	for k, v := range req.Metadata {
		meta[k] = v
	}
	for k, v := range req.ExtraTokenTypes {
		meta[k] = v
	}

	row := models.AibridgeTokenUsage{
		ID:                    uuid.NewString(),
		InterceptionID:        req.InterceptionID,
		ProviderResponseID:    req.MsgID,
		InputTokens:           req.Input,
		OutputTokens:          req.Output,
		CacheReadInputTokens:  req.CacheReadInputTokens,
		CacheWriteInputTokens: req.CacheWriteInputTokens,
		Metadata:              marshalMeta(meta),
		CreatedAt:             req.CreatedAt,
	}
	return database.DB.Create(&row).Error
}

func (r *GORMRecorder) RecordPromptUsage(_ context.Context, req *aibrecorder.PromptUsageRecord) error {
	row := models.AibridgeUserPrompt{
		ID:                 uuid.NewString(),
		InterceptionID:     req.InterceptionID,
		ProviderResponseID: req.MsgID,
		Prompt:             req.Prompt,
		Metadata:           marshalMeta(req.Metadata),
		CreatedAt:          req.CreatedAt,
	}
	return database.DB.Create(&row).Error
}

func (r *GORMRecorder) RecordToolUsage(_ context.Context, req *aibrecorder.ToolUsageRecord) error {
	var invErr *string
	if req.InvocationError != nil {
		s := req.InvocationError.Error()
		invErr = &s
	}

	row := models.AibridgeToolUsage{
		ID:                 uuid.NewString(),
		InterceptionID:     req.InterceptionID,
		ProviderResponseID: req.MsgID,
		ServerURL:          req.ServerURL,
		Tool:               req.Tool,
		Input:              marshalAny(req.Args),
		Injected:           req.Injected,
		InvocationError:    invErr,
		Metadata:           marshalMeta(req.Metadata),
		CreatedAt:          req.CreatedAt,
	}
	return database.DB.Create(&row).Error
}

func (r *GORMRecorder) RecordModelThought(_ context.Context, req *aibrecorder.ModelThoughtRecord) error {
	row := models.AibridgeModelThought{
		ID:             uuid.NewString(),
		InterceptionID: req.InterceptionID,
		Content:        req.Content,
		Metadata:       marshalMeta(req.Metadata),
		CreatedAt:      req.CreatedAt,
	}
	if err := database.DB.Create(&row).Error; err != nil {
		return fmt.Errorf("record model thought: %w", err)
	}
	return nil
}
