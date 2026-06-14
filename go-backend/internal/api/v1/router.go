package v1

import (
	"tanzanite/internal/api/middleware"
	"tanzanite/internal/api/v1/audit"
	"tanzanite/internal/api/v1/auth"
	"tanzanite/internal/api/v1/cart"
	"tanzanite/internal/api/v1/content"
	"tanzanite/internal/api/v1/faq"
	"tanzanite/internal/api/v1/feedback"
	"tanzanite/internal/api/v1/gallery"
	"tanzanite/internal/api/v1/i18n"
	"tanzanite/internal/api/v1/marketing"
	"tanzanite/internal/api/v1/order"
	"tanzanite/internal/api/v1/payment"
	"tanzanite/internal/api/v1/product"
	"tanzanite/internal/api/v1/registration"
	"tanzanite/internal/api/v1/review"
	"tanzanite/internal/api/v1/settings"
	"tanzanite/internal/api/v1/shipping"
	"tanzanite/internal/api/v1/showcase"
	"tanzanite/internal/api/v1/subscription"
	"tanzanite/internal/api/v1/suggestionfeedback"
	"tanzanite/internal/api/v1/ticket"
	"tanzanite/internal/api/v1/wishlist"
	"tanzanite/internal/pkg/cache"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/storage"
	"tanzanite/internal/repository"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes 注册所有v1路由
func RegisterRoutes(r *gin.Engine, db *gorm.DB, redisCache *cache.RedisCache, cfg *config.Config) {
	// 初始化repositories
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	productRepo := repository.NewProductRepository(db)
	cartRepo := repository.NewCartRepository(db)
	settingRepo := repository.NewSettingRepository(db)
	faqRepo := repository.NewFAQRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	shippingRepo := repository.NewShippingRepository(db)
	couponRepo := repository.NewCouponRepository(db)
	loyaltyRepo := repository.NewLoyaltyRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	galleryRepo := repository.NewGalleryRepository(db)
	registrationRepo := repository.NewRegistrationRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	showcaseRepo := repository.NewShowcaseRepository(db)
	wishlistRepo := repository.NewWishlistRepository(db)
	feedbackRepo := repository.NewFeedbackRepository(db)
	suggestionFeedbackRepo := repository.NewSuggestionFeedbackRepository(db)

	// 初始化services
	authService := service.NewAuthService(userRepo, cfg.JWT)
	postService := service.NewPostService(postRepo, redisCache, cfg.Cache.PostTTL)
	productService := service.NewProductService(productRepo, redisCache, cfg.Cache.ProductTTL)
	cartService := service.NewCartService(cartRepo, productRepo)
	settingService := service.NewSettingService(settingRepo, redisCache, cfg.Cache.SettingsTTL)
	faqService := service.NewFAQService(faqRepo)
	galleryService := service.NewGalleryService(galleryRepo)
	// registrationService := service.NewRegistrationService(registrationRepo, productRepo)
	orderService := service.NewOrderService(orderRepo, productRepo, couponRepo, paymentRepo, shippingRepo, auditRepo)
	marketingService := service.NewMarketingService(couponRepo, loyaltyRepo)
	reviewService := service.NewReviewService(reviewRepo)
	ticketService := service.NewTicketService(ticketRepo)
	subscriptionRepo := repository.NewSubscriptionRepository(db)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo)
	sitemapService := service.NewSitemapService(postRepo, cfg.Server.BaseURL)

	storageSvc, _ := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: "./uploads", BaseURL: cfg.Server.BaseURL})
	showcaseService := service.NewShowcaseService(showcaseRepo, storageSvc)
	wishlistService := service.NewWishlistService(wishlistRepo, productRepo)
	feedbackService := service.NewFeedbackService(feedbackRepo)
	suggestionFeedbackService := service.NewSuggestionFeedbackService(suggestionFeedbackRepo)

	// 初始化handlers
	authHandler := auth.NewHandler(authService)
	contentHandler := content.NewHandler(postService, faqService)
	faqHandler := faq.NewHandler(faqService)
	productHandler := product.NewHandler(productService)
	cartHandler := cart.NewHandler(cartService)
	settingsHandler := settings.NewHandler(settingService)
	orderHandler := order.NewHandler(orderService)
	marketingHandler := marketing.NewHandler(marketingService)
	reviewHandler := review.NewHandler(reviewService)
	ticketHandler := ticket.NewHandler(ticketService)
	paymentHandler := payment.NewHandler(paymentRepo)
	shippingHandler := shipping.NewHandler(shippingRepo)
	galleryHandler := gallery.NewGalleryHandler(galleryService)
	registrationHandler := registration.NewHandler(registrationRepo)
	auditHandler := audit.NewHandler(auditRepo)
	subscriptionHandler := subscription.NewHandler(subscriptionService)
	i18nHandler := i18n.NewHandler(postService, sitemapService)
	showcaseHandler := showcase.NewShowcaseHandler(showcaseService)
	wishlistHandler := wishlist.NewHandler(wishlistService)
	feedbackHandler := feedback.NewHandler(feedbackService)
	suggestionFeedbackHandler := suggestionfeedback.NewHandler(suggestionFeedbackService)
	registerWordPressCompatRoutes(r, postService)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 认证路由（公开）
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/logout", authHandler.Logout)
			authGroup.GET("/profile", middleware.AuthMiddleware(authService), authHandler.GetProfile)
		}

		// 内容路由（公开）
		contentGroup := v1.Group("/content")
		{
			contentGroup.GET("/posts", contentHandler.ListPosts)
			contentGroup.GET("/posts/:id", contentHandler.GetPost)
			contentGroup.GET("/faqs", contentHandler.ListFAQs)
			contentGroup.GET("/faqs/:id", contentHandler.GetFAQ)
			contentGroup.GET("/faq-categories", contentHandler.GetFAQCategories)
			contentGroup.GET("/faqs/search", contentHandler.SearchFAQs)
			contentGroup.GET("/faqs/category/:category", contentHandler.GetFAQsByCategory)
			contentGroup.GET("/faqs/popular", contentHandler.GetPopularFAQs)
			contentGroup.POST("/faqs/:id/view", faqHandler.IncrementFAQView)
		}

		// FAQ 管理路由（需要认证和管理员权限）
		adminFAQGroup := v1.Group("/admin/faqs")
		adminFAQGroup.Use(middleware.AuthMiddleware(authService))
		// TODO: 添加 AdminMiddleware 检查用户角色
		{
			adminFAQGroup.POST("", faqHandler.CreateFAQ)
			adminFAQGroup.PUT("/:id", faqHandler.UpdateFAQ)
			adminFAQGroup.DELETE("/:id", faqHandler.DeleteFAQ)
			adminFAQGroup.PUT("/:id/order", faqHandler.UpdateFAQOrder)
			adminFAQGroup.POST("/batch-order", faqHandler.BatchUpdateFAQOrder)
		}

		// 产品路由（公开）
		productGroup := v1.Group("/products")
		{
			productGroup.GET("", productHandler.ListProducts)
			productGroup.GET("/:id", productHandler.GetProduct)
		}

		// 购物车路由（可选认证）
		cartGroup := v1.Group("/cart")
		cartGroup.Use(middleware.OptionalAuthMiddleware(authService))
		{
			cartGroup.GET("/summary", cartHandler.GetCartSummary)
			cartGroup.POST("/add", cartHandler.AddToCart)
			cartGroup.PUT("/items/:id", cartHandler.UpdateCartItem)
			cartGroup.DELETE("/items/:id", cartHandler.RemoveFromCart)
		}

		wishlistGroup := v1.Group("/wishlist")
		wishlistGroup.Use(middleware.AuthMiddleware(authService))
		{
			wishlistGroup.GET("", wishlistHandler.ListItems)
			wishlistGroup.POST("", wishlistHandler.CreateItem)
			wishlistGroup.DELETE("/:id", wishlistHandler.DeleteItem)
		}

		feedbackGroup := v1.Group("/feedback")
		{
			feedbackGroup.GET("", feedbackHandler.List)
			feedbackGroup.GET("/eligibility", middleware.OptionalAuthMiddleware(authService), feedbackHandler.Eligibility)
			feedbackGroup.POST("", middleware.AuthMiddleware(authService), feedbackHandler.Create)
		}

		suggestionFeedbackGroup := v1.Group("/suggestion-feedback")
		{
			suggestionFeedbackGroup.GET("/eligibility", middleware.OptionalAuthMiddleware(authService), suggestionFeedbackHandler.Eligibility)
			suggestionFeedbackGroup.POST("", middleware.AuthMiddleware(authService), suggestionFeedbackHandler.Create)
		}

		// 订单路由（需要认证）
		orderGroup := v1.Group("/orders")
		orderGroup.Use(middleware.AuthMiddleware(authService))
		{
			orderGroup.POST("", orderHandler.CreateOrder)
			orderGroup.GET("", orderHandler.ListOrders)
			orderGroup.GET("/stats", orderHandler.GetOrderStats)
			orderGroup.GET("/:id", orderHandler.GetOrder)
			orderGroup.PUT("/:id/status", orderHandler.UpdateOrderStatus)
			orderGroup.POST("/:id/cancel", orderHandler.CancelOrder)
		}

		// 营销路由
		marketingGroup := v1.Group("/marketing")
		{
			// 优惠券（公开）
			marketingGroup.GET("/coupons", marketingHandler.ListCoupons)

			// 等级配置（公开）
			marketingGroup.GET("/loyalty/levels", marketingHandler.ListMemberLevels)

			// 需要认证的营销功能
			authMarketing := marketingGroup.Group("")
			authMarketing.Use(middleware.AuthMiddleware(authService))
			{
				// 优惠券
				authMarketing.POST("/coupons/validate", marketingHandler.ValidateCoupon)

				// 礼品卡
				authMarketing.GET("/gift-cards", marketingHandler.ListGiftCards)
				authMarketing.POST("/gift-cards", marketingHandler.CreateGiftCard)
				authMarketing.POST("/gift-cards/use", marketingHandler.UseGiftCard)

				// 积分和会员
				authMarketing.GET("/loyalty/assets", marketingHandler.GetUserAssets)
				authMarketing.GET("/loyalty/points", marketingHandler.GetPoints)
				authMarketing.GET("/loyalty/info", marketingHandler.GetLoyaltyInfo)
				authMarketing.POST("/loyalty/checkin", marketingHandler.CheckIn)
				authMarketing.POST("/loyalty/referral", marketingHandler.CreateReferral)
				authMarketing.POST("/loyalty/spend", marketingHandler.SpendPoints)
			}
		}

		// 评价路由
		reviewGroup := v1.Group("/reviews")
		{
			// 公开评价
			reviewGroup.GET("", reviewHandler.ListProductReviews)
			reviewGroup.GET("/featured", reviewHandler.GetFeaturedReviews)
			reviewGroup.GET("/summary/:product_id", reviewHandler.GetReviewSummary)
			reviewGroup.GET("/:id", reviewHandler.GetReview)

			// 需要认证的评价功能
			authReview := reviewGroup.Group("")
			authReview.Use(middleware.AuthMiddleware(authService))
			{
				authReview.POST("", reviewHandler.CreateReview)
				authReview.GET("/my", reviewHandler.ListUserReviews)
				authReview.DELETE("/:id", reviewHandler.DeleteReview)
				authReview.POST("/:id/helpful", reviewHandler.MarkHelpful)
			}
		}

		// 工单路由（需要认证）
		ticketGroup := v1.Group("/tickets")
		ticketGroup.Use(middleware.AuthMiddleware(authService))
		{
			ticketGroup.POST("", ticketHandler.CreateTicket)
			ticketGroup.GET("", ticketHandler.ListTickets)
			ticketGroup.GET("/stats", ticketHandler.GetTicketStats)
			ticketGroup.GET("/:id", ticketHandler.GetTicket)
			ticketGroup.PUT("/:id/status", ticketHandler.UpdateTicketStatus)
			ticketGroup.POST("/:id/close", ticketHandler.CloseTicket)
			ticketGroup.POST("/:id/messages", ticketHandler.AddMessage)
			ticketGroup.GET("/:id/messages", ticketHandler.GetMessages)
		}

		// Showcase (Picture Warehouse)
		showcaseGroup := v1.Group("/showcase")
		{
			showcaseGroup.GET("/gallery", showcaseHandler.List)
			showcaseGroup.GET("/comments", showcaseHandler.ListComments)
			showcaseGroup.POST("/upload", middleware.AuthMiddleware(authService), showcaseHandler.Upload)
			showcaseGroup.POST("/comments", middleware.AuthMiddleware(authService), showcaseHandler.AddComment)
		}

		// Admin Showcase
		adminShowcaseGroup := v1.Group("/admin/showcase")
		adminShowcaseGroup.Use(middleware.AuthMiddleware(authService))
		{
			adminShowcaseGroup.GET("", showcaseHandler.List)
			adminShowcaseGroup.PUT("/:id/approve", showcaseHandler.Approve)
			adminShowcaseGroup.PUT("/:id/reject", showcaseHandler.Reject)
		}

		// 设置路由
		settingsGroup := v1.Group("/settings")
		{
			// 公开设置
			settingsGroup.GET("/site", settingsHandler.GetSiteSettings)
			settingsGroup.GET("/quick-buy", settingsHandler.GetQuickBuySettings)
			settingsGroup.GET("/seo", settingsHandler.GetSEOSettings)
			settingsGroup.GET("/social", settingsHandler.GetSocialSettings)
			settingsGroup.GET("/public", settingsHandler.GetAllPublicSettings)
			settingsGroup.GET("/groups", settingsHandler.GetGroups)
			settingsGroup.GET("/group/:group", settingsHandler.GetSettingsByGroup)
			settingsGroup.GET("/:key", settingsHandler.GetSetting)
		}

		// 管理员设置路由（需要认证和管理员权限）
		adminSettingsGroup := v1.Group("/admin/settings")
		adminSettingsGroup.Use(middleware.AuthMiddleware(authService))
		// TODO: 添加 AdminMiddleware 检查用户角色
		{
			adminSettingsGroup.GET("", settingsHandler.GetAllSettings)
			adminSettingsGroup.GET("/email", settingsHandler.GetEmailSettings)
			adminSettingsGroup.POST("", settingsHandler.UpdateSetting)
			adminSettingsGroup.POST("/batch", settingsHandler.BatchUpdateSettings)
			adminSettingsGroup.DELETE("/:key", settingsHandler.DeleteSetting)
		}

		// i18n 路由（公开）
		i18nGroup := v1.Group("/i18n")
		{
			i18nGroup.GET("/languages", i18nHandler.GetLanguages)
			i18nGroup.GET("/translations/:post_id", i18nHandler.GetPostTranslations)
			i18nGroup.GET("/detect", i18nHandler.DetectLanguage)
			i18nGroup.POST("/set-language", i18nHandler.SetLanguage)
		}

		// 订阅路由
		subscriptionGroup := v1.Group("/subscriptions")
		{
			// 公开端点
			subscriptionGroup.POST("", subscriptionHandler.Subscribe)
			subscriptionGroup.GET("/unsubscribe/:token", subscriptionHandler.Unsubscribe)
			subscriptionGroup.POST("/unsubscribe", subscriptionHandler.UnsubscribeByEmail)
			subscriptionGroup.POST("/resubscribe", subscriptionHandler.Resubscribe)
			subscriptionGroup.GET("/status/:email", subscriptionHandler.GetSubscription)
		}

		// 支付路由
		paymentGroup := v1.Group("/payment")
		{
			// 公开端点
			paymentGroup.GET("/methods", paymentHandler.ListPaymentMethods)
			paymentGroup.GET("/methods/:id", paymentHandler.GetPaymentMethod)
			paymentGroup.GET("/tax-rates", paymentHandler.ListTaxRates)
			paymentGroup.GET("/tax-rates/:id", paymentHandler.GetTaxRate)
			paymentGroup.POST("/calculate-tax", paymentHandler.CalculateTax)

			// 需要认证的端点
			authPayment := paymentGroup.Group("")
			authPayment.Use(middleware.AuthMiddleware(authService))
			{
				authPayment.POST("/transactions", paymentHandler.CreateTransaction)
				authPayment.GET("/transactions/:id", paymentHandler.GetTransaction)
				authPayment.GET("/orders/:order_id/transactions", paymentHandler.GetOrderTransactions)
				authPayment.POST("/refunds", paymentHandler.CreateRefund)
				authPayment.GET("/refunds/:id", paymentHandler.GetRefund)
				authPayment.GET("/orders/:order_id/refunds", paymentHandler.GetOrderRefunds)
			}
		}

		// 物流路由
		shippingGroup := v1.Group("/shipping")
		{
			// 公开端点
			shippingGroup.GET("/templates", shippingHandler.ListTemplates)
			shippingGroup.GET("/templates/:id", shippingHandler.GetTemplate)
			shippingGroup.POST("/calculate", shippingHandler.CalculateShipping)
			shippingGroup.GET("/carriers", shippingHandler.ListCarriers)
			shippingGroup.GET("/carriers/:id", shippingHandler.GetCarrier)
			shippingGroup.GET("/zones", shippingHandler.ListZones)
			shippingGroup.GET("/zones/:id", shippingHandler.GetZone)
			shippingGroup.GET("/track/:tracking_number", shippingHandler.TrackShipment)
			shippingGroup.GET("/orders/:order_id/tracking", shippingHandler.GetOrderTracking)
		}

		// 图片库路由
		galleryGroup := v1.Group("/galleries")
		{
			// 公开端点
			galleryGroup.GET("", galleryHandler.GetGalleries)
			galleryGroup.GET("/:id", galleryHandler.GetGalleryByID)
			galleryGroup.GET("/:id/images", galleryHandler.GetGalleryImages)
			galleryGroup.GET("/images/search", galleryHandler.SearchImages)
			galleryGroup.GET("/images/tags", galleryHandler.GetImagesByTags)
		}

		// 产品注册路由
		registrationGroup := v1.Group("/registrations")
		{
			// 公开端点
			registrationGroup.POST("/verify", registrationHandler.VerifySerialNumber)

			// 需要认证的端点
			authRegistration := registrationGroup.Group("")
			authRegistration.Use(middleware.AuthMiddleware(authService))
			{
				authRegistration.POST("", registrationHandler.CreateRegistration)
				authRegistration.GET("", registrationHandler.ListUserRegistrations)
				authRegistration.GET("/:id", registrationHandler.GetRegistration)
				authRegistration.PUT("/:id", registrationHandler.UpdateRegistration)
				authRegistration.POST("/warranty-claims", registrationHandler.CreateWarrantyClaim)
				authRegistration.GET("/warranty-claims/:id", registrationHandler.GetWarrantyClaim)
				authRegistration.GET("/:registration_id/warranty-claims", registrationHandler.ListRegistrationClaims)
			}
		}

		// 管理员路由
		adminGroup := v1.Group("/admin")
		adminGroup.Use(middleware.AuthMiddleware(authService))
		// TODO: 添加管理员权限验证中间件
		{
			// 订单管理
			adminGroup.GET("/orders", orderHandler.ListAllOrders)

			// 营销管理
			adminGroup.GET("/marketing/coupons/all", marketingHandler.GetAllCoupons)
			adminGroup.POST("/marketing/coupons", marketingHandler.CreateCoupon)
			adminGroup.PUT("/marketing/coupons/:id", marketingHandler.UpdateCoupon)
			adminGroup.DELETE("/marketing/coupons/:id", marketingHandler.DeleteCoupon)

			// 会员与积分管理
			adminGroup.GET("/loyalty/levels", marketingHandler.ListMemberLevels)
			adminGroup.POST("/loyalty/levels", marketingHandler.CreateMemberLevel)
			adminGroup.PUT("/loyalty/levels/:id", marketingHandler.UpdateMemberLevel)
			adminGroup.POST("/loyalty/users/:id/adjust", marketingHandler.AdminAdjustPoints)

			// 评价管理
			adminGroup.GET("/reviews/pending", reviewHandler.GetPendingReviews)
			adminGroup.POST("/reviews/:id/approve", reviewHandler.ApproveReview)
			adminGroup.POST("/reviews/:id/reject", reviewHandler.RejectReview)
			adminGroup.PUT("/reviews/:id/featured", reviewHandler.SetFeatured)

			// 工单管理
			adminGroup.GET("/tickets", ticketHandler.ListAllTickets)
			adminGroup.POST("/tickets/:id/assign", ticketHandler.AssignTicket)
			adminGroup.GET("/tickets/dashboard", ticketHandler.GetDashboard)
			adminGroup.GET("/tickets/recent", ticketHandler.GetRecentTickets)

			// 支付管理
			adminGroup.POST("/payment/methods", paymentHandler.CreatePaymentMethod)
			adminGroup.PUT("/payment/methods/:id", paymentHandler.UpdatePaymentMethod)
			adminGroup.DELETE("/payment/methods/:id", paymentHandler.DeletePaymentMethod)
			adminGroup.POST("/payment/tax-rates", paymentHandler.CreateTaxRate)
			adminGroup.PUT("/payment/tax-rates/:id", paymentHandler.UpdateTaxRate)
			adminGroup.DELETE("/payment/tax-rates/:id", paymentHandler.DeleteTaxRate)
			adminGroup.PUT("/payment/refunds/:id/status", paymentHandler.UpdateRefundStatus)

			// 物流管理
			adminGroup.POST("/shipping/templates", shippingHandler.CreateTemplate)
			adminGroup.PUT("/shipping/templates/:id", shippingHandler.UpdateTemplate)
			adminGroup.DELETE("/shipping/templates/:id", shippingHandler.DeleteTemplate)
			adminGroup.POST("/shipping/carriers", shippingHandler.CreateCarrier)
			adminGroup.PUT("/shipping/carriers/:id", shippingHandler.UpdateCarrier)
			adminGroup.DELETE("/shipping/carriers/:id", shippingHandler.DeleteCarrier)
			adminGroup.POST("/shipping/tracking", shippingHandler.CreateTrackingEvent)
			adminGroup.POST("/shipping/zones", shippingHandler.CreateZone)
			adminGroup.PUT("/shipping/zones/:id", shippingHandler.UpdateZone)
			adminGroup.DELETE("/shipping/zones/:id", shippingHandler.DeleteZone)

			// 图片库管理
			adminGroup.POST("/galleries", galleryHandler.CreateGallery)
			adminGroup.PUT("/galleries/:id", galleryHandler.UpdateGallery)
			adminGroup.DELETE("/galleries/:id", galleryHandler.DeleteGallery)
			adminGroup.POST("/galleries/:id/images", galleryHandler.CreateGalleryImage)
			adminGroup.POST("/galleries/:id/images/batch", galleryHandler.BatchCreateImages)
			adminGroup.PUT("/galleries/images/:id", galleryHandler.UpdateGalleryImage)
			adminGroup.DELETE("/galleries/images/:id", galleryHandler.DeleteGalleryImage)
			adminGroup.DELETE("/galleries/images/batch", galleryHandler.BatchDeleteImages)
			adminGroup.POST("/galleries/images/batch-order", galleryHandler.BatchUpdateOrder)

			// 产品注册管理
			adminGroup.GET("/registrations", registrationHandler.ListAllRegistrations)
			adminGroup.PUT("/registrations/:id/status", registrationHandler.UpdateRegistrationStatus)
			adminGroup.GET("/registrations/expiring", registrationHandler.GetExpiringWarranties)
			adminGroup.GET("/registrations/stats", registrationHandler.GetRegistrationStats)
			adminGroup.GET("/registrations/warranty-claims", registrationHandler.ListAllWarrantyClaims)
			adminGroup.PUT("/registrations/warranty-claims/:id/status", registrationHandler.UpdateWarrantyClaimStatus)

			// 订阅管理
			adminGroup.GET("/subscriptions", subscriptionHandler.GetAllSubscriptions)
			adminGroup.GET("/subscriptions/tags", subscriptionHandler.GetSubscriptionsByTags)
			adminGroup.GET("/subscriptions/stats", subscriptionHandler.GetStats)
			adminGroup.DELETE("/subscriptions/:email", subscriptionHandler.DeleteSubscription)
			adminGroup.GET("/subscriptions/export", subscriptionHandler.ExportEmails)

			// 审计日志管理
			adminGroup.GET("/audit/logs", auditHandler.ListAllAuditLogs)
			adminGroup.GET("/audit/logs/:id", auditHandler.GetAuditLog)
			adminGroup.GET("/audit/users/:user_id/logs", auditHandler.ListUserAuditLogs)
			adminGroup.GET("/audit/entities/logs", auditHandler.ListEntityAuditLogs)
			adminGroup.GET("/audit/logs/date-range", auditHandler.ListAuditLogsByDateRange)
			adminGroup.GET("/audit/ip/:ip_address/logs", auditHandler.ListAuditLogsByIP)
			adminGroup.GET("/audit/logs/search", auditHandler.SearchAuditLogs)
			adminGroup.GET("/audit/stats", auditHandler.GetAuditStats)
			adminGroup.GET("/audit/activities/recent", auditHandler.GetRecentActivities)
			adminGroup.POST("/audit/logs/cleanup", auditHandler.DeleteOldAuditLogs)
		}
	}

	// Sitemap 路由（根路径）
	r.GET("/sitemap.xml", i18nHandler.GetSitemapIndex)
	r.GET("/sitemap-hreflang.xml", i18nHandler.GetHreflangSitemap)
	r.GET("/sitemap-:locale.xml", i18nHandler.GetLocaleSitemap)

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": "1.4.0",
		})
	})
}
