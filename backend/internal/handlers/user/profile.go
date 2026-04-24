package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/internal/middleware"
)

// GetMe returns the profile of the currently authenticated user.
func GetMe(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	c.JSON(http.StatusOK, user)
}
