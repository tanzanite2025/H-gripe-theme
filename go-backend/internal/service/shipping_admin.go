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
	return s.CreateTemplateWithCarrierServices(template, nil)
}

func (s *ShippingService) CreateTemplateWithCarrierServices(template *shipping.ShippingTemplate, carrierServices []shipping.CarrierService) error {
	if template == nil {
		return errors.New("shipping template is required")
	}
	template.NormalizeSystemFreeShippingTemplateIdentity()
	if template.IsSystemFreeShippingTemplate() && !template.FreeShipping {
		return errors.New("system free-shipping template must keep free shipping enabled")
	}
	template.ApplySystemFreeShippingTemplatePricingDefaults()
	template.FreeShippingCountries = shipping.NormalizeShippingCountryCodes(template.FreeShippingCountries)
	if err := shipping.ValidateShippingCountryCodes(template.FreeShippingCountries); err != nil {
		return err
	}
	if template.IsSystemFreeShippingTemplate() {
		template.IsSystemManaged = true
		if !template.FreeShipping {
			return errors.New("system free-shipping template must keep free shipping enabled")
		}
		if len(template.ShippingCountryCodes()) == 0 {
			return errors.New("system free-shipping template requires at least one country")
		}
		if len(carrierServices) > 0 {
			return errors.New("system free-shipping template cannot be bound to carrier services")
		}
	}
	if err := s.prepareShippingTemplateCurrencies(template); err != nil {
		return err
	}
	if err := s.validateAndHydratePublishedCarrierServiceCollections(carrierServices); err != nil {
		return err
	}
	return s.shippingRepo.CreateTemplateWithRulesAndCarrierServices(template, template.Rules, carrierServices)
}

func (s *ShippingService) UpdateTemplate(template *shipping.ShippingTemplate) error {
	return s.UpdateTemplateWithCarrierServices(template, nil)
}

func (s *ShippingService) UpdateTemplateWithCarrierServices(template *shipping.ShippingTemplate, carrierServices []shipping.CarrierService) error {
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
	if existing.IsSystemManaged {
		// IsSystemManaged is an ownership flag. Preserve it for future system
		// template kinds without making every such template behave like free
		// shipping.
		template.IsSystemManaged = true
		template.TemplateKind = existing.TemplateKind
	}
	if existing.IsSystemFreeShippingTemplate() {
		template.TemplateKind = existing.TemplateKind
		template.Type = shipping.ShippingTemplateTypeSystemFreeShipping
		template.IsSystemManaged = true
		if !template.FreeShippingCountriesProvided && len(template.ShippingCountryCodes()) == 0 && len(existing.ShippingCountryCodes()) > 0 {
			// Preserve the configured scope for legacy clients that predate the
			// optional free_shipping_countries request field.
			template.FreeShippingCountries = existing.FreeShippingCountries
		}
	}
	template.NormalizeSystemFreeShippingTemplateIdentity()
	if template.IsSystemFreeShippingTemplate() && !template.FreeShipping {
		return errors.New("system free-shipping template must keep free shipping enabled")
	}
	template.ApplySystemFreeShippingTemplatePricingDefaults()
	template.FreeShippingCountries = shipping.NormalizeShippingCountryCodes(template.FreeShippingCountries)
	if err := shipping.ValidateShippingCountryCodes(template.FreeShippingCountries); err != nil {
		return err
	}
	if template.IsSystemFreeShippingTemplate() {
		template.IsSystemManaged = true
		if !template.FreeShipping {
			return errors.New("system free-shipping template must keep free shipping enabled")
		}
		if len(template.ShippingCountryCodes()) == 0 {
			return errors.New("system free-shipping template requires at least one country")
		}
		if len(carrierServices) > 0 {
			return errors.New("system free-shipping template cannot be bound to carrier services")
		}
		// Force-detach any legacy accidental binding even when an older client
		// omits the optional carrier_services field on update.
		carrierServices = []shipping.CarrierService{}
	}
	if err := s.prepareShippingTemplateCurrencies(template); err != nil {
		return err
	}
	if err := s.validateAndHydratePublishedCarrierServiceCollections(carrierServices); err != nil {
		return err
	}
	return s.shippingRepo.UpdateTemplateWithRulesAndCarrierServices(template, template.Rules, carrierServices)
}

// validateAndHydratePublishedCarrierServiceCollections requires 4PX and Yanwen
// routes to reference an enabled production collection record by ID. Generic
// carriers retain their existing manually managed country scopes.
func (s *ShippingService) validateAndHydratePublishedCarrierServiceCollections(services []shipping.CarrierService) error {
	if len(services) == 0 {
		return nil
	}

	carriers, err := s.shippingRepo.FindAllCarriers(false)
	if err != nil {
		return fmt.Errorf("load carriers for service collection validation: %w", err)
	}
	carriersByID := make(map[uint]shipping.Carrier, len(carriers))
	for _, carrier := range carriers {
		carriersByID[carrier.ID] = carrier
	}

	needsFpx := false
	needsYanwen := false
	for i := range services {
		carrier, ok := carriersByID[services[i].CarrierID]
		if !ok {
			return fmt.Errorf("carrier %d for service %q does not exist", services[i].CarrierID, services[i].ServiceCode)
		}

		carrierProvider := normalizePublishedCarrierProviderCode(carrier.Code)
		requestedProvider := normalizePublishedCarrierProviderCode(services[i].ProviderCode)
		code := strings.ToUpper(strings.TrimSpace(services[i].ServiceCode))
		codeProvider := ""
		if strings.HasPrefix(code, "YANWEN:") {
			codeProvider = "YANWEN"
		}
		if requestedProvider != "" && requestedProvider != carrierProvider {
			return fmt.Errorf("provider %q does not match carrier %q for service %q", requestedProvider, carrier.Code, services[i].ServiceCode)
		}
		if codeProvider != "" && codeProvider != carrierProvider {
			return fmt.Errorf("service code %q does not match carrier %q", services[i].ServiceCode, carrier.Code)
		}
		provider := carrierProvider
		if provider == "4PX" {
			needsFpx = true
		}
		if provider == "YANWEN" {
			needsYanwen = true
		}
		services[i].ProviderCode = provider
	}

	fpxByID := map[uint]shipping.FpxChannel{}
	if needsFpx {
		channels, err := s.shippingRepo.FindAllFpxChannelsForEnvironment(shipping.FpxChannelEnvironmentProduction, false)
		if err != nil {
			return fmt.Errorf("load 4PX service collection: %w", err)
		}
		for _, channel := range channels {
			fpxByID[channel.ID] = channel
		}
	}

	yanwenByID := map[uint]YanwenPublishedCollectionReference{}
	if needsYanwen {
		if s.yanwenPublishedCollection == nil {
			return errors.New("Yanwen published collection service is not configured")
		}
		channels, err := s.yanwenPublishedCollection.ListProductionYanwenCollectionReferencesIncludingDisabled()
		if err != nil {
			return fmt.Errorf("load Yanwen service collection: %w", err)
		}
		for _, channel := range channels {
			yanwenByID[channel.ID] = channel
		}
	}

	for i := range services {
		provider := normalizePublishedCarrierProviderCode(services[i].ProviderCode)
		if services[i].FpxChannelID != nil && services[i].YanwenPublishedChannelID != nil {
			return fmt.Errorf("service %q cannot reference both 4PX and Yanwen service collections", services[i].ServiceCode)
		}
		if provider == "YANWEN" {
			if services[i].FpxChannelID != nil {
				return fmt.Errorf("Yanwen service %q cannot reference a 4PX service collection", services[i].ServiceCode)
			}
			if services[i].YanwenPublishedChannelID == nil {
				return fmt.Errorf("select a Yanwen service collection record for service %q", services[i].ServiceCode)
			}
			channel, ok := yanwenByID[*services[i].YanwenPublishedChannelID]
			if !ok {
				return fmt.Errorf("Yanwen service collection record %d is missing or is not in production", *services[i].YanwenPublishedChannelID)
			}
			if !channel.Enabled {
				return fmt.Errorf("Yanwen product %q is disabled in the service collection", channel.ProductCode)
			}
			services[i].YanwenPublishedChannelID = &channel.ID
			services[i].ServiceCode = "YANWEN:" + normalizeYanwenProductCode(channel.ProductCode)
			services[i].ServiceName = channel.DisplayName
			services[i].Countries = channel.Countries
		} else if provider == "4PX" {
			if services[i].YanwenPublishedChannelID != nil {
				return fmt.Errorf("4PX service %q cannot reference a Yanwen service collection", services[i].ServiceCode)
			}
			if services[i].FpxChannelID == nil {
				return fmt.Errorf("select a 4PX service collection record for service %q", services[i].ServiceCode)
			}
			channel, ok := fpxByID[*services[i].FpxChannelID]
			if !ok {
				return fmt.Errorf("4PX service collection record %d is missing or is not in production", *services[i].FpxChannelID)
			}
			if !channel.Enabled {
				return fmt.Errorf("4PX service %q is disabled in the service collection", channel.ServiceCode)
			}
			services[i].FpxChannelID = &channel.ID
			services[i].ServiceCode = channel.ServiceCode
			services[i].ServiceName = channel.DisplayName
			services[i].Countries = channel.Countries
		} else if services[i].FpxChannelID != nil || services[i].YanwenPublishedChannelID != nil {
			return fmt.Errorf("generic carrier service %q cannot reference a 4PX or Yanwen service collection", services[i].ServiceCode)
		}
		services[i].Countries = shipping.NormalizeShippingServiceCollectionCountryCodes(services[i].Countries)
	}
	return nil
}

func normalizePublishedCarrierProviderCode(value string) string {
	provider := strings.ToUpper(strings.TrimSpace(value))
	if provider == "FPX" {
		return "4PX"
	}
	return provider
}

func (s *ShippingService) DeleteTemplate(id uint) error {
	template, err := s.GetTemplate(id)
	if err != nil {
		return err
	}
	if template.IsSystemManaged || template.IsSystemFreeShippingTemplate() {
		return errors.New("system-managed shipping template cannot be deleted")
	}
	return s.shippingRepo.DeleteTemplate(id)
}

func (s *ShippingService) CreateTemplateRule(templateID uint, rule *shipping.ShippingRule) error {
	if err := s.ensureTemplateSupportsShippingRules(templateID); err != nil {
		return err
	}
	rule.TemplateID = templateID
	if err := s.prepareShippingRuleCurrency(templateID, rule); err != nil {
		return err
	}
	return s.shippingRepo.CreateRule(rule)
}

func (s *ShippingService) UpdateTemplateRule(templateID uint, rule *shipping.ShippingRule) error {
	if err := s.ensureTemplateSupportsShippingRules(templateID); err != nil {
		return err
	}
	rule.TemplateID = templateID
	if err := s.prepareShippingRuleCurrency(templateID, rule); err != nil {
		return err
	}
	return s.shippingRepo.UpdateRuleForTemplate(rule)
}

func (s *ShippingService) DeleteTemplateRule(templateID uint, ruleID uint) error {
	if err := s.ensureTemplateSupportsShippingRules(templateID); err != nil {
		return err
	}
	return s.shippingRepo.DeleteRuleForTemplate(templateID, ruleID)
}

func (s *ShippingService) ensureTemplateSupportsShippingRules(templateID uint) error {
	template, err := s.GetTemplate(templateID)
	if err != nil {
		return err
	}
	if template.IsSystemFreeShippingTemplate() {
		return errors.New("system free-shipping template does not support shipping rules")
	}
	return nil
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
