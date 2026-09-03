package service

import (
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/shipping"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)

func (s *ShippingService) quoteCarrierServiceOptions(
	country string,
	currency string,
	displayCurrency string,
	resolvedItems []resolvedShippingItem,
	groups map[uint]*shippingQuoteGroup,
	cartAmount float64,
) ([]ShippingQuoteOption, error) {
	group := singleShippingQuoteGroup(groups)
	if group == nil {
		return nil, nil
	}

	carrierServices, err := s.shippingRepo.FindEnabledCarrierServicesWithTemplates()
	if err != nil {
		return nil, err
	}

	options := make([]ShippingQuoteOption, 0, len(carrierServices))
	for i := range carrierServices {
		option, ok := buildCarrierServiceQuoteOption(carrierServices[i], group, resolvedItems, country, currency, displayCurrency, cartAmount)
		if ok {
			options = append(options, option)
		}
	}

	sort.SliceStable(options, func(i, j int) bool {
		if options[i].ShippingFee != options[j].ShippingFee {
			return options[i].ShippingFee < options[j].ShippingFee
		}
		if options[i].SortOrder != options[j].SortOrder {
			return options[i].SortOrder < options[j].SortOrder
		}
		return options[i].CarrierServiceID < options[j].CarrierServiceID
	})

	return options, nil
}

func buildCarrierServiceQuoteOption(
	service shipping.CarrierService,
	group *shippingQuoteGroup,
	resolvedItems []resolvedShippingItem,
	country string,
	currency string,
	displayCurrency string,
	cartAmount float64,
) (ShippingQuoteOption, bool) {
	if group == nil || service.Template == nil || !service.Enabled || !service.Template.Enabled {
		return ShippingQuoteOption{}, false
	}
	if service.TemplateID == nil || *service.TemplateID != group.Template.ID {
		return ShippingQuoteOption{}, false
	}
	if service.Carrier == nil || !service.Carrier.Enabled {
		return ShippingQuoteOption{}, false
	}
	if !carrierServiceMatchesCountry(service.Countries, country) {
		return ShippingQuoteOption{}, false
	}

	quoteCurrency := strings.ToUpper(strings.TrimSpace(currency))
	serviceCurrency := strings.ToUpper(strings.TrimSpace(service.Currency))
	if serviceCurrency == "" {
		serviceCurrency = quoteCurrency
	}
	if quoteCurrency != "" && serviceCurrency != "" && serviceCurrency != quoteCurrency {
		return ShippingQuoteOption{}, false
	}

	actualWeightGrams := group.TotalWeightGrams
	volumetricWeightGrams, hasVolumetricWeight := carrierServiceVolumetricWeightGrams(service, resolvedItems)
	chargeWeightGrams := actualWeightGrams
	switch service.BillingMode {
	case "volumetric_weight":
		if !hasVolumetricWeight {
			return ShippingQuoteOption{}, false
		}
		chargeWeightGrams = volumetricWeightGrams
	case "greater_of_actual_and_volumetric":
		if !hasVolumetricWeight {
			chargeWeightGrams = actualWeightGrams
			break
		}
		chargeWeightGrams = maxInt(actualWeightGrams, volumetricWeightGrams)
	}

	billableWeightGrams := carrierServiceBillableWeightGrams(chargeWeightGrams, service)
	baseFee, freeShipping, baseDisplayPrices := calculateTemplateShippingFeeWithDisplayPrices(
		service.Template,
		country,
		billableWeightGrams,
		group.Quantity,
		group.Amount,
		cartAmount,
	)

	fuelSurcharge := 0.0
	remoteSurcharge := 0.0
	var shippingFee float64
	if freeShipping {
		shippingFee = 0
	} else {
		fuelSurcharge = roundMoney(baseFee * service.FuelSurchargePercent / 100)
		remoteSurcharge = roundMoney(service.RemoteSurcharge)
		shippingFee = roundMoney(baseFee + fuelSurcharge + remoteSurcharge)
	}
	displayPrices := deriveCarrierServiceDisplayPrices(baseDisplayPrices, baseFee, fuelSurcharge, remoteSurcharge, freeShipping)

	return ShippingQuoteOption{
		CarrierID:             service.Carrier.ID,
		CarrierName:           service.Carrier.Name,
		CarrierCode:           service.Carrier.Code,
		CarrierServiceID:      service.ID,
		ServiceCode:           service.ServiceCode,
		ServiceName:           service.ServiceName,
		RouteName:             service.RouteName,
		TemplateID:            service.Template.ID,
		TemplateName:          service.Template.Name,
		Currency:              serviceCurrency,
		BillingMode:           service.BillingMode,
		ActualWeightGrams:     actualWeightGrams,
		VolumetricWeightGrams: volumetricWeightGrams,
		ChargeWeightGrams:     chargeWeightGrams,
		BillableWeightGrams:   billableWeightGrams,
		BaseFee:               baseFee,
		FuelSurcharge:         fuelSurcharge,
		RemoteSurcharge:       remoteSurcharge,
		ShippingFee:           shippingFee,
		DisplayPrice:          displayPriceForCurrency(displayCurrency, displayPrices),
		DisplayPrices:         displayPrices,
		FreeShipping:          freeShipping,
		EtaMinDays:            service.EtaMinDays,
		EtaMaxDays:            service.EtaMaxDays,
		SortOrder:             service.SortOrder,
	}, true
}

func singleShippingQuoteGroup(groups map[uint]*shippingQuoteGroup) *shippingQuoteGroup {
	if len(groups) != 1 {
		return nil
	}
	for _, group := range groups {
		return group
	}
	return nil
}

func carrierServiceMatchesCountry(countriesValue string, country string) bool {
	normalizedCountry := strings.ToUpper(strings.TrimSpace(country))
	if normalizedCountry == "" {
		return false
	}

	countries := normalizeShippingRegions(countriesValue)
	if len(countries) == 0 {
		return true
	}

	for _, candidate := range countries {
		switch candidate {
		case "*", "ALL", "GLOBAL", "WORLDWIDE":
			return true
		default:
			if candidate == normalizedCountry {
				return true
			}
		}
	}
	return false
}

func carrierServiceVolumetricWeightGrams(service shipping.CarrierService, resolvedItems []resolvedShippingItem) (int, bool) {
	if service.VolumetricDivisor <= 0 {
		return 0, false
	}

	total := 0
	for _, item := range resolvedItems {
		itemVolumetricWeight, ok := packagingRuleVolumetricWeightGrams(item.PackagingRule, service.VolumetricDivisor)
		if !ok {
			return 0, false
		}
		total += itemVolumetricWeight * item.Quantity
	}
	return total, total > 0
}

func packagingRuleVolumetricWeightGrams(rule *shipping.PackagingRule, divisor int) (int, bool) {
	if rule == nil || divisor <= 0 || rule.BoxLength <= 0 || rule.BoxWidth <= 0 || rule.BoxHeight <= 0 {
		return 0, false
	}
	volumetricKg := rule.BoxLength * rule.BoxWidth * rule.BoxHeight / float64(divisor)
	if volumetricKg <= 0 {
		return 0, false
	}
	return int(math.Ceil(volumetricKg * 1000)), true
}

func carrierServiceBillableWeightGrams(chargeWeightGrams int, service shipping.CarrierService) int {
	billableWeightGrams := maxInt(chargeWeightGrams, service.MinChargeWeightGrams)
	if billableWeightGrams <= 0 {
		return 0
	}

	firstWeightGrams := service.FirstWeightGrams
	additionalWeightGrams := service.AdditionalWeightGrams
	if firstWeightGrams > 0 {
		if billableWeightGrams <= firstWeightGrams {
			return firstWeightGrams
		}
		if additionalWeightGrams > 0 {
			return firstWeightGrams + ceilDivInt(billableWeightGrams-firstWeightGrams, additionalWeightGrams)*additionalWeightGrams
		}
		return billableWeightGrams
	}
	if additionalWeightGrams > 0 {
		return ceilDivInt(billableWeightGrams, additionalWeightGrams) * additionalWeightGrams
	}
	return billableWeightGrams
}

func ceilDivInt(value int, divisor int) int {
	if value <= 0 {
		return 0
	}
	return (value + divisor - 1) / divisor
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
) (float64, bool, []currency.DisplayPriceSnapshot) {
	if template == nil {
		return 0, false, nil
	}

	if template.FreeShipping && cartAmount >= template.FreeThreshold {
		return 0, true, nil
	}

	value := float64(totalWeightGrams) / 1000
	switch template.Type {
	case "quantity", "items":
		value = float64(quantity)
	case "price", "amount":
		value = amount
	}

	shippingFee := template.DefaultFee
	displayPrices := templateFeeDisplayPrices(template)
	for _, rule := range template.Rules {
		if shippingRuleMatchesCountry(rule.Region, country) && shippingRuleMatchesValue(rule, value) {
			shippingFee = calculateRuleFee(rule, value)
			displayPrices = ruleFeeDisplayPrices(rule, value)
			break
		}
	}

	shippingFee = roundMoney(shippingFee)
	if shippingFee <= 0 {
		return shippingFee, false, nil
	}
	return shippingFee, false, displayPrices
}

func shippingRuleMatchesValue(rule shipping.ShippingRule, value float64) bool {
	return value >= rule.MinValue && (rule.MaxValue == 0 || value <= rule.MaxValue)
}

func calculateRuleFee(rule shipping.ShippingRule, value float64) float64 {
	fee := rule.Fee
	additionalUnits := calculateRuleAdditionalUnits(rule, value)
	if additionalUnits > 0 {
		fee += float64(additionalUnits) * rule.Additional
	}
	return fee
}

func calculateRuleAdditionalUnits(rule shipping.ShippingRule, value float64) int {
	if rule.Additional <= 0 || value <= rule.MinValue {
		return 0
	}

	additionalUnits := int(math.Ceil(value-rule.MinValue)) - 1
	if additionalUnits < 0 {
		return 0
	}
	return additionalUnits
}

func templateFeeDisplayPrices(template *shipping.ShippingTemplate) []currency.DisplayPriceSnapshot {
	if template == nil || template.DefaultFee <= 0 {
		return nil
	}
	snapshots := currency.ParseDisplayPriceSnapshotMap(template.DisplayPriceData, shipping.ShippingTemplateDisplayPriceFields...)
	return roundDisplayPriceSnapshots(snapshots[shipping.ShippingTemplateDisplayPriceFieldDefaultFee])
}

func ruleFeeDisplayPrices(rule shipping.ShippingRule, value float64) []currency.DisplayPriceSnapshot {
	snapshots := currency.ParseDisplayPriceSnapshotMap(rule.DisplayPriceData, shipping.ShippingRuleDisplayPriceFields...)
	additionalUnits := calculateRuleAdditionalUnits(rule, value)

	needsFee := rule.Fee > 0
	needsAdditional := additionalUnits > 0 && rule.Additional > 0
	if !needsFee && !needsAdditional {
		return nil
	}

	feeByCurrency := displayPriceSnapshotsByCurrency(snapshots[shipping.ShippingRuleDisplayPriceFieldFee])
	additionalByCurrency := displayPriceSnapshotsByCurrency(snapshots[shipping.ShippingRuleDisplayPriceFieldAdditional])
	candidateCurrencies := map[string]struct{}{}
	if needsFee {
		for code := range feeByCurrency {
			candidateCurrencies[code] = struct{}{}
		}
	}
	if needsAdditional {
		for code := range additionalByCurrency {
			candidateCurrencies[code] = struct{}{}
		}
	}

	codes := sortedDisplayPriceCurrencyCodes(candidateCurrencies)
	result := make([]currency.DisplayPriceSnapshot, 0, len(codes))
	for _, code := range codes {
		amount := 0.0
		var combined currency.DisplayPriceSnapshot
		initialized := false
		if needsFee {
			feeSnapshot, ok := feeByCurrency[code]
			if !ok {
				continue
			}
			amount += feeSnapshot.Amount
			combined = mergeDisplayPriceSnapshotMetadata(combined, feeSnapshot, initialized)
			initialized = true
		}
		if needsAdditional {
			additionalSnapshot, ok := additionalByCurrency[code]
			if !ok {
				continue
			}
			amount += float64(additionalUnits) * additionalSnapshot.Amount
			combined = mergeDisplayPriceSnapshotMetadata(combined, additionalSnapshot, initialized)
			initialized = true
		}
		if amount <= 0 || !initialized {
			continue
		}
		combined.Amount = roundMoney(amount)
		combined.Currency = code
		combined.QuoteCurrency = code
		result = append(result, combined)
	}
	return result
}

func combineDisplayPriceSets(sets [][]currency.DisplayPriceSnapshot) []currency.DisplayPriceSnapshot {
	if len(sets) == 0 {
		return nil
	}

	totals := map[string]currency.DisplayPriceSnapshot{}
	counts := map[string]int{}
	for _, set := range sets {
		for code, snapshot := range displayPriceSnapshotsByCurrency(set) {
			total, exists := totals[code]
			if exists {
				total = mergeDisplayPriceSnapshotMetadata(total, snapshot, true)
			} else {
				total = snapshot
				total.Amount = 0
			}
			total.Amount += snapshot.Amount
			total.Currency = code
			total.QuoteCurrency = code
			totals[code] = total
			counts[code]++
		}
	}

	candidates := map[string]struct{}{}
	for code := range totals {
		candidates[code] = struct{}{}
	}

	codes := sortedDisplayPriceCurrencyCodes(candidates)
	result := make([]currency.DisplayPriceSnapshot, 0, len(codes))
	for _, code := range codes {
		if counts[code] != len(sets) {
			continue
		}
		snapshot := totals[code]
		snapshot.Amount = roundMoney(snapshot.Amount)
		if snapshot.Amount > 0 {
			result = append(result, snapshot)
		}
	}
	return result
}

func deriveCarrierServiceDisplayPrices(
	baseDisplayPrices []currency.DisplayPriceSnapshot,
	baseFee float64,
	fuelSurcharge float64,
	remoteSurcharge float64,
	freeShipping bool,
) []currency.DisplayPriceSnapshot {
	if freeShipping || baseFee <= 0 || remoteSurcharge > 0 {
		return nil
	}

	multiplier := 1.0
	if fuelSurcharge > 0 {
		multiplier += fuelSurcharge / baseFee
	}

	result := make([]currency.DisplayPriceSnapshot, 0, len(baseDisplayPrices))
	for _, snapshot := range baseDisplayPrices {
		code := currency.NormalizeCode(snapshot.Currency)
		if code == "" {
			code = currency.NormalizeCode(snapshot.QuoteCurrency)
		}
		amount := roundMoney(snapshot.Amount * multiplier)
		if amount <= 0 || code == "" {
			continue
		}
		snapshot.Amount = amount
		snapshot.Currency = code
		snapshot.QuoteCurrency = code
		result = append(result, snapshot)
	}
	return result
}

func displayPriceForCurrency(displayCurrency string, displayPrices []currency.DisplayPriceSnapshot) *currency.DisplayPriceSnapshot {
	displayCurrency = currency.NormalizeCode(displayCurrency)
	if displayCurrency == "" {
		return nil
	}
	for i := range displayPrices {
		if currency.NormalizeCode(displayPrices[i].Currency) == displayCurrency {
			return &displayPrices[i]
		}
	}
	return nil
}

func roundDisplayPriceSnapshots(snapshots []currency.DisplayPriceSnapshot) []currency.DisplayPriceSnapshot {
	if len(snapshots) == 0 {
		return nil
	}
	result := make([]currency.DisplayPriceSnapshot, 0, len(snapshots))
	for _, snapshot := range snapshots {
		code := currency.NormalizeCode(snapshot.Currency)
		if code == "" {
			code = currency.NormalizeCode(snapshot.QuoteCurrency)
		}
		amount := roundMoney(snapshot.Amount)
		if amount <= 0 || code == "" {
			continue
		}
		snapshot.Amount = amount
		snapshot.Currency = code
		snapshot.QuoteCurrency = code
		result = append(result, snapshot)
	}
	return result
}

func displayPriceSnapshotsByCurrency(snapshots []currency.DisplayPriceSnapshot) map[string]currency.DisplayPriceSnapshot {
	result := make(map[string]currency.DisplayPriceSnapshot, len(snapshots))
	for _, snapshot := range snapshots {
		code := currency.NormalizeCode(snapshot.Currency)
		if code == "" {
			code = currency.NormalizeCode(snapshot.QuoteCurrency)
		}
		amount := roundMoney(snapshot.Amount)
		if amount <= 0 || code == "" {
			continue
		}
		snapshot.Amount = amount
		snapshot.Currency = code
		snapshot.QuoteCurrency = code
		if _, exists := result[code]; !exists {
			result[code] = snapshot
		}
	}
	return result
}

func mergeDisplayPriceSnapshotMetadata(
	current currency.DisplayPriceSnapshot,
	next currency.DisplayPriceSnapshot,
	initialized bool,
) currency.DisplayPriceSnapshot {
	if !initialized {
		return next
	}
	if current.Source != next.Source {
		current.Source = "stored_display_snapshot"
	}
	if current.Rate != next.Rate {
		current.Rate = 0
	}
	current.Converted = current.Converted || next.Converted
	return current
}

func sortedDisplayPriceCurrencyCodes(values map[string]struct{}) []string {
	codes := make([]string, 0, len(values))
	for code := range values {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}

var shippingEUCountryCodes = map[string]struct{}{
	"AT": {},
	"BE": {},
	"BG": {},
	"HR": {},
	"CY": {},
	"CZ": {},
	"DE": {},
	"DK": {},
	"EE": {},
	"ES": {},
	"FI": {},
	"FR": {},
	"GR": {},
	"HU": {},
	"IE": {},
	"IT": {},
	"LT": {},
	"LU": {},
	"LV": {},
	"MT": {},
	"NL": {},
	"PL": {},
	"PT": {},
	"RO": {},
	"SE": {},
	"SI": {},
	"SK": {},
}

var shippingEEACountryCodes = map[string]struct{}{
	"AT": {},
	"BE": {},
	"BG": {},
	"CY": {},
	"CZ": {},
	"DE": {},
	"DK": {},
	"EE": {},
	"ES": {},
	"FI": {},
	"FR": {},
	"GR": {},
	"HR": {},
	"HU": {},
	"IS": {},
	"IE": {},
	"IT": {},
	"LI": {},
	"LT": {},
	"LU": {},
	"LV": {},
	"MT": {},
	"NL": {},
	"NO": {},
	"PL": {},
	"PT": {},
	"RO": {},
	"SE": {},
	"SI": {},
	"SK": {},
}

// shippingRegionCountryCodes maps supported shipping region macros to ISO alpha-2 country codes.
var shippingRegionCountryCodes = map[string]map[string]struct{}{
	"EU":                     shippingEUCountryCodes,
	"EU27":                   shippingEUCountryCodes,
	"EUROPEAN_UNION":         shippingEUCountryCodes,
	"EUROPEAN UNION":         shippingEUCountryCodes,
	"EEA":                    shippingEEACountryCodes,
	"EUROPEAN_ECONOMIC_AREA": shippingEEACountryCodes,
	"EUROPEAN ECONOMIC AREA": shippingEEACountryCodes,
}

func shippingRuleMatchesCountry(region string, country string) bool {
	normalizedCountry := strings.ToUpper(strings.TrimSpace(country))
	if normalizedCountry == "" {
		return false
	}

	regions := normalizeShippingRegions(region)
	if len(regions) == 0 {
		return true
	}

	for _, candidate := range regions {
		switch candidate {
		case "*", "ALL", "GLOBAL", "WORLDWIDE":
			return true
		default:
			if candidate == normalizedCountry || shippingRegionMatchesCountry(candidate, normalizedCountry) {
				return true
			}
		}
	}
	return false
}

func shippingRegionMatchesCountry(region string, country string) bool {
	countries, ok := shippingRegionCountryCodes[region]
	if !ok {
		return false
	}
	_, ok = countries[country]
	return ok
}

func normalizeShippingRegions(region string) []string {
	trimmed := strings.TrimSpace(region)
	if trimmed == "" {
		return nil
	}

	var parsed []string
	if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
		return normalizeRegionList(parsed)
	}

	parts := strings.FieldsFunc(trimmed, func(r rune) bool {
		return r == ',' || r == ';' || r == '|' || r == '\n' || r == '\r' || r == '\t'
	})
	return normalizeRegionList(parts)
}

func normalizeRegionList(regions []string) []string {
	normalized := make([]string, 0, len(regions))
	for _, region := range regions {
		candidate := strings.ToUpper(strings.TrimSpace(region))
		if candidate != "" {
			normalized = append(normalized, candidate)
		}
	}
	return normalized
}

func distributeGroupFee(
	group *shippingQuoteGroup,
	resolvedItems []resolvedShippingItem,
	quoteItems []ShippingQuoteItem,
	groupFee float64,
	groupFree bool,
) {
	if group == nil || len(group.ItemIndexes) == 0 {
		return
	}

	var totalBasis float64
	for _, index := range group.ItemIndexes {
		totalBasis += shippingFeeDistributionBasis(group.Template.Type, resolvedItems[index])
	}

	remaining := roundMoney(groupFee)
	for position, index := range group.ItemIndexes {
		itemFee := 0.0
		if position == len(group.ItemIndexes)-1 {
			itemFee = remaining
		} else if totalBasis > 0 {
			itemFee = roundMoney(groupFee * shippingFeeDistributionBasis(group.Template.Type, resolvedItems[index]) / totalBasis)
			remaining = roundMoney(remaining - itemFee)
		} else {
			itemFee = roundMoney(groupFee / float64(len(group.ItemIndexes)))
			remaining = roundMoney(remaining - itemFee)
		}

		quoteItems[index].ShippingFee = itemFee
		quoteItems[index].FreeShipping = groupFree
	}
}

func shippingFeeDistributionBasis(templateType string, item resolvedShippingItem) float64 {
	switch templateType {
	case "quantity", "items":
		return float64(item.Quantity)
	case "price", "amount":
		return item.Amount
	default:
		return float64(item.ChargeWeightGrams * item.Quantity)
	}
}

func roundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}

func hasPositiveID(id *uint) bool {
	return id != nil && *id > 0
}

func uintPtr(id uint) *uint {
	return &id
}
