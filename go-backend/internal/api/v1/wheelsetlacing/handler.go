package wheelsetlacing

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	wheelsetlacingdomain "commerce-platform/internal/domain/wheelsetlacing"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

const maxValidationBodyBytes = 64 * 1024

type Handler struct {
	service *service.WheelsetLacingService
}

func NewHandler(wheelsetLacingService *service.WheelsetLacingService) *Handler {
	if wheelsetLacingService == nil {
		wheelsetLacingService = service.NewWheelsetLacingService()
	}
	return &Handler{service: wheelsetLacingService}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/topologies", h.ListTopologies)
	group.POST("/validate", h.Validate)
}

// ListTopologies returns the immutable hole-mapping contract used by the
// independent wheelset lacing reference page. It intentionally contains no
// ERD/PCD, coordinate, spoke-length, tension, stiffness, efficiency, or
// assembly-safety calculation.
func (h *Handler) ListTopologies(c *gin.Context) {
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
func (h *Handler) Validate(c *gin.Context) {
	request, err := decodeValidateRequest(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_WHEELSET_LACING_REQUEST", "invalid wheelset lacing validation request", err)
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
		respondError(c, status, code, message, err)
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

func decodeValidateRequest(c *gin.Context) (wheelsetlacingdomain.ValidateRequest, error) {
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

func respondError(c *gin.Context, status int, code, message string, err error) {
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
