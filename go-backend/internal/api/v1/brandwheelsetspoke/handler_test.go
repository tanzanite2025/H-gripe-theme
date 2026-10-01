package brandwheelsetspoke

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newTestRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/wheelset-spoke-specs")
	handler.RegisterRoutes(group)
	return router
}

func TestListModelsDoesNotExposePrivateRepairKitFields(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/wheelset-spoke-specs/models", nil)
	newTestRouter(NewHandler()).ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	body := recorder.Body.String()
	for _, forbidden := range []string{"lengthMm", "nippleModel", "nippleLengthMm", "sourceUrl", "verificationStatus"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("public model index contains private field %q", forbidden)
		}
	}

	var payload struct {
		Data struct {
			Models []modelIndex `json:"models"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	require.NotEmpty(t, payload.Data.Models)
	modelSlugs := make([]string, 0, len(payload.Data.Models))
	for _, model := range payload.Data.Models {
		modelSlugs = append(modelSlugs, model.Slug)
	}
	require.Contains(t, modelSlugs, "arc-1100-dicut-db-38")
	require.Contains(t, modelSlugs, "wh-r9270-c50-tl")
}

func TestGetModelRequiresAuthentication(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/wheelset-spoke-specs/models/arc-1100-dicut-db-38", nil)
	newTestRouter(NewHandler()).ServeHTTP(recorder, request)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Contains(t, recorder.Body.String(), "registration_required")
}

func TestGetModelReturnsOnlyRequestedPrivateRecord(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/wheelset-spoke-specs/models/arc-1100-dicut-db-38", nil)
	context, _ := gin.CreateTestContext(recorder)
	context.Request = request
	context.Params = gin.Params{{Key: "slug", Value: "arc-1100-dicut-db-38"}}
	context.Set("user_id", uint(42))

	NewHandler().GetModel(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	body := recorder.Body.String()
	require.Contains(t, body, `"slug":"arc-1100-dicut-db-38"`)
	require.Contains(t, body, `"lengthMm":285`)
	require.Contains(t, body, `"nippleModel":"DT Pro Lock Hidden Aluminum"`)
	require.NotContains(t, body, "arc-1100-dicut-db-55")
}

func TestGetShimanoModelReturnsRepairKitFields(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/wheelset-spoke-specs/models/wh-m8100-tl-29", nil)
	context, _ := gin.CreateTestContext(recorder)
	context.Request = request
	context.Params = gin.Params{{Key: "slug", Value: "wh-m8100-tl-29"}}
	context.Set("user_id", uint(42))

	NewHandler().GetModel(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	body := recorder.Body.String()
	require.Contains(t, body, `"model":"DEORE XT WH-M8100-TL-29 (Boost)"`)
	require.Contains(t, body, `"lengthMm":298`)
	require.Contains(t, body, `"spokeModel":"Shimano XT butted 2.0-1.5-2.0 mm"`)
	require.Contains(t, body, `"nippleModel":"Shimano 14G aluminum nipple with spherical washer"`)
	require.NotContains(t, body, "sourceUrl")
}
