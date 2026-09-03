package app

import (
	"fmt"
	"os"

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

	"gorm.io/gorm"
)

type dependencySupport struct {
	StorageSvc                          storage.StorageService
	SiteLogoStorageSvc                  storage.StorageService
	EmailSvc                            email.EmailService
	TxManager                           *repository.TxManager
	ShippingService                     *service.ShippingService
	OutboundHTTPResilience              outboundHTTPResilience
	AntiBotService                      *antibot.Service
	AntiFraudService                    *antifraud.Service
	CardBINLimiter                      *cardtesting.Service
	PaymentGatewayCircuitBreaker        *service.PaymentGatewayCircuitBreakerService
	OrderAbuseService                   *orderabuse.Service
	StorefrontBaseURL                   string
	StorefrontInternalOrigin            string
	SiteQualityTargetOrigin             string
	StorefrontHTMLCacheInvalidator      *service.StorefrontHTMLCacheInvalidator
	StorefrontContentReleaseNotifier    *service.StorefrontContentReleaseNotifier
	UGCShowcaseUploadProtectionService  *service.UGCShowcaseUploadProtectionService
	UGCShowcaseUploadEligibilityService *service.UGCShowcaseUploadEligibilityService
	OrderNumberGenerator                *ordernumber.Generator
}

func newDependencySupport(
	db *gorm.DB,
	redisCache *cache.RedisCache,
	cfg *config.Config,
	repos Repositories,
) (*dependencySupport, error) {
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
	shippingService.ConfigureOrderRepository(repos.Order)
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
	ugcShowcaseUploadProtectionService := service.NewUGCShowcaseUploadProtectionService(redisCache.Client(), cfg.ShowcaseUploadProtection)
	ugcShowcaseUploadEligibilityService := service.NewUGCShowcaseUploadEligibilityService(repos.Order)

	storefrontBaseURL, storefrontInternalOrigin := resolveStorefrontOrigins(cfg)
	siteQualityTargetOrigin := resolveSiteQualityTargetOrigin(storefrontBaseURL)
	storefrontHTMLCacheInvalidator := service.NewStorefrontHTMLCacheInvalidatorFromEnv()
	storefrontContentReleaseNotifier := service.NewStorefrontContentReleaseNotifierFromEnv()
	orderNumberGenerator, err := ordernumber.NewGeneratorWithPreviousSecret(
		cfg.OrderNumber.EffectiveSecret(cfg.JWT.Secret),
		cfg.OrderNumber.EffectivePreviousSecret(),
		cfg.OrderNumber.NodeID,
	)
	if err != nil {
		return nil, fmt.Errorf("initialize order number generator: %w", err)
	}

	return &dependencySupport{
		StorageSvc:                          storageSvc,
		SiteLogoStorageSvc:                  siteLogoStorageSvc,
		EmailSvc:                            emailSvc,
		TxManager:                           txManager,
		ShippingService:                     shippingService,
		OutboundHTTPResilience:              outboundHTTPResilience,
		AntiBotService:                      antiBotService,
		AntiFraudService:                    antiFraudService,
		CardBINLimiter:                      cardBINLimiter,
		PaymentGatewayCircuitBreaker:        paymentGatewayCircuitBreaker,
		OrderAbuseService:                   orderAbuseService,
		StorefrontBaseURL:                   storefrontBaseURL,
		StorefrontInternalOrigin:            storefrontInternalOrigin,
		SiteQualityTargetOrigin:             siteQualityTargetOrigin,
		StorefrontHTMLCacheInvalidator:      storefrontHTMLCacheInvalidator,
		StorefrontContentReleaseNotifier:    storefrontContentReleaseNotifier,
		UGCShowcaseUploadProtectionService:  ugcShowcaseUploadProtectionService,
		UGCShowcaseUploadEligibilityService: ugcShowcaseUploadEligibilityService,
		OrderNumberGenerator:                orderNumberGenerator,
	}, nil
}
