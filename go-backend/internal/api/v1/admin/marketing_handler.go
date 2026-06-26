package admin

import (
	"tanzanite/internal/repository"
)

// MarketingHandler 营销管理处理器
type MarketingHandler struct {
	couponRepo  *repository.CouponRepository
	loyaltyRepo *repository.LoyaltyRepository
}

// NewMarketingHandler 创建营销管理处理器
func NewMarketingHandler(couponRepo *repository.CouponRepository, loyaltyRepo *repository.LoyaltyRepository) *MarketingHandler {
	return &MarketingHandler{
		couponRepo:  couponRepo,
		loyaltyRepo: loyaltyRepo,
	}
}
