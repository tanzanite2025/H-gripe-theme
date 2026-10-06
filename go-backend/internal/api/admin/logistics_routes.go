package admin

import (
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"

	"github.com/gin-gonic/gin"
)

func registerLogisticsDomainRoutes(authenticated *gin.RouterGroup, fpxHandler *FpxHandler, yanwenCollectionHandler *YanwenCollectionHandler) {
	fpxGroup := authenticated.Group("/logistics/4px")
	fpxGroup.Use(middleware.RequirePermission(auth.PermFPXView))
	{
		fpxGroup.GET("/overview", fpxHandler.GetOverview)
		fpxGroup.GET("/channels", fpxHandler.ListChannels)
		fpxGroup.PUT("/channels/:id", middleware.RequirePermission(auth.PermFPXManage), fpxHandler.UpdateChannel)
	}
	fpxCollectionGroup := authenticated.Group("/logistics/4px")
	fpxCollectionGroup.GET(
		"/collection",
		middleware.RequireAnyPermission(auth.PermShippingView, auth.PermFPXView),
		fpxHandler.ListPublishedCollection,
	)

	yanwenGroup := authenticated.Group("/logistics/yanwen")
	yanwenGroup.Use(middleware.RequirePermission(auth.PermYanwenView))
	{
		yanwenGroup.GET("/channels", yanwenCollectionHandler.ListYanwenPublishedChannels)
		yanwenGroup.POST("/channels", middleware.RequirePermission(auth.PermYanwenManage), yanwenCollectionHandler.CreateYanwenPublishedChannel)
		yanwenGroup.PUT("/channels/:id", middleware.RequirePermission(auth.PermYanwenManage), yanwenCollectionHandler.UpdateYanwenPublishedChannel)
		yanwenGroup.DELETE("/channels/:id", middleware.RequirePermission(auth.PermYanwenManage), yanwenCollectionHandler.DeleteYanwenPublishedChannel)
	}
	yanwenCollectionGroup := authenticated.Group("/logistics/yanwen")
	yanwenCollectionGroup.GET(
		"/collection",
		middleware.RequireAnyPermission(auth.PermShippingView, auth.PermYanwenView),
		yanwenCollectionHandler.ListYanwenPublishedCollectionReferences,
	)
}
