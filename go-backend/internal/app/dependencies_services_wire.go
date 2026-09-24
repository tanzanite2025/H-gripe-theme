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

	if support != nil && support.ShippingService != nil {
		support.ShippingService.ConfigureAuditRecorder(services.Audit)
	}

	recommendationsService := service.NewRecommendationService(services.Product, repos.RecommendationEvent)
	services.Recommendations = recommendationsService
	services.Recommendations.ConfigureMediaService(services.Media)
	services.Recommendations.ConfigureReviewService(services.Review)
	services.QuickBuy.ConfigureMediaService(services.Media)
	services.GoogleMerchant.ConfigureMediaService(services.Media)
	services.Ticket.ConfigureCustomerServiceRealtimeOutbox(repos.Outbox)
	services.CustomerServiceAvatar = service.NewCustomerServiceAvatarService(repos.User, support.StorageSvc, repos.Outbox)
	services.Media.ConfigureObjectCleanupOutbox(repos.Outbox)
	services.SiteLogo.ConfigureObjectCleanupOutbox(repos.Outbox)
	services.HomeVisualTiles.ConfigureObjectCleanupOutbox(repos.Outbox)
	services.UGCShowcase.ConfigureObjectCleanupOutbox(repos.Outbox)
	services.PublicUploadAccess = service.NewPublicUploadAccessService(services.Media, services.UGCShowcase, services.CustomerServiceAvatar)
	services.PublicUploadAccess.ConfigureSiteLogoService(services.SiteLogo)
	services.FAQ.ConfigureMediaService(services.Media)
	services.Review.ConfigureMediaService(services.Media)
	services.UGCShowcase.ConfigureUploadEligibility(services.UGCShowcaseUploadEligibility)
	if cfg.ShowcaseUploadProtection.Enabled {
		services.UGCShowcase.ConfigurePendingSubmissionLimit(cfg.ShowcaseUploadProtection.MaxPendingSubmissionsPerUser)
	}
	services.Marketing.ConfigureLoyaltyProgram(b.services.LoyaltyProgram)
	services.Marketing.ConfigureCurrencyPolicy(services.CurrencyPolicy)
	services.Checkout.ConfigureCurrencyPolicy(services.CurrencyPolicy)
	services.Checkout.ConfigureExchangeRateRepository(repos.ExchangeRate)
	services.Checkout.ConfigureReferralRepositories(repos.Referral, repos.ReferralProgram)
	services.Product.ConfigureCurrencyPolicy(services.CurrencyPolicy)
	services.Product.ConfigureInformationTemplateRepository(repos.ProductInformationTemplate)
	services.Product.ConfigureProductBrandRepository(repos.ProductBrand)
	services.Product.ConfigureCustomsClassificationRepository(repos.CustomsClassification)
	services.Product.ConfigureProductCategoryRepository(repos.ProductCategory)
	services.ProductCategory.ConfigureProductCacheInvalidator(services.Product)
	services.ProductCategory.SetStorefrontHTMLCacheInvalidator(support.StorefrontHTMLCacheInvalidator)
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
	support.ShippingService.ConfigureExchangeRateService(services.ExchangeRate)
	services.Warranty.ConfigureEmailChallenges(cfg.JWT.Secret)
	services.Warranty.ConfigureEmailBaseURL(support.StorefrontBaseURL)
	services.Subscription.ConfigureEmailChallenges(cfg.JWT.Secret)
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
	services.CustomerServiceContext.ConfigureFactRepositories(repos.AfterSales, repos.Payment)
	services.CustomerServiceContext.ConfigureFulfillmentRepositories(repos.Product, repos.Shipping, repos.Warranty)
	services.CustomerServiceContext.ConfigureAuditService(services.Audit)
	services.CustomerServiceContext.ConfigureMediaService(services.Media)
	minimumRetentionAge := time.Duration(cfg.Worker.CustomerServiceRetentionMinimumDays) * 24 * time.Hour
	if minimumRetentionAge <= 0 && cfg.Worker.CustomerServiceRetentionMinimumMonths > 0 {
		minimumRetentionAge = time.Duration(cfg.Worker.CustomerServiceRetentionMinimumMonths) * 30 * 24 * time.Hour
	}
	recoveryWindow := time.Duration(cfg.Worker.CustomerServiceRetentionRecoveryWindowDays) * 24 * time.Hour
	services.CustomerServiceRetention.ConfigureWindows(minimumRetentionAge, recoveryWindow)
	services.CustomerServiceRetention.ConfigureLegacyWorker(cfg.Worker.CustomerServiceRetentionEnabled)
	services.CustomerServiceRetention.ConfigurePolicyRepository(repos.CustomerServiceRetentionPolicy)
	services.CustomerServiceRetention.ConfigureAttachmentStorage(support.StorageSvc)
	services.CustomerServiceRetention.ConfigureMediaService(services.Media)
	services.CustomerServiceRetention.ConfigureCleanupOutbox(repos.Outbox)
	if searchIndex, err := service.NewCustomerServiceHTTPSearchIndexFromEnv(
		support.OutboundHTTPResilience.retry,
		support.OutboundHTTPResilience.breaker,
	); err != nil {
		return fmt.Errorf("configure customer-service search index: %w", err)
	} else if searchIndex != nil {
		services.CustomerServiceRetention.ConfigureSearchIndex(searchIndex)
	}
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
	services.Order.ConfigureOrderEvidenceSnapshot(services.OrderEvidenceSnapshot)
	services.Order.ConfigureOrderEvidence(services.OrderEvidence)
	services.Order.ConfigureProductCacheInvalidator(services.Product)
	services.Order.ConfigureProductCacheEventPublisher(b.productCacheOutboxPublisher)
	services.Order.ConfigureRefundCancellationPolicy(services.RefundCancellationPolicy)
	services.Payment = service.NewPaymentService(support.TxManager, repos.Payment)
	services.Payment.ConfigureProductCacheInvalidator(services.Product)
	services.Payment.ConfigureProductCacheEventPublisher(b.productCacheOutboxPublisher)
	services.Payment.ConfigureRisk(repos.Order, support.AntiFraudService)
	services.Payment.ConfigureEvidenceSources(repos.Order, repos.Ticket)
	services.Payment.ConfigureOrderEvidenceAssembler(
		service.NewOrderEvidencePackageAssembler(repos.Order, repos.OrderEvidence, repos.Shipping),
	)
	services.Payment.ConfigureOrderEvidenceSubmissionSnapshotRepository(repos.OrderEvidenceSubmission)
	services.Payment.ConfigurePolicyDisclosureRepository(repos.OrderPolicyDisclosure)
	services.Payment.ConfigurePayPalDisputeEvidenceDocumentStorage(support.StorageSvc)
	if podURLProvider, ok := support.StorageSvc.(service.PayPalDisputeEvidenceAttachmentURLProvider); ok {
		services.Payment.ConfigurePayPalDisputeEvidenceAttachmentURLProvider(podURLProvider)
	}
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
	services.PaymentThreeDS.ConfigureExchangeRateService(services.ExchangeRate)

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
	services.Outbox.RegisterHandler(outbox.EventTypeOrderPaid, orderPaidWebhookHandler.Handle)
	orderCompletionHandler := service.NewOrderCompletionOutboxHandler(services.Order, services.TransactionalNotificationTemplates)
	orderCompletionHandler.ConfigureTransactionalNotificationSender(support.EmailSvc)
	services.Outbox.RegisterHandler(outbox.EventTypeOrderCompleted, orderCompletionHandler.Handle)
	canonicalDomainEventHandler := service.NewCanonicalDomainEventOutboxHandlerWithSender(
		services.TransactionalNotificationTemplates,
		support.EmailSvc,
	)
	canonicalDomainEventHandler.ConfigureNotificationSettings(services.Setting)
	services.Outbox.RegisterHandler(outbox.EventTypeOrderPaymentSucceeded, canonicalDomainEventHandler.Handle)
	services.Outbox.RegisterHandler(outbox.EventTypeOrderPaymentExpired, canonicalDomainEventHandler.Handle)
	services.Outbox.RegisterHandler(outbox.EventTypeOrderCancelled, canonicalDomainEventHandler.Handle)
	services.Outbox.RegisterHandler(outbox.EventTypeOrderShipped, canonicalDomainEventHandler.Handle)
	services.Outbox.RegisterHandler(outbox.EventTypeOrderDelivered, canonicalDomainEventHandler.Handle)
	services.Outbox.RegisterHandler(outbox.EventTypeOrderRefunded, canonicalDomainEventHandler.Handle)
	services.Outbox.RegisterHandler(outbox.EventTypeAfterSalesStatusChanged, canonicalDomainEventHandler.Handle)
	services.Outbox.RegisterHandler(outbox.EventTypeReferralOrderPaid, services.Referral.HandleOrderPaidOutbox)
	services.Outbox.RegisterHandler(outbox.EventTypeReferralOrderDelivered, services.Referral.HandleOrderDeliveredOutbox)
	services.Outbox.RegisterHandler(outbox.EventTypeReferralOrderInvalidated, services.Referral.HandleOrderInvalidatedOutbox)
	orderDisputeContactEmailHandler := service.NewOrderDisputeContactEmailOutboxHandler(support.EmailSvc)
	services.Outbox.RegisterHandler(outbox.EventTypeOrderDisputeContactEmail, orderDisputeContactEmailHandler.Handle)
	emailChallengeDeliveryHandler := service.NewEmailChallengeDeliveryOutboxHandler(support.EmailSvc, support.AntiBotService, cfg.JWT.Secret)
	services.Outbox.RegisterHandler(outbox.EventTypeEmailChallengeDelivery, emailChallengeDeliveryHandler.Handle)
	trackingRegistrationHandler := service.NewTrackingShipmentRegistrationOutboxHandler(support.ShippingService)
	services.Outbox.RegisterHandler(outbox.EventTypeTrackingShipmentRegistration, trackingRegistrationHandler.Handle)
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
	paymentRefundWebhookHandler := service.NewPaymentRefundOutboxWebhookHandlerFromEnvWithResilience(
		support.OutboundHTTPResilience.retry,
		support.OutboundHTTPResilience.breaker,
	)
	if paymentRefundWebhookHandler.Configured() {
		services.Outbox.RegisterHandler(outbox.EventTypePaymentRefundPending, paymentRefundWebhookHandler.Handle)
		services.Outbox.RegisterHandler(outbox.EventTypePaymentRefundCompleted, paymentRefundWebhookHandler.Handle)
		services.Outbox.RegisterHandler(outbox.EventTypePaymentRefundFailed, paymentRefundWebhookHandler.Handle)
	}
	paymentRefundExecutionOutboxHandler := service.NewPaymentRefundExecutionOutboxHandler(
		services.Payment,
		services.AdminSettings,
	)
	services.Outbox.RegisterHandler(
		outbox.EventTypePaymentRefundExecutionRequested,
		paymentRefundExecutionOutboxHandler.Handle,
	)
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
		outbox.EventTypeCustomerServiceRetentionCleanup,
		services.CustomerServiceRetention.HandleCustomerServiceRetentionCleanup,
	)
	objectStorageCleanupHandler := service.NewObjectStorageCleanupOutboxHandler(
		services.Media,
		services.SiteLogo,
		services.HomeVisualTiles,
		services.UGCShowcase,
	)
	services.Outbox.RegisterHandler(outbox.EventTypeObjectStorageCleanup, objectStorageCleanupHandler.Handle)
	services.Outbox.RegisterHandler(
		outbox.EventTypeStorefrontRouteCatalogChanged,
		service.NewSiteQualityRouteCatalogOutboxHandler(services.SiteQualityEngine).Handle,
	)

	return nil
}
