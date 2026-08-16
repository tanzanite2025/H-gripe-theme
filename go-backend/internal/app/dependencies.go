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
	AntiBot                      *antibot.Service
	AntiFraud                    *antifraud.Service
	CardBINLimiter               *cardtesting.Service
	PaymentGatewayCircuitBreaker *service.PaymentGatewayCircuitBreakerService
	OrderAbuse                   *orderabuse.Service
	RedisClient                  *redis.Client
	CustomerServiceRealtimeRelay *service.CustomerServiceRealtimeRelay
}

type Repositories struct {
	User                       *repository.UserRepository
	Post                       *repository.PostRepository
	StorefrontRouteCatalog     *repository.StorefrontRouteCatalogRepository
	Product                    *repository.ProductRepository
	ProductCategory            *repository.ProductCategoryRepository
	ProductBrand               *repository.ProductBrandRepository
	ProductInformationTemplate *repository.ProductInformationTemplateRepository
	CustomsClassification      *repository.CustomsClassificationRepository
	Cart                       *repository.CartRepository
	Setting                    *repository.SettingRepository
	FAQ                        *repository.FAQRepository
	Order                      *repository.OrderRepository
	OrderAttribution           *repository.OrderAttributionRepository
	Payment                    *repository.PaymentRepository
	PaymentRisk                *repository.PaymentRiskRepository
	PaymentProtection          *repository.PaymentProtectionRepository
	PaymentRefundReview        *repository.PaymentRefundRecommendationRepository
	PaymentRefundExec          *repository.PaymentRefundExecutionRepository
	ExchangeRate               *repository.ExchangeRateRepository
	Shipping                   *repository.ShippingRepository
	Coupon                     *repository.CouponRepository
	Loyalty                    *repository.LoyaltyRepository
	LoyaltyProgram             *repository.LoyaltyProgramRepository
	GiftCardRedemption         *repository.GiftCardRedemptionRepository
	Review                     *repository.ReviewRepository
	Ticket                     *repository.TicketRepository
	Gallery                    *repository.GalleryRepository
	Media                      *repository.MediaRepository
	StorefrontMarket           *repository.StorefrontMarketRepository
	OpsDomainBinding           *repository.OpsDomainBindingRepository
	OpsConnector               *repository.OpsConnectorRepository
	OpsConnectorOAuth          *repository.OpsConnectorOAuthRepository
	OpsVPSBinding              *repository.OpsVPSBindingRepository
	OpsProjectBinding          *repository.OpsProjectBindingRepository
	OpsDeploymentWorkflow      *repository.OpsDeploymentWorkflowRepository
	GoogleMerchant             *repository.GoogleMerchantRepository
	Registration               *repository.RegistrationRepository
	Audit                      *repository.AuditRepository
	Showcase                   *repository.ShowcaseRepository
	Wishlist                   *repository.WishlistRepository
	Feedback                   *repository.FeedbackRepository
	SuggestionFeedback         *repository.SuggestionFeedbackRepository
	Spoke                      *repository.SpokeRepository
	QuickBuy                   *repository.QuickBuyRepository
	SelectionAssistant         *repository.SelectionAssistantRepository
	WheelsetFitQuestionnaire   *repository.WheelsetFitQuestionnaireRepository
	Subscription               *repository.SubscriptionRepository
	EmailChallenge             *repository.EmailChallengeRepository
	VisitorProfile             *repository.VisitorProfileRepository
	RecommendationEvent        *repository.RecommendationEventRepository
	VisitorRiskFact            *repository.VisitorRiskFactRepository
	Outbox                     *repository.OutboxRepository
}

type Services struct {
	Auth                              *service.AuthService
	AdminAccountMaintenance           *service.AdminAccountMaintenanceService
	Post                              *service.PostService
	Product                           *service.ProductService
	ProductCategory                   *service.ProductCategoryService
	ProductBrand                      *service.ProductBrandService
	ProductInformationTemplate        *service.ProductInformationTemplateService
	CustomsClassification             *service.CustomsClassificationService
	Cart                              *service.CartService
	Setting                           *service.SettingService
	WebsiteProfile                    *service.WebsiteProfileService
	PayPalDisputeInvoiceSellerProfile *service.PayPalDisputeInvoiceSellerProfileService
	AdminSettings                     *service.AdminSettingsService
	SEO                               *service.SEOService
	SEOResources                      *service.SEOResourceService
	Analytics                         *service.AnalyticsService
	AdminPublicChat                   *service.AdminPublicChatAgentService
	CustomerServiceAvatar             *service.CustomerServiceAvatarService
	FAQ                               *service.FAQService
	Gallery                           *service.GalleryService
	Media                             *service.MediaService
	Registration                      *service.RegistrationService
	Checkout                          *service.CheckoutService
	Order                             *service.OrderService
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
	OpsHostingerSync                  *service.OpsHostingerSyncService
	OpsDeploymentPreflight            *service.OpsDeploymentPreflightService
	OpsDeploymentHealthCheck          *service.OpsDeploymentHealthCheckService
	OpsCloudflareCachePurge           *service.OpsCloudflareCachePurgeService
	OpsDeploymentWorkflow             *service.OpsDeploymentWorkflowService
	OpsOverview                       *service.OpsOverviewService
	StorefrontContext                 *service.StorefrontContextService
	GoogleMerchant                    *service.GoogleMerchantService
	ShowcaseUploadProtection          *service.ShowcaseUploadProtectionService
	ShowcaseUploadEligibility         *service.ShowcaseUploadEligibilityService
	PublicUploadAccess                *service.PublicUploadAccessService
}

func NewDependencies(db *gorm.DB, redisCache *cache.RedisCache, cfg *config.Config) (*Dependencies, error) {
	repos := Repositories{
		User:                       repository.NewUserRepository(db),
		Post:                       repository.NewPostRepository(db),
		StorefrontRouteCatalog:     repository.NewStorefrontRouteCatalogRepository(db),
		Product:                    repository.NewProductRepository(db),
		ProductCategory:            repository.NewProductCategoryRepository(db),
		ProductBrand:               repository.NewProductBrandRepository(db),
		ProductInformationTemplate: repository.NewProductInformationTemplateRepository(db),
		CustomsClassification:      repository.NewCustomsClassificationRepository(db),
		Cart:                       repository.NewCartRepository(db),
		Setting:                    repository.NewSettingRepository(db),
		FAQ:                        repository.NewFAQRepository(db),
		Order:                      repository.NewOrderRepository(db),
		OrderAttribution:           repository.NewOrderAttributionRepository(db),
		Payment:                    repository.NewPaymentRepository(db),
		PaymentRisk:                repository.NewPaymentRiskRepository(db),
		PaymentProtection:          repository.NewPaymentProtectionRepository(db),
		PaymentRefundReview:        repository.NewPaymentRefundRecommendationRepository(db),
		PaymentRefundExec:          repository.NewPaymentRefundExecutionRepository(db),
		ExchangeRate:               repository.NewExchangeRateRepository(db),
		Shipping:                   repository.NewShippingRepository(db),
		Coupon:                     repository.NewCouponRepository(db),
		Loyalty:                    repository.NewLoyaltyRepository(db),
		LoyaltyProgram:             repository.NewLoyaltyProgramRepository(db),
		GiftCardRedemption:         repository.NewGiftCardRedemptionRepository(db),
		Review:                     repository.NewReviewRepository(db),
		Ticket:                     repository.NewTicketRepository(db),
		Gallery:                    repository.NewGalleryRepository(db),
		Media:                      repository.NewMediaRepository(db),
		StorefrontMarket:           repository.NewStorefrontMarketRepository(db),
		OpsDomainBinding:           repository.NewOpsDomainBindingRepository(db),
		OpsConnector:               repository.NewOpsConnectorRepository(db),
		OpsConnectorOAuth:          repository.NewOpsConnectorOAuthRepository(db),
		OpsVPSBinding:              repository.NewOpsVPSBindingRepository(db),
		OpsProjectBinding:          repository.NewOpsProjectBindingRepository(db),
		OpsDeploymentWorkflow:      repository.NewOpsDeploymentWorkflowRepository(db),
		GoogleMerchant:             repository.NewGoogleMerchantRepository(db),
		Registration:               repository.NewRegistrationRepository(db),
		Audit:                      repository.NewAuditRepository(db),
		Showcase:                   repository.NewShowcaseRepository(db),
		Wishlist:                   repository.NewWishlistRepository(db),
		Feedback:                   repository.NewFeedbackRepository(db),
		SuggestionFeedback:         repository.NewSuggestionFeedbackRepository(db),
		Spoke:                      repository.NewSpokeRepository(db),
		QuickBuy:                   repository.NewQuickBuyRepository(db),
		SelectionAssistant:         repository.NewSelectionAssistantRepository(db),
		WheelsetFitQuestionnaire:   repository.NewWheelsetFitQuestionnaireRepository(db),
		Subscription:               repository.NewSubscriptionRepository(db),
		EmailChallenge:             repository.NewEmailChallengeRepository(db),
		VisitorProfile:             repository.NewVisitorProfileRepository(db),
		RecommendationEvent:        repository.NewRecommendationEventRepository(db),
		VisitorRiskFact:            repository.NewVisitorRiskFactRepository(db),
		Outbox:                     repository.NewOutboxRepository(db),
	}

	storageConfig := storage.LoadConfigFromEnv()
	if _, configured := os.LookupEnv("STORAGE_BASE_URL"); !configured {
		storageConfig.BaseURL = cfg.Server.BaseURL
	}
	storageSvc, err := storage.NewStorageService(storageConfig)
	if err != nil {
		return nil, fmt.Errorf("initialize storage: %w", err)
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
	txManager.ConfigureOrderAttributionRepository(repos.OrderAttribution)
	txManager.ConfigureSettingRepository(repos.Setting)
	txManager.ConfigureExchangeRateRepository(repos.ExchangeRate)
	txManager.ConfigurePaymentRefundRecommendationRepository(repos.PaymentRefundReview)
	txManager.ConfigurePaymentRefundExecutionRepository(repos.PaymentRefundExec)

	shippingService := service.NewShippingService(repos.Shipping, repos.Product)
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
	seoService := service.NewSEOService(settingService)
	postService := service.NewPostService(repos.Post, redisCache, cfg.Cache.PostTTL)
	productService := service.NewProductServiceWithCacheOptions(repos.Product, redisCache, cfg.Cache.ProductTTL, cfg.Cache.ProductLockTTL)
	productCategoryService := service.NewProductCategoryService(repos.ProductCategory, repos.Media)
	productBrandService := service.NewProductBrandService(repos.ProductBrand)
	productInformationTemplateService := service.NewProductInformationTemplateService(repos.ProductInformationTemplate)
	customsClassificationService := service.NewCustomsClassificationService(repos.CustomsClassification, settingService)
	merchantOutboxPublisher := service.NewMerchantOutboxPublisher(repos.Outbox)
	productCacheOutboxPublisher := service.NewProductCacheOutboxPublisher(repos.Outbox)
	seoResourceService := service.NewSEOResourceService(postService, productService, settingService)
	analyticsService := service.NewAnalyticsService(settingService)
	currencyPolicyService := service.NewCurrencyPolicyService(repos.Setting)
	exchangeRateService := service.NewExchangeRateService(repos.ExchangeRate, repos.Setting)
	storefrontMarketService := service.NewStorefrontMarketService(repos.StorefrontMarket)
	opsDomainBindingService := service.NewOpsDomainBindingService(repos.OpsDomainBinding, repos.OpsProjectBinding, repos.OpsConnector)
	opsDomainDiffService := service.NewOpsDomainDiffService(repos.OpsDomainBinding)
	opsDomainPreviewService := service.NewOpsDomainPreviewService(repos.OpsDomainBinding)
	opsConnectorService := service.NewOpsConnectorService(repos.OpsConnector)
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
	mediaService := service.NewMediaService(repos.Media, storageSvc, settingService, storefrontBaseURL, cfg.MediaUpload.AccountStorageQuotaBytes)
	productService.ConfigureMediaService(mediaService)
	seoResourceService.ConfigureCanonicalBaseURL(storefrontBaseURL)
	storefrontRouteCatalogService := service.NewStorefrontRouteCatalogService(
		repos.StorefrontRouteCatalog,
		postService,
		productService,
		storefrontBaseURL,
	)
	googleMerchantService := service.NewGoogleMerchantService(
		repos.GoogleMerchant,
		repos.Product,
		cfg.GoogleMerchant,
		storefrontBaseURL,
	)
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

	services := Services{
		Auth:                              service.NewAuthService(repos.User, cfg.JWT, cfg.OAuth),
		AdminAccountMaintenance:           service.NewAdminAccountMaintenanceService(db),
		Post:                              postService,
		Product:                           productService,
		ProductCategory:                   productCategoryService,
		ProductBrand:                      productBrandService,
		ProductInformationTemplate:        productInformationTemplateService,
		CustomsClassification:             customsClassificationService,
		Cart:                              service.NewCartService(repos.Cart, repos.Product),
		Setting:                           settingService,
		WebsiteProfile:                    service.NewWebsiteProfileService(settingService),
		PayPalDisputeInvoiceSellerProfile: service.NewPayPalDisputeInvoiceSellerProfileService(settingService),
		SEO:                               seoService,
		SEOResources:                      seoResourceService,
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
		OpsHostingerSync:                  opsHostingerSyncService,
		OpsDeploymentPreflight:            opsDeploymentPreflightService,
		OpsDeploymentHealthCheck:          opsDeploymentHealthCheckService,
		OpsCloudflareCachePurge:           opsCloudflareCachePurgeService,
		OpsDeploymentWorkflow:             opsDeploymentWorkflowService,
		StorefrontContext:                 service.NewStorefrontContextServiceWithMarkets(currencyPolicyService, storefrontMarketService),
		GoogleMerchant:                    googleMerchantService,
		FAQ:                               service.NewFAQService(repos.FAQ, storageSvc),
		Gallery:                           service.NewGalleryService(repos.Gallery, repos.Media),
		Media:                             mediaService,
		Registration:                      service.NewRegistrationService(repos.Registration, repos.Product, repos.Order),
		Checkout:                          service.NewCheckoutService(repos.Product, repos.Coupon, repos.Payment, repos.Loyalty, shippingService),
		Marketing:                         service.NewMarketingService(txManager, repos.Coupon, repos.Loyalty, settingService),
		LoyaltyProgram:                    loyaltyProgramService,
		Review:                            service.NewReviewService(repos.Review),
		ReviewModeration:                  service.NewReviewModerationService(repos.Review),
		Ticket:                            service.NewTicketService(repos.Ticket, repos.User, repos.FAQ),
		CustomerServiceEvents:             service.NewCustomerServiceEventHub(),
		Subscription:                      service.NewSubscriptionService(repos.Subscription),
		Sitemap:                           service.NewSitemapService(repos.Post, cfg.Server.BaseURL),
		StorefrontRouteCatalog:            storefrontRouteCatalogService,
		Showcase:                          service.NewShowcaseService(repos.Showcase, storageSvc),
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
		Shipping:           shippingService,
		Spoke:              service.NewSpokeService(repos.Spoke),
		QuickBuy:           service.NewQuickBuyService(repos.QuickBuy, repos.Product, repos.ProductCategory),
		SelectionAssistant: service.NewSelectionAssistantService(repos.SelectionAssistant),
		WheelsetFitQuestionnaire: service.NewWheelsetFitQuestionnaireService(
			repos.WheelsetFitQuestionnaire,
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
	services.Ticket.ConfigureCustomerServiceRealtimeOutbox(repos.Outbox)
	services.CustomerServiceAvatar = service.NewCustomerServiceAvatarService(repos.User, storageSvc, repos.Outbox)
	services.PublicUploadAccess = service.NewPublicUploadAccessService(services.Media, services.Showcase, services.CustomerServiceAvatar)
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
	services.Registration.ConfigureEmailChallenges(repos.EmailChallenge, cfg.JWT.Secret, emailSvc)
	services.Registration.ConfigureEmailBaseURL(storefrontBaseURL)
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
	services.Payment = service.NewPaymentService(txManager, repos.Payment)
	services.Payment.ConfigureProductCacheInvalidator(services.Product)
	services.Payment.ConfigureProductCacheEventPublisher(productCacheOutboxPublisher)
	services.Payment.ConfigureRisk(repos.Order, antiFraudService)
	services.Payment.ConfigureEvidenceSources(repos.Order, repos.Shipping, repos.Ticket)
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

	orderPaidWebhookHandler := service.NewOrderPaidOutboxWebhookHandlerFromEnv()
	if orderPaidWebhookHandler.Configured() {
		services.Outbox.RegisterHandler(outbox.EventTypeOrderPaid, orderPaidWebhookHandler.Handle)
	}
	verifiedConversionWebhookHandler := service.NewVerifiedConversionOutboxWebhookHandlerFromEnv()
	if verifiedConversionWebhookHandler.Configured() {
		services.Outbox.RegisterHandler(outbox.EventTypeVerifiedConversion, verifiedConversionWebhookHandler.Handle)
	}
	paymentRiskAlertWebhookHandler := service.NewPaymentRiskAlertOutboxWebhookHandlerFromEnv()
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

	return &Dependencies{
		Repositories:                 repos,
		Services:                     services,
		Storage:                      storageSvc,
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
