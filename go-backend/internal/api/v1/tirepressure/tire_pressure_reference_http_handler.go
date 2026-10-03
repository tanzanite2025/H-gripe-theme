package tirepressure

import (
	"net/http"
	"strconv"
	"strings"

	"commerce-platform/internal/repository"

	"github.com/gin-gonic/gin"
)

const tirePressureReferenceDataSchemaVersion = "tire-pressure-reference-v1"

// SchwalbeTirePressureReferenceCatalogReader is the narrow read-only catalog
// contract used by the tire-pressure reference adapter. Keeping this contract
// separate from the dynamics handler prevents the calculation endpoint from
// depending on the product catalog implementation.
type SchwalbeTirePressureReferenceCatalogReader interface {
	ListSchwalbeTireCatalog(search string) ([]repository.SchwalbeTireCatalogItem, error)
}

// TirePressureReferenceHTTPHandler exposes the compact product snapshot that
// the tire-pressure page needs before it calls the independent dynamics API.
type TirePressureReferenceHTTPHandler struct {
	catalogReader SchwalbeTirePressureReferenceCatalogReader
}

func NewTirePressureReferenceHTTPHandler(
	catalogReader SchwalbeTirePressureReferenceCatalogReader,
) *TirePressureReferenceHTTPHandler {
	return &TirePressureReferenceHTTPHandler{catalogReader: catalogReader}
}

func (h *TirePressureReferenceHTTPHandler) RegisterTirePressureReferenceHTTPRoutes(group *gin.RouterGroup) {
	group.GET("/reference-data", h.GetTirePressureReferenceData)
	group.GET("/reference-data/catalog", h.GetTirePressureReferenceCatalogData)
}

type tirePressureReferenceProductResponse struct {
	ArticleNo          string   `json:"article_no"`
	EAN                *string  `json:"ean,omitempty"`
	ModelName          string   `json:"model_name"`
	ETRTO              string   `json:"etrto"`
	InchDesignation    *string  `json:"inch_designation,omitempty"`
	NominalTireWidthMM *int     `json:"nominal_tire_width_mm,omitempty"`
	BeadSeatDiameterMM *int     `json:"bead_seat_diameter_mm,omitempty"`
	WeightG            *float64 `json:"weight_g,omitempty"`
	VersionLabel       *string  `json:"version_label,omitempty"`
	Compound           *string  `json:"compound,omitempty"`
	Color              *string  `json:"color,omitempty"`
	Bead               *string  `json:"bead,omitempty"`
	EBikeRating        *string  `json:"e_bike_rating,omitempty"`
	EPI                *int     `json:"epi,omitempty"`
	LoadKG             *float64 `json:"load_kg,omitempty"`
	Seal               *string  `json:"seal,omitempty"`
	Tread              *string  `json:"tread,omitempty"`
	ProductExists      bool     `json:"product_exists"`
}

type tirePressureReferencePressureResponse struct {
	MinPressureBar *float64 `json:"min_pressure_bar,omitempty"`
	MaxPressureBar *float64 `json:"max_pressure_bar,omitempty"`
	MinPressurePSI *float64 `json:"min_pressure_psi,omitempty"`
	MaxPressurePSI *float64 `json:"max_pressure_psi,omitempty"`
}

type tirePressureReferenceDataResponse struct {
	SchemaVersion   string                                `json:"schema_version"`
	Product         tirePressureReferenceProductResponse  `json:"product"`
	Pressure        tirePressureReferencePressureResponse `json:"pressure"`
	SourceCheckedAt string                                `json:"source_checked_at"`
}

type tirePressureReferenceCatalogItemResponse struct {
	Product         tirePressureReferenceProductResponse  `json:"product"`
	Pressure        tirePressureReferencePressureResponse `json:"pressure"`
	SourceCheckedAt string                                `json:"source_checked_at"`
}

type tirePressureReferenceCatalogResponse struct {
	SchemaVersion string                                     `json:"schema_version"`
	Items         []tirePressureReferenceCatalogItemResponse `json:"items"`
}

// GetTirePressureReferenceData returns one exact local catalog row by article number.
func (h *TirePressureReferenceHTTPHandler) GetTirePressureReferenceData(c *gin.Context) {
	articleNo := strings.TrimSpace(c.Query("article_no"))
	if articleNo == "" {
		writeTirePressureReferenceErrorResponse(c, http.StatusBadRequest, "INVALID_FIELD", "article_no is required", "article_no")
		return
	}
	if len(articleNo) > 128 {
		writeTirePressureReferenceErrorResponse(c, http.StatusBadRequest, "INVALID_FIELD", "article_no is too long", "article_no")
		return
	}
	if h == nil || h.catalogReader == nil {
		writeTirePressureReferenceErrorResponse(c, http.StatusInternalServerError, "REFERENCE_UNAVAILABLE", "Schwalbe tire reference data is unavailable", "")
		return
	}

	items, err := h.catalogReader.ListSchwalbeTireCatalog(articleNo)
	if err != nil {
		writeTirePressureReferenceErrorResponse(c, http.StatusInternalServerError, "REFERENCE_UNAVAILABLE", "Schwalbe tire reference data is unavailable", "")
		return
	}

	item, found := findExactSchwalbeTireCatalogItemByArticleNumber(items, articleNo)
	if !found {
		writeTirePressureReferenceErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Schwalbe tire article was not found", "article_no")
		return
	}

	c.Header("Cache-Control", "public, max-age=300")
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": buildTirePressureReferenceDataResponse(item),
	})
}

// GetTirePressureReferenceCatalogData returns the persisted local catalog in one repository read.
func (h *TirePressureReferenceHTTPHandler) GetTirePressureReferenceCatalogData(c *gin.Context) {
	if h == nil || h.catalogReader == nil {
		writeTirePressureReferenceErrorResponse(c, http.StatusInternalServerError, "REFERENCE_UNAVAILABLE", "Schwalbe tire reference data is unavailable", "")
		return
	}

	items, err := h.catalogReader.ListSchwalbeTireCatalog("")
	if err != nil {
		writeTirePressureReferenceErrorResponse(c, http.StatusInternalServerError, "REFERENCE_UNAVAILABLE", "Schwalbe tire reference data is unavailable", "")
		return
	}

	responseItems := make([]tirePressureReferenceCatalogItemResponse, 0, len(items))
	for _, item := range items {
		responseItems = append(responseItems, buildTirePressureReferenceCatalogItemResponse(item))
	}

	c.Header("Cache-Control", "public, max-age=300")
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": tirePressureReferenceCatalogResponse{
			SchemaVersion: tirePressureReferenceDataSchemaVersion,
			Items:         responseItems,
		},
	})
}

func findExactSchwalbeTireCatalogItemByArticleNumber(
	items []repository.SchwalbeTireCatalogItem,
	articleNo string,
) (repository.SchwalbeTireCatalogItem, bool) {
	normalizedArticleNo := strings.ToLower(strings.TrimSpace(articleNo))
	for _, item := range items {
		if strings.ToLower(strings.TrimSpace(item.ArticleNo)) == normalizedArticleNo {
			return item, true
		}
	}
	return repository.SchwalbeTireCatalogItem{}, false
}

func buildTirePressureReferenceDataResponse(
	item repository.SchwalbeTireCatalogItem,
) tirePressureReferenceDataResponse {
	itemResponse := buildTirePressureReferenceCatalogItemResponse(item)
	return tirePressureReferenceDataResponse{
		SchemaVersion:   tirePressureReferenceDataSchemaVersion,
		Product:         itemResponse.Product,
		Pressure:        itemResponse.Pressure,
		SourceCheckedAt: itemResponse.SourceCheckedAt,
	}
}

func buildTirePressureReferenceCatalogItemResponse(
	item repository.SchwalbeTireCatalogItem,
) tirePressureReferenceCatalogItemResponse {
	nominalTireWidthMM, beadSeatDiameterMM := parseSchwalbeETRTODimensions(item.ETRTO)

	return tirePressureReferenceCatalogItemResponse{
		Product: tirePressureReferenceProductResponse{
			ArticleNo:          item.ArticleNo,
			EAN:                item.EAN,
			ModelName:          item.ModelName,
			ETRTO:              item.ETRTO,
			InchDesignation:    item.InchDesignation,
			NominalTireWidthMM: nominalTireWidthMM,
			BeadSeatDiameterMM: beadSeatDiameterMM,
			WeightG:            item.WeightG,
			VersionLabel:       item.VersionLabel,
			Compound:           item.Compound,
			Color:              item.Color,
			Bead:               item.Bead,
			EBikeRating:        item.EBikeRating,
			EPI:                item.EPI,
			LoadKG:             item.LoadKG,
			Seal:               item.Seal,
			Tread:              item.Tread,
			ProductExists:      item.ProductExists,
		},
		Pressure: tirePressureReferencePressureResponse{
			MinPressureBar: item.MinPressureBar,
			MaxPressureBar: item.MaxPressureBar,
			MinPressurePSI: item.MinPressurePSI,
			MaxPressurePSI: item.MaxPressurePSI,
		},
		SourceCheckedAt: item.SourceCheckedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
	}
}

func parseSchwalbeETRTODimensions(etrto string) (*int, *int) {
	normalizedETRTO := strings.NewReplacer("–", "-", "—", "-").Replace(strings.TrimSpace(etrto))
	parts := strings.Split(normalizedETRTO, "-")
	if len(parts) != 2 {
		return nil, nil
	}

	width, widthErr := strconv.Atoi(strings.TrimSpace(parts[0]))
	beadSeatDiameter, beadSeatDiameterErr := strconv.Atoi(strings.TrimSpace(parts[1]))
	if widthErr != nil || beadSeatDiameterErr != nil || width <= 0 || beadSeatDiameter <= 0 {
		return nil, nil
	}

	return &width, &beadSeatDiameter
}

func writeTirePressureReferenceErrorResponse(c *gin.Context, status int, code, message, field string) {
	c.JSON(status, gin.H{
		"code":  code,
		"error": message,
		"field": field,
	})
}
