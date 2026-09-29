package service

import (
	productdomain "commerce-platform/internal/domain/product"
	domainspoke "commerce-platform/internal/domain/spoke"
	"commerce-platform/internal/repository"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrSpokeGeometryNotFound    = errors.New("unknown rim or hub geometry")
	ErrSpokeRimGeometryMissing  = errors.New("rim geometry not available")
	ErrSpokeHubGeometryMissing  = errors.New("hub geometry not available for requested position")
	ErrInvalidSpokeCalculation  = errors.New("invalid spoke calculation input")
	ErrInvalidSpokeCatalog      = errors.New("invalid spoke catalog")
	spokeCalculationFormulaName = "v1.4-go-backend-physical-build-corrections"
	spokeCatalogIDPattern       = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,139}$`)
)

type SpokeService struct {
	spokeRepo *repository.SpokeRepository
}

type SpokeCalculationInput struct {
	RimID          string
	HubID          string
	WheelPosition  string
	SpokeCount     int
	Crossing       int
	NippleType     string
	NippleLengthMM *float64
	// SpokeHeadType selects the physical hub interface. J-bend uses the
	// flange-hole contact geometry; straight-pull uses a tangential slot
	// offset. Both paths apply the flange-hole inner-edge correction.
	SpokeHeadType               string
	SpokeHoleDiameterMM         *float64
	StraightPullTangentOffsetMM *float64
	// SpokeProfile and TargetTensionN are used to estimate elastic stretch
	// under the target build tension. A nil/zero target keeps the legacy
	// un-stretched result for API clients that do not provide this option.
	SpokeProfile   string
	TargetTensionN *float64
	// AlternatingDrillingOffsetMM is the signed axial offset of the selected
	// alternating rim hole. The right side receives the opposite offset.
	AlternatingDrillingOffsetMM *float64
	Interlacing                 bool
	InterlaceCompensationMM     *float64
	// RimOffsetMM is positive when the rim center moves toward the right
	// flange. The value changes both the spoke length geometry and bracing
	// angles used for the tension-ratio estimate.
	RimOffsetMM float64
	// Optional user-entered geometry. The public Nuxt calculator intentionally
	// sends empty RimID/HubID values and relies on this manual geometry. Catalog
	// ID resolution remains available for controlled legacy/integration callers;
	// it must not be wired to the public catalog selector as automatic fill.
	ERDMM            *float64
	LeftFlangeMM     *float64
	RightFlangeMM    *float64
	LeftFlangePCDMM  *float64
	RightFlangePCDMM *float64
	UserID           *uint
}

type SpokeCalculationResult struct {
	LeftLengthMM  float64               `json:"leftLengthMm"`
	RightLengthMM float64               `json:"rightLengthMm"`
	TensionRatio  *SpokeTensionRatio    `json:"tensionRatio,omitempty"`
	Debug         SpokeCalculationDebug `json:"debug"`
}

type SpokeCalculationDebug struct {
	// Geometry is intentionally internal-only; exposing it lets clients
	// reconstruct the proprietary catalog from a single calculation response.
	Rim                         *domainspoke.RimModel    `json:"-"`
	Hub                         *domainspoke.HubGeometry `json:"-"`
	RimOffsetMM                 float64                  `json:"rimOffsetMm"`
	SpokeHeadType               string                   `json:"spokeHeadType"`
	SpokeHoleCorrectionMM       float64                  `json:"spokeHoleCorrectionMm"`
	StraightPullTangentOffsetMM float64                  `json:"straightPullTangentOffsetMm"`
	AlternatingDrillingOffsetMM float64                  `json:"alternatingDrillingOffsetMm"`
	StretchLeftMM               float64                  `json:"stretchLeftMm"`
	StretchRightMM              float64                  `json:"stretchRightMm"`
	InterlaceCompensationMM     float64                  `json:"interlaceCompensationMm"`
	FormulaVersion              string                   `json:"formulaVersion"`
}

type SpokeTensionRatio struct {
	// LeftToRight is the estimated per-spoke tension ratio T_left / T_right.
	LeftToRight float64 `json:"leftToRight"`
	// RightToLeft is the reciprocal ratio T_right / T_left.
	RightToLeft float64 `json:"rightToLeft"`
	// LowerToHigher is always <= 1 and is the most useful wheel-builder
	// summary, for example 0.78 means 78% on the lower-tension side.
	LowerToHigher float64 `json:"lowerToHigher"`
	LowerSide     string  `json:"lowerSide"`

	LeftBracingAngleDeg  float64 `json:"leftBracingAngleDeg"`
	RightBracingAngleDeg float64 `json:"rightBracingAngleDeg"`
}

func NewSpokeService(spokeRepo *repository.SpokeRepository) *SpokeService {
	return &SpokeService{spokeRepo: spokeRepo}
}

func (s *SpokeService) GetExport() (domainspoke.ExportResponse, error) {
	export, found, err := s.spokeRepo.GetCatalogExport()
	if err != nil {
		return domainspoke.ExportResponse{}, err
	}
	if !found {
		return domainspoke.DefaultExport(), nil
	}
	return export, nil
}

// GetPublicExport intentionally omits all proprietary geometry and verified
// build lengths. The browser only needs stable identifiers and labels; the
// calculation engine resolves dimensions server-side.
func (s *SpokeService) GetPublicExport() (domainspoke.ExportResponse, error) {
	export, err := s.GetExport()
	if err != nil {
		return domainspoke.ExportResponse{}, err
	}
	for bi := range export.Rims {
		for mi := range export.Rims[bi].Items {
			export.Rims[bi].Items[mi].ERD = nil
			export.Rims[bi].Items[mi].Weight = nil
		}
	}
	for bi := range export.Hubs {
		for mi := range export.Hubs[bi].Items {
			export.Hubs[bi].Items[mi].Front = nil
			export.Hubs[bi].Items[mi].Rear = nil
		}
	}
	for pi := range export.Presets {
		export.Presets[pi].NippleLength = nil
		export.Presets[pi].ActualLengths = nil
	}
	return export, nil
}

// GetPublicRecordedResults returns the narrow result projection consumed by
// the lower browser search card. It deliberately does not reuse the public
// catalog export: that export hides recorded measurements, while this endpoint
// exposes only verified cut lengths and the metadata needed to find them.
func (s *SpokeService) GetPublicRecordedResults() (domainspoke.RecordedResultsResponse, error) {
	export, err := s.GetExport()
	if err != nil {
		return domainspoke.RecordedResultsResponse{}, err
	}

	results := make([]domainspoke.RecordedBuildResult, 0, len(export.Presets))
	for _, preset := range export.Presets {
		actual := preset.ActualLengths
		if actual == nil || !hasRecordedSpokeLength(actual) {
			continue
		}

		results = append(results, domainspoke.RecordedBuildResult{
			ID:            preset.ID,
			Name:          preset.Name,
			Keywords:      append([]string(nil), preset.Keywords...),
			Description:   preset.Description,
			RimBrandID:    preset.RimBrandID,
			RimModelID:    preset.RimModelID,
			HubBrandID:    preset.HubBrandID,
			HubModelID:    preset.HubModelID,
			WheelPosition: preset.WheelPosition,
			SpokeCount:    preset.SpokeCount,
			Crossing:      preset.Crossing,
			NippleType:    preset.NippleType,
			ActualLengths: domainspoke.RecordedBuildActualLengths{
				FrontLeft:  actual.FrontLeft,
				FrontRight: actual.FrontRight,
				RearLeft:   actual.RearLeft,
				RearRight:  actual.RearRight,
			},
		})
	}

	return domainspoke.RecordedResultsResponse{Presets: results}, nil
}

func hasRecordedSpokeLength(actual *domainspoke.WheelBuildActualLengths) bool {
	return actual != nil && (actual.FrontLeft != nil || actual.FrontRight != nil || actual.RearLeft != nil || actual.RearRight != nil)
}

func (s *SpokeService) ReplaceCatalog(export domainspoke.ExportResponse) (domainspoke.ExportResponse, error) {
	if s.spokeRepo.UsesFitmentHubSpecifications() {
		authoritativeHubs, configured, err := s.spokeRepo.GetAuthoritativeHubBrands()
		if err != nil {
			return domainspoke.ExportResponse{}, err
		}
		if configured {
			export.Hubs = authoritativeHubs
		}
	}

	productBrands, err := s.spokeRepo.ListProductBrands(true)
	if err != nil {
		return domainspoke.ExportResponse{}, err
	}

	normalized, err := normalizeSpokeCatalog(export, productBrands)
	if err != nil {
		return domainspoke.ExportResponse{}, err
	}
	if err := s.spokeRepo.ReplaceCatalog(normalized); err != nil {
		return domainspoke.ExportResponse{}, err
	}
	nextExport, err := s.GetExport()
	if err != nil {
		return domainspoke.ExportResponse{}, err
	}
	return nextExport, nil
}

func (s *SpokeService) ListHistory(search string, page, pageSize int) ([]domainspoke.History, int64, error) {
	return s.spokeRepo.ListHistory(search, page, pageSize)
}

func (s *SpokeService) ListUserHistory(userID uint, search string, page, pageSize int) ([]domainspoke.History, int64, error) {
	return s.spokeRepo.ListHistoryByUserID(userID, search, page, pageSize)
}

func (s *SpokeService) Calculate(input SpokeCalculationInput) (*SpokeCalculationResult, error) {
	if !isFinite(input.RimOffsetMM) || math.Abs(input.RimOffsetMM) > 20 {
		return nil, ErrInvalidSpokeCalculation
	}
	options := domainspoke.DefaultOptions()
	if _, exists := intOptionSet(options.SpokeCounts)[input.SpokeCount]; !exists {
		return nil, ErrInvalidSpokeCalculation
	}
	if _, exists := intOptionSet(options.Crossings)[input.Crossing]; !exists {
		return nil, ErrInvalidSpokeCalculation
	}
	if input.WheelPosition != "auto" && input.WheelPosition != "front" && input.WheelPosition != "rear" {
		return nil, ErrInvalidSpokeCalculation
	}
	input.NippleType = strings.ToLower(strings.TrimSpace(input.NippleType))
	if input.NippleType == "" {
		input.NippleType = "standard"
	}
	if input.NippleType != "" && input.NippleType != "standard" && input.NippleType != "hidden" {
		return nil, ErrInvalidSpokeCalculation
	}
	if input.NippleLengthMM != nil && (!isFinite(*input.NippleLengthMM) || *input.NippleLengthMM < 0 || *input.NippleLengthMM > 40) {
		return nil, ErrInvalidSpokeCalculation
	}

	input.SpokeHeadType = strings.ToLower(strings.TrimSpace(input.SpokeHeadType))
	if input.SpokeHeadType == "" {
		input.SpokeHeadType = spokeHeadTypeJBend
	}
	if input.SpokeHeadType != spokeHeadTypeJBend && input.SpokeHeadType != spokeHeadTypeStraightPull {
		return nil, ErrInvalidSpokeCalculation
	}
	if input.SpokeHoleDiameterMM != nil && (!isFinite(*input.SpokeHoleDiameterMM) || *input.SpokeHoleDiameterMM < 0 || *input.SpokeHoleDiameterMM > 10) {
		return nil, ErrInvalidSpokeCalculation
	}
	straightPullTangentOffsetMM := 0.0
	if input.StraightPullTangentOffsetMM != nil {
		if !isFinite(*input.StraightPullTangentOffsetMM) || math.Abs(*input.StraightPullTangentOffsetMM) > 20 {
			return nil, ErrInvalidSpokeCalculation
		}
		straightPullTangentOffsetMM = *input.StraightPullTangentOffsetMM
	}

	input.SpokeProfile = strings.ToLower(strings.TrimSpace(input.SpokeProfile))
	if input.SpokeProfile == "" {
		input.SpokeProfile = "round_2_0"
	}
	if _, exists := spokeProfileAreasMM2[input.SpokeProfile]; !exists {
		return nil, ErrInvalidSpokeCalculation
	}
	targetTensionN := 0.0
	if input.TargetTensionN != nil {
		if !isFinite(*input.TargetTensionN) || *input.TargetTensionN < 0 || *input.TargetTensionN > 3000 {
			return nil, ErrInvalidSpokeCalculation
		}
		targetTensionN = *input.TargetTensionN
	}
	alternatingDrillingOffsetMM := 0.0
	if input.AlternatingDrillingOffsetMM != nil {
		if !isFinite(*input.AlternatingDrillingOffsetMM) || math.Abs(*input.AlternatingDrillingOffsetMM) > 5 {
			return nil, ErrInvalidSpokeCalculation
		}
		alternatingDrillingOffsetMM = *input.AlternatingDrillingOffsetMM
	}
	if input.InterlaceCompensationMM != nil && (!isFinite(*input.InterlaceCompensationMM) || *input.InterlaceCompensationMM < 0 || *input.InterlaceCompensationMM > 5) {
		return nil, ErrInvalidSpokeCalculation
	}

	export, err := s.GetExport()
	if err != nil {
		return nil, err
	}
	rim := findSpokeRim(export, input.RimID)
	hub := findSpokeHub(export, input.HubID)
	if (rim == nil || hub == nil) && input.ERDMM == nil {
		return nil, ErrSpokeGeometryNotFound
	}
	if rim != nil && rim.ERD == nil && input.ERDMM == nil {
		return nil, ErrSpokeRimGeometryMissing
	}

	hubGeo := (*domainspoke.HubGeometry)(nil)
	if hub != nil {
		hubGeo = hub.Rear
		if input.WheelPosition == "front" {
			hubGeo = hub.Front
		}
	}
	if !isCompleteHubGeometry(hubGeo) {
		if input.LeftFlangeMM == nil || input.RightFlangeMM == nil || input.LeftFlangePCDMM == nil || input.RightFlangePCDMM == nil {
			return nil, ErrSpokeHubGeometryMissing
		}
		hubGeo = &domainspoke.HubGeometry{LeftFlange: input.LeftFlangeMM, RightFlange: input.RightFlangeMM, LeftFlangePCD: input.LeftFlangePCDMM, RightFlangePCD: input.RightFlangePCDMM}
	}
	erd := input.ERDMM
	if rim != nil && rim.ERD != nil {
		erd = rim.ERD
	}
	if erd == nil || !isFinite(*erd) || *erd < 250 || *erd > 800 {
		return nil, ErrInvalidSpokeCalculation
	}
	if !finiteGeometry(hubGeo) {
		return nil, ErrInvalidSpokeCalculation
	}

	leftFlange := effectiveSpokeFlangeDistance(*hubGeo.LeftFlange, input.RimOffsetMM, "left")
	rightFlange := effectiveSpokeFlangeDistance(*hubGeo.RightFlange, input.RimOffsetMM, "right")
	leftFlange += alternatingDrillingOffsetMM
	rightFlange -= alternatingDrillingOffsetMM
	if leftFlange <= 0 || rightFlange <= 0 {
		return nil, ErrInvalidSpokeCalculation
	}
	leftFlangeRadius := *hubGeo.LeftFlangePCD / 2.0
	rightFlangeRadius := *hubGeo.RightFlangePCD / 2.0
	radius := *erd / 2.0
	spokeHoleDiameterMM := input.SpokeHoleDiameterMM
	if spokeHoleDiameterMM == nil {
		spokeHoleDiameterMM = hubGeo.SpokeHoleDiameter
	}
	if spokeHoleDiameterMM == nil {
		defaultHoleDiameter := defaultSpokeHoleDiameterMM
		spokeHoleDiameterMM = &defaultHoleDiameter
	}

	// The current symmetric lacing path supplies one phase to both sides.
	// A future lacing-layout resolver (such as 2:1) can populate distinct
	// left/right phases without changing either head-type calculator.
	phaseRad := spokeLacingPhaseRadians(input.Crossing, input.SpokeCount)
	geometry, err := calculateSpokeGeometry(input.SpokeHeadType, spokeGeometryInput{
		Left: spokeGeometrySideInput{
			RimRadiusMM:      radius,
			FlangeRadiusMM:   leftFlangeRadius,
			FlangeDistanceMM: leftFlange,
			PhaseRad:         phaseRad,
		},
		Right: spokeGeometrySideInput{
			RimRadiusMM:      radius,
			FlangeRadiusMM:   rightFlangeRadius,
			FlangeDistanceMM: rightFlange,
			PhaseRad:         phaseRad,
		},
		SpokeHoleDiameterMM:         *spokeHoleDiameterMM,
		StraightPullTangentOffsetMM: straightPullTangentOffsetMM,
	})
	if err != nil {
		return nil, err
	}

	corrections, err := applySpokePhysicalCorrections(spokePhysicalCorrectionInput{
		LeftLengthMM:            geometry.LeftLengthMM,
		RightLengthMM:           geometry.RightLengthMM,
		NippleType:              input.NippleType,
		NippleLengthMM:          input.NippleLengthMM,
		Crossing:                input.Crossing,
		Interlacing:             input.Interlacing,
		InterlaceCompensationMM: input.InterlaceCompensationMM,
		SpokeProfile:            input.SpokeProfile,
		TargetTensionN:          targetTensionN,
	})
	if err != nil {
		return nil, err
	}
	left := corrections.LeftLengthMM
	right := corrections.RightLengthMM
	tensionRatio, err := computeSpokeTensionRatioSafe(leftFlange, rightFlange, left, right)
	if err != nil {
		return nil, err
	}
	if s.spokeRepo != nil {
		history := buildSpokeHistory(input, export, rim, hub, hubGeo, *erd, left, right)
		if err := s.spokeRepo.CreateHistory(history); err != nil {
			return nil, fmt.Errorf("persist spoke calculation history: %w", err)
		}
	}

	return &SpokeCalculationResult{
		LeftLengthMM:  roundSpokeLength(left),
		RightLengthMM: roundSpokeLength(right),
		TensionRatio:  tensionRatio,
		Debug: SpokeCalculationDebug{
			Rim:                         rim,
			Hub:                         hubGeo,
			RimOffsetMM:                 input.RimOffsetMM,
			SpokeHeadType:               input.SpokeHeadType,
			SpokeHoleCorrectionMM:       geometry.SpokeHoleCorrectionMM,
			StraightPullTangentOffsetMM: geometry.StraightPullTangentOffsetMM,
			AlternatingDrillingOffsetMM: alternatingDrillingOffsetMM,
			StretchLeftMM:               corrections.StretchLeftMM,
			StretchRightMM:              corrections.StretchRightMM,
			InterlaceCompensationMM:     corrections.InterlaceCompensationMM,
			FormulaVersion:              spokeCalculationFormulaName,
		},
	}, nil
}

func buildSpokeHistory(input SpokeCalculationInput, export domainspoke.ExportResponse, rim *domainspoke.RimModel, hub *domainspoke.HubModel, geometry *domainspoke.HubGeometry, erd, left, right float64) *domainspoke.History {
	history := &domainspoke.History{UserID: input.UserID, ERDMM: &erd, LeftFlangePCDMM: geometry.LeftFlangePCD, RightFlangePCDMM: geometry.RightFlangePCD, LeftFlangeToCenterMM: geometry.LeftFlange, RightFlangeToCenterMM: geometry.RightFlange, SpokeCount: &input.SpokeCount, LeftLengthMM: func() *float64 { v := roundSpokeLength(left); return &v }(), RightLengthMM: func() *float64 { v := roundSpokeLength(right); return &v }()}
	source := "calculator"
	history.SourceType = &source
	position := input.WheelPosition
	history.WheelType = &position
	nipple := input.NippleType
	if nipple != "" {
		history.NippleType = &nipple
	}
	crossing := fmt.Sprintf("%d-cross", input.Crossing)
	history.LacingPattern = &crossing
	if rim != nil {
		history.RimModel = &rim.Name
	}
	if hub != nil {
		history.HubModel = &hub.Name
	}
	if brand := spokeRimBrandName(export, input.RimID); brand != "" {
		history.RimBrand = &brand
	}
	if brand := spokeHubBrandName(export, input.HubID); brand != "" {
		history.HubBrand = &brand
	}
	return history
}

func spokeRimBrandName(export domainspoke.ExportResponse, modelID string) string {
	for _, brand := range export.Rims {
		for _, model := range brand.Items {
			if model.ID == modelID {
				return brand.Name
			}
		}
	}
	return ""
}

func spokeHubBrandName(export domainspoke.ExportResponse, modelID string) string {
	for _, brand := range export.Hubs {
		for _, model := range brand.Items {
			if model.ID == modelID {
				return brand.Name
			}
		}
	}
	return ""
}

func roundSpokeRatio(value float64) float64 {
	return math.Round(value*10000) / 10000
}

func normalizeSpokeCatalog(export domainspoke.ExportResponse, productBrands []productdomain.ProductBrand) (domainspoke.ExportResponse, error) {
	normalized := domainspoke.ExportResponse{
		Options: domainspoke.DefaultOptions(),
		Rims:    make([]domainspoke.RimBrand, 0, len(export.Rims)),
		Hubs:    make([]domainspoke.HubBrand, 0, len(export.Hubs)),
		Presets: make([]domainspoke.WheelBuildPreset, 0, len(export.Presets)),
	}

	rimBrandIDs := make(map[string]struct{})
	rimModelIDs := make(map[string]struct{})
	rimModelBrandIDs := make(map[string]string)
	hubBrandIDs := make(map[string]struct{})
	hubModelIDs := make(map[string]struct{})
	hubModelBrandIDs := make(map[string]string)
	productBrandIDs := make(map[string]productdomain.ProductBrand, len(productBrands)*2)
	allowedSpokeCounts := intOptionSet(normalized.Options.SpokeCounts)
	allowedCrossings := intOptionSet(normalized.Options.Crossings)
	allowedNippleTypes := stringOptionSet(normalized.Options.NippleTypes)
	allowedWheelPositions := stringOptionSet(normalized.Options.WheelPositions)
	for _, brand := range productBrands {
		brandID := strconv.FormatUint(uint64(brand.ID), 10)
		productBrandIDs[brandID] = brand
		productBrandIDs[normalizeCatalogID(brand.Slug)] = brand
	}

	for _, brand := range export.Rims {
		brandID, productBrand, exists := resolveProductBrandReference(brand.ID, productBrandIDs)
		if !exists {
			return domainspoke.ExportResponse{}, fmt.Errorf("%w: rim brand %q references an unknown product brand", ErrInvalidSpokeCatalog, brand.ID)
		}
		brand.ID = brandID
		brand.Name = productBrand.Name
		if err := validateCatalogID("rim brand id", brand.ID); err != nil {
			return domainspoke.ExportResponse{}, err
		}
		if _, exists := rimBrandIDs[brand.ID]; exists {
			return domainspoke.ExportResponse{}, fmt.Errorf("%w: duplicate rim brand id %q", ErrInvalidSpokeCatalog, brand.ID)
		}
		rimBrandIDs[brand.ID] = struct{}{}

		items := make([]domainspoke.RimModel, 0, len(brand.Items))
		for _, model := range brand.Items {
			model.ID = normalizeCatalogID(model.ID)
			model.Name = strings.TrimSpace(model.Name)
			if err := validateCatalogID("rim model id", model.ID); err != nil {
				return domainspoke.ExportResponse{}, err
			}
			if model.Name == "" {
				return domainspoke.ExportResponse{}, fmt.Errorf("%w: rim model %q is missing name", ErrInvalidSpokeCatalog, model.ID)
			}
			if _, exists := rimModelIDs[model.ID]; exists {
				return domainspoke.ExportResponse{}, fmt.Errorf("%w: duplicate rim model id %q", ErrInvalidSpokeCatalog, model.ID)
			}
			if model.ERD != nil && (*model.ERD < 250 || *model.ERD > 800) {
				return domainspoke.ExportResponse{}, fmt.Errorf("%w: rim model %q erd must be between 250 and 800mm", ErrInvalidSpokeCatalog, model.ID)
			}
			if model.Weight != nil && (*model.Weight < 0 || *model.Weight > 5000) {
				return domainspoke.ExportResponse{}, fmt.Errorf("%w: rim model %q weight is out of range", ErrInvalidSpokeCatalog, model.ID)
			}
			rimModelIDs[model.ID] = struct{}{}
			rimModelBrandIDs[model.ID] = brand.ID
			items = append(items, model)
		}
		brand.Items = items
		normalized.Rims = append(normalized.Rims, brand)
	}

	for _, brand := range export.Hubs {
		brand.ID = normalizeCatalogID(brand.ID)
		brand.Name = strings.TrimSpace(brand.Name)
		if err := validateCatalogID("hub brand id", brand.ID); err != nil {
			return domainspoke.ExportResponse{}, err
		}
		if brand.Name == "" {
			return domainspoke.ExportResponse{}, fmt.Errorf("%w: hub brand %q is missing name", ErrInvalidSpokeCatalog, brand.ID)
		}
		if _, exists := hubBrandIDs[brand.ID]; exists {
			return domainspoke.ExportResponse{}, fmt.Errorf("%w: duplicate hub brand id %q", ErrInvalidSpokeCatalog, brand.ID)
		}
		hubBrandIDs[brand.ID] = struct{}{}

		items := make([]domainspoke.HubModel, 0, len(brand.Items))
		for _, model := range brand.Items {
			model.ID = normalizeCatalogID(model.ID)
			model.Name = strings.TrimSpace(model.Name)
			if err := validateCatalogID("hub model id", model.ID); err != nil {
				return domainspoke.ExportResponse{}, err
			}
			if model.Name == "" {
				return domainspoke.ExportResponse{}, fmt.Errorf("%w: hub model %q is missing name", ErrInvalidSpokeCatalog, model.ID)
			}
			if _, exists := hubModelIDs[model.ID]; exists {
				return domainspoke.ExportResponse{}, fmt.Errorf("%w: duplicate hub model id %q", ErrInvalidSpokeCatalog, model.ID)
			}
			if err := validateHubGeometry(model.ID, "front", model.Front); err != nil {
				return domainspoke.ExportResponse{}, err
			}
			if err := validateHubGeometry(model.ID, "rear", model.Rear); err != nil {
				return domainspoke.ExportResponse{}, err
			}
			hubModelIDs[model.ID] = struct{}{}
			hubModelBrandIDs[model.ID] = brand.ID
			items = append(items, model)
		}
		brand.Items = items
		normalized.Hubs = append(normalized.Hubs, brand)
	}

	for _, preset := range export.Presets {
		preset.ID = normalizeCatalogID(preset.ID)
		preset.Name = strings.TrimSpace(preset.Name)
		preset.Description = strings.TrimSpace(preset.Description)
		preset.RimBrandID = normalizeCatalogID(preset.RimBrandID)
		preset.RimModelID = normalizeCatalogID(preset.RimModelID)
		preset.HubBrandID = normalizeCatalogID(preset.HubBrandID)
		preset.HubModelID = normalizeCatalogID(preset.HubModelID)
		preset.WheelPosition = normalizeWheelPosition(preset.WheelPosition)
		preset.NippleType = strings.ToLower(strings.TrimSpace(preset.NippleType))
		preset.Keywords = normalizeKeywords(preset.Keywords)
		preset.ActualLengths = normalizeActualLengths(preset.ActualLengths)

		if err := validateCatalogID("preset id", preset.ID); err != nil {
			return domainspoke.ExportResponse{}, err
		}
		if preset.Name == "" {
			return domainspoke.ExportResponse{}, fmt.Errorf("%w: preset %q is missing name", ErrInvalidSpokeCatalog, preset.ID)
		}
		if resolvedRimBrandID, _, exists := resolveProductBrandReference(preset.RimBrandID, productBrandIDs); exists {
			preset.RimBrandID = resolvedRimBrandID
		}
		if _, exists := rimBrandIDs[preset.RimBrandID]; !exists {
			return domainspoke.ExportResponse{}, fmt.Errorf("%w: preset %q references unknown rim brand %q", ErrInvalidSpokeCatalog, preset.ID, preset.RimBrandID)
		}
		if _, exists := rimModelIDs[preset.RimModelID]; !exists {
			return domainspoke.ExportResponse{}, fmt.Errorf("%w: preset %q references unknown rim model %q", ErrInvalidSpokeCatalog, preset.ID, preset.RimModelID)
		}
		if rimModelBrandIDs[preset.RimModelID] != preset.RimBrandID {
			return domainspoke.ExportResponse{}, fmt.Errorf("%w: preset %q rim model %q does not belong to rim brand %q", ErrInvalidSpokeCatalog, preset.ID, preset.RimModelID, preset.RimBrandID)
		}
		if _, exists := hubBrandIDs[preset.HubBrandID]; !exists {
			return domainspoke.ExportResponse{}, fmt.Errorf("%w: preset %q references unknown hub brand %q", ErrInvalidSpokeCatalog, preset.ID, preset.HubBrandID)
		}
		if _, exists := hubModelIDs[preset.HubModelID]; !exists {
			return domainspoke.ExportResponse{}, fmt.Errorf("%w: preset %q references unknown hub model %q", ErrInvalidSpokeCatalog, preset.ID, preset.HubModelID)
		}
		if hubModelBrandIDs[preset.HubModelID] != preset.HubBrandID {
			return domainspoke.ExportResponse{}, fmt.Errorf("%w: preset %q hub model %q does not belong to hub brand %q", ErrInvalidSpokeCatalog, preset.ID, preset.HubModelID, preset.HubBrandID)
		}
		if _, exists := allowedSpokeCounts[preset.SpokeCount]; !exists {
			return domainspoke.ExportResponse{}, fmt.Errorf("%w: preset %q spoke count must match a calculator option", ErrInvalidSpokeCatalog, preset.ID)
		}
		if _, exists := allowedCrossings[preset.Crossing]; !exists {
			return domainspoke.ExportResponse{}, fmt.Errorf("%w: preset %q crossing must match a calculator option", ErrInvalidSpokeCatalog, preset.ID)
		}
		if _, exists := allowedNippleTypes[preset.NippleType]; !exists {
			return domainspoke.ExportResponse{}, fmt.Errorf("%w: preset %q nipple type must match a calculator option", ErrInvalidSpokeCatalog, preset.ID)
		}
		if _, exists := allowedWheelPositions[preset.WheelPosition]; !exists {
			return domainspoke.ExportResponse{}, fmt.Errorf("%w: preset %q wheel position must match a calculator option", ErrInvalidSpokeCatalog, preset.ID)
		}
		if preset.NippleLength != nil && (*preset.NippleLength < 0 || *preset.NippleLength > 40) {
			return domainspoke.ExportResponse{}, fmt.Errorf("%w: preset %q nipple length is out of range", ErrInvalidSpokeCatalog, preset.ID)
		}
		if err := validateActualLengths(preset.ID, preset.ActualLengths); err != nil {
			return domainspoke.ExportResponse{}, err
		}
		normalized.Presets = append(normalized.Presets, preset)
	}

	return normalized, nil
}

func validateCatalogID(label, value string) error {
	if !spokeCatalogIDPattern.MatchString(value) {
		return fmt.Errorf("%w: %s %q must use lowercase letters, numbers, underscores or hyphens", ErrInvalidSpokeCatalog, label, value)
	}
	return nil
}

func validateHubGeometry(modelID, position string, geometry *domainspoke.HubGeometry) error {
	if geometry == nil {
		return nil
	}
	if !floatInRange(geometry.LeftFlange, 0, 100) ||
		!floatInRange(geometry.RightFlange, 0, 100) ||
		!floatInRange(geometry.LeftFlangePCD, 10, 150) ||
		!floatInRange(geometry.RightFlangePCD, 10, 150) {
		return fmt.Errorf("%w: hub model %q %s geometry is out of range", ErrInvalidSpokeCatalog, modelID, position)
	}
	if geometry.SpokeHoleDiameter != nil && (*geometry.SpokeHoleDiameter < 0 || *geometry.SpokeHoleDiameter > 10) {
		return fmt.Errorf("%w: hub model %q %s spoke hole diameter is out of range", ErrInvalidSpokeCatalog, modelID, position)
	}
	return nil
}

func isCompleteHubGeometry(geometry *domainspoke.HubGeometry) bool {
	return geometry != nil &&
		geometry.LeftFlange != nil &&
		geometry.RightFlange != nil &&
		geometry.LeftFlangePCD != nil &&
		geometry.RightFlangePCD != nil
}

func floatInRange(value *float64, minValue, maxValue float64) bool {
	return value == nil || (*value >= minValue && *value <= maxValue)
}

func normalizeCatalogID(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func resolveProductBrandReference(reference string, brands map[string]productdomain.ProductBrand) (string, productdomain.ProductBrand, bool) {
	normalized := normalizeCatalogID(reference)
	if normalized == "" {
		return "", productdomain.ProductBrand{}, false
	}
	if brand, exists := brands[normalized]; exists {
		return normalizeCatalogID(brand.Slug), brand, true
	}
	return "", productdomain.ProductBrand{}, false
}

func normalizeWheelPosition(value string) string {
	switch normalized := strings.ToLower(strings.TrimSpace(value)); normalized {
	case "front", "rear", "auto":
		return normalized
	case "":
		return "auto"
	default:
		return normalized
	}
}

func normalizeKeywords(values []string) []string {
	seen := make(map[string]struct{})
	keywords := make([]string, 0, len(values))
	for _, value := range values {
		keyword := strings.TrimSpace(value)
		if keyword == "" {
			continue
		}
		key := strings.ToLower(keyword)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		keywords = append(keywords, keyword)
	}
	return keywords
}

func normalizeActualLengths(actual *domainspoke.WheelBuildActualLengths) *domainspoke.WheelBuildActualLengths {
	if actual == nil {
		return nil
	}

	normalized := &domainspoke.WheelBuildActualLengths{
		FrontLeft:  actual.FrontLeft,
		FrontRight: actual.FrontRight,
		RearLeft:   actual.RearLeft,
		RearRight:  actual.RearRight,
		Notes:      strings.TrimSpace(actual.Notes),
	}
	if normalized.FrontLeft == nil &&
		normalized.FrontRight == nil &&
		normalized.RearLeft == nil &&
		normalized.RearRight == nil &&
		normalized.Notes == "" {
		return nil
	}
	return normalized
}

func validateActualLengths(presetID string, actual *domainspoke.WheelBuildActualLengths) error {
	if actual == nil {
		return nil
	}

	fields := []struct {
		label string
		value *float64
	}{
		{label: "front left", value: actual.FrontLeft},
		{label: "front right", value: actual.FrontRight},
		{label: "rear left", value: actual.RearLeft},
		{label: "rear right", value: actual.RearRight},
	}
	for _, field := range fields {
		if field.value != nil && (*field.value <= 0 || *field.value > 500) {
			return fmt.Errorf("%w: preset %q actual %s spoke length is out of range", ErrInvalidSpokeCatalog, presetID, field.label)
		}
	}
	return nil
}

func intOptionSet(options []domainspoke.IntOption) map[int]struct{} {
	result := make(map[int]struct{}, len(options))
	for _, option := range options {
		result[option.Value] = struct{}{}
	}
	return result
}

func stringOptionSet(options []domainspoke.StringOption) map[string]struct{} {
	result := make(map[string]struct{}, len(options))
	for _, option := range options {
		result[strings.ToLower(strings.TrimSpace(option.Value))] = struct{}{}
	}
	return result
}

func findSpokeRim(export domainspoke.ExportResponse, rimID string) *domainspoke.RimModel {
	for _, brand := range export.Rims {
		for _, rim := range brand.Items {
			if rim.ID == rimID {
				foundRim := rim
				return &foundRim
			}
		}
	}
	return nil
}

func findSpokeHub(export domainspoke.ExportResponse, hubID string) *domainspoke.HubModel {
	for _, brand := range export.Hubs {
		for _, hub := range brand.Items {
			if hub.ID == hubID {
				foundHub := hub
				return &foundHub
			}
		}
	}
	return nil
}
