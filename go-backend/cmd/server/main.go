package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tanzanite/internal/api/middleware"
	v1 "tanzanite/internal/api/v1"
	"tanzanite/internal/api/v1/admin"
	"tanzanite/internal/pkg/cache"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/database"
	"tanzanite/internal/pkg/logger"
	"tanzanite/internal/pkg/worker"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err := logger.Init(cfg.Log); err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer func() {
		_ = logger.Log.Sync()
	}()

	gin.SetMode(cfg.Server.Mode)

	db, err := database.Init(cfg.Database)
	if err != nil {
		logger.Fatal("database init failed", zap.Error(err))
	}

	if cfg.Database.AutoMigrate {
		if err := database.AutoMigrate(db, cfg.Server.Mode); err != nil {
			logger.Fatal("database auto-migration failed", zap.Error(err))
		}
	}

	sqlDB, err := db.DB()
	if err == nil {
		if err := database.RunSQLMigrations(sqlDB, &cfg.Database); err != nil {
			logger.Warn("SQL migrations failed or skipped", zap.Error(err))
		}
	} else {
		logger.Warn("failed to get sql.DB for migrations", zap.Error(err))
	}

	redisCache, err := cache.Init(cfg.Redis)
	if err != nil {
		logger.Fatal("redis init failed", zap.Error(err))
	}
	defer func() {
		_ = redisCache.Close()
	}()

	router := setupRouter(db, redisCache, cfg)

	server := &http.Server{
		Addr:         cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	var workerServer *worker.Server
	if cfg.Worker.Enabled {
		workerServer = worker.NewServer(&cfg.Redis)
		if err := workerServer.Start(); err != nil {
			logger.Fatal("worker server failed to start", zap.Error(err))
		}
	} else {
		logger.Info("Asynq worker disabled")
	}

	go func() {
		logger.Info("server started",
			zap.String("addr", cfg.Server.Port),
			zap.String("mode", cfg.Server.Mode),
			zap.String("version", Version),
			zap.String("build_time", BuildTime),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("server shutdown failed", zap.Error(err))
	}

	if workerServer != nil {
		workerServer.Stop()
	}

	logger.Info("server stopped")
}

func setupRouter(db *gorm.DB, redisCache *cache.RedisCache, cfg *config.Config) *gin.Engine {
	router := gin.New()
	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS(cfg.CORS))
	router.Use(middleware.SecurityHeaders())

	// 全局限流 - 保护整个服务
	router.Use(middleware.GlobalRateLimit(1000)) // 1000 RPS globally

	if cfg.Server.Mode != gin.ReleaseMode {
		setupSwagger(router)
	}

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "tanzanite-api",
			"version": Version,
		})
	})

	// 使用专用的健康检查处理器（支持Kubernetes探针）
	// 注意: 这些端点使用新的health handler，提供更详细的健康状态
	healthGroup := router.Group("")
	{
		// 导入health handler
		// health.RegisterRoutes(healthGroup, db, redisCache.Client)

		// 临时保留简单的健康检查（直到集成health handler）
		healthGroup.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "ok",
				"service":   "tanzanite-api",
				"version":   Version,
				"buildTime": BuildTime,
			})
		})

		healthGroup.GET("/readiness", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ready",
			})
		})

		healthGroup.GET("/liveness", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "alive",
			})
		})
	}

	v1.RegisterRoutes(router, db, redisCache, cfg)
	admin.RegisterAdminRoutes(router, db, redisCache, cfg)

	return router
}
