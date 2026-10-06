package innertubefitment

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

const (
	ModelVersion  = "inner-tube-valve-fitment-v1"
	KnowledgeAsOf = "2026-10-05"

	OuterLipOffsetMM             = 6.5
	MinimumSafeExposureMM        = 15.0
	PreferredExposureMM          = 18.0
	MinimumEngineeringMarginMM   = 15
	PreferredEngineeringMarginMM = 20
	MinimumPumpHeadGripDepthMM   = 10
	MaximumPumpHeadGripDepthMM   = 30
	DefaultPumpHeadGripDepthMM   = MinimumEngineeringMarginMM
	PreferredExposureDeltaMM     = 3
	PreferredLengthDeltaMM       = PreferredEngineeringMarginMM - MinimumEngineeringMarginMM
	MinimumRimDepthMM            = 20
	MaximumRimDepthMM            = 100
	MaximumRimDepthUncertaintyMM = 5
	DefaultRimDepthUncertaintyMM = 2
	FitmentCalculationMethod     = "passage_depth = rim_depth - outer_lip_offset; effective_exposure = total_assembly_length - passage_depth; minimum_required_length = rim_depth + pump_head_grip_depth; preferred_minimum_length = minimum_required_length + 5 mm; status = optimal at pump_head_grip_depth + 3 mm, marginal at pump_head_grip_depth, unsafe below pump_head_grip_depth"
)

var (
	ErrInvalidRimDepth            = errors.New("rim depth is outside the supported engineering range")
	ErrInvalidValveLength         = errors.New("valve length is not a supported option")
	ErrInvalidExtenderLength      = errors.New("extender length is not a supported option")
	ErrInvalidFitmentMode         = errors.New("fitment mode is invalid")
	ErrMissingManualLength        = errors.New("manual fitment requires valve and extender lengths")
	ErrInvalidPumpHeadGripDepth   = errors.New("pump head grip depth is outside the supported range")
	ErrInvalidRimDepthUncertainty = errors.New("rim depth uncertainty is outside the supported range")
	ErrNonFiniteCalculation       = errors.New("inner-tube fitment calculation produced a non-finite metric")
)

type FitmentMode string

const (
	FitmentModeAutomatic FitmentMode = "automatic"
	FitmentModeManual    FitmentMode = "manual"
)

type FitmentStatus string

const (
	FitmentStatusOptimal  FitmentStatus = "optimal"
	FitmentStatusMarginal FitmentStatus = "marginal"
	FitmentStatusUnsafe   FitmentStatus = "unsafe"
)

type RecommendationKey string

const (
	RecommendationShortValve               RecommendationKey = "shortValve"
	RecommendationStandardValve            RecommendationKey = "standardValve"
	RecommendationMediumValve              RecommendationKey = "mediumValve"
	RecommendationHighRimValve             RecommendationKey = "highRimValve"
	RecommendationMediumValveWithExtender  RecommendationKey = "mediumValveWithExtender"
	RecommendationHighRimValveWithExtender RecommendationKey = "highRimValveWithExtender"
)

type RecommendationReasonKey string

const (
	RecommendationReasonShortestNativeValve      RecommendationReasonKey = "shortestNativeValveMeetsMinimum"
	RecommendationReasonShortestExtenderAssembly RecommendationReasonKey = "shortestExtenderAssemblyReachesMinimum"
	RecommendationReasonManualCombination        RecommendationReasonKey = "manualCombination"
)

type SolveInput struct {
	RimDepthMM            int
	Mode                  FitmentMode
	ValveLengthMM         *int
	ExtenderLengthMM      *int
	PumpHeadGripDepthMM   int
	RimDepthUncertaintyMM int
}

type Recommendation struct {
	ValveLengthMM            int                     `json:"valve_length_mm"`
	ExtenderLengthMM         int                     `json:"extender_length_mm"`
	TotalAssemblyLengthMM    int                     `json:"total_assembly_length_mm"`
	MinimumRequiredLengthMM  int                     `json:"minimum_required_length_mm"`
	PreferredMinimumLengthMM int                     `json:"preferred_minimum_length_mm"`
	RecommendationKey        RecommendationKey       `json:"recommendation_key"`
	RecommendationReasonKey  RecommendationReasonKey `json:"recommendation_reason_key"`
}

type FitmentResult struct {
	ModelVersion             string                    `json:"model_version"`
	KnowledgeAsOf            string                    `json:"knowledge_as_of"`
	RimDepthMM               int                       `json:"rim_depth_mm"`
	OuterLipOffsetMM         float64                   `json:"outer_lip_offset_mm"`
	PassageDepthMM           float64                   `json:"passage_depth_mm"`
	MinimumRequiredLengthMM  int                       `json:"minimum_required_length_mm"`
	PreferredMinimumLengthMM int                       `json:"preferred_minimum_length_mm"`
	MinimumLengthMarginMM    int                       `json:"minimum_length_margin_mm"`
	PreferredLengthMarginMM  int                       `json:"preferred_length_margin_mm"`
	ValveLengthMM            int                       `json:"valve_length_mm"`
	ExtenderLengthMM         int                       `json:"extender_length_mm"`
	TotalAssemblyLengthMM    int                       `json:"total_assembly_length_mm"`
	EffectiveExposureMM      float64                   `json:"effective_exposure_mm"`
	Status                   FitmentStatus             `json:"status"`
	RecommendationKey        RecommendationKey         `json:"recommendation_key"`
	RecommendationReasonKey  RecommendationReasonKey   `json:"recommendation_reason_key"`
	RecommendationReason     string                    `json:"recommendation_reason"`
	Recommendation           Recommendation            `json:"recommendation"`
	Alternatives             []FitmentAlternative      `json:"alternatives"`
	PumpHeadGripDepthMM      int                       `json:"pump_head_grip_depth_mm"`
	PreferredExposureMM      float64                   `json:"preferred_exposure_mm"`
	RimDepthUncertaintyMM    int                       `json:"rim_depth_uncertainty_mm"`
	UncertaintyReview        *FitmentUncertaintyReview `json:"uncertainty_review,omitempty"`
	CalculationMethod        string                    `json:"calculation_method"`
	Source                   ModelSource               `json:"source"`
	Limitations              []string                  `json:"limitations"`
}

type ModelSource struct {
	Name       string `json:"name"`
	Provenance string `json:"provenance"`
}

type ClearanceCell struct {
	ValveLengthMM       int           `json:"valve_length_mm"`
	EffectiveExposureMM float64       `json:"effective_exposure_mm"`
	Status              FitmentStatus `json:"status"`
}

type FitmentAlternative struct {
	ValveLengthMM             int           `json:"valve_length_mm"`
	ExtenderLengthMM          int           `json:"extender_length_mm"`
	TotalAssemblyLengthMM     int           `json:"total_assembly_length_mm"`
	EffectiveExposureMM       float64       `json:"effective_exposure_mm"`
	MinimumLengthMarginMM     int           `json:"minimum_length_margin_mm"`
	PreferredLengthMarginMM   int           `json:"preferred_length_margin_mm"`
	Status                    FitmentStatus `json:"status"`
	IsAutomaticRecommendation bool          `json:"is_automatic_recommendation"`
}

type FitmentUncertaintyBoundary struct {
	RimDepthMM              int           `json:"rim_depth_mm"`
	PassageDepthMM          float64       `json:"passage_depth_mm"`
	EffectiveExposureMM     float64       `json:"effective_exposure_mm"`
	MinimumRequiredLengthMM int           `json:"minimum_required_length_mm"`
	MinimumLengthMarginMM   int           `json:"minimum_length_margin_mm"`
	Status                  FitmentStatus `json:"status"`
}

type FitmentUncertaintyReview struct {
	RimDepthUncertaintyMM int                        `json:"rim_depth_uncertainty_mm"`
	LowerBound            FitmentUncertaintyBoundary `json:"lower_bound"`
	UpperBound            FitmentUncertaintyBoundary `json:"upper_bound"`
	IsStatusStable        bool                       `json:"is_status_stable"`
	WorstCaseStatus       FitmentStatus              `json:"worst_case_status"`
}

type MatrixRow struct {
	RimDepthMM        int             `json:"rim_depth_mm"`
	PassageDepthMM    float64         `json:"passage_depth_mm"`
	MinimumRequiredMM int             `json:"minimum_required_length_mm"`
	Clearances        []ClearanceCell `json:"clearances"`
	RecommendedResult FitmentResult   `json:"recommended_result"`
}

type MatrixMetadata struct {
	ModelVersion                          string      `json:"model_version"`
	KnowledgeAsOf                         string      `json:"knowledge_as_of"`
	RimDepthMinMM                         int         `json:"rim_depth_min_mm"`
	RimDepthMaxMM                         int         `json:"rim_depth_max_mm"`
	OuterLipOffsetMM                      float64     `json:"outer_lip_offset_mm"`
	MinimumSafeExposureMM                 float64     `json:"minimum_safe_exposure_mm"`
	PreferredExposureMM                   float64     `json:"preferred_exposure_mm"`
	PumpHeadGripDepthMinMM                int         `json:"pump_head_grip_depth_min_mm"`
	PumpHeadGripDepthMaxMM                int         `json:"pump_head_grip_depth_max_mm"`
	DefaultPumpHeadGripDepthMM            int         `json:"default_pump_head_grip_depth_mm"`
	RimDepthUncertaintyMaxMM              int         `json:"rim_depth_uncertainty_max_mm"`
	DefaultRimDepthUncertaintyMM          int         `json:"default_rim_depth_uncertainty_mm"`
	BaseInnerTubeValveLengthOptionsMM     []int       `json:"base_valve_length_options_mm"`
	InnerTubeValveExtenderLengthOptionsMM []int       `json:"extender_length_options_mm"`
	PresetInnerTubeRimDepthsMM            []int       `json:"preset_rim_depths_mm"`
	Rows                                  []MatrixRow `json:"rows"`
	CalculationMethod                     string      `json:"calculation_method"`
	Source                                ModelSource `json:"source"`
	Limitations                           []string    `json:"limitations"`
}

var innerTubeBaseValveLengthOptionsMM = []int{40, 48, 60, 80}
var innerTubeValveExtenderLengthOptionsMM = []int{0, 20, 30, 40, 60}
var innerTubePresetRimDepthsMM = []int{28, 35, 40, 45, 50, 55, 60, 70, 75, 80, 85, 90, 95, 100}

var innerTubeFitmentLimitations = []string{
	"This is a reference model for valve fitment, not a certification for a specific rim, tube, pump, or wheelset.",
	"Rim depth must be checked against the exact rim cross-section and valve-hole geometry before installation.",
	"The 6.5 mm outer-lip offset and the default 15 mm pump-head grip depth with 18 mm preferred exposure are model parameters; manufacturer instructions take precedence.",
	"The calculation does not certify tire pressure, hookless compatibility, or the safety of a tire and rim combination.",
}

var innerTubeFitmentModelSource = ModelSource{
	Name:       "Tanzanite inner-tube valve fitment reference model",
	Provenance: "Provided board formula and engineering assumptions; not a manufacturer certification",
}

func BaseInnerTubeValveLengthOptionsMM() []int {
	return append([]int(nil), innerTubeBaseValveLengthOptionsMM...)
}

func InnerTubeValveExtenderLengthOptionsMM() []int {
	return append([]int(nil), innerTubeValveExtenderLengthOptionsMM...)
}

func PresetInnerTubeRimDepthsMM() []int {
	return append([]int(nil), innerTubePresetRimDepthsMM...)
}

func ValidateInnerTubeRimDepth(rimDepthMM int) error {
	if rimDepthMM < MinimumRimDepthMM || rimDepthMM > MaximumRimDepthMM {
		return fmt.Errorf("%w: %d mm is outside [%d, %d]", ErrInvalidRimDepth, rimDepthMM, MinimumRimDepthMM, MaximumRimDepthMM)
	}
	return nil
}

func ValidateInnerTubeFitmentMode(mode FitmentMode) error {
	if mode != FitmentModeAutomatic && mode != FitmentModeManual {
		return fmt.Errorf("%w: %q", ErrInvalidFitmentMode, mode)
	}
	return nil
}

func ValidateInnerTubePumpHeadGripDepth(pumpHeadGripDepthMM int) error {
	if pumpHeadGripDepthMM < MinimumPumpHeadGripDepthMM || pumpHeadGripDepthMM > MaximumPumpHeadGripDepthMM {
		return fmt.Errorf("%w: %d mm is outside [%d, %d]", ErrInvalidPumpHeadGripDepth, pumpHeadGripDepthMM, MinimumPumpHeadGripDepthMM, MaximumPumpHeadGripDepthMM)
	}
	return nil
}

func ValidateInnerTubeRimDepthUncertainty(rimDepthUncertaintyMM int) error {
	if rimDepthUncertaintyMM < 0 || rimDepthUncertaintyMM > MaximumRimDepthUncertaintyMM {
		return fmt.Errorf("%w: %d mm is outside [0, %d]", ErrInvalidRimDepthUncertainty, rimDepthUncertaintyMM, MaximumRimDepthUncertaintyMM)
	}
	return nil
}

func getInnerTubeMinimumRequiredLengthMM(rimDepthMM, pumpHeadGripDepthMM int) int {
	return rimDepthMM + pumpHeadGripDepthMM
}

func getInnerTubePreferredMinimumLengthMM(rimDepthMM, pumpHeadGripDepthMM int) int {
	return rimDepthMM + pumpHeadGripDepthMM + PreferredLengthDeltaMM
}

func getInnerTubePreferredExposureMM(pumpHeadGripDepthMM int) float64 {
	return float64(pumpHeadGripDepthMM + PreferredExposureDeltaMM)
}

func SolveInnerTubeValveFitment(input SolveInput) (FitmentResult, error) {
	if err := ValidateInnerTubeRimDepth(input.RimDepthMM); err != nil {
		return FitmentResult{}, err
	}
	if err := ValidateInnerTubeFitmentMode(input.Mode); err != nil {
		return FitmentResult{}, err
	}
	pumpHeadGripDepthMM := input.PumpHeadGripDepthMM
	if pumpHeadGripDepthMM == 0 {
		pumpHeadGripDepthMM = DefaultPumpHeadGripDepthMM
	}
	if err := ValidateInnerTubePumpHeadGripDepth(pumpHeadGripDepthMM); err != nil {
		return FitmentResult{}, err
	}
	if err := ValidateInnerTubeRimDepthUncertainty(input.RimDepthUncertaintyMM); err != nil {
		return FitmentResult{}, err
	}

	recommendation, err := buildAutomaticInnerTubeValveRecommendation(input.RimDepthMM, pumpHeadGripDepthMM)
	if err != nil {
		return FitmentResult{}, err
	}
	valveLengthMM := recommendation.ValveLengthMM
	extenderLengthMM := recommendation.ExtenderLengthMM
	if input.Mode == FitmentModeManual {
		if input.ValveLengthMM == nil || input.ExtenderLengthMM == nil {
			return FitmentResult{}, ErrMissingManualLength
		}
		valveLengthMM = *input.ValveLengthMM
		extenderLengthMM = *input.ExtenderLengthMM
		if !containsSupportedInnerTubeFitmentOption(innerTubeBaseValveLengthOptionsMM, valveLengthMM) {
			return FitmentResult{}, fmt.Errorf("%w: %d mm", ErrInvalidValveLength, valveLengthMM)
		}
		if !containsSupportedInnerTubeFitmentOption(innerTubeValveExtenderLengthOptionsMM, extenderLengthMM) {
			return FitmentResult{}, fmt.Errorf("%w: %d mm", ErrInvalidExtenderLength, extenderLengthMM)
		}
	}

	minimumRequiredLengthMM := getInnerTubeMinimumRequiredLengthMM(input.RimDepthMM, pumpHeadGripDepthMM)
	preferredMinimumLengthMM := getInnerTubePreferredMinimumLengthMM(input.RimDepthMM, pumpHeadGripDepthMM)
	passageDepthMM, err := roundInnerTubeFitmentMetricToTenthMillimetre(float64(input.RimDepthMM) - OuterLipOffsetMM)
	if err != nil {
		return FitmentResult{}, err
	}
	totalAssemblyLengthMM := valveLengthMM + extenderLengthMM
	effectiveExposureMM, err := roundInnerTubeFitmentMetricToTenthMillimetre(float64(totalAssemblyLengthMM) - passageDepthMM)
	if err != nil {
		return FitmentResult{}, err
	}
	status := classifyInnerTubeValveExposureStatus(effectiveExposureMM, pumpHeadGripDepthMM)
	selectedRecommendation := recommendation
	recommendationReasonKey := recommendation.RecommendationReasonKey
	if input.Mode == FitmentModeManual {
		selectedRecommendation = buildManualInnerTubeValveRecommendation(input.RimDepthMM, valveLengthMM, extenderLengthMM, pumpHeadGripDepthMM)
		recommendationReasonKey = RecommendationReasonManualCombination
	}
	alternatives, err := buildInnerTubeValveFitmentAlternatives(
		input.RimDepthMM,
		minimumRequiredLengthMM,
		pumpHeadGripDepthMM,
		recommendation.ValveLengthMM,
		recommendation.ExtenderLengthMM,
	)
	if err != nil {
		return FitmentResult{}, err
	}
	uncertaintyReview, err := buildInnerTubeValveFitmentUncertaintyReview(
		input.RimDepthMM,
		input.RimDepthUncertaintyMM,
		totalAssemblyLengthMM,
		pumpHeadGripDepthMM,
	)
	if err != nil {
		return FitmentResult{}, err
	}

	return FitmentResult{
		ModelVersion:             ModelVersion,
		KnowledgeAsOf:            KnowledgeAsOf,
		RimDepthMM:               input.RimDepthMM,
		OuterLipOffsetMM:         OuterLipOffsetMM,
		PassageDepthMM:           passageDepthMM,
		MinimumRequiredLengthMM:  minimumRequiredLengthMM,
		PreferredMinimumLengthMM: preferredMinimumLengthMM,
		MinimumLengthMarginMM:    totalAssemblyLengthMM - minimumRequiredLengthMM,
		PreferredLengthMarginMM:  totalAssemblyLengthMM - preferredMinimumLengthMM,
		ValveLengthMM:            valveLengthMM,
		ExtenderLengthMM:         extenderLengthMM,
		TotalAssemblyLengthMM:    totalAssemblyLengthMM,
		EffectiveExposureMM:      effectiveExposureMM,
		Status:                   status,
		RecommendationKey:        selectedRecommendation.RecommendationKey,
		RecommendationReasonKey:  recommendationReasonKey,
		RecommendationReason:     getInnerTubeValveRecommendationReasonText(recommendationReasonKey),
		Recommendation:           selectedRecommendation,
		Alternatives:             alternatives,
		PumpHeadGripDepthMM:      pumpHeadGripDepthMM,
		PreferredExposureMM:      getInnerTubePreferredExposureMM(pumpHeadGripDepthMM),
		RimDepthUncertaintyMM:    input.RimDepthUncertaintyMM,
		UncertaintyReview:        uncertaintyReview,
		CalculationMethod:        FitmentCalculationMethod,
		Source:                   innerTubeFitmentModelSource,
		Limitations:              cloneInnerTubeFitmentLimitations(innerTubeFitmentLimitations),
	}, nil
}

func GetInnerTubeValveFitmentMatrixMetadata() MatrixMetadata {
	rows := make([]MatrixRow, 0, len(innerTubePresetRimDepthsMM))
	for _, rimDepthMM := range innerTubePresetRimDepthsMM {
		result, err := SolveInnerTubeValveFitment(SolveInput{RimDepthMM: rimDepthMM, Mode: FitmentModeAutomatic})
		if err != nil {
			continue
		}
		clearances := make([]ClearanceCell, 0, len(innerTubeBaseValveLengthOptionsMM))
		for _, valveLengthMM := range innerTubeBaseValveLengthOptionsMM {
			manualResult, manualErr := SolveInnerTubeValveFitment(SolveInput{
				RimDepthMM: rimDepthMM, Mode: FitmentModeManual,
				ValveLengthMM: &valveLengthMM, ExtenderLengthMM: newInnerTubeFitmentIntegerPointer(0),
			})
			if manualErr != nil {
				continue
			}
			clearances = append(clearances, ClearanceCell{
				ValveLengthMM: valveLengthMM, EffectiveExposureMM: manualResult.EffectiveExposureMM, Status: manualResult.Status,
			})
		}
		rows = append(rows, MatrixRow{
			RimDepthMM: rimDepthMM, PassageDepthMM: result.PassageDepthMM,
			MinimumRequiredMM: result.MinimumRequiredLengthMM, Clearances: clearances,
			RecommendedResult: result,
		})
	}
	return MatrixMetadata{
		ModelVersion: ModelVersion, KnowledgeAsOf: KnowledgeAsOf,
		RimDepthMinMM: MinimumRimDepthMM, RimDepthMaxMM: MaximumRimDepthMM,
		OuterLipOffsetMM: OuterLipOffsetMM, MinimumSafeExposureMM: MinimumSafeExposureMM,
		PreferredExposureMM:               PreferredExposureMM,
		PumpHeadGripDepthMinMM:            MinimumPumpHeadGripDepthMM,
		PumpHeadGripDepthMaxMM:            MaximumPumpHeadGripDepthMM,
		DefaultPumpHeadGripDepthMM:        DefaultPumpHeadGripDepthMM,
		RimDepthUncertaintyMaxMM:          MaximumRimDepthUncertaintyMM,
		DefaultRimDepthUncertaintyMM:      DefaultRimDepthUncertaintyMM,
		BaseInnerTubeValveLengthOptionsMM: BaseInnerTubeValveLengthOptionsMM(), InnerTubeValveExtenderLengthOptionsMM: InnerTubeValveExtenderLengthOptionsMM(),
		PresetInnerTubeRimDepthsMM: PresetInnerTubeRimDepthsMM(), Rows: rows,
		CalculationMethod: "The Go domain engine owns the valve passage and exposure calculation; the browser only renders the returned values. Recommendations prefer the shortest native valve meeting the minimum length, then the shortest supported extender assembly. " + FitmentCalculationMethod,
		Source:            innerTubeFitmentModelSource,
		Limitations:       cloneInnerTubeFitmentLimitations(innerTubeFitmentLimitations),
	}
}

func buildAutomaticInnerTubeValveRecommendation(rimDepthMM, pumpHeadGripDepthMM int) (Recommendation, error) {
	minimumRequiredLengthMM := getInnerTubeMinimumRequiredLengthMM(rimDepthMM, pumpHeadGripDepthMM)
	preferredMinimumLengthMM := getInnerTubePreferredMinimumLengthMM(rimDepthMM, pumpHeadGripDepthMM)
	for _, valveLengthMM := range innerTubeBaseValveLengthOptionsMM {
		if valveLengthMM >= minimumRequiredLengthMM {
			return Recommendation{
				ValveLengthMM:            valveLengthMM,
				TotalAssemblyLengthMM:    valveLengthMM,
				MinimumRequiredLengthMM:  minimumRequiredLengthMM,
				PreferredMinimumLengthMM: preferredMinimumLengthMM,
				RecommendationKey:        getInnerTubeValveRecommendationKey(valveLengthMM, 0),
				RecommendationReasonKey:  RecommendationReasonShortestNativeValve,
			}, nil
		}
	}
	type candidate struct {
		valveLengthMM, extenderLengthMM, totalLengthMM int
	}
	candidates := make([]candidate, 0, len(innerTubeBaseValveLengthOptionsMM)*len(innerTubeValveExtenderLengthOptionsMM))
	for _, valveLengthMM := range innerTubeBaseValveLengthOptionsMM {
		for _, extenderLengthMM := range innerTubeValveExtenderLengthOptionsMM {
			if valveLengthMM+extenderLengthMM < minimumRequiredLengthMM {
				continue
			}
			candidates = append(candidates, candidate{
				valveLengthMM: valveLengthMM, extenderLengthMM: extenderLengthMM,
				totalLengthMM: valveLengthMM + extenderLengthMM,
			})
		}
	}
	sort.Slice(candidates, func(left, right int) bool {
		if candidates[left].totalLengthMM != candidates[right].totalLengthMM {
			return candidates[left].totalLengthMM < candidates[right].totalLengthMM
		}
		if candidates[left].extenderLengthMM != candidates[right].extenderLengthMM {
			return candidates[left].extenderLengthMM < candidates[right].extenderLengthMM
		}
		return candidates[left].valveLengthMM > candidates[right].valveLengthMM
	})
	if len(candidates) == 0 {
		return Recommendation{}, fmt.Errorf("no supported valve and extender assembly reaches %d mm", minimumRequiredLengthMM)
	}
	selected := candidates[0]
	return Recommendation{
		ValveLengthMM: selected.valveLengthMM, ExtenderLengthMM: selected.extenderLengthMM,
		TotalAssemblyLengthMM:   selected.totalLengthMM,
		MinimumRequiredLengthMM: minimumRequiredLengthMM, PreferredMinimumLengthMM: preferredMinimumLengthMM,
		RecommendationKey:       getInnerTubeValveRecommendationKey(selected.valveLengthMM, selected.extenderLengthMM),
		RecommendationReasonKey: RecommendationReasonShortestExtenderAssembly,
	}, nil
}

func buildManualInnerTubeValveRecommendation(rimDepthMM, valveLengthMM, extenderLengthMM, pumpHeadGripDepthMM int) Recommendation {
	minimumRequiredLengthMM := getInnerTubeMinimumRequiredLengthMM(rimDepthMM, pumpHeadGripDepthMM)
	preferredMinimumLengthMM := getInnerTubePreferredMinimumLengthMM(rimDepthMM, pumpHeadGripDepthMM)
	return Recommendation{
		ValveLengthMM: valveLengthMM, ExtenderLengthMM: extenderLengthMM,
		TotalAssemblyLengthMM:   valveLengthMM + extenderLengthMM,
		MinimumRequiredLengthMM: minimumRequiredLengthMM, PreferredMinimumLengthMM: preferredMinimumLengthMM,
		RecommendationKey:       getInnerTubeValveRecommendationKey(valveLengthMM, extenderLengthMM),
		RecommendationReasonKey: RecommendationReasonManualCombination,
	}
}

func getInnerTubeValveRecommendationKey(valveLengthMM, extenderLengthMM int) RecommendationKey {
	switch {
	case extenderLengthMM == 0 && valveLengthMM <= 40:
		return RecommendationShortValve
	case extenderLengthMM == 0 && valveLengthMM <= 48:
		return RecommendationStandardValve
	case extenderLengthMM == 0 && valveLengthMM <= 60:
		return RecommendationMediumValve
	case extenderLengthMM == 0:
		return RecommendationHighRimValve
	case valveLengthMM <= 60:
		return RecommendationMediumValveWithExtender
	default:
		return RecommendationHighRimValveWithExtender
	}
}

func getInnerTubeValveRecommendationReasonText(reasonKey RecommendationReasonKey) string {
	switch reasonKey {
	case RecommendationReasonShortestNativeValve:
		return "shortest supported native valve that meets the minimum required total length"
	case RecommendationReasonShortestExtenderAssembly:
		return "shortest supported valve and extender assembly that reaches the minimum required total length"
	case RecommendationReasonManualCombination:
		return "manual valve and extender combination supplied for fitment checking"
	default:
		return "reference fitment recommendation"
	}
}

func buildInnerTubeValveFitmentAlternatives(rimDepthMM, minimumRequiredLengthMM, pumpHeadGripDepthMM, recommendedValveLengthMM, recommendedExtenderLengthMM int) ([]FitmentAlternative, error) {
	passageDepthMM, err := roundInnerTubeFitmentMetricToTenthMillimetre(float64(rimDepthMM) - OuterLipOffsetMM)
	if err != nil {
		return nil, err
	}
	preferredMinimumLengthMM := getInnerTubePreferredMinimumLengthMM(rimDepthMM, pumpHeadGripDepthMM)
	alternatives := make([]FitmentAlternative, 0, len(innerTubeBaseValveLengthOptionsMM))
	for _, valveLengthMM := range innerTubeBaseValveLengthOptionsMM {
		for _, extenderLengthMM := range innerTubeValveExtenderLengthOptionsMM {
			totalAssemblyLengthMM := valveLengthMM + extenderLengthMM
			if totalAssemblyLengthMM < minimumRequiredLengthMM {
				continue
			}
			exposureMM, exposureErr := roundInnerTubeFitmentMetricToTenthMillimetre(float64(totalAssemblyLengthMM) - passageDepthMM)
			if exposureErr != nil {
				return nil, exposureErr
			}
			alternatives = append(alternatives, FitmentAlternative{
				ValveLengthMM:             valveLengthMM,
				ExtenderLengthMM:          extenderLengthMM,
				TotalAssemblyLengthMM:     totalAssemblyLengthMM,
				EffectiveExposureMM:       exposureMM,
				MinimumLengthMarginMM:     totalAssemblyLengthMM - minimumRequiredLengthMM,
				PreferredLengthMarginMM:   totalAssemblyLengthMM - preferredMinimumLengthMM,
				Status:                    classifyInnerTubeValveExposureStatus(exposureMM, pumpHeadGripDepthMM),
				IsAutomaticRecommendation: valveLengthMM == recommendedValveLengthMM && extenderLengthMM == recommendedExtenderLengthMM,
			})
			break
		}
	}
	sort.SliceStable(alternatives, func(left, right int) bool {
		if alternatives[left].IsAutomaticRecommendation != alternatives[right].IsAutomaticRecommendation {
			return alternatives[left].IsAutomaticRecommendation
		}
		if alternatives[left].TotalAssemblyLengthMM != alternatives[right].TotalAssemblyLengthMM {
			return alternatives[left].TotalAssemblyLengthMM < alternatives[right].TotalAssemblyLengthMM
		}
		if alternatives[left].ExtenderLengthMM != alternatives[right].ExtenderLengthMM {
			return alternatives[left].ExtenderLengthMM < alternatives[right].ExtenderLengthMM
		}
		return alternatives[left].ValveLengthMM > alternatives[right].ValveLengthMM
	})
	return alternatives, nil
}

func buildInnerTubeValveFitmentUncertaintyReview(rimDepthMM, rimDepthUncertaintyMM, totalAssemblyLengthMM, pumpHeadGripDepthMM int) (*FitmentUncertaintyReview, error) {
	if rimDepthUncertaintyMM <= 0 {
		return nil, nil
	}
	lowerBoundMM := rimDepthMM - rimDepthUncertaintyMM
	if lowerBoundMM < MinimumRimDepthMM {
		lowerBoundMM = MinimumRimDepthMM
	}
	upperBoundMM := rimDepthMM + rimDepthUncertaintyMM
	if upperBoundMM > MaximumRimDepthMM {
		upperBoundMM = MaximumRimDepthMM
	}
	lowerBoundary, err := buildInnerTubeValveFitmentUncertaintyBoundary(lowerBoundMM, totalAssemblyLengthMM, pumpHeadGripDepthMM)
	if err != nil {
		return nil, err
	}
	upperBoundary, err := buildInnerTubeValveFitmentUncertaintyBoundary(upperBoundMM, totalAssemblyLengthMM, pumpHeadGripDepthMM)
	if err != nil {
		return nil, err
	}
	worstCaseStatus := upperBoundary.Status
	if innerTubeFitmentStatusSeverityRank(lowerBoundary.Status) > innerTubeFitmentStatusSeverityRank(worstCaseStatus) {
		worstCaseStatus = lowerBoundary.Status
	}
	return &FitmentUncertaintyReview{
		RimDepthUncertaintyMM: rimDepthUncertaintyMM,
		LowerBound:            lowerBoundary,
		UpperBound:            upperBoundary,
		IsStatusStable:        lowerBoundary.Status == upperBoundary.Status,
		WorstCaseStatus:       worstCaseStatus,
	}, nil
}

func buildInnerTubeValveFitmentUncertaintyBoundary(rimDepthMM, totalAssemblyLengthMM, pumpHeadGripDepthMM int) (FitmentUncertaintyBoundary, error) {
	passageDepthMM, err := roundInnerTubeFitmentMetricToTenthMillimetre(float64(rimDepthMM) - OuterLipOffsetMM)
	if err != nil {
		return FitmentUncertaintyBoundary{}, err
	}
	effectiveExposureMM, err := roundInnerTubeFitmentMetricToTenthMillimetre(float64(totalAssemblyLengthMM) - passageDepthMM)
	if err != nil {
		return FitmentUncertaintyBoundary{}, err
	}
	minimumRequiredLengthMM := getInnerTubeMinimumRequiredLengthMM(rimDepthMM, pumpHeadGripDepthMM)
	return FitmentUncertaintyBoundary{
		RimDepthMM:              rimDepthMM,
		PassageDepthMM:          passageDepthMM,
		EffectiveExposureMM:     effectiveExposureMM,
		MinimumRequiredLengthMM: minimumRequiredLengthMM,
		MinimumLengthMarginMM:   totalAssemblyLengthMM - minimumRequiredLengthMM,
		Status:                  classifyInnerTubeValveExposureStatus(effectiveExposureMM, pumpHeadGripDepthMM),
	}, nil
}

func innerTubeFitmentStatusSeverityRank(status FitmentStatus) int {
	switch status {
	case FitmentStatusUnsafe:
		return 3
	case FitmentStatusMarginal:
		return 2
	default:
		return 1
	}
}

func classifyInnerTubeValveExposureStatus(exposureMM float64, pumpHeadGripDepthMM int) FitmentStatus {
	if exposureMM >= getInnerTubePreferredExposureMM(pumpHeadGripDepthMM) {
		return FitmentStatusOptimal
	}
	if exposureMM >= float64(pumpHeadGripDepthMM) {
		return FitmentStatusMarginal
	}
	return FitmentStatusUnsafe
}

func containsSupportedInnerTubeFitmentOption(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func newInnerTubeFitmentIntegerPointer(value int) *int {
	return &value
}

func roundInnerTubeFitmentMetricToTenthMillimetre(value float64) (float64, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, ErrNonFiniteCalculation
	}
	return math.Round(value*10) / 10, nil
}

func cloneInnerTubeFitmentLimitations(values []string) []string {
	return append([]string(nil), values...)
}
