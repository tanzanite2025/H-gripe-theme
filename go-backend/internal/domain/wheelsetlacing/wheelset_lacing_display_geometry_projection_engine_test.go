package wheelsetlacing

import (
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
	for _, topologyID := range []string{"21h-g3-2to1", "24h-uniform-2to1"} {
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
			if len(result.Spokes) != topology.HoleCount || result.Metrics.DriveSideSpokeCount == 0 {
				t.Fatalf("unexpected special topology result: %+v", result)
			}
		})
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
