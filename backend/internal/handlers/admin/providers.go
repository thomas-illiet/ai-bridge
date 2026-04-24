package admin

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thomas-illiet/ai-bridge/internal/models"
	"github.com/thomas-illiet/ai-bridge/internal/services"
	"gorm.io/gorm"
)

type providerResponse struct {
	models.Provider
	APIKeySet bool `json:"apiKeySet"`
}

func toResponse(p *models.Provider) providerResponse {
	return providerResponse{
		Provider:  *p,
		APIKeySet: p.APIKey != "",
	}
}

func triggerReload(c *gin.Context, pub services.ReloadPublisher) error {
	return pub.PublishReload(c.Request.Context())
}

// ListProviders returns all AI providers, optionally filtered by a search term.
func ListProviders(c *gin.Context) {
	search := c.Query("search")
	providers, err := services.ListProviders(search)
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"provider": toResponse(p)})
}

// CreateProvider creates a new AI provider and signals the bridge to reload.
func CreateProvider(pub services.ReloadPublisher) gin.HandlerFunc {
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
		if reloadErr := triggerReload(c, pub); reloadErr != nil {
			c.JSON(http.StatusCreated, gin.H{"provider": toResponse(p), "reload_error": reloadErr.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"provider": toResponse(p)})
	}
}

// UpdateProvider updates an existing provider and signals the bridge to reload.
func UpdateProvider(pub services.ReloadPublisher) gin.HandlerFunc {
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
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if reloadErr := triggerReload(c, pub); reloadErr != nil {
			c.JSON(http.StatusOK, gin.H{"provider": toResponse(p), "reload_error": reloadErr.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"provider": toResponse(p)})
	}
}

// DeleteProvider soft-deletes a provider and signals the bridge to reload.
func DeleteProvider(pub services.ReloadPublisher) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		if err := services.DeleteProvider(id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if reloadErr := triggerReload(c, pub); reloadErr != nil {
			c.JSON(http.StatusOK, gin.H{"reload_error": reloadErr.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// ReloadProviders forces a reload of the bridge from the current DB state.
func ReloadProviders(pub services.ReloadPublisher) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := triggerReload(c, pub); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "reload signal sent"})
	}
}
