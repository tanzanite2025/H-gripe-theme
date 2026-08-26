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
	showcaseService := services.Showcase
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
	visitorRiskHandler := NewVisitorRiskHandler(services.VisitorRisk)
	visitorRiskHandler.ConfigureAuditService(services.Audit)
	marketingHandler := NewMarketingHandler(marketingService, services.LoyaltyProgram)
	settingsHandler := NewSettingsHandler(services.AdminSettings)
	refundReturnPolicyHandler := NewRefundReturnPolicyHandler(services.RefundReturnPolicy)
	siteLogoHandler := NewSiteLogoHandler(services.SiteLogo, services.AdminSettings)
	visualShowcaseHandler := NewVisualShowcaseHandler(services.VisualShowcase)
	websiteProfileHandler := NewWebsiteProfileHandler(services.WebsiteProfile)
	websiteNameHandler := NewWebsiteNameHandler(services.WebsiteName)
	seoHomeHandler := seoapi.NewHomeHandler(services.SEO)
	seoArticlesHandler := seoapi.NewArticlesHandler(services.SEOResources)
	seoProductsHandler := seoapi.NewProductsHandler(services.SEOResources)
	seoCategoriesHandler := seoapi.NewCategoriesHandler(services.SEOResources)
	urlRoutesHandler := urlapi.NewRoutesHandler(services.StorefrontRouteCatalog)
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
	showcaseHandler := NewShowcaseHandler(showcaseService)
	showcaseHandler.ConfigureAuditService(services.Audit)
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
	{
		// 认证路由（公开）
		authGroup := admin.Group("/auth")
		authGroup.Use(middleware.RateLimit(10)) // 10 RPS for auth endpoints
		{
			authGroup.GET("/config", authHandler.GetAuthConfig)
			authGroup.POST("/login", authHandler.AdminLogin)
			authGroup.POST("/google-login", authHandler.AdminGoogleLogin)
			authGroup.POST("/refresh", authHandler.RefreshToken)
		}

		admin.GET("/google-merchant/oauth/callback", googleMerchantHandler.CompleteOAuth)
		admin.GET("/ops/connectors/oauth/callback", opsConnectorHandler.CompleteOAuth)
		admin.GET("/social/oauth/callback", socialOAuthHandler.CompleteOAuth)

		// 需要认证的路由
		authenticated := admin.Group("")
		authenticated.Use(middleware.AuthMiddleware(authService), middleware.RequireBackofficeAccess())
		{
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

			// 认证相关
			authGroup := authenticated.Group("/auth")
			{
				authGroup.GET("/profile", authHandler.GetProfile)
				authGroup.POST("/logout", authHandler.Logout)
				authGroup.GET("/permissions", authHandler.GetPermissions)
			}

			// 仪表板（所有管理员都可以访问）
			dashboardGroup := authenticated.Group("/dashboard")
			dashboardGroup.Use(middleware.RequireAnyPermission(auth.PermOrderView, auth.PermUserView, auth.PermTicketView, auth.PermSubscriptionView))
			{
				dashboardGroup.GET("/stats", dashboardHandler.GetStats)
				dashboardGroup.GET("/recent-orders", dashboardHandler.GetRecentOrders)
				dashboardGroup.GET("/recent-users", dashboardHandler.GetRecentUsers)
				dashboardGroup.GET("/sales-chart", dashboardHandler.GetSalesChart)
			}

			// 用户管理（需要用户管理权限）
			usersGroup := authenticated.Group("/users")
			usersGroup.Use(middleware.RequirePermission(auth.PermUserView))
			{
				usersGroup.GET("", userHandler.ListUsers)
				usersGroup.GET("/stats", userHandler.GetUserStats)
				usersGroup.GET("/:id", userHandler.GetUser)
				usersGroup.POST("", middleware.RequirePermission(auth.PermUserCreate), userHandler.CreateUser)
				usersGroup.PUT("/:id", middleware.RequirePermission(auth.PermUserEdit), userHandler.UpdateUser)
				usersGroup.PATCH("/:id/status", middleware.RequirePermission(auth.PermUserEdit), userHandler.UpdateUserStatus)
				usersGroup.DELETE("/:id", middleware.RequirePermission(auth.PermUserDelete), userHandler.DeleteUser)
				usersGroup.POST("/batch-delete", middleware.RequirePermission(auth.PermUserDelete), userHandler.BatchDeleteUsers)
			}

			customersGroup := authenticated.Group("/customers")
			customersGroup.Use(middleware.RequirePermission(auth.PermUserView))
			{
				customersGroup.GET("", customerHandler.ListCustomers)
			}

			// 商品管理（需要商品管理权限）
			// 采购资料是独立附加域，只读写自身产品编码/名称快照。
			productProcurementGroup := authenticated.Group("/procurement/records")
			productProcurementGroup.Use(middleware.RequirePermission(auth.PermProcurementView))
			{
				productProcurementGroup.GET("", productProcurementHandler.List)
				productProcurementGroup.GET("/by-codes", productProcurementHandler.ListByCodes)
				productProcurementGroup.GET("/:id", productProcurementHandler.Get)
				productProcurementGroup.POST("", middleware.RequirePermission(auth.PermProcurementCreate), productProcurementHandler.Create)
				productProcurementGroup.PUT("/:id", middleware.RequirePermission(auth.PermProcurementEdit), productProcurementHandler.Update)
				productProcurementGroup.DELETE("/:id", middleware.RequirePermission(auth.PermProcurementDelete), productProcurementHandler.Delete)
			}

			productProcurementOptionsGroup := authenticated.Group("/procurement/product-options")
			productProcurementOptionsGroup.Use(middleware.RequirePermission(auth.PermProcurementView))
			{
				productProcurementOptionsGroup.GET("", productProcurementHandler.ProductOptions)
			}

			frameFitmentEntriesGroup := authenticated.Group("/fitment-catalog/frame-entries")
			frameFitmentEntriesGroup.Use(middleware.RequirePermission(auth.PermFitmentCatalogView))
			{
				frameFitmentEntriesGroup.GET("", frameFitmentEntryHandler.List)
				frameFitmentEntriesGroup.GET("/:id", frameFitmentEntryHandler.Get)
				frameFitmentEntriesGroup.POST("", middleware.RequirePermission(auth.PermFitmentCatalogCreate), frameFitmentEntryHandler.Create)
				frameFitmentEntriesGroup.PUT("/:id", middleware.RequirePermission(auth.PermFitmentCatalogEdit), frameFitmentEntryHandler.Update)
				frameFitmentEntriesGroup.PATCH("/:id/status", middleware.RequirePermission(auth.PermFitmentCatalogEdit), frameFitmentEntryHandler.UpdateStatus)
				frameFitmentEntriesGroup.DELETE("/:id", middleware.RequirePermission(auth.PermFitmentCatalogDelete), frameFitmentEntryHandler.Delete)
			}

			forkFitmentEntriesGroup := authenticated.Group("/fitment-catalog/fork-entries")
			forkFitmentEntriesGroup.Use(middleware.RequirePermission(auth.PermFitmentCatalogView))
			{
				forkFitmentEntriesGroup.GET("", forkFitmentEntryHandler.List)
				forkFitmentEntriesGroup.GET("/:id", forkFitmentEntryHandler.Get)
				forkFitmentEntriesGroup.POST("", middleware.RequirePermission(auth.PermFitmentCatalogCreate), forkFitmentEntryHandler.Create)
				forkFitmentEntriesGroup.PUT("/:id", middleware.RequirePermission(auth.PermFitmentCatalogEdit), forkFitmentEntryHandler.Update)
				forkFitmentEntriesGroup.PATCH("/:id/status", middleware.RequirePermission(auth.PermFitmentCatalogEdit), forkFitmentEntryHandler.UpdateStatus)
				forkFitmentEntriesGroup.DELETE("/:id", middleware.RequirePermission(auth.PermFitmentCatalogDelete), forkFitmentEntryHandler.Delete)
			}

			fitmentHubSpecificationsGroup := authenticated.Group("/fitment-catalog/hub-specifications")
			fitmentHubSpecificationsGroup.Use(middleware.RequirePermission(auth.PermFitmentCatalogView))
			{
				fitmentHubSpecificationsGroup.GET("", fitmentHubSpecificationHandler.List)
				fitmentHubSpecificationsGroup.GET("/:id", fitmentHubSpecificationHandler.Get)
				fitmentHubSpecificationsGroup.POST("", middleware.RequirePermission(auth.PermFitmentCatalogCreate), fitmentHubSpecificationHandler.Create)
				fitmentHubSpecificationsGroup.PUT("/:id", middleware.RequirePermission(auth.PermFitmentCatalogEdit), fitmentHubSpecificationHandler.Update)
				fitmentHubSpecificationsGroup.PATCH("/:id/status", middleware.RequirePermission(auth.PermFitmentCatalogEdit), fitmentHubSpecificationHandler.UpdateStatus)
				fitmentHubSpecificationsGroup.DELETE("/:id", middleware.RequirePermission(auth.PermFitmentCatalogDelete), fitmentHubSpecificationHandler.Delete)
			}

			productProfitabilityGroup := authenticated.Group("/procurement/profitability")
			productProfitabilityGroup.Use(middleware.RequirePermission(auth.PermProcurementView))
			{
				productProfitabilityGroup.GET("/by-codes", productProfitabilityHandler.ListByCodes)
				productProfitabilityGroup.POST("/preview", productProfitabilityHandler.Preview)
				productProfitabilityGroup.POST("/bulk-upsert", middleware.RequirePermission(auth.PermProcurementEdit), productProfitabilityHandler.BulkUpsert)
			}

			productSpecificationTemplatesGroup := authenticated.Group("/product-specification-templates")
			productSpecificationTemplatesGroup.Use(middleware.RequirePermission(auth.PermProductView))
			{
				productSpecificationTemplatesGroup.GET("", productHandler.ListProductSpecificationTemplates)
				productSpecificationTemplatesGroup.GET("/:id", productHandler.GetProductSpecificationTemplate)
				productSpecificationTemplatesGroup.POST("", middleware.RequirePermission(auth.PermProductCreate), productHandler.CreateProductSpecificationTemplate)
				productSpecificationTemplatesGroup.PUT("/:id", middleware.RequirePermission(auth.PermProductEdit), productHandler.UpdateProductSpecificationTemplate)
				productSpecificationTemplatesGroup.DELETE("/:id", middleware.RequirePermission(auth.PermProductDelete), productHandler.DeleteProductSpecificationTemplate)
			}

			productBrandsGroup := authenticated.Group("/product-brands")
			productBrandsGroup.Use(middleware.RequirePermission(auth.PermProductView))
			{
				productBrandsGroup.GET("", productBrandHandler.List)
				productBrandsGroup.GET("/:id", productBrandHandler.Get)
				productBrandsGroup.POST("", middleware.RequirePermission(auth.PermProductCreate), productBrandHandler.Create)
				productBrandsGroup.PUT("/:id", middleware.RequirePermission(auth.PermProductEdit), productBrandHandler.Update)
				productBrandsGroup.DELETE("/:id", middleware.RequirePermission(auth.PermProductDelete), productBrandHandler.Delete)
			}

			productCategoriesGroup := authenticated.Group("/product-categories")
			productCategoriesGroup.Use(middleware.RequirePermission(auth.PermProductView))
			{
				productCategoriesGroup.GET("", productCategoryHandler.List)
				productCategoriesGroup.GET("/:id", productCategoryHandler.Get)
				productCategoriesGroup.GET("/:id/translations", productCategoryHandler.ListTranslations)
				productCategoriesGroup.POST("", middleware.RequirePermission(auth.PermProductCreate), productCategoryHandler.Create)
				productCategoriesGroup.PUT("/:id", middleware.RequirePermission(auth.PermProductEdit), productCategoryHandler.Update)
				productCategoriesGroup.PUT("/:id/translations", middleware.RequirePermission(auth.PermProductEdit), productCategoryHandler.UpdateTranslations)
				productCategoriesGroup.DELETE("/:id", middleware.RequirePermission(auth.PermProductDelete), productCategoryHandler.Delete)
			}

			productsGroup := authenticated.Group("/products")
			productsGroup.Use(middleware.RequirePermission(auth.PermProductView))
			{
				productsGroup.GET("", productHandler.ListProducts)
				productsGroup.GET("/stats", productHandler.GetProductStats)
				productsGroup.GET("/customs-summary", productHandler.GetCustomsSummary)
				productsGroup.GET("/:id", productHandler.GetProduct)
				productsGroup.GET("/:id/translations", productHandler.GetProductTranslations)
				productsGroup.POST("", middleware.RequirePermission(auth.PermProductCreate), productHandler.CreateProduct)
				productsGroup.PUT("/:id", middleware.RequirePermission(auth.PermProductEdit), productHandler.UpdateProduct)
				productsGroup.POST("/:id/translations/copy", middleware.RequirePermission(auth.PermProductCreate), productHandler.CopyProductTranslation)
				productsGroup.PATCH("/:id/status", middleware.RequirePermission(auth.PermProductEdit), productHandler.UpdateProductStatus)
				productsGroup.DELETE("/:id", middleware.RequirePermission(auth.PermProductDelete), productHandler.DeleteProduct)
				productsGroup.POST("/batch-status", middleware.RequirePermission(auth.PermProductEdit), productHandler.BatchUpdateStatus)
				productsGroup.POST("/batch-delete", middleware.RequirePermission(auth.PermProductDelete), productHandler.BatchDelete)
			}

			productInformationTemplatesGroup := authenticated.Group("/product-information-templates")
			productInformationTemplatesGroup.Use(middleware.RequirePermission(auth.PermProductView))
			{
				productInformationTemplatesGroup.GET("", productInformationTemplateHandler.List)
				productInformationTemplatesGroup.GET("/:id", productInformationTemplateHandler.Get)
				productInformationTemplatesGroup.POST("", middleware.RequirePermission(auth.PermProductCreate), productInformationTemplateHandler.Create)
				productInformationTemplatesGroup.PUT("/:id", middleware.RequirePermission(auth.PermProductEdit), productInformationTemplateHandler.Update)
				productInformationTemplatesGroup.DELETE("/:id", middleware.RequirePermission(auth.PermProductDelete), productInformationTemplateHandler.Delete)
			}

			customsClassificationsGroup := authenticated.Group("/customs-classifications")
			customsClassificationsGroup.Use(middleware.RequirePermission(auth.PermProductView))
			{
				customsClassificationsGroup.GET("", customsClassificationHandler.List)
				customsClassificationsGroup.GET("/lookup", customsClassificationHandler.Lookup)
				customsClassificationsGroup.GET("/:id", customsClassificationHandler.Get)
				customsClassificationsGroup.POST("", middleware.RequirePermission(auth.PermProductCreate), customsClassificationHandler.Create)
				customsClassificationsGroup.PUT("/:id", middleware.RequirePermission(auth.PermProductEdit), customsClassificationHandler.Update)
				customsClassificationsGroup.DELETE("/:id", middleware.RequirePermission(auth.PermProductDelete), customsClassificationHandler.Delete)
			}

			spokeCatalogGroup := authenticated.Group("/spoke-catalog")
			spokeCatalogGroup.Use(middleware.RequirePermission(auth.PermProductView))
			{
				spokeCatalogGroup.GET("", spokeCatalogHandler.Get)
				spokeCatalogGroup.PUT("", middleware.RequirePermission(auth.PermProductEdit), spokeCatalogHandler.Replace)
				spokeCatalogGroup.POST("/import", middleware.RequirePermission(auth.PermProductEdit), spokeCatalogHandler.Import)
				spokeCatalogGroup.GET("/preset-template", spokeCatalogHandler.DownloadPresetTemplate)
				spokeCatalogGroup.POST("/preset-template/import", middleware.RequirePermission(auth.PermProductEdit), spokeCatalogHandler.ImportPresetTemplate)
			}

			quickBuyGroup := authenticated.Group("/quick-buy")
			quickBuyGroup.Use(middleware.RequirePermission(auth.PermProductView))
			{
				quickBuyGroup.GET("/flows", quickBuyHandler.ListFlows)
				quickBuyGroup.GET("/flows/:id", quickBuyHandler.GetFlow)
				quickBuyGroup.POST("/flows", middleware.RequirePermission(auth.PermProductEdit), quickBuyHandler.CreateFlow)
				quickBuyGroup.PUT("/flows/:id", middleware.RequirePermission(auth.PermProductEdit), quickBuyHandler.UpdateFlow)
				quickBuyGroup.PUT("/flows/:id/configuration", middleware.RequirePermission(auth.PermProductEdit), quickBuyHandler.SaveFlowConfiguration)
				quickBuyGroup.POST("/flows/:id/draft", middleware.RequirePermission(auth.PermProductEdit), quickBuyHandler.CreateDraftVersion)
				quickBuyGroup.PUT("/flow-versions/:version_id", middleware.RequirePermission(auth.PermProductEdit), quickBuyHandler.UpdateDraftVersion)
				quickBuyGroup.POST("/flow-versions/:version_id/validate", quickBuyHandler.ValidateVersion)
				quickBuyGroup.POST("/flow-versions/:version_id/preview", quickBuyHandler.PreviewVersionStepCandidates)
				quickBuyGroup.POST("/flow-versions/:version_id/publish", middleware.RequirePermission(auth.PermProductEdit), quickBuyHandler.PublishVersion)
			}

			selectionAssistantGroup := authenticated.Group("/selection-assistant")
			selectionAssistantGroup.Use(middleware.RequirePermission(auth.PermProductView))
			{
				selectionAssistantGroup.GET("/flows", selectionAssistantHandler.ListFlows)
				selectionAssistantGroup.GET("/flows/:id", selectionAssistantHandler.GetFlow)
				selectionAssistantGroup.POST("/flows", middleware.RequirePermission(auth.PermProductEdit), selectionAssistantHandler.CreateFlow)
				selectionAssistantGroup.PUT("/flows/:id/configuration", middleware.RequirePermission(auth.PermProductEdit), selectionAssistantHandler.SaveFlowConfiguration)
				selectionAssistantGroup.POST("/flow-versions/:version_id/validate", selectionAssistantHandler.ValidateVersion)
				selectionAssistantGroup.POST("/flow-versions/:version_id/publish", middleware.RequirePermission(auth.PermProductEdit), selectionAssistantHandler.PublishVersion)
			}

			selectionConfigurationKeyGroup := authenticated.Group("/selection-configuration")
			selectionConfigurationKeyGroup.Use(middleware.RequirePermission(auth.PermProductView))
			{
				selectionConfigurationKeyGroup.GET("/keys", selectionConfigurationKeyHandler.List)
				selectionConfigurationKeyGroup.GET("/keys/options", selectionConfigurationKeyHandler.ListOptions)
				selectionConfigurationKeyGroup.POST("/keys", middleware.RequirePermission(auth.PermProductEdit), selectionConfigurationKeyHandler.Create)
				selectionConfigurationKeyGroup.PUT("/keys/:id", middleware.RequirePermission(auth.PermProductEdit), selectionConfigurationKeyHandler.Update)
			}

			wheelsetFitQuestionnaireGroup := authenticated.Group("/wheelset-fit-questionnaire")
			wheelsetFitQuestionnaireGroup.Use(middleware.RequirePermission(auth.PermProductView))
			{
				wheelsetFitQuestionnaireGroup.GET("/current", wheelsetFitQuestionnaireHandler.GetCurrentVersion)
				wheelsetFitQuestionnaireGroup.GET("/product-filter-options", wheelsetFitQuestionnaireHandler.GetProductFilterOptions)
				wheelsetFitQuestionnaireGroup.POST("/draft", middleware.RequirePermission(auth.PermProductEdit), wheelsetFitQuestionnaireHandler.CreateDraft)
				wheelsetFitQuestionnaireGroup.POST("/questions", middleware.RequirePermission(auth.PermProductEdit), wheelsetFitQuestionnaireHandler.CreateQuestion)
				wheelsetFitQuestionnaireGroup.PUT("/questions/order", middleware.RequirePermission(auth.PermProductEdit), wheelsetFitQuestionnaireHandler.ReorderQuestions)
				wheelsetFitQuestionnaireGroup.PUT("/questions/:id", middleware.RequirePermission(auth.PermProductEdit), wheelsetFitQuestionnaireHandler.UpdateQuestion)
				wheelsetFitQuestionnaireGroup.DELETE("/questions/:id", middleware.RequirePermission(auth.PermProductEdit), wheelsetFitQuestionnaireHandler.DeleteQuestion)
				wheelsetFitQuestionnaireGroup.POST("/versions/:version_id/validate", wheelsetFitQuestionnaireHandler.ValidateVersion)
				wheelsetFitQuestionnaireGroup.POST("/versions/:version_id/publish", middleware.RequirePermission(auth.PermProductEdit), wheelsetFitQuestionnaireHandler.PublishVersion)
			}

			googleMerchantGroup := authenticated.Group("/google-merchant")
			googleMerchantGroup.Use(middleware.RequirePermission(auth.PermMerchantView))
			{
				googleMerchantGroup.GET("/connection", googleMerchantHandler.GetConnection)
				googleMerchantGroup.PATCH("/connection", middleware.RequirePermission(auth.PermMerchantEdit), googleMerchantHandler.UpdateConnection)
				googleMerchantGroup.POST("/oauth/start", middleware.RequirePermission(auth.PermMerchantEdit), googleMerchantHandler.StartOAuth)
				googleMerchantGroup.POST("/disconnect", middleware.RequirePermission(auth.PermMerchantEdit), googleMerchantHandler.Disconnect)
				googleMerchantGroup.GET("/remote-products", googleMerchantHandler.ListRemoteProducts)
				googleMerchantGroup.GET("/offers", googleMerchantHandler.ListOffers)
				googleMerchantGroup.POST("/reconcile", middleware.RequirePermission(auth.PermMerchantSync), googleMerchantHandler.Reconcile)
				googleMerchantGroup.POST("/offers", middleware.RequirePermission(auth.PermMerchantEdit), googleMerchantHandler.CreateOffer)
				googleMerchantGroup.PUT("/offers/:id", middleware.RequirePermission(auth.PermMerchantEdit), googleMerchantHandler.UpdateOffer)
				googleMerchantGroup.POST("/offers/:id/validate", middleware.RequirePermission(auth.PermMerchantEdit), googleMerchantHandler.ValidateOffer)
				googleMerchantGroup.POST("/offers/:id/sync", middleware.RequirePermission(auth.PermMerchantSync), googleMerchantHandler.SyncOffer)
				googleMerchantGroup.POST("/offers/:id/remove-remote", middleware.RequirePermission(auth.PermMerchantSync), googleMerchantHandler.RemoveRemoteOffer)
				googleMerchantGroup.DELETE("/offers/:id", middleware.RequirePermission(auth.PermMerchantEdit), googleMerchantHandler.DeleteOffer)
			}

			socialOAuthGroup := authenticated.Group("/social/oauth")
			socialOAuthGroup.Use(middleware.RequirePermission(auth.PermSettingsView))
			{
				socialOAuthGroup.GET("", socialOAuthHandler.ListConnections)
				socialOAuthGroup.POST(
					"/:provider/start",
					middleware.RequirePermission(auth.PermSettingsEdit),
					socialOAuthHandler.Start,
				)
				socialOAuthGroup.DELETE(
					"/:provider",
					middleware.RequirePermission(auth.PermSettingsEdit),
					socialOAuthHandler.Disconnect,
				)
			}

			mediaGroup := authenticated.Group("/media")
			mediaGroup.Use(middleware.RequirePermission(auth.PermMediaView))
			{
				mediaGroup.GET("/derivative-presets", mediaHandler.ListDerivativePresets)
				mediaGroup.POST("/derivative-presets", middleware.RequirePermission(auth.PermMediaConfigure), mediaHandler.CreateDerivativePreset)
				mediaGroup.PUT("/derivative-presets/:id", middleware.RequirePermission(auth.PermMediaConfigure), mediaHandler.UpdateDerivativePreset)
				mediaGroup.PATCH("/derivative-presets/:id/enabled", middleware.RequirePermission(auth.PermMediaConfigure), mediaHandler.UpdateDerivativePresetEnabled)
				mediaGroup.DELETE("/derivative-presets/:id", middleware.RequirePermission(auth.PermMediaConfigure), mediaHandler.DeleteDerivativePreset)
				mediaGroup.GET("/derivative-rebuild-jobs", mediaHandler.ListDerivativeRebuildJobs)
				mediaGroup.POST("/derivative-rebuild-jobs", middleware.RequirePermission(auth.PermMediaConfigure), mediaHandler.RequestDerivativeRebuild)

				mediaGroup.GET("/assets", mediaHandler.ListAssets)
				mediaGroup.GET("/assets/:id/file", mediaHandler.ServeAssetFile)
				mediaGroup.GET("/assets/:id/copyright-evidence", middleware.RequirePermission(auth.PermMediaEdit), mediaHandler.ExportCopyrightEvidence)
				mediaGroup.GET("/assets/:id/references", mediaHandler.GetAssetReferences)
				mediaGroup.GET("/assets/:id", mediaHandler.GetAsset)
				mediaGroup.POST(
					"/assets",
					middleware.RequirePermission(auth.PermMediaCreate),
					middleware.RateLimitByUserPerMinute(3, 2),
					mediaHandler.UploadAsset,
				)
				mediaGroup.PATCH("/assets/:id", middleware.RequirePermission(auth.PermMediaEdit), mediaHandler.UpdateAsset)
				mediaGroup.DELETE("/assets/:id", middleware.RequirePermission(auth.PermMediaDelete), mediaHandler.DeleteAsset)
			}

			preflightGroup := authenticated.Group("/preflight")
			preflightGroup.Use(middleware.RequireAnyPermission(auth.PermServicesView, auth.PermMediaView))
			{
				siteQualityGroup := preflightGroup.Group("/site-quality")
				siteQualityGroup.Use(middleware.RequirePermission(auth.PermServicesView))
				registerSiteQualityRoutes(siteQualityGroup, siteQualityHandler, auth.PermServicesManage)

				imageDimensionsGroup := preflightGroup.Group("/image-dimensions")
				imageDimensionsGroup.Use(middleware.RequirePermission(auth.PermMediaView))
				registerMediaImageDimensionRoutes(imageDimensionsGroup, mediaImageDimensionsHandler)

				contentLinksGroup := preflightGroup.Group("/content-links")
				contentLinksGroup.Use(middleware.RequirePermission(auth.PermServicesView))
				registerPreflightContentLinkRoutes(contentLinksGroup, contentLinkPreflightHandler, auth.PermServicesManage)

				preflightGroup.GET("/fonts", middleware.RequirePermission(auth.PermServicesView), fontPreflightHandler.Get)
			}

			// 属性管理（需要商品管理权限）
			attributesGroup := authenticated.Group("/attributes")
			attributesGroup.Use(middleware.RequirePermission(auth.PermProductView))
			{
				attributesGroup.GET("", productHandler.ListAttributes)
				attributesGroup.GET("/:id", productHandler.GetAttribute)
				attributesGroup.POST("", middleware.RequirePermission(auth.PermProductCreate), productHandler.CreateAttribute)
				attributesGroup.PUT("/:id", middleware.RequirePermission(auth.PermProductEdit), productHandler.UpdateAttribute)
				attributesGroup.DELETE("/:id", middleware.RequirePermission(auth.PermProductDelete), productHandler.DeleteAttribute)

				// 属性值管理
				attributesGroup.GET("/:id/values", productHandler.GetAttributeValues)
				attributesGroup.POST("/:id/values", middleware.RequirePermission(auth.PermProductCreate), productHandler.CreateAttributeValue)
				attributesGroup.PUT("/:id/values/:valueId", middleware.RequirePermission(auth.PermProductEdit), productHandler.UpdateAttributeValue)
				attributesGroup.DELETE("/:id/values/:valueId", middleware.RequirePermission(auth.PermProductDelete), productHandler.DeleteAttributeValue)
			}

			// 订单管理（需要订单管理权限）
			ordersGroup := authenticated.Group("/orders")
			ordersGroup.Use(middleware.RequirePermission(auth.PermOrderView))
			{
				ordersGroup.GET("", orderHandler.ListOrders)
				ordersGroup.GET("/disputes", orderHandler.ListDisputeOrders)
				ordersGroup.GET("/stats", orderHandler.GetOrderStats)
				ordersGroup.GET("/sales-chart", orderHandler.GetSalesChart)
				ordersGroup.GET("/export", orderHandler.ExportOrders)
				ordersGroup.GET("/:id/after-sales", afterSalesHandler.ListByOrder)
				ordersGroup.POST("/:id/after-sales", middleware.RequirePermission(auth.PermOrderEdit), afterSalesHandler.Create)
				ordersGroup.PATCH("/after-sales/:id/status", middleware.RequirePermission(auth.PermOrderEdit), afterSalesHandler.UpdateStatus)
				ordersGroup.GET("/:id/dispute-analysis", orderHandler.GetOrderDisputeAnalysis)
				ordersGroup.GET("/:id/customs-export", orderHandler.ExportOrderCustoms)
				ordersGroup.GET("/:id", orderHandler.GetOrder)
				ordersGroup.PATCH("/:id/status", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.UpdateOrderStatus)
				ordersGroup.PATCH("/:id/shipping-status", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.UpdateShippingStatus)
				ordersGroup.PATCH("/:id/tracking", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.UpdateTrackingInfo)
				ordersGroup.POST("/:id/fulfillment", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.FulfillOrder)
				ordersGroup.POST("/:id/tracking/sync", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.SyncTrackingInfo)
				ordersGroup.POST("/:id/dispute-contact-email", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.SendDisputeContactEmail)
				ordersGroup.PATCH("/:id/admin-note", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.UpdateAdminNote)
				ordersGroup.PATCH("/:id/items/:item_id/customs", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.UpdateOrderItemCustoms)
				ordersGroup.POST("/batch-status", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.BatchUpdateStatus)
				ordersGroup.DELETE("/:id", middleware.RequirePermission(auth.PermOrderDelete), orderHandler.DeleteOrder)
			}

			afterSalesGroup := authenticated.Group("/after-sales")
			afterSalesGroup.Use(middleware.RequirePermission(auth.PermOrderView))
			{
				afterSalesGroup.GET("", afterSalesHandler.List)
				afterSalesGroup.GET("/:id", afterSalesHandler.Get)
				afterSalesGroup.GET("/:id/attachments/:attachment_id", afterSalesHandler.ServeAttachment)
				afterSalesGroup.PATCH("/:id/status", middleware.RequirePermission(auth.PermOrderEdit), afterSalesHandler.UpdateStatus)
				afterSalesGroup.GET("/:id/refund-review", afterSalesHandler.GetRefundReview)
				afterSalesGroup.PUT("/:id/refund-review", middleware.RequirePermission(auth.PermOrderEdit), afterSalesHandler.SaveRefundReview)
				afterSalesGroup.PATCH("/:id/refund-review/decision", middleware.RequirePermission(auth.PermOrderRefund), afterSalesHandler.DecideRefundReview)
				afterSalesGroup.POST("/:id/refund-review/pending-refund", middleware.RequirePermission(auth.PermOrderRefund), afterSalesHandler.CreatePendingRefund)
			}

			paymentGroup := authenticated.Group("/payment")
			paymentGroup.Use(middleware.RequirePermission(auth.PermOrderView))
			{
				paymentGroup.GET("/transactions/:id", paymentHandler.GetTransaction)
				paymentGroup.GET("/orders/:order_id/transactions", paymentHandler.GetOrderTransactions)
				paymentGroup.GET("/refunds/:id", paymentHandler.GetRefund)
				paymentGroup.GET("/orders/:order_id/refunds", paymentHandler.GetOrderRefunds)
				paymentGroup.POST("/refunds", middleware.RequirePermission(auth.PermOrderRefund), paymentHandler.CreateRefund)
				paymentGroup.POST("/refunds/:id/execute", middleware.RequirePermission(auth.PermOrderRefund), paymentRefundExecutionHandler.ExecutePendingRefund)
				paymentGroup.GET("/disputes", paymentHandler.ListStripeDisputes)
				paymentGroup.GET("/disputes/:id", paymentHandler.GetStripeDispute)
				paymentGroup.GET("/disputes/:id/evidence", paymentHandler.GetStripeDisputeEvidence)
				paymentGroup.POST("/disputes/:id/evidence/submit", middleware.RequirePermission(auth.PermOrderEdit), paymentHandler.SubmitStripeDisputeEvidence)
				paymentGroup.GET("/paypal-disputes", paymentHandler.ListPayPalDisputes)
				paymentGroup.GET("/paypal-disputes/:id", paymentHandler.GetPayPalDispute)
				paymentGroup.GET("/paypal-disputes/:id/evidence", paymentHandler.GetPayPalDisputeEvidence)
				paymentGroup.POST("/paypal-disputes/:id/evidence/submit", middleware.RequirePermission(auth.PermOrderEdit), paymentHandler.SubmitPayPalDisputeEvidence)
				paymentGroup.GET("/paypal-disputes/:id/evidence/invoice.pdf", paymentHandler.PreviewPayPalDisputeCommercialInvoicePDF)
				paymentGroup.POST("/paypal-invoice-preview.pdf", paymentHandler.PreviewPayPalCommercialInvoicePDF)
				paymentGroup.GET("/reviews", paymentHandler.ListPaymentReviews)
				paymentGroup.GET("/reviews/:id", paymentHandler.GetPaymentReview)
				paymentGroup.POST("/reviews", middleware.RequirePermission(auth.PermOrderEdit), paymentHandler.CreatePaymentReview)
				paymentGroup.PATCH("/reviews/:id", middleware.RequirePermission(auth.PermOrderEdit), paymentHandler.UpdatePaymentReview)
				paymentGroup.GET("/risk/summary", paymentRiskMonitoringHandler.GetSummary)
				paymentGroup.POST("/risk/recompute", middleware.RequirePermission(auth.PermOrderEdit), paymentRiskMonitoringHandler.RecomputeSummary)
				paymentGroup.GET("/risk/refund-recommendations", paymentRefundRecommendationHandler.ListRecommendations)
				paymentGroup.PATCH("/risk/refund-recommendations/:id", middleware.RequirePermission(auth.PermOrderEdit), paymentRefundRecommendationHandler.UpdateRecommendation)
				paymentGroup.POST("/risk/refund-recommendations/:id/pending-refund", middleware.RequirePermission(auth.PermOrderRefund), paymentRefundRecommendationHandler.CreatePendingRefund)
				paymentGroup.GET("/risk/controls", paymentProtectionHandler.ListControls)
				paymentGroup.GET("/risk/controls/:id/audit", paymentProtectionHandler.ListControlAuditLogs)
				paymentGroup.POST("/risk/controls", middleware.RequirePermission(auth.PermOrderEdit), paymentProtectionHandler.CreateControl)
				paymentGroup.POST("/risk/controls/:id/revoke", middleware.RequirePermission(auth.PermOrderEdit), paymentProtectionHandler.RevokeControl)
			}

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

				visualShowcaseGroup := contentGroup.Group("/visual-showcases")
				{
					visualShowcaseGroup.GET("/:showcase_key", visualShowcaseHandler.GetItems)
					visualShowcaseGroup.POST("/:showcase_key/assets", middleware.RequirePermission(auth.PermContentEdit), middleware.RateLimitByUserPerMinute(3, 2), visualShowcaseHandler.UploadImage)
					visualShowcaseGroup.PUT("/:showcase_key", middleware.RequirePermission(auth.PermContentEdit), visualShowcaseHandler.ReplaceItems)
				}
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
				showcaseGroup.GET("", showcaseHandler.List)
				showcaseGroup.GET("/:id/images/:image_index/file", showcaseHandler.ServeImageFile)
				showcaseGroup.PUT("/:id/approve", middleware.RequirePermission(auth.PermGalleryEdit), showcaseHandler.Approve)
				showcaseGroup.PUT("/:id/reject", middleware.RequirePermission(auth.PermGalleryEdit), showcaseHandler.Reject)
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

			// 订阅管理（需要订阅管理权限）
			subscriptionsGroup := authenticated.Group("/subscriptions")
			subscriptionsGroup.Use(middleware.RequirePermission(auth.PermSubscriptionView))
			{
				subscriptionsGroup.GET("", subscriptionHandler.ListSubscriptions)
				subscriptionsGroup.GET("/stats", subscriptionHandler.GetSubscriptionStats)
				subscriptionsGroup.GET("/active-emails", subscriptionHandler.GetActiveEmails)
				subscriptionsGroup.GET("/:email", subscriptionHandler.GetSubscription)
				subscriptionsGroup.PATCH("/:email/status", middleware.RequirePermission(auth.PermSubscriptionEdit), subscriptionHandler.UpdateSubscriptionStatus)
				subscriptionsGroup.DELETE("/:email", middleware.RequirePermission(auth.PermSubscriptionDelete), subscriptionHandler.DeleteSubscription)
				subscriptionsGroup.POST("/batch-delete", middleware.RequirePermission(auth.PermSubscriptionDelete), subscriptionHandler.BatchDelete)
			}

			// 在线客服对话（独立于普通工单列表；底层仍使用 customer_service 工单作为唯一事实源）
			customerServiceGroup := authenticated.Group("/customer-service")
			customerServiceGroup.Use(middleware.RequirePermission(auth.PermTicketView))
			{
				customerServiceGroup.GET("/agents", ticketHandler.ListCustomerServiceAgents)
				customerServiceGroup.GET("/groups", ticketHandler.ListCustomerServiceGroups)
				customerServiceGroup.GET("/auto-reply/faqs", autoReplyHandler.ListPublishedFAQs)
				customerServiceGroup.GET("/auto-reply/rules", autoReplyHandler.ListRules)
				customerServiceGroup.GET("/auto-reply/rules/:id", autoReplyHandler.GetRule)
				customerServiceGroup.POST("/auto-reply/rules", middleware.RequirePermission(auth.PermTicketEdit), autoReplyHandler.CreateRule)
				customerServiceGroup.PUT("/auto-reply/rules/:id", middleware.RequirePermission(auth.PermTicketEdit), autoReplyHandler.UpdateRule)
				customerServiceGroup.DELETE("/auto-reply/rules/:id", middleware.RequirePermission(auth.PermTicketDelete), autoReplyHandler.DeleteRule)
				customerServiceGroup.GET("/analytics", ticketHandler.GetCustomerServiceAnalytics)
				customerServiceGroup.GET("/conversations", ticketHandler.ListCustomerServiceConversations)
				customerServiceGroup.GET("/ws", ticketHandler.StreamCustomerServiceWebSocket)
				customerServiceGroup.GET("/visitor-profiles", visitorProfileHandler.ListVisitorProfiles)
				customerServiceGroup.GET("/visitor-profiles/stats", visitorProfileHandler.GetVisitorProfileStats)
				customerServiceGroup.POST("/visitor-profiles/cleanup", middleware.AdminOnly(), visitorProfileHandler.CleanupExpiredVisitorProfiles)
				customerServiceGroup.GET("/visitor-risk-facts", visitorRiskHandler.ListVisitorRiskFacts)
				customerServiceGroup.GET("/visitor-risk-facts/stats", visitorRiskHandler.GetVisitorRiskStats)
				customerServiceGroup.POST("/visitor-risk-facts/cleanup", middleware.AdminOnly(), visitorRiskHandler.CleanupExpiredVisitorRiskFacts)
				customerServiceGroup.GET("/visitor-risk-facts/:id/decision", visitorRiskHandler.GetVisitorRiskDecision)
				customerServiceGroup.POST("/visitor-risk-facts/:id/decision", middleware.AdminOnly(), visitorRiskHandler.CreateVisitorRiskDecision)
				customerServiceGroup.GET("/conversations/:id/context", ticketHandler.GetCustomerServiceConversationContext)
				customerServiceGroup.GET("/conversations/:id/messages", ticketHandler.GetCustomerServiceConversationMessages)
				customerServiceGroup.POST("/conversations/:id/messages", middleware.RequirePermission(auth.PermTicketEdit), ticketHandler.CreateCustomerServiceConversationMessage)
				customerServiceGroup.POST("/conversations/:id/messages/mark-read", ticketHandler.MarkCustomerServiceConversationMessagesRead)
				customerServiceGroup.PATCH("/conversations/:id/transfer", middleware.RequirePermission(auth.PermTicketEdit), ticketHandler.TransferCustomerServiceConversation)
			}

			// 营销管理（需要营销管理权限）
			marketingGroup := authenticated.Group("/marketing")
			marketingGroup.Use(middleware.RequirePermission(auth.PermMarketingView))
			{
				// 营销统计
				marketingGroup.GET("/stats", marketingHandler.GetMarketingStats)

				// 优惠券管理
				couponsGroup := marketingGroup.Group("/coupons")
				{
					couponsGroup.GET("", marketingHandler.ListCoupons)
					couponsGroup.GET("/stats", marketingHandler.GetCouponStats)
					couponsGroup.GET("/:id", marketingHandler.GetCoupon)
					couponsGroup.POST("", middleware.RequirePermission(auth.PermMarketingCreate), marketingHandler.CreateCoupon)
					couponsGroup.PUT("/:id", middleware.RequirePermission(auth.PermMarketingEdit), marketingHandler.UpdateCoupon)
					couponsGroup.DELETE("/:id", middleware.RequirePermission(auth.PermMarketingDelete), marketingHandler.DeleteCoupon)
				}

				// 礼品卡管理
				giftCardsGroup := marketingGroup.Group("/gift-cards")
				{
					giftCardsGroup.GET("", marketingHandler.ListGiftCards)
					giftCardsGroup.GET("/:id", marketingHandler.GetGiftCard)
					giftCardsGroup.PATCH("/:id/status", middleware.RequirePermission(auth.PermMarketingEdit), marketingHandler.UpdateGiftCardStatus)
				}

				// 积分交易管理
				loyaltyGroup := marketingGroup.Group("/loyalty")
				{
					loyaltyGroup.GET("/transactions", marketingHandler.ListLoyaltyTransactions)
					loyaltyGroup.GET("/redemptions", marketingHandler.ListGiftCardRedemptions)
					loyaltyGroup.POST("/transactions", middleware.RequirePermission(auth.PermMarketingCreate), marketingHandler.CreateLoyaltyTransaction)
					loyaltyGroup.GET("/check-ins", marketingHandler.ListCheckIns)
					loyaltyGroup.GET("/referrals", marketingHandler.ListReferrals)
					loyaltyGroup.PATCH("/referrals/:id/status", middleware.RequirePermission(auth.PermMarketingEdit), marketingHandler.UpdateReferralStatus)
					loyaltyGroup.GET("/program-config", marketingHandler.GetLoyaltyProgramConfig)
					loyaltyGroup.PUT("/program-config", middleware.RequirePermission(auth.PermMarketingEdit), marketingHandler.UpdateLoyaltyProgramConfig)
				}

				// 会员等级管理
				levelsGroup := marketingGroup.Group("/levels")
				{
					levelsGroup.GET("", marketingHandler.ListMemberLevels)
					levelsGroup.GET("/:id", marketingHandler.GetMemberLevel)
					levelsGroup.POST("", middleware.RequirePermission(auth.PermMarketingCreate), marketingHandler.CreateMemberLevel)
					levelsGroup.PUT("/:id", middleware.RequirePermission(auth.PermMarketingEdit), marketingHandler.UpdateMemberLevel)
					levelsGroup.DELETE("/:id", middleware.RequirePermission(auth.PermMarketingDelete), marketingHandler.DeleteMemberLevel)
				}
			}

			// 后台定价辅助（商品、运费和设置编辑均可使用）
			pricingGroup := authenticated.Group("/pricing")
			pricingGroup.Use(middleware.RequireAnyPermission(auth.PermSettingsEdit, auth.PermProductEdit, auth.PermShippingEdit))
			{
				pricingGroup.POST("/exchange-rates/convert", exchangeRateHandler.ConvertDisplayPrices)
			}

			seoGroup := authenticated.Group("/seo")
			seoGroup.Use(middleware.RequirePermission(auth.PermSEOView))
			{
				seoGroup.GET("/indexing/status", seoProductsHandler.IndexingStatus)
				seoGroup.GET("/home", seoHomeHandler.Get)
				seoGroup.PUT("/home", middleware.RequirePermission(auth.PermSEOEdit), seoHomeHandler.Update)
				seoGroup.GET("/articles", seoArticlesHandler.Get)
				seoGroup.PUT("/articles/:id", middleware.RequirePermission(auth.PermSEOEdit), seoArticlesHandler.Update)
				seoGroup.GET("/products", seoProductsHandler.Get)
				seoGroup.PUT("/products/:id", middleware.RequirePermission(auth.PermSEOEdit), seoProductsHandler.Update)
				seoGroup.GET("/categories", seoCategoriesHandler.Get)
				seoGroup.PUT("/categories/:id", middleware.RequirePermission(auth.PermSEOEdit), seoCategoriesHandler.Update)
				seoGroup.POST(
					"/products/:id/indexing",
					middleware.RequirePermission(auth.PermSEOEdit),
					middleware.Idempotency(deps.RedisClient),
					middleware.RateLimitByUserPerMinuteRedis(deps.RedisClient),
					seoProductsHandler.PushIndexing,
				)
			}

			urlGroup := authenticated.Group("/urls")
			urlGroup.Use(middleware.RequirePermission(auth.PermURLView))
			{
				urlGroup.GET("/stats", urlRoutesHandler.Stats)
				urlGroup.GET("/routes", urlRoutesHandler.List)
				urlGroup.GET("/routes/:id", urlRoutesHandler.Get)
				urlGroup.GET("/routes/:id/history", urlRoutesHandler.History)
				urlGroup.GET("/issues", urlIssuesHandler.List)
				urlGroup.GET("/issues/summary", urlIssuesHandler.Summary)
				urlGroup.GET("/issues/:id", urlIssuesHandler.Get)
				urlGroup.GET("/issues/:id/events", urlIssuesHandler.Events)
				urlGroup.GET("/redirects", urlRedirectsHandler.List)
				urlGroup.GET("/sitemap", urlRoutesHandler.Sitemap)
				urlGroup.POST("/issues/:id/acknowledge", middleware.RequirePermission(auth.PermURLEdit), urlIssuesHandler.Acknowledge)
				urlGroup.POST("/issues/:id/claim", middleware.RequirePermission(auth.PermURLEdit), urlIssuesHandler.Claim)
				urlGroup.POST("/issues/:id/comments", middleware.RequirePermission(auth.PermURLEdit), urlIssuesHandler.Comment)
				urlGroup.POST("/issues/:id/link-redirect", middleware.RequirePermission(auth.PermURLEdit), urlIssuesHandler.LinkRedirect)
				urlGroup.POST("/issues/:id/resolve", middleware.RequirePermission(auth.PermURLEdit), urlIssuesHandler.Resolve)
				urlGroup.POST("/issues/:id/suppress", middleware.RequirePermission(auth.PermURLEdit), urlIssuesHandler.Suppress)
				urlGroup.POST("/issues/:id/recheck", middleware.RequirePermission(auth.PermURLEdit), urlIssuesHandler.Recheck)
				urlGroup.POST("/issues/:id/verify", middleware.RequirePermission(auth.PermURLEdit), urlIssuesHandler.Verify)
				urlGroup.POST("/redirects", middleware.RequirePermission(auth.PermURLEdit), urlRedirectsHandler.Create)
				urlGroup.POST("/redirects/:id/publish", middleware.RequirePermission(auth.PermURLEdit), urlRedirectsHandler.Publish)
				urlGroup.POST("/redirects/:id/disable", middleware.RequirePermission(auth.PermURLEdit), urlRedirectsHandler.Disable)
				urlGroup.POST("/sync", middleware.RequirePermission(auth.PermURLEdit), urlRoutesHandler.Sync)
				urlGroup.POST("/sitemap/sync", middleware.RequirePermission(auth.PermURLEdit), urlRoutesHandler.SyncSitemap)
				urlGroup.POST("/check", middleware.RequirePermission(auth.PermURLEdit), urlRoutesHandler.Check)
				urlGroup.POST("/routes/:id/check", middleware.RequirePermission(auth.PermURLEdit), urlRoutesHandler.CheckOne)
			}

			analyticsGroup := authenticated.Group("/analytics")
			analyticsGroup.Use(middleware.RequirePermission(auth.PermAnalyticsView))
			{
				analyticsGroup.GET("", analyticsHandler.Get)
				analyticsGroup.PUT("", middleware.RequirePermission(auth.PermAnalyticsEdit), analyticsHandler.Update)
			}

			servicesGroup := authenticated.Group("/services")
			servicesGroup.Use(middleware.RequirePermission(auth.PermServicesView))
			{
				servicesGroup.GET("/overview", serviceCenterHandler.Overview)
				cloudflareGroup := servicesGroup.Group("/cloudflare")
				{
					cloudflareGroup.GET("", serviceCenterHandler.Cloudflare)
					cloudflareGroup.GET("/cache-rules", serviceCenterHandler.GetCloudflareCacheRules)
					cloudflareGroup.POST("/oauth/start", middleware.RequirePermission(auth.PermServicesManage), serviceCenterHandler.StartCloudflareOAuth)
					cloudflareGroup.POST("/connectors/:id/test", middleware.RequirePermission(auth.PermServicesManage), serviceCenterHandler.TestCloudflareConnection)
					cloudflareGroup.PATCH("/cache-rules/:rule_id/enabled", middleware.RequirePermission(auth.PermServicesManage), serviceCenterHandler.UpdateCloudflareCacheRuleEnabled)
				}
			}

			// 设置管理（需要设置管理权限）
			settingsGroup := authenticated.Group("/settings")
			settingsGroup.Use(middleware.RequirePermission(auth.PermSettingsView))
			{
				settingsGroup.GET("", settingsHandler.GetAllSettings)
				settingsGroup.GET("/groups", settingsHandler.GetGroups)
				settingsGroup.GET("/commercial-crawler-protection", commercialCrawlerHandler.GetStatus)
				settingsGroup.GET("/public-chat-agents", publicChatAgentHandler.ListPublicChatAgents)
				settingsGroup.GET("/public-chat-agent-candidates", publicChatAgentHandler.ListPublicChatAgentCandidates)
				settingsGroup.POST("/public-chat-agents", middleware.RequirePermission(auth.PermSettingsEdit), publicChatAgentHandler.UpsertPublicChatAgent)
				settingsGroup.POST(
					"/public-chat-agents/:userID/avatar",
					middleware.RequirePermission(auth.PermSettingsEdit),
					middleware.RateLimitByUserPerMinute(3, 2),
					customerServiceAvatarHandler.UploadAvatar,
				)
				settingsGroup.DELETE(
					"/public-chat-agents/:userID/avatar",
					middleware.RequirePermission(auth.PermSettingsEdit),
					middleware.RateLimitByUserPerMinute(3, 2),
					customerServiceAvatarHandler.DeleteAvatar,
				)
				settingsGroup.GET("/public-chat-groups", publicChatAgentHandler.ListPublicChatGroups)
				settingsGroup.POST("/public-chat-groups", middleware.RequirePermission(auth.PermSettingsEdit), publicChatAgentHandler.UpsertPublicChatGroup)
				settingsGroup.PUT("/public-chat-groups/:id", middleware.RequirePermission(auth.PermSettingsEdit), publicChatAgentHandler.UpdatePublicChatGroup)
				settingsGroup.DELETE("/public-chat-groups/:id", middleware.RequirePermission(auth.PermSettingsEdit), publicChatAgentHandler.DeletePublicChatGroup)
				settingsGroup.GET("/payment-runtime", paymentHandler.GetGatewayRuntimeStatus)
				settingsGroup.GET("/refund-return-policy", refundReturnPolicyHandler.Get)
				settingsGroup.GET("/paypal-invoice-seller-profile", paymentHandler.GetPayPalDisputeInvoiceSellerProfile)
				settingsGroup.PUT("/paypal-invoice-seller-profile", middleware.RequirePermission(auth.PermSettingsEdit), paymentHandler.UpdatePayPalDisputeInvoiceSellerProfile)
				settingsGroup.POST("/payment-runtime/:provider/callback-check", middleware.RequirePermission(auth.PermSettingsEdit), paymentHandler.CheckGatewayCallback)
				settingsGroup.GET("/currency-policy", currencyPolicyHandler.GetPolicy)
				settingsGroup.GET("/currency-policy/audit", currencyPolicyHandler.GetBackendEntryCurrencyAudit)
				settingsGroup.PUT("/currency-policy", middleware.RequirePermission(auth.PermSettingsEdit), currencyPolicyHandler.UpdatePolicy)
				settingsGroup.POST(
					"/site-logo",
					middleware.RequirePermission(auth.PermSettingsEdit),
					middleware.RateLimitByUserPerMinute(3, 2),
					siteLogoHandler.Upload,
				)
				settingsGroup.DELETE(
					"/site-logo",
					middleware.RequirePermission(auth.PermSettingsEdit),
					middleware.RateLimitByUserPerMinute(3, 2),
					siteLogoHandler.Delete,
				)
				settingsGroup.GET("/exchange-rates", exchangeRateHandler.GetExchangeRates)
				settingsGroup.POST("/exchange-rates/sync", middleware.RequirePermission(auth.PermSettingsEdit), exchangeRateHandler.SyncExchangeRates)
				settingsGroup.POST("/exchange-rates/convert", middleware.RequirePermission(auth.PermSettingsEdit), exchangeRateHandler.ConvertDisplayPrices)
				settingsGroup.PUT("/payment-gateways/:provider", middleware.RequirePermission(auth.PermSettingsEdit), paymentHandler.UpsertGatewayConfig)
				settingsGroup.PUT("/refund-return-policy", middleware.RequirePermission(auth.PermSettingsEdit), refundReturnPolicyHandler.Update)
				settingsGroup.DELETE("/payment-gateways/:provider", middleware.RequirePermission(auth.PermSettingsEdit), paymentHandler.DeleteGatewayConfig)
				settingsGroup.GET("/payment-installments/:provider", paymentHandler.GetPaymentProviderInstallments)
				settingsGroup.PUT("/payment-installments/:provider", middleware.RequirePermission(auth.PermSettingsEdit), paymentHandler.UpdatePaymentProviderInstallments)
				settingsGroup.DELETE("/payment-installments/:provider", middleware.RequirePermission(auth.PermSettingsEdit), paymentHandler.DeletePaymentProviderInstallments)
				settingsGroup.GET("/payment-methods", paymentHandler.ListPaymentMethods)
				settingsGroup.GET("/payment-methods/:id", paymentHandler.GetPaymentMethod)
				settingsGroup.POST("/payment-methods", middleware.RequirePermission(auth.PermSettingsEdit), paymentHandler.CreatePaymentMethod)
				settingsGroup.PUT("/payment-methods/:id", middleware.RequirePermission(auth.PermSettingsEdit), paymentHandler.UpdatePaymentMethod)
				settingsGroup.DELETE("/payment-methods/:id", middleware.RequirePermission(auth.PermSettingsEdit), paymentHandler.DeletePaymentMethod)
				settingsGroup.PUT("", middleware.RequirePermission(auth.PermSettingsEdit), settingsHandler.UpdateSetting)
				settingsGroup.POST("/batch", middleware.RequirePermission(auth.PermSettingsEdit), settingsHandler.BatchUpdateSettings)
				settingsGroup.DELETE("/:key", middleware.RequirePermission(auth.PermSettingsEdit), settingsHandler.DeleteSetting)

				// 分组设置
				settingsGroup.GET("/site", settingsHandler.GetSiteSettings)
				settingsGroup.GET("/email", settingsHandler.GetEmailSettings)
				settingsGroup.GET("/social", settingsHandler.GetSocialSettings)
				settingsGroup.GET("/website-profile", websiteProfileHandler.Get)
				settingsGroup.PUT("/website-profile", middleware.RequirePermission(auth.PermSettingsEdit), websiteProfileHandler.Update)
				settingsGroup.GET("/website-name", websiteNameHandler.Get)
				settingsGroup.PUT("/website-name", middleware.RequirePermission(auth.PermSettingsEdit), websiteNameHandler.Update)
				settingsGroup.GET("/payment", settingsHandler.GetPaymentSettings)
				settingsGroup.GET("/api", settingsHandler.GetAPISettings)
				settingsGroup.GET("/loyalty", settingsHandler.GetLoyaltySettings)
				settingsGroup.GET("/redeem", settingsHandler.GetRedeemSettings)
				settingsGroup.GET("/:key", settingsHandler.GetSetting)
			}

			// 物流包装箱规则与承运商管理（需要物流管理权限）
			shippingGroup := authenticated.Group("/shipping")
			shippingGroup.Use(middleware.RequirePermission(auth.PermShippingView))
			{
				shippingGroup.POST("/quote", shippingHandler.QuoteShipping)

				shippingGroup.GET("/templates", shippingHandler.ListTemplates)
				shippingGroup.GET("/templates/:id", shippingHandler.GetTemplate)
				shippingGroup.POST("/templates", middleware.RequirePermission(auth.PermShippingCreate), shippingHandler.CreateTemplate)
				shippingGroup.PUT("/templates/:id", middleware.RequirePermission(auth.PermShippingEdit), shippingHandler.UpdateTemplate)
				shippingGroup.DELETE("/templates/:id", middleware.RequirePermission(auth.PermShippingDelete), shippingHandler.DeleteTemplate)
				shippingGroup.POST("/templates/:id/rules", middleware.RequirePermission(auth.PermShippingEdit), shippingHandler.CreateTemplateRule)
				shippingGroup.PUT("/templates/:id/rules/:ruleId", middleware.RequirePermission(auth.PermShippingEdit), shippingHandler.UpdateTemplateRule)
				shippingGroup.DELETE("/templates/:id/rules/:ruleId", middleware.RequirePermission(auth.PermShippingEdit), shippingHandler.DeleteTemplateRule)

				shippingGroup.GET("/zones", shippingHandler.ListZones)
				shippingGroup.GET("/zones/:id", shippingHandler.GetZone)
				shippingGroup.POST("/zones", middleware.RequirePermission(auth.PermShippingCreate), shippingHandler.CreateZone)
				shippingGroup.PUT("/zones/:id", middleware.RequirePermission(auth.PermShippingEdit), shippingHandler.UpdateZone)
				shippingGroup.DELETE("/zones/:id", middleware.RequirePermission(auth.PermShippingDelete), shippingHandler.DeleteZone)

				shippingGroup.GET("/packaging-rules", shippingHandler.ListPackagingRules)
				shippingGroup.GET("/packaging-rules/:id", shippingHandler.GetPackagingRule)
				shippingGroup.POST("/packaging-rules", middleware.RequirePermission(auth.PermShippingCreate), shippingHandler.CreatePackagingRule)
				shippingGroup.PUT("/packaging-rules/:id", middleware.RequirePermission(auth.PermShippingEdit), shippingHandler.UpdatePackagingRule)
				shippingGroup.DELETE("/packaging-rules/:id", middleware.RequirePermission(auth.PermShippingDelete), shippingHandler.DeletePackagingRule)
				shippingGroup.POST("/packaging-rules/apply", middleware.RequirePermission(auth.PermShippingEdit), shippingHandler.CreatePackagingRuleApply)
				shippingGroup.DELETE("/packaging-rules/apply/:applyId", middleware.RequirePermission(auth.PermShippingEdit), shippingHandler.DeletePackagingRuleApply)

				// 承运商（Carriers）CRUD 管理端点
				shippingGroup.GET("/carriers", shippingHandler.ListCarriers)
				shippingGroup.GET("/carriers/:id", shippingHandler.GetCarrier)
				shippingGroup.POST("/carriers", middleware.RequirePermission(auth.PermShippingCreate), shippingHandler.CreateCarrier)
				shippingGroup.PUT("/carriers/:id", middleware.RequirePermission(auth.PermShippingEdit), shippingHandler.UpdateCarrier)
				shippingGroup.DELETE("/carriers/:id", middleware.RequirePermission(auth.PermShippingDelete), shippingHandler.DeleteCarrier)

				shippingGroup.GET("/carrier-services", shippingHandler.ListCarrierServices)
				shippingGroup.GET("/carrier-services/:id", shippingHandler.GetCarrierService)
				shippingGroup.POST("/carrier-services", middleware.RequirePermission(auth.PermShippingCreate), shippingHandler.CreateCarrierService)
				shippingGroup.PUT("/carrier-services/:id", middleware.RequirePermission(auth.PermShippingEdit), shippingHandler.UpdateCarrierService)
				shippingGroup.DELETE("/carrier-services/:id", middleware.RequirePermission(auth.PermShippingDelete), shippingHandler.DeleteCarrierService)

				shippingGroup.GET("/tracking-providers", shippingHandler.ListTrackingProviderConfigs)
				shippingGroup.GET("/tracking-providers/:id", shippingHandler.GetTrackingProviderConfig)
				shippingGroup.POST("/tracking-providers", middleware.RequirePermission(auth.PermShippingCreate), shippingHandler.CreateTrackingProviderConfig)
				shippingGroup.PUT("/tracking-providers/:id", middleware.RequirePermission(auth.PermShippingEdit), shippingHandler.UpdateTrackingProviderConfig)
				shippingGroup.DELETE("/tracking-providers/:id", middleware.RequirePermission(auth.PermShippingDelete), shippingHandler.DeleteTrackingProviderConfig)

				shippingGroup.GET("/tracking-carrier-mappings", shippingHandler.ListTrackingCarrierMappings)
				shippingGroup.GET("/tracking-carrier-mappings/:id", shippingHandler.GetTrackingCarrierMapping)
				shippingGroup.POST("/tracking-carrier-mappings", middleware.RequirePermission(auth.PermShippingCreate), shippingHandler.CreateTrackingCarrierMapping)
				shippingGroup.PUT("/tracking-carrier-mappings/:id", middleware.RequirePermission(auth.PermShippingEdit), shippingHandler.UpdateTrackingCarrierMapping)
				shippingGroup.DELETE("/tracking-carrier-mappings/:id", middleware.RequirePermission(auth.PermShippingDelete), shippingHandler.DeleteTrackingCarrierMapping)
				shippingGroup.GET("/tracking-shipments", shippingHandler.ListTrackingShipments)
				shippingGroup.GET("/tracking-polling", shippingHandler.GetTrackingPollingState)
				shippingGroup.GET("/tracking-webhook", shippingHandler.GetTrackingWebhookState)
				shippingGroup.GET("/tracking-shipments/:orderID/events", shippingHandler.ListTrackingEvents)
				shippingGroup.POST("/tracking-shipments/:orderID/register", middleware.RequirePermission(auth.PermShippingEdit), shippingHandler.RegisterTrackingShipment)
				shippingGroup.POST("/tracking-shipments/:orderID/sync", middleware.RequirePermission(auth.PermShippingEdit), shippingHandler.SyncTrackingShipment)
				shippingGroup.POST("/tracking-shipments/sync-due", middleware.RequirePermission(auth.PermShippingEdit), shippingHandler.SyncDueTrackingShipments)
			}

			// 运维中心：当前维护域名、连接器、VPS 和项目的声明式台账。
			opsGroup := authenticated.Group("/ops")
			opsGroup.Use(middleware.RequirePermission(auth.PermOpsView))
			{
				opsGroup.GET("/overview", opsOverviewHandler.Get)
				opsGroup.GET("/network/summary", opsNetworkSummaryHandler.Get)
				opsGroup.GET("/outbox/unknown", outboxReconciliationHandler.ListUnknown)
				opsGroup.POST("/outbox/unknown/:id/resume", middleware.RequirePermission(auth.PermSystemManage), outboxReconciliationHandler.Resume)
				opsGroup.POST("/outbox/unknown/:id/mark-processed", middleware.RequirePermission(auth.PermSystemManage), outboxReconciliationHandler.MarkProcessed)
				opsGroup.GET("/admin-accounts", middleware.AdminOnly(), adminAccountHandler.List)
				opsGroup.POST("/admin-accounts/ensure", middleware.AdminOnly(), adminAccountHandler.Ensure)
				opsGroup.GET("/deployments/preflight-overview", middleware.RequirePermission(auth.PermOpsDeployView), opsDeploymentPreflightHandler.GetOverview)
				opsGroup.GET("/deployments/preflight", middleware.RequirePermission(auth.PermOpsDeployView), opsDeploymentPreflightHandler.GetProjectReportByQuery)
				opsGroup.GET("/workflows", middleware.RequirePermission(auth.PermOpsDeployView), opsDeploymentWorkflowHandler.List)
				opsGroup.GET("/workflows/:id", middleware.RequirePermission(auth.PermOpsDeployView), opsDeploymentWorkflowHandler.Get)
				opsGroup.POST("/workflows", middleware.RequireAnyPermission(auth.PermOpsDeployDryRun, auth.PermOpsDeployExecute), opsDeploymentWorkflowHandler.Create)
				opsGroup.POST("/workflows/:id/validate", middleware.RequireAnyPermission(auth.PermOpsDeployDryRun, auth.PermOpsDeployExecute), opsDeploymentWorkflowHandler.Validate)
				opsGroup.POST("/workflows/:id/approve", middleware.RequirePermission(auth.PermOpsWorkflowApprove), opsDeploymentWorkflowHandler.Approve)
				opsGroup.POST("/workflows/:id/execute", middleware.RequireAnyPermission(auth.PermOpsDeployDryRun, auth.PermOpsDeployExecute), opsDeploymentWorkflowHandler.Execute)
				opsGroup.POST("/workflows/:id/retry", middleware.RequireAnyPermission(auth.PermOpsDeployDryRun, auth.PermOpsDeployExecute), opsDeploymentWorkflowHandler.RetryFailedStep)
				opsGroup.POST("/workflows/:id/rollback", middleware.RequirePermission(auth.PermOpsDeployRollback), opsDeploymentWorkflowHandler.Rollback)
				opsGroup.POST("/workflows/:id/cancel", middleware.RequireAnyPermission(auth.PermOpsDeployDryRun, auth.PermOpsDeployExecute), opsDeploymentWorkflowHandler.Cancel)

				domainsGroup := opsGroup.Group("/domains")
				domainsGroup.Use(middleware.RequirePermission(auth.PermOpsDomainView))
				{
					domainsGroup.GET("", opsDomainBindingHandler.List)
					domainsGroup.GET("/:id", opsDomainBindingHandler.Get)
					domainsGroup.GET("/:id/diff", opsDomainBindingHandler.Diff)
					domainsGroup.GET("/:id/preview", opsDomainBindingHandler.Preview)
					domainsGroup.POST("/:id/sync", middleware.RequirePermission(auth.PermOpsDomainSync), opsDomainBindingHandler.Sync)
					domainsGroup.POST("", middleware.RequirePermission(auth.PermOpsDomainEdit), opsDomainBindingHandler.Create)
					domainsGroup.PUT("/:id", middleware.RequirePermission(auth.PermOpsDomainEdit), opsDomainBindingHandler.Update)
					domainsGroup.PATCH("/:id/enabled", middleware.RequirePermission(auth.PermOpsDomainEdit), opsDomainBindingHandler.UpdateStatus)
				}

				connectorsGroup := opsGroup.Group("/connectors")
				connectorsGroup.Use(middleware.RequirePermission(auth.PermOpsConnectorView))
				{
					connectorsGroup.GET("", opsConnectorHandler.List)
					connectorsGroup.GET("/:id", opsConnectorHandler.Get)
					connectorsGroup.POST("", middleware.RequirePermission(auth.PermOpsConnectorEdit), opsConnectorHandler.Create)
					connectorsGroup.PUT("/:id", middleware.RequirePermission(auth.PermOpsConnectorEdit), opsConnectorHandler.Update)
					connectorsGroup.PATCH("/:id/enabled", middleware.RequirePermission(auth.PermOpsConnectorEdit), opsConnectorHandler.UpdateStatus)
					connectorsGroup.POST("/:id/test", middleware.RequirePermission(auth.PermOpsConnectorEdit), opsConnectorHandler.Test)
					connectorsGroup.POST("/oauth/start", middleware.RequirePermission(auth.PermOpsConnectorEdit), opsConnectorHandler.StartOAuth)
				}

				vpsGroup := opsGroup.Group("/vps")
				vpsGroup.Use(middleware.RequirePermission(auth.PermOpsVPSView))
				{
					vpsGroup.GET("", opsVPSBindingHandler.List)
					vpsGroup.GET("/:id", opsVPSBindingHandler.Get)
					vpsGroup.POST("/:id/sync", middleware.RequirePermission(auth.PermOpsVPSSync), opsVPSBindingHandler.Sync)
					vpsGroup.POST("", middleware.RequirePermission(auth.PermOpsVPSEdit), opsVPSBindingHandler.Create)
					vpsGroup.PUT("/:id", middleware.RequirePermission(auth.PermOpsVPSEdit), opsVPSBindingHandler.Update)
					vpsGroup.PATCH("/:id/enabled", middleware.RequirePermission(auth.PermOpsVPSEdit), opsVPSBindingHandler.UpdateStatus)
				}

				projectsGroup := opsGroup.Group("/projects")
				projectsGroup.Use(middleware.RequirePermission(auth.PermOpsProjectView))
				{
					projectsGroup.GET("", opsProjectBindingHandler.List)
					projectsGroup.GET("/:id/preflight", middleware.RequirePermission(auth.PermOpsDeployView), opsDeploymentPreflightHandler.GetProjectReport)
					projectsGroup.GET("/:id", opsProjectBindingHandler.Get)
					projectsGroup.POST("/:id/sync", middleware.RequirePermission(auth.PermOpsProjectSync), opsProjectBindingHandler.Sync)
					projectsGroup.POST("", middleware.RequirePermission(auth.PermOpsProjectEdit), opsProjectBindingHandler.Create)
					projectsGroup.PUT("/:id", middleware.RequirePermission(auth.PermOpsProjectEdit), opsProjectBindingHandler.Update)
					projectsGroup.PATCH("/:id/enabled", middleware.RequirePermission(auth.PermOpsProjectEdit), opsProjectBindingHandler.UpdateStatus)
				}
			}

			// 审计日志（需要日志查看权限）
			logsGroup := authenticated.Group("/logs")
			logsGroup.Use(middleware.RequirePermission(auth.PermLogsView))
			{
				logsGroup.GET("", auditHandler.ListAuditLogs)
				logsGroup.GET("/stats", auditHandler.GetAuditStats)
				logsGroup.GET("/recent", auditHandler.GetRecentActivities)
				logsGroup.GET("/search", auditHandler.SearchAuditLogs)
				logsGroup.GET("/:id", auditHandler.GetAuditLog)
				logsGroup.GET("/user/:user_id", auditHandler.GetUserAuditLogs)
				logsGroup.POST("/cleanup", middleware.AdminOnly(), auditHandler.DeleteOldLogs)
			}
		}
	}
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

func registerSiteQualityRoutes(group *gin.RouterGroup, handler *SiteQualityHandler, managePermission auth.Permission) {
	group.GET("", handler.ListSiteQualityRuns)
	group.GET("/targets", handler.ListSiteQualityTargets)
	group.POST("/jobs", middleware.RequirePermission(managePermission), handler.CreateSiteQualityJob)
	group.GET("/jobs/:id", handler.GetSiteQualityJob)
	group.GET("/findings", handler.ListSiteQualityFindings)
	group.GET("/findings/:id", handler.GetSiteQualityFinding)
	group.GET("/findings/:id/events", handler.ListSiteQualityFindingEvents)
	group.POST("/findings/:id/acknowledge", middleware.RequirePermission(managePermission), handler.AcknowledgeSiteQualityFinding)
	group.POST("/findings/:id/resolve", middleware.RequirePermission(managePermission), handler.ResolveSiteQualityFinding)
	group.POST("/findings/:id/recheck", middleware.RequirePermission(managePermission), handler.RecheckSiteQualityFinding)
}

func registerMediaImageDimensionRoutes(group *gin.RouterGroup, handler *MediaImageDimensionsHandler) {
	group.GET("", handler.List)
	group.POST("/:id/reconcile", middleware.RequirePermission(auth.PermMediaEdit), handler.Reconcile)
}

func registerPreflightContentLinkRoutes(group *gin.RouterGroup, handler *ContentLinkPreflightHandler, managePermission auth.Permission) {
	group.GET("/targets", handler.ListTargets)
	group.POST("/runs", middleware.RequirePermission(managePermission), handler.Run)
	group.GET("/issues", handler.ListIssues)
	group.GET("/issues/:id", handler.GetIssue)
	group.GET("/issues/:id/events", handler.ListIssueEvents)
	group.GET("/stats", handler.Stats)
	group.POST("/issues/:id/apply", middleware.RequirePermission(managePermission), handler.ApplySuggestion)
	group.POST("/issues/:id/resolve", middleware.RequirePermission(managePermission), handler.Resolve)
	group.POST("/issues/:id/recheck", middleware.RequirePermission(managePermission), handler.Recheck)
}
