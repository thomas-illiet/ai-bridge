package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/handlers/common"
	"github.com/thomas-illiet/ai-bridge/models"
	"gorm.io/gorm"
)

type adminTokenRow struct {
	models.APIToken
	Username string `json:"username"`
}

var allowedAdminTokenSortColumns = map[string]string{
	"name":         "api_tokens.name",
	"username":     "users.username",
	"created_at":   "api_tokens.created_at",
	"expires_at":   "api_tokens.expires_at",
	"last_used_at": "api_tokens.last_used_at",
}

const adminTokenStatusExpr = "CASE WHEN api_tokens.revoked_at IS NOT NULL THEN 'revoked' WHEN api_tokens.expires_at IS NOT NULL AND api_tokens.expires_at < NOW() THEN 'expired' ELSE 'active' END"

// ListTokens returns a paginated list of all client tokens with their owner's username.
func ListTokens(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	includeInactive := c.Query("include_inactive") == "true"
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortDir := c.DefaultQuery("sort_dir", "desc")
	if sortDir != "asc" {
		sortDir = "desc"
	}
	var adminTokenOrderExpr string
	if sortBy == "status" {
		adminTokenOrderExpr = adminTokenStatusExpr + " " + sortDir
	} else {
		sortCol, ok := allowedAdminTokenSortColumns[sortBy]
		if !ok {
			sortCol = "api_tokens.created_at"
		}
		adminTokenOrderExpr = sortCol + " " + sortDir
	}

	q := database.DB.Model(&models.APIToken{}).
		Joins("LEFT JOIN users ON users.id = api_tokens.user_id").
		Where("api_tokens.deleted_at IS NULL")

	if common.CallerIsManager(c) {
		q = q.Where("users.role != ?", models.RoleService)
	}

	if !includeInactive {
		q = q.Where("api_tokens.revoked_at IS NULL AND (api_tokens.expires_at IS NULL OR api_tokens.expires_at > NOW())")
	}

	if search != "" {
		like := "%" + search + "%"
		q = q.Where("api_tokens.name ILIKE ? OR users.username ILIKE ?", like, like)
	}

	var total int64
	q.Count(&total)

	var rows []adminTokenRow
	q.Select("api_tokens.*, users.username").
		Order(gorm.Expr(adminTokenOrderExpr)).
		Limit(pageSize).Offset(offset).
		Scan(&rows)

	if rows == nil {
		rows = []adminTokenRow{}
	}

	c.JSON(http.StatusOK, gin.H{
		"tokens":   rows,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// RevokeToken revokes any client token by ID, with scope restrictions for managers.
func RevokeToken(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token id"})
		return
	}

	if common.CallerIsManager(c) {
		var token models.APIToken
		if err := database.DB.Where("id = ?", id).First(&token).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
			return
		}
		var owner models.User
		if err := database.DB.Where("id = ?", token.UserID).First(&owner).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "owner lookup failed"})
			return
		}
		if owner.Role == models.RoleService {
			c.JSON(http.StatusForbidden, gin.H{"error": "managers cannot revoke service account tokens"})
			return
		}
	}

	result := database.DB.Model(&models.APIToken{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", gorm.Expr("NOW()"))

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "token not found or already revoked"})
		return
	}
	c.Status(http.StatusNoContent)
}

// UnrevokeToken restores a revoked (non-expired) token by clearing its revoked_at timestamp.
func UnrevokeToken(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token id"})
		return
	}

	if common.CallerIsManager(c) {
		var token models.APIToken
		if err := database.DB.Where("id = ?", id).First(&token).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
			return
		}
		var owner models.User
		if err := database.DB.Where("id = ?", token.UserID).First(&owner).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "owner lookup failed"})
			return
		}
		if owner.Role == models.RoleService {
			c.JSON(http.StatusForbidden, gin.H{"error": "managers cannot unrevoke service account tokens"})
			return
		}
	}

	result := database.DB.Model(&models.APIToken{}).
		Where("id = ? AND revoked_at IS NOT NULL AND (expires_at IS NULL OR expires_at > NOW())", id).
		Update("revoked_at", nil)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "token not found, not revoked, or already expired"})
		return
	}
	c.Status(http.StatusNoContent)
}
