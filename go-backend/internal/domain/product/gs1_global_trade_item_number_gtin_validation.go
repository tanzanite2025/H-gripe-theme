package product

// IsValidGS1GlobalTradeItemNumber reports whether value is a canonical GTIN-8,
// GTIN-12, GTIN-13, or GTIN-14 digit string with a valid GS1 check digit.
func IsValidGS1GlobalTradeItemNumber(value string) bool {
	if len(value) != 8 && len(value) != 12 && len(value) != 13 && len(value) != 14 {
		return false
	}

	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return false
		}
	}

	sum := 0
	weight := 3
	for index := len(value) - 2; index >= 0; index-- {
		sum += int(value[index]-'0') * weight
		if weight == 3 {
			weight = 1
		} else {
			weight = 3
		}
	}

	expectedCheckDigit := (10 - sum%10) % 10
	return int(value[len(value)-1]-'0') == expectedCheckDigit
}
