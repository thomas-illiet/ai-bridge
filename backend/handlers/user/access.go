package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thomas-illiet/ai-bridge/config"
	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/middleware"
	"github.com/thomas-illiet/ai-bridge/models"
	"github.com/thomas-illiet/ai-bridge/services"
)

func CreateAccessRequest(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := middleware.GetUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}

		var body struct {
			Reason string `json:"reason" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "reason is required"})
			return
		}

		var existing models.AccessRequest
		err := database.DB.Where("user_id = ? AND status = ?", user.ID, models.AccessRequestPending).First(&existing).Error
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "you already have a pending request"})
			return
		}

		req := models.AccessRequest{
			ID:     uuid.NewString(),
			UserID: user.ID,
			Status: models.AccessRequestPending,
			Reason: body.Reason,
		}
		if err := database.DB.Create(&req).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		dbUser, _ := services.GetUserByID(user.ID)
		if dbUser != nil {
			go services.SendNewRequestNotification(cfg, dbUser, &req)
		}

		c.JSON(http.StatusCreated, req)
	}
}

func GetMyAccessRequest(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	var req models.AccessRequest
	err := database.DB.Where("user_id = ?", user.ID).
		Order("created_at DESC").
		First(&req).Error
	if err != nil {
		c.JSON(http.StatusOK, nil)
		return
	}
	c.JSON(http.StatusOK, req)
}
