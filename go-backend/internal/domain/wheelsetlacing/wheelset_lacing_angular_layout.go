package wheelsetlacing

import (
	"fmt"
	"math"
	"strings"
)

// WheelsetLacingAngularLayoutRequest contains the continuous inputs needed to
// resolve a registered topology. The two optional G3 radii derive its automatic
// parallel spacing; the resulting layout remains coordinate-free so both the
// reference drawing and spoke-length calculator can use the same hole angles.
type WheelsetLacingAngularLayoutRequest struct {
	TopologyID                           string
	G3RimHoleCircleRadiusMM              float64
	G3SideAFlangeHoleCircleRadiusMM      float64
	G3RimHoleSpacingAToBDegrees          float64
	G3RimHoleSpacingBToADegrees          float64
	G3RimHoleSpacingAToNextGroupADegrees float64
}

// WheelsetLacingAngularPoint is a coordinate-free polar position. AngleRadians
// is the canonical drilling phase; a consumer chooses its own radial scale.
type WheelsetLacingAngularPoint struct {
	ID           int
	Side         Side
	AngleRadians float64
}

// WheelsetLacingAngularSpoke combines the immutable topology mapping with the
// two canonical angular endpoints. RelativeAngleRadians is the only lacing
// phase the physical spoke-length calculation needs.
type WheelsetLacingAngularSpoke struct {
	ID                   int
	Side                 Side
	Type                 SpokeType
	HubHoleID            int
	RimHoleID            int
	Hub                  WheelsetLacingAngularPoint
	Rim                  WheelsetLacingAngularPoint
	RelativeAngleRadians float64
}

// WheelsetLacingAngularLayout is the shared result consumed by display and
// physical calculation layers. No field in this type is an SVG coordinate.
type WheelsetLacingAngularLayout struct {
	Topology  Topology
	RimHoles  []WheelsetLacingAngularPoint
	HubHolesA []WheelsetLacingAngularPoint
	HubHolesB []WheelsetLacingAngularPoint
	Spokes    []WheelsetLacingAngularSpoke
	G3Spacing WheelsetLacingG3GroupSpacing
}

// WheelsetLacingG3GroupSpacing is the validated angular spacing profile for a
// G3 A-B-A rim-hole group. The third gap closes one seven-group pitch.
type WheelsetLacingG3GroupSpacing struct {
	Enabled                     bool
	GroupCount                  int
	GroupPitchDegrees           float64
	SpacingAToBDegrees          float64
	SpacingBToADegrees          float64
	SpacingAToNextGroupADegrees float64
}

// CalculateWheelsetLacingAngularLayout resolves registered hole mappings into
// canonical angles. It is intentionally independent from canvas projection and
// physical dimensions so both consumers use exactly the same lacing phases.
func CalculateWheelsetLacingAngularLayout(
	request WheelsetLacingAngularLayoutRequest,
	topology Topology,
) (WheelsetLacingAngularLayout, error) {
	if strings.TrimSpace(request.TopologyID) == "" {
		return WheelsetLacingAngularLayout{}, fmt.Errorf("%w: topology_id is required", ErrInvalidRequest)
	}
	if strings.TrimSpace(request.TopologyID) != topology.ID {
		return WheelsetLacingAngularLayout{}, fmt.Errorf("%w: topology_id %q does not match topology %q", ErrTopologyMismatch, request.TopologyID, topology.ID)
	}
	if err := validateTopology(topology); err != nil {
		return WheelsetLacingAngularLayout{}, fmt.Errorf("%w: angular layout topology is invalid: %v", ErrInvalidTopology, err)
	}
	if err := validateWheelsetLacingAngularLayoutRequest(request); err != nil {
		return WheelsetLacingAngularLayout{}, err
	}

	points, g3Spacing, err := buildWheelsetLacingAngularPoints(topology, request)
	if err != nil {
		return WheelsetLacingAngularLayout{}, err
	}

	rimByID := make(map[int]WheelsetLacingAngularPoint, len(points.rimHoles))
	for _, point := range points.rimHoles {
		rimByID[point.ID] = point
	}
	hubByKey := make(map[string]WheelsetLacingAngularPoint, len(points.hubHolesA)+len(points.hubHolesB))
	for _, point := range points.hubHolesA {
		hubByKey[wheelsetLacingAngularHubKey(SideA, point.ID)] = point
	}
	for _, point := range points.hubHolesB {
		hubByKey[wheelsetLacingAngularHubKey(SideB, point.ID)] = point
	}

	spokes := make([]WheelsetLacingAngularSpoke, 0, len(topology.Spokes))
	for _, mapping := range topology.Spokes {
		hub, hubOK := hubByKey[wheelsetLacingAngularHubKey(mapping.Side, mapping.HubHoleID)]
		rim, rimOK := rimByID[mapping.RimHoleID]
		if !hubOK || !rimOK {
			return WheelsetLacingAngularLayout{}, fmt.Errorf("%w: spoke %d references a missing angular point", ErrInvalidTopology, mapping.ID)
		}
		spokes = append(spokes, WheelsetLacingAngularSpoke{
			ID:                   mapping.ID,
			Side:                 mapping.Side,
			Type:                 mapping.Type,
			HubHoleID:            mapping.HubHoleID,
			RimHoleID:            mapping.RimHoleID,
			Hub:                  hub,
			Rim:                  rim,
			RelativeAngleRadians: normalizeWheelsetLacingAngularDifference(rim.AngleRadians - hub.AngleRadians),
		})
	}

	return WheelsetLacingAngularLayout{
		Topology:  cloneTopology(topology),
		RimHoles:  points.rimHoles,
		HubHolesA: points.hubHolesA,
		HubHolesB: points.hubHolesB,
		Spokes:    spokes,
		G3Spacing: g3Spacing,
	}, nil
}

type wheelsetLacingAngularPointSet struct {
	rimHoles  []WheelsetLacingAngularPoint
	hubHolesA []WheelsetLacingAngularPoint
	hubHolesB []WheelsetLacingAngularPoint
}

func validateWheelsetLacingAngularLayoutRequest(request WheelsetLacingAngularLayoutRequest) error {
	for field, value := range map[string]float64{
		"g3_rim_hole_spacing_a_to_b_degrees":            request.G3RimHoleSpacingAToBDegrees,
		"g3_rim_hole_spacing_b_to_a_degrees":            request.G3RimHoleSpacingBToADegrees,
		"g3_rim_hole_spacing_a_to_next_group_a_degrees": request.G3RimHoleSpacingAToNextGroupADegrees,
	} {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > WheelsetLacingMaximumTwentyOneHoleG3RimHoleSpacingDegrees {
			return fmt.Errorf("%w: %s must be finite and within [0, %g] degrees", ErrInvalidRequest, field, WheelsetLacingMaximumTwentyOneHoleG3RimHoleSpacingDegrees)
		}
	}
	if (request.G3RimHoleCircleRadiusMM == 0) != (request.G3SideAFlangeHoleCircleRadiusMM == 0) {
		return fmt.Errorf("%w: G3 rim and side-A flange hole-circle radii must be supplied together", ErrInvalidRequest)
	}
	for field, value := range map[string]float64{
		"g3_rim_hole_circle_radius_mm":           request.G3RimHoleCircleRadiusMM,
		"g3_side_a_flange_hole_circle_radius_mm": request.G3SideAFlangeHoleCircleRadiusMM,
	} {
		if value != 0 && (!isFinitePositiveWheelsetLacingNumber(value)) {
			return fmt.Errorf("%w: %s must be finite and positive", ErrInvalidRequest, field)
		}
	}
	return nil
}

func buildWheelsetLacingAngularPoints(
	topology Topology,
	request WheelsetLacingAngularLayoutRequest,
) (wheelsetLacingAngularPointSet, WheelsetLacingG3GroupSpacing, error) {
	switch topology.DisplayLayout {
	case DisplayGeometryLayoutG3TwentyOneHoleTriplet2To1:
		return buildWheelsetLacingTwentyOneHoleG3AngularPoints(topology, request)
	case DisplayGeometryLayoutUniform2To1:
		return buildWheelsetLacingUniformTwoToOneAngularPoints(topology, 24, 16, 8), WheelsetLacingG3GroupSpacing{}, nil
	case DisplayGeometryLayoutUniform18H2To1:
		return buildWheelsetLacingUniformTwoToOneAngularPoints(topology, 18, 12, 6), WheelsetLacingG3GroupSpacing{}, nil
	case DisplayGeometryLayoutSymmetric1To1:
		return buildWheelsetLacingSymmetricAngularPoints(topology), WheelsetLacingG3GroupSpacing{}, nil
	default:
		return wheelsetLacingAngularPointSet{}, WheelsetLacingG3GroupSpacing{}, fmt.Errorf("%w: unsupported angular layout %q", ErrInvalidTopology, topology.DisplayLayout)
	}
}

func buildWheelsetLacingSymmetricAngularPoints(topology Topology) wheelsetLacingAngularPointSet {
	holeCount := topology.HoleCount
	rimHoles := make([]WheelsetLacingAngularPoint, 0, len(topology.RimHoles))
	rimBySide := map[Side][]WheelsetLacingAngularPoint{SideA: {}, SideB: {}}
	for index, hole := range topology.RimHoles {
		point := WheelsetLacingAngularPoint{
			ID:           hole.ID,
			Side:         hole.Side,
			AngleRadians: (float64(index) * 2 * math.Pi / float64(holeCount)) - math.Pi/2 + math.Pi/float64(holeCount),
		}
		rimHoles = append(rimHoles, point)
		rimBySide[hole.Side] = append(rimBySide[hole.Side], point)
	}
	hubHolesA := make([]WheelsetLacingAngularPoint, 0, len(topology.HubHolesA))
	hubHolesB := make([]WheelsetLacingAngularPoint, 0, len(topology.HubHolesB))
	for index, hole := range topology.HubHolesA {
		hubHolesA = append(hubHolesA, WheelsetLacingAngularPoint{ID: hole.ID, Side: SideA, AngleRadians: rimBySide[SideA][index].AngleRadians})
	}
	for index, hole := range topology.HubHolesB {
		hubHolesB = append(hubHolesB, WheelsetLacingAngularPoint{ID: hole.ID, Side: SideB, AngleRadians: rimBySide[SideB][index].AngleRadians})
	}
	return wheelsetLacingAngularPointSet{rimHoles: rimHoles, hubHolesA: hubHolesA, hubHolesB: hubHolesB}
}

func buildWheelsetLacingTwentyOneHoleG3AngularPoints(
	topology Topology,
	request WheelsetLacingAngularLayoutRequest,
) (wheelsetLacingAngularPointSet, WheelsetLacingG3GroupSpacing, error) {
	if topology.HoleCount != 21 || len(topology.RimHoles) != 21 || len(topology.HubHolesA) != 14 || len(topology.HubHolesB) != 7 {
		return wheelsetLacingAngularPointSet{}, WheelsetLacingG3GroupSpacing{}, fmt.Errorf("%w: G3 angular layout requires a 21-hole 14/7 topology", ErrInvalidTopology)
	}
	spacingAToB, spacingBToA, spacingAToNextGroupA, err := resolveWheelsetLacingG3RimHoleSpacingValues(
		request.G3RimHoleSpacingAToBDegrees,
		request.G3RimHoleSpacingBToADegrees,
		request.G3RimHoleSpacingAToNextGroupADegrees,
		request.G3RimHoleCircleRadiusMM,
		request.G3SideAFlangeHoleCircleRadiusMM,
	)
	if err != nil {
		return wheelsetLacingAngularPointSet{}, WheelsetLacingG3GroupSpacing{}, err
	}
	if err := validateWheelsetLacingG3RimHoleSpacingClosure(spacingAToB, spacingBToA, spacingAToNextGroupA); err != nil {
		return wheelsetLacingAngularPointSet{}, WheelsetLacingG3GroupSpacing{}, err
	}

	spacingAToBRadians := spacingAToB * math.Pi / 180
	spacingBToARadians := spacingBToA * math.Pi / 180
	rimHoles := make([]WheelsetLacingAngularPoint, 0, len(topology.RimHoles))
	for group := 0; group < WheelsetLacingTwentyOneHoleG3GroupCount; group++ {
		centerAngle := (float64(group) * 2 * math.Pi / float64(WheelsetLacingTwentyOneHoleG3GroupCount)) - math.Pi/2
		rimHoles = append(rimHoles,
			WheelsetLacingAngularPoint{ID: group * 3, Side: SideA, AngleRadians: centerAngle - spacingAToBRadians},
			WheelsetLacingAngularPoint{ID: group*3 + 1, Side: SideB, AngleRadians: centerAngle},
			WheelsetLacingAngularPoint{ID: group*3 + 2, Side: SideA, AngleRadians: centerAngle + spacingBToARadians},
		)
	}
	hubHolesA := make([]WheelsetLacingAngularPoint, 0, len(topology.HubHolesA))
	hubHolesB := make([]WheelsetLacingAngularPoint, 0, len(topology.HubHolesB))
	for group := 0; group < WheelsetLacingTwentyOneHoleG3GroupCount; group++ {
		centerAngle := (float64(group) * 2 * math.Pi / float64(WheelsetLacingTwentyOneHoleG3GroupCount)) - math.Pi/2
		midAngle := centerAngle + math.Pi/float64(WheelsetLacingTwentyOneHoleG3GroupCount)
		hubHolesA = append(hubHolesA,
			WheelsetLacingAngularPoint{ID: group * 2, Side: SideA, AngleRadians: midAngle - math.Pi/14},
			WheelsetLacingAngularPoint{ID: group*2 + 1, Side: SideA, AngleRadians: midAngle + math.Pi/14},
		)
		hubHolesB = append(hubHolesB, WheelsetLacingAngularPoint{ID: group, Side: SideB, AngleRadians: centerAngle})
	}
	return wheelsetLacingAngularPointSet{rimHoles: rimHoles, hubHolesA: hubHolesA, hubHolesB: hubHolesB}, WheelsetLacingG3GroupSpacing{
		Enabled:                     true,
		GroupCount:                  WheelsetLacingTwentyOneHoleG3GroupCount,
		GroupPitchDegrees:           WheelsetLacingTwentyOneHoleG3GroupPitchDegrees,
		SpacingAToBDegrees:          spacingAToB,
		SpacingBToADegrees:          spacingBToA,
		SpacingAToNextGroupADegrees: spacingAToNextGroupA,
	}, nil
}

func buildWheelsetLacingUniformTwoToOneAngularPoints(
	topology Topology,
	totalHoleCount int,
	driveSideHoleCount int,
	nonDriveSideHoleCount int,
) wheelsetLacingAngularPointSet {
	rimHoles := make([]WheelsetLacingAngularPoint, 0, len(topology.RimHoles))
	rimBySide := map[Side][]WheelsetLacingAngularPoint{SideA: {}, SideB: {}}
	for index, hole := range topology.RimHoles {
		point := WheelsetLacingAngularPoint{
			ID:           hole.ID,
			Side:         hole.Side,
			AngleRadians: (float64(index) * 2 * math.Pi / float64(totalHoleCount)) - math.Pi/2 + math.Pi/float64(totalHoleCount),
		}
		rimHoles = append(rimHoles, point)
		rimBySide[hole.Side] = append(rimBySide[hole.Side], point)
	}
	hubHolesA := make([]WheelsetLacingAngularPoint, 0, len(topology.HubHolesA))
	for index, hole := range topology.HubHolesA {
		hubHolesA = append(hubHolesA, WheelsetLacingAngularPoint{
			ID:           hole.ID,
			Side:         SideA,
			AngleRadians: (float64(index) * 2 * math.Pi / float64(driveSideHoleCount)) - math.Pi/2,
		})
	}
	hubHolesB := make([]WheelsetLacingAngularPoint, 0, len(topology.HubHolesB))
	for index, hole := range topology.HubHolesB {
		if index >= len(rimBySide[SideB]) {
			return wheelsetLacingAngularPointSet{}
		}
		hubHolesB = append(hubHolesB, WheelsetLacingAngularPoint{
			ID:           hole.ID,
			Side:         SideB,
			AngleRadians: rimBySide[SideB][index].AngleRadians,
		})
	}
	_ = nonDriveSideHoleCount
	return wheelsetLacingAngularPointSet{rimHoles: rimHoles, hubHolesA: hubHolesA, hubHolesB: hubHolesB}
}

func resolveWheelsetLacingG3RimHoleSpacingValues(
	spacingAToB, spacingBToA, spacingAToNextGroupA float64,
	rimHoleCircleRadiusMM, sideAFlangeHoleCircleRadiusMM float64,
) (float64, float64, float64, error) {
	if spacingAToB == 0 && spacingBToA == 0 && spacingAToNextGroupA == 0 {
		return calculateWheelsetLacingTwentyOneHoleG3ParallelSpacing(
			rimHoleCircleRadiusMM,
			sideAFlangeHoleCircleRadiusMM,
		)
	}
	if spacingAToB <= 0 || spacingBToA <= 0 {
		return 0, 0, 0, fmt.Errorf("%w: custom 21-hole G3 spacing requires positive A-to-B and B-to-A gaps", ErrInvalidRequest)
	}
	if spacingAToNextGroupA == 0 {
		spacingAToNextGroupA = WheelsetLacingTwentyOneHoleG3GroupPitchDegrees - spacingAToB - spacingBToA
	}
	return spacingAToB, spacingBToA, spacingAToNextGroupA, nil
}

// calculateWheelsetLacingTwentyOneHoleG3ParallelSpacing derives the symmetric
// rim-hole gaps that make each mapped pair of side-A spokes share a direction.
// It equates the rim-side and flange-side pair chord lengths at their matched
// angular midline; the spoke mapping must preserve the corresponding endpoint
// order for this condition to produce parallel spoke vectors.
func calculateWheelsetLacingTwentyOneHoleG3ParallelSpacing(
	rimHoleCircleRadiusMM, sideAFlangeHoleCircleRadiusMM float64,
) (float64, float64, float64, error) {
	if !isFinitePositiveWheelsetLacingNumber(rimHoleCircleRadiusMM) || !isFinitePositiveWheelsetLacingNumber(sideAFlangeHoleCircleRadiusMM) {
		return 0, 0, 0, fmt.Errorf("%w: automatic 21-hole G3 parallel spacing requires positive rim and side-A flange hole-circle radii", ErrInvalidRequest)
	}
	const hubPairHalfAngleRadians = WheelsetLacingTwentyOneHoleG3GroupPitchDegrees * math.Pi / 720
	closingGapSine := (sideAFlangeHoleCircleRadiusMM / rimHoleCircleRadiusMM) * math.Sin(hubPairHalfAngleRadians)
	maximumClosingGapSine := math.Sin(WheelsetLacingTwentyOneHoleG3GroupPitchDegrees * math.Pi / 360)
	if !isFinitePositiveWheelsetLacingNumber(closingGapSine) || closingGapSine >= maximumClosingGapSine {
		return 0, 0, 0, fmt.Errorf("%w: rim and side-A flange radii cannot form a valid 21-hole G3 parallel group", ErrInvalidRequest)
	}
	closingGapDegrees := 2 * math.Asin(closingGapSine) * 180 / math.Pi
	withinGroupGapDegrees := (WheelsetLacingTwentyOneHoleG3GroupPitchDegrees - closingGapDegrees) / 2
	if !isFinitePositiveWheelsetLacingNumber(withinGroupGapDegrees) {
		return 0, 0, 0, fmt.Errorf("%w: automatic 21-hole G3 parallel spacing produced an invalid within-group gap", ErrInvalidRequest)
	}
	return withinGroupGapDegrees, withinGroupGapDegrees, closingGapDegrees, nil
}

func validateWheelsetLacingG3RimHoleSpacingClosure(spacingAToB, spacingBToA, spacingAToNextGroupA float64) error {
	for field, value := range map[string]float64{
		"A-to-B":            spacingAToB,
		"B-to-A":            spacingBToA,
		"A-to-next-group-A": spacingAToNextGroupA,
	} {
		if value <= 0 || value > WheelsetLacingMaximumTwentyOneHoleG3RimHoleSpacingDegrees {
			return fmt.Errorf("%w: 21-hole G3 rim-hole %s spacing must be within (0, %g] degrees", ErrInvalidRequest, field, WheelsetLacingMaximumTwentyOneHoleG3RimHoleSpacingDegrees)
		}
	}
	spacingTotal := spacingAToB + spacingBToA + spacingAToNextGroupA
	if math.Abs(spacingTotal-WheelsetLacingTwentyOneHoleG3GroupPitchDegrees) > WheelsetLacingTwentyOneHoleG3SpacingClosureToleranceDegrees {
		return fmt.Errorf("%w: 21-hole G3 rim-hole A-to-B, B-to-A, and A-to-next-group-A spacings must sum to %.6g degrees", ErrInvalidRequest, WheelsetLacingTwentyOneHoleG3GroupPitchDegrees)
	}
	return nil
}

func wheelsetLacingAngularHubKey(side Side, id int) string {
	return fmt.Sprintf("%s:%d", side, id)
}

func normalizeWheelsetLacingAngularDifference(value float64) float64 {
	return math.Atan2(math.Sin(value), math.Cos(value))
}

func isFinitePositiveWheelsetLacingNumber(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}
