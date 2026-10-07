package wheelsetlacing

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Distribution describes how rim and flange holes are split between the two
// sides. It is a topology fact; it is not a tension ratio or a mechanical
// efficiency value.
type Distribution string

const (
	DistributionSymmetric1To1 Distribution = "symmetric_1to1"
	DistributionUniform2To1   Distribution = "uniform_2to1"
	DistributionG32To1        Distribution = "g3_2to1"
)

// DisplayGeometryLayout identifies the exact coordinate generator used by the
// display-geometry projection. It is deliberately separate from Distribution:
// multiple 2:1 topologies can share a ratio while requiring different rim-hole
// geometry. A new topology must register a layout and a matching generator;
// unknown layouts fail closed instead of falling back to symmetric geometry.
type DisplayGeometryLayout string

const (
	DisplayGeometryLayoutSymmetric1To1              DisplayGeometryLayout = "symmetric_1to1"
	DisplayGeometryLayoutUniform2To1                DisplayGeometryLayout = "uniform_2to1"
	DisplayGeometryLayoutUniform18H2To1             DisplayGeometryLayout = "uniform_18h_2to1"
	DisplayGeometryLayoutG3TwentyOneHoleTriplet2To1 DisplayGeometryLayout = "g3_21h_triplet_2to1"
)

type Side string

const (
	SideA Side = "A"
	SideB Side = "B"
)

type SpokeType string

const (
	SpokeTypeLeading  SpokeType = "leading"
	SpokeTypeTrailing SpokeType = "trailing"
	SpokeTypeNonDrive SpokeType = "nondrive"
)

// SpokeHeadStyle identifies the hub-side end geometry required by a topology.
// It is separate from the lacing map: conventional topologies retain their
// existing J-bend mappings, paired-hole straight-pull variants share each
// anchor between two spokes, and 21-hole G3 retains its special spoke anchors.
type SpokeHeadStyle string

const (
	SpokeHeadStyleJBend        SpokeHeadStyle = "j_bend"
	SpokeHeadStyleStraightPull SpokeHeadStyle = "straight_pull"
)

var (
	ErrUnknownTopology  = errors.New("unknown wheelset lacing topology")
	ErrInvalidTopology  = errors.New("wheelset lacing topology is invalid")
	ErrTopologyMismatch = errors.New("wheelset lacing topology selection does not match the requested fields")
	ErrInvalidRequest   = errors.New("wheelset lacing topology request is invalid")
)

// Hole is a stable integer identifier and its side assignment. It deliberately
// has no coordinate: the page's SVG display radius is not a physical input.
type Hole struct {
	ID   int  `json:"id"`
	Side Side `json:"side"`
}

// SpokeMapping is the complete discrete mapping for one spoke. HubHoleID is
// scoped to Side, while RimHoleID is scoped to the complete rim hole list.
type SpokeMapping struct {
	ID        int       `json:"id"`
	Side      Side      `json:"side"`
	Type      SpokeType `json:"type"`
	HubHoleID int       `json:"hub_hole_id"`
	RimHoleID int       `json:"rim_hole_id"`
}

// Topology is the backend's read-only topology contract. It contains only
// discrete hole assignments and no ERD/PCD, lengths, angles, or force values.
type Topology struct {
	ID        string `json:"topology_id"`
	Selection string `json:"selection"`
	HoleCount int    `json:"hole_count"`
	// Cross is selectable for conventional lacing patterns. For G3, its
	// registered value is topology metadata, not a J-bend cross selector.
	Cross          int                   `json:"cross"`
	Distribution   Distribution          `json:"distribution"`
	DisplayLayout  DisplayGeometryLayout `json:"display_layout"`
	SpokeHeadStyle SpokeHeadStyle        `json:"spoke_head_style"`
	RimHoles       []Hole                `json:"rim_holes"`
	HubHolesA      []Hole                `json:"hub_holes_a"`
	HubHolesB      []Hole                `json:"hub_holes_b"`
	Spokes         []SpokeMapping        `json:"spokes"`
}

// Catalog owns an immutable, validated set of topology facts.
type Catalog struct {
	topologies []Topology
	byID       map[string]int
}

// NewCatalog validates and indexes a topology set. The input is copied so a
// caller cannot mutate the catalog after construction.
func NewCatalog(topologies []Topology) (*Catalog, error) {
	catalog := &Catalog{
		topologies: make([]Topology, len(topologies)),
		byID:       make(map[string]int, len(topologies)),
	}
	for index := range topologies {
		copied := cloneTopology(topologies[index])
		if err := validateTopology(copied); err != nil {
			return nil, fmt.Errorf("%w: [CRITICAL] topology %q: %w", ErrInvalidTopology, copied.ID, err)
		}
		key := strings.TrimSpace(copied.ID)
		if key == "" {
			return nil, fmt.Errorf("%w: [CRITICAL] topology_id is required", ErrInvalidTopology)
		}
		if _, exists := catalog.byID[key]; exists {
			return nil, fmt.Errorf("%w: duplicate topology_id %q", ErrInvalidTopology, key)
		}
		copied.ID = key
		catalog.byID[key] = index
		catalog.topologies[index] = copied
	}
	return catalog, nil
}

// NewDefaultCatalog builds the topology contract used by the public API.
func NewDefaultCatalog() *Catalog {
	set := make([]Topology, 0, 26)
	for _, holes := range []int{16, 20, 24, 28, 32, 36} {
		set = append(set, buildSymmetricTopology(holes, 0))
		for _, cross := range supportedSymmetricCrosses(holes) {
			if cross == 0 {
				continue
			}
			set = append(set, buildSymmetricTopology(holes, cross))
		}
	}
	set = append(set, buildTwentyOneHoleG3Topology(), buildUniformTwoToOneTopology(), buildUniform18TwoToOneTopology())
	straightPullVariants := make([]Topology, 0, len(set)-1)
	for _, topology := range set {
		if topology.SpokeHeadStyle != SpokeHeadStyleJBend {
			continue
		}
		straightPullTopology, err := buildStraightPullPairedHoleTopologyVariant(topology)
		if err != nil {
			panic(fmt.Errorf("build straight-pull variant for %q: %w", topology.ID, err))
		}
		straightPullVariants = append(straightPullVariants, straightPullTopology)
	}
	set = append(set, straightPullVariants...)
	catalog, err := NewCatalog(set)
	if err != nil {
		// The checked-in contract is a program invariant. Serving a partial or
		// malformed topology directory would hide an engineering data error.
		panic(err)
	}
	return catalog
}

// List returns a deep copy in stable ID order. Stable ordering keeps the API
// response and its ETag deterministic across processes.
func (c *Catalog) List() []Topology {
	if c == nil {
		return nil
	}
	result := make([]Topology, len(c.topologies))
	for index := range c.topologies {
		result[index] = cloneTopology(c.topologies[index])
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

// Get returns one topology by its canonical ID.
func (c *Catalog) Get(id string) (Topology, error) {
	if c == nil {
		return Topology{}, fmt.Errorf("%w: catalog is nil", ErrInvalidTopology)
	}
	key := strings.TrimSpace(id)
	index, ok := c.byID[key]
	if !ok {
		return Topology{}, fmt.Errorf("%w: %q", ErrUnknownTopology, key)
	}
	return cloneTopology(c.topologies[index]), nil
}

// ValidateRequest is the strict selection contract accepted by the API. The
// topology_id identifies one canonical model. Optional fields are consistency
// checks and are rejected when they disagree with the catalog fact.
type ValidateRequest struct {
	TopologyID   string        `json:"topology_id"`
	Selection    *string       `json:"selection,omitempty"`
	HoleCount    *int          `json:"hole_count,omitempty"`
	Cross        *int          `json:"cross,omitempty"`
	Distribution *Distribution `json:"distribution,omitempty"`
}

func (c *Catalog) Validate(request ValidateRequest) (Topology, error) {
	if strings.TrimSpace(request.TopologyID) == "" {
		return Topology{}, fmt.Errorf("%w: topology_id is required", ErrInvalidRequest)
	}
	topology, err := c.Get(request.TopologyID)
	if err != nil {
		return Topology{}, err
	}
	if request.Selection != nil && strings.TrimSpace(*request.Selection) != topology.Selection {
		return Topology{}, fmt.Errorf("%w: selection %q does not match %q", ErrTopologyMismatch, *request.Selection, topology.Selection)
	}
	if request.HoleCount != nil && *request.HoleCount != topology.HoleCount {
		return Topology{}, fmt.Errorf("%w: hole_count %d does not match %d", ErrTopologyMismatch, *request.HoleCount, topology.HoleCount)
	}
	if request.Cross != nil && *request.Cross != topology.Cross {
		return Topology{}, fmt.Errorf("%w: cross %d does not match %d", ErrTopologyMismatch, *request.Cross, topology.Cross)
	}
	if request.Distribution != nil && *request.Distribution != topology.Distribution {
		return Topology{}, fmt.Errorf("%w: distribution %q does not match %q", ErrTopologyMismatch, *request.Distribution, topology.Distribution)
	}
	return topology, nil
}

func supportedSymmetricCrosses(holes int) []int {
	switch holes {
	case 16:
		return []int{0, 1}
	case 20:
		return []int{0, 1, 2}
	case 24:
		return []int{0, 1, 2, 3}
	case 28:
		return []int{0, 1, 2, 3}
	case 32:
		return []int{0, 1, 2, 3, 4}
	case 36:
		return []int{0, 1, 2, 3, 4}
	default:
		return nil
	}
}

func symmetricSelection(holes, cross int) string {
	return fmt.Sprintf("%dh-symmetric-1to1-%dx", holes, cross)
}

func buildSymmetricTopology(holes, cross int) Topology {
	flangeCount := holes / 2
	topology := Topology{
		ID:             symmetricSelection(holes, cross),
		Selection:      fmt.Sprintf("%d", holes),
		HoleCount:      holes,
		Cross:          cross,
		Distribution:   DistributionSymmetric1To1,
		DisplayLayout:  DisplayGeometryLayoutSymmetric1To1,
		SpokeHeadStyle: SpokeHeadStyleJBend,
		RimHoles:       make([]Hole, 0, holes),
		HubHolesA:      make([]Hole, 0, flangeCount),
		HubHolesB:      make([]Hole, 0, flangeCount),
		Spokes:         make([]SpokeMapping, 0, holes),
	}
	for index := 0; index < holes; index++ {
		side := SideA
		if index%2 == 1 {
			side = SideB
		}
		topology.RimHoles = append(topology.RimHoles, Hole{ID: index, Side: side})
	}
	for index := 0; index < flangeCount; index++ {
		topology.HubHolesA = append(topology.HubHolesA, Hole{ID: index, Side: SideA})
		topology.HubHolesB = append(topology.HubHolesB, Hole{ID: index, Side: SideB})
	}
	rimA := make([]Hole, 0, flangeCount)
	rimB := make([]Hole, 0, flangeCount)
	for _, hole := range topology.RimHoles {
		if hole.Side == SideA {
			rimA = append(rimA, hole)
		} else {
			rimB = append(rimB, hole)
		}
	}
	for index := 0; index < flangeCount; index++ {
		if cross == 0 {
			topology.Spokes = append(topology.Spokes,
				SpokeMapping{ID: len(topology.Spokes), Side: SideA, Type: SpokeTypeLeading, HubHoleID: index, RimHoleID: rimA[index].ID},
				SpokeMapping{ID: len(topology.Spokes) + 1, Side: SideB, Type: SpokeTypeNonDrive, HubHoleID: index, RimHoleID: rimB[index].ID},
			)
			continue
		}
		isLeading := index%2 == 0
		targetIndex := index - cross
		spokeType := SpokeTypeTrailing
		if isLeading {
			targetIndex = index + cross
			spokeType = SpokeTypeLeading
		}
		targetIndex = mod(targetIndex, flangeCount)
		topology.Spokes = append(topology.Spokes, SpokeMapping{
			ID: len(topology.Spokes), Side: SideA, Type: spokeType, HubHoleID: index, RimHoleID: rimA[targetIndex].ID,
		})
	}
	for index := 0; index < flangeCount && cross != 0; index++ {
		isLeading := index%2 == 0
		targetIndex := index - cross
		if isLeading {
			targetIndex = index + cross
		}
		topology.Spokes = append(topology.Spokes, SpokeMapping{
			ID: len(topology.Spokes), Side: SideB, Type: SpokeTypeNonDrive, HubHoleID: index, RimHoleID: rimB[mod(targetIndex, flangeCount)].ID,
		})
	}
	return topology
}

func buildTwentyOneHoleG3Topology() Topology {
	const groups = 7
	topology := Topology{
		ID:             "21h-g3-2to1",
		Selection:      "21_g3",
		HoleCount:      21,
		Cross:          2,
		Distribution:   DistributionG32To1,
		DisplayLayout:  DisplayGeometryLayoutG3TwentyOneHoleTriplet2To1,
		SpokeHeadStyle: SpokeHeadStyleStraightPull,
		RimHoles:       make([]Hole, 0, 21),
		HubHolesA:      make([]Hole, 0, groups),
		HubHolesB:      make([]Hole, 0, 7),
		Spokes:         make([]SpokeMapping, 0, 21),
	}
	for group := 0; group < groups; group++ {
		topology.RimHoles = append(topology.RimHoles,
			Hole{ID: group * 3, Side: SideA},
			Hole{ID: group*3 + 1, Side: SideB},
			Hole{ID: group*3 + 2, Side: SideA},
		)
	}
	for index := 0; index < groups; index++ {
		topology.HubHolesA = append(topology.HubHolesA, Hole{ID: index, Side: SideA})
	}
	for index := 0; index < 7; index++ {
		topology.HubHolesB = append(topology.HubHolesB, Hole{ID: index, Side: SideB})
	}
	for group := 0; group < groups; group++ {
		topology.Spokes = append(topology.Spokes, SpokeMapping{
			ID: len(topology.Spokes), Side: SideB, Type: SpokeTypeNonDrive, HubHoleID: group, RimHoleID: group*3 + 1,
		})
	}
	// G3 is a straight-pull topology: each three-hole A-B-A rim group is kept
	// together. The two A-side spokes in a group share one straight-pull flange
	// hole and extend in opposite tangential directions; the B-side spoke uses
	// the group's matching non-drive flange hole. There are seven drive-side
	// flange holes for fourteen drive-side spokes, not fourteen drive-side
	// flange holes.
	for group := 0; group < groups; group++ {
		topology.Spokes = append(topology.Spokes,
			SpokeMapping{ID: len(topology.Spokes), Side: SideA, Type: SpokeTypeTrailing, HubHoleID: group, RimHoleID: group * 3},
			SpokeMapping{ID: len(topology.Spokes) + 1, Side: SideA, Type: SpokeTypeLeading, HubHoleID: group, RimHoleID: group*3 + 2},
		)
	}
	return topology
}

func buildUniformTwoToOneTopology() Topology {
	return buildUniformTwoToOneTopologyWithExplicitHoleCountsAndDisplayLayout(
		"24h-uniform-2to1",
		"24_2to1",
		24,
		16,
		8,
		DisplayGeometryLayoutUniform2To1,
	)
}

// buildUniform18TwoToOneTopology is the non-G3 18H variant: its rim holes are
// uniformly spaced, with 12 drive-side and 6 non-drive-side assignments.
func buildUniform18TwoToOneTopology() Topology {
	return buildUniformTwoToOneTopologyWithExplicitHoleCountsAndDisplayLayout(
		"18h-uniform-2to1",
		"18_2to1",
		18,
		12,
		6,
		DisplayGeometryLayoutUniform18H2To1,
	)
}

func buildUniformTwoToOneTopologyWithExplicitHoleCountsAndDisplayLayout(id, selection string, total, driveSideHoleCount, nonDriveSideHoleCount int, displayLayout DisplayGeometryLayout) Topology {
	topology := Topology{
		ID:             id,
		Selection:      selection,
		HoleCount:      total,
		Cross:          2,
		Distribution:   DistributionUniform2To1,
		DisplayLayout:  displayLayout,
		SpokeHeadStyle: SpokeHeadStyleJBend,
		RimHoles:       make([]Hole, 0, total),
		HubHolesA:      make([]Hole, 0, driveSideHoleCount),
		HubHolesB:      make([]Hole, 0, nonDriveSideHoleCount),
		Spokes:         make([]SpokeMapping, 0, total),
	}
	for index := 0; index < total; index++ {
		side := SideA
		if index%3 == 1 {
			side = SideB
		}
		topology.RimHoles = append(topology.RimHoles, Hole{ID: index, Side: side})
	}
	rimA := make([]Hole, 0, 16)
	rimB := make([]Hole, 0, 8)
	for _, hole := range topology.RimHoles {
		if hole.Side == SideA {
			rimA = append(rimA, hole)
		} else {
			rimB = append(rimB, hole)
		}
	}
	for index := 0; index < driveSideHoleCount; index++ {
		topology.HubHolesA = append(topology.HubHolesA, Hole{ID: index, Side: SideA})
	}
	for index := 0; index < nonDriveSideHoleCount; index++ {
		topology.HubHolesB = append(topology.HubHolesB, Hole{ID: index, Side: SideB})
	}
	for index := 0; index < nonDriveSideHoleCount; index++ {
		topology.Spokes = append(topology.Spokes, SpokeMapping{
			ID: len(topology.Spokes), Side: SideB, Type: SpokeTypeNonDrive, HubHoleID: index, RimHoleID: rimB[index].ID,
		})
	}
	for index := 0; index < driveSideHoleCount; index++ {
		targetIndex := index - 2
		spokeType := SpokeTypeTrailing
		if index%2 == 0 {
			targetIndex = index + 2
			spokeType = SpokeTypeLeading
		}
		topology.Spokes = append(topology.Spokes, SpokeMapping{
			ID: len(topology.Spokes), Side: SideA, Type: spokeType, HubHoleID: index, RimHoleID: rimA[mod(targetIndex, driveSideHoleCount)].ID,
		})
	}
	return topology
}

func validateTopology(topology Topology) error {
	if strings.TrimSpace(topology.ID) == "" {
		return errors.New("topology_id is required")
	}
	if topology.HoleCount <= 0 {
		return errors.New("hole_count must be positive")
	}
	if topology.Cross < 0 || topology.Cross > 4 {
		return fmt.Errorf("cross must be an integer from 0 to 4, received %d", topology.Cross)
	}
	if topology.Distribution != DistributionSymmetric1To1 && topology.Distribution != DistributionUniform2To1 && topology.Distribution != DistributionG32To1 {
		return fmt.Errorf("unsupported distribution %q", topology.Distribution)
	}
	if topology.SpokeHeadStyle != SpokeHeadStyleJBend && topology.SpokeHeadStyle != SpokeHeadStyleStraightPull {
		return fmt.Errorf("unsupported spoke head style %q", topology.SpokeHeadStyle)
	}
	if topology.DisplayLayout != DisplayGeometryLayoutSymmetric1To1 && topology.DisplayLayout != DisplayGeometryLayoutUniform2To1 && topology.DisplayLayout != DisplayGeometryLayoutUniform18H2To1 && topology.DisplayLayout != DisplayGeometryLayoutG3TwentyOneHoleTriplet2To1 {
		return fmt.Errorf("unsupported display layout %q", topology.DisplayLayout)
	}
	expected := expectedTopology(topology.Selection, topology.Cross, topology.SpokeHeadStyle)
	if expected == nil {
		return fmt.Errorf("unsupported selection %q", topology.Selection)
	}
	if topology.HoleCount != expected.HoleCount || topology.Distribution != expected.Distribution || topology.DisplayLayout != expected.DisplayLayout || topology.Cross != expected.Cross || topology.SpokeHeadStyle != expected.SpokeHeadStyle {
		return fmt.Errorf("selection %q does not match hole_count=%d, cross=%d, distribution=%q, display_layout=%q, spoke_head_style=%q", topology.Selection, topology.HoleCount, topology.Cross, topology.Distribution, topology.DisplayLayout, topology.SpokeHeadStyle)
	}
	if len(topology.RimHoles) != len(expected.RimHoles) || len(topology.HubHolesA) != len(expected.HubHolesA) || len(topology.HubHolesB) != len(expected.HubHolesB) || len(topology.Spokes) != len(expected.Spokes) {
		return fmt.Errorf("hole or spoke counts do not match the selection")
	}
	if err := validateHoles(topology.RimHoles, "rim"); err != nil {
		return err
	}
	if err := validateHoles(topology.HubHolesA, "hub A"); err != nil {
		return err
	}
	if err := validateHoles(topology.HubHolesB, "hub B"); err != nil {
		return err
	}
	if err := validateSpokeMappings(topology); err != nil {
		return err
	}
	return compareExpectedSpokes(topology.Spokes, expected.Spokes)
}

func expectedTopology(selection string, cross int, spokeHeadStyle SpokeHeadStyle) *Topology {
	var expected *Topology
	switch selection {
	case "21_g3":
		if cross != 2 {
			return nil
		}
		g3Topology := buildTwentyOneHoleG3Topology()
		expected = &g3Topology
	case "24_2to1":
		if cross != 2 {
			return nil
		}
		uniformTopology := buildUniformTwoToOneTopology()
		expected = &uniformTopology
	case "18_2to1":
		if cross != 2 {
			return nil
		}
		uniform18Topology := buildUniform18TwoToOneTopology()
		expected = &uniform18Topology
	default:
		var holes int
		if _, err := fmt.Sscanf(selection, "%d", &holes); err != nil || fmt.Sprintf("%d", holes) != selection {
			return nil
		}
		allowed := supportedSymmetricCrosses(holes)
		for _, candidate := range allowed {
			if candidate == cross {
				symmetricTopology := buildSymmetricTopology(holes, cross)
				expected = &symmetricTopology
				break
			}
		}
	}
	if expected == nil {
		return nil
	}
	if expected.SpokeHeadStyle == spokeHeadStyle {
		return expected
	}
	if expected.SpokeHeadStyle != SpokeHeadStyleJBend || spokeHeadStyle != SpokeHeadStyleStraightPull {
		return nil
	}
	straightPullTopology, err := buildStraightPullPairedHoleTopologyVariant(*expected)
	if err != nil {
		return nil
	}
	return &straightPullTopology
}

// buildStraightPullPairedHoleTopologyVariant preserves every rim assignment
// while pairing consecutive J-bend flange holes into one straight-pull anchor
// shared by two spokes. G3 has its own registered straight-pull mapping.
func buildStraightPullPairedHoleTopologyVariant(source Topology) (Topology, error) {
	if source.SpokeHeadStyle != SpokeHeadStyleJBend || source.Distribution == DistributionG32To1 {
		return Topology{}, fmt.Errorf("topology %q is not a conventional J-bend topology", source.ID)
	}
	straightPullTopology := cloneTopology(source)
	straightPullTopology.ID += "-straight-pull"
	straightPullTopology.SpokeHeadStyle = SpokeHeadStyleStraightPull
	straightPullTopology.HubHolesA = make([]Hole, 0, len(source.HubHolesA)/2)
	straightPullTopology.HubHolesB = make([]Hole, 0, len(source.HubHolesB)/2)
	hubHoleIDByOriginalKey := make(map[string]int, len(source.HubHolesA)+len(source.HubHolesB))

	for _, sideHoles := range []struct {
		side  Side
		holes []Hole
	}{
		{side: SideA, holes: source.HubHolesA},
		{side: SideB, holes: source.HubHolesB},
	} {
		orderedHoles := append([]Hole(nil), sideHoles.holes...)
		sort.Slice(orderedHoles, func(i, j int) bool { return orderedHoles[i].ID < orderedHoles[j].ID })
		if len(orderedHoles)%2 != 0 {
			return Topology{}, fmt.Errorf("side %s has an odd spoke-hole count %d", sideHoles.side, len(orderedHoles))
		}
		for pairIndex := 0; pairIndex < len(orderedHoles)/2; pairIndex++ {
			straightPullHole := Hole{ID: pairIndex, Side: sideHoles.side}
			if sideHoles.side == SideA {
				straightPullTopology.HubHolesA = append(straightPullTopology.HubHolesA, straightPullHole)
			} else {
				straightPullTopology.HubHolesB = append(straightPullTopology.HubHolesB, straightPullHole)
			}
			firstOriginalHoleID := orderedHoles[pairIndex*2].ID
			secondOriginalHoleID := orderedHoles[pairIndex*2+1].ID
			hubHoleIDByOriginalKey[hubKey(sideHoles.side, firstOriginalHoleID)] = pairIndex
			hubHoleIDByOriginalKey[hubKey(sideHoles.side, secondOriginalHoleID)] = pairIndex
		}
	}

	for index := range straightPullTopology.Spokes {
		spoke := &straightPullTopology.Spokes[index]
		pairedHoleID, exists := hubHoleIDByOriginalKey[hubKey(spoke.Side, spoke.HubHoleID)]
		if !exists {
			return Topology{}, fmt.Errorf("spoke %d does not map to a paired flange hole", spoke.ID)
		}
		spoke.HubHoleID = pairedHoleID
	}
	return straightPullTopology, nil
}

func validateHoles(holes []Hole, label string) error {
	seen := make(map[int]struct{}, len(holes))
	for _, hole := range holes {
		if hole.ID < 0 {
			return fmt.Errorf("%s hole %d has a negative id", label, hole.ID)
		}
		if hole.Side != SideA && hole.Side != SideB {
			return fmt.Errorf("%s hole %d has unsupported side %q", label, hole.ID, hole.Side)
		}
		if _, exists := seen[hole.ID]; exists {
			return fmt.Errorf("%s holes contain duplicate id %d", label, hole.ID)
		}
		seen[hole.ID] = struct{}{}
	}
	return nil
}

func validateSpokeMappings(topology Topology) error {
	rimByID := make(map[int]Hole, len(topology.RimHoles))
	for _, hole := range topology.RimHoles {
		rimByID[hole.ID] = hole
	}
	hubByKey := make(map[string]Hole, len(topology.HubHolesA)+len(topology.HubHolesB))
	for _, hole := range topology.HubHolesA {
		hubByKey[hubKey(SideA, hole.ID)] = hole
	}
	for _, hole := range topology.HubHolesB {
		hubByKey[hubKey(SideB, hole.ID)] = hole
	}
	spokeIDs := make(map[int]struct{}, len(topology.Spokes))
	rimUsage := make(map[int]int, len(topology.RimHoles))
	hubUsage := make(map[string]int, len(hubByKey))
	for _, spoke := range topology.Spokes {
		if _, exists := spokeIDs[spoke.ID]; exists {
			return fmt.Errorf("spokes contain duplicate id %d", spoke.ID)
		}
		spokeIDs[spoke.ID] = struct{}{}
		if spoke.Side != SideA && spoke.Side != SideB {
			return fmt.Errorf("spoke %d has unsupported side %q", spoke.ID, spoke.Side)
		}
		if spoke.Type != SpokeTypeLeading && spoke.Type != SpokeTypeTrailing && spoke.Type != SpokeTypeNonDrive {
			return fmt.Errorf("spoke %d has unsupported type %q", spoke.ID, spoke.Type)
		}
		if spoke.Side == SideB && spoke.Type != SpokeTypeNonDrive {
			return fmt.Errorf("spoke %d on side B must be nondrive", spoke.ID)
		}
		if spoke.Side == SideA && spoke.Type == SpokeTypeNonDrive {
			return fmt.Errorf("spoke %d on side A cannot be nondrive", spoke.ID)
		}
		rim, exists := rimByID[spoke.RimHoleID]
		if !exists || rim.Side != spoke.Side {
			return fmt.Errorf("spoke %d maps to an invalid rim hole %d", spoke.ID, spoke.RimHoleID)
		}
		hub, exists := hubByKey[hubKey(spoke.Side, spoke.HubHoleID)]
		if !exists || hub.Side != spoke.Side {
			return fmt.Errorf("spoke %d maps to an invalid hub hole %s:%d", spoke.ID, spoke.Side, spoke.HubHoleID)
		}
		rimUsage[spoke.RimHoleID]++
		hubUsage[hubKey(spoke.Side, spoke.HubHoleID)]++
	}
	for _, hole := range topology.RimHoles {
		if rimUsage[hole.ID] != 1 {
			return fmt.Errorf("rim hole %d must map to exactly one spoke", hole.ID)
		}
	}
	for key := range hubByKey {
		hub := hubByKey[key]
		wantSpokesPerHole := 1
		// Conventional straight-pull variants pair both flange sides. G3 is
		// asymmetric: only the fourteen-spoke drive side has two spokes per
		// anchor; its seven non-drive anchors remain one spoke each.
		if topology.SpokeHeadStyle == SpokeHeadStyleStraightPull &&
			(topology.Distribution != DistributionG32To1 || hub.Side == SideA) {
			wantSpokesPerHole = 2
		}
		if hubUsage[key] != wantSpokesPerHole {
			return fmt.Errorf("hub hole %s must map to exactly %d spoke(s)", key, wantSpokesPerHole)
		}
	}
	return nil
}

func compareExpectedSpokes(actual, expected []SpokeMapping) error {
	if len(actual) != len(expected) {
		return errors.New("spoke mapping count does not match the canonical assignment")
	}
	byID := make(map[int]SpokeMapping, len(actual))
	for _, spoke := range actual {
		byID[spoke.ID] = spoke
	}
	for _, want := range expected {
		got, ok := byID[want.ID]
		if !ok || got.Side != want.Side || got.Type != want.Type || got.HubHoleID != want.HubHoleID || got.RimHoleID != want.RimHoleID {
			return fmt.Errorf("spoke %d does not match the canonical hole assignment", want.ID)
		}
	}
	return nil
}

func hubKey(side Side, id int) string {
	return fmt.Sprintf("%s:%d", side, id)
}

func mod(value, base int) int {
	if base <= 0 {
		return 0
	}
	value %= base
	if value < 0 {
		value += base
	}
	return value
}

func cloneTopology(source Topology) Topology {
	clone := source
	clone.RimHoles = append([]Hole(nil), source.RimHoles...)
	clone.HubHolesA = append([]Hole(nil), source.HubHolesA...)
	clone.HubHolesB = append([]Hole(nil), source.HubHolesB...)
	clone.Spokes = append([]SpokeMapping(nil), source.Spokes...)
	return clone
}
