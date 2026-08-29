package service

import (
	domainspoke "commerce-platform/internal/domain/spoke"
	"commerce-platform/internal/repository"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
)

var (
	ErrSpokeGeometryNotFound    = errors.New("unknown rim or hub geometry")
	ErrSpokeRimGeometryMissing  = errors.New("rim geometry not available")
	ErrSpokeHubGeometryMissing  = errors.New("hub geometry not available for requested position")
	ErrInvalidSpokeCalculation  = errors.New("invalid spoke calculation input")
	ErrInvalidSpokeCatalog      = errors.New("invalid spoke catalog")
	spokeCalculationFormulaName = "v1.1-go-backend"
	spokeCatalogIDPattern       = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,139}$`)
)

type SpokeService struct {
	spokeRepo *repository.SpokeRepository
}

type SpokeCalculationInput struct {
	RimID         string
	HubID         string
	WheelPosition string
	SpokeCount    int
	Crossing      int
	// RimOffsetMM is positive when the rim center moves toward the right
	// flange. The value changes both the spoke length geometry and bracing
	// angles used for the tension-ratio estimate.
	RimOffsetMM float64
}

type SpokeCalculationResult struct {
	LeftLengthMM  float64               `json:"leftLengthMm"`
	RightLengthMM float64               `json:"rightLengthMm"`
	TensionRatio  *SpokeTensionRatio    `json:"tensionRatio,omitempty"`
	Debug         SpokeCalculationDebug `json:"debug"`
}

type SpokeCalculationDebug struct {
	Rim            *domainspoke.RimModel    `json:"rim"`
	Hub            *domainspoke.HubGeometry `json:"hub"`
	RimOffsetMM    float64                  `json:"rimOffsetMm"`
	FormulaVersion string                   `json:"formulaVersion"`
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

	normalized, err := normalizeSpokeCatalog(export)
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
	options := domainspoke.DefaultOptions()
	if _, exists := intOptionSet(options.SpokeCounts)[input.SpokeCount]; !exists {
		return nil, ErrInvalidSpokeCalculation
	}
	if _, exists := intOptionSet(options.Crossings)[input.Crossing]; !exists {
		return nil, ErrInvalidSpokeCalculation
	}
	if math.Abs(input.RimOffsetMM) > 20 {
		return nil, ErrInvalidSpokeCalculation
	}

	export, err := s.GetExport()
	if err != nil {
		return nil, err
	}
	rim := findSpokeRim(export, input.RimID)
	hub := findSpokeHub(export, input.HubID)
	if rim == nil || hub == nil {
		return nil, ErrSpokeGeometryNotFound
	}
	if rim.ERD == nil {
		return nil, ErrSpokeRimGeometryMissing
	}

	hubGeo := hub.Rear
	if input.WheelPosition == "front" {
		hubGeo = hub.Front
	}
	if !isCompleteHubGeometry(hubGeo) {
		return nil, ErrSpokeHubGeometryMissing
	}

	leftFlange := effectiveSpokeFlangeDistance(*hubGeo.LeftFlange, input.RimOffsetMM, "left")
	rightFlange := effectiveSpokeFlangeDistance(*hubGeo.RightFlange, input.RimOffsetMM, "right")
	if leftFlange <= 0 || rightFlange <= 0 {
		return nil, ErrInvalidSpokeCalculation
	}
	leftFlangeRadius := *hubGeo.LeftFlangePCD / 2.0
	rightFlangeRadius := *hubGeo.RightFlangePCD / 2.0
	radius := *rim.ERD / 2.0
	angleRad := (720.0 * float64(input.Crossing) / float64(input.SpokeCount)) * math.Pi / 180.0

	left := math.Sqrt(radius*radius + leftFlangeRadius*leftFlangeRadius + leftFlange*leftFlange - 2*radius*leftFlangeRadius*math.Cos(angleRad))
	right := math.Sqrt(radius*radius + rightFlangeRadius*rightFlangeRadius + rightFlange*rightFlange - 2*radius*rightFlangeRadius*math.Cos(angleRad))
	tensionRatio := computeSpokeTensionRatio(leftFlange, rightFlange, left, right)

	return &SpokeCalculationResult{
		LeftLengthMM:  math.Round(left*10) / 10,
		RightLengthMM: math.Round(right*10) / 10,
		TensionRatio:  tensionRatio,
		Debug: SpokeCalculationDebug{
			Rim:            rim,
			Hub:            hubGeo,
			RimOffsetMM:    input.RimOffsetMM,
			FormulaVersion: spokeCalculationFormulaName,
		},
	}, nil
}

func effectiveSpokeFlangeDistance(flangeDistance, rimOffset float64, side string) float64 {
	if side == "left" {
		return flangeDistance + rimOffset
	}
	return flangeDistance - rimOffset
}

func computeSpokeTensionRatio(leftBracingDistance, rightBracingDistance, leftLength, rightLength float64) *SpokeTensionRatio {
	if leftBracingDistance <= 0 || rightBracingDistance <= 0 || leftLength <= 0 || rightLength <= 0 {
		return nil
	}

	leftSin := math.Min(1, leftBracingDistance/leftLength)
	rightSin := math.Min(1, rightBracingDistance/rightLength)
	if leftSin <= 0 || rightSin <= 0 {
		return nil
	}

	leftToRight := rightSin / leftSin
	rightToLeft := leftSin / rightSin
	lowerToHigher := math.Min(leftToRight, rightToLeft)

	lowerSide := "balanced"
	switch {
	case leftToRight < 0.995:
		lowerSide = "left"
	case leftToRight > 1.005:
		lowerSide = "right"
	}

	return &SpokeTensionRatio{
		LeftToRight:          roundSpokeRatio(leftToRight),
		RightToLeft:          roundSpokeRatio(rightToLeft),
		LowerToHigher:        roundSpokeRatio(lowerToHigher),
		LowerSide:            lowerSide,
		LeftBracingAngleDeg:  roundSpokeRatio(math.Asin(leftSin) * 180 / math.Pi),
		RightBracingAngleDeg: roundSpokeRatio(math.Asin(rightSin) * 180 / math.Pi),
	}
}

func roundSpokeRatio(value float64) float64 {
	return math.Round(value*10000) / 10000
}

func normalizeSpokeCatalog(export domainspoke.ExportResponse) (domainspoke.ExportResponse, error) {
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
	allowedSpokeCounts := intOptionSet(normalized.Options.SpokeCounts)
	allowedCrossings := intOptionSet(normalized.Options.Crossings)
	allowedNippleTypes := stringOptionSet(normalized.Options.NippleTypes)
	allowedWheelPositions := stringOptionSet(normalized.Options.WheelPositions)

	for _, brand := range export.Rims {
		brand.ID = normalizeCatalogID(brand.ID)
		brand.Name = strings.TrimSpace(brand.Name)
		if err := validateCatalogID("rim brand id", brand.ID); err != nil {
			return domainspoke.ExportResponse{}, err
		}
		if brand.Name == "" {
			return domainspoke.ExportResponse{}, fmt.Errorf("%w: rim brand %q is missing name", ErrInvalidSpokeCatalog, brand.ID)
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
