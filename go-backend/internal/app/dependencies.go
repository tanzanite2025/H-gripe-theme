package app

import (
	"os"
	"strings"

	"commerce-platform/internal/pkg/antibot"
	"commerce-platform/internal/pkg/antifraud"
	"commerce-platform/internal/pkg/cache"
	"commerce-platform/internal/pkg/cardtesting"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/orderabuse"
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

const defaultDevStorefrontOrigin = "http://localhost:9199"

type Repositories struct {
	User                         *repository.UserRepository
	Post                         *repository.PostRepository
	StorefrontRouteCatalog       *repository.StorefrontRouteCatalogRepository
	StorefrontURLSearchProfiles  *repository.StorefrontURLSearchProfileRepository
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
	GlobalIPBlockRule            *repository.GlobalIPBlockRuleRepository
	OpsDeploymentWorkflow        *repository.OpsDeploymentWorkflowRepository
	GoogleMerchant               *repository.GoogleMerchantRepository
	SocialOAuth                  *repository.SocialOAuthRepository
	Warranty                     *repository.WarrantyRepository
	ShipmentRecord               *repository.ShipmentRecordRepository
	Audit                        *repository.AuditRepository
	UGCShowcase                  *repository.UGCShowcaseRepository
	HomeVisualTiles              *repository.HomeVisualTileRepository
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
	StorefrontURLSearchProfiles       *service.StorefrontURLSearchProfileService
	StorefrontRedirectRules           *service.StorefrontRedirectRuleService
	StorefrontURLIssues               *service.StorefrontURLIssueService
	PreflightContentLinks             *service.PreflightContentLinkService
	LighthouseRunner                  *service.LighthouseRunnerService
	SiteQualityEngine                 *service.SiteQualityEngineService
	HotDataArchive                    *service.HotDataArchiveService
	UGCShowcase                       *service.UGCShowcaseService
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
	GlobalIPBlock                     *service.GlobalIPBlockService
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
	UGCShowcaseUploadProtection       *service.UGCShowcaseUploadProtectionService
	UGCShowcaseUploadEligibility      *service.UGCShowcaseUploadEligibilityService
	PublicUploadAccess                *service.PublicUploadAccessService
	HomeVisualTiles                   *service.HomeVisualTileService
}

func NewDependencies(db *gorm.DB, redisCache *cache.RedisCache, cfg *config.Config) (*Dependencies, error) {
	repos, err := newDependencyRepositories(db)
	if err != nil {
		return nil, err
	}
	support, err := newDependencySupport(db, redisCache, cfg, repos)
	if err != nil {
		return nil, err
	}
	services, customerServiceRealtimeRelay, err := newDependencyServices(db, redisCache, cfg, repos, support)
	if err != nil {
		return nil, err
	}

	return &Dependencies{
		Repositories:                 repos,
		Services:                     services,
		Storage:                      support.StorageSvc,
		SiteLogoStorage:              support.SiteLogoStorageSvc,
		AntiBot:                      support.AntiBotService,
		AntiFraud:                    support.AntiFraudService,
		CardBINLimiter:               support.CardBINLimiter,
		PaymentGatewayCircuitBreaker: support.PaymentGatewayCircuitBreaker,
		OrderAbuse:                   support.OrderAbuseService,
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
