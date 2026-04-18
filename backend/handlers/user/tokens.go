package user

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
	maxDaysUser  = 5
	maxDaysAdmin = 30
)

type createTokenRequest struct {
	Name         string `json:"name" binding:"required,min=1,max=100"`
	Description  string `json:"description"`
	DurationDays int    `json:"durationDays" binding:"required,min=1"`
}

type updateTokenRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description"`
}

// CreateToken creates a new client token for the authenticated user with the given name and duration.
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
		for _, r := range user.Roles {
			if r == models.RoleAdmin || r == models.RoleManager {
				maxDays = maxDaysAdmin
				break
			}
		}
		if req.DurationDays > maxDays {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("duration exceeds maximum allowed (%d days)", maxDays)})
			return
		}

		record, rawToken, err := services.CreateToken(user.ID, req.Name, req.Description, secret, req.DurationDays)
		if err != nil {
			if errors.Is(err, services.ErrTokenNameTaken) {
				c.JSON(http.StatusConflict, gin.H{"error": "a token with this name already exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"token": record, "rawToken": rawToken})
	}
}

// UpdateToken updates the name and description of a token owned by the authenticated user.
func UpdateToken(c *gin.Context) {
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

	var req updateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := services.UpdateToken(id, user.ID, req.Name, req.Description)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
			return
		}
		if errors.Is(err, services.ErrTokenNameTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "a token with this name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": record})
}

// ListTokens returns all tokens belonging to the authenticated user.
func ListTokens(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	includeRevoked := c.Query("include_revoked") == "true"
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortDir := c.DefaultQuery("sort_dir", "desc")

	tokens, err := services.ListUserTokens(user.ID, includeRevoked, sortBy, sortDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tokens": tokens})
}

// RevokeToken revokes a token owned by the authenticated user.
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
