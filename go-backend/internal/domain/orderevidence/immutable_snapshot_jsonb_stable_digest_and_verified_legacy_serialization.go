package orderevidence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"gorm.io/datatypes"
)

// CanonicalImmutableSnapshotJSONForStorageIndependentIntegrity preserves exact
// decimal values, including IDs beyond float64 precision. JSONB may change key
// order, whitespace and exponent notation without changing the document.
func CanonicalImmutableSnapshotJSONForStorageIndependentIntegrity(data []byte) ([]byte, error) {
	if !json.Valid(data) {
		return nil, errors.New("snapshot data must be valid JSON")
	}
	canonical, err := canonicalJSON(string(data))
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.UseNumber()
	var document any
	if err := decoder.Decode(&document); err != nil {
		return nil, err
	}
	if number, ok := document.(json.Number); ok {
		normalized, err := immutableSnapshotExactDecimalNumber(number.String())
		if err != nil {
			return nil, err
		}
		document = json.Number(normalized)
	}
	if err := normalizeImmutableSnapshotExactDecimalNumbers(document); err != nil {
		return nil, err
	}
	return json.Marshal(document)
}

func normalizeImmutableSnapshotExactDecimalNumbers(document any) error {
	switch value := document.(type) {
	case map[string]any:
		for key, child := range value {
			if number, ok := child.(json.Number); ok {
				normalized, err := immutableSnapshotExactDecimalNumber(number.String())
				if err != nil {
					return err
				}
				value[key] = json.Number(normalized)
			} else if err := normalizeImmutableSnapshotExactDecimalNumbers(child); err != nil {
				return err
			}
		}
	case []any:
		for index, child := range value {
			if number, ok := child.(json.Number); ok {
				normalized, err := immutableSnapshotExactDecimalNumber(number.String())
				if err != nil {
					return err
				}
				value[index] = json.Number(normalized)
			} else if err := normalizeImmutableSnapshotExactDecimalNumbers(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func immutableSnapshotExactDecimalNumber(number string) (string, error) {
	sign := ""
	if strings.HasPrefix(number, "-") {
		sign, number = "-", number[1:]
	}
	exponent := 0
	if index := strings.IndexAny(number, "eE"); index >= 0 {
		parsed, err := strconv.Atoi(number[index+1:])
		if err != nil || parsed < -200000 || parsed > 200000 {
			return "", errors.New("snapshot number exceeds JSONB decimal range")
		}
		exponent, number = parsed, number[:index]
	}
	if index := strings.IndexByte(number, '.'); index >= 0 {
		exponent -= len(number) - index - 1
		number = number[:index] + number[index+1:]
	}
	number = strings.TrimLeft(number, "0")
	if number == "" {
		return "0", nil
	}
	for strings.HasSuffix(number, "0") {
		number = strings.TrimSuffix(number, "0")
		exponent++
	}
	if exponent >= 0 {
		return sign + number + strings.Repeat("0", exponent), nil
	}
	point := len(number) + exponent
	if point > 0 {
		return sign + number[:point] + "." + number[point:], nil
	}
	return sign + "0." + strings.Repeat("0", -point) + number, nil
}

func immutableSnapshotDigestMatchesDocumentOrVerifiedLegacySerialization(data []byte, digest string, legacyPayload any) bool {
	if !json.Valid(data) {
		return false
	}
	matches := func(candidate []byte) bool {
		hash := sha256.Sum256(candidate)
		return strings.EqualFold(strings.TrimSpace(digest), hex.EncodeToString(hash[:]))
	}
	if matches(data) {
		return true
	}
	if canonical, err := CanonicalImmutableSnapshotJSONForStorageIndependentIntegrity(data); err == nil && matches(canonical) {
		return true
	}
	// Compatibility with the first stable submission writer, which retained
	// the lexical representation of numbers. Never replace the persisted hash.
	if canonical, err := canonicalJSON(string(data)); err == nil && matches(canonical) {
		return true
	}
	if legacyPayload == nil {
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(legacyPayload); err != nil {
		return false
	}
	// Historical initial payload raw specs were sorted by Go. Export raw JSON
	// came from JSONB, whose keys are ordered by UTF-8 byte length then bytes.
	for _, postgresRawOrder := range []bool{false, true} {
		if err := restoreLegacyImmutableSnapshotRawJSONFieldOrder(reflect.ValueOf(legacyPayload), postgresRawOrder); err != nil {
			return false
		}
		reconstructed, err := json.Marshal(legacyPayload)
		if err != nil || !matches(reconstructed) {
			continue
		}
		originalContent, originalErr := CanonicalImmutableSnapshotJSONForStorageIndependentIntegrity(data)
		reconstructedContent, reconstructedErr := CanonicalImmutableSnapshotJSONForStorageIndependentIntegrity(reconstructed)
		// This equality prevents ignored, omitted or defaulted fields from
		// allowing a valid digest to authenticate only part of the stored body.
		if originalErr == nil && reconstructedErr == nil && bytes.Equal(originalContent, reconstructedContent) {
			return true
		}
	}
	return false
}

func restoreLegacyImmutableSnapshotRawJSONFieldOrder(value reflect.Value, postgresOrder bool) error {
	if !value.IsValid() {
		return nil
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		return restoreLegacyImmutableSnapshotRawJSONFieldOrder(value.Elem(), postgresOrder)
	}
	if value.Type() == reflect.TypeOf(json.RawMessage{}) || value.Type() == reflect.TypeOf(datatypes.JSON{}) {
		if value.Len() == 0 || !value.CanSet() {
			return nil
		}
		restored, err := serializeLegacyImmutableSnapshotRawJSONWithKnownKeyOrder(value.Bytes(), postgresOrder)
		if err != nil {
			return err
		}
		value.SetBytes(restored)
		return nil
	}
	switch value.Kind() {
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			if value.Type().Field(index).PkgPath != "" {
				continue
			}
			if err := restoreLegacyImmutableSnapshotRawJSONFieldOrder(value.Field(index), postgresOrder); err != nil {
				return err
			}
		}
	case reflect.Slice:
		for index := 0; index < value.Len(); index++ {
			if err := restoreLegacyImmutableSnapshotRawJSONFieldOrder(value.Index(index), postgresOrder); err != nil {
				return err
			}
		}
	}
	return nil
}

func serializeLegacyImmutableSnapshotRawJSONWithKnownKeyOrder(data []byte, postgresOrder bool) ([]byte, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, errors.New("empty legacy raw snapshot JSON")
	}
	var output bytes.Buffer
	switch trimmed[0] {
	case '{':
		var object map[string]json.RawMessage
		if err := json.Unmarshal(trimmed, &object); err != nil {
			return nil, err
		}
		keys := make([]string, 0, len(object))
		for key := range object {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			if postgresOrder && len(keys[i]) != len(keys[j]) {
				return len(keys[i]) < len(keys[j])
			}
			return keys[i] < keys[j]
		})
		output.WriteByte('{')
		for index, key := range keys {
			if index > 0 {
				output.WriteByte(',')
			}
			encodedKey, _ := json.Marshal(key)
			output.Write(encodedKey)
			output.WriteByte(':')
			child, err := serializeLegacyImmutableSnapshotRawJSONWithKnownKeyOrder(object[key], postgresOrder)
			if err != nil {
				return nil, err
			}
			output.Write(child)
		}
		output.WriteByte('}')
	case '[':
		var array []json.RawMessage
		if err := json.Unmarshal(trimmed, &array); err != nil {
			return nil, err
		}
		output.WriteByte('[')
		for index, child := range array {
			if index > 0 {
				output.WriteByte(',')
			}
			restored, err := serializeLegacyImmutableSnapshotRawJSONWithKnownKeyOrder(child, postgresOrder)
			if err != nil {
				return nil, err
			}
			output.Write(restored)
		}
		output.WriteByte(']')
	default:
		if err := json.Compact(&output, trimmed); err != nil {
			return nil, err
		}
	}
	return output.Bytes(), nil
}
