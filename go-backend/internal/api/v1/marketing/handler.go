package marketing

import (
	"commerce-platform/internal/service"
	"fmt"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	marketingService *service.MarketingService
	settingService   *service.SettingService
	programService   *service.LoyaltyProgramService
	mediaResolver    service.PublicMediaURLResolver
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

func (h *Handler) ConfigureMediaService(resolver service.PublicMediaURLResolver) {
	if h == nil {
		return
	}
	h.mediaResolver = resolver
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

func (h *Handler) ListUserGiftCards(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "[CRITICAL] Unauthorized access"})
		return
	}

	giftCards, total, err := h.marketingService.ListUserGiftCards(userID.(uint), 1, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"gift_cards": publicGiftCardsFromDomain(giftCards, h.mediaResolver),
		"total":      total,
	})
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

func (h *Handler) ListRedeemGiftCardOptions(c *gin.Context) {
	if h.programService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "loyalty program service is unavailable"})
		return
	}

	config, err := h.programService.GetActive()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("[CRITICAL] Failed to load redeem config: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enabled":                 config.Enabled,
		"exchange_rate":           config.ExchangeRatePoints,
		"min_points":              config.MinRedeemPoints,
		"max_value_per_day":       float64(config.MaxValuePerDayCents) / 100,
		"max_value_per_day_cents": config.MaxValuePerDayCents,
		"card_expiry_days":        config.CardExpiryDays,
		"currency":                config.Currency,
		"items":                   h.marketingService.ListRedeemGiftCardOptionsFromConfig(config),
	})
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
		"redemption_exchange_rate":               config.ExchangeRatePoints,
	})
}

func (h *Handler) RedeemPointsToGiftCard(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "[CRITICAL] Unauthorized access"})
		return
	}

	var req struct {
		OptionID           uint    `json:"option_id"`
		GiftCardValueCents int64   `json:"giftcard_value_cents"`
		GiftCardValue      float64 `json:"giftcard_value"`
		IdempotencyKey     string  `json:"idempotency_key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("[CRITICAL] Invalid request arguments: %v", err)})
		return
	}

	if req.GiftCardValueCents <= 0 && req.GiftCardValue > 0 {
		req.GiftCardValueCents = int64(math.Round(req.GiftCardValue * 100))
	}
	idempotencyKey := c.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		idempotencyKey = req.IdempotencyKey
	}
	if h.programService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "loyalty program service is unavailable"})
		return
	}
	config, err := h.programService.GetActive()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("[CRITICAL] Failed to load redeem config: %v", err)})
		return
	}

	result, err := h.marketingService.RedeemPointsForGiftCard(userID.(uint), service.RedeemGiftCardRequest{
		OptionID:           req.OptionID,
		GiftCardValueCents: req.GiftCardValueCents,
		IdempotencyKey:     idempotencyKey,
	}, config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"giftcard_id":          result.GiftCardID,
		"card_code":            result.CardCode,
		"balance":              result.Balance,
		"balance_cents":        result.BalanceCents,
		"giftcard_value_cents": result.GiftCardValueCents,
		"redemption_id":        result.RedemptionID,
		"points_spent":         result.PointsSpent,
		"points_remaining":     result.PointsRemaining,
		"expires_at":           result.ExpiresAt,
		"message":              "redeemed successfully",
	})
}
