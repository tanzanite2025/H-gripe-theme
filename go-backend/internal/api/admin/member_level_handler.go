package admin

import (
	"strconv"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *MarketingHandler) ListMemberLevels(c *gin.Context) {
	levels, err := h.marketingService.ListMemberLevels()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"levels": levels})
}

func (h *MarketingHandler) GetMemberLevel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid member level ID")
		return
	}

	level, err := h.marketingService.GetMemberLevel(uint(id))
	if err != nil {
		respondMarketingError(c, err, "member level")
		return
	}

	response.Success(c, gin.H{"level": level})
}

func (h *MarketingHandler) CreateMemberLevel(c *gin.Context) {
	var req struct {
		Name             string   `json:"name" binding:"required"`
		MinPoints        *int     `json:"min_points" binding:"required"`
		MaxPoints        *int     `json:"max_points" binding:"required"`
		DiscountRate     float64  `json:"discount_rate"`
		PointsMultiplier *float64 `json:"points_multiplier"`
		Benefits         string   `json:"benefits"`
		Icon             string   `json:"icon"`
		Color            string   `json:"color"`
		SortOrder        int      `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	pointsMultiplier := 1.0
	if req.PointsMultiplier != nil {
		pointsMultiplier = *req.PointsMultiplier
	}

	level, err := h.marketingService.CreateMemberLevelAdmin(service.MemberLevelCreateInput{
		Name:             req.Name,
		MinPoints:        *req.MinPoints,
		MaxPoints:        *req.MaxPoints,
		DiscountRate:     req.DiscountRate,
		PointsMultiplier: pointsMultiplier,
		Benefits:         req.Benefits,
		Icon:             req.Icon,
		Color:            req.Color,
		SortOrder:        req.SortOrder,
	})
	if err != nil {
		respondMarketingError(c, err, "member level")
		return
	}

	response.Created(c, gin.H{"level": level})
}

func (h *MarketingHandler) UpdateMemberLevel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid member level ID")
		return
	}

	var req struct {
		Name             *string  `json:"name"`
		MinPoints        *int     `json:"min_points"`
		MaxPoints        *int     `json:"max_points"`
		DiscountRate     *float64 `json:"discount_rate"`
		PointsMultiplier *float64 `json:"points_multiplier"`
		Benefits         *string  `json:"benefits"`
		Icon             *string  `json:"icon"`
		Color            *string  `json:"color"`
		SortOrder        *int     `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	level, err := h.marketingService.UpdateMemberLevelAdmin(uint(id), service.MemberLevelUpdateInput{
		Name:             req.Name,
		MinPoints:        req.MinPoints,
		MaxPoints:        req.MaxPoints,
		DiscountRate:     req.DiscountRate,
		PointsMultiplier: req.PointsMultiplier,
		Benefits:         req.Benefits,
		Icon:             req.Icon,
		Color:            req.Color,
		SortOrder:        req.SortOrder,
	})
	if err != nil {
		respondMarketingError(c, err, "member level")
		return
	}

	response.Success(c, gin.H{"level": level})
}

func (h *MarketingHandler) DeleteMemberLevel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid member level ID")
		return
	}

	if err := h.marketingService.DeleteMemberLevelAdmin(uint(id)); err != nil {
		respondMarketingError(c, err, "member level")
		return
	}

	response.SuccessWithMessage(c, "deleted successfully", nil)
}
