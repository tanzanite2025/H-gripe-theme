package spoke

import (
	"commerce-platform/internal/service"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	spokeService *service.SpokeService
}

func NewHandler(spokeService *service.SpokeService) *Handler {
	return &Handler{spokeService: spokeService}
}

func (h *Handler) GetSpokeCalculatorEngineeringMetadata(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=3600")
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": h.spokeService.GetSpokeCalculatorEngineeringMetadata(),
	})
}

func (h *Handler) GetExport(c *gin.Context) {
	// Public callers must never receive CAD geometry or verified measurements.
	export, err := h.spokeService.GetPublicExport()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "spoke_export_error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, export)
}

// GetPublicCatalog is an explicit alias for the browser-facing projection.
func (h *Handler) GetPublicCatalog(c *gin.Context) {
	h.GetExport(c)
}

func (h *Handler) GetPublicResults(c *gin.Context) {
	results, err := h.spokeService.GetPublicRecordedResults()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "spoke_results_error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

func (h *Handler) ListHistory(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("per_page", "5"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 5
	}

	items, total, err := h.spokeService.ListUserHistory(userIDValue.(uint), c.Query("search"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "spoke_history_error", "message": err.Error()})
		return
	}
	for index := range items {
		items[index].UserID = nil
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"meta": gin.H{
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			"page":        page,
			"per_page":    pageSize,
		},
	})
}

type CalcRequest struct {
	RimID                                string   `json:"rimId"`
	HubID                                string   `json:"hubId"`
	WheelPosition                        string   `json:"wheelPosition" binding:"required"`
	TopologyID                           string   `json:"topologyId"`
	SpokeCount                           int      `json:"spokeCount"`
	Crossing                             int      `json:"crossing"`
	G3RimHoleSpacingAToBDegrees          float64  `json:"g3RimHoleSpacingAToBDegrees"`
	G3RimHoleSpacingBToADegrees          float64  `json:"g3RimHoleSpacingBToADegrees"`
	G3RimHoleSpacingAToNextGroupADegrees float64  `json:"g3RimHoleSpacingAToNextGroupADegrees"`
	RimOffsetMM                          float64  `json:"rimOffsetMm"`
	NippleType                           string   `json:"nippleType"`
	NippleLengthMM                       *float64 `json:"nippleLengthMm"`
	SpokeHeadType                        string   `json:"spokeHeadType"`
	SpokeHoleDiameterMM                  *float64 `json:"spokeHoleDiameterMm"`
	StraightPullTangentOffsetMM          *float64 `json:"straightPullTangentOffsetMm"`
	SpokeProfile                         string   `json:"spokeProfile"`
	TargetTensionN                       *float64 `json:"targetTensionN"`
	AlternatingDrillingOffsetMM          *float64 `json:"alternatingDrillingOffsetMm"`
	Interlacing                          bool     `json:"interlacing"`
	InterlaceCompensationMM              *float64 `json:"interlaceCompensationMm"`
	SpokeElongationCompensationMM        *float64 `json:"spokeElongationCompensationMm"`
	ERDMM                                *float64 `json:"erdMm"`
	LeftFlangeMM                         *float64 `json:"leftFlangeMm"`
	RightFlangeMM                        *float64 `json:"rightFlangeMm"`
	LeftFlangePCDMM                      *float64 `json:"leftFlangePcdMm"`
	RightFlangePCDMM                     *float64 `json:"rightFlangePcdMm"`
}

// UnmarshalJSON keeps the public camelCase contract while accepting the
// snake_case names used by the physical-build specification. This lets shop
// floor integrations send the field names from their existing machine API
// without making the browser payload or the rest of this API inconsistent.
func (r *CalcRequest) UnmarshalJSON(data []byte) error {
	type calcRequestAlias CalcRequest
	var camel calcRequestAlias
	if err := json.Unmarshal(data, &camel); err != nil {
		return err
	}

	var snake struct {
		RimID                                string   `json:"rim_id"`
		HubID                                string   `json:"hub_id"`
		WheelPosition                        string   `json:"wheel_position"`
		TopologyID                           string   `json:"topology_id"`
		SpokeCount                           *int     `json:"spoke_count"`
		Crossing                             *int     `json:"crossing"`
		G3RimHoleSpacingAToBDegrees          *float64 `json:"g3_rim_hole_spacing_a_to_b_degrees"`
		G3RimHoleSpacingBToADegrees          *float64 `json:"g3_rim_hole_spacing_b_to_a_degrees"`
		G3RimHoleSpacingAToNextGroupADegrees *float64 `json:"g3_rim_hole_spacing_a_to_next_group_a_degrees"`
		RimOffsetMM                          *float64 `json:"rim_offset_mm"`
		NippleType                           string   `json:"nipple_type"`
		NippleLengthMM                       *float64 `json:"nipple_length_mm"`
		SpokeHeadType                        string   `json:"spoke_head_type"`
		SpokeHoleDiameterMM                  *float64 `json:"spoke_hole_diameter_mm"`
		StraightPullTangentOffsetMM          *float64 `json:"straight_pull_tangent_offset_mm"`
		SpokeProfile                         string   `json:"spoke_profile"`
		TargetTensionN                       *float64 `json:"target_tension_n"`
		AlternatingDrillingOffsetMM          *float64 `json:"alternating_drilling_offset_mm"`
		InterlaceCompensationMM              *float64 `json:"interlace_compensation_mm"`
		SpokeElongationCompensationMM        *float64 `json:"spoke_elongation_compensation_mm"`
		ERDMM                                *float64 `json:"erd_mm"`
		LeftFlangeMM                         *float64 `json:"left_flange_mm"`
		RightFlangeMM                        *float64 `json:"right_flange_mm"`
		LeftFlangePCDMM                      *float64 `json:"left_flange_pcd_mm"`
		RightFlangePCDMM                     *float64 `json:"right_flange_pcd_mm"`
	}
	if err := json.Unmarshal(data, &snake); err != nil {
		return err
	}

	*r = CalcRequest(camel)
	if r.RimID == "" {
		r.RimID = snake.RimID
	}
	if r.HubID == "" {
		r.HubID = snake.HubID
	}
	if r.WheelPosition == "" {
		r.WheelPosition = snake.WheelPosition
	}
	if r.TopologyID == "" {
		r.TopologyID = snake.TopologyID
	}
	if snake.SpokeCount != nil {
		r.SpokeCount = *snake.SpokeCount
	}
	if snake.Crossing != nil {
		r.Crossing = *snake.Crossing
	}
	if snake.G3RimHoleSpacingAToBDegrees != nil {
		r.G3RimHoleSpacingAToBDegrees = *snake.G3RimHoleSpacingAToBDegrees
	}
	if snake.G3RimHoleSpacingBToADegrees != nil {
		r.G3RimHoleSpacingBToADegrees = *snake.G3RimHoleSpacingBToADegrees
	}
	if snake.G3RimHoleSpacingAToNextGroupADegrees != nil {
		r.G3RimHoleSpacingAToNextGroupADegrees = *snake.G3RimHoleSpacingAToNextGroupADegrees
	}
	if snake.RimOffsetMM != nil {
		r.RimOffsetMM = *snake.RimOffsetMM
	}
	if r.NippleType == "" {
		r.NippleType = snake.NippleType
	}
	if r.NippleLengthMM == nil {
		r.NippleLengthMM = snake.NippleLengthMM
	}
	if r.SpokeHeadType == "" {
		r.SpokeHeadType = snake.SpokeHeadType
	}
	if r.SpokeHoleDiameterMM == nil {
		r.SpokeHoleDiameterMM = snake.SpokeHoleDiameterMM
	}
	if r.StraightPullTangentOffsetMM == nil {
		r.StraightPullTangentOffsetMM = snake.StraightPullTangentOffsetMM
	}
	if r.SpokeProfile == "" {
		r.SpokeProfile = snake.SpokeProfile
	}
	if r.TargetTensionN == nil {
		r.TargetTensionN = snake.TargetTensionN
	}
	if r.AlternatingDrillingOffsetMM == nil {
		r.AlternatingDrillingOffsetMM = snake.AlternatingDrillingOffsetMM
	}
	if r.InterlaceCompensationMM == nil {
		r.InterlaceCompensationMM = snake.InterlaceCompensationMM
	}
	if r.SpokeElongationCompensationMM == nil {
		r.SpokeElongationCompensationMM = snake.SpokeElongationCompensationMM
	}
	if r.ERDMM == nil {
		r.ERDMM = snake.ERDMM
	}
	if r.LeftFlangeMM == nil {
		r.LeftFlangeMM = snake.LeftFlangeMM
	}
	if r.RightFlangeMM == nil {
		r.RightFlangeMM = snake.RightFlangeMM
	}
	if r.LeftFlangePCDMM == nil {
		r.LeftFlangePCDMM = snake.LeftFlangePCDMM
	}
	if r.RightFlangePCDMM == nil {
		r.RightFlangePCDMM = snake.RightFlangePCDMM
	}
	return nil
}

func (h *Handler) Calculate(c *gin.Context) {
	var req CalcRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	var userID *uint
	if value, exists := c.Get("user_id"); exists {
		if id, ok := value.(uint); ok && id > 0 {
			userID = &id
		}
	}
	result, err := h.spokeService.Calculate(service.SpokeCalculationInput{
		RimID:                                req.RimID,
		HubID:                                req.HubID,
		WheelPosition:                        req.WheelPosition,
		TopologyID:                           req.TopologyID,
		SpokeCount:                           req.SpokeCount,
		Crossing:                             req.Crossing,
		G3RimHoleSpacingAToBDegrees:          req.G3RimHoleSpacingAToBDegrees,
		G3RimHoleSpacingBToADegrees:          req.G3RimHoleSpacingBToADegrees,
		G3RimHoleSpacingAToNextGroupADegrees: req.G3RimHoleSpacingAToNextGroupADegrees,
		RimOffsetMM:                          req.RimOffsetMM,
		NippleType:                           req.NippleType,
		NippleLengthMM:                       req.NippleLengthMM,
		SpokeHeadType:                        req.SpokeHeadType,
		SpokeHoleDiameterMM:                  req.SpokeHoleDiameterMM,
		StraightPullTangentOffsetMM:          req.StraightPullTangentOffsetMM,
		SpokeProfile:                         req.SpokeProfile,
		TargetTensionN:                       req.TargetTensionN,
		AlternatingDrillingOffsetMM:          req.AlternatingDrillingOffsetMM,
		Interlacing:                          req.Interlacing,
		InterlaceCompensationMM:              req.InterlaceCompensationMM,
		SpokeElongationCompensationMM:        req.SpokeElongationCompensationMM,
		ERDMM:                                req.ERDMM,
		LeftFlangeMM:                         req.LeftFlangeMM,
		RightFlangeMM:                        req.RightFlangeMM,
		LeftFlangePCDMM:                      req.LeftFlangePCDMM,
		RightFlangePCDMM:                     req.RightFlangePCDMM,
		UserID:                               userID,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSpokeGeometryNotFound):
			c.JSON(http.StatusBadRequest, gin.H{"error": "not_found", "message": "Unknown rim or hub geometry"})
		case errors.Is(err, service.ErrSpokeRimGeometryMissing):
			c.JSON(http.StatusBadRequest, gin.H{"error": "not_found", "message": "Rim geometry not available for requested model"})
		case errors.Is(err, service.ErrSpokeHubGeometryMissing):
			c.JSON(http.StatusBadRequest, gin.H{"error": "not_found", "message": "Hub geometry not available for requested position"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}
