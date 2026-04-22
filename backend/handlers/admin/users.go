package admin

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/handlers/common"
	"github.com/thomas-illiet/ai-bridge/middleware"
	"github.com/thomas-illiet/ai-bridge/models"
	"gorm.io/gorm"
)

type userSummary struct {
	models.User
	TotalRequests int64 `json:"totalRequests"`
	TotalInput    int64 `json:"totalInput"`
	TotalOutput   int64 `json:"totalOutput"`
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

var allowedUserSortColumns = map[string]string{
	"username":        "u.username",
	"email":           "u.email",
	"role":            "u.role",
	"role_expires_at": "u.role_expires_at",
	"created_at":      "u.created_at",
	"total_requests":  "total_requests",
	"total_input":     "total_input",
	"total_output":    "total_output",
}

// ListUsers returns all registered users with their aggregated request and token statistics.
func ListUsers(c *gin.Context) {
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortDir := c.DefaultQuery("sort_dir", "asc")
	col, ok := allowedUserSortColumns[sortBy]
	if !ok {
		col = "u.created_at"
	}
	if sortDir != "asc" {
		sortDir = "desc"
	}

	var result []userSummary
	rawSQL := `
		SELECT u.*,
			COALESCE(req.cnt, 0)        AS total_requests,
			COALESCE(tok.total_input, 0)  AS total_input,
			COALESCE(tok.total_output, 0) AS total_output
		FROM users u
		LEFT JOIN (
			SELECT initiator_id, COUNT(*) AS cnt
			FROM interceptions
			GROUP BY initiator_id
		) req ON req.initiator_id = u.id
		LEFT JOIN (
			SELECT ai.initiator_id,
				COALESCE(SUM(atu.input_tokens),  0) AS total_input,
				COALESCE(SUM(atu.output_tokens), 0) AS total_output
			FROM token_usages atu
			JOIN interceptions ai ON ai.id = atu.interception_id
			GROUP BY ai.initiator_id
		) tok ON tok.initiator_id = u.id
		ORDER BY ` + col + ` ` + sortDir

	if err := database.DB.Raw(rawSQL).Scan(&result).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result == nil {
		result = []userSummary{}
	}
	c.JSON(http.StatusOK, gin.H{"users": result})
}

// GetUserStats returns detailed usage statistics for a specific user.
func GetUserStats(c *gin.Context) {
	id := c.Param("id")
	if common.CallerIsManager(c) {
		var target models.User
		if err := database.DB.Where("id = ?", id).First(&target).Error; err != nil ||
			target.Role == models.RoleAdmin || target.Role == models.RoleService {
			c.JSON(http.StatusForbidden, gin.H{"error": "managers cannot view stats for this user"})
			return
		}
	}

	var totalRequests int64
	database.DB.Model(&models.Interception{}).
		Where("initiator_id = ?", id).Count(&totalRequests)

	var tokens userTokenSum
	database.DB.Model(&models.TokenUsage{}).
		Select("'' as initiator_id, COALESCE(SUM(token_usages.input_tokens),0) as total_input, COALESCE(SUM(token_usages.output_tokens),0) as total_output").
		Joins("JOIN interceptions ai ON ai.id = token_usages.interception_id").
		Where("ai.initiator_id = ?", id).
		Scan(&tokens)

	since := time.Now().UTC().AddDate(0, 0, -6).Truncate(24 * time.Hour)
	daily := make([]userDailyCount, 0)
	database.DB.Model(&models.Interception{}).
		Select("TO_CHAR(started_at, 'YYYY-MM-DD') as date, COUNT(*) as count").
		Where("initiator_id = ? AND started_at >= ?", id, since).
		Group("TO_CHAR(started_at, 'YYYY-MM-DD')").Order("date ASC").
		Scan(&daily)

	byProvider := make([]userProviderCount, 0)
	database.DB.Model(&models.Interception{}).
		Select("provider, COUNT(*) as count").
		Where("initiator_id = ?", id).
		Group("provider").Order("count DESC").
		Scan(&byProvider)

	byModel := make([]userModelCount, 0)
	database.DB.Model(&models.Interception{}).
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

// DeleteUser removes a user and all their tokens from the database.
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	caller := middleware.GetUser(c)
	if caller != nil && caller.ID == id {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete your own account"})
		return
	}
	if common.CallerIsManager(c) {
		var target models.User
		if err := database.DB.Where("id = ?", id).First(&target).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if target.Role == models.RoleAdmin || target.Role == models.RoleService {
			c.JSON(http.StatusForbidden, gin.H{"error": "managers cannot delete this user"})
			return
		}
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", id).Delete(&models.APIToken{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&models.User{}, "id = ?", id)
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

// UpdateUserRole updates the role and optional expiry date of a user.
func UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Role      string `json:"role" binding:"required"`
		ExpiresAt string `json:"expiresAt"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Role != models.RoleAdmin && body.Role != models.RoleManager &&
		body.Role != models.RoleUser && body.Role != models.RoleNone {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role: must be admin, manager, user, or none"})
		return
	}
	caller := middleware.GetUser(c)
	if common.CallerIsManager(c) {
		if body.Role != models.RoleUser && body.Role != models.RoleNone {
			c.JSON(http.StatusForbidden, gin.H{"error": "managers can only assign user or none roles"})
			return
		}
		var target models.User
		if err := database.DB.Where("id = ?", id).First(&target).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if target.Role == models.RoleAdmin || target.Role == models.RoleService {
			c.JSON(http.StatusForbidden, gin.H{"error": "managers cannot modify this user"})
			return
		}
	}
	if caller != nil && caller.ID == id && body.Role != models.RoleAdmin {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot change your own role"})
		return
	}

	var expiresAt *time.Time
	if body.ExpiresAt != "" {
		t, err := time.Parse("2006-01-02", body.ExpiresAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "expiresAt must be YYYY-MM-DD"})
			return
		}
		eod := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.UTC)
		expiresAt = &eod
	}

	updates := map[string]any{"role": body.Role, "role_expires_at": expiresAt}
	result := database.DB.Model(&models.User{}).Where("id = ?", id).Updates(updates)
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
