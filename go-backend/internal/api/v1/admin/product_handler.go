package admin

import (
	"net/http"
	"strconv"
	"tanzanite/internal/domain/product"
	"tanzanite/internal/repository"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productRepo *repository.ProductRepository
}

func NewProductHandler(productRepo *repository.ProductRepository) *ProductHandler {
	return &ProductHandler{
		productRepo: productRepo,
	}
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

	products, total, err := h.productRepo.FindAllWithFilters(page, pageSize, status, locale, search, featured)
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

	product, err := h.productRepo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"product": product,
	})
}

// CreateProduct 创建商品
// POST /api/admin/products
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req struct {
		SKU         string   `json:"sku" binding:"required"`
		Name        string   `json:"name" binding:"required"`
		Slug        string   `json:"slug" binding:"required"`
		Description string   `json:"description"`
		ShortDesc   string   `json:"short_description"`
		Price       float64  `json:"price" binding:"required,gt=0"`
		SalePrice   *float64 `json:"sale_price"`
		Stock       int      `json:"stock" binding:"gte=0"`
		Weight      int      `json:"weight_grams"`
		Status      string   `json:"status" binding:"required,oneof=active inactive out_of_stock"`
		Locale      string   `json:"locale"`
		ParentID    *uint    `json:"parent_id"`
		Featured    bool     `json:"featured"`
		MetaTitle   string   `json:"meta_title"`
		MetaDesc    string   `json:"meta_description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查 SKU 是否已存在
	existingProduct, _ := h.productRepo.FindBySKU(req.SKU)
	if existingProduct != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "SKU already exists"})
		return
	}

	// 创建商品
	newProduct := &product.Product{
		SKU:         req.SKU,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		ShortDesc:   req.ShortDesc,
		Price:       req.Price,
		SalePrice:   req.SalePrice,
		Stock:       req.Stock,
		Weight:      req.Weight,
		Status:      req.Status,
		Locale:      req.Locale,
		ParentID:    req.ParentID,
		Featured:    req.Featured,
		MetaTitle:   req.MetaTitle,
		MetaDesc:    req.MetaDesc,
	}

	if err := h.productRepo.Create(newProduct); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
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

	var req struct {
		SKU         string   `json:"sku"`
		Name        string   `json:"name"`
		Slug        string   `json:"slug"`
		Description string   `json:"description"`
		ShortDesc   string   `json:"short_description"`
		Price       float64  `json:"price" binding:"omitempty,gt=0"`
		SalePrice   *float64 `json:"sale_price"`
		Stock       int      `json:"stock" binding:"omitempty,gte=0"`
		Weight      int      `json:"weight_grams"`
		Status      string   `json:"status" binding:"omitempty,oneof=active inactive out_of_stock"`
		Locale      string   `json:"locale"`
		ParentID    *uint    `json:"parent_id"`
		Featured    bool     `json:"featured"`
		MetaTitle   string   `json:"meta_title"`
		MetaDesc    string   `json:"meta_description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取现有商品
	existingProduct, err := h.productRepo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// 更新字段
	if req.SKU != "" && req.SKU != existingProduct.SKU {
		// 检查新 SKU 是否已被使用
		skuProduct, _ := h.productRepo.FindBySKU(req.SKU)
		if skuProduct != nil && skuProduct.ID != existingProduct.ID {
			c.JSON(http.StatusConflict, gin.H{"error": "SKU already exists"})
			return
		}
		existingProduct.SKU = req.SKU
	}

	if req.Name != "" {
		existingProduct.Name = req.Name
	}
	if req.Slug != "" {
		existingProduct.Slug = req.Slug
	}
	if req.Description != "" {
		existingProduct.Description = req.Description
	}
	if req.ShortDesc != "" {
		existingProduct.ShortDesc = req.ShortDesc
	}
	if req.Price > 0 {
		existingProduct.Price = req.Price
	}
	existingProduct.SalePrice = req.SalePrice
	if req.Stock >= 0 {
		existingProduct.Stock = req.Stock
	}
	if req.Weight > 0 {
		existingProduct.Weight = req.Weight
	}
	if req.Status != "" {
		existingProduct.Status = req.Status
	}
	if req.Locale != "" {
		existingProduct.Locale = req.Locale
	}
	existingProduct.ParentID = req.ParentID
	existingProduct.Featured = req.Featured
	if req.MetaTitle != "" {
		existingProduct.MetaTitle = req.MetaTitle
	}
	if req.MetaDesc != "" {
		existingProduct.MetaDesc = req.MetaDesc
	}

	if err := h.productRepo.Update(existingProduct); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product updated successfully",
		"product": existingProduct,
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

	if err := h.productRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
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

	if err := h.productRepo.UpdateStatus(uint(id), req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product status updated successfully",
	})
}

// UpdateProductStock 更新商品库存
// PATCH /api/admin/products/:id/stock
func (h *ProductHandler) UpdateProductStock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req struct {
		Stock int `json:"stock" binding:"required,gte=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.productRepo.UpdateStock(uint(id), req.Stock); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product stock"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product stock updated successfully",
	})
}

// GetProductStats 获取商品统计
// GET /api/admin/products/stats
func (h *ProductHandler) GetProductStats(c *gin.Context) {
	stats, err := h.productRepo.GetStats()
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

	updated := 0
	for _, id := range req.ProductIDs {
		if err := h.productRepo.UpdateStatus(id, req.Status); err == nil {
			updated++
		}
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

	deleted := 0
	for _, id := range req.ProductIDs {
		if err := h.productRepo.Delete(id); err == nil {
			deleted++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch delete completed",
		"deleted": deleted,
		"total":   len(req.ProductIDs),
	})
}
