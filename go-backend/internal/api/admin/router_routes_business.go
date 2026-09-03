package admin

import (
	seoapi "commerce-platform/internal/api/admin/seo"
	urlapi "commerce-platform/internal/api/admin/urlmanagement"
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func registerBusinessRoutes(
	authenticated *gin.RouterGroup,
	marketingHandler *MarketingHandler,
	exchangeRateHandler *ExchangeRateHandler,
	seoHomeHandler *seoapi.HomeHandler,
	seoArticlesHandler *seoapi.ArticlesHandler,
	seoProductsHandler *seoapi.ProductsHandler,
	seoCategoriesHandler *seoapi.CategoriesHandler,
	urlRoutesHandler *urlapi.RoutesHandler,
	urlSearchProfilesHandler *urlapi.SearchProfilesHandler,
	urlRedirectsHandler *urlapi.RedirectsHandler,
	urlIssuesHandler *urlapi.IssuesHandler,
	analyticsHandler *AnalyticsHandler,
	serviceCenterHandler *ServiceCenterHandler,
	redisClient redis.UniversalClient,
) {
	// 营销管理（需要营销管理权限）
	marketingGroup := authenticated.Group("/marketing")
	marketingGroup.Use(middleware.RequirePermission(auth.PermMarketingView))
	{
		// 营销统计
		marketingGroup.GET("/stats", marketingHandler.GetMarketingStats)
		marketingGroup.GET("/risk-analysis", marketingHandler.GetPromotionRiskAnalysis)

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
			middleware.Idempotency(redisClient),
			middleware.RateLimitByUserPerMinuteRedis(redisClient),
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
		urlGroup.GET("/search-profiles", urlSearchProfilesHandler.List)
		urlGroup.GET("/search-profiles/:id", urlSearchProfilesHandler.Get)
		urlGroup.PUT("/search-profiles/:id", middleware.RequirePermission(auth.PermURLEdit), urlSearchProfilesHandler.Upsert)
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
		githubGroup := servicesGroup.Group("/github")
		{
			githubGroup.GET("", serviceCenterHandler.GitHub)
			githubGroup.POST("/oauth/start", middleware.RequirePermission(auth.PermServicesManage), serviceCenterHandler.StartGitHubOAuth)
			githubGroup.POST("/connectors/:id/test", middleware.RequirePermission(auth.PermServicesManage), serviceCenterHandler.TestGitHubConnection)
		}
		cloudflareGroup := servicesGroup.Group("/cloudflare")
		{
			cloudflareGroup.GET("", serviceCenterHandler.Cloudflare)
			cloudflareGroup.GET("/cache-rules", serviceCenterHandler.GetCloudflareCacheRules)
			cloudflareGroup.POST("/oauth/start", middleware.RequirePermission(auth.PermServicesManage), serviceCenterHandler.StartCloudflareOAuth)
			cloudflareGroup.POST("/connectors/:id/test", middleware.RequirePermission(auth.PermServicesManage), serviceCenterHandler.TestCloudflareConnection)
			cloudflareGroup.PATCH("/cache-rules/:rule_id/enabled", middleware.RequirePermission(auth.PermServicesManage), serviceCenterHandler.UpdateCloudflareCacheRuleEnabled)
		}
	}
}
