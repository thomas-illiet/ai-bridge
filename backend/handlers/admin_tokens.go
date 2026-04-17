package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/models"
	"gorm.io/gorm"
)

// adminTokenRow joins ClientToken with the owner's username for display.
type adminTokenRow struct {
	models.ClientToken
	Username string `json:"username"`
}

// AdminListTokens returns all tokens across all users with pagination and search.
func AdminListTokens(c *gin.Context) {
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

	q := database.DB.Model(&models.ClientToken{}).
		Joins("LEFT JOIN registered_users ON registered_users.id = client_tokens.user_id").
		Where("client_tokens.deleted_at IS NULL")

	if search != "" {
		like := "%" + search + "%"
		q = q.Where("client_tokens.name ILIKE ? OR registered_users.username ILIKE ?", like, like)
	}

	var total int64
	q.Count(&total)

	var rows []adminTokenRow
	q.Select("client_tokens.*, registered_users.username").
		Order("client_tokens.created_at DESC").
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

// AdminRevokeToken revokes any token by ID regardless of owner.
func AdminRevokeToken(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token id"})
		return
	}

	result := database.DB.Model(&models.ClientToken{}).
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
