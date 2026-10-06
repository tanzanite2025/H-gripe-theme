package service

import (
	"commerce-platform/internal/domain/shipping"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Yanwen official country, warehouse, and product directory gateway methods.

func (c *YanwenGatewayClient) FetchYanwenOfficialCountries(ctx context.Context, credentials yanwenGatewayCredentials) ([]shipping.YanwenCountryCatalogEntry, error) {
	responseBody, _, err := c.sendYanwenGatewayRequest(ctx, credentials, yanwenCountryListMethod, []byte(`{}`))
	if err != nil {
		return nil, err
	}
	return parseYanwenCountryResponse(responseBody)
}

func (c *YanwenGatewayClient) FetchYanwenOfficialWarehouses(ctx context.Context, credentials yanwenGatewayCredentials) ([]shipping.YanwenWarehouseCatalogEntry, error) {
	responseBody, _, err := c.sendYanwenGatewayRequest(ctx, credentials, yanwenWarehouseListMethod, []byte(`{}`))
	if err != nil {
		return nil, err
	}
	return parseYanwenWarehouseResponse(responseBody)
}

func (c *YanwenGatewayClient) FetchYanwenOfficialProducts(ctx context.Context, credentials yanwenGatewayCredentials) ([]shipping.YanwenProductCatalogEntry, error) {
	responseBody, _, err := c.sendYanwenGatewayRequest(ctx, credentials, yanwenProductListMethod, []byte(`{}`))
	if err != nil {
		return nil, err
	}
	return parseYanwenProductResponse(responseBody)
}

func parseYanwenProductResponse(responseBody []byte) ([]shipping.YanwenProductCatalogEntry, error) {
	response, err := decodeYanwenGatewayResponse(responseBody)
	if err != nil {
		return nil, err
	}
	if err := validateYanwenGatewayResponse(response); err != nil {
		return nil, err
	}
	if len(response.Data) == 0 || strings.EqualFold(strings.TrimSpace(string(response.Data)), "null") {
		return nil, errors.New("Yanwen product response is missing data")
	}
	var records []struct {
		ID     json.RawMessage `json:"id"`
		NameCh string          `json:"nameCh"`
		NameEn string          `json:"nameEn"`
	}
	if err := json.Unmarshal(response.Data, &records); err != nil {
		return nil, fmt.Errorf("decode Yanwen product data: %w", err)
	}
	products := make([]shipping.YanwenProductCatalogEntry, 0, len(records))
	seenIDs := make(map[string]struct{}, len(records))
	for index, record := range records {
		productID := normalizeYanwenProductID(record.ID)
		if productID == "" {
			return nil, fmt.Errorf("Yanwen product data record %d is missing id", index)
		}
		if strings.TrimSpace(record.NameCh) == "" && strings.TrimSpace(record.NameEn) == "" {
			return nil, fmt.Errorf("Yanwen product %s is missing name", productID)
		}
		if _, exists := seenIDs[productID]; exists {
			return nil, fmt.Errorf("Yanwen product response contains duplicate id %s", productID)
		}
		seenIDs[productID] = struct{}{}
		products = append(products, shipping.YanwenProductCatalogEntry{
			ProductID:   productID,
			NameChinese: strings.TrimSpace(record.NameCh),
			NameEnglish: strings.TrimSpace(record.NameEn),
		})
	}
	return products, nil
}

func parseYanwenCountryResponse(responseBody []byte) ([]shipping.YanwenCountryCatalogEntry, error) {
	response, err := decodeYanwenGatewayResponse(responseBody)
	if err != nil {
		return nil, err
	}
	if err := validateYanwenGatewayResponse(response); err != nil {
		return nil, err
	}
	if len(response.Data) == 0 || strings.EqualFold(strings.TrimSpace(string(response.Data)), "null") {
		return nil, errors.New("Yanwen country response is missing data")
	}
	var records []struct {
		ID     json.RawMessage `json:"id"`
		Code   string          `json:"code"`
		NameCh string          `json:"nameCh"`
		NameEn string          `json:"nameEn"`
	}
	if err := json.Unmarshal(response.Data, &records); err != nil {
		return nil, fmt.Errorf("decode Yanwen country data: %w", err)
	}
	countries := make([]shipping.YanwenCountryCatalogEntry, 0, len(records))
	seenIDs := make(map[string]struct{}, len(records))
	seenCodes := make(map[string]struct{}, len(records))
	for index, record := range records {
		countryID := normalizeYanwenProductID(record.ID)
		if countryID == "" {
			return nil, fmt.Errorf("Yanwen country data record %d is missing id", index)
		}
		countryCode := strings.ToUpper(strings.TrimSpace(record.Code))
		if countryCode == "" {
			return nil, fmt.Errorf("Yanwen country %s is missing code", countryID)
		}
		if strings.TrimSpace(record.NameCh) == "" && strings.TrimSpace(record.NameEn) == "" {
			return nil, fmt.Errorf("Yanwen country %s is missing name", countryID)
		}
		if _, exists := seenIDs[countryID]; exists {
			return nil, fmt.Errorf("Yanwen country response contains duplicate id %s", countryID)
		}
		if _, exists := seenCodes[countryCode]; exists {
			return nil, fmt.Errorf("Yanwen country response contains duplicate code %s", countryCode)
		}
		seenIDs[countryID] = struct{}{}
		seenCodes[countryCode] = struct{}{}
		countries = append(countries, shipping.YanwenCountryCatalogEntry{
			CountryID:   countryID,
			CountryCode: countryCode,
			NameChinese: strings.TrimSpace(record.NameCh),
			NameEnglish: strings.TrimSpace(record.NameEn),
		})
	}
	return countries, nil
}

func parseYanwenWarehouseResponse(responseBody []byte) ([]shipping.YanwenWarehouseCatalogEntry, error) {
	response, err := decodeYanwenGatewayResponse(responseBody)
	if err != nil {
		return nil, err
	}
	if err := validateYanwenGatewayResponse(response); err != nil {
		return nil, err
	}
	if len(response.Data) == 0 || strings.EqualFold(strings.TrimSpace(string(response.Data)), "null") {
		return nil, errors.New("Yanwen warehouse response is missing data")
	}
	var records []struct {
		Code string `json:"code"`
		Name string `json:"name"`
		Area string `json:"area"`
	}
	if err := json.Unmarshal(response.Data, &records); err != nil {
		return nil, fmt.Errorf("decode Yanwen warehouse data: %w", err)
	}
	warehouses := make([]shipping.YanwenWarehouseCatalogEntry, 0, len(records))
	seenCodes := make(map[string]struct{}, len(records))
	for index, record := range records {
		code := strings.TrimSpace(record.Code)
		if code == "" {
			return nil, fmt.Errorf("Yanwen warehouse data record %d is missing code", index)
		}
		if strings.TrimSpace(record.Name) == "" {
			return nil, fmt.Errorf("Yanwen warehouse %s is missing name", code)
		}
		if _, exists := seenCodes[code]; exists {
			return nil, fmt.Errorf("Yanwen warehouse response contains duplicate code %s", code)
		}
		seenCodes[code] = struct{}{}
		warehouses = append(warehouses, shipping.YanwenWarehouseCatalogEntry{
			WarehouseCode: code,
			Name:          strings.TrimSpace(record.Name),
			Area:          strings.TrimSpace(record.Area),
		})
	}
	return warehouses, nil
}

func normalizeYanwenProductID(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var stringValue string
	if json.Unmarshal(raw, &stringValue) == nil {
		return strings.TrimSpace(stringValue)
	}
	var numberValue json.Number
	if json.Unmarshal(raw, &numberValue) == nil {
		return strings.TrimSpace(numberValue.String())
	}
	return ""
}

func parseRequiredYanwenIntegerField(raw json.RawMessage, fieldName string) (int, error) {
	if len(raw) == 0 || strings.EqualFold(strings.TrimSpace(string(raw)), "null") {
		return 0, fmt.Errorf("Yanwen get order response is missing %s", fieldName)
	}
	value := strings.TrimSpace(string(raw))
	if strings.HasPrefix(value, "\"") {
		var stringValue string
		if err := json.Unmarshal(raw, &stringValue); err != nil {
			return 0, fmt.Errorf("Yanwen get order response has invalid %s: %w", fieldName, err)
		}
		value = strings.TrimSpace(stringValue)
	}
	if value == "" {
		return 0, fmt.Errorf("Yanwen get order response is missing %s", fieldName)
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("Yanwen get order response has invalid %s", fieldName)
	}
	return parsed, nil
}
