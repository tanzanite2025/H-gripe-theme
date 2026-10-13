package service

import (
	"errors"
	"fmt"
	"sort"

	domainmoney "commerce-platform/internal/domain/money"
)

// quoteProductShippingTemplateAlternatives chooses exactly one billing template
// per cart item before grouping weights and prices. Shared templates may cover
// a subset of the cart when other items select another available template.
func (s *ShippingService) quoteProductShippingTemplateAlternatives(country, postalCode, quoteCurrency, displayCurrency string, resolvedItems []resolvedShippingItem) ([]ShippingQuotePlan, error) {
	alternatives := make(map[int][]resolvedShippingItem)
	for _, item := range resolvedItems {
		alternatives[item.SourceItemIndex] = append(alternatives[item.SourceItemIndex], item)
	}
	indexes := make([]int, 0, len(alternatives))
	for index := range alternatives {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)
	plans := make([]ShippingQuotePlan, 0)
	attempts := 0
	unavailable := ErrShippingQuotePlanUnavailable
	var visit func(int, []resolvedShippingItem) error
	visit = func(position int, selected []resolvedShippingItem) error {
		if position < len(indexes) {
			for _, item := range alternatives[indexes[position]] {
				if err := visit(position+1, append(selected, item)); err != nil {
					return err
				}
			}
			return nil
		}
		attempts++
		if attempts > maxShippingQuotePlans {
			return fmt.Errorf("%w: template combinations exceed %d", ErrShippingRateConfigurationInvalid, maxShippingQuotePlans)
		}
		groups := make(map[uint]*shippingQuoteGroup)
		for _, item := range selected {
			group := groups[item.Template.ID]
			if group == nil {
				amount, err := domainmoney.New(0, quoteCurrency)
				if err != nil {
					return err
				}
				group = &shippingQuoteGroup{Template: item.Template, Amount: amount}
				groups[item.Template.ID] = group
			}
			amount, err := group.Amount.Add(item.Amount)
			if err != nil {
				return err
			}
			group.Amount = amount
			group.Quantity += item.Quantity
			group.TotalWeightGrams += item.ChargeWeightGrams * item.Quantity
			group.ItemIndexes = append(group.ItemIndexes, item.SourceItemIndex)
		}
		rates := make(map[uint]shippingQuoteGroupRate, len(groups))
		for templateID, group := range groups {
			// Route-backed templates are priced after route and dimensional eligibility
			// are resolved. Standalone/system templates use the template fee directly.
			if group.Template.RouteGroupID != nil {
				continue
			}
			fee, free, displayPrices, err := s.calculateTemplateShippingFeeForQuoteMoney(group.Template, country, group.TotalWeightGrams, group.Quantity, group.Amount, group.Amount, quoteCurrency)
			if errors.Is(err, ErrCountryNotSupported) || errors.Is(err, ErrShippingTemplateRateNotApplicable) {
				unavailable = err
				return nil
			}
			if err != nil {
				return err
			}
			rates[templateID] = shippingQuoteGroupRate{FeeMinor: fee.AmountMinor(), FreeShipping: free, DisplayPrices: displayPrices}
		}
		choices, err := s.quoteShipmentPlans(country, postalCode, quoteCurrency, displayCurrency, selected, groups, rates)
		if errors.Is(err, ErrShippingQuotePlanUnavailable) || errors.Is(err, ErrCountryNotSupported) || errors.Is(err, ErrShippingTemplateRateNotApplicable) {
			unavailable = err
			return nil
		}
		if err != nil {
			return err
		}
		plans = append(plans, choices...)
		if len(plans) > maxShippingQuotePlans {
			return fmt.Errorf("%w: shipment plan combinations exceed %d", ErrShippingRateConfigurationInvalid, maxShippingQuotePlans)
		}
		return nil
	}
	if err := visit(0, nil); err != nil {
		return nil, err
	}
	if len(plans) == 0 {
		return nil, unavailable
	}
	sort.SliceStable(plans, func(i, j int) bool {
		if plans[i].ShippingFeeMinor != plans[j].ShippingFeeMinor {
			return plans[i].ShippingFeeMinor < plans[j].ShippingFeeMinor
		}
		if plans[i].EtaMaxDays != plans[j].EtaMaxDays {
			return plans[i].EtaMaxDays < plans[j].EtaMaxDays
		}
		return shippingQuotePlanSortKey(plans[i]) < shippingQuotePlanSortKey(plans[j])
	})
	return plans, nil
}
