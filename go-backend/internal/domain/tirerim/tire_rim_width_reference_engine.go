package tirerim

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

const (
	MinInputTireWidthMM = 18
	MaxInputTireWidthMM = 132
)

var (
	ErrInvalidRimSystem                          = errors.New("invalid rim system")
	ErrInvalidTireWidth                          = errors.New("invalid tire width")
	ErrNoPublishedBracket                        = errors.New("no published tire-width bracket")
	ErrInvalidTireRimApplicationProfile          = errors.New("invalid tire/rim application profile")
	ErrTireRimApplicationProfileOutsideAeroScope = errors.New("application profile does not match the aero-rule scope")
)

type RimSystem string

const (
	RimSystemHookless RimSystem = "hookless"
	RimSystemHooked   RimSystem = "hooked"
)

type WidthRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type SourceRow struct {
	TireWidthMM int    `json:"tire_width_mm"`
	Kind        string `json:"kind"`
}

type Calculation struct {
	Method             string  `json:"method"`
	LowerTireWidthMM   int     `json:"lower_tire_width_mm"`
	UpperTireWidthMM   int     `json:"upper_tire_width_mm"`
	InterpolationRatio float64 `json:"interpolation_ratio"`
}

type NumericRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// TireRimBestFitment is one selected pairing, separate from the full normal
// fitment ranges and the entered pairing's verdict. Wide-tire selections use
// purpose-specific project windows; they are pending full calibration and
// do not claim to simulate impact absorption, tire pressure, or casing load.
type TireRimBestFitment struct {
	RimInnerWidthMM             int                         `json:"rim_inner_width_mm"`
	TireWidthMM                 int                         `json:"tire_width_mm"`
	ApplicationProfile          TireRimApplicationProfile   `json:"application_profile"`
	SelectionMethod             string                      `json:"selection_method"`
	ReasonCode                  string                      `json:"reason_code"`
	ModelStatus                 string                      `json:"model_status"`
	TargetTireWidthMM           float64                     `json:"target_tire_width_mm"`
	TargetTireWidthDifferenceMM float64                     `json:"target_tire_width_difference_mm"`
	FitmentRuleReference        TireRimFitmentRuleReference `json:"fitment_rule_reference"`
}

// TireRimFitmentRuleReference identifies the range layer's construction.
// Normal results cite ISO Table 7; best_fitment retains its project guidance.
// Neither reference grants exact product approval.
type TireRimFitmentRuleReference struct {
	NormalRangeBasis       string   `json:"normal_range_basis"`
	RecommendedRangeBasis  string   `json:"recommended_range_basis"`
	ManufacturerSourceURLs []string `json:"manufacturer_source_urls"`
	Interpolated           bool     `json:"interpolated"`
}

// TireRimApplicationProfile identifies the project's default use-oriented
// calculation layer. The calculator infers it from the entered tire width.
// Explicit API profiles affect recommendation windows, not ISO normal ranges.
// These are project model labels, not
// ISO or ETRTO classifications.
type TireRimApplicationProfile string

const (
	TireRimApplicationProfileRoadAero      TireRimApplicationProfile = "road_aero"
	TireRimApplicationProfileGravelAllRoad TireRimApplicationProfile = "gravel_all_road"
	TireRimApplicationProfileMTBXC         TireRimApplicationProfile = "mtb_xc_trail"
	TireRimApplicationProfileMTBPlus       TireRimApplicationProfile = "mtb_plus"
	TireRimApplicationProfileFatSnow       TireRimApplicationProfile = "fat_snow"
	TireRimApplicationProfileAutoByWidth   TireRimApplicationProfile = "automatic_by_width"
)

// TireRimStandardReferenceContext describes what the named standards can
// provide to the result. It deliberately contains source identity and role,
// not a fabricated universal "best" value.
type TireRimStandardReferenceContext struct {
	ISO5775Part1             string `json:"iso_5775_part_1"`
	ISO5775Part2             string `json:"iso_5775_part_2"`
	ETRTOEdition             string `json:"etrto_edition"`
	DTSwissUsedByCalculation bool   `json:"dt_swiss_used_by_calculation"`
	DTSwissReferenceRole     string `json:"dt_swiss_reference_role"`
}

var tireRimStandardReferenceContext = TireRimStandardReferenceContext{
	ISO5775Part1: "ISO 5775-1:2023",
	ISO5775Part2: "ISO 5775-2:2021",
	// No ETRTO manual table is applied by this width-table integration.
	ETRTOEdition:             "",
	DTSwissUsedByCalculation: false,
	DTSwissReferenceRole:     "display_only",
}

// TireRimNormalFitmentTableReference identifies the actual source of normal
// width ranges, independently of best_fitment's optimization guidance.
type TireRimNormalFitmentTableReference struct {
	Standard   string `json:"standard"`
	Clause     string `json:"clause"`
	Table      string `json:"table"`
	SourceKind string `json:"source_kind"`
	Scope      string `json:"scope"`
}

var iso5775Part12023NormalFitmentTableReference = TireRimNormalFitmentTableReference{
	Standard: "ISO 5775-1:2023", Clause: "4.4", Table: "7",
	SourceKind: "user_provided_standard_width_matrix_image",
	Scope:      "nominal_width_pairing_only",
}

const TireRimStandardNormalFitmentMinimumInnerWidthMM = 13

// These cells are transcribed from the user-provided ISO 5775-1:2023 width
// matrix, recorded in the specification section 19. They are not interpolated,
// expanded by manufacturer optimization windows, or claimed as ETRTO 2026 data.
// The shared width relationship does not assert exact-model hookless approval.
var iso5775Part12023NormalWidthPairingTable = [...]struct {
	TireWidthRangeMM     WidthRange
	RimInnerWidthRangeMM WidthRange
}{
	{WidthRange{18, 19}, WidthRange{13, 15}},
	{WidthRange{20, 21}, WidthRange{13, 17}},
	{WidthRange{22, 24}, WidthRange{13, 20}},
	{WidthRange{25, 27}, WidthRange{13, 22}},
	{WidthRange{28, 28}, WidthRange{15, 23}},
	{WidthRange{29, 34}, WidthRange{16, 25}},
	{WidthRange{35, 46}, WidthRange{17, 27}},
	{WidthRange{47, 57}, WidthRange{17, 30}},
	{WidthRange{58, 65}, WidthRange{21, 35}},
	{WidthRange{66, 71}, WidthRange{25, 43}},
	{WidthRange{72, 83}, WidthRange{31, 53}},
	{WidthRange{84, 95}, WidthRange{41, 64}},
	{WidthRange{96, 113}, WidthRange{59, 89}},
	{WidthRange{114, 132}, WidthRange{72, 100}},
}

func findISO5775Part12023NormalRimInnerWidthRangesForTireWidth(tireWidthMM int) []WidthRange {
	for _, row := range iso5775Part12023NormalWidthPairingTable {
		if tireWidthMM >= row.TireWidthRangeMM.Min && tireWidthMM <= row.TireWidthRangeMM.Max {
			return []WidthRange{row.RimInnerWidthRangeMM}
		}
	}
	return []WidthRange{}
}

func findISO5775Part12023NormalTireWidthRangeForRimInnerWidth(innerWidthMM int) WidthRange {
	var result WidthRange
	for _, row := range iso5775Part12023NormalWidthPairingTable {
		if innerWidthMM < row.RimInnerWidthRangeMM.Min || innerWidthMM > row.RimInnerWidthRangeMM.Max {
			continue
		}
		if result == (WidthRange{}) {
			result.Min = row.TireWidthRangeMM.Min
		}
		result.Max = row.TireWidthRangeMM.Max
	}
	return result
}

func buildISO5775Part12023NormalFitmentRuleReference() TireRimFitmentRuleReference {
	return TireRimFitmentRuleReference{
		NormalRangeBasis:       "iso_5775_1_2023_clause_4_4_table_7",
		ManufacturerSourceURLs: []string{},
	}
}

func classifyTireRimApplicationProfile(tireWidthMM int) TireRimApplicationProfile {
	switch {
	case tireWidthMM <= TireRimEngineeringAeroRuleMaximumTireWidthMM:
		return TireRimApplicationProfileRoadAero
	case tireWidthMM <= 50:
		return TireRimApplicationProfileGravelAllRoad
	case tireWidthMM <= 70:
		return TireRimApplicationProfileMTBXC
	case tireWidthMM <= 85:
		return TireRimApplicationProfileMTBPlus
	default:
		return TireRimApplicationProfileFatSnow
	}
}

func NormalizeTireRimApplicationProfile(profile string) TireRimApplicationProfile {
	return TireRimApplicationProfile(strings.ToLower(strings.TrimSpace(profile)))
}

func validateOptionalTireRimApplicationProfile(profiles []TireRimApplicationProfile) (TireRimApplicationProfile, error) {
	if len(profiles) > 1 {
		return "", ErrInvalidTireRimApplicationProfile
	}
	profile := TireRimApplicationProfileAutoByWidth
	if len(profiles) == 1 && profiles[0] != "" {
		profile = profiles[0]
	}
	switch profile {
	case TireRimApplicationProfileAutoByWidth, TireRimApplicationProfileRoadAero,
		TireRimApplicationProfileGravelAllRoad, TireRimApplicationProfileMTBXC,
		TireRimApplicationProfileMTBPlus, TireRimApplicationProfileFatSnow:
		return profile, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidTireRimApplicationProfile, profile)
	}
}

func resolveTireRimApplicationProfileForKnownTireWidth(tireWidthMM int, profiles []TireRimApplicationProfile) (TireRimApplicationProfile, string, error) {
	requestedProfile, err := validateOptionalTireRimApplicationProfile(profiles)
	if err != nil {
		return "", "", err
	}
	if requestedProfile == TireRimApplicationProfileAutoByWidth {
		return classifyTireRimApplicationProfile(tireWidthMM), "automatic_by_tire_width", nil
	}
	if (tireWidthMM <= TireRimEngineeringAeroRuleMaximumTireWidthMM) != (requestedProfile == TireRimApplicationProfileRoadAero) {
		return "", "", ErrTireRimApplicationProfileOutsideAeroScope
	}
	return requestedProfile, "explicit", nil
}

// RimOuterWidthEstimate uses a nominal wall thickness per side for display
// and the narrow-tire geometric target. Exact width comes from the rim maker.
type RimOuterWidthEstimate struct {
	RimInnerWidthMM                int     `json:"rim_inner_width_mm"`
	WallThicknessPerSideEstimateMM float64 `json:"wall_thickness_per_side_estimate_mm"`
	RimOuterWidthEstimateMM        float64 `json:"rim_outer_width_estimate_mm"`
}

type DerivedMetrics struct {
	Status          string       `json:"status"`
	TireWidthMM     NumericRange `json:"tire_width_mm"`
	AeroRuleApplies bool         `json:"aero_rule_applies"`
	// AeroTargetOuterWidthMM is populated only when the entered tire width is
	// within the 105% rule's <=30C scope. For wider tires, normal fitment
	// ranges are used and this field is omitted.
	AeroTargetOuterWidthMM *NumericRange `json:"aero_target_outer_width_mm,omitempty"`
}

type Suggestion struct {
	TireWidthMM                 int                                     `json:"tire_width_mm"`
	RimSystem                   RimSystem                               `json:"rim_system"`
	ApplicationProfile          TireRimApplicationProfile               `json:"application_profile"`
	ApplicationProfileSelection string                                  `json:"application_profile_selection"`
	ResultKind                  string                                  `json:"result_kind"`
	AeroRuleApplies             bool                                    `json:"aero_rule_applies"`
	BestFitment                 *TireRimBestFitment                     `json:"best_fitment"`
	RimWidthRanges              []WidthRange                            `json:"rim_width_ranges"`
	NormalFitmentRimWidthRanges []WidthRange                            `json:"normal_fitment_rim_width_ranges"`
	RecommendedRimWidthRanges   []WidthRange                            `json:"recommended_rim_width_ranges,omitempty"`
	RimOuterWidthEstimates      []RimOuterWidthEstimate                 `json:"rim_outer_width_estimates"`
	SourceRows                  []SourceRow                             `json:"source_rows"`
	Calculation                 *Calculation                            `json:"calculation,omitempty"`
	DerivedMetrics              DerivedMetrics                          `json:"derived_metrics"`
	ModelVersion                string                                  `json:"model_version"`
	KnowledgeAsOf               string                                  `json:"knowledge_as_of"`
	Limitations                 []string                                `json:"limitations"`
	StandardReference           TireRimStandardReferenceContext         `json:"standard_reference"`
	NormalFitmentReference      TireRimNormalFitmentTableReference      `json:"normal_fitment_reference"`
	Engineering                 *TireRimEngineeringReferenceCalculation `json:"engineering,omitempty"`
}

// TireWidthMatch is the single tire width whose input width is closest to the
// target tire width for a known rim inner width.
// The 105% relationship is used once to derive that target from the estimated
// rim outer width; the candidate tire input is compared directly.
// The match is chosen from the model-supported 18–132 mm search range; the
// browser must not rank a separate list of candidates.
type TireWidthMatch struct {
	TireWidthMM                 int     `json:"tire_width_mm"`
	TargetTireWidthDifferenceMM float64 `json:"target_tire_width_difference_mm"`
}

// RimToTireSuggestion is the reverse-lookup response for a known rim inner
// width. The 105% aero target is returned only when the resulting target tire
// width is within the <=30C rule scope. Wider rims return their normal and
// preferred tire-width fitment ranges instead.
type RimToTireSuggestion struct {
	RimInnerWidthMM                  int                                `json:"rim_inner_width_mm"`
	RimSystem                        RimSystem                          `json:"rim_system"`
	ApplicationProfile               TireRimApplicationProfile          `json:"application_profile"`
	ResolvedApplicationProfile       TireRimApplicationProfile          `json:"resolved_application_profile"`
	ApplicationProfileSelection      string                             `json:"application_profile_selection"`
	AeroRuleApplies                  bool                               `json:"aero_rule_applies"`
	AeroTargetTireWidthMM            *float64                           `json:"aero_target_tire_width_mm,omitempty"`
	BestMatch                        *TireWidthMatch                    `json:"best_match,omitempty"`
	BestFitment                      *TireRimBestFitment                `json:"best_fitment"`
	NormalFitmentTireWidthRange      WidthRange                         `json:"normal_fitment_tire_width_range"`
	RecommendedFitmentTireWidthRange *WidthRange                        `json:"recommended_fitment_tire_width_range"`
	RimOuterWidthEstimate            RimOuterWidthEstimate              `json:"rim_outer_width_estimate"`
	ModelVersion                     string                             `json:"model_version"`
	KnowledgeAsOf                    string                             `json:"knowledge_as_of"`
	Limitations                      []string                           `json:"limitations"`
	StandardReference                TireRimStandardReferenceContext    `json:"standard_reference"`
	NormalFitmentReference           TireRimNormalFitmentTableReference `json:"normal_fitment_reference"`
	FitmentRuleReference             TireRimFitmentRuleReference        `json:"fitment_rule_reference"`
}

// tireRimEngineeringTireToRimWidthRatio expresses the aero relationship used
// by the calculator: tire width = rim outer width × 1.05.
const tireRimEngineeringTireToRimWidthRatio = 1.05

// TireRimEngineeringAeroRuleMaximumTireWidthMM is the inclusive upper limit
// of the 105% aero rule. Wider tires use the normal fitment model only.
const TireRimEngineeringAeroRuleMaximumTireWidthMM = 30

func ValidateTireWidth(width int) error {
	if width < MinInputTireWidthMM || width > MaxInputTireWidthMM {
		return fmt.Errorf("%w: %d mm is outside [%d, %d]", ErrInvalidTireWidth, width, MinInputTireWidthMM, MaxInputTireWidthMM)
	}
	return nil
}

func ValidateRimSystem(system RimSystem) error {
	if system != RimSystemHookless && system != RimSystemHooked {
		return fmt.Errorf("%w: %q", ErrInvalidRimSystem, system)
	}
	return nil
}

func ResolveTireRimWidthReference(width int, system RimSystem, profiles ...TireRimApplicationProfile) (*Suggestion, error) {
	if err := ValidateTireWidth(width); err != nil {
		return nil, err
	}
	if err := ValidateRimSystem(system); err != nil {
		return nil, err
	}
	profile, profileSelection, err := resolveTireRimApplicationProfileForKnownTireWidth(width, profiles)
	if err != nil {
		return nil, err
	}
	if width > TireRimEngineeringAeroRuleMaximumTireWidthMM {
		return resolveNormalTireRimFitmentForKnownTireWidth(width, system, profile, profileSelection)
	}
	result, err := resolveAeroTireRimFitmentForKnownTireWidth(width, system)
	if result != nil {
		result.ApplicationProfileSelection = profileSelection
	}
	return result, err
}

// resolveAeroTireRimFitmentForKnownTireWidth applies the 105% rule to a tire
// width within its inclusive <=30C scope. It returns one closest inner-width
// target and does not use the normal fitment ranges to rank alternatives.
func resolveAeroTireRimFitmentForKnownTireWidth(width int, system RimSystem) (*Suggestion, error) {
	targetOuterWidthMM := float64(width) / tireRimEngineeringTireToRimWidthRatio
	targetInnerWidthMM := targetOuterWidthMM - 2*TireRimEngineeringRimWallThicknessEstimatePerSideMM
	selectedInnerWidthMM := int(math.Round(targetInnerWidthMM))
	selectedInnerWidths := []int{}
	if selectedInnerWidthMM >= TireRimEngineeringMinInnerWidthMM && selectedInnerWidthMM <= TireRimEngineeringMaxInnerWidthMM {
		selectedInnerWidths = append(selectedInnerWidths, selectedInnerWidthMM)
	}
	selectedRimTargetTireWidthMM := (float64(selectedInnerWidthMM) + 2*TireRimEngineeringRimWallThicknessEstimatePerSideMM) * tireRimEngineeringTireToRimWidthRatio
	normalInnerWidthRanges := findInnerWidthRangesForTireWidth(system, width, false)
	tireWidth := NumericRange{
		Min: float64(width),
		Max: float64(width),
	}
	return &Suggestion{
		TireWidthMM:                 width,
		RimSystem:                   system,
		ApplicationProfile:          classifyTireRimApplicationProfile(width),
		ResultKind:                  "engineering_recommended",
		AeroRuleApplies:             true,
		BestFitment:                 buildTireRimAeroBestFitmentWithinNormalRange(system, selectedInnerWidthMM, width, selectedRimTargetTireWidthMM),
		RimWidthRanges:              buildSingleRimWidthRangeForExistingAeroRecommendation(selectedInnerWidths),
		NormalFitmentRimWidthRanges: normalInnerWidthRanges,
		RimOuterWidthEstimates:      buildTireRimEngineeringRimOuterWidthEstimates(selectedInnerWidths),
		DerivedMetrics: DerivedMetrics{
			Status:          "engineering",
			TireWidthMM:     tireWidth,
			AeroRuleApplies: true,
			AeroTargetOuterWidthMM: &NumericRange{
				Min: roundTireRimWidthReferenceMetric(tireWidth.Min/tireRimEngineeringTireToRimWidthRatio, 1),
				Max: roundTireRimWidthReferenceMetric(tireWidth.Max/tireRimEngineeringTireToRimWidthRatio, 1),
			},
		},
		ModelVersion:           TireRimEngineeringReferenceModelVersion,
		KnowledgeAsOf:          TireRimEngineeringReferenceKnowledgeAsOf,
		Limitations:            cloneTireRimWidthReferenceStrings(tireRimEngineeringLimitations),
		StandardReference:      tireRimStandardReferenceContext,
		NormalFitmentReference: iso5775Part12023NormalFitmentTableReference,
	}, nil
}

// resolveNormalTireRimFitmentForKnownTireWidth returns every contiguous inner
// width range whose normal engineering envelope contains the entered tire.
// The 105% aero rule is deliberately not used for widths above 30C.
func resolveNormalTireRimFitmentForKnownTireWidth(width int, system RimSystem, profile TireRimApplicationProfile, profileSelection string) (*Suggestion, error) {
	normalInnerWidthRanges := findInnerWidthRangesForTireWidth(system, width, false, profile)
	if len(normalInnerWidthRanges) == 0 {
		return nil, fmt.Errorf("%w: %d mm has no normal fitment range for %s", ErrNoPublishedBracket, width, system)
	}

	return &Suggestion{
		TireWidthMM:                 width,
		RimSystem:                   system,
		ApplicationProfile:          profile,
		ApplicationProfileSelection: profileSelection,
		ResultKind:                  "engineering_normal_fitment",
		AeroRuleApplies:             false,
		BestFitment:                 selectBestNormalRimInnerWidthForKnownTireWidth(system, width, profile),
		RimWidthRanges:              normalInnerWidthRanges,
		NormalFitmentRimWidthRanges: normalInnerWidthRanges,
		RecommendedRimWidthRanges:   findInnerWidthRangesForTireWidth(system, width, true, profile),
		DerivedMetrics: DerivedMetrics{
			Status:          "engineering",
			TireWidthMM:     NumericRange{Min: float64(width), Max: float64(width)},
			AeroRuleApplies: false,
		},
		ModelVersion:           TireRimEngineeringReferenceModelVersion,
		KnowledgeAsOf:          TireRimEngineeringReferenceKnowledgeAsOf,
		Limitations:            cloneTireRimWidthReferenceStrings(tireRimEngineeringLimitations),
		StandardReference:      tireRimStandardReferenceContext,
		NormalFitmentReference: iso5775Part12023NormalFitmentTableReference,
	}, nil
}

func buildSingleRimWidthRangeForExistingAeroRecommendation(innerWidthsMM []int) []WidthRange {
	if len(innerWidthsMM) == 0 {
		return []WidthRange{}
	}
	return []WidthRange{{Min: innerWidthsMM[0], Max: innerWidthsMM[0]}}
}

// selectBestNormalRimInnerWidthForKnownTireWidth considers only pairings
// inside both normal and recommended windows. Eligible reviewed windows take
// priority over legacy fallback windows, then center distance is minimized.
// An exact tie selects the wider inner width as a stable project convention,
// not as a claim that the wider option has measured physical superiority.
func selectBestNormalRimInnerWidthForKnownTireWidth(system RimSystem, tireWidthMM int, profile TireRimApplicationProfile) *TireRimBestFitment {
	if tireWidthMM <= TireRimEngineeringAeroRuleMaximumTireWidthMM {
		return nil
	}
	var bestFitment *TireRimBestFitment
	bestDifferenceMM := math.Inf(1)
	bestUsesReviewedWindow := false
	for _, innerWidthMM := range tireRimEngineeringSupportedInnerWidthsMM {
		rule := tireRimEngineeringRules[system][profile][innerWidthMM]
		if rule.DefaultVerdict == TireRimEngineeringVerdictReference ||
			tireWidthMM < rule.AllowedTireWidthRange.Min || tireWidthMM > rule.AllowedTireWidthRange.Max ||
			tireWidthMM < rule.RecommendedTireWidthRange.Min || tireWidthMM > rule.RecommendedTireWidthRange.Max {
			continue
		}
		targetTireWidthMM := recommendedTireWidthWindowCenterForRule(rule)
		differenceMM := math.Abs(float64(tireWidthMM) - targetTireWidthMM)
		if bestFitment != nil && bestUsesReviewedWindow && !rule.RecommendedWindowUsesReviewedGuidance {
			continue
		}
		if bestFitment != nil && rule.RecommendedWindowUsesReviewedGuidance == bestUsesReviewedWindow && differenceMM > bestDifferenceMM {
			continue
		}
		bestDifferenceMM = differenceMM
		bestUsesReviewedWindow = rule.RecommendedWindowUsesReviewedGuidance
		reasonCode := "closest_recommended_tire_window_center"
		if bestUsesReviewedWindow {
			reasonCode = "closest_reviewed_recommended_tire_window_center"
		}
		bestFitment = &TireRimBestFitment{
			RimInnerWidthMM:             innerWidthMM,
			TireWidthMM:                 tireWidthMM,
			ApplicationProfile:          profile,
			SelectionMethod:             "recommended_window_center",
			ReasonCode:                  reasonCode,
			ModelStatus:                 "pending_profile_calibration",
			TargetTireWidthMM:           targetTireWidthMM,
			TargetTireWidthDifferenceMM: roundTireRimWidthReferenceMetric(differenceMM, 2),
			FitmentRuleReference:        cloneTireRimFitmentRuleReference(rule.FitmentRuleReference),
		}
	}
	return bestFitment
}

// selectBestNormalTireWidthForKnownRimInnerWidth chooses the nearest integer
// to the existing recommended-window center, within the normal window and
// the >30 mm path. Half-millimetre ties round upward. No outer rim width,
// wall-thickness assumption, or 105% ratio enters this selection.
func selectBestNormalTireWidthForKnownRimInnerWidth(system RimSystem, innerWidthMM int, profile TireRimApplicationProfile) *TireRimBestFitment {
	if innerWidthMM < TireRimEngineeringMinInnerWidthMM || innerWidthMM > TireRimEngineeringMaxInnerWidthMM {
		return nil
	}
	baseRule := tireRimEngineeringRules[system][profile][innerWidthMM]
	if baseRule.DefaultVerdict == TireRimEngineeringVerdictReference {
		return nil
	}
	minimumTireWidthMM := max(baseRule.AllowedTireWidthRange.Min, baseRule.RecommendedTireWidthRange.Min, TireRimEngineeringAeroRuleMaximumTireWidthMM+1)
	maximumTireWidthMM := min(baseRule.AllowedTireWidthRange.Max, baseRule.RecommendedTireWidthRange.Max, MaxInputTireWidthMM)
	if minimumTireWidthMM > maximumTireWidthMM {
		return nil
	}
	targetTireWidthMM := recommendedTireWidthWindowCenterForRule(baseRule)
	selectedTireWidthMM := max(minimumTireWidthMM, min(maximumTireWidthMM, int(math.Round(targetTireWidthMM))))
	return &TireRimBestFitment{
		RimInnerWidthMM:             innerWidthMM,
		TireWidthMM:                 selectedTireWidthMM,
		ApplicationProfile:          profile,
		SelectionMethod:             "recommended_window_center",
		ReasonCode:                  "nearest_integer_recommended_tire_window_center",
		ModelStatus:                 "pending_profile_calibration",
		TargetTireWidthMM:           targetTireWidthMM,
		TargetTireWidthDifferenceMM: roundTireRimWidthReferenceMetric(math.Abs(float64(selectedTireWidthMM)-targetTireWidthMM), 2),
		FitmentRuleReference:        cloneTireRimFitmentRuleReference(baseRule.FitmentRuleReference),
	}
}

// buildTireRimAeroBestFitmentWithinNormalRange preserves the project's
// existing geometric target, but never labels a blocked or transitional
// pairing as a new best_fitment. Legacy geometric fields remain unchanged.
func buildTireRimAeroBestFitmentWithinNormalRange(system RimSystem, innerWidthMM, tireWidthMM int, targetTireWidthMM float64) *TireRimBestFitment {
	rule := tireRimEngineeringRules[system][TireRimApplicationProfileRoadAero][innerWidthMM]
	if tireWidthMM > TireRimEngineeringAeroRuleMaximumTireWidthMM ||
		rule.DefaultVerdict == TireRimEngineeringVerdictReference ||
		tireWidthMM < rule.AllowedTireWidthRange.Min || tireWidthMM > rule.AllowedTireWidthRange.Max {
		return nil
	}
	return &TireRimBestFitment{
		RimInnerWidthMM:             innerWidthMM,
		TireWidthMM:                 tireWidthMM,
		ApplicationProfile:          TireRimApplicationProfileRoadAero,
		SelectionMethod:             "project_aero_geometric_target",
		ReasonCode:                  "closest_integer_project_aero_target",
		ModelStatus:                 "project_geometric_reference",
		TargetTireWidthMM:           roundTireRimWidthReferenceMetric(targetTireWidthMM, 1),
		TargetTireWidthDifferenceMM: roundTireRimWidthReferenceMetric(math.Abs(float64(tireWidthMM)-targetTireWidthMM), 1),
		FitmentRuleReference:        cloneTireRimFitmentRuleReference(rule.FitmentRuleReference),
	}
}

func buildTireRimEngineeringRimOuterWidthEstimates(innerWidthsMM []int) []RimOuterWidthEstimate {
	estimates := make([]RimOuterWidthEstimate, 0, len(innerWidthsMM))
	for _, innerWidthMM := range innerWidthsMM {
		outerWidthMM := float64(innerWidthMM) + 2*TireRimEngineeringRimWallThicknessEstimatePerSideMM
		estimates = append(estimates, RimOuterWidthEstimate{
			RimInnerWidthMM:                innerWidthMM,
			WallThicknessPerSideEstimateMM: TireRimEngineeringRimWallThicknessEstimatePerSideMM,
			RimOuterWidthEstimateMM:        roundTireRimWidthReferenceMetric(outerWidthMM, 1),
		})
	}
	return estimates
}

func BuildTireRimWidthGuidanceForTireWidth(width int, systems []RimSystem) []Suggestion {
	guidance := make([]Suggestion, 0, len(systems))
	seenSystems := make(map[RimSystem]struct{}, len(systems))
	for _, system := range systems {
		if _, exists := seenSystems[system]; exists {
			continue
		}
		seenSystems[system] = struct{}{}
		result, err := ResolveTireRimWidthReference(width, system)
		if err == nil && result != nil {
			guidance = append(guidance, *result)
		}
	}
	return guidance
}

func cloneTireRimWidthRanges(values []WidthRange) []WidthRange {
	if len(values) == 0 {
		return []WidthRange{}
	}
	result := make([]WidthRange, len(values))
	copy(result, values)
	return result
}

// findInnerWidthRangesForTireWidth returns contiguous rim inner-width ranges
// whose allowed or recommended engineering envelope contains the tire width.
// It is the normal-fitment path for tires above the 30C aero-rule boundary.
func findInnerWidthRangesForTireWidth(system RimSystem, tireWidthMM int, recommendedOnly bool, profiles ...TireRimApplicationProfile) []WidthRange {
	if !recommendedOnly {
		return findISO5775Part12023NormalRimInnerWidthRangesForTireWidth(tireWidthMM)
	}
	profile := classifyTireRimApplicationProfile(tireWidthMM)
	if len(profiles) > 0 {
		profile = profiles[0]
	}
	ranges := make([]WidthRange, 0)
	rangeStart := 0
	for innerWidthMM := TireRimEngineeringMinInnerWidthMM; innerWidthMM <= TireRimEngineeringMaxInnerWidthMM; innerWidthMM++ {
		rule := tireRimEngineeringRules[system][profile][innerWidthMM]
		fitmentRange := rule.AllowedTireWidthRange
		if recommendedOnly {
			fitmentRange = rule.RecommendedTireWidthRange
		}
		isInRange := tireWidthMM >= fitmentRange.Min && tireWidthMM <= fitmentRange.Max
		if isInRange && rangeStart == 0 {
			rangeStart = innerWidthMM
		}
		if !isInRange && rangeStart != 0 {
			ranges = append(ranges, WidthRange{Min: rangeStart, Max: innerWidthMM - 1})
			rangeStart = 0
		}
	}
	if rangeStart != 0 {
		ranges = append(ranges, WidthRange{Min: rangeStart, Max: TireRimEngineeringMaxInnerWidthMM})
	}
	return ranges
}

func cloneTireRimWidthReferenceStrings(values []string) []string {
	return append([]string(nil), values...)
}

func roundTireRimWidthReferenceMetric(value float64, places int) float64 {
	factor := math.Pow10(places)
	return math.Round(value*factor) / factor
}

func NormalizeRimSystem(value string) RimSystem {
	return RimSystem(strings.ToLower(strings.TrimSpace(value)))
}

// TireRimEngineeringReferenceModelVersion identifies the engineering model
// used by the interactive tire/rim calculator. The guide's DT Swiss comparison
// table is static front-end reference data and is never used by this model.
const (
	TireRimEngineeringReferenceModelVersion  = "tire-rim-engineering-reference-v11"
	TireRimEngineeringReferenceKnowledgeAsOf = "2026-10-11"
	// TireRimEngineeringMinInnerWidthMM and TireRimEngineeringMaxInnerWidthMM
	// define the integer inner-width envelope exposed by the calculator. The
	// upper bound is the current project envelope, not a market maximum.
	TireRimEngineeringMinInnerWidthMM = 19
	TireRimEngineeringMaxInnerWidthMM = 100
)

// TireRimEngineeringRimWallThicknessEstimatePerSideMM is the approximate
// outer-width assumption used for display and the project's narrow-tire
// geometric target. Actual outer width must come from the rim manufacturer;
// this assumption does not enter the wide-tire recommended-window selection.
const TireRimEngineeringRimWallThicknessEstimatePerSideMM = 3.0

var ErrInvalidTireRimEngineeringInnerWidth = errors.New("invalid tire/rim engineering inner width")

var tireRimEngineeringLimitations = []string{
	"This is a versioned engineering reference model, not certification for a specific tire, rim, wheelset, or frame.",
	"Normal width ranges directly apply ISO 5775-1:2023 clause 4.4 Table 7, transcribed from the user-provided width matrix image. ISO 5775-2:2021 provides rim terminology context; no ETRTO manual table is applied by this integration.",
	"Tire width varies by casing construction, pressure, and rim geometry; verify the exact tire and rim manufacturer's limits before installation.",
	"The normal width table accepts integer rim inner widths from 13 mm through 100 mm and tire widths from 18 mm through 132 mm. Existing project recommendation windows cover 19–100 mm rims separately; a standard-covered width can have no best recommendation.",
	"Normal ranges are looked up without interpolation and are independent of riding purpose or manufacturer optimization windows. The shared width matrix for hooked and hookless inputs does not establish exact-model hookless approval.",
	"Estimated rim outer width adds about 3 mm of wall thickness per side for display and the narrow-tire geometric target; actual rim profiles and maximum-width locations vary by model.",
	"The selected tire-width input is used directly as the tire width; the model does not add an inner-width-dependent correction.",
	"The 105% value means tire width = target rim outer width × 1.05 for tire widths up to and including 30C; it is an aerodynamic shape target, not a model-specific result, pressure limit, or bead-retention safety boundary.",
	"For a known rim inner width, the reverse lookup uses the 105% target only when the derived tire width is within the inclusive 30C scope. Wider rims return their normal and preferred fitment ranges without applying the aero rule.",
	"Wide-tire best_fitment selects one pairing from purpose-specific project windows by center distance. Reviewed model guidance is identified separately from unreviewed legacy windows; full calibration remains pending and pressure, rider load, casing stiffness, and impact absorption are not simulated.",
	"Non-Fat known-tire selection first considers eligible reviewed-purpose windows; legacy project windows are fallback candidates only when no reviewed candidate fits. This evidence priority is a project convention, not a measured physical score. Fat/snow does not use unreviewed fallback windows.",
	"Hooked MTB project windows use WTB optimal-performance cells, not its compatible-only cells. Their enclosing tire-width bands bridge omitted 53 mm and 63 mm chart rows as project estimates, not official WTB records.",
	"Normal fitment ranges describe the ISO nominal-width pairing relationship, not a guarantee against tire detachment. Exact bead-seat type and product approval remain separate conditions.",
	"This integration preserves the existing best-fitment algorithm and values. Some project recommendations fall outside the newly applied ISO normal ranges; reconciling those candidates is a separate pending step.",
	"A null recommended window or best_fitment means no recommendation is available for the selected purpose; it does not independently establish that the entered pairing is incompatible.",
	"Fat/snow recommendations require reviewed tire-size suggestions, not compatibility-only ranges or legacy window midpoints. The current reviewed default is scoped to 85 mm hookless inner width using ENVE M685's 3.8–4.5 inch suggestion; it is a project default, not approval for every 85 mm rim or a flotation simulation. Other Fat/snow inner widths retain reference envelopes without an optimum.",
	"HED compatibility ranges and pressure instructions remain product-specific evidence and do not define the generic Fat/snow envelope. Fat/snow results, and any tire wider than 85 mm, return null maximum_pressure because width and rim system alone cannot establish an exact product pressure limit.",
	"The 132 mm tire-width and 100 mm inner-width limits are current normal-table coverage, not maximum market sizes. Surly's dedicated 6.2 inch Molenda/Clown Royal system is outside this coverage and is never clamped to 132 mm or mapped from its unspecified 100 mm rim width to an inner-width rule.",
	"Pressure, tubeless approval, bead-seat compatibility, and frame clearance must follow the strictest exact product documentation.",
}

type TireRimEngineeringVerdict string

const (
	TireRimEngineeringVerdictCritical    TireRimEngineeringVerdict = "critical"
	TireRimEngineeringVerdictRecommended TireRimEngineeringVerdict = "recommended"
	TireRimEngineeringVerdictReference   TireRimEngineeringVerdict = "reference"
)

type TireRimEngineeringVerdictReason string

const (
	TireRimEngineeringReasonBelowMinimum     TireRimEngineeringVerdictReason = "below_minimum"
	TireRimEngineeringReasonAboveMaximum     TireRimEngineeringVerdictReason = "above_maximum"
	TireRimEngineeringReasonRecommended      TireRimEngineeringVerdictReason = "recommended_window"
	TireRimEngineeringReasonReference        TireRimEngineeringVerdictReason = "reference_window"
	TireRimEngineeringReasonNoRecommendation TireRimEngineeringVerdictReason = "no_recommended_window"
	TireRimEngineeringReasonTransitionalRim  TireRimEngineeringVerdictReason = "transitional_rim"
)

type TireRimEngineeringPressureLimit struct {
	Bar float64 `json:"bar"`
	PSI float64 `json:"psi"`
}

type TireRimEngineeringTubeRecommendation struct {
	Band       string `json:"band"`
	WidthMinMM int    `json:"width_min_mm"`
	WidthMaxMM int    `json:"width_max_mm"`
}

type TireRimEngineeringReferenceCalculation struct {
	ModelVersion                string                          `json:"model_version"`
	Verdict                     TireRimEngineeringVerdict       `json:"verdict"`
	VerdictReason               TireRimEngineeringVerdictReason `json:"verdict_reason"`
	RimSystem                   RimSystem                       `json:"rim_system"`
	ApplicationProfile          TireRimApplicationProfile       `json:"application_profile"`
	ApplicationProfileSelection string                          `json:"application_profile_selection"`
	RimInnerWidthMM             int                             `json:"rim_inner_width_mm"`
	TireWidthMM                 int                             `json:"tire_width_mm"`
	AllowedTireWidthRange       WidthRange                      `json:"allowed_tire_width_range"`
	RecommendedTireWidthRange   *WidthRange                     `json:"recommended_tire_width_range"`
	AeroRuleApplies             bool                            `json:"aero_rule_applies"`
	// AeroTargetOuterWidthMM is derived as tire width ÷ 1.05 only when the
	// entered tire width is within the inclusive <=30C aero-rule scope.
	AeroTargetOuterWidthMM *float64 `json:"aero_target_outer_width_mm,omitempty"`
	// AeroTargetTireWidthMM is derived as the estimated rim outer width × 1.05
	// only when the entered tire width is within the inclusive <=30C scope.
	AeroTargetTireWidthMM  *float64                             `json:"aero_target_tire_width_mm,omitempty"`
	MaximumPressure        *TireRimEngineeringPressureLimit     `json:"maximum_pressure"`
	TubeRecommendation     TireRimEngineeringTubeRecommendation `json:"tube_recommendation"`
	StandardReference      TireRimStandardReferenceContext      `json:"standard_reference"`
	NormalFitmentReference TireRimNormalFitmentTableReference   `json:"normal_fitment_reference"`
	FitmentRuleReference   TireRimFitmentRuleReference          `json:"fitment_rule_reference"`
}

type tireRimEngineeringRule struct {
	AllowedTireWidthRange                 WidthRange
	RecommendedTireWidthRange             WidthRange
	DefaultVerdict                        TireRimEngineeringVerdict
	RecommendedWindowCenterMM             *float64
	RecommendedWindowUsesReviewedGuidance bool
	FitmentRuleReference                  TireRimFitmentRuleReference
	ProfileGuidance                       map[TireRimApplicationProfile]tireRimReviewedManufacturerWidthGuidance
}

// tireRimReviewedManufacturerWidthGuidance records a model's suggested width
// band, never an ISO approval or a universal tire/rim retention boundary.
type tireRimReviewedManufacturerWidthGuidance struct {
	MinimumTireWidthInches float64
	MaximumTireWidthInches float64
	SourceURLs             []string
}

// tireRimEngineeringRuleAnchors contains the existing project model width bands
// that define the engineering envelope. Every integer width between the
// anchors is generated by interpolation so the API can accept a real 1 mm
// rim-width increment instead of a short list of presets.
var tireRimEngineeringRuleAnchors = map[RimSystem]map[int]tireRimEngineeringRule{
	RimSystemHookless: {
		19: {AllowedTireWidthRange: WidthRange{Min: 28, Max: 32}, RecommendedTireWidthRange: WidthRange{Min: 28, Max: 28}, DefaultVerdict: TireRimEngineeringVerdictReference},
		21: {AllowedTireWidthRange: WidthRange{Min: 28, Max: 32}, RecommendedTireWidthRange: WidthRange{Min: 28, Max: 30}},
		23: {AllowedTireWidthRange: WidthRange{Min: 28, Max: 32}, RecommendedTireWidthRange: WidthRange{Min: 28, Max: 32}},
		25: {
			AllowedTireWidthRange: WidthRange{Min: 29, Max: 34}, RecommendedTireWidthRange: WidthRange{Min: 30, Max: 32},
			ProfileGuidance: map[TireRimApplicationProfile]tireRimReviewedManufacturerWidthGuidance{
				TireRimApplicationProfileMTBXC: {2.1, 2.4, []string{"https://enve.com/pages/tire-compatibility"}},
			},
		},
		28: {
			AllowedTireWidthRange: WidthRange{Min: 38, Max: 50}, RecommendedTireWidthRange: WidthRange{Min: 40, Max: 45},
			ProfileGuidance: map[TireRimApplicationProfile]tireRimReviewedManufacturerWidthGuidance{
				TireRimApplicationProfileMTBXC: {2.2, 2.5, []string{"https://stans.com/products/arch-mk4-rim"}},
			},
		},
		30: {
			AllowedTireWidthRange: WidthRange{Min: 42, Max: 60}, RecommendedTireWidthRange: WidthRange{Min: 48, Max: 57},
			ProfileGuidance: map[TireRimApplicationProfile]tireRimReviewedManufacturerWidthGuidance{
				TireRimApplicationProfileMTBXC: {2.3, 2.6, []string{"https://enve.com/pages/tire-compatibility", "https://stans.com/products/flow-mk4-rim"}},
			},
		},
		35: {
			AllowedTireWidthRange: WidthRange{Min: 50, Max: 72}, RecommendedTireWidthRange: WidthRange{Min: 56, Max: 66},
			ProfileGuidance: map[TireRimApplicationProfile]tireRimReviewedManufacturerWidthGuidance{
				TireRimApplicationProfileMTBXC:   {2.5, 2.8, []string{"https://enve.com/pages/tire-compatibility"}},
				TireRimApplicationProfileMTBPlus: {2.6, 3.0, []string{"https://stans.com/products/baron-mk3-rim"}},
			},
		},
		40: {
			AllowedTireWidthRange: WidthRange{Min: 58, Max: 84}, RecommendedTireWidthRange: WidthRange{Min: 66, Max: 76},
			ProfileGuidance: map[TireRimApplicationProfile]tireRimReviewedManufacturerWidthGuidance{
				TireRimApplicationProfileMTBPlus: {2.8, 3.2, []string{"https://enve.com/pages/tire-compatibility"}},
			},
		},
		45:  {AllowedTireWidthRange: WidthRange{Min: 66, Max: 96}, RecommendedTireWidthRange: WidthRange{Min: 74, Max: 86}},
		50:  {AllowedTireWidthRange: WidthRange{Min: 74, Max: 108}, RecommendedTireWidthRange: WidthRange{Min: 82, Max: 96}},
		55:  {AllowedTireWidthRange: WidthRange{Min: 82, Max: 118}, RecommendedTireWidthRange: WidthRange{Min: 90, Max: 104}},
		60:  {AllowedTireWidthRange: WidthRange{Min: 90, Max: 127}, RecommendedTireWidthRange: WidthRange{Min: 98, Max: 112}},
		65:  {AllowedTireWidthRange: WidthRange{Min: 96, Max: 127}, RecommendedTireWidthRange: WidthRange{Min: 104, Max: 120}},
		70:  {AllowedTireWidthRange: WidthRange{Min: 100, Max: 127}, RecommendedTireWidthRange: WidthRange{Min: 110, Max: 127}},
		76:  {AllowedTireWidthRange: WidthRange{Min: 102, Max: 127}, RecommendedTireWidthRange: WidthRange{Min: 102, Max: 127}},
		80:  {AllowedTireWidthRange: WidthRange{Min: 108, Max: 127}, RecommendedTireWidthRange: WidthRange{Min: 110, Max: 127}},
		90:  {AllowedTireWidthRange: WidthRange{Min: 118, Max: 127}, RecommendedTireWidthRange: WidthRange{Min: 120, Max: 127}},
		100: {AllowedTireWidthRange: WidthRange{Min: 127, Max: 127}, RecommendedTireWidthRange: WidthRange{Min: 127, Max: 127}},
	},
	RimSystemHooked: {
		19:  {AllowedTireWidthRange: WidthRange{Min: 23, Max: 35}, RecommendedTireWidthRange: WidthRange{Min: 23, Max: 28}},
		21:  {AllowedTireWidthRange: WidthRange{Min: 23, Max: 40}, RecommendedTireWidthRange: WidthRange{Min: 25, Max: 30}},
		23:  {AllowedTireWidthRange: WidthRange{Min: 23, Max: 45}, RecommendedTireWidthRange: WidthRange{Min: 28, Max: 32}},
		25:  {AllowedTireWidthRange: WidthRange{Min: 23, Max: 45}, RecommendedTireWidthRange: WidthRange{Min: 26, Max: 35}},
		28:  {AllowedTireWidthRange: WidthRange{Min: 29, Max: 60}, RecommendedTireWidthRange: WidthRange{Min: 36, Max: 60}},
		30:  {AllowedTireWidthRange: WidthRange{Min: 33, Max: 60}, RecommendedTireWidthRange: WidthRange{Min: 46, Max: 60}},
		35:  {AllowedTireWidthRange: WidthRange{Min: 40, Max: 72}, RecommendedTireWidthRange: WidthRange{Min: 52, Max: 66}},
		40:  {AllowedTireWidthRange: WidthRange{Min: 50, Max: 84}, RecommendedTireWidthRange: WidthRange{Min: 60, Max: 76}},
		45:  {AllowedTireWidthRange: WidthRange{Min: 60, Max: 96}, RecommendedTireWidthRange: WidthRange{Min: 70, Max: 86}},
		50:  {AllowedTireWidthRange: WidthRange{Min: 70, Max: 108}, RecommendedTireWidthRange: WidthRange{Min: 80, Max: 96}},
		55:  {AllowedTireWidthRange: WidthRange{Min: 80, Max: 118}, RecommendedTireWidthRange: WidthRange{Min: 90, Max: 104}},
		60:  {AllowedTireWidthRange: WidthRange{Min: 90, Max: 127}, RecommendedTireWidthRange: WidthRange{Min: 100, Max: 112}},
		65:  {AllowedTireWidthRange: WidthRange{Min: 96, Max: 127}, RecommendedTireWidthRange: WidthRange{Min: 106, Max: 120}},
		70:  {AllowedTireWidthRange: WidthRange{Min: 100, Max: 127}, RecommendedTireWidthRange: WidthRange{Min: 112, Max: 127}},
		76:  {AllowedTireWidthRange: WidthRange{Min: 102, Max: 127}, RecommendedTireWidthRange: WidthRange{Min: 102, Max: 127}},
		80:  {AllowedTireWidthRange: WidthRange{Min: 108, Max: 127}, RecommendedTireWidthRange: WidthRange{Min: 110, Max: 127}},
		90:  {AllowedTireWidthRange: WidthRange{Min: 118, Max: 127}, RecommendedTireWidthRange: WidthRange{Min: 120, Max: 127}},
		100: {AllowedTireWidthRange: WidthRange{Min: 127, Max: 127}, RecommendedTireWidthRange: WidthRange{Min: 127, Max: 127}},
	},
}

var tireRimEngineeringSupportedInnerWidthsMM = buildTireRimEngineeringSupportedInnerWidths()

// reviewedWTBHookedMountainOptimalPerformanceBands records only solid optimal
// cells from the official chart. Compatible dots and its ISO-colored backdrop
// are not optimization evidence. The chart covers grouped integer rim widths.
var reviewedWTBHookedMountainOptimalPerformanceBands = []struct {
	TireWidthRangeMM     WidthRange
	RimInnerWidthRangeMM WidthRange
}{
	{WidthRange{50, 52}, WidthRange{25, 30}},
	{WidthRange{54, 54}, WidthRange{25, 30}},
	{WidthRange{55, 56}, WidthRange{25, 30}},
	{WidthRange{57, 57}, WidthRange{26, 30}},
	{WidthRange{58, 59}, WidthRange{26, 32}},
	{WidthRange{60, 62}, WidthRange{28, 33}},
	{WidthRange{64, 64}, WidthRange{28, 35}},
	{WidthRange{65, 65}, WidthRange{30, 35}},
	{WidthRange{66, 67}, WidthRange{32, 40}},
	{WidthRange{68, 71}, WidthRange{35, 43}},
}

var tireRimEngineeringRules = buildTireRimEngineeringRulesFromAnchors()

func buildTireRimEngineeringSupportedInnerWidths() []int {
	widths := make([]int, 0, TireRimEngineeringMaxInnerWidthMM-TireRimEngineeringMinInnerWidthMM+1)
	for innerWidthMM := TireRimEngineeringMinInnerWidthMM; innerWidthMM <= TireRimEngineeringMaxInnerWidthMM; innerWidthMM++ {
		widths = append(widths, innerWidthMM)
	}
	return widths
}

func buildTireRimEngineeringRulesFromAnchors() map[RimSystem]map[TireRimApplicationProfile]map[int]tireRimEngineeringRule {
	rulesBySystem := make(map[RimSystem]map[TireRimApplicationProfile]map[int]tireRimEngineeringRule, len(tireRimEngineeringRuleAnchors))
	for system, anchorRules := range tireRimEngineeringRuleAnchors {
		rulesBySystem[system] = make(map[TireRimApplicationProfile]map[int]tireRimEngineeringRule)
		for _, profile := range []TireRimApplicationProfile{TireRimApplicationProfileRoadAero, TireRimApplicationProfileGravelAllRoad, TireRimApplicationProfileMTBXC, TireRimApplicationProfileMTBPlus, TireRimApplicationProfileFatSnow} {
			profileAnchors := make(map[int]tireRimEngineeringRule, len(anchorRules))
			for innerWidthMM, anchorRule := range anchorRules {
				profileAnchors[innerWidthMM] = buildTireRimApplicationProfileRuleFromReviewedAnchor(anchorRule, profile)
			}
			rulesBySystem[system][profile] = make(map[int]tireRimEngineeringRule, len(tireRimEngineeringSupportedInnerWidthsMM))
			for _, innerWidthMM := range tireRimEngineeringSupportedInnerWidthsMM {
				var rule tireRimEngineeringRule
				if exactRule, exists := profileAnchors[innerWidthMM]; exists {
					rule = exactRule
				} else {
					rule = interpolateTireRimEngineeringRule(profileAnchors, innerWidthMM)
				}
				if system == RimSystemHooked && profile == TireRimApplicationProfileMTBXC {
					rule = applyReviewedWTBHookedMountainOptimalPerformanceBandsToRule(rule, innerWidthMM)
				}
				if profile == TireRimApplicationProfileFatSnow {
					rule.FitmentRuleReference.RecommendedRangeBasis = "no_reviewed_tire_size_suggestion_for_fat_snow"
					// One reviewed model provides a size suggestion at this exact
					// inner width. Compatibility-only records and neighboring
					// legacy widths do not supply additional optimization anchors.
					if system == RimSystemHookless && innerWidthMM == 85 {
						rule = applyReviewedManufacturerTireSizeSuggestionToProjectRule(rule, tireRimReviewedManufacturerWidthGuidance{
							MinimumTireWidthInches: 3.8,
							MaximumTireWidthInches: 4.5,
							SourceURLs:             []string{"https://enve.com/pages/tire-compatibility"},
						})
					}
				}
				rulesBySystem[system][profile][innerWidthMM] = rule
			}
		}
	}
	return rulesBySystem
}

// buildTireRimApplicationProfileRuleFromReviewedAnchor retains the legacy
// project envelope and explicitly includes reviewed model suggestions in it.
// It never turns the suggestion endpoints into universal safety limits. Only
// the corresponding purpose receives the reviewed recommended window.
func buildTireRimApplicationProfileRuleFromReviewedAnchor(anchor tireRimEngineeringRule, profile TireRimApplicationProfile) tireRimEngineeringRule {
	rule := anchor
	rule.ProfileGuidance = nil
	rule.FitmentRuleReference = TireRimFitmentRuleReference{
		NormalRangeBasis:       "legacy_project_engineering_envelope",
		RecommendedRangeBasis:  "legacy_project_window_assigned_to_application_profile",
		ManufacturerSourceURLs: []string{},
	}
	if profile == TireRimApplicationProfileFatSnow {
		// Legacy Fat centers include a 70→76 mm reversal and an artificial
		// 100/127 boundary optimum. Retain their normal reference envelopes,
		// but do not interpolate these unreviewed optimization windows.
		rule.RecommendedTireWidthRange = WidthRange{}
		rule.RecommendedWindowCenterMM = nil
		rule.RecommendedWindowUsesReviewedGuidance = false
		rule.FitmentRuleReference.RecommendedRangeBasis = "no_reviewed_tire_size_suggestion_for_fat_snow"
	} else if profile != TireRimApplicationProfileRoadAero {
		profileBand := defaultNominalTireWidthBandForApplicationProfile(profile)
		minimum := max(rule.RecommendedTireWidthRange.Min, profileBand.Min)
		maximum := min(rule.RecommendedTireWidthRange.Max, profileBand.Max)
		rule.RecommendedTireWidthRange = WidthRange{}
		if minimum <= maximum {
			rule.RecommendedTireWidthRange = WidthRange{Min: minimum, Max: maximum}
		} else {
			rule.FitmentRuleReference.RecommendedRangeBasis = "no_recommended_window_for_application_profile"
		}
	}
	if guidance, exists := anchor.ProfileGuidance[profile]; exists {
		rule = applyReviewedManufacturerTireSizeSuggestionToProjectRule(rule, guidance)
	}
	return rule
}

// applyReviewedManufacturerTireSizeSuggestionToProjectRule includes a model's
// suggested sizes in the existing project envelope. Only the recommendation
// uses the exact source midpoint; neither it nor the enclosing normal range
// is a universal compatibility approval. Any legacy normal interpolation
// remains identified separately from the source-based recommendation.
func applyReviewedManufacturerTireSizeSuggestionToProjectRule(rule tireRimEngineeringRule, guidance tireRimReviewedManufacturerWidthGuidance) tireRimEngineeringRule {
	minimumMM := guidance.MinimumTireWidthInches * 25.4
	maximumMM := guidance.MaximumTireWidthInches * 25.4
	rule.RecommendedTireWidthRange = WidthRange{Min: int(math.Round(minimumMM)), Max: int(math.Round(maximumMM))}
	rule.AllowedTireWidthRange.Min = min(rule.AllowedTireWidthRange.Min, rule.RecommendedTireWidthRange.Min)
	rule.AllowedTireWidthRange.Max = max(rule.AllowedTireWidthRange.Max, rule.RecommendedTireWidthRange.Max)
	centerMM := roundTireRimWidthReferenceMetric((minimumMM+maximumMM)/2, 2)
	rule.RecommendedWindowCenterMM = &centerMM
	rule.RecommendedWindowUsesReviewedGuidance = true
	rule.FitmentRuleReference.NormalRangeBasis = "project_envelope_including_reviewed_model_guidance"
	rule.FitmentRuleReference.RecommendedRangeBasis = "reviewed_manufacturer_guidance_project_window"
	rule.FitmentRuleReference.ManufacturerSourceURLs = cloneTireRimWidthReferenceStrings(guidance.SourceURLs)
	return rule
}

// applyReviewedWTBHookedMountainOptimalPerformanceBandsToRule calibrates the
// existing rule for the requested inner width. The enclosing project window
// bridges source row gaps; it is never a copied compatibility or safety band.
func applyReviewedWTBHookedMountainOptimalPerformanceBandsToRule(rule tireRimEngineeringRule, innerWidthMM int) tireRimEngineeringRule {
	var projectWindow WidthRange
	for _, sourceBand := range reviewedWTBHookedMountainOptimalPerformanceBands {
		if innerWidthMM < sourceBand.RimInnerWidthRangeMM.Min || innerWidthMM > sourceBand.RimInnerWidthRangeMM.Max {
			continue
		}
		if projectWindow == (WidthRange{}) {
			projectWindow = sourceBand.TireWidthRangeMM
		} else {
			projectWindow.Min = min(projectWindow.Min, sourceBand.TireWidthRangeMM.Min)
			projectWindow.Max = max(projectWindow.Max, sourceBand.TireWidthRangeMM.Max)
		}
	}
	rule.RecommendedTireWidthRange = projectWindow
	rule.RecommendedWindowCenterMM = nil
	rule.RecommendedWindowUsesReviewedGuidance = projectWindow != (WidthRange{})
	rule.FitmentRuleReference.ManufacturerSourceURLs = []string{"https://www.wtb.com/pages/tire-rim-fit-chart", "https://www.wtb.com/cdn/shop/files/Fit_Right_Chart.png?v=1787649652&width=2481"}
	if projectWindow == (WidthRange{}) {
		rule.FitmentRuleReference.RecommendedRangeBasis = "no_optimal_performance_window_in_reviewed_manufacturer_chart"
		return rule
	}
	rule.AllowedTireWidthRange.Min = min(rule.AllowedTireWidthRange.Min, projectWindow.Min)
	rule.AllowedTireWidthRange.Max = max(rule.AllowedTireWidthRange.Max, projectWindow.Max)
	rule.FitmentRuleReference.NormalRangeBasis = "project_envelope_including_reviewed_model_guidance"
	rule.FitmentRuleReference.RecommendedRangeBasis = "reviewed_manufacturer_optimal_bands_project_envelope"
	// The normal envelope may still interpolate old anchors. The recommended
	// window above is transposed from source cells at this exact inner width.
	return rule
}

// defaultNominalTireWidthBandForApplicationProfile only assigns unreviewed
// legacy recommendations to a default purpose. Reviewed model guidance can
// cross these heuristic bands; they are not tire compatibility boundaries.
func defaultNominalTireWidthBandForApplicationProfile(profile TireRimApplicationProfile) WidthRange {
	switch profile {
	case TireRimApplicationProfileRoadAero:
		return WidthRange{Min: MinInputTireWidthMM, Max: 30}
	case TireRimApplicationProfileGravelAllRoad:
		return WidthRange{Min: 31, Max: 50}
	case TireRimApplicationProfileMTBXC:
		return WidthRange{Min: 51, Max: 70}
	case TireRimApplicationProfileMTBPlus:
		return WidthRange{Min: 71, Max: 85}
	default:
		return WidthRange{Min: 86, Max: MaxInputTireWidthMM}
	}
}

func recommendedTireWidthWindowCenterForRule(rule tireRimEngineeringRule) float64 {
	if rule.RecommendedWindowCenterMM != nil {
		return *rule.RecommendedWindowCenterMM
	}
	return float64(rule.RecommendedTireWidthRange.Min+rule.RecommendedTireWidthRange.Max) / 2
}

func cloneTireRimFitmentRuleReference(reference TireRimFitmentRuleReference) TireRimFitmentRuleReference {
	manufacturerSources := make([]string, len(reference.ManufacturerSourceURLs))
	copy(manufacturerSources, reference.ManufacturerSourceURLs)
	reference.ManufacturerSourceURLs = manufacturerSources
	return reference
}

// nullableRecommendedTireWidthRange prevents an absent project recommendation
// from being serialized as a misleading zero-millimetre tire window.
func nullableRecommendedTireWidthRange(widthRange WidthRange) *WidthRange {
	if widthRange == (WidthRange{}) {
		return nil
	}
	return &widthRange
}

func mergeTireRimManufacturerSourceURLs(lowerSources, upperSources []string) []string {
	result := make([]string, 0, len(lowerSources)+len(upperSources))
	for _, source := range append(cloneTireRimWidthReferenceStrings(lowerSources), upperSources...) {
		found := false
		for _, existing := range result {
			if source == existing {
				found = true
				break
			}
		}
		if !found {
			result = append(result, source)
		}
	}
	return result
}

func interpolateTireRimEngineeringRule(anchorRules map[int]tireRimEngineeringRule, innerWidthMM int) tireRimEngineeringRule {
	lowerAnchorWidthMM, upperAnchorWidthMM := 0, 0
	for anchorWidthMM := TireRimEngineeringMinInnerWidthMM; anchorWidthMM <= TireRimEngineeringMaxInnerWidthMM; anchorWidthMM++ {
		if _, exists := anchorRules[anchorWidthMM]; !exists {
			continue
		}
		if anchorWidthMM < innerWidthMM {
			lowerAnchorWidthMM = anchorWidthMM
			continue
		}
		upperAnchorWidthMM = anchorWidthMM
		break
	}
	if lowerAnchorWidthMM == 0 || upperAnchorWidthMM == 0 || lowerAnchorWidthMM == upperAnchorWidthMM {
		return anchorRules[lowerAnchorWidthMM]
	}

	lowerRule := anchorRules[lowerAnchorWidthMM]
	upperRule := anchorRules[upperAnchorWidthMM]
	ratio := float64(innerWidthMM-lowerAnchorWidthMM) / float64(upperAnchorWidthMM-lowerAnchorWidthMM)
	var recommendedRange WidthRange
	var recommendedCenterMM *float64
	if lowerRule.RecommendedTireWidthRange != (WidthRange{}) && upperRule.RecommendedTireWidthRange != (WidthRange{}) {
		recommendedRange = interpolateTireRimEngineeringWidthRange(lowerRule.RecommendedTireWidthRange, upperRule.RecommendedTireWidthRange, ratio)
		centerMM := roundTireRimWidthReferenceMetric(recommendedTireWidthWindowCenterForRule(lowerRule)+(recommendedTireWidthWindowCenterForRule(upperRule)-recommendedTireWidthWindowCenterForRule(lowerRule))*ratio, 2)
		recommendedCenterMM = &centerMM
	}
	manufacturerSources := mergeTireRimManufacturerSourceURLs(lowerRule.FitmentRuleReference.ManufacturerSourceURLs, upperRule.FitmentRuleReference.ManufacturerSourceURLs)
	reference := TireRimFitmentRuleReference{
		NormalRangeBasis:       "project_anchor_interpolation",
		RecommendedRangeBasis:  "project_profile_window_interpolation",
		ManufacturerSourceURLs: manufacturerSources,
		Interpolated:           true,
	}
	if recommendedRange == (WidthRange{}) {
		reference.RecommendedRangeBasis = "no_recommended_window_for_application_profile"
	}
	return tireRimEngineeringRule{
		AllowedTireWidthRange:                 interpolateTireRimEngineeringWidthRange(lowerRule.AllowedTireWidthRange, upperRule.AllowedTireWidthRange, ratio),
		RecommendedTireWidthRange:             recommendedRange,
		RecommendedWindowCenterMM:             recommendedCenterMM,
		RecommendedWindowUsesReviewedGuidance: lowerRule.RecommendedWindowUsesReviewedGuidance && upperRule.RecommendedWindowUsesReviewedGuidance,
		DefaultVerdict:                        TireRimEngineeringVerdictRecommended,
		FitmentRuleReference:                  reference,
	}
}

func interpolateTireRimEngineeringWidthRange(lowerRange, upperRange WidthRange, ratio float64) WidthRange {
	minimum := int(math.Round(float64(lowerRange.Min) + float64(upperRange.Min-lowerRange.Min)*ratio))
	maximum := int(math.Round(float64(lowerRange.Max) + float64(upperRange.Max-lowerRange.Max)*ratio))
	if minimum > maximum {
		minimum, maximum = maximum, minimum
	}
	return WidthRange{Min: minimum, Max: maximum}
}

func ValidateTireRimEngineeringInnerWidth(innerWidthMM int) error {
	if innerWidthMM >= TireRimStandardNormalFitmentMinimumInnerWidthMM && innerWidthMM <= TireRimEngineeringMaxInnerWidthMM {
		return nil
	}
	return fmt.Errorf(
		"%w: %d mm is outside the supported integer inner-width range [%d, %d]",
		ErrInvalidTireRimEngineeringInnerWidth,
		innerWidthMM,
		TireRimStandardNormalFitmentMinimumInnerWidthMM,
		TireRimEngineeringMaxInnerWidthMM,
	)
}

// CalculateTireRimEngineeringReference applies the versioned engineering
// model in the backend. The browser must render this result rather than
// reimplementing the verdict or the 105% projection locally.
func CalculateTireRimEngineeringReference(
	system RimSystem,
	innerWidthMM int,
	tireWidthMM int,
	profiles ...TireRimApplicationProfile,
) (*TireRimEngineeringReferenceCalculation, error) {
	if err := ValidateRimSystem(system); err != nil {
		return nil, err
	}
	if err := ValidateTireWidth(tireWidthMM); err != nil {
		return nil, err
	}
	if err := ValidateTireRimEngineeringInnerWidth(innerWidthMM); err != nil {
		return nil, err
	}
	profile, profileSelection, err := resolveTireRimApplicationProfileForKnownTireWidth(tireWidthMM, profiles)
	if err != nil {
		return nil, err
	}
	rule := tireRimEngineeringRules[system][profile][innerWidthMM]
	normalTireWidthRange := findISO5775Part12023NormalTireWidthRangeForRimInnerWidth(innerWidthMM)
	if normalTireWidthRange == (WidthRange{}) {
		return nil, fmt.Errorf("%w: %d mm is not defined for %s", ErrInvalidTireRimEngineeringInnerWidth, innerWidthMM, system)
	}

	verdict := TireRimEngineeringVerdictReference
	reason := TireRimEngineeringReasonReference
	if tireWidthMM < normalTireWidthRange.Min {
		verdict = TireRimEngineeringVerdictCritical
		reason = TireRimEngineeringReasonBelowMinimum
	} else if tireWidthMM > normalTireWidthRange.Max {
		verdict = TireRimEngineeringVerdictCritical
		reason = TireRimEngineeringReasonAboveMaximum
	} else if rule.DefaultVerdict == TireRimEngineeringVerdictReference {
		verdict = TireRimEngineeringVerdictReference
		reason = TireRimEngineeringReasonTransitionalRim
	} else if rule.RecommendedTireWidthRange == (WidthRange{}) {
		reason = TireRimEngineeringReasonNoRecommendation
	} else if tireWidthMM >= rule.RecommendedTireWidthRange.Min && tireWidthMM <= rule.RecommendedTireWidthRange.Max {
		verdict = TireRimEngineeringVerdictRecommended
		reason = TireRimEngineeringReasonRecommended
	}

	tireWidthValueMM := float64(tireWidthMM)
	rimOuterWidthValueMM := float64(innerWidthMM) + 2*TireRimEngineeringRimWallThicknessEstimatePerSideMM
	aeroRuleApplies := tireWidthMM <= TireRimEngineeringAeroRuleMaximumTireWidthMM
	var aeroTargetOuterWidthMM *float64
	var aeroTargetTireWidthMM *float64
	if aeroRuleApplies {
		aeroTargetOuterWidth := roundTireRimWidthReferenceMetric(tireWidthValueMM/tireRimEngineeringTireToRimWidthRatio, 1)
		aeroTargetTireWidth := roundTireRimWidthReferenceMetric(rimOuterWidthValueMM*tireRimEngineeringTireToRimWidthRatio, 1)
		aeroTargetOuterWidthMM = &aeroTargetOuterWidth
		aeroTargetTireWidthMM = &aeroTargetTireWidth
	}
	var maximumPressure *TireRimEngineeringPressureLimit
	if profile != TireRimApplicationProfileFatSnow && tireWidthMM <= 85 {
		pressureReference := calculateTireRimEngineeringMaximumPressure(system, tireWidthMM)
		maximumPressure = &pressureReference
	}
	return &TireRimEngineeringReferenceCalculation{
		ModelVersion:                TireRimEngineeringReferenceModelVersion,
		Verdict:                     verdict,
		VerdictReason:               reason,
		RimSystem:                   system,
		ApplicationProfile:          profile,
		ApplicationProfileSelection: profileSelection,
		RimInnerWidthMM:             innerWidthMM,
		TireWidthMM:                 tireWidthMM,
		AllowedTireWidthRange:       normalTireWidthRange,
		RecommendedTireWidthRange:   nullableRecommendedTireWidthRange(rule.RecommendedTireWidthRange),
		AeroRuleApplies:             aeroRuleApplies,
		AeroTargetOuterWidthMM:      aeroTargetOuterWidthMM,
		AeroTargetTireWidthMM:       aeroTargetTireWidthMM,
		MaximumPressure:             maximumPressure,
		TubeRecommendation:          calculateTireRimEngineeringTubeRecommendation(tireWidthMM),
		StandardReference:           tireRimStandardReferenceContext,
		FitmentRuleReference:        buildISO5775Part12023NormalFitmentRuleReference(),
		NormalFitmentReference:      iso5775Part12023NormalFitmentTableReference,
	}, nil
}

// ResolveTireRimWidthReferenceForRim returns an aero target only when that
// target is within the inclusive <=30C rule scope. Wider rims return their
// normal and preferred tire-width fitment ranges without applying 105%.
func ResolveTireRimWidthReferenceForRim(
	system RimSystem,
	innerWidthMM int,
	profiles ...TireRimApplicationProfile,
) (*RimToTireSuggestion, error) {
	if err := ValidateRimSystem(system); err != nil {
		return nil, err
	}
	if err := ValidateTireRimEngineeringInnerWidth(innerWidthMM); err != nil {
		return nil, err
	}

	requestedProfile, err := validateOptionalTireRimApplicationProfile(profiles)
	if err != nil {
		return nil, err
	}
	rimOuterWidthEstimate := buildTireRimEngineeringRimOuterWidthEstimates([]int{innerWidthMM})[0]
	var targetTireWidthMM float64
	if requestedProfile == TireRimApplicationProfileAutoByWidth || requestedProfile == TireRimApplicationProfileRoadAero {
		// Automatic reverse lookup checks only whether the existing narrow-tire
		// path applies. Explicit wide purposes never calculate an aero target.
		targetTireWidthMM = rimOuterWidthEstimate.RimOuterWidthEstimateMM * tireRimEngineeringTireToRimWidthRatio
	}
	profile := requestedProfile
	profileSelection := "explicit"
	if requestedProfile == TireRimApplicationProfileAutoByWidth {
		profileSelection = "automatic_by_rim_default_window"
		profile = TireRimApplicationProfileRoadAero
		if targetTireWidthMM > TireRimEngineeringAeroRuleMaximumTireWidthMM {
			legacyRule := tireRimEngineeringRules[system][TireRimApplicationProfileRoadAero][innerWidthMM]
			legacyCenterMM := (legacyRule.RecommendedTireWidthRange.Min + legacyRule.RecommendedTireWidthRange.Max) / 2
			profile = classifyTireRimApplicationProfile(max(31, legacyCenterMM))
		}
	}
	if profile == TireRimApplicationProfileRoadAero && targetTireWidthMM > TireRimEngineeringAeroRuleMaximumTireWidthMM {
		return nil, ErrTireRimApplicationProfileOutsideAeroScope
	}
	rule := tireRimEngineeringRules[system][profile][innerWidthMM]
	normalTireWidthRange := findISO5775Part12023NormalTireWidthRangeForRimInnerWidth(innerWidthMM)
	if normalTireWidthRange == (WidthRange{}) {
		return nil, fmt.Errorf("%w: %d mm is not defined for %s", ErrInvalidTireRimEngineeringInnerWidth, innerWidthMM, system)
	}

	topLevelProfile := requestedProfile
	if profile == TireRimApplicationProfileRoadAero {
		bestMatch := findClosestTireWidthForTargetTireWidth(
			system,
			innerWidthMM,
			targetTireWidthMM,
		)
		roundedTargetTireWidthMM := roundTireRimWidthReferenceMetric(targetTireWidthMM, 1)
		return &RimToTireSuggestion{
			RimInnerWidthMM:                  innerWidthMM,
			RimSystem:                        system,
			ApplicationProfile:               topLevelProfile,
			ResolvedApplicationProfile:       profile,
			ApplicationProfileSelection:      profileSelection,
			AeroRuleApplies:                  true,
			AeroTargetTireWidthMM:            &roundedTargetTireWidthMM,
			BestMatch:                        &bestMatch,
			BestFitment:                      buildTireRimAeroBestFitmentWithinNormalRange(system, innerWidthMM, bestMatch.TireWidthMM, targetTireWidthMM),
			NormalFitmentTireWidthRange:      normalTireWidthRange,
			RecommendedFitmentTireWidthRange: nullableRecommendedTireWidthRange(rule.RecommendedTireWidthRange),
			RimOuterWidthEstimate:            rimOuterWidthEstimate,
			ModelVersion:                     TireRimEngineeringReferenceModelVersion,
			KnowledgeAsOf:                    TireRimEngineeringReferenceKnowledgeAsOf,
			Limitations:                      cloneTireRimWidthReferenceStrings(tireRimEngineeringLimitations),
			StandardReference:                tireRimStandardReferenceContext,
			FitmentRuleReference:             buildISO5775Part12023NormalFitmentRuleReference(),
			NormalFitmentReference:           iso5775Part12023NormalFitmentTableReference,
		}, nil
	}

	return &RimToTireSuggestion{
		RimInnerWidthMM:                  innerWidthMM,
		RimSystem:                        system,
		ApplicationProfile:               topLevelProfile,
		ResolvedApplicationProfile:       profile,
		ApplicationProfileSelection:      profileSelection,
		AeroRuleApplies:                  false,
		BestFitment:                      selectBestNormalTireWidthForKnownRimInnerWidth(system, innerWidthMM, profile),
		NormalFitmentTireWidthRange:      normalTireWidthRange,
		RecommendedFitmentTireWidthRange: nullableRecommendedTireWidthRange(rule.RecommendedTireWidthRange),
		RimOuterWidthEstimate:            rimOuterWidthEstimate,
		ModelVersion:                     TireRimEngineeringReferenceModelVersion,
		KnowledgeAsOf:                    TireRimEngineeringReferenceKnowledgeAsOf,
		Limitations:                      cloneTireRimWidthReferenceStrings(tireRimEngineeringLimitations),
		StandardReference:                tireRimStandardReferenceContext,
		FitmentRuleReference:             buildISO5775Part12023NormalFitmentRuleReference(),
		NormalFitmentReference:           iso5775Part12023NormalFitmentTableReference,
	}, nil
}

func findClosestTireWidthForTargetTireWidth(
	system RimSystem,
	innerWidthMM int,
	targetTireWidthMM float64,
) TireWidthMatch {
	bestMatch := TireWidthMatch{}
	bestDifferenceMM := math.Inf(1)
	for tireWidthMM := MinInputTireWidthMM; tireWidthMM <= MaxInputTireWidthMM; tireWidthMM++ {
		calculation, err := CalculateTireRimEngineeringReference(system, innerWidthMM, tireWidthMM)
		if err != nil {
			continue
		}
		differenceMM := math.Abs(float64(calculation.TireWidthMM) - targetTireWidthMM)
		if differenceMM > bestDifferenceMM {
			continue
		}
		if differenceMM == bestDifferenceMM && bestMatch.TireWidthMM != 0 && tireWidthMM >= bestMatch.TireWidthMM {
			continue
		}
		bestDifferenceMM = differenceMM
		bestMatch = TireWidthMatch{
			TireWidthMM:                 tireWidthMM,
			TargetTireWidthDifferenceMM: roundTireRimWidthReferenceMetric(differenceMM, 1),
		}
	}
	return bestMatch
}

// BuildTireRimEngineeringSuggestion keeps the entered pairing's calculation
// and verdict intact, and adds a separate best tire for the fixed rim. The
// entered tire's application layer determines which selection path is used.
func BuildTireRimEngineeringSuggestion(
	calculation *TireRimEngineeringReferenceCalculation,
) *Suggestion {
	if calculation == nil {
		return nil
	}
	var aeroTargetOuterWidthRange *NumericRange
	if calculation.AeroTargetOuterWidthMM != nil {
		aeroTargetOuterWidthRange = &NumericRange{
			Min: *calculation.AeroTargetOuterWidthMM,
			Max: *calculation.AeroTargetOuterWidthMM,
		}
	}
	bestFitment := selectBestNormalTireWidthForKnownRimInnerWidth(calculation.RimSystem, calculation.RimInnerWidthMM, calculation.ApplicationProfile)
	if calculation.AeroRuleApplies {
		bestFitment = nil
		if target := calculation.AeroTargetTireWidthMM; target != nil && *target <= TireRimEngineeringAeroRuleMaximumTireWidthMM {
			bestFitment = buildTireRimAeroBestFitmentWithinNormalRange(calculation.RimSystem, calculation.RimInnerWidthMM, int(math.Round(*target)), *target)
		}
	}
	return &Suggestion{
		TireWidthMM:                 calculation.TireWidthMM,
		RimSystem:                   calculation.RimSystem,
		ApplicationProfile:          calculation.ApplicationProfile,
		ApplicationProfileSelection: calculation.ApplicationProfileSelection,
		ResultKind:                  "engineering_exact",
		AeroRuleApplies:             calculation.AeroRuleApplies,
		BestFitment:                 bestFitment,
		RimWidthRanges: []WidthRange{{
			Min: calculation.RimInnerWidthMM,
			Max: calculation.RimInnerWidthMM,
		}},
		NormalFitmentRimWidthRanges: findInnerWidthRangesForTireWidth(calculation.RimSystem, calculation.TireWidthMM, false, calculation.ApplicationProfile),
		RimOuterWidthEstimates:      buildTireRimEngineeringRimOuterWidthEstimates([]int{calculation.RimInnerWidthMM}),
		DerivedMetrics: DerivedMetrics{
			Status:          "engineering",
			AeroRuleApplies: calculation.AeroRuleApplies,
			TireWidthMM: NumericRange{
				Min: float64(calculation.TireWidthMM),
				Max: float64(calculation.TireWidthMM),
			},
			AeroTargetOuterWidthMM: aeroTargetOuterWidthRange,
		},
		ModelVersion:           TireRimEngineeringReferenceModelVersion,
		KnowledgeAsOf:          TireRimEngineeringReferenceKnowledgeAsOf,
		Limitations:            cloneTireRimWidthReferenceStrings(tireRimEngineeringLimitations),
		Engineering:            calculation,
		StandardReference:      calculation.StandardReference,
		NormalFitmentReference: calculation.NormalFitmentReference,
	}
}

func calculateTireRimEngineeringMaximumPressure(system RimSystem, tireWidthMM int) TireRimEngineeringPressureLimit {
	if system == RimSystemHookless {
		switch {
		case tireWidthMM <= 24:
			return TireRimEngineeringPressureLimit{Bar: 5.5, PSI: 80}
		case tireWidthMM <= 29:
			return TireRimEngineeringPressureLimit{Bar: 5, PSI: 72.5}
		case tireWidthMM <= 34:
			return TireRimEngineeringPressureLimit{Bar: 4.5, PSI: 65}
		case tireWidthMM <= 39:
			return TireRimEngineeringPressureLimit{Bar: 4, PSI: 58}
		case tireWidthMM <= 44:
			return TireRimEngineeringPressureLimit{Bar: 3.5, PSI: 51}
		default:
			return TireRimEngineeringPressureLimit{Bar: 3, PSI: 43.5}
		}
	}
	switch {
	case tireWidthMM <= 24:
		return TireRimEngineeringPressureLimit{Bar: 8, PSI: 116}
	case tireWidthMM <= 26:
		return TireRimEngineeringPressureLimit{Bar: 7.5, PSI: 110}
	case tireWidthMM <= 29:
		return TireRimEngineeringPressureLimit{Bar: 6.5, PSI: 95}
	case tireWidthMM <= 34:
		return TireRimEngineeringPressureLimit{Bar: 5.5, PSI: 80}
	case tireWidthMM <= 39:
		return TireRimEngineeringPressureLimit{Bar: 4.5, PSI: 65}
	case tireWidthMM <= 45:
		return TireRimEngineeringPressureLimit{Bar: 3.5, PSI: 50}
	default:
		return TireRimEngineeringPressureLimit{Bar: 2.8, PSI: 40}
	}
}

func calculateTireRimEngineeringTubeRecommendation(tireWidthMM int) TireRimEngineeringTubeRecommendation {
	switch {
	case tireWidthMM >= 90:
		return TireRimEngineeringTubeRecommendation{Band: "fat", WidthMinMM: 90, WidthMaxMM: 127}
	case tireWidthMM >= 60:
		return TireRimEngineeringTubeRecommendation{Band: "wide", WidthMinMM: 60, WidthMaxMM: 90}
	case tireWidthMM >= 38:
		return TireRimEngineeringTubeRecommendation{Band: "wide", WidthMinMM: 38, WidthMaxMM: 45}
	case tireWidthMM >= 29:
		return TireRimEngineeringTubeRecommendation{Band: "mid", WidthMinMM: 28, WidthMaxMM: 40}
	default:
		return TireRimEngineeringTubeRecommendation{Band: "road", WidthMinMM: 23, WidthMaxMM: 28}
	}
}
