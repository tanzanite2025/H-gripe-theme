package v1

import (
	"commerce-platform/internal/api/middleware"
	analyticsapi "commerce-platform/internal/api/v1/analytics"
	"commerce-platform/internal/api/v1/auth"
	"commerce-platform/internal/api/v1/behavior"
	"commerce-platform/internal/api/v1/cart"
	"commerce-platform/internal/api/v1/checkout"
	"commerce-platform/internal/api/v1/content"
	currencyapi "commerce-platform/internal/api/v1/currency"
	"commerce-platform/internal/api/v1/faq"
	"commerce-platform/internal/api/v1/feedback"
	fitmentcatalogapi "commerce-platform/internal/api/v1/fitmentcatalog"
	"commerce-platform/internal/api/v1/gallery"
	"commerce-platform/internal/api/v1/i18n"
	"commerce-platform/internal/api/v1/marketing"
	"commerce-platform/internal/api/v1/order"
	"commerce-platform/internal/api/v1/payment"
	"commerce-platform/internal/api/v1/product"
	quickbuyapi "commerce-platform/internal/api/v1/quickbuy"
	"commerce-platform/internal/api/v1/recommendation"
	"commerce-platform/internal/api/v1/review"
	selectionassistantapi "commerce-platform/internal/api/v1/selectionassistant"
	seohomeapi "commerce-platform/internal/api/v1/seo/home"
	"commerce-platform/internal/api/v1/settings"
	"commerce-platform/internal/api/v1/shipping"
	"commerce-platform/internal/api/v1/showcase"
	"commerce-platform/internal/api/v1/spoke"
	"commerce-platform/internal/api/v1/storefront"
	"commerce-platform/internal/api/v1/subscription"
	"commerce-platform/internal/api/v1/suggestionfeedback"
	"commerce-platform/internal/api/v1/ticket"
	visualshowcaseapi "commerce-platform/internal/api/v1/visualshowcase"
	"commerce-platform/internal/api/v1/warranty"
	wheelsetfitapi "commerce-platform/internal/api/v1/wheelsetfit"
	"commerce-platform/internal/api/v1/wishlist"
	"commerce-platform/internal/app"
	attributionpkg "commerce-platform/internal/pkg/attribution"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/securecookie"

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
	warrantyService := services.Warranty
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
	contentHandler := content.NewHandler(postService, faqService, services.Media)
	faqHandler := faq.NewHandler(faqService)
	productHandler := product.NewHandler(productService)
	productHandler.ConfigureProductCategoryService(services.ProductCategory)
	productHandler.ConfigureMediaService(services.Media)
	productHandler.ConfigureStorefrontContext(services.StorefrontContext)
	productHandler.ConfigureReviewService(reviewService)
	productHandler.ConfigureShippingService(services.Shipping)
	quickBuyHandler := quickbuyapi.NewHandler(services.QuickBuy, services.Media)
	selectionAssistantHandler := selectionassistantapi.NewHandler(services.SelectionAssistant)
	wheelsetFitQuestionnaireHandler := wheelsetfitapi.NewHandler(services.WheelsetFitQuestionnaire)
	fitmentCatalogHandler := fitmentcatalogapi.NewHandler(
		services.FrameFitmentEntry,
		services.ForkFitmentEntry,
		services.FitmentHubSpecification,
	)
	cartHandler := cart.NewHandler(cartService, cart.Options{
		MediaService:          services.Media,
		VisitorProfileService: services.VisitorProfile,
		VisitorSecret:         cfg.JWT.Secret,
	})
	settingsHandler := settings.NewHandler(settingService, services.WebsiteProfile)
	settingsHandler.ConfigureMediaService(services.Media)
	settingsHandler.ConfigureSiteLogoService(services.SiteLogo)
	settingsHandler.ConfigureWebsiteNameService(services.WebsiteName)
	settingsHandler.ConfigureRefundReturnPolicyService(services.RefundReturnPolicy)
	seoHomeHandler := seohomeapi.NewHandler(services.SEO)
	analyticsHandler := analyticsapi.NewHandler(services.Analytics)
	storefrontContextHandler := storefront.NewContextHandler(services.StorefrontContext)
	storefrontRedirectsHandler := storefront.NewRedirectsHandler(services.StorefrontRedirectRules)
	storefrontSitemapHandler := storefront.NewSitemapHandler(services.StorefrontRouteCatalog)
	storefrontURLSearchHandler := storefront.NewURLSearchHandler(services.StorefrontURLSearchProfiles)
	currencyHandler := currencyapi.NewHandler(services.CurrencyPolicy, services.ExchangeRate)
	orderHandler := order.NewHandler(orderService, cartService, deps.AntiFraud)
	orderHandler.ConfigureOrderAbuse(deps.OrderAbuse)
	orderHandler.ConfigurePaymentProtection(services.PaymentProtection)
	orderHandler.ConfigureAfterSales(services.AfterSales, storageSvc)
	checkoutHandler := checkout.NewHandler(checkoutService, cartService)
	marketingHandler := marketing.NewHandler(marketingService, settingService, services.LoyaltyProgram)
	marketingHandler.ConfigureMediaService(services.Media)
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
	paymentHandler.ConfigureCardBINLimiter(deps.CardBINLimiter)
	paymentHandler.ConfigurePaymentGatewayCircuitBreaker(deps.PaymentGatewayCircuitBreaker)
	shippingHandler := shipping.NewHandler(services.Shipping, orderService)
	galleryHandler := gallery.NewGalleryHandler(galleryService, services.Media)
	warrantyHandler := warranty.NewHandler(warrantyService, storageSvc, deps.AntiBot)
	warrantyHandler.ConfigureMediaService(services.Media)
	warrantyHandler.ConfigureShipmentRecordService(services.ShipmentRecord)
	subscriptionHandler := subscription.NewHandler(subscriptionService, deps.AntiBot)
	i18nHandler := i18n.NewHandler(postService, sitemapService)
	showcaseHandler := showcase.NewShowcaseHandler(showcaseService)
	showcaseHandler.ConfigureUploadProtection(services.ShowcaseUploadProtection)
	showcaseHandler.ConfigureUploadEligibility(services.ShowcaseUploadEligibility)
	visualShowcaseHandler := visualshowcaseapi.NewHandler(services.VisualShowcase, services.Media)
	wishlistHandler := wishlist.NewHandler(wishlistService, services.Media)
	feedbackHandler := feedback.NewHandler(feedbackService)
	feedbackHandler.ConfigureSourceHashSecret(cfg.JWT.Secret)
	suggestionFeedbackHandler := suggestionfeedback.NewHandler(suggestionFeedbackService, storageSvc, services.Media)
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
		v1.GET("/upload-specs", GetUploadSpecs)

		storefrontGroup := v1.Group("/storefront")
		{
			storefrontGroup.GET("/context", storefrontContextHandler.GetContext)
			storefrontGroup.GET("/redirects", storefrontRedirectsHandler.ListPublished)
			storefrontGroup.GET("/sitemap-routes", storefrontSitemapHandler.List)
			storefrontGroup.GET("/url-search-index", storefrontURLSearchHandler.List)
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
			productGroup.GET("/specification-templates", productHandler.ListProductSpecificationTemplates)
			productGroup.GET("/categories", productHandler.ListCategories)
			productGroup.GET("/categories/:slug", productHandler.GetCategory)
			productGroup.GET("/attributes/filterable", productHandler.GetFilterableAttributes)
			productGroup.GET("/:id", productHandler.GetProduct)
		}

		quickBuyGroup := v1.Group("/quick-buy")
		quickBuyGroup.Use(middleware.QuickBuyRateLimit(deps.RedisClient, cfg.QuickBuyRateLimit))
		{
			quickBuyGroup.GET("/flows/current", quickBuyHandler.GetCurrentFlow)
			quickBuyGroup.POST("/sessions", middleware.OptionalAuthMiddleware(authService), quickBuyHandler.CreateSession)
			quickBuyGroup.GET("/sessions/:token", middleware.OptionalAuthMiddleware(authService), quickBuyHandler.GetSession)
			quickBuyGroup.GET("/sessions/:token/steps/:step_key/candidates", middleware.OptionalAuthMiddleware(authService), quickBuyHandler.ListSessionStepCandidates)
			quickBuyGroup.PATCH("/sessions/:token/selections", middleware.OptionalAuthMiddleware(authService), quickBuyHandler.UpdateSessionSelections)
			quickBuyGroup.POST("/sessions/:token/validate", middleware.OptionalAuthMiddleware(authService), quickBuyHandler.ValidateSession)
		}

		selectionAssistantGroup := v1.Group("/selection-assistant")
		{
			selectionAssistantGroup.GET("/flows/:slug", selectionAssistantHandler.GetPublishedFlow)
		}

		wheelsetFitQuestionnaireGroup := v1.Group("/wheelset-fit-questionnaire")
		{
			wheelsetFitQuestionnaireGroup.GET("/current", wheelsetFitQuestionnaireHandler.GetCurrentFlow)
		}

		fitmentCatalogGroup := v1.Group("/fitment-catalog")
		{
			fitmentCatalogHandler.RegisterRoutes(fitmentCatalogGroup)
		}

		visualShowcaseGroup := v1.Group("/visual-showcases")
		{
			visualShowcaseGroup.GET("/:showcase_key", visualShowcaseHandler.Get)
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
			feedbackGroup.GET("", middleware.FeedbackReadRateLimit(deps.RedisClient, cfg.FeedbackRateLimit), middleware.RateLimit(20), feedbackHandler.List)
			feedbackGroup.GET("/eligibility", middleware.OptionalAuthMiddleware(authService), feedbackHandler.Eligibility)
			feedbackGroup.POST(
				"",
				middleware.AuthMiddleware(authService),
				middleware.FeedbackWriteRateLimit(deps.RedisClient, cfg.FeedbackRateLimit),
				middleware.RateLimitByUserPerMinute(6, 2),
				feedbackHandler.Create,
			)
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
			orderGroup.POST("", middleware.Idempotency(deps.RedisClient), middleware.RateLimitByUser(2), orderHandler.CreateOrder)
			orderGroup.GET("", orderHandler.ListOrders)
			orderGroup.GET("/stats", orderHandler.GetOrderStats)
			orderGroup.POST(
				"/:order_number/after-sales",
				middleware.RateLimitByUser(2),
				orderHandler.CreateAfterSalesRequest,
			)
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

		customerServiceGroup := v1.Group("/customer-service")
		{
			customerServiceGroup.GET("/agents", ticketHandler.ListPublicCustomerServiceAgents)
			customerServiceGroup.GET("/products", commercialInventoryProbeGuard, productHandler.ListPublicChatProducts)
			customerServiceGroup.GET("/orders", middleware.AuthMiddleware(authService), orderHandler.ListPublicChatOrders)
			customerServiceGroup.POST("/conversations", middleware.OptionalAuthMiddleware(authService), middleware.RateLimitByUser(3), ticketHandler.EnsurePublicCustomerServiceConversation)
			customerServiceGroup.GET("/has-conversation", middleware.OptionalAuthMiddleware(authService), ticketHandler.HasPublicCustomerServiceConversation)
			customerServiceGroup.POST("/attachments", middleware.OptionalAuthMiddleware(authService), middleware.RateLimitByUserPerMinute(6, 2), ticketHandler.UploadPublicCustomerServiceAttachment)
			customerServiceGroup.POST("/messages", middleware.OptionalAuthMiddleware(authService), middleware.RateLimitByUser(5), ticketHandler.SendPublicCustomerServiceMessage)
			customerServiceGroup.GET("/messages/:conversation_id", middleware.OptionalAuthMiddleware(authService), ticketHandler.GetPublicCustomerServiceMessages)
			customerServiceGroup.GET("/auto-reply/welcome", middleware.OptionalAuthMiddleware(authService), ticketHandler.GetWelcomeMessage)
			customerServiceGroup.POST("/auto-reply/match", middleware.OptionalAuthMiddleware(authService), middleware.RateLimitByUser(5), ticketHandler.MatchKeywordMessage)
			customerServiceGroup.GET("/ws", middleware.OptionalAuthMiddleware(authService), ticketHandler.StreamPublicCustomerServiceWebSocket)
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
			showcaseGroup.GET("/upload-orders", middleware.AuthMiddleware(authService), middleware.RateLimitByUser(10), showcaseHandler.ListUploadOrders)
			showcaseGroup.GET("/:id/images/:image_index/file", showcaseHandler.ServePublicImageFile)
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
			settingsGroup.GET("/website-profile", settingsHandler.GetWebsiteProfile)
			settingsGroup.GET("/website-name", settingsHandler.GetWebsiteName)
			settingsGroup.GET("/refund-return-policy", settingsHandler.GetRefundReturnPolicy)
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
			paymentGroup.GET("/stripe/express-checkout/config", paymentHandler.GetStripeExpressCheckoutConfiguration)
			paymentGroup.POST("/calculate-tax", middleware.RateLimit(5), paymentHandler.CalculateTax)

			// 需要认证的端点
			authPayment := paymentGroup.Group("")
			authPayment.Use(middleware.AuthMiddleware(authService))
			{
				authPayment.GET("/transactions/:id", paymentHandler.GetTransaction)
				authPayment.GET("/orders/:order_id/transactions", paymentHandler.GetOrderTransactions)
				authPayment.GET("/refunds/:id", paymentHandler.GetRefund)
				authPayment.GET("/orders/:order_id/refunds", paymentHandler.GetOrderRefunds)
				authPayment.POST("/stripe/payment-intents", middleware.Idempotency(deps.RedisClient), middleware.RateLimitByUser(3), paymentHandler.CreateStripePaymentIntent)
				authPayment.POST("/paypal/orders", middleware.Idempotency(deps.RedisClient), middleware.RateLimitByUser(3), paymentHandler.CreatePayPalOrder)
				authPayment.POST("/paypal/orders/:paypal_order_id/capture", middleware.Idempotency(deps.RedisClient), middleware.RateLimitByUser(5), paymentHandler.CapturePayPalOrder)
				authPayment.POST("/alipay/orders", middleware.Idempotency(deps.RedisClient), middleware.RateLimitByUser(3), paymentHandler.CreateAlipayOrder)
				authPayment.POST("/alipay/orders/:order_number/confirm", middleware.Idempotency(deps.RedisClient), middleware.RateLimitByUser(5), paymentHandler.ConfirmAlipayOrder)
				authPayment.POST("/wechat/orders", middleware.Idempotency(deps.RedisClient), middleware.RateLimitByUser(3), paymentHandler.CreateWechatOrder)
				authPayment.POST("/wechat/orders/:order_number/confirm", middleware.Idempotency(deps.RedisClient), middleware.RateLimitByUser(5), paymentHandler.ConfirmWechatOrder)
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
			galleryGroup.GET("/slug/:slug", galleryHandler.GetGalleryBySlug)
			galleryGroup.GET("/images/search", galleryHandler.SearchImages)
			galleryGroup.GET("/images/tags", galleryHandler.GetImagesByTags)
			galleryGroup.GET("/:id/images", galleryHandler.GetGalleryImages)
			galleryGroup.GET("/:id", galleryHandler.GetGalleryByID)
		}

		// 订单保修路由
		warrantyGroup := v1.Group("/warranty")
		{
			// 公开端点
			warrantyGroup.POST("/verify-order", middleware.RateLimit(2), warrantyHandler.VerifyWarrantyOrder)
			warrantyGroup.GET("/verify/:token", middleware.RateLimit(5), warrantyHandler.VerifyWarrantyOrderToken)
			warrantyGroup.POST("/claim", middleware.RateLimit(1), warrantyHandler.SubmitWarrantyClaim)

			// 需要认证的端点
			authWarranty := warrantyGroup.Group("")
			authWarranty.Use(middleware.AuthMiddleware(authService))
			{
				authWarranty.GET("/orders/:order_number", warrantyHandler.GetWarrantyStatus)
				authWarranty.GET("/claims/:id", warrantyHandler.GetWarrantyClaim)
			}
		}

	}

	// Sitemap 路由（根路径）
	r.GET("/sitemap.xml", i18nHandler.GetSitemapIndex)
	r.GET("/sitemap-hreflang.xml", i18nHandler.GetHreflangSitemap)
	r.GET("/sitemap-:locale.xml", i18nHandler.GetLocaleSitemap)

	// 健康检查
}
