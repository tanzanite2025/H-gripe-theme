package tirepressure

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"commerce-platform/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type testTirePressureReferenceCatalogReader struct {
	items      []repository.SchwalbeTireCatalogItem
	err        error
	lastSearch string
	callCount  int
}

func (reader *testTirePressureReferenceCatalogReader) ListSchwalbeTireCatalog(search string) ([]repository.SchwalbeTireCatalogItem, error) {
	reader.callCount++
	reader.lastSearch = search
	return reader.items, reader.err
}

func TestGetTirePressureReferenceDataReturnsExactArticleWithDimensionsAndPressureRange(t *testing.T) {
	reader := &testTirePressureReferenceCatalogReader{
		items: []repository.SchwalbeTireCatalogItem{
			{
				ArticleNo:       "11654432",
				ModelName:       "Marathon Efficiency",
				ETRTO:           "40-622",
				MinPressureBar:  newTirePressureReferenceFloat64Pointer(2.2),
				MaxPressureBar:  newTirePressureReferenceFloat64Pointer(4.5),
				MinPressurePSI:  newTirePressureReferenceFloat64Pointer(32),
				MaxPressurePSI:  newTirePressureReferenceFloat64Pointer(65),
				SourceCheckedAt: time.Date(2026, time.October, 3, 10, 20, 30, 0, time.UTC),
				ProductExists:   true,
			},
			{
				ArticleNo: "116544320",
				ModelName: "A different size",
				ETRTO:     "45-622",
			},
		},
	}

	recorder := requestTirePressureReferenceData(t, reader, " 11654432 ")
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.Equal(t, "public, max-age=300", recorder.Header().Get("Cache-Control"))

	var response struct {
		Code int                               `json:"code"`
		Data tirePressureReferenceDataResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, 0, response.Code)
	require.Equal(t, tirePressureReferenceDataSchemaVersion, response.Data.SchemaVersion)
	require.Equal(t, "11654432", response.Data.Product.ArticleNo)
	require.Equal(t, "Marathon Efficiency", response.Data.Product.ModelName)
	require.Equal(t, "40-622", response.Data.Product.ETRTO)
	require.NotNil(t, response.Data.Pressure.MinPressureBar)
	require.Equal(t, 2.2, *response.Data.Pressure.MinPressureBar)
	require.NotNil(t, response.Data.Pressure.MaxPressureBar)
	require.Equal(t, 4.5, *response.Data.Pressure.MaxPressureBar)
	require.NotNil(t, response.Data.Pressure.MinPressurePSI)
	require.Equal(t, float64(32), *response.Data.Pressure.MinPressurePSI)
	require.NotNil(t, response.Data.Pressure.MaxPressurePSI)
	require.Equal(t, float64(65), *response.Data.Pressure.MaxPressurePSI)
	require.Equal(t, "2026-10-03T10:20:30.000Z", response.Data.SourceCheckedAt)
	require.NotContains(t, recorder.Body.String(), `"weight_g"`)
	require.NotContains(t, recorder.Body.String(), `"compound"`)
	require.Equal(t, "11654432", reader.lastSearch)
}

func TestGetTirePressureReferenceDataRejectsMissingArticleNumber(t *testing.T) {
	reader := &testTirePressureReferenceCatalogReader{}
	recorder := requestTirePressureReferenceData(t, reader, "")

	require.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"INVALID_FIELD"`)
	require.Contains(t, recorder.Body.String(), `"article_no"`)
	require.Empty(t, reader.lastSearch)
}

func TestGetTirePressureReferenceDataDoesNotTreatArticleNumberSubstringAsExactMatch(t *testing.T) {
	reader := &testTirePressureReferenceCatalogReader{
		items: []repository.SchwalbeTireCatalogItem{{
			ArticleNo: "11654432",
			ETRTO:     "40-622",
		}},
	}
	recorder := requestTirePressureReferenceData(t, reader, "1165443")

	require.Equal(t, http.StatusNotFound, recorder.Code, recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"NOT_FOUND"`)
}

func TestGetTirePressureReferenceDataReturnsNotFoundWhenCatalogHasNoArticle(t *testing.T) {
	reader := &testTirePressureReferenceCatalogReader{
		items: []repository.SchwalbeTireCatalogItem{{
			ArticleNo: "11654432",
		}},
	}
	recorder := requestTirePressureReferenceData(t, reader, "11654433")

	require.Equal(t, http.StatusNotFound, recorder.Code, recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"NOT_FOUND"`)
}

func TestGetTirePressureReferenceCatalogDataReturnsLocalCatalogInOneRead(t *testing.T) {
	reader := &testTirePressureReferenceCatalogReader{
		items: []repository.SchwalbeTireCatalogItem{
			{
				ArticleNo:      "11654432",
				ModelName:      "Marathon Efficiency",
				ETRTO:          "40-622",
				MinPressureBar: newTirePressureReferenceFloat64Pointer(2.2),
				MaxPressureBar: newTirePressureReferenceFloat64Pointer(4.5),
			},
			{
				ArticleNo:      "11100770",
				ModelName:      "Marathon Plus",
				ETRTO:          "40-622",
				MinPressureBar: newTirePressureReferenceFloat64Pointer(3.5),
				MaxPressureBar: newTirePressureReferenceFloat64Pointer(6),
			},
		},
	}

	recorder := requestTirePressureReferenceCatalogData(t, reader)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.Equal(t, "public, max-age=300", recorder.Header().Get("Cache-Control"))

	var response struct {
		Code int                                  `json:"code"`
		Data tirePressureReferenceCatalogResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, 0, response.Code)
	require.Equal(t, tirePressureReferenceDataSchemaVersion, response.Data.SchemaVersion)
	require.Len(t, response.Data.Items, 2)
	require.Equal(t, "11654432", response.Data.Items[0].Product.ArticleNo)
	require.Equal(t, 2.2, *response.Data.Items[0].Pressure.MinPressureBar)
	require.Equal(t, 6.0, *response.Data.Items[1].Pressure.MaxPressureBar)
	require.Empty(t, reader.lastSearch)
	require.Equal(t, 1, reader.callCount)
}

func TestGetTirePressureReferenceDataMapsCatalogFailureToInternalServerError(t *testing.T) {
	reader := &testTirePressureReferenceCatalogReader{err: errors.New("catalog unavailable")}
	recorder := requestTirePressureReferenceData(t, reader, "11654432")

	require.Equal(t, http.StatusInternalServerError, recorder.Code, recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"REFERENCE_UNAVAILABLE"`)
}

func requestTirePressureReferenceData(t *testing.T, reader SchwalbeTirePressureReferenceCatalogReader, articleNo string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewTirePressureReferenceHTTPHandler(reader).RegisterTirePressureReferenceHTTPRoutes(router.Group("/api/v1/engineering/tire-pressure"))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/engineering/tire-pressure/reference-data", nil)
	query := request.URL.Query()
	if articleNo != "" {
		query.Set("article_no", articleNo)
		request.URL.RawQuery = query.Encode()
	}
	router.ServeHTTP(recorder, request)
	return recorder
}

func requestTirePressureReferenceCatalogData(t *testing.T, reader SchwalbeTirePressureReferenceCatalogReader) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewTirePressureReferenceHTTPHandler(reader).RegisterTirePressureReferenceHTTPRoutes(router.Group("/api/v1/engineering/tire-pressure"))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/engineering/tire-pressure/reference-data/catalog", nil)
	router.ServeHTTP(recorder, request)
	return recorder
}

func newTirePressureReferenceFloat64Pointer(value float64) *float64 {
	return &value
}
