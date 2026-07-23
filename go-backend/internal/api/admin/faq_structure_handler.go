package admin

import (
	"net/http"
	"strconv"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

// GetCategories 获取所有分类
// GET /api/admin/faqs/categories
func (h *FAQHandler) GetCategories(c *gin.Context) {
	locale := c.DefaultQuery("locale", "en")

	categories, err := h.faqService.GetCategories(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
	})
}

// ListStructure 获取 FAQ 页面与分类结构
// GET /api/admin/faqs/structure
func (h *FAQHandler) ListStructure(c *gin.Context) {
	locale := c.DefaultQuery("locale", "zh")

	pages, err := h.faqService.ListAdminStructure(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get FAQ structure"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pages": pages,
	})
}

// UpdatePage 更新 FAQ 页面元信息
// PUT /api/admin/faqs/pages/:page_id
func (h *FAQHandler) UpdatePage(c *gin.Context) {
	pageID := c.Param("page_id")
	var req faqPageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	page, err := h.faqService.UpsertAdminPage(pageID, service.FAQPageAdminInput{
		RoutePath: req.RoutePath,
		Domain:    req.Domain,
		Locale:    req.Locale,
		Title:     req.Title,
		Subtitle:  req.Subtitle,
		Status:    req.Status,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update FAQ page"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "FAQ page updated successfully",
		"page":    page,
	})
}

// CreateCategory 创建 FAQ 分类
// POST /api/admin/faqs/categories
func (h *FAQHandler) CreateCategory(c *gin.Context) {
	var req faqCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.faqService.CreateAdminCategory(service.FAQCategoryAdminInput{
		PageID:      req.PageID,
		CategoryKey: req.CategoryKey,
		Name:        req.Name,
		Icon:        req.Icon,
		Locale:      req.Locale,
		Status:      req.Status,
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create FAQ category"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "FAQ category created successfully",
		"category": category,
	})
}

// UpdateCategory 更新 FAQ 分类
// PUT /api/admin/faqs/categories/:id
func (h *FAQHandler) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid FAQ category ID"})
		return
	}

	var req faqCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.faqService.UpdateAdminCategory(uint(id), service.FAQCategoryAdminInput{
		PageID:      req.PageID,
		CategoryKey: req.CategoryKey,
		Name:        req.Name,
		Icon:        req.Icon,
		Locale:      req.Locale,
		Status:      req.Status,
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		if service.IsRecordNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "FAQ category not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update FAQ category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "FAQ category updated successfully",
		"category": category,
	})
}

// DeleteCategory 删除 FAQ 分类
// DELETE /api/admin/faqs/categories/:id
func (h *FAQHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid FAQ category ID"})
		return
	}

	if err := h.faqService.DeleteAdminCategory(uint(id)); err != nil {
		if service.IsRecordNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "FAQ category not found"})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "FAQ category deleted successfully",
	})
}
