package admin

import (
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"

	"github.com/gin-gonic/gin"
)

func registerContentRoutes(
	authenticated *gin.RouterGroup,
	contentHandler *ContentHandler,
	pageFeedbackHandler *PageFeedbackHandler,
	homeVisualTileHandler *HomeVisualTileHandler,
	refundReturnPolicyHandler *RefundReturnPolicyHandler,
	mediaHandler *MediaHandler,
	reviewModerationHandler *ReviewModerationHandler,
	faqHandler *FAQHandler,
	galleryHandler *GalleryHandler,
	ugcShowcaseHandler *UGCShowcaseHandler,
	warrantyHandler *WarrantyHandler,
	shipmentRecordHandler *ShipmentRecordHandler,
	subscriptionHandler *SubscriptionHandler,
	ticketHandler *TicketHandler,
	autoReplyHandler *AutoReplyHandler,
	visitorProfileHandler *VisitorProfileHandler,
	visitorRiskHandler *VisitorRiskHandler,
	globalIPBlockHandler *GlobalIPBlockHandler,
) {
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
		customerServiceGroup.POST("/visitor-profiles/:id/ip-block", middleware.AdminOnly(), visitorProfileHandler.BlockVisitorProfileIP)
		customerServiceGroup.DELETE("/visitor-profiles/:id/ip-block", middleware.AdminOnly(), visitorProfileHandler.UnblockVisitorProfileIP)
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

	// 全局 IP/CIDR 封禁规则（访客画像或设置查看权限可查看，变更仅限管理员）
	securityGroup := authenticated.Group("/security")
	securityGroup.Use(middleware.RequireAnyPermission(auth.PermSettingsView, auth.PermTicketView))
	{
		securityGroup.GET("/ip-blocks", globalIPBlockHandler.List)
		securityGroup.POST("/ip-blocks", middleware.AdminOnly(), globalIPBlockHandler.Create)
		securityGroup.DELETE("/ip-blocks/:id", middleware.AdminOnly(), globalIPBlockHandler.Disable)
	}
}
