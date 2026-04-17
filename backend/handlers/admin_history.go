package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/models"
)

// AdminGetHistory returns all interceptions with optional user and search filters.
func AdminGetHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	search := c.Query("search")
	userID := c.Query("userId")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	where := " WHERE 1=1 "
	args := []any{}

	if userID != "" {
		where += " AND ai.initiator_id = ? "
		args = append(args, userID)
	}
	if search != "" {
		where += " AND (ai.model ILIKE ? OR ai.provider ILIKE ? OR ru.username ILIKE ?) "
		like := "%" + search + "%"
		args = append(args, like, like, like)
	}

	rows, total, err := historyQuery(where, args, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"interceptions": rows, "total": total, "page": page, "pageSize": pageSize})
}

// AdminGetHistoryDetail returns any interception detail regardless of owner.
func AdminGetHistoryDetail(c *gin.Context) {
	id := c.Param("id")

	var row interceptionRow
	err := database.DB.Raw(historyBaseSQL+
		" WHERE ai.id = ? "+
		" GROUP BY ai.id, ai.initiator_id, ai.provider, ai.model, ai.started_at, ai.ended_at, ru.username",
		id).Scan(&row).Error
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
