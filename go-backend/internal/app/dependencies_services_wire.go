package app

import (
	"fmt"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/service"
)

func (b *dependencyServicesBuilder) wire() error {
	services := &b.services
	repos := b.repos
	cfg := b.cfg
	support := b.support

	recommendationsService := service.NewRecommendationService(services.Product, repos.RecommendationEvent)
	services.Recommendations = recommendationsService
	services.Recommendations.ConfigureMediaService(services.Media)
	services.Recommendations.ConfigureReviewService(services.Review)
	services.QuickBuy.ConfigureMediaService(services.Media)
	services.GoogleMerchant.ConfigureMediaService(services.Media)
	services.Ticket.ConfigureCustomerServiceRealtimeOutbox(repos.Outbox)
	services.CustomerServiceAvatar = service.NewCustomerServiceAvatarService(repos.User, support.StorageSvc, repos.Outbox)
	services.PublicUploadAccess = service.NewPublicUploadAccessService(services.Media, services.UGCShowcase, services.CustomerServiceAvatar)
	services.PublicUploadAccess.ConfigureSiteLogoService(services.SiteLogo)
	services.FAQ.ConfigureMediaService(services.Media)
	services.Review.ConfigureMediaService(services.Media)
	services.UGCShowcase.ConfigureUploadEligibility(services.UGCShowcaseUploadEligibility)
	if cfg.ShowcaseUploadProtection.Enabled {
		services.UGCShowcase.ConfigurePendingSubmissionLimit(cfg.ShowcaseUploadProtection.MaxPendingSubmissionsPerUser)
	}
	services.Marketing.ConfigureLoyaltyProgram(b.services.LoyaltyProgram)
	services.Marketing.ConfigureGiftCardRedemptions(repos.GiftCardRedemption)
	services.Marketing.ConfigureCurrencyPolicy(services.CurrencyPolicy)
	services.Checkout.ConfigureLoyaltyProgram(b.services.LoyaltyProgram)
	services.Checkout.ConfigureCurrencyPolicy(services.CurrencyPolicy)
	services.Checkout.ConfigureExchangeRateRepository(repos.ExchangeRate)
	services.Product.ConfigureCurrencyPolicy(services.CurrencyPolicy)
	services.Product.ConfigureInformationTemplateRepository(repos.ProductInformationTemplate)
	services.Product.ConfigureProductBrandRepository(repos.ProductBrand)
	services.Product.ConfigureCustomsClassificationRepository(repos.CustomsClassification)
	services.Product.ConfigureProductCategoryRepository(repos.ProductCategory)
	services.Product.ConfigureMerchantEventPublisher(b.merchantOutboxPublisher)
	services.Product.ConfigureProductCacheEventPublisher(b.productCacheOutboxPublisher)
	services.ProductBrand.ConfigureProductDependencies(repos.Product, b.productCacheOutboxPublisher, b.merchantOutboxPublisher)
	services.Product.ConfigureTxManager(support.TxManager)
	services.ProductBrand.ConfigureTxManager(support.TxManager)
	services.ProductInformationTemplate.ConfigureProductCacheInvalidator(services.Product)
	services.ProductInformationTemplate.ConfigureProductCacheEventPublisher(b.productCacheOutboxPublisher)
	services.GoogleMerchant.ConfigureMerchantEventPublisher(b.merchantOutboxPublisher)
	services.ExchangeRate.ConfigureCurrencyPolicy(services.CurrencyPolicy)
	services.ExchangeRate.ConfigureStorefrontMarkets(services.StorefrontMarket)
	services.ExchangeRate.ConfigureProductService(services.Product)
	services.ExchangeRate.ConfigureShippingService(support.ShippingService)
	services.Warranty.ConfigureEmailChallenges(repos.EmailChallenge, cfg.JWT.Secret, support.EmailSvc)
	services.Warranty.ConfigureEmailBaseURL(support.StorefrontBaseURL)
	services.Subscription.ConfigureEmailChallenges(repos.EmailChallenge, cfg.JWT.Secret, support.EmailSvc)
	services.Subscription.ConfigureEmailBaseURL(support.StorefrontBaseURL)
	services.Product.SetStorefrontHTMLCacheInvalidator(support.StorefrontHTMLCacheInvalidator)
	services.ProductBrand.SetStorefrontHTMLCacheInvalidator(support.StorefrontHTMLCacheInvalidator)
	services.Post.SetStorefrontHTMLCacheInvalidator(support.StorefrontHTMLCacheInvalidator)
	services.FAQ.SetStorefrontHTMLCacheInvalidator(support.StorefrontHTMLCacheInvalidator)
	services.FAQ.SetStorefrontContentReleaseNotifier(support.StorefrontContentReleaseNotifier)
	services.SEO.SetStorefrontHTMLCacheInvalidator(support.StorefrontHTMLCacheInvalidator)
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
		support.TxManager,
		repos.Order,
		services.Checkout,
		support.ShippingService,
		support.OrderNumberGenerator,
	)
	services.Order.ConfigureProductCacheInvalidator(services.Product)
	services.Order.ConfigureProductCacheEventPublisher(b.productCacheOutboxPublisher)
	services.Order.ConfigureRefundReturnPolicy(services.RefundReturnPolicy)
	services.Payment = service.NewPaymentService(support.TxManager, repos.Payment)
	services.Payment.ConfigureProductCacheInvalidator(services.Product)
	services.Payment.ConfigureProductCacheEventPublisher(b.productCacheOutboxPublisher)
	services.Payment.ConfigureRisk(repos.Order, support.AntiFraudService)
	services.Payment.ConfigureEvidenceSources(repos.Order, repos.Shipping, repos.Ticket)
	services.Payment.ConfigurePolicyDisclosureRepository(repos.OrderPolicyDisclosure)
	services.Payment.ConfigurePayPalDisputeEvidenceDocumentStorage(support.StorageSvc)
	services.Payment.ConfigurePayPalDisputeInvoiceSellerProfileProvider(services.PayPalDisputeInvoiceSellerProfile)
	services.Payment.ConfigurePayPalDisputeInvoiceOptions(service.PayPalDisputeInvoiceOptions{
		AutoAttachPDF: envBoolDefault("PAYPAL_DISPUTE_AUTO_ATTACH_INVOICE_PDF", true),
	})
	services.Order.ConfigurePaymentDisputeAnalysis(services.Payment)
	services.Order.ConfigureAdminEmailSender(support.EmailSvc)
	services.PaymentThreeDS = service.NewPaymentThreeDSPolicyService(
		repos.Order,
		services.VisitorRisk,
		support.AntiFraudService,
		cfg.PaymentThreeDS,
	)
	services.PaymentThreeDS.ConfigureRiskMonitoring(services.PaymentRiskMonitoring)
	services.PaymentThreeDS.ConfigurePaymentProtection(services.PaymentProtection)

	if cfg.CustomerServiceRealtime.Enabled {
		var err error
		b.customerServiceRealtimeRelay, err = service.NewCustomerServiceRealtimeRelay(
			b.redisCache.Client(),
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
			return fmt.Errorf("initialize customer-service realtime relay: %w", err)
		}
		services.CustomerServiceEvents.ConfigureReplayProvider(b.customerServiceRealtimeRelay)
	}

	orderPaidWebhookHandler := service.NewOrderPaidOutboxWebhookHandlerFromEnvWithResilience(
		support.OutboundHTTPResilience.retry,
		support.OutboundHTTPResilience.breaker,
	)
	if orderPaidWebhookHandler.Configured() {
		services.Outbox.RegisterHandler(outbox.EventTypeOrderPaid, orderPaidWebhookHandler.Handle)
	}
	verifiedConversionWebhookHandler := service.NewVerifiedConversionOutboxWebhookHandlerFromEnvWithResilience(
		support.OutboundHTTPResilience.retry,
		support.OutboundHTTPResilience.breaker,
	)
	if verifiedConversionWebhookHandler.Configured() {
		services.Outbox.RegisterHandler(outbox.EventTypeVerifiedConversion, verifiedConversionWebhookHandler.Handle)
	}
	paymentRiskAlertWebhookHandler := service.NewPaymentRiskAlertOutboxWebhookHandlerFromEnvWithResilience(
		support.OutboundHTTPResilience.retry,
		support.OutboundHTTPResilience.breaker,
	)
	if cfg.PaymentRiskMonitoring.AlertEnabled && paymentRiskAlertWebhookHandler.Configured() {
		services.PaymentRiskMonitoring.ConfigureAlerting(true)
		services.Outbox.RegisterHandler(outbox.EventTypePaymentRiskLevelChanged, paymentRiskAlertWebhookHandler.Handle)
		services.Outbox.RegisterHandler(outbox.EventTypePaymentRiskFailOpen, paymentRiskAlertWebhookHandler.Handle)
		services.PaymentThreeDS.ConfigureFailOpenAlertPublisher(service.NewPaymentRiskFailOpenOutboxPublisher(repos.Outbox))
	}
	merchantOutboxHandler := service.NewGoogleMerchantOutboxHandler(services.GoogleMerchant)
	services.Outbox.RegisterHandler(outbox.EventTypeMerchantProductUpsert, merchantOutboxHandler.Handle)
	services.Outbox.RegisterHandler(outbox.EventTypeMerchantProductWithdraw, merchantOutboxHandler.Handle)
	services.Outbox.RegisterHandler(outbox.EventTypeMerchantOfferRevalidate, merchantOutboxHandler.Handle)
	productCacheOutboxHandler := service.NewProductCacheOutboxHandler(services.Product)
	services.Outbox.RegisterHandler(outbox.EventTypeProductCacheInvalidate, productCacheOutboxHandler.Handle)
	customerServiceRealtimeOutboxHandler := service.NewCustomerServiceRealtimeOutboxHandler(services.CustomerServiceEvents, b.customerServiceRealtimeRelay)
	services.Outbox.RegisterHandler(outbox.EventTypeCustomerServiceRealtime, customerServiceRealtimeOutboxHandler.Handle)
	customerServiceAvatarCleanupHandler := service.NewCustomerServiceAvatarCleanupHandler(repos.User, support.StorageSvc)
	services.Outbox.RegisterHandler(outbox.EventTypeCustomerServiceAvatarCleanup, customerServiceAvatarCleanupHandler.Handle)
	services.Outbox.RegisterHandler(
		outbox.EventTypeStorefrontRouteCatalogChanged,
		service.NewSiteQualityRouteCatalogOutboxHandler(services.SiteQualityEngine).Handle,
	)

	return nil
}
