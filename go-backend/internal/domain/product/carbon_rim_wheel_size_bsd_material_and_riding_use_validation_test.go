package product

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCarbonRimWheelSizeMustAgreeWithBSDAndDoesNotInferRidingUse(t *testing.T) {
	spec := &CarbonRimSpecification{WheelSize: "700c", BeadSeatDiameterMM: 622, Material: "carbon_fiber", RidingUses: []string{"road", "gravel"}}
	require.NoError(t, ValidateCarbonRimWheelSizeMaterialAndRidingUses(spec, true))
	spec.WheelSize = "29-inch"
	require.NoError(t, ValidateCarbonRimWheelSizeMaterialAndRidingUses(spec, true))
	require.Equal(t, []string{"road", "gravel"}, spec.RidingUses)
	spec.WheelSize = "650b"
	require.ErrorContains(t, ValidateCarbonRimWheelSizeMaterialAndRidingUses(spec, true), "bead_seat_diameter_mm")
	spec.BeadSeatDiameterMM = 584
	require.NoError(t, ValidateCarbonRimWheelSizeMaterialAndRidingUses(spec, true))
	spec.Material = "aluminum"
	require.ErrorContains(t, ValidateCarbonRimWheelSizeMaterialAndRidingUses(spec, true), "material")
	spec.Material = "carbon_fiber"
	for _, uses := range [][]string{nil, {"road", "road"}, {"unverified-discipline"}} {
		spec.RidingUses = uses
		require.Error(t, ValidateCarbonRimWheelSizeMaterialAndRidingUses(spec, true))
	}
}

func TestCarbonRimOlderDocumentsDoNotAcquireInventedWheelSizeOrRidingUses(t *testing.T) {
	old := &CarbonRimSpecification{BeadSeatDiameterMM: 622}
	require.NoError(t, ValidateCarbonRimWheelSizeMaterialAndRidingUses(old, false))
	require.Error(t, ValidateCarbonRimWheelSizeMaterialAndRidingUses(old, true))
	require.Empty(t, CarbonRimWheelSizeDisplayLabel(old.WheelSize))
	require.Empty(t, CarbonRimRidingUseDisplayLabels(old.RidingUses, "en"))
}
