package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/internal/database"
	"github.com/thomas-illiet/ai-bridge/internal/models"
	"gorm.io/gorm"
)

// GetModels returns the list of available models for the requested provider name.
func GetModels() gin.HandlerFunc {
	return func(c *gin.Context) {
		providerName := c.Query("provider")
		if providerName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provider query parameter is required"})
			return
		}

		var p models.Provider
		if err := database.DB.Where("name = ? AND enabled = true", providerName).First(&p).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "provider not found or disabled"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		switch p.Type {
		case models.ProviderTypeOpenAI:
			modelList, err := fetchOpenAIModels(p.APIKey, p.BaseURL)
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"models": modelList})
		case models.ProviderTypeOllama:
			modelList, err := fetchOllamaModels(p.BaseURL)
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"models": modelList})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported provider type"})
		}
	}
}

// fetchOpenAIModels queries an OpenAI-compatible API and returns the list of model IDs.
func fetchOpenAIModels(apiKey, baseURL string) ([]string, error) {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1/"
	}
	url := strings.TrimRight(baseURL, "/") + "/models"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai returned %d", resp.StatusCode)
	}

	var body struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	ids := make([]string, 0, len(body.Data))
	for _, m := range body.Data {
		ids = append(ids, m.ID)
	}
	return ids, nil
}

// fetchOllamaModels queries the Ollama API and returns the list of model names.
func fetchOllamaModels(baseURL string) ([]string, error) {
	url := strings.TrimRight(baseURL, "/") + "/api/tags"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned %d", resp.StatusCode)
	}

	var body struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	names := make([]string, 0, len(body.Models))
	for _, m := range body.Models {
		names = append(names, m.Name)
	}
	return names, nil
}
