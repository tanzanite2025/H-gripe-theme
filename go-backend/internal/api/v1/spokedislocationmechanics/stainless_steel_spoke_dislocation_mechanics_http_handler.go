package spokedislocationmechanics

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	spokedislocationmechanicsdomain "commerce-platform/internal/domain/spokedislocationmechanics"
	"github.com/gin-gonic/gin"
)

const maximumStainlessSteelSpokeDislocationMechanicsRequestBodyBytes = 16 * 1024

// StainlessSteelSpokeDislocationMechanicsHTTPHandler exposes the read-only
// backend calculation and metadata contract for the spoke engineering guide.
type StainlessSteelSpokeDislocationMechanicsHTTPHandler struct{}

func NewStainlessSteelSpokeDislocationMechanicsHTTPHandler() *StainlessSteelSpokeDislocationMechanicsHTTPHandler {
	return &StainlessSteelSpokeDislocationMechanicsHTTPHandler{}
}

type stainlessSteelSpokeDislocationMechanicsAPIErrorResponse struct {
	Code    string `json:"code"`
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Field   string `json:"field,omitempty"`
}

// RegisterStainlessSteelSpokeDislocationMechanicsRoutes registers the public
// metadata and calculation endpoints under an engineering-only route group.
func (h *StainlessSteelSpokeDislocationMechanicsHTTPHandler) RegisterStainlessSteelSpokeDislocationMechanicsRoutes(group *gin.RouterGroup) {
	group.GET("/metadata", h.GetStainlessSteelSpokeDislocationMechanicsMetadata)
	group.POST("/calculate", h.CalculateStainlessSteelSpokeDislocationMechanics)
}

func (h *StainlessSteelSpokeDislocationMechanicsHTTPHandler) GetStainlessSteelSpokeDislocationMechanicsMetadata(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=86400")
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": spokedislocationmechanicsdomain.GetStainlessSteelSpokeDislocationMechanicsMetadata(),
	})
}

func (h *StainlessSteelSpokeDislocationMechanicsHTTPHandler) CalculateStainlessSteelSpokeDislocationMechanics(c *gin.Context) {
	if contentType := c.GetHeader("Content-Type"); contentType != "" && !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		writeStainlessSteelSpokeDislocationMechanicsAPIError(c, http.StatusBadRequest, "INVALID_JSON", "Content-Type must be application/json", "content_type")
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maximumStainlessSteelSpokeDislocationMechanicsRequestBodyBytes)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	var request spokedislocationmechanicsdomain.StainlessSteelSpokeDislocationMechanicsCalculationRequest
	if err := decoder.Decode(&request); err != nil {
		writeStainlessSteelSpokeDislocationMechanicsAPIError(c, http.StatusBadRequest, "INVALID_JSON", "Request body must be a single JSON object", "")
		return
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		writeStainlessSteelSpokeDislocationMechanicsAPIError(c, http.StatusBadRequest, "INVALID_JSON", "Request body must contain exactly one JSON object", "")
		return
	}

	result, err := spokedislocationmechanicsdomain.CalculateStainlessSteelSpokeDislocationMechanics(request)
	if err != nil {
		status, code, field := mapStainlessSteelSpokeDislocationMechanicsValidationError(err)
		writeStainlessSteelSpokeDislocationMechanicsAPIError(c, status, code, err.Error(), field)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

func mapStainlessSteelSpokeDislocationMechanicsValidationError(err error) (int, string, string) {
	switch {
	case errors.Is(err, spokedislocationmechanicsdomain.ErrStainlessSteelSpokeStressOutsideMaterialCurve):
		return http.StatusUnprocessableEntity, "CURVE_OUT_OF_RANGE", "nominal_working_tension_n"
	case errors.Is(err, spokedislocationmechanicsdomain.ErrUnknownStainlessSteelSpokeModel):
		return http.StatusUnprocessableEntity, "UNKNOWN_MODEL", "model_id"
	case errors.Is(err, spokedislocationmechanicsdomain.ErrCustomSpokeAreaRequired):
		return http.StatusUnprocessableEntity, "OUT_OF_RANGE", "custom_effective_area_mm2"
	case errors.Is(err, spokedislocationmechanicsdomain.ErrStainlessSteelSpokeDislocationMechanicsFieldOutOfRange):
		return http.StatusUnprocessableEntity, "OUT_OF_RANGE", ""
	case errors.Is(err, spokedislocationmechanicsdomain.ErrInvalidStainlessSteelSpokeDislocationMechanicsField):
		return http.StatusBadRequest, "INVALID_FIELD", ""
	default:
		return http.StatusBadRequest, "INVALID_FIELD", ""
	}
}

func writeStainlessSteelSpokeDislocationMechanicsAPIError(c *gin.Context, status int, code, message, field string) {
	c.JSON(status, stainlessSteelSpokeDislocationMechanicsAPIErrorResponse{
		Code:    code,
		Error:   message,
		Message: message,
		Field:   field,
	})
}
