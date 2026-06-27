package admin

import (
	"tanzanite/internal/api/middleware"
	"tanzanite/internal/api/v1/registration"
	"tanzanite/internal/api/v1/shipping"
	"tanzanite/internal/api/v1/showcase"
	"tanzanite/internal/domain/auth"
	"tanzanite/internal/pkg/cache"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/storage"
	"tanzanite/internal/repository"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterAdminRoutes 注册管理后台路由
func RegisterAdminRoutes(r *gin.Engine, db *gorm.DB, redisCache *cache.RedisCache, cfg *config.Config) {
	// 初始化 repositories
	userRepo := repository.NewUserRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	subscriptionRepo := repository.NewSubscriptionRepository(db)
	productRepo := repository.NewProductRepository(db)
	postRepo := repository.NewPostRepository(db)
	faqRepo := repository.NewFAQRepository(db)
	galleryRepo := repository.NewGalleryRepository(db)
	couponRepo := repository.NewCouponRepository(db)
	loyaltyRepo := repository.NewLoyaltyRepository(db)
	settingRepo := repository.NewSettingRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	showcaseRepo := repository.NewShowcaseRepository(db)
	registrationRepo := repository.NewRegistrationRepository(db)
	shippingRepo := repository.NewShippingRepository(db)

	// 初始化 services
	authService := service.NewAuthService(userRepo, cfg.JWT, cfg.OAuth)
	storageSvc, _ := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: "./uploads", BaseURL: cfg.Server.BaseURL})
	showcaseService := service.NewShowcaseService(showcaseRepo, storageSvc)
	registrationService := service.NewRegistrationService(registrationRepo, productRepo)

	// 初始化 handlers
	authHandler := NewAuthHandler(authService)
	dashboardHandler := NewDashboardHandler(orderRepo, userRepo, ticketRepo, subscriptionRepo)
	userHandler := NewUserHandler(userRepo)
	productHandler := NewProductHandler(productRepo)
	orderHandler := NewOrderHandler(orderRepo)
	contentHandler := NewContentHandler(postRepo)
	faqHandler := NewFAQHandler(faqRepo)
	galleryHandler := NewGalleryHandler(galleryRepo)
	subscriptionHandler := NewSubscriptionHandler(subscriptionRepo)
	ticketHandler := NewTicketHandler(ticketRepo)
	marketingHandler := NewMarketingHandler(couponRepo, loyaltyRepo)
	settingsHandler := NewSettingsHandler(settingRepo, userRepo)
	auditHandler := NewAuditHandler(auditRepo)
	showcaseHandler := showcase.NewShowcaseHandler(showcaseService)
	registrationHandler := registration.NewHandler(registrationRepo, registrationService, orderRepo, storageSvc)
	shippingHandler := shipping.NewHandler(shippingRepo)

	// 管理后台 API 路由组
	admin := r.Group("/api/admin")
	{
		// 认证路由（公开）
		authGroup := admin.Group("/auth")
		authGroup.Use(middleware.RateLimit(10)) // 10 RPS for auth endpoints
		{
			authGroup.POST("/login", authHandler.AdminLogin)
		}

		// 需要认证的路由
		authenticated := admin.Group("")
		authenticated.Use(middleware.AuthMiddleware(authService), middleware.RequireBackofficeAccess())
		{
			// 认证相关
			authGroup := authenticated.Group("/auth")
			{
				authGroup.GET("/profile", authHandler.GetProfile)
				authGroup.POST("/refresh", authHandler.RefreshToken)
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

			// 商品管理（需要商品管理权限）
			productsGroup := authenticated.Group("/products")
			productsGroup.Use(middleware.RequirePermission(auth.PermProductView))
			{
				productsGroup.GET("", productHandler.ListProducts)
				productsGroup.GET("/stats", productHandler.GetProductStats)
				productsGroup.GET("/:id", productHandler.GetProduct)
				productsGroup.POST("", middleware.RequirePermission(auth.PermProductCreate), productHandler.CreateProduct)
				productsGroup.PUT("/:id", middleware.RequirePermission(auth.PermProductEdit), productHandler.UpdateProduct)
				productsGroup.PATCH("/:id/status", middleware.RequirePermission(auth.PermProductEdit), productHandler.UpdateProductStatus)
				productsGroup.PATCH("/:id/stock", middleware.RequirePermission(auth.PermProductEdit), productHandler.UpdateProductStock)
				productsGroup.DELETE("/:id", middleware.RequirePermission(auth.PermProductDelete), productHandler.DeleteProduct)
				productsGroup.POST("/batch-status", middleware.RequirePermission(auth.PermProductEdit), productHandler.BatchUpdateStatus)
				productsGroup.POST("/batch-delete", middleware.RequirePermission(auth.PermProductDelete), productHandler.BatchDelete)
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
				ordersGroup.PATCH("/:id/payment-status", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.UpdatePaymentStatus)
				ordersGroup.PATCH("/:id/shipping-status", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.UpdateShippingStatus)
				ordersGroup.PATCH("/:id/tracking", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.UpdateTrackingInfo)
				ordersGroup.PATCH("/:id/admin-note", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.UpdateAdminNote)
				ordersGroup.POST("/batch-status", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.BatchUpdateStatus)
				ordersGroup.DELETE("/:id", middleware.RequirePermission(auth.PermOrderDelete), orderHandler.DeleteOrder)
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
				faqsGroup.GET("/categories", faqHandler.GetCategories)
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
				registrationsGroup.PUT("/warranty-claims/:id/status", middleware.RequirePermission(auth.PermProductEdit), registrationHandler.UpdateWarrantyClaimStatus)
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
					giftCardsGroup.POST("", middleware.RequirePermission(auth.PermMarketingCreate), marketingHandler.CreateGiftCard)
					giftCardsGroup.PATCH("/:id/status", middleware.RequirePermission(auth.PermMarketingEdit), marketingHandler.UpdateGiftCardStatus)
				}

				// 积分交易管理
				loyaltyGroup := marketingGroup.Group("/loyalty")
				{
					loyaltyGroup.GET("/transactions", marketingHandler.ListLoyaltyTransactions)
					loyaltyGroup.POST("/transactions", middleware.RequirePermission(auth.PermMarketingCreate), marketingHandler.CreateLoyaltyTransaction)
					loyaltyGroup.GET("/check-ins", marketingHandler.ListCheckIns)
					loyaltyGroup.GET("/referrals", marketingHandler.ListReferrals)
					loyaltyGroup.PATCH("/referrals/:id/status", middleware.RequirePermission(auth.PermMarketingEdit), marketingHandler.UpdateReferralStatus)
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

			// 设置管理（需要设置管理权限）
			settingsGroup := authenticated.Group("/settings")
			settingsGroup.Use(middleware.RequirePermission(auth.PermSettingsView))
			{
				settingsGroup.GET("", settingsHandler.GetAllSettings)
				settingsGroup.GET("/groups", settingsHandler.GetGroups)
				settingsGroup.GET("/public-chat-agent-compatibility", settingsHandler.GetPublicChatAgentCompatibility)
				settingsGroup.PUT("", middleware.RequirePermission(auth.PermSettingsEdit), settingsHandler.UpdateSetting)
				settingsGroup.POST("/batch", middleware.RequirePermission(auth.PermSettingsEdit), settingsHandler.BatchUpdateSettings)
				settingsGroup.DELETE("/:key", middleware.RequirePermission(auth.PermSettingsEdit), settingsHandler.DeleteSetting)

				// 分组设置
				settingsGroup.GET("/site", settingsHandler.GetSiteSettings)
				settingsGroup.GET("/email", settingsHandler.GetEmailSettings)
				settingsGroup.GET("/seo", settingsHandler.GetSEOSettings)
				settingsGroup.GET("/social", settingsHandler.GetSocialSettings)
				settingsGroup.GET("/payment", settingsHandler.GetPaymentSettings)
				settingsGroup.GET("/:key", settingsHandler.GetSetting)
			}

			// 物流包装箱规则与承运商管理（需要设置管理权限）
			shippingGroup := authenticated.Group("/shipping")
			shippingGroup.Use(middleware.RequirePermission(auth.PermSettingsView))
			{
				shippingGroup.GET("/packaging-rules", shippingHandler.ListPackagingRules)
				shippingGroup.GET("/packaging-rules/:id", shippingHandler.GetPackagingRule)
				shippingGroup.POST("/packaging-rules", middleware.RequirePermission(auth.PermSettingsEdit), shippingHandler.CreatePackagingRule)
				shippingGroup.PUT("/packaging-rules/:id", middleware.RequirePermission(auth.PermSettingsEdit), shippingHandler.UpdatePackagingRule)
				shippingGroup.DELETE("/packaging-rules/:id", middleware.RequirePermission(auth.PermSettingsEdit), shippingHandler.DeletePackagingRule)
				shippingGroup.POST("/packaging-rules/apply", middleware.RequirePermission(auth.PermSettingsEdit), shippingHandler.CreatePackagingRuleApply)
				shippingGroup.DELETE("/packaging-rules/apply/:applyId", middleware.RequirePermission(auth.PermSettingsEdit), shippingHandler.DeletePackagingRuleApply)

				// 承运商（Carriers）CRUD 管理端点
				shippingGroup.GET("/carriers", shippingHandler.ListCarriers)
				shippingGroup.GET("/carriers/:id", shippingHandler.GetCarrier)
				shippingGroup.POST("/carriers", middleware.RequirePermission(auth.PermSettingsEdit), shippingHandler.CreateCarrier)
				shippingGroup.PUT("/carriers/:id", middleware.RequirePermission(auth.PermSettingsEdit), shippingHandler.UpdateCarrier)
				shippingGroup.DELETE("/carriers/:id", middleware.RequirePermission(auth.PermSettingsEdit), shippingHandler.DeleteCarrier)
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
