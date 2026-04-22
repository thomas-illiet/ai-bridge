package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/models"
)

type providerInfo struct {
	Name        string              `json:"name"`
	DisplayName string              `json:"displayName"`
	Type        models.ProviderType `json:"type"`
}

// ListAvailableProviders returns the name and type of all enabled providers.
func ListAvailableProviders(c *gin.Context) {
	var dbProviders []models.AIProvider
	if err := database.DB.Where("enabled = true").Order("name asc").Find(&dbProviders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]providerInfo, 0, len(dbProviders))
	for _, p := range dbProviders {
		result = append(result, providerInfo{Name: p.Name, DisplayName: p.DisplayName, Type: p.Type})
	}
	c.JSON(http.StatusOK, gin.H{"providers": result})
}
