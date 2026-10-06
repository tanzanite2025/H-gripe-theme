package shipping

import "testing"

func TestNormalizeShippingServiceCollectionCountryCodes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "json array is normalized and deduplicated", input: `["de","FR","DE"]`, want: `["DE","FR"]`},
		{name: "csv input is accepted", input: "US, ca GB", want: `["US","CA","GB"]`},
		{name: "empty input is an unrestricted array", input: "", want: `[]`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := NormalizeShippingServiceCollectionCountryCodes(test.input); got != test.want {
				t.Fatalf("NormalizeShippingServiceCollectionCountryCodes(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}
