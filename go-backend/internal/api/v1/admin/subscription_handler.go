package admin

import (
	"net/http"
	"strconv"
	"tanzanite/internal/repository"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	subscriptionRepo *repository.SubscriptionRepository
}

func NewSubscriptionHandler(subscriptionRepo *repository.SubscriptionRepository) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionRepo: subscriptionRepo,
	}
}

// ListSubscriptions 获取订阅列表
// GET /api/admin/subscriptions
func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	subscriptions, total, err := h.subscriptionRepo.FindAll(page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscriptions"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"subscriptions": subscriptions,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetSubscription 获取订阅详情
// GET /api/admin/subscriptions/:email
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	email := c.Param("email")

	subscription, err := h.subscriptionRepo.FindByEmail(email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscription": subscription,
	})
}

// UpdateSubscriptionStatus 更新订阅状态
// PATCH /api/admin/subscriptions/:email/status
func (h *SubscriptionHandler) UpdateSubscriptionStatus(c *gin.Context) {
	email := c.Param("email")

	var req struct {
		Status string `json:"status" binding:"required,oneof=active unsubscribed"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.subscriptionRepo.UpdateStatus(email, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription status updated successfully",
	})
}

// DeleteSubscription 删除订阅
// DELETE /api/admin/subscriptions/:email
func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	email := c.Param("email")

	if err := h.subscriptionRepo.Delete(email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription deleted successfully",
	})
}

// GetSubscriptionStats 获取订阅统计
// GET /api/admin/subscriptions/stats
func (h *SubscriptionHandler) GetSubscriptionStats(c *gin.Context) {
	stats, err := h.subscriptionRepo.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetActiveEmails 获取所有活跃订阅邮箱
// GET /api/admin/subscriptions/active-emails
func (h *SubscriptionHandler) GetActiveEmails(c *gin.Context) {
	emails, err := h.subscriptionRepo.GetActiveEmails()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active emails"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"emails": emails,
		"count":  len(emails),
	})
}

// BatchDelete 批量删除订阅
// POST /api/admin/subscriptions/batch-delete
func (h *SubscriptionHandler) BatchDelete(c *gin.Context) {
	var req struct {
		Emails []string `json:"emails" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deleted := 0
	for _, email := range req.Emails {
		if err := h.subscriptionRepo.Delete(email); err == nil {
			deleted++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch delete completed",
		"deleted": deleted,
		"total":   len(req.Emails),
	})
}
