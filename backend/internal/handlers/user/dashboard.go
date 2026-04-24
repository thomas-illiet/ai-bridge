package user

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/internal/database"
	"github.com/thomas-illiet/ai-bridge/internal/handlers/common"
	"github.com/thomas-illiet/ai-bridge/internal/middleware"
	"github.com/thomas-illiet/ai-bridge/internal/models"
)

type dailyCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type dailyTokens struct {
	Date  string `json:"date"`
	Total int64  `json:"total"`
}

type providerCount struct {
	Provider string `json:"provider"`
	Count    int64  `json:"count"`
}

type modelCount struct {
	Model string `json:"model"`
	Count int64  `json:"count"`
}

type modelTokens struct {
	Model string `json:"model"`
	Total int64  `json:"total"`
}

type toolCount struct {
	Tool  string `json:"tool"`
	Count int64  `json:"count"`
}

type tokenTotals struct {
	TotalInput  int64 `json:"totalInput"`
	TotalOutput int64 `json:"totalOutput"`
}

type lastRequestInfo struct {
	Model     string    `json:"model"`
	Provider  string    `json:"provider"`
	StartedAt time.Time `json:"startedAt"`
}

// dashboardScope resolves scope and returns isGlobal + uid for user scope.
func dashboardScope(c *gin.Context) (isGlobal bool, uid string) {
	user := middleware.GetUser(c)
	if c.DefaultQuery("scope", "user") == "global" && common.CallerIsElevated(c) {
		return true, ""
	}
	if user != nil {
		return false, user.ID
	}
	return false, ""
}

// GetTotalRequests returns the total number of interceptions.
func GetTotalRequests(c *gin.Context) {
	isGlobal, uid := dashboardScope(c)
	var total int64
	q := database.DB.Model(&models.Interception{})
	if !isGlobal {
		q = q.Where("initiator_id = ?", uid)
	}
	q.Count(&total)
	c.JSON(http.StatusOK, gin.H{"totalRequests": total})
}

// GetTokenTotals returns aggregated input and output token counts.
func GetTokenTotals(c *gin.Context) {
	isGlobal, uid := dashboardScope(c)
	var totals tokenTotals
	q := database.DB.Model(&models.TokenUsage{}).
		Select("COALESCE(SUM(input_tokens), 0) as total_input, COALESCE(SUM(output_tokens), 0) as total_output")
	if !isGlobal {
		q = q.Joins("JOIN interceptions ON interceptions.id = token_usages.interception_id").
			Where("interceptions.initiator_id = ?", uid)
	}
	q.Scan(&totals)
	c.JSON(http.StatusOK, totals)
}

// GetDailyRequests returns per-day request counts for the last 7 days.
func GetDailyRequests(c *gin.Context) {
	isGlobal, uid := dashboardScope(c)
	since := time.Now().UTC().AddDate(0, 0, -6).Truncate(24 * time.Hour)
	var rows []dailyCount
	q := database.DB.Model(&models.Interception{}).
		Select("TO_CHAR(started_at, 'YYYY-MM-DD') as date, COUNT(*) as count").
		Group("TO_CHAR(started_at, 'YYYY-MM-DD')").
		Order("date ASC")
	if isGlobal {
		q = q.Where("started_at >= ?", since)
	} else {
		q = q.Where("initiator_id = ? AND started_at >= ?", uid, since)
	}
	q.Scan(&rows)
	if rows == nil {
		rows = []dailyCount{}
	}
	c.JSON(http.StatusOK, gin.H{"daily": rows})
}

// GetDailyTokens returns per-day token usage for the last 7 days.
func GetDailyTokens(c *gin.Context) {
	isGlobal, uid := dashboardScope(c)
	since := time.Now().UTC().AddDate(0, 0, -6).Truncate(24 * time.Hour)
	var rows []dailyTokens
	q := database.DB.Model(&models.TokenUsage{}).
		Joins("JOIN interceptions ON interceptions.id = token_usages.interception_id").
		Select("TO_CHAR(interceptions.started_at, 'YYYY-MM-DD') as date, SUM(token_usages.input_tokens + token_usages.output_tokens) as total").
		Group("TO_CHAR(interceptions.started_at, 'YYYY-MM-DD')").
		Order("date ASC")
	if isGlobal {
		q = q.Where("interceptions.started_at >= ?", since)
	} else {
		q = q.Where("interceptions.initiator_id = ? AND interceptions.started_at >= ?", uid, since)
	}
	q.Scan(&rows)
	if rows == nil {
		rows = []dailyTokens{}
	}
	c.JSON(http.StatusOK, gin.H{"dailyTokens": rows})
}

// GetByProvider returns request counts grouped by provider (top 5).
func GetByProvider(c *gin.Context) {
	isGlobal, uid := dashboardScope(c)
	var rows []providerCount
	q := database.DB.Model(&models.Interception{}).
		Select("provider, COUNT(*) as count").
		Group("provider").Order("count DESC").Limit(5)
	if !isGlobal {
		q = q.Where("initiator_id = ?", uid)
	}
	q.Scan(&rows)
	if rows == nil {
		rows = []providerCount{}
	}
	c.JSON(http.StatusOK, gin.H{"byProvider": rows})
}

// GetByModel returns request counts grouped by model (top 5).
func GetByModel(c *gin.Context) {
	isGlobal, uid := dashboardScope(c)
	var rows []modelCount
	q := database.DB.Model(&models.Interception{}).
		Select("model, COUNT(*) as count").
		Group("model").Order("count DESC").Limit(5)
	if !isGlobal {
		q = q.Where("initiator_id = ?", uid)
	}
	q.Scan(&rows)
	if rows == nil {
		rows = []modelCount{}
	}
	c.JSON(http.StatusOK, gin.H{"byModel": rows})
}

// GetTokensByModel returns token consumption grouped by model (top 8).
func GetTokensByModel(c *gin.Context) {
	isGlobal, uid := dashboardScope(c)
	var rows []modelTokens
	q := database.DB.Model(&models.TokenUsage{}).
		Joins("JOIN interceptions ON interceptions.id = token_usages.interception_id").
		Select("interceptions.model, SUM(token_usages.input_tokens + token_usages.output_tokens) as total").
		Group("interceptions.model").Order("total DESC").Limit(8)
	if !isGlobal {
		q = q.Where("interceptions.initiator_id = ?", uid)
	}
	q.Scan(&rows)
	if rows == nil {
		rows = []modelTokens{}
	}
	c.JSON(http.StatusOK, gin.H{"tokensByModel": rows})
}

// GetToolsUsed returns tool call counts (top 8).
func GetToolsUsed(c *gin.Context) {
	isGlobal, uid := dashboardScope(c)
	var rows []toolCount
	q := database.DB.Model(&models.ToolUsage{}).
		Joins("JOIN interceptions ON interceptions.id = tool_usages.interception_id").
		Select("tool_usages.tool as tool, COUNT(*) as count").
		Group("tool_usages.tool").Order("count DESC").Limit(8)
	if !isGlobal {
		q = q.Where("interceptions.initiator_id = ?", uid)
	}
	q.Scan(&rows)
	if rows == nil {
		rows = []toolCount{}
	}
	c.JSON(http.StatusOK, gin.H{"toolsUsed": rows})
}

// GetLastRequest returns metadata for the most recent interception.
func GetLastRequest(c *gin.Context) {
	isGlobal, uid := dashboardScope(c)
	var tmp lastRequestInfo
	q := database.DB.Model(&models.Interception{}).
		Select("model, provider, started_at").
		Order("started_at DESC").Limit(1)
	if !isGlobal {
		q = q.Where("initiator_id = ?", uid)
	}
	q.Scan(&tmp)
	if tmp.Model == "" {
		c.JSON(http.StatusOK, gin.H{"lastRequest": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"lastRequest": tmp})
}

// GetActiveUsers returns the count of distinct users with at least one interception. Requires elevated role.
func GetActiveUsers(c *gin.Context) {
	if !common.CallerIsElevated(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "elevated role required"})
		return
	}
	var count int64
	database.DB.Model(&models.Interception{}).Distinct("initiator_id").Count(&count)
	c.JSON(http.StatusOK, gin.H{"activeUsers": count})
}
