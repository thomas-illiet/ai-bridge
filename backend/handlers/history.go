package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/middleware"
	"github.com/thomas-illiet/ai-bridge/models"
)

type interceptionRow struct {
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

type interceptionDetail struct {
	interceptionRow
	Prompts []string `json:"prompts"`
}

const historyBaseSQL = `
	SELECT
		ai.id, ai.initiator_id, ai.provider, ai.model, ai.started_at, ai.ended_at,
		COALESCE(ru.username, ai.initiator_id) AS username,
		COALESCE(SUM(atu.input_tokens),  0) AS input_tokens,
		COALESCE(SUM(atu.output_tokens), 0) AS output_tokens
	FROM aibridge_interceptions ai
	LEFT JOIN registered_users          ru  ON ru.id  = ai.initiator_id
	LEFT JOIN aibridge_token_usages     atu ON atu.interception_id = ai.id
`

// historyQuery runs a paginated interception query with a shared WHERE clause.
func historyQuery(whereSQL string, whereArgs []any, page, pageSize int) ([]interceptionRow, int64, error) {
	countSQL := `SELECT COUNT(DISTINCT ai.id)
		FROM aibridge_interceptions ai
		LEFT JOIN registered_users      ru  ON ru.id = ai.initiator_id
		LEFT JOIN aibridge_token_usages atu ON atu.interception_id = ai.id
		` + whereSQL

	var total int64
	if err := database.DB.Raw(countSQL, whereArgs...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	dataSQL := historyBaseSQL + whereSQL +
		` GROUP BY ai.id, ai.initiator_id, ai.provider, ai.model, ai.started_at, ai.ended_at, ru.username
		 ORDER BY ai.started_at DESC LIMIT ? OFFSET ?`
	dataArgs := append(append([]any{}, whereArgs...), pageSize, (page-1)*pageSize)

	rows := make([]interceptionRow, 0)
	if err := database.DB.Raw(dataSQL, dataArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// GetHistory returns paginated interception history for the authenticated user.
func GetHistory(c *gin.Context) {
	user := middleware.GetUser(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	where := " WHERE ai.initiator_id = ? "
	args := []any{user.ID}

	if search != "" {
		where += " AND (ai.model ILIKE ? OR ai.provider ILIKE ?) "
		like := "%" + search + "%"
		args = append(args, like, like)
	}

	rows, total, err := historyQuery(where, args, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"interceptions": rows, "total": total, "page": page, "pageSize": pageSize})
}

// GetHistoryDetail returns a single interception with its prompts (owner only).
func GetHistoryDetail(c *gin.Context) {
	user := middleware.GetUser(c)
	id := c.Param("id")

	var row interceptionRow
	err := database.DB.Raw(historyBaseSQL+
		" WHERE ai.id = ? AND ai.initiator_id = ? "+
		" GROUP BY ai.id, ai.initiator_id, ai.provider, ai.model, ai.started_at, ai.ended_at, ru.username",
		id, user.ID).Scan(&row).Error
	if err != nil || row.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var prompts []string
	database.DB.Model(&models.AibridgeUserPrompt{}).
		Where("interception_id = ?", id).
		Order("created_at ASC").
		Pluck("prompt", &prompts)

	c.JSON(http.StatusOK, interceptionDetail{interceptionRow: row, Prompts: prompts})
}
