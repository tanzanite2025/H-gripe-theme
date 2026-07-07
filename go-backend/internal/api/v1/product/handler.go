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

// ListProducts 获取产品列表
func (h *Handler) ListProducts(c *gin.Context) {
	locale := middleware.GetLocale(c)
	status := c.DefaultQuery("status", "active")
	featured := c.Query("featured") == "true"
	params := pagination.ParsePagination(c)

	// 覆盖默认pageSize为12（产品展示常用）
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

// GetProduct 获取单个产品
func (h *Handler) GetProduct(c *gin.Context) {
	idOrSlug := c.Param("id")
	locale := middleware.GetLocale(c)

	// 尝试作为ID解析
	if id, err := strconv.ParseUint(idOrSlug, 10, 32); err == nil {
		product, err := h.productService.GetByID(uint(id))
		if err != nil {
			apierror.RespondNotFound(c, "Product")
			return
		}
		response.Success(c, product)
		return
	}

	// 作为slug查询
	product, err := h.productService.GetBySlug(idOrSlug, locale)
	if err != nil {
		apierror.RespondNotFound(c, "Product")
		return
	}

	response.Success(c, product)
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

// GetFilterableAttributes 获取可过滤属性列表 (公开端点)
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

// ListAttributes 获取属性列表
func (h *Handler) ListAttributes(c *gin.Context) {
	params := pagination.ParsePagination(c)

	attrs, total, err := h.productService.ListAttributes(params.Page, params.PageSize)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Paged(c, attrs, params.Page, params.PageSize, total)
}

// GetAttribute 获取属性详情
func (h *Handler) GetAttribute(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid attribute ID")
		return
	}

	attr, err := h.productService.GetAttributeByID(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Attribute")
		return
	}

	response.Success(c, attr)
}

// CreateAttribute 创建属性
func (h *Handler) CreateAttribute(c *gin.Context) {
	var attr productdomain.ProductAttribute
	if err := c.ShouldBindJSON(&attr); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	if err := h.productService.CreateAttribute(&attr); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Created(c, attr)
}

// UpdateAttribute 更新属性
func (h *Handler) UpdateAttribute(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid attribute ID")
		return
	}

	var attr productdomain.ProductAttribute
	if err := c.ShouldBindJSON(&attr); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}
	attr.ID = uint(id)

	if err := h.productService.UpdateAttribute(&attr); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, attr)
}

// DeleteAttribute 删除属性
func (h *Handler) DeleteAttribute(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid attribute ID")
		return
	}

	if err := h.productService.DeleteAttribute(uint(id)); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Attribute deleted successfully", nil)
}

// GetAttributeValues 获取属性值列表
func (h *Handler) GetAttributeValues(c *gin.Context) {
	attrID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid attribute ID")
		return
	}

	values, err := h.productService.GetValuesByAttributeID(uint(attrID))
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, values)
}

// CreateAttributeValue 创建属性值
func (h *Handler) CreateAttributeValue(c *gin.Context) {
	attrID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid attribute ID")
		return
	}

	var val productdomain.AttributeValue
	if err := c.ShouldBindJSON(&val); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}
	val.AttributeID = uint(attrID)

	if err := h.productService.CreateAttributeValue(&val); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Created(c, val)
}

// UpdateAttributeValue 更新属性值
func (h *Handler) UpdateAttributeValue(c *gin.Context) {
	_, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid attribute ID")
		return
	}
	valID, err := strconv.ParseUint(c.Param("valueId"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid value ID")
		return
	}

	var val productdomain.AttributeValue
	if err := c.ShouldBindJSON(&val); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}
	val.ID = uint(valID)

	if err := h.productService.UpdateAttributeValue(&val); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, val)
}

// DeleteAttributeValue 删除属性值
func (h *Handler) DeleteAttributeValue(c *gin.Context) {
	_, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid attribute ID")
		return
	}
	valID, err := strconv.ParseUint(c.Param("valueId"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid value ID")
		return
	}

	if err := h.productService.DeleteAttributeValue(uint(valID)); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Attribute value deleted successfully", nil)
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
