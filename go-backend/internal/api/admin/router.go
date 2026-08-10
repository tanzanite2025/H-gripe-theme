package admin

import (
	seoapi "tanzanite/internal/api/admin/seo"
	"tanzanite/internal/api/middleware"
	"tanzanite/internal/api/v1/showcase"
	"tanzanite/internal/app"
	"tanzanite/internal/domain/auth"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/securecookie"

	"github.com/gin-gonic/gin"
)

// RegisterAdminRoutes 注册管理后台路由
func RegisterAdminRoutes(r *gin.Engine, deps *app.Dependencies, cfg *config.Config) {
	// 初始化 repositories
	services := deps.Services
	authService := services.Auth
	showcaseService := services.Showcase
	registrationService := services.Registration
	userService := services.User
	postService := services.Post
	productService := services.Product
	orderService := services.Order
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
	dashboardHandler := NewDashboardHandler(dashboardService)
	userHandler := NewUserHandler(userService)
	customerHandler := NewCustomerHandler(userService)
	productHandler := NewProductHandler(productService)
	productTypeImageHandler := NewProductTypeImageHandler(productService, services.Media)
	productInformationTemplateHandler := NewProductInformationTemplateHandler(services.ProductInformationTemplate)
	spokeCatalogHandler := NewSpokeCatalogHandler(services.Spoke)
	quickBuyHandler := NewQuickBuyHandler(services.QuickBuy)
	mediaHandler := NewMediaHandler(services.Media)
	orderHandler := NewOrderHandler(orderService)
	paymentHandler := NewPaymentHandler(paymentService, services.AdminSettings)
	paymentHandler.ConfigurePublicBaseURL(cfg.Server.BaseURL)
	paymentHandler.ConfigureAuditService(services.Audit)
	paymentRefundExecutionHandler := NewPaymentRefundExecutionHandler(paymentService, services.AdminSettings)
	paymentRefundExecutionHandler.ConfigureAuditService(services.Audit)
	paymentRiskMonitoringHandler := NewPaymentRiskMonitoringHandler(services.PaymentRiskMonitoring)
	paymentRiskMonitoringHandler.ConfigureAuditService(services.Audit)
	paymentProtectionHandler := NewPaymentProtectionHandler(services.PaymentProtection)
	paymentProtectionHandler.ConfigureAuditService(services.Audit)
	paymentRefundRecommendationHandler := NewPaymentRefundRecommendationHandler(services.PaymentRefundReview)
	paymentRefundRecommendationHandler.ConfigureAuditService(services.Audit)
	contentHandler := NewContentHandler(postService)
	faqHandler := NewFAQHandler(services.FAQ)
	galleryHandler := NewGalleryHandler(services.Gallery)
	subscriptionHandler := NewSubscriptionHandler(services.Subscription)
	ticketHandler := NewTicketHandler(services.Ticket, services.CustomerServiceContext, services.CustomerServiceEvents, services.Media)
	autoReplyHandler := NewAutoReplyHandler(services.Ticket, services.FAQ)
	visitorProfileHandler := NewVisitorProfileHandler(services.VisitorProfile)
	visitorRiskHandler := NewVisitorRiskHandler(services.VisitorRisk)
	visitorRiskHandler.ConfigureAuditService(services.Audit)
	marketingHandler := NewMarketingHandler(marketingService, services.LoyaltyProgram)
	settingsHandler := NewSettingsHandler(services.AdminSettings)
	seoHomeHandler := seoapi.NewHomeHandler(services.SEO)
	seoArticlesHandler := seoapi.NewArticlesHandler(services.SEOResources)
	seoProductsHandler := seoapi.NewProductsHandler(services.SEOResources)
	seoHomeHandler.ConfigureAuditService(services.Audit)
	seoArticlesHandler.ConfigureAuditService(services.Audit)
	seoProductsHandler.ConfigureAuditService(services.Audit)
	analyticsHandler := NewAnalyticsHandler(services.Analytics)
	commercialCrawlerHandler := NewCommercialCrawlerProtectionHandler(orderService)
	currencyPolicyHandler := NewCurrencyPolicyHandler(services.CurrencyPolicy)
	currencyPolicyHandler.ConfigureAuditService(services.Audit)
	exchangeRateHandler := NewExchangeRateHandler(services.ExchangeRate)
	exchangeRateHandler.ConfigureAuditService(services.Audit)
	storefrontMarketHandler := NewStorefrontMarketHandler(services.StorefrontMarket)
	storefrontMarketHandler.ConfigureAuditService(services.Audit)
	googleMerchantHandler := NewGoogleMerchantHandler(services.GoogleMerchant, cfg.GoogleMerchant.PostConnectURL)
	publicChatAgentHandler := NewPublicChatAgentHandler(services.AdminPublicChat)
	auditHandler := NewAuditHandler(services.Audit)
	showcaseHandler := showcase.NewShowcaseHandler(showcaseService)
	registrationHandler := NewRegistrationHandler(registrationService)
	shippingHandler := NewShippingHandler(services.Shipping)

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
				dashboardGroup.GET("/recent-tickets", dashboardHandler.GetRecentTickets)
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
			productTypesGroup := authenticated.Group("/product-types")
			productTypesGroup.Use(middleware.RequirePermission(auth.PermProductView))
			{
				productTypesGroup.GET("", productHandler.ListProductTypes)
				productTypesGroup.GET("/:id", productHandler.GetProductType)
				productTypesGroup.POST("", middleware.RequirePermission(auth.PermProductCreate), productHandler.CreateProductType)
				productTypesGroup.PUT("/:id", middleware.RequirePermission(auth.PermProductEdit), productHandler.UpdateProductType)
				productTypesGroup.POST("/:id/image", middleware.RequirePermission(auth.PermProductEdit), productTypeImageHandler.UploadImage)
				productTypesGroup.DELETE("/:id/image", middleware.RequirePermission(auth.PermProductEdit), productTypeImageHandler.DeleteImage)
				productTypesGroup.DELETE("/:id", middleware.RequirePermission(auth.PermProductDelete), productHandler.DeleteProductType)
			}

			productsGroup := authenticated.Group("/products")
			productsGroup.Use(middleware.RequirePermission(auth.PermProductView))
			{
				productsGroup.GET("", productHandler.ListProducts)
				productsGroup.GET("/stats", productHandler.GetProductStats)
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
				quickBuyGroup.POST("/flows/:id/draft", middleware.RequirePermission(auth.PermProductEdit), quickBuyHandler.CreateDraftVersion)
				quickBuyGroup.PUT("/flow-versions/:version_id", middleware.RequirePermission(auth.PermProductEdit), quickBuyHandler.UpdateDraftVersion)
				quickBuyGroup.POST("/flow-versions/:version_id/validate", quickBuyHandler.ValidateVersion)
				quickBuyGroup.POST("/flow-versions/:version_id/preview", quickBuyHandler.PreviewVersionCandidates)
				quickBuyGroup.POST("/flow-versions/:version_id/publish", middleware.RequirePermission(auth.PermProductEdit), quickBuyHandler.PublishVersion)
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

			mediaGroup := authenticated.Group("/media")
			mediaGroup.Use(middleware.RequirePermission(auth.PermMediaView))
			{
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
				ordersGroup.GET("/stats", orderHandler.GetOrderStats)
				ordersGroup.GET("/sales-chart", orderHandler.GetSalesChart)
				ordersGroup.GET("/export", orderHandler.ExportOrders)
				ordersGroup.GET("/:id", orderHandler.GetOrder)
				ordersGroup.PATCH("/:id/status", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.UpdateOrderStatus)
				ordersGroup.PATCH("/:id/shipping-status", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.UpdateShippingStatus)
				ordersGroup.PATCH("/:id/tracking", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.UpdateTrackingInfo)
				ordersGroup.POST("/:id/tracking/sync", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.SyncTrackingInfo)
				ordersGroup.PATCH("/:id/admin-note", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.UpdateAdminNote)
				ordersGroup.POST("/batch-status", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.BatchUpdateStatus)
				ordersGroup.DELETE("/:id", middleware.RequirePermission(auth.PermOrderDelete), orderHandler.DeleteOrder)
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
				showcaseGroup.PUT("/:id/approve", middleware.RequirePermission(auth.PermGalleryEdit), showcaseHandler.Approve)
				showcaseGroup.PUT("/:id/reject", middleware.RequirePermission(auth.PermGalleryEdit), showcaseHandler.Reject)
			}

			// 产品注册与保修管理（需要商品管理权限）
			registrationsGroup := authenticated.Group("/registrations")
			registrationsGroup.Use(middleware.RequirePermission(auth.PermProductView))
			{
				registrationsGroup.GET("", registrationHandler.ListAllRegistrations)
				registrationsGroup.PUT("/:id/status", middleware.RequirePermission(auth.PermProductEdit), registrationHandler.UpdateRegistrationStatus)
				registrationsGroup.GET("/expiring", registrationHandler.GetExpiringWarranties)
				registrationsGroup.GET("/stats", registrationHandler.GetRegistrationStats)
				registrationsGroup.GET("/warranty-claims", registrationHandler.ListAllWarrantyClaims)
				registrationsGroup.GET("/warranty-claims/:id", registrationHandler.GetWarrantyClaim)
				registrationsGroup.GET("/warranty-claims/:id/order-items", registrationHandler.ListWarrantyClaimOrderItems)
				registrationsGroup.PUT("/warranty-claims/:id/order-item", middleware.RequirePermission(auth.PermProductEdit), registrationHandler.BindWarrantyClaimOrderItem)
				registrationsGroup.GET("/warranty-claims/:id/service-records", registrationHandler.ListWarrantyServiceRecords)
				registrationsGroup.POST("/warranty-claims/:id/service-records", middleware.RequirePermission(auth.PermProductEdit), registrationHandler.CreateWarrantyServiceRecord)
				registrationsGroup.PUT("/warranty-claims/:id/status", middleware.RequirePermission(auth.PermProductEdit), registrationHandler.UpdateWarrantyClaimStatus)
				registrationsGroup.PUT("/warranty-claims/:id/resolution", middleware.RequirePermission(auth.PermProductEdit), registrationHandler.UpdateWarrantyClaimResolution)
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

			// 工单管理（需要工单管理权限）
			ticketsGroup := authenticated.Group("/tickets")
			ticketsGroup.Use(middleware.RequirePermission(auth.PermTicketView))
			{
				ticketsGroup.GET("", ticketHandler.ListTickets)
				ticketsGroup.GET("/stats", ticketHandler.GetTicketStats)
				ticketsGroup.GET("/:id", ticketHandler.GetTicket)
				ticketsGroup.PUT("/:id", middleware.RequirePermission(auth.PermTicketEdit), ticketHandler.UpdateTicket)
				ticketsGroup.PATCH("/:id/status", middleware.RequirePermission(auth.PermTicketEdit), ticketHandler.UpdateTicketStatus)
				ticketsGroup.PATCH("/:id/assign", middleware.RequirePermission(auth.PermTicketEdit), ticketHandler.AssignTicket)
				ticketsGroup.DELETE("/:id", middleware.RequirePermission(auth.PermTicketDelete), ticketHandler.DeleteTicket)

				// 工单消息
				ticketsGroup.GET("/:id/messages", ticketHandler.GetMessages)
				ticketsGroup.POST("/:id/messages", middleware.RequirePermission(auth.PermTicketEdit), ticketHandler.CreateMessage)
				ticketsGroup.POST("/:id/messages/mark-read", ticketHandler.MarkMessagesAsRead)
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
				customerServiceGroup.GET("/analytics/regions", ticketHandler.GetCustomerServiceRegionAnalytics)
				customerServiceGroup.GET("/conversations", ticketHandler.ListCustomerServiceConversations)
				customerServiceGroup.GET("/events", ticketHandler.StreamCustomerServiceEvents)
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
				customerServiceGroup.POST("/conversations/:id/typing", middleware.RequirePermission(auth.PermTicketEdit), ticketHandler.SendCustomerServiceConversationTyping)
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
				seoGroup.GET("/home", seoHomeHandler.Get)
				seoGroup.PUT("/home", middleware.RequirePermission(auth.PermSEOEdit), seoHomeHandler.Update)
				seoGroup.GET("/articles", seoArticlesHandler.Get)
				seoGroup.PUT("/articles/:id", middleware.RequirePermission(auth.PermSEOEdit), seoArticlesHandler.Update)
				seoGroup.GET("/products", seoProductsHandler.Get)
				seoGroup.PUT("/products/:id", middleware.RequirePermission(auth.PermSEOEdit), seoProductsHandler.Update)
			}

			analyticsGroup := authenticated.Group("/analytics")
			analyticsGroup.Use(middleware.RequirePermission(auth.PermAnalyticsView))
			{
				analyticsGroup.GET("", analyticsHandler.Get)
				analyticsGroup.PUT("", middleware.RequirePermission(auth.PermAnalyticsEdit), analyticsHandler.Update)
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
				settingsGroup.GET("/public-chat-groups", publicChatAgentHandler.ListPublicChatGroups)
				settingsGroup.POST("/public-chat-groups", middleware.RequirePermission(auth.PermSettingsEdit), publicChatAgentHandler.UpsertPublicChatGroup)
				settingsGroup.PUT("/public-chat-groups/:id", middleware.RequirePermission(auth.PermSettingsEdit), publicChatAgentHandler.UpdatePublicChatGroup)
				settingsGroup.DELETE("/public-chat-groups/:id", middleware.RequirePermission(auth.PermSettingsEdit), publicChatAgentHandler.DeletePublicChatGroup)
				settingsGroup.GET("/payment-runtime", paymentHandler.GetGatewayRuntimeStatus)
				settingsGroup.POST("/payment-runtime/:provider/callback-check", middleware.RequirePermission(auth.PermSettingsEdit), paymentHandler.CheckGatewayCallback)
				settingsGroup.GET("/currency-policy", currencyPolicyHandler.GetPolicy)
				settingsGroup.PUT("/currency-policy", middleware.RequirePermission(auth.PermSettingsEdit), currencyPolicyHandler.UpdatePolicy)
				settingsGroup.GET("/exchange-rates", exchangeRateHandler.GetExchangeRates)
				settingsGroup.POST("/exchange-rates/sync", middleware.RequirePermission(auth.PermSettingsEdit), exchangeRateHandler.SyncExchangeRates)
				settingsGroup.POST("/exchange-rates/convert", middleware.RequirePermission(auth.PermSettingsEdit), exchangeRateHandler.ConvertDisplayPrices)
				settingsGroup.PUT("/payment-gateways/:provider", middleware.RequirePermission(auth.PermSettingsEdit), paymentHandler.UpsertGatewayConfig)
				settingsGroup.DELETE("/payment-gateways/:provider", middleware.RequirePermission(auth.PermSettingsEdit), paymentHandler.DeleteGatewayConfig)
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
