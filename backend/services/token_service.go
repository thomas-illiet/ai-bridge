package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/models"
	"gorm.io/gorm"
)

type PATClaims struct {
	jwt.RegisteredClaims
}

var ErrTokenNameTaken = errors.New("a token with this name already exists")

func IsTokenNameTaken(userID, name string, excludeID *uuid.UUID) (bool, error) {
	q := database.DB.Model(&models.ClientToken{}).
		Where("user_id = ? AND name = ? AND revoked_at IS NULL", userID, name)
	if excludeID != nil {
		q = q.Where("id != ?", *excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func CreateToken(userID, name, description, secret string, durationDays int) (*models.ClientToken, string, error) {
	taken, err := IsTokenNameTaken(userID, name, nil)
	if err != nil {
		return nil, "", fmt.Errorf("check name: %w", err)
	}
	if taken {
		return nil, "", ErrTokenNameTaken
	}

	id := uuid.New()
	expiresAt := time.Now().UTC().AddDate(0, 0, durationDays)

	claims := PATClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ID:        id.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return nil, "", fmt.Errorf("sign token: %w", err)
	}

	record := &models.ClientToken{
		ID:          id,
		UserID:      userID,
		Name:        name,
		Description: description,
		TokenHash:   HashToken(raw),
		ExpiresAt:   &expiresAt,
	}

	if err := database.DB.Create(record).Error; err != nil {
		return nil, "", fmt.Errorf("store token: %w", err)
	}

	return record, raw, nil
}

func UpdateToken(id uuid.UUID, userID, name, description string) (*models.ClientToken, error) {
	var record models.ClientToken
	if err := database.DB.Where("id = ? AND user_id = ? AND revoked_at IS NULL", id, userID).First(&record).Error; err != nil {
		return nil, err
	}

	taken, err := IsTokenNameTaken(userID, name, &id)
	if err != nil {
		return nil, fmt.Errorf("check name: %w", err)
	}
	if taken {
		return nil, ErrTokenNameTaken
	}

	if err := database.DB.Model(&record).Updates(map[string]interface{}{
		"name":        name,
		"description": description,
	}).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func LookupAndVerify(jti, raw, secret string) (*models.ClientToken, error) {
	id, err := uuid.Parse(jti)
	if err != nil {
		return nil, fmt.Errorf("invalid jti: %w", err)
	}

	var record models.ClientToken
	if err := database.DB.Where("id = ?", id).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("db lookup: %w", err)
	}

	if HashToken(raw) != record.TokenHash {
		return nil, fmt.Errorf("invalid token")
	}

	_, err = jwt.ParseWithClaims(raw, &PATClaims{}, func(_ *jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, fmt.Errorf("invalid token signature")
	}

	if record.IsRevoked() {
		return nil, fmt.Errorf("token revoked")
	}

	if record.IsExpired() {
		return nil, fmt.Errorf("token expired")
	}

	now := time.Now()
	database.DB.Model(&record).Update("last_used_at", now)

	return &record, nil
}

var allowedTokenSortColumns = map[string]string{
	"name":         "name",
	"created_at":   "created_at",
	"expires_at":   "expires_at",
	"last_used_at": "last_used_at",
}

const tokenStatusExpr = "CASE WHEN revoked_at IS NOT NULL THEN 'revoked' WHEN expires_at IS NOT NULL AND expires_at < NOW() THEN 'expired' ELSE 'active' END"

func ListUserTokens(userID string, includeRevoked bool, sortBy, sortDir string) ([]models.ClientToken, error) {
	if sortDir != "asc" {
		sortDir = "desc"
	}

	var orderExpr string
	if sortBy == "status" {
		orderExpr = tokenStatusExpr + " " + sortDir
	} else {
		col, ok := allowedTokenSortColumns[sortBy]
		if !ok {
			col = "created_at"
		}
		orderExpr = col + " " + sortDir
	}

	var tokens []models.ClientToken
	q := database.DB.Where("user_id = ?", userID)
	if !includeRevoked {
		q = q.Where("revoked_at IS NULL")
	}
	if err := q.Order(gorm.Expr(orderExpr)).Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

func RevokeToken(id uuid.UUID, userID string) error {
	result := database.DB.Model(&models.ClientToken{}).
		Where("id = ? AND user_id = ? AND revoked_at IS NULL", id, userID).
		Update("revoked_at", time.Now())

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
