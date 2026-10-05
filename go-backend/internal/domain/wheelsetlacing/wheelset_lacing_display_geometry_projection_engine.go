package wheelsetlacing

import (
	"fmt"
	"math"
)

// WheelsetLacingDisplayGeometryContractVersion identifies the backend contract
// used by the wheelset lacing reference page. Radial values are canvas
// coordinates; flange offsets are physical millimetres used only for the
// generated axial reference profile. Neither represents spoke length,
// stiffness, tension, efficiency, or assembly safety.
const WheelsetLacingDisplayGeometryContractVersion = "v1.4-backend-display-geometry"

// These values are centralized in the domain package so SSR and browser
// requests use the same reference geometry. Radial constants are canvas
// coordinates; the flange-offset defaults are illustrative physical profile
// inputs and are not used as spoke-length or safety calculations.
const (
	DefaultWheelsetLacingDisplayRimHoleRingRadius        = 232.0
	DefaultWheelsetLacingDisplayHubFlangeHoleRingRadiusA = 66.0
	DefaultWheelsetLacingDisplayHubFlangeHoleRingRadiusB = 54.0
	DefaultWheelsetLacingFlangeOffsetAMM                 = 20.0
	DefaultWheelsetLacingFlangeOffsetBMM                 = 35.0
	DefaultWheelsetLacingG3RimHoleSpacingAToBDegrees     = 4.87
	DefaultWheelsetLacingG3RimHoleSpacingBToADegrees     = 4.87
	WheelsetLacingDisplayProfileHalfSpan                 = 160.0
	WheelsetLacingDisplayProfileAxleHalfSpan             = 194.0
	WheelsetLacingDisplayMaximumCanvasRadius             = 280.0
	WheelsetLacingMaximumFlangeOffsetMM                  = 100.0
	WheelsetLacingG3GroupCount                           = 7
	WheelsetLacingG3GroupPitchDegrees                    = 360.0 / WheelsetLacingG3GroupCount
	WheelsetLacingMaximumG3RimHoleSpacingDegrees         = 20.0
)

type DisplayGeometryProjectionRequest struct {
	TopologyID                  string  `json:"topology_id"`
	RimRadius                   float64 `json:"rim_radius"`
	FlangeRadiusA               float64 `json:"flange_radius_a"`
	FlangeRadiusB               float64 `json:"flange_radius_b"`
	FlangeOffsetAMM             float64 `json:"flange_offset_a_mm"`
	FlangeOffsetBMM             float64 `json:"flange_offset_b_mm"`
	G3RimHoleSpacingAToBDegrees float64 `json:"g3_rim_hole_spacing_a_to_b_degrees"`
	G3RimHoleSpacingBToADegrees float64 `json:"g3_rim_hole_spacing_b_to_a_degrees"`
}

// NewWheelsetLacingDefaultDisplayGeometryProjectionRequest creates the
// canonical canvas-only request used by the server-rendered reference page.
// Custom display radii and flange offsets remain available to the POST endpoint
// for explicit consumers, while the public GET endpoint uses these stable
// reference values.
func NewWheelsetLacingDefaultDisplayGeometryProjectionRequest(topologyID string) DisplayGeometryProjectionRequest {
	return DisplayGeometryProjectionRequest{
		TopologyID:      topologyID,
		RimRadius:       DefaultWheelsetLacingDisplayRimHoleRingRadius,
		FlangeRadiusA:   DefaultWheelsetLacingDisplayHubFlangeHoleRingRadiusA,
		FlangeRadiusB:   DefaultWheelsetLacingDisplayHubFlangeHoleRingRadiusB,
		FlangeOffsetAMM: DefaultWheelsetLacingFlangeOffsetAMM,
		FlangeOffsetBMM: DefaultWheelsetLacingFlangeOffsetBMM,
	}
}

type DisplayGeometryPoint struct {
	ID    int     `json:"id"`
	Side  Side    `json:"side"`
	Angle float64 `json:"angle"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
}

type DisplayGeometrySpoke struct {
	ID   int                  `json:"id"`
	Side Side                 `json:"side"`
	Type SpokeType            `json:"type"`
	Hub  DisplayGeometryPoint `json:"hub"`
	Rim  DisplayGeometryPoint `json:"rim"`
}

// DisplayGeometryProjectionMetrics intentionally describes only the generated
// canvas vectors. It is not wheel efficiency, stiffness, tension, strength,
// flange clearance, or assembly safety.
type DisplayGeometryProjectionMetrics struct {
	AggregateMeanAbsoluteProjectionAngleDegrees    float64 `json:"aggregate_mean_absolute_projection_angle_degrees"`
	MinimumAbsoluteDriveSideProjectionAngleDegrees float64 `json:"minimum_absolute_drive_side_projection_angle_degrees"`
	MaximumAbsoluteDriveSideProjectionAngleDegrees float64 `json:"maximum_absolute_drive_side_projection_angle_degrees"`
	MeanAbsoluteTangentialProjectionPercent        float64 `json:"mean_absolute_tangential_projection_percent"`
	MeanAbsoluteRadialProjectionPercent            float64 `json:"mean_absolute_radial_projection_percent"`
	DriveSideSpokeCount                            int     `json:"drive_side_spoke_count"`
}

// DisplayGeometryFlangeProfile is a compact axial reference projection. The
// X coordinates are profile canvas coordinates; the offset fields retain the
// physical input in millimetres for labels and accessibility text.
type DisplayGeometryFlangeProfile struct {
	CenterlineX     float64 `json:"centerline_x"`
	FlangeAX        float64 `json:"flange_a_x"`
	FlangeBX        float64 `json:"flange_b_x"`
	AxleLeftX       float64 `json:"axle_left_x"`
	AxleRightX      float64 `json:"axle_right_x"`
	FlangeOffsetAMM float64 `json:"flange_offset_a_mm"`
	FlangeOffsetBMM float64 `json:"flange_offset_b_mm"`
	TotalSpanMM     float64 `json:"total_flange_span_mm"`
}

// DisplayGeometryG3GroupSpacing describes the two angular gaps in each G3
// A-B-A rim-hole triplet. It is display geometry only; it does not represent
// measured rim drilling dimensions in millimetres.
type DisplayGeometryG3GroupSpacing struct {
	Enabled            bool    `json:"enabled"`
	GroupCount         int     `json:"group_count"`
	GroupPitchDegrees  float64 `json:"group_pitch_degrees"`
	SpacingAToBDegrees float64 `json:"spacing_a_to_b_degrees"`
	SpacingBToADegrees float64 `json:"spacing_b_to_a_degrees"`
}

type DisplayGeometryProjectionResult struct {
	ContractVersion string                           `json:"contract_version"`
	DisplayLayout   DisplayGeometryLayout            `json:"display_layout"`
	Topology        Topology                         `json:"topology"`
	RimHoles        []DisplayGeometryPoint           `json:"rim_holes"`
	HubHolesA       []DisplayGeometryPoint           `json:"hub_holes_a"`
	HubHolesB       []DisplayGeometryPoint           `json:"hub_holes_b"`
	Spokes          []DisplayGeometrySpoke           `json:"spokes"`
	Metrics         DisplayGeometryProjectionMetrics `json:"metrics"`
	FlangeProfile   DisplayGeometryFlangeProfile     `json:"flange_profile"`
	G3GroupSpacing  DisplayGeometryG3GroupSpacing    `json:"g3_group_spacing"`
}

func CalculateWheelsetLacingDisplayGeometryProjection(
	request DisplayGeometryProjectionRequest,
	topology Topology,
) (DisplayGeometryProjectionResult, error) {
	if err := validateDisplayGeometryProjectionRequest(request); err != nil {
		return DisplayGeometryProjectionResult{}, err
	}
	if err := validateTopology(topology); err != nil {
		return DisplayGeometryProjectionResult{}, fmt.Errorf("%w: display geometry topology is invalid: %v", ErrInvalidTopology, err)
	}

	rimHoles, hubHolesA, hubHolesB, err := buildWheelsetLacingDisplayGeometryPoints(topology, request)
	if err != nil {
		return DisplayGeometryProjectionResult{}, err
	}

	rimByID := make(map[int]DisplayGeometryPoint, len(rimHoles))
	for _, point := range rimHoles {
		rimByID[point.ID] = point
	}
	hubsByKey := make(map[string]DisplayGeometryPoint, len(hubHolesA)+len(hubHolesB))
	for _, point := range hubHolesA {
		hubsByKey[displayGeometryHubKey(SideA, point.ID)] = point
	}
	for _, point := range hubHolesB {
		hubsByKey[displayGeometryHubKey(SideB, point.ID)] = point
	}

	spokes := make([]DisplayGeometrySpoke, 0, len(topology.Spokes))
	for _, mapping := range topology.Spokes {
		hub, hubOK := hubsByKey[displayGeometryHubKey(mapping.Side, mapping.HubHoleID)]
		rim, rimOK := rimByID[mapping.RimHoleID]
		if !hubOK || !rimOK {
			return DisplayGeometryProjectionResult{}, fmt.Errorf("%w: spoke %d references a missing display point", ErrInvalidTopology, mapping.ID)
		}
		spokes = append(spokes, DisplayGeometrySpoke{ID: mapping.ID, Side: mapping.Side, Type: mapping.Type, Hub: hub, Rim: rim})
	}

	metrics, err := calculateWheelsetLacingDisplayGeometryProjectionMetrics(spokes)
	if err != nil {
		return DisplayGeometryProjectionResult{}, err
	}
	flangeProfile := buildWheelsetLacingDisplayGeometryFlangeProfile(request)
	g3GroupSpacing, err := buildWheelsetLacingDisplayGeometryG3GroupSpacing(request, topology)
	if err != nil {
		return DisplayGeometryProjectionResult{}, err
	}
	return DisplayGeometryProjectionResult{
		ContractVersion: WheelsetLacingDisplayGeometryContractVersion,
		DisplayLayout:   topology.DisplayLayout,
		Topology:        topology,
		RimHoles:        rimHoles,
		HubHolesA:       hubHolesA,
		HubHolesB:       hubHolesB,
		Spokes:          spokes,
		Metrics:         metrics,
		FlangeProfile:   flangeProfile,
		G3GroupSpacing:  g3GroupSpacing,
	}, nil
}

func validateDisplayGeometryProjectionRequest(request DisplayGeometryProjectionRequest) error {
	if request.TopologyID == "" {
		return fmt.Errorf("%w: topology_id is required", ErrInvalidRequest)
	}
	for field, value := range map[string]float64{
		"rim_radius":      request.RimRadius,
		"flange_radius_a": request.FlangeRadiusA,
		"flange_radius_b": request.FlangeRadiusB,
	} {
		if math.IsNaN(value) || math.IsInf(value, 0) || value <= 0 || value > WheelsetLacingDisplayMaximumCanvasRadius {
			return fmt.Errorf("%w: %s must be finite and within (0, %g] canvas units", ErrInvalidRequest, field, WheelsetLacingDisplayMaximumCanvasRadius)
		}
	}
	for field, value := range map[string]float64{
		"flange_offset_a_mm": request.FlangeOffsetAMM,
		"flange_offset_b_mm": request.FlangeOffsetBMM,
	} {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > WheelsetLacingMaximumFlangeOffsetMM {
			return fmt.Errorf("%w: %s must be finite and within [0, %g] mm", ErrInvalidRequest, field, WheelsetLacingMaximumFlangeOffsetMM)
		}
	}
	for field, value := range map[string]float64{
		"g3_rim_hole_spacing_a_to_b_degrees": request.G3RimHoleSpacingAToBDegrees,
		"g3_rim_hole_spacing_b_to_a_degrees": request.G3RimHoleSpacingBToADegrees,
	} {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > WheelsetLacingMaximumG3RimHoleSpacingDegrees {
			return fmt.Errorf("%w: %s must be finite and within [0, %g] degrees", ErrInvalidRequest, field, WheelsetLacingMaximumG3RimHoleSpacingDegrees)
		}
	}
	return nil
}

func buildWheelsetLacingDisplayGeometryFlangeProfile(request DisplayGeometryProjectionRequest) DisplayGeometryFlangeProfile {
	maximumOffset := math.Max(request.FlangeOffsetAMM, request.FlangeOffsetBMM)
	profileScale := 0.0
	if maximumOffset > 0 {
		profileScale = WheelsetLacingDisplayProfileHalfSpan / maximumOffset
	}
	return DisplayGeometryFlangeProfile{
		CenterlineX:     0,
		FlangeAX:        roundWheelsetLacingDisplayGeometryValue(request.FlangeOffsetAMM*profileScale, 2),
		FlangeBX:        roundWheelsetLacingDisplayGeometryValue(-request.FlangeOffsetBMM*profileScale, 2),
		AxleLeftX:       -WheelsetLacingDisplayProfileAxleHalfSpan,
		AxleRightX:      WheelsetLacingDisplayProfileAxleHalfSpan,
		FlangeOffsetAMM: roundWheelsetLacingDisplayGeometryValue(request.FlangeOffsetAMM, 1),
		FlangeOffsetBMM: roundWheelsetLacingDisplayGeometryValue(request.FlangeOffsetBMM, 1),
		TotalSpanMM:     roundWheelsetLacingDisplayGeometryValue(request.FlangeOffsetAMM+request.FlangeOffsetBMM, 1),
	}
}

func buildWheelsetLacingDisplayGeometryG3GroupSpacing(
	request DisplayGeometryProjectionRequest,
	topology Topology,
) (DisplayGeometryG3GroupSpacing, error) {
	if topology.DisplayLayout != DisplayGeometryLayoutG3Triplet2To1 {
		return DisplayGeometryG3GroupSpacing{}, nil
	}
	spacingAToB, spacingBToA := resolveWheelsetLacingG3RimHoleSpacing(request)
	if spacingAToB+spacingBToA >= WheelsetLacingG3GroupPitchDegrees {
		return DisplayGeometryG3GroupSpacing{}, fmt.Errorf("%w: G3 rim-hole triplet spacing must leave a positive gap between groups", ErrInvalidRequest)
	}
	return DisplayGeometryG3GroupSpacing{
		Enabled:            true,
		GroupCount:         WheelsetLacingG3GroupCount,
		GroupPitchDegrees:  roundWheelsetLacingDisplayGeometryValue(WheelsetLacingG3GroupPitchDegrees, 2),
		SpacingAToBDegrees: roundWheelsetLacingDisplayGeometryValue(spacingAToB, 2),
		SpacingBToADegrees: roundWheelsetLacingDisplayGeometryValue(spacingBToA, 2),
	}, nil
}

func resolveWheelsetLacingG3RimHoleSpacing(request DisplayGeometryProjectionRequest) (float64, float64) {
	spacingAToB := request.G3RimHoleSpacingAToBDegrees
	spacingBToA := request.G3RimHoleSpacingBToADegrees
	if spacingAToB == 0 {
		spacingAToB = DefaultWheelsetLacingG3RimHoleSpacingAToBDegrees
	}
	if spacingBToA == 0 {
		spacingBToA = DefaultWheelsetLacingG3RimHoleSpacingBToADegrees
	}
	return spacingAToB, spacingBToA
}

func buildWheelsetLacingDisplayGeometryPoints(
	topology Topology,
	request DisplayGeometryProjectionRequest,
) ([]DisplayGeometryPoint, []DisplayGeometryPoint, []DisplayGeometryPoint, error) {
	switch topology.DisplayLayout {
	case DisplayGeometryLayoutG3Triplet2To1:
		return buildG3DisplayGeometryPoints(topology, request)
	case DisplayGeometryLayoutUniform2To1:
		return buildUniformTwoToOneDisplayGeometryPoints(topology, request)
	case DisplayGeometryLayoutSymmetric1To1:
		return buildSymmetricDisplayGeometryPoints(topology, topology.HoleCount, request)
	default:
		return nil, nil, nil, fmt.Errorf("%w: unsupported display layout %q", ErrInvalidTopology, topology.DisplayLayout)
	}
}

func buildSymmetricDisplayGeometryPoints(topology Topology, holeCount int, request DisplayGeometryProjectionRequest) ([]DisplayGeometryPoint, []DisplayGeometryPoint, []DisplayGeometryPoint, error) {
	if holeCount <= 0 || len(topology.RimHoles) != holeCount || len(topology.HubHolesA) != holeCount/2 || len(topology.HubHolesB) != holeCount/2 {
		return nil, nil, nil, fmt.Errorf("%w: symmetric display layout shape does not match %d rim holes", ErrInvalidTopology, holeCount)
	}
	rimHoles := make([]DisplayGeometryPoint, 0, len(topology.RimHoles))
	rimBySide := map[Side][]DisplayGeometryPoint{SideA: {}, SideB: {}}
	for index, hole := range topology.RimHoles {
		point := displayGeometryPoint(hole.ID, hole.Side, (float64(index)*2*math.Pi/float64(holeCount))-math.Pi/2+math.Pi/float64(holeCount), request.RimRadius)
		rimHoles = append(rimHoles, point)
		rimBySide[hole.Side] = append(rimBySide[hole.Side], point)
	}
	hubHolesA := make([]DisplayGeometryPoint, 0, len(topology.HubHolesA))
	hubHolesB := make([]DisplayGeometryPoint, 0, len(topology.HubHolesB))
	for index, hole := range topology.HubHolesA {
		if index >= len(rimBySide[SideA]) {
			return nil, nil, nil, fmt.Errorf("%w: symmetric side A hole %d has no rim reference", ErrInvalidTopology, hole.ID)
		}
		hubHolesA = append(hubHolesA, displayGeometryPoint(hole.ID, SideA, rimBySide[SideA][index].Angle, request.FlangeRadiusA))
	}
	for index, hole := range topology.HubHolesB {
		if index >= len(rimBySide[SideB]) {
			return nil, nil, nil, fmt.Errorf("%w: symmetric side B hole %d has no rim reference", ErrInvalidTopology, hole.ID)
		}
		hubHolesB = append(hubHolesB, displayGeometryPoint(hole.ID, SideB, rimBySide[SideB][index].Angle, request.FlangeRadiusB))
	}
	return rimHoles, hubHolesA, hubHolesB, nil
}

func buildG3DisplayGeometryPoints(topology Topology, request DisplayGeometryProjectionRequest) ([]DisplayGeometryPoint, []DisplayGeometryPoint, []DisplayGeometryPoint, error) {
	if topology.HoleCount != 21 || len(topology.RimHoles) != 21 || len(topology.HubHolesA) != 14 || len(topology.HubHolesB) != 7 {
		return nil, nil, nil, fmt.Errorf("%w: G3 display layout requires a 21-hole 14/7 topology", ErrInvalidTopology)
	}
	spacingAToB, spacingBToA := resolveWheelsetLacingG3RimHoleSpacing(request)
	if spacingAToB+spacingBToA >= WheelsetLacingG3GroupPitchDegrees {
		return nil, nil, nil, fmt.Errorf("%w: G3 rim-hole triplet spacing must leave a positive gap between groups", ErrInvalidRequest)
	}
	spacingAToBRadians := spacingAToB * math.Pi / 180
	spacingBToARadians := spacingBToA * math.Pi / 180
	rimHoles := make([]DisplayGeometryPoint, 0, len(topology.RimHoles))
	for group := 0; group < WheelsetLacingG3GroupCount; group++ {
		centerAngle := (float64(group) * 2 * math.Pi / float64(WheelsetLacingG3GroupCount)) - math.Pi/2
		rimHoles = append(rimHoles,
			displayGeometryPoint(group*3, SideA, centerAngle-spacingAToBRadians, request.RimRadius),
			displayGeometryPoint(group*3+1, SideB, centerAngle, request.RimRadius),
			displayGeometryPoint(group*3+2, SideA, centerAngle+spacingBToARadians, request.RimRadius),
		)
	}
	hubHolesA := make([]DisplayGeometryPoint, 0, len(topology.HubHolesA))
	hubHolesB := make([]DisplayGeometryPoint, 0, len(topology.HubHolesB))
	for group := 0; group < WheelsetLacingG3GroupCount; group++ {
		centerAngle := (float64(group) * 2 * math.Pi / float64(WheelsetLacingG3GroupCount)) - math.Pi/2
		midAngle := centerAngle + math.Pi/float64(WheelsetLacingG3GroupCount)
		hubHolesA = append(hubHolesA,
			displayGeometryPoint(group*2, SideA, midAngle-math.Pi/14, request.FlangeRadiusA),
			displayGeometryPoint(group*2+1, SideA, midAngle+math.Pi/14, request.FlangeRadiusA),
		)
		hubHolesB = append(hubHolesB, displayGeometryPoint(group, SideB, centerAngle, request.FlangeRadiusB))
	}
	return rimHoles, hubHolesA, hubHolesB, nil
}

func buildUniformTwoToOneDisplayGeometryPoints(topology Topology, request DisplayGeometryProjectionRequest) ([]DisplayGeometryPoint, []DisplayGeometryPoint, []DisplayGeometryPoint, error) {
	if topology.HoleCount != 24 || len(topology.RimHoles) != 24 || len(topology.HubHolesA) != 16 || len(topology.HubHolesB) != 8 {
		return nil, nil, nil, fmt.Errorf("%w: uniform 2:1 display layout requires a 24-hole 16/8 topology", ErrInvalidTopology)
	}
	const total = 24
	rimHoles := make([]DisplayGeometryPoint, 0, len(topology.RimHoles))
	rimBySide := map[Side][]DisplayGeometryPoint{SideA: {}, SideB: {}}
	for index, hole := range topology.RimHoles {
		angle := (float64(index) * 2 * math.Pi / total) - math.Pi/2 + math.Pi/total
		point := displayGeometryPoint(hole.ID, hole.Side, angle, request.RimRadius)
		rimHoles = append(rimHoles, point)
		rimBySide[hole.Side] = append(rimBySide[hole.Side], point)
	}
	hubHolesA := make([]DisplayGeometryPoint, 0, len(topology.HubHolesA))
	for index, hole := range topology.HubHolesA {
		angle := (float64(index) * 2 * math.Pi / 16) - math.Pi/2
		hubHolesA = append(hubHolesA, displayGeometryPoint(hole.ID, SideA, angle, request.FlangeRadiusA))
	}
	hubHolesB := make([]DisplayGeometryPoint, 0, len(topology.HubHolesB))
	for index, hole := range topology.HubHolesB {
		if index >= len(rimBySide[SideB]) {
			return nil, nil, nil, fmt.Errorf("%w: uniform side B hole %d has no rim reference", ErrInvalidTopology, hole.ID)
		}
		hubHolesB = append(hubHolesB, displayGeometryPoint(hole.ID, SideB, rimBySide[SideB][index].Angle, request.FlangeRadiusB))
	}
	return rimHoles, hubHolesA, hubHolesB, nil
}

func displayGeometryPoint(id int, side Side, angle, radius float64) DisplayGeometryPoint {
	return DisplayGeometryPoint{
		ID: id, Side: side,
		Angle: roundWheelsetLacingDisplayGeometryValue(angle, 6),
		X:     roundWheelsetLacingDisplayGeometryValue(radius*math.Cos(angle), 2),
		Y:     roundWheelsetLacingDisplayGeometryValue(radius*math.Sin(angle), 2),
	}
}

func calculateWheelsetLacingDisplayGeometryProjectionMetrics(spokes []DisplayGeometrySpoke) (DisplayGeometryProjectionMetrics, error) {
	tangential := make([]float64, 0)
	radial := make([]float64, 0)
	angles := make([]float64, 0)
	for _, spoke := range spokes {
		if spoke.Side != SideA {
			continue
		}
		hubRadius := math.Hypot(spoke.Hub.X, spoke.Hub.Y)
		vectorX := spoke.Rim.X - spoke.Hub.X
		vectorY := spoke.Rim.Y - spoke.Hub.Y
		spokeLength := math.Hypot(vectorX, vectorY)
		if !isFinitePositiveWheelsetLacingDisplayGeometryNumber(hubRadius) || !isFinitePositiveWheelsetLacingDisplayGeometryNumber(spokeLength) {
			return DisplayGeometryProjectionMetrics{}, fmt.Errorf("%w: spoke %d has invalid display vector geometry", ErrInvalidTopology, spoke.ID)
		}
		tangentialComponent := math.Abs((-spoke.Hub.Y*vectorX + spoke.Hub.X*vectorY) / (hubRadius * spokeLength))
		radialComponent := math.Abs((spoke.Hub.X*vectorX + spoke.Hub.Y*vectorY) / (hubRadius * spokeLength))
		if !isFiniteWheelsetLacingDisplayGeometryNumber(tangentialComponent) || !isFiniteWheelsetLacingDisplayGeometryNumber(radialComponent) {
			return DisplayGeometryProjectionMetrics{}, fmt.Errorf("%w: spoke %d projection is not finite", ErrInvalidTopology, spoke.ID)
		}
		tangential = append(tangential, math.Min(1, tangentialComponent))
		radial = append(radial, math.Min(1, radialComponent))
		angles = append(angles, math.Atan2(tangentialComponent, radialComponent)*180/math.Pi)
	}
	if len(tangential) == 0 {
		return DisplayGeometryProjectionMetrics{}, fmt.Errorf("%w: no drive-side spokes are available for display projection", ErrInvalidTopology)
	}
	meanTangential := displayGeometryMean(tangential)
	meanRadial := displayGeometryMean(radial)
	return DisplayGeometryProjectionMetrics{
		AggregateMeanAbsoluteProjectionAngleDegrees:    roundWheelsetLacingDisplayGeometryValue(math.Atan2(meanTangential, meanRadial)*180/math.Pi, 1),
		MinimumAbsoluteDriveSideProjectionAngleDegrees: roundWheelsetLacingDisplayGeometryValue(minDisplayGeometryValue(angles), 1),
		MaximumAbsoluteDriveSideProjectionAngleDegrees: roundWheelsetLacingDisplayGeometryValue(maxDisplayGeometryValue(angles), 1),
		MeanAbsoluteTangentialProjectionPercent:        roundWheelsetLacingDisplayGeometryValue(meanTangential*100, 1),
		MeanAbsoluteRadialProjectionPercent:            roundWheelsetLacingDisplayGeometryValue(meanRadial*100, 1),
		DriveSideSpokeCount:                            len(tangential),
	}, nil
}

func displayGeometryMean(values []float64) float64 {
	var total float64
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

func minDisplayGeometryValue(values []float64) float64 {
	result := values[0]
	for _, value := range values[1:] {
		if value < result {
			result = value
		}
	}
	return result
}

func maxDisplayGeometryValue(values []float64) float64 {
	result := values[0]
	for _, value := range values[1:] {
		if value > result {
			result = value
		}
	}
	return result
}

func displayGeometryHubKey(side Side, id int) string {
	return fmt.Sprintf("%s:%d", side, id)
}

func isFinitePositiveWheelsetLacingDisplayGeometryNumber(value float64) bool {
	return isFiniteWheelsetLacingDisplayGeometryNumber(value) && value > 0
}

func isFiniteWheelsetLacingDisplayGeometryNumber(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func roundWheelsetLacingDisplayGeometryValue(value float64, decimals int) float64 {
	scale := math.Pow10(decimals)
	return math.Round(value*scale) / scale
}
