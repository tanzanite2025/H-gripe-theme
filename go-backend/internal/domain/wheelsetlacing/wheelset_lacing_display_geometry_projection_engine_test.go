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
