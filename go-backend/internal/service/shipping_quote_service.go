package service

import (
	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/shipping"
	shippingrating "commerce-platform/internal/domain/shipping/rating"
	"errors"
	"fmt"
)

func (s *ShippingService) CalculateShipping(input ShippingCalculationInput) (*ShippingQuote, error) {
	template, err := s.GetPublicTemplate(input.TemplateID)
	if err != nil {
		return nil, err
	}

	country, ok := shippingrating.NormalizeCountry(input.Country)
	if !ok {
		return nil, fmt.Errorf("%w: country must be an ISO alpha-2 code", ErrInvalidShippingDestination)
	}
	input.Country = country
	if template.FreeShipping && input.Amount >= template.FreeThreshold {
		return &ShippingQuote{ShippingFee: 0, FreeShipping: true}, nil
	}

	value := input.Weight
	switch template.Type {
	case "quantity":
		value = float64(input.Quantity)
	case "price", "amount":
		value = input.Amount
	}
	if err := validateTemplateWeightBillingForValue(template, input.Country, value); err != nil {
		return nil, err
	}

	shippingFee, freeShipping, _, err := calculateTemplateShippingFeeForValue(template, input.Country, value)
	if err != nil {
		return nil, err
	}

	return &ShippingQuote{
		ShippingFee:  shippingFee,
		FreeShipping: freeShipping,
	}, nil
}

func (s *ShippingService) QuoteCart(input ShippingQuoteInput) (*ShippingQuote, error) {
	if s.productRepo == nil {
		return nil, errors.New("shipping quote product repository is not configured")
	}
	if len(input.Items) == 0 {
		return nil, errors.New("shipping quote requires at least one item")
	}

	items := make([]ShippingQuoteItemInput, 0, len(input.Items))
	quoteCurrency := currency.NormalizeCode(input.Currency)
	if !currency.IsCatalogCode(quoteCurrency) {
		return nil, errors.New("shipping quote currency is required")
	}
	totalAmount, err := domainmoney.New(0, quoteCurrency)
	for _, item := range input.Items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity for product ID %d", item.ProductID)
		}

		product, variant, err := s.productRepo.FindPurchasableVariant(item.ProductID, item.VariantID)
		if err != nil {
			return nil, fmt.Errorf("product ID %d is not available for shipping quote: %w", item.ProductID, err)
		}
		if variant == nil {
			return nil, fmt.Errorf("product ID %d has no purchasable SKU", item.ProductID)
		}
		if variant.Weight <= 0 {
			return nil, fmt.Errorf("shipping weight is missing for SKU %s", variant.SKU)
		}

		templateID, err := resolveProductShippingTemplateID(product, variant)
		if err != nil {
			return nil, err
		}

		resolvedVariantID := variant.ID
		unitPrice := variant.EffectivePrice()
		unitPriceMoney, err := domainmoney.FromMajorFloat(unitPrice, quoteCurrency)
		if err != nil {
			return nil, fmt.Errorf("invalid price for SKU %s: %w", variant.SKU, err)
		}
		lineAmount, err := unitPriceMoney.MultiplyInt(int64(item.Quantity))
		if err != nil {
			return nil, fmt.Errorf("calculate amount for SKU %s: %w", variant.SKU, err)
		}
		totalAmount, err = totalAmount.Add(lineAmount)
		if err != nil {
			return nil, fmt.Errorf("accumulate shipping quote amount: %w", err)
		}
		items = append(items, ShippingQuoteItemInput{
			ProductID:                      product.ID,
			VariantID:                      &resolvedVariantID,
			ProductSpecificationTemplateID: product.ProductSpecificationTemplateID,
			ShippingTemplateID:             uintPtr(templateID),
			Quantity:                       item.Quantity,
			UnitPrice:                      unitPrice,
			WeightGrams:                    variant.Weight,
		})
	}

	input.Items = items
	input.Amount, err = totalAmount.MajorFloat()
	if err != nil {
		return nil, fmt.Errorf("format shipping quote amount: %w", err)
	}
	return s.QuoteResolvedItems(input)
}

func (s *ShippingService) QuoteResolvedItems(input ShippingQuoteInput) (*ShippingQuote, error) {
	country, ok := shippingrating.NormalizeCountry(input.Country)
	if !ok {
		return nil, fmt.Errorf("%w: country must be an ISO alpha-2 code", ErrInvalidShippingDestination)
	}
	quoteCurrency := currency.NormalizeCode(input.Currency)
	if !currency.IsCatalogCode(quoteCurrency) {
		return nil, errors.New("shipping quote currency is required")
	}
	displayCurrency := currency.NormalizeCode(input.DisplayCurrency)
	if len(input.Items) == 0 {
		return nil, errors.New("shipping quote requires at least one item")
	}
	input.Country = country
	input.Currency = quoteCurrency
	input.DisplayCurrency = displayCurrency
	if input.ShippingQuoteID != "" {
		return s.restoreShippingQuote(input)
	}
	if input.SelectedQuotePlanID != "" {
		return nil, fmt.Errorf("%w: shipping_quote_id is required when selecting a plan", ErrShippingQuotePlanUnavailable)
	}

	productIDs := uniqueShippingQuoteProductIDs(input.Items)
	packagingRulesByProductAndVariant, err := s.shippingRepo.FindActivePackagingRulesByProductIDsAndVariants(productIDs)
	if err != nil {
		return nil, err
	}
	templateIDs, err := uniqueShippingQuoteTemplateIDs(input.Items)
	if err != nil {
		return nil, err
	}
	templatesByID, err := s.shippingRepo.FindTemplatesByIDs(templateIDs)
	if err != nil {
		return nil, err
	}

	resolvedItems := make([]resolvedShippingItem, 0, len(input.Items))
	for _, item := range input.Items {
		if item.ProductID == 0 {
			return nil, errors.New("shipping quote item product_id is required")
		}
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity for product ID %d", item.ProductID)
		}
		if item.WeightGrams <= 0 {
			if item.VariantID != nil {
				return nil, fmt.Errorf("shipping weight is missing for variant ID %d", *item.VariantID)
			}
			return nil, fmt.Errorf("shipping weight is missing for product ID %d", item.ProductID)
		}

		template := templatesByID[*item.ShippingTemplateID]
		if template == nil {
			return nil, fmt.Errorf("shipping template ID %d is not configured", *item.ShippingTemplateID)
		}
		if !template.Enabled {
			return nil, fmt.Errorf("shipping template ID %d is disabled", template.ID)
		}
		templateCurrency := currency.NormalizeCode(template.Currency)
		if !currency.IsCatalogCode(templateCurrency) {
			return nil, fmt.Errorf("shipping template ID %d has invalid source currency", template.ID)
		}
		for _, rule := range template.Rules {
			ruleCurrency := currency.NormalizeCode(rule.Currency)
			if !currency.IsCatalogCode(ruleCurrency) {
				return nil, fmt.Errorf("shipping rule ID %d has invalid source currency", rule.ID)
			}
			if ruleCurrency != templateCurrency {
				return nil, fmt.Errorf("shipping rule ID %d currency %s does not match template currency %s", rule.ID, ruleCurrency, templateCurrency)
			}
		}

		unitPriceMoney, err := domainmoney.FromMajorFloat(item.UnitPrice, quoteCurrency)
		if err != nil {
			return nil, fmt.Errorf("invalid shipping quote item price: %w", err)
		}
		amount, err := unitPriceMoney.MultiplyInt(int64(item.Quantity))
		if err != nil {
			return nil, fmt.Errorf("calculate shipping quote item amount: %w", err)
		}
		packagingRule := lookupPackagingRule(packagingRulesByProductAndVariant, item.ProductID, item.VariantID)
		packagingWeightGrams := packagingRuleWeightGrams(packagingRule)
		chargeWeightGrams := item.WeightGrams + packagingWeightGrams
		resolvedItems = append(resolvedItems, resolvedShippingItem{
			ShippingQuoteItemInput: item,
			Amount:                 amount,
			Template:               template,
			PackagingRule:          packagingRule,
			PackagingWeightGrams:   packagingWeightGrams,
			ChargeWeightGrams:      chargeWeightGrams,
		})
	}

	groups := make(map[uint]*shippingQuoteGroup)
	quoteItems := make([]ShippingQuoteItem, len(resolvedItems))
	for index, item := range resolvedItems {
		templateID := item.Template.ID
		group := groups[templateID]
		if group == nil {
			group = &shippingQuoteGroup{Template: item.Template}
			groups[templateID] = group
		}

		group.ItemIndexes = append(group.ItemIndexes, index)
		if group.Amount.AmountMinor() == 0 && group.Amount.Currency().String() == "" {
			group.Amount, err = domainmoney.New(0, quoteCurrency)
			if err != nil {
				return nil, err
			}
		}
		group.Amount, err = group.Amount.Add(item.Amount)
		if err != nil {
			return nil, fmt.Errorf("accumulate shipping group amount: %w", err)
		}
		group.Quantity += item.Quantity
		group.TotalWeightGrams += item.ChargeWeightGrams * item.Quantity

		var packagingRuleID *uint
		var packagingRuleName string
		if item.PackagingRule != nil {
			packagingRuleID = uintPtr(item.PackagingRule.ID)
			packagingRuleName = item.PackagingRule.RuleName
		}

		itemAmount, amountErr := item.Amount.MajorFloat()
		if amountErr != nil {
			return nil, fmt.Errorf("format shipping quote item amount: %w", amountErr)
		}
		quoteItems[index] = ShippingQuoteItem{
			ProductID:                      item.ProductID,
			VariantID:                      item.VariantID,
			ProductSpecificationTemplateID: item.ProductSpecificationTemplateID,
			TemplateID:                     item.Template.ID,
			TemplateName:                   item.Template.Name,
			PackagingRuleID:                packagingRuleID,
			PackagingRuleName:              packagingRuleName,
			Quantity:                       item.Quantity,
			UnitPrice:                      item.UnitPrice,
			Amount:                         itemAmount,
			WeightGrams:                    item.WeightGrams,
			PackagingWeightGrams:           item.PackagingWeightGrams,
			ChargeWeightGrams:              item.ChargeWeightGrams,
		}
	}

	groupRates := make(map[uint]shippingQuoteGroupRate, len(groups))
	for _, group := range groups {
		groupAmount, err := group.Amount.MajorFloat()
		if err != nil {
			return nil, fmt.Errorf("format shipping group amount: %w", err)
		}
		groupFee, groupFree, groupDisplayPrices, err := s.calculateTemplateShippingFeeForQuote(
			group.Template,
			country,
			group.TotalWeightGrams,
			group.Quantity,
			groupAmount,
			groupAmount,
			quoteCurrency,
		)
		if err != nil {
			return nil, err
		}
		groupRates[group.Template.ID] = shippingQuoteGroupRate{
			Fee: groupFee, FreeShipping: groupFree, DisplayPrices: groupDisplayPrices,
		}
	}

	plans, err := s.quoteShipmentPlans(
		country,
		input.PostalCode,
		quoteCurrency,
		displayCurrency,
		resolvedItems,
		groups,
		groupRates,
	)
	if err != nil {
		return nil, err
	}
	if len(plans) == 0 {
		return nil, ErrShippingQuotePlanUnavailable
	}

	quote := &ShippingQuote{
		Currency:        quoteCurrency,
		DisplayCurrency: displayCurrency,
		Items:           quoteItems,
		Plans:           plans,
	}
	if err := s.persistShippingQuote(input, quote); err != nil {
		return nil, err
	}
	return quote, nil
}

func lookupPackagingRule(
	rulesByProductAndVariant map[uint]map[uint]*shipping.PackagingRule,
	productID uint,
	variantID *uint,
) *shipping.PackagingRule {
	rules := rulesByProductAndVariant[productID]
	if len(rules) == 0 {
		return nil
	}
	if variantID != nil {
		if rule := rules[*variantID]; rule != nil {
			return rule
		}
	}
	return rules[0]
}
