package marketing

import (
	"fmt"
	"net/http"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	marketingService *service.MarketingService
	settingService   *service.SettingService
}

func NewHandler(marketingService *service.MarketingService, settingService *service.SettingService) *Handler {
	return &Handler{
		marketingService: marketingService,
		settingService:   settingService,
	}
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
		Code   string  `json:"code" binding:"required"`
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coupon, discount, err := h.marketingService.ValidateCoupon(req.Code, userID.(uint), req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":    true,
		"coupon":   coupon,
		"discount": discount,
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
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		RefereeID uint `json:"referee_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.marketingService.CreateReferral(userID.(uint), req.RefereeID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "referral created"})
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

func (h *Handler) GetUserAssets(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "[CRITICAL] Unauthorized access"})
		return
	}

	redeemedGiftCards, err := h.marketingService.CountRedeemedGiftCards(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"coupons":     0,
		"point_cards": redeemedGiftCards,
	})
}

func (h *Handler) ListRedeemGiftCardOptions(c *gin.Context) {
	locale := c.GetHeader("X-Locale")
	if locale == "" {
		locale = c.DefaultQuery("lang", "en")
	}

	config, err := h.settingService.GetRedeemSettings(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("[CRITICAL] Failed to load redeem settings: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enabled":           config.Enabled,
		"exchange_rate":     config.ExchangeRate,
		"min_points":        config.MinPoints,
		"max_value_per_day": config.MaxValuePerDay,
		"card_expiry_days":  config.CardExpiryDays,
		"items":             h.marketingService.ListRedeemGiftCardOptions(config),
	})
}

func (h *Handler) RedeemPointsToGiftCard(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "[CRITICAL] Unauthorized access"})
		return
	}

	var req struct {
		GiftCardValue float64 `json:"giftcard_value" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("[CRITICAL] Invalid request arguments: %v", err)})
		return
	}

	locale := c.GetHeader("X-Locale")
	if locale == "" {
		locale = c.DefaultQuery("lang", "en")
	}
	config, err := h.settingService.GetRedeemSettings(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("[CRITICAL] Failed to load redeem settings: %v", err)})
		return
	}

	result, err := h.marketingService.RedeemPointsForGiftCard(userID.(uint), req.GiftCardValue, config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"giftcard_id":      result.GiftCardID,
		"card_code":        result.CardCode,
		"balance":          result.Balance,
		"points_spent":     result.PointsSpent,
		"points_remaining": result.PointsRemaining,
		"expires_at":       result.ExpiresAt,
		"message":          "redeemed successfully",
	})
}
