package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type ProductHandler struct {
	productService *service.ProductService
}

var productSEORequestFields = []string{
	"meta_title",
	"meta_description",
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

func rejectProductSEORequestFields(raw map[string]json.RawMessage) (string, bool) {
	for _, field := range productSEORequestFields {
		if _, exists := raw[field]; exists {
			return field, true
		}
	}
	return "", false
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

// GetProductTranslations 获取商品翻译组
// GET /api/admin/products/:id/translations
func (h *ProductHandler) GetProductTranslations(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	group, err := h.productService.GetAdminProductTranslationGroup(uint(id))
	if err != nil {
		respondProductServiceError(c, err, "Failed to fetch product translations")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"translation_group": group,
	})
}

// CopyProductTranslation 复制商品到目标语言
// POST /api/admin/products/:id/translations/copy
func (h *ProductHandler) CopyProductTranslation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req struct {
		TargetLocale string `json:"target_locale" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdProduct, group, err := h.productService.CopyAdminProductTranslation(uint(id), req.TargetLocale)
	if err != nil {
		respondProductServiceError(c, err, "Failed to copy product translation")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":           "Product translation copied successfully",
		"product":           createdProduct,
		"translation_group": group,
	})
}

// CreateProduct 创建商品
// POST /api/admin/products
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req productCreateRequest

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var raw map[string]json.RawMessage
	if err := c.ShouldBindBodyWith(&raw, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if field, blocked := rejectProductSEORequestFields(raw); blocked {
		c.JSON(http.StatusBadRequest, gin.H{"error": field + " must be maintained from SEO / 产品"})
		return
	}

	newProduct, err := h.productService.CreateAdminProduct(service.ProductCreateInput{
		ProductTypeID:        req.ProductTypeID,
		ShippingTemplateID:   req.ShippingTemplateID,
		AfterSalesTemplateID: req.AfterSalesTemplateID,
		PackagingTemplateID:  req.PackagingTemplateID,
		Name:                 req.Name,
		Slug:                 req.Slug,
		Description:          req.Description,
		ShortDesc:            req.ShortDesc,
		Currency:             req.Currency,
		Status:               req.Status,
		Locale:               req.Locale,
		ParentID:             req.ParentID,
		Featured:             req.Featured,
		SpecValues:           normalizeRequestSpecs(req.Specs),
		Variants:             normalizeVariantRequests(req.Variants),
		VariantOptionValues:  normalizeVariantOptionValueRequests(req.VariantOptionValues),
		Media:                normalizeMediaRequests(req.Media),
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
	if field, blocked := rejectProductSEORequestFields(raw); blocked {
		c.JSON(http.StatusBadRequest, gin.H{"error": field + " must be maintained from SEO / 产品"})
		return
	}

	_, updateParentID := raw["parent_id"]
	_, updateProductTypeID := raw["product_type_id"]
	_, updateShippingTemplateID := raw["shipping_template_id"]
	_, updateAfterSalesTemplateID := raw["after_sales_template_id"]
	_, updatePackagingTemplateID := raw["packaging_template_id"]
	_, updateCurrency := raw["currency"]
	_, updateSpecs := raw["specs"]
	_, updateVariants := raw["variants"]
	_, updateVariantOptionValues := raw["variant_option_values"]
	_, updateMedia := raw["media"]
	if updateProductTypeID && !updateSpecs {
		updateSpecs = true
	}
	if updateProductTypeID && !updateVariants {
		updateVariants = true
	}
	if updateProductTypeID && !updateVariantOptionValues {
		updateVariantOptionValues = true
	}

	updatedProduct, err := h.productService.UpdateAdminProduct(uint(id), service.ProductUpdateInput{
		ProductTypeID:              req.ProductTypeID,
		UpdateProductTypeID:        updateProductTypeID,
		ShippingTemplateID:         req.ShippingTemplateID,
		UpdateShippingTemplateID:   updateShippingTemplateID,
		AfterSalesTemplateID:       req.AfterSalesTemplateID,
		UpdateAfterSalesTemplateID: updateAfterSalesTemplateID,
		PackagingTemplateID:        req.PackagingTemplateID,
		UpdatePackagingTemplateID:  updatePackagingTemplateID,
		Name:                       req.Name,
		Slug:                       req.Slug,
		Description:                req.Description,
		ShortDesc:                  req.ShortDesc,
		Currency:                   req.Currency,
		UpdateCurrency:             updateCurrency,
		Status:                     req.Status,
		Locale:                     req.Locale,
		ParentID:                   req.ParentID,
		UpdateParentID:             updateParentID,
		Featured:                   req.Featured,
		SpecValues:                 normalizeRequestSpecs(req.Specs),
		UpdateSpecValues:           updateSpecs,
		Variants:                   normalizeVariantRequests(req.Variants),
		UpdateVariants:             updateVariants,
		VariantOptionValues:        normalizeVariantOptionValueRequests(req.VariantOptionValues),
		UpdateVariantOptionValues:  updateVariantOptionValues,
		Media:                      normalizeMediaRequests(req.Media),
		UpdateMedia:                updateMedia,
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
	includeDisabled := c.Query("include_disabled") == "true"
	productTypes, err := h.productService.ListProductTypes(includeDisabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": productTypes})
}
