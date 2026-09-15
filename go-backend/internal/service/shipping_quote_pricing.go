package service

import (
	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/shipping"
	shippingrating "commerce-platform/internal/domain/shipping/rating"
	"errors"
	"fmt"
	"math"
	"strings"
)

func carrierServiceVolumetricWeightGrams(service shipping.CarrierService, resolvedItems []resolvedShippingItem) (int, int, bool) {
	if service.VolumetricDivisor <= 0 {
		return 0, 0, false
	}

	knownVolumetricWeightGrams := 0
	dimensionalChargeWeightGrams := 0
	complete := len(resolvedItems) > 0
	for _, item := range resolvedItems {
		itemVolumetricWeight, ok := packagingRuleVolumetricWeightGrams(item.PackagingRule, service.VolumetricDivisor)
		if !ok {
			dimensionalChargeWeightGrams += item.ChargeWeightGrams * item.Quantity
			complete = false
			continue
		}
		itemTotalVolumetricWeight := itemVolumetricWeight * item.Quantity
		knownVolumetricWeightGrams += itemTotalVolumetricWeight
		dimensionalChargeWeightGrams += itemTotalVolumetricWeight
	}
	return knownVolumetricWeightGrams, dimensionalChargeWeightGrams, complete && knownVolumetricWeightGrams > 0
}

func packagingRuleVolumetricWeightGrams(rule *shipping.PackagingRule, divisor int) (int, bool) {
	if rule == nil {
		return 0, false
	}
	return shippingrating.VolumetricWeightGrams(rule.BoxLength, rule.BoxWidth, rule.BoxHeight, divisor)
}

func carrierServiceBillableWeightGrams(chargeWeightGrams int, service shipping.CarrierService) int {
	return shippingrating.BillableWeightGrams(
		chargeWeightGrams,
		service.MinChargeWeightGrams,
		shippingrating.WeightScale{
			FirstWeightGrams:      service.FirstWeightGrams,
			AdditionalWeightGrams: service.AdditionalWeightGrams,
		},
	)
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func uniqueShippingQuoteProductIDs(items []ShippingQuoteItemInput) []uint {
	seen := make(map[uint]struct{})
	productIDs := make([]uint, 0, len(items))
	for _, item := range items {
		if item.ProductID == 0 {
			continue
		}
		if _, ok := seen[item.ProductID]; ok {
			continue
		}
		seen[item.ProductID] = struct{}{}
		productIDs = append(productIDs, item.ProductID)
	}
	return productIDs
}

func uniqueShippingQuoteTemplateIDs(items []ShippingQuoteItemInput) ([]uint, error) {
	seen := make(map[uint]struct{})
	templateIDs := make([]uint, 0, len(items))
	for _, item := range items {
		if item.ShippingTemplateID == nil || *item.ShippingTemplateID == 0 {
			if item.VariantID != nil {
				return nil, fmt.Errorf("shipping template is missing for variant ID %d", *item.VariantID)
			}
			return nil, fmt.Errorf("shipping template is missing for product ID %d", item.ProductID)
		}

		templateID := *item.ShippingTemplateID
		if _, ok := seen[templateID]; ok {
			continue
		}
		seen[templateID] = struct{}{}
		templateIDs = append(templateIDs, templateID)
	}
	return templateIDs, nil
}

func resolveProductShippingTemplateID(p *product.Product, variant *product.ProductVariant) (uint, error) {
	if variant != nil && variant.ShippingTemplateID != nil && *variant.ShippingTemplateID != 0 {
		return *variant.ShippingTemplateID, nil
	}
	if p != nil && p.ShippingTemplateID != nil && *p.ShippingTemplateID != 0 {
		return *p.ShippingTemplateID, nil
	}
	if variant != nil {
		return 0, fmt.Errorf("shipping template is missing for SKU %s", variant.SKU)
	}
	if p != nil {
		return 0, fmt.Errorf("shipping template is missing for product ID %d", p.ID)
	}
	return 0, errors.New("shipping template is missing")
}

func packagingRuleWeightGrams(rule *shipping.PackagingRule) int {
	if rule == nil || rule.BoxWeight <= 0 {
		return 0
	}
	return int(math.Round(rule.BoxWeight * 1000))
}

func calculateTemplateShippingFeeWithDisplayPrices(
	template *shipping.ShippingTemplate,
	country string,
	totalWeightGrams int,
	quantity int,
	amount float64,
	cartAmount float64,
	weightBilling ...shippingWeightBilling,
) (float64, bool, []currency.DisplayPriceSnapshot, error) {
	if template == nil {
		return 0, false, nil, errors.New("shipping template is required")
	}

	freeThresholdReached := cartAmount >= template.FreeThreshold
	if template.Type == "price" || template.Type == "amount" {
		if reached, ok := majorAmountAtLeast(cartAmount, template.FreeThreshold, template.Currency); ok {
			freeThresholdReached = reached
		}
	}
	if template.FreeShipping && freeThresholdReached {
		return 0, true, nil, nil
	}

	value := float64(totalWeightGrams) / 1000
	switch template.Type {
	case "quantity", "items":
		value = float64(quantity)
	case "price", "amount":
		value = amount
	}
	return calculateTemplateShippingFeeForValue(template, country, value, weightBilling...)
}

// calculateTemplateShippingFeeForQuote evaluates thresholds in the template's
// source currency and returns the resulting fee in the checkout currency.
func (s *ShippingService) calculateTemplateShippingFeeForQuote(
	template *shipping.ShippingTemplate,
	country string,
	totalWeightGrams int,
	quantity int,
	amount float64,
	cartAmount float64,
	quoteCurrency string,
	weightBilling ...shippingWeightBilling,
) (float64, bool, []currency.DisplayPriceSnapshot, error) {
	if template == nil {
		return 0, false, nil, errors.New("shipping template is required")
	}

	sourceCurrency := currency.NormalizeCode(template.Currency)
	quoteCurrency = currency.NormalizeCode(quoteCurrency)
	if !currency.IsCatalogCode(sourceCurrency) {
		return 0, false, nil, fmt.Errorf("shipping template ID %d has invalid source currency", template.ID)
	}
	if !currency.IsCatalogCode(quoteCurrency) {
		return 0, false, nil, fmt.Errorf("shipping quote currency %s is invalid", quoteCurrency)
	}

	sourceAmount := amount
	sourceCartAmount := cartAmount
	var err error
	if sourceCurrency != quoteCurrency && (template.Type == "price" || template.Type == "amount") {
		sourceAmount, err = s.convertShippingAmount(amount, quoteCurrency, sourceCurrency)
		if err != nil {
			return 0, false, nil, err
		}
	}
	if sourceCurrency != quoteCurrency && template.FreeShipping && template.FreeThreshold > 0 {
		sourceCartAmount, err = s.convertShippingAmount(cartAmount, quoteCurrency, sourceCurrency)
		if err != nil {
			return 0, false, nil, err
		}
	}
	fee, freeShipping, displayPrices, err := calculateTemplateShippingFeeWithDisplayPrices(
		template,
		country,
		totalWeightGrams,
		quantity,
		sourceAmount,
		sourceCartAmount,
		weightBilling...,
	)
	if err != nil {
		return 0, false, nil, err
	}
	if sourceCurrency == quoteCurrency || fee <= 0 {
		return roundMoney(fee, quoteCurrency), freeShipping, displayPrices, nil
	}

	fee, err = s.convertShippingAmount(fee, sourceCurrency, quoteCurrency)
	if err != nil {
		return 0, false, nil, err
	}
	return roundMoney(fee, quoteCurrency), freeShipping, displayPrices, nil
}

func (s *ShippingService) convertShippingAmount(amount float64, fromCurrency string, toCurrency string) (float64, error) {
	fromCurrency = currency.NormalizeCode(fromCurrency)
	toCurrency = currency.NormalizeCode(toCurrency)
	if amount == 0 || fromCurrency == toCurrency {
		return roundMoney(amount, toCurrency), nil
	}
	if amount < 0 {
		return 0, fmt.Errorf("shipping amount cannot be negative")
	}
	if !currency.IsCatalogCode(fromCurrency) || !currency.IsCatalogCode(toCurrency) {
		return 0, fmt.Errorf("shipping currency conversion from %s to %s is invalid", fromCurrency, toCurrency)
	}
	if s == nil || s.exchangeRates == nil {
		return 0, fmt.Errorf("%w: exchange rate unavailable for %s to %s shipping conversion", ErrShippingRateUnavailable, fromCurrency, toCurrency)
	}

	amountMoney, amountErr := domainmoney.FromMajorFloat(amount, fromCurrency)
	if amountErr != nil {
		return 0, amountErr
	}
	convertedMoney, conversionErr := s.exchangeRates.ConvertMoneyStrict(amountMoney, toCurrency)
	if conversionErr != nil {
		return 0, fmt.Errorf("%w: exchange rate unavailable for %s to %s shipping conversion: %w", ErrShippingRateUnavailable, fromCurrency, toCurrency, conversionErr)
	}
	return convertedMoney.MajorFloat()
}

func calculateTemplateShippingFeeForValue(
	template *shipping.ShippingTemplate,
	country string,
	value float64,
	weightBilling ...shippingWeightBilling,
) (float64, bool, []currency.DisplayPriceSnapshot, error) {
	if template == nil {
		return 0, false, nil, errors.New("shipping template is required")
	}
	shippingFee := template.DefaultFee
	displayPrices := templateFeeDisplayPrices(template)
	matchedRule := false
	countryMatched := false
	ruleWeightBilling := weightBilling
	if template.Type != "weight" {
		ruleWeightBilling = nil
	}
	for _, rule := range template.Rules {
		if shippingrating.MatchesRegion(rule.Region, country) {
			countryMatched = true
		}
		if shippingrating.MatchesRegion(rule.Region, country) && shippingRuleMatchesTemplateValue(template.Type, template.Currency, rule, value) {
			if len(ruleWeightBilling) > 0 {
				if err := validateShippingWeightBilling(rule, ruleWeightBilling[0]); err != nil {
					return 0, false, nil, err
				}
			}
			shippingFee = calculateRuleFee(template.Type, template.Currency, rule, value, ruleWeightBilling...)
			displayPrices = ruleFeeDisplayPrices(template.Type, template.Currency, rule, value, ruleWeightBilling...)
			matchedRule = true
			break
		}
	}
	if !matchedRule && template.DefaultFee <= 0 {
		cause := ErrShippingRateUnavailable
		if !countryMatched {
			cause = ErrCountryNotSupported
		}
		return 0, false, nil, fmt.Errorf(
			"%w: shipping template %q (ID %d) has no applicable rule for country %s and charge value %.2f, and no positive default fee",
			cause,
			template.Name,
			template.ID,
			strings.ToUpper(strings.TrimSpace(country)),
			value,
		)
	}

	shippingFee = roundMoney(shippingFee, template.Currency)
	if shippingFee <= 0 {
		return shippingFee, matchedRule, nil, nil
	}
	return shippingFee, false, displayPrices, nil
}

func shippingRuleMatchesValue(rule shipping.ShippingRule, value float64) bool {
	return value >= rule.MinValue && (rule.MaxValue == 0 || value <= rule.MaxValue)
}

func shippingRuleMatchesTemplateValue(templateType, templateCurrency string, rule shipping.ShippingRule, value float64) bool {
	if templateType != "price" && templateType != "amount" {
		return shippingRuleMatchesValue(rule, value)
	}
	valueMoney, err := domainmoney.FromMajorFloat(value, templateCurrency)
	if err != nil {
		return false
	}
	minMoney, err := domainmoney.FromMajorFloat(rule.MinValue, templateCurrency)
	if err != nil || valueMoney.AmountMinor() < minMoney.AmountMinor() {
		return false
	}
	if rule.MaxValue <= 0 {
		return true
	}
	maxMoney, err := domainmoney.FromMajorFloat(rule.MaxValue, templateCurrency)
	return err == nil && valueMoney.AmountMinor() <= maxMoney.AmountMinor()
}

func validateTemplateWeightBillingForValue(
	template *shipping.ShippingTemplate,
	country string,
	value float64,
	weightBilling ...shippingWeightBilling,
) error {
	if template == nil || template.Type != "weight" {
		return nil
	}
	for _, rule := range template.Rules {
		if !shippingrating.MatchesRegion(rule.Region, country) || !shippingRuleMatchesTemplateValue(template.Type, template.Currency, rule, value) {
			continue
		}
		if rule.Additional <= 0 {
			return nil
		}
		if len(weightBilling) == 0 {
			return fmt.Errorf(
				"%w: weight rule ID %d requires carrier first/additional weight configuration",
				ErrShippingRateConfigurationInvalid,
				rule.ID,
			)
		}
		return validateShippingWeightBilling(rule, weightBilling[0])
	}
	return nil
}

func validateShippingWeightBilling(rule shipping.ShippingRule, billing shippingWeightBilling) error {
	if err := shippingrating.ValidateWeightScale(rule.Additional, shippingrating.WeightScale{
		FirstWeightGrams: billing.firstWeightGrams, AdditionalWeightGrams: billing.additionalWeightGrams,
	}); err != nil {
		return fmt.Errorf(
			"%w: weight rule ID %d requires positive carrier first/additional weight configuration",
			ErrShippingRateConfigurationInvalid,
			rule.ID,
		)
	}
	return nil
}

// shippingWeightBilling carries the carrier-specific weight scale. Weight
// rules with an additional fee must receive this configuration explicitly.
type shippingWeightBilling struct {
	firstWeightGrams      int
	additionalWeightGrams int
}

func calculateRuleFee(templateType, templateCurrency string, rule shipping.ShippingRule, value float64, weightBilling ...shippingWeightBilling) float64 {
	feeCurrency := currency.NormalizeCode(rule.Currency)
	if feeCurrency == "" {
		feeCurrency = currency.NormalizeCode(templateCurrency)
	}
	if feeCurrency != currency.NormalizeCode(templateCurrency) {
		return 0
	}
	feeMoney, err := domainmoney.FromMajorFloat(rule.Fee, feeCurrency)
	if err != nil {
		return 0
	}
	additionalUnits := calculateRuleAdditionalUnitsForTemplate(templateType, templateCurrency, rule, value, weightBilling...)
	if additionalUnits > 0 {
		additionalMoney, addErr := domainmoney.FromMajorFloat(rule.Additional, feeCurrency)
		if addErr != nil {
			return 0
		}
		additionalMoney, addErr = additionalMoney.MultiplyInt(int64(additionalUnits))
		if addErr != nil {
			return 0
		}
		feeMoney, addErr = feeMoney.Add(additionalMoney)
		if addErr != nil {
			return 0
		}
	}
	fee, err := feeMoney.MajorFloat()
	if err != nil {
		return 0
	}
	return fee
}

func majorAmountAtLeast(value, threshold float64, code string) (bool, bool) {
	valueMoney, err := domainmoney.FromMajorFloat(value, code)
	if err != nil {
		return false, false
	}
	thresholdMoney, err := domainmoney.FromMajorFloat(threshold, code)
	if err != nil {
		return false, false
	}
	return valueMoney.AmountMinor() >= thresholdMoney.AmountMinor(), true
}

func calculateRuleAdditionalUnitsForTemplate(
	templateType string,
	templateCurrency string,
	rule shipping.ShippingRule,
	value float64,
	weightBilling ...shippingWeightBilling,
) int {
	if templateType == "weight" {
		return calculateRuleAdditionalUnits(rule, value, weightBilling...)
	}
	if templateType == "price" || templateType == "amount" {
		feeCurrency := currency.NormalizeCode(rule.Currency)
		if feeCurrency == "" {
			feeCurrency = currency.NormalizeCode(templateCurrency)
		}
		minMoney, err := domainmoney.FromMajorFloat(rule.MinValue, feeCurrency)
		valueMoney, valueErr := domainmoney.FromMajorFloat(value, feeCurrency)
		additionalMoney, additionalErr := domainmoney.FromMajorFloat(rule.Additional, feeCurrency)
		if err != nil || valueErr != nil || additionalErr != nil || additionalMoney.AmountMinor() <= 0 || valueMoney.AmountMinor() <= minMoney.AmountMinor() {
			return 0
		}
		excess := valueMoney.AmountMinor() - minMoney.AmountMinor()
		units := (excess + additionalMoney.AmountMinor() - 1) / additionalMoney.AmountMinor()
		if units <= 0 {
			return 0
		}
		return int(units)
	}
	return shippingrating.AdditionalWholeUnits(rule.MinValue, value, rule.Additional)
}

func calculateRuleAdditionalUnits(rule shipping.ShippingRule, value float64, weightBilling ...shippingWeightBilling) int {
	if len(weightBilling) == 0 {
		return 0
	}
	billing := weightBilling[0]
	valueGrams := int(math.Round(value * 1000))
	units, err := shippingrating.AdditionalWeightUnits(valueGrams, rule.Additional, shippingrating.WeightScale{
		FirstWeightGrams: billing.firstWeightGrams, AdditionalWeightGrams: billing.additionalWeightGrams,
	})
	if err != nil {
		return 0
	}
	return units
}

func hasPositiveID(id *uint) bool {
	return id != nil && *id > 0
}

func uintPtr(id uint) *uint {
	return &id
}
