package product

import (
	"strings"
	"tanzanite/internal/api/middleware"
	productdomain "tanzanite/internal/domain/product"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/pagination"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	productService *service.ProductService
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

func (h *Handler) ListProducts(c *gin.Context) {
	locale := middleware.GetLocale(c)
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
		"data":      PublicProductsFromDomain(publicProducts),
		"page_size": params.PageSize,
		"has_more":  hasMore,
	})
}

func (h *Handler) GetProduct(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("id"))
	locale := middleware.GetLocale(c)

	product, err := h.productService.GetPublicBySlug(slug, locale)
	if err != nil {
		apierror.RespondNotFound(c, "Product")
		return
	}

	response.Success(c, PublicProductFromDomain(*product))
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
	productTypes, err := h.productService.ListProductTypes(false)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, PublicProductTypesFromDomain(productTypes))
}

func trimPublicProductPage(products []productdomain.Product, pageSize int) ([]productdomain.Product, bool) {
	if pageSize < 1 || len(products) <= pageSize {
		return products, false
	}
	return products[:pageSize], true
}
