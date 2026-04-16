package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/config"
	"github.com/thomas-illiet/ai-bridge/database"
)

type serviceStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "up", "down", "disabled"
	Message string `json:"message,omitempty"`
}

type statusResponse struct {
	Status   string          `json:"status"` // "healthy", "degraded"
	Services []serviceStatus `json:"services"`
}

func GetStatus(cfg *config.Config) gin.HandlerFunc {
	client := &http.Client{Timeout: 3 * time.Second}

	return func(c *gin.Context) {
		ctx := c.Request.Context()
		services := []serviceStatus{
			checkDatabase(ctx),
			checkKeycloak(client, cfg),
			checkAnthropic(client, cfg),
			checkOpenAI(client, cfg),
			checkOllama(client, cfg),
		}

		overall := "healthy"
		for _, s := range services {
			if s.Status == "down" {
				overall = "degraded"
				break
			}
		}

		c.JSON(http.StatusOK, statusResponse{Status: overall, Services: services})
	}
}

func checkDatabase(ctx context.Context) serviceStatus {
	db, err := database.DB.DB()
	if err != nil {
		return serviceStatus{Name: "database", Status: "down", Message: err.Error()}
	}
	if err := db.PingContext(ctx); err != nil {
		return serviceStatus{Name: "database", Status: "down", Message: err.Error()}
	}
	return serviceStatus{Name: "database", Status: "up"}
}

func checkKeycloak(client *http.Client, cfg *config.Config) serviceStatus {
	url := fmt.Sprintf("%s/health/ready", cfg.KeycloakBaseURL)
	resp, err := client.Get(url)
	if err != nil {
		return serviceStatus{Name: "keycloak", Status: "down", Message: err.Error()}
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return serviceStatus{Name: "keycloak", Status: "down", Message: fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}
	return serviceStatus{Name: "keycloak", Status: "up"}
}

func checkAnthropic(client *http.Client, cfg *config.Config) serviceStatus {
	if cfg.AnthropicAPIKey == "" {
		return serviceStatus{Name: "anthropic", Status: "disabled"}
	}
	req, _ := http.NewRequest(http.MethodGet, "https://api.anthropic.com/v1/models", nil)
	req.Header.Set("x-api-key", cfg.AnthropicAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	resp, err := client.Do(req)
	if err != nil {
		return serviceStatus{Name: "anthropic", Status: "down", Message: err.Error()}
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return serviceStatus{Name: "anthropic", Status: "down", Message: fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}
	return serviceStatus{Name: "anthropic", Status: "up"}
}

func checkOpenAI(client *http.Client, cfg *config.Config) serviceStatus {
	if cfg.OpenAIAPIKey == "" {
		return serviceStatus{Name: "openai", Status: "disabled"}
	}
	req, _ := http.NewRequest(http.MethodGet, "https://api.openai.com/v1/models", nil)
	req.Header.Set("Authorization", "Bearer "+cfg.OpenAIAPIKey)
	resp, err := client.Do(req)
	if err != nil {
		return serviceStatus{Name: "openai", Status: "down", Message: err.Error()}
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return serviceStatus{Name: "openai", Status: "down", Message: fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}
	return serviceStatus{Name: "openai", Status: "up"}
}

func checkOllama(client *http.Client, cfg *config.Config) serviceStatus {
	if cfg.OllamaBaseURL == "" {
		return serviceStatus{Name: "ollama", Status: "disabled"}
	}
	url := fmt.Sprintf("%s/api/version", cfg.OllamaBaseURL)
	resp, err := client.Get(url)
	if err != nil {
		return serviceStatus{Name: "ollama", Status: "down", Message: err.Error()}
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return serviceStatus{Name: "ollama", Status: "down", Message: fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}
	return serviceStatus{Name: "ollama", Status: "up"}
}
