package user

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/handlers/common"
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

// GetMe returns the profile of the currently authenticated user.
func GetMe(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// GetDashboard returns usage statistics scoped to the current user, or globally for elevated roles.
func GetDashboard(c *gin.Context) {
	user := middleware.GetUser(c)

	scope := c.DefaultQuery("scope", "user")
	if scope == "global" && !common.CallerIsElevated(c) {
		scope = "user"
	}

	since := time.Now().UTC().AddDate(0, 0, -6).Truncate(24 * time.Hour)

	var totalRequests int64
	var tokens tokenTotals
	daily := []dailyCount{}
	dailyToks := []dailyTokens{}
	byProvider := []providerCount{}
	byModel := []modelCount{}
	var lastReq *lastRequestInfo
	var activeUsers int64

	if scope == "global" {
		database.DB.Model(&models.AibridgeInterception{}).Count(&totalRequests)

		database.DB.Model(&models.AibridgeTokenUsage{}).
			Select("COALESCE(SUM(input_tokens), 0) as total_input, COALESCE(SUM(output_tokens), 0) as total_output").
			Scan(&tokens)

		database.DB.Model(&models.AibridgeInterception{}).
			Select("TO_CHAR(started_at, 'YYYY-MM-DD') as date, COUNT(*) as count").
			Where("started_at >= ?", since).
			Group("TO_CHAR(started_at, 'YYYY-MM-DD')").
			Order("date ASC").
			Scan(&daily)

		database.DB.Model(&models.AibridgeTokenUsage{}).
			Joins("JOIN aibridge_interceptions ON aibridge_interceptions.id = aibridge_token_usages.interception_id").
			Select("TO_CHAR(aibridge_interceptions.started_at, 'YYYY-MM-DD') as date, SUM(aibridge_token_usages.input_tokens + aibridge_token_usages.output_tokens) as total").
			Where("aibridge_interceptions.started_at >= ?", since).
			Group("TO_CHAR(aibridge_interceptions.started_at, 'YYYY-MM-DD')").
			Order("date ASC").
			Scan(&dailyToks)

		database.DB.Model(&models.AibridgeInterception{}).
			Select("provider, COUNT(*) as count").
			Group("provider").Order("count DESC").Limit(5).
			Scan(&byProvider)

		database.DB.Model(&models.AibridgeInterception{}).
			Select("model, COUNT(*) as count").
			Group("model").Order("count DESC").Limit(5).
			Scan(&byModel)

		database.DB.Model(&models.AibridgeInterception{}).
			Distinct("initiator_id").Count(&activeUsers)

		var tmp lastRequestInfo
		database.DB.Model(&models.AibridgeInterception{}).
			Select("model, provider, started_at").
			Order("started_at DESC").Limit(1).
			Scan(&tmp)
		if tmp.Model != "" {
			lastReq = &tmp
		}
	} else {
		uid := user.ID

		database.DB.Model(&models.AibridgeInterception{}).
			Where("initiator_id = ?", uid).Count(&totalRequests)

		database.DB.Model(&models.AibridgeTokenUsage{}).
			Select("COALESCE(SUM(input_tokens), 0) as total_input, COALESCE(SUM(output_tokens), 0) as total_output").
			Joins("JOIN aibridge_interceptions ON aibridge_interceptions.id = aibridge_token_usages.interception_id").
			Where("aibridge_interceptions.initiator_id = ?", uid).
			Scan(&tokens)

		database.DB.Model(&models.AibridgeInterception{}).
			Select("TO_CHAR(started_at, 'YYYY-MM-DD') as date, COUNT(*) as count").
			Where("initiator_id = ? AND started_at >= ?", uid, since).
			Group("TO_CHAR(started_at, 'YYYY-MM-DD')").
			Order("date ASC").
			Scan(&daily)

		database.DB.Model(&models.AibridgeTokenUsage{}).
			Joins("JOIN aibridge_interceptions ON aibridge_interceptions.id = aibridge_token_usages.interception_id").
			Select("TO_CHAR(aibridge_interceptions.started_at, 'YYYY-MM-DD') as date, SUM(aibridge_token_usages.input_tokens + aibridge_token_usages.output_tokens) as total").
			Where("aibridge_interceptions.initiator_id = ? AND aibridge_interceptions.started_at >= ?", uid, since).
			Group("TO_CHAR(aibridge_interceptions.started_at, 'YYYY-MM-DD')").
			Order("date ASC").
			Scan(&dailyToks)

		database.DB.Model(&models.AibridgeInterception{}).
			Select("provider, COUNT(*) as count").
			Where("initiator_id = ?", uid).
			Group("provider").Order("count DESC").Limit(5).
			Scan(&byProvider)

		database.DB.Model(&models.AibridgeInterception{}).
			Select("model, COUNT(*) as count").
			Where("initiator_id = ?", uid).
			Group("model").Order("count DESC").Limit(5).
			Scan(&byModel)

		var tmp lastRequestInfo
		database.DB.Model(&models.AibridgeInterception{}).
			Select("model, provider, started_at").
			Where("initiator_id = ?", uid).
			Order("started_at DESC").Limit(1).
			Scan(&tmp)
		if tmp.Model != "" {
			lastReq = &tmp
		}
	}

	resp := gin.H{
		"user":          user.PreferredUsername,
		"scope":         scope,
		"totalRequests": totalRequests,
		"tokens":        tokens,
		"daily":         daily,
		"dailyTokens":   dailyToks,
		"byProvider":    byProvider,
		"byModel":       byModel,
		"lastRequest":   lastReq,
	}
	if scope == "global" {
		resp["activeUsers"] = activeUsers
	}
	c.JSON(http.StatusOK, resp)
}
