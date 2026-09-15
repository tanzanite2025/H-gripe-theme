package admin

import (
	workbenchfeedapi "commerce-platform/internal/api/admin/workbenchfeed"
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"

	"github.com/gin-gonic/gin"
)

func registerWorkbenchFeedRoutes(authenticated *gin.RouterGroup, handler *workbenchfeedapi.Handler) {
	group := authenticated.Group("/workbench-feed")
	group.Use(middleware.RequirePermission(auth.PermWorkbenchFeedView))
	{
		group.GET("/entries", handler.List)
		group.GET("/entries/:id", handler.Get)
		group.GET("/product-options", handler.ProductOptions)
		group.POST("/media", middleware.RequirePermission(auth.PermWorkbenchFeedCreate), handler.UploadMedia)
		group.POST("/entries", middleware.RequirePermission(auth.PermWorkbenchFeedCreate), handler.Create)
		group.PUT("/entries/:id", middleware.RequirePermission(auth.PermWorkbenchFeedEdit), handler.Update)
		group.PATCH("/entries/:id/status", middleware.RequirePermission(auth.PermWorkbenchFeedEdit), handler.UpdateStatus)
		group.DELETE("/entries/:id", middleware.RequirePermission(auth.PermWorkbenchFeedDelete), handler.Delete)
	}
}
