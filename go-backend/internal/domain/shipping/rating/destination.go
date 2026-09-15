package rating

import (
	"encoding/json"
	"strings"
)

var euCountryCodes = countrySet(
	"AT", "BE", "BG", "HR", "CY", "CZ", "DE", "DK", "EE", "ES", "FI", "FR", "GR",
	"HU", "IE", "IT", "LT", "LU", "LV", "MT", "NL", "PL", "PT", "RO", "SE", "SI", "SK",
)

var eeaCountryCodes = countrySet(
	"AT", "BE", "BG", "CY", "CZ", "DE", "DK", "EE", "ES", "FI", "FR", "GR", "HR", "HU",
	"IS", "IE", "IT", "LI", "LT", "LU", "LV", "MT", "NL", "NO", "PL", "PT", "RO", "SE", "SI", "SK",
)

var regionCountryCodes = map[string]map[string]struct{}{
	"EU":                     euCountryCodes,
	"EU27":                   euCountryCodes,
	"EUROPEAN_UNION":         euCountryCodes,
	"EUROPEAN UNION":         euCountryCodes,
	"EEA":                    eeaCountryCodes,
	"EUROPEAN_ECONOMIC_AREA": eeaCountryCodes,
	"EUROPEAN ECONOMIC AREA": eeaCountryCodes,
}

// NormalizeCountry accepts the ISO 3166-1 alpha-2 representation used by the
// shipping domain. Country support is evaluated separately against rate data.
func NormalizeCountry(value string) (string, bool) {
	country := strings.ToUpper(strings.TrimSpace(value))
	if len(country) != 2 {
		return "", false
	}
	for _, character := range country {
		if character < 'A' || character > 'Z' {
			return "", false
		}
	}
	return country, true
}

// MatchesRegion evaluates a country code against a direct country, a supported
// region macro, a JSON list, or a delimited list. An empty region is global.
func MatchesRegion(regionValue string, countryValue string) bool {
	country, ok := NormalizeCountry(countryValue)
	if !ok {
		return false
	}

	regions := NormalizeRegions(regionValue)
	if len(regions) == 0 {
		return true
	}
	for _, candidate := range regions {
		switch candidate {
		case "*", "ALL", "GLOBAL", "WORLDWIDE":
			return true
		default:
			if candidate == country || regionMatchesCountry(candidate, country) {
				return true
			}
		}
	}
	return false
}

func NormalizeRegions(value string) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	var parsed []string
	if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
		return normalizeRegionList(parsed)
	}
	parts := strings.FieldsFunc(trimmed, func(character rune) bool {
		return character == ',' || character == ';' || character == '|' ||
			character == '\n' || character == '\r' || character == '\t'
	})
	return normalizeRegionList(parts)
}

func normalizeRegionList(regions []string) []string {
	normalized := make([]string, 0, len(regions))
	for _, region := range regions {
		if candidate := strings.ToUpper(strings.TrimSpace(region)); candidate != "" {
			normalized = append(normalized, candidate)
		}
	}
	return normalized
}

func regionMatchesCountry(region string, country string) bool {
	countries, ok := regionCountryCodes[region]
	if !ok {
		return false
	}
	_, ok = countries[country]
	return ok
}

func countrySet(codes ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		result[code] = struct{}{}
	}
	return result
}
