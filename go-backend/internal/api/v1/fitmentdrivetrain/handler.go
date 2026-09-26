package fitmentdrivetrain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	fitmentcatalogdomain "commerce-platform/internal/domain/fitmentcatalog"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	engine  *fitmentcatalogdomain.DrivetrainFitmentEngine
	version string
	etag    string
}

const knowledgeAsOf = "2026-09-25"

type calculateRequest struct {
	Brand           string                                `json:"brand"`
	CassetteSpec    string                                `json:"cassette_spec"`
	FreehubStandard *fitmentcatalogdomain.FreehubStandard `json:"freehub_standard,omitempty"`
}

type calculationResponse struct {
	Brand              string                                      `json:"brand"`
	CassetteSpec       string                                      `json:"cassette_spec"`
	DisplayName        string                                      `json:"display_name"`
	HintGroupsets      string                                      `json:"hint_groupsets"`
	Speed              int                                         `json:"speed"`
	MinCogTeeth        int                                         `json:"min_cog_teeth"`
	MaxCogTeeth        int                                         `json:"max_cog_teeth"`
	RecommendedFreehub fitmentcatalogdomain.FreehubStandard        `json:"recommended_freehub"`
	FitmentOptions     []fitmentcatalogdomain.FreehubFitmentOption `json:"fitment_options"`
	SelectedFreehub    *fitmentcatalogdomain.FreehubFitmentOption  `json:"selected_freehub,omitempty"`
	Spacer             fitmentcatalogdomain.SpacerRequirement      `json:"spacer"`
	ImageSrc           string                                      `json:"image_src"`
	MechanicalFact     string                                      `json:"mechanical_fact"`
	KnowledgeAsOf      string                                      `json:"knowledge_as_of"`
	RuleVersion        string                                      `json:"rule_version"`
}

// NewHandler creates a handler around the supplied immutable engine. Tests
// can inject a small rule set while production uses NewDefaultHandler.
func NewHandler(engine *fitmentcatalogdomain.DrivetrainFitmentEngine) *Handler {
	if engine == nil {
		engine = fitmentcatalogdomain.NewDefaultDrivetrainFitmentEngine()
	}
	matrix := engine.Matrix()
	version := "v1.0"
	if len(matrix) > 0 && matrix[0].RuleVersion != "" {
		version = matrix[0].RuleVersion
	}
	payload, _ := json.Marshal(matrix)
	digest := sha256.Sum256(payload)

	return &Handler{
		engine:  engine,
		version: version,
		etag:    `"` + hex.EncodeToString(digest[:]) + `"`,
	}
}

func NewDefaultHandler() *Handler {
	return NewHandler(nil)
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/calculate", h.Calculate)
	group.GET("/matrix", h.Matrix)
}

// Calculate returns the authoritative cassette/freehub fitment result.
// @Summary Calculate drivetrain and freehub fitment
// @Description Resolves a versioned cassette rule and optionally validates a caller-selected freehub standard. Unknown specifications fail loudly; no generic HG fallback is used.
// @Tags Fitment - Drivetrain
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Cassette and optional freehub selection"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 422 {object} map[string]interface{}
// @Router /api/v1/fitment/drivetrain/calculate [post]
func (h *Handler) Calculate(c *gin.Context) {
	var request calculateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondCalculationError(c, http.StatusBadRequest, "invalid drivetrain calculation request", err)
		return
	}
	if strings.TrimSpace(request.Brand) == "" || strings.TrimSpace(request.CassetteSpec) == "" {
		respondCalculationError(c, http.StatusBadRequest, "brand and cassette_spec are required", nil)
		return
	}

	rule, err := h.engine.CalculateByCassette(request.Brand, request.CassetteSpec)
	if err != nil {
		respondCalculationError(c, http.StatusBadRequest, "drivetrain calculation failed", err)
		return
	}

	response := newCalculationResponse(rule, nil)
	if request.FreehubStandard != nil {
		option, _, compatibilityErr := h.engine.CalculateByCassetteAndFreehub(
			request.Brand,
			request.CassetteSpec,
			*request.FreehubStandard,
		)
		if compatibilityErr != nil {
			status := http.StatusUnprocessableEntity
			message := "drivetrain fitment is incompatible"
			if errors.Is(compatibilityErr, fitmentcatalogdomain.ErrUnknownFreehub) {
				status = http.StatusBadRequest
				message = "invalid freehub standard"
			}
			respondCalculationError(c, status, message, compatibilityErr)
			return
		}
		response = newCalculationResponse(rule, option)
	} else {
		for index := range rule.FitmentOptions {
			if rule.FitmentOptions[index].Standard == rule.RecommendedFreehub {
				response = newCalculationResponse(rule, &rule.FitmentOptions[index])
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": response})
}

// Matrix returns the complete read-only drivetrain compatibility matrix.
// @Summary Get drivetrain compatibility matrix
// @Description Returns the versioned cassette/freehub/spacer matrix used by SSR and the storefront helper. The response is cacheable for 24 hours and supports ETag revalidation.
// @Tags Fitment - Drivetrain
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/fitment/drivetrain/matrix [get]
func (h *Handler) Matrix(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=86400")
	c.Header("ETag", h.etag)
	if strings.TrimSpace(c.GetHeader("If-None-Match")) == h.etag {
		c.Status(http.StatusNotModified)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"rules":           h.engine.Matrix(),
			"rule_version":    h.version,
			"knowledge_as_of": knowledgeAsOf,
		},
	})
}

func newCalculationResponse(rule *fitmentcatalogdomain.CassetteFitmentRule, selected *fitmentcatalogdomain.FreehubFitmentOption) calculationResponse {
	response := calculationResponse{
		Brand:              rule.Brand,
		CassetteSpec:       rule.CassetteSpec,
		DisplayName:        rule.DisplayName,
		HintGroupsets:      rule.HintGroupsets,
		Speed:              rule.Speed,
		MinCogTeeth:        rule.MinCogTeeth,
		MaxCogTeeth:        rule.MaxCogTeeth,
		RecommendedFreehub: rule.RecommendedFreehub,
		FitmentOptions:     rule.FitmentOptions,
		SelectedFreehub:    selected,
		ImageSrc:           rule.ImageSrc,
		MechanicalFact:     rule.MechanicalNotes,
		KnowledgeAsOf:      knowledgeAsOf,
		RuleVersion:        rule.RuleVersion,
	}
	if selected != nil {
		response.Spacer = selected.Spacer
		response.ImageSrc = selected.ImageSrc
	}
	return response
}

func respondCalculationError(c *gin.Context, status int, message string, err error) {
	details := ""
	if err != nil {
		details = err.Error()
	}
	if details != "" {
		message += ": " + details
	}
	code := "DRIVETRAIN_CALCULATION_ERROR"
	if errors.Is(err, fitmentcatalogdomain.ErrUnknownCassetteSpec) {
		code = "UNKNOWN_CASSETTE_SPEC"
	}
	if errors.Is(err, fitmentcatalogdomain.ErrUnknownFreehub) {
		code = "UNKNOWN_FREEHUB_STANDARD"
	}
	if errors.Is(err, fitmentcatalogdomain.ErrIncompatibleFreehub) {
		code = "INCOMPATIBLE_FREEHUB"
	}
	c.JSON(status, gin.H{"code": code, "message": message, "error": details})
}
