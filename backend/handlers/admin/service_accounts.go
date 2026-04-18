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

func ListServiceAccounts(c *gin.Context) {
	var accounts []models.RegisteredUser
	if err := database.DB.Where("role = ?", models.RoleService).Order("created_at desc").Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if accounts == nil {
		accounts = []models.RegisteredUser{}
	}
	c.JSON(http.StatusOK, gin.H{"serviceAccounts": accounts})
}

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

func ListServiceTokens(c *gin.Context) {
	id := c.Param("id")

	var count int64
	database.DB.Model(&models.RegisteredUser{}).Where("id = ? AND role = ?", id, models.RoleService).Count(&count)
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "service account not found"})
		return
	}

	includeRevoked := c.Query("include_revoked") == "true"
	tokens, err := services.ListUserTokens(id, includeRevoked)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tokens == nil {
		tokens = []models.ClientToken{}
	}
	c.JSON(http.StatusOK, gin.H{"tokens": tokens})
}

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

		record, rawToken, err := services.CreateToken(id, body.Name, secret, body.DurationDays)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"token": record, "rawToken": rawToken})
	}
}
