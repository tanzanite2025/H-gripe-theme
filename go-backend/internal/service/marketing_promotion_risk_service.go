package service

import (
	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/loyalty"
	"errors"
	"math"
	"sort"
	"strings"
	"time"
)

const (
	promotionRiskCouponCandidateLimit = 1000
	promotionRiskGatewayMinimumAmount = 0.50
	promotionRiskMinimumSubtotal      = 0.01
)

type MarketingPromotionRiskAnalysis struct {
	GeneratedAt          time.Time                     `json:"generated_at"`
	Currency             string                        `json:"currency"`
	GatewayMinimumAmount float64                       `json:"gateway_minimum_amount"`
	Summary              MarketingPromotionRiskSummary `json:"summary"`
	Items                []MarketingPromotionRiskItem  `json:"items"`
}

type MarketingPromotionRiskSummary struct {
	Severity                    string  `json:"severity"`
	CandidateCouponCount        int     `json:"candidate_coupon_count"`
	RiskItemCount               int     `json:"risk_item_count"`
	ZeroTotalRiskCount          int     `json:"zero_total_risk_count"`
	GatewayMinimumRiskCount     int     `json:"gateway_minimum_risk_count"`
	MemberLevelCount            int     `json:"member_level_count"`
	MaxMemberDiscountRate       float64 `json:"max_member_discount_rate"`
	MaxMemberDiscountLevelName  string  `json:"max_member_discount_level_name"`
	PointsRedemptionEnabled     bool    `json:"points_redemption_enabled"`
	DirectPointsDiscountCapRate float64 `json:"direct_points_discount_cap_rate"`
	MaxRedeemGiftCardValue      float64 `json:"max_redeem_gift_card_value"`
}

type MarketingPromotionRiskItem struct {
	Severity                   string    `json:"severity"`
	Kind                       string    `json:"kind"`
	Scenario                   string    `json:"scenario"`
	CouponID                   uint      `json:"coupon_id,omitempty"`
	CouponCode                 string    `json:"coupon_code,omitempty"`
	CouponType                 string    `json:"coupon_type,omitempty"`
	CouponStatus               string    `json:"coupon_status,omitempty"`
	CouponValue                float64   `json:"coupon_value,omitempty"`
	CouponMinAmount            float64   `json:"coupon_min_amount"`
	CouponMaxDiscount          float64   `json:"coupon_max_discount,omitempty"`
	MemberLevelID              uint      `json:"member_level_id,omitempty"`
	MemberLevelName            string    `json:"member_level_name,omitempty"`
	MemberDiscountRate         float64   `json:"member_discount_rate"`
	PointsDiscountRate         float64   `json:"points_discount_rate"`
	FullCoverSubtotalThreshold float64   `json:"full_cover_subtotal_threshold,omitempty"`
	GatewayMinimumThreshold    float64   `json:"gateway_minimum_threshold,omitempty"`
	EstimatedSubtotal          float64   `json:"estimated_subtotal"`
	EstimatedCouponDiscount    float64   `json:"estimated_coupon_discount"`
	EstimatedMemberDiscount    float64   `json:"estimated_member_discount"`
	EstimatedPointsDiscount    float64   `json:"estimated_points_discount"`
	EstimatedDiscountAmount    float64   `json:"estimated_discount_amount"`
	EstimatedPayableAmount     float64   `json:"estimated_payable_amount"`
	Factors                    []string  `json:"factors"`
	Recommendation             string    `json:"recommendation"`
	StartsAt                   time.Time `json:"starts_at,omitempty"`
	EndsAt                     time.Time `json:"ends_at,omitempty"`
}

type promotionDiscountShape struct {
	scenario   string
	fixed      float64
	rate       float64
	minAmount  float64
	maxAmount  float64
	couponRate float64
}

type promotionRiskThreshold struct {
	severity     string
	kind         string
	fullCover    float64
	gatewayFloor float64
	estimate     float64
}

func (s *MarketingService) AnalyzePromotionStackingRisk() (*MarketingPromotionRiskAnalysis, error) {
	coupons, err := s.couponRepo.FindCouponsForRiskAnalysis(promotionRiskCouponCandidateLimit)
	if err != nil {
		return nil, err
	}

	levels, err := s.loyaltyRepo.FindAllMemberLevels()
	if err != nil {
		return nil, err
	}
	maxLevel := promotionRiskMaxMemberLevel(levels)
	memberRate := 0.0
	if maxLevel != nil {
		memberRate = boundedRate(maxLevel.DiscountRate)
	}

	config, err := s.promotionRiskProgramConfig()
	if err != nil {
		return nil, err
	}
	pointsEnabled := config != nil && config.Enabled && config.ExchangeRatePoints > 0
	pointsRate := 0.0
	if pointsEnabled {
		pointsRate = checkoutMaxPointsDiscountSubtotalRate
	}

	items := make([]MarketingPromotionRiskItem, 0, len(coupons))
	if noCouponItem, ok := analyzeNoCouponPromotionRisk(maxLevel, pointsRate); ok {
		items = append(items, noCouponItem)
	}
	for _, cp := range coupons {
		if item, ok := analyzeCouponPromotionRisk(cp, maxLevel, pointsRate); ok {
			items = append(items, item)
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		leftRank := promotionRiskSeverityRank(items[i].Severity)
		rightRank := promotionRiskSeverityRank(items[j].Severity)
		if leftRank != rightRank {
			return leftRank > rightRank
		}
		if items[i].EstimatedPayableAmount != items[j].EstimatedPayableAmount {
			return items[i].EstimatedPayableAmount < items[j].EstimatedPayableAmount
		}
		return strings.Compare(items[i].CouponCode, items[j].CouponCode) < 0
	})

	summary := MarketingPromotionRiskSummary{
		Severity:                    "info",
		CandidateCouponCount:        len(coupons),
		RiskItemCount:               len(items),
		MemberLevelCount:            len(levels),
		MaxMemberDiscountRate:       roundPromotionRate(memberRate * 100),
		PointsRedemptionEnabled:     pointsEnabled,
		DirectPointsDiscountCapRate: roundPromotionRate(pointsRate * 100),
		MaxRedeemGiftCardValue:      maxRedeemGiftCardValue(config),
	}
	if maxLevel != nil {
		summary.MaxMemberDiscountLevelName = maxLevel.Name
	}
	for _, item := range items {
		switch item.Kind {
		case "zero_total":
			summary.ZeroTotalRiskCount++
		case "below_gateway_minimum":
			summary.GatewayMinimumRiskCount++
		}
	}
	if summary.ZeroTotalRiskCount > 0 {
		summary.Severity = "critical"
	} else if summary.GatewayMinimumRiskCount > 0 {
		summary.Severity = "warning"
	}

	return &MarketingPromotionRiskAnalysis{
		GeneratedAt:          time.Now().UTC(),
		Currency:             currency.DefaultPrimaryCurrency,
		GatewayMinimumAmount: promotionRiskGatewayMinimumAmount,
		Summary:              summary,
		Items:                items,
	}, nil
}

func (s *MarketingService) promotionRiskProgramConfig() (*loyalty.ProgramConfig, error) {
	if s.program == nil {
		return nil, nil
	}
	config, err := s.program.GetActive()
	if err != nil {
		if errors.Is(err, ErrLoyaltyProgramConfigNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return config, nil
}

func analyzeCouponPromotionRisk(cp coupon.Coupon, maxLevel *loyalty.MemberLevel, pointsRate float64) (MarketingPromotionRiskItem, bool) {
	memberRate := 0.0
	if maxLevel != nil {
		memberRate = boundedRate(maxLevel.DiscountRate)
	}
	baseRate := memberRate + boundedFraction(pointsRate)
	shapes := couponRiskDiscountShapes(cp, baseRate)

	var selected MarketingPromotionRiskItem
	found := false
	for _, shape := range shapes {
		threshold, ok := promotionRiskThresholdForShape(shape)
		if !ok {
			continue
		}

		item := promotionRiskItemFromCoupon(cp, maxLevel, shape, threshold, memberRate, pointsRate)
		if !found ||
			promotionRiskSeverityRank(item.Severity) > promotionRiskSeverityRank(selected.Severity) ||
			(promotionRiskSeverityRank(item.Severity) == promotionRiskSeverityRank(selected.Severity) && item.EstimatedPayableAmount < selected.EstimatedPayableAmount) {
			selected = item
			found = true
		}
	}
	return selected, found
}

func analyzeNoCouponPromotionRisk(maxLevel *loyalty.MemberLevel, pointsRate float64) (MarketingPromotionRiskItem, bool) {
	memberRate := 0.0
	if maxLevel != nil {
		memberRate = boundedRate(maxLevel.DiscountRate)
	}
	rate := memberRate + boundedFraction(pointsRate)
	if rate < 1 {
		return MarketingPromotionRiskItem{}, false
	}

	shape := promotionDiscountShape{
		scenario:   "member_points",
		rate:       rate,
		minAmount:  promotionRiskMinimumSubtotal,
		maxAmount:  0,
		couponRate: 0,
	}
	threshold, ok := promotionRiskThresholdForShape(shape)
	if !ok {
		return MarketingPromotionRiskItem{}, false
	}

	item := promotionRiskItemFromCoupon(coupon.Coupon{}, maxLevel, shape, threshold, memberRate, pointsRate)
	item.Scenario = "member_points"
	item.CouponStatus = ""
	item.Factors = promotionRiskFactors("", "", maxLevel, memberRate, pointsRate)
	item.Recommendation = promotionRiskRecommendation(item.Kind, "")
	return item, true
}

func couponRiskDiscountShapes(cp coupon.Coupon, baseRate float64) []promotionDiscountShape {
	minAmount := math.Max(cp.MinAmount, promotionRiskMinimumSubtotal)
	switch cp.Type {
	case "fixed":
		return []promotionDiscountShape{{
			scenario:   "coupon_member_points",
			fixed:      math.Max(0, cp.Value),
			rate:       baseRate,
			minAmount:  minAmount,
			couponRate: 0,
		}}
	case "percentage":
		couponRate := boundedRate(cp.Value)
		if cp.MaxDiscount <= 0 || couponRate <= 0 {
			return []promotionDiscountShape{{
				scenario:   "coupon_member_points",
				rate:       baseRate + couponRate,
				minAmount:  minAmount,
				couponRate: couponRate,
				fixed:      0,
				maxAmount:  0,
			}}
		}

		capStart := cp.MaxDiscount / couponRate
		return []promotionDiscountShape{
			{
				scenario:   "coupon_member_points_before_cap",
				rate:       baseRate + couponRate,
				minAmount:  minAmount,
				maxAmount:  capStart,
				couponRate: couponRate,
			},
			{
				scenario:  "coupon_member_points_after_cap",
				fixed:     cp.MaxDiscount,
				rate:      baseRate,
				minAmount: math.Max(minAmount, capStart),
				maxAmount: 0,
			},
		}
	default:
		return nil
	}
}

func promotionRiskThresholdForShape(shape promotionDiscountShape) (promotionRiskThreshold, bool) {
	minAmount := math.Max(shape.minAmount, promotionRiskMinimumSubtotal)
	if shape.maxAmount > 0 && minAmount > shape.maxAmount {
		return promotionRiskThreshold{}, false
	}

	rate := boundedFraction(shape.rate)
	if rate >= 1 {
		return promotionRiskThreshold{
			severity:  "critical",
			kind:      "zero_total",
			fullCover: shape.maxAmount,
			estimate:  minAmount,
		}, true
	}

	remainingRate := 1 - rate
	fullCover := 0.0
	if shape.fixed > 0 {
		fullCover = shape.fixed / remainingRate
	}
	gatewayFloor := (shape.fixed + promotionRiskGatewayMinimumAmount) / remainingRate
	fullCoverApplies := fullCover >= minAmount && (shape.maxAmount <= 0 || fullCover <= shape.maxAmount || minAmount <= shape.maxAmount)
	gatewayFloorApplies := gatewayFloor >= minAmount && (shape.maxAmount <= 0 || gatewayFloor <= shape.maxAmount || minAmount <= shape.maxAmount)

	if fullCoverApplies {
		return promotionRiskThreshold{
			severity:     "critical",
			kind:         "zero_total",
			fullCover:    fullCover,
			gatewayFloor: gatewayFloor,
			estimate:     minAmount,
		}, true
	}
	if gatewayFloorApplies {
		return promotionRiskThreshold{
			severity:     "warning",
			kind:         "below_gateway_minimum",
			fullCover:    fullCover,
			gatewayFloor: gatewayFloor,
			estimate:     minAmount,
		}, true
	}
	return promotionRiskThreshold{}, false
}

func promotionRiskItemFromCoupon(
	cp coupon.Coupon,
	maxLevel *loyalty.MemberLevel,
	shape promotionDiscountShape,
	threshold promotionRiskThreshold,
	memberRate float64,
	pointsRate float64,
) MarketingPromotionRiskItem {
	estimateSubtotal := roundMoney(math.Max(threshold.estimate, promotionRiskMinimumSubtotal))
	couponDiscount := 0.0
	if cp.ID > 0 {
		couponDiscount = roundMoney(cp.CalculateDiscount(estimateSubtotal))
	}
	memberDiscount := roundMoney(estimateSubtotal * memberRate)
	pointsDiscount := roundMoney(estimateSubtotal * boundedFraction(pointsRate))
	totalDiscount := roundMoney(couponDiscount + memberDiscount + pointsDiscount)
	payable := roundMoney(estimateSubtotal - totalDiscount)
	if payable < 0 {
		payable = 0
	}

	item := MarketingPromotionRiskItem{
		Severity:                   threshold.severity,
		Kind:                       threshold.kind,
		Scenario:                   shape.scenario,
		CouponID:                   cp.ID,
		CouponCode:                 cp.Code,
		CouponType:                 cp.Type,
		CouponStatus:               couponRiskStatus(cp),
		CouponValue:                roundMoney(cp.Value),
		CouponMinAmount:            roundMoney(cp.MinAmount),
		CouponMaxDiscount:          roundMoney(cp.MaxDiscount),
		MemberDiscountRate:         roundPromotionRate(memberRate * 100),
		PointsDiscountRate:         roundPromotionRate(boundedFraction(pointsRate) * 100),
		FullCoverSubtotalThreshold: roundMoney(threshold.fullCover),
		GatewayMinimumThreshold:    roundMoney(threshold.gatewayFloor),
		EstimatedSubtotal:          estimateSubtotal,
		EstimatedCouponDiscount:    couponDiscount,
		EstimatedMemberDiscount:    memberDiscount,
		EstimatedPointsDiscount:    pointsDiscount,
		EstimatedDiscountAmount:    totalDiscount,
		EstimatedPayableAmount:     payable,
		Factors:                    promotionRiskFactors(cp.Code, cp.Type, maxLevel, memberRate, pointsRate),
		Recommendation:             promotionRiskRecommendation(threshold.kind, cp.Code),
		StartsAt:                   cp.StartDate,
		EndsAt:                     cp.EndDate,
	}
	if maxLevel != nil {
		item.MemberLevelID = maxLevel.ID
		item.MemberLevelName = maxLevel.Name
	}
	return item
}

func promotionRiskMaxMemberLevel(levels []loyalty.MemberLevel) *loyalty.MemberLevel {
	var selected *loyalty.MemberLevel
	for index := range levels {
		level := &levels[index]
		if selected == nil || level.DiscountRate > selected.DiscountRate {
			selected = level
		}
	}
	return selected
}

func maxRedeemGiftCardValue(config *loyalty.ProgramConfig) float64 {
	if config == nil || !config.Enabled {
		return 0
	}
	maxValue := 0.0
	for _, option := range config.RedeemOptions {
		pointsRequired, err := PointsForGiftCardValue(option.ValueCents, config.ExchangeRatePoints)
		if err != nil || pointsRequired < config.MinRedeemPoints || option.RemainingQuantity() <= 0 {
			continue
		}
		maxValue = math.Max(maxValue, float64(option.ValueCents)/100)
	}
	return roundMoney(maxValue)
}

func promotionRiskFactors(couponCode string, couponType string, maxLevel *loyalty.MemberLevel, memberRate float64, pointsRate float64) []string {
	factors := make([]string, 0, 3)
	if couponCode != "" {
		if couponType == "percentage" {
			factors = append(factors, "percentage_coupon")
		} else {
			factors = append(factors, "fixed_coupon")
		}
	}
	if maxLevel != nil && memberRate > 0 {
		factors = append(factors, "member_level_discount")
	}
	if pointsRate > 0 {
		factors = append(factors, "direct_points_discount")
	}
	return factors
}

func promotionRiskRecommendation(kind string, couponCode string) string {
	switch kind {
	case "zero_total":
		return "raise minimum spend, cap the discount, or mark zero-total orders as internal settlement"
	case "below_gateway_minimum":
		return "raise minimum spend or ensure the payable amount stays above the payment gateway minimum"
	default:
		return "review promotion stacking limits"
	}
}

func couponRiskStatus(cp coupon.Coupon) string {
	now := time.Now()
	if !cp.Enabled {
		return "disabled"
	}
	if cp.EndDate.Before(now) {
		return "expired"
	}
	if cp.StartDate.After(now) {
		return "scheduled"
	}
	return "active"
}

func boundedRate(percent float64) float64 {
	if percent <= 0 {
		return 0
	}
	if percent >= 100 {
		return 1
	}
	return percent / 100
}

func boundedFraction(value float64) float64 {
	if value <= 0 {
		return 0
	}
	if value >= 1 {
		return 1
	}
	return value
}

func roundPromotionRate(value float64) float64 {
	return math.Round(value*100) / 100
}

func promotionRiskSeverityRank(severity string) int {
	switch severity {
	case "critical":
		return 3
	case "warning":
		return 2
	default:
		return 1
	}
}
