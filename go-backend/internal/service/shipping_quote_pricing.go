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

// calculateTemplateShippingFeeWithDisplayPricesMoney is the money-native
// pricing entry point. All monetary thresholds and fees stay in minor units;
// float64 is used only for dimensional (weight/quantity) rule values and for
// display snapshots at the boundary.
func calculateTemplateShippingFeeWithDisplayPricesMoney(
	template *shipping.ShippingTemplate,
	country string,
	totalWeightGrams int,
	quantity int,
	amount domainmoney.Money,
	cartAmount domainmoney.Money,
	weightBilling ...shippingWeightBilling,
) (domainmoney.Money, bool, []currency.DisplayPriceSnapshot, error) {
	if template == nil {
		return domainmoney.Money{}, false, nil, errors.New("shipping template is required")
	}

	freeThresholdMoney, thresholdErr := template.FreeThresholdMoney()
	if thresholdErr != nil {
		return domainmoney.Money{}, false, nil, thresholdErr
	}
	if amount.Currency().String() != currency.NormalizeCode(template.Currency) || cartAmount.Currency().String() != currency.NormalizeCode(template.Currency) {
		return domainmoney.Money{}, false, nil, domainmoney.ErrCurrencyMismatch
	}
	freeThresholdReached := cartAmount.AmountMinor() >= freeThresholdMoney.AmountMinor()
	if template.Type == "price" || template.Type == "amount" {
		freeThresholdReached = cartAmount.AmountMinor() >= freeThresholdMoney.AmountMinor()
	}
	if template.FreeShipping && freeThresholdReached {
		zero, err := domainmoney.New(0, template.Currency)
		return zero, true, nil, err
	}

	value := float64(totalWeightGrams) / 1000
	switch template.Type {
	case "quantity", "items":
		value = float64(quantity)
	case "price", "amount":
		// Price/amount rules stay entirely in Money/minor units. The scalar
		// value is only used by dimensional (weight/quantity) rules.
		value = 0
	}
	return calculateTemplateShippingFeeForMoneyValue(template, country, value, amount, weightBilling...)
}

// calculateTemplateShippingFeeForMoneyValue evaluates a template while
// retaining the exact price/amount value as Money. Weight and quantity remain
// scalar dimensions; only their rule thresholds are non-monetary values.
func calculateTemplateShippingFeeForMoneyValue(
	template *shipping.ShippingTemplate,
	country string,
	value float64,
	amount domainmoney.Money,
	weightBilling ...shippingWeightBilling,
) (domainmoney.Money, bool, []currency.DisplayPriceSnapshot, error) {
	if template == nil {
		return domainmoney.Money{}, false, nil, errors.New("shipping template is required")
	}
	defaultFeeMoney, err := template.DefaultFeeMoney()
	if err != nil {
		return domainmoney.Money{}, false, nil, err
	}
	displayPrices := templateFeeDisplayPrices(template)
	matchedRule := false
	countryMatched := false
	ruleWeightBilling := weightBilling
	if template.Type != "weight" {
		ruleWeightBilling = nil
	}
	for _, rule := range template.Rules {
		regionMatches := shippingrating.MatchesRegion(rule.Region, country)
		if regionMatches {
			countryMatched = true
		}
		matched := false
		if regionMatches {
			if template.Type == "price" || template.Type == "amount" {
				matched = shippingRuleMatchesTemplateMoneyValue(template.Currency, rule, amount)
			} else {
				matched = shippingRuleMatchesTemplateValue(template.Type, template.Currency, rule, value)
			}
		}
		if !matched {
			continue
		}
		if len(ruleWeightBilling) > 0 {
			if err := validateShippingWeightBilling(rule, ruleWeightBilling[0]); err != nil {
				return domainmoney.Money{}, false, nil, err
			}
		}
		feeMoney := calculateRuleFeeMoney(template.Type, template.Currency, rule, value, &amount, ruleWeightBilling...)
		if feeMoney.Validate() != nil {
			return domainmoney.Money{}, false, nil, fmt.Errorf("invalid shipping fee for rule ID %d", rule.ID)
		}
		displayPrices = ruleFeeDisplayPrices(template.Type, template.Currency, rule, value, &amount, ruleWeightBilling...)
		defaultFeeMoney = feeMoney
		matchedRule = true
		break
	}
	if !matchedRule && defaultFeeMoney.AmountMinor() <= 0 {
		cause := ErrShippingRateUnavailable
		if !countryMatched {
			cause = ErrCountryNotSupported
		}
		return domainmoney.Money{}, false, nil, fmt.Errorf(
			"%w: shipping template %q (ID %d) has no applicable rule for country %s and charge value %.2f, and no positive default fee",
			cause,
			template.Name,
			template.ID,
			strings.ToUpper(strings.TrimSpace(country)),
			value,
		)
	}
	return defaultFeeMoney, false, displayPrices, nil
}

func shippingRuleMatchesTemplateMoneyValue(templateCurrency string, rule shipping.ShippingRule, value domainmoney.Money) bool {
	code := currency.NormalizeCode(templateCurrency)
	if value.Currency().String() != code {
		return false
	}
	minMoney, err := rule.MinValueMoney(code)
	if err != nil || value.AmountMinor() < minMoney.AmountMinor() {
		return false
	}
	maxMoney, err := rule.MaxValueMoney(code)
	if err != nil || maxMoney.AmountMinor() <= 0 {
		return true
	}
	return value.AmountMinor() <= maxMoney.AmountMinor()
}

// calculateTemplateShippingFeeForQuoteMoney converts the exact quote amounts
// into the template currency, evaluates the template, then converts the fee
// back to the checkout currency. No major-unit arithmetic is performed.
func (s *ShippingService) calculateTemplateShippingFeeForQuoteMoney(
	template *shipping.ShippingTemplate,
	country string,
	totalWeightGrams int,
	quantity int,
	amount domainmoney.Money,
	cartAmount domainmoney.Money,
	quoteCurrency string,
	weightBilling ...shippingWeightBilling,
) (domainmoney.Money, bool, []currency.DisplayPriceSnapshot, error) {
	if template == nil {
		return domainmoney.Money{}, false, nil, errors.New("shipping template is required")
	}

	sourceCurrency := currency.NormalizeCode(template.Currency)
	quoteCurrency = currency.NormalizeCode(quoteCurrency)
	if !currency.IsCatalogCode(sourceCurrency) {
		return domainmoney.Money{}, false, nil, fmt.Errorf("shipping template ID %d has invalid source currency", template.ID)
	}
	if !currency.IsCatalogCode(quoteCurrency) {
		return domainmoney.Money{}, false, nil, fmt.Errorf("shipping quote currency %s is invalid", quoteCurrency)
	}
	if amount.Currency().String() != quoteCurrency || cartAmount.Currency().String() != quoteCurrency {
		return domainmoney.Money{}, false, nil, domainmoney.ErrCurrencyMismatch
	}

	sourceAmount := amount
	sourceCartAmount := cartAmount
	var err error
	if sourceCurrency != quoteCurrency && (template.Type == "price" || template.Type == "amount") {
		sourceAmount, err = s.convertShippingMoney(amount, sourceCurrency)
		if err != nil {
			return domainmoney.Money{}, false, nil, err
		}
	} else if sourceCurrency != quoteCurrency {
		// Weight/quantity templates do not inspect merchandise amount, but the
		// money-native evaluator still requires a value in the template currency.
		sourceAmount, err = domainmoney.New(0, sourceCurrency)
		if err != nil {
			return domainmoney.Money{}, false, nil, err
		}
	}
	if sourceCurrency != quoteCurrency && template.FreeShipping && template.FreeThresholdMinor > 0 {
		sourceCartAmount, err = s.convertShippingMoney(cartAmount, sourceCurrency)
		if err != nil {
			return domainmoney.Money{}, false, nil, err
		}
	} else if sourceCurrency != quoteCurrency {
		sourceCartAmount, err = domainmoney.New(0, sourceCurrency)
		if err != nil {
			return domainmoney.Money{}, false, nil, err
		}
	}
	fee, freeShipping, displayPrices, err := calculateTemplateShippingFeeWithDisplayPricesMoney(
		template,
		country,
		totalWeightGrams,
		quantity,
		sourceAmount,
		sourceCartAmount,
		weightBilling...,
	)
	if err != nil {
		return domainmoney.Money{}, false, nil, err
	}
	if sourceCurrency == quoteCurrency || fee.AmountMinor() <= 0 {
		if sourceCurrency == quoteCurrency {
			return fee, freeShipping, displayPrices, nil
		}
		converted, convertErr := s.convertShippingMoney(fee, quoteCurrency)
		return converted, freeShipping, displayPrices, convertErr
	}
	converted, err := s.convertShippingMoney(fee, quoteCurrency)
	if err != nil {
		return domainmoney.Money{}, false, nil, err
	}
	return converted, freeShipping, displayPrices, nil
}

func (s *ShippingService) convertShippingMoney(amount domainmoney.Money, toCurrency string) (domainmoney.Money, error) {
	if err := amount.Validate(); err != nil {
		return domainmoney.Money{}, err
	}
	toCurrency = currency.NormalizeCode(toCurrency)
	if !currency.IsCatalogCode(toCurrency) {
		return domainmoney.Money{}, fmt.Errorf("shipping target currency %s is invalid", toCurrency)
	}
	if amount.AmountMinor() == 0 || amount.Currency().String() == toCurrency {
		return domainmoney.New(amount.AmountMinor(), toCurrency)
	}
	if amount.AmountMinor() < 0 {
		return domainmoney.Money{}, fmt.Errorf("shipping amount cannot be negative")
	}
	if s == nil || s.exchangeRates == nil {
		return domainmoney.Money{}, fmt.Errorf("%w: exchange rate unavailable for %s to %s shipping conversion", ErrShippingRateUnavailable, amount.Currency(), toCurrency)
	}
	converted, err := s.exchangeRates.ConvertMoneyStrict(amount, toCurrency)
	if err != nil {
		return domainmoney.Money{}, fmt.Errorf("%w: exchange rate unavailable for %s to %s shipping conversion: %w", ErrShippingRateUnavailable, amount.Currency(), toCurrency, err)
	}
	return converted, nil
}

func validateShippingWeightBilling(rule shipping.ShippingRule, billing shippingWeightBilling) error {
	additionalFee := 0.0
	if rule.AdditionalMinor > 0 {
		additionalFee = 1
	}
	if err := shippingrating.ValidateWeightScale(additionalFee, shippingrating.WeightScale{
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

func shippingRuleMatchesValue(rule shipping.ShippingRule, value float64) bool {
	return value >= rule.MinValue && (rule.MaxValue == 0 || value <= rule.MaxValue)
}

func shippingRuleMatchesTemplateValue(templateType, _ string, rule shipping.ShippingRule, value float64) bool {
	// Monetary templates use shippingRuleMatchesTemplateMoneyValue. This
	// helper is only for dimensional weight/quantity thresholds.
	if templateType == "price" || templateType == "amount" {
		return false
	}
	return shippingRuleMatchesValue(rule, value)
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
		additionalMoney, additionalErr := rule.AdditionalMoney(template.Currency)
		if additionalErr != nil || additionalMoney.AmountMinor() <= 0 {
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

// shippingWeightBilling carries the carrier-specific weight scale. Weight
// rules with an additional fee must receive this configuration explicitly.
type shippingWeightBilling struct {
	firstWeightGrams      int
	additionalWeightGrams int
}

// calculateRuleFeeMoney performs all fee composition in minor units. The
// optional amount is supplied for price/amount rules so additional-unit
// calculations never round-trip through float64.
func calculateRuleFeeMoney(templateType, templateCurrency string, rule shipping.ShippingRule, value float64, amount *domainmoney.Money, weightBilling ...shippingWeightBilling) domainmoney.Money {
	feeCurrency := currency.NormalizeCode(rule.Currency)
	if feeCurrency == "" {
		feeCurrency = currency.NormalizeCode(templateCurrency)
	}
	if feeCurrency != currency.NormalizeCode(templateCurrency) {
		return domainmoney.Money{}
	}
	feeMoney, err := rule.FeeMoney(feeCurrency)
	if err != nil {
		return domainmoney.Money{}
	}
	additionalUnits := 0
	if (templateType == "price" || templateType == "amount") && amount != nil {
		additionalUnits = calculateRuleAdditionalUnitsForMoneyTemplate(templateCurrency, rule, *amount)
	} else {
		additionalUnits = calculateRuleAdditionalUnitsForTemplate(templateType, templateCurrency, rule, value, weightBilling...)
	}
	if additionalUnits > 0 {
		additionalMoney, addErr := rule.AdditionalMoney(feeCurrency)
		if addErr != nil {
			return domainmoney.Money{}
		}
		additionalMoney, addErr = additionalMoney.MultiplyInt(int64(additionalUnits))
		if addErr != nil {
			return domainmoney.Money{}
		}
		feeMoney, addErr = feeMoney.Add(additionalMoney)
		if addErr != nil {
			return domainmoney.Money{}
		}
	}
	return feeMoney
}

func calculateRuleAdditionalUnitsForMoneyTemplate(templateCurrency string, rule shipping.ShippingRule, value domainmoney.Money) int {
	feeCurrency := currency.NormalizeCode(rule.Currency)
	if feeCurrency == "" {
		feeCurrency = currency.NormalizeCode(templateCurrency)
	}
	if feeCurrency != value.Currency().String() {
		return 0
	}
	minMoney, err := rule.MinValueMoney(feeCurrency)
	additionalMoney, additionalErr := rule.AdditionalMoney(feeCurrency)
	if err != nil || additionalErr != nil || additionalMoney.AmountMinor() <= 0 || value.AmountMinor() <= minMoney.AmountMinor() {
		return 0
	}
	excess := value.AmountMinor() - minMoney.AmountMinor()
	return int((excess + additionalMoney.AmountMinor() - 1) / additionalMoney.AmountMinor())
}

func calculateRuleAdditionalUnitsForTemplate(
	templateType string,
	_ string,
	rule shipping.ShippingRule,
	value float64,
	weightBilling ...shippingWeightBilling,
) int {
	if templateType == "weight" {
		return calculateRuleAdditionalUnits(rule, value, weightBilling...)
	}
	additionalFee := 0.0
	if rule.AdditionalMinor > 0 {
		additionalFee = 1
	}
	return shippingrating.AdditionalWholeUnits(rule.MinValue, value, additionalFee)
}

func calculateRuleAdditionalUnits(rule shipping.ShippingRule, value float64, weightBilling ...shippingWeightBilling) int {
	if len(weightBilling) == 0 {
		return 0
	}
	billing := weightBilling[0]
	valueGrams := int(math.Round(value * 1000))
	additionalFee := 0.0
	if rule.AdditionalMinor > 0 {
		additionalFee = 1
	}
	units, err := shippingrating.AdditionalWeightUnits(valueGrams, additionalFee, shippingrating.WeightScale{
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
