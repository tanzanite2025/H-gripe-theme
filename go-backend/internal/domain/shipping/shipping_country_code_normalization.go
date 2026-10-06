package shipping

import (
	"encoding/json"
	"sort"
	"strings"
)

// parseShippingCountryCodes reads the country-list formats accepted by the
// admin API. Invalid empty entries are discarded and remaining codes are
// upper-cased and de-duplicated in lexical order.
func parseShippingCountryCodes(value string) []string {
	raw := strings.TrimSpace(value)
	if raw == "" || raw == "null" {
		return nil
	}

	var parsed []string
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		var quoted string
		if stringErr := json.Unmarshal([]byte(raw), &quoted); stringErr == nil && strings.TrimSpace(quoted) != raw {
			return parseShippingCountryCodes(quoted)
		}
		parsed = strings.FieldsFunc(raw, func(character rune) bool {
			return character == ',' || character == '，' || character == ';' || character == '|' ||
				character == '\n' || character == '\r' || character == '\t' || character == ' '
		})
	}

	seen := make(map[string]struct{}, len(parsed))
	codes := make([]string, 0, len(parsed))
	for _, item := range parsed {
		code := strings.ToUpper(strings.TrimSpace(item))
		if code == "" {
			continue
		}
		if _, exists := seen[code]; exists {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}

func marshalShippingCountryCodes(codes []string) string {
	encoded, err := json.Marshal(codes)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}
