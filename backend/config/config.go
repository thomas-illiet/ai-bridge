package config

import (
	"fmt"
	"os"
	"strconv"
	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort     string
	OIDCIssuerURL     string
	OIDCInternalURL   string // internal URL for health checks (defaults to OIDCIssuerURL)
	OIDCClientID      string
	OIDCJWKSUrl       string
	AllowedOrigins string
	DatabaseDSN    string
	TokenSecret    string
	TrustedProxies string
	SMTPHost              string
	SMTPPort              int
	SMTPUser              string
	SMTPPassword          string
	SMTPFrom              string
	SMTPTo                string // comma-separated admin emails
	RoleExpiryIntervalSec int    // how often to check for expired roles (seconds)
	AppURL                string // public base URL used in email links
	RedisURL              string // e.g. "redis://redis:6379/0"
	BridgeServerPort      string // default "8586", bridge service only
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

	cfg := &Config{
		ServerPort:     getEnv("SERVER_PORT", "8585"),
		OIDCIssuerURL:   getEnv("OIDC_ISSUER_URL", "http://localhost:8180/realms/ai-bridge"),
		OIDCInternalURL: getEnv("OIDC_INTERNAL_URL", ""),
		OIDCClientID:    getEnv("OIDC_CLIENT_ID", "ai-bridge-frontend"),
		OIDCJWKSUrl:     getEnv("OIDC_JWKS_URL", ""),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:5173"),
		DatabaseDSN:    getEnv("DATABASE_DSN", ""),
		TokenSecret:    getEnv("TOKEN_SECRET", ""),
		TrustedProxies: getEnv("TRUSTED_PROXIES", "127.0.0.0/8,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"),
		SMTPHost:          getEnv("SMTP_HOST", ""),
		SMTPPort:          getEnvInt("SMTP_PORT", 587),
		SMTPUser:          getEnv("SMTP_USER", ""),
		SMTPPassword:      getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:          getEnv("SMTP_FROM", ""),
		SMTPTo:                getEnv("SMTP_TO", ""),
		RoleExpiryIntervalSec: getEnvInt("ROLE_EXPIRY_INTERVAL_SEC", 60),
		AppURL:                getEnv("APP_URL", "http://localhost:5173"),
		RedisURL:              getEnv("REDIS_URL", "redis://localhost:6379/0"),
		BridgeServerPort:      getEnv("BRIDGE_SERVER_PORT", "8586"),
	}

	if cfg.OIDCIssuerURL == "" {
		return nil, fmt.Errorf("OIDC_ISSUER_URL is required")
	}
	if cfg.DatabaseDSN == "" {
		return nil, fmt.Errorf("DATABASE_DSN is required")
	}
	if cfg.TokenSecret == "" {
		return nil, fmt.Errorf("TOKEN_SECRET is required")
	}

	return cfg, nil
}

// JWKSUrl returns the explicit JWKS URL override if set, or empty string to trigger OIDC discovery.
func (c *Config) JWKSUrl() string {
	return c.OIDCJWKSUrl
}

// OIDCHealthURL returns the URL used for OIDC health checks.
// Uses OIDCInternalURL when set (for container-internal routing), otherwise falls back to OIDCIssuerURL.
func (c *Config) OIDCHealthURL() string {
	if c.OIDCInternalURL != "" {
		return c.OIDCInternalURL
	}
	return c.OIDCIssuerURL
}
