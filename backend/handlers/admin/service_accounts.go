package admin

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/models"
	"github.com/thomas-illiet/ai-bridge/services"
	"gorm.io/gorm"
)

const maxDaysServiceToken = 365

var allowedServiceAccountSortColumns = map[string]string{
	"username":   "username",
	"created_at": "created_at",
}

// ListServiceAccounts returns all users with the service role.
func ListServiceAccounts(c *gin.Context) {
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortDir := c.DefaultQuery("sort_dir", "desc")
	col, ok := allowedServiceAccountSortColumns[sortBy]
	if !ok {
		col = "created_at"
	}
	if sortDir != "asc" {
		sortDir = "desc"
	}

	var accounts []models.RegisteredUser
	if err := database.DB.Where("role = ?", models.RoleService).Order(col + " " + sortDir).Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if accounts == nil {
		accounts = []models.RegisteredUser{}
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
	database.DB.Model(&models.RegisteredUser{}).Where("id = ? AND role = ?", id, models.RoleService).Count(&count)
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
		tokens = []models.ClientToken{}
	}
	c.JSON(http.StatusOK, gin.H{"tokens": tokens})
}

// CreateServiceToken creates a new token for the specified service account.
func CreateServiceToken(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var account models.RegisteredUser
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
