package fitmentcatalog

import "testing"

func TestFrameFitmentEntryValidateYearModes(t *testing.T) {
	singleYear := 2024
	rangeStart := 2022
	rangeEnd := 2024

	tests := []struct {
		name    string
		entry   FrameFitmentEntry
		wantErr bool
	}{
		{
			name: "single year",
			entry: FrameFitmentEntry{
				BrandName: " Example ",
				ModelName: " Road X ",
				YearMode:  YearModeSingle,
				YearFrom:  &singleYear,
			},
		},
		{
			name: "year range",
			entry: FrameFitmentEntry{
				BrandName: "Example",
				ModelName: "Road X",
				YearMode:  YearModeRange,
				YearFrom:  &rangeStart,
				YearTo:    &rangeEnd,
			},
		},
		{
			name: "all years",
			entry: FrameFitmentEntry{
				BrandName: "Example",
				ModelName: "Road X",
				YearMode:  YearModeAll,
			},
		},
		{
			name: "range without end",
			entry: FrameFitmentEntry{
				BrandName: "Example",
				ModelName: "Road X",
				YearMode:  YearModeRange,
				YearFrom:  &rangeStart,
			},
			wantErr: true,
		},
		{
			name: "year range reversed",
			entry: FrameFitmentEntry{
				BrandName: "Example",
				ModelName: "Road X",
				YearMode:  YearModeRange,
				YearFrom:  &rangeEnd,
				YearTo:    &rangeStart,
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.entry.Validate()
			if (err != nil) != test.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
