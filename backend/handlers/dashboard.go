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

type providerCount struct {
	Provider string `json:"provider"`
	Count    int64  `json:"count"`
}

type tokenTotals struct {
	TotalInput  int64 `json:"totalInput"`
	TotalOutput int64 `json:"totalOutput"`
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
	var daily []dailyCount
	database.DB.Model(&models.AibridgeInterception{}).
		Select("TO_CHAR(started_at, 'YYYY-MM-DD') as date, COUNT(*) as count").
		Where("started_at >= ?", since).
		Group("TO_CHAR(started_at, 'YYYY-MM-DD')").
		Order("date ASC").
		Scan(&daily)

	var byProvider []providerCount
	database.DB.Model(&models.AibridgeInterception{}).
		Select("provider, COUNT(*) as count").
		Group("provider").
		Order("count DESC").
		Scan(&byProvider)

	c.JSON(http.StatusOK, gin.H{
		"user":          user.PreferredUsername,
		"totalRequests": totalRequests,
		"tokens":        tokens,
		"daily":         daily,
		"byProvider":    byProvider,
	})
}
