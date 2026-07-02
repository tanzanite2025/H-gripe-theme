package v1

import (
	"tanzanite/internal/api/middleware"
	"tanzanite/internal/api/v1/auth"
	"tanzanite/internal/api/v1/cart"
	"tanzanite/internal/api/v1/chat"
	"tanzanite/internal/api/v1/checkout"
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
	"tanzanite/internal/api/v1/spoke"
	"tanzanite/internal/api/v1/subscription"
	"tanzanite/internal/api/v1/suggestionfeedback"
	"tanzanite/internal/api/v1/ticket"
	"tanzanite/internal/api/v1/wishlist"
	"tanzanite/internal/app"
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
	chatService := services.Chat

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
	cartHandler := cart.NewHandler(cartService)
	settingsHandler := settings.NewHandler(settingService)
	orderHandler := order.NewHandler(orderService, cartService)
	checkoutHandler := checkout.NewHandler(checkoutService, cartService)
	marketingHandler := marketing.NewHandler(marketingService, settingService)
	reviewHandler := review.NewHandler(reviewService)
	ticketHandler := ticket.NewHandler(ticketService, ticket.Options{
		AllowedOrigins: cfg.CORS.AllowedOrigins,
		VisitorSecret:  cfg.JWT.Secret,
	})
	paymentHandler := payment.NewHandler(paymentService, orderService)
	shippingHandler := shipping.NewHandler(services.Shipping)
	galleryHandler := gallery.NewGalleryHandler(galleryService)
	registrationHandler := registration.NewHandler(registrationService, storageSvc)
	subscriptionHandler := subscription.NewHandler(subscriptionService)
	i18nHandler := i18n.NewHandler(postService, sitemapService)
	showcaseHandler := showcase.NewShowcaseHandler(showcaseService)
	wishlistHandler := wishlist.NewHandler(wishlistService)
	feedbackHandler := feedback.NewHandler(feedbackService)
	suggestionFeedbackHandler := suggestionfeedback.NewHandler(suggestionFeedbackService, storageSvc)
	spokeHandler := spoke.NewHandler(services.Spoke)
	chatHandler := chat.NewChatHandler(chatService)
	// API v1 路由组
	v1 := r.Group("/api/v1")
	v1.Use(middleware.CSRFProtection(cfg.CORS.AllowedOrigins))
	{
		// 认证路由（公开）
		authGroup := v1.Group("/auth")
		authGroup.Use(middleware.RateLimit(10)) // 10 RPS for auth endpoints
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/google-login", authHandler.GoogleLogin)
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

		// 产品路由（公开）
		productGroup := v1.Group("/products")
		{
			productGroup.GET("", productHandler.ListProducts)
			productGroup.GET("/attributes/filterable", productHandler.GetFilterableAttributes)
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
			feedbackGroup.POST("", middleware.AuthMiddleware(authService), feedbackHandler.Create)
		}

		suggestionFeedbackGroup := v1.Group("/suggestion-feedback")
		{
			suggestionFeedbackGroup.GET("/eligibility", middleware.OptionalAuthMiddleware(authService), suggestionFeedbackHandler.Eligibility)
			suggestionFeedbackGroup.POST("/upload", middleware.AuthMiddleware(authService), suggestionFeedbackHandler.Upload)
			suggestionFeedbackGroup.POST("", middleware.AuthMiddleware(authService), suggestionFeedbackHandler.Create)
		}

		spokeGroup := v1.Group("/spoke")
		{
			spokeGroup.POST("/calc", spokeHandler.Calculate)
			spokeGroup.GET("/export", spokeHandler.GetExport)
			spokeGroup.GET("/history", spokeHandler.ListHistory)
		}

		checkoutGroup := v1.Group("/checkout")
		checkoutGroup.Use(middleware.AuthMiddleware(authService))
		{
			checkoutGroup.POST("/quote", checkoutHandler.Quote)
		}

		// 订单路由（需要认证）
		orderGroup := v1.Group("/orders")
		orderGroup.Use(middleware.AuthMiddleware(authService))
		{
			orderGroup.POST("", orderHandler.CreateOrder)
			orderGroup.GET("", orderHandler.ListOrders)
			orderGroup.GET("/stats", orderHandler.GetOrderStats)
			orderGroup.GET("/:id", orderHandler.GetOrder)
			orderGroup.POST("/:id/cancel", orderHandler.CancelOrder)
		}

		// 管理员订单路由
		adminOrderGroup := v1.Group("/admin/orders")
		adminOrderGroup.Use(middleware.AuthMiddleware(authService), middleware.RequireRole("admin"))
		{
			adminOrderGroup.GET("", orderHandler.ListAllOrders)
			adminOrderGroup.PUT("/:id/status", orderHandler.UpdateOrderStatus)
		}

		// 营销路由
		marketingGroup := v1.Group("/marketing")
		{
			// 优惠券（公开）
			marketingGroup.GET("/coupons", marketingHandler.ListCoupons)

			// 等级配置（公开）
			marketingGroup.GET("/loyalty/levels", marketingHandler.ListMemberLevels)
			marketingGroup.GET("/loyalty/redeem-options", marketingHandler.ListRedeemGiftCardOptions)

			// 需要认证的营销功能
			authMarketing := marketingGroup.Group("")
			authMarketing.Use(middleware.AuthMiddleware(authService))
			{
				// 优惠券
				authMarketing.POST("/coupons/validate", marketingHandler.ValidateCoupon)

				// 积分和会员
				authMarketing.GET("/loyalty/assets", marketingHandler.GetUserAssets)
				authMarketing.GET("/loyalty/points", marketingHandler.GetPoints)
				authMarketing.GET("/loyalty/info", marketingHandler.GetLoyaltyInfo)
				authMarketing.POST("/loyalty/checkin", marketingHandler.CheckIn)
				authMarketing.POST("/loyalty/referral", marketingHandler.CreateReferral)
				authMarketing.POST("/loyalty/redeem", marketingHandler.RedeemPointsToGiftCard)
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

		customerServiceGroup := v1.Group("/customer-service")
		{
			customerServiceGroup.GET("/agents", ticketHandler.ListPublicCustomerServiceAgents)
			customerServiceGroup.GET("/products", productHandler.ListPublicChatProducts)
			customerServiceGroup.GET("/orders", middleware.AuthMiddleware(authService), orderHandler.ListPublicChatOrders)
			customerServiceGroup.POST("/conversations", middleware.OptionalAuthMiddleware(authService), ticketHandler.EnsurePublicCustomerServiceConversation)
			customerServiceGroup.GET("/has-conversation", middleware.OptionalAuthMiddleware(authService), ticketHandler.HasPublicCustomerServiceConversation)
			customerServiceGroup.POST("/messages", middleware.OptionalAuthMiddleware(authService), ticketHandler.SendPublicCustomerServiceMessage)
			customerServiceGroup.GET("/messages/:conversation_id", middleware.OptionalAuthMiddleware(authService), ticketHandler.GetPublicCustomerServiceMessages)
			customerServiceGroup.GET("/auto-reply/welcome", middleware.OptionalAuthMiddleware(authService), ticketHandler.GetWelcomeMessage)
			customerServiceGroup.POST("/auto-reply/match", middleware.OptionalAuthMiddleware(authService), ticketHandler.MatchKeywordMessage)
			customerServiceGroup.GET("/ws", middleware.OptionalAuthMiddleware(authService), ticketHandler.ServeWS)

			agentGroup := customerServiceGroup.Group("/agent")
			agentGroup.Use(middleware.AuthMiddleware(authService), middleware.RequireRole("admin", "manager", "support"))
			{
				agentGroup.GET("/conversations", ticketHandler.ListCustomerServiceConversations)
				agentGroup.GET("/conversations/:id/messages", ticketHandler.GetCustomerServiceMessages)
				agentGroup.POST("/conversations/:id/transfer", ticketHandler.TransferCustomerServiceConversation)
				agentGroup.POST("/messages", ticketHandler.SendCustomerServiceMessage)
				agentGroup.POST("/messages/read", ticketHandler.MarkCustomerServiceMessagesRead)
				agentGroup.GET("/status", ticketHandler.GetCustomerServiceAgentStatus)
				agentGroup.POST("/status", ticketHandler.UpdateCustomerServiceAgentStatus)
			}
		}

		// 聊天消息持久化路由（新增）
		chatGroup := v1.Group("/chat")
		chatGroup.Use(middleware.AuthMiddleware(authService))
		{
			chatGroup.POST("/messages", chatHandler.SaveMessage)
			chatGroup.GET("/messages", chatHandler.GetMessages)
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
			showcaseGroup.POST("/upload", middleware.AuthMiddleware(authService), showcaseHandler.Upload)
			showcaseGroup.POST("/comments", middleware.AuthMiddleware(authService), showcaseHandler.AddComment)
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
			// 公网暴露的 Webhook 回调路由（免鉴权，内部负责验签）
			paymentGroup.POST("/webhook/:provider", paymentHandler.HandleWebhook)

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
				authPayment.GET("/transactions/:id", paymentHandler.GetTransaction)
				authPayment.GET("/orders/:order_id/transactions", paymentHandler.GetOrderTransactions)
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
			registrationGroup.POST("/verify", registrationHandler.VerifySerialNumber)
			registrationGroup.POST("/warranty/verify-order", registrationHandler.VerifyWarrantyOrder)
			registrationGroup.POST("/warranty/claim", registrationHandler.SubmitWarrantyClaim)

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
				authRegistration.GET("/:registration_id/warranty-claims", registrationHandler.ListRegistrationClaims)
			}
		}

	}

	// Sitemap 路由（根路径）
	r.GET("/sitemap.xml", i18nHandler.GetSitemapIndex)
	r.GET("/sitemap-hreflang.xml", i18nHandler.GetHreflangSitemap)
	r.GET("/sitemap-:locale.xml", i18nHandler.GetLocaleSitemap)

	// 健康检查
}
