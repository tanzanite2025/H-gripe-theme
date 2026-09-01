package admin

import (
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"

	"github.com/gin-gonic/gin"
)

func registerSystemRoutes(
	authenticated *gin.RouterGroup,
	settingsHandler *SettingsHandler,
	commercialCrawlerHandler *CommercialCrawlerProtectionHandler,
	publicChatAgentHandler *PublicChatAgentHandler,
	customerServiceAvatarHandler *CustomerServiceAvatarHandler,
	paymentHandler *PaymentHandler,
	currencyPolicyHandler *CurrencyPolicyHandler,
	siteLogoHandler *SiteLogoHandler,
	exchangeRateHandler *ExchangeRateHandler,
	websiteProfileHandler *WebsiteProfileHandler,
	websiteNameHandler *WebsiteNameHandler,
	shippingHandler *ShippingHandler,
) {
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
}
