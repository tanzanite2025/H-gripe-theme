package admin

import (
	"strconv"
	"tanzanite/internal/domain/coupon"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/pagination"
	"tanzanite/internal/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

// ============ 礼品卡管理 ============

// ListGiftCards 获取礼品卡列表
func (h *MarketingHandler) ListGiftCards(c *gin.Context) {
	params := pagination.ParsePagination(c)
	_ = c.Query("status") // status 参数暂时未使用

	// 注意：需要在 CouponRepository 中添加 FindAllGiftCards 方法
	// 这里暂时返回空列表
	response.SuccessWithMessage(c, "礼品卡列表功能需要在 Repository 中添加 FindAllGiftCards 方法", gin.H{
		"gift_cards": []coupon.GiftCard{},
		"total":      0,
		"page":       params.Page,
		"page_size":  params.PageSize,
	})
}

// GetGiftCard 获取礼品卡详情
func (h *MarketingHandler) GetGiftCard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "无效的礼品卡ID")
		return
	}

	gc, err := h.couponRepo.FindGiftCardByID(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "礼品卡")
		return
	}

	// 获取交易记录
	transactions, _ := h.couponRepo.FindGiftCardTransactionsByCardID(uint(id))

	response.Success(c, gin.H{
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
		apierror.RespondBadRequest(c, err.Error())
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
		apierror.RespondInternalError(c, err)
		return
	}

	response.Created(c, gin.H{"gift_card": gc})
}

// UpdateGiftCardStatus 更新礼品卡状态
func (h *MarketingHandler) UpdateGiftCardStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "无效的礼品卡ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active used expired cancelled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	gc, err := h.couponRepo.FindGiftCardByID(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "礼品卡")
		return
	}

	gc.Status = req.Status
	if err := h.couponRepo.UpdateGiftCard(gc); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"gift_card": gc})
}
