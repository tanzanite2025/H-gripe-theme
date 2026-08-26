package app

import (
	"fmt"
	"os"
	"strings"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/antibot"
	"commerce-platform/internal/pkg/antifraud"
	"commerce-platform/internal/pkg/cache"
	"commerce-platform/internal/pkg/cardtesting"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/email"
	"commerce-platform/internal/pkg/orderabuse"
	"commerce-platform/internal/pkg/ordernumber"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Dependencies struct {
	Repositories                 Repositories
	Services                     Services
	Storage                      storage.StorageService
	SiteLogoStorage              storage.StorageService
	AntiBot                      *antibot.Service
	AntiFraud                    *antifraud.Service
	CardBINLimiter               *cardtesting.Service
	PaymentGatewayCircuitBreaker *service.PaymentGatewayCircuitBreakerService
	OrderAbuse                   *orderabuse.Service
	RedisClient                  redis.UniversalClient
	CustomerServiceRealtimeRelay *service.CustomerServiceRealtimeRelay
}

type Repositories struct {
	User                         *repository.UserRepository
	Post                         *repository.PostRepository
	StorefrontRouteCatalog       *repository.StorefrontRouteCatalogRepository
	StorefrontRedirectRules      *repository.StorefrontRedirectRuleRepository
	StorefrontURLIssues          *repository.StorefrontURLIssueRepository
	PreflightContentLinks        *repository.PreflightContentLinkRepository
	SiteQualityTargets           *repository.SiteQualityTargetRepository
	SiteQualityJobs              *repository.SiteQualityJobRepository
	SiteQualityRuns              *repository.SiteQualityRunRepository
	SiteQualityFindings          *repository.SiteQualityFindingRepository
	HotDataArchive               *repository.HotDataArchiveRepository
	Product                      *repository.ProductRepository
	ProductProcurement           *repository.ProductProcurementRepository
	FrameFitmentEntry            *repository.FrameFitmentEntryRepository
	ForkFitmentEntry             *repository.ForkFitmentEntryRepository
	FitmentFrameHubSpecification *repository.FitmentFrameHubSpecificationRepository
	FitmentForkHubSpecification  *repository.FitmentForkHubSpecificationRepository
	FitmentHubSpecification      *repository.FitmentHubSpecificationRepository
	ProductProcurementCatalog    *repository.ProductProcurementCatalogRepository
	ProductProfitCalculation     *repository.ProductProfitCalculationRepository
	ProductCategory              *repository.ProductCategoryRepository
	ProductBrand                 *repository.ProductBrandRepository
	ProductInformationTemplate   *repository.ProductInformationTemplateRepository
	CustomsClassification        *repository.CustomsClassificationRepository
	Cart                         *repository.CartRepository
	Setting                      *repository.SettingRepository
	FAQ                          *repository.FAQRepository
	Order                        *repository.OrderRepository
	AfterSales                   *repository.AfterSalesCaseRepository
	AfterSalesRefundReview       *repository.AfterSalesRefundReviewRepository
	OrderIdempotency             *repository.OrderIdempotencyRepository
	OrderPolicyDisclosure        *repository.OrderPolicyDisclosureRepository
	OrderAttribution             *repository.OrderAttributionRepository
	Payment                      *repository.PaymentRepository
	PaymentRisk                  *repository.PaymentRiskRepository
	PaymentProtection            *repository.PaymentProtectionRepository
	PaymentRefundReview          *repository.PaymentRefundRecommendationRepository
	PaymentRefundExec            *repository.PaymentRefundExecutionRepository
	ExchangeRate                 *repository.ExchangeRateRepository
	Shipping                     *repository.ShippingRepository
	Coupon                       *repository.CouponRepository
	Loyalty                      *repository.LoyaltyRepository
	LoyaltyProgram               *repository.LoyaltyProgramRepository
	GiftCardRedemption           *repository.GiftCardRedemptionRepository
	Review                       *repository.ReviewRepository
	Ticket                       *repository.TicketRepository
	Gallery                      *repository.GalleryRepository
	Media                        *repository.MediaRepository
	SiteLogo                     *repository.SiteLogoRepository
	MediaDerivativePresets       *repository.MediaDerivativePresetRepository
	MediaDerivativeRebuildJobs   *repository.MediaDerivativeRebuildJobRepository
	StorefrontMarket             *repository.StorefrontMarketRepository
	OpsDomainBinding             *repository.OpsDomainBindingRepository
	OpsConnector                 *repository.OpsConnectorRepository
	OpsConnectorOAuth            *repository.OpsConnectorOAuthRepository
	OpsVPSBinding                *repository.OpsVPSBindingRepository
	OpsProjectBinding            *repository.OpsProjectBindingRepository
	OpsNetworkRule               *repository.OpsNetworkRuleRepository
	OpsDeploymentWorkflow        *repository.OpsDeploymentWorkflowRepository
	GoogleMerchant               *repository.GoogleMerchantRepository
	SocialOAuth                  *repository.SocialOAuthRepository
	Warranty                     *repository.WarrantyRepository
	ShipmentRecord               *repository.ShipmentRecordRepository
	Audit                        *repository.AuditRepository
	Showcase                     *repository.ShowcaseRepository
	VisualShowcase               *repository.VisualShowcaseRepository
	Wishlist                     *repository.WishlistRepository
	Feedback                     *repository.FeedbackRepository
	SuggestionFeedback           *repository.SuggestionFeedbackRepository
	Spoke                        *repository.SpokeRepository
	QuickBuy                     *repository.QuickBuyRepository
	SelectionAssistant           *repository.SelectionAssistantRepository
	SelectionConfigurationKey    *repository.SelectionConfigurationKeyRepository
	WheelsetFitQuestionnaire     *repository.WheelsetFitQuestionnaireRepository
	Subscription                 *repository.SubscriptionRepository
	EmailChallenge               *repository.EmailChallengeRepository
	VisitorProfile               *repository.VisitorProfileRepository
	RecommendationEvent          *repository.RecommendationEventRepository
	VisitorRiskFact              *repository.VisitorRiskFactRepository
	Outbox                       *repository.OutboxRepository
}

type Services struct {
	Auth                              *service.AuthService
	AdminAccountMaintenance           *service.AdminAccountMaintenanceService
	Post                              *service.PostService
	Product                           *service.ProductService
	ProductProcurement                *service.ProductProcurementService
	FrameFitmentEntry                 *service.FrameFitmentEntryService
	ForkFitmentEntry                  *service.ForkFitmentEntryService
	FitmentHubSpecification           *service.FitmentHubSpecificationService
	ProductProfitability              *service.ProductProfitabilityService
	ProductCategory                   *service.ProductCategoryService
	ProductBrand                      *service.ProductBrandService
	ProductInformationTemplate        *service.ProductInformationTemplateService
	CustomsClassification             *service.CustomsClassificationService
	Cart                              *service.CartService
	Setting                           *service.SettingService
	WebsiteProfile                    *service.WebsiteProfileService
	WebsiteName                       *service.WebsiteNameService
	RefundReturnPolicy                *service.RefundReturnPolicyService
	PayPalDisputeInvoiceSellerProfile *service.PayPalDisputeInvoiceSellerProfileService
	AdminSettings                     *service.AdminSettingsService
	SEO                               *service.SEOService
	SEOResources                      *service.SEOResourceService
	GoogleIndexing                    *service.GoogleIndexingService
	Analytics                         *service.AnalyticsService
	AdminPublicChat                   *service.AdminPublicChatAgentService
	CustomerServiceAvatar             *service.CustomerServiceAvatarService
	FAQ                               *service.FAQService
	Gallery                           *service.GalleryService
	Media                             *service.MediaService
	SiteLogo                          *service.SiteLogoService
	Warranty                          *service.WarrantyService
	ShipmentRecord                    *service.ShipmentRecordService
	Checkout                          *service.CheckoutService
	Order                             *service.OrderService
	AfterSales                        *service.AfterSalesService
	Payment                           *service.PaymentService
	Marketing                         *service.MarketingService
	LoyaltyProgram                    *service.LoyaltyProgramService
	Review                            *service.ReviewService
	ReviewModeration                  *service.ReviewModerationService
	Ticket                            *service.TicketService
	CustomerServiceContext            *service.CustomerServiceContextService
	CustomerServiceAnalytics          *service.CustomerServiceAnalyticsService
	CustomerServiceEvents             *service.CustomerServiceEventHub
	Subscription                      *service.SubscriptionService
	Sitemap                           *service.SitemapService
	StorefrontRouteCatalog            *service.StorefrontRouteCatalogService
	StorefrontRedirectRules           *service.StorefrontRedirectRuleService
	StorefrontURLIssues               *service.StorefrontURLIssueService
	PreflightContentLinks             *service.PreflightContentLinkService
	LighthouseRunner                  *service.LighthouseRunnerService
	SiteQualityEngine                 *service.SiteQualityEngineService
	HotDataArchive                    *service.HotDataArchiveService
	Showcase                          *service.ShowcaseService
	Wishlist                          *service.WishlistService
	Feedback                          *service.FeedbackService
	SuggestionFeedback                *service.SuggestionFeedbackService
	User                              *service.UserService
	Dashboard                         *service.DashboardService
	Audit                             *service.AuditService
	Shipping                          *service.ShippingService
	Spoke                             *service.SpokeService
	QuickBuy                          *service.QuickBuyService
	SelectionAssistant                *service.SelectionAssistantService
	SelectionConfigurationKey         *service.SelectionConfigurationKeyService
	WheelsetFitQuestionnaire          *service.WheelsetFitQuestionnaireService
	VisitorProfile                    *service.VisitorProfileService
	BehaviorEvents                    *service.BehaviorEventService
	Recommendations                   *service.RecommendationService
	VisitorRisk                       *service.VisitorRiskService
	PaymentRiskMonitoring             *service.PaymentRiskMonitoringService
	PaymentProtection                 *service.PaymentProtectionService
	PaymentRefundReview               *service.PaymentRefundRecommendationService
	PaymentThreeDS                    *service.PaymentThreeDSPolicyService
	Outbox                            *service.OutboxService
	CurrencyPolicy                    *service.CurrencyPolicyService
	ExchangeRate                      *service.ExchangeRateService
	StorefrontMarket                  *service.StorefrontMarketService
	OpsDomainBinding                  *service.OpsDomainBindingService
	OpsDomainDiff                     *service.OpsDomainDiffService
	OpsDomainPreview                  *service.OpsDomainPreviewService
	OpsDomainSync                     *service.OpsDomainSyncService
	OpsConnector                      *service.OpsConnectorService
	OpsConnectorOAuth                 *service.OpsConnectorOAuthService
	OpsVPSBinding                     *service.OpsVPSBindingService
	OpsProjectBinding                 *service.OpsProjectBindingService
	OpsNetworkSummary                 *service.OpsNetworkSummaryService
	OpsHostingerSync                  *service.OpsHostingerSyncService
	OpsDeploymentPreflight            *service.OpsDeploymentPreflightService
	OpsDeploymentHealthCheck          *service.OpsDeploymentHealthCheckService
	OpsCloudflareCachePurge           *service.OpsCloudflareCachePurgeService
	OpsDeploymentWorkflow             *service.OpsDeploymentWorkflowService
	OpsOverview                       *service.OpsOverviewService
	StorefrontContext                 *service.StorefrontContextService
	GoogleMerchant                    *service.GoogleMerchantService
	SocialOAuth                       *service.SocialOAuthService
	ShowcaseUploadProtection          *service.ShowcaseUploadProtectionService
	ShowcaseUploadEligibility         *service.ShowcaseUploadEligibilityService
	PublicUploadAccess                *service.PublicUploadAccessService
	VisualShowcase                    *service.VisualShowcaseService
}

func NewDependencies(db *gorm.DB, redisCache *cache.RedisCache, cfg *config.Config) (*Dependencies, error) {
	fitmentFrameHubSpecificationRepository := repository.NewFitmentFrameHubSpecificationRepository(db)
	fitmentForkHubSpecificationRepository := repository.NewFitmentForkHubSpecificationRepository(db)
	fitmentHubSpecificationRepository := repository.NewFitmentHubSpecificationRepository(db, fitmentFrameHubSpecificationRepository)
	fitmentHubSpecificationRepository.ConfigureForkHubSpecificationRepository(fitmentForkHubSpecificationRepository)
	repos := Repositories{
		User:                         repository.NewUserRepository(db),
		Post:                         repository.NewPostRepository(db),
		StorefrontRouteCatalog:       repository.NewStorefrontRouteCatalogRepository(db),
		StorefrontRedirectRules:      repository.NewStorefrontRedirectRuleRepository(db),
		StorefrontURLIssues:          repository.NewStorefrontURLIssueRepository(db),
		PreflightContentLinks:        repository.NewPreflightContentLinkRepository(db),
		SiteQualityTargets:           repository.NewSiteQualityTargetRepository(db),
		SiteQualityJobs:              repository.NewSiteQualityJobRepository(db),
		SiteQualityRuns:              repository.NewSiteQualityRunRepository(db),
		SiteQualityFindings:          repository.NewSiteQualityFindingRepository(db),
		HotDataArchive:               repository.NewHotDataArchiveRepository(db),
		Product:                      repository.NewProductRepository(db),
		ProductProcurement:           repository.NewProductProcurementRepository(db),
		FrameFitmentEntry:            repository.NewFrameFitmentEntryRepository(db),
		ForkFitmentEntry:             repository.NewForkFitmentEntryRepository(db),
		FitmentFrameHubSpecification: fitmentFrameHubSpecificationRepository,
		FitmentForkHubSpecification:  fitmentForkHubSpecificationRepository,
		FitmentHubSpecification:      fitmentHubSpecificationRepository,
		ProductProcurementCatalog:    repository.NewProductProcurementCatalogRepository(db),
		ProductProfitCalculation:     repository.NewProductProfitCalculationRepository(db),
		ProductCategory:              repository.NewProductCategoryRepository(db),
		ProductBrand:                 repository.NewProductBrandRepository(db),
		ProductInformationTemplate:   repository.NewProductInformationTemplateRepository(db),
		CustomsClassification:        repository.NewCustomsClassificationRepository(db),
		Cart:                         repository.NewCartRepository(db),
		Setting:                      repository.NewSettingRepository(db),
		FAQ:                          repository.NewFAQRepository(db),
		Order:                        repository.NewOrderRepository(db),
		AfterSales:                   repository.NewAfterSalesCaseRepository(db),
		AfterSalesRefundReview:       repository.NewAfterSalesRefundReviewRepository(db),
		OrderIdempotency:             repository.NewOrderIdempotencyRepository(db),
		OrderPolicyDisclosure:        repository.NewOrderPolicyDisclosureRepository(db),
		OrderAttribution:             repository.NewOrderAttributionRepository(db),
		Payment:                      repository.NewPaymentRepository(db),
		PaymentRisk:                  repository.NewPaymentRiskRepository(db),
		PaymentProtection:            repository.NewPaymentProtectionRepository(db),
		PaymentRefundReview:          repository.NewPaymentRefundRecommendationRepository(db),
		PaymentRefundExec:            repository.NewPaymentRefundExecutionRepository(db),
		ExchangeRate:                 repository.NewExchangeRateRepository(db),
		Shipping:                     repository.NewShippingRepository(db),
		Coupon:                       repository.NewCouponRepository(db),
		Loyalty:                      repository.NewLoyaltyRepository(db),
		LoyaltyProgram:               repository.NewLoyaltyProgramRepository(db),
		GiftCardRedemption:           repository.NewGiftCardRedemptionRepository(db),
		Review:                       repository.NewReviewRepository(db),
		Ticket:                       repository.NewTicketRepository(db),
		Gallery:                      repository.NewGalleryRepository(db),
		Media:                        repository.NewMediaRepository(db),
		SiteLogo:                     repository.NewSiteLogoRepository(db),
		MediaDerivativePresets:       repository.NewMediaDerivativePresetRepository(db),
		MediaDerivativeRebuildJobs:   repository.NewMediaDerivativeRebuildJobRepository(db),
		StorefrontMarket:             repository.NewStorefrontMarketRepository(db),
		OpsDomainBinding:             repository.NewOpsDomainBindingRepository(db),
		OpsConnector:                 repository.NewOpsConnectorRepository(db),
		OpsConnectorOAuth:            repository.NewOpsConnectorOAuthRepository(db),
		OpsVPSBinding:                repository.NewOpsVPSBindingRepository(db),
		OpsProjectBinding:            repository.NewOpsProjectBindingRepository(db),
		OpsNetworkRule:               repository.NewOpsNetworkRuleRepository(db),
		OpsDeploymentWorkflow:        repository.NewOpsDeploymentWorkflowRepository(db),
		GoogleMerchant:               repository.NewGoogleMerchantRepository(db),
		SocialOAuth:                  repository.NewSocialOAuthRepository(db),
		Warranty:                     repository.NewWarrantyRepository(db),
		ShipmentRecord:               repository.NewShipmentRecordRepository(db),
		Audit:                        repository.NewAuditRepository(db),
		Showcase:                     repository.NewShowcaseRepository(db),
		VisualShowcase:               repository.NewVisualShowcaseRepository(db),
		Wishlist:                     repository.NewWishlistRepository(db),
		Feedback:                     repository.NewFeedbackRepository(db),
		SuggestionFeedback:           repository.NewSuggestionFeedbackRepository(db),
		Spoke:                        repository.NewSpokeRepository(db),
		QuickBuy:                     repository.NewQuickBuyRepository(db),
		SelectionAssistant:           repository.NewSelectionAssistantRepository(db),
		SelectionConfigurationKey:    repository.NewSelectionConfigurationKeyRepository(db),
		WheelsetFitQuestionnaire:     repository.NewWheelsetFitQuestionnaireRepository(db),
		Subscription:                 repository.NewSubscriptionRepository(db),
		EmailChallenge:               repository.NewEmailChallengeRepository(db),
		VisitorProfile:               repository.NewVisitorProfileRepository(db),
		RecommendationEvent:          repository.NewRecommendationEventRepository(db),
		VisitorRiskFact:              repository.NewVisitorRiskFactRepository(db),
		Outbox:                       repository.NewOutboxRepository(db),
	}
	repos.StorefrontRouteCatalog.ConfigureOutbox(repos.Outbox)
	if err := service.SeedDefaultMediaDerivativePresets(repos.MediaDerivativePresets); err != nil {
		return nil, fmt.Errorf("seed media derivative presets: %w", err)
	}

	storageConfig := storage.LoadConfigFromEnv()
	if _, configured := os.LookupEnv("STORAGE_BASE_URL"); !configured {
		storageConfig.BaseURL = cfg.Server.BaseURL
	}
	storageSvc, err := storage.NewStorageService(storageConfig)
	if err != nil {
		return nil, fmt.Errorf("initialize storage: %w", err)
	}
	siteLogoStorageConfig := storage.LoadSiteLogoConfigFromEnv(storageConfig)
	siteLogoStorageSvc, err := storage.NewStorageService(siteLogoStorageConfig)
	if err != nil {
		return nil, fmt.Errorf("initialize site logo storage: %w", err)
	}
	emailSvc, err := email.NewEmailService(email.LoadConfigFromEnv())
	if err != nil {
		return nil, fmt.Errorf("initialize email service: %w", err)
	}
	txManager := repository.NewTxManager(db, repos.Order, repos.Product, repos.Coupon, repos.Loyalty, repos.Payment, repos.Shipping)
	txManager.ConfigureGiftCardRedemptionRepository(repos.GiftCardRedemption)
	txManager.ConfigureLoyaltyProgramRepository(repos.LoyaltyProgram)
	txManager.ConfigureOutboxRepository(repos.Outbox)
	txManager.ConfigureProductBrandRepository(repos.ProductBrand)
	txManager.ConfigureOrderIdempotencyRepository(repos.OrderIdempotency)
	txManager.ConfigureOrderAttributionRepository(repos.OrderAttribution)
	txManager.ConfigureSettingRepository(repos.Setting)
	txManager.ConfigureExchangeRateRepository(repos.ExchangeRate)
	txManager.ConfigureOrderPolicyDisclosureRepository(repos.OrderPolicyDisclosure)
	txManager.ConfigurePaymentRefundRecommendationRepository(repos.PaymentRefundReview)
	txManager.ConfigurePaymentRefundExecutionRepository(repos.PaymentRefundExec)
	txManager.ConfigureAfterSalesRefundReviewRepository(repos.AfterSalesRefundReview)

	shippingService := service.NewShippingService(repos.Shipping, repos.Product)
	outboundHTTPResilience := newOutboundHTTPResilience(
		redisCache.Client(),
		cfg.OutboundHTTPResilience,
	)
	shippingService.ConfigureOutboundTrackingResilience(
		outboundHTTPResilience.retry,
		outboundHTTPResilience.breaker,
	)
	antiBotService := antibot.New(redisCache.Client(), cfg.AntiAbuse)
	antiFraudService := antifraud.New(redisCache.Client(), cfg.PaymentRisk)
	cardBINLimiter := cardtesting.New(redisCache.Client(), cfg.PaymentBINRateLimit)
	paymentGatewayCircuitBreaker := service.NewPaymentGatewayCircuitBreakerService(
		redisCache.Client(),
		cfg.PaymentGatewayCircuitBreaker,
	)
	orderAbuseService := orderabuse.New(redisCache.Client(), cfg.OrderAbuse)
	showcaseUploadProtectionService := service.NewShowcaseUploadProtectionService(redisCache.Client(), cfg.ShowcaseUploadProtection)
	showcaseUploadEligibilityService := service.NewShowcaseUploadEligibilityService(repos.Order)

	storefrontHTMLCacheInvalidator := service.NewStorefrontHTMLCacheInvalidatorFromEnv()
	storefrontContentReleaseNotifier := service.NewStorefrontContentReleaseNotifierFromEnv()
	settingService := service.NewSettingService(repos.Setting, redisCache, cfg.Cache.SettingsTTL)
	refundReturnPolicyService := service.NewRefundReturnPolicyService(repos.Setting)
	seoService := service.NewSEOService(settingService)
	postService := service.NewPostService(repos.Post, redisCache, cfg.Cache.PostTTL)
	productService := service.NewProductServiceWithCacheOptions(repos.Product, redisCache, cfg.Cache.ProductTTL, cfg.Cache.ProductLockTTL)
	productProcurementService := service.NewProductProcurementServiceWithProfitability(
		repos.ProductProcurement,
		repos.ProductProfitCalculation,
	)
	productProcurementService.ConfigureCatalogRepository(repos.ProductProcurementCatalog)
	fitmentHubSpecificationService := service.NewFitmentHubSpecificationService(
		repos.FitmentHubSpecification,
		repos.FitmentFrameHubSpecification,
	)
	fitmentHubSpecificationService.ConfigureForkHubSpecificationRepository(repos.FitmentForkHubSpecification)
	frameFitmentEntryService := service.NewFrameFitmentEntryService(
		repos.FrameFitmentEntry,
		repos.FitmentHubSpecification,
		repos.FitmentFrameHubSpecification,
	)
	forkFitmentEntryService := service.NewForkFitmentEntryService(
		repos.ForkFitmentEntry,
		repos.FitmentHubSpecification,
		repos.FitmentForkHubSpecification,
	)
	productProfitabilityService := service.NewProductProfitabilityServiceWithProcurement(
		repos.ProductProfitCalculation,
		repos.ProductProcurement,
	)
	productCategoryService := service.NewProductCategoryService(repos.ProductCategory, repos.Media)
	productBrandService := service.NewProductBrandService(repos.ProductBrand)
	productInformationTemplateService := service.NewProductInformationTemplateService(repos.ProductInformationTemplate)
	customsClassificationService := service.NewCustomsClassificationService(repos.CustomsClassification, settingService)
	merchantOutboxPublisher := service.NewMerchantOutboxPublisher(repos.Outbox)
	productCacheOutboxPublisher := service.NewProductCacheOutboxPublisher(repos.Outbox)
	seoResourceService := service.NewSEOResourceService(postService, productService, productCategoryService)
	analyticsService := service.NewAnalyticsService(settingService)
	currencyPolicyService := service.NewCurrencyPolicyService(repos.Setting)
	exchangeRateService := service.NewExchangeRateService(repos.ExchangeRate, repos.Setting)
	shippingService.ConfigureCurrencyPolicy(currencyPolicyService)
	storefrontMarketService := service.NewStorefrontMarketService(repos.StorefrontMarket)
	opsDomainBindingService := service.NewOpsDomainBindingService(repos.OpsDomainBinding, repos.OpsProjectBinding, repos.OpsConnector)
	opsDomainDiffService := service.NewOpsDomainDiffService(repos.OpsDomainBinding)
	opsDomainPreviewService := service.NewOpsDomainPreviewService(repos.OpsDomainBinding)
	opsConnectorService := service.NewOpsConnectorService(repos.OpsConnector)
	opsNetworkSummaryService := service.NewOpsNetworkSummaryService(
		repos.OpsNetworkRule,
		repos.OpsVPSBinding,
		repos.OpsProjectBinding,
		repos.OpsDomainBinding,
		repos.OpsConnector,
	)
	opsDomainSyncService := service.NewOpsDomainSyncService(repos.OpsDomainBinding, opsConnectorService)
	opsHostingerSyncService := service.NewOpsHostingerSyncService(
		repos.OpsVPSBinding,
		repos.OpsProjectBinding,
		opsConnectorService,
	)
	opsConnectorOAuthService := service.NewOpsConnectorOAuthService(
		repos.OpsConnectorOAuth,
		repos.OpsConnector,
		opsConnectorService,
		repos.OpsVPSBinding,
		repos.OpsProjectBinding,
		repos.OpsDomainBinding,
		opsHostingerSyncService,
		opsHostingerSyncService,
		opsDomainSyncService,
		cfg.Server.BaseURL,
	)
	opsDeploymentPreflightService := service.NewOpsDeploymentPreflightService(
		repos.OpsProjectBinding,
		repos.OpsVPSBinding,
		repos.OpsConnector,
		repos.OpsDomainBinding,
	)
	opsDeploymentHealthCheckService := service.NewOpsDeploymentHealthCheckService(repos.OpsDomainBinding)
	opsCloudflareCachePurgeService := service.NewOpsCloudflareCachePurgeService(repos.OpsDomainBinding, opsConnectorService)
	opsDeploymentWorkflowService := service.NewOpsDeploymentWorkflowService(
		repos.OpsDeploymentWorkflow,
		repos.OpsProjectBinding,
		opsDeploymentPreflightService,
	)
	opsDeploymentWorkflowService.ConfigureCachePurgeService(opsCloudflareCachePurgeService)
	opsDeploymentWorkflowService.ConfigureRollbackExecutor(service.NewOpsDeploymentSSHRollbackExecutorFromEnv())
	opsDeploymentWorkflowService.ConfigureProductionDependencies(
		repos.OpsVPSBinding,
		opsConnectorService,
		opsHostingerSyncService,
		opsDeploymentHealthCheckService,
	)
	storefrontBaseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("STOREFRONT_BASE_URL")), "/")
	if storefrontBaseURL == "" {
		storefrontBaseURL = strings.TrimRight(strings.TrimSpace(cfg.Server.BaseURL), "/")
	}
	lighthouseRunnerService := service.NewLighthouseRunnerService(
		repos.SiteQualityRuns,
		repos.SiteQualityFindings,
		service.LighthouseRunnerConfig{
			RunnerURL:         cfg.SiteQuality.RunnerURL,
			RunnerToken:       cfg.SiteQuality.RunnerToken,
			StorefrontBaseURL: storefrontBaseURL,
		},
	)
	lighthouseRunnerService.ConfigureJobRepository(repos.SiteQualityJobs)
	siteQualityEngineService := service.NewSiteQualityEngineService(
		repos.SiteQualityTargets,
		repos.SiteQualityJobs,
		repos.SiteQualityRuns,
		repos.SiteQualityFindings,
		repos.StorefrontRouteCatalog,
		lighthouseRunnerService,
		service.SiteQualityEngineConfig{
			BaseURL:                  storefrontBaseURL,
			WorkerEnabled:            cfg.Worker.SiteQualityEnabled,
			WorkerInterval:           time.Duration(cfg.Worker.SiteQualityDispatchIntervalSeconds) * time.Second,
			SampleCount:              cfg.Worker.SiteQualitySampleCount,
			RequiredConfirmations:    cfg.Worker.SiteQualityConfirmations,
			RequiredCleanEvaluations: cfg.Worker.SiteQualityCleanEvaluations,
			WorkerBatchLimit:         cfg.Worker.SiteQualityBatchLimit,
			LeaseTimeout:             time.Duration(cfg.Worker.SiteQualityLeaseTimeoutSeconds) * time.Second,
			ProviderConcurrency:      cfg.Worker.SiteQualityProviderConcurrency,
			ProviderRequestInterval:  time.Duration(cfg.Worker.SiteQualityProviderSpacingSeconds) * time.Second,
		},
	)
	mediaService := service.NewMediaService(repos.Media, storageSvc, settingService, storefrontBaseURL, cfg.MediaUpload.AccountStorageQuotaBytes)
	mediaService.ConfigureDerivativePresetRepository(repos.MediaDerivativePresets)
	mediaService.ConfigureDerivativeRebuildJobRepository(repos.MediaDerivativeRebuildJobs)
	siteLogoService := service.NewSiteLogoService(repos.SiteLogo, siteLogoStorageSvc, storefrontBaseURL)
	productService.ConfigureMediaService(mediaService)
	seoResourceService.ConfigureMediaService(mediaService)
	seoResourceService.ConfigureCanonicalBaseURL(storefrontBaseURL)
	storefrontRouteCatalogService := service.NewStorefrontRouteCatalogService(
		repos.StorefrontRouteCatalog,
		postService,
		productService,
		storefrontBaseURL,
	)
	storefrontRedirectRuleService := service.NewStorefrontRedirectRuleService(
		repos.StorefrontRedirectRules,
		repos.StorefrontRouteCatalog,
	)
	storefrontURLIssueService := service.NewStorefrontURLIssueService(
		repos.StorefrontURLIssues,
		repos.StorefrontRouteCatalog,
		repos.StorefrontRedirectRules,
	)
	preflightContentLinkService := service.NewPreflightContentLinkService(
		repos.PreflightContentLinks,
		repos.StorefrontRouteCatalog,
		postService,
		service.PreflightContentLinkConfig{
			BaseURL: storefrontBaseURL,
		},
	)
	storefrontRouteCatalogService.ConfigureIssueReconciler(storefrontURLIssueService)
	googleMerchantService := service.NewGoogleMerchantService(
		repos.GoogleMerchant,
		repos.Product,
		cfg.GoogleMerchant,
		storefrontBaseURL,
	)
	googleMerchantService.ConfigureOutboundHTTPResilience(
		outboundHTTPResilience.retry,
		outboundHTTPResilience.breaker,
	)
	googleIndexingService, err := service.NewGoogleIndexingService(
		productService,
		cfg.GoogleIndexing,
		storefrontBaseURL,
	)
	if err != nil {
		return nil, fmt.Errorf("initialize Google Indexing service: %w", err)
	}
	googleIndexingService.ConfigureOutboundHTTPResilience(
		outboundHTTPResilience.retry,
		outboundHTTPResilience.breaker,
	)
	socialOAuthService := service.NewSocialOAuthService(repos.SocialOAuth, cfg.SocialOAuth)
	loyaltyProgramService := service.NewLoyaltyProgramService(repos.LoyaltyProgram)
	loyaltyProgramService.ConfigureCurrencyPolicy(currencyPolicyService)
	orderNumberGenerator, err := ordernumber.NewGeneratorWithPreviousSecret(
		cfg.OrderNumber.EffectiveSecret(cfg.JWT.Secret),
		cfg.OrderNumber.EffectivePreviousSecret(),
		cfg.OrderNumber.NodeID,
	)
	if err != nil {
		return nil, fmt.Errorf("initialize order number generator: %w", err)
	}
	afterSalesService := service.NewAfterSalesService(repos.AfterSales, repos.Order, repos.AfterSalesRefundReview)
	afterSalesService.ConfigureUserRepository(repos.User)
	afterSalesService.ConfigureTxManager(txManager)
	afterSalesService.ConfigureAttachmentStorage(storageSvc)
	hotDataArchiveService := service.NewHotDataArchiveService(
		repos.HotDataArchive,
		cfg.Worker.HotDataArchiveBatchLimit,
	)
	authService := service.NewAuthService(repos.User, cfg.JWT, cfg.OAuth)
	authService.ConfigureGoogleOAuthResilience(
		outboundHTTPResilience.retry,
		outboundHTTPResilience.breaker,
	)

	services := Services{
		Auth:                              authService,
		AdminAccountMaintenance:           service.NewAdminAccountMaintenanceService(db),
		Post:                              postService,
		Product:                           productService,
		ProductProcurement:                productProcurementService,
		FrameFitmentEntry:                 frameFitmentEntryService,
		ForkFitmentEntry:                  forkFitmentEntryService,
		FitmentHubSpecification:           fitmentHubSpecificationService,
		ProductProfitability:              productProfitabilityService,
		ProductCategory:                   productCategoryService,
		ProductBrand:                      productBrandService,
		ProductInformationTemplate:        productInformationTemplateService,
		CustomsClassification:             customsClassificationService,
		Cart:                              service.NewCartService(repos.Cart, repos.Product),
		Setting:                           settingService,
		WebsiteProfile:                    service.NewWebsiteProfileService(settingService),
		WebsiteName:                       service.NewWebsiteNameService(settingService),
		RefundReturnPolicy:                refundReturnPolicyService,
		PayPalDisputeInvoiceSellerProfile: service.NewPayPalDisputeInvoiceSellerProfileService(settingService),
		SEO:                               seoService,
		SEOResources:                      seoResourceService,
		GoogleIndexing:                    googleIndexingService,
		Analytics:                         analyticsService,
		CurrencyPolicy:                    currencyPolicyService,
		ExchangeRate:                      exchangeRateService,
		StorefrontMarket:                  storefrontMarketService,
		OpsDomainBinding:                  opsDomainBindingService,
		OpsDomainDiff:                     opsDomainDiffService,
		OpsDomainPreview:                  opsDomainPreviewService,
		OpsDomainSync:                     opsDomainSyncService,
		OpsConnector:                      opsConnectorService,
		OpsConnectorOAuth:                 opsConnectorOAuthService,
		OpsVPSBinding:                     service.NewOpsVPSBindingService(repos.OpsVPSBinding, repos.OpsConnector, repos.OpsProjectBinding),
		OpsProjectBinding:                 service.NewOpsProjectBindingService(repos.OpsProjectBinding, repos.OpsVPSBinding, repos.OpsConnector),
		OpsNetworkSummary:                 opsNetworkSummaryService,
		OpsHostingerSync:                  opsHostingerSyncService,
		OpsDeploymentPreflight:            opsDeploymentPreflightService,
		OpsDeploymentHealthCheck:          opsDeploymentHealthCheckService,
		OpsCloudflareCachePurge:           opsCloudflareCachePurgeService,
		OpsDeploymentWorkflow:             opsDeploymentWorkflowService,
		StorefrontContext:                 service.NewStorefrontContextServiceWithMarkets(currencyPolicyService, storefrontMarketService),
		GoogleMerchant:                    googleMerchantService,
		SocialOAuth:                       socialOAuthService,
		FAQ:                               service.NewFAQService(repos.FAQ, storageSvc),
		Gallery:                           service.NewGalleryService(repos.Gallery, repos.Media),
		Media:                             mediaService,
		SiteLogo:                          siteLogoService,
		Warranty:                          service.NewWarrantyService(repos.Warranty, repos.Order),
		ShipmentRecord:                    service.NewShipmentRecordService(repos.ShipmentRecord),
		Checkout:                          service.NewCheckoutService(repos.Product, repos.Coupon, repos.Payment, repos.Loyalty, shippingService),
		AfterSales:                        afterSalesService,
		Marketing:                         service.NewMarketingService(txManager, repos.Coupon, repos.Loyalty, settingService),
		LoyaltyProgram:                    loyaltyProgramService,
		Review:                            service.NewReviewService(repos.Review),
		ReviewModeration:                  service.NewReviewModerationService(repos.Review),
		Ticket:                            service.NewTicketService(repos.Ticket, repos.User, repos.FAQ),
		CustomerServiceEvents:             service.NewCustomerServiceEventHub(),
		Subscription:                      service.NewSubscriptionService(repos.Subscription),
		Sitemap:                           service.NewSitemapService(repos.Post, cfg.Server.BaseURL),
		StorefrontRouteCatalog:            storefrontRouteCatalogService,
		StorefrontRedirectRules:           storefrontRedirectRuleService,
		StorefrontURLIssues:               storefrontURLIssueService,
		PreflightContentLinks:             preflightContentLinkService,
		LighthouseRunner:                  lighthouseRunnerService,
		SiteQualityEngine:                 siteQualityEngineService,
		HotDataArchive:                    hotDataArchiveService,
		Showcase:                          service.NewShowcaseService(repos.Showcase, storageSvc),
		VisualShowcase:                    service.NewVisualShowcaseService(repos.VisualShowcase, storageSvc),
		ShowcaseUploadProtection:          showcaseUploadProtectionService,
		ShowcaseUploadEligibility:         showcaseUploadEligibilityService,
		Wishlist:                          service.NewWishlistService(repos.Wishlist, repos.Product),
		Feedback:                          service.NewFeedbackService(repos.Feedback),
		SuggestionFeedback: service.NewSuggestionFeedbackService(
			repos.SuggestionFeedback,
		),
		User:      service.NewUserService(repos.User),
		Dashboard: service.NewDashboardService(repos.Order, repos.User, repos.Subscription),
		Audit:     service.NewAuditService(repos.Audit),
		OpsOverview: service.NewOpsOverviewService(
			repos.OpsDomainBinding,
			repos.OpsConnector,
			repos.OpsVPSBinding,
			repos.OpsProjectBinding,
			service.NewAuditService(repos.Audit),
		),
		Shipping:                  shippingService,
		Spoke:                     service.NewSpokeService(repos.Spoke),
		QuickBuy:                  service.NewQuickBuyService(repos.QuickBuy, repos.Product, repos.ProductCategory),
		SelectionAssistant:        service.NewSelectionAssistantService(repos.SelectionAssistant),
		SelectionConfigurationKey: service.NewSelectionConfigurationKeyService(repos.SelectionConfigurationKey),
		WheelsetFitQuestionnaire: service.NewWheelsetFitQuestionnaireService(
			repos.WheelsetFitQuestionnaire,
			repos.SelectionConfigurationKey,
		),
		VisitorProfile: service.NewVisitorProfileService(
			repos.VisitorProfile,
		),
		BehaviorEvents: service.NewBehaviorEventService(
			repos.RecommendationEvent,
			cfg.BehaviorEvents,
		),
		VisitorRisk: service.NewVisitorRiskService(
			repos.VisitorRiskFact,
			cfg.VisitorRisk,
			cfg.JWT.Secret,
		),
		PaymentRiskMonitoring: service.NewPaymentRiskMonitoringService(
			repos.PaymentRisk,
			cfg.PaymentRiskMonitoring,
		),
		PaymentProtection: service.NewPaymentProtectionService(
			repos.PaymentProtection,
			cfg.PaymentProtection,
		),
		PaymentRefundReview: service.NewPaymentRefundRecommendationService(repos.PaymentRefundReview, txManager),
		Outbox:              service.NewOutboxService(repos.Outbox),
	}
	services.Recommendations = service.NewRecommendationService(services.Product, repos.RecommendationEvent)
	services.Recommendations.ConfigureMediaService(services.Media)
	services.Recommendations.ConfigureReviewService(services.Review)
	services.QuickBuy.ConfigureMediaService(services.Media)
	services.GoogleMerchant.ConfigureMediaService(services.Media)
	services.Ticket.ConfigureCustomerServiceRealtimeOutbox(repos.Outbox)
	services.CustomerServiceAvatar = service.NewCustomerServiceAvatarService(repos.User, storageSvc, repos.Outbox)
	services.PublicUploadAccess = service.NewPublicUploadAccessService(services.Media, services.Showcase, services.CustomerServiceAvatar)
	services.PublicUploadAccess.ConfigureSiteLogoService(services.SiteLogo)
	services.FAQ.ConfigureMediaService(services.Media)
	services.Review.ConfigureMediaService(services.Media)
	services.Showcase.ConfigureUploadEligibility(services.ShowcaseUploadEligibility)
	if cfg.ShowcaseUploadProtection.Enabled {
		services.Showcase.ConfigurePendingSubmissionLimit(cfg.ShowcaseUploadProtection.MaxPendingSubmissionsPerUser)
	}
	services.Marketing.ConfigureLoyaltyProgram(loyaltyProgramService)
	services.Marketing.ConfigureGiftCardRedemptions(repos.GiftCardRedemption)
	services.Marketing.ConfigureCurrencyPolicy(currencyPolicyService)
	services.Checkout.ConfigureLoyaltyProgram(loyaltyProgramService)
	services.Checkout.ConfigureCurrencyPolicy(currencyPolicyService)
	services.Checkout.ConfigureExchangeRateRepository(repos.ExchangeRate)
	services.Product.ConfigureCurrencyPolicy(currencyPolicyService)
	services.Product.ConfigureInformationTemplateRepository(repos.ProductInformationTemplate)
	services.Product.ConfigureProductBrandRepository(repos.ProductBrand)
	services.Product.ConfigureCustomsClassificationRepository(repos.CustomsClassification)
	services.Product.ConfigureProductCategoryRepository(repos.ProductCategory)
	services.Product.ConfigureMerchantEventPublisher(merchantOutboxPublisher)
	services.Product.ConfigureProductCacheEventPublisher(productCacheOutboxPublisher)
	services.ProductBrand.ConfigureProductDependencies(repos.Product, productCacheOutboxPublisher, merchantOutboxPublisher)
	services.Product.ConfigureTxManager(txManager)
	services.ProductBrand.ConfigureTxManager(txManager)
	services.ProductInformationTemplate.ConfigureProductCacheInvalidator(services.Product)
	services.ProductInformationTemplate.ConfigureProductCacheEventPublisher(productCacheOutboxPublisher)
	services.GoogleMerchant.ConfigureMerchantEventPublisher(merchantOutboxPublisher)
	services.ExchangeRate.ConfigureCurrencyPolicy(currencyPolicyService)
	services.ExchangeRate.ConfigureStorefrontMarkets(storefrontMarketService)
	services.ExchangeRate.ConfigureProductService(productService)
	services.ExchangeRate.ConfigureShippingService(shippingService)
	services.Warranty.ConfigureEmailChallenges(repos.EmailChallenge, cfg.JWT.Secret, emailSvc)
	services.Warranty.ConfigureEmailBaseURL(storefrontBaseURL)
	services.Subscription.ConfigureEmailChallenges(repos.EmailChallenge, cfg.JWT.Secret, emailSvc)
	services.Subscription.ConfigureEmailBaseURL(storefrontBaseURL)
	services.Product.SetStorefrontHTMLCacheInvalidator(storefrontHTMLCacheInvalidator)
	services.ProductBrand.SetStorefrontHTMLCacheInvalidator(storefrontHTMLCacheInvalidator)
	services.Post.SetStorefrontHTMLCacheInvalidator(storefrontHTMLCacheInvalidator)
	services.FAQ.SetStorefrontHTMLCacheInvalidator(storefrontHTMLCacheInvalidator)
	services.FAQ.SetStorefrontContentReleaseNotifier(storefrontContentReleaseNotifier)
	services.SEO.SetStorefrontHTMLCacheInvalidator(storefrontHTMLCacheInvalidator)
	services.AdminSettings = service.NewAdminSettingsService(services.Setting)
	services.AdminPublicChat = service.NewAdminPublicChatAgentService(repos.User)
	services.CustomerServiceContext = service.NewCustomerServiceContextService(
		services.Ticket,
		repos.User,
		repos.Cart,
		repos.Wishlist,
		repos.Order,
		repos.Loyalty,
		services.VisitorProfile,
	)
	services.CustomerServiceContext.ConfigureMediaService(services.Media)
	services.CustomerServiceAnalytics = service.NewCustomerServiceAnalyticsService(
		services.Ticket,
		services.CustomerServiceContext,
		repos.Order,
	)
	services.Order = service.NewOrderService(
		txManager,
		repos.Order,
		services.Checkout,
		shippingService,
		orderNumberGenerator,
	)
	services.Order.ConfigureProductCacheInvalidator(services.Product)
	services.Order.ConfigureProductCacheEventPublisher(productCacheOutboxPublisher)
	services.Order.ConfigureRefundReturnPolicy(services.RefundReturnPolicy)
	services.Payment = service.NewPaymentService(txManager, repos.Payment)
	services.Payment.ConfigureProductCacheInvalidator(services.Product)
	services.Payment.ConfigureProductCacheEventPublisher(productCacheOutboxPublisher)
	services.Payment.ConfigureRisk(repos.Order, antiFraudService)
	services.Payment.ConfigureEvidenceSources(repos.Order, repos.Shipping, repos.Ticket)
	services.Payment.ConfigurePolicyDisclosureRepository(repos.OrderPolicyDisclosure)
	services.Payment.ConfigurePayPalDisputeEvidenceDocumentStorage(storageSvc)
	services.Payment.ConfigurePayPalDisputeInvoiceSellerProfileProvider(services.PayPalDisputeInvoiceSellerProfile)
	services.Payment.ConfigurePayPalDisputeInvoiceOptions(service.PayPalDisputeInvoiceOptions{
		AutoAttachPDF: envBoolDefault("PAYPAL_DISPUTE_AUTO_ATTACH_INVOICE_PDF", true),
	})
	services.Order.ConfigurePaymentDisputeAnalysis(services.Payment)
	services.Order.ConfigureAdminEmailSender(emailSvc)
	services.PaymentThreeDS = service.NewPaymentThreeDSPolicyService(
		repos.Order,
		services.VisitorRisk,
		antiFraudService,
		cfg.PaymentThreeDS,
	)
	services.PaymentThreeDS.ConfigureRiskMonitoring(services.PaymentRiskMonitoring)
	services.PaymentThreeDS.ConfigurePaymentProtection(services.PaymentProtection)

	var customerServiceRealtimeRelay *service.CustomerServiceRealtimeRelay
	if cfg.CustomerServiceRealtime.Enabled {
		var err error
		customerServiceRealtimeRelay, err = service.NewCustomerServiceRealtimeRelay(
			redisCache.Client(),
			services.CustomerServiceEvents,
			service.CustomerServiceRealtimeRelayConfig{
				Stream:         cfg.CustomerServiceRealtime.Stream,
				StreamMaxLen:   int64(cfg.CustomerServiceRealtime.StreamMaxLen),
				ReplayLimit:    cfg.CustomerServiceRealtime.ReplayLimit,
				ConsumerBlock:  time.Duration(cfg.CustomerServiceRealtime.ConsumerBlockSeconds) * time.Second,
				DedupRetention: time.Duration(cfg.CustomerServiceRealtime.DedupRetentionSeconds) * time.Second,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("initialize customer-service realtime relay: %w", err)
		}
		services.CustomerServiceEvents.ConfigureReplayProvider(customerServiceRealtimeRelay)
	}

	orderPaidWebhookHandler := service.NewOrderPaidOutboxWebhookHandlerFromEnvWithResilience(
		outboundHTTPResilience.retry,
		outboundHTTPResilience.breaker,
	)
	if orderPaidWebhookHandler.Configured() {
		services.Outbox.RegisterHandler(outbox.EventTypeOrderPaid, orderPaidWebhookHandler.Handle)
	}
	verifiedConversionWebhookHandler := service.NewVerifiedConversionOutboxWebhookHandlerFromEnvWithResilience(
		outboundHTTPResilience.retry,
		outboundHTTPResilience.breaker,
	)
	if verifiedConversionWebhookHandler.Configured() {
		services.Outbox.RegisterHandler(outbox.EventTypeVerifiedConversion, verifiedConversionWebhookHandler.Handle)
	}
	paymentRiskAlertWebhookHandler := service.NewPaymentRiskAlertOutboxWebhookHandlerFromEnvWithResilience(
		outboundHTTPResilience.retry,
		outboundHTTPResilience.breaker,
	)
	if cfg.PaymentRiskMonitoring.AlertEnabled && paymentRiskAlertWebhookHandler.Configured() {
		services.PaymentRiskMonitoring.ConfigureAlerting(true)
		services.Outbox.RegisterHandler(outbox.EventTypePaymentRiskLevelChanged, paymentRiskAlertWebhookHandler.Handle)
	}
	merchantOutboxHandler := service.NewGoogleMerchantOutboxHandler(services.GoogleMerchant)
	services.Outbox.RegisterHandler(outbox.EventTypeMerchantProductUpsert, merchantOutboxHandler.Handle)
	services.Outbox.RegisterHandler(outbox.EventTypeMerchantProductWithdraw, merchantOutboxHandler.Handle)
	services.Outbox.RegisterHandler(outbox.EventTypeMerchantOfferRevalidate, merchantOutboxHandler.Handle)
	productCacheOutboxHandler := service.NewProductCacheOutboxHandler(services.Product)
	services.Outbox.RegisterHandler(outbox.EventTypeProductCacheInvalidate, productCacheOutboxHandler.Handle)
	customerServiceRealtimeOutboxHandler := service.NewCustomerServiceRealtimeOutboxHandler(services.CustomerServiceEvents, customerServiceRealtimeRelay)
	services.Outbox.RegisterHandler(outbox.EventTypeCustomerServiceRealtime, customerServiceRealtimeOutboxHandler.Handle)
	customerServiceAvatarCleanupHandler := service.NewCustomerServiceAvatarCleanupHandler(repos.User, storageSvc)
	services.Outbox.RegisterHandler(outbox.EventTypeCustomerServiceAvatarCleanup, customerServiceAvatarCleanupHandler.Handle)
	services.Outbox.RegisterHandler(
		outbox.EventTypeStorefrontRouteCatalogChanged,
		service.NewSiteQualityRouteCatalogOutboxHandler(services.SiteQualityEngine).Handle,
	)

	return &Dependencies{
		Repositories:                 repos,
		Services:                     services,
		Storage:                      storageSvc,
		SiteLogoStorage:              siteLogoStorageSvc,
		AntiBot:                      antiBotService,
		AntiFraud:                    antiFraudService,
		CardBINLimiter:               cardBINLimiter,
		PaymentGatewayCircuitBreaker: paymentGatewayCircuitBreaker,
		OrderAbuse:                   orderAbuseService,
		RedisClient:                  redisCache.Client(),
		CustomerServiceRealtimeRelay: customerServiceRealtimeRelay,
	}, nil
}

func envBoolDefault(key string, fallback bool) bool {
	value, configured := os.LookupEnv(key)
	if !configured || strings.TrimSpace(value) == "" {
		return fallback
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on", "enabled":
		return true
	case "0", "false", "no", "n", "off", "disabled":
		return false
	default:
		return fallback
	}
}
