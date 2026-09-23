package admin

import (
	"net/http"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *MarketingHandler) GetLoyaltyProgramConfig(c *gin.Context) {
	if h.programService == nil {
		apierror.RespondInternalError(c, service.ErrLoyaltyProgramConfigNotFound)
		return
	}

	config, err := h.programService.GetPublicConfig()
	if err != nil {
		respondMarketingError(c, err, "loyalty program config")
		return
	}
	c.JSON(http.StatusOK, gin.H{"config": config})
}

func (h *MarketingHandler) UpdateLoyaltyProgramConfig(c *gin.Context) {
	if h.programService == nil {
		apierror.RespondInternalError(c, service.ErrLoyaltyProgramConfigNotFound)
		return
	}

	var req struct {
		Enabled                   bool   `json:"enabled"`
		Currency                  string `json:"currency" binding:"required"`
		PurchaseEarnPointsPerUnit int    `json:"purchase_earn_points_per_currency_unit" binding:"gte=0"`
		ExchangeRatePoints        int    `json:"exchange_rate_points" binding:"required,gt=0"`
		ReferralReferrerPoints    int    `json:"referral_referrer_points" binding:"gte=0"`
		ReferralRefereePoints     int    `json:"referral_referee_points" binding:"gte=0"`
		CheckInBasePoints         int    `json:"checkin_base_points" binding:"gte=0"`
		CheckInStreakIntervalDays int    `json:"checkin_streak_interval_days" binding:"gt=0"`
		CheckInStreakBonusPoints  int    `json:"checkin_streak_bonus_points" binding:"gte=0"`
		CheckInMaxPoints          int    `json:"checkin_max_points" binding:"gte=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	var createdBy *uint
	if userID, exists := c.Get("user_id"); exists {
		if value, ok := userID.(uint); ok {
			createdBy = &value
		}
	}

	if _, err := h.programService.Update(service.LoyaltyProgramConfigInput{
		Enabled:                   req.Enabled,
		Currency:                  req.Currency,
		PurchaseEarnPointsPerUnit: req.PurchaseEarnPointsPerUnit,
		ExchangeRatePoints:        req.ExchangeRatePoints,
		ReferralReferrerPoints:    req.ReferralReferrerPoints,
		ReferralRefereePoints:     req.ReferralRefereePoints,
		CheckInBasePoints:         req.CheckInBasePoints,
		CheckInStreakIntervalDays: req.CheckInStreakIntervalDays,
		CheckInStreakBonusPoints:  req.CheckInStreakBonusPoints,
		CheckInMaxPoints:          req.CheckInMaxPoints,
		CreatedBy:                 createdBy,
	}); err != nil {
		respondMarketingError(c, err, "loyalty program config")
		return
	}

	config, err := h.programService.GetPublicConfig()
	if err != nil {
		respondMarketingError(c, err, "loyalty program config")
		return
	}

	c.JSON(http.StatusOK, gin.H{"config": config})
}
