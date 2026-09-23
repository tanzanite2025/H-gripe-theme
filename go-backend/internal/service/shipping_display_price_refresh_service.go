package service

import (
	"errors"
	"fmt"
	"math/big"

	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
)

var ErrShippingCurrencyInvalid = errors.New("shipping source currency is invalid")

type ShippingDisplayPriceRefreshResult struct {
	BaseCurrency          string   `json:"base_currency"`
	QuoteCurrencies       []string `json:"quote_currencies"`
	TemplatesScanned      int      `json:"templates_scanned"`
	TemplatesUpdated      int      `json:"templates_updated"`
	RulesScanned          int      `json:"rules_scanned"`
	RulesUpdated          int      `json:"rules_updated"`
	CurrencyMismatchCount int      `json:"currency_mismatch_count"`
}

// RefreshDisplayPriceSnapshots updates only shipping display snapshots.
// Template/rule source currencies and amounts are never changed here.
func (s *ShippingService) RefreshDisplayPriceSnapshots(
	baseCurrency string,
	quoteCurrencies []string,
	rates []currency.ExchangeRate,
) (ShippingDisplayPriceRefreshResult, error) {
	if s == nil || s.shippingRepo == nil {
		return ShippingDisplayPriceRefreshResult{}, errors.New("shipping service is not configured")
	}

	baseCurrency = currency.NormalizeCode(baseCurrency)
	if !currency.IsCatalogCode(baseCurrency) {
		return ShippingDisplayPriceRefreshResult{}, ErrShippingCurrencyInvalid
	}

	quoteCurrencies = normalizeDisplayPriceRefreshQuotes(quoteCurrencies, baseCurrency)
	ratesByQuote := displayPriceRatesByQuoteRat(rates, baseCurrency)
	templates, err := s.shippingRepo.FindAllTemplates()
	if err != nil {
		return ShippingDisplayPriceRefreshResult{}, err
	}

	result := ShippingDisplayPriceRefreshResult{
		BaseCurrency:     baseCurrency,
		QuoteCurrencies:  append([]string(nil), quoteCurrencies...),
		TemplatesScanned: len(templates),
	}
	updates := make([]repository.ShippingDisplayPriceSnapshotUpdate, 0, len(templates))

	for i := range templates {
		template := &templates[i]
		result.RulesScanned += len(template.Rules)

		templateCurrency := currency.NormalizeCode(template.Currency)
		if templateCurrency != baseCurrency {
			result.CurrencyMismatchCount++
			for _, rule := range template.Rules {
				if currency.NormalizeCode(rule.Currency) != baseCurrency {
					result.CurrencyMismatchCount++
				}
			}
			continue
		}

		templateFields, fieldsErr := shippingTemplateDisplayPriceAmounts(template)
		if fieldsErr != nil {
			return result, fmt.Errorf("resolve display amounts for shipping template %d: %w", template.ID, fieldsErr)
		}
		templateDisplayPriceData := refreshShippingDisplayPriceMap(
			template.DisplayPriceData,
			templateFields,
			baseCurrency,
			quoteCurrencies,
			ratesByQuote,
			shippingdomain.ShippingTemplateDisplayPriceFields,
		)
		update := repository.ShippingDisplayPriceSnapshotUpdate{
			TemplateID:       template.ID,
			DisplayPriceData: templateDisplayPriceData,
		}
		result.TemplatesUpdated++

		for _, rule := range template.Rules {
			if currency.NormalizeCode(rule.Currency) != baseCurrency {
				result.CurrencyMismatchCount++
				continue
			}

			ruleFields, fieldsErr := shippingRuleDisplayPriceAmounts(template.Type, template.Currency, rule)
			if fieldsErr != nil {
				return result, fmt.Errorf("resolve display amounts for shipping rule %d: %w", rule.ID, fieldsErr)
			}
			update.RuleUpdates = append(update.RuleUpdates, repository.ShippingRuleDisplayPriceSnapshotUpdate{
				RuleID: rule.ID,
				DisplayPriceData: refreshShippingDisplayPriceMap(
					rule.DisplayPriceData,
					ruleFields,
					baseCurrency,
					quoteCurrencies,
					ratesByQuote,
					shippingRuleDisplayPriceFieldsForType(template.Type),
				),
			})
			result.RulesUpdated++
		}

		updates = append(updates, update)
	}

	if err := s.shippingRepo.UpdateDisplayPriceSnapshots(updates); err != nil {
		return result, err
	}
	return result, nil
}

func shippingTemplateDisplayPriceAmounts(template *shippingdomain.ShippingTemplate) (map[string]domainmoney.Money, error) {
	if template == nil {
		return nil, errors.New("shipping template is required")
	}
	amounts := make(map[string]domainmoney.Money, 2)
	defaultFee, err := template.DefaultFeeMoney()
	if err != nil {
		return nil, fmt.Errorf("default fee: %w", err)
	}
	freeThreshold, err := template.FreeThresholdMoney()
	if err != nil {
		return nil, fmt.Errorf("free threshold: %w", err)
	}
	amounts[shippingdomain.ShippingTemplateDisplayPriceFieldDefaultFee] = defaultFee
	amounts[shippingdomain.ShippingTemplateDisplayPriceFieldFreeThreshold] = freeThreshold
	return amounts, nil
}

func shippingRuleDisplayPriceAmounts(templateType, templateCurrency string, rule shippingdomain.ShippingRule) (map[string]domainmoney.Money, error) {
	code := rule.Currency
	if code == "" {
		code = templateCurrency
	}
	amounts := make(map[string]domainmoney.Money, 4)
	fee, err := rule.FeeMoney(code)
	if err != nil {
		return nil, fmt.Errorf("fee: %w", err)
	}
	additional, err := rule.AdditionalMoney(code)
	if err != nil {
		return nil, fmt.Errorf("additional: %w", err)
	}
	amounts[shippingdomain.ShippingRuleDisplayPriceFieldFee] = fee
	amounts[shippingdomain.ShippingRuleDisplayPriceFieldAdditional] = additional
	if templateType == "price" {
		minMoney, err := rule.MinValueMoney(code)
		if err != nil {
			return nil, fmt.Errorf("minimum value: %w", err)
		}
		maxMoney, err := rule.MaxValueMoney(code)
		if err != nil {
			return nil, fmt.Errorf("maximum value: %w", err)
		}
		amounts[shippingdomain.ShippingRuleDisplayPriceFieldMinValue] = minMoney
		amounts[shippingdomain.ShippingRuleDisplayPriceFieldMaxValue] = maxMoney
	}
	return amounts, nil
}

func shippingRuleDisplayPriceFieldsForType(templateType string) []string {
	if templateType == "price" {
		return shippingdomain.ShippingRuleDisplayPriceFields
	}
	return []string{
		shippingdomain.ShippingRuleDisplayPriceFieldFee,
		shippingdomain.ShippingRuleDisplayPriceFieldAdditional,
	}
}

func refreshShippingDisplayPriceMap(
	raw datatypes.JSON,
	amounts map[string]domainmoney.Money,
	baseCurrency string,
	quoteCurrencies []string,
	ratesByQuote map[string]*big.Rat,
	allowedFields []string,
) datatypes.JSON {
	previous := currency.ParseDisplayPriceSnapshotMap(raw, allowedFields...)
	next := make(map[string][]currency.DisplayPriceSnapshot, len(amounts))
	for _, field := range allowedFields {
		amount, ok := amounts[field]
		if !ok || amount.AmountMinor() <= 0 {
			continue
		}

		snapshotJSON := displayPriceSnapshotMoneyJSON(
			amount,
			nil,
			baseCurrency,
			quoteCurrencies,
			ratesByQuote,
			previous[field],
		)
		snapshots := currency.ParseDisplayPriceSnapshots(snapshotJSON)
		if len(snapshots) > 0 {
			next[field] = snapshots
		}
	}
	return currency.DisplayPriceSnapshotMapJSON(next, baseCurrency, allowedFields...)
}
