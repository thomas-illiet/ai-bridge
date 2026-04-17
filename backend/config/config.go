package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort        string
	KeycloakBaseURL   string
	KeycloakIssuerURL string
	KeycloakRealm     string
	KeycloakClientID  string
	AllowedOrigins    string
	DatabaseDSN       string
	TokenSecret       string
	OpenAIAPIKey      string
	OllamaBaseURL     string
	OllamaNumCtx      int
	TrustedProxies    string
	SMTPHost              string
	SMTPPort              int
	SMTPUser              string
	SMTPPassword          string
	SMTPFrom              string
	SMTPTo                string // comma-separated admin emails
	RoleExpiryIntervalSec int    // how often to check for expired roles (seconds)
	AppURL                string // public base URL used in email links
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			return val
		}
	}
	return fallback
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
		OpenAIAPIKey:      getEnv("OPENAI_API_KEY", ""),
		OllamaBaseURL:     getEnv("OLLAMA_BASE_URL", ""),
		OllamaNumCtx:      getEnvInt("OLLAMA_NUM_CTX", 4096),
		TrustedProxies:    getEnv("TRUSTED_PROXIES", "127.0.0.0/8,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"),
		SMTPHost:          getEnv("SMTP_HOST", ""),
		SMTPPort:          getEnvInt("SMTP_PORT", 587),
		SMTPUser:          getEnv("SMTP_USER", ""),
		SMTPPassword:      getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:          getEnv("SMTP_FROM", ""),
		SMTPTo:                getEnv("SMTP_TO", ""),
		RoleExpiryIntervalSec: getEnvInt("ROLE_EXPIRY_INTERVAL_SEC", 60),
		AppURL:                getEnv("APP_URL", "http://localhost:5173"),
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

func (c *Config) JWTSUrl() string {
	return fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", c.KeycloakBaseURL, c.KeycloakRealm)
}

func (c *Config) IssuerURL() string {
	return fmt.Sprintf("%s/realms/%s", c.KeycloakIssuerURL, c.KeycloakRealm)
}
