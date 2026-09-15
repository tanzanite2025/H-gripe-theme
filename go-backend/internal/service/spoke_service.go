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
	spokeCalculationFormulaName = "v1.2-go-backend-hidden-nipple-safe"
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
	// RimOffsetMM is positive when the rim center moves toward the right
	// flange. The value changes both the spoke length geometry and bracing
	// angles used for the tension-ratio estimate.
	RimOffsetMM float64
	// Optional user-entered geometry. Catalog geometry is preferred whenever
	// RimID/HubID resolve to an authoritative record.
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
	Rim            *domainspoke.RimModel    `json:"-"`
	Hub            *domainspoke.HubGeometry `json:"-"`
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
	if leftFlange <= 0 || rightFlange <= 0 {
		return nil, ErrInvalidSpokeCalculation
	}
	leftFlangeRadius := *hubGeo.LeftFlangePCD / 2.0
	rightFlangeRadius := *hubGeo.RightFlangePCD / 2.0
	radius := *erd / 2.0
	angleRad := (720.0 * float64(input.Crossing) / float64(input.SpokeCount)) * math.Pi / 180.0

	leftSquared := radius*radius + leftFlangeRadius*leftFlangeRadius + leftFlange*leftFlange - 2*radius*leftFlangeRadius*math.Cos(angleRad)
	rightSquared := radius*radius + rightFlangeRadius*rightFlangeRadius + rightFlange*rightFlange - 2*radius*rightFlangeRadius*math.Cos(angleRad)
	if !isFinite(leftSquared) || !isFinite(rightSquared) || leftSquared < 0 || rightSquared < 0 {
		return nil, fmt.Errorf("%w: spoke geometry produced an invalid triangle", ErrInvalidSpokeCalculation)
	}
	left := math.Sqrt(leftSquared)
	right := math.Sqrt(rightSquared)
	if !isFinite(left) || !isFinite(right) || left <= 0 || right <= 0 {
		return nil, fmt.Errorf("%w: spoke length calculation diverged", ErrInvalidSpokeCalculation)
	}
	if input.NippleType == "hidden" && input.NippleLengthMM != nil {
		correction := *input.NippleLengthMM - 3
		if correction < 0 {
			return nil, fmt.Errorf("%w: hidden nipple length must be at least 3mm", ErrInvalidSpokeCalculation)
		}
		left += correction
		right += correction
	}
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

// computeSpokeTensionRatio is retained for package-level compatibility. The
// service path uses the error-returning variant so invalid values fail loudly.
func computeSpokeTensionRatio(leftBracingDistance, rightBracingDistance, leftLength, rightLength float64) *SpokeTensionRatio {
	result, _ := computeSpokeTensionRatioSafe(leftBracingDistance, rightBracingDistance, leftLength, rightLength)
	return result
}

func computeSpokeTensionRatioSafe(leftBracingDistance, rightBracingDistance, leftLength, rightLength float64) (*SpokeTensionRatio, error) {
	if leftBracingDistance <= 0 || rightBracingDistance <= 0 || leftLength <= 0 || rightLength <= 0 {
		return nil, fmt.Errorf("%w: bracing geometry must be positive", ErrInvalidSpokeCalculation)
	}
	if !isFinite(leftBracingDistance) || !isFinite(rightBracingDistance) || !isFinite(leftLength) || !isFinite(rightLength) {
		return nil, fmt.Errorf("%w: tension ratio calculation received a non-finite value", ErrInvalidSpokeCalculation)
	}

	leftSin := math.Min(1, leftBracingDistance/leftLength)
	rightSin := math.Min(1, rightBracingDistance/rightLength)
	if leftSin <= 0 || rightSin <= 0 || !isFinite(leftSin) || !isFinite(rightSin) {
		return nil, fmt.Errorf("%w: tension ratio calculation diverged", ErrInvalidSpokeCalculation)
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

	result := &SpokeTensionRatio{
		LeftToRight:          roundSpokeRatio(leftToRight),
		RightToLeft:          roundSpokeRatio(rightToLeft),
		LowerToHigher:        roundSpokeRatio(lowerToHigher),
		LowerSide:            lowerSide,
		LeftBracingAngleDeg:  roundSpokeRatio(math.Asin(leftSin) * 180 / math.Pi),
		RightBracingAngleDeg: roundSpokeRatio(math.Asin(rightSin) * 180 / math.Pi),
	}
	if !isFinite(result.LeftToRight) || !isFinite(result.RightToLeft) || !isFinite(result.LowerToHigher) {
		return nil, fmt.Errorf("%w: tension ratio calculation diverged", ErrInvalidSpokeCalculation)
	}
	return result, nil
}

func roundSpokeLength(value float64) float64 { return math.Round(value*100) / 100 }

func isFinite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func finiteGeometry(geometry *domainspoke.HubGeometry) bool {
	if geometry == nil || geometry.LeftFlange == nil || geometry.RightFlange == nil || geometry.LeftFlangePCD == nil || geometry.RightFlangePCD == nil {
		return false
	}
	return isFinite(*geometry.LeftFlange) && isFinite(*geometry.RightFlange) &&
		isFinite(*geometry.LeftFlangePCD) && isFinite(*geometry.RightFlangePCD) &&
		*geometry.LeftFlange > 0 && *geometry.LeftFlange <= 100 &&
		*geometry.RightFlange > 0 && *geometry.RightFlange <= 100 &&
		*geometry.LeftFlangePCD >= 10 && *geometry.LeftFlangePCD <= 150 &&
		*geometry.RightFlangePCD >= 10 && *geometry.RightFlangePCD <= 150
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
