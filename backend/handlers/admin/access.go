package admin

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/config"
	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/handlers/common"
	"github.com/thomas-illiet/ai-bridge/middleware"
	"github.com/thomas-illiet/ai-bridge/models"
	"github.com/thomas-illiet/ai-bridge/services"
)

func ListAccessRequests(c *gin.Context) {
	status := c.Query("status")
	q := database.DB.Preload("User").Order("created_at DESC")
	if status != "" {
		q = q.Where("status = ?", status)
	}

	var requests []models.AccessRequest
	if err := q.Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var pending int64
	database.DB.Model(&models.AccessRequest{}).Where("status = ?", models.AccessRequestPending).Count(&pending)

	c.JSON(http.StatusOK, gin.H{"requests": requests, "pendingCount": pending})
}

func ApproveRequest(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		admin := middleware.GetUser(c)

		var body struct {
			Role      string `json:"role" binding:"required"`
			ExpiresAt string `json:"expiresAt"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "role is required"})
			return
		}
		if common.CallerIsManager(c) {
			if body.Role != models.RoleUser {
				c.JSON(http.StatusBadRequest, gin.H{"error": "managers can only grant the 'user' role"})
				return
			}
		} else if body.Role != models.RoleUser && body.Role != models.RoleAdmin {
			c.JSON(http.StatusBadRequest, gin.H{"error": "role must be 'user' or 'admin'"})
			return
		}

		var req models.AccessRequest
		if err := database.DB.Where("id = ?", id).First(&req).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "request not found"})
			return
		}
		if req.Status != models.AccessRequestPending {
			c.JSON(http.StatusConflict, gin.H{"error": "request already reviewed"})
			return
		}

		now := time.Now()
		req.Status = models.AccessRequestApproved
		req.ReviewedBy = admin.ID
		req.ReviewedAt = &now

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

		database.DB.Save(&req)

		updates := map[string]any{"role": body.Role, "role_expires_at": expiresAt}
		database.DB.Model(&models.RegisteredUser{}).Where("id = ?", req.UserID).Updates(updates)

		dbUser, _ := services.GetUserByID(req.UserID)
		if dbUser != nil {
			go services.SendRequestApproved(cfg, dbUser)
		}

		c.JSON(http.StatusOK, req)
	}
}

func RejectRequest(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		admin := middleware.GetUser(c)

		var body struct {
			Note string `json:"note" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "note is required"})
			return
		}

		var req models.AccessRequest
		if err := database.DB.Where("id = ?", id).First(&req).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "request not found"})
			return
		}
		if req.Status != models.AccessRequestPending {
			c.JSON(http.StatusConflict, gin.H{"error": "request already reviewed"})
			return
		}

		now := time.Now()
		req.Status = models.AccessRequestRejected
		req.ReviewNote = body.Note
		req.ReviewedBy = admin.ID
		req.ReviewedAt = &now
		database.DB.Save(&req)

		dbUser, _ := services.GetUserByID(req.UserID)
		if dbUser != nil {
			go services.SendRequestRejected(cfg, dbUser, body.Note)
		}

		c.JSON(http.StatusOK, req)
	}
}
