package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/middleware"
	"github.com/thomas-illiet/ai-bridge/models"
	"gorm.io/gorm"
)

type userSummary struct {
	models.RegisteredUser
	TotalRequests int64 `json:"totalRequests"`
	TotalInput    int64 `json:"totalInput"`
	TotalOutput   int64 `json:"totalOutput"`
}

type userRequestCount struct {
	InitiatorID string `json:"initiatorId"`
	Count       int64  `json:"count"`
}

type userTokenSum struct {
	InitiatorID string `json:"initiatorId"`
	TotalInput  int64  `json:"totalInput"`
	TotalOutput int64  `json:"totalOutput"`
}

type userDailyCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type userProviderCount struct {
	Provider string `json:"provider"`
	Count    int64  `json:"count"`
}

type userModelCount struct {
	Model string `json:"model"`
	Count int64  `json:"count"`
}

// ListUsers returns all registered users with aggregated request and token stats.
func ListUsers(c *gin.Context) {
	var users []models.RegisteredUser
	if err := database.DB.Order("created_at asc").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var reqCounts []userRequestCount
	database.DB.Model(&models.AibridgeInterception{}).
		Select("initiator_id, COUNT(*) as count").
		Group("initiator_id").
		Scan(&reqCounts)

	var tokenSums []userTokenSum
	database.DB.Model(&models.AibridgeTokenUsage{}).
		Select("ai.initiator_id, COALESCE(SUM(aibridge_token_usages.input_tokens),0) as total_input, COALESCE(SUM(aibridge_token_usages.output_tokens),0) as total_output").
		Joins("JOIN aibridge_interceptions ai ON ai.id = aibridge_token_usages.interception_id").
		Group("ai.initiator_id").
		Scan(&tokenSums)

	reqMap := make(map[string]int64, len(reqCounts))
	for _, r := range reqCounts {
		reqMap[r.InitiatorID] = r.Count
	}
	tokMap := make(map[string]userTokenSum, len(tokenSums))
	for _, t := range tokenSums {
		tokMap[t.InitiatorID] = t
	}

	result := make([]userSummary, len(users))
	for i, u := range users {
		result[i] = userSummary{
			RegisteredUser: u,
			TotalRequests:  reqMap[u.ID],
			TotalInput:     tokMap[u.ID].TotalInput,
			TotalOutput:    tokMap[u.ID].TotalOutput,
		}
	}
	c.JSON(http.StatusOK, gin.H{"users": result})
}

// GetUserStats returns detailed usage stats for a specific user.
func GetUserStats(c *gin.Context) {
	id := c.Param("id")

	var totalRequests int64
	database.DB.Model(&models.AibridgeInterception{}).
		Where("initiator_id = ?", id).Count(&totalRequests)

	var tokens userTokenSum
	database.DB.Model(&models.AibridgeTokenUsage{}).
		Select("'' as initiator_id, COALESCE(SUM(aibridge_token_usages.input_tokens),0) as total_input, COALESCE(SUM(aibridge_token_usages.output_tokens),0) as total_output").
		Joins("JOIN aibridge_interceptions ai ON ai.id = aibridge_token_usages.interception_id").
		Where("ai.initiator_id = ?", id).
		Scan(&tokens)

	since := time.Now().UTC().AddDate(0, 0, -6).Truncate(24 * time.Hour)
	daily := make([]userDailyCount, 0)
	database.DB.Model(&models.AibridgeInterception{}).
		Select("TO_CHAR(started_at, 'YYYY-MM-DD') as date, COUNT(*) as count").
		Where("initiator_id = ? AND started_at >= ?", id, since).
		Group("TO_CHAR(started_at, 'YYYY-MM-DD')").Order("date ASC").
		Scan(&daily)

	byProvider := make([]userProviderCount, 0)
	database.DB.Model(&models.AibridgeInterception{}).
		Select("provider, COUNT(*) as count").
		Where("initiator_id = ?", id).
		Group("provider").Order("count DESC").
		Scan(&byProvider)

	byModel := make([]userModelCount, 0)
	database.DB.Model(&models.AibridgeInterception{}).
		Select("model, COUNT(*) as count").
		Where("initiator_id = ?", id).
		Group("model").Order("count DESC").Limit(10).
		Scan(&byModel)

	c.JSON(http.StatusOK, gin.H{
		"totalRequests": totalRequests,
		"totalInput":    tokens.TotalInput,
		"totalOutput":   tokens.TotalOutput,
		"daily":         daily,
		"byProvider":    byProvider,
		"byModel":       byModel,
	})
}

// DeleteUser removes a user and all their tokens; admin cannot delete themselves.
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	caller := middleware.GetUser(c)
	if caller != nil && caller.ID == id {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete your own account"})
		return
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", id).Delete(&models.ClientToken{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&models.RegisteredUser{}, "id = ?", id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("not found")
		}
		return nil
	})
	if err != nil {
		if err.Error() == "not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// UpdateUserRole sets the role of a user (admin, user, or none).
func UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Role != models.RoleAdmin && body.Role != models.RoleUser && body.Role != models.RoleNone {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role: must be admin, user, or none"})
		return
	}
	caller := middleware.GetUser(c)
	if caller != nil && caller.ID == id && body.Role != models.RoleAdmin {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot change your own role"})
		return
	}
	result := database.DB.Model(&models.RegisteredUser{}).Where("id = ?", id).Update("role", body.Role)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.Status(http.StatusNoContent)
}
