package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tanzanite/internal/api/admin"
	"tanzanite/internal/api/middleware"
	v1 "tanzanite/internal/api/v1"
	"tanzanite/internal/api/v1/health"
	"tanzanite/internal/app"
	"tanzanite/internal/pkg/cache"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/database"
	"tanzanite/internal/pkg/logger"
	"tanzanite/internal/pkg/scheduler"
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
	cfg, err := config.Load(os.Getenv("CONFIG_FILE"))
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err := logger.Init(cfg.Log); err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer func() {
		_ = logger.Log.Sync()
	}()

	command := "serve"
	if len(os.Args) > 1 {
		if len(os.Args) != 2 || os.Args[1] != "migrate" {
			logger.Fatal("unsupported command", zap.Strings("arguments", os.Args[1:]))
		}
		command = os.Args[1]
	}

	gin.SetMode(cfg.Server.Mode)

	db, err := database.Init(cfg.Database)
	if err != nil {
		logger.Fatal("database init failed", zap.Error(err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("get database connection failed", zap.Error(err))
	}
	defer func() {
		_ = sqlDB.Close()
	}()

	if command == "migrate" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		migrationErr := database.PrepareSchema(ctx, db, &cfg.Database, cfg.Server.Mode)
		cancel()
		if migrationErr != nil {
			logger.Fatal("database migration failed", zap.Error(migrationErr))
		}
		logger.Info("database migration completed")
		return
	}

	if cfg.Database.AutoMigrate {
		if cfg.Server.Mode == gin.ReleaseMode {
			logger.Fatal("DB_AUTO_MIGRATE must be false in release mode; run the migrate command before starting the API")
		}
		if cfg.Database.Driver == "postgres" {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			migrationErr := database.PrepareSchema(ctx, db, &cfg.Database, cfg.Server.Mode)
			cancel()
			if migrationErr != nil {
				logger.Fatal("database migration failed", zap.Error(migrationErr))
			}
		} else if err := database.AutoMigrate(db, cfg.Server.Mode); err != nil {
			logger.Fatal("database auto-migration failed", zap.Error(err))
		}
	}

	redisCache, err := cache.Init(cfg.Redis)
	if err != nil {
		logger.Fatal("redis init failed", zap.Error(err))
	}
	defer func() {
		_ = redisCache.Close()
	}()

	deps, err := app.NewDependencies(db, redisCache, cfg)
	if err != nil {
		logger.Fatal("dependency initialization failed", zap.Error(err))
	}

	router := setupRouter(db, redisCache, cfg, deps)

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

	var trackingScheduler *scheduler.TrackingScheduler
	if cfg.Worker.TrackingPollingEnabled {
		trackingScheduler = scheduler.NewTrackingScheduler(deps.Services.Shipping, cfg.Worker)
		trackingScheduler.Start(context.Background())
	} else {
		deps.Services.Shipping.ConfigureTrackingPolling(false, time.Duration(cfg.Worker.TrackingPollingIntervalSeconds)*time.Second, cfg.Worker.TrackingPollingBatchLimit)
		logger.Info("tracking scheduler disabled")
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
	if trackingScheduler != nil {
		trackingScheduler.Stop()
	}

	logger.Info("server stopped")
}

func setupRouter(db *gorm.DB, redisCache *cache.RedisCache, cfg *config.Config, deps *app.Dependencies) *gin.Engine {
	router := gin.New()
	if err := router.SetTrustedProxies(cfg.Server.TrustedProxies); err != nil {
		logger.Fatal("trusted proxy configuration failed", zap.Error(err))
	}
	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS(cfg.CORS))
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.GlobalRateLimit(1000))

	if cfg.Server.Mode != gin.ReleaseMode {
		setupSwagger(router)
	}

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "tanzanite-api",
			"version": Version,
		})
	})

	health.RegisterRoutes(router.Group(""), db, redisCache.Client(), Version, BuildTime)

	v1.RegisterRoutes(router, deps, cfg)
	admin.RegisterAdminRoutes(router, deps, cfg)

	return router
}
