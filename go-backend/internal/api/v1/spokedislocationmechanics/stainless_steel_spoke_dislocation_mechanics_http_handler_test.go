package spokedislocationmechanics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newStainlessSteelSpokeDislocationMechanicsTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewStainlessSteelSpokeDislocationMechanicsHTTPHandler().RegisterStainlessSteelSpokeDislocationMechanicsRoutes(router.Group("/api/v1/engineering/spoke-dislocation-mechanics"))
	return router
}

func TestGetStainlessSteelSpokeDislocationMechanicsMetadataReturnsCacheableBackendContract(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/engineering/spoke-dislocation-mechanics/metadata", nil)
	newStainlessSteelSpokeDislocationMechanicsTestRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Cache-Control") != "public, max-age=86400" {
		t.Fatalf("expected metadata cache policy, got %q", recorder.Header().Get("Cache-Control"))
	}
	for _, fragment := range []string{`"model_version":"stainless-steel-spoke-dislocation-mechanics-v3.0"`, `"model_catalog_id":"stainless-steel-spoke-model-reference-catalog"`, `"model_catalog_version":"1.3.0"`, `"material_catalog_id":"stainless-steel-spoke-material-reference-catalog"`, `"material_catalog_version":"2.1.0"`, `"id":"aisi-304-cold-drawn-wire-true-strain-0-585"`, `"id":"cx-ray"`, `"id":"dt-aerolite"`, `"effective_area_mm2":1.77`, `"elastic_modulus_mpa":193000`, `"yield_strength_mpa":1137`, `"ultimate_tensile_strength_mpa":1346`, `"total_elongation_to_failure_percent":5`, `"effective_spoke_length_max_mm":1000`, `"overload_ratio_max_percent":200`, `"data_status":"published-curve-reference-digitized-from-figure"`, `"data_status":"catalog-geometry-reference"`, `"url":"https://open.uct.ac.za/server/api/core/bitstreams/869005f7-b771-4f0f-96c4-53eeecc8a628/content"`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("expected metadata to contain %s, got %s", fragment, recorder.Body.String())
		}
	}
	for _, fragment := range []string{`"reference_calculation":{"effective_spoke_length_mm":270`, `"reference_calculation_tension_kgf":180`, `"reference_calculation_tension_n":1765.197`, `"yield_strength_mpa":1137`, `"curve_endpoint_stress_mpa":1346`, `"curve_endpoint_total_strain_percent":3`, `"curve_endpoint_total_elongation_mm":8.1`, `"total_elongation_to_failure_percent":5`, `"fracture_reference_total_elongation_mm":13.5`, `"reference_calculation_stress_mpa":1131.537`, `"reference_calculation_total_elongation_mm":1.583`, `"yield_reference_force_n":1774`, `"curve_endpoint_force_n":2100`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("expected metadata to contain fixed reference calculation %s, got %s", fragment, recorder.Body.String())
		}
	}
	if strings.Contains(recorder.Body.String(), `"macro_yield_reference_mpa"`) || strings.Contains(recorder.Body.String(), `"mobile_dislocation_threshold_mpa"`) {
		t.Fatalf("metadata must expose the selected material curve contract instead of the retired generic thresholds: %s", recorder.Body.String())
	}
}

func TestCalculateStainlessSteelSpokeDislocationMechanicsAcceptsJSONCharsetAndReturnsRoundedResult(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/spoke-dislocation-mechanics/calculate", strings.NewReader(`{"model_id":"cx-ray","effective_spoke_length_mm":290,"nominal_working_tension_n":1200,"overload_ratio_percent":135}`))
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	newStainlessSteelSpokeDislocationMechanicsTestRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range []string{`"work_stress_mpa":769`, `"work_total_elongation_mm":1.156`, `"work_elastic_elongation_mm":1.156`, `"work_permanent_elongation_mm":0`, `"overload_force_n":1620`, `"overload_total_elongation_mm":1.56`, `"overload_elastic_elongation_mm":1.56`, `"overload_permanent_elongation_mm":0`, `"overload_delta_elastic_elongation_mm":0.405`, `"overload_delta_total_elongation_mm":0.405`, `"overload_delta_permanent_elongation_mm":0`, `"fracture_reference_total_elongation_mm":14.5`, `"overload_yield_ratio_percent":91.3`, `"safety_level":"warning"`, `"model_reference":{"id":"cx-ray"`, `"material_reference_id":"aisi-304-cold-drawn-wire-true-strain-0-585"`, `"model_catalog_version":"1.3.0"`, `"material_catalog_version":"2.1.0"`, `"material_reference":{"id":"aisi-304-cold-drawn-wire-true-strain-0-585"`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("expected result to contain %s, got %s", fragment, recorder.Body.String())
		}
	}
}

func TestCalculateStainlessSteelSpokeDislocationMechanicsAcceptsScreenshotInputValues(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/spoke-dislocation-mechanics/calculate", strings.NewReader(`{"model_id":"cx-ray","effective_spoke_length_mm":200,"nominal_working_tension_n":1200,"overload_ratio_percent":135}`))
	request.Header.Set("Content-Type", "application/json")
	newStainlessSteelSpokeDislocationMechanicsTestRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected screenshot input values to return 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range []string{`"effective_spoke_length_mm":200`, `"work_total_elongation_mm":0.797`, `"work_elastic_elongation_mm":0.797`, `"work_permanent_elongation_mm":0`, `"overload_total_elongation_mm":1.076`, `"overload_elastic_elongation_mm":1.076`, `"overload_delta_total_elongation_mm":0.279`, `"fracture_reference_total_elongation_mm":10`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("expected screenshot calculation result to contain %s, got %s", fragment, recorder.Body.String())
		}
	}
}

func TestCalculateStainlessSteelSpokeDislocationMechanicsRejectsStressOutsideMaterialCurve(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/spoke-dislocation-mechanics/calculate", strings.NewReader(`{"model_id":"custom","custom_effective_area_mm2":1,"effective_spoke_length_mm":290,"nominal_working_tension_n":1200,"overload_ratio_percent":120}`))
	request.Header.Set("Content-Type", "application/json")
	newStainlessSteelSpokeDislocationMechanicsTestRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), `"code":"CURVE_OUT_OF_RANGE"`) {
		t.Fatalf("expected stress outside the selected AISI 304 curve to be rejected, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestCalculateStainlessSteelSpokeDislocationMechanicsReturnsCriticalStateWithoutFalseSuccess(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/spoke-dislocation-mechanics/calculate", strings.NewReader(`{"model_id":"leader","effective_spoke_length_mm":290,"nominal_working_tension_n":1400,"overload_ratio_percent":150}`))
	request.Header.Set("Content-Type", "application/json")
	newStainlessSteelSpokeDislocationMechanicsTestRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"safety_level":"critical"`) {
		t.Fatalf("expected critical result state, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestCalculateStainlessSteelSpokeDislocationMechanicsRejectsUnknownFieldsAndOutOfRangeInputs(t *testing.T) {
	cases := []struct {
		name           string
		body           string
		expectedCode   string
		expectedStatus int
	}{
		{name: "unknown field", body: `{"model_id":"cx-ray","effective_spoke_length_mm":290,"nominal_working_tension_n":1200,"overload_ratio_percent":135,"work_stress":700}`, expectedCode: "INVALID_JSON", expectedStatus: http.StatusBadRequest},
		{name: "out of range", body: `{"model_id":"cx-ray","effective_spoke_length_mm":290,"nominal_working_tension_n":1200,"overload_ratio_percent":99}`, expectedCode: "OUT_OF_RANGE", expectedStatus: http.StatusUnprocessableEntity},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/spoke-dislocation-mechanics/calculate", strings.NewReader(testCase.body))
			request.Header.Set("Content-Type", "application/json")
			newStainlessSteelSpokeDislocationMechanicsTestRouter().ServeHTTP(recorder, request)
			if recorder.Code != testCase.expectedStatus || !strings.Contains(recorder.Body.String(), `"code":"`+testCase.expectedCode+`"`) {
				t.Fatalf("expected %s/%d, got %d: %s", testCase.expectedCode, testCase.expectedStatus, recorder.Code, recorder.Body.String())
			}
		})
	}
}
