package v1

import (
	"tanzanite/internal/api/middleware"
	analyticsapi "tanzanite/internal/api/v1/analytics"
	"tanzanite/internal/api/v1/auth"
	"tanzanite/internal/api/v1/behavior"
	"tanzanite/internal/api/v1/cart"
	"tanzanite/internal/api/v1/checkout"
	"tanzanite/internal/api/v1/content"
	currencyapi "tanzanite/internal/api/v1/currency"
	"tanzanite/internal/api/v1/faq"
	"tanzanite/internal/api/v1/feedback"
	"tanzanite/internal/api/v1/gallery"
	"tanzanite/internal/api/v1/i18n"
	"tanzanite/internal/api/v1/marketing"
	"tanzanite/internal/api/v1/order"
	"tanzanite/internal/api/v1/payment"
	"tanzanite/internal/api/v1/product"
	quickbuyapi "tanzanite/internal/api/v1/quickbuy"
	"tanzanite/internal/api/v1/recommendation"
	"tanzanite/internal/api/v1/registration"
	"tanzanite/internal/api/v1/review"
	seohomeapi "tanzanite/internal/api/v1/seo/home"
	"tanzanite/internal/api/v1/settings"
	"tanzanite/internal/api/v1/shipping"
	"tanzanite/internal/api/v1/showcase"
	"tanzanite/internal/api/v1/spoke"
	"tanzanite/internal/api/v1/storefront"
	"tanzanite/internal/api/v1/subscription"
	"tanzanite/internal/api/v1/suggestionfeedback"
	"tanzanite/internal/api/v1/ticket"
	"tanzanite/internal/api/v1/wishlist"
	"tanzanite/internal/app"
	attributionpkg "tanzanite/internal/pkg/attribution"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/securecookie"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RegisterRoutes 注册所有v1路由
func RegisterRoutes(r *gin.Engine, deps *app.Dependencies, cfg *config.Config) {
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.Use(middleware.TraceMiddleware())
	// 初始化repositories
	services := deps.Services
	authService := services.Auth
	postService := services.Post
	productService := services.Product
	cartService := services.Cart
	settingService := services.Setting
	faqService := services.FAQ
	galleryService := services.Gallery
	registrationService := services.Registration
	checkoutService := services.Checkout
	orderService := services.Order
	paymentService := services.Payment
	marketingService := services.Marketing
	reviewService := services.Review
	ticketService := services.Ticket
	subscriptionService := services.Subscription
	sitemapService := services.Sitemap
	storageSvc := deps.Storage
	showcaseService := services.Showcase
	wishlistService := services.Wishlist
	feedbackService := services.Feedback
	suggestionFeedbackService := services.SuggestionFeedback

	// 初始化handlers
	cookieOptions := securecookie.Options{
		Secure:   cfg.Cookie.SecureEnabled(cfg.Server),
		SameSite: cfg.Cookie.SameSiteMode(),
		Domain:   cfg.Cookie.Domain,
	}
	authHandler := auth.NewHandler(authService, cookieOptions)
	browsingHistoryHandler := auth.NewBrowsingHistoryHandler(services.User)
	contentHandler := content.NewHandler(postService, faqService)
	faqHandler := faq.NewHandler(faqService)
	productHandler := product.NewHandler(productService)
	productHandler.ConfigureStorefrontContext(services.StorefrontContext)
	quickBuyHandler := quickbuyapi.NewHandler(services.QuickBuy)
	cartHandler := cart.NewHandler(cartService, cart.Options{
		VisitorProfileService: services.VisitorProfile,
		VisitorSecret:         cfg.JWT.Secret,
	})
	settingsHandler := settings.NewHandler(settingService)
	seoHomeHandler := seohomeapi.NewHandler(services.SEO)
	analyticsHandler := analyticsapi.NewHandler(services.Analytics)
	storefrontContextHandler := storefront.NewContextHandler(services.StorefrontContext)
	currencyHandler := currencyapi.NewHandler(services.CurrencyPolicy, services.ExchangeRate)
	orderHandler := order.NewHandler(orderService, cartService, deps.AntiFraud)
	orderHandler.ConfigureOrderAbuse(deps.OrderAbuse)
	orderHandler.ConfigurePaymentProtection(services.PaymentProtection)
	checkoutHandler := checkout.NewHandler(checkoutService, cartService)
	marketingHandler := marketing.NewHandler(marketingService, settingService, services.LoyaltyProgram)
	reviewHandler := review.NewHandler(reviewService)
	ticketHandler := ticket.NewHandler(ticketService, ticket.Options{
		MediaService:          services.Media,
		AllowedOrigins:        cfg.CORS.AllowedOrigins,
		VisitorSecret:         cfg.JWT.Secret,
		VisitorProfileService: services.VisitorProfile,
		CustomerServiceEvents: services.CustomerServiceEvents,
	})
	paymentHandler := payment.NewHandler(
		paymentService,
		orderService,
		services.AdminSettings,
		services.PaymentThreeDS,
		services.PaymentRiskMonitoring,
		services.PaymentProtection,
		services.PaymentRefundReview,
		deps.AntiBot,
		deps.AntiFraud,
		services.StorefrontContext,
	)
	paymentHandler.ConfigurePublicBaseURL(cfg.Server.BaseURL)
	shippingHandler := shipping.NewHandler(services.Shipping, orderService)
	galleryHandler := gallery.NewGalleryHandler(galleryService)
	registrationHandler := registration.NewHandler(registrationService, storageSvc, deps.AntiBot)
	subscriptionHandler := subscription.NewHandler(subscriptionService, deps.AntiBot)
	i18nHandler := i18n.NewHandler(postService, sitemapService)
	showcaseHandler := showcase.NewShowcaseHandler(showcaseService)
	wishlistHandler := wishlist.NewHandler(wishlistService)
	feedbackHandler := feedback.NewHandler(feedbackService)
	suggestionFeedbackHandler := suggestionfeedback.NewHandler(suggestionFeedbackService, storageSvc)
	spokeHandler := spoke.NewHandler(services.Spoke)
	behaviorEventHandler := behavior.NewHandler(services.BehaviorEvents)
	recommendationHandler := recommendation.NewHandler(services.Recommendations)
	attributionSigner, attributionErr := attributionpkg.NewSigner(cfg.JWT.Secret)
	if attributionErr == nil {
		behaviorEventHandler.ConfigureAttribution(attributionSigner, cookieOptions)
		orderHandler.ConfigureAttribution(attributionSigner)
	}

	// 公网 Webhook 回调入口不挂 CSRF。
	// 第三方平台（支付网关、17TRACK 等）不会携带浏览器 CSRF token，安全边界由各自 handler 内的签名验签负责。
	webhookV1 := r.Group("/api/v1")
	{
		paymentWebhookGroup := webhookV1.Group("/payment")
		{
			paymentWebhookGroup.POST("/webhook/:provider", paymentHandler.HandleWebhook)
		}

		shippingWebhookGroup := webhookV1.Group("/shipping")
		{
			shippingWebhookGroup.POST("/webhook/:provider", shippingHandler.HandleTrackingWebhook)
		}
	}

	// API v1 路由组
	v1 := r.Group("/api/v1")
	v1.Use(middleware.CSRFProtection(cfg.CORS.AllowedOrigins))
	v1.Use(middleware.I18n())
	v1.Use(middleware.VisitorRiskTelemetry(services.VisitorRisk))
	{
		storefrontGroup := v1.Group("/storefront")
		{
			storefrontGroup.GET("/context", storefrontContextHandler.GetContext)
		}

		// 认证路由（公开）
		authGroup := v1.Group("/auth")
		authGroup.Use(middleware.RateLimit(10)) // 10 RPS for auth endpoints
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/google-login", authHandler.GoogleLogin)
			authGroup.POST("/logout", authHandler.Logout)
			authGroup.GET("/profile", middleware.OptionalAuthMiddleware(authService), authHandler.GetProfile)
		}

		// 内容路由（公开）
		contentGroup := v1.Group("/content")
		{
			contentGroup.GET("/posts", contentHandler.ListPosts)
			contentGroup.GET("/posts/:id", contentHandler.GetPost)
			contentGroup.GET("/faq-pages", contentHandler.ListFAQPages)
			contentGroup.GET("/faq-pages/by-route", contentHandler.GetFAQPageByRoute)
			contentGroup.GET("/faq-pages/:page_id", contentHandler.GetFAQPage)
			contentGroup.GET("/faqs", contentHandler.ListFAQs)
			contentGroup.GET("/faqs/:id", contentHandler.GetFAQ)
			contentGroup.GET("/faq-categories", contentHandler.GetFAQCategories)
			contentGroup.GET("/faqs/search", contentHandler.SearchFAQs)
			contentGroup.GET("/faqs/category/:category", contentHandler.GetFAQsByCategory)
			contentGroup.GET("/faqs/popular", contentHandler.GetPopularFAQs)
			contentGroup.POST("/faqs/:id/view", faqHandler.IncrementFAQView)
		}

		// 推荐行为事件路由（公开，可选认证）
		behaviorEventsGroup := v1.Group("/behavior-events")
		behaviorEventsGroup.Use(
			middleware.OptionalAuthMiddleware(authService),
			middleware.RateLimit(20),
		)
		{
			behaviorEventsGroup.POST("/batch", behaviorEventHandler.IngestBatch)
		}

		// 推荐读取路由（公开，可选认证）
		recommendationsGroup := v1.Group("/recommendations")
		recommendationsGroup.Use(
			middleware.OptionalAuthMiddleware(authService),
			middleware.RateLimit(30),
		)
		{
			recommendationsGroup.POST("", recommendationHandler.GetRecommendations)
		}

		// 产品路由（公开）
		commercialInventoryProbeGuard := middleware.CommercialInventoryProbeGuard(deps.RedisClient)
		productGroup := v1.Group("/products")
		productGroup.Use(commercialInventoryProbeGuard)
		{
			productGroup.GET("", productHandler.ListProducts)
			productGroup.GET("/types", productHandler.ListProductTypes)
			productGroup.GET("/attributes/filterable", productHandler.GetFilterableAttributes)
			productGroup.GET("/:id", productHandler.GetProduct)
		}

		quickBuyGroup := v1.Group("/quick-buy")
		{
			quickBuyGroup.GET("/flows/current", quickBuyHandler.GetCurrentFlow)
			quickBuyGroup.POST("/sessions", middleware.OptionalAuthMiddleware(authService), middleware.RateLimit(10), quickBuyHandler.CreateSession)
			quickBuyGroup.GET("/sessions/:token", middleware.OptionalAuthMiddleware(authService), middleware.RateLimit(20), quickBuyHandler.GetSession)
			quickBuyGroup.GET("/sessions/:token/steps/:step_key/candidates", middleware.OptionalAuthMiddleware(authService), middleware.RateLimit(20), quickBuyHandler.ListStepCandidates)
			quickBuyGroup.PATCH("/sessions/:token/selections", middleware.OptionalAuthMiddleware(authService), middleware.RateLimit(20), quickBuyHandler.UpdateSelections)
			quickBuyGroup.POST("/sessions/:token/validate", middleware.OptionalAuthMiddleware(authService), middleware.RateLimit(20), quickBuyHandler.ValidateSession)
		}

		// 购物车路由（可选认证）
		cartGroup := v1.Group("/cart")
		cartGroup.Use(
			middleware.OptionalAuthMiddleware(authService),
			middleware.CommercialCartProbeGuard(deps.RedisClient),
		)
		{
			cartGroup.GET("/summary", cartHandler.GetCartSummary)
			cartGroup.POST("/add", cartHandler.AddToCart)
			cartGroup.PUT("/items/:id", cartHandler.UpdateCartItem)
			cartGroup.DELETE("/items/:id", cartHandler.RemoveFromCart)
			cartGroup.POST("/sync", cartHandler.SyncCart)
			cartGroup.POST("/clear", cartHandler.ClearCart)
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
			feedbackGroup.POST("", middleware.AuthMiddleware(authService), middleware.RateLimitByUser(2), feedbackHandler.Create)
		}

		suggestionFeedbackGroup := v1.Group("/suggestion-feedback")
		{
			suggestionFeedbackGroup.GET("/eligibility", middleware.OptionalAuthMiddleware(authService), suggestionFeedbackHandler.Eligibility)
			suggestionFeedbackGroup.POST("/upload", middleware.AuthMiddleware(authService), middleware.RateLimitByUserPerMinute(6, 2), suggestionFeedbackHandler.Upload)
			suggestionFeedbackGroup.POST("", middleware.AuthMiddleware(authService), middleware.RateLimitByUser(2), suggestionFeedbackHandler.Create)
		}

		spokeGroup := v1.Group("/spoke")
		spokeGroup.Use(middleware.RateLimit(5))
		{
			spokeGroup.POST("/calc", spokeHandler.Calculate)
			spokeGroup.GET("/export", spokeHandler.GetExport)
			spokeGroup.GET("/history", middleware.AuthMiddleware(authService), spokeHandler.ListHistory)
		}

		checkoutGroup := v1.Group("/checkout")
		checkoutGroup.Use(middleware.AuthMiddleware(authService))
		{
			checkoutGroup.POST("/quote", checkoutHandler.Quote)
		}

		// 订单路由（需要认证）
		orderGroup := v1.Group("/orders")
		orderGroup.Use(
			middleware.AuthMiddleware(authService),
			middleware.CommercialOrderEnumerationGuard(deps.RedisClient),
		)
		{
			orderGroup.POST("", middleware.RateLimitByUser(2), orderHandler.CreateOrder)
			orderGroup.GET("", orderHandler.ListOrders)
			orderGroup.GET("/stats", orderHandler.GetOrderStats)
			orderGroup.GET("/:order_number", orderHandler.GetOrder)
			orderGroup.POST("/:order_number/cancel", orderHandler.CancelOrder)
		}

		// 营销路由
		marketingGroup := v1.Group("/marketing")
		{
			// 优惠券（公开）
			marketingGroup.GET("/coupons", marketingHandler.ListCoupons)

			// 等级配置（公开）
			marketingGroup.GET("/loyalty/levels", marketingHandler.ListMemberLevels)
			marketingGroup.GET("/loyalty/config", marketingHandler.GetLoyaltyProgramConfig)
			marketingGroup.GET("/loyalty/redeem-options", marketingHandler.ListRedeemGiftCardOptions)
			marketingGroup.GET("/loyalty/rules", marketingHandler.GetLoyaltyRules)

			// 需要认证的营销功能
			authMarketing := marketingGroup.Group("")
			authMarketing.Use(middleware.AuthMiddleware(authService))
			{
				// 优惠券
				authMarketing.POST("/coupons/validate", middleware.RateLimitByUser(3), marketingHandler.ValidateCoupon)

				// 积分和会员
				authMarketing.GET("/loyalty/assets", marketingHandler.GetUserAssets)
				authMarketing.GET("/loyalty/gift-cards", marketingHandler.ListUserGiftCards)
				authMarketing.GET("/loyalty/points", marketingHandler.GetPoints)
				authMarketing.GET("/loyalty/info", marketingHandler.GetLoyaltyInfo)
				authMarketing.POST("/loyalty/checkin", middleware.RateLimitByUser(1), marketingHandler.CheckIn)
				authMarketing.POST("/loyalty/referral", middleware.RateLimitByUser(2), marketingHandler.CreateReferral)
				authMarketing.POST("/loyalty/redeem", middleware.RateLimitByUser(1), marketingHandler.RedeemPointsToGiftCard)
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
				authReview.POST("", middleware.RateLimitByUser(2), reviewHandler.CreateReview)
				authReview.GET("/my", reviewHandler.ListUserReviews)
				authReview.DELETE("/:id", reviewHandler.DeleteReview)
				authReview.POST("/:id/helpful", middleware.RateLimitByUser(3), reviewHandler.MarkHelpful)
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

		customerServiceGroup := v1.Group("/customer-service")
		{
			customerServiceGroup.GET("/agents", ticketHandler.ListPublicCustomerServiceAgents)
			customerServiceGroup.GET("/products", commercialInventoryProbeGuard, productHandler.ListPublicChatProducts)
			customerServiceGroup.GET("/orders", middleware.AuthMiddleware(authService), orderHandler.ListPublicChatOrders)
			customerServiceGroup.POST("/conversations", middleware.OptionalAuthMiddleware(authService), middleware.RateLimitByUser(3), ticketHandler.EnsurePublicCustomerServiceConversation)
			customerServiceGroup.GET("/has-conversation", middleware.OptionalAuthMiddleware(authService), ticketHandler.HasPublicCustomerServiceConversation)
			customerServiceGroup.POST("/attachments", middleware.OptionalAuthMiddleware(authService), middleware.RateLimitByUserPerMinute(6, 2), ticketHandler.UploadPublicCustomerServiceAttachment)
			customerServiceGroup.POST("/messages", middleware.OptionalAuthMiddleware(authService), middleware.RateLimitByUser(5), ticketHandler.SendPublicCustomerServiceMessage)
			customerServiceGroup.POST("/typing", middleware.OptionalAuthMiddleware(authService), middleware.RateLimitByUser(5), ticketHandler.SendPublicCustomerServiceTyping)
			customerServiceGroup.GET("/messages/:conversation_id", middleware.OptionalAuthMiddleware(authService), ticketHandler.GetPublicCustomerServiceMessages)
			customerServiceGroup.GET("/auto-reply/welcome", middleware.OptionalAuthMiddleware(authService), ticketHandler.GetWelcomeMessage)
			customerServiceGroup.POST("/auto-reply/match", middleware.OptionalAuthMiddleware(authService), middleware.RateLimitByUser(5), ticketHandler.MatchKeywordMessage)
			customerServiceGroup.GET("/events", middleware.OptionalAuthMiddleware(authService), ticketHandler.StreamPublicCustomerServiceEvents)
			customerServiceGroup.GET("/ws", middleware.OptionalAuthMiddleware(authService), ticketHandler.ServeWS)
		}

		// 用户浏览历史路由（需要认证）
		userGroup := v1.Group("/user")
		userGroup.Use(middleware.AuthMiddleware(authService))
		{
			userGroup.POST("/browsing-history", browsingHistoryHandler.AddBrowsingHistory)
			userGroup.GET("/browsing-history", browsingHistoryHandler.GetBrowsingHistory)
			userGroup.DELETE("/browsing-history/:product_id", browsingHistoryHandler.DeleteBrowsingHistory)
			userGroup.DELETE("/browsing-history", browsingHistoryHandler.ClearBrowsingHistory)
		}

		// Showcase (Picture Warehouse)
		showcaseGroup := v1.Group("/showcase")
		{
			showcaseGroup.GET("/gallery", showcaseHandler.List)
			showcaseGroup.GET("/comments", showcaseHandler.ListComments)
			showcaseGroup.POST("/upload", middleware.AuthMiddleware(authService), middleware.RateLimitByUserPerMinute(3, 1), showcaseHandler.Upload)
			showcaseGroup.POST("/comments", middleware.AuthMiddleware(authService), middleware.RateLimitByUser(2), showcaseHandler.AddComment)
		}

		seoGroup := v1.Group("/seo")
		{
			seoGroup.GET("/home", seoHomeHandler.Get)
		}

		analyticsGroup := v1.Group("/analytics")
		{
			analyticsGroup.GET("", analyticsHandler.Get)
		}

		// 设置路由
		settingsGroup := v1.Group("/settings")
		{
			settingsGroup.GET("/currency-policy", currencyHandler.GetPolicy)
			// 公开设置
			settingsGroup.GET("/site", settingsHandler.GetSiteSettings)
			settingsGroup.GET("/social", settingsHandler.GetSocialSettings)
			settingsGroup.GET("/public", settingsHandler.GetAllPublicSettings)
			settingsGroup.GET("/groups", settingsHandler.GetGroups)
			settingsGroup.GET("/group/:group", settingsHandler.GetSettingsByGroup)
			settingsGroup.GET("/:key", settingsHandler.GetSetting)
		}

		currencyGroup := v1.Group("/currency")
		{
			currencyGroup.GET("/exchange-rates", currencyHandler.ListExchangeRates)
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
			subscriptionGroup.POST("", middleware.RateLimit(2), subscriptionHandler.Subscribe)
			subscriptionGroup.GET("/confirm/:token", middleware.RateLimit(5), subscriptionHandler.ConfirmSubscription)
			subscriptionGroup.GET("/unsubscribe/:token", middleware.RateLimit(5), subscriptionHandler.Unsubscribe)
			subscriptionGroup.POST("/unsubscribe", middleware.RateLimit(2), subscriptionHandler.UnsubscribeByEmail)
			subscriptionGroup.POST("/resubscribe", middleware.RateLimit(2), subscriptionHandler.Resubscribe)
			subscriptionGroup.GET("/resubscribe/:token", middleware.RateLimit(5), subscriptionHandler.ResubscribeByToken)
			subscriptionGroup.GET("/status/:email", middleware.RateLimit(2), subscriptionHandler.GetSubscription)
			subscriptionGroup.GET("/status-token/:token", middleware.RateLimit(5), subscriptionHandler.GetSubscriptionByToken)
		}

		// 支付路由
		paymentGroup := v1.Group("/payment")
		{
			// 公开端点
			paymentGroup.GET("/methods", paymentHandler.ListPaymentMethods)
			paymentGroup.GET("/methods/:id", paymentHandler.GetPaymentMethod)
			paymentGroup.GET("/tax-rates", paymentHandler.ListTaxRates)
			paymentGroup.GET("/tax-rates/:id", paymentHandler.GetTaxRate)
			paymentGroup.POST("/calculate-tax", middleware.RateLimit(5), paymentHandler.CalculateTax)

			// 需要认证的端点
			authPayment := paymentGroup.Group("")
			authPayment.Use(middleware.AuthMiddleware(authService))
			{
				authPayment.GET("/transactions/:id", paymentHandler.GetTransaction)
				authPayment.GET("/orders/:order_id/transactions", paymentHandler.GetOrderTransactions)
				authPayment.GET("/refunds/:id", paymentHandler.GetRefund)
				authPayment.GET("/orders/:order_id/refunds", paymentHandler.GetOrderRefunds)
				authPayment.POST("/stripe/payment-intents", middleware.RateLimitByUser(3), paymentHandler.CreateStripePaymentIntent)
				authPayment.POST("/paypal/orders", middleware.RateLimitByUser(3), paymentHandler.CreatePayPalOrder)
				authPayment.POST("/paypal/orders/:paypal_order_id/capture", middleware.RateLimitByUser(5), paymentHandler.CapturePayPalOrder)
				authPayment.POST("/alipay/orders", middleware.RateLimitByUser(3), paymentHandler.CreateAlipayOrder)
				authPayment.POST("/alipay/orders/:order_number/confirm", middleware.RateLimitByUser(5), paymentHandler.ConfirmAlipayOrder)
				authPayment.POST("/wechat/orders", middleware.RateLimitByUser(3), paymentHandler.CreateWechatOrder)
				authPayment.POST("/wechat/orders/:order_number/confirm", middleware.RateLimitByUser(5), paymentHandler.ConfirmWechatOrder)
			}
		}

		// 物流路由
		shippingGroup := v1.Group("/shipping")
		{
			// 公开端点
			shippingGroup.GET("/templates", shippingHandler.ListTemplates)
			shippingGroup.GET("/templates/:id", shippingHandler.GetTemplate)
			shippingGroup.POST("/calculate", middleware.RateLimit(5), shippingHandler.CalculateShipping)
			shippingGroup.POST("/quote", middleware.RateLimit(5), shippingHandler.QuoteShipping)
			shippingGroup.GET("/carriers", shippingHandler.ListCarriers)
			shippingGroup.GET("/carriers/:id", shippingHandler.GetCarrier)
			shippingGroup.GET("/carrier-services", shippingHandler.ListCarrierServices)
			shippingGroup.GET("/carrier-services/:id", shippingHandler.GetCarrierService)
			shippingGroup.GET("/zones", shippingHandler.ListZones)
			shippingGroup.GET("/zones/:id", shippingHandler.GetZone)
			shippingGroup.GET("/track/:tracking_number", middleware.AuthMiddleware(authService), middleware.RateLimit(5), shippingHandler.TrackShipment)
			shippingGroup.GET("/orders/:order_id/tracking", middleware.AuthMiddleware(authService), shippingHandler.GetOrderTracking)
			shippingGroup.GET("/packaging-rules", shippingHandler.ListPackagingRules)
			shippingGroup.GET("/packaging-rules/:id", shippingHandler.GetPackagingRule)
			shippingGroup.GET("/products/:id/packaging-rules", shippingHandler.GetProductPackagingRules)
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
			registrationGroup.POST("/verify", middleware.RateLimit(2), registrationHandler.VerifySerialNumber)
			registrationGroup.POST("/warranty/verify-order", middleware.RateLimit(2), registrationHandler.VerifyWarrantyOrder)
			registrationGroup.GET("/warranty/verify/:token", middleware.RateLimit(5), registrationHandler.VerifyWarrantyOrderToken)
			registrationGroup.POST("/warranty/claim", middleware.RateLimit(1), registrationHandler.SubmitWarrantyClaim)

			// 需要认证的端点
			authRegistration := registrationGroup.Group("")
			authRegistration.Use(middleware.AuthMiddleware(authService))
			{
				authRegistration.GET("/warranty/:code", registrationHandler.GetWarrantyStatus)
				authRegistration.POST("", registrationHandler.CreateRegistration)
				authRegistration.GET("", registrationHandler.ListUserRegistrations)
				authRegistration.GET("/:id", registrationHandler.GetRegistration)
				authRegistration.PUT("/:id", registrationHandler.UpdateRegistration)
				authRegistration.POST("/warranty-claims", registrationHandler.CreateWarrantyClaim)
				authRegistration.GET("/warranty-claims/:id", registrationHandler.GetWarrantyClaim)
				authRegistration.GET("/:id/warranty-claims", registrationHandler.ListRegistrationClaims)
			}
		}

	}

	// Sitemap 路由（根路径）
	r.GET("/sitemap.xml", i18nHandler.GetSitemapIndex)
	r.GET("/sitemap-hreflang.xml", i18nHandler.GetHreflangSitemap)
	r.GET("/sitemap-:locale.xml", i18nHandler.GetLocaleSitemap)

	// 健康检查
}
