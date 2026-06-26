package admin

import (
	"strconv"
	"tanzanite/internal/domain/loyalty"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ============ 会员等级管理 ============

// ListMemberLevels 获取会员等级列表
func (h *MarketingHandler) ListMemberLevels(c *gin.Context) {
	levels, err := h.loyaltyRepo.FindAllMemberLevels()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"levels": levels})
}

// GetMemberLevel 获取会员等级详情
func (h *MarketingHandler) GetMemberLevel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "无效的等级ID")
		return
	}

	level, err := h.loyaltyRepo.FindMemberLevelByID(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "会员等级")
		return
	}

	response.Success(c, gin.H{"level": level})
}

// CreateMemberLevel 创建会员等级
func (h *MarketingHandler) CreateMemberLevel(c *gin.Context) {
	var req struct {
		Name             string  `json:"name" binding:"required"`
		MinPoints        int     `json:"min_points" binding:"required"`
		MaxPoints        int     `json:"max_points" binding:"required"`
		DiscountRate     float64 `json:"discount_rate"`
		PointsMultiplier float64 `json:"points_multiplier"`
		Benefits         string  `json:"benefits"`
		Icon             string  `json:"icon"`
		Color            string  `json:"color"`
		SortOrder        int     `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	level := &loyalty.MemberLevel{
		Name:             req.Name,
		MinPoints:        req.MinPoints,
		MaxPoints:        req.MaxPoints,
		DiscountRate:     req.DiscountRate,
		PointsMultiplier: req.PointsMultiplier,
		Benefits:         req.Benefits,
		Icon:             req.Icon,
		Color:            req.Color,
		SortOrder:        req.SortOrder,
	}

	if err := h.loyaltyRepo.CreateMemberLevel(level); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Created(c, gin.H{"level": level})
}

// UpdateMemberLevel 更新会员等级
func (h *MarketingHandler) UpdateMemberLevel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "无效的等级ID")
		return
	}

	level, err := h.loyaltyRepo.FindMemberLevelByID(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "会员等级")
		return
	}

	var req struct {
		Name             string  `json:"name"`
		MinPoints        int     `json:"min_points"`
		MaxPoints        int     `json:"max_points"`
		DiscountRate     float64 `json:"discount_rate"`
		PointsMultiplier float64 `json:"points_multiplier"`
		Benefits         string  `json:"benefits"`
		Icon             string  `json:"icon"`
		Color            string  `json:"color"`
		SortOrder        int     `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	// 更新字段
	if req.Name != "" {
		level.Name = req.Name
	}
	if req.MinPoints > 0 {
		level.MinPoints = req.MinPoints
	}
	if req.MaxPoints > 0 {
		level.MaxPoints = req.MaxPoints
	}
	level.DiscountRate = req.DiscountRate
	level.PointsMultiplier = req.PointsMultiplier
	level.Benefits = req.Benefits
	level.Icon = req.Icon
	level.Color = req.Color
	level.SortOrder = req.SortOrder

	if err := h.loyaltyRepo.UpdateMemberLevel(level); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"level": level})
}

// DeleteMemberLevel 删除会员等级
func (h *MarketingHandler) DeleteMemberLevel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "无效的等级ID")
		return
	}

	if err := h.loyaltyRepo.DeleteMemberLevel(uint(id)); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}
