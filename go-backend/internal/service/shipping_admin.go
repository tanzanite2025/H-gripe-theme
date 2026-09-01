package service

import (
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/shipping"
	"errors"
	"fmt"
	"strings"
)

func (s *ShippingService) ListTemplates() ([]shipping.ShippingTemplate, error) {
	return s.shippingRepo.FindAllTemplates()
}

func (s *ShippingService) GetTemplate(id uint) (*shipping.ShippingTemplate, error) {
	return s.shippingRepo.FindTemplateByID(id)
}

func (s *ShippingService) CreateTemplate(template *shipping.ShippingTemplate) error {
	if err := s.prepareShippingTemplateCurrencies(template); err != nil {
		return err
	}
	return s.shippingRepo.CreateTemplateWithRules(template, template.Rules)
}

func (s *ShippingService) UpdateTemplate(template *shipping.ShippingTemplate) error {
	if template == nil {
		return errors.New("shipping template is required")
	}
	existing, err := s.GetTemplate(template.ID)
	if err != nil {
		return err
	}
	if currency.NormalizeCode(template.Currency) == "" {
		template.Currency = existing.Currency
	}
	if len(template.DisplayPriceData) == 0 {
		template.DisplayPriceData = existing.DisplayPriceData
	}
	if err := s.prepareShippingTemplateCurrencies(template); err != nil {
		return err
	}
	return s.shippingRepo.UpdateTemplateWithRules(template, template.Rules)
}

func (s *ShippingService) DeleteTemplate(id uint) error {
	return s.shippingRepo.DeleteTemplate(id)
}

func (s *ShippingService) CreateTemplateRule(templateID uint, rule *shipping.ShippingRule) error {
	rule.TemplateID = templateID
	if err := s.prepareShippingRuleCurrency(templateID, rule); err != nil {
		return err
	}
	return s.shippingRepo.CreateRule(rule)
}

func (s *ShippingService) UpdateTemplateRule(templateID uint, rule *shipping.ShippingRule) error {
	rule.TemplateID = templateID
	if err := s.prepareShippingRuleCurrency(templateID, rule); err != nil {
		return err
	}
	return s.shippingRepo.UpdateRuleForTemplate(rule)
}

func (s *ShippingService) DeleteTemplateRule(templateID uint, ruleID uint) error {
	return s.shippingRepo.DeleteRuleForTemplate(templateID, ruleID)
}

func (s *ShippingService) CalculateShipping(input ShippingCalculationInput) (*ShippingQuote, error) {
	template, err := s.GetPublicTemplate(input.TemplateID)
	if err != nil {
		return nil, err
	}

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

	shippingFee := template.DefaultFee
	for _, rule := range template.Rules {
		if shippingRuleMatchesCountry(rule.Region, input.Country) && value >= rule.MinValue && (rule.MaxValue == 0 || value <= rule.MaxValue) {
			shippingFee = calculateRuleFee(rule, value)
			break
		}
	}

	return &ShippingQuote{ShippingFee: roundMoney(shippingFee), FreeShipping: false}, nil
}

func (s *ShippingService) QuoteCart(input ShippingQuoteInput) (*ShippingQuote, error) {
	if s.productRepo == nil {
		return nil, errors.New("shipping quote product repository is not configured")
	}
	if len(input.Items) == 0 {
		return nil, errors.New("shipping quote requires at least one item")
	}

	items := make([]ShippingQuoteItemInput, 0, len(input.Items))
	var amount float64
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
		amount += unitPrice * float64(item.Quantity)
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
	input.Amount = amount
	return s.QuoteResolvedItems(input)
}

func (s *ShippingService) QuoteResolvedItems(input ShippingQuoteInput) (*ShippingQuote, error) {
	country := strings.ToUpper(strings.TrimSpace(input.Country))
	if country == "" {
		return nil, errors.New("shipping country is required")
	}
	quoteCurrency := currency.NormalizeCode(input.Currency)
	if !currency.IsCatalogCode(quoteCurrency) {
		return nil, errors.New("shipping quote currency is required")
	}
	if len(input.Items) == 0 {
		return nil, errors.New("shipping quote requires at least one item")
	}

	productIDs := uniqueShippingQuoteProductIDs(input.Items)
	packagingRulesByProduct, err := s.shippingRepo.FindActivePackagingRulesByProductIDs(productIDs)
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
	var cartAmount float64
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
		if templateCurrency != quoteCurrency {
			return nil, fmt.Errorf("shipping template %s currency %s does not match quote currency %s", template.Name, templateCurrency, quoteCurrency)
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

		amount := item.UnitPrice * float64(item.Quantity)
		packagingRule := packagingRulesByProduct[item.ProductID]
		packagingWeightGrams := packagingRuleWeightGrams(packagingRule)
		chargeWeightGrams := item.WeightGrams + packagingWeightGrams
		cartAmount += amount
		resolvedItems = append(resolvedItems, resolvedShippingItem{
			ShippingQuoteItemInput: item,
			Amount:                 amount,
			Template:               template,
			PackagingRule:          packagingRule,
			PackagingWeightGrams:   packagingWeightGrams,
			ChargeWeightGrams:      chargeWeightGrams,
		})
	}

	if input.Amount > 0 {
		cartAmount = input.Amount
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
		group.Amount += item.Amount
		group.Quantity += item.Quantity
		group.TotalWeightGrams += item.ChargeWeightGrams * item.Quantity

		var packagingRuleID *uint
		var packagingRuleName string
		if item.PackagingRule != nil {
			packagingRuleID = uintPtr(item.PackagingRule.ID)
			packagingRuleName = item.PackagingRule.RuleName
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
			Amount:                         roundMoney(item.Amount),
			WeightGrams:                    item.WeightGrams,
			PackagingWeightGrams:           item.PackagingWeightGrams,
			ChargeWeightGrams:              item.ChargeWeightGrams,
		}
	}

	displayCurrency := currency.NormalizeCode(input.DisplayCurrency)
	var shippingFee float64
	freeShipping := false
	groupDisplayPriceSets := make([][]currency.DisplayPriceSnapshot, 0, len(groups))
	for _, group := range groups {
		groupFee, groupFree, groupDisplayPrices := calculateTemplateShippingFeeWithDisplayPrices(
			group.Template,
			country,
			group.TotalWeightGrams,
			group.Quantity,
			group.Amount,
			cartAmount,
		)
		if groupFree {
			freeShipping = true
		}
		shippingFee += groupFee
		if groupFee > 0 {
			groupDisplayPriceSets = append(groupDisplayPriceSets, groupDisplayPrices)
		}
		distributeGroupFee(group, resolvedItems, quoteItems, groupFee, groupFree)
	}

	shippingFee = roundMoney(shippingFee)
	if shippingFee > 0 {
		freeShipping = false
	}
	displayPrices := combineDisplayPriceSets(groupDisplayPriceSets)

	options, err := s.quoteCarrierServiceOptions(country, quoteCurrency, displayCurrency, resolvedItems, groups, cartAmount)
	if err != nil {
		return nil, err
	}

	source := "template"
	var selectedOption *ShippingQuoteOption
	if len(options) > 0 {
		source = "carrier_service"
		selectedOption = &options[0]
		shippingFee = selectedOption.ShippingFee
		freeShipping = selectedOption.FreeShipping
		displayPrices = selectedOption.DisplayPrices
		if group := singleShippingQuoteGroup(groups); group != nil {
			distributeGroupFee(group, resolvedItems, quoteItems, selectedOption.ShippingFee, selectedOption.FreeShipping)
		}
	}

	return &ShippingQuote{
		ShippingFee:     shippingFee,
		FreeShipping:    freeShipping,
		Currency:        quoteCurrency,
		DisplayPrice:    displayPriceForCurrency(displayCurrency, displayPrices),
		DisplayPrices:   displayPrices,
		DisplayCurrency: displayCurrency,
		Source:          source,
		Items:           quoteItems,
		Options:         options,
		SelectedOption:  selectedOption,
	}, nil
}

func (s *ShippingService) prepareShippingTemplateCurrencies(template *shipping.ShippingTemplate) error {
	if template == nil {
		return errors.New("shipping template is required")
	}
	templateCurrency, err := s.normalizeShippingSourceCurrency(template.Currency)
	if err != nil {
		return err
	}
	template.Currency = templateCurrency
	for i := range template.Rules {
		ruleCurrency := currency.NormalizeCode(template.Rules[i].Currency)
		if ruleCurrency == "" {
			template.Rules[i].Currency = templateCurrency
			continue
		}
		ruleCurrency, err = s.normalizeShippingSourceCurrency(ruleCurrency)
		if err != nil {
			return err
		}
		if ruleCurrency != templateCurrency {
			return fmt.Errorf("shipping rule currency %s does not match template currency %s", ruleCurrency, templateCurrency)
		}
		template.Rules[i].Currency = ruleCurrency
	}
	return nil
}

func (s *ShippingService) prepareShippingRuleCurrency(templateID uint, rule *shipping.ShippingRule) error {
	if rule == nil {
		return errors.New("shipping rule is required")
	}
	template, err := s.GetTemplate(templateID)
	if err != nil {
		return err
	}
	templateCurrency := currency.NormalizeCode(template.Currency)
	ruleCurrency := currency.NormalizeCode(rule.Currency)
	if ruleCurrency == "" {
		ruleCurrency = templateCurrency
	}
	normalized, err := s.normalizeShippingSourceCurrency(ruleCurrency)
	if err != nil {
		return err
	}
	if templateCurrency != "" && normalized != templateCurrency {
		return fmt.Errorf("shipping rule currency %s does not match template currency %s", normalized, templateCurrency)
	}
	rule.Currency = normalized
	return nil
}

func (s *ShippingService) normalizeShippingSourceCurrency(value string) (string, error) {
	code := currency.NormalizeCode(value)
	if code == "" {
		code = currency.DefaultPrimaryCurrency
		if s != nil && s.currencyPolicy != nil {
			primaryCurrency, err := s.currencyPolicy.BackendEntryCurrency()
			if err != nil {
				return "", fmt.Errorf("resolve backend entry currency: %w", err)
			}
			code = currency.NormalizeCode(primaryCurrency)
		}
	}
	if !currency.IsCatalogCode(code) {
		return "", ErrShippingCurrencyInvalid
	}
	return code, nil
}
