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

// ShippingTemplate 相关方法

// CreateTemplate 创建运费模板
func (r *ShippingRepository) CreateTemplate(t *shipping.ShippingTemplate) error {
	return r.db.Create(t).Error
}

// FindTemplateByID 根据ID查找模板
func (r *ShippingRepository) FindTemplateByID(id uint) (*shipping.ShippingTemplate, error) {
	var t shipping.ShippingTemplate
	err := r.db.Preload("Rules").First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// FindAllTemplates 查找所有模板
func (r *ShippingRepository) FindAllTemplates() ([]shipping.ShippingTemplate, error) {
	var templates []shipping.ShippingTemplate
	err := r.db.Preload("Rules").Find(&templates).Error
	return templates, err
}

// UpdateTemplate 更新模板
func (r *ShippingRepository) UpdateTemplate(t *shipping.ShippingTemplate) error {
	return r.db.Save(t).Error
}

// DeleteTemplate 删除模板
func (r *ShippingRepository) DeleteTemplate(id uint) error {
	// 先删除关联的规则
	if err := r.db.Where("template_id = ?", id).Delete(&shipping.ShippingRule{}).Error; err != nil {
		return err
	}
	return r.db.Delete(&shipping.ShippingTemplate{}, id).Error
}

// ShippingRule 相关方法

// CreateRule 创建运费规则
func (r *ShippingRepository) CreateRule(rule *shipping.ShippingRule) error {
	return r.db.Create(rule).Error
}

// FindRulesByTemplateID 根据模板ID查找规则
func (r *ShippingRepository) FindRulesByTemplateID(templateID uint) ([]shipping.ShippingRule, error) {
	var rules []shipping.ShippingRule
	err := r.db.Where("template_id = ?", templateID).Order("min_value ASC").Find(&rules).Error
	return rules, err
}

// UpdateRule 更新规则
func (r *ShippingRepository) UpdateRule(rule *shipping.ShippingRule) error {
	return r.db.Save(rule).Error
}

// DeleteRule 删除规则
func (r *ShippingRepository) DeleteRule(id uint) error {
	return r.db.Delete(&shipping.ShippingRule{}, id).Error
}

// Carrier 相关方法

// CreateCarrier 创建物流公司
func (r *ShippingRepository) CreateCarrier(c *shipping.Carrier) error {
	return r.db.Create(c).Error
}

// FindCarrierByID 根据ID查找物流公司
func (r *ShippingRepository) FindCarrierByID(id uint) (*shipping.Carrier, error) {
	var c shipping.Carrier
	err := r.db.First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// FindCarrierByCode 根据代码查找物流公司
func (r *ShippingRepository) FindCarrierByCode(code string) (*shipping.Carrier, error) {
	var c shipping.Carrier
	err := r.db.Where("code = ?", code).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// FindAllCarriers 查找所有物流公司
func (r *ShippingRepository) FindAllCarriers(enabledOnly bool) ([]shipping.Carrier, error) {
	var carriers []shipping.Carrier
	query := r.db.Order("name ASC")
	
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}
	
	err := query.Find(&carriers).Error
	return carriers, err
}

// UpdateCarrier 更新物流公司
func (r *ShippingRepository) UpdateCarrier(c *shipping.Carrier) error {
	return r.db.Save(c).Error
}

// DeleteCarrier 删除物流公司
func (r *ShippingRepository) DeleteCarrier(id uint) error {
	return r.db.Delete(&shipping.Carrier{}, id).Error
}

// TrackingEvent 相关方法

// CreateTrackingEvent 创建物流追踪事件
func (r *ShippingRepository) CreateTrackingEvent(e *shipping.TrackingEvent) error {
	return r.db.Create(e).Error
}

// FindTrackingEventsByOrderID 根据订单ID查找追踪事件
func (r *ShippingRepository) FindTrackingEventsByOrderID(orderID uint) ([]shipping.TrackingEvent, error) {
	var events []shipping.TrackingEvent
	err := r.db.Where("order_id = ?", orderID).Order("event_time DESC").Find(&events).Error
	return events, err
}

// FindTrackingEventsByTrackingNumber 根据追踪号查找事件
func (r *ShippingRepository) FindTrackingEventsByTrackingNumber(trackingNumber string) ([]shipping.TrackingEvent, error) {
	var events []shipping.TrackingEvent
	err := r.db.Where("tracking_number = ?", trackingNumber).Order("event_time DESC").Find(&events).Error
	return events, err
}

// ShippingZone 相关方法

// CreateZone 创建配送区域
func (r *ShippingRepository) CreateZone(z *shipping.ShippingZone) error {
	return r.db.Create(z).Error
}

// FindZoneByID 根据ID查找区域
func (r *ShippingRepository) FindZoneByID(id uint) (*shipping.ShippingZone, error) {
	var z shipping.ShippingZone
	err := r.db.First(&z, id).Error
	if err != nil {
		return nil, err
	}
	return &z, nil
}

// FindAllZones 查找所有区域
func (r *ShippingRepository) FindAllZones() ([]shipping.ShippingZone, error) {
	var zones []shipping.ShippingZone
	err := r.db.Order("name ASC").Find(&zones).Error
	return zones, err
}

// FindZoneByCountry 根据国家查找区域
func (r *ShippingRepository) FindZoneByCountry(country string) (*shipping.ShippingZone, error) {
	var z shipping.ShippingZone
	err := r.db.Where("? = ANY(countries)", country).First(&z).Error
	if err != nil {
		return nil, err
	}
	return &z, nil
}

// UpdateZone 更新区域
func (r *ShippingRepository) UpdateZone(z *shipping.ShippingZone) error {
	return r.db.Save(z).Error
}

// DeleteZone 删除区域
func (r *ShippingRepository) DeleteZone(id uint) error {
	return r.db.Delete(&shipping.ShippingZone{}, id).Error
}
