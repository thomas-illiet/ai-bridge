package services

import (
	"errors"

	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/models"
	"gorm.io/gorm"
)

// GetOrCreateUser auto-registers the user on first login.
// The very first user ever registered receives the "admin" role.
// Subsequent new users receive the "none" role.
// Existing users have their username/email refreshed but role unchanged.
func GetOrCreateUser(id, username, email string) (*models.RegisteredUser, error) {
	var user models.RegisteredUser
	err := database.DB.Where("id = ?", id).First(&user).Error

	if err == nil {
		// Existing user — refresh profile fields, keep role.
		database.DB.Model(&user).Updates(map[string]any{
			"username": username,
			"email":    email,
		})
		return &user, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// New user — determine role.
	role := models.RoleNone
	var count int64
	if database.DB.Model(&models.RegisteredUser{}).Count(&count); count == 0 {
		role = models.RoleAdmin
	}

	user = models.RegisteredUser{
		ID:       id,
		Username: username,
		Email:    email,
		Role:     role,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID returns the registered user or nil if not found.
func GetUserByID(id string) (*models.RegisteredUser, error) {
	var user models.RegisteredUser
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
