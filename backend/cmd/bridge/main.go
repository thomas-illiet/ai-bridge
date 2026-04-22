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
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"

	"github.com/coder/aibridge"
	aibpkg "github.com/thomas-illiet/ai-bridge/aibridge"
	"github.com/thomas-illiet/ai-bridge/config"
	"github.com/thomas-illiet/ai-bridge/database"
	handlerPublic "github.com/thomas-illiet/ai-bridge/handlers/public"
	"github.com/thomas-illiet/ai-bridge/middleware"
	"github.com/thomas-illiet/ai-bridge/services"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	logger := slog.Make(sloghuman.Sink(os.Stderr)).Leveled(slog.LevelInfo)

	if err := database.Connect(cfg.DatabaseDSN); err != nil {
		log.Fatalf("database error: %v", err)
	}

	reg := prometheus.NewRegistry()
	metrics := aibridge.NewMetrics(reg)
	tracer := otel.GetTracerProvider().Tracer("ai-bridge")
	recorder := aibpkg.NewGORMRecorder()

	bridgeManager := aibpkg.NewBridgeManager(ctx, recorder, logger, metrics, tracer)

	if initialProviders, err := services.BuildProviders(); err != nil {
		log.Printf("warning: could not load initial providers: %v", err)
	} else if err := bridgeManager.Reload(initialProviders); err != nil {
		log.Printf("warning: could not initialize bridge: %v", err)
	}

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis url error: %v", err)
	}
	rdb := redis.NewClient(redisOpts)

	go subscribeReload(ctx, rdb, bridgeManager)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Split(cfg.AllowedOrigins, ","),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Cache-Control"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "Transfer-Encoding", "X-Content-Type-Options"},
		AllowCredentials: true,
	}))

	r.GET("/health", handlerPublic.HealthCheck)
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))

	aib := r.Group("")
	aib.Use(middleware.JWTAuth(cfg))
	aib.Use(middleware.RequireAnyRole(middleware.RoleUser, middleware.RoleAdmin, middleware.RoleService, middleware.RoleManager))
	aib.Use(middleware.IPWhitelist(cfg.TrustedProxies))
	aib.Use(middleware.AIBridgeActor())
	aib.Any("/:provider/*path", gin.WrapH(bridgeManager))

	log.Printf("bridge service listening on :%s", cfg.BridgeServerPort)
	if err := r.Run(":" + cfg.BridgeServerPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func subscribeReload(ctx context.Context, rdb *redis.Client, bm *aibpkg.BridgeManager) {
	sub := rdb.Subscribe(ctx, services.ReloadChannel())
	defer sub.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-sub.Channel():
			if !ok {
				return
			}
			_ = msg
			providers, err := services.BuildProviders()
			if err != nil {
				log.Printf("bridge reload: build providers: %v", err)
				continue
			}
			if err := bm.Reload(providers); err != nil {
				log.Printf("bridge reload: %v", err)
			} else {
				log.Printf("bridge reloaded: %d providers active", len(providers))
			}
		}
	}
}
