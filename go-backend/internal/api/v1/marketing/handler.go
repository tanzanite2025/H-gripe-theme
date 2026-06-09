package marketing

import (
	"net/http"
	"tanzanite/internal/domain/coupon"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	marketingService *service.MarketingService
}

func NewHandler(marketingService *service.MarketingService) *Handler {
	return &Handler{
		marketingService: marketingService,
	}
}

// Coupon 相关接口

// ListCoupons 获取优惠券列表
// @Summary 获取优惠券列表
// @Tags Marketing
// @Produce json
// @Success 200 {array} coupon.Coupon
// @Router /api/v1/marketing/coupons [get]
func (h *Handler) ListCoupons(c *gin.Context) {
	coupons, err := h.marketingService.GetActiveCoupons()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": coupons})
}

// ValidateCoupon 验证优惠券
// @Summary 验证优惠券
// @Tags Marketing
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "验证请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/marketing/coupons/validate [post]
func (h *Handler) ValidateCoupon(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Code   string  `json:"code" binding:"required"`
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coupon, discount, err := h.marketingService.ValidateCoupon(req.Code, userID.(uint), req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":    true,
		"coupon":   coupon,
		"discount": discount,
	})
}

// CreateCoupon 创建优惠券（管理员）
// @Summary 创建优惠券
// @Tags Marketing
// @Accept json
// @Produce json
// @Param coupon body coupon.Coupon true "优惠券信息"
// @Success 201 {object} coupon.Coupon
// @Router /api/v1/admin/marketing/coupons [post]
func (h *Handler) CreateCoupon(c *gin.Context) {
	var coupon coupon.Coupon
	if err := c.ShouldBindJSON(&coupon); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.marketingService.CreateCoupon(&coupon); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, coupon)
}

// Gift Card 相关接口

// CreateGiftCard 创建礼品卡
// @Summary 创建礼品卡
// @Tags Marketing
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "礼品卡信息"
// @Success 201 {object} coupon.GiftCard
// @Router /api/v1/marketing/gift-cards [post]
func (h *Handler) CreateGiftCard(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	card, err := h.marketingService.CreateGiftCard(userID.(uint), req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, card)
}

// UseGiftCard 使用礼品卡
// @Summary 使用礼品卡
// @Tags Marketing
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "使用请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/marketing/gift-cards/use [post]
func (h *Handler) UseGiftCard(c *gin.Context) {
	var req struct {
		Code    string  `json:"code" binding:"required"`
		Amount  float64 `json:"amount" binding:"required,gt=0"`
		OrderID uint    `json:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.marketingService.UseGiftCard(req.Code, req.Amount, req.OrderID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "gift card used successfully"})
}

// Loyalty 相关接口

// GetPoints 获取积分余额
// @Summary 获取积分余额
// @Tags Marketing
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/marketing/loyalty/points [get]
func (h *Handler) GetPoints(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	loyalty, err := h.marketingService.GetUserLoyalty(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "loyalty info not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"current_points": loyalty.AvailablePoints,
		"total_points":   loyalty.TotalPoints,
		"level_id":       loyalty.MemberLevelID,
	})
}

// CheckIn 每日签到
// @Summary 每日签到
// @Tags Marketing
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/marketing/loyalty/checkin [post]
func (h *Handler) CheckIn(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	points, err := h.marketingService.CheckIn(userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "check-in successful",
		"points":  points,
	})
}

// CreateReferral 创建推荐
// @Summary 创建推荐
// @Tags Marketing
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "推荐信息"
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/marketing/loyalty/referral [post]
func (h *Handler) CreateReferral(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		RefereeID uint `json:"referee_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.marketingService.CreateReferral(userID.(uint), req.RefereeID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "referral created"})
}

// SpendPoints 消费积分
// @Summary 消费积分
// @Tags Marketing
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "消费请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/marketing/loyalty/spend [post]
func (h *Handler) SpendPoints(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Points  int  `json:"points" binding:"required,gt=0"`
		OrderID uint `json:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.marketingService.SpendPoints(userID.(uint), req.Points, req.OrderID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "points spent successfully"})
}

// GetLoyaltyInfo 获取会员信息
// @Summary 获取会员信息
// @Tags Marketing
// @Produce json
// @Success 200 {object} loyalty.UserLoyalty
// @Router /api/v1/marketing/loyalty/info [get]
func (h *Handler) GetLoyaltyInfo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	loyalty, err := h.marketingService.GetUserLoyalty(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "loyalty info not found"})
		return
	}

	c.JSON(http.StatusOK, loyalty)
}

// ==========================================
// B端 (Admin) 管理接口
// ==========================================

// ListMemberLevels 获取所有会员等级
// @Summary 获取所有会员等级
// @Tags Admin/Loyalty
// @Produce json
// @Success 200 {array} loyalty.MemberLevel
// @Router /api/v1/admin/loyalty/levels [get]
func (h *Handler) ListMemberLevels(c *gin.Context) {
	levels, err := h.marketingService.ListMemberLevels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, levels)
}

// CreateMemberLevel 创建会员等级
// @Summary 创建会员等级
// @Tags Admin/Loyalty
// @Accept json
// @Produce json
// @Param level body loyalty.MemberLevel true "会员等级"
// @Success 201 {object} loyalty.MemberLevel
// @Router /api/v1/admin/loyalty/levels [post]
func (h *Handler) CreateMemberLevel(c *gin.Context) {
	var level loyalty.MemberLevel
	if err := c.ShouldBindJSON(&level); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.marketingService.CreateMemberLevel(&level); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, level)
}

// UpdateMemberLevel 更新会员等级
// @Summary 更新会员等级
// @Tags Admin/Loyalty
// @Accept json
// @Produce json
// @Param id path int true "等级ID"
// @Param level body loyalty.MemberLevel true "会员等级"
// @Success 200 {object} loyalty.MemberLevel
// @Router /api/v1/admin/loyalty/levels/{id} [put]
func (h *Handler) UpdateMemberLevel(c *gin.Context) {
	// 简单解析ID并绑定
	var level loyalty.MemberLevel
	if err := c.ShouldBindJSON(&level); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 忽略路径ID强行覆盖等细节，直接以传入对象的ID为准，在真实业务中应做一致性校验
	if err := h.marketingService.UpdateMemberLevel(&level); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, level)
}

// AdminAdjustPoints 管理员调整积分
// @Summary 手动调整用户积分
// @Tags Admin/Loyalty
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param request body map[string]interface{} true "调整请求 (points: int, reason: string)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/loyalty/users/{id}/adjust [post]
func (h *Handler) AdminAdjustPoints(c *gin.Context) {
	var req struct {
		Points int    `json:"points" binding:"required"`
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取路由参数 user_id
	var uriParams struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uriParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.marketingService.AdminAdjustPoints(uriParams.ID, req.Points, req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "points adjusted successfully"})
}

