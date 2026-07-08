package product

import (
	"encoding/json"
	"strconv"
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

func parseOptionalFloatQuery(c *gin.Context, key string) *float64 {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil
	}

	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil
	}
	return &value
}

func parseSpecFilterQuery(c *gin.Context) map[string][]string {
	filters := make(map[string][]string)
	query := c.Request.URL.Query()

	for key, values := range query {
		switch {
		case key == "attributes":
			for _, raw := range values {
				mergeAttributeJSON(filters, raw)
			}
		case strings.HasPrefix(key, "attributes["):
			slug := strings.TrimPrefix(key, "attributes[")
			switch {
			case strings.HasSuffix(slug, "][]"):
				slug = strings.TrimSuffix(slug, "][]")
			case strings.HasSuffix(slug, "]"):
				slug = strings.TrimSuffix(slug, "]")
			}
			appendSpecFilterValues(filters, slug, values)
		case strings.HasPrefix(key, "attributes."):
			slug := strings.TrimPrefix(key, "attributes.")
			slug = strings.TrimSuffix(slug, "[]")
			appendSpecFilterValues(filters, slug, values)
		}
	}

	return filters
}

func mergeAttributeJSON(filters map[string][]string, raw string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return
	}

	var listValues map[string][]string
	if err := json.Unmarshal([]byte(raw), &listValues); err == nil {
		for slug, values := range listValues {
			appendSpecFilterValues(filters, slug, values)
		}
		return
	}

	var singleValues map[string]string
	if err := json.Unmarshal([]byte(raw), &singleValues); err == nil {
		for slug, value := range singleValues {
			appendSpecFilterValues(filters, slug, []string{value})
		}
	}
}

func appendSpecFilterValues(filters map[string][]string, slug string, values []string) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return
	}

	for _, raw := range values {
		for _, part := range strings.Split(raw, ",") {
			value := strings.TrimSpace(part)
			if value == "" {
				continue
			}
			filters[slug] = append(filters[slug], value)
		}
	}
}
