package spoke

import (
	"encoding/json"
	"testing"
)

func TestCalcRequestAcceptsPhysicalSnakeCaseFields(t *testing.T) {
	var request CalcRequest
	err := json.Unmarshal([]byte(`{
		"wheel_position": "front",
		"topology_id": "21h-g3-2to1",
		"spoke_count": 24,
		"crossing": 3,
		"g3_rim_hole_spacing_a_to_b_degrees": 2.5,
		"g3_rim_hole_spacing_b_to_a_degrees": 8.5,
		"g3_rim_hole_spacing_a_to_next_group_a_degrees": 40.428571,
		"spoke_head_type": "straight_pull",
		"spoke_hole_diameter_mm": 2.5,
		"straight_pull_tangent_offset_mm": 1.2,
		"spoke_profile": "bladed_0_9x2_2",
		"target_tension_n": 1200,
		"alternating_drilling_offset_mm": 0.75,
		"interlacing": true,
		"interlace_compensation_mm": 0.45,
		"spoke_elongation_compensation_mm": 1.18,
		"erd_mm": 598,
		"left_flange_mm": 22.5,
		"right_flange_mm": 35.6,
		"left_flange_pcd_mm": 44,
		"right_flange_pcd_mm": 44
	}`), &request)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if request.WheelPosition != "front" || request.TopologyID != "21h-g3-2to1" || request.SpokeCount != 24 || request.Crossing != 3 {
		t.Fatalf("base fields were not decoded: %+v", request)
	}
	if request.G3RimHoleSpacingAToBDegrees != 2.5 || request.G3RimHoleSpacingBToADegrees != 8.5 || request.G3RimHoleSpacingAToNextGroupADegrees != 40.428571 {
		t.Fatalf("G3 spacing fields were not decoded: %+v", request)
	}
	if request.SpokeHeadType != "straight_pull" || request.SpokeProfile != "bladed_0_9x2_2" || !request.Interlacing {
		t.Fatalf("physical enum fields were not decoded: %+v", request)
	}
	if request.SpokeHoleDiameterMM == nil || *request.SpokeHoleDiameterMM != 2.5 {
		t.Fatalf("spoke hole diameter = %v, want 2.5", request.SpokeHoleDiameterMM)
	}
	if request.TargetTensionN == nil || *request.TargetTensionN != 1200 {
		t.Fatalf("target tension = %v, want 1200", request.TargetTensionN)
	}
	if request.SpokeElongationCompensationMM == nil || *request.SpokeElongationCompensationMM != 1.18 {
		t.Fatalf("spoke elongation compensation = %v, want 1.18", request.SpokeElongationCompensationMM)
	}
}
