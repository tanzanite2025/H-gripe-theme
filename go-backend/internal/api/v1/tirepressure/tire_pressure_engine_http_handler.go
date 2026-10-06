package tirepressure

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strings"

	"commerce-platform/internal/domain/tirepressure"

	"github.com/gin-gonic/gin"
)

const maxRequestBodyBytes = 16 * 1024

type TirePressureEngineeringCalculationHTTPHandler struct{}

func NewTirePressureEngineeringCalculationHandler() *TirePressureEngineeringCalculationHTTPHandler {
	return &TirePressureEngineeringCalculationHTTPHandler{}
}

type TirePressureCalculationErrorResponse struct {
	Code  string `json:"code"`
	Error string `json:"error"`
	Field string `json:"field,omitempty"`
}

func (h *TirePressureEngineeringCalculationHTTPHandler) RegisterTirePressureEngineeringCalculationRoutes(group *gin.RouterGroup) {
	group.POST("/solve", h.HandleTirePressureMetadataSolve)
	group.POST("/dynamics", h.HandleTirePressureGroundFrameDynamicsSolve)
}

// HandleTirePressureGroundFrameDynamicsSolve exposes the transparent first-order force model without pretending
// that manufacturer pressure limits or a calibrated pressure recommendation exist.
func (h *TirePressureEngineeringCalculationHTTPHandler) HandleTirePressureGroundFrameDynamicsSolve(c *gin.Context) {
	if contentType := c.GetHeader("Content-Type"); contentType != "" && !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		WriteTirePressureCalculationErrorResponse(c, http.StatusBadRequest, "INVALID_JSON", "Content-Type must be application/json", "content_type")
		return
	}
	var req tirepressure.Request
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		WriteTirePressureCalculationErrorResponse(c, http.StatusBadRequest, "INVALID_JSON", "Request body must be valid JSON", "")
		return
	}
	// Rim system is metadata for manufacturer-limit/compatibility checks. The
	// dynamics page does not calculate a rim-specific force term, so omitted
	// values use the explicit unknown state for backward-compatible validation.
	if req.RimSystem == "" {
		req.RimSystem = tirepressure.RimUnknown
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		WriteTirePressureCalculationErrorResponse(c, http.StatusBadRequest, "INVALID_JSON", "Request body must contain exactly one JSON object", "")
		return
	}
	if err := tirepressure.ValidateTirePressureDynamicCalculationRequest(req); err != nil {
		var validationErr *tirepressure.ValidationError
		if errors.As(err, &validationErr) {
			status := http.StatusBadRequest
			if validationErr.Code == "OUT_OF_RANGE" {
				status = http.StatusUnprocessableEntity
			}
			WriteTirePressureCalculationErrorResponse(c, status, validationErr.Code, validationErr.Reason, validationErr.Field)
			return
		}
		WriteTirePressureCalculationErrorResponse(c, http.StatusBadRequest, "INVALID_FIELD", "Request fields are invalid", "")
		return
	}
	loads, err := tirepressure.ResolveTirePressureWheelLoads(req)
	if err != nil {
		if errors.Is(err, tirepressure.ErrLoadMismatch) {
			WriteTirePressureCalculationErrorResponse(c, http.StatusUnprocessableEntity, "LOAD_TOTAL_MISMATCH", "Front and rear loads must sum to rider plus bike mass", "front_load_kg")
			return
		}
		WriteTirePressureCalculationErrorResponse(c, http.StatusBadRequest, "INVALID_FIELD", "Request fields are invalid", "")
		return
	}
	dynamics, err := tirepressure.CalculateTirePressureGroundFrameCorneringDynamics(req, loads)
	if err != nil {
		WriteTirePressureCalculationErrorResponse(c, http.StatusBadRequest, "INVALID_FIELD", "Dynamic inputs are invalid", "")
		return
	}
	warnings := []string{
		"Demo estimate only; not a pressure recommendation or safety guarantee.",
		"The data required for this calculation currently lacks complete and accurate support; the result is for relative change demonstration only.",
	}
	if req.WetPressureDemonstrationEnabled {
		warnings = append(warnings, "Wet friction retention and same-area equivalent pressure use an explicit first-order demonstration proxy; the area comparison does not claim to restore lost grip or represent a calibrated tire-road measurement.")
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"model_version": tirepressure.DynamicsModelVersion,
		"load_source":   loads.Source,
		"front_load_kg": loads.FrontKg,
		"rear_load_kg":  loads.RearKg,
		"dynamics":      dynamics,
		"warnings":      warnings,
	}})
}

func (h *TirePressureEngineeringCalculationHTTPHandler) HandleTirePressureMetadataSolve(c *gin.Context) {
	if contentType := c.GetHeader("Content-Type"); contentType != "" && !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		WriteTirePressureCalculationErrorResponse(c, http.StatusBadRequest, "INVALID_JSON", "Content-Type must be application/json", "content_type")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	var raw map[string]json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		WriteTirePressureCalculationErrorResponse(c, http.StatusBadRequest, "INVALID_JSON", "Request body must be valid JSON", "")
		return
	}
	if raw == nil {
		WriteTirePressureCalculationErrorResponse(c, http.StatusBadRequest, "INVALID_JSON", "Request body must be a JSON object", "")
		return
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		WriteTirePressureCalculationErrorResponse(c, http.StatusBadRequest, "INVALID_JSON", "Request body must be valid JSON", "")
		return
	}
	var req tirepressure.Request
	requestDecoder := json.NewDecoder(bytes.NewReader(encoded))
	requestDecoder.DisallowUnknownFields()
	if err := requestDecoder.Decode(&req); err != nil {
		WriteTirePressureCalculationErrorResponse(c, http.StatusBadRequest, "INVALID_JSON", "Request body must be valid JSON", "")
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		WriteTirePressureCalculationErrorResponse(c, http.StatusBadRequest, "INVALID_JSON", "Request body must contain exactly one JSON object", "")
		return
	}
	if err := tirepressure.ValidateTirePressureDynamicCalculationRequest(req); err != nil {
		var validationErr *tirepressure.ValidationError
		if errors.As(err, &validationErr) {
			status := http.StatusBadRequest
			if validationErr.Code == "OUT_OF_RANGE" {
				status = http.StatusUnprocessableEntity
			}
			WriteTirePressureCalculationErrorResponse(c, status, validationErr.Code, validationErr.Reason, validationErr.Field)
			return
		}
		WriteTirePressureCalculationErrorResponse(c, http.StatusBadRequest, "INVALID_FIELD", "Request fields are invalid", "")
		return
	}
	loads, err := tirepressure.ResolveTirePressureWheelLoads(req)
	if err != nil {
		if errors.Is(err, tirepressure.ErrLoadMismatch) {
			WriteTirePressureCalculationErrorResponse(c, http.StatusUnprocessableEntity, "LOAD_TOTAL_MISMATCH", "Front and rear loads must sum to rider plus bike mass", "front_load_kg")
			return
		}
		var validationErr *tirepressure.ValidationError
		if errors.As(err, &validationErr) {
			WriteTirePressureCalculationErrorResponse(c, http.StatusBadRequest, validationErr.Code, validationErr.Reason, validationErr.Field)
			return
		}
		WriteTirePressureCalculationErrorResponse(c, http.StatusBadRequest, "INVALID_FIELD", "Request fields are invalid", "")
		return
	}
	maxBar, limitStatus := tirepressure.CalculateEffectiveProvidedTirePressureLimitBar(req)
	if limitStatus == "LIMITS_NOT_PROVIDED" {
		WriteTirePressureCalculationErrorResponse(c, http.StatusUnprocessableEntity, "LIMIT_UNVERIFIED", "Provide pressure limits with source records before using this metadata response", "limit_sources")
		return
	}
	if math.IsNaN(maxBar) || math.IsInf(maxBar, 0) || maxBar <= 0 {
		WriteTirePressureCalculationErrorResponse(c, http.StatusBadRequest, "INVALID_FIELD", "Pressure limit must be a positive finite number", "limit_sources")
		return
	}
	dynamics, err := tirepressure.CalculateTirePressureGroundFrameCorneringDynamics(req, loads)
	if err != nil {
		var validationErr *tirepressure.ValidationError
		if errors.As(err, &validationErr) {
			WriteTirePressureCalculationErrorResponse(c, http.StatusBadRequest, validationErr.Code, validationErr.Reason, validationErr.Field)
			return
		}
		WriteTirePressureCalculationErrorResponse(c, http.StatusBadRequest, "INVALID_FIELD", "Dynamic inputs are invalid", "")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"model_version":           tirepressure.ModelVersion,
			"recommendation_status":   "MODEL_NOT_CALIBRATED",
			"load_source":             loads.Source,
			"front_load_kg":           loads.FrontKg,
			"rear_load_kg":            loads.RearKg,
			"effective_limit_bar":     maxBar,
			"effective_limit_psi":     maxBar * tirepressure.PsiPerBar,
			"limit_status":            limitStatus,
			"pressure_recommendation": nil,
			"dynamics":                dynamics,
			"warnings": []string{
				"The pressure model is not approved for production; no pressure recommendation is returned.",
				"The data required for this calculation currently lacks complete and accurate support; the result is for relative change demonstration only.",
				"Follow tire, rim, and wheel manufacturer instructions.",
			},
		},
	})
}

func WriteTirePressureCalculationErrorResponse(c *gin.Context, status int, code, message, field string) {
	c.JSON(status, TirePressureCalculationErrorResponse{Code: code, Error: message, Field: field})
}
