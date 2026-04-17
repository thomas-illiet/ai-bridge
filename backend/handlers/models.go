package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/config"
)

// GetModels returns available model IDs for the given provider.
func GetModels(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		provider := c.Query("provider")
		switch provider {
		case "openai":
			if cfg.OpenAIAPIKey == "" {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "OpenAI not configured"})
				return
			}
			models, err := fetchOpenAIModels(cfg.OpenAIAPIKey)
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"models": models})
		case "ollama":
			if cfg.OllamaBaseURL == "" {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Ollama not configured"})
				return
			}
			models, err := fetchOllamaModels(cfg.OllamaBaseURL)
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"models": models})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "provider must be 'openai' or 'ollama'"})
		}
	}
}

func fetchOpenAIModels(apiKey string) ([]string, error) {
	req, _ := http.NewRequest("GET", "https://api.openai.com/v1/models", nil)
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
