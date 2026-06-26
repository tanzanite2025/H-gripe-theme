package subscription

import (
	"net/http"
	"strconv"
	"strings"

	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	subscriptionService *service.SubscriptionService
}

func NewHandler(subscriptionService *service.SubscriptionService) *Handler {
	return &Handler{
		subscriptionService: subscriptionService,
	}
}

// Subscribe 订阅
// POST /api/v1/subscriptions
func (h *Handler) Subscribe(c *gin.Context) {
	var req struct {
		Email  string   `json:"email" binding:"required,email"`
		Source string   `json:"source"`
		Locale string   `json:"locale"`
		Tags   []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.Source == "" {
		req.Source = "website"
	}
	if req.Locale == "" {
		req.Locale = "en"
	}

	sub, err := h.subscriptionService.Subscribe(req.Email, req.Source, req.Locale, req.Tags)
	if err != nil {
		if err.Error() == "email already subscribed" {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already subscribed"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Subscribed successfully",
		"data":    sub,
	})
}

// Unsubscribe 退订
// GET /api/v1/subscriptions/unsubscribe/:token
func (h *Handler) Unsubscribe(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	if err := h.subscriptionService.Unsubscribe(token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unsubscribed successfully"})
}

// UnsubscribeByEmail 通过邮箱退订
// POST /api/v1/subscriptions/unsubscribe
func (h *Handler) UnsubscribeByEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.subscriptionService.UnsubscribeByEmail(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unsubscribed successfully"})
}

// Resubscribe 重新订阅
// POST /api/v1/subscriptions/resubscribe
func (h *Handler) Resubscribe(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.subscriptionService.Resubscribe(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Resubscribed successfully"})
}

// GetSubscription 获取订阅状态
// GET /api/v1/subscriptions/status/:email
func (h *Handler) GetSubscription(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	sub, err := h.subscriptionService.GetSubscription(email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": sub})
}

// ========== 管理员 API ==========

// GetAllSubscriptions 获取所有订阅（管理员）
// GET /api/v1/admin/subscriptions
func (h *Handler) GetAllSubscriptions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	subscriptions, total, err := h.subscriptionService.GetAllSubscriptions(page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"data": subscriptions,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetSubscriptionsByTags 根据标签获取订阅（管理员）
// GET /api/v1/admin/subscriptions/tags?tags=tag1,tag2
func (h *Handler) GetSubscriptionsByTags(c *gin.Context) {
	tagsStr := c.Query("tags")
	if tagsStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tags parameter is required"})
		return
	}

	tags := strings.Split(tagsStr, ",")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	subscriptions, total, err := h.subscriptionService.GetSubscriptionsByTags(tags, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"data": subscriptions,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
		"tags": tags,
	})
}

// GetStats 获取订阅统计（管理员）
// GET /api/v1/admin/subscriptions/stats
func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.subscriptionService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// DeleteSubscription 删除订阅（管理员）
// DELETE /api/v1/admin/subscriptions/:email
func (h *Handler) DeleteSubscription(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	if err := h.subscriptionService.DeleteSubscription(email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription deleted successfully"})
}

// ExportEmails 导出邮箱列表（管理员）
// GET /api/v1/admin/subscriptions/export
func (h *Handler) ExportEmails(c *gin.Context) {
	tagsStr := c.Query("tags")

	var emails []string
	var err error

	if tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		emails, err = h.subscriptionService.GetActiveEmailsByTags(tags)
	} else {
		emails, err = h.subscriptionService.GetActiveEmails()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"emails": emails,
		"count":  len(emails),
	})
}
