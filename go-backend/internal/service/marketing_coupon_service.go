package service

import (
	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/repository"
	"errors"
	"fmt"
	"strings"
	"time"
)

type CouponCreateInput struct {
	Code                 string
	Type                 string
	ValueMinor           int64
	ValueRateDecimal     string
	Currency             string
	Description          string
	MinAmountMinor       int64
	MaxDiscountMinor     int64
	UsageLimit           int
	UsageLimitPerUser    int
	StartDate            time.Time
	EndDate              time.Time
	ApplicableProducts   string
	ExcludedProducts     string
	ApplicableCategories string
	Enabled              bool
}

type CouponUpdateInput struct {
	Code                 *string
	Type                 *string
	ValueMinor           *int64
	ValueRateDecimal     *string
	Currency             *string
	Description          *string
	MinAmountMinor       *int64
	MaxDiscountMinor     *int64
	UsageLimit           *int
	UsageLimitPerUser    *int
	StartDate            *time.Time
	EndDate              *time.Time
	ApplicableProducts   *string
	ExcludedProducts     *string
	ApplicableCategories *string
	Enabled              *bool
}

// ValidateCoupon validates a coupon against an amount expressed in the
// coupon's canonical minor unit. No major-unit float is accepted in the
// transactional coupon path.
func (s *MarketingService) ValidateCoupon(code string, userID uint, amountMinor int64, emails ...string) (*coupon.Coupon, int64, error) {
	c, err := s.couponRepo.FindCouponByCode(code)
	if err != nil {
		return nil, 0, errors.New("coupon not found")
	}

	if !c.Enabled {
		return nil, 0, errors.New("coupon is disabled")
	}
	if err := validateCouponRecipient(c, userID); err != nil {
		return nil, 0, err
	}
	couponCurrency, err := parseCouponCurrency(c.Currency)
	if err != nil {
		return nil, 0, err
	}

	now := time.Now()
	if now.Before(c.StartDate) || now.After(c.EndDate) {
		return nil, 0, errors.New("coupon is expired")
	}

	if err := validateCouponUsageLimit(c); err != nil {
		return nil, 0, err
	}
	var email string
	if len(emails) > 0 {
		email = emails[0]
	}
	if err := validateCouponPerUserUsageLimit(s.couponRepo, c, userID, email); err != nil {
		return nil, 0, err
	}

	if amountMinor < 0 {
		return nil, 0, errors.New("amount cannot be negative")
	}
	amountMoney, err := domainmoney.New(amountMinor, couponCurrency)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid amount: %w", err)
	}
	minimumMoney, err := domainmoney.New(c.MinAmountMinor, couponCurrency)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid coupon minimum amount: %w", err)
	}
	if amountMoney.AmountMinor() < minimumMoney.AmountMinor() {
		minimumMajor, _ := minimumMoney.FormatMajor()
		return nil, 0, fmt.Errorf("minimum amount %s required", minimumMajor)
	}

	discountMoney, err := c.CalculateDiscountMoney(amountMoney)
	if err != nil {
		return nil, 0, err
	}
	return c, discountMoney.AmountMinor(), nil
}

func validateCouponRecipient(c *coupon.Coupon, userID uint) error {
	if c == nil || c.ReferralRecipientUserID == nil {
		return nil
	}
	if userID == 0 || *c.ReferralRecipientUserID != userID {
		return errors.New("coupon is not available for this customer")
	}
	return nil
}

func (s *MarketingService) UseCoupon(couponID, userID, orderID uint, discountAmountMinor int64, emails ...string) error {
	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		c, err := repos.Coupon.FindCouponByIDForUpdate(couponID)
		if err != nil {
			return err
		}
		if err := validateCouponUsageLimit(c); err != nil {
			return err
		}
		if err := validateCouponRecipient(c, userID); err != nil {
			return err
		}
		if _, err := parseCouponCurrency(c.Currency); err != nil {
			return err
		}
		var email string
		if len(emails) > 0 {
			email = emails[0]
		}
		if err := validateCouponPerUserUsageLimit(repos.Coupon, c, userID, email); err != nil {
			return err
		}
		if err := repos.Coupon.IncrementUsedCount(couponID); err != nil {
			return err
		}

		usage := &coupon.CouponUsage{
			CouponID: couponID,
			UserID:   userID,
			Email:    coupon.NormalizeEmail(email),
			OrderID:  orderID,
			Currency: c.Currency,
		}
		if discountAmountMinor < 0 {
			return errors.New("coupon discount cannot be negative")
		}
		if discountMoney, moneyErr := domainmoney.New(discountAmountMinor, c.Currency); moneyErr != nil {
			return moneyErr
		} else {
			usage.DiscountMinor = discountMoney.AmountMinor()
		}

		return repos.Coupon.CreateCouponUsage(usage)
	})
}

func validateCouponUsageLimit(c *coupon.Coupon) error {
	if c.UsageLimit > 0 && c.UsedCount >= c.UsageLimit {
		return repository.ErrCouponUsageLimitReached
	}
	return nil
}

func validateCouponPerUserUsageLimit(couponRepo *repository.CouponRepository, c *coupon.Coupon, userID uint, emails ...string) error {
	if c.UsageLimitPerUser <= 0 {
		return nil
	}
	if userID > 0 {
		usedCount, err := couponRepo.CountUserCouponUsage(userID, c.ID)
		if err != nil {
			return err
		}
		if int(usedCount) >= c.UsageLimitPerUser {
			return ErrCouponPerUserUsageLimitReached
		}
		return nil
	}

	var email string
	if len(emails) > 0 {
		email = emails[0]
	}
	email = coupon.NormalizeEmail(email)
	if email == "" {
		return ErrCouponUsageIdentityRequired
	}
	usedCount, err := couponRepo.CountEmailCouponUsage(email, c.ID)
	if err != nil {
		return err
	}
	if int(usedCount) >= c.UsageLimitPerUser {
		return ErrCouponPerUserUsageLimitReached
	}
	return nil
}

func (s *MarketingService) GetActiveCoupons() ([]coupon.Coupon, error) {
	return s.couponRepo.FindActiveCoupons()
}

func (s *MarketingService) ListCouponsAdmin(page, pageSize int, status string) ([]coupon.Coupon, int64, error) {
	return s.couponRepo.FindAllCouponsByStatus(page, pageSize, status)
}

func (s *MarketingService) GetCoupon(id uint) (*coupon.Coupon, error) {
	cp, err := s.couponRepo.FindCouponByID(id)
	if err != nil {
		return nil, normalizeMarketingError(err)
	}
	return cp, nil
}

func (s *MarketingService) CreateCouponAdmin(input CouponCreateInput) (*coupon.Coupon, error) {
	if err := s.ensureCouponCodeAvailable(input.Code, 0); err != nil {
		return nil, err
	}
	couponCurrency, err := parseCouponCurrency(input.Currency)
	if err != nil {
		return nil, err
	}

	cp := &coupon.Coupon{
		Code:                 input.Code,
		Type:                 input.Type,
		ValueMinor:           input.ValueMinor,
		ValueRateDecimal:     input.ValueRateDecimal,
		Currency:             couponCurrency,
		Description:          input.Description,
		MinAmountMinor:       input.MinAmountMinor,
		MaxDiscountMinor:     input.MaxDiscountMinor,
		UsageLimit:           input.UsageLimit,
		UsageLimitPerUser:    input.UsageLimitPerUser,
		StartDate:            input.StartDate,
		EndDate:              input.EndDate,
		ApplicableProducts:   input.ApplicableProducts,
		ExcludedProducts:     input.ExcludedProducts,
		ApplicableCategories: input.ApplicableCategories,
		Enabled:              input.Enabled,
	}
	if err := s.couponRepo.CreateCoupon(cp); err != nil {
		return nil, err
	}

	return cp, nil
}

func (s *MarketingService) UpdateCouponAdmin(id uint, input CouponUpdateInput) (*coupon.Coupon, error) {
	cp, err := s.GetCoupon(id)
	if err != nil {
		return nil, err
	}

	if input.Code != nil && *input.Code != cp.Code {
		if err := s.ensureCouponCodeAvailable(*input.Code, cp.ID); err != nil {
			return nil, err
		}
		cp.Code = *input.Code
	}
	if input.Currency != nil {
		couponCurrency, currencyErr := parseCouponCurrency(*input.Currency)
		if currencyErr != nil {
			return nil, currencyErr
		}
		if couponCurrency != cp.Currency && input.ValueMinor == nil && input.ValueRateDecimal == nil && input.MinAmountMinor == nil && input.MaxDiscountMinor == nil {
			return nil, errors.New("changing coupon currency requires canonical monetary values")
		}
		cp.Currency = couponCurrency
	}
	if input.Type != nil {
		cp.Type = *input.Type
		if strings.EqualFold(*input.Type, "percentage") {
			cp.ValueMinor = 0
		} else if strings.EqualFold(*input.Type, "fixed") {
			cp.ValueRateDecimal = ""
		}
	}
	if input.ValueMinor != nil {
		cp.ValueMinor = *input.ValueMinor
	}
	if input.ValueRateDecimal != nil {
		cp.ValueRateDecimal = *input.ValueRateDecimal
	}
	if input.Description != nil {
		cp.Description = *input.Description
	}
	if input.MinAmountMinor != nil {
		cp.MinAmountMinor = *input.MinAmountMinor
	}
	if input.MaxDiscountMinor != nil {
		cp.MaxDiscountMinor = *input.MaxDiscountMinor
	}
	if input.UsageLimit != nil {
		cp.UsageLimit = *input.UsageLimit
	}
	if input.UsageLimitPerUser != nil {
		cp.UsageLimitPerUser = *input.UsageLimitPerUser
	}
	if input.StartDate != nil {
		cp.StartDate = *input.StartDate
	}
	if input.EndDate != nil {
		cp.EndDate = *input.EndDate
	}
	if input.ApplicableProducts != nil {
		cp.ApplicableProducts = *input.ApplicableProducts
	}
	if input.ExcludedProducts != nil {
		cp.ExcludedProducts = *input.ExcludedProducts
	}
	if input.ApplicableCategories != nil {
		cp.ApplicableCategories = *input.ApplicableCategories
	}
	if input.Enabled != nil {
		cp.Enabled = *input.Enabled
	}

	if err := s.couponRepo.UpdateCoupon(cp); err != nil {
		return nil, err
	}

	return cp, nil
}

func parseCouponCurrency(value string) (string, error) {
	value = currency.NormalizeCode(value)
	if value == "" {
		return currency.DefaultPrimaryCurrency, nil
	}
	if !currency.IsCatalogCode(value) {
		return "", fmt.Errorf("unsupported coupon currency %s", value)
	}
	return value, nil
}

func (s *MarketingService) DeleteCouponAdmin(id uint) error {
	if _, err := s.GetCoupon(id); err != nil {
		return err
	}
	return s.couponRepo.DeleteCoupon(id)
}

func (s *MarketingService) GetCouponStats() (map[string]interface{}, error) {
	coupons, _, err := s.couponRepo.FindAllCoupons(1, 1000)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	stats := map[string]interface{}{
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

	return stats, nil
}

func (s *MarketingService) ensureCouponCodeAvailable(code string, excludeID uint) error {
	existing, err := s.couponRepo.FindCouponByCode(code)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil
		}
		return err
	}
	if existing != nil && existing.ID != excludeID {
		return ErrCouponCodeExists
	}
	return nil
}
