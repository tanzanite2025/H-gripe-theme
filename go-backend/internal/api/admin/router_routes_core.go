package admin

import (
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"

	"github.com/gin-gonic/gin"
)

func registerPublicAdminRoutes(
	admin *gin.RouterGroup,
	authHandler *AuthHandler,
	googleMerchantHandler *GoogleMerchantHandler,
	opsConnectorHandler *OpsConnectorHandler,
	socialOAuthHandler *SocialOAuthHandler,
) {
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
}

func registerAuthenticatedCoreRoutes(
	authenticated *gin.RouterGroup,
	authHandler *AuthHandler,
	dashboardHandler *DashboardHandler,
	userHandler *UserHandler,
	customerHandler *CustomerHandler,
) {
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
}
