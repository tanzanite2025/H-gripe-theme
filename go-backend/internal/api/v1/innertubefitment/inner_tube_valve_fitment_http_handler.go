package innertubefitment

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	domain "commerce-platform/internal/domain/innertubefitment"

	"github.com/gin-gonic/gin"
)

const maxInnerTubeValveFitmentRequestBodyBytes = 16 * 1024

type InnerTubeValveFitmentHTTPHandler struct{}

func NewInnerTubeValveFitmentHTTPHandler() *InnerTubeValveFitmentHTTPHandler {
	return &InnerTubeValveFitmentHTTPHandler{}
}

type innerTubeValveFitmentSolveRequest struct {
	RimDepthMM            int                `json:"rim_depth_mm"`
	Mode                  domain.FitmentMode `json:"mode"`
	ValveLengthMM         *int               `json:"valve_length_mm,omitempty"`
	ExtenderLengthMM      *int               `json:"extender_length_mm,omitempty"`
	PumpHeadGripDepthMM   *int               `json:"pump_head_grip_depth_mm,omitempty"`
	RimDepthUncertaintyMM *int               `json:"rim_depth_uncertainty_mm,omitempty"`
}

func (h *InnerTubeValveFitmentHTTPHandler) RegisterInnerTubeValveFitmentHTTPRoutes(group *gin.RouterGroup) {
	group.GET("/matrix", h.GetInnerTubeValveFitmentMatrix)
	group.POST("/solve", h.SolveInnerTubeValveFitment)
}

func (h *InnerTubeValveFitmentHTTPHandler) GetInnerTubeValveFitmentMatrix(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=86400")
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": domain.GetInnerTubeValveFitmentMatrixMetadata()})
}

func (h *InnerTubeValveFitmentHTTPHandler) SolveInnerTubeValveFitment(c *gin.Context) {
	if contentType := c.GetHeader("Content-Type"); contentType != "" && !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		writeInnerTubeValveFitmentErrorResponse(c, http.StatusBadRequest, "INVALID_JSON", "Content-Type must be application/json", "content_type")
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxInnerTubeValveFitmentRequestBodyBytes)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	var request innerTubeValveFitmentSolveRequest
	if err := decoder.Decode(&request); err != nil {
		writeInnerTubeValveFitmentErrorResponse(c, http.StatusBadRequest, "INVALID_JSON", "Request body must be a single JSON object", "")
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		writeInnerTubeValveFitmentErrorResponse(c, http.StatusBadRequest, "INVALID_JSON", "Request body must contain exactly one JSON object", "")
		return
	}
	if request.PumpHeadGripDepthMM != nil {
		if err := domain.ValidateInnerTubePumpHeadGripDepth(*request.PumpHeadGripDepthMM); err != nil {
			status, code, field := innerTubeValveFitmentErrorDetails(err)
			writeInnerTubeValveFitmentErrorResponse(c, status, code, err.Error(), field)
			return
		}
	}

	result, err := domain.SolveInnerTubeValveFitment(domain.SolveInput{
		RimDepthMM: request.RimDepthMM, Mode: request.Mode,
		ValveLengthMM: request.ValveLengthMM, ExtenderLengthMM: request.ExtenderLengthMM,
		PumpHeadGripDepthMM:   valueOrInnerTubeFitmentDefault(request.PumpHeadGripDepthMM, domain.DefaultPumpHeadGripDepthMM),
		RimDepthUncertaintyMM: valueOrInnerTubeFitmentDefault(request.RimDepthUncertaintyMM, domain.DefaultRimDepthUncertaintyMM),
	})
	if err != nil {
		status, code, field := innerTubeValveFitmentErrorDetails(err)
		writeInnerTubeValveFitmentErrorResponse(c, status, code, err.Error(), field)
		return
	}

	c.Header("Cache-Control", "no-store, max-age=0")
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

func innerTubeValveFitmentErrorDetails(err error) (int, string, string) {
	switch {
	case errors.Is(err, domain.ErrInvalidRimDepth):
		return http.StatusUnprocessableEntity, "OUT_OF_RANGE", "rim_depth_mm"
	case errors.Is(err, domain.ErrInvalidFitmentMode):
		return http.StatusBadRequest, "INVALID_FIELD", "mode"
	case errors.Is(err, domain.ErrMissingManualLength):
		return http.StatusBadRequest, "INVALID_FIELD", "valve_length_mm"
	case errors.Is(err, domain.ErrInvalidPumpHeadGripDepth):
		return http.StatusUnprocessableEntity, "OUT_OF_RANGE", "pump_head_grip_depth_mm"
	case errors.Is(err, domain.ErrInvalidRimDepthUncertainty):
		return http.StatusUnprocessableEntity, "OUT_OF_RANGE", "rim_depth_uncertainty_mm"
	case errors.Is(err, domain.ErrInvalidValveLength):
		return http.StatusUnprocessableEntity, "OUT_OF_RANGE", "valve_length_mm"
	case errors.Is(err, domain.ErrInvalidExtenderLength):
		return http.StatusUnprocessableEntity, "OUT_OF_RANGE", "extender_length_mm"
	default:
		return http.StatusBadRequest, "INVALID_FIELD", ""
	}
}

func valueOrInnerTubeFitmentDefault(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}

func writeInnerTubeValveFitmentErrorResponse(c *gin.Context, status int, code, message, field string) {
	c.JSON(status, gin.H{"code": code, "error": message, "message": message, "field": field})
}
