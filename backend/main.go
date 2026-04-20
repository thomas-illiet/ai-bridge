package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

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
	handlerAdmin "github.com/thomas-illiet/ai-bridge/handlers/admin"
	handlerPublic "github.com/thomas-illiet/ai-bridge/handlers/public"
	handlerUser "github.com/thomas-illiet/ai-bridge/handlers/user"
	"github.com/thomas-illiet/ai-bridge/middleware"
	"github.com/thomas-illiet/ai-bridge/models"
	"github.com/thomas-illiet/ai-bridge/services"
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
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Cache-Control"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "Transfer-Encoding", "X-Content-Type-Options"},
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
		&models.AccessRequest{},
		&models.AIProvider{},
	); err != nil {
		log.Fatalf("auto-migrate error: %v", err)
	}

	reg := prometheus.NewRegistry()
	metrics := aibridge.NewMetrics(reg)
	tracer := otel.GetTracerProvider().Tracer("ai-bridge")
	recorder := aibpkg.NewGORMRecorder()

	bridgeManager := aibpkg.NewBridgeManager(ctx, recorder, logger, metrics, tracer)

	// Load providers already in DB and initialize the bridge.
	if initialProviders, err := services.BuildProviders(); err != nil {
		log.Printf("warning: could not load initial providers: %v", err)
	} else if err := bridgeManager.Reload(initialProviders); err != nil {
		log.Printf("warning: could not initialize bridge: %v", err)
	}

	r.GET("/health", handlerPublic.HealthCheck)
	r.GET("/api/status", handlerPublic.GetStatus(cfg))
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))

	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth(cfg))
	{
		api.GET("/me", handlerUser.GetMe)

		access := api.Group("")
		access.Use(middleware.RequireAnyRole(middleware.RoleNone))
		access.POST("/access-requests", handlerUser.CreateAccessRequest(cfg))
		access.GET("/access-requests/me", handlerUser.GetMyAccessRequest)

		user := api.Group("")
		user.Use(middleware.RequireAnyRole(middleware.RoleUser, middleware.RoleAdmin, middleware.RoleManager))
		api.GET("/dashboard", handlerUser.GetDashboard)
		api.GET("/providers", handlerUser.ListAvailableProviders)
		api.GET("/models", handlerUser.GetModels())
		user.POST("/tokens", handlerUser.CreateToken(cfg.TokenSecret))
		user.GET("/tokens", handlerUser.ListTokens)
		user.PATCH("/tokens/:id", handlerUser.UpdateToken)
		user.DELETE("/tokens/:id", handlerUser.RevokeToken)
		user.GET("/history", handlerUser.GetHistory)
		user.GET("/history/stats", handlerUser.GetHistoryStats)
		user.GET("/history/:id", handlerUser.GetHistoryDetail)

		elevated := api.Group("/admin")
		elevated.Use(middleware.RequireAnyRole(middleware.RoleAdmin, middleware.RoleManager))
		elevated.GET("/users", handlerAdmin.ListUsers)
		elevated.PATCH("/users/:id", handlerAdmin.UpdateUserRole)
		elevated.DELETE("/users/:id", handlerAdmin.DeleteUser)
		elevated.GET("/users/:id/stats", handlerAdmin.GetUserStats)
		elevated.GET("/tokens", handlerAdmin.ListTokens)
		elevated.DELETE("/tokens/:id", handlerAdmin.RevokeToken)
		elevated.POST("/tokens/:id/unrevoke", handlerAdmin.UnrevokeToken)
		elevated.GET("/history", handlerAdmin.GetHistory)
		elevated.GET("/history/:id", handlerAdmin.GetHistoryDetail)
		elevated.GET("/access-requests", handlerAdmin.ListAccessRequests)
		elevated.POST("/access-requests/:id/approve", handlerAdmin.ApproveRequest(cfg))
		elevated.POST("/access-requests/:id/reject", handlerAdmin.RejectRequest(cfg))

		admin := api.Group("/admin")
		admin.Use(middleware.RequireRole(middleware.RoleAdmin))
		admin.GET("/whitelist", handlerAdmin.ListWhitelist)
		admin.POST("/whitelist", handlerAdmin.AddWhitelist)
		admin.DELETE("/whitelist/:id", handlerAdmin.DeleteWhitelist)
		admin.PATCH("/whitelist/:id", handlerAdmin.ToggleWhitelist)
		admin.GET("/service-accounts", handlerAdmin.ListServiceAccounts)
		admin.POST("/service-accounts", handlerAdmin.CreateServiceAccount)
		admin.DELETE("/service-accounts/:id", handlerAdmin.DeleteServiceAccount)
		admin.GET("/service-accounts/:id/tokens", handlerAdmin.ListServiceTokens)
		admin.POST("/service-accounts/:id/tokens", handlerAdmin.CreateServiceToken(cfg.TokenSecret))
		admin.GET("/providers", handlerAdmin.ListProviders)
		admin.POST("/providers", handlerAdmin.CreateProvider(bridgeManager))
		admin.GET("/providers/:id", handlerAdmin.GetProvider)
		admin.PUT("/providers/:id", handlerAdmin.UpdateProvider(bridgeManager))
		admin.DELETE("/providers/:id", handlerAdmin.DeleteProvider(bridgeManager))
		admin.POST("/providers/reload", handlerAdmin.ReloadProviders(bridgeManager))
	}

	// AI proxy: any path not matched above is forwarded to the bridge manager.
	aib := r.Group("")
	aib.Use(middleware.JWTAuth(cfg))
	aib.Use(middleware.RequireAnyRole(middleware.RoleUser, middleware.RoleAdmin, middleware.RoleService, middleware.RoleManager))
	aib.Use(middleware.IPWhitelist(cfg.TrustedProxies))
	aib.Use(middleware.AIBridgeActor())
	aib.Any("/:provider/*path", gin.WrapH(bridgeManager))

	// Background job: revoke roles that have passed their expiry date.
	go func() {
		interval := time.Duration(cfg.RoleExpiryIntervalSec) * time.Second
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			database.DB.Model(&models.RegisteredUser{}).
				Where("role != ? AND role_expires_at IS NOT NULL AND role_expires_at <= ?",
					models.RoleNone, time.Now()).
				Update("role", models.RoleNone)
		}
	}()

	log.Printf("server listening on :%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
