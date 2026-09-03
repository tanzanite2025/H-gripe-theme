package spoke

func DefaultOptions() CatalogOptions {
	return CatalogOptions{
		SpokeCounts: []IntOption{
			{Value: 16, Label: "16"},
			{Value: 18, Label: "18"},
			{Value: 20, Label: "20"},
			{Value: 24, Label: "24"},
			{Value: 28, Label: "28"},
			{Value: 32, Label: "32"},
			{Value: 36, Label: "36"},
		},
		Crossings: []IntOption{
			{Value: 0, Label: "0-cross (Radial)"},
			{Value: 1, Label: "1-cross"},
			{Value: 2, Label: "2-cross"},
			{Value: 3, Label: "3-cross"},
			{Value: 4, Label: "4-cross"},
		},
		NippleTypes: []StringOption{
			{Value: "standard", Label: "Standard external"},
			{Value: "hidden", Label: "Hidden / aero"},
		},
		WheelPositions: []StringOption{
			{Value: "auto", Label: "Auto"},
			{Value: "front", Label: "Front"},
			{Value: "rear", Label: "Rear"},
		},
	}
}

func DefaultExport() ExportResponse {
	return ExportResponse{
		Options: DefaultOptions(),
		Rims:    []RimBrand{},
		Hubs:    []HubBrand{},
		Presets: []WheelBuildPreset{},
	}
}

func floatPtr(value float64) *float64 {
	return &value
}

func hubGeometry(leftFlange, rightFlange, leftFlangePCD, rightFlangePCD float64) *HubGeometry {
	return &HubGeometry{
		LeftFlange:     floatPtr(leftFlange),
		RightFlange:    floatPtr(rightFlange),
		LeftFlangePCD:  floatPtr(leftFlangePCD),
		RightFlangePCD: floatPtr(rightFlangePCD),
	}
}
