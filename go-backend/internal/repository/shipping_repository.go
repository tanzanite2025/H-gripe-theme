package repository

import (
	"encoding/json"
	"fmt"
	"strings"
	"tanzanite/internal/domain/shipping"
	"time"

	"gorm.io/gorm"
)

type ShippingRepository struct {
	db *gorm.DB
}

type TrackingShipmentFilter struct {
	SyncStatus          string
	RegistrationStatus  string
	TrackingNumber      string
	ProviderCarrierCode string
	Keyword             string
	OrderID             uint
	ProviderID          uint
	CarrierID           uint
	CarrierServiceID    uint
	Enabled             *bool
	DueOnly             bool
	Limit               int
}

func NewShippingRepository(db *gorm.DB) *ShippingRepository {
	return &ShippingRepository{db: db}
}

func (r *ShippingRepository) WithTx(tx *gorm.DB) *ShippingRepository {
	return &ShippingRepository{db: tx}
}

// ShippingTemplate 閻╃鍙ч弬瑙勭《

// FindTemplateByID 閺嶈宓両D閺屻儲澹樺Ο鈩冩緲
func (r *ShippingRepository) FindTemplateByID(id uint) (*shipping.ShippingTemplate, error) {
	var t shipping.ShippingTemplate
	err := r.db.Preload("Rules", func(db *gorm.DB) *gorm.DB {
		return db.Order("min_value ASC, id ASC")
	}).First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// FindAllTemplates 閺屻儲澹橀幍鈧張澶嬆侀弶?
func (r *ShippingRepository) FindAllTemplates() ([]shipping.ShippingTemplate, error) {
	var templates []shipping.ShippingTemplate
	err := r.db.Preload("Rules", func(db *gorm.DB) *gorm.DB {
		return db.Order("min_value ASC, id ASC")
	}).Find(&templates).Error
	return templates, err
}

func (r *ShippingRepository) CreateTemplateWithRules(template *shipping.ShippingTemplate, rules []shipping.ShippingRule) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		template.Rules = nil
		if err := tx.Create(template).Error; err != nil {
			return err
		}

		if len(rules) > 0 {
			for i := range rules {
				rules[i].ID = 0
				rules[i].TemplateID = template.ID
			}
			if err := tx.Create(&rules).Error; err != nil {
				return err
			}
		}

		return tx.Preload("Rules").First(template, template.ID).Error
	})
}

func (r *ShippingRepository) UpdateTemplateWithRules(template *shipping.ShippingTemplate, rules []shipping.ShippingRule) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"name":           template.Name,
			"type":           template.Type,
			"free_shipping":  template.FreeShipping,
			"free_threshold": template.FreeThreshold,
			"default_fee":    template.DefaultFee,
			"description":    template.Description,
			"enabled":        template.Enabled,
		}
		if err := tx.Model(&shipping.ShippingTemplate{}).Where("id = ?", template.ID).Updates(updates).Error; err != nil {
			return err
		}

		if err := tx.Where("template_id = ?", template.ID).Delete(&shipping.ShippingRule{}).Error; err != nil {
			return err
		}

		if len(rules) > 0 {
			for i := range rules {
				rules[i].ID = 0
				rules[i].TemplateID = template.ID
			}
			if err := tx.Create(&rules).Error; err != nil {
				return err
			}
		}

		return tx.Preload("Rules").First(template, template.ID).Error
	})
}

func (r *ShippingRepository) DeleteTemplate(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("template_id = ?", id).Delete(&shipping.ShippingRule{}).Error; err != nil {
			return err
		}
		if err := tx.Where("template_id = ?", id).Delete(&shipping.ShippingTemplateBinding{}).Error; err != nil {
			return err
		}
		return tx.Delete(&shipping.ShippingTemplate{}, id).Error
	})
}

func (r *ShippingRepository) FindAllTemplateBindings() ([]shipping.ShippingTemplateBinding, error) {
	var bindings []shipping.ShippingTemplateBinding
	err := r.db.Preload("Template").Order("priority DESC").Order("id DESC").Find(&bindings).Error
	return bindings, err
}

func (r *ShippingRepository) FindEnabledTemplateBindingsWithTemplates() ([]shipping.ShippingTemplateBinding, error) {
	var bindings []shipping.ShippingTemplateBinding
	err := r.db.
		Preload("Template.Rules", func(db *gorm.DB) *gorm.DB {
			return db.Order("min_value ASC, id ASC")
		}).
		Preload("Template").
		Where("enabled = ?", true).
		Order("priority DESC").
		Order("id DESC").
		Find(&bindings).Error
	return bindings, err
}

func (r *ShippingRepository) FindTemplateBindingByID(id uint) (*shipping.ShippingTemplateBinding, error) {
	var binding shipping.ShippingTemplateBinding
	err := r.db.Preload("Template").First(&binding, id).Error
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

func (r *ShippingRepository) CreateTemplateBinding(binding *shipping.ShippingTemplateBinding) error {
	return r.db.Create(binding).Error
}

func (r *ShippingRepository) UpdateTemplateBinding(binding *shipping.ShippingTemplateBinding) error {
	updates := map[string]interface{}{
		"template_id":     binding.TemplateID,
		"scope":           binding.Scope,
		"product_type_id": binding.ProductTypeID,
		"product_id":      binding.ProductID,
		"variant_id":      binding.VariantID,
		"priority":        binding.Priority,
		"enabled":         binding.Enabled,
	}
	return r.db.Model(&shipping.ShippingTemplateBinding{}).Where("id = ?", binding.ID).Updates(updates).Error
}

func (r *ShippingRepository) DeleteTemplateBinding(id uint) error {
	return r.db.Delete(&shipping.ShippingTemplateBinding{}, id).Error
}

// ShippingRule 閻╃鍙ч弬瑙勭《

// CreateRule 閸掓稑缂撴潻鎰瀭鐟欏嫬鍨?
func (r *ShippingRepository) CreateRule(rule *shipping.ShippingRule) error {
	return r.db.Create(rule).Error
}

// FindRulesByTemplateID 閺嶈宓佸Ο鈩冩緲ID閺屻儲澹樼憴鍕灟
func (r *ShippingRepository) FindRulesByTemplateID(templateID uint) ([]shipping.ShippingRule, error) {
	var rules []shipping.ShippingRule
	err := r.db.Where("template_id = ?", templateID).Order("min_value ASC").Find(&rules).Error
	return rules, err
}

// UpdateRule 閺囧瓨鏌婄憴鍕灟
func (r *ShippingRepository) UpdateRule(rule *shipping.ShippingRule) error {
	return r.db.Save(rule).Error
}

func (r *ShippingRepository) UpdateRuleForTemplate(rule *shipping.ShippingRule) error {
	updates := map[string]interface{}{
		"region":      rule.Region,
		"min_value":   rule.MinValue,
		"max_value":   rule.MaxValue,
		"fee":         rule.Fee,
		"additional":  rule.Additional,
		"template_id": rule.TemplateID,
	}
	return r.db.Model(&shipping.ShippingRule{}).
		Where("id = ? AND template_id = ?", rule.ID, rule.TemplateID).
		Updates(updates).Error
}

// DeleteRule 閸掔娀娅庣憴鍕灟
func (r *ShippingRepository) DeleteRule(id uint) error {
	return r.db.Delete(&shipping.ShippingRule{}, id).Error
}

func (r *ShippingRepository) DeleteRuleForTemplate(templateID uint, ruleID uint) error {
	return r.db.Where("id = ? AND template_id = ?", ruleID, templateID).Delete(&shipping.ShippingRule{}).Error
}

// Carrier 閻╃鍙ч弬瑙勭《

// CreateCarrier 閸掓稑缂撻悧鈺傜ウ閸忣剙寰?
func (r *ShippingRepository) CreateCarrier(c *shipping.Carrier) error {
	return r.db.Create(c).Error
}

// FindCarrierByID 閺嶈宓両D閺屻儲澹橀悧鈺傜ウ閸忣剙寰?
func (r *ShippingRepository) FindCarrierByID(id uint) (*shipping.Carrier, error) {
	var c shipping.Carrier
	err := r.db.First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// FindCarrierByCode 閺嶈宓佹禒锝囩垳閺屻儲澹橀悧鈺傜ウ閸忣剙寰?
func (r *ShippingRepository) FindCarrierByCode(code string) (*shipping.Carrier, error) {
	var c shipping.Carrier
	err := r.db.Where("code = ?", code).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// FindAllCarriers 閺屻儲澹橀幍鈧張澶屽⒖濞翠礁鍙曢崣?
func (r *ShippingRepository) FindAllCarriers(enabledOnly bool) ([]shipping.Carrier, error) {
	var carriers []shipping.Carrier
	query := r.db.Order("name ASC")

	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}

	err := query.Find(&carriers).Error
	return carriers, err
}

// UpdateCarrier 閺囧瓨鏌婇悧鈺傜ウ閸忣剙寰?
func (r *ShippingRepository) UpdateCarrier(c *shipping.Carrier) error {
	return r.db.Save(c).Error
}

// DeleteCarrier 閸掔娀娅庨悧鈺傜ウ閸忣剙寰?
func (r *ShippingRepository) DeleteCarrier(id uint) error {
	return r.db.Delete(&shipping.Carrier{}, id).Error
}

func (r *ShippingRepository) FindAllTrackingProviderConfigs(enabledOnly bool) ([]shipping.TrackingProviderConfig, error) {
	var providers []shipping.TrackingProviderConfig
	query := r.db.
		Order("sort_order ASC").
		Order("provider_name ASC")

	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}

	err := query.Find(&providers).Error
	return providers, err
}

func (r *ShippingRepository) FindTrackingProviderConfigByID(id uint) (*shipping.TrackingProviderConfig, error) {
	var provider shipping.TrackingProviderConfig
	err := r.db.First(&provider, id).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *ShippingRepository) FindTrackingProviderConfigByCode(providerCode string) (*shipping.TrackingProviderConfig, error) {
	var provider shipping.TrackingProviderConfig
	err := r.db.
		Where("LOWER(provider_code) = LOWER(?)", strings.TrimSpace(providerCode)).
		Order("enabled DESC").
		Order("id ASC").
		First(&provider).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *ShippingRepository) CreateTrackingProviderConfig(provider *shipping.TrackingProviderConfig) error {
	return r.db.Create(provider).Error
}

func (r *ShippingRepository) UpdateTrackingProviderConfig(provider *shipping.TrackingProviderConfig) error {
	updates := map[string]interface{}{
		"provider_code":            provider.ProviderCode,
		"provider_name":            provider.ProviderName,
		"environment":              provider.Environment,
		"base_url":                 provider.BaseURL,
		"api_key":                  provider.APIKey,
		"webhook_secret":           provider.WebhookSecret,
		"webhook_enabled":          provider.WebhookEnabled,
		"auto_register":            provider.AutoRegister,
		"polling_enabled":          provider.PollingEnabled,
		"polling_interval_minutes": provider.PollingIntervalMinutes,
		"request_timeout_seconds":  provider.RequestTimeoutSeconds,
		"enabled":                  provider.Enabled,
		"sort_order":               provider.SortOrder,
		"description":              provider.Description,
	}
	return r.db.Model(&shipping.TrackingProviderConfig{}).Where("id = ?", provider.ID).Updates(updates).Error
}

func (r *ShippingRepository) DeleteTrackingProviderConfig(id uint) error {
	return r.db.Delete(&shipping.TrackingProviderConfig{}, id).Error
}

func (r *ShippingRepository) FindAllTrackingCarrierMappings(enabledOnly bool) ([]shipping.TrackingCarrierMapping, error) {
	var mappings []shipping.TrackingCarrierMapping
	query := r.db.
		Preload("Provider").
		Preload("Carrier").
		Preload("CarrierService").
		Preload("CarrierService.Carrier").
		Order("provider_id ASC").
		Order("priority DESC").
		Order("id DESC")

	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}

	err := query.Find(&mappings).Error
	return mappings, err
}

func (r *ShippingRepository) FindTrackingCarrierMappingByID(id uint) (*shipping.TrackingCarrierMapping, error) {
	var mapping shipping.TrackingCarrierMapping
	err := r.db.
		Preload("Provider").
		Preload("Carrier").
		Preload("CarrierService").
		Preload("CarrierService.Carrier").
		First(&mapping, id).Error
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *ShippingRepository) FindEnabledTrackingCarrierMappingByCarrierService(providerID uint, carrierServiceID uint) (*shipping.TrackingCarrierMapping, error) {
	var mapping shipping.TrackingCarrierMapping
	err := r.db.
		Preload("Provider").
		Preload("CarrierService").
		Preload("CarrierService.Carrier").
		Where("provider_id = ? AND scope = ? AND carrier_service_id = ? AND enabled = ?", providerID, "carrier_service", carrierServiceID, true).
		Order("priority DESC").
		Order("id DESC").
		First(&mapping).Error
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *ShippingRepository) FindEnabledTrackingCarrierMappingByCarrier(providerID uint, carrierID uint) (*shipping.TrackingCarrierMapping, error) {
	var mapping shipping.TrackingCarrierMapping
	err := r.db.
		Preload("Provider").
		Preload("Carrier").
		Where("provider_id = ? AND scope = ? AND carrier_id = ? AND enabled = ?", providerID, "carrier", carrierID, true).
		Order("priority DESC").
		Order("id DESC").
		First(&mapping).Error
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *ShippingRepository) CreateTrackingCarrierMapping(mapping *shipping.TrackingCarrierMapping) error {
	return r.db.Create(mapping).Error
}

func (r *ShippingRepository) UpdateTrackingCarrierMapping(mapping *shipping.TrackingCarrierMapping) error {
	updates := map[string]interface{}{
		"provider_id":           mapping.ProviderID,
		"scope":                 mapping.Scope,
		"carrier_id":            mapping.CarrierID,
		"carrier_service_id":    mapping.CarrierServiceID,
		"provider_carrier_code": mapping.ProviderCarrierCode,
		"provider_carrier_name": mapping.ProviderCarrierName,
		"enabled":               mapping.Enabled,
		"priority":              mapping.Priority,
		"description":           mapping.Description,
	}
	return r.db.Model(&shipping.TrackingCarrierMapping{}).Where("id = ?", mapping.ID).Updates(updates).Error
}

func (r *ShippingRepository) DeleteTrackingCarrierMapping(id uint) error {
	return r.db.Delete(&shipping.TrackingCarrierMapping{}, id).Error
}

func (r *ShippingRepository) FindTrackingShipmentByOrderID(orderID uint) (*shipping.TrackingShipment, error) {
	var shipment shipping.TrackingShipment
	err := r.db.
		Preload("Provider").
		Preload("Carrier").
		Preload("CarrierService").
		Preload("Mapping").
		Where("order_id = ?", orderID).
		First(&shipment).Error
	if err != nil {
		return nil, err
	}
	return &shipment, nil
}

func (r *ShippingRepository) FindTrackingShipmentByProviderTrackingNumber(providerID uint, trackingNumber string, providerCarrierCode string) (*shipping.TrackingShipment, error) {
	var shipment shipping.TrackingShipment
	baseQuery := func() *gorm.DB {
		return r.db.
			Preload("Provider").
			Preload("Carrier").
			Preload("CarrierService").
			Preload("Mapping").
			Where("tracking_provider_id = ? AND tracking_number = ?", providerID, strings.TrimSpace(trackingNumber))
	}

	providerCarrierCode = strings.TrimSpace(providerCarrierCode)
	if providerCarrierCode != "" {
		err := baseQuery().
			Where("provider_carrier_code = ?", providerCarrierCode).
			First(&shipment).Error
		if err == nil {
			return &shipment, nil
		}
		if !IsRecordNotFound(err) {
			return nil, err
		}
	}

	var shipments []shipping.TrackingShipment
	err := baseQuery().
		Order("id ASC").
		Limit(2).
		Find(&shipments).Error
	if err != nil {
		return nil, err
	}
	if len(shipments) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if len(shipments) > 1 {
		return nil, fmt.Errorf("multiple tracking shipments match provider ID %d and tracking number %s", providerID, strings.TrimSpace(trackingNumber))
	}
	return &shipments[0], nil
}

func (r *ShippingRepository) FindAllTrackingShipments(filter TrackingShipmentFilter) ([]shipping.TrackingShipment, error) {
	var shipments []shipping.TrackingShipment
	query := r.db.
		Preload("Provider").
		Preload("Carrier").
		Preload("CarrierService").
		Preload("Mapping").
		Order("updated_at DESC").
		Order("id DESC")

	syncStatus := strings.TrimSpace(strings.ToLower(filter.SyncStatus))
	if syncStatus != "" && syncStatus != "all" {
		query = query.Where("sync_status = ?", syncStatus)
	}

	registrationStatus := strings.TrimSpace(strings.ToLower(filter.RegistrationStatus))
	if registrationStatus != "" && registrationStatus != "all" {
		query = query.Where("registration_status = ?", registrationStatus)
	}

	if filter.OrderID > 0 {
		query = query.Where("order_id = ?", filter.OrderID)
	}
	if filter.ProviderID > 0 {
		query = query.Where("tracking_provider_id = ?", filter.ProviderID)
	}
	if filter.CarrierID > 0 {
		query = query.Where("carrier_id = ?", filter.CarrierID)
	}
	if filter.CarrierServiceID > 0 {
		query = query.Where("carrier_service_id = ?", filter.CarrierServiceID)
	}
	if filter.Enabled != nil {
		query = query.Where("enabled = ?", *filter.Enabled)
	}

	trackingNumber := strings.TrimSpace(filter.TrackingNumber)
	if trackingNumber != "" {
		query = query.Where("tracking_number = ?", trackingNumber)
	}

	providerCarrierCode := strings.TrimSpace(filter.ProviderCarrierCode)
	if providerCarrierCode != "" {
		query = query.Where("provider_carrier_code = ?", providerCarrierCode)
	}

	keyword := strings.TrimSpace(filter.Keyword)
	if keyword != "" {
		like := "%" + strings.ToLower(keyword) + "%"
		query = query.Where(
			"LOWER(tracking_number) LIKE ? OR LOWER(provider_carrier_code) LIKE ? OR CAST(order_id AS TEXT) LIKE ? OR LOWER(last_error) LIKE ?",
			like,
			like,
			"%"+keyword+"%",
			like,
		)
	}

	if filter.DueOnly {
		now := time.Now()
		query = query.Where(
			"(sync_status = ? OR (sync_status = ? AND (next_sync_at IS NULL OR next_sync_at <= ?)) OR (sync_status = ? AND next_sync_at <= ?))",
			"pending",
			"failed",
			now,
			"synced",
			now,
		)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	err := query.Find(&shipments).Error
	return shipments, err
}

func (r *ShippingRepository) FindDueTrackingShipments(limit int, now time.Time) ([]shipping.TrackingShipment, error) {
	var shipments []shipping.TrackingShipment
	query := r.db.
		Preload("Provider").
		Preload("Carrier").
		Preload("CarrierService").
		Preload("Mapping").
		Where("enabled = ?", true).
		Where(
			"(sync_status = ? OR (sync_status = ? AND (next_sync_at IS NULL OR next_sync_at <= ?)) OR (sync_status = ? AND next_sync_at <= ?))",
			"pending",
			"failed",
			now,
			"synced",
			now,
		).
		Order("COALESCE(next_sync_at, created_at) ASC").
		Order("id ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&shipments).Error
	return shipments, err
}

func (r *ShippingRepository) UpsertTrackingShipment(shipment *shipping.TrackingShipment) error {
	existing, err := r.FindTrackingShipmentByOrderID(shipment.OrderID)
	if err != nil {
		if IsRecordNotFound(err) {
			return r.db.Create(shipment).Error
		}
		return err
	}

	shipment.ID = existing.ID
	sourceUnchanged := trackingShipmentExternalSourceUnchanged(existing, shipment)
	updates := map[string]interface{}{
		"tracking_provider_id":        shipment.TrackingProviderID,
		"tracking_number":             shipment.TrackingNumber,
		"provider_carrier_code":       shipment.ProviderCarrierCode,
		"carrier_id":                  shipment.CarrierID,
		"carrier_service_id":          shipment.CarrierServiceID,
		"tracking_carrier_mapping_id": shipment.TrackingCarrierMappingID,
		"enabled":                     shipment.Enabled,
		"updated_at":                  time.Now(),
	}
	if !sourceUnchanged {
		updates["registration_status"] = shipment.RegistrationStatus
		updates["sync_status"] = shipment.SyncStatus
		updates["event_count"] = shipment.EventCount
		updates["last_event_at"] = shipment.LastEventAt
		updates["last_synced_at"] = shipment.LastSyncedAt
		updates["next_sync_at"] = shipment.NextSyncAt
		updates["last_error"] = shipment.LastError
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&shipping.TrackingShipment{}).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
			return err
		}
		if !sourceUnchanged {
			return tx.Where("order_id = ?", existing.OrderID).Delete(&shipping.TrackingEvent{}).Error
		}
		return nil
	})
}

func trackingShipmentExternalSourceUnchanged(existing *shipping.TrackingShipment, next *shipping.TrackingShipment) bool {
	if existing == nil || next == nil {
		return false
	}
	return existing.TrackingProviderID == next.TrackingProviderID &&
		strings.TrimSpace(existing.TrackingNumber) == strings.TrimSpace(next.TrackingNumber) &&
		strings.TrimSpace(existing.ProviderCarrierCode) == strings.TrimSpace(next.ProviderCarrierCode)
}

func (r *ShippingRepository) UpdateTrackingShipmentSyncing(orderID uint) error {
	now := time.Now()
	return r.db.Model(&shipping.TrackingShipment{}).
		Where("order_id = ?", orderID).
		Updates(map[string]interface{}{
			"sync_status": "syncing",
			"last_error":  "",
			"updated_at":  now,
		}).Error
}

func (r *ShippingRepository) UpdateTrackingShipmentSyncSuccess(orderID uint, eventCount int, lastEventAt *time.Time, nextSyncAt *time.Time) error {
	now := time.Now()
	return r.db.Model(&shipping.TrackingShipment{}).
		Where("order_id = ?", orderID).
		Updates(map[string]interface{}{
			"registration_status": "registered",
			"sync_status":         "synced",
			"event_count":         eventCount,
			"last_event_at":       lastEventAt,
			"last_synced_at":      &now,
			"next_sync_at":        nextSyncAt,
			"last_error":          "",
			"updated_at":          now,
		}).Error
}

func (r *ShippingRepository) UpdateTrackingShipmentRegistrationStatus(orderID uint, status string, lastError string) error {
	now := time.Now()
	return r.db.Model(&shipping.TrackingShipment{}).
		Where("order_id = ?", orderID).
		Updates(map[string]interface{}{
			"registration_status": status,
			"last_error":          lastError,
			"updated_at":          now,
		}).Error
}

func (r *ShippingRepository) UpdateTrackingShipmentSyncFailure(orderID uint, lastError string, nextSyncAt *time.Time) error {
	now := time.Now()
	return r.db.Model(&shipping.TrackingShipment{}).
		Where("order_id = ?", orderID).
		Updates(map[string]interface{}{
			"sync_status":    "failed",
			"last_synced_at": &now,
			"next_sync_at":   nextSyncAt,
			"last_error":     lastError,
			"updated_at":     now,
		}).Error
}

func (r *ShippingRepository) FindAllCarrierServices(enabledOnly bool) ([]shipping.CarrierService, error) {
	var services []shipping.CarrierService
	query := r.db.
		Preload("Carrier").
		Preload("Template").
		Order("carrier_id ASC").
		Order("sort_order ASC").
		Order("service_name ASC")

	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}

	err := query.Find(&services).Error
	return services, err
}

func (r *ShippingRepository) FindEnabledCarrierServicesWithTemplates() ([]shipping.CarrierService, error) {
	var services []shipping.CarrierService
	err := r.db.
		Preload("Carrier").
		Preload("Template.Rules", func(db *gorm.DB) *gorm.DB {
			return db.Order("min_value ASC, id ASC")
		}).
		Preload("Template").
		Where("enabled = ?", true).
		Where("template_id IS NOT NULL").
		Order("sort_order ASC").
		Order("id ASC").
		Find(&services).Error
	return services, err
}

func (r *ShippingRepository) FindCarrierServiceByID(id uint) (*shipping.CarrierService, error) {
	var service shipping.CarrierService
	err := r.db.
		Preload("Carrier").
		Preload("Template").
		First(&service, id).Error
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *ShippingRepository) CreateCarrierService(service *shipping.CarrierService) error {
	return r.db.Create(service).Error
}

func (r *ShippingRepository) UpdateCarrierService(service *shipping.CarrierService) error {
	updates := map[string]interface{}{
		"carrier_id":              service.CarrierID,
		"template_id":             service.TemplateID,
		"service_code":            service.ServiceCode,
		"service_name":            service.ServiceName,
		"route_name":              service.RouteName,
		"countries":               service.Countries,
		"currency":                service.Currency,
		"billing_mode":            service.BillingMode,
		"first_weight_grams":      service.FirstWeightGrams,
		"additional_weight_grams": service.AdditionalWeightGrams,
		"min_charge_weight_grams": service.MinChargeWeightGrams,
		"volumetric_divisor":      service.VolumetricDivisor,
		"fuel_surcharge_percent":  service.FuelSurchargePercent,
		"remote_surcharge":        service.RemoteSurcharge,
		"eta_min_days":            service.EtaMinDays,
		"eta_max_days":            service.EtaMaxDays,
		"enabled":                 service.Enabled,
		"sort_order":              service.SortOrder,
		"description":             service.Description,
	}
	return r.db.Model(&shipping.CarrierService{}).Where("id = ?", service.ID).Updates(updates).Error
}

func (r *ShippingRepository) DeleteCarrierService(id uint) error {
	return r.db.Delete(&shipping.CarrierService{}, id).Error
}

// TrackingEvent 閻╃鍙ч弬瑙勭《

// FindTrackingEventsByOrderID 閺嶈宓佺拋銏犲礋ID閺屻儲澹樻潻鍊熼嚋娴滃娆?
func (r *ShippingRepository) FindTrackingEventsByOrderID(orderID uint) ([]shipping.TrackingEvent, error) {
	var events []shipping.TrackingEvent
	err := r.db.Where("order_id = ?", orderID).Order("event_time DESC").Find(&events).Error
	return events, err
}

// FindTrackingEventsByTrackingNumber 閺嶈宓佹潻鍊熼嚋閸欓攱鐓￠幍鍙ョ皑娴?
func (r *ShippingRepository) FindTrackingEventsByTrackingNumber(trackingNumber string) ([]shipping.TrackingEvent, error) {
	var events []shipping.TrackingEvent
	err := r.db.Where("tracking_number = ?", trackingNumber).Order("event_time DESC").Find(&events).Error
	return events, err
}

func (r *ShippingRepository) ReplaceTrackingEvents(orderID uint, trackingNumber string, events []shipping.TrackingEvent) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ? AND tracking_number = ?", orderID, trackingNumber).Delete(&shipping.TrackingEvent{}).Error; err != nil {
			return err
		}
		if len(events) == 0 {
			return nil
		}
		return tx.Create(&events).Error
	})
}

// ShippingZone 閻╃鍙ч弬瑙勭《

// FindZoneByID 閺嶈宓両D閺屻儲澹橀崠鍝勭厵
func (r *ShippingRepository) FindZoneByID(id uint) (*shipping.ShippingZone, error) {
	var z shipping.ShippingZone
	err := r.db.First(&z, id).Error
	if err != nil {
		return nil, err
	}
	return &z, nil
}

// FindAllZones 閺屻儲澹橀幍鈧張澶婂隘閸?
func (r *ShippingRepository) FindAllZones() ([]shipping.ShippingZone, error) {
	var zones []shipping.ShippingZone
	err := r.db.Order("name ASC").Find(&zones).Error
	return zones, err
}

func (r *ShippingRepository) CreateZone(zone *shipping.ShippingZone) error {
	return r.db.Create(zone).Error
}

func (r *ShippingRepository) UpdateZone(zone *shipping.ShippingZone) error {
	updates := map[string]interface{}{
		"name":         zone.Name,
		"countries":    zone.Countries,
		"states":       zone.States,
		"postal_codes": zone.PostalCodes,
		"enabled":      zone.Enabled,
	}
	return r.db.Model(&shipping.ShippingZone{}).Where("id = ?", zone.ID).Updates(updates).Error
}

func (r *ShippingRepository) DeleteZone(id uint) error {
	return r.db.Delete(&shipping.ShippingZone{}, id).Error
}

// FindZoneByCountry 閺嶈宓侀崶钘夘啀閺屻儲澹橀崠鍝勭厵
func (r *ShippingRepository) FindZoneByCountry(country string) (*shipping.ShippingZone, error) {
	var zones []shipping.ShippingZone
	if err := r.db.Where("enabled = ?", true).Order("name ASC").Find(&zones).Error; err != nil {
		return nil, err
	}

	normalizedCountry := strings.ToUpper(strings.TrimSpace(country))
	for i := range zones {
		if countryMatchesZone(normalizedCountry, zones[i].Countries) {
			return &zones[i], nil
		}
	}

	return nil, gorm.ErrRecordNotFound
}

func countryMatchesZone(country string, countriesValue string) bool {
	if country == "" || strings.TrimSpace(countriesValue) == "" {
		return false
	}

	var countries []string
	if err := json.Unmarshal([]byte(countriesValue), &countries); err == nil {
		for _, candidate := range countries {
			if strings.ToUpper(strings.TrimSpace(candidate)) == country {
				return true
			}
		}
		return false
	}

	for _, candidate := range strings.FieldsFunc(countriesValue, func(r rune) bool {
		return r == ',' || r == ';' || r == '|' || r == '\n' || r == '\r' || r == '\t'
	}) {
		if strings.ToUpper(strings.TrimSpace(candidate)) == country {
			return true
		}
	}
	return false
}

// FindPackagingRuleByID 閺嶈宓両D閺屻儲澹橀崠鍛邦棅鐟欏嫭鐗哥憴鍕灟
func (r *ShippingRepository) FindPackagingRuleByID(id uint) (*shipping.PackagingRule, error) {
	var pr shipping.PackagingRule
	err := r.db.Preload("Applies").First(&pr, id).Error
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

// FindAllPackagingRules 閼惧嘲褰囬幍鈧張澶婂瘶鐟佸懓顫夐弽鑹邦潐閸?
func (r *ShippingRepository) FindAllPackagingRules() ([]shipping.PackagingRule, error) {
	var rules []shipping.PackagingRule
	err := r.db.Preload("Applies").Find(&rules).Error
	return rules, err
}

// CreatePackagingRule 閸掓稑缂撻崠鍛邦棅鐟欏嫭鐗哥憴鍕灟
func (r *ShippingRepository) CreatePackagingRule(rule *shipping.PackagingRule) error {
	return r.db.Create(rule).Error
}

// UpdatePackagingRule 閺囧瓨鏌婇崠鍛邦棅鐟欏嫭鐗哥憴鍕灟
func (r *ShippingRepository) UpdatePackagingRule(rule *shipping.PackagingRule) error {
	return r.db.Save(rule).Error
}

// DeletePackagingRule 閸掔娀娅庨崠鍛邦棅鐟欏嫭鐗哥憴鍕灟
func (r *ShippingRepository) DeletePackagingRule(id uint) error {
	// 閸忓牆鍨归梽銈呯安閻劌鍙ч懕鏃傛畱閺夛紕娲?
	if err := r.db.Where("rule_id = ?", id).Delete(&shipping.PackagingRuleApply{}).Error; err != nil {
		return err
	}
	return r.db.Delete(&shipping.PackagingRule{}, id).Error
}

// CreatePackagingRuleApply 婢х偛濮為崠鍛邦棅鐟欏嫬鍨惃鍕安閻劋楠囬崫浣筋唶瑜?
func (r *ShippingRepository) CreatePackagingRuleApply(apply *shipping.PackagingRuleApply) error {
	return r.db.Create(apply).Error
}

func (r *ShippingRepository) FindPackagingRuleApply(ruleID uint, productID uint) (*shipping.PackagingRuleApply, error) {
	var apply shipping.PackagingRuleApply
	err := r.db.Where("rule_id = ? AND product_id = ?", ruleID, productID).First(&apply).Error
	if err != nil {
		return nil, err
	}
	return &apply, nil
}

func (r *ShippingRepository) FindPackagingRuleApplyByProductID(productID uint) (*shipping.PackagingRuleApply, error) {
	var apply shipping.PackagingRuleApply
	err := r.db.Where("product_id = ?", productID).First(&apply).Error
	if err != nil {
		return nil, err
	}
	return &apply, nil
}

// DeletePackagingRuleApply 閸掔娀娅庨崠鍛邦棅鐟欏嫬鍨惃鍕安閻劋楠囬崫浣筋唶瑜?
func (r *ShippingRepository) DeletePackagingRuleApply(id uint) error {
	return r.db.Delete(&shipping.PackagingRuleApply{}, id).Error
}

func (r *ShippingRepository) FindActivePackagingRulesByProductIDs(productIDs []uint) (map[uint]*shipping.PackagingRule, error) {
	rulesByProduct := make(map[uint]*shipping.PackagingRule)
	if len(productIDs) == 0 {
		return rulesByProduct, nil
	}

	var applies []shipping.PackagingRuleApply
	err := r.db.
		Preload("Rule").
		Joins("JOIN shipping_packaging_rules ON shipping_packaging_rules.id = shipping_packaging_rule_applies.rule_id").
		Where("shipping_packaging_rule_applies.product_id IN ? AND shipping_packaging_rules.is_active = ?", productIDs, true).
		Order("shipping_packaging_rule_applies.product_id ASC, shipping_packaging_rule_applies.id DESC").
		Find(&applies).Error
	if err != nil {
		return nil, err
	}

	for i := range applies {
		apply := &applies[i]
		if apply.Rule == nil {
			continue
		}
		if _, exists := rulesByProduct[apply.ProductID]; exists {
			return nil, fmt.Errorf("product ID %d has multiple active packaging rules", apply.ProductID)
		}
		rulesByProduct[apply.ProductID] = apply.Rule
	}

	return rulesByProduct, nil
}

// FindPackagingRulesByProductID 閺嶈宓佹禍褍鎼D閺屻儲澹橀崠褰掑帳閻ㄥ嫭绺哄ú璇插瘶鐟佸懓顫夐弽鑹邦潐閸?
func (r *ShippingRepository) FindPackagingRulesByProductID(productID uint) ([]shipping.PackagingRule, error) {
	var rules []shipping.PackagingRule
	err := r.db.Joins("JOIN shipping_packaging_rule_applies ON shipping_packaging_rule_applies.rule_id = shipping_packaging_rules.id").
		Where("shipping_packaging_rule_applies.product_id = ? AND shipping_packaging_rules.is_active = ?", productID, true).
		Find(&rules).Error
	return rules, err
}
