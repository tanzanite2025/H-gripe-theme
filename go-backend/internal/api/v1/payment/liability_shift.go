package payment

import (
	"bytes"
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	pgateway "commerce-platform/internal/pkg/payment"
)

type liabilityShiftJSONPath struct {
	keys    []string
	outcome bool
}

func stripeLiabilityShiftedFromPaymentIntentSucceeded(eventObjectRaw []byte, eventRaw []byte) *bool {
	paths := stripePaymentIntentLiabilityShiftPaths()
	if shifted := liabilityShiftedFromJSON(eventObjectRaw, paths...); shifted != nil {
		return shifted
	}
	return liabilityShiftedFromJSON(eventRaw, prefixedLiabilityShiftPaths([]string{"data", "object"}, paths)...)
}

func paypalLiabilityShiftedFromWebhookPayload(resourceRaw []byte, eventRaw []byte) *bool {
	paths := []liabilityShiftJSONPath{
		{keys: []string{"purchase_units", "0", "payments", "captures", "0", "liability_shifted"}},
		{keys: []string{"purchase_units", "0", "payments", "captures", "0", "liability_shift"}, outcome: true},
		{keys: []string{"purchase_units", "0", "payments", "captures", "0", "authentication_result", "liability_shifted"}},
		{keys: []string{"purchase_units", "0", "payments", "captures", "0", "authentication_result", "liability_shift"}, outcome: true},
		{keys: []string{"payment_source", "card", "authentication_result", "liability_shifted"}},
		{keys: []string{"payment_source", "card", "authentication_result", "liability_shift"}, outcome: true},
		{keys: []string{"metadata", "liability_shifted"}},
		{keys: []string{"liability_shifted"}},
		{keys: []string{"liability_shift"}, outcome: true},
	}
	if shifted := liabilityShiftedFromJSON(resourceRaw, paths...); shifted != nil {
		return shifted
	}
	return liabilityShiftedFromJSON(eventRaw, prefixedLiabilityShiftPaths([]string{"resource"}, paths)...)
}

func paymentResponseLiabilityShifted(response *pgateway.PaymentResponse, rawResponse []byte) *bool {
	if response != nil {
		if response.LiabilityShifted != nil {
			return response.LiabilityShifted
		}
		if shifted := liabilityShiftedFromStringMap(response.Metadata); shifted != nil {
			return shifted
		}
	}
	return liabilityShiftedFromJSON(rawResponse)
}

func stripePaymentIntentLiabilityShiftPaths() []liabilityShiftJSONPath {
	return []liabilityShiftJSONPath{
		{keys: []string{"latest_charge", "payment_method_details", "card", "three_d_secure", "liability_shifted"}},
		{keys: []string{"latest_charge", "payment_method_details", "card", "three_d_secure", "liability_shift"}, outcome: true},
		{keys: []string{"charges", "data", "0", "payment_method_details", "card", "three_d_secure", "liability_shifted"}},
		{keys: []string{"charges", "data", "0", "payment_method_details", "card", "three_d_secure", "liability_shift"}, outcome: true},
		{keys: []string{"payment_method_details", "card", "three_d_secure", "liability_shifted"}},
		{keys: []string{"payment_method_details", "card", "three_d_secure", "liability_shift"}, outcome: true},
		{keys: []string{"metadata", "liability_shifted"}},
		{keys: []string{"metadata", "three_ds_liability_shifted"}},
		{keys: []string{"metadata", "three_d_secure_liability_shifted"}},
		{keys: []string{"metadata", "liability_shift"}, outcome: true},
		{keys: []string{"liability_shifted"}},
		{keys: []string{"liability_shift"}, outcome: true},
	}
}

func prefixedLiabilityShiftPaths(prefix []string, paths []liabilityShiftJSONPath) []liabilityShiftJSONPath {
	result := make([]liabilityShiftJSONPath, 0, len(paths))
	for _, path := range paths {
		keys := make([]string, 0, len(prefix)+len(path.keys))
		keys = append(keys, prefix...)
		keys = append(keys, path.keys...)
		result = append(result, liabilityShiftJSONPath{keys: keys, outcome: path.outcome})
	}
	return result
}

func liabilityShiftedFromStringMap(values map[string]string) *bool {
	if len(values) == 0 {
		return nil
	}
	asJSON := make(map[string]interface{}, len(values))
	for key, value := range values {
		asJSON[key] = value
	}
	return findLiabilityShiftedInJSON(asJSON)
}

func liabilityShiftedFromJSON(raw []byte, paths ...liabilityShiftJSONPath) *bool {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil
	}

	var payload interface{}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return nil
	}

	for _, path := range paths {
		value, ok := findLiabilityShiftJSONPath(payload, path.keys)
		if !ok {
			continue
		}
		var shifted *bool
		if path.outcome {
			shifted = parseLiabilityShiftOutcomeValue(value)
		} else {
			shifted = parseLiabilityShiftedValue(value)
		}
		if shifted != nil {
			return shifted
		}
	}
	return findLiabilityShiftedInJSON(payload)
}

func findLiabilityShiftJSONPath(value interface{}, keys []string) (interface{}, bool) {
	current := value
	for _, key := range keys {
		switch typed := current.(type) {
		case map[string]interface{}:
			nested, ok := lookupJSONMapValue(typed, key)
			if !ok {
				return nil, false
			}
			current = nested
		case []interface{}:
			index, err := strconv.Atoi(strings.TrimSpace(key))
			if err != nil || index < 0 || index >= len(typed) {
				return nil, false
			}
			current = typed[index]
		default:
			return nil, false
		}
	}
	return current, true
}

func lookupJSONMapValue(values map[string]interface{}, wanted string) (interface{}, bool) {
	wanted = compactJSONKey(wanted)
	for key, value := range values {
		if compactJSONKey(key) == wanted {
			return value, true
		}
	}
	return nil, false
}

func findLiabilityShiftedInJSON(value interface{}) *bool {
	switch typed := value.(type) {
	case map[string]interface{}:
		for _, key := range []string{"liability_shifted", "three_ds_liability_shifted", "three_d_secure_liability_shifted"} {
			if nested, ok := lookupJSONMapValue(typed, key); ok {
				if shifted := parseLiabilityShiftedValue(nested); shifted != nil {
					return shifted
				}
			}
		}
		for _, key := range []string{"liability_shift", "three_ds_liability_shift", "three_d_secure_liability_shift"} {
			if nested, ok := lookupJSONMapValue(typed, key); ok {
				if shifted := parseLiabilityShiftOutcomeValue(nested); shifted != nil {
					return shifted
				}
			}
		}

		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if shifted := findLiabilityShiftedInJSON(typed[key]); shifted != nil {
				return shifted
			}
		}
	case []interface{}:
		for _, nested := range typed {
			if shifted := findLiabilityShiftedInJSON(nested); shifted != nil {
				return shifted
			}
		}
	}
	return nil
}

func parseLiabilityShiftedValue(value interface{}) *bool {
	switch typed := value.(type) {
	case bool:
		return boolPointer(typed)
	case string:
		return parseLiabilityShiftedString(typed)
	case json.Number:
		return parseLiabilityShiftedString(typed.String())
	case float64:
		return parseLiabilityShiftedString(strconv.FormatFloat(typed, 'f', -1, 64))
	case int:
		return parseLiabilityShiftedString(strconv.Itoa(typed))
	default:
		return nil
	}
}

func parseLiabilityShiftOutcomeValue(value interface{}) *bool {
	if shifted := parseLiabilityShiftedValue(value); shifted != nil {
		return shifted
	}
	stringValue, ok := value.(string)
	if !ok {
		return nil
	}
	switch normalizeLiabilityShiftString(stringValue) {
	case "issuer", "issuershifted", "shifted", "liabilityshifted":
		return boolPointer(true)
	case "merchant", "merchantliability", "notshifted", "noliabilityshift", "nonliabilityshift", "none":
		return boolPointer(false)
	default:
		return nil
	}
}

func parseLiabilityShiftedString(value string) *bool {
	switch normalizeLiabilityShiftString(value) {
	case "true", "t", "1", "yes", "y":
		return boolPointer(true)
	case "false", "f", "0", "no", "n":
		return boolPointer(false)
	default:
		return nil
	}
}

func normalizeLiabilityShiftString(value string) string {
	return compactJSONKey(value)
}

func compactJSONKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer("_", "", "-", "", " ", "", ".", "")
	return replacer.Replace(value)
}

func boolPointer(value bool) *bool {
	return &value
}
