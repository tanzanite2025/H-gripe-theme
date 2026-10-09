package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sort"
	"strings"
	"time"

	"commerce-platform/internal/domain/taxrate"
)

const (
	vatcomplyTaxRateProviderName = "VATcomply"
	vatcomplyTaxRateAPIEndpoint  = "https://api.vatcomply.com/vat_rates"
	vatcomplyTaxRateAPIDocURL    = "https://www.vatcomply.com/api/vat-rates/"
	taxRateSourceRequestTimeout  = 20 * time.Second
	taxRateSourceMaximumBodySize = 5 * 1024 * 1024
)

// Keep the provider's current EU member-state coverage explicit so truncated
// or unexpected responses cannot be published as complete snapshots. Update
// this contract when EU membership changes.
var vatcomplyExpectedEUMemberStateCountryCodes = map[string]struct{}{
	"AT": {}, "BE": {}, "BG": {}, "CY": {}, "CZ": {}, "DE": {}, "DK": {},
	"EE": {}, "ES": {}, "FI": {}, "FR": {}, "GR": {}, "HR": {}, "HU": {},
	"IE": {}, "IT": {}, "LT": {}, "LU": {}, "LV": {}, "MT": {}, "NL": {},
	"PL": {}, "PT": {}, "RO": {}, "SE": {}, "SI": {}, "SK": {},
}

type vatcomplyTaxRateSourceAdapter struct {
	httpClient *http.Client
}

type vatcomplyCountryTaxRatesResponse struct {
	CountryCode      string                   `json:"country_code"`
	CountryName      string                   `json:"country_name"`
	StandardRate     json.Number              `json:"standard_rate"`
	ReducedRates     []json.Number            `json:"reduced_rates"`
	SuperReducedRate *json.Number             `json:"super_reduced_rate"`
	ParkingRate      *json.Number             `json:"parking_rate"`
	Currency         string                   `json:"currency"`
	MemberState      bool                     `json:"member_state"`
	RateCategories   map[string][]json.Number `json:"rate_categories"`
}

func newVATComplyTaxRateSourceAdapter() *vatcomplyTaxRateSourceAdapter {
	return &vatcomplyTaxRateSourceAdapter{
		httpClient: &http.Client{
			Timeout: taxRateSourceRequestTimeout,
			CheckRedirect: func(request *http.Request, previousRequests []*http.Request) error {
				if request.URL.Scheme != "https" || !strings.EqualFold(request.URL.Host, "api.vatcomply.com") {
					return errors.New("VATcomply redirected outside its configured HTTPS host")
				}
				return nil
			},
		},
	}
}

func (adapter *vatcomplyTaxRateSourceAdapter) FetchVATComplyTaxRateEntries(
	ctx context.Context,
	capturedAt time.Time,
) ([]taxrate.TaxRateSourceSnapshotEntry, int, error) {
	if adapter == nil || adapter.httpClient == nil {
		return nil, 0, errors.New("VATcomply tax rate adapter is not configured")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, vatcomplyTaxRateAPIEndpoint, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create VATcomply request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "TanzaniteCommerce-TaxRateSnapshot/1.0")
	response, err := adapter.httpClient.Do(request)
	if err != nil {
		return nil, 0, fmt.Errorf("request VATcomply tax rates: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, 0, fmt.Errorf("VATcomply tax rate API returned HTTP %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, taxRateSourceMaximumBodySize+1))
	if err != nil {
		return nil, 0, fmt.Errorf("read VATcomply response: %w", err)
	}
	if len(body) > taxRateSourceMaximumBodySize {
		return nil, 0, fmt.Errorf("VATcomply response exceeds %d bytes", taxRateSourceMaximumBodySize)
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var providerCountries []vatcomplyCountryTaxRatesResponse
	if err := decoder.Decode(&providerCountries); err != nil {
		return nil, 0, fmt.Errorf("decode VATcomply response: %w", err)
	}
	var trailingData interface{}
	if err := decoder.Decode(&trailingData); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, 0, errors.New("VATcomply response contains trailing JSON data")
		}
		return nil, 0, fmt.Errorf("decode trailing VATcomply response data: %w", err)
	}

	entries := make([]taxrate.TaxRateSourceSnapshotEntry, 0, len(providerCountries)*4)
	countryCodes := make(map[string]struct{}, len(providerCountries))
	rateEntryKeys := make(map[string]struct{}, len(providerCountries)*4)
	for _, country := range providerCountries {
		if !country.MemberState {
			continue
		}
		sourceCountryCode := strings.ToUpper(strings.TrimSpace(country.CountryCode))
		countryCode := normalizeVATComplyCountryCode(sourceCountryCode)
		countryName := strings.TrimSpace(country.CountryName)
		if !isTwoLetterCountryCode(countryCode) || countryName == "" || len([]rune(countryName)) > 120 {
			return nil, 0, fmt.Errorf("VATcomply response contains an invalid country code or name %q", sourceCountryCode)
		}
		if _, exists := countryCodes[countryCode]; exists {
			return nil, 0, fmt.Errorf("VATcomply response contains duplicate normalized country %s", countryCode)
		}
		countryCodes[countryCode] = struct{}{}
		currencyCode := strings.ToUpper(strings.TrimSpace(country.Currency))
		if currencyCode != "" && !isThreeLetterCurrencyCode(currencyCode) {
			return nil, 0, fmt.Errorf("VATcomply response contains invalid currency %q for %s", currencyCode, countryCode)
		}

		appendTaxRateSnapshotEntry := func(rateType, rateCategory string, value json.Number) error {
			decimal, parseErr := normalizeTaxRatePercentage(value)
			if parseErr != nil {
				return fmt.Errorf("VATcomply %s rate for %s: %w", rateType, countryCode, parseErr)
			}
			entryKey := strings.Join([]string{countryCode, rateType, rateCategory, decimal}, "|")
			if _, exists := rateEntryKeys[entryKey]; exists {
				return nil
			}
			rateEntryKeys[entryKey] = struct{}{}
			entries = append(entries, taxrate.TaxRateSourceSnapshotEntry{
				CountryCode:       countryCode,
				SourceCountryCode: sourceCountryCode,
				CountryName:       countryName,
				Currency:          currencyCode,
				RateType:          rateType,
				RateCategory:      rateCategory,
				RateDecimal:       decimal,
				CreatedAt:         capturedAt,
			})
			return nil
		}

		if err := appendTaxRateSnapshotEntry("standard", "", country.StandardRate); err != nil {
			return nil, 0, err
		}
		for _, rate := range country.ReducedRates {
			if err := appendTaxRateSnapshotEntry("reduced", "", rate); err != nil {
				return nil, 0, err
			}
		}
		if country.SuperReducedRate != nil {
			if err := appendTaxRateSnapshotEntry("super_reduced", "", *country.SuperReducedRate); err != nil {
				return nil, 0, err
			}
		}
		if country.ParkingRate != nil {
			if err := appendTaxRateSnapshotEntry("parking", "", *country.ParkingRate); err != nil {
				return nil, 0, err
			}
		}
		categoryNames := make([]string, 0, len(country.RateCategories))
		for category := range country.RateCategories {
			categoryNames = append(categoryNames, category)
		}
		sort.Strings(categoryNames)
		for _, category := range categoryNames {
			category = strings.TrimSpace(category)
			if category == "" || len([]rune(category)) > 96 {
				return nil, 0, fmt.Errorf("VATcomply response contains invalid product rate category for %s", countryCode)
			}
			for _, rate := range country.RateCategories[category] {
				if err := appendTaxRateSnapshotEntry("product_category", category, rate); err != nil {
					return nil, 0, err
				}
			}
		}
	}
	if err := validateVATComplyMemberStateCoverage(countryCodes); err != nil {
		return nil, 0, err
	}
	if len(entries) == 0 {
		return nil, 0, errors.New("VATcomply response contains no tax rate entries")
	}
	sort.Slice(entries, func(left, right int) bool {
		if entries[left].CountryCode != entries[right].CountryCode {
			return entries[left].CountryCode < entries[right].CountryCode
		}
		if entries[left].RateType != entries[right].RateType {
			return entries[left].RateType < entries[right].RateType
		}
		if entries[left].RateCategory != entries[right].RateCategory {
			return entries[left].RateCategory < entries[right].RateCategory
		}
		return entries[left].RateDecimal < entries[right].RateDecimal
	})
	return entries, len(countryCodes), nil
}

func validateVATComplyMemberStateCoverage(actualCountryCodes map[string]struct{}) error {
	missingCountryCodes := make([]string, 0)
	for expectedCountryCode := range vatcomplyExpectedEUMemberStateCountryCodes {
		if _, exists := actualCountryCodes[expectedCountryCode]; !exists {
			missingCountryCodes = append(missingCountryCodes, expectedCountryCode)
		}
	}
	sort.Strings(missingCountryCodes)

	unexpectedCountryCodes := make([]string, 0)
	for actualCountryCode := range actualCountryCodes {
		if _, exists := vatcomplyExpectedEUMemberStateCountryCodes[actualCountryCode]; !exists {
			unexpectedCountryCodes = append(unexpectedCountryCodes, actualCountryCode)
		}
	}
	sort.Strings(unexpectedCountryCodes)

	if len(missingCountryCodes) > 0 || len(unexpectedCountryCodes) > 0 {
		return fmt.Errorf(
			"VATcomply EU member-state coverage mismatch (missing: %s; unexpected: %s)",
			strings.Join(missingCountryCodes, ","),
			strings.Join(unexpectedCountryCodes, ","),
		)
	}
	return nil
}

func normalizeVATComplyCountryCode(sourceCountryCode string) string {
	if sourceCountryCode == "EL" {
		return "GR"
	}
	return sourceCountryCode
}

func isTwoLetterCountryCode(value string) bool {
	if len(value) != 2 {
		return false
	}
	for _, character := range value {
		if character < 'A' || character > 'Z' {
			return false
		}
	}
	return true
}

func isThreeLetterCurrencyCode(value string) bool {
	if len(value) != 3 {
		return false
	}
	for _, character := range value {
		if character < 'A' || character > 'Z' {
			return false
		}
	}
	return true
}

func normalizeTaxRatePercentage(value json.Number) (string, error) {
	text := strings.TrimSpace(value.String())
	if text == "" {
		return "", errors.New("rate is missing")
	}
	rate, ok := new(big.Rat).SetString(text)
	if !ok || rate.Sign() < 0 || rate.Cmp(big.NewRat(100, 1)) > 0 {
		return "", fmt.Errorf("rate %q is not between 0 and 100", text)
	}
	scaled := new(big.Rat).Mul(rate, new(big.Rat).SetInt64(100_000_000))
	if scaled.Denom().Cmp(big.NewInt(1)) != 0 {
		return "", fmt.Errorf("rate %q has more than 8 fractional digits", text)
	}
	decimal := strings.TrimRight(strings.TrimRight(rate.FloatString(8), "0"), ".")
	if decimal == "" || decimal == "-0" {
		decimal = "0"
	}
	return decimal, nil
}
