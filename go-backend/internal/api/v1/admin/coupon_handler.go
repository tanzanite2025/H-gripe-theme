package admin

import (
	"net/http"
	"strconv"
	"tanzanite/internal/domain/coupon"
	"time"

	"github.com/gin-gonic/gin"
)

// ListCoupons 获取优惠券列表
// @Summary 获取优惠券列表
// @Tags Admin-Marketing
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param status query string false "状态"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/marketing/coupons [get]
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
		"coupons":   coupons,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetCoupon 获取优惠券详情
// @Summary 获取优惠券详情
// @Tags Admin-Marketing
// @Produce json
// @Param id path int true "优惠券ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/marketing/coupons/{id} [get]
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
// @Summary 创建优惠券
// @Tags Admin-Marketing
// @Accept json
// @Produce json
// @Param coupon body coupon.Coupon true "优惠券信息"
// @Success 201 {object} map[string]interface{}
// @Router /api/admin/marketing/coupons [post]
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
// @Summary 更新优惠券
// @Tags Admin-Marketing
// @Accept json
// @Produce json
// @Param id path int true "优惠券ID"
// @Param coupon body coupon.Coupon true "优惠券信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/marketing/coupons/{id} [put]
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
// @Summary 删除优惠券
// @Tags Admin-Marketing
// @Produce json
// @Param id path int true "优惠券ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/marketing/coupons/{id} [delete]
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
// @Summary 获取优惠券统计
// @Tags Admin-Marketing
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/marketing/coupons/stats [get]
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
