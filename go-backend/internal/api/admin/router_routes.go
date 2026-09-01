package admin

import (
	"os"
	"strings"

	seoapi "commerce-platform/internal/api/admin/seo"
	urlapi "commerce-platform/internal/api/admin/urlmanagement"
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/app"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/securecookie"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterAdminRoutes 注册管理后台路由
func RegisterAdminRoutes(r *gin.Engine, deps *app.Dependencies, cfg *config.Config) {
	// 初始化 repositories
	services := deps.Services
	authService := services.Auth
	ugcShowcaseService := services.UGCShowcase
	warrantyService := services.Warranty
	userService := services.User
	postService := services.Post
	productService := services.Product
	productProcurementService := services.ProductProcurement
	frameFitmentEntryService := services.FrameFitmentEntry
	forkFitmentEntryService := services.ForkFitmentEntry
	fitmentHubSpecificationService := services.FitmentHubSpecification
	productProfitabilityService := services.ProductProfitability
	orderService := services.Order
	afterSalesService := services.AfterSales
	paymentService := services.Payment
	marketingService := services.Marketing
	dashboardService := services.Dashboard

	// 初始化 handlers
	cookieOptions := securecookie.Options{
		Secure:   cfg.Cookie.SecureEnabled(cfg.Server),
		SameSite: cfg.Cookie.SameSiteMode(),
		Domain:   cfg.Cookie.Domain,
	}
	authHandler := NewAuthHandler(authService, cookieOptions)
	adminAccountHandler := NewAdminAccountHandler(services.AdminAccountMaintenance)
	dashboardHandler := NewDashboardHandler(dashboardService)
	userHandler := NewUserHandler(userService)
	customerHandler := NewCustomerHandler(userService)
	productHandler := NewProductHandler(productService)
	productProcurementHandler := NewProductProcurementHandler(productProcurementService)
	frameFitmentEntryHandler := NewFrameFitmentEntryHandler(frameFitmentEntryService)
	forkFitmentEntryHandler := NewForkFitmentEntryHandler(forkFitmentEntryService)
	fitmentHubSpecificationHandler := NewFitmentHubSpecificationHandler(fitmentHubSpecificationService)
	productProfitabilityHandler := NewProductProfitabilityHandler(productProfitabilityService)
	productCategoryHandler := NewProductCategoryHandler(services.ProductCategory)
	productBrandHandler := NewProductBrandHandler(services.ProductBrand)
	productInformationTemplateHandler := NewProductInformationTemplateHandler(services.ProductInformationTemplate)
	customsClassificationHandler := NewCustomsClassificationHandler(services.CustomsClassification)
	spokeCatalogHandler := NewSpokeCatalogHandler(services.Spoke)
	quickBuyHandler := NewQuickBuyHandler(services.QuickBuy)
	selectionAssistantHandler := NewSelectionAssistantHandler(services.SelectionAssistant)
	selectionConfigurationKeyHandler := NewSelectionConfigurationKeyHandler(services.SelectionConfigurationKey)
	wheelsetFitQuestionnaireHandler := NewWheelsetFitQuestionnaireHandler(services.WheelsetFitQuestionnaire, services.Product)
	mediaHandler := NewMediaHandler(services.Media)
	mediaHandler.ConfigureAuditService(services.Audit)
	mediaImageDimensionsHandler := NewMediaImageDimensionsHandler(service.NewMediaImageDimensionEngine(services.Media))
	fontPreflightHandler := NewFontPreflightHandler(fontPreflightStorefrontOrigin(cfg), nil)
	contentLinkPreflightHandler := NewContentLinkPreflightHandler(services.PreflightContentLinks)
	orderHandler := NewOrderHandler(orderService)
	orderHandler.ConfigureAuditService(services.Audit)
	afterSalesHandler := NewAfterSalesHandler(afterSalesService)
	paymentHandler := NewPaymentHandler(paymentService, services.AdminSettings, services.PayPalDisputeInvoiceSellerProfile)
	paymentHandler.ConfigurePublicBaseURL(cfg.Server.BaseURL)
	paymentHandler.ConfigureAuditService(services.Audit)
	paymentRefundExecutionHandler := NewPaymentRefundExecutionHandler(paymentService, services.AdminSettings)
	paymentRefundExecutionHandler.ConfigureAuditService(services.Audit)
	paymentRiskMonitoringHandler := NewPaymentRiskMonitoringHandler(services.PaymentRiskMonitoring)
	paymentRiskMonitoringHandler.ConfigureRiskConfiguration(
		cfg,
		services.PaymentThreeDS,
		services.PaymentProtection,
		deps.PaymentGatewayCircuitBreaker,
	)
	paymentRiskMonitoringHandler.ConfigureGatewayRuntimeReader(paymentHandler.RuntimeReadiness)
	paymentRiskMonitoringHandler.ConfigureAuditService(services.Audit)
	paymentProtectionHandler := NewPaymentProtectionHandler(services.PaymentProtection)
	paymentProtectionHandler.ConfigureAuditService(services.Audit)
	paymentRefundRecommendationHandler := NewPaymentRefundRecommendationHandler(services.PaymentRefundReview)
	paymentRefundRecommendationHandler.ConfigureAuditService(services.Audit)
	contentHandler := NewContentHandler(postService)
	faqHandler := NewFAQHandler(services.FAQ)
	galleryHandler := NewGalleryHandler(services.Gallery)
	subscriptionHandler := NewSubscriptionHandler(services.Subscription)
	ticketHandler := NewTicketHandler(services.Ticket, services.CustomerServiceContext, services.CustomerServiceAnalytics, services.CustomerServiceEvents, services.Media)
	ticketHandler.ConfigureAllowedOrigins(cfg.CORS.AllowedOrigins)
	autoReplyHandler := NewAutoReplyHandler(services.Ticket, services.FAQ)
	visitorProfileHandler := NewVisitorProfileHandler(services.VisitorProfile)
	visitorProfileHandler.ConfigureAuditService(services.Audit)
	visitorRiskHandler := NewVisitorRiskHandler(services.VisitorRisk)
	visitorRiskHandler.ConfigureAuditService(services.Audit)
	globalIPBlockHandler := NewGlobalIPBlockHandler(services.GlobalIPBlock)
	globalIPBlockHandler.ConfigureAuditService(services.Audit)
	marketingHandler := NewMarketingHandler(marketingService, services.LoyaltyProgram)
	settingsHandler := NewSettingsHandler(services.AdminSettings)
	refundReturnPolicyHandler := NewRefundReturnPolicyHandler(services.RefundReturnPolicy)
	siteLogoHandler := NewSiteLogoHandler(services.SiteLogo, services.AdminSettings)
	homeVisualTileHandler := NewHomeVisualTileHandler(services.HomeVisualTiles)
	websiteProfileHandler := NewWebsiteProfileHandler(services.WebsiteProfile)
	websiteNameHandler := NewWebsiteNameHandler(services.WebsiteName)
	seoHomeHandler := seoapi.NewHomeHandler(services.SEO)
	seoArticlesHandler := seoapi.NewArticlesHandler(services.SEOResources)
	seoProductsHandler := seoapi.NewProductsHandler(services.SEOResources)
	seoCategoriesHandler := seoapi.NewCategoriesHandler(services.SEOResources)
	urlRoutesHandler := urlapi.NewRoutesHandler(services.StorefrontRouteCatalog)
	urlSearchProfilesHandler := urlapi.NewSearchProfilesHandler(services.StorefrontURLSearchProfiles)
	urlRedirectsHandler := urlapi.NewRedirectsHandler(services.StorefrontRedirectRules)
	urlIssuesHandler := urlapi.NewIssuesHandler(services.StorefrontURLIssues, services.StorefrontRouteCatalog)
	seoHomeHandler.ConfigureAuditService(services.Audit)
	seoArticlesHandler.ConfigureAuditService(services.Audit)
	seoProductsHandler.ConfigureAuditService(services.Audit)
	seoCategoriesHandler.ConfigureAuditService(services.Audit)
	if services.GoogleIndexing != nil {
		services.GoogleIndexing.ConfigureRedisClient(deps.RedisClient)
	}
	seoProductsHandler.ConfigureGoogleIndexingService(services.GoogleIndexing)
	urlIssuesHandler.ConfigureAuditService(services.Audit)
	analyticsHandler := NewAnalyticsHandler(services.Analytics)
	commercialCrawlerHandler := NewCommercialCrawlerProtectionHandler(orderService)
	currencyPolicyHandler := NewCurrencyPolicyHandler(services.CurrencyPolicy)
	currencyPolicyHandler.ConfigureAuditService(services.Audit)
	currencyPolicyHandler.ConfigureProductService(services.Product)
	exchangeRateHandler := NewExchangeRateHandler(services.ExchangeRate)
	exchangeRateHandler.ConfigureAuditService(services.Audit)
	storefrontMarketHandler := NewStorefrontMarketHandler(services.StorefrontMarket)
	storefrontMarketHandler.ConfigureAuditService(services.Audit)
	googleMerchantHandler := NewGoogleMerchantHandler(services.GoogleMerchant, cfg.GoogleMerchant.PostConnectURL)
	socialOAuthHandler := NewSocialOAuthHandler(services.SocialOAuth, cfg.SocialOAuth.PostConnectURL)
	publicChatAgentHandler := NewPublicChatAgentHandler(services.AdminPublicChat)
	customerServiceAvatarHandler := NewCustomerServiceAvatarHandler(services.CustomerServiceAvatar)
	auditHandler := NewAuditHandler(services.Audit)
	ugcShowcaseHandler := NewUGCShowcaseHandler(ugcShowcaseService)
	ugcShowcaseHandler.ConfigureAuditService(services.Audit)
	warrantyHandler := NewWarrantyHandler(warrantyService)
	shipmentRecordHandler := NewShipmentRecordHandler(services.ShipmentRecord, deps.Storage)
	shippingHandler := NewShippingHandler(services.Shipping)
	opsDomainBindingHandler := NewOpsDomainBindingHandler(
		services.OpsDomainBinding,
		services.OpsDomainDiff,
		services.OpsDomainPreview,
		services.OpsDomainSync,
	)
	opsDomainBindingHandler.ConfigureAuditService(services.Audit)
	opsConnectorHandler := NewOpsConnectorHandler(services.OpsConnector)
	opsConnectorHandler.ConfigureAuditService(services.Audit)
	opsConnectorHandler.ConfigureOAuthService(services.OpsConnectorOAuth)
	serviceCenterHandler := NewServiceCenterHandler(
		services.OpsConnector,
		services.OpsDomainBinding,
		opsConnectorHandler,
	)
	serviceCenterHandler.ConfigureOpsOverviewService(services.OpsOverview)
	serviceCenterHandler.ConfigureOpsNetworkSummaryService(services.OpsNetworkSummary)
	serviceCenterHandler.ConfigureCloudflareCacheRulesService(
		service.NewCloudflareCacheRulesService(services.OpsConnector),
	)
	serviceCenterHandler.ConfigureAuditService(services.Audit)
	siteQualityHandler := NewSiteQualityHandler(services.LighthouseRunner, services.SiteQualityEngine)
	siteQualityHandler.ConfigureAuditService(services.Audit)
	opsVPSBindingHandler := NewOpsVPSBindingHandler(services.OpsVPSBinding)
	opsVPSBindingHandler.ConfigureAuditService(services.Audit)
	opsVPSBindingHandler.ConfigureSyncService(services.OpsHostingerSync)
	opsProjectBindingHandler := NewOpsProjectBindingHandler(services.OpsProjectBinding)
	opsProjectBindingHandler.ConfigureAuditService(services.Audit)
	opsProjectBindingHandler.ConfigureSyncService(services.OpsHostingerSync)
	opsDeploymentPreflightHandler := NewOpsDeploymentPreflightHandler(services.OpsDeploymentPreflight)
	opsDeploymentWorkflowHandler := NewOpsDeploymentWorkflowHandler(services.OpsDeploymentWorkflow)
	opsDeploymentWorkflowHandler.ConfigureAuditService(services.Audit)
	opsOverviewHandler := NewOpsOverviewHandler(services.OpsOverview)
	opsNetworkSummaryHandler := NewOpsNetworkSummaryHandler(services.OpsNetworkSummary)
	outboxReconciliationHandler := NewOutboxReconciliationHandler(services.Outbox)
	outboxReconciliationHandler.ConfigureAuditService(services.Audit)
	reviewModerationHandler := NewReviewModerationHandler(services.ReviewModeration)
	reviewModerationHandler.ConfigureAuditService(services.Audit)
	pageFeedbackHandler := NewPageFeedbackHandler(services.Feedback)
	pageFeedbackHandler.ConfigureAuditService(services.Audit)
	pageFeedbackHandler.ConfigureRedisClient(deps.RedisClient)

	// 管理后台 API 路由组
	admin := r.Group("/api/admin")
	admin.Use(middleware.CSRFProtection(cfg.CORS.AllowedOrigins))
	registerPublicAdminRoutes(admin, authHandler, googleMerchantHandler, opsConnectorHandler, socialOAuthHandler)

	// 需要认证的路由
	authenticated := admin.Group("")
	authenticated.Use(middleware.AuthMiddleware(authService), middleware.RequireBackofficeAccess())

	storefrontGroup := authenticated.Group("/storefront")
	storefrontGroup.Use(middleware.RequirePermission(auth.PermSettingsView))
	{
		marketGroup := storefrontGroup.Group("/markets")
		{
			marketGroup.GET("", storefrontMarketHandler.ListMarkets)
			marketGroup.GET("/options", storefrontMarketHandler.GetOptions)
			marketGroup.GET("/:id", storefrontMarketHandler.GetMarket)
			marketGroup.POST("", middleware.RequirePermission(auth.PermSettingsEdit), storefrontMarketHandler.CreateMarket)
			marketGroup.PUT("/:id", middleware.RequirePermission(auth.PermSettingsEdit), storefrontMarketHandler.UpdateMarket)
			marketGroup.DELETE("/:id", middleware.RequirePermission(auth.PermSettingsEdit), storefrontMarketHandler.DeleteMarket)
		}
	}

	registerAuthenticatedCoreRoutes(authenticated, authHandler, dashboardHandler, userHandler, customerHandler)
	registerProductRoutes(
		authenticated,
		productHandler,
		productProcurementHandler,
		frameFitmentEntryHandler,
		forkFitmentEntryHandler,
		fitmentHubSpecificationHandler,
		productProfitabilityHandler,
		productCategoryHandler,
		productBrandHandler,
		productInformationTemplateHandler,
		customsClassificationHandler,
		spokeCatalogHandler,
		quickBuyHandler,
		selectionAssistantHandler,
		selectionConfigurationKeyHandler,
		wheelsetFitQuestionnaireHandler,
	)
	registerIntegrationRoutes(authenticated, googleMerchantHandler, socialOAuthHandler)
	registerMediaAndPreflightRoutes(
		authenticated,
		mediaHandler,
		mediaImageDimensionsHandler,
		fontPreflightHandler,
		contentLinkPreflightHandler,
		siteQualityHandler,
	)
	registerCommerceRoutes(
		authenticated,
		orderHandler,
		afterSalesHandler,
		paymentHandler,
		paymentRefundExecutionHandler,
		paymentRiskMonitoringHandler,
		paymentProtectionHandler,
		paymentRefundRecommendationHandler,
	)

	// 内容管理（需要内容管理权限）
	contentGroup := authenticated.Group("/content")
	contentGroup.Use(middleware.RequirePermission(auth.PermContentView))
	{
		// 文章管理
		postsGroup := contentGroup.Group("/posts")
		{
			postsGroup.GET("", contentHandler.ListPosts)
			postsGroup.GET("/stats", contentHandler.GetPostStats)
			postsGroup.GET("/:id", contentHandler.GetPost)
			postsGroup.GET("/:id/translations", contentHandler.GetTranslations)
			postsGroup.POST("", middleware.RequirePermission(auth.PermContentCreate), contentHandler.CreatePost)
			postsGroup.PUT("/:id", middleware.RequirePermission(auth.PermContentEdit), contentHandler.UpdatePost)
			postsGroup.PATCH("/:id/status", middleware.RequirePermission(auth.PermContentEdit), contentHandler.UpdatePostStatus)
			postsGroup.DELETE("/:id", middleware.RequirePermission(auth.PermContentDelete), contentHandler.DeletePost)
			postsGroup.POST("/batch-status", middleware.RequirePermission(auth.PermContentEdit), contentHandler.BatchUpdateStatus)
			postsGroup.POST("/batch-delete", middleware.RequirePermission(auth.PermContentDelete), contentHandler.BatchDelete)
		}

		feedbackGroup := contentGroup.Group("/feedback")
		{
			feedbackGroup.GET("", pageFeedbackHandler.List)
			feedbackGroup.GET("/risk-overview", pageFeedbackHandler.RiskOverview)
			feedbackGroup.GET("/:id", pageFeedbackHandler.Get)
			feedbackGroup.PATCH("/:id", middleware.RequirePermission(auth.PermContentEdit), pageFeedbackHandler.Update)
		}

		pageFeedbackGroup := contentGroup.Group("/page-feedback")
		{
			pageFeedbackGroup.GET("", pageFeedbackHandler.List)
			pageFeedbackGroup.GET("/risk-overview", pageFeedbackHandler.RiskOverview)
			pageFeedbackGroup.GET("/:id", pageFeedbackHandler.Get)
			pageFeedbackGroup.PATCH("/:id", middleware.RequirePermission(auth.PermContentEdit), pageFeedbackHandler.Update)
		}

		homeVisualTileGroup := contentGroup.Group("/visual-showcases")
		{
			homeVisualTileGroup.GET("/:showcase_key", homeVisualTileHandler.GetItems)
			homeVisualTileGroup.POST("/:showcase_key/assets", middleware.RequirePermission(auth.PermContentEdit), middleware.RateLimitByUserPerMinute(3, 2), homeVisualTileHandler.UploadImage)
			homeVisualTileGroup.PUT("/:showcase_key", middleware.RequirePermission(auth.PermContentEdit), homeVisualTileHandler.ReplaceItems)
		}

		contentGroup.GET("/refund-return-policy", refundReturnPolicyHandler.Get)
		contentGroup.PUT("/refund-return-policy", middleware.RequirePermission(auth.PermContentEdit), refundReturnPolicyHandler.Update)
		contentGroup.POST(
			"/refund-return-policy/assets",
			middleware.RequirePermission(auth.PermContentEdit),
			middleware.RateLimitByUserPerMinute(3, 2),
			mediaHandler.UploadRefundReturnPolicyImage,
		)
	}

	// 商品评价审核（独立 review 域）
	reviewsGroup := authenticated.Group("/reviews")
	reviewsGroup.Use(middleware.RequirePermission(auth.PermReviewView))
	{
		reviewsGroup.GET("", reviewModerationHandler.List)
		reviewsGroup.GET("/:id", reviewModerationHandler.Get)
		reviewsGroup.PATCH("/:id/status", middleware.RequirePermission(auth.PermReviewModerate), reviewModerationHandler.UpdateStatus)
	}

	// FAQ 管理（需要 FAQ 管理权限）
	faqsGroup := authenticated.Group("/faqs")
	faqsGroup.Use(middleware.RequirePermission(auth.PermFAQView))
	{
		faqsGroup.GET("", faqHandler.ListFAQs)
		faqsGroup.GET("/grouped", faqHandler.ListFAQGroups)
		faqsGroup.GET("/structure", faqHandler.ListStructure)
		faqsGroup.GET("/categories", faqHandler.GetCategories)
		faqsGroup.POST("/categories", middleware.RequirePermission(auth.PermFAQCreate), faqHandler.CreateCategory)
		faqsGroup.PUT("/categories/:id", middleware.RequirePermission(auth.PermFAQEdit), faqHandler.UpdateCategory)
		faqsGroup.DELETE("/categories/:id", middleware.RequirePermission(auth.PermFAQDelete), faqHandler.DeleteCategory)
		faqsGroup.PUT("/pages/:page_id", middleware.RequirePermission(auth.PermFAQEdit), faqHandler.UpdatePage)
		faqsGroup.POST(
			"/answer-image",
			middleware.RequireAnyPermission(auth.PermFAQCreate, auth.PermFAQEdit),
			middleware.RateLimitByUserPerMinute(6, 2),
			faqHandler.UploadAnswerImage,
		)
		faqsGroup.GET("/:id", faqHandler.GetFAQ)
		faqsGroup.POST("", middleware.RequirePermission(auth.PermFAQCreate), faqHandler.CreateFAQ)
		faqsGroup.PUT("/:id", middleware.RequirePermission(auth.PermFAQEdit), faqHandler.UpdateFAQ)
		faqsGroup.PATCH("/:id/order", middleware.RequirePermission(auth.PermFAQEdit), faqHandler.UpdateOrder)
		faqsGroup.DELETE("/:id", middleware.RequirePermission(auth.PermFAQDelete), faqHandler.DeleteFAQ)
		faqsGroup.POST("/batch-delete", middleware.RequirePermission(auth.PermFAQDelete), faqHandler.BatchDelete)
	}

	// 图库管理（需要图库管理权限）
	galleriesGroup := authenticated.Group("/galleries")
	galleriesGroup.Use(middleware.RequirePermission(auth.PermGalleryView))
	{
		galleriesGroup.GET("", galleryHandler.ListGalleries)
		galleriesGroup.GET("/:id", galleryHandler.GetGallery)
		galleriesGroup.POST("", middleware.RequirePermission(auth.PermGalleryCreate), galleryHandler.CreateGallery)
		galleriesGroup.PUT("/:id", middleware.RequirePermission(auth.PermGalleryEdit), galleryHandler.UpdateGallery)
		galleriesGroup.DELETE("/:id", middleware.RequirePermission(auth.PermGalleryDelete), galleryHandler.DeleteGallery)

		// 图片管理
		galleriesGroup.GET("/:id/images", galleryHandler.ListImages)
		galleriesGroup.POST("/:id/images", middleware.RequirePermission(auth.PermGalleryCreate), galleryHandler.CreateImage)
		galleriesGroup.PUT("/:id/images/:imageId", middleware.RequirePermission(auth.PermGalleryEdit), galleryHandler.UpdateImage)
		galleriesGroup.DELETE("/:id/images/:imageId", middleware.RequirePermission(auth.PermGalleryDelete), galleryHandler.DeleteImage)
		galleriesGroup.POST("/:id/images/batch-delete", middleware.RequirePermission(auth.PermGalleryDelete), galleryHandler.BatchDeleteImages)
	}

	// 买家秀审批管理（需要图库管理权限）
	showcaseGroup := authenticated.Group("/showcase")
	showcaseGroup.Use(middleware.RequirePermission(auth.PermGalleryView))
	{
		showcaseGroup.GET("", ugcShowcaseHandler.List)
		showcaseGroup.GET("/:id/images/:image_index/file", ugcShowcaseHandler.ServeImageFile)
		showcaseGroup.PUT("/:id/approve", middleware.RequirePermission(auth.PermGalleryEdit), ugcShowcaseHandler.Approve)
		showcaseGroup.PUT("/:id/reject", middleware.RequirePermission(auth.PermGalleryEdit), ugcShowcaseHandler.Reject)
	}

	// 保修与已发货订单附加凭据（需要商品管理权限）
	warrantyGroup := authenticated.Group("/warranty")
	warrantyGroup.Use(middleware.RequirePermission(auth.PermProductView))
	{
		warrantyGroup.GET("/shipment-records", shipmentRecordHandler.List)
		warrantyGroup.GET("/shipment-records/stats", shipmentRecordHandler.Stats)
		warrantyGroup.GET("/shipment-records/:id", shipmentRecordHandler.Get)
		warrantyGroup.PUT("/shipment-records/:id", middleware.RequirePermission(auth.PermProductEdit), shipmentRecordHandler.Update)
		warrantyGroup.POST("/shipment-records/:id/images", middleware.RequirePermission(auth.PermProductEdit), shipmentRecordHandler.UploadImages)
		warrantyGroup.GET("/claims", warrantyHandler.ListAllWarrantyClaims)
		warrantyGroup.GET("/claims/:id", warrantyHandler.GetWarrantyClaim)
		warrantyGroup.GET("/claims/:id/order-items", warrantyHandler.ListWarrantyClaimOrderItems)
		warrantyGroup.PUT("/claims/:id/order-item", middleware.RequirePermission(auth.PermProductEdit), warrantyHandler.BindWarrantyClaimOrderItem)
		warrantyGroup.GET("/claims/:id/service-records", warrantyHandler.ListWarrantyServiceRecords)
		warrantyGroup.POST("/claims/:id/service-records", middleware.RequirePermission(auth.PermProductEdit), warrantyHandler.CreateWarrantyServiceRecord)
		warrantyGroup.PUT("/claims/:id/status", middleware.RequirePermission(auth.PermProductEdit), warrantyHandler.UpdateWarrantyClaimStatus)
		warrantyGroup.PUT("/claims/:id/resolution", middleware.RequirePermission(auth.PermProductEdit), warrantyHandler.UpdateWarrantyClaimResolution)
	}

	registerContentRoutes(
		authenticated,
		contentHandler,
		pageFeedbackHandler,
		homeVisualTileHandler,
		refundReturnPolicyHandler,
		mediaHandler,
		reviewModerationHandler,
		faqHandler,
		galleryHandler,
		ugcShowcaseHandler,
		warrantyHandler,
		shipmentRecordHandler,
		subscriptionHandler,
		ticketHandler,
		autoReplyHandler,
		visitorProfileHandler,
		visitorRiskHandler,
		globalIPBlockHandler,
	)
	registerBusinessRoutes(
		authenticated,
		marketingHandler,
		exchangeRateHandler,
		seoHomeHandler,
		seoArticlesHandler,
		seoProductsHandler,
		seoCategoriesHandler,
		urlRoutesHandler,
		urlSearchProfilesHandler,
		urlRedirectsHandler,
		urlIssuesHandler,
		analyticsHandler,
		serviceCenterHandler,
		deps.RedisClient,
	)
	registerSystemRoutes(
		authenticated,
		settingsHandler,
		commercialCrawlerHandler,
		publicChatAgentHandler,
		customerServiceAvatarHandler,
		paymentHandler,
		currencyPolicyHandler,
		siteLogoHandler,
		exchangeRateHandler,
		websiteProfileHandler,
		websiteNameHandler,
		shippingHandler,
	)
	registerOperationsRoutes(
		authenticated,
		adminAccountHandler,
		opsOverviewHandler,
		opsNetworkSummaryHandler,
		outboxReconciliationHandler,
		opsDeploymentPreflightHandler,
		opsDeploymentWorkflowHandler,
		opsDomainBindingHandler,
		opsConnectorHandler,
		opsVPSBindingHandler,
		opsProjectBindingHandler,
		auditHandler,
	)
}

func fontPreflightStorefrontOrigin(cfg *config.Config) string {
	if origin := strings.TrimRight(strings.TrimSpace(os.Getenv("STOREFRONT_INTERNAL_ORIGIN")), "/"); origin != "" {
		return origin
	}

	// Local development may run the API on the host while Nuxt listens on 9199.
	// Containerized development supplies the service origin explicitly.
	if cfg == nil || !strings.EqualFold(strings.TrimSpace(cfg.Server.Mode), gin.ReleaseMode) {
		if origin := strings.TrimRight(strings.TrimSpace(os.Getenv("STOREFRONT_BASE_URL")), "/"); origin != "" {
			return origin
		}
		return "http://localhost:9199"
	}

	// Production must use an internal storefront origin. The public edge
	// intentionally returns 404 for /_internal/*.
	return ""
}
