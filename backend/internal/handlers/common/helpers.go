package common

import (
	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/internal/middleware"
	"github.com/thomas-illiet/ai-bridge/internal/models"
)

// CallerIsManager returns true if the authenticated user has the manager role.
func CallerIsManager(c *gin.Context) bool {
	user := middleware.GetUser(c)
	if user == nil {
		return false
	}
	for _, r := range user.Roles {
		if r == models.RoleManager {
			return true
		}
	}
	return false
}

// CallerIsElevated returns true if the authenticated user has the admin or manager role.
func CallerIsElevated(c *gin.Context) bool {
	user := middleware.GetUser(c)
	if user == nil {
		return false
	}
	for _, r := range user.Roles {
		if r == models.RoleAdmin || r == models.RoleManager {
			return true
		}
	}
	return false
}
