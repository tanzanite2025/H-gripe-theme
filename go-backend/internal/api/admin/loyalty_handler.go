package admin

import (
	"bytes"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/pagination"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
	"encoding/csv"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *MarketingHandler) ListReferralLedger(c *gin.Context) {
	if h == nil || h.referralService == nil {
		apierror.RespondInternalError(c, service.ErrReferralServiceUnavailable)
		return
	}
	params := pagination.ParsePagination(c)
	ledger, err := h.referralService.AdminLedger(repository.ReferralAdminFilters{
		Status: strings.TrimSpace(c.Query("status")), Keyword: strings.TrimSpace(c.Query("keyword")),
	}, params.Page, params.PageSize, time.Now().UTC())
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Paged(c, ledger, params.Page, params.PageSize, ledger.Total)
}

func (h *MarketingHandler) GetReferralDetail(c *gin.Context) {
	if h == nil || h.referralService == nil {
		apierror.RespondInternalError(c, service.ErrReferralServiceUnavailable)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		apierror.RespondBadRequest(c, "invalid referral id")
		return
	}
	detail, err := h.referralService.AdminDetail(uint(id), time.Now().UTC())
	if err != nil {
		if repository.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "referral")
		} else {
			apierror.RespondInternalError(c, err)
		}
		return
	}
	response.Success(c, detail)
}

func (h *MarketingHandler) SettleReferral(c *gin.Context) {
	if h == nil || h.referralService == nil {
		apierror.RespondInternalError(c, service.ErrReferralServiceUnavailable)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		apierror.RespondBadRequest(c, "invalid referral id")
		return
	}
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}
	var actorID *uint
	if value, ok := c.Get("user_id"); ok {
		if userID, ok := value.(uint); ok && userID > 0 {
			actorID = &userID
		}
	}
	result, err := h.referralService.SettleReferral(uint(id), req.Reason, actorID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrReferralReasonRequired), errors.Is(err, service.ErrReferralActionInvalid):
			apierror.RespondBadRequest(c, err.Error())
		case repository.IsRecordNotFound(err):
			apierror.RespondNotFound(c, "referral")
		default:
			apierror.RespondInternalError(c, err)
		}
		return
	}
	response.Success(c, result)
}

func (h *MarketingHandler) RevokeReferral(c *gin.Context) {
	if h == nil || h.referralService == nil {
		apierror.RespondInternalError(c, service.ErrReferralServiceUnavailable)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		apierror.RespondBadRequest(c, "invalid referral id")
		return
	}
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}
	var actorID *uint
	if value, ok := c.Get("user_id"); ok {
		if userID, ok := value.(uint); ok && userID > 0 {
			actorID = &userID
		}
	}
	result, err := h.referralService.RevokeReferral(uint(id), req.Reason, actorID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrReferralReasonRequired), errors.Is(err, service.ErrReferralActionInvalid):
			apierror.RespondBadRequest(c, err.Error())
		case repository.IsRecordNotFound(err):
			apierror.RespondNotFound(c, "referral")
		default:
			apierror.RespondInternalError(c, err)
		}
		return
	}
	response.Success(c, result)
}

func (h *MarketingHandler) ExportReferralLedger(c *gin.Context) {
	if h == nil || h.referralService == nil {
		apierror.RespondInternalError(c, service.ErrReferralServiceUnavailable)
		return
	}
	filters := repository.ReferralAdminFilters{Status: strings.TrimSpace(c.Query("status")), Keyword: strings.TrimSpace(c.Query("keyword"))}
	const exportPageSize = 100
	items := make([]service.ReferralAdminItem, 0, exportPageSize)
	for page := 1; ; page++ {
		ledger, err := h.referralService.AdminLedger(filters, page, exportPageSize, time.Now().UTC())
		if err != nil {
			apierror.RespondInternalError(c, err)
			return
		}
		items = append(items, ledger.Items...)
		if len(ledger.Items) == 0 || page*exportPageSize >= int(ledger.Total) {
			break
		}
	}
	var body bytes.Buffer
	writer := csv.NewWriter(&body)
	_ = writer.Write([]string{"id", "referral_code", "referrer", "referee", "order_number", "amount_minor", "currency", "status", "vesting_until", "reward_points", "risk_flags", "created_at"})
	for _, item := range items {
		orderNumber, amountMinor, currencyCode := "", "", ""
		if item.Order != nil {
			orderNumber = item.Order.OrderNumber
			amountMinor = strconv.FormatInt(item.Order.AmountMinor, 10)
			currencyCode = item.Order.Currency
		}
		flags := make([]string, 0, len(item.RiskFlags))
		for _, flag := range item.RiskFlags {
			if flagType, ok := flag["type"].(string); ok && strings.TrimSpace(flagType) != "" {
				flags = append(flags, flagType)
			}
		}
		if err := writer.Write([]string{
			strconv.FormatUint(uint64(item.ID), 10), item.ReferralCode, item.Referrer.Name, item.Referee.Name,
			orderNumber, amountMinor, currencyCode, item.Status, formatReferralTime(item.VestingUntil),
			strconv.Itoa(item.RewardPoints), strings.Join(flags, ";"), item.CreatedAt.UTC().Format(time.RFC3339),
		}); err != nil {
			apierror.RespondInternalError(c, err)
			return
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	c.Header("Content-Disposition", `attachment; filename="referral-ledger.csv"`)
	c.Data(200, "text/csv; charset=utf-8", body.Bytes())
}

func formatReferralTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func (h *MarketingHandler) GetReferralProgramConfig(c *gin.Context) {
	if h == nil || h.referralService == nil {
		apierror.RespondInternalError(c, service.ErrReferralServiceUnavailable)
		return
	}
	config, err := h.referralService.GetAdminProgramConfig()
	if err != nil {
		if repository.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "referral program config")
		} else {
			apierror.RespondInternalError(c, err)
		}
		return
	}
	response.Success(c, gin.H{"config": config})
}

func (h *MarketingHandler) UpdateReferralProgramConfig(c *gin.Context) {
	if h == nil || h.referralService == nil {
		apierror.RespondInternalError(c, service.ErrReferralServiceUnavailable)
		return
	}
	var req struct {
		ExpectedVersion              int    `json:"expected_version" binding:"required,gt=0"`
		Enabled                      bool   `json:"enabled"`
		Currency                     string `json:"currency" binding:"required"`
		MinOrderAmountMinor          int64  `json:"min_order_amount_minor" binding:"gte=0"`
		ReferrerRewardPoints         int    `json:"referrer_reward_points" binding:"gte=0"`
		RefereeBenefitType           string `json:"referee_benefit_type" binding:"required"`
		RefereeBenefitValue          int64  `json:"referee_benefit_value" binding:"gte=0"`
		RefereeBenefitMaxAmountMinor int64  `json:"referee_benefit_max_amount_minor" binding:"gte=0"`
		CouponStackable              bool   `json:"coupon_stackable"`
		VestingPeriodDays            int    `json:"vesting_period_days" binding:"gt=0"`
		UndeliveredFallbackDays      int    `json:"undelivered_fallback_days" binding:"gt=0"`
		AttributionTTLDays           int    `json:"attribution_ttl_days" binding:"gt=0"`
		MonthlyCapPerReferrer        int    `json:"monthly_cap_per_referrer" binding:"gt=0"`
		AntiFraudMode                string `json:"anti_fraud_mode" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}
	var createdBy *uint
	if value, ok := c.Get("user_id"); ok {
		if userID, ok := value.(uint); ok && userID > 0 {
			createdBy = &userID
		}
	}
	config, err := h.referralService.PublishAdminProgramConfig(service.ReferralProgramConfigInput{
		Enabled: req.Enabled, Currency: req.Currency, MinOrderAmountMinor: req.MinOrderAmountMinor,
		ReferrerRewardPoints: req.ReferrerRewardPoints, RefereeBenefitType: req.RefereeBenefitType,
		RefereeBenefitValue: req.RefereeBenefitValue, RefereeBenefitMaxAmountMinor: req.RefereeBenefitMaxAmountMinor,
		CouponStackable: req.CouponStackable, VestingPeriodDays: req.VestingPeriodDays,
		UndeliveredFallbackDays: req.UndeliveredFallbackDays, AttributionTTLDays: req.AttributionTTLDays,
		MonthlyCapPerReferrer: req.MonthlyCapPerReferrer, AntiFraudMode: req.AntiFraudMode, CreatedBy: createdBy,
	}, req.ExpectedVersion)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrReferralProgramVersionConflict):
			apierror.RespondConflict(c, err.Error())
		case errors.Is(err, service.ErrInvalidReferralProgramConfig):
			apierror.RespondBadRequest(c, err.Error())
		default:
			apierror.RespondInternalError(c, err)
		}
		return
	}
	response.Success(c, gin.H{"config": config})
}

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

func (h *MarketingHandler) ListGiftCardRedemptions(c *gin.Context) {
	params := pagination.ParsePagination(c)
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)
	if userID == 0 {
		apierror.RespondBadRequest(c, "please provide user_id")
		return
	}

	redemptions, total, err := h.marketingService.ListGiftCardRedemptionsAdmin(uint(userID), params.Page, params.PageSize)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Paged(c, gin.H{"redemptions": redemptions}, params.Page, params.PageSize, total)
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
	apierror.RespondError(c, 410, "referral_legacy_endpoint_retired", "Legacy referral status mutation is disabled; the new referral state-machine actions are not enabled yet")
}
