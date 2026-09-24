package service

import (
	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/loyalty"
	domainmoney "commerce-platform/internal/domain/money"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	promotionRiskCouponCandidateLimit = 1000
)

var (
	promotionRiskGatewayMinimumMoney  = domainmoney.MustNew(50, currency.DefaultPrimaryCurrency)
	promotionRiskMinimumSubtotalMoney = domainmoney.MustNew(1, currency.DefaultPrimaryCurrency)
)

type MarketingPromotionRiskAnalysis struct {
	GeneratedAt          time.Time                     `json:"generated_at"`
	Currency             string                        `json:"currency"`
	GatewayMinimumAmount string                        `json:"gateway_minimum_amount"`
	Summary              MarketingPromotionRiskSummary `json:"summary"`
	Items                []MarketingPromotionRiskItem  `json:"items"`
}

type MarketingPromotionRiskSummary struct {
	Severity                   string  `json:"severity"`
	CandidateCouponCount       int     `json:"candidate_coupon_count"`
	RiskItemCount              int     `json:"risk_item_count"`
	ZeroTotalRiskCount         int     `json:"zero_total_risk_count"`
	GatewayMinimumRiskCount    int     `json:"gateway_minimum_risk_count"`
	MemberLevelCount           int     `json:"member_level_count"`
	MaxMemberDiscountRate      float64 `json:"max_member_discount_rate"`
	MaxMemberDiscountLevelName string  `json:"max_member_discount_level_name"`
}

type MarketingPromotionRiskItem struct {
	Severity                   string    `json:"severity"`
	Kind                       string    `json:"kind"`
	Scenario                   string    `json:"scenario"`
	CouponID                   uint      `json:"coupon_id,omitempty"`
	CouponCode                 string    `json:"coupon_code,omitempty"`
	CouponType                 string    `json:"coupon_type,omitempty"`
	CouponStatus               string    `json:"coupon_status,omitempty"`
	CouponValue                string    `json:"coupon_value,omitempty"`
	CouponMinAmount            string    `json:"coupon_min_amount"`
	CouponMaxDiscount          string    `json:"coupon_max_discount,omitempty"`
	MemberLevelID              uint      `json:"member_level_id,omitempty"`
	MemberLevelName            string    `json:"member_level_name,omitempty"`
	MemberDiscountRate         float64   `json:"member_discount_rate"`
	FullCoverSubtotalThreshold string    `json:"full_cover_subtotal_threshold,omitempty"`
	GatewayMinimumThreshold    string    `json:"gateway_minimum_threshold,omitempty"`
	EstimatedSubtotal          string    `json:"estimated_subtotal"`
	EstimatedCouponDiscount    string    `json:"estimated_coupon_discount"`
	EstimatedMemberDiscount    string    `json:"estimated_member_discount"`
	EstimatedDiscountAmount    string    `json:"estimated_discount_amount"`
	EstimatedPayableAmount     string    `json:"estimated_payable_amount"`
	Factors                    []string  `json:"factors"`
	Recommendation             string    `json:"recommendation"`
	StartsAt                   time.Time `json:"starts_at,omitempty"`
	EndsAt                     time.Time `json:"ends_at,omitempty"`
	estimatedPayableMinor      int64     `json:"-"`
}

type promotionDiscountShape struct {
	scenario     string
	fixed        domainmoney.Money
	rate         *big.Rat
	minAmount    domainmoney.Money
	maxAmount    domainmoney.Money
	hasMaxAmount bool
}

type promotionRiskThreshold struct {
	severity     string
	kind         string
	fullCover    domainmoney.Money
	gatewayFloor domainmoney.Money
	estimate     domainmoney.Money
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
	memberRate := memberLevelDiscountRateFraction(maxLevel)

	items := make([]MarketingPromotionRiskItem, 0, len(coupons))
	if noCouponItem, ok := analyzeNoCouponPromotionRisk(maxLevel); ok {
		items = append(items, noCouponItem)
	}
	for _, cp := range coupons {
		if item, ok := analyzeCouponPromotionRisk(cp, maxLevel); ok {
			items = append(items, item)
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		leftRank := promotionRiskSeverityRank(items[i].Severity)
		rightRank := promotionRiskSeverityRank(items[j].Severity)
		if leftRank != rightRank {
			return leftRank > rightRank
		}
		if items[i].estimatedPayableMinor != items[j].estimatedPayableMinor {
			return items[i].estimatedPayableMinor < items[j].estimatedPayableMinor
		}
		return strings.Compare(items[i].CouponCode, items[j].CouponCode) < 0
	})

	summary := MarketingPromotionRiskSummary{
		Severity:              "info",
		CandidateCouponCount:  len(coupons),
		RiskItemCount:         len(items),
		MemberLevelCount:      len(levels),
		MaxMemberDiscountRate: promotionRiskRatePercent(memberRate),
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
		GatewayMinimumAmount: promotionRiskMoneyMajor(promotionRiskGatewayMinimumMoney),
		Summary:              summary,
		Items:                items,
	}, nil
}

func analyzeCouponPromotionRisk(cp coupon.Coupon, maxLevel *loyalty.MemberLevel) (MarketingPromotionRiskItem, bool) {
	// Risk analysis is emitted in the primary settlement currency. A coupon
	// denominated in another currency cannot be safely compared without an
	// explicit FX rate, so leave it out instead of treating its minor units as
	// primary-currency money.
	couponCurrency := currency.NormalizeCode(cp.Currency)
	if couponCurrency != "" && couponCurrency != currency.DefaultPrimaryCurrency {
		return MarketingPromotionRiskItem{}, false
	}
	memberRate := memberLevelDiscountRateFraction(maxLevel)
	baseRate := memberRate
	shapes := couponRiskDiscountShapes(cp, baseRate)

	var selected MarketingPromotionRiskItem
	found := false
	for _, shape := range shapes {
		threshold, ok := promotionRiskThresholdForShape(shape)
		if !ok {
			continue
		}

		item := promotionRiskItemFromCoupon(cp, maxLevel, shape, threshold, memberRate)
		if !found ||
			promotionRiskSeverityRank(item.Severity) > promotionRiskSeverityRank(selected.Severity) ||
			(promotionRiskSeverityRank(item.Severity) == promotionRiskSeverityRank(selected.Severity) && item.EstimatedPayableAmount < selected.EstimatedPayableAmount) {
			selected = item
			found = true
		}
	}
	return selected, found
}

func analyzeNoCouponPromotionRisk(maxLevel *loyalty.MemberLevel) (MarketingPromotionRiskItem, bool) {
	memberRate := memberLevelDiscountRateFraction(maxLevel)
	rate := memberRate
	if rate.Cmp(big.NewRat(1, 1)) < 0 {
		return MarketingPromotionRiskItem{}, false
	}

	shape := promotionDiscountShape{
		scenario:  "member_points",
		rate:      rate,
		minAmount: promotionRiskMinimumSubtotalMoney,
		maxAmount: domainmoney.MustNew(0, currency.DefaultPrimaryCurrency),
	}
	threshold, ok := promotionRiskThresholdForShape(shape)
	if !ok {
		return MarketingPromotionRiskItem{}, false
	}

	item := promotionRiskItemFromCoupon(coupon.Coupon{}, maxLevel, shape, threshold, memberRate)
	item.Scenario = "member_points"
	item.CouponStatus = ""
	item.Factors = promotionRiskFactors("", "", maxLevel, memberRate)
	item.Recommendation = promotionRiskRecommendation(item.Kind, "")
	return item, true
}

func couponRiskDiscountShapes(cp coupon.Coupon, baseRate *big.Rat) []promotionDiscountShape {
	minMinor := cp.MinAmountMinor
	if minMinor < promotionRiskMinimumSubtotalMoney.AmountMinor() {
		minMinor = promotionRiskMinimumSubtotalMoney.AmountMinor()
	}
	minAmount, err := domainmoney.New(minMinor, currency.DefaultPrimaryCurrency)
	if err != nil || cp.MinAmountMinor < 0 || cp.ValueMinor < 0 || cp.MaxDiscountMinor < 0 {
		return nil
	}
	baseRateRat := promotionRiskCloneRate(baseRate)
	switch cp.Type {
	case "fixed":
		fixed, fixedErr := domainmoney.New(cp.ValueMinor, currency.DefaultPrimaryCurrency)
		if fixedErr != nil {
			return nil
		}
		return []promotionDiscountShape{{
			scenario:  "coupon_member_points",
			fixed:     fixed,
			rate:      baseRateRat,
			minAmount: minAmount,
		}}
	case "percentage":
		couponRate, ok := promotionRiskCouponRate(cp.ValueRateDecimal)
		if !ok {
			return nil
		}
		if cp.MaxDiscountMinor <= 0 || couponRate.Sign() <= 0 {
			return []promotionDiscountShape{{
				scenario:  "coupon_member_points",
				rate:      promotionRiskAddRates(baseRateRat, couponRate),
				minAmount: minAmount,
			}}
		}

		maxDiscount, maxErr := domainmoney.New(cp.MaxDiscountMinor, currency.DefaultPrimaryCurrency)
		if maxErr != nil {
			return nil
		}
		inverseCouponRate := new(big.Rat).Inv(couponRate)
		capStart, capErr := maxDiscount.MultiplyRat(inverseCouponRate)
		if capErr != nil {
			return nil
		}
		return []promotionDiscountShape{
			{
				scenario:     "coupon_member_points_before_cap",
				rate:         promotionRiskAddRates(baseRateRat, couponRate),
				minAmount:    minAmount,
				maxAmount:    capStart,
				hasMaxAmount: true,
			},
			{
				scenario:  "coupon_member_points_after_cap",
				fixed:     maxDiscount,
				rate:      baseRateRat,
				minAmount: promotionRiskMaxMoney(minAmount, capStart),
			},
		}
	default:
		return nil
	}
}

func promotionRiskThresholdForShape(shape promotionDiscountShape) (promotionRiskThreshold, bool) {
	minAmount := promotionRiskMaxMoney(shape.minAmount, promotionRiskMinimumSubtotalMoney)
	if shape.hasMaxAmount && minAmount.AmountMinor() > shape.maxAmount.AmountMinor() {
		return promotionRiskThreshold{}, false
	}

	rate := promotionRiskBoundedFraction(shape.rate)
	if rate.Cmp(big.NewRat(1, 1)) >= 0 {
		fullCover := domainmoney.MustNew(0, currency.DefaultPrimaryCurrency)
		if shape.hasMaxAmount {
			fullCover = shape.maxAmount
		}
		return promotionRiskThreshold{
			severity:  "critical",
			kind:      "zero_total",
			fullCover: fullCover,
			estimate:  minAmount,
		}, true
	}

	remainingRate := new(big.Rat).Sub(big.NewRat(1, 1), rate)
	fullCover := domainmoney.MustNew(0, currency.DefaultPrimaryCurrency)
	if shape.fixed.AmountMinor() > 0 {
		fullCover, _ = shape.fixed.MultiplyRat(new(big.Rat).Inv(remainingRate))
	}
	gatewayBase, _ := shape.fixed.Add(promotionRiskGatewayMinimumMoney)
	gatewayFloor, _ := gatewayBase.MultiplyRat(new(big.Rat).Inv(remainingRate))
	fullCoverApplies := promotionRiskThresholdInRange(fullCover, minAmount, shape)
	gatewayFloorApplies := promotionRiskThresholdInRange(gatewayFloor, minAmount, shape)

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

func promotionRiskThresholdInRange(candidate, minimum domainmoney.Money, shape promotionDiscountShape) bool {
	if candidate.AmountMinor() < minimum.AmountMinor() {
		return false
	}
	return !shape.hasMaxAmount || candidate.AmountMinor() <= shape.maxAmount.AmountMinor()
}

func promotionRiskItemFromCoupon(
	cp coupon.Coupon,
	maxLevel *loyalty.MemberLevel,
	shape promotionDiscountShape,
	threshold promotionRiskThreshold,
	memberRate *big.Rat,
) MarketingPromotionRiskItem {
	estimateMoney := promotionRiskMaxMoney(threshold.estimate, promotionRiskMinimumSubtotalMoney)
	couponDiscountMoney := domainmoney.MustNew(0, currency.DefaultPrimaryCurrency)
	if cp.ID > 0 && (currency.NormalizeCode(cp.Currency) == "" || currency.NormalizeCode(cp.Currency) == currency.DefaultPrimaryCurrency) {
		discountMoney, discountErr := cp.CalculateDiscountMoney(estimateMoney)
		if discountErr == nil {
			couponDiscountMoney = discountMoney
		}
	}
	memberMoney := promotionRiskRateAmount(estimateMoney, memberRate)
	memberDiscount := promotionRiskMoneyMajor(memberMoney)
	totalMoney, totalErr := couponDiscountMoney.Add(memberMoney)
	if totalErr != nil {
		totalMoney = estimateMoney
	}
	if totalMoney.AmountMinor() > estimateMoney.AmountMinor() {
		totalMoney = estimateMoney
	}
	estimateSubtotal := promotionRiskMoneyMajor(estimateMoney)
	couponDiscount := promotionRiskMoneyMajor(couponDiscountMoney)
	totalDiscount := promotionRiskMoneyMajor(totalMoney)
	payableMoney, payableErr := estimateMoney.Subtract(totalMoney)
	if payableErr != nil || payableMoney.AmountMinor() < 0 {
		payableMoney = domainmoney.MustNew(0, currency.DefaultPrimaryCurrency)
	}
	payable := promotionRiskMoneyMajor(payableMoney)

	item := MarketingPromotionRiskItem{
		Severity:                   threshold.severity,
		Kind:                       threshold.kind,
		Scenario:                   shape.scenario,
		CouponID:                   cp.ID,
		CouponCode:                 cp.Code,
		CouponType:                 cp.Type,
		CouponStatus:               couponRiskStatus(cp),
		CouponValue:                couponRiskDisplayValue(cp),
		CouponMinAmount:            couponRiskDisplayMoney(cp.MinAmountMinor),
		CouponMaxDiscount:          couponRiskDisplayMoney(cp.MaxDiscountMinor),
		MemberDiscountRate:         promotionRiskRatePercent(memberRate),
		FullCoverSubtotalThreshold: promotionRiskMoneyMajor(threshold.fullCover),
		GatewayMinimumThreshold:    promotionRiskMoneyMajor(threshold.gatewayFloor),
		EstimatedSubtotal:          estimateSubtotal,
		EstimatedCouponDiscount:    couponDiscount,
		EstimatedMemberDiscount:    memberDiscount,
		EstimatedDiscountAmount:    totalDiscount,
		EstimatedPayableAmount:     payable,
		Factors:                    promotionRiskFactors(cp.Code, cp.Type, maxLevel, memberRate),
		Recommendation:             promotionRiskRecommendation(threshold.kind, cp.Code),
		StartsAt:                   cp.StartDate,
		EndsAt:                     cp.EndDate,
		estimatedPayableMinor:      payableMoney.AmountMinor(),
	}
	if maxLevel != nil {
		item.MemberLevelID = maxLevel.ID
		item.MemberLevelName = maxLevel.Name
	}
	return item
}

func promotionRiskRateAmount(base domainmoney.Money, rate *big.Rat) domainmoney.Money {
	if rate == nil || rate.Sign() <= 0 {
		return domainmoney.MustNew(0, base.Currency().String())
	}
	value, err := base.MultiplyRat(rate)
	if err != nil {
		return domainmoney.MustNew(0, base.Currency().String())
	}
	return value
}

func promotionRiskCouponRate(value string) (*big.Rat, bool) {
	rat, ok := new(big.Rat).SetString(strings.TrimSpace(value))
	if !ok || rat.Sign() < 0 || rat.Cmp(big.NewRat(100, 1)) > 0 {
		return nil, false
	}
	rat.Quo(rat, big.NewRat(100, 1))
	return rat, true
}

func promotionRiskCloneRate(value *big.Rat) *big.Rat {
	if value == nil {
		return new(big.Rat)
	}
	return new(big.Rat).Set(value)
}

func promotionRiskAddRates(values ...*big.Rat) *big.Rat {
	total := new(big.Rat)
	for _, value := range values {
		if value != nil {
			total.Add(total, value)
		}
	}
	if total.Sign() < 0 {
		return new(big.Rat)
	}
	if total.Cmp(big.NewRat(1, 1)) > 0 {
		return big.NewRat(1, 1)
	}
	return total
}

func promotionRiskBoundedFraction(value *big.Rat) *big.Rat {
	if value == nil || value.Sign() <= 0 {
		return new(big.Rat)
	}
	if value.Cmp(big.NewRat(1, 1)) >= 0 {
		return big.NewRat(1, 1)
	}
	return promotionRiskCloneRate(value)
}

func promotionRiskMaxMoney(left, right domainmoney.Money) domainmoney.Money {
	if left.AmountMinor() >= right.AmountMinor() {
		return left
	}
	return right
}

func promotionRiskMoneyMajor(value domainmoney.Money) string {
	major, err := value.FormatMajor()
	if err != nil {
		return "0"
	}
	return major
}

func couponRiskDisplayMoney(amountMinor int64) string {
	if amountMinor <= 0 {
		return "0"
	}
	value, err := domainmoney.New(amountMinor, currency.DefaultPrimaryCurrency)
	if err != nil {
		return "0"
	}
	return promotionRiskMoneyMajor(value)
}

func couponRiskDisplayValue(cp coupon.Coupon) string {
	if cp.Type == "percentage" {
		value, err := strconv.ParseFloat(strings.TrimSpace(cp.ValueRateDecimal), 64)
		if err != nil || value < 0 {
			return "0"
		}
		return strconv.FormatFloat(roundPromotionRate(value), 'f', -1, 64)
	}
	return couponRiskDisplayMoney(cp.ValueMinor)
}

func promotionRiskMaxMemberLevel(levels []loyalty.MemberLevel) *loyalty.MemberLevel {
	var selected *loyalty.MemberLevel
	for index := range levels {
		level := &levels[index]
		levelRate := memberLevelDiscountRateRat(level)
		selectedRate := memberLevelDiscountRateRat(selected)
		if selected == nil || (levelRate != nil && (selectedRate == nil || levelRate.Cmp(selectedRate) > 0)) {
			selected = level
		}
	}
	return selected
}

func memberLevelDiscountRateRat(level *loyalty.MemberLevel) *big.Rat {
	if level == nil {
		return nil
	}
	rate, ok := new(big.Rat).SetString(strings.TrimSpace(level.DiscountRateDecimal))
	if !ok || rate.Sign() < 0 {
		return nil
	}
	return rate
}

func memberLevelDiscountRateFraction(level *loyalty.MemberLevel) *big.Rat {
	rate := memberLevelDiscountRateRat(level)
	if rate == nil {
		return new(big.Rat)
	}
	rate.Quo(rate, big.NewRat(100, 1))
	return promotionRiskBoundedFraction(rate)
}

func promotionRiskFactors(couponCode string, couponType string, maxLevel *loyalty.MemberLevel, memberRate *big.Rat) []string {
	factors := make([]string, 0, 3)
	if couponCode != "" {
		if couponType == "percentage" {
			factors = append(factors, "percentage_coupon")
		} else {
			factors = append(factors, "fixed_coupon")
		}
	}
	if maxLevel != nil && memberRate != nil && memberRate.Sign() > 0 {
		factors = append(factors, "member_level_discount")
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

// promotionRiskRatePercent is the read-model boundary. Risk arithmetic stays
// exact; only the final API percentage projection uses float64.
func promotionRiskRatePercent(value *big.Rat) float64 {
	if value == nil {
		return 0
	}
	percent := new(big.Rat).Mul(value, big.NewRat(100, 1))
	result, _ := percent.Float64()
	return roundPromotionRate(result)
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
