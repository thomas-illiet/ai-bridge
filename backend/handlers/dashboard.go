package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/middleware"
	"github.com/thomas-illiet/ai-bridge/models"
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

type tokenTotals struct {
	TotalInput  int64 `json:"totalInput"`
	TotalOutput int64 `json:"totalOutput"`
}

type lastRequestInfo struct {
	Model     string    `json:"model"`
	Provider  string    `json:"provider"`
	StartedAt time.Time `json:"startedAt"`
}

// GetMe returns the authenticated user's profile.
func GetMe(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// GetDashboard returns aggregated usage stats for the authenticated user.
func GetDashboard(c *gin.Context) {
	user := middleware.GetUser(c)

	var totalRequests int64
	database.DB.Model(&models.AibridgeInterception{}).Count(&totalRequests)

	var tokens tokenTotals
	database.DB.Model(&models.AibridgeTokenUsage{}).
		Select("COALESCE(SUM(input_tokens), 0) as total_input, COALESCE(SUM(output_tokens), 0) as total_output").
		Scan(&tokens)

	since := time.Now().UTC().AddDate(0, 0, -6).Truncate(24 * time.Hour)

	daily := []dailyCount{}
	database.DB.Model(&models.AibridgeInterception{}).
		Select("TO_CHAR(started_at, 'YYYY-MM-DD') as date, COUNT(*) as count").
		Where("started_at >= ?", since).
		Group("TO_CHAR(started_at, 'YYYY-MM-DD')").
		Order("date ASC").
		Scan(&daily)

	dailyToks := []dailyTokens{}
	database.DB.Model(&models.AibridgeTokenUsage{}).
		Joins("JOIN aibridge_interceptions ON aibridge_interceptions.id = aibridge_token_usages.interception_id").
		Select("TO_CHAR(aibridge_interceptions.started_at, 'YYYY-MM-DD') as date, SUM(aibridge_token_usages.input_tokens + aibridge_token_usages.output_tokens) as total").
		Where("aibridge_interceptions.started_at >= ?", since).
		Group("TO_CHAR(aibridge_interceptions.started_at, 'YYYY-MM-DD')").
		Order("date ASC").
		Scan(&dailyToks)

	byProvider := []providerCount{}
	database.DB.Model(&models.AibridgeInterception{}).
		Select("provider, COUNT(*) as count").
		Group("provider").
		Order("count DESC").
		Limit(5).
		Scan(&byProvider)

	byModel := []modelCount{}
	database.DB.Model(&models.AibridgeInterception{}).
		Select("model, COUNT(*) as count").
		Group("model").
		Order("count DESC").
		Limit(5).
		Scan(&byModel)

	var activeUsers int64
	database.DB.Model(&models.AibridgeInterception{}).
		Distinct("initiator_id").Count(&activeUsers)

	var lastReq *lastRequestInfo
	var tmp lastRequestInfo
	database.DB.Model(&models.AibridgeInterception{}).
		Select("model, provider, started_at").
		Order("started_at DESC").
		Limit(1).
		Scan(&tmp)
	if tmp.Model != "" {
		lastReq = &tmp
	}

	c.JSON(http.StatusOK, gin.H{
		"user":          user.PreferredUsername,
		"totalRequests": totalRequests,
		"activeUsers":   activeUsers,
		"tokens":        tokens,
		"daily":         daily,
		"dailyTokens":   dailyToks,
		"byProvider":    byProvider,
		"byModel":       byModel,
		"lastRequest":   lastReq,
	})
}
