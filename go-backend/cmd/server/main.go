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
	"commerce-platform/internal/pkg/maintenancegate"
	"commerce-platform/internal/pkg/scheduler"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/worker"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/flock"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	command := parseCommand()
	maintenanceContext, cancelMaintenanceWatch := context.WithCancel(context.Background())
	defer cancelMaintenanceWatch()
	// The OS releases this shared lease only after the entire process exits,
	// including every independent worker and scheduler. A persistent fence also
	// blocks restart, migrations and backfill commands before database access.
	var lease *flock.Flock
	var err error
	if command == "register-backup-boundary" {
		directory := os.Getenv("LOCAL_BACKUP_MAINTENANCE_DIRECTORY")
		if directory == "" {
			log.Fatal("LOCAL_BACKUP_MAINTENANCE_DIRECTORY is required") //nolint:gocritic // Fatal startup exits the whole process and releases its process-lifetime maintenance lease.
		}
		timeout, cancel := context.WithTimeout(maintenanceContext, 60*time.Second)
		lease, err = maintenancegate.AcquireExclusiveRecoveryLease(timeout, directory)
		cancel()
	} else {
		lease, err = maintenancegate.AcquireApplicationProcessLease(maintenanceContext, os.Getenv("LOCAL_BACKUP_MAINTENANCE_DIRECTORY"), func() {
			log.Print("application process stopped for independent backup/recovery maintenance")
			os.Exit(0)
		})
	}
	if err != nil {
		log.Fatalf("maintenance gate: %v", err)
	}
	if lease != nil {
		defer func(closeResource func() error) { _ = closeResource() }(lease.Close)
	}
	if command == "render-edge-config" {
		applyEdgeConfigLoadDefaults()
	}

	cfg, err := config.Load(os.Getenv("CONFIG_FILE"))
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if directory := os.Getenv("LOCAL_BACKUP_MAINTENANCE_DIRECTORY"); directory != "" {
		sharedStorage := storage.LoadConfigFromEnv()
		logoStorage := storage.LoadSiteLogoConfigFromEnv(sharedStorage)
		if string(sharedStorage.Type) != "local" || string(logoStorage.Type) != "local" {
			log.Fatal("local computer backup requires explicitly configured local application media")
		}
		recoveryKeys := collectRecoverySigningAndEncryptionKeysFromApplicationConfiguration(cfg)
		externalProjections, err := collectRegisteredRecoveryExternalProjectionConfiguration(cfg, sharedStorage, logoStorage)
		if err != nil {
			log.Fatalf("register recovery external projections: %v", err)
		}
		if err := maintenancegate.RegisterApplicationBoundary(directory, cfg.Database.Database, Version,
			map[string]string{"uploads": sharedStorage.LocalPath, "uploads_private": sharedStorage.PrivateLocalPath, "site_logo_uploads": logoStorage.LocalPath},
			recoveryKeys, externalProjections); err != nil {
			log.Fatalf("register recovery boundary: %v", err)
		}
	}

	if command == "register-backup-boundary" {
		log.Print("offline recovery boundary registered; no database, migration or business job was started")
		return
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
		service.ConfigureMediaDerivativeGenerationCapacity(cfg.Worker.MediaDerivativeGenerationCapacity)

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
	if os.Getenv("LOCAL_BACKUP_PRODUCT_VIEW_REDIS_OWNERSHIP_ENABLED") == "true" {
		if !cfg.Worker.ProductViewCountFlushEnabled || (cfg.Redis.NormalizedMode() != "standalone" && cfg.Redis.NormalizedMode() != "sentinel") {
			logger.Fatal("product view Redis recovery ownership requires standalone or Sentinel buffered views")
		}
		ownershipContext, cancelOwnership := context.WithTimeout(maintenanceContext, 15*time.Second)
		projection := service.DescribeProductCatalogRuntimeProjectionBoundaryForRecovery(true)
		ownershipErr := maintenancegate.RegisterProductViewRedisRuntimeOwnershipBeforeStartingWriters(ownershipContext,
			os.Getenv("LOCAL_BACKUP_MAINTENANCE_DIRECTORY"), os.Getenv("LOCAL_BACKUP_SITE_ID"), os.Getenv("LOCAL_BACKUP_ENVIRONMENT"), cfg.Database.Database,
			maintenancegate.FingerprintRedisRecoveryDestination(cfg.Redis.NormalizedMode(), cfg.Redis.GetRedisAddrs(), cfg.Redis.DB, cfg.Redis.MasterName), projection.ViewCountKeysSHA256, redisCache.Client())
		cancelOwnership()
		if ownershipErr != nil {
			logger.Fatal("product view Redis ownership registration refused", zap.Error(ownershipErr))
		}
	}

	deps, err := app.NewDependencies(db, redisCache, cfg)
	if err != nil {
		logger.Fatal("dependency initialization failed", zap.Error(err))
	}
	// Keep the backend route catalog and FAQ page structure self-healing after
	// every deployment. A failed storefront fetch is non-fatal so the API can
	// still serve existing content while the next explicit URL sync retries it.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		var summary service.StorefrontRouteCatalogSyncSummary
		var syncErr error
		for attempt := 1; attempt <= 3; attempt++ {
			summary, syncErr = deps.Services.StorefrontRouteCatalog.Sync(ctx)
			if syncErr == nil {
				break
			}
			if attempt < 3 {
				timer := time.NewTimer(5 * time.Second)
				select {
				case <-ctx.Done():
					timer.Stop()
				case <-timer.C:
				}
			}
		}
		if syncErr != nil {
			logger.Warn("storefront route and FAQ startup sync failed", zap.Error(syncErr))
			return
		}
		logger.Info(
			"storefront route and FAQ startup sync completed",
			zap.String("manifest_version", summary.ManifestVersion),
			zap.Int("route_entries", summary.Entries),
			zap.Int("faq_pages_created", summary.FAQ.Created),
			zap.Int("faq_pages_stale", summary.FAQ.Stale),
		)
	}()
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
		trackingScheduler = scheduler.NewTrackingScheduler(deps.Services.Shipping, cfg.Worker, deps.RedisClient)
		trackingScheduler.Start(context.Background())
	} else {
		deps.Services.Shipping.ConfigureTrackingPolling(false, time.Duration(cfg.Worker.TrackingPollingIntervalSeconds)*time.Second, cfg.Worker.TrackingPollingBatchLimit)
		logger.Info("tracking scheduler disabled")
	}

	var yanwenTrackingScheduler *scheduler.YanwenTrackingScheduler
	if cfg.Worker.YanwenTrackingPollingEnabled {
		yanwenTrackingScheduler = scheduler.NewYanwenTrackingScheduler(deps.Services.YanwenTrackingPolling, cfg.Worker, deps.RedisClient)
		yanwenTrackingScheduler.Start(context.Background())
	} else {
		logger.Info("Yanwen tracking scheduler disabled")
	}

	var yanwenCatalogSyncScheduler *scheduler.YanwenCatalogSyncScheduler
	if cfg.Worker.YanwenCatalogSyncEnabled {
		yanwenCatalogSyncScheduler = scheduler.NewYanwenCatalogSyncScheduler(deps.Services.YanwenOfficialCatalog, cfg.Worker, deps.RedisClient)
		yanwenCatalogSyncScheduler.Start(context.Background())
	} else {
		logger.Info("Yanwen catalog sync scheduler disabled")
	}

	var visitorProfileCleanupScheduler *scheduler.VisitorProfileCleanupScheduler
	if cfg.Worker.VisitorProfileCleanupEnabled {
		visitorProfileCleanupScheduler = scheduler.NewVisitorProfileCleanupScheduler(deps.Services.VisitorProfile, cfg.Worker)
		visitorProfileCleanupScheduler.Start(context.Background())
	} else {
		logger.Info("visitor profile cleanup scheduler disabled")
	}

	customerServiceRetentionScheduler := scheduler.NewCustomerServiceRetentionScheduler(deps.Services.CustomerServiceRetention, cfg.Worker)
	customerServiceRetentionScheduler.ConfigureBusinessDraftMediaUploadExpiry(deps.Services.Media)
	customerServiceRetentionScheduler.Start(context.Background())

	var behaviorEventCleanupScheduler *scheduler.BehaviorEventCleanupScheduler
	if cfg.Worker.BehaviorEventCleanupEnabled {
		behaviorEventCleanupScheduler = scheduler.NewBehaviorEventCleanupScheduler(deps.Services.BehaviorEvents, cfg.Worker)
		behaviorEventCleanupScheduler.Start(context.Background())
	} else {
		logger.Info("behavior event cleanup scheduler disabled")
	}

	var quickBuyAbandonmentScheduler *scheduler.QuickBuyAbandonmentScheduler
	if cfg.Worker.QuickBuyAbandonmentEnabled {
		quickBuyAbandonmentScheduler = scheduler.NewQuickBuyAbandonmentScheduler(deps.Services.QuickBuy, cfg.Worker)
		quickBuyAbandonmentScheduler.Start(context.Background())
	} else {
		logger.Info("quick-buy abandonment scheduler disabled")
	}

	var shippingQuoteCleanupScheduler *scheduler.ShippingQuoteCleanupScheduler
	if cfg.Worker.ShippingQuoteCleanupEnabled {
		shippingQuoteCleanupScheduler = scheduler.NewShippingQuoteCleanupScheduler(deps.Services.Shipping, cfg.Worker)
		shippingQuoteCleanupScheduler.Start(context.Background())
	} else {
		logger.Info("shipping quote cleanup scheduler disabled")
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

	var referralLifecycleScheduler *scheduler.ReferralLifecycleScheduler
	if cfg.Worker.ReferralLifecycleEnabled {
		referralLifecycleScheduler = scheduler.NewReferralLifecycleScheduler(deps.Services.Referral, cfg.Worker, deps.RedisClient)
		referralLifecycleScheduler.Start(context.Background())
	} else {
		logger.Info("referral lifecycle scheduler disabled")
	}

	var paymentExpirationScheduler *scheduler.PaymentExpirationScheduler
	if cfg.Worker.PaymentExpirationEnabled {
		paymentExpirationScheduler = scheduler.NewPaymentExpirationScheduler(deps.Services.Order, cfg.Worker, deps.RedisClient)
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

	taxRateSourceSnapshotSyncScheduler := scheduler.NewTaxRateSourceSnapshotSyncScheduler(deps.Services.TaxRateSourceSnapshot)
	taxRateSourceSnapshotSyncScheduler.Start(context.Background())

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
	var productViewCountFlushScheduler *scheduler.ProductViewCountFlushScheduler
	if cfg.Worker.ProductViewCountFlushEnabled {
		productViewCountFlushScheduler = scheduler.NewProductViewCountFlushScheduler(deps.Services.Product, deps.RedisClient, cfg.Worker)
		productViewCountFlushScheduler.Start(context.Background())
	} else {
		logger.Info("product view count flush scheduler disabled")
	}
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
	if yanwenTrackingScheduler != nil {
		yanwenTrackingScheduler.Stop()
	}
	if yanwenCatalogSyncScheduler != nil {
		yanwenCatalogSyncScheduler.Stop()
	}
	if visitorProfileCleanupScheduler != nil {
		visitorProfileCleanupScheduler.Stop()
	}
	if customerServiceRetentionScheduler != nil {
		customerServiceRetentionScheduler.Stop()
	}
	if behaviorEventCleanupScheduler != nil {
		behaviorEventCleanupScheduler.Stop()
	}
	if quickBuyAbandonmentScheduler != nil {
		quickBuyAbandonmentScheduler.Stop()
	}
	if shippingQuoteCleanupScheduler != nil {
		shippingQuoteCleanupScheduler.Stop()
	}
	if ugcShowcaseCleanupScheduler != nil {
		ugcShowcaseCleanupScheduler.Stop()
	}
	if outboxDispatchScheduler != nil {
		outboxDispatchScheduler.Stop()
	}
	if referralLifecycleScheduler != nil {
		referralLifecycleScheduler.Stop()
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
	if taxRateSourceSnapshotSyncScheduler != nil {
		taxRateSourceSnapshotSyncScheduler.Stop()
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
	if productViewCountFlushScheduler != nil {
		productViewCountFlushScheduler.Stop()
	}
	if deps.CustomerServiceRealtimeRelay != nil {
		deps.CustomerServiceRealtimeRelay.Stop()
	}
	deps.Services.GlobalIPBlock.StopCacheInvalidationListener()

	logger.Info("server stopped")
}

func collectRecoverySigningAndEncryptionKeysFromApplicationConfiguration(cfg *config.Config) map[string]string {
	keys := map[string]string{
		"JWT_SECRET":                cfg.JWT.Secret,
		"PAYMENT_CONFIG_MASTER_KEY": os.Getenv("PAYMENT_CONFIG_MASTER_KEY"),
		"OPS_CONNECTOR_MASTER_KEY":  os.Getenv("OPS_CONNECTOR_MASTER_KEY"),
		"ORDER_NUMBER_SECRET":       cfg.OrderNumber.EffectiveSecret(cfg.JWT.Secret),
	}
	for name, value := range map[string]string{
		"ORDER_NUMBER_PREVIOUS_SECRET":         cfg.OrderNumber.EffectivePreviousSecret(),
		"GOOGLE_MERCHANT_TOKEN_ENCRYPTION_KEY": cfg.GoogleMerchant.TokenEncryptionKey,
		"SOCIAL_OAUTH_TOKEN_ENCRYPTION_KEY":    cfg.SocialOAuth.TokenEncryptionKey,
	} {
		if value != "" {
			keys[name] = value
		}
	}
	for _, name := range []string{"EMAIL_PROVIDER_MASTER_KEY", "SECURITY_MASTER_KEY", "FPX_API_MASTER_KEY", "YANWEN_API_MASTER_KEY"} {
		if value := os.Getenv(name); value != "" {
			keys[name] = value
		}
	}
	return keys
}

func parseCommand() string {
	if len(os.Args) <= 1 {
		return "serve"
	}
	if len(os.Args) == 2 && (os.Args[1] == "migrate" || os.Args[1] == "render-edge-config" || os.Args[1] == "backfill-media-derivatives" || os.Args[1] == "register-backup-boundary") {
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
