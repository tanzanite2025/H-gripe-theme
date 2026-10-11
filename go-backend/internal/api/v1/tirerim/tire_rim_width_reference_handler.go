package tirerim

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	domain "commerce-platform/internal/domain/tirerim"

	"github.com/gin-gonic/gin"
)

const maxTireRimWidthReferenceRequestBodyBytes = 16 * 1024

type TireRimWidthReferenceHandler struct{}

func NewTireRimWidthReferenceHandler() *TireRimWidthReferenceHandler {
	return &TireRimWidthReferenceHandler{}
}

type tireRimWidthReferenceSolveRequest struct {
	TireWidthMM        *int   `json:"tire_width_mm,omitempty"`
	RimSystem          string `json:"rim_system"`
	RimInnerWidthMM    *int   `json:"rim_inner_width_mm,omitempty"`
	ApplicationProfile string `json:"application_profile,omitempty"`
}

func (h *TireRimWidthReferenceHandler) RegisterTireRimWidthReferenceHTTPRoutes(group *gin.RouterGroup) {
	group.POST("/solve", h.SolveTireRimWidthReference)
}

func (h *TireRimWidthReferenceHandler) SolveTireRimWidthReference(c *gin.Context) {
	if contentType := c.GetHeader("Content-Type"); contentType != "" && !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		tireRimWidthReferenceWriteAPIErrorResponse(c, http.StatusBadRequest, "INVALID_JSON", "Content-Type must be application/json", "content_type")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxTireRimWidthReferenceRequestBodyBytes)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	var request tireRimWidthReferenceSolveRequest
	if err := decoder.Decode(&request); err != nil {
		tireRimWidthReferenceWriteAPIErrorResponse(c, http.StatusBadRequest, "INVALID_JSON", "Request body must be a single JSON object", "")
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		tireRimWidthReferenceWriteAPIErrorResponse(c, http.StatusBadRequest, "INVALID_JSON", "Request body must contain exactly one JSON object", "")
		return
	}
	normalizedRimSystem := domain.NormalizeRimSystem(request.RimSystem)
	normalizedApplicationProfile := domain.NormalizeTireRimApplicationProfile(request.ApplicationProfile)
	if request.TireWidthMM == nil && request.RimInnerWidthMM != nil {
		reverseSuggestion, err := domain.ResolveTireRimWidthReferenceForRim(
			normalizedRimSystem,
			*request.RimInnerWidthMM,
			normalizedApplicationProfile,
		)
		if err != nil {
			status := http.StatusBadRequest
			code := "INVALID_FIELD"
			field := ""
			switch {
			case errors.Is(err, domain.ErrInvalidRimSystem):
				field = "rim_system"
			case errors.Is(err, domain.ErrInvalidTireRimApplicationProfile), errors.Is(err, domain.ErrTireRimApplicationProfileOutsideAeroScope):
				field = "application_profile"
			case errors.Is(err, domain.ErrInvalidTireRimEngineeringInnerWidth):
				status = http.StatusUnprocessableEntity
				code = "OUT_OF_RANGE"
				field = "rim_inner_width_mm"
			}
			tireRimWidthReferenceWriteAPIErrorResponse(c, status, code, err.Error(), field)
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": reverseSuggestion})
		return
	}

	if request.TireWidthMM == nil {
		tireRimWidthReferenceWriteAPIErrorResponse(c, http.StatusBadRequest, "INVALID_FIELD", "tire_width_mm or rim_inner_width_mm is required", "tire_width_mm")
		return
	}

	if request.RimInnerWidthMM != nil {
		engineeringCalculation, err := domain.CalculateTireRimEngineeringReference(
			normalizedRimSystem,
			*request.RimInnerWidthMM,
			*request.TireWidthMM,
			normalizedApplicationProfile,
		)
		if err != nil {
			status := http.StatusBadRequest
			code := "INVALID_FIELD"
			field := ""
			switch {
			case errors.Is(err, domain.ErrInvalidTireWidth):
				status = http.StatusUnprocessableEntity
				code = "OUT_OF_RANGE"
				field = "tire_width_mm"
			case errors.Is(err, domain.ErrInvalidRimSystem):
				field = "rim_system"
			case errors.Is(err, domain.ErrInvalidTireRimApplicationProfile), errors.Is(err, domain.ErrTireRimApplicationProfileOutsideAeroScope):
				field = "application_profile"
			case errors.Is(err, domain.ErrInvalidTireRimEngineeringInnerWidth):
				status = http.StatusUnprocessableEntity
				code = "OUT_OF_RANGE"
				field = "rim_inner_width_mm"
			}
			tireRimWidthReferenceWriteAPIErrorResponse(c, status, code, err.Error(), field)
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": domain.BuildTireRimEngineeringSuggestion(engineeringCalculation)})
		return
	}

	result, err := domain.ResolveTireRimWidthReference(*request.TireWidthMM, normalizedRimSystem, normalizedApplicationProfile)
	if err != nil {
		status := http.StatusBadRequest
		code := "INVALID_FIELD"
		field := ""
		switch {
		case errors.Is(err, domain.ErrInvalidTireWidth):
			status = http.StatusUnprocessableEntity
			code = "OUT_OF_RANGE"
			field = "tire_width_mm"
		case errors.Is(err, domain.ErrInvalidRimSystem):
			field = "rim_system"
		case errors.Is(err, domain.ErrInvalidTireRimApplicationProfile), errors.Is(err, domain.ErrTireRimApplicationProfileOutsideAeroScope):
			field = "application_profile"
		case errors.Is(err, domain.ErrNoPublishedBracket):
			status = http.StatusUnprocessableEntity
			code = "NO_PUBLISHED_BRACKET"
			field = "tire_width_mm"
		}
		tireRimWidthReferenceWriteAPIErrorResponse(c, status, code, err.Error(), field)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

func tireRimWidthReferenceWriteAPIErrorResponse(c *gin.Context, status int, code, message, field string) {
	c.JSON(status, gin.H{"code": code, "error": message, "message": message, "field": field})
}
