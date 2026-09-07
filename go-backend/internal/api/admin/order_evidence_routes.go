package admin

import (
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"

	"github.com/gin-gonic/gin"
)

func registerOrderEvidenceRoutes(
	authenticated *gin.RouterGroup,
	handler *OrderEvidenceHandler,
) {
	workbenchGroup := authenticated.Group("/order-evidence")
	workbenchGroup.Use(middleware.RequirePermission(auth.PermOrderView))
	{
		workbenchGroup.GET("", handler.ListOrders)
	}

	ordersGroup := authenticated.Group("/orders")
	ordersGroup.Use(middleware.RequirePermission(auth.PermOrderView))
	{
		ordersGroup.GET("/:id/evidence", handler.GetPackage)
		ordersGroup.GET("/:id/evidence/export", handler.ExportSnapshot)
		ordersGroup.PATCH("/:id/evidence/items/:item_id", middleware.RequirePermission(auth.PermOrderEdit), handler.UpdateItem)
		ordersGroup.POST("/:id/evidence/items/:item_id/attachments", middleware.RequirePermission(auth.PermOrderEdit), handler.UploadAttachment)
		ordersGroup.GET("/:id/evidence/items/:item_id/attachments/:attachment_id", handler.ServeAttachment)
		ordersGroup.DELETE("/:id/evidence/items/:item_id/attachments/:attachment_id", middleware.RequirePermission(auth.PermOrderEdit), handler.DeleteAttachment)
		ordersGroup.POST("/:id/evidence/lock", middleware.RequirePermission(auth.PermOrderEdit), handler.LockPackage)
		ordersGroup.POST("/:id/evidence/revisions", middleware.RequirePermission(auth.PermOrderEdit), handler.CreateRevision)
	}
}
