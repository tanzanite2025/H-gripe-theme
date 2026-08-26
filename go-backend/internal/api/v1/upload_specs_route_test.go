package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"commerce-platform/internal/app"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/upload"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestUploadSpecsRouteReturnsStablePublicContract(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterRoutes(router, &app.Dependencies{}, &config.Config{
		CORS: config.CORSConfig{},
		JWT:  config.JWTConfig{Secret: "test-secret"},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/upload-specs", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

	var payload struct {
		Specs []upload.UploadSpec `json:"specs"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	require.Len(t, payload.Specs, len(upload.ListUploadSpecs()))

	codes := make([]string, 0, len(payload.Specs))
	for _, spec := range payload.Specs {
		codes = append(codes, spec.Code)
	}
	require.True(t, sort.StringsAreSorted(codes), "spec codes are not sorted: %v", codes)

	var homeVisual *upload.UploadSpec
	for index := range payload.Specs {
		if payload.Specs[index].Code == string(upload.SpecVisualShowcaseHomeCategories) {
			homeVisual = &payload.Specs[index]
			break
		}
	}
	require.NotNil(t, homeVisual)
	require.Equal(t, "16:9", homeVisual.AspectRatioLabel)
	require.Equal(t, 1920, homeVisual.RecommendedWidth)
	require.Equal(t, 1080, homeVisual.RecommendedHeight)
}
