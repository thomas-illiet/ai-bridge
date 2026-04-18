package admin

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/middleware"
	"github.com/thomas-illiet/ai-bridge/models"
)

func ListWhitelist(c *gin.Context) {
	var entries []models.IPWhitelistEntry
	if err := database.DB.Order("created_at desc").Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"entries": entries})
}

func AddWhitelist(c *gin.Context) {
	var body struct {
		CIDR        string `json:"cidr" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cidr := body.CIDR
	if !strings.Contains(cidr, "/") {
		cidr += "/32"
	}
	if _, _, err := net.ParseCIDR(cidr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid IP or CIDR: %s", body.CIDR)})
		return
	}

	user := middleware.GetUser(c)
	entry := models.IPWhitelistEntry{
		ID:          uuid.New(),
		CIDR:        body.CIDR,
		Description: body.Description,
		Enabled:     true,
		CreatedBy:   user.PreferredUsername,
	}
	if err := database.DB.Create(&entry).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "entry already exists or DB error: " + err.Error()})
		return
	}

	middleware.InvalidateIPCache()
	c.JSON(http.StatusCreated, entry)
}

func DeleteWhitelist(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result := database.DB.Delete(&models.IPWhitelistEntry{}, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	middleware.InvalidateIPCache()
	c.Status(http.StatusNoContent)
}

func ToggleWhitelist(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result := database.DB.Model(&models.IPWhitelistEntry{}).Where("id = ?", id).Update("enabled", body.Enabled)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	middleware.InvalidateIPCache()
	c.Status(http.StatusNoContent)
}
