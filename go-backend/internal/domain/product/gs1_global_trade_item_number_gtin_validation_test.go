package product

import "testing"

func TestIsValidGS1GlobalTradeItemNumberAcceptsCanonicalGTINLengthsAndCheckDigits(t *testing.T) {
	for _, value := range []string{
		"96385074",
		"036000291452",
		"4006381333931",
		"00012345678905",
	} {
		t.Run(value, func(t *testing.T) {
			if !IsValidGS1GlobalTradeItemNumber(value) {
				t.Fatalf("expected valid GS1 GTIN %q", value)
			}
		})
	}
}

func TestIsValidGS1GlobalTradeItemNumberRejectsInvalidLengthCharactersAndCheckDigits(t *testing.T) {
	for _, value := range []string{
		"",
		"9638507",
		"963850741",
		"4006381333932",
		"40063813339A1",
		"400-6381333931",
	} {
		t.Run(value, func(t *testing.T) {
			if IsValidGS1GlobalTradeItemNumber(value) {
				t.Fatalf("expected invalid GS1 GTIN %q", value)
			}
		})
	}
}
