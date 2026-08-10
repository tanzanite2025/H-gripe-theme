package app

import (
	"fmt"
	"os"
	"strings"

	"tanzanite/internal/domain/outbox"
	"tanzanite/internal/pkg/antibot"
	"tanzanite/internal/pkg/antifraud"
	"tanzanite/internal/pkg/cache"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/email"
	"tanzanite/internal/pkg/orderabuse"
	"tanzanite/internal/pkg/ordernumber"
	"tanzanite/internal/pkg/storage"
	"tanzanite/internal/repository"
	"tanzanite/internal/service"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Dependencies struct {
	Repositories Repositories
	Services     Services
	Storage      storage.StorageService
	AntiBot      *antibot.Service
	AntiFraud    *antifraud.Service
	OrderAbuse   *orderabuse.Service
	RedisClient  *redis.Client
}

type Repositories struct {
	User                       *repository.UserRepository
	Post                       *repository.PostRepository
	Product                    *repository.ProductRepository
	ProductInformationTemplate *repository.ProductInformationTemplateRepository
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
	GoogleMerchant             *repository.GoogleMerchantRepository
	Registration               *repository.RegistrationRepository
	Audit                      *repository.AuditRepository
	Showcase                   *repository.ShowcaseRepository
	Wishlist                   *repository.WishlistRepository
	Feedback                   *repository.FeedbackRepository
	SuggestionFeedback         *repository.SuggestionFeedbackRepository
	Spoke                      *repository.SpokeRepository
	QuickBuy                   *repository.QuickBuyRepository
	Subscription               *repository.SubscriptionRepository
	EmailChallenge             *repository.EmailChallengeRepository
	VisitorProfile             *repository.VisitorProfileRepository
	RecommendationEvent        *repository.RecommendationEventRepository
	VisitorRiskFact            *repository.VisitorRiskFactRepository
	Outbox                     *repository.OutboxRepository
}

type Services struct {
	Auth                       *service.AuthService
	Post                       *service.PostService
	Product                    *service.ProductService
	ProductInformationTemplate *service.ProductInformationTemplateService
	Cart                       *service.CartService
	Setting                    *service.SettingService
	AdminSettings              *service.AdminSettingsService
	SEO                        *service.SEOService
	SEOResources               *service.SEOResourceService
	Analytics                  *service.AnalyticsService
	AdminPublicChat            *service.AdminPublicChatAgentService
	FAQ                        *service.FAQService
	Gallery                    *service.GalleryService
	Media                      *service.MediaService
	Registration               *service.RegistrationService
	Checkout                   *service.CheckoutService
	Order                      *service.OrderService
	Payment                    *service.PaymentService
	Marketing                  *service.MarketingService
	LoyaltyProgram             *service.LoyaltyProgramService
	Review                     *service.ReviewService
	Ticket                     *service.TicketService
	CustomerServiceContext     *service.CustomerServiceContextService
	CustomerServiceEvents      *service.CustomerServiceEventHub
	Subscription               *service.SubscriptionService
	Sitemap                    *service.SitemapService
	Showcase                   *service.ShowcaseService
	Wishlist                   *service.WishlistService
	Feedback                   *service.FeedbackService
	SuggestionFeedback         *service.SuggestionFeedbackService
	User                       *service.UserService
	Dashboard                  *service.DashboardService
	Audit                      *service.AuditService
	Shipping                   *service.ShippingService
	Spoke                      *service.SpokeService
	QuickBuy                   *service.QuickBuyService
	VisitorProfile             *service.VisitorProfileService
	BehaviorEvents             *service.BehaviorEventService
	Recommendations            *service.RecommendationService
	VisitorRisk                *service.VisitorRiskService
	PaymentRiskMonitoring      *service.PaymentRiskMonitoringService
	PaymentProtection          *service.PaymentProtectionService
	PaymentRefundReview        *service.PaymentRefundRecommendationService
	PaymentThreeDS             *service.PaymentThreeDSPolicyService
	Outbox                     *service.OutboxService
	CurrencyPolicy             *service.CurrencyPolicyService
	ExchangeRate               *service.ExchangeRateService
	StorefrontMarket           *service.StorefrontMarketService
	StorefrontContext          *service.StorefrontContextService
	GoogleMerchant             *service.GoogleMerchantService
}

func NewDependencies(db *gorm.DB, redisCache *cache.RedisCache, cfg *config.Config) (*Dependencies, error) {
	repos := Repositories{
		User:                       repository.NewUserRepository(db),
		Post:                       repository.NewPostRepository(db),
		Product:                    repository.NewProductRepository(db),
		ProductInformationTemplate: repository.NewProductInformationTemplateRepository(db),
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
		GoogleMerchant:             repository.NewGoogleMerchantRepository(db),
		Registration:               repository.NewRegistrationRepository(db),
		Audit:                      repository.NewAuditRepository(db),
		Showcase:                   repository.NewShowcaseRepository(db),
		Wishlist:                   repository.NewWishlistRepository(db),
		Feedback:                   repository.NewFeedbackRepository(db),
		SuggestionFeedback:         repository.NewSuggestionFeedbackRepository(db),
		Spoke:                      repository.NewSpokeRepository(db),
		QuickBuy:                   repository.NewQuickBuyRepository(db),
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
	txManager.ConfigureOrderAttributionRepository(repos.OrderAttribution)
	txManager.ConfigurePaymentRefundRecommendationRepository(repos.PaymentRefundReview)
	txManager.ConfigurePaymentRefundExecutionRepository(repos.PaymentRefundExec)

	shippingService := service.NewShippingService(repos.Shipping, repos.Product)
	antiBotService := antibot.New(redisCache.Client(), cfg.AntiAbuse)
	antiFraudService := antifraud.New(redisCache.Client(), cfg.PaymentRisk)
	orderAbuseService := orderabuse.New(redisCache.Client(), cfg.OrderAbuse)

	storefrontHTMLCacheInvalidator := service.NewStorefrontHTMLCacheInvalidatorFromEnv()
	storefrontContentReleaseNotifier := service.NewStorefrontContentReleaseNotifierFromEnv()
	settingService := service.NewSettingService(repos.Setting, redisCache, cfg.Cache.SettingsTTL)
	seoService := service.NewSEOService(settingService)
	postService := service.NewPostService(repos.Post, redisCache, cfg.Cache.PostTTL)
	productService := service.NewProductService(repos.Product, redisCache, cfg.Cache.ProductTTL)
	merchantOutboxPublisher := service.NewMerchantOutboxPublisher(repos.Outbox)
	seoResourceService := service.NewSEOResourceService(postService, productService, settingService)
	analyticsService := service.NewAnalyticsService(settingService)
	currencyPolicyService := service.NewCurrencyPolicyService(repos.Setting)
	exchangeRateService := service.NewExchangeRateService(repos.ExchangeRate, repos.Setting)
	storefrontMarketService := service.NewStorefrontMarketService(repos.StorefrontMarket)
	storefrontBaseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("STOREFRONT_BASE_URL")), "/")
	if storefrontBaseURL == "" {
		storefrontBaseURL = strings.TrimRight(strings.TrimSpace(cfg.Server.BaseURL), "/")
	}
	mediaService := service.NewMediaService(repos.Media, storageSvc, settingService, storefrontBaseURL, cfg.MediaUpload.AccountStorageQuotaBytes)
	productService.ConfigureMediaService(mediaService)
	seoResourceService.ConfigureCanonicalBaseURL(storefrontBaseURL)
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
		Auth:                       service.NewAuthService(repos.User, cfg.JWT, cfg.OAuth),
		Post:                       postService,
		Product:                    productService,
		ProductInformationTemplate: service.NewProductInformationTemplateService(repos.ProductInformationTemplate),
		Cart:                       service.NewCartService(repos.Cart, repos.Product),
		Setting:                    settingService,
		SEO:                        seoService,
		SEOResources:               seoResourceService,
		Analytics:                  analyticsService,
		CurrencyPolicy:             currencyPolicyService,
		ExchangeRate:               exchangeRateService,
		StorefrontMarket:           storefrontMarketService,
		StorefrontContext:          service.NewStorefrontContextServiceWithMarkets(currencyPolicyService, storefrontMarketService),
		GoogleMerchant:             googleMerchantService,
		FAQ:                        service.NewFAQService(repos.FAQ, storageSvc),
		Gallery:                    service.NewGalleryService(repos.Gallery),
		Media:                      mediaService,
		Registration:               service.NewRegistrationService(repos.Registration, repos.Product, repos.Order),
		Checkout:                   service.NewCheckoutService(repos.Product, repos.Coupon, repos.Payment, repos.Loyalty, shippingService),
		Marketing:                  service.NewMarketingService(txManager, repos.Coupon, repos.Loyalty, settingService),
		LoyaltyProgram:             loyaltyProgramService,
		Review:                     service.NewReviewService(repos.Review),
		Ticket:                     service.NewTicketService(repos.Ticket, repos.User, repos.FAQ),
		CustomerServiceEvents:      service.NewCustomerServiceEventHub(),
		Subscription:               service.NewSubscriptionService(repos.Subscription),
		Sitemap:                    service.NewSitemapService(repos.Post, cfg.Server.BaseURL),
		Showcase:                   service.NewShowcaseService(repos.Showcase, storageSvc),
		Wishlist:                   service.NewWishlistService(repos.Wishlist, repos.Product),
		Feedback:                   service.NewFeedbackService(repos.Feedback),
		SuggestionFeedback: service.NewSuggestionFeedbackService(
			repos.SuggestionFeedback,
		),
		User:      service.NewUserService(repos.User),
		Dashboard: service.NewDashboardService(repos.Order, repos.User, repos.Ticket, repos.Subscription),
		Audit:     service.NewAuditService(repos.Audit),
		Shipping:  shippingService,
		Spoke:     service.NewSpokeService(repos.Spoke),
		QuickBuy:  service.NewQuickBuyService(repos.QuickBuy, repos.Product),
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
	services.Marketing.ConfigureLoyaltyProgram(loyaltyProgramService)
	services.Marketing.ConfigureGiftCardRedemptions(repos.GiftCardRedemption)
	services.Marketing.ConfigureCurrencyPolicy(currencyPolicyService)
	services.Checkout.ConfigureLoyaltyProgram(loyaltyProgramService)
	services.Product.ConfigureCurrencyPolicy(currencyPolicyService)
	services.Product.ConfigureInformationTemplateRepository(repos.ProductInformationTemplate)
	services.Product.ConfigureMerchantEventPublisher(merchantOutboxPublisher)
	services.GoogleMerchant.ConfigureMerchantEventPublisher(merchantOutboxPublisher)
	services.ExchangeRate.ConfigureCurrencyPolicy(currencyPolicyService)
	services.Registration.ConfigureEmailChallenges(repos.EmailChallenge, cfg.JWT.Secret, emailSvc)
	services.Registration.ConfigureEmailBaseURL(storefrontBaseURL)
	services.Subscription.ConfigureEmailChallenges(repos.EmailChallenge, cfg.JWT.Secret, emailSvc)
	services.Subscription.ConfigureEmailBaseURL(storefrontBaseURL)
	services.Product.SetStorefrontHTMLCacheInvalidator(storefrontHTMLCacheInvalidator)
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
	services.Order = service.NewOrderService(
		txManager,
		repos.Order,
		services.Checkout,
		shippingService,
		orderNumberGenerator,
	)
	services.Payment = service.NewPaymentService(txManager, repos.Payment)
	services.Payment.ConfigureRisk(repos.Order, antiFraudService)
	services.Payment.ConfigureEvidenceSources(repos.Order, repos.Shipping, repos.Ticket)
	services.PaymentThreeDS = service.NewPaymentThreeDSPolicyService(
		repos.Order,
		services.VisitorRisk,
		antiFraudService,
		cfg.PaymentThreeDS,
	)
	services.PaymentThreeDS.ConfigureRiskMonitoring(services.PaymentRiskMonitoring)
	services.PaymentThreeDS.ConfigurePaymentProtection(services.PaymentProtection)

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

	return &Dependencies{
		Repositories: repos,
		Services:     services,
		Storage:      storageSvc,
		AntiBot:      antiBotService,
		AntiFraud:    antiFraudService,
		OrderAbuse:   orderAbuseService,
		RedisClient:  redisCache.Client(),
	}, nil
}
