package service

import (
	"errors"
	"commerce-platform/internal/domain/currency"
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
	ratesByQuote := displayPriceRatesByQuote(rates, baseCurrency)
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

		templateFields := shippingTemplateDisplayPriceAmounts(template)
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

			ruleFields := shippingRuleDisplayPriceAmounts(template.Type, rule)
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

func shippingTemplateDisplayPriceAmounts(template *shippingdomain.ShippingTemplate) map[string]float64 {
	if template == nil {
		return nil
	}
	return map[string]float64{
		shippingdomain.ShippingTemplateDisplayPriceFieldDefaultFee:    template.DefaultFee,
		shippingdomain.ShippingTemplateDisplayPriceFieldFreeThreshold: template.FreeThreshold,
	}
}

func shippingRuleDisplayPriceAmounts(templateType string, rule shippingdomain.ShippingRule) map[string]float64 {
	amounts := map[string]float64{
		shippingdomain.ShippingRuleDisplayPriceFieldFee:        rule.Fee,
		shippingdomain.ShippingRuleDisplayPriceFieldAdditional: rule.Additional,
	}
	if templateType == "price" {
		amounts[shippingdomain.ShippingRuleDisplayPriceFieldMinValue] = rule.MinValue
		amounts[shippingdomain.ShippingRuleDisplayPriceFieldMaxValue] = rule.MaxValue
	}
	return amounts
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
	amounts map[string]float64,
	baseCurrency string,
	quoteCurrencies []string,
	ratesByQuote map[string]float64,
	allowedFields []string,
) datatypes.JSON {
	previous := currency.ParseDisplayPriceSnapshotMap(raw, allowedFields...)
	next := make(map[string][]currency.DisplayPriceSnapshot, len(amounts))
	for _, field := range allowedFields {
		amount := amounts[field]
		if amount <= 0 {
			continue
		}

		snapshotJSON := displayPriceSnapshotJSON(
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
