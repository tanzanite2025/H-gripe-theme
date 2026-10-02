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
	for _, fragment := range []string{`"model_version":"stainless-steel-spoke-dislocation-mechanics-v2.1"`, `"model_catalog_id":"stainless-steel-spoke-model-reference-catalog"`, `"model_catalog_version":"1.2.0"`, `"material_catalog_id":"stainless-steel-spoke-material-reference-catalog"`, `"material_catalog_version":"1.0.2"`, `"id":"cx-ray"`, `"id":"dt-aerolite"`, `"effective_area_mm2":1.77`, `"elastic_modulus_mpa":193000`, `"effective_spoke_length_max_mm":1000`, `"macro_yield_reference_mpa":1400`, `"overload_ratio_max_percent":200`, `"data_status":"generic-materials-science-reference"`, `"data_status":"catalog-geometry-reference"`, `"url":"https://www.astm.org/a0313_a0313m.html"`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("expected metadata to contain %s, got %s", fragment, recorder.Body.String())
		}
	}
	if strings.Contains(recorder.Body.String(), `"mobile_dislocation_threshold_mpa"`) {
		t.Fatalf("metadata must not expose a dislocation-threshold parameter: %s", recorder.Body.String())
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
	for _, fragment := range []string{`"work_stress_mpa":769`, `"work_elastic_elongation_mm":1.156`, `"overload_force_n":1620`, `"overload_elastic_elongation_mm":1.56`, `"overload_delta_elastic_elongation_mm":0.405`, `"overload_yield_ratio_percent":74.2`, `"safety_level":"reference"`, `"model_reference":{"id":"cx-ray"`, `"material_reference_id":"austenitic-302-304-cold-drawn-spoke-reference"`, `"model_catalog_version":"1.2.0"`, `"material_catalog_version":"1.0.2"`, `"material_reference":{"id":"austenitic-302-304-cold-drawn-spoke-reference"`} {
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
	for _, fragment := range []string{`"effective_spoke_length_mm":200`, `"work_elastic_elongation_mm":0.797`, `"overload_elastic_elongation_mm":1.076`, `"overload_delta_elastic_elongation_mm":0.279`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("expected screenshot calculation result to contain %s, got %s", fragment, recorder.Body.String())
		}
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
