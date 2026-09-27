package admin

import (
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"

	"github.com/gin-gonic/gin"
)

func registerLogisticsDomainRoutes(authenticated *gin.RouterGroup, fpxHandler *FpxHandler) {
	fpxGroup := authenticated.Group("/logistics/4px")
	fpxGroup.Use(middleware.RequirePermission(auth.PermFPXView))
	{
		fpxGroup.GET("/overview", fpxHandler.GetOverview)
		fpxGroup.GET("/channels", fpxHandler.ListChannels)
		fpxGroup.GET("/collection", fpxHandler.ListPublishedCollection)
		fpxGroup.PUT("/channels/:id", middleware.RequirePermission(auth.PermFPXManage), fpxHandler.UpdateChannel)
	}
}
