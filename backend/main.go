package main

import (
	"context"
	"log"
	"os"
	"strings"

	"cdr.dev/slog/v3"
	"cdr.dev/slog/v3/sloggers/sloghuman"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"

	"github.com/coder/aibridge"
	aibpkg "github.com/thomas-illiet/ai-bridge/aibridge"
	"github.com/thomas-illiet/ai-bridge/config"
	"github.com/thomas-illiet/ai-bridge/database"
	"github.com/thomas-illiet/ai-bridge/handlers"
	"github.com/thomas-illiet/ai-bridge/middleware"
	"github.com/thomas-illiet/ai-bridge/models"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	logger := slog.Make(sloghuman.Sink(os.Stderr)).Leveled(slog.LevelInfo)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Split(cfg.AllowedOrigins, ","),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	if err := database.Connect(cfg.DatabaseDSN); err != nil {
		log.Fatalf("database error: %v", err)
	}
	if err := database.DB.AutoMigrate(
		&models.ClientToken{},
		&models.AibridgeInterception{},
		&models.AibridgeTokenUsage{},
		&models.AibridgeUserPrompt{},
		&models.AibridgeToolUsage{},
		&models.AibridgeModelThought{},
	); err != nil {
		log.Fatalf("auto-migrate error: %v", err)
	}

	// Build aibridge providers from configured API keys.
	var providers []aibridge.Provider
	if cfg.AnthropicAPIKey != "" {
		providers = append(providers, aibridge.NewAnthropicProvider(
			aibridge.AnthropicConfig{Key: cfg.AnthropicAPIKey}, nil,
		))
	}
	if cfg.OpenAIAPIKey != "" {
		providers = append(providers, aibridge.NewOpenAIProvider(
			aibridge.OpenAIConfig{Key: cfg.OpenAIAPIKey},
		))
	}
	if cfg.OllamaBaseURL != "" {
		providers = append(providers, aibridge.NewOpenAIProvider(aibridge.OpenAIConfig{
			Name:    "ollama",
			BaseURL: strings.TrimRight(cfg.OllamaBaseURL, "/") + "/v1/",
			Key:     "ollama",
		}))
	}

	reg := prometheus.NewRegistry()
	metrics := aibridge.NewMetrics(reg)
	tracer := otel.GetTracerProvider().Tracer("ai-bridge")
	recorder := aibpkg.NewGORMRecorder()

	r.GET("/health", handlers.HealthCheck)
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))

	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth(cfg))
	{
		api.GET("/me", handlers.GetMe)
		api.GET("/dashboard", handlers.GetDashboard)
		api.POST("/tokens", handlers.CreateToken(cfg.TokenSecret))
		api.GET("/tokens", handlers.ListTokens)
		api.DELETE("/tokens/:id", handlers.RevokeToken)
	}

	if len(providers) > 0 {
		bridge, err := aibridge.NewRequestBridge(ctx, providers, recorder, nil, logger, metrics, tracer)
		if err != nil {
			log.Fatalf("aibridge error: %v", err)
		}
		defer bridge.Shutdown(ctx)

		aib := r.Group("")
		aib.Use(middleware.JWTAuth(cfg))
		aib.Use(middleware.AIBridgeActor())
		aib.Any("/anthropic/*path", gin.WrapH(bridge))
		aib.Any("/openai/*path", gin.WrapH(bridge))
		aib.Any("/ollama/*path", gin.WrapH(bridge))
	} else {
		logger.Warn(ctx, "no AI providers configured — set ANTHROPIC_API_KEY or OPENAI_API_KEY to enable the proxy")
	}

	log.Printf("server listening on :%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
