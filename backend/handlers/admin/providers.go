package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	aibpkg "github.com/thomas-illiet/ai-bridge/aibridge"
	"github.com/thomas-illiet/ai-bridge/models"
	"github.com/thomas-illiet/ai-bridge/services"
	"gorm.io/gorm"
)

type providerResponse struct {
	models.AIProvider
	APIKeySet bool `json:"apiKeySet"`
}

func toResponse(p *models.AIProvider) providerResponse {
	return providerResponse{
		AIProvider: *p,
		APIKeySet:  p.APIKey != "",
	}
}

func triggerReload(bm *aibpkg.BridgeManager) error {
	providers, err := services.BuildProviders()
	if err != nil {
		return err
	}
	return bm.Reload(providers)
}

// ListProviders returns all AI providers.
func ListProviders(c *gin.Context) {
	providers, err := services.ListProviders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := make([]providerResponse, len(providers))
	for i := range providers {
		resp[i] = toResponse(&providers[i])
	}
	c.JSON(http.StatusOK, gin.H{"providers": resp})
}

// GetProvider returns a single provider by ID.
func GetProvider(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	p, err := services.GetProvider(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"provider": toResponse(p)})
}

// CreateProvider creates a new AI provider and hot-reloads the bridge.
func CreateProvider(bm *aibpkg.BridgeManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req services.CreateProviderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		p, err := services.CreateProvider(req)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if reloadErr := triggerReload(bm); reloadErr != nil {
			c.JSON(http.StatusCreated, gin.H{"provider": toResponse(p), "reload_error": reloadErr.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"provider": toResponse(p)})
	}
}

// UpdateProvider updates an existing provider and hot-reloads the bridge.
func UpdateProvider(bm *aibpkg.BridgeManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		var req services.UpdateProviderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		p, err := services.UpdateProvider(id, req)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if reloadErr := triggerReload(bm); reloadErr != nil {
			c.JSON(http.StatusOK, gin.H{"provider": toResponse(p), "reload_error": reloadErr.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"provider": toResponse(p)})
	}
}

// DeleteProvider soft-deletes a provider and hot-reloads the bridge.
func DeleteProvider(bm *aibpkg.BridgeManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		if err := services.DeleteProvider(id); err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if reloadErr := triggerReload(bm); reloadErr != nil {
			c.JSON(http.StatusOK, gin.H{"reload_error": reloadErr.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// ReloadProviders forces a hot-reload of the bridge from the current DB state.
func ReloadProviders(bm *aibpkg.BridgeManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := triggerReload(bm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "bridge reloaded"})
	}
}
