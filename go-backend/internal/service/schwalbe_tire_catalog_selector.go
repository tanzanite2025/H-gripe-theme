package service

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/repository"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

const schwalbeTireCatalogSelectorPageSize = 20

var schwalbeTireCatalogETRTOPattern = regexp.MustCompile(`^([0-9]+)-([0-9]+)$`)
var schwalbeTireCatalogInchDesignationPattern = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)\s*[xX×]`)

var schwalbeTireCatalogCasingConstructions = []string{
	"Super Race",
	"Super Ground",
	"Super Trail",
	"Super Downhill",
	"TRAIL",
	"TRAIL PRO",
	"GRAVITY",
	"GRAVITY PRO",
}

// SchwalbeTireCatalogSelectorQuery describes the public selector URL state.
// Filters within one dimension use OR; selected dimensions combine with AND.
type SchwalbeTireCatalogSelectorQuery struct {
	Search    string
	Page      int
	MinLoadKG *float64
	// IncludeRimWidthGuidance asks the selector to attach the source-backed
	// possible-combination range for each returned tire. It is separate from
	// InnerRimWidthMM so the page can explain a tire's reference range without
	// filtering the catalog by a user-entered rim width.
	IncludeRimWidthGuidance bool
	// InnerRimWidthMM enables source-backed possible-combination matching.
	// A nil value keeps the existing unfiltered catalog behavior.
	InnerRimWidthMM *float64
	ModelName       string
	// Nominal tire-width endpoints are inclusive. A nil endpoint leaves that
	// side of the range open. The exact-value slice remains for old links and
	// clients that still send tire_width_mm.
	NominalTireWidthMinMM *int
	NominalTireWidthMaxMM *int
	NominalTireWidthsMM   []int
	BeadSeatDiametersMM   []int
	WheelSizeKeys         []string
	CasingConstructions   []string
	RadialOnly            bool
	Beads                 []string
	Seals                 []string
	EBikeRatings          []*string
	Colors                []string
	Compounds             []string
	SortBy                string
}

type SchwalbeTireCatalogSelectorStringOption struct {
	Value string `json:"value"`
}

type SchwalbeTireCatalogSelectorNumberOption struct {
	Value int `json:"value"`
}

// SchwalbeTireCatalogSelectorWheelSizeOption keeps the user-facing wheel
// diameter paired with BSD. BSD alone is ambiguous for sizes such as 28" and
// 29" (both can use BSD 622), so the pair is the stable selector value.
type SchwalbeTireCatalogSelectorWheelSizeOption struct {
	Value              string `json:"value"`
	WheelDiameterIn    string `json:"wheel_diameter_in"`
	BeadSeatDiameterMM int    `json:"bsd_mm"`
}

type SchwalbeTireCatalogSelectorNullableStringOption struct {
	Value *string `json:"value"`
}

// SchwalbeTireCatalogSelectorFilterOptions are built from the text-search
// result before any selected facet is applied, so choosing one facet does not
// remove options from the other facets.
type SchwalbeTireCatalogSelectorFilterOptions struct {
	ModelNames          []SchwalbeTireCatalogSelectorStringOption         `json:"model_names"`
	NominalTireWidthsMM []SchwalbeTireCatalogSelectorNumberOption         `json:"nominal_tire_widths_mm"`
	BeadSeatDiametersMM []SchwalbeTireCatalogSelectorNumberOption         `json:"bead_seat_diameters_mm"`
	WheelSizes          []SchwalbeTireCatalogSelectorWheelSizeOption      `json:"wheel_sizes"`
	CasingConstructions []SchwalbeTireCatalogSelectorStringOption         `json:"casing_constructions"`
	Beads               []SchwalbeTireCatalogSelectorStringOption         `json:"beads"`
	Seals               []SchwalbeTireCatalogSelectorStringOption         `json:"seals"`
	EBikeRatings        []SchwalbeTireCatalogSelectorNullableStringOption `json:"e_bike_ratings"`
	Colors              []SchwalbeTireCatalogSelectorStringOption         `json:"colors"`
	Compounds           []SchwalbeTireCatalogSelectorStringOption         `json:"compounds"`
}

// SchwalbeTireCatalogSelectorPage contains only the requested page of rows,
// while totals and filter options are calculated from their full result sets.
type SchwalbeTireCatalogSelectorPage struct {
	Items           []repository.SchwalbeTireCatalogItem        `json:"items"`
	Page            int                                         `json:"page"`
	PageSize        int                                         `json:"page_size"`
	Total           int                                         `json:"total"`
	TotalPages      int                                         `json:"total_pages"`
	FilterOptions   SchwalbeTireCatalogSelectorFilterOptions    `json:"filter_options"`
	RimWidthContext *SchwalbeTireCatalogSelectorRimWidthContext `json:"rim_width_context,omitempty"`
	// RimWidthGuidanceByArticle is converted into item-level public fields by
	// the API projection. Keeping it off the page JSON prevents an internal map
	// shape from becoming a second public response contract.
	RimWidthGuidanceByArticle map[string][]SchwalbeTireCatalogSelectorRimWidthGuidance `json:"-"`
}

// SchwalbeTireCatalogSelectorRimWidthContext explains the source and state of
// an active inner-rim-width query. SourceURL is intentionally kept out of the
// selector response; the public rules endpoint owns that provenance detail.
type SchwalbeTireCatalogSelectorRimWidthContext struct {
	InnerRimWidthMM float64    `json:"inner_rim_width_mm"`
	GuidanceStatus  string     `json:"guidance_status"`
	SourceBasis     string     `json:"source_basis,omitempty"`
	SourceVersion   string     `json:"source_version,omitempty"`
	SourceCheckedAt *time.Time `json:"source_checked_at,omitempty"`
}

type schwalbeTireCatalogSelectorDimensions struct {
	ModelName           string
	NominalTireWidthMM  *int
	BeadSeatDiameterMM  *int
	WheelDiameterIn     string
	WheelSizeKey        string
	CasingConstructions map[string]struct{}
	IsRadial            bool
	Bead                string
	Seal                string
	EBikeRating         *string
	Color               string
	Compound            string
}

type schwalbeTireCatalogSelectorRow struct {
	item       repository.SchwalbeTireCatalogItem
	dimensions schwalbeTireCatalogSelectorDimensions
}

func (s *ProductService) SearchSchwalbeTireCatalogSelector(query SchwalbeTireCatalogSelectorQuery) (*SchwalbeTireCatalogSelectorPage, error) {
	items, err := s.productRepo.ListSchwalbeTireCatalog(query.Search)
	if err != nil {
		return nil, err
	}

	var rimWidthRules []repository.SchwalbeTireRimWidthCombinationRule
	var rimWidthContext *SchwalbeTireCatalogSelectorRimWidthContext
	if query.InnerRimWidthMM != nil || query.IncludeRimWidthGuidance {
		rimWidthRules, err = s.productRepo.ListSchwalbeTireRimWidthCombinationRules()
		if err != nil {
			return nil, err
		}
		rimWidthContext = buildSchwalbeTireCatalogSelectorRimWidthContext(query.InnerRimWidthMM, query.WheelSizeKeys, rimWidthRules)
	}

	rows := make([]schwalbeTireCatalogSelectorRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, schwalbeTireCatalogSelectorRow{
			item:       item,
			dimensions: deriveSchwalbeTireCatalogSelectorDimensions(item),
		})
	}

	collator := collate.New(language.English)
	filterOptions := buildSchwalbeTireCatalogSelectorFilterOptions(rows, collator)
	filteredRows := filterSchwalbeTireCatalogSelectorRows(rows, query, rimWidthRules)
	sortSchwalbeTireCatalogSelectorRows(filteredRows, query.SortBy)

	total := len(filteredRows)
	totalPages := (total + schwalbeTireCatalogSelectorPageSize - 1) / schwalbeTireCatalogSelectorPageSize
	if totalPages < 1 {
		totalPages = 1
	}
	page := query.Page
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * schwalbeTireCatalogSelectorPageSize
	end := start + schwalbeTireCatalogSelectorPageSize
	if end > total {
		end = total
	}
	pageItems := make([]repository.SchwalbeTireCatalogItem, 0, end-start)
	rimWidthGuidanceByArticle := make(map[string][]SchwalbeTireCatalogSelectorRimWidthGuidance)
	for _, row := range filteredRows[start:end] {
		pageItems = append(pageItems, row.item)
		if query.InnerRimWidthMM != nil || query.IncludeRimWidthGuidance {
			var guidance []SchwalbeTireCatalogSelectorRimWidthGuidance
			if query.IncludeRimWidthGuidance {
				guidance = schwalbeTireRimWidthGuidanceForTireWidth(row.dimensions.NominalTireWidthMM, rimWidthRules)
			} else {
				guidance = schwalbeTireRimWidthGuidanceFor(row.dimensions.NominalTireWidthMM, query.InnerRimWidthMM, rimWidthRules)
			}
			if len(guidance) > 0 {
				rimWidthGuidanceByArticle[row.item.ArticleNo] = guidance
			}
		}
	}

	return &SchwalbeTireCatalogSelectorPage{
		Items:                     pageItems,
		Page:                      page,
		PageSize:                  schwalbeTireCatalogSelectorPageSize,
		Total:                     total,
		TotalPages:                totalPages,
		FilterOptions:             filterOptions,
		RimWidthContext:           rimWidthContext,
		RimWidthGuidanceByArticle: rimWidthGuidanceByArticle,
	}, nil
}

func buildSchwalbeTireCatalogSelectorRimWidthContext(
	innerRimWidthMM *float64,
	wheelSizeKeys []string,
	rules []repository.SchwalbeTireRimWidthCombinationRule,
) *SchwalbeTireCatalogSelectorRimWidthContext {
	if innerRimWidthMM == nil {
		return nil
	}

	context := &SchwalbeTireCatalogSelectorRimWidthContext{
		InnerRimWidthMM: *innerRimWidthMM,
		GuidanceStatus:  "no_coverage",
	}
	if len(schwalbeTireSelectorStringSet(wheelSizeKeys)) == 0 {
		context.GuidanceStatus = "wheel_size_required"
	}
	if schwalbeTireRimWidthInputCovered(innerRimWidthMM, rules) {
		if context.GuidanceStatus != "wheel_size_required" {
			context.GuidanceStatus = "covered"
		}
	}
	if len(rules) == 0 {
		return context
	}

	context.SourceBasis = rules[0].SourceBasis
	context.SourceVersion = rules[0].SourceVersion
	checkedAt := rules[0].SourceCheckedAt
	if !checkedAt.IsZero() {
		context.SourceCheckedAt = &checkedAt
	}
	return context
}

func deriveSchwalbeTireCatalogSelectorDimensions(item repository.SchwalbeTireCatalogItem) schwalbeTireCatalogSelectorDimensions {
	dimensions := schwalbeTireCatalogSelectorDimensions{
		ModelName:           strings.TrimSpace(item.ModelName),
		CasingConstructions: make(map[string]struct{}),
		Bead:                normalizedSchwalbeTireSelectorString(item.Bead),
		Seal:                normalizedSchwalbeTireSelectorString(item.Seal),
		EBikeRating:         normalizedSchwalbeTireSelectorPointer(item.EBikeRating),
		Color:               normalizedSchwalbeTireSelectorString(item.Color),
		Compound:            normalizedSchwalbeTireSelectorString(item.Compound),
	}

	if match := schwalbeTireCatalogETRTOPattern.FindStringSubmatch(strings.TrimSpace(item.ETRTO)); len(match) == 3 {
		width, widthErr := strconv.ParseInt(match[1], 10, 64)
		diameter, diameterErr := strconv.ParseInt(match[2], 10, 64)
		if widthErr == nil && diameterErr == nil && width <= 9007199254740991 && diameter <= 9007199254740991 {
			widthValue := int(width)
			diameterValue := int(diameter)
			dimensions.NominalTireWidthMM = &widthValue
			dimensions.BeadSeatDiameterMM = &diameterValue
		}
	}

	if match := schwalbeTireCatalogInchDesignationPattern.FindStringSubmatch(normalizedSchwalbeTireSelectorString(item.InchDesignation)); len(match) == 2 {
		if wheelDiameter, ok := normalizeSchwalbeTireSelectorWheelDiameter(match[1]); ok {
			dimensions.WheelDiameterIn = wheelDiameter
			dimensions.WheelSizeKey = buildSchwalbeTireSelectorWheelSizeKey(
				wheelDiameter,
				dimensions.BeadSeatDiameterMM,
			)
		}
	}

	versionLabel := normalizedSchwalbeTireSelectorString(item.VersionLabel)
	for _, token := range strings.Split(versionLabel, ",") {
		token = strings.TrimSpace(token)
		if token == "Radial" {
			dimensions.IsRadial = true
		}
		for _, casing := range schwalbeTireCatalogCasingConstructions {
			if token == casing {
				dimensions.CasingConstructions[casing] = struct{}{}
			}
		}
	}
	return dimensions
}

func normalizeSchwalbeTireSelectorWheelDiameter(value string) (string, bool) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", false
	}
	parsed, err := strconv.ParseFloat(normalized, 64)
	if err != nil || parsed <= 0 {
		return "", false
	}
	return strconv.FormatFloat(parsed, 'f', -1, 64), true
}

func buildSchwalbeTireSelectorWheelSizeKey(wheelDiameter string, bsd *int) string {
	if strings.TrimSpace(wheelDiameter) == "" || bsd == nil || *bsd <= 0 {
		return ""
	}
	return wheelDiameter + "-" + strconv.Itoa(*bsd)
}

func normalizedSchwalbeTireSelectorPointer(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil
	}
	return &normalized
}

func normalizedSchwalbeTireSelectorString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func filterSchwalbeTireCatalogSelectorRows(
	rows []schwalbeTireCatalogSelectorRow,
	query SchwalbeTireCatalogSelectorQuery,
	rimWidthRules []repository.SchwalbeTireRimWidthCombinationRule,
) []schwalbeTireCatalogSelectorRow {
	filtered := make([]schwalbeTireCatalogSelectorRow, 0, len(rows))
	modelName := strings.TrimSpace(query.ModelName)
	widths := schwalbeTireSelectorIntegerSet(query.NominalTireWidthsMM)
	hasWidthRange := query.NominalTireWidthMinMM != nil || query.NominalTireWidthMaxMM != nil
	diameters := schwalbeTireSelectorIntegerSet(query.BeadSeatDiametersMM)
	wheelSizes := schwalbeTireSelectorStringSet(query.WheelSizeKeys)
	casings := schwalbeTireSelectorStringSet(query.CasingConstructions)
	beads := schwalbeTireSelectorStringSet(query.Beads)
	seals := schwalbeTireSelectorStringSet(query.Seals)
	colors := schwalbeTireSelectorStringSet(query.Colors)
	compounds := schwalbeTireSelectorStringSet(query.Compounds)
	selectedRatings := make(map[string]struct{}, len(query.EBikeRatings))
	selectUnrated := false
	for _, rating := range query.EBikeRatings {
		if rating == nil {
			selectUnrated = true
			continue
		}
		if normalized := strings.TrimSpace(*rating); normalized != "" {
			selectedRatings[normalized] = struct{}{}
		}
	}

	for _, row := range rows {
		dimensions := row.dimensions
		if query.InnerRimWidthMM != nil && len(wheelSizes) == 0 {
			// Inner-width matching is only meaningful with the user's wheel
			// diameter + BSD pair. Keep an incomplete shared URL from silently
			// producing a cross-wheel result set.
			continue
		}
		if query.MinLoadKG != nil {
			if row.item.LoadKG == nil || *row.item.LoadKG < *query.MinLoadKG {
				continue
			}
		}
		if query.InnerRimWidthMM != nil && len(schwalbeTireRimWidthGuidanceFor(dimensions.NominalTireWidthMM, query.InnerRimWidthMM, rimWidthRules)) == 0 {
			continue
		}
		if modelName != "" && dimensions.ModelName != modelName {
			continue
		}
		if hasWidthRange {
			if dimensions.NominalTireWidthMM == nil {
				continue
			}
			if query.NominalTireWidthMinMM != nil && *dimensions.NominalTireWidthMM < *query.NominalTireWidthMinMM {
				continue
			}
			if query.NominalTireWidthMaxMM != nil && *dimensions.NominalTireWidthMM > *query.NominalTireWidthMaxMM {
				continue
			}
		} else if !schwalbeTireSelectorMatchesIntegerSet(widths, dimensions.NominalTireWidthMM) {
			continue
		}
		if !schwalbeTireSelectorMatchesIntegerSet(diameters, dimensions.BeadSeatDiameterMM) {
			continue
		}
		if !schwalbeTireSelectorMatchesStringSet(wheelSizes, dimensions.WheelSizeKey) {
			continue
		}
		if len(casings) > 0 && !schwalbeTireSelectorMatchesTokenSet(casings, dimensions.CasingConstructions) {
			continue
		}
		if query.RadialOnly && !dimensions.IsRadial {
			continue
		}
		if !schwalbeTireSelectorMatchesStringSet(beads, dimensions.Bead) {
			continue
		}
		if !schwalbeTireSelectorMatchesStringSet(seals, dimensions.Seal) {
			continue
		}
		if len(selectedRatings) > 0 || selectUnrated {
			if dimensions.EBikeRating == nil {
				if !selectUnrated {
					continue
				}
			} else if _, ok := selectedRatings[*dimensions.EBikeRating]; !ok {
				continue
			}
		}
		if !schwalbeTireSelectorMatchesStringSet(colors, dimensions.Color) {
			continue
		}
		if !schwalbeTireSelectorMatchesStringSet(compounds, dimensions.Compound) {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func schwalbeTireSelectorIntegerSet(values []int) map[int]struct{} {
	selected := make(map[int]struct{}, len(values))
	for _, value := range values {
		if value > 0 {
			selected[value] = struct{}{}
		}
	}
	return selected
}

func schwalbeTireSelectorStringSet(values []string) map[string]struct{} {
	selected := make(map[string]struct{}, len(values))
	for _, value := range values {
		if normalized := strings.TrimSpace(value); normalized != "" {
			selected[normalized] = struct{}{}
		}
	}
	return selected
}

func schwalbeTireSelectorMatchesIntegerSet(selected map[int]struct{}, actual *int) bool {
	if len(selected) == 0 {
		return true
	}
	if actual == nil {
		return false
	}
	_, ok := selected[*actual]
	return ok
}

func schwalbeTireSelectorMatchesStringSet(selected map[string]struct{}, actual string) bool {
	if len(selected) == 0 {
		return true
	}
	_, ok := selected[actual]
	return actual != "" && ok
}

func schwalbeTireSelectorMatchesTokenSet(selected, actual map[string]struct{}) bool {
	for value := range selected {
		if _, ok := actual[value]; ok {
			return true
		}
	}
	return false
}

func sortSchwalbeTireCatalogSelectorRows(rows []schwalbeTireCatalogSelectorRow, sortBy string) {
	switch strings.TrimSpace(sortBy) {
	case "", "weight":
		sortBy = "weight_asc"
	case "weight_asc", "weight_desc", "model", "etrto", "article":
		// Keep the explicitly requested legacy API orders while the selector
		// uses the weight orders above.
	default:
		sortBy = "weight_asc"
	}
	textCollator := collate.New(language.English)
	numericCollator := collate.New(language.English, collate.Numeric)
	sort.SliceStable(rows, func(leftIndex, rightIndex int) bool {
		left := rows[leftIndex].item
		right := rows[rightIndex].item
		if sortBy == "weight_asc" || sortBy == "weight_desc" || sortBy == "weight" {
			if comparison := compareSchwalbeTireSelectorWeights(left.WeightG, right.WeightG, sortBy == "weight_desc"); comparison != 0 {
				return comparison < 0
			}
		}
		switch sortBy {
		case "etrto":
			if comparison := numericCollator.CompareString(left.ETRTO, right.ETRTO); comparison != 0 {
				return comparison < 0
			}
			return textCollator.CompareString(left.ArticleNo, right.ArticleNo) < 0
		case "article":
			return numericCollator.CompareString(left.ArticleNo, right.ArticleNo) < 0
		default:
			if comparison := textCollator.CompareString(left.ModelName, right.ModelName); comparison != 0 {
				return comparison < 0
			}
			if comparison := numericCollator.CompareString(left.ETRTO, right.ETRTO); comparison != 0 {
				return comparison < 0
			}
			return textCollator.CompareString(left.ArticleNo, right.ArticleNo) < 0
		}
	})
}

// compareSchwalbeTireSelectorWeights sorts known weights before unknown
// values. Missing source weights stay at the end in either direction, while
// equal weights use the caller's deterministic catalog tie-breakers.
func compareSchwalbeTireSelectorWeights(left, right *float64, descending bool) int {
	if left == nil && right == nil {
		return 0
	}
	if left == nil {
		return 1
	}
	if right == nil {
		return -1
	}
	if *left == *right {
		return 0
	}
	if descending {
		if *left > *right {
			return -1
		}
		return 1
	}
	if *left < *right {
		return -1
	}
	return 1
}

func buildSchwalbeTireCatalogSelectorFilterOptions(rows []schwalbeTireCatalogSelectorRow, collator *collate.Collator) SchwalbeTireCatalogSelectorFilterOptions {
	modelNames := make(map[string]struct{})
	widths := make(map[int]struct{})
	diameters := make(map[int]struct{})
	wheelSizes := make(map[string]SchwalbeTireCatalogSelectorWheelSizeOption)
	casings := make(map[string]struct{})
	beads := make(map[string]struct{})
	seals := make(map[string]struct{})
	ratings := make(map[string]struct{})
	colors := make(map[string]struct{})
	compounds := make(map[string]struct{})
	includeUnrated := false

	for _, row := range rows {
		dimensions := row.dimensions
		if dimensions.ModelName != "" {
			modelNames[dimensions.ModelName] = struct{}{}
		}
		if dimensions.NominalTireWidthMM != nil {
			widths[*dimensions.NominalTireWidthMM] = struct{}{}
		}
		if dimensions.BeadSeatDiameterMM != nil {
			diameters[*dimensions.BeadSeatDiameterMM] = struct{}{}
		}
		if dimensions.WheelSizeKey != "" && dimensions.BeadSeatDiameterMM != nil {
			wheelSizes[dimensions.WheelSizeKey] = SchwalbeTireCatalogSelectorWheelSizeOption{
				Value:              dimensions.WheelSizeKey,
				WheelDiameterIn:    dimensions.WheelDiameterIn,
				BeadSeatDiameterMM: *dimensions.BeadSeatDiameterMM,
			}
		}
		for value := range dimensions.CasingConstructions {
			casings[value] = struct{}{}
		}
		addSchwalbeSelectorOptionString(beads, dimensions.Bead)
		addSchwalbeSelectorOptionString(seals, dimensions.Seal)
		if dimensions.EBikeRating == nil {
			includeUnrated = true
		} else {
			ratings[*dimensions.EBikeRating] = struct{}{}
		}
		addSchwalbeSelectorOptionString(colors, dimensions.Color)
		addSchwalbeSelectorOptionString(compounds, dimensions.Compound)
	}

	return SchwalbeTireCatalogSelectorFilterOptions{
		ModelNames:          schwalbeTireSelectorStringOptions(modelNames, collator),
		NominalTireWidthsMM: schwalbeTireSelectorNumberOptions(widths),
		BeadSeatDiametersMM: schwalbeTireSelectorNumberOptions(diameters),
		WheelSizes:          schwalbeTireSelectorWheelSizeOptions(wheelSizes),
		CasingConstructions: schwalbeTireSelectorCasingOptions(casings),
		Beads:               schwalbeTireSelectorStringOptions(beads, collator),
		Seals:               schwalbeTireSelectorStringOptions(seals, collator),
		EBikeRatings:        schwalbeTireSelectorRatingOptions(ratings, includeUnrated, collator),
		Colors:              schwalbeTireSelectorStringOptions(colors, collator),
		Compounds:           schwalbeTireSelectorStringOptions(compounds, collator),
	}
}

func schwalbeTireSelectorWheelSizeOptions(values map[string]SchwalbeTireCatalogSelectorWheelSizeOption) []SchwalbeTireCatalogSelectorWheelSizeOption {
	options := make([]SchwalbeTireCatalogSelectorWheelSizeOption, 0, len(values))
	for _, option := range values {
		options = append(options, option)
	}
	sort.Slice(options, func(leftIndex, rightIndex int) bool {
		left := options[leftIndex]
		right := options[rightIndex]
		leftDiameter, leftErr := strconv.ParseFloat(left.WheelDiameterIn, 64)
		rightDiameter, rightErr := strconv.ParseFloat(right.WheelDiameterIn, 64)
		if leftErr == nil && rightErr == nil && leftDiameter != rightDiameter {
			return leftDiameter < rightDiameter
		}
		if left.WheelDiameterIn != right.WheelDiameterIn {
			return left.WheelDiameterIn < right.WheelDiameterIn
		}
		if left.BeadSeatDiameterMM != right.BeadSeatDiameterMM {
			return left.BeadSeatDiameterMM < right.BeadSeatDiameterMM
		}
		return left.Value < right.Value
	})
	return options
}

func addSchwalbeSelectorOptionString(values map[string]struct{}, value string) {
	if value != "" {
		values[value] = struct{}{}
	}
}

func schwalbeTireSelectorStringOptions(values map[string]struct{}, collator *collate.Collator) []SchwalbeTireCatalogSelectorStringOption {
	labels := make([]string, 0, len(values))
	for value := range values {
		labels = append(labels, value)
	}
	sort.Slice(labels, func(leftIndex, rightIndex int) bool {
		return collator.CompareString(labels[leftIndex], labels[rightIndex]) < 0
	})
	options := make([]SchwalbeTireCatalogSelectorStringOption, 0, len(labels))
	for _, label := range labels {
		options = append(options, SchwalbeTireCatalogSelectorStringOption{Value: label})
	}
	return options
}

func schwalbeTireSelectorNumberOptions(values map[int]struct{}) []SchwalbeTireCatalogSelectorNumberOption {
	labels := make([]int, 0, len(values))
	for value := range values {
		labels = append(labels, value)
	}
	sort.Ints(labels)
	options := make([]SchwalbeTireCatalogSelectorNumberOption, 0, len(labels))
	for _, label := range labels {
		options = append(options, SchwalbeTireCatalogSelectorNumberOption{Value: label})
	}
	return options
}

func schwalbeTireSelectorCasingOptions(values map[string]struct{}) []SchwalbeTireCatalogSelectorStringOption {
	options := make([]SchwalbeTireCatalogSelectorStringOption, 0, len(values))
	for _, label := range schwalbeTireCatalogCasingConstructions {
		if _, ok := values[label]; ok {
			options = append(options, SchwalbeTireCatalogSelectorStringOption{Value: label})
		}
	}
	return options
}

func schwalbeTireSelectorRatingOptions(values map[string]struct{}, includeUnrated bool, collator *collate.Collator) []SchwalbeTireCatalogSelectorNullableStringOption {
	labels := make([]string, 0, len(values))
	for value := range values {
		labels = append(labels, value)
	}
	sort.Slice(labels, func(leftIndex, rightIndex int) bool {
		return collator.CompareString(labels[leftIndex], labels[rightIndex]) < 0
	})
	options := make([]SchwalbeTireCatalogSelectorNullableStringOption, 0, len(labels)+1)
	for _, label := range labels {
		value := label
		options = append(options, SchwalbeTireCatalogSelectorNullableStringOption{Value: &value})
	}
	if includeUnrated {
		options = append(options, SchwalbeTireCatalogSelectorNullableStringOption{Value: nil})
	}
	return options
}
