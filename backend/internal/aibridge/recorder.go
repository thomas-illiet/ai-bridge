package aibridge

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	aibrecorder "github.com/coder/aibridge/recorder"
	"github.com/google/uuid"
	"github.com/thomas-illiet/ai-bridge/internal/ctxutil"
	"github.com/thomas-illiet/ai-bridge/internal/database"
	"github.com/thomas-illiet/ai-bridge/internal/models"
)

// GORMRecorder implements aibridge.Recorder using the PostgresSQL database via GORM.
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

func (r *GORMRecorder) RecordInterception(ctx context.Context, req *aibrecorder.InterceptionRecord) error {
	startedAt := req.StartedAt
	if startedAt.IsZero() {
		startedAt = time.Now().UTC()
	}

	row := models.Interception{
		ID:           req.ID,
		InitiatorID:  req.InitiatorID,
		Provider:     req.ProviderName,
		ProviderType: req.Provider,
		Model:        req.Model,
		ClientIP:     ctxutil.ClientIPFromContext(ctx),
		StartedAt:    startedAt,
		Metadata:     marshalMeta(req.Metadata),
	}
	return database.DB.Create(&row).Error
}

func (r *GORMRecorder) RecordInterceptionEnded(_ context.Context, req *aibrecorder.InterceptionRecordEnded) error {
	endedAt := req.EndedAt
	if endedAt.IsZero() {
		endedAt = time.Now().UTC()
	}
	return database.DB.Model(&models.Interception{}).
		Where("id = ?", req.ID).
		Update("ended_at", endedAt).Error
}

func (r *GORMRecorder) RecordTokenUsage(_ context.Context, req *aibrecorder.TokenUsageRecord) error {
	meta := aibrecorder.Metadata{}
	for k, v := range req.Metadata {
		meta[k] = v
	}
	for k, v := range req.ExtraTokenTypes {
		meta[k] = v
	}

	row := models.TokenUsage{
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
	row := models.UserPrompt{
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

	row := models.ToolUsage{
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
	row := models.ModelThought{
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
