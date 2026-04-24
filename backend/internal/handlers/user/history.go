package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/internal/database"
	"github.com/thomas-illiet/ai-bridge/internal/handlers/common"
	"github.com/thomas-illiet/ai-bridge/internal/middleware"
	"github.com/thomas-illiet/ai-bridge/internal/models"
)

// GetHistory returns a paginated list of interceptions for the authenticated user.
func GetHistory(c *gin.Context) {
	user := middleware.GetUser(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	search := c.Query("search")
	sortBy := c.DefaultQuery("sortBy", "startedAt")
	sortDir := c.DefaultQuery("sortDir", "desc")

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

	rows, total, err := common.HistoryQuery(where, args, page, pageSize, sortBy, sortDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"interceptions": rows, "total": total, "page": page, "pageSize": pageSize})
}

// GetHistoryStats returns aggregate token and request statistics for the authenticated user.
func GetHistoryStats(c *gin.Context) {
	user := middleware.GetUser(c)

	type statsRow struct {
		Total        int64 `gorm:"column:total"`
		TotalInput   int64 `gorm:"column:total_input"`
		TotalOutput  int64 `gorm:"column:total_output"`
	}
	var stats statsRow
	database.DB.Raw(`
		SELECT
			COUNT(DISTINCT ai.id)                    AS total,
			COALESCE(SUM(atu.input_tokens),  0)      AS total_input,
			COALESCE(SUM(atu.output_tokens), 0)      AS total_output
		FROM interceptions ai
		LEFT JOIN token_usages atu ON atu.interception_id = ai.id
		WHERE ai.initiator_id = ?
`, user.ID).Scan(&stats)

	var topModel string
	database.DB.Raw(`
		SELECT model FROM interceptions
		WHERE initiator_id = ?
		GROUP BY model ORDER BY COUNT(*) DESC LIMIT 1
`, user.ID).Scan(&topModel)

	c.JSON(http.StatusOK, gin.H{
		"total":       stats.Total,
		"totalInput":  stats.TotalInput,
		"totalOutput": stats.TotalOutput,
		"topModel":    topModel,
	})
}

// GetHistoryDetail returns the full details and prompts of a single interception owned by the authenticated user.
func GetHistoryDetail(c *gin.Context) {
	user := middleware.GetUser(c)
	id := c.Param("id")

	var row common.InterceptionRow
	err := database.DB.Raw(common.HistoryBaseSQL+
		" WHERE ai.id = ? AND ai.initiator_id = ? "+
		" GROUP BY ai.id, ai.initiator_id, ai.provider, ai.model, ai.started_at, ai.ended_at, ru.username",
		id, user.ID).Scan(&row).Error
	if err != nil || row.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var prompts []string
	database.DB.Model(&models.UserPrompt{}).
		Where("interception_id = ?", id).
		Order("created_at ASC").
		Pluck("prompt", &prompts)

	c.JSON(http.StatusOK, common.InterceptionDetail{InterceptionRow: row, Prompts: prompts})
}
