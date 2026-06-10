package faq

import (
	"net/http"
	"strconv"
	"tanzanite/internal/domain/faq"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	faqService *service.FAQService
}

func NewHandler(faqService *service.FAQService) *Handler {
	return &Handler{
		faqService: faqService,
	}
}

// CreateFAQRequest 创建FAQ请求
type CreateFAQRequest struct {
	Question string `json:"question" binding:"required"`
	Answer   string `json:"answer" binding:"required"`
	PageID   string `json:"page_id"`
	Category string `json:"category" binding:"required"`
	Locale   string `json:"locale"`
	ParentID *uint  `json:"parent_id"`
	Order    int    `json:"order"`
	Status   string `json:"status"`
}

// UpdateFAQRequest 更新FAQ请求
type UpdateFAQRequest struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
	PageID   string `json:"page_id"`
	Category string `json:"category"`
	Locale   string `json:"locale"`
	ParentID *uint  `json:"parent_id"`
	Order    int    `json:"order"`
	Status   string `json:"status"`
}

// UpdateOrderRequest 更新排序请求
type UpdateOrderRequest struct {
	Order int `json:"order" binding:"required"`
}

// BatchUpdateOrderRequest 批量更新排序请求
type BatchUpdateOrderRequest struct {
	Orders map[uint]int `json:"orders" binding:"required"`
}

// CreateFAQ 创建FAQ（管理员）
func (h *Handler) CreateFAQ(c *gin.Context) {
	var req CreateFAQRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.Locale == "" {
		req.Locale = "en"
	}
	if req.Status == "" {
		req.Status = "published"
	}

	f := &faq.FAQ{
		Question: req.Question,
		Answer:   req.Answer,
		PageID:   req.PageID,
		Category: req.Category,
		Locale:   req.Locale,
		ParentID: req.ParentID,
		Order:    req.Order,
		Status:   req.Status,
	}

	if err := h.faqService.Create(f); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "FAQ created successfully",
		"data":    f,
	})
}

// UpdateFAQ 更新FAQ（管理员）
func (h *Handler) UpdateFAQ(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 获取现有FAQ
	existingFAQ, err := h.faqService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "FAQ not found"})
		return
	}

	var req UpdateFAQRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新字段
	if req.Question != "" {
		existingFAQ.Question = req.Question
	}
	if req.Answer != "" {
		existingFAQ.Answer = req.Answer
	}
	if req.PageID != "" {
		existingFAQ.PageID = req.PageID
	}
	if req.Category != "" {
		existingFAQ.Category = req.Category
	}
	if req.Locale != "" {
		existingFAQ.Locale = req.Locale
	}
	if req.ParentID != nil {
		existingFAQ.ParentID = req.ParentID
	}
	if req.Order != 0 {
		existingFAQ.Order = req.Order
	}
	if req.Status != "" {
		existingFAQ.Status = req.Status
	}

	if err := h.faqService.Update(existingFAQ); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "FAQ updated successfully",
		"data":    existingFAQ,
	})
}

// DeleteFAQ 删除FAQ（管理员）
func (h *Handler) DeleteFAQ(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.faqService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "FAQ deleted successfully",
	})
}

// UpdateFAQOrder 更新FAQ排序（管理员）
func (h *Handler) UpdateFAQOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.faqService.UpdateOrder(uint(id), req.Order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "FAQ order updated successfully",
	})
}

// BatchUpdateFAQOrder 批量更新FAQ排序（管理员）
func (h *Handler) BatchUpdateFAQOrder(c *gin.Context) {
	var req BatchUpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.faqService.BatchUpdateOrder(req.Orders); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "FAQ orders updated successfully",
		"count":   len(req.Orders),
	})
}

// IncrementFAQView 增加FAQ浏览次数
func (h *Handler) IncrementFAQView(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.faqService.IncrementViewCount(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "View count incremented",
	})
}
