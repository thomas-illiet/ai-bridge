package public

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thomas-illiet/ai-bridge/config"
	"github.com/thomas-illiet/ai-bridge/database"
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

func ms(d time.Duration) *int64 { v := d.Milliseconds(); return &v }

// checkDatabase pings the database and returns its service status.
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

// checkOIDC probes the OIDC discovery endpoint and returns its service status.
// Uses OIDCInternalURL when set so container-internal routing works correctly.
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

// checkOpenAI probes the OpenAI models endpoint, returns count of available models.
func checkOpenAI(client *http.Client, cfg *config.Config) serviceStatus {
	if cfg.OpenAIAPIKey == "" {
		return serviceStatus{Name: "openai", Status: "disabled"}
	}
	req, _ := http.NewRequest(http.MethodGet, "https://api.openai.com/v1/models", nil)
	req.Header.Set("Authorization", "Bearer "+cfg.OpenAIAPIKey)
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return serviceStatus{Name: "openai", Status: "down", Message: err.Error()}
	}
	defer resp.Body.Close()
	latency := ms(time.Since(start))
	if resp.StatusCode != http.StatusOK {
		return serviceStatus{Name: "openai", Status: "down", Message: fmt.Sprintf("HTTP %d", resp.StatusCode), LatencyMs: latency}
	}
	var body struct {
		Data []json.RawMessage `json:"data"`
	}
	count := 0
	if json.NewDecoder(resp.Body).Decode(&body) == nil {
		count = len(body.Data)
	}
	return serviceStatus{Name: "openai", Status: "up", LatencyMs: latency, ModelCount: &count}
}

// checkOllama probes the Ollama tags endpoint and returns count of local models.
func checkOllama(client *http.Client, cfg *config.Config) serviceStatus {
	if cfg.OllamaBaseURL == "" {
		return serviceStatus{Name: "ollama", Status: "disabled"}
	}
	url := fmt.Sprintf("%s/api/tags", cfg.OllamaBaseURL)
	start := time.Now()
	resp, err := client.Get(url)
	if err != nil {
		return serviceStatus{Name: "ollama", Status: "down", Message: err.Error()}
	}
	defer resp.Body.Close()
	latency := ms(time.Since(start))
	if resp.StatusCode != http.StatusOK {
		return serviceStatus{Name: "ollama", Status: "down", Message: fmt.Sprintf("HTTP %d", resp.StatusCode), LatencyMs: latency}
	}
	var body struct {
		Models []json.RawMessage `json:"models"`
	}
	count := 0
	if json.NewDecoder(resp.Body).Decode(&body) == nil {
		count = len(body.Models)
	}
	return serviceStatus{Name: "ollama", Status: "up", LatencyMs: latency, ModelCount: &count}
}
