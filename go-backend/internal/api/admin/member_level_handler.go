package admin

import (
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"
	"strconv"

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
		Name         string  `json:"name" binding:"required"`
		MinPoints    *int    `json:"min_points" binding:"required"`
		MaxPoints    *int    `json:"max_points" binding:"required"`
		DiscountRate float64 `json:"discount_rate"`
		Benefits     string  `json:"benefits"`
		Icon         string  `json:"icon"`
		Color        string  `json:"color"`
		SortOrder    int     `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	level, err := h.marketingService.CreateMemberLevelAdmin(service.MemberLevelCreateInput{
		Name:         req.Name,
		MinPoints:    *req.MinPoints,
		MaxPoints:    *req.MaxPoints,
		DiscountRate: req.DiscountRate,
		Benefits:     req.Benefits,
		Icon:         req.Icon,
		Color:        req.Color,
		SortOrder:    req.SortOrder,
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
		Name         *string  `json:"name"`
		MinPoints    *int     `json:"min_points"`
		MaxPoints    *int     `json:"max_points"`
		DiscountRate *float64 `json:"discount_rate"`
		Benefits     *string  `json:"benefits"`
		Icon         *string  `json:"icon"`
		Color        *string  `json:"color"`
		SortOrder    *int     `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	level, err := h.marketingService.UpdateMemberLevelAdmin(uint(id), service.MemberLevelUpdateInput{
		Name:         req.Name,
		MinPoints:    req.MinPoints,
		MaxPoints:    req.MaxPoints,
		DiscountRate: req.DiscountRate,
		Benefits:     req.Benefits,
		Icon:         req.Icon,
		Color:        req.Color,
		SortOrder:    req.SortOrder,
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
