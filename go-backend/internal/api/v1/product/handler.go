package product

import (
	"net/http"
	"strconv"
	"tanzanite/internal/api/middleware"
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

// ListProducts 获取产品列表
func (h *Handler) ListProducts(c *gin.Context) {
	locale := middleware.GetLocale(c)
	status := c.DefaultQuery("status", "active")
	featured := c.Query("featured") == "true"
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 12
	}

	products, total, err := h.productService.List(locale, status, featured, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       products,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetProduct 获取单个产品
func (h *Handler) GetProduct(c *gin.Context) {
	idOrSlug := c.Param("id")
	locale := middleware.GetLocale(c)

	// 尝试作为ID解析
	if id, err := strconv.ParseUint(idOrSlug, 10, 32); err == nil {
		product, err := h.productService.GetByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusOK, product)
		return
	}

	// 作为slug查询
	product, err := h.productService.GetBySlug(idOrSlug, locale)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}
