package admin

import (
	"net/http"
	"strconv"
	"tanzanite/internal/domain/coupon"
	"tanzanite/internal/domain/loyalty"
	"tanzanite/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
)

// MarketingHandler 营销管理处理器
type MarketingHandler struct {
	couponRepo  *repository.CouponRepository
	loyaltyRepo *repository.LoyaltyRepository
}

// NewMarketingHandler 创建营销管理处理器
func NewMarketingHandler(couponRepo *repository.CouponRepository, loyaltyRepo *repository.LoyaltyRepository) *MarketingHandler {
	return &MarketingHandler{
		couponRepo:  couponRepo,
		loyaltyRepo: loyaltyRepo,
	}
}

// ============ 优惠券管理 ============

// ListCoupons 获取优惠券列表
func (h *MarketingHandler) ListCoupons(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status") // all, active, expired, disabled

	coupons, total, err := h.couponRepo.FindAllCoupons(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取优惠券列表失败"})
		return
	}

	// 根据状态筛选
	if status != "" && status != "all" {
		filtered := make([]coupon.Coupon, 0)
		now := time.Now()
		for _, cp := range coupons {
			switch status {
			case "active":
				if cp.Enabled && now.After(cp.StartDate) && now.Before(cp.EndDate) {
					filtered = append(filtered, cp)
				}
			case "expired":
				if now.After(cp.EndDate) {
					filtered = append(filtered, cp)
				}
			case "disabled":
				if !cp.Enabled {
					filtered = append(filtered, cp)
				}
			}
		}
		coupons = filtered
		total = int64(len(filtered))
	}

	c.JSON(http.StatusOK, gin.H{
		"coupons": coupons,
		"total":   total,
		"page":    page,
		"page_size": pageSize,
	})
}

// GetCoupon 获取优惠券详情
func (h *MarketingHandler) GetCoupon(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的优惠券ID"})
		return
	}

	cp, err := h.couponRepo.FindCouponByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "优惠券不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"coupon": cp})
}

// CreateCoupon 创建优惠券
func (h *MarketingHandler) CreateCoupon(c *gin.Context) {
	var req struct {
		Code                 string    `json:"code" binding:"required"`
		Type                 string    `json:"type" binding:"required,oneof=fixed percentage"`
		Value                float64   `json:"value" binding:"required,gt=0"`
		Description          string    `json:"description"`
		MinAmount            float64   `json:"min_amount"`
		MaxDiscount          float64   `json:"max_discount"`
		UsageLimit           int       `json:"usage_limit"`
		UsageLimitPerUser    int       `json:"usage_limit_per_user"`
		StartDate            time.Time `json:"start_date" binding:"required"`
		EndDate              time.Time `json:"end_date" binding:"required"`
		ApplicableProducts   string    `json:"applicable_products"`
		ExcludedProducts     string    `json:"excluded_products"`
		ApplicableCategories string    `json:"applicable_categories"`
		Enabled              bool      `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cp := &coupon.Coupon{
		Code:                 req.Code,
		Type:                 req.Type,
		Value:                req.Value,
		Description:          req.Description,
		MinAmount:            req.MinAmount,
		MaxDiscount:          req.MaxDiscount,
		UsageLimit:           req.UsageLimit,
		UsageLimitPerUser:    req.UsageLimitPerUser,
		StartDate:            req.StartDate,
		EndDate:              req.EndDate,
		ApplicableProducts:   req.ApplicableProducts,
		ExcludedProducts:     req.ExcludedProducts,
		ApplicableCategories: req.ApplicableCategories,
		Enabled:              req.Enabled,
	}

	if err := h.couponRepo.CreateCoupon(cp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建优惠券失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"coupon": cp})
}

// UpdateCoupon 更新优惠券
func (h *MarketingHandler) UpdateCoupon(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的优惠券ID"})
		return
	}

	cp, err := h.couponRepo.FindCouponByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "优惠券不存在"})
		return
	}

	var req struct {
		Code                 string    `json:"code"`
		Type                 string    `json:"type" binding:"omitempty,oneof=fixed percentage"`
		Value                float64   `json:"value"`
		Description          string    `json:"description"`
		MinAmount            float64   `json:"min_amount"`
		MaxDiscount          float64   `json:"max_discount"`
		UsageLimit           int       `json:"usage_limit"`
		UsageLimitPerUser    int       `json:"usage_limit_per_user"`
		StartDate            time.Time `json:"start_date"`
		EndDate              time.Time `json:"end_date"`
		ApplicableProducts   string    `json:"applicable_products"`
		ExcludedProducts     string    `json:"excluded_products"`
		ApplicableCategories string    `json:"applicable_categories"`
		Enabled              bool      `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新字段
	if req.Code != "" {
		cp.Code = req.Code
	}
	if req.Type != "" {
		cp.Type = req.Type
	}
	if req.Value > 0 {
		cp.Value = req.Value
	}
	cp.Description = req.Description
	cp.MinAmount = req.MinAmount
	cp.MaxDiscount = req.MaxDiscount
	cp.UsageLimit = req.UsageLimit
	cp.UsageLimitPerUser = req.UsageLimitPerUser
	if !req.StartDate.IsZero() {
		cp.StartDate = req.StartDate
	}
	if !req.EndDate.IsZero() {
		cp.EndDate = req.EndDate
	}
	cp.ApplicableProducts = req.ApplicableProducts
	cp.ExcludedProducts = req.ExcludedProducts
	cp.ApplicableCategories = req.ApplicableCategories
	cp.Enabled = req.Enabled

	if err := h.couponRepo.UpdateCoupon(cp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新优惠券失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"coupon": cp})
}

// DeleteCoupon 删除优惠券
func (h *MarketingHandler) DeleteCoupon(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的优惠券ID"})
		return
	}

	if err := h.couponRepo.DeleteCoupon(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除优惠券失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetCouponStats 获取优惠券统计
func (h *MarketingHandler) GetCouponStats(c *gin.Context) {
	coupons, _, err := h.couponRepo.FindAllCoupons(1, 1000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计失败"})
		return
	}

	now := time.Now()
	stats := gin.H{
		"total":    len(coupons),
		"active":   0,
		"expired":  0,
		"disabled": 0,
		"used":     0,
	}

	totalUsed := 0
	for _, cp := range coupons {
		if cp.Enabled && now.After(cp.StartDate) && now.Before(cp.EndDate) {
			stats["active"] = stats["active"].(int) + 1
		} else if now.After(cp.EndDate) {
			stats["expired"] = stats["expired"].(int) + 1
		} else if !cp.Enabled {
			stats["disabled"] = stats["disabled"].(int) + 1
		}
		totalUsed += cp.UsedCount
	}
	stats["used"] = totalUsed

	c.JSON(http.StatusOK, stats)
}

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

// ============ 会员等级管理 ============

// ListMemberLevels 获取会员等级列表
func (h *MarketingHandler) ListMemberLevels(c *gin.Context) {
	levels, err := h.loyaltyRepo.FindAllMemberLevels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取会员等级失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"levels": levels})
}

// GetMemberLevel 获取会员等级详情
func (h *MarketingHandler) GetMemberLevel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的等级ID"})
		return
	}

	level, err := h.loyaltyRepo.FindMemberLevelByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会员等级不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"level": level})
}

// CreateMemberLevel 创建会员等级
func (h *MarketingHandler) CreateMemberLevel(c *gin.Context) {
	var req struct {
		Name             string  `json:"name" binding:"required"`
		MinPoints        int     `json:"min_points" binding:"required"`
		MaxPoints        int     `json:"max_points" binding:"required"`
		DiscountRate     float64 `json:"discount_rate"`
		PointsMultiplier float64 `json:"points_multiplier"`
		Benefits         string  `json:"benefits"`
		Icon             string  `json:"icon"`
		Color            string  `json:"color"`
		SortOrder        int     `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	level := &loyalty.MemberLevel{
		Name:             req.Name,
		MinPoints:        req.MinPoints,
		MaxPoints:        req.MaxPoints,
		DiscountRate:     req.DiscountRate,
		PointsMultiplier: req.PointsMultiplier,
		Benefits:         req.Benefits,
		Icon:             req.Icon,
		Color:            req.Color,
		SortOrder:        req.SortOrder,
	}

	if err := h.loyaltyRepo.CreateMemberLevel(level); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建会员等级失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"level": level})
}

// UpdateMemberLevel 更新会员等级
func (h *MarketingHandler) UpdateMemberLevel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的等级ID"})
		return
	}

	level, err := h.loyaltyRepo.FindMemberLevelByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会员等级不存在"})
		return
	}

	var req struct {
		Name             string  `json:"name"`
		MinPoints        int     `json:"min_points"`
		MaxPoints        int     `json:"max_points"`
		DiscountRate     float64 `json:"discount_rate"`
		PointsMultiplier float64 `json:"points_multiplier"`
		Benefits         string  `json:"benefits"`
		Icon             string  `json:"icon"`
		Color            string  `json:"color"`
		SortOrder        int     `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新字段
	if req.Name != "" {
		level.Name = req.Name
	}
	if req.MinPoints > 0 {
		level.MinPoints = req.MinPoints
	}
	if req.MaxPoints > 0 {
		level.MaxPoints = req.MaxPoints
	}
	level.DiscountRate = req.DiscountRate
	level.PointsMultiplier = req.PointsMultiplier
	level.Benefits = req.Benefits
	level.Icon = req.Icon
	level.Color = req.Color
	level.SortOrder = req.SortOrder

	if err := h.loyaltyRepo.UpdateMemberLevel(level); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新会员等级失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"level": level})
}

// DeleteMemberLevel 删除会员等级
func (h *MarketingHandler) DeleteMemberLevel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的等级ID"})
		return
	}

	if err := h.loyaltyRepo.DeleteMemberLevel(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除会员等级失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ============ 营销统计 ============

// GetMarketingStats 获取营销统计
func (h *MarketingHandler) GetMarketingStats(c *gin.Context) {
	stats := gin.H{}

	// 优惠券统计
	coupons, _, _ := h.couponRepo.FindAllCoupons(1, 1000)
	now := time.Now()
	couponStats := gin.H{
		"total":  len(coupons),
		"active": 0,
		"used":   0,
	}
	totalUsed := 0
	for _, cp := range coupons {
		if cp.Enabled && now.After(cp.StartDate) && now.Before(cp.EndDate) {
			couponStats["active"] = couponStats["active"].(int) + 1
		}
		totalUsed += cp.UsedCount
	}
	couponStats["used"] = totalUsed
	stats["coupons"] = couponStats

	// 积分统计
	loyaltyStats, _ := h.loyaltyRepo.GetLoyaltyStats()
	stats["loyalty"] = loyaltyStats

	c.JSON(http.StatusOK, stats)
}
