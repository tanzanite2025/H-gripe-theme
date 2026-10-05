package innertubefitment

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetInnerTubeValveFitmentMatrixIncludes100MillimetreRow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewInnerTubeValveFitmentHTTPHandler().RegisterInnerTubeValveFitmentHTTPRoutes(router.Group("/api/v1/engineering/inner-tube-fitment"))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/engineering/inner-tube-fitment/matrix", nil)
	router.ServeHTTP(recorder, request)

	body := recorder.Body.String()
	if recorder.Code != http.StatusOK || !strings.Contains(body, `"rim_depth_max_mm":100`) || !strings.Contains(body, `"rim_depth_mm":100`) || !strings.Contains(body, `"extender_length_mm":40`) {
		t.Fatalf("status=%d body=%s", recorder.Code, body)
	}
}

func TestSolveInnerTubeValveFitmentReturns100MillimetreRecommendation(t *testing.T) {
	recorder := requestInnerTubeValveFitment(t, `{"rim_depth_mm":100,"mode":"automatic"}`)
	body := recorder.Body.String()
	if recorder.Code != http.StatusOK || !strings.Contains(body, `"valve_length_mm":80`) || !strings.Contains(body, `"extender_length_mm":40`) || !strings.Contains(body, `"effective_exposure_mm":26.5`) || !strings.Contains(body, `"minimum_length_margin_mm":5`) || !strings.Contains(body, `"preferred_length_margin_mm":0`) || !strings.Contains(body, `"recommendation_reason_key":"shortestExtenderAssemblyReachesMinimum"`) || !strings.Contains(body, `"recommendation_reason":"shortest supported valve and extender assembly that reaches the minimum required total length"`) || !strings.Contains(body, `"alternatives":[`) || !strings.Contains(body, `"model_version":"inner-tube-valve-fitment-v1"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, body)
	}
}

func TestSolveInnerTubeValveFitmentAcceptsGripDepthAndUncertaintyParameters(t *testing.T) {
	recorder := requestInnerTubeValveFitment(t, `{"rim_depth_mm":50,"mode":"automatic","pump_head_grip_depth_mm":20,"rim_depth_uncertainty_mm":3}`)
	body := recorder.Body.String()
	if recorder.Code != http.StatusOK || !strings.Contains(body, `"pump_head_grip_depth_mm":20`) || !strings.Contains(body, `"rim_depth_uncertainty_mm":3`) || !strings.Contains(body, `"uncertainty_review":`) || !strings.Contains(body, `"lower_bound":`) {
		t.Fatalf("status=%d body=%s", recorder.Code, body)
	}
}

func TestSolveInnerTubeValveFitmentRejectsUnknownFields(t *testing.T) {
	recorder := requestInnerTubeValveFitment(t, `{"rim_depth_mm":50,"mode":"automatic","unexpected":true}`)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"INVALID_JSON"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSolveInnerTubeValveFitmentRejectsOutOfRangeRimDepth(t *testing.T) {
	recorder := requestInnerTubeValveFitment(t, `{"rim_depth_mm":101,"mode":"automatic"}`)
	if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), `"OUT_OF_RANGE"`) || !strings.Contains(recorder.Body.String(), `"rim_depth_mm"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSolveInnerTubeValveFitmentRejectsOutOfRangeReviewParameters(t *testing.T) {
	tests := []struct {
		name  string
		body  string
		field string
	}{
		{name: "pump head grip depth", body: `{"rim_depth_mm":50,"mode":"automatic","pump_head_grip_depth_mm":31}`, field: "pump_head_grip_depth_mm"},
		{name: "zero pump head grip depth", body: `{"rim_depth_mm":50,"mode":"automatic","pump_head_grip_depth_mm":0}`, field: "pump_head_grip_depth_mm"},
		{name: "rim depth uncertainty", body: `{"rim_depth_mm":50,"mode":"automatic","rim_depth_uncertainty_mm":6}`, field: "rim_depth_uncertainty_mm"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := requestInnerTubeValveFitment(t, test.body)
			if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), `"OUT_OF_RANGE"`) || !strings.Contains(recorder.Body.String(), `"`+test.field+`"`) {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestSolveInnerTubeValveFitmentRejectsUnsupportedContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewInnerTubeValveFitmentHTTPHandler().RegisterInnerTubeValveFitmentHTTPRoutes(router.Group("/api/v1/engineering/inner-tube-fitment"))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/inner-tube-fitment/solve", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "text/plain")
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"INVALID_JSON"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func requestInnerTubeValveFitment(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewInnerTubeValveFitmentHTTPHandler().RegisterInnerTubeValveFitmentHTTPRoutes(router.Group("/api/v1/engineering/inner-tube-fitment"))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/inner-tube-fitment/solve", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	return recorder
}
