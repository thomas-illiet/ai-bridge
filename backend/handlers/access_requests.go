package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thomas-illiet/ai-bridge/config"
	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/middleware"
	"github.com/thomas-illiet/ai-bridge/models"
	"github.com/thomas-illiet/ai-bridge/services"
)

// CreateAccessRequest submits a new access request for the authenticated user.
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

		// Only one pending request at a time.
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

		// Load user record for email.
		dbUser, _ := services.GetUserByID(user.ID)
		if dbUser != nil {
			go services.SendNewRequestNotification(cfg, dbUser, &req)
		}

		c.JSON(http.StatusCreated, req)
	}
}

// GetMyAccessRequest returns the most recent access request for the authenticated user.
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

// AdminListAccessRequests returns all access requests, optionally filtered by status.
func AdminListAccessRequests(c *gin.Context) {
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

// AdminApproveRequest approves an access request and grants the user a role.
func AdminApproveRequest(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		admin := middleware.GetUser(c)

		var body struct {
			Role      string `json:"role" binding:"required"`
			ExpiresAt string `json:"expiresAt"` // optional RFC3339 date string
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "role is required"})
			return
		}
		if body.Role != models.RoleUser && body.Role != models.RoleAdmin {
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

		// Parse optional expiry.
		var expiresAt *time.Time
		if body.ExpiresAt != "" {
			t, err := time.Parse("2006-01-02", body.ExpiresAt)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "expiresAt must be YYYY-MM-DD"})
				return
			}
			// Set to end of day UTC.
			eod := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.UTC)
			expiresAt = &eod
		}

		database.DB.Save(&req)

		// Update user role.
		updates := map[string]any{"role": body.Role, "role_expires_at": expiresAt}
		database.DB.Model(&models.RegisteredUser{}).Where("id = ?", req.UserID).Updates(updates)

		dbUser, _ := services.GetUserByID(req.UserID)
		if dbUser != nil {
			go services.SendRequestApproved(cfg, dbUser)
		}

		c.JSON(http.StatusOK, req)
	}
}

// AdminRejectRequest rejects an access request with an explanatory note.
func AdminRejectRequest(cfg *config.Config) gin.HandlerFunc {
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
