package admin

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thomas-illiet/ai-bridge/internal/database"
	"github.com/thomas-illiet/ai-bridge/internal/middleware"
	"github.com/thomas-illiet/ai-bridge/internal/models"
	"github.com/thomas-illiet/ai-bridge/internal/services"
	"gorm.io/gorm"
)


var allowedFirewallSortColumns = map[string]string{
	"cidr":       "cidr",
	"created_at": "created_at",
	"enabled":    "enabled",
	"priority":   "priority",
	"action":     "action",
}

// ReloadFirewall forces an immediate invalidation of the firewall cache on both the API and bridge.
func ReloadFirewall(pub services.ReloadPublisher) gin.HandlerFunc {
	return func(c *gin.Context) {
		middleware.InvalidateFirewallCache()
		if err := pub.PublishReload(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "reload signal sent"})
	}
}

// ListFirewallRules returns all firewall rules.
func ListFirewallRules(c *gin.Context) {
	sortBy := c.DefaultQuery("sort_by", "priority")
	sortDir := c.DefaultQuery("sort_dir", "asc")
	col, ok := allowedFirewallSortColumns[sortBy]
	if !ok {
		col = "priority"
	}
	if sortDir != "asc" {
		sortDir = "desc"
	}

	search := c.Query("search")

	q := database.DB.Order(gorm.Expr(col + " " + sortDir))
	if search != "" {
		q = q.Where("cidr ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var entries []models.FirewallRule
	if err := q.Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"entries": entries})
}

// AddFirewallRule adds a new firewall rule and invalidates the firewall cache.
func AddFirewallRule(pub services.ReloadPublisher) gin.HandlerFunc {
	return func(c *gin.Context) { addFirewallRule(c, pub) }
}

func addFirewallRule(c *gin.Context, pub services.ReloadPublisher) {
	var body struct {
		CIDR        string `json:"cidr" binding:"required"`
		Description string `json:"description"`
		Action      string `json:"action"`
		Priority    int    `json:"priority"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	action := body.Action
	if action == "" {
		action = "allow"
	}
	if action != "allow" && action != "deny" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "action must be 'allow' or 'deny'"})
		return
	}

	priority := body.Priority
	if priority == 0 {
		priority = 100
	}

	var existing models.FirewallRule
	if err := database.DB.Where("priority = ?", priority).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("priority %d is already used by another rule", priority)})
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

	entry := models.FirewallRule{
		ID:          uuid.New(),
		CIDR:        body.CIDR,
		Description: body.Description,
		Action:      action,
		Priority:    priority,
		Enabled:     true,
	}
	if err := database.DB.Create(&entry).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "entry already exists or DB error: " + err.Error()})
		return
	}

	middleware.InvalidateFirewallCache()
	_ = pub.PublishReload(c.Request.Context())
	c.JSON(http.StatusCreated, entry)
}

// DeleteFirewallRule removes a firewall rule by ID and invalidates the firewall cache.
func DeleteFirewallRule(pub services.ReloadPublisher) gin.HandlerFunc {
	return func(c *gin.Context) { deleteFirewallRule(c, pub) }
}

func deleteFirewallRule(c *gin.Context, pub services.ReloadPublisher) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result := database.DB.Delete(&models.FirewallRule{}, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	middleware.InvalidateFirewallCache()
	_ = pub.PublishReload(c.Request.Context())
	c.Status(http.StatusNoContent)
}

// ToggleFirewallRule enables or disables a firewall rule and invalidates the firewall cache.
func ToggleFirewallRule(pub services.ReloadPublisher) gin.HandlerFunc {
	return func(c *gin.Context) { toggleFirewallRule(c, pub) }
}

func toggleFirewallRule(c *gin.Context, pub services.ReloadPublisher) {
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
	result := database.DB.Model(&models.FirewallRule{}).Where("id = ?", id).Update("enabled", body.Enabled)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	middleware.InvalidateFirewallCache()
	_ = pub.PublishReload(c.Request.Context())
	c.Status(http.StatusNoContent)
}

// MoveFirewallRulePriority moves a rule one position up or down in priority order,
// reassigning all priorities with step 10 to keep them unique and well-spaced.
func MoveFirewallRulePriority(pub services.ReloadPublisher) gin.HandlerFunc {
	return func(c *gin.Context) { moveFirewallRulePriority(c, pub) }
}

func moveFirewallRulePriority(c *gin.Context, pub services.ReloadPublisher) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body struct {
		Direction  string   `json:"direction" binding:"required"`
		OrderedIDs []string `json:"ordered_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Direction != "up" && body.Direction != "down" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "direction must be 'up' or 'down'"})
		return
	}

	// Parse and validate all IDs from the frontend-provided order.
	orderedIDs := make([]uuid.UUID, 0, len(body.OrderedIDs))
	for _, raw := range body.OrderedIDs {
		uid, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id in ordered_ids: " + raw})
			return
		}
		orderedIDs = append(orderedIDs, uid)
	}

	// Find position of target rule in the frontend-provided order.
	pos := -1
	for i, uid := range orderedIDs {
		if uid == id {
			pos = i
			break
		}
	}
	if pos == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id not found in ordered_ids"})
		return
	}

	swapPos := pos - 1
	if body.Direction == "down" {
		swapPos = pos + 1
	}
	if swapPos < 0 || swapPos >= len(orderedIDs) {
		c.JSON(http.StatusOK, gin.H{"message": "already at boundary"})
		return
	}

	orderedIDs[pos], orderedIDs[swapPos] = orderedIDs[swapPos], orderedIDs[pos]

	// Reassign priorities as multiples of 10 in the new order.
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		for i, uid := range orderedIDs {
			tmp := -((i + 1) * 10)
			if err := tx.Model(&models.FirewallRule{}).Where("id = ?", uid).Update("priority", tmp).Error; err != nil {
				return err
			}
		}
		for i, uid := range orderedIDs {
			if err := tx.Model(&models.FirewallRule{}).Where("id = ?", uid).Update("priority", (i+1)*10).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	middleware.InvalidateFirewallCache()
	_ = pub.PublishReload(c.Request.Context())
	c.Status(http.StatusNoContent)
}
