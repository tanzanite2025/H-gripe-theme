package product

import (
	"strconv"
	"strings"
	"tanzanite/internal/api/middleware"
	productdomain "tanzanite/internal/domain/product"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListPublicChatProducts(c *gin.Context) {
	locale := middleware.GetLocale(c)
	status := normalizePublicChatProductStatus(c.DefaultQuery("status", "active"))
	keyword := strings.TrimSpace(c.Query("keyword"))
	typeSlug := strings.TrimSpace(c.Query("product_type"))
	priceMin := parseOptionalFloatQuery(c, "price_min")
	priceMax := parseOptionalFloatQuery(c, "price_max")
	specFilters := parseSpecFilterQuery(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("per_page", c.DefaultQuery("page_size", "20")))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	products, total, err := h.productService.SearchPublic(service.ProductSearchInput{
		Locale:      locale,
		Status:      status,
		Keyword:     keyword,
		TypeSlug:    typeSlug,
		PriceMin:    priceMin,
		PriceMax:    priceMax,
		SpecFilters: specFilters,
		Page:        page,
		PageSize:    pageSize,
	})
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	items := make([]gin.H, 0, len(products))
	for _, item := range products {
		items = append(items, makePublicChatProduct(item))
	}

	c.JSON(200, gin.H{
		"success": true,
		"items":   items,
		"meta": gin.H{
			"page":        page,
			"per_page":    pageSize,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			"total":       total,
			"filters":     []string{"keyword", "status", "product_type", "price_min", "price_max", "attributes"},
			"sorting":     []string{"updated_at"},
		},
	})
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
	price, sale := item.DisplayPrices()
	if sale != nil {
		salePrice = *sale
	}
	defaultVariantID := uint(0)
	if defaultVariant := item.DefaultVariant(); defaultVariant != nil {
		defaultVariantID = defaultVariant.ID
	}

	return gin.H{
		"id":                 item.ID,
		"title":              item.Name,
		"status":             item.Status,
		"excerpt":            item.ShortDesc,
		"slug":               item.Slug,
		"sku":                item.DisplaySKU(),
		"thumbnail":          thumbnail,
		"default_variant_id": defaultVariantID,
		"variant_count":      len(item.ActiveVariants()),
		"variants":           makePublicVariants(item.ActiveVariants()),
		"prices": gin.H{
			"regular": price,
			"sale":    salePrice,
			"member":  0,
		},
		"stock": gin.H{
			"quantity": item.TotalVariantStock(),
			"alert":    0,
		},
		"preview_url": "/shop/" + item.Slug,
		"updated_at":  item.UpdatedAt,
		"created_at":  item.CreatedAt,
	}
}

func makePublicVariants(variants []productdomain.ProductVariant) []gin.H {
	items := make([]gin.H, 0, len(variants))
	for _, variant := range variants {
		salePrice := 0.0
		if variant.SalePrice != nil {
			salePrice = *variant.SalePrice
		}
		items = append(items, gin.H{
			"id":            variant.ID,
			"sku":           variant.SKU,
			"title":         variant.Title,
			"option_values": variant.OptionValues,
			"price":         variant.Price,
			"sale_price":    salePrice,
			"stock":         variant.Stock,
			"is_default":    variant.IsDefault,
		})
	}
	return items
}
