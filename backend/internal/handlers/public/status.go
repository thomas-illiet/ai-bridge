package public

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/internal/config"
	"github.com/thomas-illiet/ai-bridge/internal/database"
	"github.com/thomas-illiet/ai-bridge/internal/models"
)

type serviceStatus struct {
	Name       string `json:"name"`
	Status     string `json:"status"` // "up", "down", "disabled"
	Message    string `json:"message,omitempty"`
	LatencyMs  *int64 `json:"latency_ms,omitempty"`
	ModelCount *int   `json:"model_count,omitempty"`
}

type statusResponse struct {
	Status   string          `json:"status"` // "healthy", "degraded"
	Services []serviceStatus `json:"services"`
}

// GetStatus returns the overall health of the service and the status of each dependency.
func GetStatus(cfg *config.Config) gin.HandlerFunc {
	client := &http.Client{Timeout: 3 * time.Second}

	return func(c *gin.Context) {
		ctx := c.Request.Context()
		services := []serviceStatus{
			checkDatabase(ctx),
			checkOIDC(client, cfg),
		}

		var providers []models.Provider
		database.DB.Where("enabled = true").Find(&providers)
		for i := range providers {
			services = append(services, checkProvider(client, &providers[i]))
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

func ms(d time.Duration) *int64 { v := d.Milliseconds(); return &v }

func checkDatabase(ctx context.Context) serviceStatus {
	db, err := database.DB.DB()
	if err != nil {
		return serviceStatus{Name: "database", Status: "down", Message: err.Error()}
	}
	start := time.Now()
	if err := db.PingContext(ctx); err != nil {
		return serviceStatus{Name: "database", Status: "down", Message: err.Error()}
	}
	return serviceStatus{Name: "database", Status: "up", LatencyMs: ms(time.Since(start))}
}

func checkOIDC(client *http.Client, cfg *config.Config) serviceStatus {
	url := cfg.OIDCHealthURL() + "/.well-known/openid-configuration"
	start := time.Now()
	resp, err := client.Get(url)
	if err != nil {
		return serviceStatus{Name: "oidc", Status: "down", Message: err.Error()}
	}
	resp.Body.Close()
	latency := ms(time.Since(start))
	if resp.StatusCode != http.StatusOK {
		return serviceStatus{Name: "oidc", Status: "down", Message: fmt.Sprintf("HTTP %d", resp.StatusCode), LatencyMs: latency}
	}
	return serviceStatus{Name: "oidc", Status: "up", LatencyMs: latency}
}

func checkProvider(client *http.Client, p *models.Provider) serviceStatus {
	switch p.Type {
	case models.ProviderTypeOpenAI:
		return checkOpenAIProvider(client, p)
	case models.ProviderTypeOllama:
		return checkOllamaProvider(client, p)
	default:
		return serviceStatus{Name: p.Name, Status: "disabled", Message: "unknown type"}
	}
}

func checkOpenAIProvider(client *http.Client, p *models.Provider) serviceStatus {
	baseURL := p.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1/"
	}
	url := strings.TrimRight(baseURL, "/") + "/models"
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return serviceStatus{Name: p.Name, Status: "down", Message: err.Error()}
	}
	defer resp.Body.Close()
	latency := ms(time.Since(start))
	if resp.StatusCode != http.StatusOK {
		return serviceStatus{Name: p.Name, Status: "down", Message: fmt.Sprintf("HTTP %d", resp.StatusCode), LatencyMs: latency}
	}
	var body struct {
		Data []json.RawMessage `json:"data"`
	}
	count := 0
	if json.NewDecoder(resp.Body).Decode(&body) == nil {
		count = len(body.Data)
	}
	return serviceStatus{Name: p.Name, Status: "up", LatencyMs: latency, ModelCount: &count}
}

func checkOllamaProvider(client *http.Client, p *models.Provider) serviceStatus {
	url := strings.TrimRight(p.BaseURL, "/") + "/api/tags"
	start := time.Now()
	resp, err := client.Get(url)
	if err != nil {
		return serviceStatus{Name: p.Name, Status: "down", Message: err.Error()}
	}
	defer resp.Body.Close()
	latency := ms(time.Since(start))
	if resp.StatusCode != http.StatusOK {
		return serviceStatus{Name: p.Name, Status: "down", Message: fmt.Sprintf("HTTP %d", resp.StatusCode), LatencyMs: latency}
	}
	var body struct {
		Models []json.RawMessage `json:"models"`
	}
	count := 0
	if json.NewDecoder(resp.Body).Decode(&body) == nil {
		count = len(body.Models)
	}
	return serviceStatus{Name: p.Name, Status: "up", LatencyMs: latency, ModelCount: &count}
}
