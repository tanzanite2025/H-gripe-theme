package admin

import (
	"net/http"
	"strconv"
	"tanzanite/internal/domain/loyalty"
	"time"

	"github.com/gin-gonic/gin"
)

// ============ 积分交易管理 ============

// ListLoyaltyTransactions 获取积分交易列表
func (h *MarketingHandler) ListLoyaltyTransactions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)

	if userID > 0 {
		transactions, total, err := h.loyaltyRepo.FindTransactionsByUserID(uint(userID), page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取交易记录失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"transactions": transactions,
			"total":        total,
			"page":         page,
			"page_size":    pageSize,
		})
	} else {
		// 需要在 Repository 中添加 FindAllTransactions 方法
		c.JSON(http.StatusOK, gin.H{
			"transactions": []loyalty.LoyaltyTransaction{},
			"total":        0,
			"page":         page,
			"page_size":    pageSize,
			"message":      "请提供 user_id 参数",
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前余额
	balance, err := h.loyaltyRepo.GetUserPointsBalance(req.UserID)
	if err != nil {
		balance = 0
	}

	// 创建交易记录
	transaction := &loyalty.LoyaltyTransaction{
		UserID:      req.UserID,
		Type:        "adjust",
		Points:      req.Points,
		Balance:     balance + req.Points,
		Source:      "admin",
		Description: req.Description,
	}

	if err := h.loyaltyRepo.CreateTransaction(transaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建交易记录失败"})
		return
	}

	// 更新用户积分
	userLoyalty, err := h.loyaltyRepo.FindUserLoyaltyByUserID(req.UserID)
	if err != nil {
		// 如果用户没有积分记录，创建一个
		userLoyalty = &loyalty.UserLoyalty{
			UserID:          req.UserID,
			TotalPoints:     req.Points,
			AvailablePoints: req.Points,
		}
		h.loyaltyRepo.CreateUserLoyalty(userLoyalty)
	} else {
		if req.Points > 0 {
			userLoyalty.TotalPoints += req.Points
			userLoyalty.AvailablePoints += req.Points
		} else {
			userLoyalty.UsedPoints += -req.Points
			userLoyalty.AvailablePoints += req.Points
		}
		h.loyaltyRepo.UpdateUserLoyalty(userLoyalty)
	}

	c.JSON(http.StatusCreated, gin.H{"transaction": transaction})
}

// ============ 签到管理 ============

// ListCheckIns 获取签到记录列表
func (h *MarketingHandler) ListCheckIns(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)

	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供 user_id 参数"})
		return
	}

	checkIns, total, err := h.loyaltyRepo.FindCheckInsByUserID(uint(userID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取签到记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"check_ins": checkIns,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ============ 推荐管理 ============

// ListReferrals 获取推荐记录列表
func (h *MarketingHandler) ListReferrals(c *gin.Context) {
	referrerID, _ := strconv.ParseUint(c.Query("referrer_id"), 10, 32)

	if referrerID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供 referrer_id 参数"})
		return
	}

	referrals, err := h.loyaltyRepo.FindReferralsByReferrerID(uint(referrerID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取推荐记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"referrals": referrals})
}

// UpdateReferralStatus 更新推荐状态
func (h *MarketingHandler) UpdateReferralStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的推荐ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=pending completed expired"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	referral, err := h.loyaltyRepo.FindReferralByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "推荐记录不存在"})
		return
	}

	referral.Status = req.Status
	if req.Status == "completed" && referral.CompletedAt == nil {
		now := time.Now()
		referral.CompletedAt = &now
	}

	if err := h.loyaltyRepo.UpdateReferral(referral); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新推荐状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"referral": referral})
}
