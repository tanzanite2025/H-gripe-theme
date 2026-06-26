package admin

import (
	"tanzanite/internal/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

// ============ 营销统计 ============

// GetMarketingStats 获取营销统计
func (h *MarketingHandler) GetMarketingStats(c *gin.Context) {
	stats := gin.H{}

	// 优惠券统计
	coupons, _, _ := h.couponRepo.FindAllCoupons(1, 1000)
	now := time.Now()
	couponStats := gin.H{
		"total":  len(coupons),
		"active": 0,
		"used":   0,
	}
	totalUsed := 0
	for _, cp := range coupons {
		if cp.Enabled && now.After(cp.StartDate) && now.Before(cp.EndDate) {
			couponStats["active"] = couponStats["active"].(int) + 1
		}
		totalUsed += cp.UsedCount
	}
	couponStats["used"] = totalUsed
	stats["coupons"] = couponStats

	// 积分统计
	loyaltyStats, _ := h.loyaltyRepo.GetLoyaltyStats()
	stats["loyalty"] = loyaltyStats

	response.Success(c, stats)
}
