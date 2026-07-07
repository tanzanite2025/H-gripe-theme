package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"tanzanite/internal/domain/product"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

type productCreateRequest struct {
	ProductTypeID *uint                   `json:"product_type_id"`
	Name          string                  `json:"name" binding:"required"`
	Slug          string                  `json:"slug" binding:"required"`
	Description   string                  `json:"description"`
	ShortDesc     string                  `json:"short_description"`
	Weight        int                     `json:"weight_grams" binding:"gte=0"`
	Status        string                  `json:"status" binding:"required,oneof=active inactive out_of_stock"`
	Locale        string                  `json:"locale"`
	ParentID      *uint                   `json:"parent_id"`
	Featured      bool                    `json:"featured"`
	MetaTitle     string                  `json:"meta_title"`
	MetaDesc      string                  `json:"meta_description"`
	Specs         map[string]interface{}  `json:"specs"`
	Variants      []productVariantRequest `json:"variants"`
}

type productUpdateRequest struct {
	ProductTypeID *uint                   `json:"product_type_id"`
	Name          *string                 `json:"name" binding:"omitempty,min=1"`
	Slug          *string                 `json:"slug" binding:"omitempty,min=1"`
	Description   *string                 `json:"description"`
	ShortDesc     *string                 `json:"short_description"`
	Weight        *int                    `json:"weight_grams" binding:"omitempty,gte=0"`
	Status        *string                 `json:"status" binding:"omitempty,oneof=active inactive out_of_stock"`
	Locale        *string                 `json:"locale"`
	ParentID      *uint                   `json:"parent_id"`
	Featured      *bool                   `json:"featured"`
	MetaTitle     *string                 `json:"meta_title"`
	MetaDesc      *string                 `json:"meta_description"`
	Specs         map[string]interface{}  `json:"specs"`
	Variants      []productVariantRequest `json:"variants"`
}

type productVariantRequest struct {
	ID           *uint                  `json:"id"`
	SKU          string                 `json:"sku"`
	Title        string                 `json:"title"`
	OptionValues map[string]interface{} `json:"option_values"`
	Price        float64                `json:"price"`
	SalePrice    *float64               `json:"sale_price"`
	Stock        int                    `json:"stock"`
	Weight       int                    `json:"weight_grams"`
	IsDefault    bool                   `json:"is_default"`
	IsActive     *bool                  `json:"is_active"`
	SortOrder    int                    `json:"sort_order"`
}

func respondProductServiceError(c *gin.Context, err error, fallbackMessage string) {
	switch {
	case errors.Is(err, service.ErrProductNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
	case errors.Is(err, service.ErrProductSKUExists):
		c.JSON(http.StatusConflict, gin.H{"error": "SKU already exists"})
	case errors.Is(err, service.ErrProductTypeNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product type not found"})
	case errors.Is(err, service.ErrProductSpecInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProductVariantInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": fallbackMessage})
	}
}

func normalizeRequestSpecs(raw map[string]interface{}) map[string]string {
	if len(raw) == 0 {
		return nil
	}

	specs := make(map[string]string, len(raw))
	for key, value := range raw {
		switch typed := value.(type) {
		case nil:
			continue
		case string:
			specs[key] = typed
		case bool:
			specs[key] = strconv.FormatBool(typed)
		case float64:
			specs[key] = strconv.FormatFloat(typed, 'f', -1, 64)
		case int:
			specs[key] = strconv.Itoa(typed)
		default:
			encoded, err := json.Marshal(typed)
			if err != nil {
				continue
			}
			specs[key] = string(encoded)
		}
	}
	return specs
}

func normalizeVariantRequests(raw []productVariantRequest) []service.ProductVariantInput {
	if len(raw) == 0 {
		return nil
	}

	variants := make([]service.ProductVariantInput, 0, len(raw))
	for _, item := range raw {
		variants = append(variants, service.ProductVariantInput{
			ID:           item.ID,
			SKU:          item.SKU,
			Title:        item.Title,
			OptionValues: normalizeRequestSpecs(item.OptionValues),
			Price:        item.Price,
			SalePrice:    item.SalePrice,
			Stock:        item.Stock,
			Weight:       item.Weight,
			IsDefault:    item.IsDefault,
			IsActive:     item.IsActive,
			SortOrder:    item.SortOrder,
		})
	}
	return variants
}

// ListProducts 获取商品列表
// GET /api/admin/products
func (h *ProductHandler) ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	locale := c.Query("locale")
	search := c.Query("search")
	featured := c.Query("featured")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	products, total, err := h.productService.ListAdmin(page, pageSize, status, locale, search, featured)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"products": products,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetProduct 获取商品详情
// GET /api/admin/products/:id
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	product, err := h.productService.GetAdminProduct(uint(id))
	if err != nil {
		respondProductServiceError(c, err, "Failed to fetch product")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"product": product,
	})
}

// CreateProduct 创建商品
// POST /api/admin/products
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req productCreateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newProduct, err := h.productService.CreateAdminProduct(service.ProductCreateInput{
		ProductTypeID: req.ProductTypeID,
		Name:          req.Name,
		Slug:          req.Slug,
		Description:   req.Description,
		ShortDesc:     req.ShortDesc,
		Weight:        req.Weight,
		Status:        req.Status,
		Locale:        req.Locale,
		ParentID:      req.ParentID,
		Featured:      req.Featured,
		MetaTitle:     req.MetaTitle,
		MetaDesc:      req.MetaDesc,
		SpecValues:    normalizeRequestSpecs(req.Specs),
		Variants:      normalizeVariantRequests(req.Variants),
	})
	if err != nil {
		respondProductServiceError(c, err, "Failed to create product")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product created successfully",
		"product": newProduct,
	})
}

// UpdateProduct 更新商品
// PUT /api/admin/products/:id
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req productUpdateRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var raw map[string]json.RawMessage
	if err := c.ShouldBindBodyWith(&raw, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, updateParentID := raw["parent_id"]
	_, updateProductTypeID := raw["product_type_id"]
	_, updateSpecs := raw["specs"]
	_, updateVariants := raw["variants"]
	if updateProductTypeID && !updateSpecs {
		updateSpecs = true
	}
	if updateProductTypeID && !updateVariants {
		updateVariants = true
	}

	updatedProduct, err := h.productService.UpdateAdminProduct(uint(id), service.ProductUpdateInput{
		ProductTypeID:       req.ProductTypeID,
		UpdateProductTypeID: updateProductTypeID,
		Name:                req.Name,
		Slug:                req.Slug,
		Description:         req.Description,
		ShortDesc:           req.ShortDesc,
		Weight:              req.Weight,
		Status:              req.Status,
		Locale:              req.Locale,
		ParentID:            req.ParentID,
		UpdateParentID:      updateParentID,
		Featured:            req.Featured,
		MetaTitle:           req.MetaTitle,
		MetaDesc:            req.MetaDesc,
		SpecValues:          normalizeRequestSpecs(req.Specs),
		UpdateSpecValues:    updateSpecs,
		Variants:            normalizeVariantRequests(req.Variants),
		UpdateVariants:      updateVariants,
	})
	if err != nil {
		respondProductServiceError(c, err, "Failed to update product")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product updated successfully",
		"product": updatedProduct,
	})
}

// DeleteProduct 删除商品
// DELETE /api/admin/products/:id
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	if err := h.productService.Delete(uint(id)); err != nil {
		respondProductServiceError(c, err, "Failed to delete product")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product deleted successfully",
	})
}

// UpdateProductStatus 更新商品状态
// PATCH /api/admin/products/:id/status
func (h *ProductHandler) UpdateProductStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active inactive out_of_stock"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.productService.UpdateStatus(uint(id), req.Status); err != nil {
		respondProductServiceError(c, err, "Failed to update product status")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product status updated successfully",
	})
}

// GetProductStats 获取商品统计
// GET /api/admin/products/stats
func (h *ProductHandler) GetProductStats(c *gin.Context) {
	stats, err := h.productService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// BatchUpdateStatus 批量更新商品状态
// POST /api/admin/products/batch-status
func (h *ProductHandler) BatchUpdateStatus(c *gin.Context) {
	var req struct {
		ProductIDs []uint `json:"product_ids" binding:"required,min=1"`
		Status     string `json:"status" binding:"required,oneof=active inactive out_of_stock"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.productService.BatchUpdateStatus(req.ProductIDs, req.Status)
	if err != nil {
		respondProductServiceError(c, err, "Failed to batch update product status")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch update completed",
		"updated": updated,
		"total":   len(req.ProductIDs),
	})
}

// BatchDelete 批量删除商品
// POST /api/admin/products/batch-delete
func (h *ProductHandler) BatchDelete(c *gin.Context) {
	var req struct {
		ProductIDs []uint `json:"product_ids" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deleted, err := h.productService.BatchDelete(req.ProductIDs)
	if err != nil {
		respondProductServiceError(c, err, "Failed to batch delete products")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch delete completed",
		"deleted": deleted,
		"total":   len(req.ProductIDs),
	})
}

// GetFilterableAttributes 获取可过滤属性列表 (公开端点可借用)
func (h *ProductHandler) GetFilterableAttributes(c *gin.Context) {
	attrs, err := h.productService.GetFilterableAttributes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    attrs,
	})
}

func (h *ProductHandler) ListProductTypes(c *gin.Context) {
	productTypes, err := h.productService.ListProductTypes(true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": productTypes})
}

// ListAttributes 获取属性列表
func (h *ProductHandler) ListAttributes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	attrs, total, err := h.productService.ListAttributes(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        attrs,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetAttribute 获取属性详情
func (h *ProductHandler) GetAttribute(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attribute id"})
		return
	}

	attr, err := h.productService.GetAttributeByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "attribute not found"})
		return
	}

	c.JSON(http.StatusOK, attr)
}

// CreateAttribute 创建属性
func (h *ProductHandler) CreateAttribute(c *gin.Context) {
	var attr product.ProductAttribute
	if err := c.ShouldBindJSON(&attr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.productService.CreateAttribute(&attr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, attr)
}

// UpdateAttribute 更新属性
func (h *ProductHandler) UpdateAttribute(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attribute id"})
		return
	}

	var attr product.ProductAttribute
	if err := c.ShouldBindJSON(&attr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	attr.ID = uint(id)

	if err := h.productService.UpdateAttribute(&attr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, attr)
}

// DeleteAttribute 删除属性
func (h *ProductHandler) DeleteAttribute(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attribute id"})
		return
	}

	if err := h.productService.DeleteAttribute(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "attribute deleted successfully"})
}

// GetAttributeValues 获取属性值列表
func (h *ProductHandler) GetAttributeValues(c *gin.Context) {
	attrID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attribute id"})
		return
	}

	values, err := h.productService.GetValuesByAttributeID(uint(attrID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, values)
}

// CreateAttributeValue 创建属性值
func (h *ProductHandler) CreateAttributeValue(c *gin.Context) {
	attrID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attribute id"})
		return
	}

	var val product.AttributeValue
	if err := c.ShouldBindJSON(&val); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	val.AttributeID = uint(attrID)

	if err := h.productService.CreateAttributeValue(&val); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, val)
}

// UpdateAttributeValue 更新属性值
func (h *ProductHandler) UpdateAttributeValue(c *gin.Context) {
	_, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attribute id"})
		return
	}
	valID, err := strconv.ParseUint(c.Param("valueId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value id"})
		return
	}

	var val product.AttributeValue
	if err := c.ShouldBindJSON(&val); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	val.ID = uint(valID)

	if err := h.productService.UpdateAttributeValue(&val); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, val)
}

// DeleteAttributeValue 删除属性值
func (h *ProductHandler) DeleteAttributeValue(c *gin.Context) {
	_, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attribute id"})
		return
	}
	valID, err := strconv.ParseUint(c.Param("valueId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value id"})
		return
	}

	if err := h.productService.DeleteAttributeValue(uint(valID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "attribute value deleted successfully"})
}
