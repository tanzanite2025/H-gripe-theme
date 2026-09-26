package tirepressure

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSolveRejectsUnverifiedLimitsBeforeModelOutput(t *testing.T) {
	router := testRouter()
	body := `{"rider_weight_kg":72,"bike_weight_kg":8.5,"nominal_tire_width_mm":28,"inner_rim_width_mm":23,"rim_system":"HOOKLESS","riding_position":"AGGRESSIVE_RACE"}`
	recorder := request(router, body)
	if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), `"LIMIT_UNVERIFIED"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSolveRejectsUnknownFields(t *testing.T) {
	recorder := request(testRouter(), `{"rider_weight_kg":72,"unexpected":true}`)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"INVALID_JSON"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSolveRejectsNonObjectJSON(t *testing.T) {
	recorder := request(testRouter(), `[]`)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"INVALID_JSON"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSolveRejectsUnsupportedContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/tire-pressure/solve", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "text/plain")
	recorder := httptest.NewRecorder()
	testRouter().ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"INVALID_JSON"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSolveRejectsEmptyBody(t *testing.T) {
	recorder := request(testRouter(), "")
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"INVALID_JSON"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSolveRejectsMultipleJSONValues(t *testing.T) {
	recorder := request(testRouter(), `{}`+`{}`)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"INVALID_JSON"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSolveRejectsOversizedBody(t *testing.T) {
	largeBody := `{"rider_weight_kg":72,"bike_weight_kg":8.5,"nominal_tire_width_mm":28,"inner_rim_width_mm":23,"rim_system":"HOOKLESS","riding_position":"AGGRESSIVE_RACE","padding":"` + strings.Repeat("x", 17*1024) + `"}`
	recorder := request(testRouter(), largeBody)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"INVALID_JSON"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSolveReturnsExplicitUncalibratedStatusWithVerifiedLimit(t *testing.T) {
	recorder := request(testRouter(), `{"rider_weight_kg":72,"bike_weight_kg":8.5,"nominal_tire_width_mm":28,"inner_rim_width_mm":23,"rim_system":"HOOKLESS","tire_max_pressure_bar":5,"limit_sources":[{"name":"rim maker","version":"manual-v1","applies_to":"selected setup","limit_type":"TIRE","limit_bar":5}],"riding_position":"AGGRESSIVE_RACE"}`)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"recommendation_status":"MODEL_NOT_CALIBRATED"`) || !strings.Contains(recorder.Body.String(), `"pressure_recommendation":null`) || !strings.Contains(recorder.Body.String(), `"limit_status":"PROVIDED_UNVERIFIED"`) || !strings.Contains(recorder.Body.String(), `"effective_limit_bar":5`) || !strings.Contains(recorder.Body.String(), `"effective_limit_psi"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSolveRejectsLimitWithoutSourceRecord(t *testing.T) {
	recorder := request(testRouter(), `{"rider_weight_kg":72,"bike_weight_kg":8.5,"nominal_tire_width_mm":28,"inner_rim_width_mm":23,"rim_system":"HOOKLESS","tire_max_pressure_bar":5,"riding_position":"AGGRESSIVE_RACE"}`)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"limit_sources"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSolveMapsRangeValidationTo422(t *testing.T) {
	recorder := request(testRouter(), `{"rider_weight_kg":151,"bike_weight_kg":8.5,"nominal_tire_width_mm":28,"inner_rim_width_mm":23,"rim_system":"HOOKLESS","riding_position":"AGGRESSIVE_RACE"}`)
	if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), `"OUT_OF_RANGE"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSolveMapsMeasuredLoadMismatchTo422(t *testing.T) {
	recorder := request(testRouter(), `{"rider_weight_kg":72,"bike_weight_kg":8.5,"front_load_kg":20,"rear_load_kg":20,"nominal_tire_width_mm":28,"inner_rim_width_mm":23,"rim_system":"HOOKLESS","riding_position":"AGGRESSIVE_RACE"}`)
	if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), `"LOAD_TOTAL_MISMATCH"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestDynamicsReturnsGroundFrameDemoModel(t *testing.T) {
	recorder := requestPath(testRouter(), "/dynamics", "{\"rider_weight_kg\":72,\"bike_weight_kg\":8.5,\"nominal_tire_width_mm\":28,\"inner_rim_width_mm\":23,\"rim_system\":\"HOOKLESS\",\"riding_position\":\"AGGRESSIVE_RACE\",\"surface_condition\":\"ROUGH_CHIP\",\"casing_type\":\"TUBELESS\",\"weather_condition\":\"DRY\",\"lean_angle_deg\":30,\"front_operating_pressure_psi\":48,\"rear_operating_pressure_psi\":52}")
	body := recorder.Body.String()
	if recorder.Code != http.StatusOK || !strings.Contains(body, "\"force_frame\":\"GROUND\"") || !strings.Contains(body, "\"model_status\":\"DEMO_ESTIMATE_UNCALIBRATED\"") || !strings.Contains(body, "\"lateral_demand_n\"") {
		t.Fatalf("status=%d body=%s", recorder.Code, body)
	}
}

func TestDynamicsRejectsMultipleJSONValues(t *testing.T) {
	recorder := requestPath(testRouter(), "/dynamics", "{}{}")
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "\"INVALID_JSON\"") {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func testRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewHandler().RegisterRoutes(router.Group("/api/v1/engineering/tire-pressure"))
	return router
}

func request(router *gin.Engine, body string) *httptest.ResponseRecorder {
	return requestPath(router, "/solve", body)
}

func requestPath(router *gin.Engine, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/tire-pressure"+path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}
