package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/middleware"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func GetMe(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func GetDashboard(c *gin.Context) {
	user := middleware.GetUser(c)
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to the dashboard",
		"user":    user.PreferredUsername,
	})
}
