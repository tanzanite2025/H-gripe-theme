package currency

import (
	"encoding/json"
	"strings"

	"gorm.io/datatypes"
)

type DisplayPriceSnapshot struct {
	Amount         float64 `json:"amount"`
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
		if value.Amount <= 0 || strings.TrimSpace(value.FallbackReason) != "" {
			continue
		}
		code := NormalizeCode(value.QuoteCurrency)
		if code == "" {
			code = NormalizeCode(value.Currency)
		}
		if code == "" || code == baseCurrency || !IsCatalogCode(code) {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		result = append(result, DisplayPriceSnapshot{
			Amount:        value.Amount,
			Currency:      code,
			QuoteCurrency: code,
			Rate:          value.Rate,
			Source:        strings.TrimSpace(value.Source),
			Converted:     value.Converted || value.Rate > 0,
		})
	}

	return result
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
