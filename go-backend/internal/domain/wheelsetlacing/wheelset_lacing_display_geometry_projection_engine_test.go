package wheelsetlacing

import (
	"errors"
	"math"
	"testing"
)

func TestCalculateWheelsetLacingDisplayGeometryProjectionMatchesCanonicalSymmetricPreview(t *testing.T) {
	topology, err := NewDefaultCatalog().Get("24h-symmetric-1to1-2x")
	if err != nil {
		t.Fatal(err)
	}
	result, err := CalculateWheelsetLacingDisplayGeometryProjection(DisplayGeometryProjectionRequest{
		TopologyID: "24h-symmetric-1to1-2x", RimRadius: 232, FlangeRadiusA: 66, FlangeRadiusB: 54,
		FlangeOffsetAMM: 20, FlangeOffsetBMM: 35,
	}, topology)
	if err != nil {
		t.Fatal(err)
	}
	if result.DisplayLayout != DisplayGeometryLayoutSymmetric1To1 || result.Topology.DisplayLayout != DisplayGeometryLayoutSymmetric1To1 {
		t.Fatalf("unexpected symmetric display layout: %+v", result)
	}
	if len(result.RimHoles) != 24 || len(result.HubHolesA) != 12 || len(result.HubHolesB) != 12 || len(result.Spokes) != 24 {
		t.Fatalf("unexpected display geometry counts: %+v", result)
	}
	if math.Abs(result.Metrics.AggregateMeanAbsoluteProjectionAngleDegrees-76.0) > 0.1 {
		t.Fatalf("aggregate projection angle = %v, want 76.0", result.Metrics.AggregateMeanAbsoluteProjectionAngleDegrees)
	}
	if result.RimHoles[0].X != 30.28 || result.RimHoles[0].Y != -230.02 {
		t.Fatalf("unexpected rounded first rim point: %+v", result.RimHoles[0])
	}
	if result.FlangeProfile.FlangeAX != 91.43 || result.FlangeProfile.FlangeBX != -160 || result.FlangeProfile.TotalSpanMM != 55 {
		t.Fatalf("unexpected flange profile: %+v", result.FlangeProfile)
	}
}

func TestCalculateWheelsetLacingDisplayGeometryProjectionKeepsSpecialTopologyMappings(t *testing.T) {
	catalog := NewDefaultCatalog()
	for _, topologyID := range []string{"21h-g3-2to1", "24h-uniform-2to1", "18h-uniform-2to1"} {
		t.Run(topologyID, func(t *testing.T) {
			topology, err := catalog.Get(topologyID)
			if err != nil {
				t.Fatal(err)
			}
			result, err := CalculateWheelsetLacingDisplayGeometryProjection(DisplayGeometryProjectionRequest{
				TopologyID: topologyID, RimRadius: 232, FlangeRadiusA: 66, FlangeRadiusB: 54,
				FlangeOffsetAMM: 20, FlangeOffsetBMM: 35,
			}, topology)
			if err != nil {
				t.Fatal(err)
			}
			if result.DisplayLayout != topology.DisplayLayout {
				t.Fatalf("display layout = %q, want %q", result.DisplayLayout, topology.DisplayLayout)
			}
			if len(result.Spokes) != topology.HoleCount || result.Metrics.DriveSideSpokeCount == 0 {
				t.Fatalf("unexpected special topology result: %+v", result)
			}
		})
	}
}

func TestCalculateWheelsetLacingDisplayGeometryProjectionUsesIndependentUniform18HGeometry(t *testing.T) {
	topology, err := NewDefaultCatalog().Get("18h-uniform-2to1")
	if err != nil {
		t.Fatal(err)
	}
	result, err := CalculateWheelsetLacingDisplayGeometryProjection(DisplayGeometryProjectionRequest{
		TopologyID: "18h-uniform-2to1", RimRadius: 232, FlangeRadiusA: 66, FlangeRadiusB: 54,
		FlangeOffsetAMM: 20, FlangeOffsetBMM: 35,
	}, topology)
	if err != nil {
		t.Fatal(err)
	}
	if result.DisplayLayout != DisplayGeometryLayoutUniform18H2To1 {
		t.Fatalf("display layout = %q, want %q", result.DisplayLayout, DisplayGeometryLayoutUniform18H2To1)
	}
	if len(result.RimHoles) != 18 || len(result.HubHolesA) != 12 || len(result.HubHolesB) != 6 || len(result.Spokes) != 18 {
		t.Fatalf("unexpected uniform 18H display geometry counts: rim=%d hubA=%d hubB=%d spokes=%d", len(result.RimHoles), len(result.HubHolesA), len(result.HubHolesB), len(result.Spokes))
	}
	if result.G3GroupSpacing.Enabled {
		t.Fatal("uniform 18H display geometry must not expose G3 spacing")
	}
	for index, point := range result.RimHoles {
		if point.Side == SideB && index%3 != 1 {
			t.Fatalf("uniform 18H rim point %d has B side outside A-B-A sequence", index)
		}
		if point.Side == SideA && index%3 == 1 {
			t.Fatalf("uniform 18H rim point %d has A side in B slot", index)
		}
	}
	if result.Metrics.DriveSideSpokeCount != 12 {
		t.Fatalf("drive-side spoke count = %d, want 12", result.Metrics.DriveSideSpokeCount)
	}
	for index, point := range result.RimHoles {
		wantAngle := (float64(index) * 2 * math.Pi / 18) - math.Pi/2 + math.Pi/18
		if math.Abs(point.Angle-wantAngle) > 0.000001 {
			t.Fatalf("uniform 18H rim point %d angle = %v, want %v", index, point.Angle, wantAngle)
		}
	}
	for index, point := range result.HubHolesB {
		rimPoint := result.RimHoles[1+index*3]
		if math.Abs(point.Angle-rimPoint.Angle) > 0.000001 {
			t.Fatalf("uniform 18H non-drive hub point %d angle = %v, want radial rim angle %v", index, point.Angle, rimPoint.Angle)
		}
	}
}

func TestCalculateWheelsetLacingDisplayGeometryProjectionUsesG3DefaultsWhenSpacingIsUnset(t *testing.T) {
	topology, err := NewDefaultCatalog().Get("21h-g3-2to1")
	if err != nil {
		t.Fatal(err)
	}
	result, err := CalculateWheelsetLacingDisplayGeometryProjection(DisplayGeometryProjectionRequest{
		TopologyID: "21h-g3-2to1", RimRadius: 232, FlangeRadiusA: 66, FlangeRadiusB: 54,
		FlangeOffsetAMM: 20, FlangeOffsetBMM: 35,
	}, topology)
	if err != nil {
		t.Fatal(err)
	}
	if result.G3GroupSpacing.SpacingAToBDegrees != DefaultWheelsetLacingTwentyOneHoleG3RimHoleSpacingAToBDegrees || result.G3GroupSpacing.SpacingBToADegrees != DefaultWheelsetLacingTwentyOneHoleG3RimHoleSpacingBToADegrees || result.G3GroupSpacing.SpacingAToNextGroupADegrees != roundWheelsetLacingDisplayGeometryValue(DefaultWheelsetLacingTwentyOneHoleG3RimHoleSpacingAToNextGroupADegrees, 2) {
		t.Fatalf("unexpected G3 default spacing profile: %+v", result.G3GroupSpacing)
	}
}

func TestCalculateWheelsetLacingDisplayGeometryProjectionDerivesOmittedG3ClosingGapForOlderCallers(t *testing.T) {
	topology, err := NewDefaultCatalog().Get("21h-g3-2to1")
	if err != nil {
		t.Fatal(err)
	}
	result, err := CalculateWheelsetLacingDisplayGeometryProjection(DisplayGeometryProjectionRequest{
		TopologyID: "21h-g3-2to1", RimRadius: 232, FlangeRadiusA: 66, FlangeRadiusB: 54,
		FlangeOffsetAMM: 20, FlangeOffsetBMM: 35,
		G3RimHoleSpacingAToBDegrees: 2.5, G3RimHoleSpacingBToADegrees: 8.5,
	}, topology)
	if err != nil {
		t.Fatal(err)
	}
	if result.G3GroupSpacing.SpacingAToNextGroupADegrees != 40.43 {
		t.Fatalf("derived G3 closing gap = %v°, want 40.43°", result.G3GroupSpacing.SpacingAToNextGroupADegrees)
	}
}

func TestCalculateWheelsetLacingDisplayGeometryProjectionRejectsInvalidDisplayRadii(t *testing.T) {
	topology, err := NewDefaultCatalog().Get("24h-symmetric-1to1-2x")
	if err != nil {
		t.Fatal(err)
	}
	_, err = CalculateWheelsetLacingDisplayGeometryProjection(DisplayGeometryProjectionRequest{
		TopologyID: "24h-symmetric-1to1-2x", RimRadius: math.NaN(), FlangeRadiusA: 66, FlangeRadiusB: 54,
	}, topology)
	if err == nil || !containsDisplayGeometryError(err, ErrInvalidRequest) {
		t.Fatalf("error = %v, want invalid request", err)
	}
}

func TestCalculateWheelsetLacingDisplayGeometryProjectionKeepsIndependentFlangeOffsets(t *testing.T) {
	topology, err := NewDefaultCatalog().Get("24h-symmetric-1to1-2x")
	if err != nil {
		t.Fatal(err)
	}
	result, err := CalculateWheelsetLacingDisplayGeometryProjection(DisplayGeometryProjectionRequest{
		TopologyID: "24h-symmetric-1to1-2x", RimRadius: 232, FlangeRadiusA: 74, FlangeRadiusB: 48,
		FlangeOffsetAMM: 12, FlangeOffsetBMM: 28,
	}, topology)
	if err != nil {
		t.Fatal(err)
	}
	if result.HubHolesA[0].X == 66 || result.HubHolesB[0].X == 54 {
		t.Fatalf("custom flange radii were not applied: A=%+v B=%+v", result.HubHolesA[0], result.HubHolesB[0])
	}
	if result.FlangeProfile.FlangeAX != 68.57 || result.FlangeProfile.FlangeBX != -160 {
		t.Fatalf("custom flange offsets were not projected: %+v", result.FlangeProfile)
	}
}

func TestCalculateWheelsetLacingDisplayGeometryProjectionSupportsIndependentG3RimHoleSpacing(t *testing.T) {
	topology, err := NewDefaultCatalog().Get("21h-g3-2to1")
	if err != nil {
		t.Fatal(err)
	}
	result, err := CalculateWheelsetLacingDisplayGeometryProjection(DisplayGeometryProjectionRequest{
		TopologyID: "21h-g3-2to1", RimRadius: 232, FlangeRadiusA: 66, FlangeRadiusB: 54,
		FlangeOffsetAMM: 20, FlangeOffsetBMM: 35,
		G3RimHoleSpacingAToBDegrees: 2.5, G3RimHoleSpacingBToADegrees: 8.5, G3RimHoleSpacingAToNextGroupADegrees: 40.428571,
	}, topology)
	if err != nil {
		t.Fatal(err)
	}
	if !result.G3GroupSpacing.Enabled || result.G3GroupSpacing.SpacingAToBDegrees != 2.5 || result.G3GroupSpacing.SpacingBToADegrees != 8.5 || result.G3GroupSpacing.SpacingAToNextGroupADegrees != 40.43 {
		t.Fatalf("unexpected G3 group spacing profile: %+v", result.G3GroupSpacing)
	}
	if got, want := result.RimHoles[1].Angle, -math.Pi/2; math.Abs(got-want) > 0.000001 {
		t.Fatalf("G3 center hole angle = %v, want %v", got, want)
	}
	if got, want := result.RimHoles[0].Angle, -math.Pi/2-2.5*math.Pi/180; math.Abs(got-want) > 0.000001 {
		t.Fatalf("G3 A-to-B hole angle = %v, want %v", got, want)
	}
	if got, want := result.RimHoles[2].Angle, -math.Pi/2+8.5*math.Pi/180; math.Abs(got-want) > 0.000001 {
		t.Fatalf("G3 B-to-A hole angle = %v, want %v", got, want)
	}
}

func TestCalculateWheelsetLacingDisplayGeometryProjectionPreservesAllThreeG3RimHoleGapsAcrossGroups(t *testing.T) {
	topology, err := NewDefaultCatalog().Get("21h-g3-2to1")
	if err != nil {
		t.Fatal(err)
	}
	const spacingAToB = 2.5
	const spacingBToA = 8.5
	const spacingAToNextGroupA = 40.428571
	result, err := CalculateWheelsetLacingDisplayGeometryProjection(DisplayGeometryProjectionRequest{
		TopologyID: "21h-g3-2to1", RimRadius: 232, FlangeRadiusA: 66, FlangeRadiusB: 54,
		FlangeOffsetAMM: 20, FlangeOffsetBMM: 35,
		G3RimHoleSpacingAToBDegrees: spacingAToB, G3RimHoleSpacingBToADegrees: spacingBToA,
		G3RimHoleSpacingAToNextGroupADegrees: spacingAToNextGroupA,
	}, topology)
	if err != nil {
		t.Fatal(err)
	}
	for group := 0; group < WheelsetLacingTwentyOneHoleG3GroupCount; group++ {
		firstHoleIndex := group * 3
		nextGroupFirstHoleIndex := ((group + 1) % WheelsetLacingTwentyOneHoleG3GroupCount) * 3
		if got := positiveG3RimHoleAngleDifferenceInDegrees(result.RimHoles[firstHoleIndex+1].Angle, result.RimHoles[firstHoleIndex].Angle); math.Abs(got-spacingAToB) > 0.0001 {
			t.Fatalf("group %d A-to-B gap = %v°, want %v°", group, got, spacingAToB)
		}
		if got := positiveG3RimHoleAngleDifferenceInDegrees(result.RimHoles[firstHoleIndex+2].Angle, result.RimHoles[firstHoleIndex+1].Angle); math.Abs(got-spacingBToA) > 0.0001 {
			t.Fatalf("group %d B-to-A gap = %v°, want %v°", group, got, spacingBToA)
		}
		if got := positiveG3RimHoleAngleDifferenceInDegrees(result.RimHoles[nextGroupFirstHoleIndex].Angle, result.RimHoles[firstHoleIndex+2].Angle); math.Abs(got-spacingAToNextGroupA) > 0.0001 {
			t.Fatalf("group %d A-to-next-group-A gap = %v°, want %v°", group, got, spacingAToNextGroupA)
		}
	}
}

func positiveG3RimHoleAngleDifferenceInDegrees(toAngle, fromAngle float64) float64 {
	angleDifference := math.Mod(toAngle-fromAngle, 2*math.Pi)
	if angleDifference < 0 {
		angleDifference += 2 * math.Pi
	}
	return angleDifference * 180 / math.Pi
}

func TestCalculateWheelsetLacingDisplayGeometryProjectionRejectsNonClosingG3RimHoleSpacing(t *testing.T) {
	topology, err := NewDefaultCatalog().Get("21h-g3-2to1")
	if err != nil {
		t.Fatal(err)
	}
	_, err = CalculateWheelsetLacingDisplayGeometryProjection(DisplayGeometryProjectionRequest{
		TopologyID: "21h-g3-2to1", RimRadius: 232, FlangeRadiusA: 66, FlangeRadiusB: 54,
		FlangeOffsetAMM: 20, FlangeOffsetBMM: 35,
		G3RimHoleSpacingAToBDegrees: 5, G3RimHoleSpacingBToADegrees: 5, G3RimHoleSpacingAToNextGroupADegrees: 5,
	}, topology)
	if err == nil || !containsDisplayGeometryError(err, ErrInvalidRequest) {
		t.Fatalf("error = %v, want non-closing G3 rim-hole spacing request rejection", err)
	}
}

func TestCalculateWheelsetLacingDisplayGeometryProjectionRejectsUnregisteredDisplayLayout(t *testing.T) {
	topology, err := NewDefaultCatalog().Get("24h-symmetric-1to1-2x")
	if err != nil {
		t.Fatal(err)
	}
	topology.DisplayLayout = DisplayGeometryLayout("future_18h_2to1")
	_, err = CalculateWheelsetLacingDisplayGeometryProjection(DisplayGeometryProjectionRequest{
		TopologyID: "24h-symmetric-1to1-2x", RimRadius: 232, FlangeRadiusA: 66, FlangeRadiusB: 54,
		FlangeOffsetAMM: 20, FlangeOffsetBMM: 35,
	}, topology)
	if err == nil || !errors.Is(err, ErrInvalidTopology) {
		t.Fatalf("error = %v, want ErrInvalidTopology for an unregistered display layout", err)
	}
}

func TestBuildWheelsetLacingDisplayGeometryPointsDoesNotFallbackUnknownLayoutToSymmetric(t *testing.T) {
	topology, err := NewDefaultCatalog().Get("24h-symmetric-1to1-2x")
	if err != nil {
		t.Fatal(err)
	}
	topology.DisplayLayout = DisplayGeometryLayout("future_18h_2to1")
	_, _, _, err = buildWheelsetLacingDisplayGeometryPoints(topology, DisplayGeometryProjectionRequest{
		TopologyID: "24h-symmetric-1to1-2x", RimRadius: 232, FlangeRadiusA: 66, FlangeRadiusB: 54,
		FlangeOffsetAMM: 20, FlangeOffsetBMM: 35,
	})
	if err == nil || !errors.Is(err, ErrInvalidTopology) {
		t.Fatalf("error = %v, want unsupported display layout", err)
	}
}

func TestBuildWheelsetLacingDisplayGeometryPointsRejectsWrongSpecialLayoutShape(t *testing.T) {
	topology, err := NewDefaultCatalog().Get("24h-uniform-2to1")
	if err != nil {
		t.Fatal(err)
	}
	topology.DisplayLayout = DisplayGeometryLayoutSymmetric1To1
	_, _, _, err = buildWheelsetLacingDisplayGeometryPoints(topology, DisplayGeometryProjectionRequest{
		TopologyID: "24h-uniform-2to1", RimRadius: 232, FlangeRadiusA: 66, FlangeRadiusB: 54,
	})
	if err == nil || !errors.Is(err, ErrInvalidTopology) {
		t.Fatalf("error = %v, want special layout shape mismatch", err)
	}
}

func TestCalculateWheelsetLacingDisplayGeometryProjectionRejectsInvalidFlangeOffsets(t *testing.T) {
	topology, err := NewDefaultCatalog().Get("24h-symmetric-1to1-2x")
	if err != nil {
		t.Fatal(err)
	}
	_, err = CalculateWheelsetLacingDisplayGeometryProjection(DisplayGeometryProjectionRequest{
		TopologyID: "24h-symmetric-1to1-2x", RimRadius: 232, FlangeRadiusA: 66, FlangeRadiusB: 54,
		FlangeOffsetAMM: -1, FlangeOffsetBMM: 35,
	}, topology)
	if err == nil || !containsDisplayGeometryError(err, ErrInvalidRequest) {
		t.Fatalf("error = %v, want invalid flange offset request", err)
	}
}

func containsDisplayGeometryError(err, target error) bool {
	for err != nil {
		if err == target {
			return true
		}
		type unwrap interface{ Unwrap() error }
		unwrapped, ok := err.(unwrap)
		if !ok {
			return false
		}
		err = unwrapped.Unwrap()
	}
	return false
}
