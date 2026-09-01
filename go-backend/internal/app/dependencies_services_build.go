package app

import (
	"fmt"
	"time"

	"commerce-platform/internal/service"
)

func (b *dependencyServicesBuilder) build() error {
	storageSvc := b.support.StorageSvc
	siteLogoStorageSvc := b.support.SiteLogoStorageSvc
	txManager := b.support.TxManager
	shippingService := b.support.ShippingService
	outboundHTTPResilience := b.support.OutboundHTTPResilience
	UGCShowcaseUploadProtectionService := b.support.UGCShowcaseUploadProtectionService
	UGCShowcaseUploadEligibilityService := b.support.UGCShowcaseUploadEligibilityService
	storefrontBaseURL := b.support.StorefrontBaseURL
	storefrontInternalOrigin := b.support.StorefrontInternalOrigin
	siteQualityTargetOrigin := b.support.SiteQualityTargetOrigin

	settingService := service.NewSettingService(b.repos.Setting, b.redisCache, b.cfg.Cache.SettingsTTL)
	refundReturnPolicyService := service.NewRefundReturnPolicyService(b.repos.Setting)
	seoService := service.NewSEOService(settingService)
	postService := service.NewPostService(b.repos.Post, b.redisCache, b.cfg.Cache.PostTTL)
	productService := service.NewProductServiceWithCacheOptions(b.repos.Product, b.redisCache, b.cfg.Cache.ProductTTL, b.cfg.Cache.ProductLockTTL)
	productProcurementService := service.NewProductProcurementServiceWithProfitability(
		b.repos.ProductProcurement,
		b.repos.ProductProfitCalculation,
	)
	productProcurementService.ConfigureCatalogRepository(b.repos.ProductProcurementCatalog)
	fitmentHubSpecificationService := service.NewFitmentHubSpecificationService(
		b.repos.FitmentHubSpecification,
		b.repos.FitmentFrameHubSpecification,
	)
	fitmentHubSpecificationService.ConfigureForkHubSpecificationRepository(b.repos.FitmentForkHubSpecification)
	fitmentHubSpecificationService.ConfigureSpokeRepository(b.repos.Spoke)
	frameFitmentEntryService := service.NewFrameFitmentEntryService(
		b.repos.FrameFitmentEntry,
		b.repos.FitmentHubSpecification,
		b.repos.FitmentFrameHubSpecification,
	)
	forkFitmentEntryService := service.NewForkFitmentEntryService(
		b.repos.ForkFitmentEntry,
		b.repos.FitmentHubSpecification,
		b.repos.FitmentForkHubSpecification,
	)
	productProfitabilityService := service.NewProductProfitabilityServiceWithProcurement(
		b.repos.ProductProfitCalculation,
		b.repos.ProductProcurement,
	)
	productCategoryService := service.NewProductCategoryService(b.repos.ProductCategory, b.repos.Media)
	productBrandService := service.NewProductBrandService(b.repos.ProductBrand)
	productInformationTemplateService := service.NewProductInformationTemplateService(b.repos.ProductInformationTemplate)
	customsClassificationService := service.NewCustomsClassificationService(b.repos.CustomsClassification, settingService)
	merchantOutboxPublisher := service.NewMerchantOutboxPublisher(b.repos.Outbox)
	productCacheOutboxPublisher := service.NewProductCacheOutboxPublisher(b.repos.Outbox)
	b.merchantOutboxPublisher = merchantOutboxPublisher
	b.productCacheOutboxPublisher = productCacheOutboxPublisher
	seoResourceService := service.NewSEOResourceService(postService, productService, productCategoryService)
	analyticsService := service.NewAnalyticsService(settingService)
	currencyPolicyService := service.NewCurrencyPolicyService(b.repos.Setting)
	exchangeRateService := service.NewExchangeRateService(b.repos.ExchangeRate, b.repos.Setting)
	shippingService.ConfigureCurrencyPolicy(currencyPolicyService)
	storefrontMarketService := service.NewStorefrontMarketService(b.repos.StorefrontMarket)
	opsDomainBindingService := service.NewOpsDomainBindingService(b.repos.OpsDomainBinding, b.repos.OpsProjectBinding, b.repos.OpsConnector)
	opsDomainDiffService := service.NewOpsDomainDiffService(b.repos.OpsDomainBinding)
	opsDomainPreviewService := service.NewOpsDomainPreviewService(b.repos.OpsDomainBinding)
	opsConnectorService := service.NewOpsConnectorService(b.repos.OpsConnector)
	opsNetworkSummaryService := service.NewOpsNetworkSummaryService(
		b.repos.OpsNetworkRule,
		b.repos.OpsVPSBinding,
		b.repos.OpsProjectBinding,
		b.repos.OpsDomainBinding,
		b.repos.OpsConnector,
	)
	opsDomainSyncService := service.NewOpsDomainSyncService(b.repos.OpsDomainBinding, opsConnectorService)
	opsHostingerSyncService := service.NewOpsHostingerSyncService(
		b.repos.OpsVPSBinding,
		b.repos.OpsProjectBinding,
		opsConnectorService,
	)
	opsConnectorOAuthService := service.NewOpsConnectorOAuthService(
		b.repos.OpsConnectorOAuth,
		b.repos.OpsConnector,
		opsConnectorService,
		b.repos.OpsVPSBinding,
		b.repos.OpsProjectBinding,
		b.repos.OpsDomainBinding,
		opsHostingerSyncService,
		opsHostingerSyncService,
		opsDomainSyncService,
		b.cfg.Server.BaseURL,
	)
	opsDeploymentPreflightService := service.NewOpsDeploymentPreflightService(
		b.repos.OpsProjectBinding,
		b.repos.OpsVPSBinding,
		b.repos.OpsConnector,
		b.repos.OpsDomainBinding,
	)
	opsDeploymentHealthCheckService := service.NewOpsDeploymentHealthCheckService(b.repos.OpsDomainBinding)
	opsCloudflareCachePurgeService := service.NewOpsCloudflareCachePurgeService(b.repos.OpsDomainBinding, opsConnectorService)
	opsDeploymentWorkflowService := service.NewOpsDeploymentWorkflowService(
		b.repos.OpsDeploymentWorkflow,
		b.repos.OpsProjectBinding,
		opsDeploymentPreflightService,
	)
	opsDeploymentWorkflowService.ConfigureCachePurgeService(opsCloudflareCachePurgeService)
	opsDeploymentWorkflowService.ConfigureRollbackExecutor(service.NewOpsDeploymentSSHRollbackExecutorFromEnv())
	opsDeploymentWorkflowService.ConfigureProductionDependencies(
		b.repos.OpsVPSBinding,
		opsConnectorService,
		opsHostingerSyncService,
		opsDeploymentHealthCheckService,
	)
	lighthouseRunnerService := service.NewLighthouseRunnerService(
		b.repos.SiteQualityRuns,
		b.repos.SiteQualityFindings,
		service.LighthouseRunnerConfig{
			RunnerURL:           b.cfg.SiteQuality.RunnerURL,
			RunnerToken:         b.cfg.SiteQuality.RunnerToken,
			StorefrontBaseURL:   storefrontBaseURL,
			StorefrontTargetURL: siteQualityTargetOrigin,
		},
	)
	lighthouseRunnerService.ConfigureJobRepository(b.repos.SiteQualityJobs)
	siteQualityEngineService := service.NewSiteQualityEngineService(
		b.repos.SiteQualityTargets,
		b.repos.SiteQualityJobs,
		b.repos.SiteQualityRuns,
		b.repos.SiteQualityFindings,
		b.repos.StorefrontRouteCatalog,
		lighthouseRunnerService,
		service.SiteQualityEngineConfig{
			BaseURL:                  storefrontBaseURL,
			WorkerEnabled:            b.cfg.Worker.SiteQualityEnabled,
			AutoScanEnabled:          b.cfg.Worker.SiteQualityAutoScanEnabled,
			WorkerInterval:           time.Duration(b.cfg.Worker.SiteQualityDispatchIntervalSeconds) * time.Second,
			SampleCount:              b.cfg.Worker.SiteQualitySampleCount,
			RequiredConfirmations:    b.cfg.Worker.SiteQualityConfirmations,
			RequiredCleanEvaluations: b.cfg.Worker.SiteQualityCleanEvaluations,
			WorkerBatchLimit:         b.cfg.Worker.SiteQualityBatchLimit,
			LeaseTimeout:             time.Duration(b.cfg.Worker.SiteQualityLeaseTimeoutSeconds) * time.Second,
			ProviderConcurrency:      b.cfg.Worker.SiteQualityProviderConcurrency,
			ProviderRequestInterval:  time.Duration(b.cfg.Worker.SiteQualityProviderSpacingSeconds) * time.Second,
		},
	)
	mediaService := service.NewMediaService(b.repos.Media, storageSvc, settingService, storefrontBaseURL, b.cfg.MediaUpload.AccountStorageQuotaBytes)
	mediaService.ConfigureDerivativePresetRepository(b.repos.MediaDerivativePresets)
	mediaService.ConfigureDerivativeRebuildJobRepository(b.repos.MediaDerivativeRebuildJobs)
	siteLogoService := service.NewSiteLogoService(b.repos.SiteLogo, siteLogoStorageSvc, storefrontBaseURL)
	productService.ConfigureMediaService(mediaService)
	seoResourceService.ConfigureMediaService(mediaService)
	seoResourceService.ConfigureCanonicalBaseURL(storefrontBaseURL)
	storefrontRouteCatalogService := service.NewStorefrontRouteCatalogService(
		b.repos.StorefrontRouteCatalog,
		postService,
		productService,
		storefrontBaseURL,
		storefrontInternalOrigin,
	)
	storefrontRedirectRuleService := service.NewStorefrontRedirectRuleService(
		b.repos.StorefrontRedirectRules,
		b.repos.StorefrontRouteCatalog,
	)
	storefrontURLIssueService := service.NewStorefrontURLIssueService(
		b.repos.StorefrontURLIssues,
		b.repos.StorefrontRouteCatalog,
		b.repos.StorefrontRedirectRules,
	)
	storefrontURLSearchProfileService := service.NewStorefrontURLSearchProfileService(
		b.repos.StorefrontURLSearchProfiles,
		b.repos.StorefrontRouteCatalog,
	)
	preflightContentLinkService := service.NewPreflightContentLinkService(
		b.repos.PreflightContentLinks,
		b.repos.StorefrontRouteCatalog,
		postService,
		service.PreflightContentLinkConfig{
			BaseURL: storefrontBaseURL,
		},
	)
	storefrontRouteCatalogService.ConfigureIssueReconciler(storefrontURLIssueService)
	googleMerchantService := service.NewGoogleMerchantService(
		b.repos.GoogleMerchant,
		b.repos.Product,
		b.cfg.GoogleMerchant,
		storefrontBaseURL,
	)
	googleMerchantService.ConfigureOutboundHTTPResilience(
		outboundHTTPResilience.retry,
		outboundHTTPResilience.breaker,
	)
	googleIndexingService, err := service.NewGoogleIndexingService(
		productService,
		b.cfg.GoogleIndexing,
		storefrontBaseURL,
	)
	if err != nil {
		return fmt.Errorf("initialize Google Indexing service: %w", err)
	}
	googleIndexingService.ConfigureOutboundHTTPResilience(
		outboundHTTPResilience.retry,
		outboundHTTPResilience.breaker,
	)
	socialOAuthService := service.NewSocialOAuthService(b.repos.SocialOAuth, b.cfg.SocialOAuth)
	loyaltyProgramService := service.NewLoyaltyProgramService(b.repos.LoyaltyProgram)
	loyaltyProgramService.ConfigureCurrencyPolicy(currencyPolicyService)
	afterSalesService := service.NewAfterSalesService(b.repos.AfterSales, b.repos.Order, b.repos.AfterSalesRefundReview)
	afterSalesService.ConfigureUserRepository(b.repos.User)
	afterSalesService.ConfigureTxManager(txManager)
	afterSalesService.ConfigureAttachmentStorage(storageSvc)
	hotDataArchiveService := service.NewHotDataArchiveService(
		b.repos.HotDataArchive,
		b.cfg.Worker.HotDataArchiveBatchLimit,
	)
	authService := service.NewAuthService(b.repos.User, b.cfg.JWT, b.cfg.OAuth)
	authService.ConfigureGoogleOAuthResilience(
		outboundHTTPResilience.retry,
		outboundHTTPResilience.breaker,
	)

	globalIPBlockService := service.NewGlobalIPBlockService(b.repos.GlobalIPBlockRule)
	globalIPBlockService.ConfigureAuditRepository(b.repos.Audit)
	globalIPBlockService.ConfigureCacheInvalidation(b.redisCache.Client())
	visitorProfileService := service.NewVisitorProfileService(b.repos.VisitorProfile)
	visitorProfileService.ConfigureIPAddressRetentionDays(b.cfg.Worker.VisitorProfileIPAddressRetentionDays)
	visitorProfileService.ConfigureGlobalIPBlockService(globalIPBlockService)

	b.services = Services{
		Auth:                              authService,
		AdminAccountMaintenance:           service.NewAdminAccountMaintenanceService(b.db),
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
		Cart:                              service.NewCartService(b.repos.Cart, b.repos.Product),
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
		OpsVPSBinding:                     service.NewOpsVPSBindingService(b.repos.OpsVPSBinding, b.repos.OpsConnector, b.repos.OpsProjectBinding),
		OpsProjectBinding:                 service.NewOpsProjectBindingService(b.repos.OpsProjectBinding, b.repos.OpsVPSBinding, b.repos.OpsConnector),
		OpsNetworkSummary:                 opsNetworkSummaryService,
		OpsHostingerSync:                  opsHostingerSyncService,
		OpsDeploymentPreflight:            opsDeploymentPreflightService,
		OpsDeploymentHealthCheck:          opsDeploymentHealthCheckService,
		OpsCloudflareCachePurge:           opsCloudflareCachePurgeService,
		OpsDeploymentWorkflow:             opsDeploymentWorkflowService,
		StorefrontContext:                 service.NewStorefrontContextServiceWithMarkets(currencyPolicyService, storefrontMarketService),
		GoogleMerchant:                    googleMerchantService,
		SocialOAuth:                       socialOAuthService,
		FAQ:                               service.NewFAQService(b.repos.FAQ, storageSvc),
		Gallery:                           service.NewGalleryService(b.repos.Gallery, b.repos.Media),
		Media:                             mediaService,
		SiteLogo:                          siteLogoService,
		Warranty:                          service.NewWarrantyService(b.repos.Warranty, b.repos.Order),
		ShipmentRecord:                    service.NewShipmentRecordService(b.repos.ShipmentRecord),
		Checkout:                          service.NewCheckoutService(b.repos.Product, b.repos.Coupon, b.repos.Payment, b.repos.Loyalty, shippingService),
		AfterSales:                        afterSalesService,
		Marketing:                         service.NewMarketingService(txManager, b.repos.Coupon, b.repos.Loyalty, settingService),
		LoyaltyProgram:                    loyaltyProgramService,
		Review:                            service.NewReviewService(b.repos.Review),
		ReviewModeration:                  service.NewReviewModerationService(b.repos.Review),
		Ticket:                            service.NewTicketService(b.repos.Ticket, b.repos.User, b.repos.FAQ),
		CustomerServiceEvents:             service.NewCustomerServiceEventHub(),
		Subscription:                      service.NewSubscriptionService(b.repos.Subscription),
		Sitemap:                           service.NewSitemapService(b.repos.Post, b.cfg.Server.BaseURL),
		StorefrontRouteCatalog:            storefrontRouteCatalogService,
		StorefrontURLSearchProfiles:       storefrontURLSearchProfileService,
		StorefrontRedirectRules:           storefrontRedirectRuleService,
		StorefrontURLIssues:               storefrontURLIssueService,
		PreflightContentLinks:             preflightContentLinkService,
		LighthouseRunner:                  lighthouseRunnerService,
		SiteQualityEngine:                 siteQualityEngineService,
		HotDataArchive:                    hotDataArchiveService,
		UGCShowcase:                       service.NewUGCShowcaseService(b.repos.UGCShowcase, storageSvc),
		HomeVisualTiles:                   service.NewHomeVisualTileService(b.repos.HomeVisualTiles, storageSvc),
		UGCShowcaseUploadProtection:       UGCShowcaseUploadProtectionService,
		UGCShowcaseUploadEligibility:      UGCShowcaseUploadEligibilityService,
		Wishlist:                          service.NewWishlistService(b.repos.Wishlist, b.repos.Product),
		Feedback:                          service.NewFeedbackService(b.repos.Feedback),
		SuggestionFeedback: service.NewSuggestionFeedbackService(
			b.repos.SuggestionFeedback,
		),
		User:      service.NewUserService(b.repos.User),
		Dashboard: service.NewDashboardService(b.repos.Order, b.repos.User, b.repos.Subscription),
		Audit:     service.NewAuditService(b.repos.Audit),
		OpsOverview: service.NewOpsOverviewService(
			b.repos.OpsDomainBinding,
			b.repos.OpsConnector,
			b.repos.OpsVPSBinding,
			b.repos.OpsProjectBinding,
			service.NewAuditService(b.repos.Audit),
		),
		Shipping:                  shippingService,
		Spoke:                     service.NewSpokeService(b.repos.Spoke),
		QuickBuy:                  service.NewQuickBuyService(b.repos.QuickBuy, b.repos.Product, b.repos.ProductCategory),
		SelectionAssistant:        service.NewSelectionAssistantService(b.repos.SelectionAssistant),
		SelectionConfigurationKey: service.NewSelectionConfigurationKeyService(b.repos.SelectionConfigurationKey),
		WheelsetFitQuestionnaire: service.NewWheelsetFitQuestionnaireService(
			b.repos.WheelsetFitQuestionnaire,
			b.repos.SelectionConfigurationKey,
		),
		VisitorProfile: visitorProfileService,
		GlobalIPBlock:  globalIPBlockService,
		BehaviorEvents: service.NewBehaviorEventService(
			b.repos.RecommendationEvent,
			b.cfg.BehaviorEvents,
		),
		VisitorRisk: service.NewVisitorRiskService(
			b.repos.VisitorRiskFact,
			b.cfg.VisitorRisk,
			b.cfg.JWT.Secret,
		),
		PaymentRiskMonitoring: service.NewPaymentRiskMonitoringService(
			b.repos.PaymentRisk,
			b.cfg.PaymentRiskMonitoring,
		),
		PaymentProtection: service.NewPaymentProtectionService(
			b.repos.PaymentProtection,
			b.cfg.PaymentProtection,
		),
		PaymentRefundReview: service.NewPaymentRefundRecommendationService(b.repos.PaymentRefundReview, txManager),
		Outbox:              service.NewOutboxService(b.repos.Outbox),
	}
	return nil
}
