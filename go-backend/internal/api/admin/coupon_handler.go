package admin

import (
	"strconv"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/pagination"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *MarketingHandler) ListCoupons(c *gin.Context) {
	params := pagination.ParsePagination(c)
	status := c.Query("status")

	coupons, total, err := h.marketingService.ListCouponsAdmin(params.Page, params.PageSize, status)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Paged(c, coupons, params.Page, params.PageSize, total)
}

func (h *MarketingHandler) GetCoupon(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid coupon ID")
		return
	}

	cp, err := h.marketingService.GetCoupon(uint(id))
	if err != nil {
		respondMarketingError(c, err, "coupon")
		return
	}

	response.Success(c, gin.H{"coupon": cp})
}

func (h *MarketingHandler) CreateCoupon(c *gin.Context) {
	var req struct {
		Code                 string    `json:"code" binding:"required"`
		Type                 string    `json:"type" binding:"required,oneof=fixed percentage"`
		Value                float64   `json:"value" binding:"required,gt=0"`
		Description          string    `json:"description"`
		MinAmount            float64   `json:"min_amount"`
		MaxDiscount          float64   `json:"max_discount"`
		UsageLimit           int       `json:"usage_limit"`
		UsageLimitPerUser    int       `json:"usage_limit_per_user"`
		StartDate            time.Time `json:"start_date" binding:"required"`
		EndDate              time.Time `json:"end_date" binding:"required"`
		ApplicableProducts   string    `json:"applicable_products"`
		ExcludedProducts     string    `json:"excluded_products"`
		ApplicableCategories string    `json:"applicable_categories"`
		Enabled              *bool     `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	cp, err := h.marketingService.CreateCouponAdmin(service.CouponCreateInput{
		Code:                 req.Code,
		Type:                 req.Type,
		Value:                req.Value,
		Description:          req.Description,
		MinAmount:            req.MinAmount,
		MaxDiscount:          req.MaxDiscount,
		UsageLimit:           req.UsageLimit,
		UsageLimitPerUser:    req.UsageLimitPerUser,
		StartDate:            req.StartDate,
		EndDate:              req.EndDate,
		ApplicableProducts:   req.ApplicableProducts,
		ExcludedProducts:     req.ExcludedProducts,
		ApplicableCategories: req.ApplicableCategories,
		Enabled:              enabled,
	})
	if err != nil {
		respondMarketingError(c, err, "coupon")
		return
	}

	response.Created(c, gin.H{"coupon": cp})
}

func (h *MarketingHandler) UpdateCoupon(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid coupon ID")
		return
	}

	var req struct {
		Code                 *string    `json:"code"`
		Type                 *string    `json:"type" binding:"omitempty,oneof=fixed percentage"`
		Value                *float64   `json:"value" binding:"omitempty,gt=0"`
		Description          *string    `json:"description"`
		MinAmount            *float64   `json:"min_amount"`
		MaxDiscount          *float64   `json:"max_discount"`
		UsageLimit           *int       `json:"usage_limit"`
		UsageLimitPerUser    *int       `json:"usage_limit_per_user"`
		StartDate            *time.Time `json:"start_date"`
		EndDate              *time.Time `json:"end_date"`
		ApplicableProducts   *string    `json:"applicable_products"`
		ExcludedProducts     *string    `json:"excluded_products"`
		ApplicableCategories *string    `json:"applicable_categories"`
		Enabled              *bool      `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	cp, err := h.marketingService.UpdateCouponAdmin(uint(id), service.CouponUpdateInput{
		Code:                 req.Code,
		Type:                 req.Type,
		Value:                req.Value,
		Description:          req.Description,
		MinAmount:            req.MinAmount,
		MaxDiscount:          req.MaxDiscount,
		UsageLimit:           req.UsageLimit,
		UsageLimitPerUser:    req.UsageLimitPerUser,
		StartDate:            req.StartDate,
		EndDate:              req.EndDate,
		ApplicableProducts:   req.ApplicableProducts,
		ExcludedProducts:     req.ExcludedProducts,
		ApplicableCategories: req.ApplicableCategories,
		Enabled:              req.Enabled,
	})
	if err != nil {
		respondMarketingError(c, err, "coupon")
		return
	}

	response.Success(c, gin.H{"coupon": cp})
}

func (h *MarketingHandler) DeleteCoupon(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid coupon ID")
		return
	}

	if err := h.marketingService.DeleteCouponAdmin(uint(id)); err != nil {
		respondMarketingError(c, err, "coupon")
		return
	}

	response.SuccessWithMessage(c, "deleted successfully", nil)
}

func (h *MarketingHandler) GetCouponStats(c *gin.Context) {
	stats, err := h.marketingService.GetCouponStats()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, stats)
}
