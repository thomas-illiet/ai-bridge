package services

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/coder/aibridge"
	aibpkg "github.com/thomas-illiet/ai-bridge/aibridge"
	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var providerNameRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

type CreateProviderRequest struct {
	Name    string                `json:"name"`
	Type    models.ProviderType   `json:"type"`
	BaseURL string                `json:"base_url"`
	APIKey  string                `json:"api_key"`
	Config  models.ProviderConfig `json:"config"`
	Enabled bool                  `json:"enabled"`
}

type UpdateProviderRequest struct {
	Name    *string               `json:"name"`
	BaseURL *string               `json:"base_url"`
	APIKey  *string               `json:"api_key"`
	Config  models.ProviderConfig `json:"config"`
	Enabled *bool                 `json:"enabled"`
}

// BuildProviders loads all enabled providers from DB and converts them to aibridge.Provider.
func BuildProviders() ([]aibridge.Provider, error) {
	var dbProviders []models.AIProvider
	if err := database.DB.Where("enabled = true").Find(&dbProviders).Error; err != nil {
		return nil, fmt.Errorf("load providers: %w", err)
	}

	providers := make([]aibridge.Provider, 0, len(dbProviders))
	for i := range dbProviders {
		ap, err := ToAIBridgeProvider(&dbProviders[i])
		if err != nil {
			return nil, fmt.Errorf("convert provider %s: %w", dbProviders[i].Name, err)
		}
		providers = append(providers, ap)
	}
	return providers, nil
}

// ToAIBridgeProvider converts a DB provider row to the aibridge.Provider interface.
func ToAIBridgeProvider(p *models.AIProvider) (aibridge.Provider, error) {
	switch p.Type {
	case models.ProviderTypeOpenAI:
		return aibpkg.NewNamedOpenAIProvider(p.Name, p.APIKey, p.BaseURL), nil
	case models.ProviderTypeOllama:
		baseURL := strings.TrimRight(p.BaseURL, "/") + "/v1/"
		return aibpkg.NewNamedOllamaProvider(p.Name, baseURL, p.APIKey), nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", p.Type)
	}
}

func validateProviderName(name string) error {
	if !providerNameRegex.MatchString(name) {
		return fmt.Errorf("invalid name %q: must match ^[a-z0-9]+(-[a-z0-9]+)*$", name)
	}
	return nil
}

func CreateProvider(req CreateProviderRequest) (*models.AIProvider, error) {
	if err := validateProviderName(req.Name); err != nil {
		return nil, err
	}
	if req.Type != models.ProviderTypeOpenAI && req.Type != models.ProviderTypeOllama {
		return nil, fmt.Errorf("invalid type: must be 'openai' or 'ollama'")
	}
	if req.BaseURL == "" && req.Type == models.ProviderTypeOllama {
		return nil, fmt.Errorf("base_url is required for ollama providers")
	}
	if req.Config == nil {
		req.Config = models.ProviderConfig{}
	}

	p := &models.AIProvider{
		Name:    req.Name,
		Type:    req.Type,
		BaseURL: req.BaseURL,
		APIKey:  req.APIKey,
		Config:  req.Config,
		Enabled: req.Enabled,
	}
	if err := database.DB.Create(p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

func UpdateProvider(id uuid.UUID, req UpdateProviderRequest) (*models.AIProvider, error) {
	var p models.AIProvider
	if err := database.DB.First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		if err := validateProviderName(*req.Name); err != nil {
			return nil, err
		}
		updates["name"] = *req.Name
	}
	if req.BaseURL != nil {
		updates["base_url"] = *req.BaseURL
	}
	if req.APIKey != nil {
		updates["api_key"] = *req.APIKey
	}
	if req.Config != nil {
		updates["config"] = req.Config
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if err := database.DB.Model(&p).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func DeleteProvider(id uuid.UUID) error {
	result := database.DB.Unscoped().Delete(&models.AIProvider{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func ListProviders() ([]models.AIProvider, error) {
	var providers []models.AIProvider
	if err := database.DB.Order("created_at desc").Find(&providers).Error; err != nil {
		return nil, err
	}
	return providers, nil
}

func GetProvider(id uuid.UUID) (*models.AIProvider, error) {
	var p models.AIProvider
	if err := database.DB.First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}
