package admin

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/internal/database"
	"github.com/thomas-illiet/ai-bridge/internal/handlers/common"
	"github.com/thomas-illiet/ai-bridge/internal/models"
)

// checkUserStatsAccess denies managers from viewing admin/service user stats. Returns false and writes 403 on denial.
func checkUserStatsAccess(c *gin.Context, id string) bool {
	if common.CallerIsManager(c) {
		var target models.User
		if err := database.DB.Where("id = ?", id).First(&target).Error; err != nil ||
			target.Role == models.RoleAdmin || target.Role == models.RoleService {
			c.JSON(http.StatusForbidden, gin.H{"error": "managers cannot view stats for this user"})
			return false
		}
	}
	return true
}

// GetUserTotalRequests returns the total interception count for a user.
func GetUserTotalRequests(c *gin.Context) {
	id := c.Param("id")
	if !checkUserStatsAccess(c, id) {
		return
	}
	var total int64
	database.DB.Model(&models.Interception{}).Where("initiator_id = ?", id).Count(&total)
	c.JSON(http.StatusOK, gin.H{"totalRequests": total})
}

// GetUserTokenTotals returns aggregated token usage for a user.
func GetUserTokenTotals(c *gin.Context) {
	id := c.Param("id")
	if !checkUserStatsAccess(c, id) {
		return
	}
	var totals userTokenSum
	database.DB.Model(&models.TokenUsage{}).
		Select("'' as initiator_id, COALESCE(SUM(token_usages.input_tokens),0) as total_input, COALESCE(SUM(token_usages.output_tokens),0) as total_output").
		Joins("JOIN interceptions ai ON ai.id = token_usages.interception_id").
		Where("ai.initiator_id = ?", id).
		Scan(&totals)
	c.JSON(http.StatusOK, gin.H{"totalInput": totals.TotalInput, "totalOutput": totals.TotalOutput})
}

// GetUserDailyRequests returns per-day request counts for a user over the last 7 days.
func GetUserDailyRequests(c *gin.Context) {
	id := c.Param("id")
	if !checkUserStatsAccess(c, id) {
		return
	}
	since := time.Now().UTC().AddDate(0, 0, -6).Truncate(24 * time.Hour)
	daily := make([]userDailyCount, 0)
	database.DB.Model(&models.Interception{}).
		Select("TO_CHAR(started_at, 'YYYY-MM-DD') as date, COUNT(*) as count").
		Where("initiator_id = ? AND started_at >= ?", id, since).
		Group("TO_CHAR(started_at, 'YYYY-MM-DD')").Order("date ASC").
		Scan(&daily)
	c.JSON(http.StatusOK, gin.H{"daily": daily})
}

// GetUserByProvider returns provider breakdown for a user.
func GetUserByProvider(c *gin.Context) {
	id := c.Param("id")
	if !checkUserStatsAccess(c, id) {
		return
	}
	rows := make([]userProviderCount, 0)
	database.DB.Model(&models.Interception{}).
		Select("provider, COUNT(*) as count").
		Where("initiator_id = ?", id).
		Group("provider").Order("count DESC").
		Scan(&rows)
	c.JSON(http.StatusOK, gin.H{"byProvider": rows})
}

// GetUserByModel returns top models by request count for a user.
func GetUserByModel(c *gin.Context) {
	id := c.Param("id")
	if !checkUserStatsAccess(c, id) {
		return
	}
	rows := make([]userModelCount, 0)
	database.DB.Model(&models.Interception{}).
		Select("model, COUNT(*) as count").
		Where("initiator_id = ?", id).
		Group("model").Order("count DESC").Limit(10).
		Scan(&rows)
	c.JSON(http.StatusOK, gin.H{"byModel": rows})
}
