package repository

import (
	"commerce-platform/internal/domain/shipping"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ShippingRepository struct {
	db *gorm.DB
}

var (
	ErrTrackingSyncLeaseUnavailable = errors.New("tracking shipment sync lease is already held")
	ErrTrackingSyncLeaseLost        = errors.New("tracking shipment sync lease was lost")
)

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

// attachShippingDisplayPriceSnapshots hydrates the storefront read model in a
// separate query. Transactional shipping rows never carry converted display
// prices; stale snapshots are suppressed by comparing their source fingerprint
// with the current template/rule amounts.
func (r *ShippingRepository) attachShippingDisplayPriceSnapshots(templates []shipping.ShippingTemplate) error {
	if len(templates) == 0 {
		return nil
	}
	for i := range templates {
		templates[i].DisplayPriceData = datatypes.JSON([]byte("{}"))
		templates[i].DisplayPriceSnapshot = nil
		for j := range templates[i].Rules {
			templates[i].Rules[j].DisplayPriceData = datatypes.JSON([]byte("{}"))
			templates[i].Rules[j].DisplayPriceSnapshot = nil
		}
	}
	if !r.db.Migrator().HasTable(&shipping.ShippingDisplayPriceSnapshot{}) {
		return nil
	}

	templateIDs := make([]uint, 0, len(templates))
	templatesByID := make(map[uint]*shipping.ShippingTemplate, len(templates))
	rulesByID := make(map[uint]*shipping.ShippingRule)
	for i := range templates {
		template := &templates[i]
		templateIDs = append(templateIDs, template.ID)
		templatesByID[template.ID] = template
		for j := range template.Rules {
			rulesByID[template.Rules[j].ID] = &template.Rules[j]
		}
	}

	var snapshots []shipping.ShippingDisplayPriceSnapshot
	if err := r.db.Where("template_id IN ?", templateIDs).Find(&snapshots).Error; err != nil {
		return err
	}
	for i := range snapshots {
		snapshot := &snapshots[i]
		if snapshot.RuleID == nil {
			template := templatesByID[snapshot.TemplateID]
			if template != nil && shippingTemplateDisplaySnapshotMatchesSource(snapshot, template) {
				template.DisplayPriceData = append(datatypes.JSON(nil), snapshot.DisplayPriceData...)
				template.DisplayPriceSnapshot = snapshot
			}
			continue
		}
		rule := rulesByID[*snapshot.RuleID]
		if rule != nil && shippingRuleDisplaySnapshotMatchesSource(snapshot, rule) {
			rule.DisplayPriceData = append(datatypes.JSON(nil), snapshot.DisplayPriceData...)
			rule.DisplayPriceSnapshot = snapshot
		}
	}
	return nil
}

func shippingTemplateDisplaySnapshotMatchesSource(snapshot *shipping.ShippingDisplayPriceSnapshot, template *shipping.ShippingTemplate) bool {
	return snapshot != nil && template != nil &&
		strings.EqualFold(strings.TrimSpace(snapshot.SourceCurrency), strings.TrimSpace(template.Currency)) &&
		int64PointerEqual(snapshot.SourceDefaultFeeMinor, template.DefaultFeeMinor) &&
		int64PointerEqual(snapshot.SourceFreeThresholdMinor, template.FreeThresholdMinor)
}

func shippingRuleDisplaySnapshotMatchesSource(snapshot *shipping.ShippingDisplayPriceSnapshot, rule *shipping.ShippingRule) bool {
	return snapshot != nil && rule != nil &&
		strings.EqualFold(strings.TrimSpace(snapshot.SourceCurrency), strings.TrimSpace(rule.Currency)) &&
		int64PointerEqual(snapshot.SourceMinValueMinor, rule.MinValueMinor) &&
		int64PointerEqual(snapshot.SourceMaxValueMinor, rule.MaxValueMinor) &&
		int64PointerEqual(snapshot.SourceFeeMinor, rule.FeeMinor) &&
		int64PointerEqual(snapshot.SourceAdditionalMinor, rule.AdditionalMinor)
}

func int64PointerEqual(value *int64, expected int64) bool {
	return value != nil && *value == expected
}

// ShippingTemplate 閻╃鍙ч弬瑙勭《

// FindTemplateByID 閺嶈宓両D閺屻儲澹樺Ο鈩冩緲
func (r *ShippingRepository) FindTemplateByID(id uint) (*shipping.ShippingTemplate, error) {
	var t shipping.ShippingTemplate
	err := r.db.Preload("Rules", func(db *gorm.DB) *gorm.DB {
		return db.Order("min_value_minor ASC, min_value ASC, id ASC")
	}).First(&t, id).Error
	if err != nil {
		return nil, err
	}
	hydrated := []shipping.ShippingTemplate{t}
	if err := r.attachShippingDisplayPriceSnapshots(hydrated); err != nil {
		return nil, err
	}
	t = hydrated[0]
	return &t, nil
}

func (r *ShippingRepository) FindTemplatesByIDs(ids []uint) (map[uint]*shipping.ShippingTemplate, error) {
	templatesByID := make(map[uint]*shipping.ShippingTemplate, len(ids))
	if len(ids) == 0 {
		return templatesByID, nil
	}

	var templates []shipping.ShippingTemplate
	if err := r.db.Preload("Rules", func(db *gorm.DB) *gorm.DB {
		return db.Order("min_value_minor ASC, min_value ASC, id ASC")
	}).Where("id IN ?", ids).Find(&templates).Error; err != nil {
		return nil, err
	}

	for i := range templates {
		templatesByID[templates[i].ID] = &templates[i]
	}
	if err := r.attachShippingDisplayPriceSnapshots(templates); err != nil {
		return nil, err
	}
	for i := range templates {
		templatesByID[templates[i].ID] = &templates[i]
	}
	return templatesByID, nil
}

// FindAllTemplates 閺屻儲澹橀幍鈧張澶嬆侀弶?
func (r *ShippingRepository) FindAllTemplates() ([]shipping.ShippingTemplate, error) {
	var templates []shipping.ShippingTemplate
	err := r.db.Preload("Rules", func(db *gorm.DB) *gorm.DB {
		return db.Order("min_value_minor ASC, min_value ASC, id ASC")
	}).Find(&templates).Error
	if err == nil {
		err = r.attachShippingDisplayPriceSnapshots(templates)
	}
	return templates, err
}

func (r *ShippingRepository) CreateTemplateWithRules(template *shipping.ShippingTemplate, rules []shipping.ShippingRule) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		enabled := template.Enabled
		template.Rules = nil
		if err := tx.Create(template).Error; err != nil {
			return err
		}
		if !enabled {
			if err := tx.Model(template).Update("enabled", false).Error; err != nil {
				return err
			}
			template.Enabled = false
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
		if err := upsertShippingTemplateDisplayPriceSnapshot(tx, template); err != nil {
			return err
		}
		for i := range rules {
			if err := upsertShippingRuleDisplayPriceSnapshot(tx, &rules[i]); err != nil {
				return err
			}
		}

		if err := tx.Preload("Rules").First(template, template.ID).Error; err != nil {
			return err
		}
		hydrated := []shipping.ShippingTemplate{*template}
		if err := (&ShippingRepository{db: tx}).attachShippingDisplayPriceSnapshots(hydrated); err != nil {
			return err
		}
		*template = hydrated[0]
		return nil
	})
}

func (r *ShippingRepository) UpdateTemplateWithRules(template *shipping.ShippingTemplate, rules []shipping.ShippingRule) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"name":                 template.Name,
			"type":                 template.Type,
			"currency":             template.Currency,
			"free_shipping":        template.FreeShipping,
			"free_threshold_minor": template.FreeThresholdMinor,
			"default_fee_minor":    template.DefaultFeeMinor,
			"description":          template.Description,
			"enabled":              template.Enabled,
		}
		if err := tx.Model(&shipping.ShippingTemplate{}).Where("id = ?", template.ID).Updates(updates).Error; err != nil {
			return err
		}

		if err := tx.Where("template_id = ?", template.ID).Delete(&shipping.ShippingRule{}).Error; err != nil {
			return err
		}
		if tx.Migrator().HasTable(&shipping.ShippingDisplayPriceSnapshot{}) {
			if err := tx.Where("template_id = ?", template.ID).Delete(&shipping.ShippingDisplayPriceSnapshot{}).Error; err != nil {
				return err
			}
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
		if err := upsertShippingTemplateDisplayPriceSnapshot(tx, template); err != nil {
			return err
		}
		for i := range rules {
			if err := upsertShippingRuleDisplayPriceSnapshot(tx, &rules[i]); err != nil {
				return err
			}
		}

		if err := tx.Preload("Rules").First(template, template.ID).Error; err != nil {
			return err
		}
		hydrated := []shipping.ShippingTemplate{*template}
		if err := (&ShippingRepository{db: tx}).attachShippingDisplayPriceSnapshots(hydrated); err != nil {
			return err
		}
		*template = hydrated[0]
		return nil
	})
}

func (r *ShippingRepository) DeleteTemplate(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if tx.Migrator().HasTable(&shipping.ShippingDisplayPriceSnapshot{}) {
			if err := tx.Where("template_id = ?", id).Delete(&shipping.ShippingDisplayPriceSnapshot{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("template_id = ?", id).Delete(&shipping.ShippingRule{}).Error; err != nil {
			return err
		}
		return tx.Delete(&shipping.ShippingTemplate{}, id).Error
	})
}

// ShippingRule 閻╃鍙ч弬瑙勭《

// CreateRule 閸掓稑缂撴潻鎰瀭鐟欏嫬鍨?
func (r *ShippingRepository) CreateRule(rule *shipping.ShippingRule) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(rule).Error; err != nil {
			return err
		}
		return upsertShippingRuleDisplayPriceSnapshot(tx, rule)
	})
}

// FindRulesByTemplateID 閺嶈宓佸Ο鈩冩緲ID閺屻儲澹樼憴鍕灟
func (r *ShippingRepository) FindRulesByTemplateID(templateID uint) ([]shipping.ShippingRule, error) {
	var rules []shipping.ShippingRule
	err := r.db.Where("template_id = ?", templateID).Order("min_value_minor ASC, min_value ASC, id ASC").Find(&rules).Error
	if err == nil && len(rules) > 0 {
		wrapped := shipping.ShippingTemplate{ID: templateID, Rules: rules}
		if attachErr := r.attachShippingDisplayPriceSnapshots([]shipping.ShippingTemplate{wrapped}); attachErr != nil {
			err = attachErr
		} else {
			rules = wrapped.Rules
		}
	}
	return rules, err
}

// UpdateRule 閺囧瓨鏌婄憴鍕灟
func (r *ShippingRepository) UpdateRule(rule *shipping.ShippingRule) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(rule).Error; err != nil {
			return err
		}
		return upsertShippingRuleDisplayPriceSnapshot(tx, rule)
	})
}

func (r *ShippingRepository) UpdateRuleForTemplate(rule *shipping.ShippingRule) error {
	updates := map[string]interface{}{
		"region":           rule.Region,
		"currency":         rule.Currency,
		"min_value_minor":  rule.MinValueMinor,
		"max_value_minor":  rule.MaxValueMinor,
		"fee_minor":        rule.FeeMinor,
		"additional_minor": rule.AdditionalMinor,
		"template_id":      rule.TemplateID,
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&shipping.ShippingRule{}).
			Where("id = ? AND template_id = ?", rule.ID, rule.TemplateID).
			Updates(updates).Error; err != nil {
			return err
		}
		return upsertShippingRuleDisplayPriceSnapshot(tx, rule)
	})
}

type ShippingDisplayPriceSnapshotUpdate struct {
	TemplateID       uint
	DisplayPriceData datatypes.JSON
	RuleUpdates      []ShippingRuleDisplayPriceSnapshotUpdate
}

type ShippingRuleDisplayPriceSnapshotUpdate struct {
	RuleID           uint
	DisplayPriceData datatypes.JSON
}

func (r *ShippingRepository) UpdateDisplayPriceSnapshots(updates []ShippingDisplayPriceSnapshotUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if !tx.Migrator().HasTable(&shipping.ShippingDisplayPriceSnapshot{}) {
			return nil
		}
		for _, update := range updates {
			if update.TemplateID == 0 {
				continue
			}
			var template shipping.ShippingTemplate
			if err := tx.First(&template, update.TemplateID).Error; err != nil {
				return err
			}
			if err := upsertShippingTemplateDisplayPriceSnapshotWithData(tx, &template, update.DisplayPriceData); err != nil {
				return err
			}
			for _, ruleUpdate := range update.RuleUpdates {
				if ruleUpdate.RuleID == 0 {
					continue
				}
				var rule shipping.ShippingRule
				if err := tx.Where("id = ? AND template_id = ?", ruleUpdate.RuleID, update.TemplateID).First(&rule).Error; err != nil {
					return err
				}
				if err := upsertShippingRuleDisplayPriceSnapshotWithData(tx, &rule, ruleUpdate.DisplayPriceData); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func upsertShippingTemplateDisplayPriceSnapshot(tx *gorm.DB, template *shipping.ShippingTemplate) error {
	if template == nil || !tx.Migrator().HasTable(&shipping.ShippingDisplayPriceSnapshot{}) {
		return nil
	}
	return upsertShippingTemplateDisplayPriceSnapshotWithData(tx, template, template.DisplayPriceData)
}

func upsertShippingTemplateDisplayPriceSnapshotWithData(tx *gorm.DB, template *shipping.ShippingTemplate, displayPriceData datatypes.JSON) error {
	if template == nil || template.ID == 0 {
		return nil
	}
	if len(displayPriceData) == 0 {
		displayPriceData = datatypes.JSON([]byte("{}"))
	}
	snapshot := shipping.ShippingDisplayPriceSnapshot{
		ScopeKey:                 fmt.Sprintf("template:%d", template.ID),
		TemplateID:               template.ID,
		SourceCurrency:           template.Currency,
		SourceDefaultFeeMinor:    shippingInt64Ptr(template.DefaultFeeMinor),
		SourceFreeThresholdMinor: shippingInt64Ptr(template.FreeThresholdMinor),
		DisplayPriceData:         displayPriceData,
	}
	return upsertShippingDisplayPriceSnapshot(tx, snapshot)
}

func upsertShippingRuleDisplayPriceSnapshot(tx *gorm.DB, rule *shipping.ShippingRule) error {
	if rule == nil || !tx.Migrator().HasTable(&shipping.ShippingDisplayPriceSnapshot{}) {
		return nil
	}
	return upsertShippingRuleDisplayPriceSnapshotWithData(tx, rule, rule.DisplayPriceData)
}

func upsertShippingRuleDisplayPriceSnapshotWithData(tx *gorm.DB, rule *shipping.ShippingRule, displayPriceData datatypes.JSON) error {
	if rule == nil || rule.ID == 0 {
		return nil
	}
	if len(displayPriceData) == 0 {
		displayPriceData = datatypes.JSON([]byte("{}"))
	}
	snapshot := shipping.ShippingDisplayPriceSnapshot{
		ScopeKey:              fmt.Sprintf("rule:%d", rule.ID),
		TemplateID:            rule.TemplateID,
		RuleID:                shippingUintPtr(rule.ID),
		SourceCurrency:        rule.Currency,
		SourceMinValueMinor:   shippingInt64Ptr(rule.MinValueMinor),
		SourceMaxValueMinor:   shippingInt64Ptr(rule.MaxValueMinor),
		SourceFeeMinor:        shippingInt64Ptr(rule.FeeMinor),
		SourceAdditionalMinor: shippingInt64Ptr(rule.AdditionalMinor),
		DisplayPriceData:      displayPriceData,
	}
	return upsertShippingDisplayPriceSnapshot(tx, snapshot)
}

func upsertShippingDisplayPriceSnapshot(tx *gorm.DB, snapshot shipping.ShippingDisplayPriceSnapshot) error {
	var existing shipping.ShippingDisplayPriceSnapshot
	err := tx.Where("scope_key = ?", snapshot.ScopeKey).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return tx.Create(&snapshot).Error
	}
	if err != nil {
		return err
	}
	return tx.Model(&shipping.ShippingDisplayPriceSnapshot{}).
		Where("id = ?", existing.ID).
		Updates(map[string]interface{}{
			"template_id":                 snapshot.TemplateID,
			"rule_id":                     snapshot.RuleID,
			"source_currency":             snapshot.SourceCurrency,
			"source_default_fee_minor":    snapshot.SourceDefaultFeeMinor,
			"source_free_threshold_minor": snapshot.SourceFreeThresholdMinor,
			"source_min_value_minor":      snapshot.SourceMinValueMinor,
			"source_max_value_minor":      snapshot.SourceMaxValueMinor,
			"source_fee_minor":            snapshot.SourceFeeMinor,
			"source_additional_minor":     snapshot.SourceAdditionalMinor,
			"display_prices":              snapshot.DisplayPriceData,
		}).Error
}

func shippingInt64Ptr(value int64) *int64 { return &value }

func shippingUintPtr(value uint) *uint { return &value }

// DeleteRule 閸掔娀娅庣憴鍕灟
func (r *ShippingRepository) DeleteRule(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if tx.Migrator().HasTable(&shipping.ShippingDisplayPriceSnapshot{}) {
			if err := tx.Where("rule_id = ?", id).Delete(&shipping.ShippingDisplayPriceSnapshot{}).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&shipping.ShippingRule{}, id).Error
	})
}

func (r *ShippingRepository) DeleteRuleForTemplate(templateID uint, ruleID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if tx.Migrator().HasTable(&shipping.ShippingDisplayPriceSnapshot{}) {
			if err := tx.Where("rule_id = ? AND template_id = ?", ruleID, templateID).Delete(&shipping.ShippingDisplayPriceSnapshot{}).Error; err != nil {
				return err
			}
		}
		return tx.Where("id = ? AND template_id = ?", ruleID, templateID).Delete(&shipping.ShippingRule{}).Error
	})
}

// Carrier 閻╃鍙ч弬瑙勭《

// CreateCarrier 閸掓稑缂撻悧鈺傜ウ閸忣剙寰?
func (r *ShippingRepository) CreateCarrier(c *shipping.Carrier) error {
	enabled := c.Enabled
	if err := r.db.Create(c).Error; err != nil {
		return err
	}
	if !enabled {
		if err := r.db.Model(c).Update("enabled", false).Error; err != nil {
			return err
		}
		c.Enabled = false
	}
	return nil
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

// FindTrackingShipmentsByOrderID returns every non-deleted package shipment
// for an order, ordered by creation.
func (r *ShippingRepository) FindTrackingShipmentsByOrderID(orderID uint) ([]shipping.TrackingShipment, error) {
	var shipments []shipping.TrackingShipment
	err := r.db.
		Preload("Provider").
		Preload("Carrier").
		Preload("CarrierService").
		Preload("Mapping").
		Where("order_id = ?", orderID).
		Order("id ASC").
		Find(&shipments).Error
	return shipments, err
}

func (r *ShippingRepository) FindTrackingShipmentByOrderIDAndTrackingNumber(orderID uint, trackingNumber string) (*shipping.TrackingShipment, error) {
	var shipment shipping.TrackingShipment
	err := r.db.
		Preload("Provider").
		Preload("Carrier").
		Preload("CarrierService").
		Preload("Mapping").
		Where("order_id = ? AND tracking_number = ?", orderID, strings.TrimSpace(trackingNumber)).
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

func (r *ShippingRepository) FindTrackingShipmentByTrackingNumber(trackingNumber string) (*shipping.TrackingShipment, error) {
	trackingNumber = strings.TrimSpace(trackingNumber)
	if trackingNumber == "" {
		return nil, gorm.ErrRecordNotFound
	}

	var shipments []shipping.TrackingShipment
	err := r.db.
		Preload("Provider").
		Preload("Carrier").
		Preload("CarrierService").
		Preload("Mapping").
		Where("tracking_number = ?", trackingNumber).
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
		return nil, fmt.Errorf("multiple tracking shipments match tracking number %s", trackingNumber)
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
		// Keep the application time location for SQLite compatibility; Postgres
		// normalizes timestamptz comparisons itself.
		now := time.Now()
		query = query.Where(
			"(sync_status = ? OR (sync_status = ? AND (next_sync_at IS NULL OR next_sync_at <= ?)) OR (sync_status = ? AND next_sync_at <= ?) OR (sync_status = ? AND (sync_lease_owner = '' OR sync_lease_expires_at IS NULL OR sync_lease_expires_at <= ?)))",
			"pending",
			"failed",
			now,
			"synced",
			now,
			"syncing",
			now,
		)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	err := query.Find(&shipments).Error
	return shipments, err
}

func (r *ShippingRepository) ClaimDueTrackingShipments(limit int, now time.Time, owner string, leaseTimeout time.Duration) ([]shipping.TrackingShipment, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return nil, errors.New("tracking sync lease owner is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if leaseTimeout <= 0 {
		leaseTimeout = 10 * time.Minute
	}

	claimed := make([]shipping.TrackingShipment, 0, limit)
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var candidates []shipping.TrackingShipment
		query := tx.
			Preload("Provider").
			Preload("Carrier").
			Preload("CarrierService").
			Preload("Mapping").
			Where("enabled = ?", true).
			Where(trackingShipmentDueForClaimSQL, trackingShipmentDueForClaimArgs(now)...).
			Order("CASE WHEN sync_status = 'syncing' THEN 0 ELSE 1 END ASC").
			Order("COALESCE(sync_lease_expires_at, next_sync_at, created_at) ASC").
			Order("id ASC").
			Limit(limit)
		query = lockTrackingShipmentClaims(r.db, query)
		if err := query.Find(&candidates).Error; err != nil {
			return err
		}

		for index := range candidates {
			candidate := &candidates[index]
			expiresAt := now.Add(leaseTimeout)
			result := tx.Model(&shipping.TrackingShipment{}).
				Where("id = ? AND enabled = ?", candidate.ID, true).
				Where(trackingShipmentDueForClaimSQL, trackingShipmentDueForClaimArgs(now)...).
				Updates(map[string]interface{}{
					"sync_status":           "syncing",
					"sync_lease_owner":      owner,
					"sync_lease_generation": gorm.Expr("sync_lease_generation + 1"),
					"sync_lease_expires_at": expiresAt,
					"last_error":            "",
					"updated_at":            now,
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				continue
			}
			candidate.SyncStatus = "syncing"
			candidate.SyncLeaseOwner = owner
			candidate.SyncLeaseGeneration++
			candidate.SyncLeaseExpiresAt = &expiresAt
			candidate.LastError = ""
			candidate.UpdatedAt = now
			claimed = append(claimed, *candidate)
		}
		return nil
	})
	return claimed, err
}

func (r *ShippingRepository) ClaimTrackingShipmentForSync(orderID uint, trackingNumber string, now time.Time, owner string, leaseTimeout time.Duration) (*shipping.TrackingShipment, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	trackingNumber = strings.TrimSpace(trackingNumber)
	owner = strings.TrimSpace(owner)
	if orderID == 0 || trackingNumber == "" || owner == "" {
		return nil, errors.New("tracking sync claim input is incomplete")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if leaseTimeout <= 0 {
		leaseTimeout = 10 * time.Minute
	}

	var claimed *shipping.TrackingShipment
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var candidate shipping.TrackingShipment
		query := tx.
			Preload("Provider").
			Preload("Carrier").
			Preload("CarrierService").
			Preload("Mapping").
			Where("order_id = ? AND tracking_number = ? AND enabled = ?", orderID, trackingNumber, true)
		query = lockTrackingShipmentClaims(r.db, query)
		if err := query.First(&candidate).Error; err != nil {
			return err
		}
		if trackingShipmentHasActiveLease(&candidate, now) {
			return ErrTrackingSyncLeaseUnavailable
		}

		expiresAt := now.Add(leaseTimeout)
		result := tx.Model(&shipping.TrackingShipment{}).
			Where("id = ? AND enabled = ?", candidate.ID, true).
			Where("sync_status <> ? OR sync_lease_owner = '' OR sync_lease_expires_at IS NULL OR sync_lease_expires_at <= ?", "syncing", now).
			Updates(map[string]interface{}{
				"sync_status":           "syncing",
				"sync_lease_owner":      owner,
				"sync_lease_generation": gorm.Expr("sync_lease_generation + 1"),
				"sync_lease_expires_at": expiresAt,
				"last_error":            "",
				"updated_at":            now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrTrackingSyncLeaseUnavailable
		}
		candidate.SyncStatus = "syncing"
		candidate.SyncLeaseOwner = owner
		candidate.SyncLeaseGeneration++
		candidate.SyncLeaseExpiresAt = &expiresAt
		candidate.LastError = ""
		candidate.UpdatedAt = now
		claimed = &candidate
		return nil
	})
	return claimed, err
}

const trackingShipmentDueForClaimSQL = "(sync_status = ? OR (sync_status = ? AND (next_sync_at IS NULL OR next_sync_at <= ?)) OR (sync_status = ? AND next_sync_at <= ?) OR (sync_status = ? AND (sync_lease_owner = '' OR sync_lease_expires_at IS NULL OR sync_lease_expires_at <= ?)))"

func trackingShipmentDueForClaimArgs(now time.Time) []interface{} {
	return []interface{}{"pending", "failed", now, "synced", now, "syncing", now}
}

func trackingShipmentHasActiveLease(shipment *shipping.TrackingShipment, now time.Time) bool {
	return shipment != nil &&
		shipment.SyncStatus == "syncing" &&
		strings.TrimSpace(shipment.SyncLeaseOwner) != "" &&
		shipment.SyncLeaseExpiresAt != nil &&
		shipment.SyncLeaseExpiresAt.After(now)
}

func lockTrackingShipmentClaims(db *gorm.DB, query *gorm.DB) *gorm.DB {
	switch db.Dialector.Name() {
	case "postgres", "mysql":
		return query.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"})
	case "sqlserver":
		return query.Clauses(clause.Locking{Strength: "UPDATE"})
	default:
		return query
	}
}

func (r *ShippingRepository) UpsertTrackingShipment(shipment *shipping.TrackingShipment) error {
	existing, err := r.FindTrackingShipmentByOrderIDAndTrackingNumber(shipment.OrderID, shipment.TrackingNumber)
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
		updates["sync_lease_owner"] = ""
		updates["sync_lease_expires_at"] = nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&shipping.TrackingShipment{}).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
			return err
		}
		if !sourceUnchanged {
			return tx.Where("order_id = ? AND tracking_number = ?", existing.OrderID, existing.TrackingNumber).Delete(&shipping.TrackingEvent{}).Error
		}
		return nil
	})
}

func (r *ShippingRepository) RenewTrackingShipmentSyncLease(claim *shipping.TrackingShipment, now time.Time, leaseTimeout time.Duration) error {
	if claim == nil || claim.ID == 0 || strings.TrimSpace(claim.SyncLeaseOwner) == "" || claim.SyncLeaseGeneration <= 0 {
		return ErrTrackingSyncLeaseLost
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if leaseTimeout <= 0 {
		leaseTimeout = 10 * time.Minute
	}
	expiresAt := now.Add(leaseTimeout)
	result := r.db.Model(&shipping.TrackingShipment{}).
		Where(
			"id = ? AND sync_status = ? AND sync_lease_owner = ? AND sync_lease_generation = ? AND sync_lease_expires_at > ?",
			claim.ID,
			"syncing",
			claim.SyncLeaseOwner,
			claim.SyncLeaseGeneration,
			now,
		).
		Updates(map[string]interface{}{
			"sync_lease_expires_at": expiresAt,
			"updated_at":            now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTrackingSyncLeaseLost
	}
	claim.SyncLeaseExpiresAt = &expiresAt
	claim.UpdatedAt = now
	return nil
}

func (r *ShippingRepository) CompleteTrackingShipmentSync(claim *shipping.TrackingShipment, eventCount int, lastEventAt *time.Time, nextSyncAt *time.Time, now time.Time) error {
	if claim == nil || claim.ID == 0 || strings.TrimSpace(claim.SyncLeaseOwner) == "" || claim.SyncLeaseGeneration <= 0 {
		return ErrTrackingSyncLeaseLost
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	result := r.db.Model(&shipping.TrackingShipment{}).
		Where(
			"id = ? AND sync_status = ? AND sync_lease_owner = ? AND sync_lease_generation = ? AND sync_lease_expires_at > ?",
			claim.ID,
			"syncing",
			claim.SyncLeaseOwner,
			claim.SyncLeaseGeneration,
			now,
		).
		Updates(map[string]interface{}{
			"registration_status":   "registered",
			"sync_status":           "synced",
			"event_count":           eventCount,
			"last_event_at":         lastEventAt,
			"last_synced_at":        &now,
			"next_sync_at":          nextSyncAt,
			"last_error":            "",
			"sync_lease_owner":      "",
			"sync_lease_expires_at": nil,
			"updated_at":            now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTrackingSyncLeaseLost
	}
	return nil
}

func (r *ShippingRepository) FailTrackingShipmentSync(claim *shipping.TrackingShipment, lastError string, nextSyncAt *time.Time, now time.Time) error {
	if claim == nil || claim.ID == 0 || strings.TrimSpace(claim.SyncLeaseOwner) == "" || claim.SyncLeaseGeneration <= 0 {
		return ErrTrackingSyncLeaseLost
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	result := r.db.Model(&shipping.TrackingShipment{}).
		Where(
			"id = ? AND sync_status = ? AND sync_lease_owner = ? AND sync_lease_generation = ? AND sync_lease_expires_at > ?",
			claim.ID,
			"syncing",
			claim.SyncLeaseOwner,
			claim.SyncLeaseGeneration,
			now,
		).
		Updates(map[string]interface{}{
			"sync_status":           "failed",
			"last_synced_at":        &now,
			"next_sync_at":          nextSyncAt,
			"last_error":            lastError,
			"sync_lease_owner":      "",
			"sync_lease_expires_at": nil,
			"updated_at":            now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTrackingSyncLeaseLost
	}
	return nil
}

func (r *ShippingRepository) ApplyTrackingWebhookSyncSuccess(orderID uint, trackingNumber string, eventCount int, lastEventAt *time.Time, now time.Time) error {
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	result := r.db.Model(&shipping.TrackingShipment{}).
		Where("order_id = ? AND tracking_number = ?", orderID, strings.TrimSpace(trackingNumber)).
		Updates(map[string]interface{}{
			"registration_status":   "registered",
			"sync_status":           "synced",
			"event_count":           eventCount,
			"last_event_at":         lastEventAt,
			"last_synced_at":        &now,
			"next_sync_at":          nil,
			"last_error":            "",
			"sync_lease_owner":      "",
			"sync_lease_expires_at": nil,
			"updated_at":            now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *ShippingRepository) UpdateTrackingShipmentRegistrationStatusForTracking(orderID uint, trackingNumber string, status string, lastError string) error {
	now := time.Now().UTC()
	return r.db.Model(&shipping.TrackingShipment{}).
		Where("order_id = ? AND tracking_number = ?", orderID, strings.TrimSpace(trackingNumber)).
		Updates(map[string]interface{}{"registration_status": status, "last_error": lastError, "updated_at": now}).Error
}

// AreAllTrackingShipmentsDelivered reports whether every package currently
// registered for an order has a delivery event. An order with no package rows
// is never considered delivered.
func (r *ShippingRepository) AreAllTrackingShipmentsDelivered(orderID uint) (bool, error) {
	shipments, err := r.FindTrackingShipmentsByOrderID(orderID)
	if err != nil {
		return false, err
	}
	if len(shipments) == 0 {
		return false, nil
	}
	activeCount := 0
	for _, shipment := range shipments {
		if !shipment.Enabled {
			continue
		}
		activeCount++
		var count int64
		if err := r.db.Model(&shipping.TrackingEvent{}).
			Where("order_id = ? AND tracking_number = ?", orderID, shipment.TrackingNumber).
			Where("LOWER(status) LIKE ? OR LOWER(status) LIKE ? OR status LIKE ? OR status LIKE ?", "%delivered%", "%signed%", "%妥投%", "%签收%").
			Count(&count).Error; err != nil {
			return false, err
		}
		if count == 0 {
			return false, nil
		}
	}
	return activeCount > 0, nil
}

func trackingShipmentExternalSourceUnchanged(existing *shipping.TrackingShipment, next *shipping.TrackingShipment) bool {
	if existing == nil || next == nil {
		return false
	}
	return existing.TrackingProviderID == next.TrackingProviderID &&
		strings.TrimSpace(existing.TrackingNumber) == strings.TrimSpace(next.TrackingNumber) &&
		strings.TrimSpace(existing.ProviderCarrierCode) == strings.TrimSpace(next.ProviderCarrierCode)
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
	if err == nil {
		err = r.attachCarrierServiceTemplateSnapshots(services)
	}
	return services, err
}

func (r *ShippingRepository) FindEnabledCarrierServicesWithTemplates() ([]shipping.CarrierService, error) {
	var services []shipping.CarrierService
	err := r.db.
		Preload("Carrier").
		Preload("Template.Rules", func(db *gorm.DB) *gorm.DB {
			return db.Order("min_value_minor ASC, min_value ASC, id ASC")
		}).
		Preload("Template").
		Where("enabled = ?", true).
		Where("template_id IS NOT NULL").
		Order("sort_order ASC").
		Order("id ASC").
		Find(&services).Error
	if err == nil {
		err = r.attachCarrierServiceTemplateSnapshots(services)
	}
	return services, err
}

func (r *ShippingRepository) attachCarrierServiceTemplateSnapshots(services []shipping.CarrierService) error {
	if len(services) == 0 {
		return nil
	}
	templates := make([]shipping.ShippingTemplate, 0, len(services))
	indexes := make(map[uint]int, len(services))
	for i := range services {
		if services[i].Template == nil || services[i].Template.ID == 0 {
			continue
		}
		index := len(templates)
		templates = append(templates, *services[i].Template)
		indexes[services[i].Template.ID] = index
	}
	if err := r.attachShippingDisplayPriceSnapshots(templates); err != nil {
		return err
	}
	for i := range services {
		if services[i].Template == nil {
			continue
		}
		if index, ok := indexes[services[i].Template.ID]; ok {
			*services[i].Template = templates[index]
		}
	}
	return nil
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
	if err := r.attachCarrierServiceTemplateSnapshots([]shipping.CarrierService{service}); err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *ShippingRepository) CreateCarrierService(service *shipping.CarrierService) error {
	enabled := service.Enabled
	if err := r.db.Create(service).Error; err != nil {
		return err
	}
	if !enabled {
		if err := r.db.Model(service).Update("enabled", false).Error; err != nil {
			return err
		}
		service.Enabled = false
	}
	return nil
}

func (r *ShippingRepository) UpdateCarrierService(service *shipping.CarrierService) error {
	updates := map[string]interface{}{
		"carrier_id":                     service.CarrierID,
		"template_id":                    service.TemplateID,
		"service_code":                   service.ServiceCode,
		"service_name":                   service.ServiceName,
		"route_name":                     service.RouteName,
		"countries":                      service.Countries,
		"currency":                       service.Currency,
		"billing_mode":                   service.BillingMode,
		"first_weight_grams":             service.FirstWeightGrams,
		"additional_weight_grams":        service.AdditionalWeightGrams,
		"min_charge_weight_grams":        service.MinChargeWeightGrams,
		"volumetric_divisor":             service.VolumetricDivisor,
		"fuel_surcharge_percent_decimal": service.FuelSurchargePercentDecimal,
		"remote_surcharge_minor":         service.RemoteSurchargeMinor,
		"remote_postal_codes":            service.RemotePostalCodes,
		"eta_min_days":                   service.EtaMinDays,
		"eta_max_days":                   service.EtaMaxDays,
		"enabled":                        service.Enabled,
		"sort_order":                     service.SortOrder,
		"description":                    service.Description,
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

func (r *ShippingRepository) FindTrackingEventsByOrderIDAndTrackingNumber(orderID uint, trackingNumber string) ([]shipping.TrackingEvent, error) {
	var events []shipping.TrackingEvent
	err := r.db.Where("order_id = ? AND tracking_number = ?", orderID, strings.TrimSpace(trackingNumber)).Order("event_time DESC").Find(&events).Error
	return events, err
}

func (r *ShippingRepository) UpsertTrackingEvents(orderID uint, trackingNumber string, events []shipping.TrackingEvent) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if len(events) == 0 {
			return nil
		}
		return tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "order_id"},
				{Name: "tracking_number"},
				{Name: "event_time"},
				{Name: "status"},
			},
			DoNothing: true,
		}).Create(&events).Error
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
	enabled := zone.Enabled
	if err := r.db.Create(zone).Error; err != nil {
		return err
	}
	if !enabled {
		if err := r.db.Model(zone).Update("enabled", false).Error; err != nil {
			return err
		}
		zone.Enabled = false
	}
	return nil
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
	isActive := rule.IsActive
	if err := r.db.Create(rule).Error; err != nil {
		return err
	}
	if !isActive {
		if err := r.db.Model(rule).Update("is_active", false).Error; err != nil {
			return err
		}
		rule.IsActive = false
	}
	return nil
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

func (r *ShippingRepository) FindPackagingRuleApply(ruleID uint, productID uint, variantID ...*uint) (*shipping.PackagingRuleApply, error) {
	var apply shipping.PackagingRuleApply
	query := r.db.Where("rule_id = ? AND product_id = ?", ruleID, productID)
	if len(variantID) > 0 {
		if variantID[0] == nil {
			query = query.Where("variant_id IS NULL")
		} else {
			query = query.Where("variant_id = ?", *variantID[0])
		}
	}
	err := query.First(&apply).Error
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

// FindPackagingRuleApplyByProductAndVariant returns the binding for one
// product target. A nil variant ID addresses the product-level default.
func (r *ShippingRepository) FindPackagingRuleApplyByProductAndVariant(productID uint, variantID *uint) (*shipping.PackagingRuleApply, error) {
	var apply shipping.PackagingRuleApply
	query := r.db.Where("product_id = ?", productID)
	if variantID == nil {
		query = query.Where("variant_id IS NULL")
	} else {
		query = query.Where("variant_id = ?", *variantID)
	}
	if err := query.First(&apply).Error; err != nil {
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

	rulesByTarget, err := r.FindActivePackagingRulesByProductIDsAndVariants(productIDs)
	if err != nil {
		return nil, err
	}
	for productID, rules := range rulesByTarget {
		if rule := rules[0]; rule != nil {
			rulesByProduct[productID] = rule
			continue
		}
		// Preserve a useful result for legacy callers when a product only has
		// one variant-specific rule and no product-level default.
		var firstVariantID uint
		for variantID, rule := range rules {
			if variantID == 0 || rule == nil || (firstVariantID != 0 && variantID >= firstVariantID) {
				continue
			}
			firstVariantID = variantID
			rulesByProduct[productID] = rule
		}
	}
	return rulesByProduct, nil
}

// FindActivePackagingRulesByProductIDsAndVariants loads all active bindings
// for the requested products. The inner map uses variant ID 0 for the
// product-level default rule. This allows callers to apply the deterministic
// precedence of variant-specific rule over the product default.
func (r *ShippingRepository) FindActivePackagingRulesByProductIDsAndVariants(productIDs []uint) (map[uint]map[uint]*shipping.PackagingRule, error) {
	rulesByTarget := make(map[uint]map[uint]*shipping.PackagingRule)
	if len(productIDs) == 0 {
		return rulesByTarget, nil
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
		variantID := uint(0)
		if apply.VariantID != nil {
			variantID = *apply.VariantID
		}
		rules := rulesByTarget[apply.ProductID]
		if rules == nil {
			rules = make(map[uint]*shipping.PackagingRule)
			rulesByTarget[apply.ProductID] = rules
		}
		if _, exists := rules[variantID]; exists {
			if variantID == 0 {
				return nil, fmt.Errorf("product ID %d has multiple active default packaging rules", apply.ProductID)
			}
			return nil, fmt.Errorf("product ID %d variant ID %d has multiple active packaging rules", apply.ProductID, variantID)
		}
		rules[variantID] = apply.Rule
	}

	return rulesByTarget, nil
}

// FindPackagingRulesByProductID 閺嶈宓佹禍褍鎼D閺屻儲澹橀崠褰掑帳閻ㄥ嫭绺哄ú璇插瘶鐟佸懓顫夐弽鑹邦潐閸?
func (r *ShippingRepository) FindPackagingRulesByProductID(productID uint) ([]shipping.PackagingRule, error) {
	var rules []shipping.PackagingRule
	err := r.db.Joins("JOIN shipping_packaging_rule_applies ON shipping_packaging_rule_applies.rule_id = shipping_packaging_rules.id").
		Where("shipping_packaging_rule_applies.product_id = ? AND shipping_packaging_rules.is_active = ?", productID, true).
		Order("shipping_packaging_rule_applies.variant_id ASC, shipping_packaging_rule_applies.id DESC").
		Find(&rules).Error
	return rules, err
}
