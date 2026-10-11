package tirerim

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	domain "commerce-platform/internal/domain/tirerim"

	"github.com/gin-gonic/gin"
)

func newTireRimWidthReferenceTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewTireRimWidthReferenceHandler().RegisterTireRimWidthReferenceHTTPRoutes(router.Group("/api/v1/engineering/tire-rim"))
	return router
}

func TestSolveTireRimWidthReferenceReturnsSingleBestPairingThroughExistingEndpoint(t *testing.T) {
	tests := []struct {
		name                    string
		requestBody             string
		expectedInnerWidthMM    int
		expectedTireWidthMM     int
		expectedProfile         domain.TireRimApplicationProfile
		expectedSelectionMethod string
		expectedAeroRule        bool
	}{
		{"inclusive aero boundary", `{"tire_width_mm":30,"rim_system":"hookless"}`, 23, 30, domain.TireRimApplicationProfileRoadAero, "project_aero_geometric_target", true},
		{"first width without aero", `{"tire_width_mm":31,"rim_system":"hookless"}`, 25, 31, domain.TireRimApplicationProfileGravelAllRoad, "recommended_window_center", false},
		{"known MTB tire", `{"tire_width_mm":60,"rim_system":"hookless"}`, 28, 60, domain.TireRimApplicationProfileMTBXC, "recommended_window_center", false},
		{"calibrated hooked MTB tire", `{"tire_width_mm":60,"rim_system":"hooked"}`, 31, 60, domain.TireRimApplicationProfileMTBXC, "recommended_window_center", false},
		{"calibrated hooked MTB rim", `{"rim_inner_width_mm":35,"rim_system":"hooked","application_profile":"mtb_xc_trail"}`, 35, 68, domain.TireRimApplicationProfileMTBXC, "recommended_window_center", false},
		{"known gravel rim", `{"rim_inner_width_mm":28,"rim_system":"hookless"}`, 28, 43, domain.TireRimApplicationProfileGravelAllRoad, "recommended_window_center", false},
		{"reviewed snow rim suggestion", `{"rim_inner_width_mm":85,"rim_system":"hookless"}`, 85, 105, domain.TireRimApplicationProfileFatSnow, "recommended_window_center", false},
		{"explicit MTB rim", `{"rim_inner_width_mm":35,"rim_system":"hookless","application_profile":"mtb_xc_trail"}`, 35, 67, domain.TireRimApplicationProfileMTBXC, "recommended_window_center", false},
		{"explicit Plus rim", `{"rim_inner_width_mm":35,"rim_system":"hookless","application_profile":" MTB_PLUS "}`, 35, 71, domain.TireRimApplicationProfileMTBPlus, "recommended_window_center", false},
		{"explicit Plus known tire", `{"tire_width_mm":71,"rim_system":"hookless","application_profile":"mtb_plus"}`, 35, 71, domain.TireRimApplicationProfileMTBPlus, "recommended_window_center", false},
		{"explicit Plus exact pair", `{"tire_width_mm":74,"rim_inner_width_mm":35,"rim_system":"hookless","application_profile":"mtb_plus"}`, 35, 71, domain.TireRimApplicationProfileMTBPlus, "recommended_window_center", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(test.requestBody))
			request.Header.Set("Content-Type", "application/json")
			newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
			}
			var payload struct {
				Data struct {
					BestFitment       *domain.TireRimBestFitment             `json:"best_fitment"`
					AeroRuleApplies   bool                                   `json:"aero_rule_applies"`
					StandardReference domain.TireRimStandardReferenceContext `json:"standard_reference"`
				} `json:"data"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
				t.Fatal(err)
			}
			best := payload.Data.BestFitment
			if best == nil || best.RimInnerWidthMM != test.expectedInnerWidthMM || best.TireWidthMM != test.expectedTireWidthMM || best.ApplicationProfile != test.expectedProfile || best.SelectionMethod != test.expectedSelectionMethod || payload.Data.AeroRuleApplies != test.expectedAeroRule {
				t.Fatalf("unexpected best pairing: %s", recorder.Body.String())
			}
			if !test.expectedAeroRule && (best.ModelStatus != "pending_profile_calibration" || strings.Contains(recorder.Body.String(), `"aero_target_`)) {
				t.Fatalf("wide path must identify calibration status and omit aero targets: %s", recorder.Body.String())
			}
			if references := payload.Data.StandardReference; references.ISO5775Part1 != "ISO 5775-1:2023" || references.ISO5775Part2 != "ISO 5775-2:2021" || references.ETRTOEdition != "" || references.DTSwissUsedByCalculation || references.DTSwissReferenceRole != "display_only" {
				t.Fatalf("unexpected standard reference roles: %#v", references)
			}
		})
	}
}

func TestSolveTireRimWidthReferenceKeepsCriticalEnteredPairingSeparateFromBestAlternative(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(`{"tire_width_mm":127,"rim_inner_width_mm":35,"rim_system":"hookless","application_profile":"mtb_xc_trail"}`))
	request.Header.Set("Content-Type", "application/json")
	newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data domain.Suggestion `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	result := payload.Data
	if result.ResultKind != "engineering_exact" || result.TireWidthMM != 127 || result.Engineering == nil || result.Engineering.TireWidthMM != 127 || result.Engineering.RimInnerWidthMM != 35 || result.Engineering.Verdict != domain.TireRimEngineeringVerdictCritical || result.Engineering.VerdictReason != domain.TireRimEngineeringReasonAboveMaximum {
		t.Fatalf("entered pairing was replaced or its critical verdict was lost: %#v", result)
	}
	if result.BestFitment == nil || result.BestFitment.RimInnerWidthMM != 35 || result.BestFitment.TireWidthMM != 67 || result.BestFitment.ApplicationProfile != domain.TireRimApplicationProfileMTBXC || result.ApplicationProfile != domain.TireRimApplicationProfileMTBXC || len(result.NormalFitmentRimWidthRanges) == 0 {
		t.Fatalf("missing separate best alternative or entered-width normal range: %#v", result)
	}
}

func TestSolveTireRimWidthReferenceSerializesNullBestForBlockedLegacyAeroTarget(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(`{"rim_inner_width_mm":19,"rim_system":"hookless"}`))
	request.Header.Set("Content-Type", "application/json")
	newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if string(payload.Data["best_fitment"]) != "null" || payload.Data["best_match"] == nil || payload.Data["normal_fitment_tire_width_range"] == nil {
		t.Fatalf("blocked legacy target must not become a best pairing or lose legacy fields: %s", recorder.Body.String())
	}
}

func TestSolveTireRimWidthReferenceReturnsNormalFitmentRangeFor31MmTire(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(`{"tire_width_mm":31,"rim_system":"hookless"}`))
	request.Header.Set("Content-Type", "application/json")
	newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data struct {
			ResultKind         string `json:"result_kind"`
			ApplicationProfile string `json:"application_profile"`
			AeroRuleApplies    bool   `json:"aero_rule_applies"`
			RimWidthRanges     []struct {
				Min int `json:"min"`
				Max int `json:"max"`
			} `json:"rim_width_ranges"`
			NormalFitmentRimWidthRanges []struct {
				Min int `json:"min"`
				Max int `json:"max"`
			} `json:"normal_fitment_rim_width_ranges"`
			StandardReference struct {
				ISO5775Part1             string `json:"iso_5775_part_1"`
				ISO5775Part2             string `json:"iso_5775_part_2"`
				ETRTOEdition             string `json:"etrto_edition"`
				DTSwissUsedByCalculation bool   `json:"dt_swiss_used_by_calculation"`
			} `json:"standard_reference"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.ResultKind != "engineering_normal_fitment" || payload.Data.ApplicationProfile != "gravel_all_road" || payload.Data.AeroRuleApplies || len(payload.Data.RimWidthRanges) != 1 || payload.Data.RimWidthRanges[0].Min != 16 || payload.Data.RimWidthRanges[0].Max != 25 || len(payload.Data.NormalFitmentRimWidthRanges) != 1 || payload.Data.NormalFitmentRimWidthRanges[0].Min != 16 || payload.Data.NormalFitmentRimWidthRanges[0].Max != 25 {
		t.Fatalf("unexpected response: %#v", payload.Data)
	}
	if payload.Data.StandardReference.ISO5775Part1 != "ISO 5775-1:2023" || payload.Data.StandardReference.ISO5775Part2 != "ISO 5775-2:2021" || payload.Data.StandardReference.ETRTOEdition != "" || payload.Data.StandardReference.DTSwissUsedByCalculation {
		t.Fatalf("unexpected standard reference response: %#v", payload.Data.StandardReference)
	}
}

func TestSolveTireRimWidthReferenceKeepsThe105PercentRuleAt30Mm(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(`{"tire_width_mm":30,"rim_system":"hookless"}`))
	request.Header.Set("Content-Type", "application/json")
	newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data struct {
			AeroRuleApplies bool                     `json:"aero_rule_applies"`
			RimWidthRanges  []struct{ Min, Max int } `json:"rim_width_ranges"`
			DerivedMetrics  struct {
				AeroRuleApplies        bool `json:"aero_rule_applies"`
				AeroTargetOuterWidthMM *struct {
					Min float64 `json:"min"`
					Max float64 `json:"max"`
				} `json:"aero_target_outer_width_mm"`
			} `json:"derived_metrics"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !payload.Data.AeroRuleApplies || !payload.Data.DerivedMetrics.AeroRuleApplies {
		t.Fatal("expected 30C to remain within the 105% rule scope")
	}
	if payload.Data.DerivedMetrics.AeroTargetOuterWidthMM == nil || payload.Data.DerivedMetrics.AeroTargetOuterWidthMM.Min != 28.6 {
		t.Fatalf("unexpected 30C aero target: %#v", payload.Data.DerivedMetrics.AeroTargetOuterWidthMM)
	}
	if len(payload.Data.RimWidthRanges) != 1 || payload.Data.RimWidthRanges[0].Min != 23 || payload.Data.RimWidthRanges[0].Max != 23 {
		t.Fatalf("unexpected 30C aero inner-width result: %#v", payload.Data.RimWidthRanges)
	}
}

func TestSolveTireRimWidthReferenceReturnsEngineeringCalculationWhenInnerWidthIsProvided(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(`{"tire_width_mm":32,"rim_system":"hookless","rim_inner_width_mm":25}`))
	request.Header.Set("Content-Type", "application/json")
	newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data struct {
			ResultKind             string `json:"result_kind"`
			RimOuterWidthEstimates []struct {
				RimInnerWidthMM                int     `json:"rim_inner_width_mm"`
				WallThicknessPerSideEstimateMM float64 `json:"wall_thickness_per_side_estimate_mm"`
				RimOuterWidthEstimateMM        float64 `json:"rim_outer_width_estimate_mm"`
			} `json:"rim_outer_width_estimates"`
			Engineering struct {
				Verdict                string   `json:"verdict"`
				AeroRuleApplies        bool     `json:"aero_rule_applies"`
				TireWidthMM            float64  `json:"tire_width_mm"`
				AeroTargetOuterWidthMM *float64 `json:"aero_target_outer_width_mm"`
				AeroTargetTireWidthMM  *float64 `json:"aero_target_tire_width_mm"`
			} `json:"engineering"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.ResultKind != "engineering_exact" || payload.Data.Engineering.Verdict != "recommended" {
		t.Fatalf("unexpected engineering result: %#v", payload.Data)
	}
	if payload.Data.Engineering.TireWidthMM != 32 || payload.Data.Engineering.AeroRuleApplies || payload.Data.Engineering.AeroTargetOuterWidthMM != nil || payload.Data.Engineering.AeroTargetTireWidthMM != nil {
		t.Fatalf("unexpected inverse engineering aero metrics: %#v", payload.Data.Engineering)
	}
	if len(payload.Data.RimOuterWidthEstimates) != 1 || payload.Data.RimOuterWidthEstimates[0].RimInnerWidthMM != 25 || payload.Data.RimOuterWidthEstimates[0].WallThicknessPerSideEstimateMM != 3 || payload.Data.RimOuterWidthEstimates[0].RimOuterWidthEstimateMM != 31 {
		t.Fatalf("unexpected rim outer-width estimate in solve response: %#v", payload.Data.RimOuterWidthEstimates)
	}
}

func TestSolveTireRimWidthReferenceReturnsClosestTireWidthMatchWhenOnlyInnerWidthIsProvided(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(`{"rim_system":"hooked","rim_inner_width_mm":19}`))
	request.Header.Set("Content-Type", "application/json")
	newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data struct {
			AeroTargetTireWidthMM float64 `json:"aero_target_tire_width_mm"`
			BestMatch             struct {
				TireWidthMM                 int     `json:"tire_width_mm"`
				TargetTireWidthDifferenceMM float64 `json:"target_tire_width_difference_mm"`
			} `json:"best_match"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.AeroTargetTireWidthMM != 26.3 || payload.Data.BestMatch.TireWidthMM != 26 || payload.Data.BestMatch.TargetTireWidthDifferenceMM != 0.3 {
		t.Fatalf("unexpected closest tire-width match: %#v", payload.Data.BestMatch)
	}
}

func TestSolveTireRimWidthReferenceSupportsSnowBikeInnerWidthAtOneMillimetreResolution(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(`{"rim_system":"hookless","rim_inner_width_mm":76}`))
	request.Header.Set("Content-Type", "application/json")
	newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 for 76 mm snow-bike rim, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data struct {
			RimInnerWidthMM       int      `json:"rim_inner_width_mm"`
			AeroRuleApplies       bool     `json:"aero_rule_applies"`
			AeroTargetTireWidthMM *float64 `json:"aero_target_tire_width_mm"`
			BestMatch             *struct {
				TireWidthMM                 int     `json:"tire_width_mm"`
				TargetTireWidthDifferenceMM float64 `json:"target_tire_width_difference_mm"`
			} `json:"best_match"`
			NormalFitmentTireWidthRange struct {
				Min int `json:"min"`
				Max int `json:"max"`
			} `json:"normal_fitment_tire_width_range"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode snow-bike response: %v", err)
	}
	if payload.Data.RimInnerWidthMM != 76 || payload.Data.AeroRuleApplies || payload.Data.AeroTargetTireWidthMM != nil || payload.Data.BestMatch != nil || payload.Data.NormalFitmentTireWidthRange.Min != 96 || payload.Data.NormalFitmentTireWidthRange.Max != 132 {
		t.Fatalf("unexpected snow-bike response: %#v", payload.Data)
	}
}

func TestSolveTireRimWidthReferenceRejectsInnerWidthAboveSnowBikeBoundary(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(`{"rim_system":"hookless","rim_inner_width_mm":101}`))
	request.Header.Set("Content-Type", "application/json")
	newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), `"code":"OUT_OF_RANGE"`) {
		t.Fatalf("expected 101 mm to be rejected as out of range, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestSolveTireRimWidthReferenceRejectsUnknownFields(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(`{"tire_width_mm":31,"rim_system":"hookless","unsafe":true}`))
	request.Header.Set("Content-Type", "application/json")
	newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"code":"INVALID_JSON"`) {
		t.Fatalf("expected strict JSON rejection, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestSolveTireRimWidthReferenceReturnsStandardNormalRangeAndNullBestForHookless18Mm(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(`{"tire_width_mm":18,"rim_system":"hookless"}`))
	request.Header.Set("Content-Type", "application/json")
	newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
	var payload struct {
		Data domain.Suggestion `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusOK || payload.Data.BestFitment != nil || !reflect.DeepEqual(payload.Data.NormalFitmentRimWidthRanges, []domain.WidthRange{{Min: 13, Max: 15}}) || payload.Data.NormalFitmentReference.Table != "7" {
		t.Fatalf("expected standard range without a new best value, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestTireRimWidthReferenceMatrixEndpointIsNotRegistered(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/engineering/tire-rim/matrix", nil)
	newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected the display-only DT Swiss matrix endpoint to be absent, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestSolveTireRimWidthReferenceRejectsUnknownAndOutOfScopeApplicationProfilesInEveryInputMode(t *testing.T) {
	for _, body := range []string{
		`{"rim_system":"hookless","tire_width_mm":60,"application_profile":"unknown"}`,
		`{"rim_system":"hookless","rim_inner_width_mm":30,"application_profile":"unknown"}`,
		`{"rim_system":"hookless","tire_width_mm":60,"rim_inner_width_mm":30,"application_profile":"unknown"}`,
		`{"rim_system":"hookless","tire_width_mm":30,"application_profile":"mtb_plus"}`,
		`{"rim_system":"hookless","tire_width_mm":31,"rim_inner_width_mm":25,"application_profile":"road_aero"}`,
		`{"rim_system":"hookless","rim_inner_width_mm":30,"application_profile":"road_aero"}`,
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(body))
		newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
		if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"code":"INVALID_FIELD"`) || !strings.Contains(recorder.Body.String(), `"field":"application_profile"`) {
			t.Fatalf("profile errors must use the existing field-error contract: %d %s", recorder.Code, recorder.Body.String())
		}
	}
}

func TestSolveTireRimWidthReferenceSerializesAbsentPurposeRecommendationsAsNullWithoutLosingNormalRanges(t *testing.T) {
	for _, body := range []string{
		`{"rim_system":"hookless","rim_inner_width_mm":35,"application_profile":"fat_snow"}`,
		`{"rim_system":"hookless","rim_inner_width_mm":35,"tire_width_mm":127}`,
		`{"rim_system":"hookless","rim_inner_width_mm":40,"tire_width_mm":60,"application_profile":"gravel_all_road"}`,
		`{"rim_system":"hooked","rim_inner_width_mm":45,"tire_width_mm":70,"application_profile":"mtb_xc_trail"}`,
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(body))
		newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected a reference result: %d %s", recorder.Code, recorder.Body.String())
		}
		var payload struct {
			Data map[string]json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		if string(payload.Data["best_fitment"]) != "null" {
			t.Fatalf("missing purpose must not receive an unrelated best: %s", recorder.Body.String())
		}
		if rawEngineering, exact := payload.Data["engineering"]; exact {
			var engineering map[string]json.RawMessage
			if err := json.Unmarshal(rawEngineering, &engineering); err != nil {
				t.Fatal(err)
			}
			if string(engineering["recommended_tire_width_range"]) != "null" || engineering["allowed_tire_width_range"] == nil {
				t.Fatalf("exact result lost nullable recommendation or normal envelope: %s", recorder.Body.String())
			}
		} else if string(payload.Data["recommended_fitment_tire_width_range"]) != "null" || payload.Data["normal_fitment_tire_width_range"] == nil {
			t.Fatalf("reverse result lost nullable recommendation or normal envelope: %s", recorder.Body.String())
		}
		if strings.Contains(recorder.Body.String(), `"aero_target_`) {
			t.Fatalf("wide result contains aero output: %s", recorder.Body.String())
		}
	}
}

func TestExistingSolveAPIReturnsNullFatOptimaAndPressureWithoutLosingNormalRangesOrEnteredVerdicts(t *testing.T) {
	for _, body := range []string{
		`{"rim_system":"hookless","tire_width_mm":115}`,
		`{"rim_system":"hooked","tire_width_mm":105}`,
		`{"rim_system":"hookless","rim_inner_width_mm":76}`,
		`{"rim_system":"hookless","rim_inner_width_mm":84}`,
		`{"rim_system":"hookless","rim_inner_width_mm":86}`,
		`{"rim_system":"hookless","rim_inner_width_mm":100}`,
		`{"rim_system":"hookless","rim_inner_width_mm":76,"tire_width_mm":115}`,
		`{"rim_system":"hookless","rim_inner_width_mm":100,"tire_width_mm":126}`,
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(body))
		newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("Fat reference request failed: %d %s", recorder.Code, recorder.Body.String())
		}
		var payload struct {
			Data map[string]json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		if string(payload.Data["best_fitment"]) != "null" || strings.Contains(recorder.Body.String(), `"aero_target_`) {
			t.Fatalf("unreviewed Fat request acquired an optimum or aero output: %s", recorder.Body.String())
		}
		if engineeringJSON, exact := payload.Data["engineering"]; exact {
			var engineering map[string]json.RawMessage
			if err := json.Unmarshal(engineeringJSON, &engineering); err != nil {
				t.Fatal(err)
			}
			if string(engineering["maximum_pressure"]) != "null" || string(engineering["recommended_tire_width_range"]) != "null" || engineering["allowed_tire_width_range"] == nil {
				t.Fatalf("Fat exact response lost explicit nulls or normal range: %s", recorder.Body.String())
			}
			if strings.Contains(body, `"tire_width_mm":126`) && (string(engineering["verdict"]) != `"reference"` || string(engineering["verdict_reason"]) != `"no_recommended_window"` || string(engineering["tire_width_mm"]) != "126") {
				t.Fatalf("standard-covered Fat pairing lost its normal verdict or input width: %s", recorder.Body.String())
			}
		} else if _, reverse := payload.Data["rim_inner_width_mm"]; reverse {
			if string(payload.Data["recommended_fitment_tire_width_range"]) != "null" || payload.Data["normal_fitment_tire_width_range"] == nil {
				t.Fatalf("Fat reverse response lost explicit null or normal range: %s", recorder.Body.String())
			}
		} else if payload.Data["normal_fitment_rim_width_ranges"] == nil || strings.Contains(recorder.Body.String(), `"recommended_rim_width_ranges"`) {
			t.Fatalf("Fat tire response lost normal ranges or retained a legacy optimization window: %s", recorder.Body.String())
		}
	}
}

func TestExistingSolveAPIRetainsReviewedFatSuggestionAndRejectsDedicatedOversizeSystemsWithoutClamping(t *testing.T) {
	for _, body := range []string{
		`{"rim_system":"hookless","tire_width_mm":105}`,
		`{"rim_system":"hookless","rim_inner_width_mm":85}`,
		`{"rim_system":"hookless","rim_inner_width_mm":85,"tire_width_mm":115}`,
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(body))
		newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
		var payload struct {
			Data struct {
				BestFitment *domain.TireRimBestFitment `json:"best_fitment"`
			} `json:"data"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		best := payload.Data.BestFitment
		if recorder.Code != http.StatusOK || best == nil || best.RimInnerWidthMM != 85 || best.TireWidthMM != 105 || best.TargetTireWidthMM != 105.41 || best.ModelStatus != "pending_profile_calibration" || best.FitmentRuleReference.ManufacturerSourceURLs[0] != "https://enve.com/pages/tire-compatibility" || strings.Contains(recorder.Body.String(), `"aero_target_`) {
			t.Fatalf("reviewed Fat suggestion did not pass through the shared API: %s", recorder.Body.String())
		}
		if strings.Contains(body, `"tire_width_mm":115`) {
			var exactPayload struct {
				Data domain.Suggestion `json:"data"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &exactPayload); err != nil {
				t.Fatal(err)
			}
			if exactPayload.Data.TireWidthMM != 115 || exactPayload.Data.Engineering.Verdict != domain.TireRimEngineeringVerdictReference || exactPayload.Data.Engineering.MaximumPressure != nil {
				t.Fatalf("suggestion overwrote the entered Fat pairing or restored generic pressure: %s", recorder.Body.String())
			}
		}
	}
	for _, body := range []string{
		`{"rim_system":"hookless","tire_width_mm":158}`,
		`{"rim_system":"hookless","rim_inner_width_mm":100,"tire_width_mm":158}`,
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(body))
		newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
		if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), `"code":"OUT_OF_RANGE"`) || !strings.Contains(recorder.Body.String(), `"field":"tire_width_mm"`) {
			t.Fatalf("dedicated 6.2 inch request was clamped or lost the error contract: %s", recorder.Body.String())
		}
	}
}

func TestSolveTireRimWidthReferenceRetainsCalibratedManufacturerRolesAndEnteredPairingThroughTheExistingAPI(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(`{"rim_system":"hooked","rim_inner_width_mm":35,"tire_width_mm":65,"application_profile":"mtb_xc_trail"}`))
	newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected calibrated result: %d %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data domain.Suggestion `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	result := payload.Data
	if result.TireWidthMM != 65 || result.Engineering == nil || result.Engineering.TireWidthMM != 65 || result.Engineering.Verdict != domain.TireRimEngineeringVerdictRecommended || result.BestFitment == nil || result.BestFitment.TireWidthMM != 68 {
		t.Fatalf("best alternative replaced the entered pairing: %#v", result)
	}
	if result.ModelVersion != "tire-rim-engineering-reference-v11" || result.Engineering.FitmentRuleReference.NormalRangeBasis != "iso_5775_1_2023_clause_4_4_table_7" || result.BestFitment.FitmentRuleReference.RecommendedRangeBasis != "reviewed_manufacturer_optimal_bands_project_envelope" || len(result.BestFitment.FitmentRuleReference.ManufacturerSourceURLs) != 2 || len(result.Engineering.FitmentRuleReference.ManufacturerSourceURLs) != 0 || result.StandardReference.DTSwissUsedByCalculation || result.BestFitment.ModelStatus != "pending_profile_calibration" {
		t.Fatalf("calibration lost project and source identities: %#v", result)
	}
	if strings.Contains(recorder.Body.String(), `"aero_target_`) {
		t.Fatalf("mountain calibration emitted aero output: %s", recorder.Body.String())
	}
}
