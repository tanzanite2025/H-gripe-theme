package service

import (
	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/shipping"
	"fmt"
	"math/big"
	"sort"
	"strconv"
)

func templateFeeDisplayPrices(template *shipping.ShippingTemplate) []currency.DisplayPriceSnapshot {
	if template == nil {
		return nil
	}
	fee, err := template.DefaultFeeMoney()
	if err != nil || fee.AmountMinor() <= 0 {
		return nil
	}
	snapshots := currency.ParseDisplayPriceSnapshotMap(template.DisplayPriceData, shipping.ShippingTemplateDisplayPriceFields...)
	return roundDisplayPriceSnapshots(snapshots[shipping.ShippingTemplateDisplayPriceFieldDefaultFee])
}

func ruleFeeDisplayPrices(templateType, templateCurrency string, rule shipping.ShippingRule, value float64, amount *domainmoney.Money, weightBilling ...shippingWeightBilling) []currency.DisplayPriceSnapshot {
	snapshots := currency.ParseDisplayPriceSnapshotMap(rule.DisplayPriceData, shipping.ShippingRuleDisplayPriceFields...)
	additionalUnits := 0
	if templateType == "price" || templateType == "amount" {
		if amount != nil {
			additionalUnits = calculateRuleAdditionalUnitsForMoneyTemplate(templateCurrency, rule, *amount)
		}
	} else {
		additionalUnits = calculateRuleAdditionalUnitsForTemplate(templateType, templateCurrency, rule, value, weightBilling...)
	}

	feeMoney, feeErr := rule.FeeMoney(templateCurrency)
	additionalMoney, additionalErr := rule.AdditionalMoney(templateCurrency)
	needsFee := feeErr == nil && feeMoney.AmountMinor() > 0
	needsAdditional := additionalUnits > 0 && additionalErr == nil && additionalMoney.AmountMinor() > 0
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
		var amount domainmoney.Money
		amountInitialized := false
		var combined currency.DisplayPriceSnapshot
		initialized := false
		if needsFee {
			feeSnapshot, ok := feeByCurrency[code]
			if !ok {
				continue
			}
			feeMoney, err := domainmoney.ParseMajor(feeSnapshot.AmountDecimal, code)
			if err != nil {
				continue
			}
			amount = feeMoney
			amountInitialized = true
			combined = mergeDisplayPriceSnapshotMetadata(combined, feeSnapshot, initialized)
			initialized = true
		}
		if needsAdditional {
			additionalSnapshot, ok := additionalByCurrency[code]
			if !ok {
				continue
			}
			additionalMoney, err := domainmoney.ParseMajor(additionalSnapshot.AmountDecimal, code)
			if err != nil {
				continue
			}
			additionalMoney, err = additionalMoney.MultiplyInt(int64(additionalUnits))
			if err != nil {
				continue
			}
			if !amountInitialized {
				amount = additionalMoney
				amountInitialized = true
			} else {
				amount, err = amount.Add(additionalMoney)
				if err != nil {
					continue
				}
			}
			combined = mergeDisplayPriceSnapshotMetadata(combined, additionalSnapshot, initialized)
			initialized = true
		}
		if !amountInitialized || amount.AmountMinor() <= 0 || !initialized {
			continue
		}
		combined.AmountDecimal, _ = amount.FormatMajor()
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
	amounts := map[string]domainmoney.Money{}
	counts := map[string]int{}
	for _, set := range sets {
		for code, snapshot := range displayPriceSnapshotsByCurrency(set) {
			total, exists := totals[code]
			value, err := domainmoney.ParseMajor(snapshot.AmountDecimal, code)
			if err != nil {
				continue
			}
			if exists {
				total = mergeDisplayPriceSnapshotMetadata(total, snapshot, true)
				value, err = amounts[code].Add(value)
				if err != nil {
					continue
				}
			} else {
				total = snapshot
			}
			amounts[code] = value
			total.AmountDecimal, _ = value.FormatMajor()
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
		value, ok := amounts[code]
		if !ok || value.AmountMinor() <= 0 {
			continue
		}
		snapshot.AmountDecimal, _ = value.FormatMajor()
		if value.AmountMinor() > 0 {
			result = append(result, snapshot)
		}
	}
	return result
}

func deriveCarrierServiceDisplayPrices(
	baseDisplayPrices []currency.DisplayPriceSnapshot,
	baseFee domainmoney.Money,
	fuelSurcharge domainmoney.Money,
	remoteSurcharge domainmoney.Money,
	freeShipping bool,
) []currency.DisplayPriceSnapshot {
	if freeShipping || baseFee.AmountMinor() <= 0 {
		return nil
	}

	baseRat := big.NewRat(baseFee.AmountMinor(), 1)
	multiplier := big.NewRat(1, 1)
	if fuelSurcharge.AmountMinor() > 0 {
		fuelRat := big.NewRat(fuelSurcharge.AmountMinor(), 1)
		multiplier.Add(multiplier, new(big.Rat).Quo(fuelRat, baseRat))
	}
	remoteRatio := (*big.Rat)(nil)
	if remoteSurcharge.AmountMinor() > 0 {
		remoteRat := big.NewRat(remoteSurcharge.AmountMinor(), 1)
		remoteRatio = new(big.Rat).Quo(remoteRat, baseRat)
	}

	result := make([]currency.DisplayPriceSnapshot, 0, len(baseDisplayPrices))
	for _, snapshot := range baseDisplayPrices {
		code := currency.NormalizeCode(snapshot.Currency)
		if code == "" {
			code = currency.NormalizeCode(snapshot.QuoteCurrency)
		}
		snapshotMoney, moneyErr := domainmoney.ParseMajor(snapshot.AmountDecimal, code)
		if moneyErr != nil {
			continue
		}
		amountMoney, moneyErr := snapshotMoney.MultiplyRat(multiplier)
		if moneyErr != nil {
			continue
		}
		if remoteRatio != nil && snapshotMoney.AmountMinor() > 0 {
			remoteDisplay, remoteErr := snapshotMoney.MultiplyRat(remoteRatio)
			if remoteErr != nil {
				continue
			}
			amountMoney, remoteErr = amountMoney.Add(remoteDisplay)
			if remoteErr != nil {
				continue
			}
		} else if remoteSurcharge.AmountMinor() > 0 && snapshot.Rate > 0 {
			remoteRat := big.NewRat(remoteSurcharge.AmountMinor(), 1)
			rateRat, rateErr := majorFloatRat(snapshot.Rate)
			if rateErr != nil {
				continue
			}
			remoteRat.Mul(remoteRat, rateRat)
			remoteMoney, remoteErr := domainmoney.FromMajorRat(remoteRat, code)
			if remoteErr != nil {
				continue
			}
			amountMoney, remoteErr = amountMoney.Add(remoteMoney)
			if remoteErr != nil {
				continue
			}
		}
		amountDecimal, amountErr := amountMoney.FormatMajor()
		if amountErr != nil || amountMoney.AmountMinor() <= 0 || code == "" {
			continue
		}
		snapshot.AmountDecimal = amountDecimal
		snapshot.Currency = code
		snapshot.QuoteCurrency = code
		result = append(result, snapshot)
	}
	return result
}

func majorFloatRat(value float64) (*big.Rat, error) {
	rat, ok := new(big.Rat).SetString(strconv.FormatFloat(value, 'f', -1, 64))
	if !ok {
		return nil, fmt.Errorf("invalid decimal amount %v", value)
	}
	return rat, nil
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
		amount, err := domainmoney.ParseMajor(snapshot.AmountDecimal, code)
		if err != nil || amount.AmountMinor() <= 0 || code == "" {
			continue
		}
		snapshot.AmountDecimal, _ = amount.FormatMajor()
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
		amount, err := domainmoney.ParseMajor(snapshot.AmountDecimal, code)
		if err != nil || amount.AmountMinor() <= 0 || code == "" {
			continue
		}
		snapshot.AmountDecimal, _ = amount.FormatMajor()
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
