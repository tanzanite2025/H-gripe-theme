package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	tireRimAPI "commerce-platform/internal/api/v1/tirerim"
	"commerce-platform/internal/domain/tirerim"
	"commerce-platform/internal/repository"

	"github.com/gin-gonic/gin"
)

func TestSchwalbeTireCatalogRimWidthReferenceForRowsUsesBackendEngineeringWindow(t *testing.T) {
	nominalWidthMM := 31
	rows := []schwalbeTireCatalogSelectorRow{{
		item: repository.SchwalbeTireCatalogItem{
			ArticleNo: "TEST-31",
			ModelName: "Reviewed model",
			ETRTO:     "31-622",
		},
		dimensions: schwalbeTireCatalogSelectorDimensions{
			NominalTireWidthMM: &nominalWidthMM,
		},
	}}
	compatibility := map[string][]SchwalbeTireCatalogSelectorRimCompatibility{
		"TEST-31": {{RimSystem: "hooked", Status: "supported"}},
	}

	guidanceByArticle := schwalbeTireCatalogRimWidthReferenceForRows(rows, compatibility)
	guidance := guidanceByArticle["TEST-31"]
	if len(guidance) != 1 {
		t.Fatalf("expected one hooked reference result, got %#v", guidance)
	}
	if guidance[0].RimSystem != tirerim.RimSystemHooked || guidance[0].ResultKind != "engineering_normal_fitment" {
		t.Fatalf("expected backend hooked normal fitment result, got %#v", guidance[0])
	}
	if len(guidance[0].RimWidthRanges) != 1 || guidance[0].RimWidthRanges[0] != (tirerim.WidthRange{Min: 16, Max: 25}) {
		t.Fatalf("unexpected normal fitment hooked range: %#v", guidance[0].RimWidthRanges)
	}
	directSuggestion, err := tirerim.ResolveTireRimWidthReference(nominalWidthMM, tirerim.RimSystemHooked)
	if err != nil {
		t.Fatal(err)
	}
	if guidance[0].BestFitment == nil || !reflect.DeepEqual(guidance[0], *directSuggestion) {
		t.Fatalf("catalog guidance must retain the same best pairing and legacy fields as the domain response: %#v", guidance[0])
	}
}

func TestSchwalbeTireCatalogUsesUpgradedMountainPurposeRulesFromTheSharedEngine(t *testing.T) {
	tireWidthMM := 60
	rows := []schwalbeTireCatalogSelectorRow{{
		item:       repository.SchwalbeTireCatalogItem{ArticleNo: "TEST-MTB-60", ETRTO: "60-622"},
		dimensions: schwalbeTireCatalogSelectorDimensions{NominalTireWidthMM: &tireWidthMM},
	}}
	compatibility := map[string][]SchwalbeTireCatalogSelectorRimCompatibility{
		"TEST-MTB-60": {{RimSystem: "hookless", Status: "supported"}, {RimSystem: "hooked", Status: "supported"}},
	}
	guidance := schwalbeTireCatalogRimWidthReferenceForRows(rows, compatibility)["TEST-MTB-60"]
	if len(guidance) != 2 {
		t.Fatalf("expected both existing system references: %#v", guidance)
	}
	for _, reference := range guidance {
		direct, err := tirerim.ResolveTireRimWidthReference(tireWidthMM, reference.RimSystem)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(reference, *direct) || reference.BestFitment == nil || reference.ApplicationProfile != tirerim.TireRimApplicationProfileMTBXC {
			t.Fatalf("catalog diverged from unique calculation source: %#v", reference)
		}
		if reference.RimSystem == tirerim.RimSystemHookless && (reference.BestFitment.RimInnerWidthMM != 28 || len(reference.BestFitment.FitmentRuleReference.ManufacturerSourceURLs) == 0) {
			t.Fatalf("catalog retained old MTB recommendation: %#v", reference)
		}
		if reference.RimSystem == tirerim.RimSystemHooked && (reference.BestFitment.RimInnerWidthMM != 31 || reference.BestFitment.FitmentRuleReference.RecommendedRangeBasis != "reviewed_manufacturer_optimal_bands_project_envelope") {
			t.Fatalf("catalog retained old hooked MTB recommendation: %#v", reference)
		}
	}
}

func TestSchwalbeTireCatalogUsesReviewedFatSuggestionAndNullUnreviewedOptimaFromTheUniqueEngine(t *testing.T) {
	for _, tireWidthMM := range []int{105, 115, 127} {
		rows := []schwalbeTireCatalogSelectorRow{{
			item:       repository.SchwalbeTireCatalogItem{ArticleNo: "TEST-FAT"},
			dimensions: schwalbeTireCatalogSelectorDimensions{NominalTireWidthMM: &tireWidthMM},
		}}
		compatibility := map[string][]SchwalbeTireCatalogSelectorRimCompatibility{
			"TEST-FAT": {{RimSystem: "hookless", Status: "supported"}, {RimSystem: "hooked", Status: "supported"}},
		}
		guidance := schwalbeTireCatalogRimWidthReferenceForRows(rows, compatibility)["TEST-FAT"]
		if len(guidance) != 2 {
			t.Fatalf("Fat engine change removed catalog reference systems: %#v", guidance)
		}
		for _, reference := range guidance {
			sharedResult, err := tirerim.ResolveTireRimWidthReference(tireWidthMM, reference.RimSystem)
			if err != nil || !reflect.DeepEqual(reference, *sharedResult) || len(reference.NormalFitmentRimWidthRanges) == 0 {
				t.Fatalf("Fat catalog diverged from the unique engine or lost reference ranges: %#v error=%v", reference, err)
			}
			if tireWidthMM == 105 && reference.RimSystem == tirerim.RimSystemHookless {
				if reference.BestFitment == nil || reference.BestFitment.RimInnerWidthMM != 85 || reference.BestFitment.ModelStatus != "pending_profile_calibration" {
					t.Fatalf("catalog did not retain reviewed Fat project suggestion: %#v", reference)
				}
			} else if reference.BestFitment != nil || len(reference.RecommendedRimWidthRanges) != 0 {
				t.Fatalf("catalog restored a legacy Fat midpoint as an optimum: %#v", reference)
			}
		}
	}
}

func TestSchwalbeCatalogGuidanceMatchesExistingSolveAPIAcrossTireWidthInputBoundsAndBothRimSystems(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	tireRimAPI.NewTireRimWidthReferenceHandler().RegisterTireRimWidthReferenceHTTPRoutes(router.Group("/api/v1/engineering/tire-rim"))

	for tireWidthMM := tirerim.MinInputTireWidthMM; tireWidthMM <= tirerim.MaxInputTireWidthMM; tireWidthMM++ {
		for _, system := range []tirerim.RimSystem{tirerim.RimSystemHooked, tirerim.RimSystemHookless} {
			t.Run(fmt.Sprintf("%s_%d_mm", system, tireWidthMM), func(t *testing.T) {
				articleNo := fmt.Sprintf("REGRESSION-%s-%d", system, tireWidthMM)
				rows := []schwalbeTireCatalogSelectorRow{{
					item:       repository.SchwalbeTireCatalogItem{ArticleNo: articleNo},
					dimensions: schwalbeTireCatalogSelectorDimensions{NominalTireWidthMM: &tireWidthMM},
				}}
				compatibility := map[string][]SchwalbeTireCatalogSelectorRimCompatibility{
					articleNo: {{RimSystem: string(system), Status: "supported"}},
				}
				catalogGuidance := schwalbeTireCatalogRimWidthReferenceForRows(rows, compatibility)[articleNo]
				requestBody := fmt.Sprintf(`{"tire_width_mm":%d,"rim_system":%q}`, tireWidthMM, system)
				request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(requestBody))
				request.Header.Set("Content-Type", "application/json")
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, request)

				if recorder.Code != http.StatusOK || len(catalogGuidance) != 1 {
					t.Fatalf("expected one catalog result and a successful solve: status=%d api=%s catalog=%#v", recorder.Code, recorder.Body.String(), catalogGuidance)
				}
				var apiResponse struct {
					Data map[string]any `json:"data"`
				}
				if err := json.Unmarshal(recorder.Body.Bytes(), &apiResponse); err != nil {
					t.Fatal(err)
				}
				catalogJSON, err := json.Marshal(catalogGuidance[0])
				if err != nil {
					t.Fatal(err)
				}
				var catalogResponse map[string]any
				if err := json.Unmarshal(catalogJSON, &catalogResponse); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(catalogResponse, apiResponse.Data) {
					t.Fatalf("catalog and existing solve API returned different recommendation, ranges, purpose or metadata: catalog=%s api=%s", catalogJSON, recorder.Body.String())
				}
			})
		}
	}
}

func TestSchwalbeCatalogDoesNotInferHooklessGuidanceFromUnknownOrUnsupportedCompatibilityFacts(t *testing.T) {
	tireWidthMM := 60
	rows := []schwalbeTireCatalogSelectorRow{{
		item:       repository.SchwalbeTireCatalogItem{ArticleNo: "REGRESSION-COMPATIBILITY"},
		dimensions: schwalbeTireCatalogSelectorDimensions{NominalTireWidthMM: &tireWidthMM},
	}}
	for _, hooklessStatus := range []string{"", "unknown", "unsupported"} {
		t.Run("hookless_status_"+hooklessStatus, func(t *testing.T) {
			compatibility := map[string][]SchwalbeTireCatalogSelectorRimCompatibility{
				"REGRESSION-COMPATIBILITY": {
					{RimSystem: "hooked", Status: "supported"},
					{RimSystem: "hookless", Status: hooklessStatus},
				},
			}
			guidance := schwalbeTireCatalogRimWidthReferenceForRows(rows, compatibility)["REGRESSION-COMPATIBILITY"]
			if len(guidance) != 1 || guidance[0].RimSystem != tirerim.RimSystemHooked {
				t.Fatalf("size recommendation must not create a hookless compatibility fact: %#v", guidance)
			}
		})
	}
	if guidance := schwalbeTireCatalogRimWidthReferenceForRows(rows, nil); len(guidance) != 0 {
		t.Fatalf("missing compatibility facts must not produce rim guidance: %#v", guidance)
	}
	rows[0].dimensions.NominalTireWidthMM = nil
	compatibility := map[string][]SchwalbeTireCatalogSelectorRimCompatibility{
		"REGRESSION-COMPATIBILITY": {{RimSystem: "hooked", Status: "supported"}, {RimSystem: "hookless", Status: "supported"}},
	}
	if guidance := schwalbeTireCatalogRimWidthReferenceForRows(rows, compatibility); len(guidance) != 0 {
		t.Fatalf("missing nominal tire width must not acquire a guessed result: %#v", guidance)
	}
}
