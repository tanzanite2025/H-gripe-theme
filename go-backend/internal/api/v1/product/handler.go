package product

import (
	"strconv"
	"tanzanite/internal/api/middleware"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/pagination"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	productService *service.ProductService
}

func NewHandler(productService *service.ProductService) *Handler {
	return &Handler{
		productService: productService,
	}
}

func (h *Handler) ListProducts(c *gin.Context) {
	locale := middleware.GetLocale(c)
	status := c.DefaultQuery("status", "active")
	featured := c.Query("featured") == "true"
	params := pagination.ParsePagination(c)

	if c.Query("page_size") == "" {
		params.PageSize = 12
	}

	products, total, err := h.productService.List(locale, status, featured, params.Page, params.PageSize)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Paged(c, products, params.Page, params.PageSize, total)
}

func (h *Handler) GetProduct(c *gin.Context) {
	idOrSlug := c.Param("id")
	locale := middleware.GetLocale(c)

	if id, err := strconv.ParseUint(idOrSlug, 10, 32); err == nil {
		product, err := h.productService.GetByID(uint(id))
		if err != nil {
			apierror.RespondNotFound(c, "Product")
			return
		}
		response.Success(c, product)
		return
	}

	product, err := h.productService.GetBySlug(idOrSlug, locale)
	if err != nil {
		apierror.RespondNotFound(c, "Product")
		return
	}

	response.Success(c, product)
}

func (h *Handler) GetFilterableAttributes(c *gin.Context) {
	attrs, err := h.productService.GetFilterableAttributes()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"data":    attrs,
	})
}

func (h *Handler) ListProductTypes(c *gin.Context) {
	productTypes, err := h.productService.ListProductTypes(false)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, productTypes)
}
