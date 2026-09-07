package admin

import (
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"

	"github.com/gin-gonic/gin"
)

func registerProductQualityRequirementRoutes(
	authenticated *gin.RouterGroup,
	handler *ProductQualityRequirementHandler,
) {
	productsGroup := authenticated.Group("/products")
	productsGroup.Use(middleware.RequirePermission(auth.PermProductView))
	{
		productsGroup.GET("/:id/fulfillment-requirements", handler.List)
		productsGroup.PUT(
			"/:id/fulfillment-requirements/spoke-tension-qc",
			middleware.RequirePermission(auth.PermProductEdit),
			handler.UpsertSpokeTensionQC,
		)
	}
}
