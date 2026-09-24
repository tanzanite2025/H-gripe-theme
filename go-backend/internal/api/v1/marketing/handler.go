package marketing

import (
	"commerce-platform/internal/service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	marketingService *service.MarketingService
	settingService   *service.SettingService
	programService   *service.LoyaltyProgramService
}

func NewHandler(marketingService *service.MarketingService, settingService *service.SettingService, programServices ...*service.LoyaltyProgramService) *Handler {
	handler := &Handler{
		marketingService: marketingService,
		settingService:   settingService,
	}
	if len(programServices) > 0 {
		handler.programService = programServices[0]
	}
	return handler
}

func (h *Handler) ListCoupons(c *gin.Context) {
	coupons, err := h.marketingService.GetActiveCoupons()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": coupons})
}

func (h *Handler) ValidateCoupon(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Code        string `json:"code" binding:"required"`
		AmountMinor int64  `json:"amount_minor" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customerEmail, _ := c.Get("email")
	email, _ := customerEmail.(string)
	coupon, discountMinor, err := h.marketingService.ValidateCoupon(req.Code, userID.(uint), req.AmountMinor, email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":          true,
		"coupon":         coupon,
		"discount_minor": discountMinor,
	})
}

func (h *Handler) GetPoints(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	loyalty, err := h.marketingService.GetUserLoyalty(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "loyalty info not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"current_points": loyalty.AvailablePoints,
		"total_points":   loyalty.TotalPoints,
		"level_id":       loyalty.MemberLevelID,
	})
}

func (h *Handler) CheckIn(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	points, err := h.marketingService.CheckIn(userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "check-in successful",
		"points":  points,
	})
}

func (h *Handler) CreateReferral(c *gin.Context) {
	c.JSON(http.StatusGone, gin.H{
		"code":  "referral_legacy_endpoint_retired",
		"error": "This referral endpoint is retired; use the referral-code flow when the new program is enabled.",
	})
}

func (h *Handler) GetLoyaltyInfo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	loyalty, err := h.marketingService.GetUserLoyalty(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "loyalty info not found"})
		return
	}

	c.JSON(http.StatusOK, loyalty)
}

func (h *Handler) ListMemberLevels(c *gin.Context) {
	levels, err := h.marketingService.ListMemberLevels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, levels)
}

func (h *Handler) GetLoyaltyProgramConfig(c *gin.Context) {
	if h.programService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "loyalty program service is unavailable"})
		return
	}

	config, err := h.programService.GetPublicConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("[CRITICAL] Failed to load loyalty program config: %v", err)})
		return
	}

	c.JSON(http.StatusOK, config)
}

func (h *Handler) GetLoyaltyRules(c *gin.Context) {
	if h.programService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "loyalty program service is unavailable"})
		return
	}
	config, err := h.programService.GetActive()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("[CRITICAL] Failed to load loyalty rules: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"version":                                config.Version,
		"currency":                               service.LoyaltyPointsBaseCurrency,
		"points_base_currency":                   service.LoyaltyPointsBaseCurrency,
		"purchase_earn_points_per_currency_unit": config.PurchaseEarnPointsPerUnit,
		"purchase_earn_trigger":                  "order_completed",
		"purchase_earn_amount_basis":             "order_subtotal_minus_discounts",
		"referral_referrer_points":               config.ReferralReferrerPoints,
		"referral_referee_points":                config.ReferralRefereePoints,
		"checkin_base_points":                    config.CheckInBasePoints,
		"checkin_streak_interval_days":           config.CheckInStreakIntervalDays,
		"checkin_streak_bonus_points":            config.CheckInStreakBonusPoints,
		"checkin_max_points":                     config.CheckInMaxPoints,
	})
}
