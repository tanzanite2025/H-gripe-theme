package service

import (
	"commerce-platform/internal/domain/shipping"
	"encoding/json"
	"strings"
)

// carrierServiceRemoteSurcharge applies the configured surcharge only to a
// matching postal-code rule. An empty rule list keeps the service-wide
// surcharge behavior for configurations without postal-code restrictions.
func carrierServiceRemoteSurcharge(service shipping.CarrierService, postalCode string) (float64, error) {
	if service.RemoteSurcharge <= 0 {
		return 0, nil
	}
	rulesValue := strings.TrimSpace(service.RemotePostalCodes)
	if rulesValue == "" || rulesValue == "[]" || rulesValue == "null" {
		return service.RemoteSurcharge, nil
	}

	postalCode = normalizeShippingPostalCode(postalCode)
	if postalCode == "" {
		return 0, nil
	}
	matched, ok := remotePostalCodeRulesMatch(postalCode, rulesValue)
	if !ok {
		// Invalid configuration must not turn into a surcharge for every
		// destination. Treat it as no match until an administrator fixes it.
		return 0, nil
	}
	if matched {
		return service.RemoteSurcharge, nil
	}
	return 0, nil
}

func normalizeShippingPostalCode(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	value = strings.Join(strings.Fields(value), "")
	return value
}

// remotePostalCodeRulesMatch accepts a JSON array of strings or objects. String
// entries support exact values, a trailing '*' prefix, and inclusive ranges
// such as "10000-10099". Object entries support {"prefix":"..."},
// {"from":"...","to":"..."}, and {"postal_code":"..."}.
func remotePostalCodeRulesMatch(postalCode string, raw string) (bool, bool) {
	var entries []json.RawMessage
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		parts := strings.FieldsFunc(raw, func(r rune) bool {
			return r == ',' || r == ';' || r == '|' || r == '\n' || r == '\r' || r == '\t'
		})
		if len(parts) == 0 {
			return false, false
		}
		for _, part := range parts {
			if remotePostalCodePatternMatches(postalCode, part) {
				return true, true
			}
		}
		return false, true
	}
	for _, entry := range entries {
		var pattern string
		if err := json.Unmarshal(entry, &pattern); err == nil {
			if remotePostalCodePatternMatches(postalCode, pattern) {
				return true, true
			}
			continue
		}

		var object map[string]interface{}
		if err := json.Unmarshal(entry, &object); err != nil {
			return false, false
		}
		for _, key := range []string{"postal_code", "postalCode", "code", "pattern", "prefix"} {
			if value, ok := object[key].(string); ok && remotePostalCodePatternMatches(postalCode, value) {
				return true, true
			}
		}
		from := firstRemotePostalCodeObjectValue(object, "from", "start", "min")
		to := firstRemotePostalCodeObjectValue(object, "to", "end", "max")
		if from != "" && to != "" && remotePostalCodeRangeMatches(postalCode, from, to) {
			return true, true
		}
	}
	return false, true
}

func firstRemotePostalCodeObjectValue(object map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := object[key].(string); ok {
			if value = normalizeShippingPostalCode(value); value != "" {
				return value
			}
		}
	}
	return ""
}

func remotePostalCodePatternMatches(postalCode string, pattern string) bool {
	pattern = normalizeShippingPostalCode(pattern)
	if pattern == "" {
		return false
	}
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(postalCode, strings.TrimSuffix(pattern, "*"))
	}
	parts := strings.Split(pattern, "-")
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return remotePostalCodeRangeMatches(postalCode, parts[0], parts[1])
	}
	return postalCode == pattern
}

func remotePostalCodeRangeMatches(postalCode, from, to string) bool {
	from = normalizeShippingPostalCode(from)
	to = normalizeShippingPostalCode(to)
	if from == "" || to == "" || len(postalCode) != len(from) || len(postalCode) != len(to) {
		return false
	}
	return postalCode >= from && postalCode <= to
}
