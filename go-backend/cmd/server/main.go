package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"commerce-platform/internal/api/admin"
	"commerce-platform/internal/api/edge"
	"commerce-platform/internal/api/middleware"
	v1 "commerce-platform/internal/api/v1"
	"commerce-platform/internal/api/v1/health"
	"commerce-platform/internal/app"
	"commerce-platform/internal/pkg/cache"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/database"
	"commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/pkg/scheduler"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/worker"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	command := parseCommand()
	if command == "render-edge-config" {
		applyEdgeConfigLoadDefaults()
	}

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

	if command == "render-edge-config" {
		environment := strings.TrimSpace(os.Getenv("OPS_EDGE_CONFIG_ENVIRONMENT"))
		outputDir := strings.TrimSpace(os.Getenv("OPS_EDGE_CONFIG_OUTPUT_DIR"))
		if environment == "" || outputDir == "" {
			logger.Fatal(
				"edge config rendering requires OPS_EDGE_CONFIG_ENVIRONMENT and OPS_EDGE_CONFIG_OUTPUT_DIR",
			)
		}

		edgeConfigService := service.NewOpsEdgeConfigService(
			repository.NewOpsDomainBindingRepository(db),
			repository.NewOpsProjectBindingRepository(db),
		)
		if _, err := edgeConfigService.RenderToDirectory(environment, outputDir); err != nil {
			logger.Fatal("edge config rendering failed", zap.Error(err))
		}
		logger.Info(
			"edge config rendering completed",
			zap.String("environment", environment),
			zap.String("output_dir", outputDir),
		)
		return
	}

	if command == "backfill-media-derivatives" {
		storageSvc, err := storage.NewStorageService(storage.LoadConfigFromEnv())
		if err != nil {
			logger.Fatal("media derivative backfill storage initialization failed", zap.Error(err))
		}
		mediaRepo := repository.NewMediaRepository(db)
		presetRepo := repository.NewMediaDerivativePresetRepository(db)
		if err := service.SeedDefaultMediaDerivativePresets(presetRepo); err != nil {
			logger.Fatal("seed media derivative presets failed", zap.Error(err))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
		mediaService := service.NewMediaService(
			mediaRepo,
			storageSvc,
			nil,
			cfg.Server.BaseURL,
			cfg.MediaUpload.AccountStorageQuotaBytes,
		)
		mediaService.ConfigureDerivativePresetRepository(presetRepo)
		result, backfillErr := mediaService.BackfillMissingImageDerivatives(ctx)
		cancel()
		if backfillErr != nil {
			logger.Fatal("media derivative backfill failed", zap.Error(backfillErr))
		}
		logger.Info(
			"media derivative backfill completed",
			zap.Int("scanned_assets", result.ScannedAssets),
			zap.Int("generated_assets", result.GeneratedAssets),
			zap.Int("generated_derivatives", result.GeneratedDerivatives),
			zap.Int64("updated_product_media_rows", result.UpdatedProductMediaRows),
		)
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
	deps.Services.GlobalIPBlock.StartCacheInvalidationListener(context.Background())
	if deps.CustomerServiceRealtimeRelay != nil {
		if err := deps.CustomerServiceRealtimeRelay.Start(context.Background()); err != nil {
			logger.Fatal("customer-service realtime relay failed to start", zap.Error(err))
		}
	}

	router := setupRouter(db, redisCache, cfg, deps)

	server := &http.Server{
		Addr:              cfg.Server.Port,
		Handler:           router,
		ReadTimeout:       time.Duration(cfg.Server.ReadTimeout) * time.Second,
		ReadHeaderTimeout: time.Duration(cfg.Server.ReadHeaderTimeout) * time.Second,
		WriteTimeout:      time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:       time.Duration(cfg.Server.IdleTimeout) * time.Second,
		MaxHeaderBytes:    cfg.Server.MaxHeaderBytes,
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

	var visitorProfileCleanupScheduler *scheduler.VisitorProfileCleanupScheduler
	if cfg.Worker.VisitorProfileCleanupEnabled {
		visitorProfileCleanupScheduler = scheduler.NewVisitorProfileCleanupScheduler(deps.Services.VisitorProfile, cfg.Worker)
		visitorProfileCleanupScheduler.Start(context.Background())
	} else {
		logger.Info("visitor profile cleanup scheduler disabled")
	}

	var behaviorEventCleanupScheduler *scheduler.BehaviorEventCleanupScheduler
	if cfg.Worker.BehaviorEventCleanupEnabled {
		behaviorEventCleanupScheduler = scheduler.NewBehaviorEventCleanupScheduler(deps.Services.BehaviorEvents, cfg.Worker)
		behaviorEventCleanupScheduler.Start(context.Background())
	} else {
		logger.Info("behavior event cleanup scheduler disabled")
	}

	var ugcShowcaseCleanupScheduler *scheduler.UGCShowcaseCleanupScheduler
	if cfg.Worker.ShowcaseCleanupEnabled {
		ugcShowcaseCleanupScheduler = scheduler.NewUGCShowcaseCleanupScheduler(deps.Services.UGCShowcase, cfg.Worker)
		ugcShowcaseCleanupScheduler.Start(context.Background())
	} else {
		logger.Info("showcase cleanup scheduler disabled")
	}

	var outboxDispatchScheduler *scheduler.OutboxDispatchScheduler
	if cfg.Worker.OutboxDispatchEnabled {
		if deps.Services.Outbox == nil || deps.Services.Outbox.HandlerCount() == 0 {
			logger.Warn("outbox dispatch scheduler enabled but no handlers are configured; pending events will be retained")
		} else {
			outboxDispatchScheduler = scheduler.NewOutboxDispatchScheduler(deps.Services.Outbox, cfg.Worker)
			outboxDispatchScheduler.Start(context.Background())
		}
	} else {
		logger.Info("outbox dispatch scheduler disabled")
	}

	var paymentExpirationScheduler *scheduler.PaymentExpirationScheduler
	if cfg.Worker.PaymentExpirationEnabled {
		paymentExpirationScheduler = scheduler.NewPaymentExpirationScheduler(deps.Services.Order, cfg.Worker)
		paymentExpirationScheduler.Start(context.Background())
	} else {
		logger.Info("payment expiration scheduler disabled")
	}

	var paymentRiskMonitoringScheduler *scheduler.PaymentRiskMonitoringScheduler
	if cfg.Worker.PaymentRiskMonitoringEnabled {
		paymentRiskMonitoringScheduler = scheduler.NewPaymentRiskMonitoringScheduler(
			deps.Services.PaymentRiskMonitoring,
			cfg.Worker,
		)
		paymentRiskMonitoringScheduler.Start(context.Background())
	} else {
		logger.Info("payment risk monitoring scheduler disabled")
	}

	var exchangeRateSyncScheduler *scheduler.ExchangeRateSyncScheduler
	if cfg.Worker.ExchangeRateSyncEnabled {
		exchangeRateSyncScheduler = scheduler.NewExchangeRateSyncScheduler(deps.Services.ExchangeRate, cfg.Worker)
		exchangeRateSyncScheduler.Start(context.Background())
	} else {
		logger.Info("exchange rate sync scheduler disabled")
	}

	var siteQualityWorker *scheduler.SiteQualityWorker
	if cfg.Worker.SiteQualityEnabled {
		siteQualityWorker = scheduler.NewSiteQualityWorker(
			deps.Services.SiteQualityEngine,
			cfg.Worker,
		)
		siteQualityWorker.Start(context.Background())
	} else {
		logger.Info("site quality job worker disabled")
	}

	var hotDataArchiveScheduler *scheduler.HotDataArchiveScheduler
	if cfg.Worker.HotDataArchiveEnabled {
		hotDataArchiveScheduler = scheduler.NewHotDataArchiveScheduler(
			deps.Services.HotDataArchive,
			cfg.Worker,
		)
		hotDataArchiveScheduler.Start(context.Background())
	} else {
		logger.Info("hot data archive scheduler disabled")
	}

	var mediaDerivativeRebuildScheduler *scheduler.MediaDerivativeRebuildScheduler
	if cfg.Worker.MediaDerivativeRebuildEnabled {
		mediaDerivativeRebuildScheduler = scheduler.NewMediaDerivativeRebuildScheduler(deps.Services.Media, cfg.Worker)
		mediaDerivativeRebuildScheduler.Start(context.Background())
	} else {
		logger.Info("media derivative rebuild scheduler disabled")
	}

	var visitorRiskFlushScheduler *scheduler.VisitorRiskFlushScheduler
	if cfg.VisitorRisk.Enabled {
		visitorRiskFlushScheduler = scheduler.NewVisitorRiskFlushScheduler(deps.Services.VisitorRisk, cfg.VisitorRisk)
		visitorRiskFlushScheduler.Start(context.Background())
	} else {
		logger.Info("visitor risk telemetry disabled")
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
	if visitorProfileCleanupScheduler != nil {
		visitorProfileCleanupScheduler.Stop()
	}
	if behaviorEventCleanupScheduler != nil {
		behaviorEventCleanupScheduler.Stop()
	}
	if ugcShowcaseCleanupScheduler != nil {
		ugcShowcaseCleanupScheduler.Stop()
	}
	if outboxDispatchScheduler != nil {
		outboxDispatchScheduler.Stop()
	}
	if paymentExpirationScheduler != nil {
		paymentExpirationScheduler.Stop()
	}
	if paymentRiskMonitoringScheduler != nil {
		paymentRiskMonitoringScheduler.Stop()
	}
	if exchangeRateSyncScheduler != nil {
		exchangeRateSyncScheduler.Stop()
	}
	if siteQualityWorker != nil {
		siteQualityWorker.Stop()
	}
	if hotDataArchiveScheduler != nil {
		hotDataArchiveScheduler.Stop()
	}
	if mediaDerivativeRebuildScheduler != nil {
		mediaDerivativeRebuildScheduler.Stop()
	}
	if visitorRiskFlushScheduler != nil {
		visitorRiskFlushScheduler.Stop()
	}
	if deps.CustomerServiceRealtimeRelay != nil {
		deps.CustomerServiceRealtimeRelay.Stop()
	}
	deps.Services.GlobalIPBlock.StopCacheInvalidationListener()

	logger.Info("server stopped")
}

func parseCommand() string {
	if len(os.Args) <= 1 {
		return "serve"
	}
	if len(os.Args) == 2 && (os.Args[1] == "migrate" || os.Args[1] == "render-edge-config" || os.Args[1] == "backfill-media-derivatives") {
		return os.Args[1]
	}
	log.Fatalf("unsupported command: %v", os.Args[1:])
	return ""
}

func applyEdgeConfigLoadDefaults() {
	setDefaultEnv("REDIS_HOST", "127.0.0.1")
	setDefaultEnv("REDIS_PORT", "6379")
	setDefaultEnv("REDIS_PASSWORD", "edge-config-unused")
}

func setDefaultEnv(key, value string) {
	if _, ok := os.LookupEnv(key); ok {
		return
	}
	_ = os.Setenv(key, value)
}

func setupRouter(db *gorm.DB, redisCache *cache.RedisCache, cfg *config.Config, deps *app.Dependencies) *gin.Engine {
	router := gin.New()
	if err := router.SetTrustedProxies(cfg.Server.TrustedProxies); err != nil {
		logger.Fatal("trusted proxy configuration failed", zap.Error(err))
	}
	trustedEdgeMetadata, err := middleware.NewTrustedEdgeMetadata(cfg.Server.TrustedProxies)
	if err != nil {
		logger.Fatal("trusted edge metadata configuration failed", zap.Error(err))
	}
	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS(cfg.CORS))
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.PrometheusMetrics())
	router.Use(trustedEdgeMetadata)
	router.Use(middleware.GlobalIPBlocker(deps.Services.GlobalIPBlock))
	router.Use(middleware.CommercialCrawlerBlocker(deps.Services.GlobalIPBlock))
	router.Use(middleware.RequestSignature(cfg.RequestSigning, redisCache.Client()))
	router.Use(middleware.GlobalRateLimit(1000))

	if cfg.Server.Mode != gin.ReleaseMode {
		setupSwagger(router)
	}

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "commerce-platform-api",
			"version": Version,
		})
	})

	edge.RegisterRoutes(router, deps.Services.GlobalIPBlock)
	health.RegisterRoutes(router.Group(""), db, redisCache.Client(), Version, BuildTime)
	registerLocalUploadsRoute(router, deps)

	v1.RegisterRoutes(router, deps, cfg)
	admin.RegisterAdminRoutes(router, deps, cfg)

	return router
}
