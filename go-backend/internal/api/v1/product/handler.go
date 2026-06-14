package product

import (
	"net/http"
	"strconv"
	"strings"
	"tanzanite/internal/api/middleware"
	productdomain "tanzanite/internal/domain/product"
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
		"data":        products,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

func (h *Handler) ListPublicChatProducts(c *gin.Context) {
	locale := middleware.GetLocale(c)
	status := normalizePublicChatProductStatus(c.DefaultQuery("status", "active"))
	keyword := strings.TrimSpace(c.Query("keyword"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("per_page", c.DefaultQuery("page_size", "20")))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	products, total, err := h.productService.SearchPublic(locale, status, keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]gin.H, 0, len(products))
	for _, item := range products {
		items = append(items, makePublicChatProduct(item))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"items":   items,
		"meta": gin.H{
			"page":        page,
			"per_page":    pageSize,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			"total":       total,
			"filters":     []string{"keyword", "status"},
			"sorting":     []string{"updated_at"},
		},
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

func normalizePublicChatProductStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", "publish", "published", "active":
		return "active"
	case "any":
		return ""
	default:
		return strings.ToLower(strings.TrimSpace(status))
	}
}

func makePublicChatProduct(item productdomain.Product) gin.H {
	thumbnail := ""
	if len(item.Images) > 0 {
		thumbnail = item.Images[0].URL
	}

	salePrice := 0.0
	if item.SalePrice != nil {
		salePrice = *item.SalePrice
	}

	return gin.H{
		"id":        item.ID,
		"title":     item.Name,
		"status":    item.Status,
		"excerpt":   item.ShortDesc,
		"slug":      item.Slug,
		"sku":       item.SKU,
		"thumbnail": thumbnail,
		"prices": gin.H{
			"regular": item.Price,
			"sale":    salePrice,
			"member":  0,
		},
		"stock": gin.H{
			"quantity": item.Stock,
			"alert":    0,
		},
		"preview_url": "/shop/" + item.Slug,
		"updated_at":  item.UpdatedAt,
		"created_at":  item.CreatedAt,
	}
}
