package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"commerce-platform/internal/domain/shipping"
)

const defaultYanwenOrderSource = "tanzanite-theme"

// YanwenCustomsDeclarationInput contains the optional tax identifiers that
// Yanwen documents on express.order.create. These values stay inside the
// Yanwen order request and are not copied into generic customs or order data.
type YanwenCustomsDeclarationInput struct {
	ReceiverTaxNumber string
	IOSS              string
	EORI              string
}

// validateYanwenChannelCustomsDeclarationRequirements enforces only the
// explicit declaration requirements configured on the selected Yanwen
// channel. It does not infer country rules or claim that a tax identifier is
// officially valid.
func validateYanwenChannelCustomsDeclarationRequirements(
	channel shipping.YanwenPublishedChannel,
	customsDeclaration YanwenCustomsDeclarationInput,
) error {
	if channel.RequireReceiverTaxNumber && strings.TrimSpace(customsDeclaration.ReceiverTaxNumber) == "" {
		return fmt.Errorf("Yanwen channel %q requires a receiver tax number", channel.ProductCode)
	}
	if channel.RequireIOSS && strings.TrimSpace(customsDeclaration.IOSS) == "" {
		return fmt.Errorf("Yanwen channel %q requires an IOSS number", channel.ProductCode)
	}
	if channel.RequireEORI && strings.TrimSpace(customsDeclaration.EORI) == "" {
		return fmt.Errorf("Yanwen channel %q requires an EORI number", channel.ProductCode)
	}
	return nil
}

const (
	maximumYanwenReceiverTaxNumberLength = 50
	maximumYanwenIOSSLength              = 50
	maximumYanwenEORILength              = 64
)

var supportedYanwenDeclarationCurrencies = map[string]struct{}{
	"USD": {},
	"EUR": {},
	"GBP": {},
	"CNY": {},
	"AUD": {},
	"CAD": {},
}

func buildYanwenCreateOrderRequestFromFacts(
	orderFacts *shipping.YanwenOrderFacts,
	productCode string,
	warehouseCode string,
	orderSource string,
	hasBattery bool,
	countryCode string,
	customsDeclaration YanwenCustomsDeclarationInput,
) (YanwenCreateOrderRequest, string, int, int, error) {
	if orderFacts == nil {
		return YanwenCreateOrderRequest{}, "", 0, 0, errors.New("Yanwen order facts are required")
	}
	customsDeclaration.ReceiverTaxNumber = strings.TrimSpace(customsDeclaration.ReceiverTaxNumber)
	customsDeclaration.IOSS = strings.TrimSpace(customsDeclaration.IOSS)
	customsDeclaration.EORI = strings.TrimSpace(customsDeclaration.EORI)
	if len(customsDeclaration.ReceiverTaxNumber) > maximumYanwenReceiverTaxNumberLength {
		return YanwenCreateOrderRequest{}, "", 0, 0, fmt.Errorf("Yanwen receiver tax number cannot exceed %d characters", maximumYanwenReceiverTaxNumberLength)
	}
	if len(customsDeclaration.IOSS) > maximumYanwenIOSSLength {
		return YanwenCreateOrderRequest{}, "", 0, 0, fmt.Errorf("Yanwen IOSS cannot exceed %d characters", maximumYanwenIOSSLength)
	}
	if len(customsDeclaration.EORI) > maximumYanwenEORILength {
		return YanwenCreateOrderRequest{}, "", 0, 0, fmt.Errorf("Yanwen EORI cannot exceed %d characters", maximumYanwenEORILength)
	}
	name := strings.TrimSpace(orderFacts.ShippingAddress.FirstName + " " + orderFacts.ShippingAddress.LastName)
	address := strings.TrimSpace(strings.Join(nonEmptyStrings(orderFacts.ShippingAddress.Address1, orderFacts.ShippingAddress.Address2), ", "))
	if name == "" || address == "" {
		return YanwenCreateOrderRequest{}, "", 0, 0, errors.New("order shipping recipient name and address are required")
	}
	if strings.TrimSpace(orderFacts.OrderNumber) == "" || strings.TrimSpace(productCode) == "" || strings.TrimSpace(warehouseCode) == "" {
		return YanwenCreateOrderRequest{}, "", 0, 0, errors.New("Yanwen order number, product and warehouse are required")
	}
	if strings.TrimSpace(orderSource) == "" {
		orderSource = defaultYanwenOrderSource
	}
	if len(orderSource) > 20 {
		return YanwenCreateOrderRequest{}, "", 0, 0, errors.New("Yanwen order source cannot exceed 20 characters")
	}
	currencyCode := strings.ToUpper(strings.TrimSpace(orderFacts.Currency))
	if _, supported := supportedYanwenDeclarationCurrencies[currencyCode]; !supported {
		return YanwenCreateOrderRequest{}, "", 0, 0, fmt.Errorf("order currency %q is not supported by Yanwen declaration API", currencyCode)
	}
	if strings.TrimSpace(countryCode) == "" {
		return YanwenCreateOrderRequest{}, "", 0, 0, errors.New("Yanwen destination country is required")
	}

	products := make([]YanwenProductDeclare, 0, len(orderFacts.Items))
	totalQuantity := 0
	totalWeight := 0
	descriptions := make([]string, 0, len(orderFacts.Items))
	for index, item := range orderFacts.Items {
		if item.Quantity <= 0 {
			return YanwenCreateOrderRequest{}, "", 0, 0, fmt.Errorf("order item %d quantity must be positive", index)
		}
		if item.WeightGrams <= 0 {
			return YanwenCreateOrderRequest{}, "", 0, 0, fmt.Errorf("order item %d weight must be positive", index)
		}
		if item.DeclaredValueMinor == nil || !item.DeclaredValueConfirmed || *item.DeclaredValueMinor <= 0 {
			return YanwenCreateOrderRequest{}, "", 0, 0, fmt.Errorf("order item %d requires a confirmed positive declared value", index)
		}
		if *item.DeclaredValueMinor%int64(item.Quantity) != 0 {
			return YanwenCreateOrderRequest{}, "", 0, 0, fmt.Errorf("order item %d declared value cannot be represented as a two-decimal unit price", index)
		}
		productName := strings.TrimSpace(item.ProductName)
		if productName == "" {
			return YanwenCreateOrderRequest{}, "", 0, 0, fmt.Errorf("order item %d product name is required", index)
		}
		declarationDescription := strings.TrimSpace(item.CustomsDescription)
		if declarationDescription == "" {
			return YanwenCreateOrderRequest{}, "", 0, 0, fmt.Errorf("order item %d customs description is required before Yanwen declaration", index)
		}
		unitPrice, err := yanwenDecimalFromMinorUnits(*item.DeclaredValueMinor / int64(item.Quantity))
		if err != nil {
			return YanwenCreateOrderRequest{}, "", 0, 0, fmt.Errorf("order item %d declared value: %w", index, err)
		}
		products = append(products, YanwenProductDeclare{
			GoodsNameChinese: productName,
			GoodsNameEnglish: declarationDescription,
			Price:            unitPrice,
			PriceExport:      unitPrice,
			HSCode:           strings.TrimSpace(item.HSCode),
			Quantity:         item.Quantity,
			Weight:           item.WeightGrams,
			SKU:              strings.TrimSpace(item.SKU),
		})
		totalQuantity += item.Quantity
		totalWeight += item.WeightGrams * item.Quantity
		descriptions = append(descriptions, declarationDescription)
	}
	if totalQuantity <= 0 || totalWeight <= 0 || len(products) == 0 {
		return YanwenCreateOrderRequest{}, "", 0, 0, errors.New("Yanwen order must contain positive quantity, weight and products")
	}
	description := strings.Join(uniqueNonEmptyStrings(descriptions), ", ")
	if len(description) > 255 {
		return YanwenCreateOrderRequest{}, "", 0, 0, errors.New("Yanwen declaration description cannot exceed 255 characters")
	}

	request := YanwenCreateOrderRequest{
		ChannelID:     strings.TrimSpace(productCode),
		OrderSource:   orderSource,
		OrderNumber:   strings.TrimSpace(orderFacts.OrderNumber),
		CompanyCode:   strings.TrimSpace(warehouseCode),
		SalesPlatform: defaultYanwenOrderSource,
		ReceiverInfo: YanwenReceiverInfo{
			Name:      name,
			Phone:     strings.TrimSpace(orderFacts.ShippingAddress.Phone),
			Email:     strings.TrimSpace(orderFacts.ShippingAddress.Email),
			Company:   strings.TrimSpace(orderFacts.ShippingAddress.Company),
			Country:   strings.ToUpper(strings.TrimSpace(countryCode)),
			State:     strings.TrimSpace(orderFacts.ShippingAddress.State),
			City:      strings.TrimSpace(orderFacts.ShippingAddress.City),
			ZipCode:   strings.TrimSpace(orderFacts.ShippingAddress.PostalCode),
			Address:   address,
			TaxNumber: customsDeclaration.ReceiverTaxNumber,
		},
		ParcelInfo: YanwenParcelInfo{
			HasBattery:    boolToYanwenBatteryValue(hasBattery),
			Currency:      currencyCode,
			TotalQuantity: totalQuantity,
			TotalWeight:   totalWeight,
			IOSS:          customsDeclaration.IOSS,
			ProductList:   products,
		},
	}
	if customsDeclaration.EORI != "" {
		request.ImportCustomsInfo = &YanwenImportCustomsInfo{EORI: customsDeclaration.EORI}
	}
	if orderFacts.PaidAt != nil {
		request.DateOfReceipt = orderFacts.PaidAt.UTC().Format("2006-01-02")
	}
	return request, description, totalQuantity, totalWeight, nil
}

func validateYanwenChannelCountryScope(rawCountries string, countryValues ...string) error {
	var countries []string
	if err := json.Unmarshal([]byte(rawCountries), &countries); err != nil {
		return fmt.Errorf("decode Yanwen channel country scope: %w", err)
	}
	if len(countries) == 0 {
		return nil
	}
	validValues := make(map[string]struct{}, len(countryValues))
	for _, value := range countryValues {
		value = strings.ToUpper(strings.TrimSpace(value))
		if value != "" {
			validValues[value] = struct{}{}
		}
	}
	for _, country := range countries {
		if _, matches := validValues[strings.ToUpper(strings.TrimSpace(country))]; matches {
			return nil
		}
	}
	return fmt.Errorf("Yanwen channel is not published for destination country %s", strings.Join(countryValues, "/"))
}

func yanwenDecimalFromMinorUnits(minor int64) (json.Number, error) {
	if minor < 0 {
		return "", errors.New("minor amount cannot be negative")
	}
	return json.Number(strconv.FormatInt(minor/100, 10) + "." + fmt.Sprintf("%02d", minor%100)), nil
}

func boolToYanwenBatteryValue(value bool) int {
	if value {
		return 1
	}
	return 0
}

func nonEmptyStrings(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, strings.TrimSpace(value))
		}
	}
	return result
}

func uniqueNonEmptyStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
