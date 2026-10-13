package service

import (
	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/shipping"
	shippingrating "commerce-platform/internal/domain/shipping/rating"
	"fmt"
	"math/big"
	"sort"
	"strings"
)

func (s *ShippingService) quoteShipmentPlans(
	country string,
	postalCode string,
	quoteCurrency string,
	displayCurrency string,
	resolvedItems []resolvedShippingItem,
	groups map[uint]*shippingQuoteGroup,
	groupRates map[uint]shippingQuoteGroupRate,
) ([]ShippingQuotePlan, error) {
	if len(groups) == 0 {
		return nil, nil
	}

	routeGroupIDs := make([]uint, 0, len(groups))
	for _, group := range groups {
		if group != nil && group.Template != nil && group.Template.RouteGroupID != nil {
			routeGroupIDs = append(routeGroupIDs, *group.Template.RouteGroupID)
		}
	}
	carrierServices, err := s.shippingRepo.FindEnabledCarrierServicesForRouteAggregates(routeGroupIDs)
	if err != nil {
		return nil, err
	}
	carrierServices, aggregatesWithPublishedCollectionRoutes, err := s.projectCarrierServicesFromPublishedCollections(carrierServices)
	if err != nil {
		return nil, err
	}

	legsByTemplate := make(map[uint][]ShippingQuoteLeg, len(groups))
	for i := range carrierServices {
		carrierService := carrierServices[i]
		for templateID, group := range groups {
			if group == nil || group.Template == nil {
				continue
			}
			if group.Template.RouteGroupID != nil {
				if carrierService.RouteGroupID == nil || *carrierService.RouteGroupID != *group.Template.RouteGroupID {
					continue
				}
			} else {
				continue
			}
			leg, ok, err := s.buildCarrierServiceQuoteLeg(
				carrierService,
				group.Template,
				group,
				shippingQuoteGroupItems(group, resolvedItems),
				country,
				postalCode,
				quoteCurrency,
				displayCurrency,
				group.Amount,
			)
			if err != nil {
				return nil, err
			}
			if ok {
				legsByTemplate[templateID] = append(legsByTemplate[templateID], leg)
			}
		}
	}

	templateIDs := make([]uint, 0, len(groups))
	for templateID := range groups {
		templateIDs = append(templateIDs, templateID)
	}
	sort.Slice(templateIDs, func(i, j int) bool { return templateIDs[i] < templateIDs[j] })

	legsByTemplateWithFallback := make(map[uint][]ShippingQuoteLeg, len(groups))
	for _, templateID := range templateIDs {
		group := groups[templateID]
		legs := legsByTemplate[templateID]
		if len(legs) == 0 {
			requiresPublishedCollectionRoute := false
			if group.Template.RouteGroupID != nil {
				_, requiresPublishedCollectionRoute = aggregatesWithPublishedCollectionRoutes[*group.Template.RouteGroupID]
			}
			if requiresPublishedCollectionRoute || group.Template.RouteGroupID != nil {
				return nil, fmt.Errorf("%w: no active published logistics service for template %q in %s", ErrShippingQuotePlanUnavailable, group.Template.Name, country)
			}
			rate := groupRates[templateID]
			if !rate.FreeShipping {
				if err := validateTemplateWeightBillingForValue(
					group.Template, country, float64(group.TotalWeightGrams)/1000,
				); err != nil {
					return nil, err
				}
			}
			legs = []ShippingQuoteLeg{{
				GroupKey:            shippingQuoteGroupKey(templateID),
				ItemIndexes:         append([]int(nil), group.ItemIndexes...),
				AllocationBasis:     group.Template.Type,
				TemplateID:          templateID,
				TemplateName:        group.Template.Name,
				Currency:            quoteCurrency,
				BillingMode:         "template",
				ActualWeightGrams:   group.TotalWeightGrams,
				ChargeWeightGrams:   group.TotalWeightGrams,
				BillableWeightGrams: group.TotalWeightGrams,
				BaseFeeDecimal: func() string {
					m, _ := domainmoney.New(rate.FeeMinor, quoteCurrency)
					v, _ := m.FormatMajor()
					return v
				}(),
				ShippingFeeDecimal: func() string {
					m, _ := domainmoney.New(rate.FeeMinor, quoteCurrency)
					v, _ := m.FormatMajor()
					return v
				}(),
				BaseFeeMinor:     rate.FeeMinor,
				ShippingFeeMinor: rate.FeeMinor,
				DisplayPrices:    rate.DisplayPrices,
				DisplayPrice:     displayPriceForCurrency(displayCurrency, rate.DisplayPrices),
				FreeShipping:     rate.FreeShipping,
			}}
		}
		sort.SliceStable(legs, func(i, j int) bool {
			if legs[i].ShippingFeeMinor != legs[j].ShippingFeeMinor {
				return legs[i].ShippingFeeMinor < legs[j].ShippingFeeMinor
			}
			if legs[i].SortOrder != legs[j].SortOrder {
				return legs[i].SortOrder < legs[j].SortOrder
			}
			return legs[i].CarrierServiceID < legs[j].CarrierServiceID
		})

		legsByTemplateWithFallback[templateID] = legs
	}

	// A product may expose several billing templates. Groups sharing an item
	// index are alternatives and must never be added to the same plan; groups
	// covering different items remain combinable for a mixed cart.
	allItemIndexes := make(map[int]struct{})
	for _, group := range groups {
		for _, itemIndex := range group.ItemIndexes {
			allItemIndexes[itemIndex] = struct{}{}
		}
	}
	plans := make([]ShippingQuotePlan, 0)
	var appendPlans func(groupPosition int, covered map[int]struct{}, plan ShippingQuotePlan) error
	appendPlans = func(groupPosition int, covered map[int]struct{}, plan ShippingQuotePlan) error {
		if len(covered) == len(allItemIndexes) {
			finalized, finalizeErr := finalizeShippingQuotePlan(plan, displayCurrency)
			if finalizeErr != nil {
				return finalizeErr
			}
			plans = append(plans, finalized)
			if len(plans) > maxShippingQuotePlans {
				return fmt.Errorf("%w: shipment plan combinations exceed %d", ErrShippingRateConfigurationInvalid, maxShippingQuotePlans)
			}
			return nil
		}
		for position := groupPosition; position < len(templateIDs); position++ {
			templateID := templateIDs[position]
			group := groups[templateID]
			if group == nil || len(legsByTemplateWithFallback[templateID]) == 0 {
				continue
			}
			overlaps := false
			for _, itemIndex := range group.ItemIndexes {
				if _, exists := covered[itemIndex]; exists {
					overlaps = true
					break
				}
			}
			if overlaps {
				continue
			}
			nextCovered := make(map[int]struct{}, len(covered)+len(group.ItemIndexes))
			for itemIndex := range covered {
				nextCovered[itemIndex] = struct{}{}
			}
			for _, itemIndex := range group.ItemIndexes {
				nextCovered[itemIndex] = struct{}{}
			}
			for _, leg := range legsByTemplateWithFallback[templateID] {
				nextPlan := plan
				nextPlan.Legs = append(append([]ShippingQuoteLeg(nil), plan.Legs...), leg)
				if err := appendPlans(position+1, nextCovered, nextPlan); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := appendPlans(0, map[int]struct{}{}, ShippingQuotePlan{Currency: quoteCurrency}); err != nil {
		return nil, err
	}
	if len(plans) == 0 {
		return nil, ErrShippingQuotePlanUnavailable
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

func (s *ShippingService) buildCarrierServiceQuoteLeg(
	service shipping.CarrierService,
	billingTemplate *shipping.ShippingTemplate,
	group *shippingQuoteGroup,
	resolvedItems []resolvedShippingItem,
	country string,
	postalCode string,
	quoteCurrencyInput string,
	displayCurrency string,
	groupAmount domainmoney.Money,
) (ShippingQuoteLeg, bool, error) {
	if group == nil || billingTemplate == nil || !service.Enabled || !billingTemplate.Enabled {
		return ShippingQuoteLeg{}, false, nil
	}
	if billingTemplate.IsSystemFreeShippingTemplate() {
		return ShippingQuoteLeg{}, false, nil
	}
	if service.Carrier == nil || !service.Carrier.Enabled {
		return ShippingQuoteLeg{}, false, nil
	}
	if service.RouteGroup != nil && !service.RouteGroup.Enabled {
		return ShippingQuoteLeg{}, false, nil
	}
	if !shippingrating.MatchesRegion(service.Countries, country) {
		return ShippingQuoteLeg{}, false, nil
	}

	quoteCurrency := strings.ToUpper(strings.TrimSpace(quoteCurrencyInput))
	templateCurrency := currency.NormalizeCode(billingTemplate.Currency)
	if !currency.IsCatalogCode(templateCurrency) {
		return ShippingQuoteLeg{}, false, fmt.Errorf("%w: customer shipping template ID %d has invalid source currency %s", ErrShippingRateConfigurationInvalid, billingTemplate.ID, templateCurrency)
	}

	actualWeightGrams := group.TotalWeightGrams
	volumetricWeightGrams, dimensionalChargeWeightGrams, hasCompleteVolumetricWeight := customerTemplateVolumetricWeightGrams(billingTemplate, resolvedItems)
	chargeWeightGrams := actualWeightGrams
	switch billingTemplate.CustomerBillingMode {
	case "actual_weight":
	case "volumetric_weight":
		if !hasCompleteVolumetricWeight {
			return ShippingQuoteLeg{}, false, nil
		}
		chargeWeightGrams = volumetricWeightGrams
	case "greater_of_actual_and_volumetric":
		chargeWeightGrams = maxInt(actualWeightGrams, dimensionalChargeWeightGrams)
	default:
		return ShippingQuoteLeg{}, false, fmt.Errorf("%w: customer shipping template ID %d has invalid billing mode %q", ErrShippingRateConfigurationInvalid, billingTemplate.ID, billingTemplate.CustomerBillingMode)
	}

	billableWeightGrams := customerTemplateBillableWeightGrams(chargeWeightGrams, billingTemplate)
	baseFee, freeShipping, baseDisplayPrices, err := s.calculateTemplateShippingFeeForQuoteMoney(
		billingTemplate,
		country,
		billableWeightGrams,
		group.Quantity,
		groupAmount,
		groupAmount,
		quoteCurrency,
		shippingWeightBilling{
			firstWeightGrams:      billingTemplate.CustomerFirstWeightGrams,
			additionalWeightGrams: billingTemplate.CustomerAdditionalWeightGrams,
		},
	)
	if err != nil {
		return ShippingQuoteLeg{}, false, err
	}

	remoteSurcharge, err := customerTemplateRemoteSurchargeMoney(billingTemplate, postalCode)
	if err != nil {
		return ShippingQuoteLeg{}, false, err
	}
	if freeShipping {
		remoteSurcharge, err = domainmoney.New(0, quoteCurrency)
		if err != nil {
			return ShippingQuoteLeg{}, false, err
		}
	} else if templateCurrency != quoteCurrency {
		remoteSurcharge, err = s.convertShippingMoney(remoteSurcharge, quoteCurrency)
		if err != nil {
			return ShippingQuoteLeg{}, false, err
		}
	}
	shippingFeeMoney, feeErr := domainmoney.New(0, quoteCurrency)
	if feeErr != nil {
		return ShippingQuoteLeg{}, false, feeErr
	}
	fuelMoney, fuelErr := domainmoney.New(0, quoteCurrency)
	if fuelErr != nil {
		return ShippingQuoteLeg{}, false, fuelErr
	}
	if !freeShipping {
		fuelRate, rateErr := billingTemplate.CustomerFuelSurchargeRate()
		if rateErr != nil {
			return ShippingQuoteLeg{}, false, rateErr
		}
		fuelMoney, fuelErr = calculateShippingPercentageMoney(baseFee, fuelRate)
		if fuelErr != nil {
			return ShippingQuoteLeg{}, false, fuelErr
		}
		shippingFeeMoney, feeErr = baseFee.Add(fuelMoney)
		if feeErr == nil {
			shippingFeeMoney, feeErr = shippingFeeMoney.Add(remoteSurcharge)
		}
		if feeErr != nil {
			return ShippingQuoteLeg{}, false, feeErr
		}
	}
	displayPrices := deriveCarrierServiceDisplayPrices(baseDisplayPrices, baseFee, fuelMoney, remoteSurcharge, freeShipping)
	baseFeeDecimal, _ := baseFee.FormatMajor()
	fuelSurchargeDecimal, _ := fuelMoney.FormatMajor()
	remoteSurchargeDecimal, _ := remoteSurcharge.FormatMajor()
	shippingFeeDecimal, _ := shippingFeeMoney.FormatMajor()

	return ShippingQuoteLeg{
		GroupKey:               shippingQuoteGroupKey(group.Template.ID),
		ItemIndexes:            append([]int(nil), group.ItemIndexes...),
		AllocationBasis:        group.Template.Type,
		CarrierID:              service.Carrier.ID,
		CarrierName:            service.Carrier.Name,
		CarrierCode:            service.Carrier.Code,
		CarrierServiceID:       service.ID,
		ServiceCode:            service.ServiceCode,
		ServiceName:            service.ServiceName,
		RouteName:              service.RouteName,
		TemplateID:             billingTemplate.ID,
		TemplateName:           billingTemplate.Name,
		Currency:               quoteCurrency,
		BillingMode:            billingTemplate.CustomerBillingMode,
		ActualWeightGrams:      actualWeightGrams,
		VolumetricWeightGrams:  volumetricWeightGrams,
		ChargeWeightGrams:      chargeWeightGrams,
		BillableWeightGrams:    billableWeightGrams,
		BaseFeeDecimal:         baseFeeDecimal,
		FuelSurchargeDecimal:   fuelSurchargeDecimal,
		RemoteSurchargeDecimal: remoteSurchargeDecimal,
		ShippingFeeDecimal:     shippingFeeDecimal,
		BaseFeeMinor:           baseFee.AmountMinor(),
		FuelSurchargeMinor:     fuelMoney.AmountMinor(),
		RemoteSurchargeMinor:   remoteSurcharge.AmountMinor(),
		ShippingFeeMinor:       shippingFeeMoney.AmountMinor(),
		DisplayPrice:           displayPriceForCurrency(displayCurrency, displayPrices),
		DisplayPrices:          displayPrices,
		FreeShipping:           freeShipping,
		EtaMinDays:             billingTemplate.CustomerEtaMinDays,
		EtaMaxDays:             billingTemplate.CustomerEtaMaxDays,
		SortOrder:              service.SortOrder,
	}, true, nil
}

func calculateShippingPercentageMoney(baseAmount domainmoney.Money, percentageRate *big.Rat) (domainmoney.Money, error) {
	if percentageRate == nil || percentageRate.Sign() < 0 || percentageRate.Cmp(big.NewRat(100, 1)) > 0 {
		return domainmoney.Money{}, fmt.Errorf("invalid shipping surcharge percentage")
	}
	rate := new(big.Rat).Set(percentageRate)
	rate.Quo(rate, big.NewRat(100, 1))
	return baseAmount.MultiplyRat(rate)
}
func shippingQuoteGroupItems(group *shippingQuoteGroup, resolvedItems []resolvedShippingItem) []resolvedShippingItem {
	if group == nil {
		return nil
	}
	items := make([]resolvedShippingItem, 0, len(group.ItemIndexes))
	for _, sourceIndex := range group.ItemIndexes {
		for _, item := range resolvedItems {
			if item.SourceItemIndex == sourceIndex && item.Template != nil && group.Template != nil && item.Template.ID == group.Template.ID {
				items = append(items, item)
				break
			}
		}
	}
	return items
}
