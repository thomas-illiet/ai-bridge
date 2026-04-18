package common

import (
	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/middleware"
	"github.com/thomas-illiet/ai-bridge/models"
)

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
