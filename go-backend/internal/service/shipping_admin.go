package service

import (
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/shipping"
	"errors"
	"fmt"
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
