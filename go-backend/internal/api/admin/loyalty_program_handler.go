package admin

import (
	"net/http"

	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/service"

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
		Enabled                   bool    `json:"enabled"`
		Currency                  string  `json:"currency" binding:"required"`
		ExchangeRatePoints        int     `json:"exchange_rate_points" binding:"required,gt=0"`
		MinRedeemPoints           int     `json:"min_redeem_points" binding:"gte=0"`
		MaxValuePerDayCents       int64   `json:"max_value_per_day_cents" binding:"gte=0"`
		CardExpiryDays            int     `json:"card_expiry_days" binding:"gte=0"`
		ReferralReferrerPoints    int     `json:"referral_referrer_points" binding:"gte=0"`
		ReferralRefereePoints     int     `json:"referral_referee_points" binding:"gte=0"`
		CheckInBasePoints         int     `json:"checkin_base_points" binding:"gte=0"`
		CheckInStreakIntervalDays int     `json:"checkin_streak_interval_days" binding:"gt=0"`
		CheckInStreakBonusPoints  int     `json:"checkin_streak_bonus_points" binding:"gte=0"`
		CheckInMaxPoints          int     `json:"checkin_max_points" binding:"gte=0"`
		RedeemValuesCents         []int64 `json:"redeem_values_cents" binding:"required,min=1"`
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

	config, err := h.programService.Update(service.LoyaltyProgramConfigInput{
		Enabled:                   req.Enabled,
		Currency:                  req.Currency,
		ExchangeRatePoints:        req.ExchangeRatePoints,
		MinRedeemPoints:           req.MinRedeemPoints,
		MaxValuePerDayCents:       req.MaxValuePerDayCents,
		CardExpiryDays:            req.CardExpiryDays,
		ReferralReferrerPoints:    req.ReferralReferrerPoints,
		ReferralRefereePoints:     req.ReferralRefereePoints,
		CheckInBasePoints:         req.CheckInBasePoints,
		CheckInStreakIntervalDays: req.CheckInStreakIntervalDays,
		CheckInStreakBonusPoints:  req.CheckInStreakBonusPoints,
		CheckInMaxPoints:          req.CheckInMaxPoints,
		RedeemValuesCents:         req.RedeemValuesCents,
		CreatedBy:                 createdBy,
	})
	if err != nil {
		respondMarketingError(c, err, "loyalty program config")
		return
	}

	c.JSON(http.StatusOK, gin.H{"config": config})
}
