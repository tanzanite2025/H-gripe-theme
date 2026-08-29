package fitmentcatalog

import "testing"

func TestHubSpecificationValidateNormalizesAndAcceptsSupportedValues(t *testing.T) {
	specification := HubSpecification{
		SpecCode:      "  hub-r-12x142 ",
		DisplayName:   " Rear Thru Axle ",
		Position:      " REAR ",
		AxleType:      " THRU_AXLE ",
		AxleSpacingMM: 142,
	}

	if err := specification.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if specification.SpecCode != "HUB-R-12X142" {
		t.Fatalf("SpecCode = %q, want normalized uppercase code", specification.SpecCode)
	}
	if specification.Position != HubPositionRear {
		t.Fatalf("Position = %q, want %q", specification.Position, HubPositionRear)
	}
	if specification.AxleType != HubAxleTypeThruAxle {
		t.Fatalf("AxleType = %q, want %q", specification.AxleType, HubAxleTypeThruAxle)
	}
}

func TestHubSpecificationValidateRejectsInvalidBaseFields(t *testing.T) {
	tests := []struct {
		name          string
		specification HubSpecification
	}{
		{
			name: "missing code",
			specification: HubSpecification{
				DisplayName:   "Front hub",
				Position:      HubPositionFront,
				AxleType:      HubAxleTypeQuickRelease,
				AxleSpacingMM: 100,
			},
		},
		{
			name: "unsupported position",
			specification: HubSpecification{
				SpecCode:      "HUB-1",
				DisplayName:   "Hub",
				Position:      "middle",
				AxleType:      HubAxleTypeQuickRelease,
				AxleSpacingMM: 100,
			},
		},
		{
			name: "non-positive spacing",
			specification: HubSpecification{
				SpecCode:    "HUB-1",
				DisplayName: "Hub",
				Position:    HubPositionFront,
				AxleType:    HubAxleTypeQuickRelease,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.specification.Validate(); err == nil {
				t.Fatal("Validate() error = nil, want validation error")
			}
		})
	}
}

func TestHubSpecificationValidateAcceptsCompleteSpokeGeometryAtPcdLowerBound(t *testing.T) {
	specification := HubSpecification{
		SpecCode:      "HUB-F-100",
		DisplayName:   "Front hub",
		Position:      HubPositionFront,
		AxleType:      HubAxleTypeQuickRelease,
		AxleSpacingMM: 100,
		WRMM:          floatPointer(35),
		WLMM:          floatPointer(22),
		PCDRMM:        floatPointer(10),
		PCDLMM:        floatPointer(10),
	}

	if err := specification.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestHubSpecificationValidateRejectsPartialSpokeGeometry(t *testing.T) {
	specification := HubSpecification{
		SpecCode:      "HUB-R-142",
		DisplayName:   "Rear hub",
		Position:      HubPositionRear,
		AxleType:      HubAxleTypeThruAxle,
		AxleSpacingMM: 142,
		WRMM:          floatPointer(20),
	}

	if err := specification.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want partial spoke geometry to be rejected")
	}
}

func TestHubSpecificationValidateRejectsCodesUnsafeForSpokeCatalog(t *testing.T) {
	for _, code := range []string{"A", "HUB/142", "HUB 142", "HUB.142"} {
		t.Run(code, func(t *testing.T) {
			specification := HubSpecification{
				SpecCode:      code,
				DisplayName:   "Rear hub",
				Position:      HubPositionRear,
				AxleType:      HubAxleTypeThruAxle,
				AxleSpacingMM: 142,
			}
			if err := specification.Validate(); err == nil {
				t.Fatalf("Validate() error = nil, want unsafe spoke catalog code %q to be rejected", code)
			}
		})
	}
}

func floatPointer(value float64) *float64 {
	return &value
}
