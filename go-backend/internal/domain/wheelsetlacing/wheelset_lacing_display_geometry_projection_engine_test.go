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

func TestCalculateWheelsetLacingDisplayGeometryProjectionSharesStraightPullAnchorsAcrossPatterns(t *testing.T) {
	catalog := NewDefaultCatalog()
	topologyIDs := []string{
		"16h-symmetric-1to1-0x-straight-pull",
		"24h-symmetric-1to1-2x-straight-pull",
		"24h-uniform-2to1-straight-pull",
		"18h-uniform-2to1-straight-pull",
	}
	for _, topologyID := range topologyIDs {
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
			if topology.SpokeHeadStyle != SpokeHeadStyleStraightPull || result.StraightPullProjection == nil {
				t.Fatalf("topology %q must return straight-pull flange projections", topologyID)
			}
			spokesBySharedHubHole := make(map[string][]DisplayGeometrySpoke)
			for _, spoke := range result.Spokes {
				key := displayGeometryHubKey(spoke.Side, spoke.Hub.ID)
				spokesBySharedHubHole[key] = append(spokesBySharedHubHole[key], spoke)
			}
			for key, pairedSpokes := range spokesBySharedHubHole {
				if len(pairedSpokes) != 2 {
					t.Fatalf("straight-pull hub hole %s has %d spokes, want 2", key, len(pairedSpokes))
				}
				firstSpoke := pairedSpokes[0]
				secondSpoke := pairedSpokes[1]
				if firstSpoke.Hub.X != secondSpoke.Hub.X || firstSpoke.Hub.Y != secondSpoke.Hub.Y {
					t.Fatalf("straight-pull hub hole %s does not share one anchor: %+v / %+v", key, firstSpoke.Hub, secondSpoke.Hub)
				}
				hubRadius := math.Hypot(firstSpoke.Hub.X, firstSpoke.Hub.Y)
				firstVectorX := firstSpoke.Rim.X - firstSpoke.Hub.X
				firstVectorY := firstSpoke.Rim.Y - firstSpoke.Hub.Y
				secondVectorX := secondSpoke.Rim.X - secondSpoke.Hub.X
				secondVectorY := secondSpoke.Rim.Y - secondSpoke.Hub.Y
				firstTangentialComponent := (-firstSpoke.Hub.Y*firstVectorX + firstSpoke.Hub.X*firstVectorY) / hubRadius
				secondTangentialComponent := (-secondSpoke.Hub.Y*secondVectorX + secondSpoke.Hub.X*secondVectorY) / hubRadius
				if firstTangentialComponent*secondTangentialComponent >= 0 {
					t.Fatalf("straight-pull spokes from hole %s do not extend in opposite tangential directions: %.4f / %.4f", key, firstTangentialComponent, secondTangentialComponent)
				}
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
	wantSpacingAToB, wantSpacingBToA, wantClosingGap, err := calculateWheelsetLacingTwentyOneHoleG3ParallelSpacing(232, 66)
	if err != nil {
		t.Fatal(err)
	}
	if result.G3GroupSpacing.SpacingAToBDegrees != roundWheelsetLacingDisplayGeometryValue(wantSpacingAToB, 2) || result.G3GroupSpacing.SpacingBToADegrees != roundWheelsetLacingDisplayGeometryValue(wantSpacingBToA, 2) || result.G3GroupSpacing.SpacingAToNextGroupADegrees != roundWheelsetLacingDisplayGeometryValue(wantClosingGap, 2) {
		t.Fatalf("unexpected radius-derived parallel G3 spacing profile: %+v", result.G3GroupSpacing)
	}
}

func TestCalculateWheelsetLacingDisplayGeometryProjectionDerivesG3FlangeSizeFromParallelHoleSpacing(t *testing.T) {
	topology, err := NewDefaultCatalog().Get("21h-g3-2to1")
	if err != nil {
		t.Fatal(err)
	}
	const parallelHoleSpacingMM = 40.0
	result, err := CalculateWheelsetLacingDisplayGeometryProjection(DisplayGeometryProjectionRequest{
		TopologyID: "21h-g3-2to1", RimRadius: 232, FlangeRadiusA: 66, FlangeRadiusB: 54,
		FlangeOffsetAMM: 20, FlangeOffsetBMM: 35, G3ParallelHoleSpacingMM: parallelHoleSpacingMM,
	}, topology)
	if err != nil {
		t.Fatal(err)
	}
	wantFlangeRadius, err := calculateWheelsetLacingTwentyOneHoleG3FlangeHoleCircleRadius(parallelHoleSpacingMM, 232)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(result.G3GroupSpacing.ParallelHoleSpacingMM-parallelHoleSpacingMM) > 0.01 {
		t.Fatalf("parallel hole spacing = %.2f mm, want %.2f mm", result.G3GroupSpacing.ParallelHoleSpacingMM, parallelHoleSpacingMM)
	}
	if math.Abs(result.G3GroupSpacing.SideAFlangeHoleCircleRadiusMM-wantFlangeRadius) > 0.01 {
		t.Fatalf("derived A-side flange hole-circle radius = %.2f mm, want %.2f mm", result.G3GroupSpacing.SideAFlangeHoleCircleRadiusMM, wantFlangeRadius)
	}
	if math.Abs(result.G3GroupSpacing.SideAFlangePCDMM-2*wantFlangeRadius) > 0.01 {
		t.Fatalf("derived A-side flange PCD = %.2f mm, want %.2f mm", result.G3GroupSpacing.SideAFlangePCDMM, 2*wantFlangeRadius)
	}
	if math.Abs(math.Hypot(result.HubHolesA[0].X, result.HubHolesA[0].Y)-wantFlangeRadius) > 0.02 {
		t.Fatalf("projected A-side hub hole radius does not match calculated flange radius %.2f mm", wantFlangeRadius)
	}
	rimPointsByID := make(map[int]DisplayGeometryPoint, len(result.RimHoles))
	for _, point := range result.RimHoles {
		rimPointsByID[point.ID] = point
	}
	if len(result.HubHolesA) != WheelsetLacingTwentyOneHoleG3GroupCount || len(result.HubHolesB) != WheelsetLacingTwentyOneHoleG3GroupCount {
		t.Fatalf("G3 flange point counts: A=%d B=%d, want 7/7", len(result.HubHolesA), len(result.HubHolesB))
	}
	for group := 0; group < WheelsetLacingTwentyOneHoleG3GroupCount; group++ {
		firstRimHole := rimPointsByID[group*3]
		secondRimHole := rimPointsByID[group*3+2]
		rimHoleSpacing := math.Hypot(secondRimHole.X-firstRimHole.X, secondRimHole.Y-firstRimHole.Y)
		if math.Abs(rimHoleSpacing-parallelHoleSpacingMM) > 0.03 {
			t.Fatalf("G3 group %d rim-hole center spacing = %.3f mm, want %.3f mm", group, rimHoleSpacing, parallelHoleSpacingMM)
		}
		nextGroup := (group + 1) % WheelsetLacingTwentyOneHoleG3GroupCount
		firstFlangeAnchor := result.HubHolesA[group]
		nextFlangeAnchor := result.HubHolesA[nextGroup]
		flangeAnchorSpacing := math.Hypot(nextFlangeAnchor.X-firstFlangeAnchor.X, nextFlangeAnchor.Y-firstFlangeAnchor.Y)
		if math.Abs(flangeAnchorSpacing-parallelHoleSpacingMM) > 0.03 {
			t.Fatalf("G3 group %d adjacent A-flange anchor spacing = %.3f mm, want %.3f mm", group, flangeAnchorSpacing, parallelHoleSpacingMM)
		}

		nonDriveSpoke := result.Spokes[group]
		trailingDriveSpoke := result.Spokes[WheelsetLacingTwentyOneHoleG3GroupCount+group*2]
		previousGroup := (group + WheelsetLacingTwentyOneHoleG3GroupCount - 1) % WheelsetLacingTwentyOneHoleG3GroupCount
		leadingDriveSpoke := result.Spokes[WheelsetLacingTwentyOneHoleG3GroupCount+previousGroup*2+1]
		for spokeLabel, spoke := range map[string]DisplayGeometrySpoke{
			"leading":   leadingDriveSpoke,
			"non-drive": nonDriveSpoke,
			"trailing":  trailingDriveSpoke,
		} {
			angularDifference := normalizeWheelsetLacingAngularDifference(
				displayGeometrySpokeOrientation(spoke) - displayGeometrySpokeOrientation(nonDriveSpoke),
			)
			if math.Abs(angularDifference) > 0.0005 {
				t.Fatalf("G3 group %d %s spoke with custom %.2f mm hole spacing is not parallel to its non-drive spoke: %.6f radians", group, spokeLabel, parallelHoleSpacingMM, angularDifference)
			}
		}
	}
}

func TestCalculateWheelsetLacingDisplayGeometryProjectionUsesSevenSharedG3DriveAnchorsForStraightPull(t *testing.T) {
	topology, err := NewDefaultCatalog().Get("21h-g3-2to1")
	if err != nil {
		t.Fatal(err)
	}
	if topology.SpokeHeadStyle != SpokeHeadStyleStraightPull {
		t.Fatalf("G3 spoke head style = %q, want %q", topology.SpokeHeadStyle, SpokeHeadStyleStraightPull)
	}
	result, err := CalculateWheelsetLacingDisplayGeometryProjection(DisplayGeometryProjectionRequest{
		TopologyID: "21h-g3-2to1", RimRadius: 232, FlangeRadiusA: 66, FlangeRadiusB: 54,
		FlangeOffsetAMM: 20, FlangeOffsetBMM: 35,
	}, topology)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.HubHolesA) != WheelsetLacingTwentyOneHoleG3GroupCount || len(result.HubHolesB) != WheelsetLacingTwentyOneHoleG3GroupCount {
		t.Fatalf("G3 flange point counts: A=%d B=%d, want 7/7", len(result.HubHolesA), len(result.HubHolesB))
	}
	for group := 0; group < WheelsetLacingTwentyOneHoleG3GroupCount; group++ {
		firstSpoke := result.Spokes[WheelsetLacingTwentyOneHoleG3GroupCount+group*2]
		secondSpoke := result.Spokes[WheelsetLacingTwentyOneHoleG3GroupCount+group*2+1]
		if firstSpoke.Type != SpokeTypeTrailing || secondSpoke.Type != SpokeTypeLeading {
			t.Fatalf("G3 group %d spoke heads = %q/%q, want paired straight-pull anchor assignments", group, firstSpoke.Type, secondSpoke.Type)
		}
		if firstSpoke.Hub.ID != group || secondSpoke.Hub.ID != group || firstSpoke.Hub.Side != SideA || secondSpoke.Hub.Side != SideA {
			t.Fatalf("G3 group %d drive spokes must share A flange anchor %d: %+v / %+v", group, group, firstSpoke.Hub, secondSpoke.Hub)
		}
		if firstSpoke.Hub.X != secondSpoke.Hub.X || firstSpoke.Hub.Y != secondSpoke.Hub.Y {
			t.Fatalf("G3 group %d drive spokes do not share one A flange point: %+v / %+v", group, firstSpoke.Hub, secondSpoke.Hub)
		}
		firstTangentialComponent := -firstSpoke.Hub.Y*(firstSpoke.Rim.X-firstSpoke.Hub.X) + firstSpoke.Hub.X*(firstSpoke.Rim.Y-firstSpoke.Hub.Y)
		secondTangentialComponent := -secondSpoke.Hub.Y*(secondSpoke.Rim.X-secondSpoke.Hub.X) + secondSpoke.Hub.X*(secondSpoke.Rim.Y-secondSpoke.Hub.Y)
		if firstTangentialComponent*secondTangentialComponent >= 0 {
			t.Fatalf("G3 group %d drive spokes do not extend in opposite tangential directions: %.4f / %.4f", group, firstTangentialComponent, secondTangentialComponent)
		}

		bSpoke := result.Spokes[group]
		previousGroup := (group + WheelsetLacingTwentyOneHoleG3GroupCount - 1) % WheelsetLacingTwentyOneHoleG3GroupCount
		leadingSpoke := result.Spokes[WheelsetLacingTwentyOneHoleG3GroupCount+previousGroup*2+1]
		for spokeLabel, spoke := range map[string]DisplayGeometrySpoke{
			"leading":   leadingSpoke,
			"non-drive": bSpoke,
			"trailing":  firstSpoke,
		} {
			if math.Abs(normalizeWheelsetLacingAngularDifference(displayGeometrySpokeOrientation(spoke)-displayGeometrySpokeOrientation(bSpoke))) > 0.0005 {
				t.Fatalf("G3 group %d %s spoke is not parallel to its non-drive spoke", group, spokeLabel)
			}
		}
		if math.Abs(normalizeWheelsetLacingAngularDifference(result.HubHolesA[group].Angle-result.HubHolesB[group].Angle)-math.Pi/float64(WheelsetLacingTwentyOneHoleG3GroupCount)) > 0.0005 {
			t.Fatalf("G3 group %d drive anchor is not half a group pitch ahead of the non-drive anchor", group)
		}
	}
	if result.StraightPullProjection == nil {
		t.Fatal("G3 must return the dedicated straight-pull flange projection")
	}
	projection := result.StraightPullProjection
	if projection.FlangeCenterA.X <= 0 || projection.FlangeCenterB.X >= 0 {
		t.Fatalf("straight-pull flange centers must extend to opposite axial sides: A=%+v B=%+v", projection.FlangeCenterA, projection.FlangeCenterB)
	}
	if len(projection.HubHolesA) != WheelsetLacingTwentyOneHoleG3GroupCount {
		t.Fatalf("G3 straight-pull projected A flange point count = %d, want 7", len(projection.HubHolesA))
	}
	for group := 0; group < WheelsetLacingTwentyOneHoleG3GroupCount; group++ {
		firstSpoke := projection.Spokes[WheelsetLacingTwentyOneHoleG3GroupCount+group*2]
		secondSpoke := projection.Spokes[WheelsetLacingTwentyOneHoleG3GroupCount+group*2+1]
		if firstSpoke.Hub.X != secondSpoke.Hub.X || firstSpoke.Hub.Y != secondSpoke.Hub.Y {
			t.Fatalf("G3 straight-pull projected drive pair %d does not share one anchor", group)
		}
	}
}

func displayGeometrySpokeOrientation(spoke DisplayGeometrySpoke) float64 {
	return math.Atan2(spoke.Rim.Y-spoke.Hub.Y, spoke.Rim.X-spoke.Hub.X)
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
