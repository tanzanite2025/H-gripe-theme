package admin

import (
	"errors"
	"net/http"
	"strconv"
	"tanzanite/internal/domain/faq"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FAQHandler struct {
	faqService *service.FAQService
}

func NewFAQHandler(faqService *service.FAQService) *FAQHandler {
	return &FAQHandler{
		faqService: faqService,
	}
}

// ListFAQs 获取FAQ列表
// GET /api/admin/faqs
func (h *FAQHandler) ListFAQs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	locale := c.Query("locale")
	category := c.Query("category")
	pageID := c.Query("page_id")
	status := c.Query("status")
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	faqs, total, err := h.faqService.ListAdmin(locale, pageID, category, status, search, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch FAQs"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"faqs": faqs,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetFAQ 获取FAQ详情
// GET /api/admin/faqs/:id
func (h *FAQHandler) GetFAQ(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid FAQ ID"})
		return
	}

	faqItem, err := h.faqService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "FAQ not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"faq": faqItem,
	})
}

// CreateFAQ 创建FAQ
// POST /api/admin/faqs
func (h *FAQHandler) CreateFAQ(c *gin.Context) {
	var req struct {
		Question string `json:"question" binding:"required"`
		Answer   string `json:"answer" binding:"required"`
		PageID   string `json:"page_id"`
		Category string `json:"category" binding:"required"`
		Locale   string `json:"locale" binding:"required"`
		Status   string `json:"status" binding:"required,oneof=draft published"`
		Order    int    `json:"order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newFAQ := &faq.FAQ{
		Question: req.Question,
		Answer:   req.Answer,
		PageID:   req.PageID,
		Category: req.Category,
		Locale:   req.Locale,
		Status:   req.Status,
		Order:    req.Order,
	}

	if err := h.faqService.Create(newFAQ); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create FAQ"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "FAQ created successfully",
		"faq":     newFAQ,
	})
}

// UpdateFAQ 更新FAQ
// PUT /api/admin/faqs/:id
func (h *FAQHandler) UpdateFAQ(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid FAQ ID"})
		return
	}

	var req struct {
		Question string `json:"question"`
		Answer   string `json:"answer"`
		PageID   string `json:"page_id"`
		Category string `json:"category"`
		Locale   string `json:"locale"`
		Status   string `json:"status" binding:"omitempty,oneof=draft published"`
		Order    int    `json:"order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedFAQ, err := h.faqService.UpdateAdminFAQ(uint(id), service.FAQAdminUpdateInput{
		Question: req.Question,
		Answer:   req.Answer,
		PageID:   req.PageID,
		Category: req.Category,
		Locale:   req.Locale,
		Status:   req.Status,
		Order:    req.Order,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "FAQ not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update FAQ"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "FAQ updated successfully",
		"faq":     updatedFAQ,
	})
}

// DeleteFAQ 删除FAQ
// DELETE /api/admin/faqs/:id
func (h *FAQHandler) DeleteFAQ(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid FAQ ID"})
		return
	}

	if err := h.faqService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete FAQ"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "FAQ deleted successfully",
	})
}

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

// UpdateOrder 更新排序
// PATCH /api/admin/faqs/:id/order
func (h *FAQHandler) UpdateOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid FAQ ID"})
		return
	}

	var req struct {
		Order int `json:"order" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.faqService.UpdateOrder(uint(id), req.Order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order updated successfully",
	})
}

// BatchDelete 批量删除FAQ
// POST /api/admin/faqs/batch-delete
func (h *FAQHandler) BatchDelete(c *gin.Context) {
	var req struct {
		FAQIDs []uint `json:"faq_ids" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deleted, err := h.faqService.BatchDelete(req.FAQIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to batch delete FAQs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch delete completed",
		"deleted": deleted,
		"total":   len(req.FAQIDs),
	})
}
