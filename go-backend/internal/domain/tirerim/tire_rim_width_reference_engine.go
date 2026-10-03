package tirerim

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)

const (
	TireRimWidthReferenceModelVersion  = "dt-swiss-tss-tc-v1"
	TireRimWidthReferenceKnowledgeAsOf = "2026-09-28"
	MinInputTireWidthMM                = 18
	MaxInputTireWidthMM                = 127
	ModelVersion                       = TireRimWidthReferenceModelVersion
	KnowledgeAsOf                      = TireRimWidthReferenceKnowledgeAsOf
)

var (
	ErrInvalidRimSystem   = errors.New("invalid rim system")
	ErrInvalidTireWidth   = errors.New("invalid tire width")
	ErrNoPublishedBracket = errors.New("no published tire-width bracket")
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

type MatrixRow struct {
	RimSystem   RimSystem    `json:"rim_system"`
	TireWidthMM int          `json:"tire_width_mm"`
	Inch        string       `json:"inch"`
	Recommended []WidthRange `json:"recommended"`
	Possible    []WidthRange `json:"possible"`
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

type DerivedMetrics struct {
	Status                 string       `json:"status"`
	InflatedTireWidthMM    NumericRange `json:"inflated_tire_width_mm"`
	AeroTargetOuterWidthMM NumericRange `json:"aero_target_outer_width_mm"`
}

type Suggestion struct {
	TireWidthMM    int            `json:"tire_width_mm"`
	RimSystem      RimSystem      `json:"rim_system"`
	ResultKind     string         `json:"result_kind"`
	RimWidthRanges []WidthRange   `json:"rim_width_ranges"`
	SourceRows     []SourceRow    `json:"source_rows"`
	Calculation    *Calculation   `json:"calculation,omitempty"`
	DerivedMetrics DerivedMetrics `json:"derived_metrics"`
	ModelVersion   string         `json:"model_version"`
	KnowledgeAsOf  string         `json:"knowledge_as_of"`
	Limitations    []string       `json:"limitations"`
}

type Metadata struct {
	ModelVersion  string      `json:"model_version"`
	KnowledgeAsOf string      `json:"knowledge_as_of"`
	Source        Source      `json:"source"`
	Rows          []MatrixRow `json:"rows"`
	Methodology   Methodology `json:"methodology"`
	Limitations   []string    `json:"limitations"`
}

type Source struct {
	Name       string `json:"name"`
	Provenance string `json:"provenance"`
}

type Methodology struct {
	ExactRows        string `json:"exact_rows"`
	InterpolatedRows string `json:"interpolated_rows"`
	PossibleRows     string `json:"possible_rows"`
	DerivedMetrics   string `json:"derived_metrics"`
}

var tireRimWidthReferenceLimitations = []string{
	"This is a DT Swiss reference chart result, not a certification for a specific tire, rim model, or wheelset.",
	"Rows marked possible are used only as a reference anchor when that tire width has no recommended row; they are not relabeled as recommended.",
	"Interpolated widths are mathematical projections between adjacent chart rows and must be checked against the exact tire and rim manufacturer specifications.",
	"The inflated-width and 105% outer-width values are display-only projections, not measured tire dimensions or safety limits.",
}

// This is the DT Swiss transcription used by the calculator. Schwalbe's broad
// ETRTO possibility matrix is intentionally absent.
var tireRimWidthReferenceMatrixByRimSystem = map[RimSystem][]MatrixRow{
	RimSystemHookless: {
		tireRimWidthReferenceMatrixRow(30, "1.20", nil, tireRimWidthReferenceRanges(23, 25)),
		tireRimWidthReferenceMatrixRow(32, "1.25", tireRimWidthReferenceRanges(23, 25), nil),
		tireRimWidthReferenceMatrixRow(34, "1.35", tireRimWidthReferenceRanges(23, 25), nil),
		tireRimWidthReferenceMatrixRow(36, "1.40", tireRimWidthReferenceRanges(23, 25), tireRimWidthReferenceRanges(26, 27)),
		tireRimWidthReferenceMatrixRow(38, "1.50", tireRimWidthReferenceRanges(23, 25), tireRimWidthReferenceRanges(26, 27)),
		tireRimWidthReferenceMatrixRow(41, "1.60", tireRimWidthReferenceRanges(23, 25), tireRimWidthReferenceRanges(26, 27)),
		tireRimWidthReferenceMatrixRow(43, "1.70", tireRimWidthReferenceRanges(23, 25, 26, 27), nil),
		tireRimWidthReferenceMatrixRow(47, "1.85", tireRimWidthReferenceRanges(23, 25, 26, 27), tireRimWidthReferenceRanges(28, 30)),
		tireRimWidthReferenceMatrixRow(50, "2.00", tireRimWidthReferenceRanges(23, 25, 26, 27), tireRimWidthReferenceRanges(28, 30)),
		tireRimWidthReferenceMatrixRow(52, "2.00", tireRimWidthReferenceRanges(26, 27), tireRimWidthReferenceRanges(23, 25, 28, 30)),
		tireRimWidthReferenceMatrixRow(53, "2.10", tireRimWidthReferenceRanges(26, 27, 28, 30), tireRimWidthReferenceRanges(23, 25)),
		tireRimWidthReferenceMatrixRow(56, "2.20", tireRimWidthReferenceRanges(26, 27, 28, 30), tireRimWidthReferenceRanges(23, 25)),
		tireRimWidthReferenceMatrixRow(60, "2.40", tireRimWidthReferenceRanges(26, 27, 28, 30), tireRimWidthReferenceRanges(23, 25, 31, 35)),
		tireRimWidthReferenceMatrixRow(64, "2.50", tireRimWidthReferenceRanges(28, 30, 31, 35), tireRimWidthReferenceRanges(23, 25, 26, 27, 36, 40)),
		tireRimWidthReferenceMatrixRow(66, "2.60", tireRimWidthReferenceRanges(28, 30, 31, 35), tireRimWidthReferenceRanges(23, 25, 26, 27, 36, 40)),
		tireRimWidthReferenceMatrixRow(69, "2.70", tireRimWidthReferenceRanges(28, 30, 31, 35), tireRimWidthReferenceRanges(26, 27, 36, 40)),
		tireRimWidthReferenceMatrixRow(71, "2.80", tireRimWidthReferenceRanges(31, 35, 36, 40), tireRimWidthReferenceRanges(26, 27, 28, 30)),
		tireRimWidthReferenceMatrixRow(75, "3.00", tireRimWidthReferenceRanges(36, 40), tireRimWidthReferenceRanges(31, 35)),
		tireRimWidthReferenceMatrixRow(85, "3.30", tireRimWidthReferenceRanges(36, 40), tireRimWidthReferenceRanges(31, 35)),
		tireRimWidthReferenceMatrixRow(96, "3.70", nil, tireRimWidthReferenceRanges(76, 76)),
		tireRimWidthReferenceMatrixRow(102, "4.00", tireRimWidthReferenceRanges(76, 76), nil),
		tireRimWidthReferenceMatrixRow(110, "4.30", tireRimWidthReferenceRanges(76, 76), nil),
		tireRimWidthReferenceMatrixRow(115, "4.50", tireRimWidthReferenceRanges(76, 76), nil),
		tireRimWidthReferenceMatrixRow(122, "4.80", tireRimWidthReferenceRanges(76, 76), nil),
		tireRimWidthReferenceMatrixRow(127, "5.00", tireRimWidthReferenceRanges(76, 76), nil),
	},
	RimSystemHooked: {
		tireRimWidthReferenceMatrixRow(20, "0.80", nil, tireRimWidthReferenceRanges(15, 17)),
		tireRimWidthReferenceMatrixRow(23, "0.90", tireRimWidthReferenceRanges(15, 17), tireRimWidthReferenceRanges(18, 20)),
		tireRimWidthReferenceMatrixRow(25, "1.00", tireRimWidthReferenceRanges(15, 17, 18, 20), tireRimWidthReferenceRanges(21, 22)),
		tireRimWidthReferenceMatrixRow(26, "1.00", tireRimWidthReferenceRanges(18, 20, 21, 22), tireRimWidthReferenceRanges(15, 17)),
		tireRimWidthReferenceMatrixRow(28, "1.10", tireRimWidthReferenceRanges(18, 20, 21, 22), tireRimWidthReferenceRanges(15, 17)),
		tireRimWidthReferenceMatrixRow(29, "1.15", tireRimWidthReferenceRanges(18, 20, 21, 22), tireRimWidthReferenceRanges(15, 17, 23, 25)),
		tireRimWidthReferenceMatrixRow(30, "1.20", tireRimWidthReferenceRanges(18, 20, 21, 22), tireRimWidthReferenceRanges(15, 17, 23, 25)),
		tireRimWidthReferenceMatrixRow(32, "1.25", tireRimWidthReferenceRanges(21, 22, 23, 25), tireRimWidthReferenceRanges(15, 17, 18, 20)),
		tireRimWidthReferenceMatrixRow(34, "1.35", tireRimWidthReferenceRanges(21, 22, 23, 25), tireRimWidthReferenceRanges(15, 17, 18, 20)),
		tireRimWidthReferenceMatrixRow(35, "1.40", tireRimWidthReferenceRanges(21, 22, 23, 25), tireRimWidthReferenceRanges(15, 17, 18, 20, 26, 27)),
		tireRimWidthReferenceMatrixRow(36, "1.40", tireRimWidthReferenceRanges(23, 25), tireRimWidthReferenceRanges(18, 20, 21, 22, 26, 27)),
		tireRimWidthReferenceMatrixRow(38, "1.50", tireRimWidthReferenceRanges(23, 25), tireRimWidthReferenceRanges(18, 20, 21, 22, 26, 27)),
		tireRimWidthReferenceMatrixRow(41, "1.60", tireRimWidthReferenceRanges(23, 25), tireRimWidthReferenceRanges(18, 20, 21, 22, 26, 27)),
		tireRimWidthReferenceMatrixRow(43, "1.70", tireRimWidthReferenceRanges(23, 25, 26, 27), tireRimWidthReferenceRanges(18, 20, 21, 22)),
		tireRimWidthReferenceMatrixRow(47, "1.85", tireRimWidthReferenceRanges(23, 25, 26, 27), tireRimWidthReferenceRanges(18, 20, 21, 22, 28, 30)),
		tireRimWidthReferenceMatrixRow(50, "2.00", tireRimWidthReferenceRanges(23, 25, 26, 27), tireRimWidthReferenceRanges(18, 20, 21, 22, 28, 30)),
		tireRimWidthReferenceMatrixRow(52, "2.00", tireRimWidthReferenceRanges(26, 27), tireRimWidthReferenceRanges(18, 20, 21, 22, 23, 25, 28, 30)),
		tireRimWidthReferenceMatrixRow(53, "2.10", tireRimWidthReferenceRanges(26, 27, 28, 30), tireRimWidthReferenceRanges(18, 20, 21, 22, 23, 25)),
		tireRimWidthReferenceMatrixRow(56, "2.20", tireRimWidthReferenceRanges(26, 27, 28, 30), tireRimWidthReferenceRanges(18, 20, 21, 22, 23, 25)),
		tireRimWidthReferenceMatrixRow(60, "2.40", tireRimWidthReferenceRanges(26, 27, 28, 30), tireRimWidthReferenceRanges(21, 22, 23, 25, 31, 35)),
		tireRimWidthReferenceMatrixRow(64, "2.50", tireRimWidthReferenceRanges(28, 30, 31, 35), tireRimWidthReferenceRanges(21, 22, 23, 25, 26, 27)),
		tireRimWidthReferenceMatrixRow(66, "2.60", tireRimWidthReferenceRanges(28, 30, 31, 35), tireRimWidthReferenceRanges(23, 25, 26, 27, 36, 40)),
		tireRimWidthReferenceMatrixRow(69, "2.70", tireRimWidthReferenceRanges(28, 30, 31, 35), tireRimWidthReferenceRanges(26, 27, 36, 40)),
		tireRimWidthReferenceMatrixRow(71, "2.80", tireRimWidthReferenceRanges(31, 35, 36, 40), tireRimWidthReferenceRanges(26, 27, 28, 30)),
		tireRimWidthReferenceMatrixRow(75, "3.00", tireRimWidthReferenceRanges(36, 40), tireRimWidthReferenceRanges(31, 35)),
		tireRimWidthReferenceMatrixRow(85, "3.30", tireRimWidthReferenceRanges(36, 40), tireRimWidthReferenceRanges(31, 35)),
		tireRimWidthReferenceMatrixRow(96, "3.70", nil, tireRimWidthReferenceRanges(76, 76)),
		tireRimWidthReferenceMatrixRow(102, "4.00", tireRimWidthReferenceRanges(76, 76), nil),
		tireRimWidthReferenceMatrixRow(110, "4.30", tireRimWidthReferenceRanges(76, 76), nil),
		tireRimWidthReferenceMatrixRow(115, "4.50", tireRimWidthReferenceRanges(76, 76), nil),
		tireRimWidthReferenceMatrixRow(122, "4.80", tireRimWidthReferenceRanges(76, 76), nil),
		tireRimWidthReferenceMatrixRow(127, "5.00", tireRimWidthReferenceRanges(76, 76), nil),
	},
}

func tireRimWidthReferenceMatrixRow(tireWidthMM int, inch string, recommended, possible []WidthRange) MatrixRow {
	return MatrixRow{
		TireWidthMM: tireWidthMM, Inch: inch,
		Recommended: cloneTireRimWidthRanges(recommended),
		Possible:    cloneTireRimWidthRanges(possible),
	}
}

func tireRimWidthReferenceRanges(values ...int) []WidthRange {
	if len(values)%2 != 0 {
		panic("tire/rim width range values must be pairs")
	}
	ranges := make([]WidthRange, 0, len(values)/2)
	for index := 0; index < len(values); index += 2 {
		ranges = append(ranges, WidthRange{Min: values[index], Max: values[index+1]})
	}
	return ranges
}

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

func GetTireRimWidthReferenceMatrix(system RimSystem) ([]MatrixRow, error) {
	if err := ValidateRimSystem(system); err != nil {
		return nil, err
	}
	return cloneTireRimWidthReferenceRowsWithRimSystem(tireRimWidthReferenceMatrixByRimSystem[system], system), nil
}

func GetTireRimWidthReferenceMetadata() Metadata {
	return Metadata{
		ModelVersion:  TireRimWidthReferenceModelVersion,
		KnowledgeAsOf: TireRimWidthReferenceKnowledgeAsOf,
		Source: Source{
			Name:       "DT Swiss TSS/TC tire-width and rim-inner-width reference chart",
			Provenance: "Repository transcription of the published chart; not a model-specific certification",
		},
		Rows: appendTireRimWidthReferenceRowsWithRimSystems(
			tireRimWidthReferenceMatrixByRimSystem[RimSystemHookless], RimSystemHookless,
			tireRimWidthReferenceMatrixByRimSystem[RimSystemHooked], RimSystemHooked,
		),
		Methodology: Methodology{
			ExactRows:        "Use the published recommended ranges without changing them.",
			InterpolatedRows: "Use piecewise linear interpolation between the nearest chart rows; no arbitrary span cutoff is applied.",
			PossibleRows:     "Keep possible-only rows labeled as possible_reference; never promote them to recommended.",
			DerivedMetrics:   "Inflated width = nominal width + 0.4 × (rim inner width − 19); 105% target is display-only.",
		},
		Limitations: cloneTireRimWidthReferenceStrings(tireRimWidthReferenceLimitations),
	}
}

func ResolveTireRimWidthReference(width int, system RimSystem) (*Suggestion, error) {
	if err := ValidateTireWidth(width); err != nil {
		return nil, err
	}
	if err := ValidateRimSystem(system); err != nil {
		return nil, err
	}
	rows := tireRimWidthReferenceMatrixByRimSystem[system]
	for _, row := range rows {
		if row.TireWidthMM != width {
			continue
		}
		if len(row.Recommended) > 0 {
			return newTireRimWidthReferenceSuggestion(width, system, "exact_recommended", row.Recommended, []SourceRow{{TireWidthMM: width, Kind: "recommended"}}, nil), nil
		}
		if len(row.Possible) > 0 {
			return newTireRimWidthReferenceSuggestion(width, system, "possible_reference", row.Possible, []SourceRow{{TireWidthMM: width, Kind: "possible_reference"}}, nil), nil
		}
	}

	type interpolationAnchor struct {
		row    MatrixRow
		bounds WidthRange
		kind   string
	}
	anchors := make([]interpolationAnchor, 0, len(rows))
	for _, row := range rows {
		ranges := row.Recommended
		kind := "recommended"
		if len(ranges) == 0 {
			ranges = row.Possible
			kind = "possible_reference"
		}
		if len(ranges) == 0 {
			continue
		}
		anchors = append(anchors, interpolationAnchor{
			row: row, bounds: calculateTireRimWidthRangeBounds(ranges), kind: kind,
		})
	}
	sort.Slice(anchors, func(leftIndex, rightIndex int) bool {
		return anchors[leftIndex].row.TireWidthMM < anchors[rightIndex].row.TireWidthMM
	})
	var lowerAnchor, upperAnchor *interpolationAnchor
	for index := range anchors {
		candidate := &anchors[index]
		if candidate.row.TireWidthMM < width {
			lowerAnchor = candidate
			continue
		}
		if candidate.row.TireWidthMM > width {
			upperAnchor = candidate
			break
		}
	}
	if lowerAnchor == nil || upperAnchor == nil {
		return nil, fmt.Errorf("%w: %d mm for %s", ErrNoPublishedBracket, width, system)
	}
	ratio := float64(width-lowerAnchor.row.TireWidthMM) / float64(upperAnchor.row.TireWidthMM-lowerAnchor.row.TireWidthMM)
	interpolatedRange := WidthRange{
		Min: int(math.Round(float64(lowerAnchor.bounds.Min) + ratio*float64(upperAnchor.bounds.Min-lowerAnchor.bounds.Min))),
		Max: int(math.Round(float64(lowerAnchor.bounds.Max) + ratio*float64(upperAnchor.bounds.Max-lowerAnchor.bounds.Max))),
	}
	if interpolatedRange.Max < interpolatedRange.Min {
		interpolatedRange.Min, interpolatedRange.Max = interpolatedRange.Max, interpolatedRange.Min
	}
	return newTireRimWidthReferenceSuggestion(width, system, "interpolated", []WidthRange{interpolatedRange}, []SourceRow{
		{TireWidthMM: lowerAnchor.row.TireWidthMM, Kind: lowerAnchor.kind},
		{TireWidthMM: upperAnchor.row.TireWidthMM, Kind: upperAnchor.kind},
	}, &Calculation{
		Method: "piecewise_linear", LowerTireWidthMM: lowerAnchor.row.TireWidthMM, UpperTireWidthMM: upperAnchor.row.TireWidthMM,
		InterpolationRatio: roundTireRimWidthReferenceMetric(ratio, 4),
	}), nil
}

func BuildTireRimWidthGuidanceForNominalWidth(width int, systems []RimSystem) []Suggestion {
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

func newTireRimWidthReferenceSuggestion(width int, system RimSystem, resultKind string, ranges []WidthRange, sourceRows []SourceRow, calculation *Calculation) *Suggestion {
	clonedRanges := cloneTireRimWidthRanges(ranges)
	rangeBounds := calculateTireRimWidthRangeBounds(clonedRanges)
	inflatedWidth := NumericRange{
		Min: roundTireRimWidthReferenceMetric(float64(width)+0.4*float64(rangeBounds.Min-19), 1),
		Max: roundTireRimWidthReferenceMetric(float64(width)+0.4*float64(rangeBounds.Max-19), 1),
	}
	return &Suggestion{
		TireWidthMM: width, RimSystem: system, ResultKind: resultKind, RimWidthRanges: clonedRanges,
		SourceRows: append([]SourceRow(nil), sourceRows...), Calculation: calculation,
		DerivedMetrics: DerivedMetrics{
			Status: "display_only", InflatedTireWidthMM: inflatedWidth,
			AeroTargetOuterWidthMM: NumericRange{
				Min: roundTireRimWidthReferenceMetric(inflatedWidth.Min*1.05, 1),
				Max: roundTireRimWidthReferenceMetric(inflatedWidth.Max*1.05, 1),
			},
		},
		ModelVersion: TireRimWidthReferenceModelVersion, KnowledgeAsOf: TireRimWidthReferenceKnowledgeAsOf,
		Limitations: cloneTireRimWidthReferenceStrings(tireRimWidthReferenceLimitations),
	}
}

func calculateTireRimWidthRangeBounds(ranges []WidthRange) WidthRange {
	if len(ranges) == 0 {
		return WidthRange{}
	}
	minimum, maximum := ranges[0].Min, ranges[0].Max
	if minimum > maximum {
		minimum, maximum = maximum, minimum
	}
	result := WidthRange{Min: minimum, Max: maximum}
	for _, item := range ranges[1:] {
		itemMinimum, itemMaximum := item.Min, item.Max
		if itemMinimum > itemMaximum {
			itemMinimum, itemMaximum = itemMaximum, itemMinimum
		}
		if itemMinimum < result.Min {
			result.Min = itemMinimum
		}
		if itemMaximum > result.Max {
			result.Max = itemMaximum
		}
	}
	return result
}

func cloneTireRimWidthRanges(values []WidthRange) []WidthRange {
	if len(values) == 0 {
		return []WidthRange{}
	}
	result := make([]WidthRange, len(values))
	copy(result, values)
	return result
}

func cloneTireRimWidthReferenceRows(values []MatrixRow) []MatrixRow {
	result := make([]MatrixRow, len(values))
	for index, row := range values {
		result[index] = row
		result[index].Recommended = cloneTireRimWidthRanges(row.Recommended)
		result[index].Possible = cloneTireRimWidthRanges(row.Possible)
	}
	return result
}

func cloneTireRimWidthReferenceRowsWithRimSystem(values []MatrixRow, system RimSystem) []MatrixRow {
	result := cloneTireRimWidthReferenceRows(values)
	for index := range result {
		result[index].RimSystem = system
	}
	return result
}

func appendTireRimWidthReferenceRows(groups ...[]MatrixRow) []MatrixRow {
	var result []MatrixRow
	for _, group := range groups {
		result = append(result, cloneTireRimWidthReferenceRows(group)...)
	}
	return result
}

func appendTireRimWidthReferenceRowsWithRimSystems(
	firstRows []MatrixRow,
	firstSystem RimSystem,
	secondRows []MatrixRow,
	secondSystem RimSystem,
) []MatrixRow {
	result := cloneTireRimWidthReferenceRowsWithRimSystem(firstRows, firstSystem)
	result = append(result, cloneTireRimWidthReferenceRowsWithRimSystem(secondRows, secondSystem)...)
	return result
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
