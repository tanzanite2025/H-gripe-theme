package product

import (
	"strings"

	"commerce-platform/internal/api/middleware"
	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/pagination"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	productService    *service.ProductService
	storefrontContext *service.StorefrontContextService
}

const (
	publicCatalogDefaultPageSize = 12
	publicCatalogMaxPageSize     = 24
)

func NewHandler(productService *service.ProductService) *Handler {
	return &Handler{
		productService: productService,
	}
}

func (h *Handler) ConfigureStorefrontContext(contextService *service.StorefrontContextService) {
	if h == nil {
		return
	}
	h.storefrontContext = contextService
}

func (h *Handler) ListProducts(c *gin.Context) {
	publicContext := h.resolvePublicContext(c)
	locale := publicContext.Locale
	featured := c.Query("featured") == "true"
	params := pagination.ParsePagination(c)

	if c.Query("page_size") == "" {
		params.PageSize = publicCatalogDefaultPageSize
	}
	if params.PageSize > publicCatalogMaxPageSize {
		params.PageSize = publicCatalogMaxPageSize
	}

	products, _, err := h.productService.ListPublic(
		locale,
		featured,
		params.Page,
		params.PageSize+1,
	)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	publicProducts, hasMore := trimPublicProductPage(products, params.PageSize)
	c.JSON(200, gin.H{
		"code":      0,
		"data":      PublicProductsFromDomainWithLocaleAndDisplayCurrency(publicProducts, publicContext.DisplayCurrency, locale),
		"context":   publicContext.Response,
		"page_size": params.PageSize,
		"has_more":  hasMore,
	})
}

func (h *Handler) GetProduct(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("id"))
	publicContext := h.resolvePublicContext(c)
	locale := publicContext.Locale

	product, translationRoutes, err := h.productService.GetPublicBySlugWithRoutes(slug, locale)
	if err != nil {
		apierror.RespondNotFound(c, "Product")
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"data":    PublicProductFromDomainWithLocaleAndRoutes(*product, publicContext.DisplayCurrency, publicContext.Locale, translationRoutes),
		"context": publicContext.Response,
	})
}

func (h *Handler) GetFilterableAttributes(c *gin.Context) {
	attrs, err := h.productService.GetFilterableAttributes()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"data":    PublicProductAttributesFromDomain(attrs),
	})
}

func (h *Handler) ListProductTypes(c *gin.Context) {
	productTypes, err := h.productService.ListPublicProductTypes(false)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, PublicProductTypesFromDomainWithLocale(productTypes, middleware.GetLocale(c)))
}

func trimPublicProductPage(products []productdomain.Product, pageSize int) ([]productdomain.Product, bool) {
	if pageSize < 1 || len(products) <= pageSize {
		return products, false
	}
	return products[:pageSize], true
}
