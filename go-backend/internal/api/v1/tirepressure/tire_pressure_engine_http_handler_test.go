package tirepressure

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandleTirePressureMetadataSolveRejectsUnverifiedLimitsBeforeModelOutput(t *testing.T) {
	router := testRouter()
	body := `{"rider_weight_kg":72,"bike_weight_kg":8.5,"nominal_tire_width_mm":28,"inner_rim_width_mm":23,"rim_system":"HOOKLESS","riding_position":"AGGRESSIVE_RACE"}`
	recorder := request(router, body)
	if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), `"LIMIT_UNVERIFIED"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestHandleTirePressureMetadataSolveRejectsUnknownFields(t *testing.T) {
	recorder := request(testRouter(), `{"rider_weight_kg":72,"unexpected":true}`)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"INVALID_JSON"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestHandleTirePressureMetadataSolveRejectsNonObjectJSON(t *testing.T) {
	recorder := request(testRouter(), `[]`)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"INVALID_JSON"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestHandleTirePressureMetadataSolveRejectsUnsupportedContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/tire-pressure/solve", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "text/plain")
	recorder := httptest.NewRecorder()
	testRouter().ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"INVALID_JSON"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestHandleTirePressureMetadataSolveRejectsEmptyBody(t *testing.T) {
	recorder := request(testRouter(), "")
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"INVALID_JSON"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestHandleTirePressureMetadataSolveRejectsMultipleJSONValues(t *testing.T) {
	recorder := request(testRouter(), `{}`+`{}`)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"INVALID_JSON"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestHandleTirePressureMetadataSolveRejectsOversizedBody(t *testing.T) {
	largeBody := `{"rider_weight_kg":72,"bike_weight_kg":8.5,"nominal_tire_width_mm":28,"inner_rim_width_mm":23,"rim_system":"HOOKLESS","riding_position":"AGGRESSIVE_RACE","padding":"` + strings.Repeat("x", 17*1024) + `"}`
	recorder := request(testRouter(), largeBody)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"INVALID_JSON"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestHandleTirePressureMetadataSolveReturnsExplicitUncalibratedStatusWithVerifiedLimit(t *testing.T) {
	recorder := request(testRouter(), `{"rider_weight_kg":72,"bike_weight_kg":8.5,"nominal_tire_width_mm":28,"inner_rim_width_mm":23,"rim_system":"HOOKLESS","tire_max_pressure_bar":5,"limit_sources":[{"name":"rim maker","version":"manual-v1","applies_to":"selected setup","limit_type":"TIRE","limit_bar":5}],"riding_position":"AGGRESSIVE_RACE"}`)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"recommendation_status":"MODEL_NOT_CALIBRATED"`) || !strings.Contains(recorder.Body.String(), `"pressure_recommendation":null`) || !strings.Contains(recorder.Body.String(), `"limit_status":"PROVIDED_UNVERIFIED"`) || !strings.Contains(recorder.Body.String(), `"effective_limit_bar":5`) || !strings.Contains(recorder.Body.String(), `"effective_limit_psi"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestHandleTirePressureMetadataSolveRejectsLimitWithoutSourceRecord(t *testing.T) {
	recorder := request(testRouter(), `{"rider_weight_kg":72,"bike_weight_kg":8.5,"nominal_tire_width_mm":28,"inner_rim_width_mm":23,"rim_system":"HOOKLESS","tire_max_pressure_bar":5,"riding_position":"AGGRESSIVE_RACE"}`)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"limit_sources"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestHandleTirePressureMetadataSolveMapsRangeValidationTo422(t *testing.T) {
	recorder := request(testRouter(), `{"rider_weight_kg":151,"bike_weight_kg":8.5,"nominal_tire_width_mm":28,"inner_rim_width_mm":23,"rim_system":"HOOKLESS","riding_position":"AGGRESSIVE_RACE"}`)
	if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), `"OUT_OF_RANGE"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestHandleTirePressureMetadataSolveMapsMeasuredLoadMismatchTo422(t *testing.T) {
	recorder := request(testRouter(), `{"rider_weight_kg":72,"bike_weight_kg":8.5,"front_load_kg":20,"rear_load_kg":20,"nominal_tire_width_mm":28,"inner_rim_width_mm":23,"rim_system":"HOOKLESS","riding_position":"AGGRESSIVE_RACE"}`)
	if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), `"LOAD_TOTAL_MISMATCH"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestDynamicsReturnsGroundFrameDemoModel(t *testing.T) {
	recorder := requestPath(testRouter(), "/dynamics", "{\"rider_weight_kg\":72,\"bike_weight_kg\":8.5,\"nominal_tire_width_mm\":28,\"inner_rim_width_mm\":23,\"riding_position\":\"AGGRESSIVE_RACE\",\"surface_condition\":\"FLAT_ROAD\",\"lean_angle_deg\":30,\"speed_kmh\":36,\"front_operating_pressure_psi\":48,\"rear_operating_pressure_psi\":52,\"front_comparison_pressure_psi\":43,\"rear_comparison_pressure_psi\":47}")
	body := recorder.Body.String()
	if recorder.Code != http.StatusOK || !strings.Contains(body, "\"force_frame\":\"GROUND\"") || !strings.Contains(body, "\"model_status\":\"DEMO_ESTIMATE_UNCALIBRATED\"") || !strings.Contains(body, "\"lateral_demand_n\"") || !strings.Contains(body, "\"surface_condition\":\"FLAT_ROAD\"") || !strings.Contains(body, "\"speed_kmh\":36") || !strings.Contains(body, "\"equivalent_turn_radius_m\"") || !strings.Contains(body, "\"tire_body_normalization_factor\":1") || !strings.Contains(body, "\"tire_body_normalization_source\":\"FIXED_GENERIC_BASELINE\"") || !strings.Contains(body, "\"pressure_contact_area_comparison\"") || !strings.Contains(body, "\"area_change_pct\"") || !strings.Contains(body, "\"vertical_deformation\"") || !strings.Contains(body, "\"dataset_version\":\"maier-2018-vertical-deformation-reference-v1\"") {
		t.Fatalf("status=%d body=%s", recorder.Code, body)
	}
}

func TestDynamicsReturnsWetEquivalentPressureCompensation(t *testing.T) {
	recorder := requestPath(testRouter(), "/dynamics", `{"rider_weight_kg":72,"bike_weight_kg":8.5,"nominal_tire_width_mm":28,"riding_position":"AGGRESSIVE_RACE","surface_condition":"FLAT_ROAD","lean_angle_deg":30,"speed_kmh":30,"front_operating_pressure_psi":48,"rear_operating_pressure_psi":52,"wet_pressure_demonstration_enabled":true,"water_film_depth_mm":1}`)
	body := recorder.Body.String()
	if recorder.Code != http.StatusOK || !strings.Contains(body, `"wet_pressure_compensation"`) || !strings.Contains(body, `"friction_retention_ratio"`) || !strings.Contains(body, `"equivalent_pressure_psi"`) || !strings.Contains(body, `"wet_grip_margin_pct"`) || strings.Contains(body, `"area_compensated_grip_limit_n"`) || !strings.Contains(body, `"water_film_depth_mm":1`) || !strings.Contains(body, "first-order demonstration proxy") {
		t.Fatalf("status=%d body=%s", recorder.Code, body)
	}
}

func TestDynamicsAcceptsFlatRoadDemoInputsWithoutRimFields(t *testing.T) {
	recorder := requestPath(testRouter(), "/dynamics", "{\"rider_weight_kg\":72,\"bike_weight_kg\":8.5,\"nominal_tire_width_mm\":28,\"riding_position\":\"AGGRESSIVE_RACE\",\"surface_condition\":\"FLAT_ROAD\",\"lean_angle_deg\":30,\"front_operating_pressure_psi\":48,\"rear_operating_pressure_psi\":52}")
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"surface_condition":"FLAT_ROAD"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
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
	NewTirePressureEngineeringCalculationHandler().RegisterTirePressureEngineeringCalculationRoutes(router.Group("/api/v1/engineering/tire-pressure"))
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
