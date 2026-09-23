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

	carrierServices, err := s.shippingRepo.FindEnabledCarrierServicesWithTemplates()
	if err != nil {
		return nil, err
	}

	legsByTemplate := make(map[uint][]ShippingQuoteLeg, len(groups))
	for i := range carrierServices {
		carrierService := carrierServices[i]
		if carrierService.TemplateID == nil {
			continue
		}
		group := groups[*carrierService.TemplateID]
		if group == nil {
			continue
		}
		leg, ok, err := s.buildCarrierServiceQuoteLeg(
			carrierService,
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
			legsByTemplate[*carrierService.TemplateID] = append(legsByTemplate[*carrierService.TemplateID], leg)
		}
	}

	templateIDs := make([]uint, 0, len(groups))
	for templateID := range groups {
		templateIDs = append(templateIDs, templateID)
	}
	sort.Slice(templateIDs, func(i, j int) bool { return templateIDs[i] < templateIDs[j] })

	plans := []ShippingQuotePlan{{Currency: quoteCurrency}}
	for _, templateID := range templateIDs {
		group := groups[templateID]
		legs := legsByTemplate[templateID]
		if len(legs) == 0 {
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

		if len(plans)*len(legs) > maxShippingQuotePlans {
			return nil, fmt.Errorf("%w: shipment plan combinations exceed %d", ErrShippingRateConfigurationInvalid, maxShippingQuotePlans)
		}
		next := make([]ShippingQuotePlan, 0, len(plans)*len(legs))
		for _, plan := range plans {
			for _, leg := range legs {
				candidate := plan
				candidate.Legs = append(append([]ShippingQuoteLeg(nil), plan.Legs...), leg)
				finalized, finalizeErr := finalizeShippingQuotePlan(candidate, displayCurrency)
				if finalizeErr != nil {
					return nil, finalizeErr
				}
				next = append(next, finalized)
			}
		}
		plans = next
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
	group *shippingQuoteGroup,
	resolvedItems []resolvedShippingItem,
	country string,
	postalCode string,
	quoteCurrencyInput string,
	displayCurrency string,
	groupAmount domainmoney.Money,
) (ShippingQuoteLeg, bool, error) {
	if group == nil || service.Template == nil || !service.Enabled || !service.Template.Enabled {
		return ShippingQuoteLeg{}, false, nil
	}
	if service.TemplateID == nil || *service.TemplateID != group.Template.ID {
		return ShippingQuoteLeg{}, false, nil
	}
	if service.Carrier == nil || !service.Carrier.Enabled {
		return ShippingQuoteLeg{}, false, nil
	}
	if !shippingrating.MatchesRegion(service.Countries, country) {
		return ShippingQuoteLeg{}, false, nil
	}

	quoteCurrency := strings.ToUpper(strings.TrimSpace(quoteCurrencyInput))
	serviceCurrency := strings.ToUpper(strings.TrimSpace(service.Currency))
	if serviceCurrency == "" {
		serviceCurrency = quoteCurrency
	}
	if !currency.IsCatalogCode(serviceCurrency) {
		return ShippingQuoteLeg{}, false, fmt.Errorf("%w: carrier service %s has invalid source currency %s", ErrShippingRateConfigurationInvalid, service.ServiceName, serviceCurrency)
	}

	actualWeightGrams := group.TotalWeightGrams
	volumetricWeightGrams, dimensionalChargeWeightGrams, hasCompleteVolumetricWeight := carrierServiceVolumetricWeightGrams(service, resolvedItems)
	chargeWeightGrams := actualWeightGrams
	switch service.BillingMode {
	case "volumetric_weight":
		if !hasCompleteVolumetricWeight {
			return ShippingQuoteLeg{}, false, nil
		}
		chargeWeightGrams = volumetricWeightGrams
	case "greater_of_actual_and_volumetric":
		chargeWeightGrams = maxInt(actualWeightGrams, dimensionalChargeWeightGrams)
	}

	billableWeightGrams := carrierServiceBillableWeightGrams(chargeWeightGrams, service)
	baseFee, freeShipping, baseDisplayPrices, err := s.calculateTemplateShippingFeeForQuoteMoney(
		service.Template,
		country,
		billableWeightGrams,
		group.Quantity,
		groupAmount,
		groupAmount,
		quoteCurrency,
		shippingWeightBilling{
			firstWeightGrams:      service.FirstWeightGrams,
			additionalWeightGrams: service.AdditionalWeightGrams,
		},
	)
	if err != nil {
		return ShippingQuoteLeg{}, false, err
	}

	remoteSurcharge, err := carrierServiceRemoteSurchargeMoney(service, postalCode)
	if err != nil {
		return ShippingQuoteLeg{}, false, err
	}
	if freeShipping {
		remoteSurcharge, err = domainmoney.New(0, quoteCurrency)
		if err != nil {
			return ShippingQuoteLeg{}, false, err
		}
	} else if serviceCurrency != quoteCurrency {
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
		fuelRate, rateErr := service.FuelSurchargeRate()
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
		TemplateID:             service.Template.ID,
		TemplateName:           service.Template.Name,
		Currency:               quoteCurrency,
		BillingMode:            service.BillingMode,
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
		EtaMinDays:             service.EtaMinDays,
		EtaMaxDays:             service.EtaMaxDays,
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
	for _, index := range group.ItemIndexes {
		if index < 0 || index >= len(resolvedItems) {
			continue
		}
		items = append(items, resolvedItems[index])
	}
	return items
}
