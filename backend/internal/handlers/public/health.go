package public

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck responds with a 200 OK status to indicate the service is alive.
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
