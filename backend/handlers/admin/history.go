package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/handlers/common"
	"github.com/thomas-illiet/ai-bridge/models"
)

func GetHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	search := c.Query("search")
	userID := c.Query("userId")
	sortBy := c.DefaultQuery("sortBy", "startedAt")
	sortDir := c.DefaultQuery("sortDir", "desc")

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

	rows, total, err := common.HistoryQuery(where, args, page, pageSize, sortBy, sortDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"interceptions": rows, "total": total, "page": page, "pageSize": pageSize})
}

func GetHistoryDetail(c *gin.Context) {
	id := c.Param("id")

	var row common.InterceptionRow
	err := database.DB.Raw(common.HistoryBaseSQL+
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

	c.JSON(http.StatusOK, common.InterceptionDetail{InterceptionRow: row, Prompts: prompts})
}
