package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thomas-illiet/ai-bridge/middleware"
	"github.com/thomas-illiet/ai-bridge/models"
	"github.com/thomas-illiet/ai-bridge/services"
	"gorm.io/gorm"
)

const (
	maxDaysUser  = 30
	maxDaysAdmin = 365
)

type createTokenRequest struct {
	Name         string `json:"name" binding:"required,min=1,max=100"`
	DurationDays int    `json:"durationDays" binding:"required,min=1"`
}

// CreateToken mints a new PAT for the authenticated user.
func CreateToken(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user := middleware.GetUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}

		maxDays := maxDaysUser
		if len(user.Roles) > 0 && user.Roles[0] == models.RoleAdmin {
			maxDays = maxDaysAdmin
		}
		if req.DurationDays > maxDays {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("duration exceeds maximum allowed (%d days)", maxDays)})
			return
		}

		record, rawToken, err := services.CreateToken(user.ID, req.Name, secret, req.DurationDays)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"token": record, "rawToken": rawToken})
	}
}

// ListTokens returns tokens for the authenticated user.
// Query param: include_revoked=true to include revoked tokens (default: false).
func ListTokens(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	includeRevoked := c.Query("include_revoked") == "true"

	tokens, err := services.ListUserTokens(user.ID, includeRevoked)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tokens": tokens})
}

// RevokeToken marks the caller's token as revoked.
func RevokeToken(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token id"})
		return
	}

	if err := services.RevokeToken(id, user.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke token"})
		return
	}

	c.Status(http.StatusNoContent)
}
