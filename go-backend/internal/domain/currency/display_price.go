package currency

import (
	"encoding/json"
	"math/big"
	"strings"

	"gorm.io/datatypes"
)

type DisplayPriceSnapshot struct {
	AmountDecimal  string  `json:"amount_decimal"`
	Currency       string  `json:"currency"`
	QuoteCurrency  string  `json:"quote_currency,omitempty"`
	Rate           float64 `json:"rate,omitempty"`
	Source         string  `json:"source,omitempty"`
	Converted      bool    `json:"converted,omitempty"`
	FallbackReason string  `json:"fallback_reason,omitempty"`
}

type DisplayPriceSnapshotMap map[string][]DisplayPriceSnapshot

func NormalizeDisplayPriceSnapshots(values []DisplayPriceSnapshot, baseCurrency string) []DisplayPriceSnapshot {
	baseCurrency = NormalizeCode(baseCurrency)
	seen := map[string]struct{}{}
	result := make([]DisplayPriceSnapshot, 0, len(values))

	for _, value := range values {
		code := NormalizeCode(value.QuoteCurrency)
		if code == "" {
			code = NormalizeCode(value.Currency)
		}
		amount, ok := normalizeDisplayAmountDecimal(value.AmountDecimal, code)
		if !ok || strings.TrimSpace(value.FallbackReason) != "" {
			continue
		}
		if code == "" || code == baseCurrency || !IsCatalogCode(code) {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		result = append(result, DisplayPriceSnapshot{
			AmountDecimal: amount,
			Currency:      code,
			QuoteCurrency: code,
			Rate:          value.Rate,
			Source:        strings.TrimSpace(value.Source),
			Converted:     value.Converted || value.Rate > 0,
		})
	}

	return result
}

// normalizeDisplayAmountDecimal keeps read-model amounts decimal-only. It
// validates the value against the quote currency scale without converting it
// through binary floating point.
func normalizeDisplayAmountDecimal(raw, code string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.HasPrefix(raw, "+") || strings.HasPrefix(raw, "-") {
		return "", false
	}
	units, ok := MinorUnits(code)
	if !ok || strings.Count(raw, ".") > 1 {
		return "", false
	}
	major, fraction := raw, ""
	if dot := strings.IndexByte(raw, '.'); dot >= 0 {
		major, fraction = raw[:dot], raw[dot+1:]
	}
	if major == "" || !allDecimalDigits(major) || (fraction != "" && !allDecimalDigits(fraction)) {
		return "", false
	}
	fraction = strings.TrimRight(fraction, "0")
	if len(fraction) > units {
		return "", false
	}
	fraction += strings.Repeat("0", units-len(fraction))
	major = strings.TrimLeft(major, "0")
	if major == "" {
		major = "0"
	}
	digits := strings.TrimLeft(major+fraction, "0")
	if digits == "" {
		return "", false
	}
	parsed, ok := new(big.Int).SetString(digits, 10)
	if !ok || parsed.Sign() <= 0 {
		return "", false
	}
	if units == 0 {
		return major, true
	}
	return major + "." + fraction, true
}

func allDecimalDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func NormalizeDisplayPriceSnapshotMap(values map[string][]DisplayPriceSnapshot, baseCurrency string, allowedKeys ...string) DisplayPriceSnapshotMap {
	allowed := map[string]struct{}{}
	for _, key := range allowedKeys {
		key = strings.TrimSpace(key)
		if key != "" {
			allowed[key] = struct{}{}
		}
	}

	result := DisplayPriceSnapshotMap{}
	for key, snapshots := range values {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if len(allowed) > 0 {
			if _, ok := allowed[key]; !ok {
				continue
			}
		}
		normalized := NormalizeDisplayPriceSnapshots(snapshots, baseCurrency)
		if len(normalized) > 0 {
			result[key] = normalized
		}
	}
	return result
}

func DisplayPriceSnapshotsJSON(values []DisplayPriceSnapshot, baseCurrency string) datatypes.JSON {
	normalized := NormalizeDisplayPriceSnapshots(values, baseCurrency)
	if len(normalized) == 0 {
		return datatypes.JSON([]byte("[]"))
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return datatypes.JSON([]byte("[]"))
	}
	return datatypes.JSON(encoded)
}

func DisplayPriceSnapshotMapJSON(values map[string][]DisplayPriceSnapshot, baseCurrency string, allowedKeys ...string) datatypes.JSON {
	normalized := NormalizeDisplayPriceSnapshotMap(values, baseCurrency, allowedKeys...)
	if len(normalized) == 0 {
		return datatypes.JSON([]byte("{}"))
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(encoded)
}

func ParseDisplayPriceSnapshots(raw datatypes.JSON) []DisplayPriceSnapshot {
	if len(raw) == 0 {
		return nil
	}
	var values []DisplayPriceSnapshot
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil
	}
	return NormalizeDisplayPriceSnapshots(values, "")
}

func ParseDisplayPriceSnapshotMap(raw datatypes.JSON, allowedKeys ...string) DisplayPriceSnapshotMap {
	if len(raw) == 0 {
		return nil
	}
	var values map[string][]DisplayPriceSnapshot
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil
	}
	return NormalizeDisplayPriceSnapshotMap(values, "", allowedKeys...)
}
