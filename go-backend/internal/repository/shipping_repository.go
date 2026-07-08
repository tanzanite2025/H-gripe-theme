package repository

import (
	"tanzanite/internal/domain/shipping"

	"gorm.io/gorm"
)

type ShippingRepository struct {
	db *gorm.DB
}

func NewShippingRepository(db *gorm.DB) *ShippingRepository {
	return &ShippingRepository{db: db}
}

// ShippingTemplate 閻╃鍙ч弬瑙勭《

// FindTemplateByID 閺嶈宓両D閺屻儲澹樺Ο鈩冩緲
func (r *ShippingRepository) FindTemplateByID(id uint) (*shipping.ShippingTemplate, error) {
	var t shipping.ShippingTemplate
	err := r.db.Preload("Rules").First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// FindAllTemplates 閺屻儲澹橀幍鈧張澶嬆侀弶?
func (r *ShippingRepository) FindAllTemplates() ([]shipping.ShippingTemplate, error) {
	var templates []shipping.ShippingTemplate
	err := r.db.Preload("Rules").Find(&templates).Error
	return templates, err
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

// DeleteRule 閸掔娀娅庣憴鍕灟
func (r *ShippingRepository) DeleteRule(id uint) error {
	return r.db.Delete(&shipping.ShippingRule{}, id).Error
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

// FindZoneByCountry 閺嶈宓侀崶钘夘啀閺屻儲澹橀崠鍝勭厵
func (r *ShippingRepository) FindZoneByCountry(country string) (*shipping.ShippingZone, error) {
	var z shipping.ShippingZone
	err := r.db.Where("? = ANY(countries)", country).First(&z).Error
	if err != nil {
		return nil, err
	}
	return &z, nil
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

// DeletePackagingRuleApply 閸掔娀娅庨崠鍛邦棅鐟欏嫬鍨惃鍕安閻劋楠囬崫浣筋唶瑜?
func (r *ShippingRepository) DeletePackagingRuleApply(id uint) error {
	return r.db.Delete(&shipping.PackagingRuleApply{}, id).Error
}

// FindPackagingRulesByProductID 閺嶈宓佹禍褍鎼D閺屻儲澹橀崠褰掑帳閻ㄥ嫭绺哄ú璇插瘶鐟佸懓顫夐弽鑹邦潐閸?
func (r *ShippingRepository) FindPackagingRulesByProductID(productID uint) ([]shipping.PackagingRule, error) {
	var rules []shipping.PackagingRule
	err := r.db.Joins("JOIN shipping_packaging_rule_applies ON shipping_packaging_rule_applies.rule_id = shipping_packaging_rules.id").
		Where("shipping_packaging_rule_applies.product_id = ? AND shipping_packaging_rules.is_active = ?", productID, true).
		Find(&rules).Error
	return rules, err
}
