package fitmentcatalog

import "testing"

func TestForkFitmentEntryValidateYearModes(t *testing.T) {
	singleYear := 2024
	rangeStart := 2022
	rangeEnd := 2024

	tests := []struct {
		name    string
		entry   ForkFitmentEntry
		wantErr bool
	}{
		{
			name: "single year",
			entry: ForkFitmentEntry{
				BrandName: " RockShox ",
				ModelName: " SID SL ",
				YearMode:  YearModeSingle,
				YearFrom:  &singleYear,
			},
		},
		{
			name: "year range",
			entry: ForkFitmentEntry{
				BrandName: "Fox",
				ModelName: "32 Step-Cast",
				YearMode:  YearModeRange,
				YearFrom:  &rangeStart,
				YearTo:    &rangeEnd,
			},
		},
		{
			name: "all years",
			entry: ForkFitmentEntry{
				BrandName: "Fox",
				ModelName: "32 Step-Cast",
				YearMode:  YearModeAll,
			},
		},
		{
			name: "range without end",
			entry: ForkFitmentEntry{
				BrandName: "Fox",
				ModelName: "32 Step-Cast",
				YearMode:  YearModeRange,
				YearFrom:  &rangeStart,
			},
			wantErr: true,
		},
		{
			name: "year range reversed",
			entry: ForkFitmentEntry{
				BrandName: "Fox",
				ModelName: "32 Step-Cast",
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
