package wheelsetlacing

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	wheelsetlacingdomain "commerce-platform/internal/domain/wheelsetlacing"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

const maxValidationBodyBytes = 64 * 1024

type WheelsetLacingEngineeringHTTPHandler struct {
	service *service.WheelsetLacingService
}

func NewWheelsetLacingEngineeringHTTPHandler(wheelsetLacingService *service.WheelsetLacingService) *WheelsetLacingEngineeringHTTPHandler {
	if wheelsetLacingService == nil {
		wheelsetLacingService = service.NewWheelsetLacingService()
	}
	return &WheelsetLacingEngineeringHTTPHandler{service: wheelsetLacingService}
}

func (h *WheelsetLacingEngineeringHTTPHandler) RegisterWheelsetLacingEngineeringRoutes(group *gin.RouterGroup) {
	group.GET("/topologies", h.ListWheelsetLacingTopologies)
	group.POST("/validate", h.ValidateWheelsetLacingTopologySelection)
	group.GET("/display-geometry", h.GetWheelsetLacingDisplayGeometryProjectionForServerRenderedReferencePage)
	group.POST("/display-geometry", h.HandleWheelsetLacingDisplayGeometryCalculation)
}

// GetWheelsetLacingDisplayGeometryProjectionForServerRenderedReferencePage
// returns the canonical default canvas and axial-profile geometry for SSR and
// GEO. The GET contract accepts only a topology identifier and uses
// domain-owned display coordinates plus reference flange offsets; it never
// accepts ERD, PCD, spoke length, tension, or safety inputs.
func (h *WheelsetLacingEngineeringHTTPHandler) GetWheelsetLacingDisplayGeometryProjectionForServerRenderedReferencePage(c *gin.Context) {
	request, err := decodeWheelsetLacingDefaultDisplayGeometryQueryParameters(c)
	if err != nil {
		WriteWheelsetLacingErrorResponse(c, http.StatusBadRequest, "INVALID_WHEELSET_LACING_GEOMETRY_QUERY", "invalid wheelset lacing display geometry query", err)
		return
	}
	c.Header("Cache-Control", "public, max-age=86400")
	h.calculateAndWriteWheelsetLacingDisplayGeometryResponse(c, request)
}

// HandleWheelsetLacingDisplayGeometryCalculation validates a canonical
// topology and returns backend-generated canvas coordinates, an axial flange
// profile, and geometry-only projection metrics. The radii are display
// coordinates; flange offsets are profile reference inputs in millimetres.
// None of these values represent ERD, PCD, spoke length, stiffness, tension,
// efficiency, or safety conclusions.
func (h *WheelsetLacingEngineeringHTTPHandler) HandleWheelsetLacingDisplayGeometryCalculation(c *gin.Context) {
	request, err := decodeWheelsetLacingDisplayGeometryProjectionRequest(c)
	if err != nil {
		WriteWheelsetLacingErrorResponse(c, http.StatusBadRequest, "INVALID_WHEELSET_LACING_GEOMETRY_REQUEST", "invalid wheelset lacing display geometry request", err)
		return
	}
	h.calculateAndWriteWheelsetLacingDisplayGeometryResponse(c, request)
}

func (h *WheelsetLacingEngineeringHTTPHandler) calculateAndWriteWheelsetLacingDisplayGeometryResponse(c *gin.Context, request wheelsetlacingdomain.DisplayGeometryProjectionRequest) {
	topology, err := h.service.Validate(wheelsetlacingdomain.ValidateRequest{TopologyID: request.TopologyID})
	if err != nil {
		status := http.StatusUnprocessableEntity
		code := "WHEELSET_LACING_VALIDATION_FAILED"
		switch {
		case errors.Is(err, wheelsetlacingdomain.ErrUnknownTopology):
			status = http.StatusNotFound
			code = "UNKNOWN_WHEELSET_LACING_TOPOLOGY"
		case errors.Is(err, wheelsetlacingdomain.ErrInvalidTopology):
			status = http.StatusInternalServerError
			code = "WHEELSET_LACING_CONTRACT_ERROR"
		}
		WriteWheelsetLacingErrorResponse(c, status, code, "wheelset lacing display geometry validation failed", err)
		return
	}
	result, err := wheelsetlacingdomain.CalculateWheelsetLacingDisplayGeometryProjection(request, topology)
	if err != nil {
		status := http.StatusUnprocessableEntity
		if errors.Is(err, wheelsetlacingdomain.ErrInvalidTopology) {
			status = http.StatusInternalServerError
		}
		WriteWheelsetLacingErrorResponse(c, status, "WHEELSET_LACING_GEOMETRY_CALCULATION_FAILED", "wheelset lacing display geometry calculation failed", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": result,
	})
}

// ListWheelsetLacingTopologies returns the immutable hole-mapping contract used by the
// independent wheelset lacing reference page. It intentionally contains no
// ERD/PCD, coordinate, spoke-length, tension, stiffness, efficiency, or
// assembly-safety calculation.
func (h *WheelsetLacingEngineeringHTTPHandler) ListWheelsetLacingTopologies(c *gin.Context) {
	etag := h.service.ETag()
	c.Header("Cache-Control", "public, max-age=86400")
	if etag != "" {
		c.Header("ETag", etag)
		if strings.TrimSpace(c.GetHeader("If-None-Match")) == etag {
			c.Status(http.StatusNotModified)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"topologies":       h.service.List(),
			"contract_version": service.WheelsetLacingContractVersion,
			"scope": gin.H{
				"physical_inputs":        false,
				"display_coordinates":    false,
				"erd_pcd_representation": "diameter",
				"length_unit":            "mm",
				"angle_unit":             "deg",
			},
			"limitations": []string{
				"This endpoint validates discrete hole topology only.",
				"It does not calculate spoke length, tension, stiffness, efficiency, strength, or assembly safety.",
			},
		},
	})
}

// Validate checks a canonical topology ID and optional consistency fields. It
// does not accept physical dimensions and never calls the spoke calculator.
func (h *WheelsetLacingEngineeringHTTPHandler) ValidateWheelsetLacingTopologySelection(c *gin.Context) {
	request, err := decodeWheelsetLacingTopologyValidationRequest(c)
	if err != nil {
		WriteWheelsetLacingErrorResponse(c, http.StatusBadRequest, "INVALID_WHEELSET_LACING_REQUEST", "invalid wheelset lacing validation request", err)
		return
	}
	topology, err := h.service.Validate(request)
	if err != nil {
		status := http.StatusUnprocessableEntity
		code := "WHEELSET_LACING_VALIDATION_FAILED"
		switch {
		case errors.Is(err, wheelsetlacingdomain.ErrInvalidRequest):
			status = http.StatusBadRequest
			code = "INVALID_WHEELSET_LACING_REQUEST"
		case errors.Is(err, wheelsetlacingdomain.ErrUnknownTopology):
			status = http.StatusNotFound
			code = "UNKNOWN_WHEELSET_LACING_TOPOLOGY"
		case errors.Is(err, wheelsetlacingdomain.ErrInvalidTopology):
			status = http.StatusInternalServerError
			code = "WHEELSET_LACING_CONTRACT_ERROR"
		case errors.Is(err, wheelsetlacingdomain.ErrTopologyMismatch):
			status = http.StatusUnprocessableEntity
			code = "TOPOLOGY_SELECTION_MISMATCH"
		}
		message := "wheelset lacing validation failed"
		if errors.Is(err, wheelsetlacingdomain.ErrInvalidTopology) {
			message = "[CRITICAL] wheelset lacing contract is invalid"
		}
		WriteWheelsetLacingErrorResponse(c, status, code, message, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"valid":            true,
			"topology":         topology,
			"contract_version": service.WheelsetLacingContractVersion,
		},
	})
}

func decodeWheelsetLacingTopologyValidationRequest(c *gin.Context) (wheelsetlacingdomain.ValidateRequest, error) {
	if c.Request == nil || c.Request.Body == nil {
		return wheelsetlacingdomain.ValidateRequest{}, errors.New("request body is required")
	}
	decoder := json.NewDecoder(io.LimitReader(c.Request.Body, maxValidationBodyBytes))
	decoder.DisallowUnknownFields()
	var request wheelsetlacingdomain.ValidateRequest
	if err := decoder.Decode(&request); err != nil {
		return wheelsetlacingdomain.ValidateRequest{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return wheelsetlacingdomain.ValidateRequest{}, errors.New("request body must contain one JSON object")
		}
		return wheelsetlacingdomain.ValidateRequest{}, err
	}
	return request, nil
}

func decodeWheelsetLacingDisplayGeometryProjectionRequest(c *gin.Context) (wheelsetlacingdomain.DisplayGeometryProjectionRequest, error) {
	if c.Request == nil || c.Request.Body == nil {
		return wheelsetlacingdomain.DisplayGeometryProjectionRequest{}, errors.New("request body is required")
	}
	decoder := json.NewDecoder(io.LimitReader(c.Request.Body, maxValidationBodyBytes))
	decoder.DisallowUnknownFields()
	var request wheelsetlacingdomain.DisplayGeometryProjectionRequest
	if err := decoder.Decode(&request); err != nil {
		return wheelsetlacingdomain.DisplayGeometryProjectionRequest{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return wheelsetlacingdomain.DisplayGeometryProjectionRequest{}, errors.New("request body must contain one JSON object")
		}
		return wheelsetlacingdomain.DisplayGeometryProjectionRequest{}, err
	}
	return request, nil
}

func decodeWheelsetLacingDefaultDisplayGeometryQueryParameters(c *gin.Context) (wheelsetlacingdomain.DisplayGeometryProjectionRequest, error) {
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return wheelsetlacingdomain.DisplayGeometryProjectionRequest{}, errors.New("request URL is required")
	}
	query := c.Request.URL.Query()
	for key := range query {
		if key != "topology_id" {
			return wheelsetlacingdomain.DisplayGeometryProjectionRequest{}, fmt.Errorf("unsupported query parameter %q", key)
		}
	}
	topologyID := strings.TrimSpace(query.Get("topology_id"))
	if topologyID == "" {
		return wheelsetlacingdomain.DisplayGeometryProjectionRequest{}, errors.New("topology_id query parameter is required")
	}
	return wheelsetlacingdomain.NewWheelsetLacingDefaultDisplayGeometryProjectionRequest(topologyID), nil
}

func WriteWheelsetLacingErrorResponse(c *gin.Context, status int, code, message string, err error) {
	detail := ""
	if err != nil {
		detail = err.Error()
	}
	if detail != "" {
		message += ": " + detail
	}
	c.JSON(status, gin.H{
		"code":    code,
		"message": message,
		"error":   detail,
	})
}
