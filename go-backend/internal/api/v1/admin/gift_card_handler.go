package admin

import (
	"net/http"
	"strconv"
	"tanzanite/internal/domain/coupon"
	"time"

	"github.com/gin-gonic/gin"
)

// ============ 礼品卡管理 ============

// ListGiftCards 获取礼品卡列表
func (h *MarketingHandler) ListGiftCards(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	_ = c.Query("status") // status 参数暂时未使用

	// 注意：需要在 CouponRepository 中添加 FindAllGiftCards 方法
	// 这里暂时返回空列表
	c.JSON(http.StatusOK, gin.H{
		"gift_cards": []coupon.GiftCard{},
		"total":      0,
		"page":       page,
		"page_size":  pageSize,
		"message":    "礼品卡列表功能需要在 Repository 中添加 FindAllGiftCards 方法",
	})
}

// GetGiftCard 获取礼品卡详情
func (h *MarketingHandler) GetGiftCard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的礼品卡ID"})
		return
	}

	gc, err := h.couponRepo.FindGiftCardByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "礼品卡不存在"})
		return
	}

	// 获取交易记录
	transactions, _ := h.couponRepo.FindGiftCardTransactionsByCardID(uint(id))

	c.JSON(http.StatusOK, gin.H{
		"gift_card":    gc,
		"transactions": transactions,
	})
}

// CreateGiftCard 创建礼品卡
func (h *MarketingHandler) CreateGiftCard(c *gin.Context) {
	var req struct {
		Code           string     `json:"code" binding:"required"`
		InitialValue   float64    `json:"initial_value" binding:"required,gt=0"`
		Currency       string     `json:"currency"`
		RecipientEmail string     `json:"recipient_email"`
		RecipientName  string     `json:"recipient_name"`
		SenderName     string     `json:"sender_name"`
		Message        string     `json:"message"`
		CoverImage     string     `json:"cover_image"`
		ExpiresAt      *time.Time `json:"expires_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	gc := &coupon.GiftCard{
		Code:           req.Code,
		InitialValue:   req.InitialValue,
		Balance:        req.InitialValue,
		Currency:       req.Currency,
		Status:         "active",
		RecipientEmail: req.RecipientEmail,
		RecipientName:  req.RecipientName,
		SenderName:     req.SenderName,
		Message:        req.Message,
		CoverImage:     req.CoverImage,
		ExpiresAt:      req.ExpiresAt,
	}

	if gc.Currency == "" {
		gc.Currency = "USD"
	}

	if err := h.couponRepo.CreateGiftCard(gc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建礼品卡失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"gift_card": gc})
}

// UpdateGiftCardStatus 更新礼品卡状态
func (h *MarketingHandler) UpdateGiftCardStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的礼品卡ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active used expired cancelled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	gc, err := h.couponRepo.FindGiftCardByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "礼品卡不存在"})
		return
	}

	gc.Status = req.Status
	if err := h.couponRepo.UpdateGiftCard(gc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新礼品卡状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"gift_card": gc})
}
