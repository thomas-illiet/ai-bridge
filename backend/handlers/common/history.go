package common

import (
	"strings"
	"time"

	"github.com/thomas-illiet/ai-bridge/database"
)

type InterceptionRow struct {
	ID           string     `gorm:"column:id"            json:"id"`
	InitiatorID  string     `gorm:"column:initiator_id"  json:"initiatorId"`
	Username     string     `gorm:"column:username"      json:"username"`
	Provider     string     `gorm:"column:provider"      json:"provider"`
	Model        string     `gorm:"column:model"         json:"model"`
	StartedAt    time.Time  `gorm:"column:started_at"    json:"startedAt"`
	EndedAt      *time.Time `gorm:"column:ended_at"      json:"endedAt"`
	InputTokens  int64      `gorm:"column:input_tokens"  json:"inputTokens"`
	OutputTokens int64      `gorm:"column:output_tokens" json:"outputTokens"`
}

type InterceptionDetail struct {
	InterceptionRow
	Prompts []string `json:"prompts"`
}

const HistoryBaseSQL = `
	SELECT
		ai.id, ai.initiator_id, ai.provider, ai.model, ai.started_at, ai.ended_at,
		COALESCE(ru.username, ai.initiator_id) AS username,
		COALESCE(SUM(atu.input_tokens),  0) AS input_tokens,
		COALESCE(SUM(atu.output_tokens), 0) AS output_tokens
	FROM aibridge_interceptions ai
	LEFT JOIN registered_users          ru  ON ru.id  = ai.initiator_id
	LEFT JOIN aibridge_token_usages     atu ON atu.interception_id = ai.id
`

// sortableColumns maps frontend sort keys to safe SQL expressions.
var sortableColumns = map[string]string{
	"provider":     "ai.provider",
	"model":        "ai.model",
	"startedAt":    "ai.started_at",
	"duration":     "EXTRACT(EPOCH FROM (COALESCE(ai.ended_at, NOW()) - ai.started_at))",
	"inputTokens":  "input_tokens",
	"outputTokens": "output_tokens",
	"username":     "COALESCE(ru.username, ai.initiator_id)",
}

// HistoryQuery runs a paginated, sortable interception query with a shared WHERE clause.
func HistoryQuery(whereSQL string, whereArgs []any, page, pageSize int, sortBy, sortDir string) ([]InterceptionRow, int64, error) {
	countSQL := `SELECT COUNT(DISTINCT ai.id)
		FROM aibridge_interceptions ai
		LEFT JOIN registered_users      ru  ON ru.id = ai.initiator_id
		LEFT JOIN aibridge_token_usages atu ON atu.interception_id = ai.id
		` + whereSQL

	var total int64
	if err := database.DB.Raw(countSQL, whereArgs...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	col, ok := sortableColumns[sortBy]
	if !ok {
		col = "ai.started_at"
	}
	dir := "DESC"
	if strings.EqualFold(sortDir, "asc") {
		dir = "ASC"
	}

	dataSQL := HistoryBaseSQL + whereSQL +
		` GROUP BY ai.id, ai.initiator_id, ai.provider, ai.model, ai.started_at, ai.ended_at, ru.username
		 ORDER BY ` + col + ` ` + dir + ` LIMIT ? OFFSET ?`
	dataArgs := append(append([]any{}, whereArgs...), pageSize, (page-1)*pageSize)

	rows := make([]InterceptionRow, 0)
	if err := database.DB.Raw(dataSQL, dataArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
