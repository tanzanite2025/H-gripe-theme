package wheelsetlacing

import (
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestDefaultCatalogContainsAllSupportedTopologyFamilies(t *testing.T) {
	catalog := NewDefaultCatalog()
	topologies := catalog.List()
	if got, want := len(topologies), 51; got != want {
		t.Fatalf("default topology count = %d, want %d", got, want)
	}

	for _, id := range []string{
		"16h-symmetric-1to1-0x",
		"24h-symmetric-1to1-3x",
		"21h-g3-2to1",
		"24h-uniform-2to1",
		"18h-uniform-2to1",
		"16h-symmetric-1to1-0x-straight-pull",
		"24h-symmetric-1to1-2x-straight-pull",
		"24h-uniform-2to1-straight-pull",
		"18h-uniform-2to1-straight-pull",
	} {
		if _, err := catalog.Get(id); err != nil {
			t.Fatalf("Get(%q) error = %v", id, err)
		}
	}
}

func TestDefaultCatalogExposesExactSupportedCrossCombinations(t *testing.T) {
	wantBySelection := map[string][]int{
		"16":      {0, 1},
		"20":      {0, 1, 2},
		"24":      {0, 1, 2, 3},
		"28":      {0, 1, 2, 3},
		"32":      {0, 1, 2, 3, 4},
		"36":      {0, 1, 2, 3, 4},
		"18_2to1": {2},
		"21_g3":   {2},
		"24_2to1": {2},
	}
	crossesBySelection := make(map[string][]int, len(wantBySelection))
	for _, topology := range NewDefaultCatalog().List() {
		if topology.SpokeHeadStyle == SpokeHeadStyleJBend || topology.Distribution == DistributionG32To1 {
			if _, ok := wantBySelection[topology.Selection]; ok {
				crossesBySelection[topology.Selection] = append(crossesBySelection[topology.Selection], topology.Cross)
			}
		}
	}
	for selection, want := range wantBySelection {
		got := crossesBySelection[selection]
		sort.Ints(got)
		if len(got) != len(want) {
			t.Fatalf("selection %q exposes crosses %v, want %v", selection, got, want)
		}
		for index := range want {
			if got[index] != want[index] {
				t.Fatalf("selection %q exposes crosses %v, want %v", selection, got, want)
			}
		}
	}
}

func TestStraightPullTopologyVariantsShareEachFlangeHoleBetweenTwoSpokesAndPreserveJbendMaps(t *testing.T) {
	catalog := NewDefaultCatalog()
	for _, straightPullTopology := range catalog.List() {
		if straightPullTopology.SpokeHeadStyle != SpokeHeadStyleStraightPull || straightPullTopology.Distribution == DistributionG32To1 {
			continue
		}
		baseID := strings.TrimSuffix(straightPullTopology.ID, "-straight-pull")
		jBendTopology, err := catalog.Get(baseID)
		if err != nil {
			t.Fatalf("straight-pull topology %q has no J-bend counterpart: %v", straightPullTopology.ID, err)
		}
		if jBendTopology.SpokeHeadStyle != SpokeHeadStyleJBend {
			t.Fatalf("counterpart %q has spoke head style %q, want J-bend", baseID, jBendTopology.SpokeHeadStyle)
		}
		if !reflect.DeepEqual(straightPullTopology.RimHoles, jBendTopology.RimHoles) {
			t.Fatalf("straight-pull topology %q changed the rim hole assignments", straightPullTopology.ID)
		}
		if len(straightPullTopology.HubHolesA)*2 != len(jBendTopology.HubHolesA) ||
			len(straightPullTopology.HubHolesB)*2 != len(jBendTopology.HubHolesB) {
			t.Fatalf("straight-pull topology %q flange holes are not paired: A=%d/%d B=%d/%d", straightPullTopology.ID,
				len(straightPullTopology.HubHolesA), len(jBendTopology.HubHolesA),
				len(straightPullTopology.HubHolesB), len(jBendTopology.HubHolesB))
		}
		if len(straightPullTopology.Spokes) != len(jBendTopology.Spokes) {
			t.Fatalf("straight-pull topology %q changed spoke count", straightPullTopology.ID)
		}
		spokeUsageByHubHole := make(map[string]int)
		for index, straightPullSpoke := range straightPullTopology.Spokes {
			jBendSpoke := jBendTopology.Spokes[index]
			if straightPullSpoke.ID != jBendSpoke.ID || straightPullSpoke.Side != jBendSpoke.Side ||
				straightPullSpoke.Type != jBendSpoke.Type || straightPullSpoke.RimHoleID != jBendSpoke.RimHoleID {
				t.Fatalf("straight-pull topology %q changed spoke-to-rim mapping at spoke %d", straightPullTopology.ID, straightPullSpoke.ID)
			}
			spokeUsageByHubHole[hubKey(straightPullSpoke.Side, straightPullSpoke.HubHoleID)]++
		}
		for key, usageCount := range spokeUsageByHubHole {
			if usageCount != 2 {
				t.Fatalf("straight-pull topology %q hub hole %s serves %d spokes, want 2", straightPullTopology.ID, key, usageCount)
			}
		}
	}
}

func TestG3AndUniformTwoToOneHaveIndependentMappings(t *testing.T) {
	catalog := NewDefaultCatalog()
	g3, err := catalog.Get("21h-g3-2to1")
	if err != nil {
		t.Fatal(err)
	}
	if g3.Distribution != DistributionG32To1 || g3.DisplayLayout != DisplayGeometryLayoutG3TwentyOneHoleTriplet2To1 || g3.SpokeHeadStyle != SpokeHeadStyleStraightPull || g3.HoleCount != 21 || len(g3.HubHolesA) != 14 || len(g3.HubHolesB) != 7 {
		t.Fatalf("unexpected G3 topology: %+v", g3)
	}
	for index, hole := range g3.RimHoles {
		want := SideA
		if index%3 == 1 {
			want = SideB
		}
		if hole.Side != want {
			t.Fatalf("G3 rim hole %d side = %q, want %q", index, hole.Side, want)
		}
	}
	for group := 0; group < WheelsetLacingTwentyOneHoleG3GroupCount; group++ {
		nextGroup := (group + 1) % WheelsetLacingTwentyOneHoleG3GroupCount
		wantDriveSpokes := []SpokeMapping{
			{ID: WheelsetLacingTwentyOneHoleG3GroupCount + group*2, Side: SideA, Type: SpokeTypeTrailing, HubHoleID: group * 2, RimHoleID: group*3 + 2},
			{ID: WheelsetLacingTwentyOneHoleG3GroupCount + group*2 + 1, Side: SideA, Type: SpokeTypeLeading, HubHoleID: group*2 + 1, RimHoleID: nextGroup * 3},
		}
		for _, want := range wantDriveSpokes {
			spoke := g3.Spokes[want.ID]
			if spoke != want {
				t.Fatalf("G3 straight-pull spoke %d = %+v, want %+v", want.ID, spoke, want)
			}
		}
	}
	for group := 0; group < WheelsetLacingTwentyOneHoleG3GroupCount; group++ {
		nonDriveSpoke := g3.Spokes[group]
		if nonDriveSpoke.Side != SideB || nonDriveSpoke.Type != SpokeTypeNonDrive || nonDriveSpoke.HubHoleID != group || nonDriveSpoke.RimHoleID != group*3+1 {
			t.Fatalf("G3 non-drive spoke %d = %+v, want radial mapping to its grouped rim hole", group, nonDriveSpoke)
		}
	}

	uniform, err := catalog.Get("24h-uniform-2to1")
	if err != nil {
		t.Fatal(err)
	}
	if uniform.Distribution != DistributionUniform2To1 || uniform.DisplayLayout != DisplayGeometryLayoutUniform2To1 || uniform.HoleCount != 24 || len(uniform.HubHolesA) != 16 || len(uniform.HubHolesB) != 8 {
		t.Fatalf("unexpected uniform 2:1 topology: %+v", uniform)
	}
	if uniform.SpokeHeadStyle != SpokeHeadStyleJBend {
		t.Fatalf("uniform 24H 2:1 must retain J-bend heads: %+v", uniform)
	}
	for index, hole := range uniform.RimHoles {
		want := SideA
		if index%3 == 1 {
			want = SideB
		}
		if hole.Side != want {
			t.Fatalf("uniform rim hole %d side = %q, want %q", index, hole.Side, want)
		}
	}
	if g3.RimHoles[0].Side != uniform.RimHoles[0].Side || len(g3.Spokes) == len(uniform.Spokes) {
		t.Fatalf("special topologies should retain distinct topology facts")
	}

	uniform18, err := catalog.Get("18h-uniform-2to1")
	if err != nil {
		t.Fatal(err)
	}
	if uniform18.Distribution != DistributionUniform2To1 || uniform18.DisplayLayout != DisplayGeometryLayoutUniform18H2To1 || uniform18.HoleCount != 18 || len(uniform18.HubHolesA) != 12 || len(uniform18.HubHolesB) != 6 {
		t.Fatalf("unexpected uniform 18H 2:1 topology: %+v", uniform18)
	}
	if uniform18.SpokeHeadStyle != SpokeHeadStyleJBend {
		t.Fatalf("uniform 18H 2:1 must retain J-bend heads: %+v", uniform18)
	}
	for index, hole := range uniform18.RimHoles {
		want := SideA
		if index%3 == 1 {
			want = SideB
		}
		if hole.Side != want {
			t.Fatalf("uniform 18H rim hole %d side = %q, want %q", index, hole.Side, want)
		}
	}
	if len(uniform18.Spokes) != 18 {
		t.Fatalf("uniform 18H spoke count = %d, want 18", len(uniform18.Spokes))
	}
	runningDriveRimHoleIDs := []int{0, 2, 3, 5, 6, 8, 9, 11, 12, 14, 15, 17}
	for index := 0; index < 6; index++ {
		spoke := uniform18.Spokes[index]
		if spoke.Side != SideB || spoke.Type != SpokeTypeNonDrive || spoke.HubHoleID != index || spoke.RimHoleID != 1+index*3 {
			t.Fatalf("uniform 18H non-drive spoke %d = %+v", index, spoke)
		}
	}
	for index := 0; index < 12; index++ {
		spoke := uniform18.Spokes[6+index]
		targetIndex := index - 2
		wantType := SpokeTypeTrailing
		if index%2 == 0 {
			targetIndex = index + 2
			wantType = SpokeTypeLeading
		}
		wantRimHoleID := runningDriveRimHoleIDs[mod(targetIndex, len(runningDriveRimHoleIDs))]
		if spoke.Side != SideA || spoke.Type != wantType || spoke.HubHoleID != index || spoke.RimHoleID != wantRimHoleID {
			t.Fatalf("uniform 18H drive spoke %d = %+v, want side=%s type=%s hub=%d rim=%d", index, spoke, SideA, wantType, index, wantRimHoleID)
		}
	}
}

func TestG3SelectionIsExplicitlySeparateFromItsHoleCount(t *testing.T) {
	catalog := NewDefaultCatalog()
	g3, err := catalog.Get("21h-g3-2to1")
	if err != nil {
		t.Fatal(err)
	}
	if g3.Selection != "21_g3" || g3.HoleCount != 21 || g3.Distribution != DistributionG32To1 {
		t.Fatalf("G3 identity must be independent from hole count: %+v", g3)
	}

	legacyHoleCountSelection := "21"
	if _, err := catalog.Validate(ValidateRequest{TopologyID: g3.ID, Selection: &legacyHoleCountSelection}); !errors.Is(err, ErrTopologyMismatch) {
		t.Fatalf("Validate() with ambiguous selection error = %v, want ErrTopologyMismatch", err)
	}
}

func TestCatalogRejectsChangedCanonicalSpokeAssignment(t *testing.T) {
	source := NewDefaultCatalog().List()
	mutated := source[0]
	mutated.Spokes[0].RimHoleID = mutated.Spokes[1].RimHoleID
	if _, err := NewCatalog([]Topology{mutated}); err == nil {
		t.Fatal("NewCatalog() error = nil, want changed mapping to fail")
	} else if !errors.Is(err, ErrInvalidTopology) {
		t.Fatalf("NewCatalog() error = %v, want ErrInvalidTopology", err)
	}
}

func TestCatalogReturnsIndependentCopies(t *testing.T) {
	catalog := NewDefaultCatalog()
	first, err := catalog.Get("21h-g3-2to1")
	if err != nil {
		t.Fatal(err)
	}
	first.RimHoles[0].Side = SideB
	first.Spokes[0].RimHoleID = 999

	second, err := catalog.Get("21h-g3-2to1")
	if err != nil {
		t.Fatal(err)
	}
	if second.RimHoles[0].Side != SideA || second.Spokes[0].RimHoleID == 999 {
		t.Fatal("catalog returned mutable internal topology data")
	}
}

func TestValidateChecksOptionalSelectionFields(t *testing.T) {
	catalog := NewDefaultCatalog()
	selection := "24_2to1"
	holes := 24
	cross := 2
	distribution := DistributionUniform2To1
	topology, err := catalog.Validate(ValidateRequest{
		TopologyID:   "24h-uniform-2to1",
		Selection:    &selection,
		HoleCount:    &holes,
		Cross:        &cross,
		Distribution: &distribution,
	})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if topology.ID != "24h-uniform-2to1" {
		t.Fatalf("validated topology ID = %q", topology.ID)
	}

	wrongCross := 3
	if _, err := catalog.Validate(ValidateRequest{TopologyID: topology.ID, Cross: &wrongCross}); !errors.Is(err, ErrTopologyMismatch) {
		t.Fatalf("Validate() error = %v, want ErrTopologyMismatch", err)
	}
}

func TestValidateReturnsIndependentUniform18HTwoToOneMapping(t *testing.T) {
	catalog := NewDefaultCatalog()
	selection := "18_2to1"
	holes := 18
	cross := 2
	distribution := DistributionUniform2To1
	topology, err := catalog.Validate(ValidateRequest{
		TopologyID:   "18h-uniform-2to1",
		Selection:    &selection,
		HoleCount:    &holes,
		Cross:        &cross,
		Distribution: &distribution,
	})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if topology.DisplayLayout != DisplayGeometryLayoutUniform18H2To1 || len(topology.HubHolesA) != 12 || len(topology.HubHolesB) != 6 {
		t.Fatalf("unexpected uniform 18H topology: %+v", topology)
	}
	for index, spoke := range topology.Spokes {
		if index < 6 && spoke.Side != SideB {
			t.Fatalf("spoke %d side = %q, want B for first six radial spokes", index, spoke.Side)
		}
		if index >= 6 && spoke.Side != SideA {
			t.Fatalf("spoke %d side = %q, want A after non-drive spokes", index, spoke.Side)
		}
	}
}
