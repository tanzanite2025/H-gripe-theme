package admin

import (
	"strconv"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/pagination"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *MarketingHandler) ListLoyaltyTransactions(c *gin.Context) {
	params := pagination.ParsePagination(c)
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)

	if userID == 0 {
		response.SuccessWithMessage(c, "please provide user_id", gin.H{
			"transactions": []interface{}{},
			"total":        0,
			"page":         params.Page,
			"page_size":    params.PageSize,
		})
		return
	}

	transactions, total, err := h.marketingService.ListLoyaltyTransactions(uint(userID), params.Page, params.PageSize)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Paged(c, gin.H{"transactions": transactions}, params.Page, params.PageSize, total)
}

func (h *MarketingHandler) CreateLoyaltyTransaction(c *gin.Context) {
	var req struct {
		UserID      uint   `json:"user_id" binding:"required"`
		Points      int    `json:"points" binding:"required"`
		Description string `json:"description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	transaction, err := h.marketingService.AdminAdjustPointsWithTransaction(req.UserID, req.Points, req.Description)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, gin.H{"transaction": transaction})
}

func (h *MarketingHandler) ListCheckIns(c *gin.Context) {
	params := pagination.ParsePagination(c)
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)

	if userID == 0 {
		apierror.RespondBadRequest(c, "please provide user_id")
		return
	}

	checkIns, total, err := h.marketingService.ListCheckIns(uint(userID), params.Page, params.PageSize)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Paged(c, gin.H{"check_ins": checkIns}, params.Page, params.PageSize, total)
}

func (h *MarketingHandler) ListReferrals(c *gin.Context) {
	referrerID, _ := strconv.ParseUint(c.Query("referrer_id"), 10, 32)
	if referrerID == 0 {
		apierror.RespondBadRequest(c, "please provide referrer_id")
		return
	}

	referrals, err := h.marketingService.ListReferrals(uint(referrerID))
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"referrals": referrals})
}

func (h *MarketingHandler) UpdateReferralStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid referral ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=pending completed expired"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	referral, err := h.marketingService.UpdateReferralStatus(uint(id), req.Status)
	if err != nil {
		respondMarketingError(c, err, "referral")
		return
	}

	response.Success(c, gin.H{"referral": referral})
}
