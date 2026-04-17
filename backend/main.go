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
		&models.IPWhitelistEntry{},
		&models.RegisteredUser{},
	); err != nil {
		log.Fatalf("auto-migrate error: %v", err)
	}

	// Build aibridge providers from configured API keys.
	var providers []aibridge.Provider
	if cfg.OpenAIAPIKey != "" {
		providers = append(providers, aibridge.NewOpenAIProvider(
			aibridge.OpenAIConfig{Key: cfg.OpenAIAPIKey},
		))
	}
	if cfg.OllamaBaseURL != "" {
		providers = append(providers, aibpkg.NewOllamaProvider(
			strings.TrimRight(cfg.OllamaBaseURL, "/")+"/v1/", "",
		))
	}

	reg := prometheus.NewRegistry()
	metrics := aibridge.NewMetrics(reg)
	tracer := otel.GetTracerProvider().Tracer("ai-bridge")
	recorder := aibpkg.NewGORMRecorder()

	r.GET("/health", handlers.HealthCheck)
	r.GET("/api/status", handlers.GetStatus(cfg))
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))

	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth(cfg))
	{
		api.GET("/me", handlers.GetMe)
		api.GET("/dashboard", handlers.GetDashboard)
		api.GET("/models", handlers.GetModels(cfg))

		user := api.Group("")
		user.Use(middleware.RequireAnyRole(middleware.RoleUser, middleware.RoleAdmin))
		user.POST("/tokens", handlers.CreateToken(cfg.TokenSecret))
		user.GET("/tokens", handlers.ListTokens)
		user.DELETE("/tokens/:id", handlers.RevokeToken)
		user.GET("/history", handlers.GetHistory)
		user.GET("/history/:id", handlers.GetHistoryDetail)

		admin := api.Group("/admin")
		admin.Use(middleware.RequireRole(middleware.RoleAdmin))
		admin.GET("/whitelist", handlers.ListWhitelist)
		admin.POST("/whitelist", handlers.AddWhitelist)
		admin.DELETE("/whitelist/:id", handlers.DeleteWhitelist)
		admin.PATCH("/whitelist/:id", handlers.ToggleWhitelist)
		admin.GET("/users", handlers.ListUsers)
		admin.PATCH("/users/:id", handlers.UpdateUserRole)
		admin.DELETE("/users/:id", handlers.DeleteUser)
		admin.GET("/users/:id/stats", handlers.GetUserStats)
		admin.GET("/tokens", handlers.AdminListTokens)
		admin.DELETE("/tokens/:id", handlers.AdminRevokeToken)
		admin.GET("/history", handlers.AdminGetHistory)
		admin.GET("/history/:id", handlers.AdminGetHistoryDetail)
	}

	if len(providers) > 0 {
		bridge, err := aibridge.NewRequestBridge(ctx, providers, recorder, nil, logger, metrics, tracer)
		if err != nil {
			log.Fatalf("aibridge error: %v", err)
		}
		defer func(bridge *aibridge.RequestBridge, ctx context.Context) {
			err := bridge.Shutdown(ctx)
			if err != nil {
				log.Fatalf("aibridge error: %v", err)
			}
		}(bridge, ctx)

		aib := r.Group("")
		aib.Use(middleware.JWTAuth(cfg))
		aib.Use(middleware.RequireAnyRole(middleware.RoleUser, middleware.RoleAdmin))
		aib.Use(middleware.IPWhitelist(cfg.TrustedProxies))
		aib.Use(middleware.AIBridgeActor())
		aib.Any("/openai/*path", gin.WrapH(bridge))
		aib.Any("/ollama/*path", gin.WrapH(bridge))
	} else {
		logger.Warn(ctx, "no AI providers configured — set OPENAI_API_KEY or OLLAMA_BASE_URL to enable the proxy")
	}

	log.Printf("server listening on :%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
