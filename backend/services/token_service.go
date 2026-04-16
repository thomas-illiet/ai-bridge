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

func CreateToken(userID, name, secret string) (*models.ClientToken, string, error) {
	id := uuid.New()

	claims := PATClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:  userID,
			ID:       id.String(),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return nil, "", fmt.Errorf("sign token: %w", err)
	}

	record := &models.ClientToken{
		ID:        id,
		UserID:    userID,
		Name:      name,
		TokenHash: HashToken(raw),
	}

	if err := database.DB.Create(record).Error; err != nil {
		return nil, "", fmt.Errorf("store token: %w", err)
	}

	return record, raw, nil
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

	now := time.Now()
	database.DB.Model(&record).Update("last_used_at", now)

	return &record, nil
}

func ListUserTokens(userID string) ([]models.ClientToken, error) {
	var tokens []models.ClientToken
	if err := database.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&tokens).Error; err != nil {
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
