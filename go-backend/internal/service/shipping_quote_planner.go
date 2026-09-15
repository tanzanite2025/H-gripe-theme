package service

import (
	"fmt"
	"strings"

	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
)

const maxShippingQuotePlans = 128

type shippingQuoteGroupRate struct {
	Fee           float64
	FreeShipping  bool
	DisplayPrices []currency.DisplayPriceSnapshot
}

func shippingQuoteGroupKey(templateID uint) string {
	return fmt.Sprintf("template:%d", templateID)
}

func finalizeShippingQuotePlan(plan ShippingQuotePlan, displayCurrency string) (ShippingQuotePlan, error) {
	planCurrency := currency.NormalizeCode(plan.Currency)
	totalFee, err := domainmoney.New(0, planCurrency)
	if err != nil {
		return ShippingQuotePlan{}, fmt.Errorf("plan currency: %w", err)
	}
	plan.EtaMinDays = 0
	plan.EtaMaxDays = 0
	displayPriceSets := make([][]currency.DisplayPriceSnapshot, 0, len(plan.Legs))
	for _, leg := range plan.Legs {
		legFee, feeErr := domainmoney.FromMajorFloat(leg.ShippingFee, planCurrency)
		if feeErr != nil {
			return ShippingQuotePlan{}, fmt.Errorf("leg shipping fee: %w", feeErr)
		}
		totalFee, feeErr = totalFee.Add(legFee)
		if feeErr != nil {
			return ShippingQuotePlan{}, fmt.Errorf("sum plan shipping fees: %w", feeErr)
		}
		if legFee.AmountMinor() > 0 {
			displayPriceSets = append(displayPriceSets, leg.DisplayPrices)
		}
		if leg.EtaMinDays > plan.EtaMinDays {
			plan.EtaMinDays = leg.EtaMinDays
		}
		if leg.EtaMaxDays > plan.EtaMaxDays {
			plan.EtaMaxDays = leg.EtaMaxDays
		}
	}
	plan.ShippingFee, err = totalFee.MajorFloat()
	if err != nil {
		return ShippingQuotePlan{}, fmt.Errorf("format plan shipping fee: %w", err)
	}
	plan.FreeShipping = totalFee.AmountMinor() <= 0
	plan.DisplayPrices = combineDisplayPriceSets(displayPriceSets)
	plan.DisplayPrice = displayPriceForCurrency(displayCurrency, plan.DisplayPrices)
	return plan, nil
}

func shippingQuotePlanSortKey(plan ShippingQuotePlan) string {
	parts := make([]string, 0, len(plan.Legs))
	for _, leg := range plan.Legs {
		parts = append(parts, fmt.Sprintf("%010d:%010d", leg.TemplateID, leg.CarrierServiceID))
	}
	return strings.Join(parts, "|")
}

func applyShippingQuotePlan(quote *ShippingQuote, plan *ShippingQuotePlan) {
	if quote == nil || plan == nil {
		return
	}
	for index := range quote.Items {
		quote.Items[index].ShippingFee = 0
		quote.Items[index].FreeShipping = false
	}
	for _, leg := range plan.Legs {
		distributeShippingQuoteItemFee(
			leg.ItemIndexes,
			quote.Items,
			leg.ShippingFee,
			leg.FreeShipping,
			leg.AllocationBasis,
			quote.Currency,
		)
	}
	selected := *plan
	quote.SelectedPlan = &selected
	quote.ShippingFee = plan.ShippingFee
	quote.FreeShipping = plan.FreeShipping
	quote.DisplayPrice = plan.DisplayPrice
	quote.DisplayPrices = append([]currency.DisplayPriceSnapshot(nil), plan.DisplayPrices...)
}

func distributeShippingQuoteItemFee(
	itemIndexes []int,
	items []ShippingQuoteItem,
	fee float64,
	free bool,
	basis string,
	currencyCode string,
) {
	if len(itemIndexes) == 0 {
		return
	}
	feeMoney, err := domainmoney.FromMajorFloat(fee, currencyCode)
	if err != nil || feeMoney.AmountMinor() < 0 {
		return
	}
	validIndexes := make([]int, 0, len(itemIndexes))
	ratios := make([]int64, 0, len(itemIndexes))
	totalBasis := int64(0)
	for _, index := range itemIndexes {
		if index >= 0 && index < len(items) {
			itemBasis, basisErr := shippingQuoteItemDistributionBasis(basis, items[index], currencyCode)
			if basisErr != nil || itemBasis < 0 {
				return
			}
			validIndexes = append(validIndexes, index)
			ratios = append(ratios, itemBasis)
			totalBasis += itemBasis
		}
	}
	if len(validIndexes) == 0 {
		return
	}
	if totalBasis == 0 {
		for i := range ratios {
			ratios[i] = 1
		}
	}
	allocations, err := feeMoney.Allocate(ratios)
	if err != nil {
		return
	}
	for i, index := range validIndexes {
		itemFee, feeErr := allocations[i].MajorFloat()
		if feeErr != nil {
			return
		}
		items[index].ShippingFee = itemFee
		items[index].FreeShipping = free
	}
}

func shippingQuoteItemDistributionBasis(templateType string, item ShippingQuoteItem, currencyCode string) (int64, error) {
	switch templateType {
	case "quantity", "items":
		return int64(item.Quantity), nil
	case "price", "amount":
		amount, err := domainmoney.FromMajorFloat(item.Amount, currencyCode)
		if err != nil {
			return 0, err
		}
		return amount.AmountMinor(), nil
	default:
		return int64(item.ChargeWeightGrams * item.Quantity), nil
	}
}
