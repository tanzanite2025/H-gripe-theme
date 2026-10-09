package tirerim

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTireRimWidthReferenceTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewTireRimWidthReferenceHandler().RegisterTireRimWidthReferenceHTTPRoutes(router.Group("/api/v1/engineering/tire-rim"))
	return router
}

func TestSolveTireRimWidthReferenceReturnsEngineering31MmResult(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(`{"tire_width_mm":31,"rim_system":"hookless"}`))
	request.Header.Set("Content-Type", "application/json")
	newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data struct {
			ResultKind     string `json:"result_kind"`
			RimWidthRanges []struct {
				Min int `json:"min"`
				Max int `json:"max"`
			} `json:"rim_width_ranges"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.ResultKind != "engineering_recommended" || len(payload.Data.RimWidthRanges) != 1 || payload.Data.RimWidthRanges[0].Min != 23 || payload.Data.RimWidthRanges[0].Max != 25 {
		t.Fatalf("unexpected response: %#v", payload.Data)
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
			ResultKind  string `json:"result_kind"`
			Engineering struct {
				Verdict                string  `json:"verdict"`
				InflatedTireWidthMM    float64 `json:"inflated_tire_width_mm"`
				AeroTargetOuterWidthMM float64 `json:"aero_target_outer_width_mm"`
			} `json:"engineering"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.ResultKind != "engineering_exact" || payload.Data.Engineering.Verdict != "recommended" {
		t.Fatalf("unexpected engineering result: %#v", payload.Data)
	}
	if payload.Data.Engineering.InflatedTireWidthMM != 34.4 || payload.Data.Engineering.AeroTargetOuterWidthMM != 36.1 {
		t.Fatalf("unexpected engineering aero metrics: %#v", payload.Data.Engineering)
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

func TestSolveTireRimWidthReferenceReturnsNoBracketForHookless18Mm(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(`{"tire_width_mm":18,"rim_system":"hookless"}`))
	request.Header.Set("Content-Type", "application/json")
	newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), `"code":"NO_PUBLISHED_BRACKET"`) {
		t.Fatalf("expected no-bracket response, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestGetTireRimWidthReferenceMatrixIncludesMethodologyAndSourceRows(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/engineering/tire-rim/matrix", nil)
	newTireRimWidthReferenceTestRouter().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	for _, fragment := range []string{`"model_version":"dt-swiss-tss-tc-v1"`, `"methodology"`, `"possible_rows"`, `"tire_width_mm":30`} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected matrix to contain %s, got %s", fragment, body)
		}
	}
}
