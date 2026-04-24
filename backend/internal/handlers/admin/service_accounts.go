package admin

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/internal/database"
	"github.com/thomas-illiet/ai-bridge/internal/models"
	"github.com/thomas-illiet/ai-bridge/internal/services"
	"gorm.io/gorm"
)

const maxDaysServiceToken = 365

type serviceAccountSummary struct {
	models.User
	TokenCount    int64 `json:"tokenCount"`
	TotalRequests int64 `json:"totalRequests"`
	TotalInput    int64 `json:"totalInput"`
	TotalOutput   int64 `json:"totalOutput"`
}

var allowedServiceAccountSortColumns = map[string]string{
	"username":       "u.username",
	"created_at":     "u.created_at",
	"token_count":    "token_count",
	"total_requests": "total_requests",
	"total_input":    "total_input",
	"total_output":   "total_output",
}

// ListServiceAccounts returns all service accounts with token and usage statistics.
func ListServiceAccounts(c *gin.Context) {
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortDir := c.DefaultQuery("sort_dir", "desc")
	col, ok := allowedServiceAccountSortColumns[sortBy]
	if !ok {
		col = "u.created_at"
	}
	if sortDir != "asc" {
		sortDir = "desc"
	}

	search := c.Query("search")

	whereClause := "\nWHERE u.role = ?"
	args := []interface{}{models.RoleService}
	if search != "" {
		whereClause += " AND (u.username ILIKE ? OR u.description ILIKE ?)"
		like := "%" + search + "%"
		args = append(args, like, like)
	}

	rawSQL := `
		SELECT u.*,
			COALESCE(tk.token_count, 0)   AS token_count,
			COALESCE(req.cnt, 0)          AS total_requests,
			COALESCE(tok.total_input, 0)  AS total_input,
			COALESCE(tok.total_output, 0) AS total_output
		FROM users u
		LEFT JOIN (
			SELECT user_id, COUNT(*) AS token_count
			FROM api_tokens WHERE revoked_at IS NULL
			GROUP BY user_id
		) tk ON tk.user_id = u.id
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
		) tok ON tok.initiator_id = u.id` + whereClause + `
		ORDER BY ` + col + ` ` + sortDir

	var accounts []serviceAccountSummary
	if err := database.DB.Raw(rawSQL, args...).Scan(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if accounts == nil {
		accounts = []serviceAccountSummary{}
	}
	c.JSON(http.StatusOK, gin.H{"serviceAccounts": accounts})
}

// CreateServiceAccount creates a new service account with the given username and description.
func CreateServiceAccount(c *gin.Context) {
	var body struct {
		Username    string `json:"username" binding:"required,min=1,max=100"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	account, err := services.CreateServiceAccount(body.Username, body.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, account)
}

// DeleteServiceAccount removes a service account and its associated tokens.
func DeleteServiceAccount(c *gin.Context) {
	id := c.Param("id")
	if err := services.DeleteServiceAccount(id); err != nil {
		if err.Error() == "not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "service account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ListServiceTokens returns all tokens belonging to a service account.
func ListServiceTokens(c *gin.Context) {
	id := c.Param("id")

	var count int64
	database.DB.Model(&models.User{}).Where("id = ? AND role = ?", id, models.RoleService).Count(&count)
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "service account not found"})
		return
	}

	includeInactive := c.Query("include_inactive") == "true"
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortDir := c.DefaultQuery("sort_dir", "desc")
	tokens, err := services.ListUserTokens(id, includeInactive, sortBy, sortDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tokens == nil {
		tokens = []models.APIToken{}
	}
	c.JSON(http.StatusOK, gin.H{"tokens": tokens})
}

// CreateServiceToken creates a new token for the specified service account.
func CreateServiceToken(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var account models.User
		if err := database.DB.Where("id = ? AND role = ?", id, models.RoleService).First(&account).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "service account not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var body struct {
			Name         string `json:"name" binding:"required,min=1,max=100"`
			DurationDays int    `json:"durationDays" binding:"required,min=1"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if body.DurationDays > maxDaysServiceToken {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("duration exceeds maximum allowed (%d days)", maxDaysServiceToken),
			})
			return
		}

		record, rawToken, err := services.CreateToken(id, body.Name, "", secret, body.DurationDays)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"token": record, "rawToken": rawToken})
	}
}
