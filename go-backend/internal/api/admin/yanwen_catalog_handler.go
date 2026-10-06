package admin

import (
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

// Yanwen official directory read and synchronization admin endpoints.

type YanwenCatalogHandler struct {
	service *service.YanwenOfficialCatalogService
}

func NewYanwenCatalogHandler(catalogService *service.YanwenOfficialCatalogService) *YanwenCatalogHandler {
	return &YanwenCatalogHandler{service: catalogService}
}

func (h *YanwenCatalogHandler) ListYanwenOfficialProducts(c *gin.Context) {
	environment := strings.ToLower(strings.TrimSpace(c.DefaultQuery("environment", "fat")))
	if environment == "test" {
		environment = "fat"
	}
	if environment != "fat" && environment != "production" {
		apierror.RespondBadRequest(c, "Yanwen environment must be fat or production")
		return
	}
	products, err := h.service.ListYanwenOfficialProducts(environment)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"data": products})
}

func (h *YanwenCatalogHandler) SyncYanwenOfficialProducts(c *gin.Context) {
	var input service.YanwenAPIConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	result, err := h.service.SyncYanwenOfficialProducts(c.Request.Context(), input)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *YanwenCatalogHandler) ListYanwenOfficialCountries(c *gin.Context) {
	environment := strings.ToLower(strings.TrimSpace(c.DefaultQuery("environment", "fat")))
	if environment == "test" {
		environment = "fat"
	}
	if environment != "fat" && environment != "production" {
		apierror.RespondBadRequest(c, "Yanwen environment must be fat or production")
		return
	}
	countries, err := h.service.ListYanwenOfficialCountries(environment)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"data": countries})
}

func (h *YanwenCatalogHandler) SyncYanwenOfficialCountries(c *gin.Context) {
	var input service.YanwenAPIConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	result, err := h.service.SyncYanwenOfficialCountries(c.Request.Context(), input)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *YanwenCatalogHandler) ListYanwenOfficialWarehouses(c *gin.Context) {
	environment := strings.ToLower(strings.TrimSpace(c.DefaultQuery("environment", "fat")))
	if environment == "test" {
		environment = "fat"
	}
	if environment != "fat" && environment != "production" {
		apierror.RespondBadRequest(c, "Yanwen environment must be fat or production")
		return
	}
	warehouses, err := h.service.ListYanwenOfficialWarehouses(environment)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"data": warehouses})
}

func (h *YanwenCatalogHandler) SyncYanwenOfficialWarehouses(c *gin.Context) {
	var input service.YanwenAPIConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	result, err := h.service.SyncYanwenOfficialWarehouses(c.Request.Context(), input)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Success(c, result)
}

func registerYanwenCatalogRoutes(authenticated *gin.RouterGroup, handler *YanwenCatalogHandler) {
	group := authenticated.Group("/logistics/yanwen/api")
	group.Use(middleware.RequirePermission(auth.PermYanwenView))
	group.GET("/countries", handler.ListYanwenOfficialCountries)
	group.GET("/warehouses", handler.ListYanwenOfficialWarehouses)
	group.GET("/products", handler.ListYanwenOfficialProducts)
	group.POST("/sync-countries", middleware.RequirePermission(auth.PermYanwenManage), handler.SyncYanwenOfficialCountries)
	group.POST("/sync-warehouses", middleware.RequirePermission(auth.PermYanwenManage), handler.SyncYanwenOfficialWarehouses)
	group.POST("/sync-products", middleware.RequirePermission(auth.PermYanwenManage), handler.SyncYanwenOfficialProducts)
}
