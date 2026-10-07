package wheelsetlacing

// DisplayGeometryCoordinate is an SVG-plane point used for the separated
// straight-pull flange projection. It is not a physical 3D coordinate.
type DisplayGeometryCoordinate struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// DisplayGeometryStraightPullProjection keeps each flange's spoke anchors in
// its own axial display plane while the rim remains on the wheel center plane.
// The spoke mapping is shared with the topology; only its display endpoints
// receive the backend-calculated flange offset projection.
type DisplayGeometryStraightPullProjection struct {
	FlangeCenterA DisplayGeometryCoordinate `json:"flange_center_a"`
	FlangeCenterB DisplayGeometryCoordinate `json:"flange_center_b"`
	HubHolesA     []DisplayGeometryPoint    `json:"hub_holes_a"`
	HubHolesB     []DisplayGeometryPoint    `json:"hub_holes_b"`
	Spokes        []DisplayGeometrySpoke    `json:"spokes"`
}

const (
	wheelsetLacingStraightPullProjectionXFactor = 0.30
	wheelsetLacingStraightPullProjectionYFactor = 0.18
)

// buildWheelsetLacingStraightPullDisplayGeometryProjection separates the two
// straight-pull flange anchor planes using the backend's axial profile. The
// projection is schematic and does not change spoke lengths or lacing angles.
func buildWheelsetLacingStraightPullDisplayGeometryProjection(
	profile DisplayGeometryFlangeProfile,
	hubHolesA, hubHolesB []DisplayGeometryPoint,
	spokes []DisplayGeometrySpoke,
) DisplayGeometryStraightPullProjection {
	flangeCenterA := wheelsetLacingStraightPullProjectionCenter(profile.FlangeAX)
	flangeCenterB := wheelsetLacingStraightPullProjectionCenter(profile.FlangeBX)
	projectedHubHolesA := shiftWheelsetLacingDisplayGeometryPoints(hubHolesA, flangeCenterA)
	projectedHubHolesB := shiftWheelsetLacingDisplayGeometryPoints(hubHolesB, flangeCenterB)
	projectedHubPointByKey := make(map[string]DisplayGeometryPoint, len(projectedHubHolesA)+len(projectedHubHolesB))
	for _, point := range projectedHubHolesA {
		projectedHubPointByKey[displayGeometryHubKey(SideA, point.ID)] = point
	}
	for _, point := range projectedHubHolesB {
		projectedHubPointByKey[displayGeometryHubKey(SideB, point.ID)] = point
	}
	projectedSpokes := make([]DisplayGeometrySpoke, 0, len(spokes))
	for _, spoke := range spokes {
		projectedSpoke := spoke
		projectedSpoke.Hub = projectedHubPointByKey[displayGeometryHubKey(spoke.Side, spoke.Hub.ID)]
		projectedSpokes = append(projectedSpokes, projectedSpoke)
	}
	return DisplayGeometryStraightPullProjection{
		FlangeCenterA: flangeCenterA,
		FlangeCenterB: flangeCenterB,
		HubHolesA:     projectedHubHolesA,
		HubHolesB:     projectedHubHolesB,
		Spokes:        projectedSpokes,
	}
}

func wheelsetLacingStraightPullProjectionCenter(axialProfileX float64) DisplayGeometryCoordinate {
	return DisplayGeometryCoordinate{
		X: roundWheelsetLacingDisplayGeometryValue(axialProfileX*wheelsetLacingStraightPullProjectionXFactor, 2),
		Y: roundWheelsetLacingDisplayGeometryValue(-axialProfileX*wheelsetLacingStraightPullProjectionYFactor, 2),
	}
}

func shiftWheelsetLacingDisplayGeometryPoints(
	points []DisplayGeometryPoint,
	center DisplayGeometryCoordinate,
) []DisplayGeometryPoint {
	shifted := make([]DisplayGeometryPoint, 0, len(points))
	for _, point := range points {
		point.X = roundWheelsetLacingDisplayGeometryValue(point.X+center.X, 2)
		point.Y = roundWheelsetLacingDisplayGeometryValue(point.Y+center.Y, 2)
		shifted = append(shifted, point)
	}
	return shifted
}
