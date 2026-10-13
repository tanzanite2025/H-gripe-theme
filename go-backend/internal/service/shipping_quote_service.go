package service

import (
	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/shipping"
	shippingrating "commerce-platform/internal/domain/shipping/rating"
	"errors"
	"fmt"
)

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
	totalAmount, initErr := domainmoney.New(0, quoteCurrency)
	if initErr != nil {
		return nil, initErr
	}
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
		configuration := ProductConfigurationResult{}
		if len(item.ConfigurationData) > 0 && string(item.ConfigurationData) != "{}" {
			selectedOptions, configErr := SelectedOptionsFromConfiguration(item.ConfigurationData)
			if configErr != nil {
				return nil, fmt.Errorf("invalid configuration for SKU %s: %w", variant.SKU, configErr)
			}
			configuration, configErr = ResolveProductConfiguration(product, variant, selectedOptions)
			if configErr != nil {
				return nil, fmt.Errorf("configuration conflict for SKU %s: %w", variant.SKU, configErr)
			}
		}
		configuredWeight := variant.Weight + configuration.WeightDeltaGrams
		if configuredWeight <= 0 {
			return nil, fmt.Errorf("shipping weight is missing for configured SKU %s", variant.SKU)
		}

		templateIDs, err := resolveProductShippingTemplateIDs(s.productRepo, product, variant)
		if err != nil {
			return nil, err
		}

		resolvedVariantID := variant.ID
		// Use the same catalog price and strict currency conversion
		// as Checkout so price rules and free thresholds use real amounts.
		unitPriceMoney, err := variant.PriceMoney()
		if err != nil {
			return nil, fmt.Errorf("invalid price for SKU %s: %w", variant.SKU, err)
		}
		if configuration.Delta.AmountMinor() != 0 {
			unitPriceMoney, err = unitPriceMoney.Add(configuration.Delta)
			if err != nil {
				return nil, fmt.Errorf("calculate configured price for SKU %s: %w", variant.SKU, err)
			}
		}
		unitPriceMoney, err = s.convertShippingMoney(unitPriceMoney, quoteCurrency)
		if err != nil {
			return nil, fmt.Errorf("convert catalog price for SKU %s: %w", variant.SKU, err)
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
			ShippingTemplateIDs:            append([]uint(nil), templateIDs...),
			Quantity:                       item.Quantity,
			UnitPriceMinor:                 unitPriceMoney.AmountMinor(),
			WeightGrams:                    configuredWeight,
			PackagingWeightDeltaGrams:      configuration.PackagingWeightDeltaGrams,
		})
	}

	input.Items = items
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
	for sourceItemIndex, item := range input.Items {
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

		templateIDsForItem := item.ShippingTemplateIDs
		if len(templateIDsForItem) == 0 {
			return nil, fmt.Errorf("shipping template is missing for product ID %d", item.ProductID)
		}

		unitPriceMoney, err := domainmoney.New(item.UnitPriceMinor, quoteCurrency)
		if err != nil {
			return nil, fmt.Errorf("invalid shipping quote item price: %w", err)
		}
		amount, err := unitPriceMoney.MultiplyInt(int64(item.Quantity))
		if err != nil {
			return nil, fmt.Errorf("calculate shipping quote item amount: %w", err)
		}
		packagingRule := lookupPackagingRule(packagingRulesByProductAndVariant, item.ProductID, item.VariantID)
		packagingWeightGrams := packagingRuleWeightGrams(packagingRule) + item.PackagingWeightDeltaGrams
		if packagingWeightGrams < 0 {
			return nil, fmt.Errorf("packaging weight cannot be negative for product ID %d", item.ProductID)
		}
		chargeWeightGrams := item.WeightGrams + packagingWeightGrams
		candidateCountBeforeItem := len(resolvedItems)
		for _, templateID := range templateIDsForItem {
			template := templatesByID[templateID]
			if template == nil {
				return nil, fmt.Errorf("shipping template ID %d is not configured", templateID)
			}
			if !template.Enabled {
				continue
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
			resolvedItem := item
			resolvedItem.ShippingTemplateIDs = []uint{templateID}
			resolvedItems = append(resolvedItems, resolvedShippingItem{
				ShippingQuoteItemInput: resolvedItem,
				SourceItemIndex:        sourceItemIndex,
				Amount:                 amount,
				Template:               template,
				PackagingRule:          packagingRule,
				PackagingWeightGrams:   packagingWeightGrams,
				ChargeWeightGrams:      chargeWeightGrams,
			})
		}
		if len(resolvedItems) == candidateCountBeforeItem {
			return nil, fmt.Errorf("%w: product ID %d has no enabled customer shipping template", ErrShippingQuotePlanUnavailable, item.ProductID)
		}
	}

	quoteItems := make([]ShippingQuoteItem, len(input.Items))
	quoteItemInitialized := make([]bool, len(input.Items))
	for _, item := range resolvedItems {
		var packagingRuleID *uint
		var packagingRuleName string
		if item.PackagingRule != nil {
			packagingRuleID = uintPtr(item.PackagingRule.ID)
			packagingRuleName = item.PackagingRule.RuleName
		}

		unitPriceDecimal := ""
		if value, moneyErr := domainmoney.New(item.UnitPriceMinor, quoteCurrency); moneyErr == nil {
			unitPriceDecimal, _ = value.FormatMajor()
		}
		amountDecimal, _ := item.Amount.FormatMajor()
		if quoteItemInitialized[item.SourceItemIndex] {
			continue
		}
		quoteItemInitialized[item.SourceItemIndex] = true
		quoteItems[item.SourceItemIndex] = ShippingQuoteItem{
			ProductID:                      item.ProductID,
			VariantID:                      item.VariantID,
			ProductSpecificationTemplateID: item.ProductSpecificationTemplateID,
			TemplateID:                     item.Template.ID,
			TemplateName:                   item.Template.Name,
			PackagingRuleID:                packagingRuleID,
			PackagingRuleName:              packagingRuleName,
			Quantity:                       item.Quantity,
			UnitPriceDecimal:               unitPriceDecimal,
			UnitPriceMinor:                 item.UnitPriceMinor,
			AmountDecimal:                  amountDecimal,
			AmountMinor:                    item.Amount.AmountMinor(),
			WeightGrams:                    item.WeightGrams,
			PackagingWeightGrams:           item.PackagingWeightGrams,
			ChargeWeightGrams:              item.ChargeWeightGrams,
			TemplateIDs:                    append([]uint(nil), input.Items[item.SourceItemIndex].ShippingTemplateIDs...),
		}
	}

	plans, err := s.quoteProductShippingTemplateAlternatives(country, input.PostalCode, quoteCurrency, displayCurrency, resolvedItems)
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

func containsShippingQuoteItemIndex(indexes []int, target int) bool {
	for _, index := range indexes {
		if index == target {
			return true
		}
	}
	return false
}
