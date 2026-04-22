package main

import (
	"log"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

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
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

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
		&models.APIToken{},
		&models.Interception{},
		&models.TokenUsage{},
		&models.UserPrompt{},
		&models.ToolUsage{},
		&models.ModelThought{},
		&models.IPAllowlist{},
		&models.User{},
		&models.AccessRequest{},
		&models.Provider{},
	); err != nil {
		log.Fatalf("auto-migrate error: %v", err)
	}

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis url error: %v", err)
	}
	rdb := redis.NewClient(redisOpts)
	reloadPub := services.NewRedisReloadPublisher(rdb)

	r.GET("/health", handlerPublic.HealthCheck)
	r.GET("/api/status", handlerPublic.GetStatus(cfg))

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
		admin.POST("/providers", handlerAdmin.CreateProvider(reloadPub))
		admin.GET("/providers/:id", handlerAdmin.GetProvider)
		admin.PUT("/providers/:id", handlerAdmin.UpdateProvider(reloadPub))
		admin.DELETE("/providers/:id", handlerAdmin.DeleteProvider(reloadPub))
		admin.POST("/providers/reload", handlerAdmin.ReloadProviders(reloadPub))
	}

	// Background job: revoke roles that have passed their expiry date.
	go func() {
		interval := time.Duration(cfg.RoleExpiryIntervalSec) * time.Second
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			database.DB.Model(&models.User{}).
				Where("role != ? AND role_expires_at IS NOT NULL AND role_expires_at <= ?",
					models.RoleNone, time.Now()).
				Update("role", models.RoleNone)
		}
	}()

	log.Printf("api service listening on :%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
