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
	if template == nil || template.DefaultFee <= 0 {
		return nil
	}
	snapshots := currency.ParseDisplayPriceSnapshotMap(template.DisplayPriceData, shipping.ShippingTemplateDisplayPriceFields...)
	return roundDisplayPriceSnapshots(snapshots[shipping.ShippingTemplateDisplayPriceFieldDefaultFee])
}

func ruleFeeDisplayPrices(templateType, templateCurrency string, rule shipping.ShippingRule, value float64, weightBilling ...shippingWeightBilling) []currency.DisplayPriceSnapshot {
	snapshots := currency.ParseDisplayPriceSnapshotMap(rule.DisplayPriceData, shipping.ShippingRuleDisplayPriceFields...)
	additionalUnits := calculateRuleAdditionalUnitsForTemplate(templateType, templateCurrency, rule, value, weightBilling...)

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
		var amount domainmoney.Money
		amountInitialized := false
		var combined currency.DisplayPriceSnapshot
		initialized := false
		if needsFee {
			feeSnapshot, ok := feeByCurrency[code]
			if !ok {
				continue
			}
			feeMoney, err := domainmoney.FromMajorFloat(feeSnapshot.Amount, code)
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
			additionalMoney, err := domainmoney.FromMajorFloat(additionalSnapshot.Amount, code)
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
		combined.Amount, _ = amount.MajorFloat()
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
			value, err := domainmoney.FromMajorFloat(snapshot.Amount, code)
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
			total.Amount, _ = value.MajorFloat()
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
		snapshot.Amount, _ = value.MajorFloat()
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
	if freeShipping || baseFee <= 0 {
		return nil
	}

	baseRat, err := majorFloatRat(baseFee)
	if err != nil || baseRat.Sign() <= 0 {
		return nil
	}
	multiplier := big.NewRat(1, 1)
	if fuelSurcharge > 0 {
		fuelRat, fuelErr := majorFloatRat(fuelSurcharge)
		if fuelErr != nil {
			return nil
		}
		multiplier.Add(multiplier, new(big.Rat).Quo(fuelRat, baseRat))
	}
	remoteRatio := (*big.Rat)(nil)
	if remoteSurcharge > 0 {
		remoteRat, remoteErr := majorFloatRat(remoteSurcharge)
		if remoteErr == nil {
			remoteRatio = new(big.Rat).Quo(remoteRat, baseRat)
		}
	}

	result := make([]currency.DisplayPriceSnapshot, 0, len(baseDisplayPrices))
	for _, snapshot := range baseDisplayPrices {
		code := currency.NormalizeCode(snapshot.Currency)
		if code == "" {
			code = currency.NormalizeCode(snapshot.QuoteCurrency)
		}
		snapshotMoney, moneyErr := domainmoney.FromMajorFloat(snapshot.Amount, code)
		if moneyErr != nil {
			continue
		}
		amountMoney, moneyErr := snapshotMoney.MultiplyRat(multiplier)
		if moneyErr != nil {
			continue
		}
		if remoteRatio != nil && snapshot.Amount > 0 {
			remoteDisplay, remoteErr := snapshotMoney.MultiplyRat(remoteRatio)
			if remoteErr != nil {
				continue
			}
			amountMoney, remoteErr = amountMoney.Add(remoteDisplay)
			if remoteErr != nil {
				continue
			}
		} else if remoteSurcharge > 0 && snapshot.Rate > 0 {
			remoteRat, remoteErr := majorFloatRat(remoteSurcharge)
			if remoteErr != nil {
				continue
			}
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
		amount, amountErr := amountMoney.MajorFloat()
		if amountErr != nil || amount <= 0 || code == "" {
			continue
		}
		snapshot.Amount = amount
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
		amount := roundMoney(snapshot.Amount, code)
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
		amount := roundMoney(snapshot.Amount, code)
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

func roundMoney(value float64, currencyCode string) float64 {
	money, err := domainmoney.FromMajorFloat(value, currencyCode)
	if err != nil {
		return 0
	}
	amount, err := money.MajorFloat()
	if err != nil {
		return 0
	}
	return amount
}
