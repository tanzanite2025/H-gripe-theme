package service

import (
	"errors"
	"fmt"
	"tanzanite/internal/domain/coupon"
	"tanzanite/internal/repository"
	"time"
)

type CouponCreateInput struct {
	Code                 string
	Type                 string
	Value                float64
	Description          string
	MinAmount            float64
	MaxDiscount          float64
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
	Value                *float64
	Description          *string
	MinAmount            *float64
	MaxDiscount          *float64
	UsageLimit           *int
	UsageLimitPerUser    *int
	StartDate            *time.Time
	EndDate              *time.Time
	ApplicableProducts   *string
	ExcludedProducts     *string
	ApplicableCategories *string
	Enabled              *bool
}

func (s *MarketingService) ValidateCoupon(code string, userID uint, amount float64) (*coupon.Coupon, float64, error) {
	c, err := s.couponRepo.FindCouponByCode(code)
	if err != nil {
		return nil, 0, errors.New("coupon not found")
	}

	if !c.Enabled {
		return nil, 0, errors.New("coupon is disabled")
	}

	now := time.Now()
	if now.Before(c.StartDate) || now.After(c.EndDate) {
		return nil, 0, errors.New("coupon is expired")
	}

	if c.UsageLimit > 0 && c.UsedCount >= c.UsageLimit {
		return nil, 0, errors.New("coupon usage limit reached")
	}

	if amount < c.MinAmount {
		return nil, 0, fmt.Errorf("minimum amount %.2f required", c.MinAmount)
	}

	var discount float64
	switch c.Type {
	case "fixed":
		discount = c.Value
	case "percentage":
		discount = amount * c.Value / 100
		if c.MaxDiscount > 0 && discount > c.MaxDiscount {
			discount = c.MaxDiscount
		}
	}

	return c, discount, nil
}

func (s *MarketingService) UseCoupon(couponID, userID, orderID uint, discountAmount float64) error {
	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if err := repos.Coupon.IncrementUsedCount(couponID); err != nil {
			return err
		}

		usage := &coupon.CouponUsage{
			CouponID: couponID,
			UserID:   userID,
			OrderID:  orderID,
			Discount: discountAmount,
		}

		return repos.Coupon.CreateCouponUsage(usage)
	})
}

func (s *MarketingService) GetActiveCoupons() ([]coupon.Coupon, error) {
	return s.couponRepo.FindActiveCoupons()
}

func (s *MarketingService) ListCouponsAdmin(page, pageSize int, status string) ([]coupon.Coupon, int64, error) {
	coupons, total, err := s.couponRepo.FindAllCoupons(page, pageSize)
	if err != nil || status == "" || status == "all" {
		return coupons, total, err
	}

	filtered := make([]coupon.Coupon, 0, len(coupons))
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
		default:
			return nil, 0, fmt.Errorf("unsupported coupon status filter %s", status)
		}
	}

	return filtered, int64(len(filtered)), nil
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

	cp := &coupon.Coupon{
		Code:                 input.Code,
		Type:                 input.Type,
		Value:                input.Value,
		Description:          input.Description,
		MinAmount:            input.MinAmount,
		MaxDiscount:          input.MaxDiscount,
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
	if input.Type != nil {
		cp.Type = *input.Type
	}
	if input.Value != nil {
		cp.Value = *input.Value
	}
	if input.Description != nil {
		cp.Description = *input.Description
	}
	if input.MinAmount != nil {
		cp.MinAmount = *input.MinAmount
	}
	if input.MaxDiscount != nil {
		cp.MaxDiscount = *input.MaxDiscount
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
