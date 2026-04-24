package admin

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thomas-illiet/ai-bridge/internal/services"
	"gorm.io/gorm"
)

// ListMCPServers returns all MCP servers, optionally filtered by search term.
func ListMCPServers(c *gin.Context) {
	servers, err := services.ListMCPServers(c.Query("search"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"mcp_servers": servers})
}

// GetMCPServer returns a single MCP server by ID.
func GetMCPServer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	s, err := services.GetMCPServer(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"mcp_server": s})
}

// CreateMCPServer creates a new MCP server and signals the bridge to reload.
func CreateMCPServer(pub services.ReloadPublisher) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req services.CreateMCPServerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		s, err := services.CreateMCPServer(req)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if reloadErr := triggerReload(c, pub); reloadErr != nil {
			c.JSON(http.StatusCreated, gin.H{"mcp_server": s, "reload_error": reloadErr.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"mcp_server": s})
	}
}

// UpdateMCPServer updates an existing MCP server and signals the bridge to reload.
func UpdateMCPServer(pub services.ReloadPublisher) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		var req services.UpdateMCPServerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		s, err := services.UpdateMCPServer(id, req)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if reloadErr := triggerReload(c, pub); reloadErr != nil {
			c.JSON(http.StatusOK, gin.H{"mcp_server": s, "reload_error": reloadErr.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"mcp_server": s})
	}
}

// DeleteMCPServer deletes an MCP server and signals the bridge to reload.
func DeleteMCPServer(pub services.ReloadPublisher) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		if err := services.DeleteMCPServer(id); err != nil {
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

// ReloadMCP forces a reload of the bridge from the current DB state.
func ReloadMCP(pub services.ReloadPublisher) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := triggerReload(c, pub); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "reload signal sent"})
	}
}
