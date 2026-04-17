package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck returns 200 OK if the server is running.
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
