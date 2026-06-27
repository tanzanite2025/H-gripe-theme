package admin

import (
	"strconv"
	"tanzanite/internal/domain/loyalty"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/pagination"
	"tanzanite/internal/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

// ============ 积分交易管理 ============

// ListLoyaltyTransactions 获取积分交易列表
func (h *MarketingHandler) ListLoyaltyTransactions(c *gin.Context) {
	params := pagination.ParsePagination(c)
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)

	if userID > 0 {
		transactions, total, err := h.loyaltyRepo.FindTransactionsByUserID(uint(userID), params.Page, params.PageSize)
		if err != nil {
			apierror.RespondInternalError(c, err)
			return
		}

		response.Paged(c, gin.H{"transactions": transactions}, params.Page, params.PageSize, total)
	} else {
		// 需要在 Repository 中添加 FindAllTransactions 方法
		response.SuccessWithMessage(c, "请提供 user_id 参数", gin.H{
			"transactions": []loyalty.LoyaltyTransaction{},
			"total":        0,
			"page":         params.Page,
			"page_size":    params.PageSize,
		})
	}
}

// CreateLoyaltyTransaction 创建积分交易（管理员调整）
func (h *MarketingHandler) CreateLoyaltyTransaction(c *gin.Context) {
	var req struct {
		UserID      uint   `json:"user_id" binding:"required"`
		Points      int    `json:"points" binding:"required"`
		Description string `json:"description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	transaction, err := h.loyaltyRepo.AdjustUserPoints(req.UserID, req.Points, "adjust", "admin", 0, req.Description)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, gin.H{"transaction": transaction})
}

// ============ 签到管理 ============

// ListCheckIns 获取签到记录列表
func (h *MarketingHandler) ListCheckIns(c *gin.Context) {
	params := pagination.ParsePagination(c)
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)

	if userID == 0 {
		apierror.RespondBadRequest(c, "请提供 user_id 参数")
		return
	}

	checkIns, total, err := h.loyaltyRepo.FindCheckInsByUserID(uint(userID), params.Page, params.PageSize)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Paged(c, gin.H{"check_ins": checkIns}, params.Page, params.PageSize, total)
}

// ============ 推荐管理 ============

// ListReferrals 获取推荐记录列表
func (h *MarketingHandler) ListReferrals(c *gin.Context) {
	referrerID, _ := strconv.ParseUint(c.Query("referrer_id"), 10, 32)

	if referrerID == 0 {
		apierror.RespondBadRequest(c, "请提供 referrer_id 参数")
		return
	}

	referrals, err := h.loyaltyRepo.FindReferralsByReferrerID(uint(referrerID))
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"referrals": referrals})
}

// UpdateReferralStatus 更新推荐状态
func (h *MarketingHandler) UpdateReferralStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "无效的推荐ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=pending completed expired"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	referral, err := h.loyaltyRepo.FindReferralByID(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "推荐记录")
		return
	}

	referral.Status = req.Status
	if req.Status == "completed" && referral.CompletedAt == nil {
		now := time.Now()
		referral.CompletedAt = &now
	}

	if err := h.loyaltyRepo.UpdateReferral(referral); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"referral": referral})
}
