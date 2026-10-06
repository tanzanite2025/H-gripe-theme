package wheelsetlacing

import (
	"math"
	"testing"
)

func TestCalculateWheelsetLacingAngularLayoutCoversRegisteredTopologySpokes(t *testing.T) {
	for _, topologyID := range []string{
		"16h-symmetric-1to1-1x",
		"24h-symmetric-1to1-2x",
		"24h-uniform-2to1",
		"18h-uniform-2to1",
		"21h-g3-2to1",
	} {
		t.Run(topologyID, func(t *testing.T) {
			topology, err := NewDefaultCatalog().Get(topologyID)
			if err != nil {
				t.Fatal(err)
			}
			request := WheelsetLacingAngularLayoutRequest{TopologyID: topologyID}
			if topologyID == "21h-g3-2to1" {
				request.G3RimHoleCircleRadiusMM = 232
				request.G3SideAFlangeHoleCircleRadiusMM = 66
			}
			layout, err := CalculateWheelsetLacingAngularLayout(
				request,
				topology,
			)
			if err != nil {
				t.Fatal(err)
			}
			if len(layout.Spokes) != topology.HoleCount {
				t.Fatalf("spoke count = %d, want %d", len(layout.Spokes), topology.HoleCount)
			}
			for _, spoke := range layout.Spokes {
				if !isFiniteAngularLayoutValue(spoke.RelativeAngleRadians) {
					t.Fatalf("spoke %d has non-finite relative angle %v", spoke.ID, spoke.RelativeAngleRadians)
				}
				if math.Abs(spoke.RelativeAngleRadians) > math.Pi {
					t.Fatalf("spoke %d relative angle %v is outside normalized range", spoke.ID, spoke.RelativeAngleRadians)
				}
			}
		})
	}
}

func TestCalculateWheelsetLacingAngularLayoutG3SpacingDefaultsAndCustomClosure(t *testing.T) {
	topology, err := NewDefaultCatalog().Get("21h-g3-2to1")
	if err != nil {
		t.Fatal(err)
	}

	defaultLayout, err := CalculateWheelsetLacingAngularLayout(
		WheelsetLacingAngularLayoutRequest{
			TopologyID: topology.ID, G3RimHoleCircleRadiusMM: 232, G3SideAFlangeHoleCircleRadiusMM: 66,
		},
		topology,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !defaultLayout.G3Spacing.Enabled || defaultLayout.G3Spacing.GroupCount != 7 {
		t.Fatalf("unexpected default G3 spacing: %+v", defaultLayout.G3Spacing)
	}

	customLayout, err := CalculateWheelsetLacingAngularLayout(
		WheelsetLacingAngularLayoutRequest{
			TopologyID:                           topology.ID,
			G3RimHoleSpacingAToBDegrees:          2.5,
			G3RimHoleSpacingBToADegrees:          8.5,
			G3RimHoleSpacingAToNextGroupADegrees: 40.4285714286,
		},
		topology,
	)
	if err != nil {
		t.Fatal(err)
	}
	if customLayout.G3Spacing.SpacingAToBDegrees != 2.5 || customLayout.G3Spacing.SpacingBToADegrees != 8.5 {
		t.Fatalf("unexpected custom G3 spacing: %+v", customLayout.G3Spacing)
	}
	if defaultLayout.Spokes[7].RelativeAngleRadians == customLayout.Spokes[7].RelativeAngleRadians {
		t.Fatal("custom G3 spacing must change the mapped spoke phase")
	}
}

func isFiniteAngularLayoutValue(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
