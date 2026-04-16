package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort          string
	KeycloakBaseURL     string
	KeycloakIssuerURL   string
	KeycloakRealm       string
	KeycloakClientID    string
	AllowedOrigins      string
	DatabaseDSN         string
	TokenSecret         string
	AnthropicAPIKey     string
	OpenAIAPIKey        string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	keycloakBaseURL := getEnv("KEYCLOAK_BASE_URL", "http://localhost:8180")
	keycloakRealm := getEnv("KEYCLOAK_REALM", "ai-bridge")

	cfg := &Config{
		ServerPort:        getEnv("SERVER_PORT", "8585"),
		KeycloakBaseURL:   keycloakBaseURL,
		KeycloakIssuerURL: getEnv("KEYCLOAK_ISSUER_URL", keycloakBaseURL),
		KeycloakRealm:     keycloakRealm,
		KeycloakClientID:  getEnv("KEYCLOAK_CLIENT_ID", "ai-bridge-frontend"),
		AllowedOrigins:    getEnv("ALLOWED_ORIGINS", "http://localhost:5173"),
		DatabaseDSN:       getEnv("DATABASE_DSN", ""),
		TokenSecret:       getEnv("TOKEN_SECRET", ""),
		AnthropicAPIKey:   getEnv("ANTHROPIC_API_KEY", ""),
		OpenAIAPIKey:      getEnv("OPENAI_API_KEY", ""),
	}

	if cfg.KeycloakBaseURL == "" || cfg.KeycloakRealm == "" {
		return nil, fmt.Errorf("KEYCLOAK_BASE_URL and KEYCLOAK_REALM are required")
	}
	if cfg.DatabaseDSN == "" {
		return nil, fmt.Errorf("DATABASE_DSN is required")
	}
	if cfg.TokenSecret == "" {
		return nil, fmt.Errorf("TOKEN_SECRET is required")
	}

	return cfg, nil
}

func (c *Config) JWKSUrl() string {
	return fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", c.KeycloakBaseURL, c.KeycloakRealm)
}

func (c *Config) IssuerURL() string {
	return fmt.Sprintf("%s/realms/%s", c.KeycloakIssuerURL, c.KeycloakRealm)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
