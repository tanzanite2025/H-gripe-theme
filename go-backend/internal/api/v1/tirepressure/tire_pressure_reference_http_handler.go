package tirepressure

import (
	"net/http"
	"strings"

	"commerce-platform/internal/repository"

	"github.com/gin-gonic/gin"
)

const tirePressureReferenceDataSchemaVersion = "tire-pressure-reference-v2"

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
	ArticleNo       string  `json:"article_no"`
	ModelName       string  `json:"model_name"`
	ETRTO           string  `json:"etrto"`
	InchDesignation *string `json:"inch_designation,omitempty"`
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
	return tirePressureReferenceCatalogItemResponse{
		Product: tirePressureReferenceProductResponse{
			ArticleNo:       item.ArticleNo,
			ModelName:       item.ModelName,
			ETRTO:           item.ETRTO,
			InchDesignation: item.InchDesignation,
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

func writeTirePressureReferenceErrorResponse(c *gin.Context, status int, code, message, field string) {
	c.JSON(status, gin.H{
		"code":  code,
		"error": message,
		"field": field,
	})
}
